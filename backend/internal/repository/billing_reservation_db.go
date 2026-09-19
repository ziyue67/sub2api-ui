package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

// TryReserveUserBalanceDatabase is the durable Redis-outage fallback for the
// in-flight reservation ledger. A PostgreSQL transaction takes a per-scope
// advisory lock, expires stale rows, then inserts only when the active sum
// remains within maxTotal. This preserves the admission invariant while Redis
// is unavailable without pretending the database row is a wallet debit.
func (r *userRepository) TryReserveUserBalanceDatabase(
	ctx context.Context,
	scope string,
	requestID string,
	amount float64,
	maxTotal float64,
	ttl time.Duration,
) (float64, bool, error) {
	if r == nil || r.db == nil {
		return 0, false, errors.New("billing reservation database is unavailable")
	}
	if scope == "" || requestID == "" {
		return 0, false, errors.New("billing reservation scope and request id are required")
	}
	if err := validateBillingReservationAmounts(amount, maxTotal); err != nil {
		return 0, false, err
	}
	expiresAt := time.Now().Add(normalizedBillingReservationTTL(ttl))

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, false, err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, scope); err != nil {
		return 0, false, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM billing_reservations WHERE scope = $1 AND expires_at <= NOW()`, scope); err != nil {
		return 0, false, err
	}

	var current float64
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount), 0)::double precision
		FROM billing_reservations
		WHERE scope = $1 AND expires_at > NOW()
	`, scope).Scan(&current)
	if err != nil {
		return 0, false, err
	}
	if amount > 0 && current > maxTotal+1e-12 {
		return current, false, nil
	}
	if current+amount > maxTotal+1e-12 {
		return current, false, nil
	}

	var inserted bool
	err = tx.QueryRowContext(ctx, `
		INSERT INTO billing_reservations (scope, request_id, amount, expires_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (scope, request_id) DO NOTHING
		RETURNING TRUE
	`, scope, requestID, amount, expiresAt).Scan(&inserted)
	if errors.Is(err, sql.ErrNoRows) {
		var existing float64
		if err := tx.QueryRowContext(ctx, `
			SELECT amount::double precision
			FROM billing_reservations
			WHERE scope = $1 AND request_id = $2 AND expires_at > NOW()
		`, scope, requestID).Scan(&existing); err != nil {
			return 0, false, err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE billing_reservations
			SET expires_at = $3, updated_at = NOW()
			WHERE scope = $1 AND request_id = $2
		`, scope, requestID, expiresAt); err != nil {
			return 0, false, err
		}
		if err := extendBillingReservationFallback(ctx, tx, expiresAt); err != nil {
			return 0, false, err
		}
		if err := tx.Commit(); err != nil {
			return 0, false, err
		}
		return current, true, nil
	}
	if err != nil {
		return 0, false, err
	}
	if !inserted {
		return 0, false, errors.New("billing reservation insert did not return a row")
	}

	if err := extendBillingReservationFallback(ctx, tx, expiresAt); err != nil {
		return 0, false, err
	}
	if err := tx.Commit(); err != nil {
		return 0, false, err
	}
	return current + amount, true, nil
}

// ReleaseUserBalanceReservation removes a durable fallback reservation.
// A missing row is an expired lease, matching the Redis receipt contract.
func (r *userRepository) ReleaseUserBalanceReservation(
	ctx context.Context,
	scope string,
	requestID string,
	_ float64,
	_ time.Duration,
) error {
	if r == nil || r.db == nil {
		return errors.New("billing reservation database is unavailable")
	}
	if scope == "" || requestID == "" {
		return nil
	}
	var deleted bool
	err := r.db.QueryRowContext(ctx, `
		DELETE FROM billing_reservations
		WHERE scope = $1 AND request_id = $2
		RETURNING TRUE
	`, scope, requestID).Scan(&deleted)
	if errors.Is(err, sql.ErrNoRows) {
		return service.ErrBillingReservationExpired
	}
	if err != nil {
		return err
	}
	return nil
}

// RenewUserBalanceReservation extends a durable fallback lease and keeps the
// global fallback window open until every row that could exist has expired.
func (r *userRepository) RenewUserBalanceReservation(
	ctx context.Context,
	scope string,
	requestID string,
	ttl time.Duration,
) error {
	if r == nil || r.db == nil {
		return errors.New("billing reservation database is unavailable")
	}
	if scope == "" || requestID == "" {
		return nil
	}
	expiresAt := time.Now().Add(normalizedBillingReservationTTL(ttl))
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var renewed bool
	err = tx.QueryRowContext(ctx, `
		UPDATE billing_reservations
		SET expires_at = $3, updated_at = NOW()
		WHERE scope = $1 AND request_id = $2 AND expires_at > NOW()
		RETURNING TRUE
	`, scope, requestID, expiresAt).Scan(&renewed)
	if errors.Is(err, sql.ErrNoRows) {
		return service.ErrBillingReservationExpired
	}
	if err != nil {
		return err
	}
	if err := extendBillingReservationFallback(ctx, tx, expiresAt); err != nil {
		return err
	}
	return tx.Commit()
}

// ActivateBillingReservationDatabaseFallback raises the shared fallback window.
// All application instances consult this state before trusting Redis, which
// avoids overlapping the Redis and database reservation ledgers during recovery.
func (r *userRepository) ActivateBillingReservationDatabaseFallback(ctx context.Context, ttl time.Duration) error {
	if r == nil || r.db == nil {
		return errors.New("billing reservation database is unavailable")
	}
	until := time.Now().Add(normalizedBillingReservationTTL(ttl))
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO billing_reservation_fallback_state (singleton, fallback_until, updated_at)
		VALUES (TRUE, $1, NOW())
		ON CONFLICT (singleton) DO UPDATE
		SET fallback_until = GREATEST(billing_reservation_fallback_state.fallback_until, EXCLUDED.fallback_until),
			updated_at = NOW()
	`, until)
	return err
}

// BillingReservationDatabaseFallbackUntil returns the shared fallback deadline.
func (r *userRepository) BillingReservationDatabaseFallbackUntil(ctx context.Context) (time.Time, error) {
	if r == nil || r.db == nil {
		return time.Time{}, errors.New("billing reservation database is unavailable")
	}
	var until time.Time
	err := r.db.QueryRowContext(ctx, `
		SELECT fallback_until
		FROM billing_reservation_fallback_state
		WHERE singleton = TRUE
	`).Scan(&until)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, nil
	}
	return until, err
}

func extendBillingReservationFallback(ctx context.Context, tx *sql.Tx, until time.Time) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO billing_reservation_fallback_state (singleton, fallback_until, updated_at)
		VALUES (TRUE, $1, NOW())
		ON CONFLICT (singleton) DO UPDATE
		SET fallback_until = GREATEST(billing_reservation_fallback_state.fallback_until, EXCLUDED.fallback_until),
			updated_at = NOW()
	`, until)
	return err
}

func validateBillingReservationAmounts(amount, maxTotal float64) error {
	if amount < 0 || maxTotal < 0 ||
		math.IsNaN(amount) || math.IsNaN(maxTotal) ||
		math.IsInf(amount, 0) || math.IsInf(maxTotal, 0) {
		return fmt.Errorf("billing reservation amount and limit must be finite and nonnegative, got amount=%v limit=%v", amount, maxTotal)
	}
	return nil
}

func normalizedBillingReservationTTL(ttl time.Duration) time.Duration {
	if ttl <= 0 {
		return 10 * time.Minute
	}
	return ttl
}

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

// DB fallback timing. The DB ledger is only used after Redis reservation calls
// fail, so these values do not affect the normal hot path.
const (
	billingReservationDBOperationTimeout    = 3 * time.Second
	billingReservationFallbackProbeInterval = time.Second
	billingReservationFallbackProbeTimeout  = 500 * time.Millisecond
	billingReservationFallbackDefaultRowTTL = 10 * time.Minute
)

// tryReserve dispatches a durable reservation to the scope-specific ledger.
//
// Numeric scopes are the wallet scopes and use billing_balance_reservations,
// which locks the users row so the admission decision stays serialized with
// the balance settlement path. All other scopes (subscriptions, user/platform
// quotas, and API-key quotas) use billing_scope_reservations with a per-scope
// advisory transaction lock.
func (s *dbReservationStore) tryReserve(
	ctx context.Context,
	scope string,
	requestID string,
	amount float64,
	maxTotal float64,
	ttl time.Duration,
) (float64, bool, error) {
	if s == nil || s.db == nil {
		return 0, false, errors.New("db reservation store unavailable")
	}
	if scope == "" || requestID == "" {
		return 0, false, errors.New("db reservation scope and request id are required")
	}
	if err := validateBillingReservationAmounts(amount, maxTotal); err != nil {
		return 0, false, err
	}
	if _, ok := reservationUserIDFromScope(scope); ok {
		return s.tryReserveUserBalance(ctx, scope, requestID, amount, maxTotal, ttl)
	}
	return s.tryReserveScopeReservation(ctx, scope, requestID, amount, maxTotal, ttl)
}

func (s *dbReservationStore) release(ctx context.Context, scope, requestID string) error {
	if s == nil || s.db == nil {
		return errors.New("db reservation store unavailable")
	}
	if _, ok := reservationUserIDFromScope(scope); ok {
		return s.releaseUserBalanceReservation(ctx, scope, requestID)
	}
	return s.releaseScopeReservation(ctx, scope, requestID)
}

func (s *dbReservationStore) renew(ctx context.Context, scope, requestID string, ttl time.Duration) error {
	if s == nil || s.db == nil {
		return errors.New("db reservation store unavailable")
	}
	if _, ok := reservationUserIDFromScope(scope); ok {
		return s.renewUserBalanceReservation(ctx, scope, requestID, ttl)
	}
	return s.renewScopeReservation(ctx, scope, requestID, ttl)
}

func (s *dbReservationStore) tryReserveScopeReservation(
	ctx context.Context,
	scope string,
	requestID string,
	amount float64,
	maxTotal float64,
	ttl time.Duration,
) (float64, bool, error) {
	if amount <= 0 {
		return 0, true, nil
	}
	expiresAt := time.Now().Add(normalizedBillingReservationTTL(ttl))

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, false, err
	}
	defer func() { _ = tx.Rollback() }()

	// Serialize every admission for the same scope. The SQL below performs a
	// read-modify-write over the active reservation sum, so the lock is the
	// database equivalent of the Redis Lua script's atomicity.
	if _, err := tx.ExecContext(ctx,
		`SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`,
		scope,
	); err != nil {
		return 0, false, err
	}
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM billing_scope_reservations WHERE scope = $1 AND expires_at <= NOW()`,
		scope,
	); err != nil {
		return 0, false, err
	}

	// A retry with the same request ID must only refresh its receipt. It must
	// not add the amount to the active sum a second time.
	var existing float64
	err = tx.QueryRowContext(ctx, `
		SELECT amount::double precision
		FROM billing_scope_reservations
		WHERE scope = $1 AND request_id = $2 AND expires_at > NOW()
	`, scope, requestID).Scan(&existing)
	switch {
	case err == nil:
		if _, err := tx.ExecContext(ctx, `
			UPDATE billing_scope_reservations
			SET expires_at = $3
			WHERE scope = $1 AND request_id = $2
		`, scope, requestID, expiresAt); err != nil {
			return 0, false, err
		}
		if err := extendBillingReservationFallbackTx(ctx, tx, expiresAt); err != nil {
			return 0, false, err
		}
		var current float64
		if err := tx.QueryRowContext(ctx, `
			SELECT COALESCE(SUM(amount), 0)::double precision
			FROM billing_scope_reservations
			WHERE scope = $1 AND expires_at > NOW()
		`, scope).Scan(&current); err != nil {
			return 0, false, err
		}
		if err := tx.Commit(); err != nil {
			return 0, false, err
		}
		return current, true, nil
	case !errors.Is(err, sql.ErrNoRows):
		return 0, false, err
	}

	var current float64
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount), 0)::double precision
		FROM billing_scope_reservations
		WHERE scope = $1 AND expires_at > NOW()
	`, scope).Scan(&current); err != nil {
		return 0, false, err
	}
	if current+amount > maxTotal+1e-12 {
		return current, false, nil
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO billing_scope_reservations (scope, request_id, amount, expires_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (scope, request_id) DO UPDATE
		SET expires_at = EXCLUDED.expires_at
	`, scope, requestID, amount, expiresAt); err != nil {
		return 0, false, err
	}
	if err := extendBillingReservationFallbackTx(ctx, tx, expiresAt); err != nil {
		return 0, false, err
	}
	if err := tx.Commit(); err != nil {
		return 0, false, err
	}
	return current + amount, true, nil
}

func (s *dbReservationStore) releaseScopeReservation(ctx context.Context, scope, requestID string) error {
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM billing_scope_reservations
		WHERE scope = $1 AND request_id = $2
	`, scope, requestID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return service.ErrBillingReservationExpired
	}
	return nil
}

func (s *dbReservationStore) renewScopeReservation(ctx context.Context, scope, requestID string, ttl time.Duration) error {
	expiresAt := time.Now().Add(normalizedBillingReservationTTL(ttl))
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.ExecContext(ctx, `
		UPDATE billing_scope_reservations
		SET expires_at = $3
		WHERE scope = $1 AND request_id = $2 AND expires_at > NOW()
	`, scope, requestID, expiresAt)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return service.ErrBillingReservationExpired
	}
	if err := extendBillingReservationFallbackTx(ctx, tx, expiresAt); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *dbReservationStore) activateFallbackWindow(ctx context.Context, ttl time.Duration) error {
	until := time.Now().Add(normalizedBillingReservationTTL(ttl))
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO billing_reservation_fallback_state (singleton, fallback_until, updated_at)
		VALUES (TRUE, $1, NOW())
		ON CONFLICT (singleton) DO UPDATE
		SET fallback_until = GREATEST(billing_reservation_fallback_state.fallback_until, EXCLUDED.fallback_until),
			updated_at = NOW()
	`, until)
	return err
}

func (s *dbReservationStore) fallbackWindowUntil(ctx context.Context) (time.Time, error) {
	var until time.Time
	err := s.db.QueryRowContext(ctx, `
		SELECT fallback_until
		FROM billing_reservation_fallback_state
		WHERE singleton = TRUE
	`).Scan(&until)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, nil
	}
	return until, err
}

func extendBillingReservationFallbackTx(ctx context.Context, tx *sql.Tx, until time.Time) error {
	if tx == nil {
		return errors.New("nil reservation fallback transaction")
	}
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
		return billingReservationFallbackDefaultRowTTL
	}
	return ttl
}

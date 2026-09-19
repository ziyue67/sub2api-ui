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

// billingReservationDBStore 是 Redis 在途预留不可用时的 DB 回落实现。
//
// 为什么需要它：Redis 故障时原来的实现是 fail-open（并发原子性直接消失），
// new-api 在同类场景下回落到 DB 条件更新（`WHERE quota >= ?`）。这里用一张
// billing_reservations 表 + scope 级 advisory lock 达到同样的"判定与写入原子"
// 语义：同一 scope 的并发预留被串行化，聚合总额超过预算的请求直接被拒。
//
// 与 Redis 实现的差异：行有 expires_at 自愈 TTL（默认 10 分钟），凭据丢失时额度
// 最多滞留一个 TTL；续期由 BillingReservationSlot 的心跳驱动，与 Redis 一致。
type billingReservationDBStore struct {
	db *sql.DB
}

// NewBillingReservationDBStore 构造 DB 回落预留仓储；db 为 nil 时返回 nil（不装配回落）。
func NewBillingReservationDBStore(db *sql.DB) service.ReservationFallbackStore {
	if db == nil {
		return nil
	}
	return &billingReservationDBStore{db: db}
}

// TryReserveUserBalance 在事务内先取 scope 级 advisory lock（串行化同一 scope 的并发预留），
// 清理逗期行、汇总当前占用，再决定是否写入本请求的凭据。
// 返回 (当前总额, 是否接受, error)；请求已存在时幂等返回当前总额。
func (s *billingReservationDBStore) TryReserveUserBalance(ctx context.Context, scope, requestID string, amount, maxTotal float64, ttl time.Duration) (float64, bool, error) {
	if s == nil || s.db == nil {
		return 0, false, errors.New("billing reservation db store not configured")
	}
	if amount < 0 || maxTotal < 0 {
		return 0, false, fmt.Errorf("reserve amount and limit must be nonnegative, got amount=%v limit=%v", amount, maxTotal)
	}
	if scope == "" || requestID == "" {
		return 0, false, errors.New("reserve scope and requestID must not be empty")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, false, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// 同 scope 串行化：并发请求在这里排队，后面的请求能看到前一笔已提交的占用。
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, scope); err != nil {
		return 0, false, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM billing_reservations WHERE scope = $1 AND expires_at <= NOW()`, scope); err != nil {
		return 0, false, err
	}
	var current float64
	if err := tx.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount), 0)::float8 FROM billing_reservations WHERE scope = $1`, scope).Scan(&current); err != nil {
		return 0, false, err
	}

	// 幂等重放：同一 requestID 重复预留不重复累加。
	var existing float64
	switch err := tx.QueryRowContext(ctx,
		`SELECT amount FROM billing_reservations WHERE scope = $1 AND request_id = $2`, scope, requestID).Scan(&existing); {
	case err == nil:
		if err := tx.Commit(); err != nil {
			return 0, false, err
		}
		return current, true, nil
	case !errors.Is(err, sql.ErrNoRows):
		return 0, false, err
	}

	if current+amount > maxTotal+1e-12 {
		if err := tx.Commit(); err != nil {
			return 0, false, err
		}
		return current, false, nil
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO billing_reservations (scope, request_id, amount, expires_at) VALUES ($1, $2, $3, NOW() + $4::interval)`,
		scope, requestID, amount, fmt.Sprintf("%d milliseconds", ttl.Milliseconds())); err != nil {
		return 0, false, err
	}
	if err := tx.Commit(); err != nil {
		return 0, false, err
	}
	return current + amount, true, nil
}

// ReserveUserBalance 是非原子接口（无预算上限）的适配：等价于"预算无限"的预留。
func (s *billingReservationDBStore) ReserveUserBalance(ctx context.Context, scope, requestID string, amount float64, ttl time.Duration) (float64, error) {
	total, _, err := s.TryReserveUserBalance(ctx, scope, requestID, amount, math.MaxFloat64, ttl)
	return total, err
}

// ReleaseUserBalanceReservation 只释放本请求的凭据；凭据已不存在（TTL 已回收）时
// 返回 service.ErrBillingReservationExpired，与 Redis 实现语义一致（不得触碰别人的预留）。
func (s *billingReservationDBStore) ReleaseUserBalanceReservation(ctx context.Context, scope, requestID string, _ float64, _ time.Duration) error {
	if s == nil || s.db == nil {
		return errors.New("billing reservation db store not configured")
	}
	res, err := s.db.ExecContext(ctx, `DELETE FROM billing_reservations WHERE scope = $1 AND request_id = $2`, scope, requestID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return service.ErrBillingReservationExpired
	}
	return nil
}

// RenewUserBalanceReservation 为长请求续期；凭据不存在时返回 service.ErrBillingReservationExpired。
func (s *billingReservationDBStore) RenewUserBalanceReservation(ctx context.Context, scope, requestID string, ttl time.Duration) error {
	if s == nil || s.db == nil {
		return errors.New("billing reservation db store not configured")
	}
	res, err := s.db.ExecContext(ctx,
		`UPDATE billing_reservations SET expires_at = NOW() + $3::interval WHERE scope = $1 AND request_id = $2`,
		scope, requestID, fmt.Sprintf("%d milliseconds", ttl.Milliseconds()))
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return service.ErrBillingReservationExpired
	}
	return nil
}

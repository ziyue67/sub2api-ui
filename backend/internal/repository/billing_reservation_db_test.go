package repository

import (
	"context"
	"database/sql"
	"math"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

// newBillingReservationRepoMock 构造以 sqlmock 为后端的 userRepository：
// DB 兜底预留（审计 R2）必须走真实的 SQL 形状（advisory 锁 + 条件插入 + 窗口抬升），
// 这里用 sqlmock 锁死这些语句与失败语义。
func newBillingReservationRepoMock(t *testing.T) (*userRepository, sqlmock.Sqlmock) {
	t.Helper()
	repo, mock := newRedeemAdjustmentRepoMock(t)
	require.NotNil(t, repo.db, "userRepository 必须持有 *sql.DB 才能提供 DB 兜底预留")
	return repo, mock
}

// TestTryReserveUserBalanceDatabase_AcceptsWithinBudget 覆盖接受路径：
// 同一 scope 上先取 advisory 锁、清过期行、求活跃合计，再条件插入，
// 成功时必须把共享兜底窗口一起抬到本笔凭据的过期时刻。
func TestTryReserveUserBalanceDatabase_AcceptsWithinBudget(t *testing.T) {
	repo, mock := newBillingReservationRepoMock(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(`SELECT pg_advisory_xact_lock\(hashtextextended\(\$1, 0\)\)`).
		WithArgs("7").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM billing_reservations WHERE scope = \$1 AND expires_at <= NOW\(\)`).
		WithArgs("7").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(amount\), 0\)::double precision FROM billing_reservations WHERE scope = \$1 AND expires_at > NOW\(\)`).
		WithArgs("7").
		WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(0.10))
	mock.ExpectQuery(`INSERT INTO billing_reservations \(scope, request_id, amount, expires_at\) VALUES \(\$1, \$2, \$3, \$4\) ON CONFLICT \(scope, request_id\) DO NOTHING RETURNING TRUE`).
		WithArgs("7", "req-1", 0.10, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"bool"}).AddRow(true))
	mock.ExpectExec(`INSERT INTO billing_reservation_fallback_state \(singleton, fallback_until, updated_at\) VALUES \(TRUE, \$1, NOW\(\)\) ON CONFLICT \(singleton\) DO UPDATE SET fallback_until = GREATEST\(billing_reservation_fallback_state.fallback_until, EXCLUDED.fallback_until\), updated_at = NOW\(\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	total, accepted, err := repo.TryReserveUserBalanceDatabase(ctx, "7", "req-1", 0.10, 0.35, 10*time.Minute)
	require.NoError(t, err)
	require.True(t, accepted)
	require.InDelta(t, 0.20, total, 1e-9, "接受时返回累加后的总额")
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestTryReserveUserBalanceDatabase_RejectsOverBudget 覆盖拒绝路径：
// 活跃合计 + 本笔超过上限时不得插入任何行（事务回滚），返回累加前的总额。
func TestTryReserveUserBalanceDatabase_RejectsOverBudget(t *testing.T) {
	repo, mock := newBillingReservationRepoMock(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(`SELECT pg_advisory_xact_lock\(hashtextextended\(\$1, 0\)\)`).
		WithArgs("7").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM billing_reservations WHERE scope = \$1 AND expires_at <= NOW\(\)`).
		WithArgs("7").
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(amount\), 0\)::double precision FROM billing_reservations WHERE scope = \$1 AND expires_at > NOW\(\)`).
		WithArgs("7").
		WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(0.30))
	mock.ExpectRollback()

	total, accepted, err := repo.TryReserveUserBalanceDatabase(ctx, "7", "req-4", 0.10, 0.35, 10*time.Minute)
	require.NoError(t, err)
	require.False(t, accepted, "超出上限必须拒绝")
	require.InDelta(t, 0.30, total, 1e-9, "拒绝时返回累加前的总额（调用方据此做 guard 判定）")
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestTryReserveUserBalanceDatabase_IdempotentReplay 覆盖凭据重放：
// 同 requestID 再次预留只刷新 TTL 与共享窗口，不重复累加金额。
func TestTryReserveUserBalanceDatabase_IdempotentReplay(t *testing.T) {
	repo, mock := newBillingReservationRepoMock(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(`SELECT pg_advisory_xact_lock\(hashtextextended\(\$1, 0\)\)`).
		WithArgs("7").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM billing_reservations WHERE scope = \$1 AND expires_at <= NOW\(\)`).
		WithArgs("7").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(amount\), 0\)::double precision FROM billing_reservations WHERE scope = \$1 AND expires_at > NOW\(\)`).
		WithArgs("7").
		WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(0.10))
	mock.ExpectQuery(`INSERT INTO billing_reservations \(scope, request_id, amount, expires_at\) VALUES \(\$1, \$2, \$3, \$4\) ON CONFLICT \(scope, request_id\) DO NOTHING RETURNING TRUE`).
		WithArgs("7", "req-1", 0.10, sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`SELECT amount::double precision FROM billing_reservations WHERE scope = \$1 AND request_id = \$2 AND expires_at > NOW\(\)`).
		WithArgs("7", "req-1").
		WillReturnRows(sqlmock.NewRows([]string{"amount"}).AddRow(0.10))
	mock.ExpectExec(`UPDATE billing_reservations SET expires_at = \$3, updated_at = NOW\(\) WHERE scope = \$1 AND request_id = \$2`).
		WithArgs("7", "req-1", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO billing_reservation_fallback_state \(singleton, fallback_until, updated_at\) VALUES \(TRUE, \$1, NOW\(\)\) ON CONFLICT \(singleton\) DO UPDATE SET fallback_until = GREATEST\(billing_reservation_fallback_state.fallback_until, EXCLUDED.fallback_until\), updated_at = NOW\(\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	total, accepted, err := repo.TryReserveUserBalanceDatabase(ctx, "7", "req-1", 0.10, 0.35, 10*time.Minute)
	require.NoError(t, err)
	require.True(t, accepted, "凭据重放是幂等接受，不得重复占用额度")
	require.InDelta(t, 0.10, total, 1e-9)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestReleaseUserBalanceReservation_ExpiredReceiptIsInert 覆盖归还契约：
// 凭据已随 TTL 过期（行不存在）时返回 ErrBillingReservationExpired，
// 调用方据此跳过递减，绝不触碰他人的预留。
func TestReleaseUserBalanceReservation_ExpiredReceiptIsInert(t *testing.T) {
	repo, mock := newBillingReservationRepoMock(t)

	mock.ExpectQuery(`DELETE FROM billing_reservations WHERE scope = \$1 AND request_id = \$2 RETURNING TRUE`).
		WithArgs("7", "req-gone").
		WillReturnError(sql.ErrNoRows)

	err := repo.ReleaseUserBalanceReservation(context.Background(), "7", "req-gone", 0.10, 10*time.Minute)
	require.ErrorIs(t, err, service.ErrBillingReservationExpired)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestReleaseUserBalanceReservation_DeletesReceipt 覆盖正常归还：按 (scope, request_id) 精确删除。
func TestReleaseUserBalanceReservation_DeletesReceipt(t *testing.T) {
	repo, mock := newBillingReservationRepoMock(t)

	mock.ExpectQuery(`DELETE FROM billing_reservations WHERE scope = \$1 AND request_id = \$2 RETURNING TRUE`).
		WithArgs("7", "req-1").
		WillReturnRows(sqlmock.NewRows([]string{"bool"}).AddRow(true))

	require.NoError(t, repo.ReleaseUserBalanceReservation(context.Background(), "7", "req-1", 0.10, 10*time.Minute))
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRenewUserBalanceReservation_ExtendsReceiptAndWindow 覆盖长请求续期：
// 只续期仍活跃的凭据（expires_at > NOW()），并同步抬高共享兜底窗口。
func TestRenewUserBalanceReservation_ExtendsReceiptAndWindow(t *testing.T) {
	repo, mock := newBillingReservationRepoMock(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`UPDATE billing_reservations SET expires_at = \$3, updated_at = NOW\(\) WHERE scope = \$1 AND request_id = \$2 AND expires_at > NOW\(\) RETURNING TRUE`).
		WithArgs("7", "req-1", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"bool"}).AddRow(true))
	mock.ExpectExec(`INSERT INTO billing_reservation_fallback_state \(singleton, fallback_until, updated_at\) VALUES \(TRUE, \$1, NOW\(\)\) ON CONFLICT \(singleton\) DO UPDATE SET fallback_until = GREATEST\(billing_reservation_fallback_state.fallback_until, EXCLUDED.fallback_until\), updated_at = NOW\(\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, repo.RenewUserBalanceReservation(context.Background(), "7", "req-1", 10*time.Minute))
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRenewUserBalanceReservation_ExpiredReceipt 覆盖"续期已过期凭据"：
// 返回 ErrBillingReservationExpired，调用方停止心跳。
func TestRenewUserBalanceReservation_ExpiredReceipt(t *testing.T) {
	repo, mock := newBillingReservationRepoMock(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`UPDATE billing_reservations SET expires_at = \$3, updated_at = NOW\(\) WHERE scope = \$1 AND request_id = \$2 AND expires_at > NOW\(\) RETURNING TRUE`).
		WithArgs("7", "req-gone", sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	err := repo.RenewUserBalanceReservation(context.Background(), "7", "req-gone", 10*time.Minute)
	require.ErrorIs(t, err, service.ErrBillingReservationExpired)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestBillingReservationDatabaseFallbackWindow 覆盖共享窗口的抬升与读取：
// 抬升用 GREATEST 单调递增（并发/多实例同时抬升不会回退），读取返回当前截止时间。
func TestBillingReservationDatabaseFallbackWindow(t *testing.T) {
	repo, mock := newBillingReservationRepoMock(t)
	ctx := context.Background()

	mock.ExpectExec(`INSERT INTO billing_reservation_fallback_state \(singleton, fallback_until, updated_at\) VALUES \(TRUE, \$1, NOW\(\)\) ON CONFLICT \(singleton\) DO UPDATE SET fallback_until = GREATEST\(billing_reservation_fallback_state.fallback_until, EXCLUDED.fallback_until\), updated_at = NOW\(\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.ActivateBillingReservationDatabaseFallback(ctx, time.Minute))

	until := time.Now().Add(42 * time.Second).UTC()
	mock.ExpectQuery(`SELECT fallback_until FROM billing_reservation_fallback_state WHERE singleton = TRUE`).
		WillReturnRows(sqlmock.NewRows([]string{"fallback_until"}).AddRow(until))

	got, err := repo.BillingReservationDatabaseFallbackUntil(ctx)
	require.NoError(t, err)
	require.WithinDuration(t, until, got, time.Millisecond)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestTryReserveUserBalanceDatabase_RejectsInvalidInputs 覆盖入参防御：
// 非有限值/负数在触库前就被拒绝，避免把 NaN 之类写进账本。
func TestTryReserveUserBalanceDatabase_RejectsInvalidInputs(t *testing.T) {
	repo, _ := newBillingReservationRepoMock(t)
	ctx := context.Background()

	_, _, err := repo.TryReserveUserBalanceDatabase(ctx, "7", "req-1", math.NaN(), 1, time.Minute)
	require.Error(t, err)

	_, _, err = repo.TryReserveUserBalanceDatabase(ctx, "7", "req-1", 0.1, math.Inf(1), time.Minute)
	require.Error(t, err)

	_, _, err = repo.TryReserveUserBalanceDatabase(ctx, "", "req-1", 0.1, 1, time.Minute)
	require.Error(t, err)
}

// TestBillingReservationDatabaseFallback_UnavailableWithoutSQLDB 覆盖降级装配：
// repository 没有 *sql.DB 时，DB 兜底能力必须显式报错（而不是静默 no-op），
// 让服务层据此 fail-closed。
func TestBillingReservationDatabaseFallback_UnavailableWithoutSQLDB(t *testing.T) {
	repo := &userRepository{}
	ctx := context.Background()

	_, _, err := repo.TryReserveUserBalanceDatabase(ctx, "7", "req-1", 0.1, 1, time.Minute)
	require.Error(t, err)
	require.Error(t, repo.ActivateBillingReservationDatabaseFallback(ctx, time.Minute))
	_, err = repo.BillingReservationDatabaseFallbackUntil(ctx)
	require.Error(t, err)
	require.Error(t, repo.ReleaseUserBalanceReservation(ctx, "7", "req-1", 0.1, time.Minute), "无 DB 时归还也必须报错")
	require.Error(t, repo.RenewUserBalanceReservation(ctx, "7", "req-1", time.Minute), "无 DB 时续期也必须报错")
}

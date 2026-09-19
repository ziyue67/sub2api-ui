//go:build unit

package service

import (
	"context"
	"errors"
	"math"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

// databaseFallbackRepoStub 用内存账本模拟 PostgreSQL 兜底预留（审计 R2）：
//
//   - TryReserveUserBalanceDatabase 复刻"按 scope 聚合、带上限判定、幂等凭据"语义；
//   - 每次成功预留/续期都把共享兜底窗口延长到"本笔凭据过期时刻"，与真实实现一致；
//   - Release/Renew 以凭据为准，凭据不存在返回 ErrBillingReservationExpired。
//
// 内嵌 UserRepository 只为满足接口：预检路径不会调用其它方法（余额未贴近底线时不回源）。
type databaseFallbackRepoStub struct {
	UserRepository

	mu           sync.Mutex
	receipts     map[string]float64
	reservedNano int64
	windowUntil  time.Time

	reserveCalls  atomic.Int64
	releaseCalls  atomic.Int64
	renewCalls    atomic.Int64
	activateCalls atomic.Int64
	probeCalls    atomic.Int64

	failReserve  atomic.Bool
	failActivate atomic.Bool
	failProbe    atomic.Bool
}

func newDatabaseFallbackRepoStub() *databaseFallbackRepoStub {
	return &databaseFallbackRepoStub{receipts: make(map[string]float64)}
}

func (s *databaseFallbackRepoStub) TryReserveUserBalanceDatabase(
	_ context.Context,
	scope string,
	requestID string,
	amount float64,
	maxTotal float64,
	ttl time.Duration,
) (float64, bool, error) {
	s.reserveCalls.Add(1)
	if s.failReserve.Load() {
		return 0, false, errors.New("database unavailable")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current := float64(s.reservedNano) / 1e9
	if current+amount > maxTotal+1e-12 {
		return current, false, nil
	}
	if _, exists := s.receipts[requestID]; !exists {
		s.receipts[requestID] = amount
		s.reservedNano += int64(math.Round(amount * 1e9))
	}
	s.extendWindowLocked(time.Now().Add(ttl))
	return float64(s.reservedNano) / 1e9, true, nil
}

func (s *databaseFallbackRepoStub) ReleaseUserBalanceReservation(_ context.Context, _ string, requestID string, _ float64, _ time.Duration) error {
	s.releaseCalls.Add(1)
	s.mu.Lock()
	defer s.mu.Unlock()
	stored, exists := s.receipts[requestID]
	if !exists {
		return ErrBillingReservationExpired
	}
	delete(s.receipts, requestID)
	s.reservedNano -= int64(math.Round(stored * 1e9))
	if s.reservedNano < 0 {
		s.reservedNano = 0
	}
	return nil
}

func (s *databaseFallbackRepoStub) RenewUserBalanceReservation(_ context.Context, _ string, requestID string, ttl time.Duration) error {
	s.renewCalls.Add(1)
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.receipts[requestID]; !exists {
		return ErrBillingReservationExpired
	}
	s.extendWindowLocked(time.Now().Add(ttl))
	return nil
}

func (s *databaseFallbackRepoStub) ActivateBillingReservationDatabaseFallback(_ context.Context, ttl time.Duration) error {
	s.activateCalls.Add(1)
	if s.failActivate.Load() {
		return errors.New("activate fallback failed")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.extendWindowLocked(time.Now().Add(ttl))
	return nil
}

func (s *databaseFallbackRepoStub) BillingReservationDatabaseFallbackUntil(_ context.Context) (time.Time, error) {
	s.probeCalls.Add(1)
	if s.failProbe.Load() {
		return time.Time{}, errors.New("probe failed")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.windowUntil, nil
}

func (s *databaseFallbackRepoStub) extendWindowLocked(until time.Time) {
	if until.After(s.windowUntil) {
		s.windowUntil = until
	}
}

func (s *databaseFallbackRepoStub) reservedAmount() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return float64(s.reservedNano) / 1e9
}

func (s *databaseFallbackRepoStub) window() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.windowUntil
}

func newDatabaseFallbackTestService(cache BillingCache, repo UserRepository) *BillingCacheService {
	cfg := &config.Config{}
	cfg.Billing.MinimumBalanceReserve = 0.10
	svc := NewBillingCacheService(cache, repo, nil, nil, nil, nil, cfg, nil)
	return svc
}

// TestReserveRequestSpend_FallsBackToDatabaseWhenRedisFails 验证审计 R2 的主线：
// Redis 预留不可用时不再 fail-open，而是抬起共享兜底窗口并改走 DB 账本；
// 槽位绑定到 DB store，因此收尾归还也必须走 DB（绝不误碰 Redis 的账）。
func TestReserveRequestSpend_FallsBackToDatabaseWhenRedisFails(t *testing.T) {
	cache := &reservationCacheStub{}
	cache.balance = 0.45
	cache.failReserve.Store(true)
	repo := newDatabaseFallbackRepoStub()
	svc := newDatabaseFallbackTestService(cache, repo)
	t.Cleanup(svc.Stop)

	var slot BillingReservationSlot
	require.NoError(t, svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, nil, nil, "", reservationEligibilityOpts(&slot, 0.10)...))

	require.Equal(t, int64(1), repo.reserveCalls.Load(), "Redis 故障后必须尝试 DB 兜底预留")
	require.Equal(t, int64(1), repo.activateCalls.Load(), "必须先抬起共享兜底窗口，其它实例才会跟随切账")
	require.InDelta(t, 0.10, repo.reservedAmount(), 1e-9, "DB 账本应记录本笔在途预留")
	require.InDelta(t, 0.0, cache.reservedAmount(), 1e-9, "Redis 账本不应被写入")
	require.True(t, repo.window().After(time.Now()), "兜底窗口必须仍在有效期")

	// 槽位绑定的是 DB store：归还走 DB，且只归还本笔凭据。
	slot.ReleaseOnExit(context.Background())
	require.Equal(t, int64(1), repo.releaseCalls.Load(), "归还必须走 DB 账本")
	require.Equal(t, int64(0), cache.releaseCalls.Load(), "不得触碰 Redis 账本")
	require.InDelta(t, 0.0, repo.reservedAmount(), 1e-9)
}

// TestReserveRequestSpend_FailsClosedWhenBothLedgersFail 验证审计 R2 的失败语义：
// Redis 与 DB 兜底同时不可用时必须 fail-closed（503），而不是退回"零预留放行"。
func TestReserveRequestSpend_FailsClosedWhenBothLedgersFail(t *testing.T) {
	cache := &reservationCacheStub{}
	cache.balance = 0.45
	cache.failReserve.Store(true)
	repo := newDatabaseFallbackRepoStub()
	repo.failReserve.Store(true)
	svc := newDatabaseFallbackTestService(cache, repo)
	t.Cleanup(svc.Stop)

	var slot BillingReservationSlot
	err := svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, nil, nil, "", reservationEligibilityOpts(&slot, 0.10)...)
	require.ErrorIs(t, err, ErrBillingServiceUnavailable, "双账本不可用时必须 fail-closed")

	// 未建立任何预留 → 槽位未绑定 → 收尾归还是 no-op。
	slot.ReleaseOnExit(context.Background())
	require.Equal(t, int64(0), repo.releaseCalls.Load())
	require.InDelta(t, 0.0, cache.reservedAmount(), 1e-9)
}

// TestReserveRequestSpend_SharedFallbackWindowBypassesRedis 验证跨实例一致性：
// 兜底窗口处于激活状态时，即使 Redis 可用也统一走 DB 账本，
// 避免同一份额度被 Redis 与 DB 各算一遍。
func TestReserveRequestSpend_SharedFallbackWindowBypassesRedis(t *testing.T) {
	cache := &reservationCacheStub{}
	cache.balance = 0.45
	repo := newDatabaseFallbackRepoStub()
	require.NoError(t, repo.ActivateBillingReservationDatabaseFallback(context.Background(), time.Minute))
	repo.activateCalls.Store(0)
	svc := newDatabaseFallbackTestService(cache, repo)
	t.Cleanup(svc.Stop)

	var slot BillingReservationSlot
	require.NoError(t, svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, nil, nil, "", reservationEligibilityOpts(&slot, 0.10)...))

	require.Equal(t, int64(0), cache.reserveCalls.Load(), "窗口激活时不得再使用 Redis 账本")
	require.Equal(t, int64(1), repo.reserveCalls.Load(), "窗口激活时必须使用 DB 账本")
	require.InDelta(t, 0.10, repo.reservedAmount(), 1e-9)
}

// TestReserveRequestSpend_DatabaseFallbackEnforcesBudget 验证 DB 账本与 Redis 账本
// 守同一条不变量：余额 0.45、封底 0.10、每笔最坏费用 0.10 时最多 3 笔，
// 第 4 笔必须在预检被拒（ErrInsufficientBalance），且 DB 余额台账不受影响。
func TestReserveRequestSpend_DatabaseFallbackEnforcesBudget(t *testing.T) {
	cache := &reservationCacheStub{}
	cache.balance = 0.45
	cache.failReserve.Store(true)
	repo := newDatabaseFallbackRepoStub()
	svc := newDatabaseFallbackTestService(cache, repo)
	t.Cleanup(svc.Stop)

	for i := 0; i < 3; i++ {
		var slot BillingReservationSlot
		require.NoError(t, svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, nil, nil, "", reservationEligibilityOpts(&slot, 0.10)...),
			"第 %d 笔应放行（0.45 - 3*0.10 == 0.15 > 0.10）", i+1)
	}
	var overflow BillingReservationSlot
	err := svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, nil, nil, "", reservationEligibilityOpts(&overflow, 0.10)...)
	require.ErrorIs(t, err, ErrInsufficientBalance, "DB 账本也必须挡住第 4 笔")
	require.InDelta(t, 0.30, repo.reservedAmount(), 1e-9, "被拒的预留不得留在 DB 账本里")
}

// TestReserveRequestSpend_NilDatabaseFallbackKeepsFailOpen 记录兼容语义：
// 装配里没有 DB 兜底能力（轻量桩/部分降级部署）时，Redis 故障仍按旧行为 fail-open，
// 该计数器是运维识别"护栏真的消失"的唯一信号。
func TestReserveRequestSpend_NilDatabaseFallbackKeepsFailOpen(t *testing.T) {
	cache := &reservationCacheStub{}
	cache.balance = 0.45
	cache.failReserve.Store(true)
	svc := newDatabaseFallbackTestService(cache, nil)
	t.Cleanup(svc.Stop)

	var slot BillingReservationSlot
	require.NoError(t, svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, nil, nil, "", reservationEligibilityOpts(&slot, 0.10)...))
	slot.ReleaseOnExit(context.Background())
	require.Equal(t, int64(0), cache.releaseCalls.Load(), "fail-open 路径不绑定槽位")
}

// TestReserveRequestSpend_NonCapableUserRepoKeepsFailOpen 覆盖"装配了 userRepo、
// 但它未实现 DB 兜底能力"的部署形态：行为必须与未装配完全一致（fail-open），
// 不得因为探测/激活而阻断既有请求。
func TestReserveRequestSpend_NonCapableUserRepoKeepsFailOpen(t *testing.T) {
	cache := &reservationCacheStub{}
	cache.balance = 0.45
	cache.failReserve.Store(true)
	svc := newDatabaseFallbackTestService(cache, &balanceLoadUserRepoStub{balance: 0.45})
	t.Cleanup(svc.Stop)

	var slot BillingReservationSlot
	require.NoError(t, svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, nil, nil, "", reservationEligibilityOpts(&slot, 0.10)...))
	slot.ReleaseOnExit(context.Background())
	require.Equal(t, int64(0), cache.releaseCalls.Load())
}

// TestReserveRequestSpend_FallbackProbeFailureKeepsRedisHotPath 验证探测失败语义：
// 探测报错时不缓存结果、按"窗口未激活"处理 —— Redis 仍可用则维持热路径，
// 失败的探测次数必须被计数（避免降级静默）。
func TestReserveRequestSpend_FallbackProbeFailureKeepsRedisHotPath(t *testing.T) {
	cache := &reservationCacheStub{}
	cache.balance = 0.45
	repo := newDatabaseFallbackRepoStub()
	repo.failProbe.Store(true)
	svc := newDatabaseFallbackTestService(cache, repo)
	t.Cleanup(svc.Stop)

	var slot BillingReservationSlot
	require.NoError(t, svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, nil, nil, "", reservationEligibilityOpts(&slot, 0.10)...))
	require.Equal(t, int64(1), cache.reserveCalls.Load(), "探测失败时应保留 Redis 热路径")
	require.Equal(t, int64(0), repo.reserveCalls.Load())
}

// TestReserveRequestSpend_DatabaseLedgerSlotSupportsRenew 验证回退到 DB 账本时，
// 槽位绑定的是 DB store 且提供续期能力（心跳可延长 DB 凭据与共享窗口），
// 而不是把 Redis store 绑上去（那会让晚到的归还去动另一本账）。
func TestReserveRequestSpend_DatabaseLedgerSlotSupportsRenew(t *testing.T) {
	cache := &reservationCacheStub{}
	cache.balance = 0.45
	cache.failReserve.Store(true)
	repo := newDatabaseFallbackRepoStub()
	svc := newDatabaseFallbackTestService(cache, repo)
	t.Cleanup(svc.Stop)

	var slot BillingReservationSlot
	require.NoError(t, svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, nil, nil, "", reservationEligibilityOpts(&slot, 0.10)...))

	slot.mu.Lock()
	bound := slot.store
	requestID := slot.requestID
	slot.mu.Unlock()

	require.Same(t, repo, bound, "槽位必须绑定 DB store，归还/续期才不会误碰 Redis")
	renewer, ok := bound.(billingReservationRenewer)
	require.True(t, ok, "DB store 必须提供续期能力（长请求心跳）")
	require.NoError(t, renewer.RenewUserBalanceReservation(context.Background(), balanceReservationScope(1), requestID, billingReservationTTL))
	require.Equal(t, int64(1), repo.renewCalls.Load())
	require.True(t, repo.window().After(time.Now()), "续期必须同时延长共享兜底窗口")
}

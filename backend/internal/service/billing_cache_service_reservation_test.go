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

// reservationCacheStub 在余额缓存桩之上实现"在途预留"能力：用互斥锁保护的累计值
// 精确模拟 Redis 端 INCRBYFLOAT / 递减归零删除的语义，用于验证并发准入的原子性。
//
// 累计刻度为 1e-9 USD（nano）：Redis INCRBYFLOAT 走十进制累加，0.1+0.1+0.1 返回 "0.3"；
// 若这里用二进制 float64 累加会得到 0.30000000000000004，令边界判定偏离真实语义。
type reservationCacheStub struct {
	balanceEligibilityCacheStub

	mu           sync.Mutex
	reservedNano int64
	reserveCalls atomic.Int64
	releaseCalls atomic.Int64
	failReserve  atomic.Bool
	failRelease  atomic.Bool
}

func (s *reservationCacheStub) ReserveUserBalance(_ context.Context, _ int64, amount float64, _ time.Duration) (float64, error) {
	s.reserveCalls.Add(1)
	if s.failReserve.Load() {
		return 0, errors.New("redis down")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reservedNano += int64(math.Round(amount * 1e9))
	return float64(s.reservedNano) / 1e9, nil
}

func (s *reservationCacheStub) ReleaseUserBalanceReservation(_ context.Context, _ int64, amount float64, _ time.Duration) error {
	s.releaseCalls.Add(1)
	if s.failRelease.Load() {
		return errors.New("redis down")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reservedNano -= int64(math.Round(amount * 1e9))
	if s.reservedNano < 0 {
		s.reservedNano = 0
	}
	return nil
}

func (s *reservationCacheStub) reservedAmount() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return float64(s.reservedNano) / 1e9
}

func newReservationTestService(cache BillingCache) *BillingCacheService {
	cfg := &config.Config{}
	cfg.Billing.MinimumBalanceReserve = 0.10
	svc := NewBillingCacheService(cache, nil, nil, nil, nil, nil, cfg, nil)
	return svc
}

func reservationEligibilityOpts(slot *BillingReservationSlot, worstSpend float64) []BillingEligibilityOption {
	return []BillingEligibilityOption{
		WithMaxRequestSpend(worstSpend),
		WithBalanceReservation(slot),
	}
}

// TestReserveRequestSpend_ConcurrentAdmissionIsAtomic 复现并锁死并发穿透：
// 余额 0.45、封底 0.10、每笔最坏费用 0.10 时，10 个并发请求只能放行 3 笔
// （0.45 - 3*0.10 == 0.15 > 0.10，第 4 笔起 0.45 - 0.4 == 0.05 < 0.10），
// 其余必须在预检即 403。
//
// 修复前：10 笔都读到同一份余额快照 0.45（10*0.10 远超 0.45 也照样全放行），
// 结算时只有前几笔扣得动，其余全部 actual_cost=0 —— 坏账。
func TestReserveRequestSpend_ConcurrentAdmissionIsAtomic(t *testing.T) {
	cache := &reservationCacheStub{}
	cache.balance = 0.45
	svc := newReservationTestService(cache)
	t.Cleanup(svc.Stop)

	const attempts = 10
	start := make(chan struct{})
	results := make(chan error, attempts)
	heldSlots := make(chan *BillingReservationSlot, attempts)
	var wg sync.WaitGroup
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			slot := new(BillingReservationSlot)
			<-start
			err := svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, nil, nil, "", reservationEligibilityOpts(slot, 0.10)...)
			if err == nil {
				heldSlots <- slot
			}
			results <- err
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	close(heldSlots)

	var admitted, rejected int
	for err := range results {
		switch {
		case err == nil:
			admitted++
		case errors.Is(err, ErrInsufficientBalance):
			rejected++
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}
	require.Equal(t, 3, admitted, "只有 3 笔能守住封底：0.45 - 3*0.10 == 0.15")
	require.Equal(t, attempts-3, rejected, "其余并发请求必须在预检被 403")
	require.InDelta(t, 0.30, cache.reservedAmount(), 1e-9, "在途预留总额应为 3*0.10")
	require.Equal(t, int64(7), cache.releaseCalls.Load(), "被拒绝的 7 笔必须已回滚预留")

	var held []*BillingReservationSlot
	for slot := range heldSlots {
		held = append(held, slot)
	}
	require.Len(t, held, 3)

	// 归还一笔后腾出额度（0.45 - 0.20 == 0.25 > 0.10）：下一笔立即可以放行。
	held[0].Release(context.Background())
	require.InDelta(t, 0.20, cache.reservedAmount(), 1e-9)
	require.Equal(t, int64(8), cache.releaseCalls.Load())

	var slot BillingReservationSlot
	require.NoError(t, svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, nil, nil, "", reservationEligibilityOpts(&slot, 0.10)...))
	require.InDelta(t, 0.30, cache.reservedAmount(), 1e-9)

	// 再打一笔：护栏不成立（0.45 - 0.40 == 0.05 < 0.10），必须回滚预留并 403。
	var overflow BillingReservationSlot
	err := svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, nil, nil, "", reservationEligibilityOpts(&overflow, 0.10)...)
	require.ErrorIs(t, err, ErrInsufficientBalance)
	require.InDelta(t, 0.30, cache.reservedAmount(), 1e-9, "被拒绝的预留必须已回滚")
	require.Equal(t, int64(9), cache.releaseCalls.Load(), "第 9 次 release 是被拒绝预留的回滚")
}

// TestReserveRequestSpend_FailsOpenWhenReservationUnavailable 确认预留能力故障
// 时 fail-open：不影响既有余额闸门的判定，也不会误绑槽位。
func TestReserveRequestSpend_FailsOpenWhenReservationUnavailable(t *testing.T) {
	cache := &reservationCacheStub{}
	cache.balance = 0.45
	cache.failReserve.Store(true)
	svc := newReservationTestService(cache)
	t.Cleanup(svc.Stop)

	var slot BillingReservationSlot
	require.NoError(t, svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, nil, nil, "", reservationEligibilityOpts(&slot, 0.10)...))

	// 预留失败 → 槽位未绑定 → 收尾归还必须是 no-op（不能去动别人的预留）。
	slot.ReleaseOnExit(context.Background())
	require.Equal(t, int64(0), cache.releaseCalls.Load())
	require.InDelta(t, 0.0, cache.reservedAmount(), 1e-9)
}

// TestReserveRequestSpend_NoopForPlainCacheStub 确认没有实现预留能力的缓存
// （历史测试桩 / 降级装配）依旧走原来的行为：不预留、不阻断。
func TestReserveRequestSpend_NoopForPlainCacheStub(t *testing.T) {
	cache := &balanceEligibilityCacheStub{balance: 0.40}
	// 普通缓存桩未实现预留能力：预检不得被预留逻辑阻断。
	svc := newReservationTestService(cache)
	t.Cleanup(svc.Stop)

	var slot BillingReservationSlot
	require.NoError(t, svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, nil, nil, "", reservationEligibilityOpts(&slot, 0.10)...))
	slot.ReleaseOnExit(context.Background())
}

// TestReserveRequestSpend_NoopWithoutWorstSpend 确认没有最坏费用上界（未挂
// WithMaxRequestSpend）时不产生预留：预留必须与放行上界同源，否则会凭空占额度。
func TestReserveRequestSpend_NoopWithoutWorstSpend(t *testing.T) {
	cache := &reservationCacheStub{}
	cache.balance = 0.45
	svc := newReservationTestService(cache)
	t.Cleanup(svc.Stop)

	var slot BillingReservationSlot
	require.NoError(t, svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, nil, nil, "", WithBalanceReservation(&slot)))
	require.Equal(t, int64(0), cache.reserveCalls.Load())
	require.InDelta(t, 0.0, cache.reservedAmount(), 1e-9)
}

// TestBillingReservationSlot_HandOffDefersReleaseToSettlement 验证归还时机：
// 已提交结算任务的请求，收尾兜底（ReleaseOnExit）不得提前归还；只有结算任务
// 的 Release 才真正归还，且重复调用幂等。
func TestBillingReservationSlot_HandOffDefersReleaseToSettlement(t *testing.T) {
	cache := &reservationCacheStub{}
	cache.balance = 0.45
	svc := newReservationTestService(cache)
	t.Cleanup(svc.Stop)

	var slot BillingReservationSlot
	require.NoError(t, svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, nil, nil, "", reservationEligibilityOpts(&slot, 0.10)...))
	require.InDelta(t, 0.10, cache.reservedAmount(), 1e-9)

	// handler 收尾：已移交给结算任务（先 HandOff），此时不允许归还。
	slot.HandOff()
	slot.ReleaseOnExit(context.Background())
	require.Equal(t, int64(0), cache.releaseCalls.Load(), "移交结算后收尾兜底不得提前归还")
	require.InDelta(t, 0.10, cache.reservedAmount(), 1e-9)

	// 结算任务在扣费完成后归还；重复调用幂等。
	slot.Release(context.Background())
	slot.Release(context.Background())
	require.Equal(t, int64(1), cache.releaseCalls.Load(), "Release 必须幂等")
	require.InDelta(t, 0.0, cache.reservedAmount(), 1e-9)
}

// TestBillingReservationSlot_ReleaseOnExitWithoutSettlement 验证没有结算的
// 错误路径：收尾兜底立即归还，不把额度钉到 TTL 到期。
func TestBillingReservationSlot_ReleaseOnExitWithoutSettlement(t *testing.T) {
	cache := &reservationCacheStub{}
	cache.balance = 0.45
	svc := newReservationTestService(cache)
	t.Cleanup(svc.Stop)

	var slot BillingReservationSlot
	require.NoError(t, svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, nil, nil, "", reservationEligibilityOpts(&slot, 0.10)...))

	slot.ReleaseOnExit(context.Background())
	require.Equal(t, int64(1), cache.releaseCalls.Load())
	require.InDelta(t, 0.0, cache.reservedAmount(), 1e-9)

	// 再做一次兜底（例如 defer 与显式归还同时存在）不得重复扣减。
	slot.ReleaseOnExit(context.Background())
	slot.Release(context.Background())
	require.Equal(t, int64(1), cache.releaseCalls.Load())
	require.InDelta(t, 0.0, cache.reservedAmount(), 1e-9)
}

//go:build integration

package repository

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// BillingReservationSuite 验证"在途预留"（并发准入护栏）在真实 Redis 上的原子语义：
// 累加与十进制往返、凭据驱动的精确归还、并发累加、TTL 自愈、心跳续期，
// 以及"凭据过期后的归还不得吃掉他人预留"这条关键正确性约束。
type BillingReservationSuite struct {
	IntegrationRedisSuite
}

// reservationCache 构造真实 Redis 上的 billingCache 并暴露内部预留方法。
func (s *BillingReservationSuite) reservationCache() (*billingCache, *redis.Client) {
	s.T().Helper()
	rdb := testRedis(s.T())
	cache, ok := NewBillingCache(rdb).(*billingCache)
	require.True(s.T(), ok, "NewBillingCache 应返回 *billingCache")
	return cache, rdb
}

func (s *BillingReservationSuite) TestReserveAccumulatesAndReportsTotal() {
	cache, rdb := s.reservationCache()
	ctx := context.Background()
	scope := "9001"
	key := billingReservedKeyPrefix + scope

	total, err := cache.ReserveUserBalance(ctx, scope, "req-a", 0.10, 10*time.Minute)
	require.NoError(s.T(), err, "ReserveUserBalance")
	require.InDelta(s.T(), 0.10, total, 1e-9, "首笔预留应从 0 累加到 0.10")

	// 第二笔累加：0.10 + 0.20 == 0.30，且小数必须精确（十进制往返，不能被整数截断）。
	total, err = cache.ReserveUserBalance(ctx, scope, "req-b", 0.20, 10*time.Minute)
	require.NoError(s.T(), err, "ReserveUserBalance 第二笔")
	require.InDelta(s.T(), 0.30, total, 1e-9, "两笔预留应累加为 0.30")

	exists, err := rdb.Exists(ctx, key).Result()
	require.NoError(s.T(), err, "Exists")
	require.Equal(s.T(), int64(1), exists, "预留键应存在")

	// 每笔请求都留有自己的凭据键。
	itemExists, err := rdb.Exists(ctx, billingReservedItemKey(scope, "req-a"), billingReservedItemKey(scope, "req-b")).Result()
	require.NoError(s.T(), err, "Exists item")
	require.Equal(s.T(), int64(2), itemExists, "每笔请求应各自留有一条凭据")

	ttl, err := rdb.TTL(ctx, key).Result()
	require.NoError(s.T(), err, "TTL")
	s.AssertTTLWithin(ttl, 9*time.Minute, 10*time.Minute)

	peeked, err := cache.ReservedUserBalanceTotal(ctx, scope)
	require.NoError(s.T(), err, "ReservedUserBalanceTotal")
	require.InDelta(s.T(), 0.30, peeked, 1e-9, "只读总额应与累加值一致")
}

func (s *BillingReservationSuite) TestReserveIsIdempotentForSameRequestID() {
	cache, _ := s.reservationCache()
	ctx := context.Background()
	scope := "9002"

	_, err := cache.ReserveUserBalance(ctx, scope, "req-a", 0.10, 10*time.Minute)
	require.NoError(s.T(), err, "第一笔")

	// 同一 requestID 重复预留（重试 / 重复绑定）不得重复累加：否则同一笔请求会被记两次额度。
	total, err := cache.ReserveUserBalance(ctx, scope, "req-a", 0.10, 10*time.Minute)
	require.NoError(s.T(), err, "重复预留同一 requestID")
	require.InDelta(s.T(), 0.10, total, 1e-9, "同 requestID 重复预留必须幂等")
}

func (s *BillingReservationSuite) TestReserveFallsBackToCacheTTLWhenTTLMissing() {
	cache, rdb := s.reservationCache()
	ctx := context.Background()
	scope := "9003"
	key := billingReservedKeyPrefix + scope

	_, err := cache.ReserveUserBalance(ctx, scope, "req-a", 0.05, 0)
	require.NoError(s.T(), err, "ReserveUserBalance 应以缺省 TTL 兜底")

	ttl, err := rdb.TTL(ctx, key).Result()
	require.NoError(s.T(), err, "TTL")
	s.AssertTTLWithin(ttl, billingCacheTTL-30*time.Second, billingCacheTTL)
}

func (s *BillingReservationSuite) TestReserveRejectsInvalidArguments() {
	cache, _ := s.reservationCache()
	ctx := context.Background()

	_, err := cache.ReserveUserBalance(ctx, "9004", "req-a", -0.10, 10*time.Minute)
	require.Error(s.T(), err, "负数预留必须被拒绝（否则可凭空放大可花余额）")

	_, err = cache.ReserveUserBalance(ctx, "", "req-a", 0.10, 10*time.Minute)
	require.Error(s.T(), err, "空 scope 必须被拒绝")

	_, err = cache.ReserveUserBalance(ctx, "9004", "", 0.10, 10*time.Minute)
	require.Error(s.T(), err, "空 requestID 必须被拒绝（凭据无法定位就无法安全归还）")
}

func (s *BillingReservationSuite) TestReleaseReducesAndDeletesAtZero() {
	cache, rdb := s.reservationCache()
	ctx := context.Background()
	scope := "9005"
	key := billingReservedKeyPrefix + scope

	_, err := cache.ReserveUserBalance(ctx, scope, "req-a", 0.30, 10*time.Minute)
	require.NoError(s.T(), err, "ReserveUserBalance")

	require.NoError(s.T(), cache.ReleaseUserBalanceReservation(ctx, scope, "req-a", 0.10, 10*time.Minute), "归还本笔预留")

	rest, err := cache.ReservedUserBalanceTotal(ctx, scope)
	require.NoError(s.T(), err, "读回总额")
	require.InDelta(s.T(), 0.0, rest, 1e-9, "归还后凭据消失、聚合键应被删除")

	exists, err := rdb.Exists(ctx, key).Result()
	require.NoError(s.T(), err, "Exists")
	require.Equal(s.T(), int64(0), exists, "归零后预留键必须被删除")

	// 凭据已删除后再归还：返回 ErrBillingReservationExpired（no-op），不得把总额减成负数。
	err = cache.ReleaseUserBalanceReservation(ctx, scope, "req-a", 0.10, 10*time.Minute)
	require.ErrorIs(s.T(), err, service.ErrBillingReservationExpired, "凭据缺失时的归还必须安全拒绝")
	exists, err = rdb.Exists(ctx, key).Result()
	require.NoError(s.T(), err, "Exists after no-op release")
	require.Equal(s.T(), int64(0), exists, "no-op 归还不得重新创建键")
}

func (s *BillingReservationSuite) TestReleaseOnMissingReceiptIsNoop() {
	cache, rdb := s.reservationCache()
	ctx := context.Background()
	scope := "9006"
	key := billingReservedKeyPrefix + scope

	err := cache.ReleaseUserBalanceReservation(ctx, scope, "never-reserved", 0.10, 10*time.Minute)
	require.ErrorIs(s.T(), err, service.ErrBillingReservationExpired, "从未预留过也要安全 no-op")

	exists, err := rdb.Exists(ctx, key).Result()
	require.NoError(s.T(), err, "Exists")
	require.Equal(s.T(), int64(0), exists, "no-op 归还不得创建键")
}

// TestExpiredReceiptReleaseDoesNotEatOtherReservations 是并发正确性缺陷的回归测试：
//
// 长请求 A 的凭据先被 TTL 回收，之后 B 建立了自己的预留，A 才结算归还。
// 旧实现只看聚合键是否存在就无条件递减，会把 B 的预留一起扣掉（护栏被错误解除，
// 并发超额窗口重开）。新实现要求凭据存在，因此 A 的归还是 no-op。
func (s *BillingReservationSuite) TestExpiredReceiptReleaseDoesNotEatOtherReservations() {
	cache, rdb := s.reservationCache()
	ctx := context.Background()
	scope := "9007"
	key := billingReservedKeyPrefix + scope

	// A 预留 0.10。
	_, err := cache.ReserveUserBalance(ctx, scope, "req-a", 0.10, 10*time.Minute)
	require.NoError(s.T(), err, "A 预留")

	// 模拟 A 的凭据被 TTL 自愈回收（聚合键仍在：保守方向，偏高不会放行超额）。
	require.NoError(s.T(), rdb.Del(ctx, billingReservedItemKey(scope, "req-a")).Err(), "回收 A 的凭据")

	// B 之后预留 0.30。
	_, err = cache.ReserveUserBalance(ctx, scope, "req-b", 0.30, 10*time.Minute)
	require.NoError(s.T(), err, "B 预留")
	total, err := cache.ReservedUserBalanceTotal(ctx, scope)
	require.NoError(s.T(), err, "读回总额")
	require.InDelta(s.T(), 0.40, total, 1e-9, "聚合总额应为 A+B")

	// A 晚到的归还：必须 no-op，绝不能吃掉 B 的 0.30。
	err = cache.ReleaseUserBalanceReservation(ctx, scope, "req-a", 0.10, 10*time.Minute)
	require.ErrorIs(s.T(), err, service.ErrBillingReservationExpired, "过期凭据的归还必须被拒绝")
	total, err = cache.ReservedUserBalanceTotal(ctx, scope)
	require.NoError(s.T(), err, "读回总额")
	require.InDelta(s.T(), 0.40, total, 1e-9, "过期凭据的归还不得影响他人预留")

	// B 正常归还：只剩 A 那笔已被 TTL 回收的残留份额。
	require.NoError(s.T(), cache.ReleaseUserBalanceReservation(ctx, scope, "req-b", 0.30, 10*time.Minute), "B 归还")
	total, err = cache.ReservedUserBalanceTotal(ctx, scope)
	require.NoError(s.T(), err, "读回总额")
	require.InDelta(s.T(), 0.10, total, 1e-9, "归还只能减掉自己的份额")

	exists, err := rdb.Exists(ctx, key).Result()
	require.NoError(s.T(), err, "Exists")
	require.Equal(s.T(), int64(1), exists, "残留份额仍挂在聚合键上，等聚合键 TTL 兜底")
}

func (s *BillingReservationSuite) TestRenewExtendsReceiptTTL() {
	cache, rdb := s.reservationCache()
	ctx := context.Background()
	scope := "9008"

	_, err := cache.ReserveUserBalance(ctx, scope, "req-a", 0.10, 5*time.Minute)
	require.NoError(s.T(), err, "Reserve")

	require.NoError(s.T(), cache.RenewUserBalanceReservation(ctx, scope, "req-a", 30*time.Minute), "续期")

	itemTTL, err := rdb.TTL(ctx, billingReservedItemKey(scope, "req-a")).Result()
	require.NoError(s.T(), err, "item TTL")
	s.AssertTTLWithin(itemTTL, 25*time.Minute, 30*time.Minute)

	totalTTL, err := rdb.TTL(ctx, billingReservedKey(scope)).Result()
	require.NoError(s.T(), err, "aggregate TTL")
	s.AssertTTLWithin(totalTTL, 25*time.Minute, 30*time.Minute)

	// 凭据不存在（已归还 / 已过期）时续期必须返回"已过期"，让调用方停止心跳。
	err = cache.RenewUserBalanceReservation(ctx, scope, "req-missing", 30*time.Minute)
	require.ErrorIs(s.T(), err, service.ErrBillingReservationExpired, "凭据缺失时续期必须报告已过期")
}

func (s *BillingReservationSuite) TestConcurrentReserveAccumulatesExactly() {
	cache, _ := s.reservationCache()
	ctx := context.Background()
	scope := "9009"

	const workers = 20
	const perWorker = 0.10

	var wg sync.WaitGroup
	totals := make([]float64, workers)
	errs := make([]error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			totals[idx], errs[idx] = cache.ReserveUserBalance(ctx, scope, fmt.Sprintf("req-%d", idx), perWorker, 10*time.Minute)
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		require.NoError(s.T(), err, "goroutine %d 预留失败", i)
	}

	// 20 × 0.10 == 2.00：INCRBYFLOAT 的十进制累加必须精确到浮点误差内。
	final, err := cache.ReservedUserBalanceTotal(ctx, scope)
	require.NoError(s.T(), err, "读回并发累加总额")
	require.InDelta(s.T(), 2.00, final, 1e-9, "20 笔并发预留应精确累加为 2.00")

	// 快照式读取曾出现"最后一个写入者"的形态：确认累计值不是单笔值。
	require.Greater(s.T(), final, perWorker, "并发累加不得退化为单笔覆盖")

	// 并发场景下每笔都必须留下自己的凭据，否则无法安全归还。
	itemCount, err := rdbItemCount(ctx, cache, scope)
	require.NoError(s.T(), err, "统计凭据数")
	require.Equal(s.T(), workers, itemCount, "每笔并发预留都应有独立凭据")
}

// rdbItemCount 统计某个 scope 下的凭据键数量（仅测试使用）。
func rdbItemCount(ctx context.Context, cache *billingCache, scope string) (int, error) {
	keys, err := cache.rdb.Keys(ctx, billingReservedItemKeyPrefix+scope+":*").Result()
	if err != nil {
		return 0, err
	}
	return len(keys), nil
}

func TestBillingReservationSuite(t *testing.T) {
	suite.Run(t, new(BillingReservationSuite))
}

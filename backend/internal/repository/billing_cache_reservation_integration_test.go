//go:build integration

package repository

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// BillingReservationSuite 验证"在途预留"（并发准入护栏）在真实 Redis 上的原子语义：
// INCRBYFLOAT 累加、递减归零删除、缺失键 no-op 不为负、并发累加精确、TTL 自愈。
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
	userID := int64(9001)
	key := fmt.Sprintf("%s%d", billingReservedKeyPrefix, userID)

	total, err := cache.ReserveUserBalance(ctx, userID, 0.10, 10*time.Minute)
	require.NoError(s.T(), err, "ReserveUserBalance")
	require.InDelta(s.T(), 0.10, total, 1e-9, "首笔预留应从 0 累加到 0.10")

	// 第二笔累加：0.10 + 0.20 == 0.30，且小数必须精确（十进制往返，不能被整数截断）。
	total, err = cache.ReserveUserBalance(ctx, userID, 0.20, 10*time.Minute)
	require.NoError(s.T(), err, "ReserveUserBalance 第二笔")
	require.InDelta(s.T(), 0.30, total, 1e-9, "两笔预留应累加为 0.30")

	exists, err := rdb.Exists(ctx, key).Result()
	require.NoError(s.T(), err, "Exists")
	require.Equal(s.T(), int64(1), exists, "预留键应存在")

	ttl, err := rdb.TTL(ctx, key).Result()
	require.NoError(s.T(), err, "TTL")
	s.AssertTTLWithin(ttl, 9*time.Minute, 10*time.Minute)
}

func (s *BillingReservationSuite) TestReserveFallsBackToCacheTTLWhenTTLMissing() {
	cache, rdb := s.reservationCache()
	ctx := context.Background()
	userID := int64(9002)
	key := fmt.Sprintf("%s%d", billingReservedKeyPrefix, userID)

	_, err := cache.ReserveUserBalance(ctx, userID, 0.05, 0)
	require.NoError(s.T(), err, "ReserveUserBalance 应以缺省 TTL 兜底")

	ttl, err := rdb.TTL(ctx, key).Result()
	require.NoError(s.T(), err, "TTL")
	s.AssertTTLWithin(ttl, billingCacheTTL-30*time.Second, billingCacheTTL)
}

func (s *BillingReservationSuite) TestReserveRejectsNegativeAmount() {
	cache, _ := s.reservationCache()
	ctx := context.Background()

	_, err := cache.ReserveUserBalance(ctx, 9003, -0.10, 10*time.Minute)
	require.Error(s.T(), err, "负数预留必须被拒绝（否则可凭空放大可花余额）")
}

func (s *BillingReservationSuite) TestReleaseReducesAndDeletesAtZero() {
	cache, rdb := s.reservationCache()
	ctx := context.Background()
	userID := int64(9004)
	key := fmt.Sprintf("%s%d", billingReservedKeyPrefix, userID)

	_, err := cache.ReserveUserBalance(ctx, userID, 0.30, 10*time.Minute)
	require.NoError(s.T(), err, "ReserveUserBalance")

	require.NoError(s.T(), cache.ReleaseUserBalanceReservation(ctx, userID, 0.10, 10*time.Minute), "归还部分预留")

	rest, err := cache.ReserveUserBalance(ctx, userID, 0, 10*time.Minute)
	require.NoError(s.T(), err, "读回总额（0 金额累加无副作用）")
	require.InDelta(s.T(), 0.20, rest, 1e-9, "归还 0.10 后应剩 0.20")

	// 归零：键必须被删除（DEL），不能留下 0 值键或负值。
	require.NoError(s.T(), cache.ReleaseUserBalanceReservation(ctx, userID, 0.20, 10*time.Minute), "归还剩余预留")
	exists, err := rdb.Exists(ctx, key).Result()
	require.NoError(s.T(), err, "Exists")
	require.Equal(s.T(), int64(0), exists, "归零后预留键必须被删除")

	// 键已删除后再归还：no-op，不得把总额减成负数（复现"凭空多放行额度"的坏账方向）。
	require.NoError(s.T(), cache.ReleaseUserBalanceReservation(ctx, userID, 0.10, 10*time.Minute), "对缺失键归还应为 no-op")
	exists, err = rdb.Exists(ctx, key).Result()
	require.NoError(s.T(), err, "Exists after no-op release")
	require.Equal(s.T(), int64(0), exists, "no-op 归还不得重新创建键")
}

func (s *BillingReservationSuite) TestReleaseOnMissingKeyIsNoop() {
	cache, rdb := s.reservationCache()
	ctx := context.Background()
	userID := int64(9005)
	key := fmt.Sprintf("%s%d", billingReservedKeyPrefix, userID)

	require.NoError(s.T(), cache.ReleaseUserBalanceReservation(ctx, userID, 0.10, 10*time.Minute), "从未预留过也要 no-op 成功")

	exists, err := rdb.Exists(ctx, key).Result()
	require.NoError(s.T(), err, "Exists")
	require.Equal(s.T(), int64(0), exists, "no-op 归还不得创建键")
}

func (s *BillingReservationSuite) TestConcurrentReserveAccumulatesExactly() {
	cache, _ := s.reservationCache()
	ctx := context.Background()
	userID := int64(9006)

	const workers = 20
	const perWorker = 0.10

	var wg sync.WaitGroup
	totals := make([]float64, workers)
	errs := make([]error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			totals[idx], errs[idx] = cache.ReserveUserBalance(ctx, userID, perWorker, 10*time.Minute)
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		require.NoError(s.T(), err, "goroutine %d 预留失败", i)
	}

	// 20 × 0.10 == 2.00：INCRBYFLOAT 的十进制累加必须精确到浮点误差内。
	final, err := cache.ReserveUserBalance(ctx, userID, 0, 10*time.Minute)
	require.NoError(s.T(), err, "读回并发累加总额")
	require.InDelta(s.T(), 2.00, final, 1e-9, "20 笔并发预留应精确累加为 2.00")

	// 快照式读取曾出现"最后一个写入者"的形态：确认累计值不是单笔值。
	require.Greater(s.T(), final, perWorker, "并发累加不得退化为单笔覆盖")
}

func TestBillingReservationSuite(t *testing.T) {
	suite.Run(t, new(BillingReservationSuite))
}

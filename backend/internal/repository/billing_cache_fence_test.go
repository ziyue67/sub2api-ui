//go:build unit

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

// newFencedBalanceCache 构造互不干扰的 miniredis 夹具，并把 cache 提升为可断言
// 栅栏能力的接口（billingBalanceFenceStoreForTest）。
type billingBalanceFenceStoreForTest interface {
	InitUserBalance(ctx context.Context, userID int64, balance float64) (bool, error)
	SetUserBalanceFenced(ctx context.Context, userID int64, balance float64) (bool, error)
}

func newFencedBalanceCache(t *testing.T) (*miniredis.Miniredis, *billingCache, billingBalanceFenceStoreForTest) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	cache, ok := NewBillingCache(rdb).(*billingCache)
	require.True(t, ok, "NewBillingCache 应返回 *billingCache")
	store, ok := any(cache).(billingBalanceFenceStoreForTest)
	require.True(t, ok, "billingCache 必须实现余额缓存栅栏能力（审计 R9）")
	return mr, cache, store
}

// TestBillingBalanceFence_ExistingKeyIsNeverOverwritten 锁死"已存在的键只刷新 TTL"：
// Redis 侧的原子扣减可能已经领先于回源快照，回源写回绝不能把余额写回高位。
func TestBillingBalanceFence_ExistingKeyIsNeverOverwritten(t *testing.T) {
	mr, cache, store := newFencedBalanceCache(t)
	ctx := context.Background()

	fenced, err := store.InitUserBalance(ctx, 11, 0.10)
	require.NoError(t, err)
	require.False(t, fenced, "冷键在没有栅栏时应完成初始化")

	// 模拟"并发扣费已把缓存降到 0.10，而回源读到的是扣费前的 0.90"。
	fenced, err = store.InitUserBalance(ctx, 11, 0.90)
	require.NoError(t, err)
	require.False(t, fenced, "键已存在不代表被栅栏拦截")

	got, err := cache.GetUserBalance(ctx, 11)
	require.NoError(t, err)
	require.InDelta(t, 0.10, got, 1e-9, "回源快照不得覆盖已存在的余额键")

	ttl := mr.TTL(billingBalanceKey(11))
	require.Greater(t, ttl, billingCacheTTL-30*time.Second, "已存在的键应刷新 TTL")
	require.LessOrEqual(t, ttl, billingCacheTTL, "键 TTL 不应超过缓存上界")
}

// TestBillingBalanceFence_InvalidateBlocksHydrationUntilExpiry 验证失效路径的核心不变式：
// 失效会同时抬栅栏并删键，栅栏存活期内任何回源写回（初始化/真值回填）都必须被拒绝；
// 栅栏自然过期后才允许重新建立缓存。
func TestBillingBalanceFence_InvalidateBlocksHydrationUntilExpiry(t *testing.T) {
	mr, cache, store := newFencedBalanceCache(t)
	ctx := context.Background()

	require.NoError(t, cache.SetUserBalance(ctx, 12, 5.0))
	require.NoError(t, cache.InvalidateUserBalance(ctx, 12))

	require.False(t, mr.Exists(billingBalanceKey(12)), "失效必须删掉余额键")
	require.True(t, mr.Exists(billingBalanceFenceKey(12)), "失效必须抬起变更栅栏")
	require.Greater(t, mr.TTL(billingBalanceFenceKey(12)), billingBalanceFenceTTL-time.Second)

	// 持有旧快照的回源写回：初始化与真值回填都必须被拦下。
	fenced, err := store.InitUserBalance(ctx, 12, 5.0)
	require.NoError(t, err)
	require.True(t, fenced, "栅栏存活时初始化必须被拒绝")
	require.False(t, mr.Exists(billingBalanceKey(12)), "被拒绝的初始化不得建立键")

	writeFenced, err := store.SetUserBalanceFenced(ctx, 12, 5.0)
	require.NoError(t, err)
	require.True(t, writeFenced, "栅栏存活时真值回填必须被拒绝")
	require.False(t, mr.Exists(billingBalanceKey(12)), "被拒绝的回填不得建立键")

	// 栅栏过期后自愈：下一次回源可以正常建立缓存。
	mr.FastForward(billingBalanceFenceTTL + time.Second)
	fenced, err = store.InitUserBalance(ctx, 12, 4.0)
	require.NoError(t, err)
	require.False(t, fenced, "栅栏过期后允许重新建立缓存")

	got, err := cache.GetUserBalance(ctx, 12)
	require.NoError(t, err)
	require.InDelta(t, 4.0, got, 1e-9)
}

// TestBillingBalanceFence_UnconditionalSetStillWins 记录有意的例外：
// SetUserBalance（无条件写）不受栅栏约束，仅用于测试/直接写入语义，生产回源路径
// 一律走 InitUserBalance / SetUserBalanceFenced（见 service 层调用点）。
func TestBillingBalanceFence_UnconditionalSetStillWins(t *testing.T) {
	mr, cache, _ := newFencedBalanceCache(t)
	ctx := context.Background()

	require.NoError(t, cache.InvalidateUserBalance(ctx, 13))
	require.True(t, mr.Exists(billingBalanceFenceKey(13)))

	require.NoError(t, cache.SetUserBalance(ctx, 13, 7.0))
	got, err := cache.GetUserBalance(ctx, 13)
	require.NoError(t, err)
	require.InDelta(t, 7.0, got, 1e-9, "无条件写语义保持不变（历史调用点/测试依赖）")
}

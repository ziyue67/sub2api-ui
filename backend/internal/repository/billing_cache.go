package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/redis/go-redis/v9"
)

const (
	billingBalanceKeyPrefix   = "billing:balance:"
	billingSubKeyPrefix       = "billing:sub:"
	billingRateLimitKeyPrefix = "apikey:rate:"
	// billingBalanceExhaustedKeyPrefix 是"钱包已耗尽（余额已到 reserve 底线）"标记键。
	// 该标记与余额缓存（billing:balance:）严格分离：只有"结算判定钱包已耗尽"与
	// "余额增加"两条路径会写它，余额未命中回源的异步写回永远不会写它，
	// 因此它不会被扣费前的旧余额快照"复活"。
	billingBalanceExhaustedKeyPrefix = "billing:balance_exhausted:"
	// billingReservedKeyPrefix 是"在途预留"键前缀：该用户当前全部"已放行、尚未结算"请求的
	// 最坏费用上界之和（USD）。预检放行前用 INCRBYFLOAT 原子累加、结算完成后归还，
	// 使"余额 - 在途预留 >= 封底"成为跨请求的原子准入护栏，堵住并发突发时用同一份
	// 余额快照全部放行、结算时却扣不动钱的坏账（键带 TTL 自愈，崩溃/丢任务不会永久钉住余额）。
	billingReservedKeyPrefix = "billing:reserved:"
	// billingReservedItemKeyPrefix 是"单笔预留凭据"键前缀：每个请求用**自己的**
	// requestID 在 Redis 上留一条金额凭据（TTL 与预留一致）。
	//
	// 为什么必须有它：上面的聚合键只有"总额"一个标量，无法回答"这笔归还到底是不是
	// 我自己那笔"。当某个请求在途时间超过预留 TTL（长流式响应完全可能），它的聚合份额
	// 已被 TTL 自愈清掉，而它结算时的递减会**吃掉同一用户后来那笔请求的预留**，
	// 甚至把聚合键 DEL 掉 —— 护栏被错误解除，并发超额窗口原样复现。
	//
	// 有了凭据键，归还变成"只有我这条凭据还在，才允许递减，且递减金额以凭据为准"：
	// 凭据过期的归还一律 no-op，绝不会触碰别人的预留。代价是聚合键可能短暂偏高
	// （已过期但未归还的份额要等聚合键 TTL 兜底），这是保守方向，不会放行超额。
	billingReservedItemKeyPrefix = "billing:resv_item:"
	subCacheInvalidateChannel = "subscription:cache:invalidate"
	billingCacheTTL           = 5 * time.Minute
	billingCacheJitter        = 30 * time.Second
	// balanceExhaustedMarkerTTL 必须 >= 余额缓存的最长存活时间（billingCacheTTL），
	// 否则标记先过期、而余额缓存里仍留着偏高的旧值，预检又会被放行。
	balanceExhaustedMarkerTTL = billingCacheTTL + time.Minute
	rateLimitCacheTTL         = 7 * 24 * time.Hour // 7 days matches the longest window

	// Rate limit window durations — must match service.RateLimitWindow* constants.
	rateLimitWindow5h = 5 * time.Hour
	rateLimitWindow1d = 24 * time.Hour
	rateLimitWindow7d = 7 * 24 * time.Hour
)

// jitteredTTL 返回带随机抖动的 TTL，防止缓存雪崩
func jitteredTTL() time.Duration {
	// 只做“减法抖动”，确保实际 TTL 不会超过 billingCacheTTL（避免上界预期被打破）。
	if billingCacheJitter <= 0 {
		return billingCacheTTL
	}
	jitter := time.Duration(rand.IntN(int(billingCacheJitter)))
	return billingCacheTTL - jitter
}

// billingBalanceKey generates the Redis key for user balance cache.
func billingBalanceKey(userID int64) string {
	return fmt.Sprintf("%s%d", billingBalanceKeyPrefix, userID)
}

// billingBalanceExhaustedKey generates the Redis key for the "wallet exhausted" marker.
func billingBalanceExhaustedKey(userID int64) string {
	return fmt.Sprintf("%s%d", billingBalanceExhaustedKeyPrefix, userID)
}

// billingReservedKey generates the Redis key for a reservation scope's in-flight total.
//
// scope 由服务层构造，必须是"同一份可被并发放大的额度"的唯一标识：
//   - 余额模式：用户的 userID；
//   - 订阅模式："sub:<userID>:<groupID>"（限额挂在用户 × 分组上，与余额互不干扰）。
func billingReservedKey(scope string) string {
	return billingReservedKeyPrefix + scope
}

// billingReservedItemKey generates the Redis key holding one request's reservation
// receipt (amount + its own TTL). scope/requestID must already be sanitized by the caller.
func billingReservedItemKey(scope, requestID string) string {
	return billingReservedItemKeyPrefix + scope + ":" + requestID
}

// billingSubKey generates the Redis key for subscription cache.
func billingSubKey(userID, groupID int64) string {
	return fmt.Sprintf("%s%d:%d", billingSubKeyPrefix, userID, groupID)
}

const (
	subFieldStatus       = "status"
	subFieldExpiresAt    = "expires_at"
	subFieldDailyUsage   = "daily_usage"
	subFieldWeeklyUsage  = "weekly_usage"
	subFieldMonthlyUsage = "monthly_usage"
	subFieldVersion      = "version"
)

// billingRateLimitKey generates the Redis key for API key rate limit cache.
func billingRateLimitKey(keyID int64) string {
	return fmt.Sprintf("%s%d", billingRateLimitKeyPrefix, keyID)
}

const (
	rateLimitFieldUsage5h  = "usage_5h"
	rateLimitFieldUsage1d  = "usage_1d"
	rateLimitFieldUsage7d  = "usage_7d"
	rateLimitFieldWindow5h = "window_5h"
	rateLimitFieldWindow1d = "window_1d"
	rateLimitFieldWindow7d = "window_7d"
)

var (
	deductBalanceScript = redis.NewScript(`
		local current = redis.call('GET', KEYS[1])
		if current == false then
			return 0
		end
		local newVal = tonumber(current) - tonumber(ARGV[1])
		redis.call('SET', KEYS[1], newVal)
		redis.call('EXPIRE', KEYS[1], ARGV[2])
		return 1
	`)

	// reserveBalanceScript 原子地"落凭据 + 累加总额"：为本次请求写入一条带 TTL 的
	// 预留凭据，并把它累加进用户的在途预留总额，返回累加后的总额（十进制文本）。
	//
	// 用 INCRBYFLOAT 而不是 GET+SET：单命令原子、天然支持"键不存在即从 0 开始"，
	// 且返回值是十进制字符串（不像 Lua number 会被 RESP 整数截断掉小数）。
	//
	// 幂等：同 requestID 重复预留（重试/重复绑定）只刷新 TTL，不重复累加，
	// 否则同一笔请求会被记两次额度。
	//
	// KEYS[1] = billing:reserved:<userID>（聚合总额）
	// KEYS[2] = billing:resv_item:<userID>:<requestID>（本笔凭据）
	// ARGV[1] = 预留金额（USD，>0），ARGV[2] = TTL（毫秒）
	reserveBalanceScript = redis.NewScript(`
		local placed = redis.call('SET', KEYS[2], ARGV[1], 'NX', 'PX', ARGV[2])
		if not placed then
			redis.call('PEXPIRE', KEYS[2], ARGV[2])
			local current = redis.call('GET', KEYS[1])
			if current == false then
				return '0'
			end
			return current
		end
		local newVal = redis.call('INCRBYFLOAT', KEYS[1], ARGV[1])
		redis.call('PEXPIRE', KEYS[1], ARGV[2])
		return newVal
	`)

	// releaseBalanceScript 原子地归还一笔在途预留，**且只归还属于本请求的那一笔**。
	//
	// 关键约束：凭据键不存在（预留已随 TTL 过期 / 从未建立）时必须整体 no-op。
	// 旧实现只看聚合键是否存在，于是"晚到的归还"会递减别人刚建立的预留 ——
	// 出现"护栏被错误解除、并发窗口重开"的方向性错误。现在递减金额取自本请求的
	// 凭据，凭据没了就绝不触碰聚合键。
	//
	// 返回值：'0'（已归还/无需归还）、新的总额字符串、或 'EXPIRED'
	// （凭据已过期，本次归还被安全忽略）。
	//
	// KEYS[1] = billing:reserved:<userID>，KEYS[2] = billing:resv_item:<userID>:<requestID>
	// ARGV[1] = 兜底 TTL（毫秒）
	releaseBalanceScript = redis.NewScript(`
		local amount = redis.call('GET', KEYS[2])
		if amount == false then
			return 'EXPIRED'
		end
		redis.call('DEL', KEYS[2])
		local exists = redis.call('EXISTS', KEYS[1])
		if exists == 0 then
			return '0'
		end
		local newVal = redis.call('INCRBYFLOAT', KEYS[1], '-' .. amount)
		if tonumber(newVal) <= 1e-07 then
			redis.call('DEL', KEYS[1])
			return '0'
		end
		local pttl = redis.call('PTTL', KEYS[1])
		if pttl < 0 then
			redis.call('PEXPIRE', KEYS[1], ARGV[1])
		end
		return newVal
	`)

	// renewBalanceScript 为**长请求**续期预留：只有凭据仍在时才刷新凭据与聚合键的
	// TTL。没有它，超过预留 TTL 的流式请求会在结算前就丢掉护栏保护。
	//
	// KEYS[1] = billing:reserved:<userID>，KEYS[2] = billing:resv_item:<userID>:<requestID>
	// ARGV[1] = TTL（毫秒）
	renewBalanceScript = redis.NewScript(`
		if redis.call('EXISTS', KEYS[2]) == 0 then
			return '0'
		end
		redis.call('PEXPIRE', KEYS[2], ARGV[1])
		redis.call('PEXPIRE', KEYS[1], ARGV[1])
		return '1'
	`)

	// peekBalanceScript 只读返回聚合总额（不存在时返回 '0'），供运维/测试观测用，
	// 不引入新的键、也不改变 TTL。
	//
	// KEYS[1] = billing:reserved:<userID>
	peekBalanceScript = redis.NewScript(`
		local current = redis.call('GET', KEYS[1])
		if current == false then
			return '0'
		end
		return current
	`)

	updateSubUsageScript = redis.NewScript(`
		local exists = redis.call('EXISTS', KEYS[1])
		if exists == 0 then
			return 0
		end
		local cost = tonumber(ARGV[1])
		redis.call('HINCRBYFLOAT', KEYS[1], 'daily_usage', cost)
		redis.call('HINCRBYFLOAT', KEYS[1], 'weekly_usage', cost)
		redis.call('HINCRBYFLOAT', KEYS[1], 'monthly_usage', cost)
		redis.call('EXPIRE', KEYS[1], ARGV[2])
		return 1
	`)

	// updateRateLimitUsageScript atomically increments all three rate limit usage counters
	// with window expiration checking. If a window has expired, its usage is reset to cost
	// (instead of accumulated) and the window timestamp is updated, matching the DB-side
	// IncrementRateLimitUsage semantics.
	//
	// ARGV: [1]=cost, [2]=ttl_seconds, [3]=now_unix, [4]=window_5h_seconds, [5]=window_1d_seconds, [6]=window_7d_seconds
	updateRateLimitUsageScript = redis.NewScript(`
		local exists = redis.call('EXISTS', KEYS[1])
		if exists == 0 then
			return 0
		end
		local cost = tonumber(ARGV[1])
		local now = tonumber(ARGV[3])
		local win5h = tonumber(ARGV[4])
		local win1d = tonumber(ARGV[5])
		local win7d = tonumber(ARGV[6])

		-- Helper: check if window is expired and update usage + window accordingly
		-- Returns nothing, modifies the hash in-place.
		local function update_window(usage_field, window_field, window_duration)
			local w = tonumber(redis.call('HGET', KEYS[1], window_field) or 0)
			if w == 0 or (now - w) >= window_duration then
				-- Window expired or never started: reset usage to cost, start new window
				redis.call('HSET', KEYS[1], usage_field, tostring(cost))
				redis.call('HSET', KEYS[1], window_field, tostring(now))
			else
				-- Window still valid: accumulate
				redis.call('HINCRBYFLOAT', KEYS[1], usage_field, cost)
			end
		end

		update_window('usage_5h', 'window_5h', win5h)
		update_window('usage_1d', 'window_1d', win1d)
		update_window('usage_7d', 'window_7d', win7d)
		redis.call('EXPIRE', KEYS[1], ARGV[2])
		return 1
	`)
)

type billingCache struct {
	rdb *redis.Client
}

func NewBillingCache(rdb *redis.Client) service.BillingCache {
	return &billingCache{rdb: rdb}
}

func (c *billingCache) GetUserBalance(ctx context.Context, userID int64) (float64, error) {
	key := billingBalanceKey(userID)
	val, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(val, 64)
}

func (c *billingCache) SetUserBalance(ctx context.Context, userID int64, balance float64) error {
	key := billingBalanceKey(userID)
	return c.rdb.Set(ctx, key, balance, jitteredTTL()).Err()
}

func (c *billingCache) DeductUserBalance(ctx context.Context, userID int64, amount float64) error {
	key := billingBalanceKey(userID)
	_, err := deductBalanceScript.Run(ctx, c.rdb, []string{key}, amount, int(jitteredTTL().Seconds())).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Printf("Warning: deduct balance cache failed for user %d: %v", userID, err)
		return err
	}
	return nil
}

func (c *billingCache) InvalidateUserBalance(ctx context.Context, userID int64) error {
	key := billingBalanceKey(userID)
	return c.rdb.Del(ctx, key).Err()
}

// MarkUserBalanceExhausted 记录"该用户的钱包已经没有可花余额（已到 reserve 底线）"。
//
// 计费是后付费：一旦结算才发现余额不足，上游成本已经发生。因此转发前的预检必须
// 尽可能 fail-closed。但预检读的是余额缓存，而余额缓存采用"未命中回源 + 异步写回"，
// 扣费后的 InvalidateUserBalance(DEL) 可能被扣费前读到的旧余额写回"复活"，
// 使预检在缓存 TTL（5 分钟）内持续放行注定扣费失败的请求。
//
// 这个标记由结算路径在"扣到底线 / 扣费被拒"时写入，回源路径永不写它，因此不受该竞态影响。
func (c *billingCache) MarkUserBalanceExhausted(ctx context.Context, userID int64) error {
	key := billingBalanceExhaustedKey(userID)
	return c.rdb.Set(ctx, key, 1, balanceExhaustedMarkerTTL).Err()
}

// ClearUserBalanceExhausted 清除"钱包已耗尽"标记。任何让余额增加的路径（充值、兑换、
// 返利、管理员调整）都必须调用它，否则刚充值的用户在标记 TTL 内仍会被预检拦截。
func (c *billingCache) ClearUserBalanceExhausted(ctx context.Context, userID int64) error {
	key := billingBalanceExhaustedKey(userID)
	return c.rdb.Del(ctx, key).Err()
}

// IsUserBalanceExhausted 查询"钱包已耗尽"标记是否存在。
func (c *billingCache) IsUserBalanceExhausted(ctx context.Context, userID int64) (bool, error) {
	key := billingBalanceExhaustedKey(userID)
	n, err := c.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// ReserveUserBalance 原子地为**某一次请求**建立在途预留：写入该请求的预留凭据，
// 并把它累加进该 scope 的在途预留总额，返回累加后的总额（USD）。
//
// 调用方（BillingCacheService 的预检）用它的返回值判断"这份可花额度减去在途预留后
// 是否仍然充足"：余额模式看 `balance - reserved >= reserve`，订阅模式看
// `usage + reserved <= limit`。只要成立，即使全部在途请求都以最坏费用结算，
// 也不会击穿额度，不存在收不满的坏账。
//
// scope 标识"同一份可被并发放大的额度"（余额用 userID，订阅用 user+group）；
// requestID 必须是**每笔请求唯一且不含分隔符 ':'** 的标识（服务层用去掉连字符的
// UUID），它决定归还时能否精确匹配到本笔凭据。
func (c *billingCache) ReserveUserBalance(ctx context.Context, scope string, requestID string, amount float64, ttl time.Duration) (float64, error) {
	if amount < 0 {
		return 0, fmt.Errorf("reserve amount must be nonnegative, got %v", amount)
	}
	if scope == "" {
		return 0, fmt.Errorf("reserve scope must not be empty")
	}
	if requestID == "" {
		return 0, fmt.Errorf("reserve requestID must not be empty")
	}
	reply, err := reserveBalanceScript.Run(ctx, c.rdb,
		[]string{billingReservedKey(scope), billingReservedItemKey(scope, requestID)},
		amount, reservationTTLMillis(ttl)).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return 0, err
	}
	return parseReservedBalanceReply(reply)
}

// ReleaseUserBalanceReservation 原子地归还"本请求"占用的在途预留。
//
// 与旧实现的关键差别：归还的前提是**本请求的凭据仍然存在**。凭据已随 TTL 过期时
// 整体 no-op 并返回 ErrBillingReservationExpired —— 少了这道闸，晚到的归还会把
// 同一 scope 后来那笔请求的预留一起扣掉，护栏被错误解除。
func (c *billingCache) ReleaseUserBalanceReservation(ctx context.Context, scope string, requestID string, amount float64, ttl time.Duration) error {
	if scope == "" || requestID == "" {
		return nil
	}
	if amount <= 0 {
		return nil
	}
	reply, err := releaseBalanceScript.Run(ctx, c.rdb,
		[]string{billingReservedKey(scope), billingReservedItemKey(scope, requestID)},
		reservationTTLMillis(ttl)).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Printf("Warning: release balance reservation failed for scope %s: %v", scope, err)
		return err
	}
	if text, ok := reply.(string); ok && text == reservationReleaseExpiredReply {
		return service.ErrBillingReservationExpired
	}
	return nil
}

// RenewUserBalanceReservation 为**长请求**续期本笔预留（心跳）。凭据已过期/已归还时返回
// ErrBillingReservationExpired，调用方据此停止心跳。
func (c *billingCache) RenewUserBalanceReservation(ctx context.Context, scope string, requestID string, ttl time.Duration) error {
	if scope == "" || requestID == "" {
		return nil
	}
	reply, err := renewBalanceScript.Run(ctx, c.rdb,
		[]string{billingReservedKey(scope), billingReservedItemKey(scope, requestID)},
		reservationTTLMillis(ttl)).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Printf("Warning: renew balance reservation failed for scope %s: %v", scope, err)
		return err
	}
	if text, ok := reply.(string); ok && text == "0" {
		return service.ErrBillingReservationExpired
	}
	return nil
}

// ReservedUserBalanceTotal 只读返回该 scope 当前的在途预留总额（USD），供运维观测使用。
// 不存在的键返回 0，不创建键、不改变 TTL。
func (c *billingCache) ReservedUserBalanceTotal(ctx context.Context, scope string) (float64, error) {
	if scope == "" {
		return 0, nil
	}
	reply, err := peekBalanceScript.Run(ctx, c.rdb, []string{billingReservedKey(scope)}).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return 0, err
	}
	return parseReservedBalanceReply(reply)
}

// reservationReleaseExpiredReply 是归还脚本在"本请求凭据已过期"时的返回值。
// 归一成一个常量，避免脚本字面量与 Go 侧判断漂移。
const reservationReleaseExpiredReply = "EXPIRED"

// reservationTTLMillis 把预留 TTL 规范化为毫秒；缺省/非法值退化为余额缓存 TTL。
func reservationTTLMillis(ttl time.Duration) int64 {
	millis := ttl.Milliseconds()
	if millis <= 0 {
		millis = billingCacheTTL.Milliseconds()
	}
	return millis
}

// parseReservedBalanceReply 解析预留脚本的十进制文本返回值，兼容 RESP2/RESP3 以及
// 集群客户端可能出现的多种返回形态（数字必须以文本往返，避免被整数截断）。
func parseReservedBalanceReply(reply any) (float64, error) {
	switch v := reply.(type) {
	case nil:
		return 0, nil
	case string:
		return strconv.ParseFloat(v, 64)
	case []byte:
		return strconv.ParseFloat(string(v), 64)
	case float64:
		return v, nil
	case int64:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("unexpected balance reservation reply type %T", reply)
	}
}

func (c *billingCache) GetSubscriptionCache(ctx context.Context, userID, groupID int64) (*service.SubscriptionCacheData, error) {
	key := billingSubKey(userID, groupID)
	result, err := c.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, redis.Nil
	}
	return c.parseSubscriptionCache(result)
}

func (c *billingCache) parseSubscriptionCache(data map[string]string) (*service.SubscriptionCacheData, error) {
	result := &service.SubscriptionCacheData{}

	result.Status = data[subFieldStatus]
	if result.Status == "" {
		return nil, errors.New("invalid cache: missing status")
	}

	if expiresStr, ok := data[subFieldExpiresAt]; ok {
		expiresAt, err := strconv.ParseInt(expiresStr, 10, 64)
		if err == nil {
			result.ExpiresAt = time.Unix(expiresAt, 0)
		}
	}

	if dailyStr, ok := data[subFieldDailyUsage]; ok {
		result.DailyUsage, _ = strconv.ParseFloat(dailyStr, 64)
	}

	if weeklyStr, ok := data[subFieldWeeklyUsage]; ok {
		result.WeeklyUsage, _ = strconv.ParseFloat(weeklyStr, 64)
	}

	if monthlyStr, ok := data[subFieldMonthlyUsage]; ok {
		result.MonthlyUsage, _ = strconv.ParseFloat(monthlyStr, 64)
	}

	if versionStr, ok := data[subFieldVersion]; ok {
		result.Version, _ = strconv.ParseInt(versionStr, 10, 64)
	}

	return result, nil
}

func (c *billingCache) SetSubscriptionCache(ctx context.Context, userID, groupID int64, data *service.SubscriptionCacheData) error {
	if data == nil {
		return nil
	}

	key := billingSubKey(userID, groupID)

	fields := map[string]any{
		subFieldStatus:       data.Status,
		subFieldExpiresAt:    data.ExpiresAt.Unix(),
		subFieldDailyUsage:   data.DailyUsage,
		subFieldWeeklyUsage:  data.WeeklyUsage,
		subFieldMonthlyUsage: data.MonthlyUsage,
		subFieldVersion:      data.Version,
	}

	pipe := c.rdb.Pipeline()
	pipe.HSet(ctx, key, fields)
	pipe.Expire(ctx, key, jitteredTTL())
	_, err := pipe.Exec(ctx)
	return err
}

func (c *billingCache) UpdateSubscriptionUsage(ctx context.Context, userID, groupID int64, cost float64) error {
	key := billingSubKey(userID, groupID)
	_, err := updateSubUsageScript.Run(ctx, c.rdb, []string{key}, cost, int(jitteredTTL().Seconds())).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Printf("Warning: update subscription usage cache failed for user %d group %d: %v", userID, groupID, err)
		return err
	}
	return nil
}

func (c *billingCache) InvalidateSubscriptionCache(ctx context.Context, userID, groupID int64) error {
	key := billingSubKey(userID, groupID)
	return c.rdb.Del(ctx, key).Err()
}

func (c *billingCache) PublishSubscriptionCacheInvalidation(ctx context.Context, cacheKey string) error {
	return c.rdb.Publish(ctx, subCacheInvalidateChannel, cacheKey).Err()
}

func (c *billingCache) SubscribeSubscriptionCacheInvalidation(ctx context.Context, handler func(cacheKey string)) error {
	pubsub := c.rdb.Subscribe(ctx, subCacheInvalidateChannel)

	_, err := pubsub.Receive(ctx)
	if err != nil {
		_ = pubsub.Close()
		return fmt.Errorf("subscribe to subscription cache invalidation: %w", err)
	}

	go func() {
		defer func() {
			if err := pubsub.Close(); err != nil {
				log.Printf("Warning: failed to close subscription cache invalidation pubsub: %v", err)
			}
		}()

		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				if msg != nil {
					handler(msg.Payload)
				}
			}
		}
	}()

	return nil
}

func (c *billingCache) GetAPIKeyRateLimit(ctx context.Context, keyID int64) (*service.APIKeyRateLimitCacheData, error) {
	key := billingRateLimitKey(keyID)
	result, err := c.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, redis.Nil
	}
	data := &service.APIKeyRateLimitCacheData{}
	if v, ok := result[rateLimitFieldUsage5h]; ok {
		data.Usage5h, _ = strconv.ParseFloat(v, 64)
	}
	if v, ok := result[rateLimitFieldUsage1d]; ok {
		data.Usage1d, _ = strconv.ParseFloat(v, 64)
	}
	if v, ok := result[rateLimitFieldUsage7d]; ok {
		data.Usage7d, _ = strconv.ParseFloat(v, 64)
	}
	if v, ok := result[rateLimitFieldWindow5h]; ok {
		data.Window5h, _ = strconv.ParseInt(v, 10, 64)
	}
	if v, ok := result[rateLimitFieldWindow1d]; ok {
		data.Window1d, _ = strconv.ParseInt(v, 10, 64)
	}
	if v, ok := result[rateLimitFieldWindow7d]; ok {
		data.Window7d, _ = strconv.ParseInt(v, 10, 64)
	}
	return data, nil
}

func (c *billingCache) SetAPIKeyRateLimit(ctx context.Context, keyID int64, data *service.APIKeyRateLimitCacheData) error {
	if data == nil {
		return nil
	}
	key := billingRateLimitKey(keyID)
	fields := map[string]any{
		rateLimitFieldUsage5h:  data.Usage5h,
		rateLimitFieldUsage1d:  data.Usage1d,
		rateLimitFieldUsage7d:  data.Usage7d,
		rateLimitFieldWindow5h: data.Window5h,
		rateLimitFieldWindow1d: data.Window1d,
		rateLimitFieldWindow7d: data.Window7d,
	}
	pipe := c.rdb.Pipeline()
	pipe.HSet(ctx, key, fields)
	pipe.Expire(ctx, key, rateLimitCacheTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *billingCache) UpdateAPIKeyRateLimitUsage(ctx context.Context, keyID int64, cost float64) error {
	key := billingRateLimitKey(keyID)
	now := time.Now().Unix()
	_, err := updateRateLimitUsageScript.Run(ctx, c.rdb, []string{key},
		cost,
		int(rateLimitCacheTTL.Seconds()),
		now,
		int(rateLimitWindow5h.Seconds()),
		int(rateLimitWindow1d.Seconds()),
		int(rateLimitWindow7d.Seconds()),
	).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Printf("Warning: update rate limit usage cache failed for api key %d: %v", keyID, err)
		return err
	}
	return nil
}

func (c *billingCache) InvalidateAPIKeyRateLimit(ctx context.Context, keyID int64) error {
	key := billingRateLimitKey(keyID)
	return c.rdb.Del(ctx, key).Err()
}

// ============================================
// user × platform quota 缓存
// ============================================

// userPlatformQuotaCacheKey 构造 Redis key
func userPlatformQuotaCacheKey(userID int64, platform string) string {
	return fmt.Sprintf("billing:user_platform_quota:%d:%s", userID, platform)
}

// parseUserPlatformQuotaHash 将 Redis HGETALL 返回的 map[string]string 反序列化为
// *service.UserPlatformQuotaCacheEntry。空 map（key 不存在）返回 nil。
// GetUserPlatformQuotaCache 和 BatchGetUserPlatformQuotaCache 共用此函数，确保解析逻辑一致。
func parseUserPlatformQuotaHash(m map[string]string) *service.UserPlatformQuotaCacheEntry {
	if len(m) == 0 {
		return nil
	}
	parseFloat := func(s string) float64 {
		if s == "" {
			return 0
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			log.Printf("billing_cache: corrupt quota usage field %q (using 0): %v", s, err)
			return 0
		}
		return f
	}
	parseFloatPtr := func(s string) *float64 {
		if s == "" {
			return nil
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil
		}
		return &f
	}
	parseTimePtr := func(s string) *time.Time {
		if s == "" {
			return nil
		}
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil
		}
		t := time.Unix(n, 0).UTC()
		return &t
	}
	parseInt64 := func(s string) int64 {
		n, _ := strconv.ParseInt(s, 10, 64)
		return n
	}
	return &service.UserPlatformQuotaCacheEntry{
		DailyUsageUSD:      parseFloat(m["daily_usage"]),
		WeeklyUsageUSD:     parseFloat(m["weekly_usage"]),
		MonthlyUsageUSD:    parseFloat(m["monthly_usage"]),
		Version:            parseInt64(m["version"]),
		SchemaVersion:      parseInt64(m["schema_version"]),
		DailyLimitUSD:      parseFloatPtr(m["daily_limit"]),
		WeeklyLimitUSD:     parseFloatPtr(m["weekly_limit"]),
		MonthlyLimitUSD:    parseFloatPtr(m["monthly_limit"]),
		DailyWindowStart:   parseTimePtr(m["daily_window_start"]),
		WeeklyWindowStart:  parseTimePtr(m["weekly_window_start"]),
		MonthlyWindowStart: parseTimePtr(m["monthly_window_start"]),
	}
}

func (c *billingCache) GetUserPlatformQuotaCache(ctx context.Context, userID int64, platform string) (*service.UserPlatformQuotaCacheEntry, bool, error) {
	key := userPlatformQuotaCacheKey(userID, platform)
	m, err := c.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, false, err
	}
	entry := parseUserPlatformQuotaHash(m)
	if entry == nil {
		// 空 map → key 不存在 → MISS
		return nil, false, nil
	}
	return entry, true, nil
}

func (c *billingCache) SetUserPlatformQuotaCache(ctx context.Context, userID int64, platform string, entry *service.UserPlatformQuotaCacheEntry, ttl time.Duration) error {
	if entry == nil {
		return nil
	}
	key := userPlatformQuotaCacheKey(userID, platform)
	pipe := c.rdb.TxPipeline()

	// 浮点可空字段：nil → 空字符串（读取时 parseFloatPtr 返回 nil，表示无限额）
	fmtFloatPtr := func(p *float64) string {
		if p == nil {
			return ""
		}
		return strconv.FormatFloat(*p, 'f', -1, 64)
	}
	// time.Time 可空字段：nil → 空字符串；有值 → unix 秒
	fmtTimePtr := func(p *time.Time) string {
		if p == nil {
			return ""
		}
		return strconv.FormatInt(p.Unix(), 10)
	}

	pipe.HSet(ctx, key,
		"daily_usage", entry.DailyUsageUSD,
		"weekly_usage", entry.WeeklyUsageUSD,
		"monthly_usage", entry.MonthlyUsageUSD,
		"version", entry.Version,
		"schema_version", entry.SchemaVersion,
		"daily_limit", fmtFloatPtr(entry.DailyLimitUSD),
		"weekly_limit", fmtFloatPtr(entry.WeeklyLimitUSD),
		"monthly_limit", fmtFloatPtr(entry.MonthlyLimitUSD),
		"daily_window_start", fmtTimePtr(entry.DailyWindowStart),
		"weekly_window_start", fmtTimePtr(entry.WeeklyWindowStart),
		"monthly_window_start", fmtTimePtr(entry.MonthlyWindowStart),
	)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *billingCache) DeleteUserPlatformQuotaCache(ctx context.Context, userID int64, platform string) error {
	return c.rdb.Del(ctx, userPlatformQuotaCacheKey(userID, platform)).Err()
}

// updateUserPlatformQuotaUsageScript 缓存累加：EXISTS + schema_version 双重守卫。
// 旧版 entry（schema_version != ARGV[3]，包括缺字段的 0 值）不参与累加，由上层走 DB fallback 后
// SetCache 重建为新版 entry —— 若此处仍累加，上层覆盖时会丢失这部分增量，导致 Redis usage 比真实偏小。
// key 不存在同样跳过（由下次 SetCache 重建）。
// KEYS[1] = hash key
// KEYS[2] = 脏集 key（dirty set）
// ARGV[1] = cost (string float)
// ARGV[2] = ttl seconds
// ARGV[3] = expected schema_version (Go 侧 UserPlatformQuotaCacheSchemaV1)
// ARGV[4] = dirty set member（空串则不 SADD）
// ARGV[5] = 脏集兜底 TTL 秒
const updateUserPlatformQuotaUsageScript = `
if redis.call("EXISTS", KEYS[1]) == 0 then
    return 0
end
local ver = redis.call("HGET", KEYS[1], "schema_version")
if ver == false or tonumber(ver) ~= tonumber(ARGV[3]) then
    return 0
end
redis.call("HINCRBYFLOAT", KEYS[1], "daily_usage", ARGV[1])
redis.call("HINCRBYFLOAT", KEYS[1], "weekly_usage", ARGV[1])
redis.call("HINCRBYFLOAT", KEYS[1], "monthly_usage", ARGV[1])
redis.call("HINCRBY", KEYS[1], "version", 1)
redis.call("EXPIRE", KEYS[1], ARGV[2])
if ARGV[4] ~= "" then
    redis.call("SADD", KEYS[2], ARGV[4])
    redis.call("EXPIRE", KEYS[2], ARGV[5])
end
return 1
`

// userPlatformQuotaDirtySetKey 返回脏集（dirty set）的 Redis key。
// 使用与 userPlatformQuotaCacheKey 相同的前缀 "billing:"。
func userPlatformQuotaDirtySetKey() string { return "billing:" + "upq:dirty" }

// userPlatformQuotaDirtyTTLSeconds 脏集兜底 TTL（秒）：初始 SADD（Lua）与 Readd 共用，
// 确保 flusher 长期停摆时脏集最终过期；正常运行因持续 SADD 不断续期。
const userPlatformQuotaDirtyTTLSeconds = 86400

// userPlatformQuotaDirtyMember 构造脏集成员字符串 "userID:platform"。
func userPlatformQuotaDirtyMember(userID int64, platform string) string {
	return strconv.FormatInt(userID, 10) + ":" + platform
}

func (c *billingCache) IncrUserPlatformQuotaUsageCache(ctx context.Context, userID int64, platform string, cost float64, ttl time.Duration, markDirty bool) error {
	member := ""
	if markDirty {
		member = userPlatformQuotaDirtyMember(userID, platform)
	}
	_, err := c.rdb.Eval(ctx, updateUserPlatformQuotaUsageScript,
		[]string{userPlatformQuotaCacheKey(userID, platform), userPlatformQuotaDirtySetKey()},
		strconv.FormatFloat(cost, 'f', -1, 64),
		int(ttl.Seconds()),
		service.UserPlatformQuotaCacheSchemaV1,
		member,
		userPlatformQuotaDirtyTTLSeconds,
	).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}
	return nil
}

// parseUserPlatformQuotaDirtyMember 将脏集成员字符串 "userID:platform" 解析为
// service.UserPlatformQuotaKey。解析失败返回 ok=false。
func parseUserPlatformQuotaDirtyMember(m string) (service.UserPlatformQuotaKey, bool) {
	parts := strings.SplitN(m, ":", 2)
	if len(parts) != 2 {
		return service.UserPlatformQuotaKey{}, false
	}
	uid, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return service.UserPlatformQuotaKey{}, false
	}
	return service.UserPlatformQuotaKey{UserID: uid, Platform: parts[1]}, true
}

// PopDirtyUserPlatformQuotaKeys 从脏集随机弹出最多 n 个 key。
// 脏集为空时返回 (nil, nil)。
func (c *billingCache) PopDirtyUserPlatformQuotaKeys(ctx context.Context, n int) ([]service.UserPlatformQuotaKey, error) {
	members, err := c.rdb.SPopN(ctx, userPlatformQuotaDirtySetKey(), int64(n)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	keys := make([]service.UserPlatformQuotaKey, 0, len(members))
	for _, m := range members {
		k, ok := parseUserPlatformQuotaDirtyMember(m)
		if !ok {
			log.Printf("billing_cache: skipping invalid dirty member %q", m)
			continue
		}
		keys = append(keys, k)
	}
	return keys, nil
}

// ReaddDirtyUserPlatformQuotaKeys 将 keys 重新加入脏集（flush 失败时回填）。
// 通过 pipeline 同时执行 SAdd + Expire，确保 Readd 后脏集具有兜底 TTL。
// 空切片时直接返回 nil。
func (c *billingCache) ReaddDirtyUserPlatformQuotaKeys(ctx context.Context, keys []service.UserPlatformQuotaKey) error {
	if len(keys) == 0 {
		return nil
	}
	dirtyKey := userPlatformQuotaDirtySetKey()
	members := make([]any, len(keys))
	for i, k := range keys {
		members[i] = userPlatformQuotaDirtyMember(k.UserID, k.Platform)
	}
	pipe := c.rdb.Pipeline()
	pipe.SAdd(ctx, dirtyKey, members...)
	pipe.Expire(ctx, dirtyKey, userPlatformQuotaDirtyTTLSeconds*time.Second)
	_, err := pipe.Exec(ctx)
	return err
}

// BatchGetUserPlatformQuotaCache 通过 Pipeline 批量 HGETALL 获取多个 user×platform 的
// quota cache。返回切片与 keys 顺序、长度对齐；MISS 或解析失败位置返回 nil。
func (c *billingCache) BatchGetUserPlatformQuotaCache(ctx context.Context, keys []service.UserPlatformQuotaKey) ([]*service.UserPlatformQuotaCacheEntry, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	pipe := c.rdb.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, len(keys))
	for i, k := range keys {
		cmds[i] = pipe.HGetAll(ctx, userPlatformQuotaCacheKey(k.UserID, k.Platform))
	}
	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}
	results := make([]*service.UserPlatformQuotaCacheEntry, len(keys))
	for i, cmd := range cmds {
		m, err := cmd.Result()
		if err != nil {
			if !errors.Is(err, redis.Nil) {
				log.Printf("billing_cache: BatchGet HGETALL cmd[%d] failed: %v (skip, self-heal)", i, err)
			}
			// 单个命令失败 → 对应位置 nil，继续
			continue
		}
		results[i] = parseUserPlatformQuotaHash(m)
	}
	return results, nil
}

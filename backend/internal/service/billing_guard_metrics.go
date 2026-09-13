package service

import (
	"sync/atomic"
	"time"
)

// 计费护栏统一可观测性。
//
// 背景：这条"后付费零超发"防线由多个修复叠加而成——结算封底（#3）、放行前最坏
// 费用预检（#4/#5）、在途预留（#6）、连接池与复核合并（#7）。但此前整条链只有
// write-off 一个指标，而且没有任何暴露点：
//
//   - 预检/预留的"降级、fail-open、被拒"全都只打日志，ops 无法判断护栏是否真的生效；
//   - "真的没钱"与"被最坏费用/预留护栏提前拦下"返回同一个 403 INSUFFICIENT_BALANCE，
//     客服与运营无法区分，用户投诉时无从解释；
//   - 上游桥不执行 max_tokens 这类"预检低估"只有结算封顶指标能间接反映。
//
// 本文件把这些状态变成可查询的数字。所有计数器都是无锁原子量（热路径不能引入
// 互斥），由 BillingGuardStatsSnapshot() 汇总，经 /admin/ops/billing-guard/health
// 暴露给 ops 面板。
//
// 健康基线：所有 reject 计数按斜率观察即可——余额不足的用户本来就会被拦，
// 关键是 *_fail_open / *_degrade / *abandoned / *release_error 这些
// **护栏失效类**计数应当恒为 0；不增长即说明护栏全程生效。

// BillingPreflightRejectReason 描述余额预检拒绝的**根本原因**。
//
// 这些原因共用同一个对外错误（ErrInsufficientBalance / 403 INSUFFICIENT_BALANCE，
// 保持既有 billing_error 形状不变），但内部语义完全不同：
//   - below_reserve / marker_active / db_truth_below_reserve：用户真的没钱，
//     属于正常业务拒绝，斜率上升是正常的；
//   - worst_case_* / inflight_reservation：用户还有余额，只是付不满"这一笔"的
//     最坏费用，或已被同用户的在途请求占用额度。这类拒绝若要上升，说明预检口径
//     偏保守（可能误伤），需要运维介入调参。
type BillingPreflightRejectReason string

const (
	// BillingRejectMarkerActive 命中"钱包已耗尽"标记（结算路径判定已扣不动）。
	BillingRejectMarkerActive BillingPreflightRejectReason = "marker_active"
	// BillingRejectBelowReserve 缓存余额已 ≤ 封底线。
	BillingRejectBelowReserve BillingPreflightRejectReason = "below_reserve"
	// BillingRejectDBTruthBelowReserve DB 真值复核确认余额 ≤ 封底线。
	BillingRejectDBTruthBelowReserve BillingPreflightRejectReason = "db_truth_below_reserve"
	// BillingRejectWorstCaseCache 无 userRepo（降级装配）时仅凭缓存判定吃不下最坏费用。
	BillingRejectWorstCaseCache BillingPreflightRejectReason = "worst_case_cache_only"
	// BillingRejectWorstCaseDBTruth DB 真值确认吃不下本次最坏费用。
	BillingRejectWorstCaseDBTruth BillingPreflightRejectReason = "worst_case_db_truth"
	// BillingRejectReservationGuard 在途预留累加后击穿封底（并发准入护栏生效）。
	BillingRejectReservationGuard BillingPreflightRejectReason = "inflight_reservation"
)

// AllBillingPreflightRejectReasons 列出全部拒绝原因，供快照与前端稳定枚举。
func AllBillingPreflightRejectReasons() []BillingPreflightRejectReason {
	return []BillingPreflightRejectReason{
		BillingRejectMarkerActive,
		BillingRejectBelowReserve,
		BillingRejectDBTruthBelowReserve,
		BillingRejectWorstCaseCache,
		BillingRejectWorstCaseDBTruth,
		BillingRejectReservationGuard,
	}
}

// 预检拒绝原因计数（按原因分离，避免"护栏误拦"淹没在"真的没钱"里）。
var (
	billingRejectMarkerActiveTotal        atomic.Int64
	billingRejectBelowReserveTotal        atomic.Int64
	billingRejectDBTruthBelowReserveTotal atomic.Int64
	billingRejectWorstCaseCacheTotal      atomic.Int64
	billingRejectWorstCaseDBTruthTotal    atomic.Int64
	billingRejectReservationGuardTotal    atomic.Int64
	billingRejectWorstCaseDBTruthLastUnix atomic.Int64
	billingRejectReservationLastUnix      atomic.Int64
)

// 在途预留计数。这些计数直接回答"并发护栏到底有没有在挡人"。
var (
	billingReservationReservedTotal       atomic.Int64
	billingReservationReleasedTotal       atomic.Int64
	billingReservationRejectedTotal       atomic.Int64
	billingReservationFailOpenTotal       atomic.Int64
	billingReservationFailOpenLastUnix    atomic.Int64
	billingReservationReleaseErrTotal     atomic.Int64
	billingReservationReleaseErrLastUnix  atomic.Int64
	billingReservationAbandonedTotal      atomic.Int64
	billingReservationAbandonedLastUnix   atomic.Int64
	billingReservationExpiredReleaseTotal atomic.Int64
	billingReservationRenewTotal          atomic.Int64
	billingReservationRenewErrTotal       atomic.Int64
)

// 复核与降级计数。这些是"护栏静默失效"的直接证据面。
var (
	billingRecheckDBReadsTotal              atomic.Int64
	billingRecheckFailClosedTotal           atomic.Int64
	billingRecheckSkippedNoUserRepoTotal    atomic.Int64
	billingPrecheckDisabledTotal            atomic.Int64
	billingPrecheckUnavailableTotal         atomic.Int64
	billingSubscriptionReservationTotal     atomic.Int64
	billingSubscriptionReserveFailOpenTotal atomic.Int64
)

// RecordBillingPreflightReject 记录一次余额预检拒绝及其原因。
func RecordBillingPreflightReject(reason BillingPreflightRejectReason) {
	now := time.Now().Unix()
	switch reason {
	case BillingRejectMarkerActive:
		billingRejectMarkerActiveTotal.Add(1)
	case BillingRejectBelowReserve:
		billingRejectBelowReserveTotal.Add(1)
	case BillingRejectDBTruthBelowReserve:
		billingRejectDBTruthBelowReserveTotal.Add(1)
	case BillingRejectWorstCaseCache:
		billingRejectWorstCaseCacheTotal.Add(1)
	case BillingRejectWorstCaseDBTruth:
		billingRejectWorstCaseDBTruthTotal.Add(1)
		billingRejectWorstCaseDBTruthLastUnix.Store(now)
	case BillingRejectReservationGuard:
		billingRejectReservationGuardTotal.Add(1)
		billingRejectReservationLastUnix.Store(now)
	}
}

// RecordBillingReservationAbandoned 记录"结算任务被丢弃、预留被提前放弃归还"。
//
// 出现即说明 usage_record worker 池按 drop/sample 语义丢掉了本应扣费的结算任务：
// 该笔请求既不会扣费（上游成本无法收回），预留也只能立刻归还而不是等结算完成。
// 健康值恒为 0。
func RecordBillingReservationAbandoned() {
	billingReservationAbandonedTotal.Add(1)
	billingReservationAbandonedLastUnix.Store(time.Now().Unix())
}

// BillingGuardStats 是护栏健康快照。
//
// 判读顺序建议：先看 ReservationFailOpen / RecheckSkippedNoUserRepo /
// ReservationAbandoned / ReservationReleaseErr —— 这四个只要非 0，就说明护栏在
// 某些条件下**静默失效**（不是拦得多，而是根本没拦）；再看
// RejectWorstCase / RejectInflightReservation 的斜率，判断预检是否过保守（误伤）。
type BillingGuardStats struct {
	// 结算侧：write-off（与 GatewayBillingShortfallStats 同源）。
	SettlementShortfallCount    int64 `json:"settlement_shortfall_count"`
	SettlementShortfallMicros   int64 `json:"settlement_shortfall_micros"`
	SettlementShortfallLastUnix int64 `json:"settlement_shortfall_last_unix"`

	// 预检拒绝（按原因）。
	RejectMarkerActive        int64 `json:"reject_marker_active"`
	RejectBelowReserve        int64 `json:"reject_below_reserve"`
	RejectDBTruthBelowReserve int64 `json:"reject_db_truth_below_reserve"`
	RejectWorstCaseCacheOnly  int64 `json:"reject_worst_case_cache_only"`
	RejectWorstCaseDBTruth    int64 `json:"reject_worst_case_db_truth"`
	RejectInflightReservation int64 `json:"reject_inflight_reservation"`
	RejectWorstCaseLastUnix   int64 `json:"reject_worst_case_last_unix"`
	RejectReservationLastUnix int64 `json:"reject_reservation_last_unix"`

	// 在途预留。
	ReservationReserved           int64 `json:"reservation_reserved"`
	ReservationReleased           int64 `json:"reservation_released"`
	ReservationRejected           int64 `json:"reservation_rejected"`
	ReservationFailOpen           int64 `json:"reservation_fail_open"`
	ReservationFailOpenLastUnix   int64 `json:"reservation_fail_open_last_unix"`
	ReservationReleaseErr         int64 `json:"reservation_release_error"`
	ReservationReleaseErrLastUnix int64 `json:"reservation_release_error_last_unix"`
	ReservationAbandoned          int64 `json:"reservation_abandoned"`
	ReservationAbandonedLastUnix  int64 `json:"reservation_abandoned_last_unix"`
	ReservationExpiredRelease     int64 `json:"reservation_expired_release"`
	ReservationRenew              int64 `json:"reservation_renew"`
	ReservationRenewError         int64 `json:"reservation_renew_error"`

	// DB 复核与降级。
	RecheckDBReads              int64 `json:"recheck_db_reads"`
	RecheckFailClosed           int64 `json:"recheck_fail_closed"`
	RecheckSkippedNoUserRepo    int64 `json:"recheck_skipped_no_user_repo"`
	PrecheckDisabled            int64 `json:"precheck_disabled"`
	PrecheckUnavailable         int64 `json:"precheck_unavailable"`
	SubscriptionReservation     int64 `json:"subscription_reservation"`
	SubscriptionReserveFailOpen int64 `json:"subscription_reservation_fail_open"`

	// 派生判据。
	Healthy              bool     `json:"healthy"`
	DegradedSignals      []string `json:"degraded_signals"`
	WorstCaseRejectRatio float64  `json:"worst_case_reject_ratio"`
}

// BillingGuardStatsSnapshot 汇总护栏健康快照。斜率由调用方按时间差分计算。
func BillingGuardStatsSnapshot() BillingGuardStats {
	count, micros, lastUnix := GatewayBillingShortfallStats()
	stats := BillingGuardStats{
		SettlementShortfallCount:    count,
		SettlementShortfallMicros:   micros,
		SettlementShortfallLastUnix: lastUnix,

		RejectMarkerActive:        billingRejectMarkerActiveTotal.Load(),
		RejectBelowReserve:        billingRejectBelowReserveTotal.Load(),
		RejectDBTruthBelowReserve: billingRejectDBTruthBelowReserveTotal.Load(),
		RejectWorstCaseCacheOnly:  billingRejectWorstCaseCacheTotal.Load(),
		RejectWorstCaseDBTruth:    billingRejectWorstCaseDBTruthTotal.Load(),
		RejectInflightReservation: billingRejectReservationGuardTotal.Load(),
		RejectWorstCaseLastUnix:   billingRejectWorstCaseDBTruthLastUnix.Load(),
		RejectReservationLastUnix: billingRejectReservationLastUnix.Load(),

		ReservationReserved:           billingReservationReservedTotal.Load(),
		ReservationReleased:           billingReservationReleasedTotal.Load(),
		ReservationRejected:           billingReservationRejectedTotal.Load(),
		ReservationFailOpen:           billingReservationFailOpenTotal.Load(),
		ReservationFailOpenLastUnix:   billingReservationFailOpenLastUnix.Load(),
		ReservationReleaseErr:         billingReservationReleaseErrTotal.Load(),
		ReservationReleaseErrLastUnix: billingReservationReleaseErrLastUnix.Load(),
		ReservationAbandoned:          billingReservationAbandonedTotal.Load(),
		ReservationAbandonedLastUnix:  billingReservationAbandonedLastUnix.Load(),
		ReservationExpiredRelease:     billingReservationExpiredReleaseTotal.Load(),
		ReservationRenew:              billingReservationRenewTotal.Load(),
		ReservationRenewError:         billingReservationRenewErrTotal.Load(),

		RecheckDBReads:              billingRecheckDBReadsTotal.Load(),
		RecheckFailClosed:           billingRecheckFailClosedTotal.Load(),
		RecheckSkippedNoUserRepo:    billingRecheckSkippedNoUserRepoTotal.Load(),
		PrecheckDisabled:            billingPrecheckDisabledTotal.Load(),
		PrecheckUnavailable:         billingPrecheckUnavailableTotal.Load(),
		SubscriptionReservation:     billingSubscriptionReservationTotal.Load(),
		SubscriptionReserveFailOpen: billingSubscriptionReserveFailOpenTotal.Load(),
	}

	// 护栏失效类信号：非 0 即说明防线在这些条件下没有生效，需要排查。
	if stats.SettlementShortfallCount > 0 {
		stats.DegradedSignals = append(stats.DegradedSignals,
			"settlement write-off 仍在发生：预检口径未能覆盖全部请求（检查 request_spend_min_output_tokens / default_max_output_tokens）")
	}
	if stats.ReservationFailOpen > 0 {
		stats.DegradedSignals = append(stats.DegradedSignals,
			"在途预留 fail-open：Redis 预留失败时并发护栏静默失效（检查 Redis 可用性）")
	}
	if stats.ReservationReleaseErr > 0 {
		stats.DegradedSignals = append(stats.DegradedSignals,
			"预留归还失败：额度会滞留到 TTL 到期（检查 Redis 写入与网络）")
	}
	if stats.ReservationRenewError > 0 {
		stats.DegradedSignals = append(stats.DegradedSignals,
			"预留续期失败：超长请求可能在结算前失去护栏保护（检查 Redis 写入与网络）")
	}
	if stats.ReservationExpiredRelease > 0 {
		stats.DegradedSignals = append(stats.DegradedSignals,
			"存在凭据已过期的归还：说明有请求在预留 TTL 内未完成结算，或结算任务丢失")
	}
	if stats.ReservationAbandoned > 0 {
		stats.DegradedSignals = append(stats.DegradedSignals,
			"结算任务被丢弃：该笔未扣费（usage_record 池 drop/sample 溢出，检查池容量）")
	}
	if stats.RecheckSkippedNoUserRepo > 0 {
		stats.DegradedSignals = append(stats.DegradedSignals,
			"预检无法读 DB 真值（userRepo 未装配）：降级模式坏账窗口回到修复前")
	}
	if stats.RecheckFailClosed > 0 {
		stats.DegradedSignals = append(stats.DegradedSignals,
			"复核失败触发 fail-closed：用户看到 503（通常伴随连接池打满）")
	}
	if stats.PrecheckDisabled > 0 {
		stats.DegradedSignals = append(stats.DegradedSignals,
			"最坏费用预检被显式关闭（request_spend_precheck_disabled=true）")
	}
	if stats.PrecheckUnavailable > 0 {
		stats.DegradedSignals = append(stats.DegradedSignals,
			"部分请求无法给出最坏费用上界（无定价/依赖缺失），闸门未生效")
	}
	stats.Healthy = len(stats.DegradedSignals) == 0

	// 误伤比例：最坏费用/预留类拒绝占全部预检拒绝的比例。持续偏高说明预检过保守。
	worstCase := stats.RejectWorstCaseCacheOnly + stats.RejectWorstCaseDBTruth + stats.RejectInflightReservation
	total := worstCase + stats.RejectMarkerActive + stats.RejectBelowReserve + stats.RejectDBTruthBelowReserve
	if total > 0 {
		stats.WorstCaseRejectRatio = float64(worstCase) / float64(total)
	}
	return stats
}

// 以下为各记录点使用的导出记录器。集中在一处，避免各调用点散落直接操纵原子量。

// RecordBillingReservationReserved 记录一次成功的在途预留。
func RecordBillingReservationReserved() { billingReservationReservedTotal.Add(1) }

// RecordBillingReservationRejected 记录一次"预留累加后击穿封底"的拒绝（并发护栏生效）。
func RecordBillingReservationRejected() { billingReservationRejectedTotal.Add(1) }

// RecordBillingReservationFailOpen 记录一次预留 fail-open（Redis 故障时护栏失效）。
func RecordBillingReservationFailOpen() {
	billingReservationFailOpenTotal.Add(1)
	billingReservationFailOpenLastUnix.Store(time.Now().Unix())
}

// RecordBillingReservationReleased 记录一次成功的预留归还。
func RecordBillingReservationReleased() { billingReservationReleasedTotal.Add(1) }

// RecordBillingReservationReleaseError 记录一次归还失败（额度将滞留到 TTL 到期）。
func RecordBillingReservationReleaseError() {
	billingReservationReleaseErrTotal.Add(1)
	billingReservationReleaseErrLastUnix.Store(time.Now().Unix())
}

// RecordBillingReservationExpiredRelease 记录一次"凭据已过期、归还被安全忽略"。
//
// 这不是故障，而是预留 TTL 自愈按预期工作的证据：健康部署下它只会因
// "结算任务丢失"或"心跳未覆盖的超长请求"而出现少量增长。
func RecordBillingReservationExpiredRelease() { billingReservationExpiredReleaseTotal.Add(1) }

// RecordBillingReservationRenew 记录一次成功的预留续期（长请求心跳）。
func RecordBillingReservationRenew() { billingReservationRenewTotal.Add(1) }

// RecordBillingReservationRenewError 记录一次续期失败（非"凭据已过期"）。
func RecordBillingReservationRenewError() { billingReservationRenewErrTotal.Add(1) }

// RecordBillingRecheckDBRead 记录一次预检的 DB 真值回源。
func RecordBillingRecheckDBRead() { billingRecheckDBReadsTotal.Add(1) }

// RecordBillingRecheckFailClosed 记录一次复核失败导致的 fail-closed（用户可见 503）。
func RecordBillingRecheckFailClosed() { billingRecheckFailClosedTotal.Add(1) }

// RecordBillingRecheckSkippedNoUserRepo 记录一次"因缺 userRepo 而无法回源复核"。
//
// 出现即代表部署处于降级装配：预检只能依据可能偏高的缓存值，
// 坏账窗口回到本次修复之前。
func RecordBillingRecheckSkippedNoUserRepo() { billingRecheckSkippedNoUserRepoTotal.Add(1) }

// RecordBillingPrecheckDisabled 记录一次"最坏费用预检被配置显式关闭"。
func RecordBillingPrecheckDisabled() { billingPrecheckDisabledTotal.Add(1) }

// RecordBillingPrecheckUnavailable 记录一次"无法给出最坏费用上界"（闸门未生效）。
func RecordBillingPrecheckUnavailable() { billingPrecheckUnavailableTotal.Add(1) }

// RecordBillingSubscriptionReservation 记录一次订阅模式占位预留。
func RecordBillingSubscriptionReservation() { billingSubscriptionReservationTotal.Add(1) }

// RecordBillingSubscriptionReserveFailOpen 记录一次订阅模式预留 fail-open。
func RecordBillingSubscriptionReserveFailOpen() { billingSubscriptionReserveFailOpenTotal.Add(1) }

// BillingGuardRuntimeInfo 返回护栏的静态运行参数。
//
// 这些值是"护栏的行为契约"（预留 TTL、心跳间隔、保守折算常量），把它们和计数一起
// 暴露出去，ops 才能在不下代码的情况下判断某个斜率是配置问题还是实现问题：
// 例如 write-off 上升而 min_output_tokens 仍是 0，就说明该把下限打开了。
func BillingGuardRuntimeInfo() map[string]any {
	return map[string]any{
		"reservation_ttl_seconds":       int64(billingReservationTTL / time.Second),
		"reservation_heartbeat_seconds": int64(billingReservationHeartbeatInterval / time.Second),
		"release_timeout_seconds":       int64(billingReservationReleaseTimeout / time.Second),
		"default_max_output_tokens":     requestSpendFallbackMaxOutputTokens,
		"fallback_safety_multiplier":    requestSpendFallbackSafetyMultiplier,
		"input_overhead_tokens":         requestSpendInputOverheadTokens,
		"min_inline_binary_run":         requestSpendMinInlineBinaryRun,
		"image_token_allowance":         requestSpendImageTokenAllowance,
		"reject_reasons":                AllBillingPreflightRejectReasons(),
	}
}

package service

import (
	"context"
	"math"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/tidwall/gjson"
)

// 本文件实现「放行前最坏费用预估」闸门（零超发保证的第二道防线）。
//
// 背景：网关计费是后付费——请求先转发到上游（成本已经发生），结算时才扣钱。
// 当余额逼近封底线（billing.minimum_balance_reserve）时，"最后一笔付不满的
// 请求"会先被服务、再在结算时只能扣到封底线为止，差额写成坏账：表现为
// usage_log.actual_cost 只收走一部分、余额停在封底线不再下降，而请求成功返回。
//
// 本闸门在**转发前**计算本次请求的"最坏费用上界"：
//   - 输入 token 上界：按请求体字节数保守折算（EstimateRequestInputTokensUpperBound）；
//   - 输出 token 上界：请求声明的 max_tokens 类字段，缺省用
//     billing.request_spend_default_max_output_tokens（默认 8192）；
//   - 费用与倍率：与结算完全同源（Resolver.Resolve + CalculateTokenCostForRequest
//     + 结算倍率链，含用户专属/分组倍率与高峰因子），再乘
//     billing.request_spend_safety_multiplier（默认 1.0）放大保守程度。
//
// 余额模式预检要求「余额 >= 封底线 + 最坏费用上界」，不满足时直接
// 403 INSUFFICIENT_BALANCE，绝不把付不满的请求转发到上游——"最后一笔"不再产生
// 坏账。余额充足的用户不受任何影响；余额贴近封底的用户，能付起最坏费用的小额
// 请求照常放行并按实际用量精确扣费。
//
// 任一环节无法给出有效上界（功能关闭、依赖缺失、无法取价）时返回 0，本闸门
// 不生效，回退到既有的封底线结算（fail-open to existing guard）。

const (
	// requestSpendFallbackMaxOutputTokens 请求未声明输出上限时使用的缺省输出上界（token）。
	requestSpendFallbackMaxOutputTokens = 8192
	// requestSpendFallbackSafetyMultiplier 预检安全系数缺省值（1.0 = 不放大）。
	requestSpendFallbackSafetyMultiplier = 1.0
	// requestSpendInputOverheadTokens 输入上界的固定附加量：覆盖请求体之外的
	// 计费输入（网关模板、头部折算等）的保守附加。
	requestSpendInputOverheadTokens = 16
)

// requestSpendPrecheckEnabled 返回最坏费用预检是否启用。
// 反向配置（request_spend_precheck_disabled）保证零值安全：未加载 viper 的
// 手工装配（测试/工具）默认启用，与生产行为一致。
func requestSpendPrecheckEnabled(cfg *config.Config) bool {
	if cfg == nil {
		return true
	}
	return !cfg.Billing.RequestSpendPrecheckDisabled
}

// requestSpendDefaultMaxOutputTokens 返回缺省输出上界（token）；未配置时回退内置缺省。
func requestSpendDefaultMaxOutputTokens(cfg *config.Config) int {
	if cfg == nil || cfg.Billing.RequestSpendDefaultMaxOutputTokens <= 0 {
		return requestSpendFallbackMaxOutputTokens
	}
	return cfg.Billing.RequestSpendDefaultMaxOutputTokens
}

// requestSpendSafetyMultiplier 返回预检安全系数；未配置或非法时回退 1.0。
func requestSpendSafetyMultiplier(cfg *config.Config) float64 {
	if cfg == nil || cfg.Billing.RequestSpendSafetyMultiplier <= 0 {
		return requestSpendFallbackSafetyMultiplier
	}
	return cfg.Billing.RequestSpendSafetyMultiplier
}

// ExtractRequestMaxOutputTokens 提取请求体声明的输出 token 上限：
// max_completion_tokens / max_tokens / max_output_tokens 中的最大正值。
// ok=false 表示请求没有声明任何输出上限（调用方应使用配置缺省值）。
func ExtractRequestMaxOutputTokens(body []byte) (int, bool) {
	best := 0
	for _, field := range [...]string{"max_completion_tokens", "max_tokens", "max_output_tokens"} {
		result := gjson.GetBytes(body, field)
		if !result.Exists() || result.Type != gjson.Number {
			continue
		}
		if value := int(result.Int()); value > best {
			best = value
		}
	}
	if best <= 0 {
		return 0, false
	}
	return best, true
}

// EstimateRequestInputTokensUpperBound 把请求体字节数折算为输入 token 上界。
//
// 保守口径：对 UTF-8 文本，BPE 分词每个 token 至少消费 2 字节，因此 len(body)/2
// 是常用的经验上界；再加固定附加量覆盖请求体之外/网关侧注入的计费输入。
// 宁可高估（多拦一笔付不满最坏费用的请求）也不低估（低估即重新产生坏账）。
// 非典型内容（如大量短 token 的特殊编码）可由
// billing.request_spend_safety_multiplier 进一步放大兜底。
func EstimateRequestInputTokensUpperBound(body []byte) int {
	if len(body) <= 0 {
		return requestSpendInputOverheadTokens
	}
	return len(body)/2 + requestSpendInputOverheadTokens
}

// estimateRequestSpendUpperBound 组装最坏费用上界（USD）。返回 0 表示本闸门
// 不生效（功能关闭、依赖缺失或无法取价），调用方保持既有行为。
func estimateRequestSpendUpperBound(
	ctx context.Context,
	cfg *config.Config,
	billingService *BillingService,
	resolver *ModelPricingResolver,
	resolveBaseMultiplier func(context.Context, *User, *APIKey) float64,
	user *User,
	apiKey *APIKey,
	model string,
	body []byte,
) float64 {
	if !requestSpendPrecheckEnabled(cfg) {
		return 0
	}
	if billingService == nil || resolveBaseMultiplier == nil || user == nil || apiKey == nil || apiKey.Group == nil {
		return 0
	}
	model = strings.TrimSpace(model)
	if model == "" {
		return 0
	}

	inputUpper := EstimateRequestInputTokensUpperBound(body)
	outputUpper, declared := ExtractRequestMaxOutputTokens(body)
	if !declared {
		outputUpper = requestSpendDefaultMaxOutputTokens(cfg)
	}
	if inputUpper <= 0 && outputUpper <= 0 {
		return 0
	}

	// 定价时刻优先取请求级冻结点（与结算同刻），未装配时回退当前时刻。
	at := GatewayTokenRequestPricingAtFromContext(ctx)
	if at.IsZero() {
		at = OpenAIPricingAtFromContext(ctx)
	}
	if at.IsZero() {
		at = timezone.Now()
	}

	baseMultiplier := resolveBaseMultiplier(ctx, user, apiKey)
	multiplier, _ := computePeakAwareMultipliers(apiKey, baseMultiplier, at)

	var resolved *ResolvedPricing
	if resolver != nil {
		gid := apiKey.Group.ID
		resolved = resolver.Resolve(ctx, PricingInput{Model: model, GroupID: &gid, Group: apiKey.Group})
	}

	cost, err := billingService.CalculateTokenCostForRequest(TokenCostRequest{
		Ctx:            ctx,
		Model:          model,
		Group:          apiKey.Group,
		Tokens:         UsageTokens{InputTokens: inputUpper, OutputTokens: outputUpper},
		RateMultiplier: multiplier,
		PricingAt:      at,
		Resolver:       resolver,
		Resolved:       resolved,
	})
	if err != nil || cost == nil {
		return 0
	}
	spend := cost.ActualCost * requestSpendSafetyMultiplier(cfg)
	if spend <= 0 || math.IsNaN(spend) || math.IsInf(spend, 0) {
		return 0
	}
	return spend
}

// EstimateRequestSpendUpperBound 预判本次请求结算时的"最坏费用上界"（USD）。
//
// 与结算同源：同一条定价解析链（Resolver.Resolve → CalculateTokenCostForRequest）
// 与倍率链（系统默认 → 用户专属/分组倍率 → 高峰因子），保证预检数字与实际
// 扣费口径一致。返回 0 表示无法给出有效上界（功能关闭/依赖缺失/无定价），
// 调用方应保持既有行为（由封底线结算兜底）。
func (s *GatewayService) EstimateRequestSpendUpperBound(ctx context.Context, user *User, apiKey *APIKey, model string, body []byte) float64 {
	if s == nil {
		return 0
	}
	return estimateRequestSpendUpperBound(
		ctx, s.cfg, s.billingService, s.resolver,
		func(ctx context.Context, user *User, apiKey *APIKey) float64 {
			return s.resolveRequestSpendBaseMultiplier(ctx, user, apiKey)
		},
		user, apiKey, model, body,
	)
}

// resolveRequestSpendBaseMultiplier 复刻 recordUsageCore 的 token 基础倍率解析：
// 系统默认倍率 → 用户专属/分组倍率（不含高峰因子，高峰由共享组装函数叠加）。
func (s *GatewayService) resolveRequestSpendBaseMultiplier(ctx context.Context, user *User, apiKey *APIKey) float64 {
	multiplier := 1.0
	if s.cfg != nil {
		multiplier = s.cfg.Default.RateMultiplier
	}
	if user != nil && apiKey != nil && apiKey.GroupID != nil && apiKey.Group != nil {
		multiplier = s.ResolveUserGroupRateMultiplier(ctx, user.ID, *apiKey.GroupID, apiKey.Group.RateMultiplier)
	}
	return multiplier
}

// EstimateRequestSpendUpperBound 见 GatewayService 同名方法：OpenAI 网关侧的
// 最坏费用上界预判，倍率链与 openai_gateway_usage.RecordUsage 保持一致。
func (s *OpenAIGatewayService) EstimateRequestSpendUpperBound(ctx context.Context, user *User, apiKey *APIKey, model string, body []byte) float64 {
	if s == nil {
		return 0
	}
	return estimateRequestSpendUpperBound(
		ctx, s.cfg, s.billingService, s.resolver,
		func(ctx context.Context, user *User, apiKey *APIKey) float64 {
			return s.resolveRequestSpendBaseMultiplier(ctx, user, apiKey)
		},
		user, apiKey, model, body,
	)
}

// resolveRequestSpendBaseMultiplier 复刻 openai_gateway_usage.RecordUsage 的
// token 基础倍率解析：系统默认倍率 → 用户专属/分组倍率（高峰因子由共享组装函数叠加）。
func (s *OpenAIGatewayService) resolveRequestSpendBaseMultiplier(ctx context.Context, user *User, apiKey *APIKey) float64 {
	multiplier := 1.0
	if s.cfg != nil {
		multiplier = s.cfg.Default.RateMultiplier
	}
	if user != nil && apiKey != nil && apiKey.GroupID != nil && apiKey.Group != nil {
		multiplier = s.ResolveUserGroupRateMultiplier(ctx, user.ID, *apiKey.GroupID, apiKey.Group.RateMultiplier)
	}
	return multiplier
}

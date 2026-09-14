package service

import (
	"context"
	"math"
	"strings"
	"unicode"
	"unicode/utf8"

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
//     配置了 billing.request_spend_min_output_tokens 时取 max(声明值, 下限)，
//     兜住不执行声明上限的上游桥（实测声明 64/190 仍产出 999 token）；
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
	// requestSpendMinInlineBinaryRun 被识别为"内联二进制负载"（base64 图片/文件）
	// 的最短连续串长度。短于它的串仍按文本折算，避免把普通长标识符误判成图片。
	requestSpendMinInlineBinaryRun = 512
	// requestSpendImageTokenAllowance 每块内联图片折算的输入 token 上界。
	//
	// 图片按 patch/分辨率计费，与 base64 字节数无关：一张 4K 图的 base64 可达数 MB，
	// 若按 len(body)/2 折算会得出几十万 token 的天文上界，让"余额充足但贴底"的用户
	// 在带图请求上被误 403。这里改用固定上限（取高分辨率场景的保守值）替代字节折算。
	requestSpendImageTokenAllowance = 1600
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

// requestSpendMinOutputTokens 返回预检输出上界下限（token）；未配置（0）或非法时
// 返回 0 = 不启用钳制。0 值安全：手工装配的 Config 保持“信任声明值”的既有行为。
func requestSpendMinOutputTokens(cfg *config.Config) int {
	if cfg == nil || cfg.Billing.RequestSpendMinOutputTokens <= 0 {
		return 0
	}
	return cfg.Billing.RequestSpendMinOutputTokens
}

// requestSpendCJKRunesPerTokenValue 返回 CJK 折算率（token/字）：未配置（0）或非法时
// 回退内置最坏值 requestSpendCJKRunesPerToken（=1）。
//
// 1 token/字 是按生产上游（deepseek-v4-flash：观测 0.5–1 token/字）校准的上界，对
// 使用**字节回退**分词器的模型（cl100k/o200k 生僻汉字、Llama 系小 CJK 词表）并不成立：
// 那些分词器对非常用汉字可达 2–3 token/字。路由到这类上游的部署应把
// billing.request_spend_cjk_tokens_per_rune 调到 2（或 3），代价是 CJK 大包并发准入变紧。
func requestSpendCJKRunesPerTokenValue(cfg *config.Config) int {
	if cfg == nil || cfg.Billing.RequestSpendCJKTokensPerRune <= 0 {
		return requestSpendCJKRunesPerToken
	}
	return cfg.Billing.RequestSpendCJKTokensPerRune
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
// 四段口径（生产实测校准，166.1.232.118 2026-09-13/14）：
//
//   - **非 CJK 普通文本**：len/2（BPE 每 token ≥2 字节的经验上界）。这是"最坏
//     类别"的兜底费率，覆盖英文短词、数字/符号/JSON 键名密集文本；实测英文
//     自然文本约 4 字节/token，2 字节/token 已留 2 倍余量。
//   - **CJK 文本**（中日韩表意文字及其标点/全角符号）：按 **1 token/字** 折算
//     （requestSpendCJKRunesPerToken，可被 billing.request_spend_cjk_tokens_per_rune
//     调高），即最坏字节/token 下界 = 3（常用 CJK 在 UTF-8 下 3 字节/字；扩展区
//     4 字节/字仍按 1 token/字，更保守）。生产实测
//     上游对中文约 0.5 token/字（≈6 字节/token），旧口径一刀切 2 字节/token
//     = 1.5 token/字：同一条真实 90124 token 的请求被估成 ≈270000 token（约 3 倍
//     高估），strict 模式下把 9 万 token 大包并发压到 2 笔。改用 1 token/字 上界后
//     同一请求降到约 18 万 token（≈2 倍实测，仍是上界）。
//     **注意**：1 token/字 是按上述上游校准的值，对字节回退型分词器（非常用汉字
//     2–3 token/字）不是上界，路由到这类模型时应调高该配置。
//   - **多模态块内的 base64**（`url`/`image_url`/`data`/`file_data` 字段承载的
//     data URI 或 `"type":"base64"` 块）：上游按图片/音频 patch 计费、与字节数无关
//     → 每块按 requestSpendImageTokenAllowance 固定折算。真实图片若按字节折算会得到
//     "一张小图 = 几十万 token"的荒谬上界，使贴底用户的带图请求被误 403；
//   - **文本内容里的 base64**（不处于多模态块内）：上游会把它当**文本**分词，
//     实测 20KB base64 ≈ 18907 token ≈ 0.92 token/字节 → 按 1 token/字节 计
//     （requestSpendDenseTokensPerByte）。若也按 1600/块折算会低估 4 倍，
//     生产实测产生过一笔 $0.00335 的 write-off（cost=0.00835208 collected=0.005）。
//     该识别不依赖 `;base64,` / `"data":"` 标记：没有任何标记的裸 base64 长串
//     同样按稠密费率计（否则仍按 2 字节/token 低估约 1.84 倍）。
//
// 保守性保证（不得低估）：
//   - CJK 折算仅在"文本语言内容几乎全为 CJK"时启用——按**非结构字节**的 CJK 占比
//     ≥ requestSpendCJKCoveragePercent 判定。中英混杂、其它语种、或含大段非 CJK
//     内容的文本一律退回 2 字节/token 的保守口径（JSON 键名/括号/引号等结构字节
//     不计入占比，避免正常 JSON 包裹把纯中文内容误判为混杂）。
//   - CJK 分支内非 CJK 字节仍按 2 字节/token 计，不会因整体分类而放松对
//     符号/数字密集字节的估计。
//   - 文本内 base64 的 1 token/字节与多模态块固定折算两条路径不受影响。
//   - 无法回溯出 JSON 键名、但字符集足够丰富（高熵）的长 base64 串一律按稠密费率
//     （1 token/字节）计，绝不因为"看起来像图片"而少估；低熵长串（重复模式、十六进制）
//     留在普通文本口径 —— 上游 BPE 对这类内容压缩良好，按 1 token/字节 反而高估。
//
// 宁可高估（多拦一笔付不满最坏费用的请求）也不低估（低估即重新产生坏账）。
// 非典型内容可由 billing.request_spend_safety_multiplier 进一步放大兜底。
func EstimateRequestInputTokensUpperBound(body []byte) int {
	return estimateRequestInputTokensUpperBound(body, requestSpendCJKRunesPerToken)
}

// estimateRequestInputTokensUpperBound 是带 CJK 折算率的内部实现：cjkTokensPerRune
// 由配置解析（见 requestSpendCJKRunesPerTokenValue），<=0 时取内置默认。
func estimateRequestInputTokensUpperBound(body []byte, cjkTokensPerRune int) int {
	if len(body) <= 0 {
		return requestSpendInputOverheadTokens
	}
	if cjkTokensPerRune <= 0 {
		cjkTokensPerRune = requestSpendCJKRunesPerToken
	}
	binaryBytes, denseBytes, blobs := inlineBinaryPayloadStats(body)
	textBytes := len(body) - binaryBytes - denseBytes
	if textBytes < 0 {
		textBytes = 0
	}
	return estimateTextInputTokensUpperBound(body, textBytes, denseBytes, binaryBytes, cjkTokensPerRune) +
		denseBytes*requestSpendDenseTokensPerByte +
		blobs*requestSpendImageTokenAllowance +
		requestSpendInputOverheadTokens
}

// estimateTextInputTokensUpperBound 给出"普通文本"部分的输入 token 上界。
//
// CJK 占比判定与折算见 EstimateRequestInputTokensUpperBound 的注释：
//   - CJK 主导：cjkRunes×cjkTokensPerRune + 非 CJKBytes/2；
//   - 否则：textBytes/2（保守口径，含混杂文本）。
func estimateTextInputTokensUpperBound(body []byte, textBytes, denseBytes, binaryBytes, cjkTokensPerRune int) int {
	if textBytes <= 0 {
		return 0
	}
	if cjkTokensPerRune <= 0 {
		cjkTokensPerRune = requestSpendCJKRunesPerToken
	}
	cjkRunes, cjkBytes, structuralBytes := countCJKText(body)
	// 语言内容字节 = 全部字节 - JSON 结构字节 - 文本内 base64 - 多模态负载。
	// 只用它做占比判定；固定折算区不计入分母，避免被 base64/图片稀释。
	languageBytes := len(body) - structuralBytes - denseBytes - binaryBytes
	if languageBytes < 0 {
		languageBytes = 0
	}
	if cjkBytes > 0 && cjkBytes*100 >= languageBytes*requestSpendCJKCoveragePercent {
		nonCJKBytes := textBytes - cjkBytes
		if nonCJKBytes < 0 {
			nonCJKBytes = 0
		}
		return cjkRunes*cjkTokensPerRune +
			nonCJKBytes/requestSpendTextBytesPerToken
	}
	return textBytes / requestSpendTextBytesPerToken
}

// countCJKText 统计 body 中的 CJK 字符数、CJK 字节数与 JSON 结构字节数。
// 只对 ≥0x80 的字节做 UTF-8 解码，纯 ASCII 请求体走快路径。
func countCJKText(body []byte) (cjkRunes, cjkBytes, structuralBytes int) {
	for i := 0; i < len(body); {
		b := body[i]
		if b < utf8.RuneSelf {
			if isJSONStructuralByte(b) {
				structuralBytes++
			}
			i++
			continue
		}
		r, size := utf8.DecodeRune(body[i:])
		if r != utf8.RuneError && isCJKTextRune(r) {
			cjkRunes++
			cjkBytes += size
		}
		i += size
	}
	return cjkRunes, cjkBytes, structuralBytes
}

// isJSONStructuralByte 判断字节是否为 JSON 结构字符。这些字节（空白、括号、冒号、
// 逗号、引号、反斜杠）不承载"语言内容"，在 CJK 占比判定中不计入分母——否则
// 一段纯中文内容只要被标准 JSON 包裹就会被拉低到混杂阈值以下，白白退回保守口径。
func isJSONStructuralByte(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\r', '{', '}', '[', ']', ':', ',', '"', '\\':
		return true
	default:
		return false
	}
}

// isCJKTextRune 判断 rune 是否属于"最坏不过 1 token/字"的 CJK 系文字：中文/日文
// 汉字（unicode.Han）、注音（Bopomofo），以及 CJK 标点/全角符号/兼容表意文字。
// 这些字符在 UTF-8 下为 3 字节（扩展区 4 字节），上游最坏按 1 token/字 计费。
//
// 注意：这里**故意不包含**假名（Hiragana/Katakana）与谚文（Hangul）——它们的
// 最坏分词可能超过 1 token/字（尤其谚文由多个字母组成），放进同一费率会低估；
// 它们连同其它文字落入 2 字节/token 的保守口径。
func isCJKTextRune(r rune) bool {
	if unicode.Is(unicode.Han, r) || unicode.Is(unicode.Bopomofo, r) {
		return true
	}
	switch {
	case r >= 0x3000 && r <= 0x303F, // CJK 符号与标点（、。「」等）
		r >= 0x31C0 && r <= 0x31EF,   // CJK 笔画
		r >= 0xFE30 && r <= 0xFE4F,   // CJK 兼容形式
		r >= 0xFF01 && r <= 0xFF60,   // 全角 ASCII 与全角标点（不含半角片假名）
		r >= 0xFFE0 && r <= 0xFFE6,   // 全角货币/符号
		r >= 0x20000 && r <= 0x3FFFD: // CJK 扩展 B 及以后（4 字节/字）
		return true
	}
	return false
}

// requestSpendTextBytesPerToken 非 CJK 普通文本的"每 token 字节数"下界
// （保守口径：2 字节/token）。这是所有非 CJK 字节的最坏类别保证。
const requestSpendTextBytesPerToken = 2

// requestSpendCJKRunesPerToken CJK 文本的每字符 token 上界。
// 生产实测上游对中文约 0.5 token/字，观测区间 0.5–1 token/字；取最坏 1 token/字
// 作为上界 —— 常用 CJK 3 字节/字 ⇒ 最坏字节/token 下界 = 3；扩展区 4 字节/字仍按
// 1 token/字计更保守。低于此上界（例如按 6 字节/token 对齐实测均值）会低估。
const requestSpendCJKRunesPerToken = 1

// requestSpendCJKCoveragePercent 判定文本主导语言为 CJK 的字节占比下限（百分比）。
// 仅当非结构语言内容中 CJK 占比 ≥ 90% 时启用 CJK 折算；中英混杂/其它语种回落
// 2 字节/token。取 90 是给混杂文本留足余量：宁可少省一点，不可放宽上界。
const requestSpendCJKCoveragePercent = 90

// requestSpendDenseTokensPerByte 文本内容里 base64 的分词密度（token/字节）。
// 实测：20KB base64 由上游分词为 18907 token ≈ 0.92 token/字节（base64 无自然词边界，
// BPE 压缩极差）。取 1.0 略保守。误按普通文本的 0.5 token/字节会低估 2 倍。
const requestSpendDenseTokensPerByte = 1

// inlineBinaryPayloadStats 统计请求体内联 base64 负载,并按其**所在位置**分类:
//
//   - binary(多模态块内):base64 是 JSON 多模态负载字段(`url`/`image_url`/`data`/
//     `file_data`)的值 —— 上游按图片/音频 patch 计费、与字节数无关 → 每块按固定
//     allowance 折算。覆盖形态:OpenAI chat 的 `image_url.url`、**OpenAI Responses 的
//     字符串形态 `image_url`**、Anthropic `source.data`(配 `"type":"base64"`)、
//     Gemini `inline_data.data`、`input_audio.data`。
//   - dense(文本内容里):base64 出现在其它位置(典型:整段 base64 粘在 text 里)——
//     上游会把它当**文本**分词(实测 ≈0.92 token/字节)→ 按
//     requestSpendDenseTokensPerByte 折算,绝不能按图片固定值折算
//     (生产实测曾因此低估 4 倍、产生一笔 write-off)。
//
// 判定方式(三步,均为纯位置/字面量判定,不依赖"负载标记"是否出现):
//  1. 扫描全部**连续 base64 字母表串**(长度 ≥ requestSpendMinInlineBinaryRun),
//     包括没有 `;base64,` / `"data":"` 标记的裸串 —— 旧实现只认标记,导致"整段
//     裸 base64 粘进 text"完全不被识别、仍按 2 字节/token 低估约 1.84 倍
//     (实测 0.92 token/字节),write-off 通道未关闭;
//  2. 用串起点向前找**最近的 JSON 键名**:键属于 multimodalPayloadKeys 才算多模态块
//     (Gemini 的 `inline_data.data` 就是**无 data: 前缀**的裸 base64,只能靠键名判定),
//     `text`/`content` 等其它键一律走第 3 步;
//  3. 其它键上的串:有 `;base64,` / `"data":"` 标记,或字符集足够丰富
//     (looksLikeBase64Payload)时才按稠密费率计 —— 后者避免把 `abab…`、长十六进制
//     串这类低熵文本按 1 token/字节 高估(上游 BPE 对重复串压缩很好),同时堵住
//     "无标记裸 base64"的低估缺口。
//
// 说明:标记本身不参与二进制/文本的二分 —— data URI 的 `data:<mime>;base64,`
// 前缀会被 jsonKeyBefore 一路回溯跳过,所以 `;base64,` 形态仍能正确定位到外层键名;
// 连续串短于阈值时不单独处理(仍按普通文本折算),避免把普通长标识符误判成图片。
func inlineBinaryPayloadStats(body []byte) (binaryBytes int, denseBytes int, blobs int) {
	for _, run := range findBase64Runs(body, requestSpendMinInlineBinaryRun) {
		payload := body[run.start:run.end]
		if multimodalPayloadKeys[jsonKeyBefore(body, run.start)] {
			binaryBytes += len(payload)
			blobs++
			continue
		}
		if !precededByBase64Marker(body, run.start) && !looksLikeBase64Payload(payload) {
			// 低熵长串(重复模式、十六进制、普通标识符):上游 BPE 能压缩,留普通文本口径。
			continue
		}
		denseBytes += len(payload)
	}
	return binaryBytes, denseBytes, blobs
}

// base64Markers 是显式的 base64 负载标记(与历史实现保持一致):data URI 的
// `;base64,` 以及 JSON `"data":"…"` / `"data": "…"`。
var base64Markers = [...][]byte{
	[]byte(";base64,"),
	[]byte(`"data":"`),
	[]byte(`"data": "`),
}

// precededByBase64Marker 判断 runStart 之前是否紧邻某个显式 base64 负载标记。
func precededByBase64Marker(body []byte, runStart int) bool {
	for _, marker := range base64Markers {
		if runStart >= len(marker) && string(body[runStart-len(marker):runStart]) == string(marker) {
			return true
		}
	}
	return false
}

// requestSpendBase64DistinctBytes 是判定"看起来确实是 base64 负载"的最小字符种类数。
//
// 依据:真实 base64(随机/压缩过的二进制或文本)在 512+ 字节的串上几乎必然用到 55+
// 种字符(base64 字母表共 64 种);而 `abab…`(2 种)、十六进制串(≤16 种)、普通标识符
// (通常 < 40 种)都远低于该阈值 —— 这些内容上游 BPE 压缩良好,按 1 token/字节 会高估。
// 取 40 兼顾两侧:既不放过真实 base64(否则重新低估约 1.84 倍),也不误伤低熵长串。
const requestSpendBase64DistinctBytes = 40

// looksLikeBase64Payload 用"字符种类数"近似判断一段 base64 字母表长串是否真的是
// base64 负载(高熵),避免把重复模式/十六进制/长标识符按稠密费率高估。
func looksLikeBase64Payload(run []byte) bool {
	if len(run) < requestSpendMinInlineBinaryRun {
		return false
	}
	var seen [256]bool
	distinct := 0
	for _, c := range run {
		if !seen[c] {
			seen[c] = true
			distinct++
			if distinct >= requestSpendBase64DistinctBytes {
				return true
			}
		}
	}
	return false
}

// base64Run 是一段连续的 base64 字母表字节区间 [start, end)。
type base64Run struct{ start, end int }

// findBase64Runs 返回 body 中所有长度 >= minLen 的连续 base64 字母表串。
// 单趟 O(n);返回的区间互不重叠,因此调用方无需再做去重。
func findBase64Runs(body []byte, minLen int) []base64Run {
	if minLen <= 0 {
		minLen = 1
	}
	var runs []base64Run
	for i := 0; i < len(body); {
		if !isBase64Alphabet(body[i]) {
			i++
			continue
		}
		start := i
		for i < len(body) && isBase64Alphabet(body[i]) {
			i++
		}
		if i-start >= minLen {
			runs = append(runs, base64Run{start: start, end: i})
		}
	}
	return runs
}

// multimodalPayloadKeys 是 JSON 中承载多模态二进制负载的字段名:
//   - OpenAI: `image_url.url`(chat 对象形态)、`image_url`(Responses 字符串形态)、
//     `input_audio.data`、`file_data`
//   - Anthropic: `source.data`(配 `"type":"base64"`)
//   - Gemini: `inline_data.data` / `inlineData.data`
//
// 出现在这些键里的 base64 由上游按 patch 计费(与字节数无关),按固定 allowance 折算;
// 其它位置(典型是 text/content 字段)的 base64 会被上游当文本分词,按稠密费率折算。
var multimodalPayloadKeys = map[string]bool{
	"url": true,
	// image_url：OpenAI Responses 的 `input_image.image_url` 是**字符串**形态的
	// data URI（chat 形态才是 `image_url.url` 对象）。漏掉它会让一张 1MB 内联图
	// 按 1 token/字节 估成 ≈100 万输入 token（实测 1 000 051），把余额只有几美元的
	// 用户误 403、并在严格预留模式下吃光并发额度 —— 而正确口径是 1600 token/块。
	"image_url": true,
	"data":      true,
	"file_data": true,
}

// jsonKeyBefore 返回 valueStart 处 JSON 字符串值所属的**键名**。
//
// valueStart 指向值的第一个字节。从它向前解析(反向)五段结构:
//
//	[URI 前缀(可空)] 值起始引号 [空白] `:` [空白] 键名收尾引号 …… 键名起始引号
//
// 其中 URI 前缀 = `data:<mime>;base64,`(data URI 形态下,负载标记 `;base64,`
// 之前的 `data:` 冒号也是值的一部分)。解析不出该结构时返回 ""(调用方按 dense
// 处理,方向保守:base64 文本按 1 token/字节计,宁多估不少估)。
func jsonKeyBefore(body []byte, valueStart int) string {
	isJSONSpace := func(c byte) bool { return c == ' ' || c == '\t' || c == '\n' || c == '\r' }

	i := valueStart - 1
	// (a) 跳过 data URI 前缀:`data:<mime>;base64,` 整段(含 scheme 名与它的冒号)
	//     都属于值的一部分,要一路回溯到值的起始引号。
	for {
		for i >= 0 && isDataURIChar(body[i]) {
			i--
		}
		if i < 0 {
			return ""
		}
		if body[i] == ':' {
			// `data:` 的 scheme 冒号:它前面还有 scheme 名(URI 字符),继续回溯
			i--
			continue
		}
		break
	}
	// (b) 此处应是值的起始引号
	if i < 0 || body[i] != '"' {
		return ""
	}
	i--
	// (c) 跳过空白,期望键值分隔符 `:`
	for i >= 0 && isJSONSpace(body[i]) {
		i--
	}
	if i < 0 || body[i] != ':' {
		return ""
	}
	i--
	// (d) 跳过空白,期望键名收尾引号
	for i >= 0 && isJSONSpace(body[i]) {
		i--
	}
	if i < 0 || body[i] != '"' {
		return ""
	}
	end := i
	// (e) 回溯键名起始引号(跳过 `\"` 转义)
	i--
	for i >= 0 {
		if body[i] == '"' && (i == 0 || body[i-1] != '\\') {
			return string(body[i+1 : end])
		}
		i--
	}
	return ""
}

// isDataURIChar 判断字节是否属于 data URI 前缀(`data:<mime>;base64,`)的组成字符。
// 注意不含 `:`(它是回溯的终止锚点)与 `"`(值的起始引号)。
func isDataURIChar(c byte) bool {
	switch {
	case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9':
		return true
	case c == ';', c == ',', c == '/', c == '+', c == '-', c == '.':
		return true
	default:
		return false
	}
}

// isBase64Alphabet 判断字节是否属于 base64 字符集（含 padding）。
func isBase64Alphabet(c byte) bool {
	switch {
	case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9':
		return true
	case c == '+', c == '/', c == '=':
		return true
	default:
		return false
	}
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
		RecordBillingPrecheckDisabled()
		return 0
	}
	if billingService == nil || resolveBaseMultiplier == nil || user == nil || apiKey == nil || apiKey.Group == nil {
		RecordBillingPrecheckUnavailable()
		return 0
	}
	model = strings.TrimSpace(model)
	if model == "" {
		RecordBillingPrecheckUnavailable()
		return 0
	}

	inputUpper := estimateRequestInputTokensUpperBound(body, requestSpendCJKRunesPerTokenValue(cfg))
	outputUpper, declared := ExtractRequestMaxOutputTokens(body)
	if !declared {
		outputUpper = requestSpendDefaultMaxOutputTokens(cfg)
	}
	// 声明上限不可信时的下限钳制：上游桥不执行 max_tokens 时（声明 64 实际
	// 999），只按声明值预检会低估最坏费用，钱包贴底时仍会产生 write-off。
	if floor := requestSpendMinOutputTokens(cfg); floor > outputUpper {
		outputUpper = floor
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
		RecordBillingPrecheckUnavailable()
		return 0
	}
	spend := cost.ActualCost * requestSpendSafetyMultiplier(cfg)
	if spend <= 0 || math.IsNaN(spend) || math.IsInf(spend, 0) {
		RecordBillingPrecheckUnavailable()
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

// imageSpendTierCandidates 返回图片预检要试探的尺寸档。
//
// 分组配置了图片单价时，按**全部档位**取最贵的一档作为上界：上游桥可能不执行请求
// 声明的 size，只按声明档预检会低估（与 request_spend_min_output_tokens 的思路一致，
// 方向保守，宁多拦不失守）。未配置单价时只试探声明档，避免在热路径上为每个档位各
// 触发一次"刷新分组媒体定价"的 DB 读。
func imageSpendTierCandidates(apiKey *APIKey, size string) []string {
	declared := NormalizeImageBillingTierOrDefault(size)
	if apiKey != nil && apiKey.Group != nil &&
		(apiKey.Group.ImagePrice1K != nil || apiKey.Group.ImagePrice2K != nil || apiKey.Group.ImagePrice4K != nil) {
		return []string{ImageBillingSize1K, ImageBillingSize2K, ImageBillingSize4K}
	}
	return []string{declared}
}

// EstimateImageRequestSpendUpperBound 预判一次"按次计费"图片生成请求的最坏费用上界（USD）。
//
// 图片是固定单价（尺寸档 × 张数），与 token 无关，因此**不能**复用 token 预估
// （token 预估对图片请求会得出天文数字或 0，两种都错）。口径与结算严格同源：
// 直接复用结算用的 calculateOpenAIImageCost（同一条分组/渠道定价 + 兜底单价链），
// 倍率取结算所用的 imageMultiplier（含高峰因子）。
//
// count <= 0（请求未声明 n）按 1 张计。返回 0 表示无法给出有效上界（功能关闭、
// 依赖缺失或无正价格），调用方保持既有行为。
//
// 已知边界：与 token 预估同理，这里是"声明 + 保守放大"的估计，不覆盖上游自行放大
// 张数的情况；`billing.request_spend_safety_multiplier` 可进一步放大兜底。
func (s *OpenAIGatewayService) EstimateImageRequestSpendUpperBound(
	ctx context.Context,
	user *User,
	apiKey *APIKey,
	model string,
	size string,
	count int,
) float64 {
	if s == nil {
		return 0
	}
	if !requestSpendPrecheckEnabled(s.cfg) {
		RecordBillingPrecheckDisabled()
		return 0
	}
	if s.billingService == nil || user == nil || apiKey == nil || apiKey.Group == nil {
		RecordBillingPrecheckUnavailable()
		return 0
	}
	model = strings.TrimSpace(model)
	if model == "" {
		RecordBillingPrecheckUnavailable()
		return 0
	}
	if count <= 0 {
		count = 1
	}

	at := timezone.Now()
	baseMultiplier := s.resolveRequestSpendBaseMultiplier(ctx, user, apiKey)
	_, imageMultiplier := computePeakAwareMultipliers(apiKey, baseMultiplier, at)

	best := 0.0
	for _, tier := range imageSpendTierCandidates(apiKey, size) {
		breakdown := s.calculateOpenAIImageCost(ctx, model, apiKey, &OpenAIForwardResult{
			ImageCount: count,
			ImageSize:  tier,
		}, imageMultiplier)
		if breakdown == nil {
			continue
		}
		if breakdown.ActualCost > best {
			best = breakdown.ActualCost
		}
	}
	if best <= 0 || math.IsNaN(best) || math.IsInf(best, 0) {
		RecordBillingPrecheckUnavailable()
		return 0
	}
	return best * requestSpendSafetyMultiplier(s.cfg)
}

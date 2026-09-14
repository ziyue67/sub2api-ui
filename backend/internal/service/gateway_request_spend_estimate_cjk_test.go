//go:build unit

package service

import (
	"encoding/base64"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

// requestSpendMeasuredCJKBytesPerToken 生产实测（166.1.232.118 2026-09-14）上游对中文
// 正文的折算：一条真实 prompt_tokens=90124 的请求，正文约 54 万字节 → ≈6 字节/token
// （≈0.5 token/字）。旧口径 2 字节/token（=1.5 token/字）把它高估成 ≈27 万 token（≈3×）。
const requestSpendMeasuredCJKBytesPerToken = 6

func cjkEstimateBody(text string) []byte {
	return []byte(`{"model":"m","messages":[{"role":"user","content":"` + text + `"}]}`)
}

func nonCJKBytesOfBody(text string) int {
	prefix := `{"model":"m","messages":[{"role":"user","content":"`
	suffix := `"}]}`
	return len(prefix) + len(suffix)
}

// TestEstimateRequestInputTokensUpperBound_PureCJK 锁死 CJK 内容感知折算：
//   - 上界不得低于按实测比例（6 字节/token）算出的真实 token 数；
//   - 上界不得低于"1 token/字"这个最坏类别保证；
//   - 相对旧的一刀切 2 字节/token 必须明显下降，且不再有约 3 倍高估。
func TestEstimateRequestInputTokensUpperBound_PureCJK(t *testing.T) {
	text := strings.Repeat("这是一段用于预检上界校准的中文长文本内容", 2000) // 20 字 × 2000
	cjkRunes := utf8.RuneCountInString(text)
	require.Equal(t, 40000, cjkRunes)
	cjkBytes := len(text)
	body := cjkEstimateBody(text)

	est := EstimateRequestInputTokensUpperBound(body)

	realTokens := cjkBytes / requestSpendMeasuredCJKBytesPerToken
	oldEstimate := len(body)/requestSpendTextBytesPerToken + requestSpendInputOverheadTokens

	// 1) 上界：按最坏 1 token/字 + 非 CJK 字节 2 字节/token。
	nonCJKBytes := nonCJKBytesOfBody(text)
	require.Equal(t, cjkRunes+nonCJKBytes/2+requestSpendInputOverheadTokens, est,
		"纯 CJK 文本必须按 1 token/字 折算（非 CJK 包裹字节仍按 2 字节/token 保守计）")

	// 2) 不得低估：高于实测真实 token，且不低于 1 token/字 的最坏保证。
	require.GreaterOrEqual(t, est, realTokens, "上界不得低于实测真实 token 数")
	require.GreaterOrEqual(t, est, cjkRunes, "上界必须覆盖最坏 1 token/字")

	// 3) 不再有约 3 倍高估：旧口径 ≈3×实测，新口径 ≈2×实测。
	require.Less(t, est, realTokens*3, "新口径不得再保留约 3 倍高估")
	require.GreaterOrEqual(t, est, realTokens, "但仍是上界")

	// 4) 相对旧口径确实下降（同一请求旧口径把正文按 2 字节/token 估）。
	require.Less(t, est, oldEstimate, "CJK 折算必须低于旧的一刀切 2 字节/token")
}

// TestEstimateRequestInputTokensUpperBound_CJKCoversWorstCasePerRune 用最坏类别口径
// 直接核对：每个汉字只按 1 token 计，绝不因"实测均值 0.5"而放到 6 字节/token
// （那会低估最坏 1 token/字的场景）。
func TestEstimateRequestInputTokensUpperBound_CJKCoversWorstCasePerRune(t *testing.T) {
	// 词表中不存在的生僻字（扩展 B 区，4 字节/字）也必须按 1 token/字 覆盖。
	text := strings.Repeat("𠀀𠀁𠀂", 100)
	body := cjkEstimateBody(text)
	est := EstimateRequestInputTokensUpperBound(body)
	require.GreaterOrEqual(t, est, 300, "生僻扩展区汉字最坏仍按 1 token/字")
}

// TestEstimateRequestInputTokensUpperBound_PureASCII 锁死 ASCII 不再被 CJK 逻辑影响：
// 纯英文/符号文本仍按 2 字节/token 折算，且高于自然文本的真实 token 数。
func TestEstimateRequestInputTokensUpperBound_PureASCII(t *testing.T) {
	text := strings.Repeat("hello world short words! ", 500)
	body := cjkEstimateBody(text)
	est := EstimateRequestInputTokensUpperBound(body)

	require.Equal(t, len(body)/2+requestSpendInputOverheadTokens, est,
		"纯 ASCII 文本必须保持 2 字节/token 的保守口径")
	// 英文自然文本约 4 字节/token → 2 字节/token 是安全上界。
	require.GreaterOrEqual(t, est, len(text)/4, "上界不得低于英文自然文本的真实 token")
}

// TestEstimateRequestInputTokensUpperBound_MixedRevertsToConservative 锁死混杂文本
// 回落到最保守的 2 字节/token：中英混杂不属于"CJK 主导"，不得启用 1 token/字。
func TestEstimateRequestInputTokensUpperBound_MixedRevertsToConservative(t *testing.T) {
	// 每个 "hello世界" 单元：5 个 ASCII 字节 + 2 个 CJK 字（6 字节）→ CJK 占比 ≈55%。
	text := strings.Repeat("hello世界", 400)
	body := cjkEstimateBody(text)

	est := EstimateRequestInputTokensUpperBound(body)
	require.Equal(t, len(body)/2+requestSpendInputOverheadTokens, est,
		"中英混杂必须回落 2 字节/token 的保守口径")
}

// TestEstimateRequestInputTokensUpperBound_MixedDominantCJKCoverageBoundary 覆盖 90%
// 阈值两侧：CJK 占比略高于阈值时才启用 1 token/字。
func TestEstimateRequestInputTokensUpperBound_MixedDominantCJKCoverageBoundary(t *testing.T) {
	// 长中文正文（4800 字节）+ 少量 ASCII 包裹：CJK 占比 >90% → 启用字级折算。
	dominantBody := cjkEstimateBody(strings.Repeat("中文内容测试用例", 200))
	require.Less(t, EstimateRequestInputTokensUpperBound(dominantBody), len(dominantBody)/2,
		"CJK 占比高于阈值时应启用字级折算")

	// 中英各半（120 字节 CJK + 60 字节 ASCII）→ CJK 占比 <90% → 保守口径。
	weakBody := cjkEstimateBody(strings.Repeat("中文内容", 10) + strings.Repeat("abcdef", 10))
	require.Equal(t, len(weakBody)/2+requestSpendInputOverheadTokens,
		EstimateRequestInputTokensUpperBound(weakBody), "CJK 占比低于阈值时必须保守折算")
}

// TestEstimateRequestInputTokensUpperBound_TextBase64StaysDenseWithCJK 锁死文本内
// base64 的 1 token/字节稠密口径在 CJK 内容旁不受影响（不回归）。
func TestEstimateRequestInputTokensUpperBound_TextBase64StaysDenseWithCJK(t *testing.T) {
	blob := strings.Repeat("iVBORw0KGgoAAAANSUhEUg", 1000) // ≈22KB base64 文本
	cjk := strings.Repeat("中文内容", 500)                     // 2000 字，6000 字节
	body := []byte(`{"model":"m","messages":[{"role":"user","content":[{"type":"text","text":"` +
		cjk + ` data:image/png;base64,` + blob + `"}]}]}`)

	upper := EstimateRequestInputTokensUpperBound(body)
	binaryBytes, denseBytes, _ := inlineBinaryPayloadStats(body)
	require.Equal(t, len(blob), denseBytes, "文本内 base64 必须归入稠密路径")
	require.Zero(t, binaryBytes, "text 字段里的 base64 不是多模态块")
	require.GreaterOrEqual(t, upper, len(blob)+2000,
		"稠密 base64(1 token/字节) 与 CJK(1 token/字) 贡献都必须计入")
	require.Less(t, upper, len(blob)*2, "稠密 base64 不得被放大成 >2 token/字节")
}

// TestEstimateRequestInputTokensUpperBound_ResponsesStringImageURL 锁死 PR#9 回归：
// OpenAI Responses 的 `input_image.image_url` 是**字符串**形态的 data URI，必须与
// chat 的 `image_url.url` 对象形态一样按「多模态块固定 allowance」折算。
//
// 回归前行为：键名 `image_url` 不在白名单 → 归入"文本内 base64" → 1 token/字节，
// 一张 1MB 内联图估出 ≈100 万输入 token（实测 1 000 051），把余额几美元的用户误 403；
// 修复后应为 1600 token/块（+ 少量文本）。
func TestEstimateRequestInputTokensUpperBound_ResponsesStringImageURL(t *testing.T) {
	blob := strings.Repeat("iVBORw0KGgoAAAANSUhEUgAAA", 40000) // 1,000,000 字节
	stringForm := []byte(`{"input":[{"type":"input_image","image_url":"data:image/png;base64,` + blob + `"}]}`)
	objectForm := []byte(`{"messages":[{"content":[{"type":"image_url","image_url":{"url":"data:image/png;base64,` + blob + `"}}]}]}`)

	stringEst := EstimateRequestInputTokensUpperBound(stringForm)
	objectEst := EstimateRequestInputTokensUpperBound(objectForm)

	require.Less(t, stringEst, 20000, "Responses 字符串形态的内联图不得按字节折算（会得到 ≈100 万 token）")
	// 两种形态的 JSON 包裹字节不同（对象形态多一层 `{}`），因此只要求同一量级：
	// 均为「1 块 × 1600 token + 少量文本」，差值远小于 1 块的折算量。
	require.InDelta(t, objectEst, stringEst, 64, "两种 image_url 形态必须同口径（均为 1 块 × 1600 token）")
}

// TestEstimateRequestInputTokensUpperBound_BareBase64InTextIsDense 锁死"裸 base64"缺口：
// **没有任何标记**（无 `data:` URI 前缀、无 `"data":"` 键）的整段 base64 粘在 text 里，
// 上游同样按文本分词（实测 ≈0.92 token/字节）。旧实现只认 `;base64,` / `"data":"`
// 三个字面标记，这类请求会回落 2 字节/token（≈0.5 token/字节），低估约 1.84 倍 →
// 钱包贴底时重新产生 write-off（正是 #9 声称关闭的那条通道）。
func TestEstimateRequestInputTokensUpperBound_BareBase64InTextIsDense(t *testing.T) {
	// 真实 base64（高熵）：随机字节编码后字符种类接近整个字母表。
	raw := make([]byte, 750000)
	for i := range raw {
		raw[i] = byte((i*7919 + i/251) % 256)
	}
	blob := base64.StdEncoding.EncodeToString(raw) // 1,000,000 字节裸 base64
	body := []byte(`{"model":"deepseek-v4-flash","messages":[{"role":"user","content":[{"type":"text","text":"` + blob + `"}]}]}`)

	binaryBytes, denseBytes, blobs := inlineBinaryPayloadStats(body)
	require.Zero(t, binaryBytes, "text 字段里的裸 base64 不是多模态块")
	require.Zero(t, blobs)
	require.Equal(t, len(blob), denseBytes, "无标记的裸 base64 长串必须归入稠密路径")

	upper := EstimateRequestInputTokensUpperBound(body)
	require.GreaterOrEqual(t, upper, len(blob), "必须按 ≥1 token/字节 计（不得再按 2 字节/token 低估约 1.84 倍）")
	require.Less(t, upper, len(blob)*2, "稠密费率不得被放大成 >2 token/字节")
}

// TestEstimateRequestInputTokensUpperBound_LowEntropyRunsStayText 反例保护：低熵的
// base64 字母表长串（重复模式、十六进制串）不得被稠密费率放大 —— 上游 BPE 对这类
// 内容压缩良好，按 1 token/字节 会把 `abab…` 这类文本高估 2 倍。
func TestEstimateRequestInputTokensUpperBound_LowEntropyRunsStayText(t *testing.T) {
	for _, tc := range []struct {
		name string
		text string
	}{
		{"repeated_pair", strings.Repeat("ab", 500)},
		{"repeated_word", strings.Repeat("iVBORw0KGgoAAAANSUhEUg", 500)},
		{"hex_string", strings.Repeat("0123456789abcdef", 64)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := []byte(`{"messages":[{"role":"user","content":"` + tc.text + `"}]}`)
			binaryBytes, denseBytes, _ := inlineBinaryPayloadStats(body)
			require.Zero(t, binaryBytes)
			require.Zero(t, denseBytes, "低熵长串必须留在普通文本口径")
			require.Equal(t, len(body)/2+requestSpendInputOverheadTokens, EstimateRequestInputTokensUpperBound(body))
		})
	}
}

// TestEstimateRequestInputTokensUpperBound_ShortBase64StaysText 反例保护：短于阈值的
// base64 字母表串（普通长标识符/短 hash）仍按普通文本折算，不得被稠密费率放大。
func TestEstimateRequestInputTokensUpperBound_ShortBase64StaysText(t *testing.T) {
	short := strings.Repeat("a", requestSpendMinInlineBinaryRun-1)
	body := []byte(`{"messages":[{"role":"user","content":"` + short + `"}]}`)

	binaryBytes, denseBytes, _ := inlineBinaryPayloadStats(body)
	require.Zero(t, binaryBytes)
	require.Zero(t, denseBytes)
	require.Equal(t, len(body)/2+requestSpendInputOverheadTokens, EstimateRequestInputTokensUpperBound(body))
}

// TestRequestSpendCJKRunesPerTokenValue_Configurable 锁死 F3 的可配置口径：
// 未配置（0）→ 内置 1；配置 2 → CJK 分支按 2 token/字（覆盖字节回退型分词器）。
func TestRequestSpendCJKRunesPerTokenValue_Configurable(t *testing.T) {
	require.Equal(t, 1, requestSpendCJKRunesPerTokenValue(nil))
	require.Equal(t, 1, requestSpendCJKRunesPerTokenValue(&config.Config{}))

	cfg := &config.Config{}
	cfg.Billing.RequestSpendCJKTokensPerRune = 2
	require.Equal(t, 2, requestSpendCJKRunesPerTokenValue(cfg))

	// 负值（配置校验会拒绝，这里再兜一层）回退内置值。
	bad := &config.Config{}
	bad.Billing.RequestSpendCJKTokensPerRune = -3
	require.Equal(t, 1, requestSpendCJKRunesPerTokenValue(bad))

	// 同一请求：rate=2 的输入上界必须严格大于 rate=1（CJK 主导文本）。
	text := strings.Repeat("这是一段用于预检上界校准的中文长文本内容", 2000)
	body := cjkEstimateBody(text)
	one := estimateRequestInputTokensUpperBound(body, 1)
	two := estimateRequestInputTokensUpperBound(body, 2)
	require.Greater(t, two, one, "调高每字 token 率必须抬高上界（保守方向）")
	require.GreaterOrEqual(t, two, 2*utf8.RuneCountInString(text))
}

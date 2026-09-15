//go:build unit

package service

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// 本文件锁死审计 N4：`data` 是极常见的通用键名，仅凭键名就把其中的 base64 当作
// 多模态图片负载（1600 token/块）会让"把文档 base64 后塞进普通 data 字段"的请求
// 被低估数个数量级。多模态的 data 必须由其**父键**上下文确认。

// TestJsonParentKeyBefore 验证外层键名回溯（N4 判定基础）。
func TestJsonParentKeyBefore(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		body string
		pos  int // valueStart（base64 串的第一个字节）
		want string
	}{
		{
			name: "anthropic-source",
			body: `{"source":{"type":"base64","media_type":"image/png","data":"XXXX"`,
			pos:  len(`{"source":{"type":"base64","media_type":"image/png","data":"`),
			want: "source",
		},
		{
			name: "gemini-inline-data",
			body: `{"parts":[{"inline_data":{"mime_type":"image/png","data":"XXXX"`,
			pos:  len(`{"parts":[{"inline_data":{"mime_type":"image/png","data":"`),
			want: "inline_data",
		},
		{
			name: "gemini-inline-data-camel",
			body: `{"inlineData":{"mime_type":"image/png","data":"XXXX"`,
			pos:  len(`{"inlineData":{"mime_type":"image/png","data":"`),
			want: "inlineData",
		},
		{
			name: "openai-input-audio",
			body: `{"input_audio":{"data":"XXXX"`,
			pos:  len(`{"input_audio":{"data":"`),
			want: "input_audio",
		},
		{
			name: "generic-payload",
			body: `{"payload":{"data":"XXXX"`,
			pos:  len(`{"payload":{"data":"`),
			want: "payload",
		},
		{
			name: "top-level-data-has-no-parent",
			body: `{"a":"x","data":"XXXX"`,
			pos:  len(`{"a":"x","data":"`),
			want: "",
		},
		{
			name: "root-object-has-no-parent",
			body: `{"data":"XXXX"`,
			pos:  len(`{"data":"`),
			want: "",
		},
		{
			name: "array-element-has-no-parent",
			body: `[{"data":"XXXX"`,
			pos:  len(`[{"data":"`),
			want: "",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, jsonParentKeyBefore([]byte(tc.body), tc.pos))
		})
	}
}

// TestMultimodalPayloadKeyIsBinary_DataRequiresParentContext 验证 `data` 键的判定：
// 已知的多模态父键 → 二进制；其它已知父键 → 不按二进制（回落稠密/文本口径）；
// 父键无法确证 → 保持既有行为（仍是二进制，避免把真实图片降级而误 403）。
func TestMultimodalPayloadKeyIsBinary_DataRequiresParentContext(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		body string
		pos  int
		want bool
	}{
		{
			name: "source-data-is-binary",
			body: `{"source":{"type":"base64","media_type":"image/png","data":"XXXX"`,
			pos:  len(`{"source":{"type":"base64","media_type":"image/png","data":"`),
			want: true,
		},
		{
			name: "inline-data-is-binary",
			body: `{"parts":[{"inline_data":{"mime_type":"image/png","data":"XXXX"`,
			pos:  len(`{"parts":[{"inline_data":{"mime_type":"image/png","data":"`),
			want: true,
		},
		{
			name: "generic-data-is-not-binary",
			body: `{"payload":{"data":"XXXX"`,
			pos:  len(`{"payload":{"data":"`),
			want: false,
		},
		{
			name: "unverifiable-parent-keeps-legacy",
			body: `{"data":"XXXX"`,
			pos:  len(`{"data":"`),
			want: true,
		},
		{
			name: "image-url-is-unaffected",
			body: `{"m":{"image_url":"XXXX"`,
			pos:  len(`{"m":{"image_url":"`),
			want: true,
		},
		{
			name: "text-key-is-not-multimodal",
			body: `{"m":{"text":"XXXX"`,
			pos:  len(`{"m":{"text":"`),
			want: false,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			body := []byte(tc.body)
			key := jsonKeyBefore(body, tc.pos)
			require.Equal(t, tc.want, multimodalPayloadKeyIsBinary(body, tc.pos, key))
		})
	}
}

// TestEstimateRequestInputTokensUpperBound_DataKeyNeedsParentContext 是 N4 的端到端
// 回归：同一份 1.2MB 高熵 base64，放在普通 data 字段必须按稠密折算（≈1 token/字节），
// 放在已知多模态父键下才按固定 allowance（1 块）折算。
func TestEstimateRequestInputTokensUpperBound_DataKeyNeedsParentContext(t *testing.T) {
	t.Parallel()

	raw := make([]byte, 900000)
	for i := range raw {
		raw[i] = byte(i*31 + 7)
	}
	blob := base64.StdEncoding.EncodeToString(raw) // ≈1.2MB 高熵

	// 普通 data 字段：必须按稠密费率（≈1 token/字节），不得按 1600/块。
	generic := EstimateRequestInputTokensUpperBound([]byte(`{"payload":{"data":"` + blob + `"}}`))
	require.Greater(t, generic, 1_000_000,
		"普通 data 字段里的高熵 base64 必须按稠密折算（审计 N4）")

	// Anthropic source.data：仍是多模态块（1 块 × 1600）。
	anthropic := EstimateRequestInputTokensUpperBound([]byte(
		`{"content":[{"type":"image","source":{"type":"base64","media_type":"image/png","data":"` + blob + `"}}]}`))
	require.Less(t, anthropic, 10_000, "Anthropic source.data 仍是多模态块")

	// Gemini inline_data.data：仍是多模态块。
	gemini := EstimateRequestInputTokensUpperBound([]byte(
		`{"parts":[{"inline_data":{"mime_type":"image/png","data":"` + blob + `"}}]}`))
	require.Less(t, gemini, 10_000, "Gemini inline_data.data 仍是多模态块")

	// 不受 N4 影响的键：image_url（Responses 字符串形态 / chat 对象形态）。
	responses := EstimateRequestInputTokensUpperBound([]byte(
		`{"input":[{"type":"input_image","image_url":"data:image/png;base64,` + blob + `"}]}`))
	require.Less(t, responses, 10_000, "Responses 字符串形态 image_url 仍是多模态块")

	chat := EstimateRequestInputTokensUpperBound([]byte(
		`{"messages":[{"content":[{"type":"image_url","image_url":{"url":"data:image/png;base64,` + blob + `"}}]}]}`))
	require.Less(t, chat, 10_000, "chat 对象形态 image_url.url 仍是多模态块")

	// 回归保护：纯文本仍按 len/2 保守折算，不被误判为稠密。
	plain := EstimateRequestInputTokensUpperBound([]byte(
		`{"model":"gpt-4o","messages":[{"role":"user","content":"` + strings.Repeat("hello world ", 8000) + `"}]}`))
	require.Greater(t, plain, 0)
	require.Less(t, plain, 200_000, "普通英文文本不应被判为稠密")
}

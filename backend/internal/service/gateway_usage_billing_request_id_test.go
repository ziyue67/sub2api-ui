//go:build unit

package service

import (
	"context"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/stretchr/testify/require"
)

func TestResolveUsageBillingRequestID_ForcedWebSearchBeatsClientID(t *testing.T) {
	t.Parallel()
	ctx := context.WithValue(context.Background(), ctxkey.ClientRequestID, "client-shared-id")
	got := resolveUsageBillingRequestID(ctx, "web_search:uuid-1")
	require.Equal(t, "web_search:uuid-1", got)
}

func TestResolveUsageBillingRequestID_IgnoresClientControlledIDs(t *testing.T) {
	// 安全回归（审计 H5）：客户端可控的 X-Client-Request-ID / X-Request-ID 绝不能
	// 决定计费幂等键 —— 固定 header 会让多笔真实调用折叠成一笔（免费调用）。
	t.Parallel()
	ctx := context.WithValue(context.Background(), ctxkey.ClientRequestID, "client-shared-id")
	ctx = context.WithValue(ctx, ctxkey.RequestID, "local-shared-id")

	// 有上游 id：用它
	require.Equal(t, "resp_abc", resolveUsageBillingRequestID(ctx, "resp_abc"))
	// 无上游 id：服务端生成，且与客户端 header 无关
	generated := resolveUsageBillingRequestID(ctx, "")
	require.True(t, strings.HasPrefix(generated, "generated:"))
	require.NotContains(t, generated, "client-shared-id")
	require.NotContains(t, generated, "local-shared-id")
	require.NotEqual(t, generated, resolveUsageBillingRequestID(ctx, ""), "每次生成都必须是新键")
	// 强持久 id 仍然优先（视频/搜索的多次轮询合并成一笔账单）
	require.Equal(t, "grok-video:task-1", resolveUsageBillingRequestID(ctx, "grok-video:task-1"))
}

func TestIsForcedUsageBillingRequestID(t *testing.T) {
	t.Parallel()
	require.True(t, isForcedUsageBillingRequestID("web_search:x"))
	require.True(t, isForcedUsageBillingRequestID("grok-video:task-1"))
	require.True(t, isForcedUsageBillingRequestID("grok_audio:up-1"))
	require.True(t, isForcedUsageBillingRequestID("grok_realtime:sess-1"))
	require.False(t, isForcedUsageBillingRequestID("resp_abc"))
}

func TestStableGrokAudioBillingRequestID(t *testing.T) {
	t.Parallel()
	require.Equal(t, "grok_audio:up-1", StableGrokAudioBillingRequestID("up-1"))
	require.Equal(t, "grok_audio:up-1", StableGrokAudioBillingRequestID("grok_audio:up-1"))
	got := StableGrokAudioBillingRequestID("")
	require.True(t, strings.HasPrefix(got, "grok_audio:"))
	require.Greater(t, len(got), len("grok_audio:"))
}

func TestStableGrokRealtimeBillingRequestID(t *testing.T) {
	t.Parallel()
	require.Equal(t, "grok_realtime:s1", StableGrokRealtimeBillingRequestID("s1"))
	require.Equal(t, "grok_realtime:s1", StableGrokRealtimeBillingRequestID("grok_realtime:s1"))
	got := StableGrokRealtimeBillingRequestID("")
	require.True(t, strings.HasPrefix(got, "grok_realtime:"))
}

func TestResolveUsageBillingRequestID_ForcedGrokAudioBeatsClientID(t *testing.T) {
	t.Parallel()
	ctx := context.WithValue(context.Background(), ctxkey.ClientRequestID, "client-shared-id")
	got := resolveUsageBillingRequestID(ctx, StableGrokAudioBillingRequestID("up-9"))
	require.Equal(t, "grok_audio:up-9", got)
}

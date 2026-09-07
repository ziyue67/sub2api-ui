<template>
  <div v-if="snapshot" class="scheme3-monitor-quota" data-testid="monitor-quota-view">
    <!-- 套餐等级徽章（如智谱 plan level / Claude 订阅档） -->
    <div v-if="snapshot.plan_level" class="scheme3-monitor-quota-plan-row">
      <span class="scheme3-monitor-quota-plan">
        {{ snapshot.plan_level }}
      </span>
    </div>

    <!-- 用量窗口条形图（样式/阈值对齐账号页 CNProviderQuotaCell） -->
    <div v-if="snapshot.success && tierRows.length" class="scheme3-monitor-quota-tiers">
      <UsageProgressBar
        v-for="row in tierRows"
        :key="row.key"
        class="scheme3-monitor-quota-tier scheme3-usage-progress"
        data-testid="monitor-quota-tier"
        :label="row.label"
        :title="row.title"
        label-width="auto"
        :color="row.color"
        scheme3
        :utilization="row.tier.used_percent"
        :resets-at="row.tier.reset_at ?? null"
      />
    </div>

    <!-- 余额（国产 payg；支持多币种） -->
    <div v-if="snapshot.success && balanceRows.length" class="scheme3-monitor-quota-balances">
      <span
        v-for="b in balanceRows"
        :key="b.currency"
        :class="['scheme3-monitor-quota-balance', { 'is-empty': b.balance <= 0 }]"
      >
        {{ b.balance.toFixed(2) }} {{ b.currency }}
      </span>
    </div>

    <div v-if="!snapshot.success" class="scheme3-monitor-quota-error" :title="snapshot.error" data-testid="monitor-quota-error">
      {{ truncatedError }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { MonitorQuotaSnapshot, MonitorQuotaTier } from '@/api/admin/channelMonitor'
import UsageProgressBar from '@/components/account/UsageProgressBar.vue'

/**
 * 配额快照渲染（管理端监控列表/运行结果 + 用户端监控卡片共用）。
 * 展示形态对齐账号管理侧的用量视图（CNProviderQuotaCell：同阈值配色、
 * 同倒计时格式）；tier 的 Window/Label 是后端约定的机器 token，
 * 已知 token 走 i18n，未知 token 原样展示（前向兼容）。
 */
const props = defineProps<{
  snapshot?: MonitorQuotaSnapshot | null
}>()

const { t, te } = useI18n()

interface QuotaTierRow {
  key: string
  label: string
  title: string
  color: 'indigo' | 'emerald' | 'purple' | 'amber'
  tier: MonitorQuotaTier
}

// 已知的 window/label 机器 token → i18n key（monitorCommon.quota.*）。
const windowI18nKeys: Record<string, string> = {
  '5h': 'monitorCommon.quota.windows.5h',
  '7d': 'monitorCommon.quota.windows.7d',
  '7d-sonnet': 'monitorCommon.quota.windows.7dSonnet',
  '7d-fable': 'monitorCommon.quota.windows.7dFable',
  weekly: 'monitorCommon.quota.windows.weekly',
  daily: 'monitorCommon.quota.windows.daily',
  '30d': 'monitorCommon.quota.windows.30d',
  total: 'monitorCommon.quota.windows.total',
}

const labelI18nKeys: Record<string, string> = {
  requests: 'monitorCommon.quota.labels.requests',
  tokens: 'monitorCommon.quota.labels.tokens',
  shared: 'monitorCommon.quota.labels.shared',
  pro: 'monitorCommon.quota.labels.pro',
  flash: 'monitorCommon.quota.labels.flash',
}

function windowLabel(window: string): string {
  const key = windowI18nKeys[window]
  return key && te(key) ? t(key) : window
}

function tierLabel(tier: MonitorQuotaTier): string {
  const window = windowLabel(tier.window)
  if (!tier.label) return window
  const labelKey = labelI18nKeys[tier.label]
  const label = labelKey && te(labelKey) ? t(labelKey) : tier.label
  return `${label}/${window}`
}

const tierColors: QuotaTierRow['color'][] = ['indigo', 'emerald', 'purple', 'amber']

const tierRows = computed<QuotaTierRow[]>(() =>
  (props.snapshot?.tiers || []).map((tier, idx) => ({
    key: `${tier.window}-${tier.label || ''}-${idx}`,
    label: tierLabel(tier),
    title: tierLabel(tier),
    color: tierColors[idx % tierColors.length],
    tier,
  })),
)

const balanceRows = computed(() => {
  const snapshot = props.snapshot
  if (!snapshot) return []
  if (snapshot.balances?.length) return snapshot.balances
  if (snapshot.balance != null) {
    return [{ currency: snapshot.currency || '?', balance: snapshot.balance }]
  }
  return []
})

const truncatedError = computed(() => {
  const error = props.snapshot?.error || t('monitorCommon.quota.unavailable')
  return error.length > 48 ? `${error.slice(0, 48)}…` : error
})

</script>

<style scoped>
.scheme3-monitor-quota {
  display: grid;
  min-width: 0;
  gap: .3rem;
  color: #6b695f;
  font-size: .625rem;
}

.scheme3-monitor-quota-plan-row,
.scheme3-monitor-quota-balances {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
}

.scheme3-monitor-quota-plan-row { gap: .35rem; }
.scheme3-monitor-quota-balances { column-gap: .55rem; row-gap: .15rem; }

.scheme3-monitor-quota-plan {
  display: inline-flex;
  align-items: center;
  border: 1px solid #dad5c8;
  border-radius: 4px;
  background: #f1eee6;
  color: #5f5b50;
  padding: .12rem .38rem;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: .58rem;
  font-weight: 800;
}

.scheme3-monitor-quota-tiers {
  display: grid;
  min-width: 0;
  gap: .28rem;
}

/* UsageProgressBar carries the upstream countdown/threshold contract. These
   selectors only skin its root for Scheme3 and never alter the shared logic. */
.scheme3-monitor-quota-tier :deep(> div) { min-width: 0; }
.scheme3-monitor-quota-tier :deep([class*="bg-indigo-100"]),
.scheme3-monitor-quota-tier :deep([class*="bg-emerald-100"]),
.scheme3-monitor-quota-tier :deep([class*="bg-purple-100"]),
.scheme3-monitor-quota-tier :deep([class*="bg-amber-100"]) {
  border: 1px solid var(--scheme3-line, #dad5c8);
  border-radius: 4px;
  background: var(--scheme3-subtle, #f1eee6);
  color: var(--scheme3-muted, #6b695f);
}
.scheme3-monitor-quota-tier :deep([class*="bg-gray-200"]) {
  height: .375rem;
  width: 4rem;
  background: var(--scheme3-track, #e5e0d5);
}
.scheme3-monitor-quota-tier :deep([class*="bg-red-500"]) { background: var(--scheme3-danger, #9e4d3d); }
.scheme3-monitor-quota-tier :deep([class*="bg-amber-500"]) { background: var(--scheme3-amber, #b7791f); }
.scheme3-monitor-quota-tier :deep([class*="bg-green-500"]) { background: var(--scheme3-accent, #1e5c42); }
.scheme3-monitor-quota-tier :deep([class*="text-red-600"]) { color: var(--scheme3-danger, #9e4d3d); }
.scheme3-monitor-quota-tier :deep([class*="text-amber-600"]) { color: var(--scheme3-amber, #b7791f); }
.scheme3-monitor-quota-tier :deep([class*="text-gray-600"]) { color: var(--scheme3-muted, #6b695f); }

.scheme3-monitor-quota-tier {
  display: block;
  min-width: 0;
}
.scheme3-usage-progress :deep(.flex.items-center.gap-1) {
  display: grid;
  min-width: 0;
  grid-template-columns: minmax(3rem, 3.5rem) 4rem auto minmax(0, 1fr);
  align-items: center;
  gap: .35rem;
}
.scheme3-usage-progress :deep([class*="w-\[32px\]"]) {
  width: auto;
  min-width: 2rem;
  color: inherit;
}
.scheme3-usage-progress :deep([class*="shrink-0 text-\[10px\]"]) {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.scheme3-monitor-quota-label,
.scheme3-monitor-quota-reset {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.scheme3-monitor-quota-label { color: #6b695f; }
.scheme3-monitor-quota-reset { color: #9a9588; }

.scheme3-monitor-quota-track {
  width: 4rem;
  height: .375rem;
  overflow: hidden;
  border-radius: 2px;
  background: #e5e0d5;
}

.scheme3-monitor-quota-fill {
  height: 100%;
  border-radius: 2px;
  transition: width 180ms ease;
}

.scheme3-monitor-quota-value,
.scheme3-monitor-quota-balance { font-weight: 800; }
.scheme3-monitor-quota-balance { color: #5f5b50; }
.scheme3-monitor-quota-fill.is-healthy { background: #1e5c42; }
.scheme3-monitor-quota-fill.is-warning { background: #8b5d14; }
.scheme3-monitor-quota-fill.is-critical { background: #9e4d3d; }
.scheme3-monitor-quota-value.is-healthy { color: #1e5c42; }
.scheme3-monitor-quota-value.is-warning { color: #8b5d14; }
.scheme3-monitor-quota-value.is-critical,
.scheme3-monitor-quota-balance.is-empty,
.scheme3-monitor-quota-error { color: #9e4d3d; }

.scheme3-monitor-quota-error {
  overflow: hidden;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

:global(html.dark .scheme3-monitor-quota) { color: #aaa69a; }
:global(html.dark .scheme3-monitor-quota-plan) {
  border-color: #47443a;
  background: #2b2924;
  color: #c7c2b6;
}
:global(html.dark .scheme3-monitor-quota-label) { color: #aaa69a; }
:global(html.dark .scheme3-monitor-quota-reset) { color: #827e72; }
:global(html.dark .scheme3-monitor-quota-track) { background: #47443a; }
:global(html.dark .scheme3-monitor-quota-balance) { color: #c7c2b6; }
:global(html.dark .scheme3-monitor-quota-fill.is-healthy) { background: #8fc2a5; }
:global(html.dark .scheme3-monitor-quota-fill.is-warning) { background: #d3a45c; }
:global(html.dark .scheme3-monitor-quota-fill.is-critical) { background: #d38b79; }
:global(html.dark .scheme3-monitor-quota-value.is-healthy) { color: #8fc2a5; }
:global(html.dark .scheme3-monitor-quota-value.is-warning) { color: #d3a45c; }
:global(html.dark .scheme3-monitor-quota-value.is-critical),
:global(html.dark .scheme3-monitor-quota-balance.is-empty),
:global(html.dark .scheme3-monitor-quota-error) { color: #d38b79; }
</style>

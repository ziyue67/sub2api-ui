<template>
  <AppLayout>
    <div class="console-shell">
      <header class="console-hero">
        <div class="console-hero-copy">
          <div class="console-eyebrow"><span class="console-live-dot"></span> SHOUR OR TOKEN / 运行控制台</div>
          <div class="console-title-row">
            <h1 class="console-title">控制台总览</h1>
            <span class="console-date">{{ todayLabel }}</span>
          </div>
          <p class="console-subtitle">把请求、余额、密钥与多路由状态放在同一张实时工作台上。</p>
        </div>
        <div class="console-hero-actions">
          <div class="console-sync-state">
            <span class="console-sync-mark"><span class="console-live-dot"></span>实时数据</span>
            <span class="console-sync-time">{{ lastSyncedLabel }}</span>
          </div>
          <button type="button" class="console-refresh" :disabled="refreshing" aria-label="刷新控制台" @click="loadDashboard">
            <Icon name="refresh" size="sm" :class="{ 'animate-spin': refreshing }" />
            <span>{{ refreshing ? '同步中' : '同步数据' }}</span>
          </button>
        </div>
      </header>

      <div v-if="loadError" class="console-alert console-alert-error" role="alert">
        <span class="console-alert-icon"><Icon name="exclamationTriangle" size="sm" /></span>
        <div class="min-w-0 flex-1"><strong>暂时无法读取控制台</strong><p>{{ loadError }}</p></div>
        <button type="button" class="console-alert-action" @click="loadDashboard">重新连接</button>
      </div>
      <div v-else-if="sectionErrors.length" class="console-alert console-alert-warn" role="status">
        <span class="console-alert-icon"><Icon name="infoCircle" size="sm" /></span>
        <div class="min-w-0 flex-1"><strong>部分数据正在等待上游返回</strong><p>{{ sectionErrors.join('、') }}暂不可用，其余模块仍可查看。</p></div>
      </div>

      <div v-if="loading" class="console-loading" aria-live="polite">
        <div class="console-loading-orbit"><LoadingSpinner /></div>
        <strong>正在整理你的调用记录</strong>
        <span>同步请求、路由与额度信息…</span>
      </div>

      <template v-else>
        <section class="console-kpi-grid" aria-label="核心指标">
          <article v-for="(card, index) in kpiCards" :key="card.label" class="console-kpi" :class="`console-kpi-${card.tone}`">
            <div class="console-kpi-head">
              <span class="console-kpi-label">{{ card.label }}</span>
              <span class="console-kpi-icon"><Icon :name="card.icon" size="sm" /></span>
            </div>
            <strong class="console-kpi-value">{{ card.value }}</strong>
            <p class="console-kpi-note">{{ card.note }}</p>
            <span class="console-kpi-index">0{{ index + 1 }}</span>
          </article>
        </section>

        <div class="console-layout">
          <main class="console-main-column">
            <section class="console-panel console-trend-panel" aria-labelledby="trend-title">
              <header class="console-panel-heading">
                <div><span class="console-panel-kicker">调用脉冲</span><h2 id="trend-title">最近七天的调用节奏</h2><p>每个柱体代表一天的请求量，颜色越深表示当天越活跃。</p></div>
                <div class="console-total"><span>七天总请求</span><strong>{{ formatNumber(trendTotalRequests) }}</strong></div>
              </header>
              <div v-if="trendBars.length" class="console-chart" aria-label="最近七天请求趋势">
                <div class="console-chart-grid" aria-hidden="true"><span></span><span></span><span></span><span></span></div>
                <div v-for="bar in trendBars" :key="bar.key" class="console-chart-column">
                  <span class="console-chart-value">{{ formatNumber(bar.requests) }}</span>
                  <div class="console-chart-track"><div class="console-chart-fill" :style="{ height: `${bar.height}%` }"></div></div>
                  <span class="console-chart-label">{{ bar.label }}</span>
                </div>
              </div>
              <div v-else class="console-empty"><Icon name="chartBar" size="md" /><span>这段时间还没有调用记录</span></div>
            </section>

            <section class="console-panel" aria-labelledby="routes-title">
              <header class="console-panel-heading console-panel-heading-tight">
                <div><span class="console-panel-kicker">路由矩阵</span><h2 id="routes-title">多路由路径</h2><p>每条路径都来自当前账号可见的密钥与分组配置。</p></div>
                <router-link to="/keys" class="console-text-link">管理密钥 <Icon name="arrowRight" size="xs" /></router-link>
              </header>
              <div v-if="routeSummary.length" class="console-route-grid">
                <article v-for="(row, index) in routeSummary" :key="row.id" class="console-route-card">
                  <div class="console-route-top"><span class="console-route-number">0{{ index + 1 }}</span><span class="console-route-platform" :class="platformTone(row.platform)">{{ platformLabel(row.platform) }}</span><span class="console-route-state"><i></i>已启用</span></div>
                  <div class="console-route-body"><h3>{{ row.name }}</h3><div class="console-route-flow"><span><Icon name="key" size="xs" />{{ row.activeKeys }} 个密钥</span><Icon name="arrowRight" size="xs" /><span><Icon name="link" size="xs" />{{ row.routes }} 条路径</span></div></div>
                  <div class="console-route-foot"><span>近 7 天 {{ formatNumber(row.requests) }} 次请求</span><strong>{{ formatMultiplier(row.multiplier) }} 倍率</strong></div>
                </article>
              </div>
              <div v-else class="console-empty"><Icon name="link" size="md" /><span>暂未发现可用的路由分组</span></div>
            </section>

            <section class="console-panel" aria-labelledby="recent-title">
              <header class="console-panel-heading console-panel-heading-tight"><div><span class="console-panel-kicker">请求账本</span><h2 id="recent-title">最近请求</h2><p>只展示当前账号可以看到的原始调用记录。</p></div><router-link to="/usage" class="console-text-link">打开明细 <Icon name="arrowRight" size="xs" /></router-link></header>
              <div v-if="recentRequests.length" class="console-table-wrap"><table class="console-table"><thead><tr><th>模型 / 分组</th><th>上游路径</th><th>耗时</th><th class="text-right">时间</th></tr></thead><tbody><tr v-for="request in recentRequests" :key="request.id"><td><strong class="console-table-model" :title="request.model">{{ request.model || '未命名模型' }}</strong><span>{{ request.group?.name || groupName(request.group_id) }}</span></td><td><code :title="request.upstream_endpoint || request.inbound_endpoint || ''">{{ compactEndpoint(request.upstream_endpoint || request.inbound_endpoint) }}</code></td><td class="console-table-mono">{{ formatDuration(request.duration_ms) }}</td><td class="console-table-mono text-right">{{ formatDateTime(request.created_at) }}</td></tr></tbody></table></div>
              <div v-else class="console-empty"><Icon name="inbox" size="md" /><span>还没有可展示的请求记录</span></div>
            </section>
          </main>

          <aside class="console-side-column">
            <section class="console-panel console-side-panel" aria-labelledby="models-title"><header class="console-side-heading"><div><span class="console-panel-kicker">模型分布</span><h2 id="models-title">模型排行</h2></div><Icon name="cube" size="sm" class="console-panel-icon" /></header><div v-if="topModels.length" class="console-model-list"><div v-for="(model, index) in topModels" :key="model.model" class="console-model-row"><span class="console-model-rank">0{{ index + 1 }}</span><div class="min-w-0 flex-1"><div class="console-model-name" :title="model.model">{{ model.model }}</div><div class="console-progress"><span :style="{ width: `${model.ratio}%` }"></span></div></div><strong>{{ formatNumber(model.requests) }}</strong></div></div><div v-else class="console-empty console-empty-small">暂无模型统计</div></section>

            <section class="console-panel console-side-panel" aria-labelledby="upstream-title"><header class="console-side-heading"><div><span class="console-panel-kicker">上游追踪</span><h2 id="upstream-title">上游节点</h2></div><Icon name="server" size="sm" class="console-panel-icon" /></header><div v-if="upstreamNodes.length" class="console-node-list"><div v-for="node in upstreamNodes" :key="node.endpoint" class="console-node-row"><span class="console-node-status"></span><div class="min-w-0 flex-1"><strong :title="node.endpoint">{{ compactEndpoint(node.endpoint) }}</strong><span>{{ node.requests }} 次调用 · 最近 {{ formatDateTime(node.lastSeen) }}</span></div><em>{{ formatDuration(node.avgDuration) }}</em></div><p class="console-side-note">节点状态只按最近日志展示，不主动推断上游健康。</p></div><div v-else class="console-empty console-empty-small">近期日志没有回传上游地址</div></section>

            <section class="console-panel console-side-panel" aria-labelledby="quota-title"><header class="console-side-heading"><div><span class="console-panel-kicker">额度监视</span><h2 id="quota-title">额度观察</h2></div><Icon name="creditCard" size="sm" class="console-panel-icon" /></header><div v-if="quotaRows.length" class="console-quota-list"><div v-for="quota in quotaRows" :key="quota.platform" class="console-quota-row"><div><strong>{{ platformLabel(quota.platform) }}</strong><span>{{ quota.limitLabel }}</span></div><div class="console-progress console-progress-quota"><span :style="{ width: `${quota.ratio}%` }"></span></div><p>已用 {{ formatMoney(quota.usage) }} · {{ quota.windowLabel }}</p></div></div><div v-else class="console-empty console-empty-small">当前没有配置平台额度</div></section>

            <section class="console-panel console-side-panel" aria-labelledby="errors-title"><header class="console-side-heading"><div><span class="console-panel-kicker">异常信号</span><h2 id="errors-title">异常记录</h2></div><span class="console-error-count">{{ errorTotal }}</span></header><div v-if="recentErrors.length" class="console-error-list"><div v-for="error in recentErrors" :key="error.id" class="console-error-row"><span class="console-error-dot"></span><div class="min-w-0 flex-1"><strong>{{ error.model || '未命名模型' }}</strong><span>{{ error.status_code }} · {{ error.category || '请求失败' }}</span></div><time>{{ formatDateTime(error.created_at) }}</time></div></div><div v-else class="console-empty console-empty-small"><Icon name="checkCircle" size="sm" class="console-success-icon" /><span>最近七天没有异常记录</span></div></section>
          </aside>
        </div>

        <footer class="console-footer"><span><i class="console-live-dot"></i>{{ userLabel }} · 仅展示当前账号可见信息</span><span>{{ lastSyncedLabel }}</span></footer>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAuthStore } from '@/stores'
import {
  usageAPI,
  type UserDashboardStats,
  type UsageDashboardSnapshotV2Response,
} from '@/api/usage'
import keysAPI from '@/api/keys'
import userGroupsAPI from '@/api/groups'
import { getMyPlatformQuotas } from '@/api/user'
import type {
  ApiKey,
  Group,
  GroupStat,
  ModelStat,
  PlatformQuotaItem,
  TrendDataPoint,
  UsageLog,
  UserErrorRequest,
} from '@/types'
import { formatDateLocalInput } from '@/utils/format'

type IconName = 'chart' | 'key' | 'dollar' | 'clock'
type Tone = 'green' | 'amber' | 'blue' | 'ink'

interface KpiCard {
  label: string
  value: string
  note: string
  icon: IconName
  tone: Tone
}

interface GroupSummaryRow {
  id: number
  name: string
  platform: string
  multiplier: number
  requests: number
  routes: number
  activeKeys: number
}

interface TrendBar {
  key: string
  label: string
  requests: number
  height: number
}

interface UpstreamNode {
  endpoint: string
  requests: number
  avgDuration: number | null
  lastSeen: string
}

interface QuotaRow extends PlatformQuotaItem {
  ratio: number
  limitLabel: string
  windowLabel: string
  usage: number
}

const authStore = useAuthStore()
const stats = ref<UserDashboardStats | null>(null)
const trendData = ref<TrendDataPoint[]>([])
const modelStats = ref<ModelStat[]>([])
const groupStats = ref<GroupStat[]>([])
const recentUsage = ref<UsageLog[]>([])
const recentErrors = ref<UserErrorRequest[]>([])
const apiKeys = ref<ApiKey[]>([])
const availableGroups = ref<Group[]>([])
const platformQuotas = ref<PlatformQuotaItem[]>([])
const loading = ref(true)
const refreshing = ref(false)
const loadError = ref('')
const sectionErrors = ref<string[]>([])
const lastSyncedAt = ref<Date | null>(null)

const startDate = formatDateLocalInput(new Date(Date.now() - 6 * 86400000))
const endDate = formatDateLocalInput(new Date())

const user = computed(() => authStore.user)
const userLabel = computed(() => user.value?.username || user.value?.email?.split('@')[0] || '当前账号')
const todayLabel = computed(() => new Intl.DateTimeFormat('zh-CN', { year: 'numeric', month: 'long', day: 'numeric' }).format(new Date()))
const lastSyncedLabel = computed(() => (lastSyncedAt.value ? `同步于 ${formatDateTime(lastSyncedAt.value.toISOString())}` : '等待同步'))

const kpiCards = computed<KpiCard[]>(() => {
  const current = stats.value
  const balance = Number(user.value?.balance || 0)
  return [
    {
      label: '今日请求',
      value: formatNumber(current?.today_requests || 0),
      note: `${formatNumber(current?.today_tokens || 0)} 个令牌 · 近 5 分钟 ${formatNumber(current?.rpm || 0)} 请求/分钟`,
      icon: 'chart',
      tone: 'green',
    },
    {
      label: '活跃密钥',
      value: `${formatNumber(current?.active_api_keys ?? apiKeys.value.filter((key) => key.status === 'active').length)} / ${formatNumber(current?.total_api_keys ?? apiKeys.value.length)}`,
      note: '可用密钥 / 全部密钥',
      icon: 'key',
      tone: 'blue',
    },
    {
      label: '可用余额',
      value: formatMoney(balance),
      note: `今日扣除 ${formatMoney(current?.today_actual_cost || 0)}`,
      icon: 'dollar',
      tone: 'amber',
    },
    {
      label: '平均响应',
      value: formatDuration(current?.average_duration_ms ?? null),
      note: `近 5 分钟 ${formatNumber(current?.tpm || 0)} 令牌/分钟`,
      icon: 'clock',
      tone: 'ink',
    },
  ]
})

const trendBars = computed<TrendBar[]>(() => {
  const rows = trendData.value.slice(-7)
  const max = Math.max(...rows.map((row) => row.requests), 1)
  return rows.map((row) => ({
    key: row.date,
    label: formatDayLabel(row.date),
    requests: row.requests,
    height: Math.max((row.requests / max) * 100, row.requests > 0 ? 5 : 0),
  }))
})

const trendTotalRequests = computed(() => trendData.value.reduce((total, row) => total + Number(row.requests || 0), 0))

const topModels = computed(() => {
  const rows = [...modelStats.value].sort((a, b) => b.requests - a.requests).slice(0, 5)
  const max = Math.max(...rows.map((row) => row.requests), 1)
  return rows.map((row) => ({ ...row, ratio: Math.max((row.requests / max) * 100, row.requests > 0 ? 4 : 0) }))
})

const routeSummary = computed<GroupSummaryRow[]>(() => {
  const byId = new Map<number, { id: number; name: string; platform: string; multiplier: number; requests: number; routes: number; keyIds: Set<number> }>()
  const groupById = new Map(availableGroups.value.map((group) => [group.id, group]))
  const requestById = new Map(groupStats.value.map((row) => [row.group_id, row]))

  const ensure = (id: number, fallbackName = '未命名分组') => {
    if (!byId.has(id)) {
      const group = groupById.get(id)
      const stat = requestById.get(id)
      byId.set(id, {
        id,
        name: group?.name || stat?.group_name || fallbackName,
        platform: group?.platform || 'composite',
        multiplier: Number(group?.rate_multiplier || 0),
        requests: Number(stat?.requests || 0),
        routes: 0,
        keyIds: new Set<number>(),
      })
    }
    return byId.get(id)!
  }

  availableGroups.value.forEach((group) => ensure(group.id, group.name))
  groupStats.value.forEach((stat) => ensure(stat.group_id, stat.group_name))
  apiKeys.value.forEach((key) => {
    const ids = new Set<number>()
    if (key.group_id !== null) ids.add(key.group_id)
    key.group_routes?.filter((route) => route.enabled).forEach((route) => ids.add(route.group_id))
    ids.forEach((id) => {
      const row = ensure(id)
      row.routes += 1
      if (key.status === 'active') row.keyIds.add(key.id)
    })
  })

  return Array.from(byId.values())
    .map((row) => ({ ...row, activeKeys: row.keyIds.size }))
    .filter((row) => row.routes > 0 || row.requests > 0)
    .sort((a, b) => b.requests - a.requests || b.routes - a.routes)
    .slice(0, 6)
})

const recentRequests = computed(() => [...recentUsage.value].sort((a, b) => Date.parse(b.created_at) - Date.parse(a.created_at)).slice(0, 6))

const upstreamNodes = computed<UpstreamNode[]>(() => {
  const byEndpoint = new Map<string, { requests: number; durations: number[]; lastSeen: string }>()
  recentUsage.value.forEach((request) => {
    const endpoint = request.upstream_endpoint?.trim()
    if (!endpoint) return
    const row = byEndpoint.get(endpoint) || { requests: 0, durations: [], lastSeen: request.created_at }
    row.requests += 1
    if (request.duration_ms !== null && request.duration_ms !== undefined) row.durations.push(request.duration_ms)
    if (Date.parse(request.created_at) > Date.parse(row.lastSeen)) row.lastSeen = request.created_at
    byEndpoint.set(endpoint, row)
  })
  return Array.from(byEndpoint.entries())
    .map(([endpoint, row]) => ({
      endpoint,
      requests: row.requests,
      avgDuration: row.durations.length ? row.durations.reduce((sum, value) => sum + value, 0) / row.durations.length : null,
      lastSeen: row.lastSeen,
    }))
    .sort((a, b) => b.requests - a.requests)
    .slice(0, 4)
})

const quotaRows = computed<QuotaRow[]>(() => platformQuotas.value.map((quota) => {
  const limit = quota.daily_limit_usd
  const usage = Number(quota.daily_usage_usd || 0)
  const ratio = limit && limit > 0 ? Math.min((usage / limit) * 100, 100) : 0
  return {
    ...quota,
    ratio,
    usage,
    limitLabel: limit && limit > 0 ? formatMoney(limit) : '不限额',
    windowLabel: limit && limit > 0 ? `今日额度 ${Math.round(ratio)}%` : '未设置每日上限',
  }
}).slice(0, 5))

const errorTotal = computed(() => recentErrors.value.length)

async function loadDashboard() {
  if (refreshing.value) return
  refreshing.value = true
  if (!stats.value) loading.value = true
  loadError.value = ''
  sectionErrors.value = []
  const failures: string[] = []

  const results = await Promise.allSettled([
    usageAPI.getDashboardStats(),
    usageAPI.getDashboardTrend({ start_date: startDate, end_date: endDate, granularity: 'day' }),
    usageAPI.getDashboardModels({ start_date: startDate, end_date: endDate }),
    usageAPI.getDashboardSnapshotV2({ start_date: startDate, end_date: endDate, granularity: 'day', include_group_stats: true }),
    usageAPI.getByDateRange(startDate, endDate),
    usageAPI.listMyErrorRequests({ page: 1, page_size: 5, start_date: startDate, end_date: endDate, sort_by: 'created_at', sort_order: 'desc' }),
    keysAPI.list(1, 50, { sort_by: 'last_used_at', sort_order: 'desc' }),
    userGroupsAPI.getAvailable(),
    getMyPlatformQuotas(),
  ])

  const [statsResult, trendResult, modelsResult, snapshotResult, usageResult, errorsResult, keysResult, groupsResult, quotaResult] = results

  if (statsResult.status === 'fulfilled') stats.value = statsResult.value
  else failures.push('指标')
  if (trendResult.status === 'fulfilled') trendData.value = trendResult.value.trend || []
  else failures.push('趋势')
  if (modelsResult.status === 'fulfilled') modelStats.value = modelsResult.value.models || []
  else failures.push('模型统计')
  if (snapshotResult.status === 'fulfilled') groupStats.value = (snapshotResult.value as UsageDashboardSnapshotV2Response).groups || []
  else failures.push('分组统计')
  if (usageResult.status === 'fulfilled') recentUsage.value = usageResult.value.items || []
  else failures.push('请求记录')
  if (errorsResult.status === 'fulfilled') recentErrors.value = errorsResult.value.items || []
  else failures.push('异常记录')
  if (keysResult.status === 'fulfilled') apiKeys.value = keysResult.value.items || []
  else failures.push('密钥')
  if (groupsResult.status === 'fulfilled') availableGroups.value = groupsResult.value || []
  else failures.push('分组')
  if (quotaResult.status === 'fulfilled') platformQuotas.value = quotaResult.value.platform_quotas || []
  else failures.push('额度')

  void authStore.refreshUser().catch(() => undefined)
  sectionErrors.value = failures.filter((name) => name !== '指标')
  if (statsResult.status === 'rejected' && !stats.value) {
    loadError.value = '无法读取当前账号的统计数据，请稍后重试。'
  }
  lastSyncedAt.value = new Date()
  loading.value = false
  refreshing.value = false
}

function formatNumber(value: number | null | undefined) {
  return new Intl.NumberFormat('zh-CN', { maximumFractionDigits: 0 }).format(Number(value || 0))
}

function formatMoney(value: number | null | undefined) {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 2, maximumFractionDigits: 4 }).format(Number(value || 0))
}

function formatDuration(value: number | null | undefined) {
  if (value === null || value === undefined || !Number.isFinite(value)) return '—'
  if (value >= 1000) return `${(value / 1000).toFixed(1)} s`
  return `${Math.round(value)} ms`
}

function formatMultiplier(value: number) {
  if (!value) return '—'
  return `${value.toFixed(3)}×`
}

function formatDayLabel(date: string) {
  const parsed = new Date(`${date}T00:00:00`)
  return Number.isNaN(parsed.getTime()) ? date.slice(-5) : new Intl.DateTimeFormat('zh-CN', { month: 'numeric', day: 'numeric' }).format(parsed)
}

function formatDateTime(date: string | null | undefined) {
  if (!date) return '—'
  const parsed = new Date(date)
  if (Number.isNaN(parsed.getTime())) return '—'
  return new Intl.DateTimeFormat('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }).format(parsed)
}

function compactEndpoint(endpoint: string | null | undefined) {
  if (!endpoint) return '未回传路径'
  try {
    const url = new URL(endpoint)
    return `${url.host}${url.pathname === '/' ? '' : url.pathname}`
  } catch {
    return endpoint.length > 42 ? `${endpoint.slice(0, 39)}…` : endpoint
  }
}

function groupName(groupId: number | null | undefined) {
  if (groupId == null) return '未绑定分组'
  return availableGroups.value.find((group) => group.id === groupId)?.name || `分组 #${groupId}`
}

function platformLabel(platform: string) {
  const labels: Record<string, string> = { anthropic: 'Claude', openai: 'OpenAI', gemini: 'Gemini', antigravity: 'Antigravity', grok: 'Grok', composite: '复合路径' }
  return labels[platform] || platform || '未标注'
}

function platformTone(platform: string) {
  const tones: Record<string, string> = { anthropic: 'tone-green', openai: 'tone-blue', gemini: 'tone-amber', antigravity: 'tone-ink', grok: 'tone-red', composite: 'tone-muted' }
  return tones[platform] || 'tone-muted'
}

onMounted(() => {
  void loadDashboard()
})
</script>

<style scoped>

/* Ledger operations console. The data layer above stays bound to the
   upstream user APIs; these styles give the console its own visual system. */
.console-shell {
  --console-ink: #16150f;
  --console-muted: #6b695f;
  --console-soft: #aaa496;
  --console-line: #dad5c8;
  --console-line-strong: #c6beaf;
  --console-surface: rgba(251, 250, 246, 0.86);
  --console-surface-strong: rgba(251, 250, 246, 0.96);
  --console-surface-hover: #f6f3eb;
  --console-bg: #f4f2ec;
  --console-indigo: #1e5c42;
  --console-cyan: #2f7658;
  --console-pink: #8f6b55;
  --console-amber: #b7791f;
  --console-red: #9e4d3d;
  position: relative;
  isolation: isolate;
  min-width: 0;
  min-height: calc(100vh - 7rem);
  overflow: hidden;
  border: 1px solid var(--console-line-strong);
  border-radius: 8px;
  background-color: var(--console-bg);
  background-image:
    linear-gradient(135deg, rgba(255, 255, 255, 0.68), rgba(244, 242, 236, 0.86)),
    linear-gradient(rgba(30, 92, 66, 0.045) 1px, transparent 1px),
    linear-gradient(90deg, rgba(30, 92, 66, 0.045) 1px, transparent 1px);
  background-size: auto, 26px 26px, 26px 26px;
  color: var(--console-ink);
  font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  box-shadow: 0 24px 70px rgba(77, 66, 42, 0.12);
}

.console-shell::before {
  position: absolute;
  z-index: -1;
  inset: 0;
  background: linear-gradient(105deg, rgba(30, 92, 66, 0.06), transparent 36%, rgba(183, 121, 31, 0.05));
  content: "";
  pointer-events: none;
}

.console-hero {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 1.5rem;
  border-bottom: 1px solid var(--console-line);
  padding: 1.45rem 1.5rem 1.35rem;
  background: var(--console-surface-strong);
  backdrop-filter: blur(18px);
}

.console-hero-copy { min-width: 0; }
.console-eyebrow,
.console-panel-kicker,
.console-kpi-label,
.console-sync-time,
.console-total span,
.console-route-foot,
.console-table th,
.console-footer,
.console-side-note {
  color: var(--console-muted);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 0.64rem;
  letter-spacing: 0.08em;
}

.console-eyebrow,
.console-panel-kicker {
  color: var(--console-indigo);
  font-weight: 700;
  letter-spacing: 0.14em;
}

.console-live-dot {
  display: inline-block;
  width: 0.42rem;
  height: 0.42rem;
  margin-right: 0.35rem;
  border-radius: 999px;
  background: var(--console-cyan);
  box-shadow: 0 0 0 0.22rem rgba(30, 92, 66, 0.14);
  vertical-align: middle;
}

.console-title-row {
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin-top: 0.55rem;
}

.console-title {
  font-family: Georgia, "Times New Roman", serif;
  font-size: 2.35rem;
  font-weight: 400;
  letter-spacing: 0;
  line-height: 1;
}

.console-date {
  color: var(--console-soft);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 0.7rem;
}

.console-subtitle {
  max-width: 40rem;
  margin-top: 0.65rem;
  color: var(--console-muted);
  font-size: 0.78rem;
  line-height: 1.55;
}

.console-hero-actions {
  display: flex;
  align-items: center;
  gap: 0.8rem;
  flex-shrink: 0;
}

.console-sync-state {
  display: flex;
  flex-direction: column;
  gap: 0.22rem;
  text-align: right;
}

.console-sync-mark {
  color: var(--console-cyan);
  font-size: 0.72rem;
  font-weight: 700;
}

.console-sync-time { white-space: nowrap; }

.console-refresh,
.console-alert-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  border: 1px solid var(--console-indigo);
  border-radius: 6px;
  background: var(--console-indigo);
  padding: 0.62rem 0.85rem;
  color: #fff;
  font-size: 0.72rem;
  font-weight: 700;
  box-shadow: 0 8px 18px rgba(30, 92, 66, 0.18);
  transition: transform 160ms ease, box-shadow 160ms ease, filter 160ms ease;
}

.console-refresh:hover:not(:disabled),
.console-alert-action:hover {
  filter: brightness(1.05);
  transform: translateY(-2px);
  box-shadow: 0 12px 24px rgba(30, 92, 66, 0.24);
}

.console-refresh:active:not(:disabled),
.console-alert-action:active { transform: translateY(1px) scale(0.98); }

.console-refresh:focus-visible,
.console-alert-action:focus-visible,
.console-text-link:focus-visible {
  outline: 2px solid var(--console-cyan);
  outline-offset: 3px;
}

.console-refresh:disabled { cursor: wait; opacity: 0.65; }

.console-alert {
  display: flex;
  align-items: center;
  gap: 0.7rem;
  margin: 1rem 1.5rem 0;
  border: 1px solid var(--console-line-strong);
  border-radius: 6px;
  padding: 0.72rem 0.8rem;
  background: var(--console-surface-strong);
  color: var(--console-ink);
  box-shadow: 0 8px 22px rgba(77, 66, 42, 0.07);
}

.console-alert-error { border-color: rgba(215, 104, 104, 0.5); color: var(--console-red); }
.console-alert-warn { border-color: rgba(217, 152, 77, 0.5); color: var(--console-amber); }
.console-alert strong { display: block; font-size: 0.76rem; }
.console-alert p { margin-top: 0.15rem; color: var(--console-muted); font-size: 0.69rem; }
.console-alert-icon { display: inline-flex; flex-shrink: 0; }
.console-alert-action { margin-left: auto; border-color: currentColor; background: transparent; color: inherit; box-shadow: none; }

.console-loading {
  display: flex;
  min-height: 24rem;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  gap: 0.55rem;
  color: var(--console-muted);
  font-size: 0.76rem;
}

.console-loading strong { color: var(--console-ink); font-size: 0.9rem; }
.console-loading-orbit { display: grid; width: 3.2rem; height: 3.2rem; place-items: center; border: 1px solid rgba(30, 92, 66, 0.35); border-radius: 999px; background: rgba(251, 250, 246, 0.75); box-shadow: 0 0 0 0.55rem rgba(30, 92, 66, 0.08); }

.console-kpi-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.75rem;
  padding: 1.2rem 1.5rem 0;
}

.console-kpi,
.console-panel {
  border: 1px solid var(--console-line);
  border-radius: 8px;
  background: var(--console-surface);
  box-shadow: 0 12px 32px rgba(77, 66, 42, 0.07);
  backdrop-filter: blur(16px);
}

.console-kpi {
  position: relative;
  min-width: 0;
  overflow: hidden;
  padding: 1rem;
  animation: console-rise 520ms cubic-bezier(0.22, 1, 0.36, 1) both;
  transition: transform 160ms ease, background-color 160ms ease, border-color 160ms ease;
}

.console-kpi:nth-child(2) { animation-delay: 60ms; }
.console-kpi:nth-child(3) { animation-delay: 120ms; }
.console-kpi:nth-child(4) { animation-delay: 180ms; }
.console-kpi:hover { border-color: var(--console-line-strong); background: var(--console-surface-hover); transform: translateY(-3px); }
.console-kpi::after { display: none; }
.console-kpi-green { color: var(--console-cyan); }
.console-kpi-blue { color: var(--console-indigo); }
.console-kpi-amber { color: var(--console-amber); }
.console-kpi-ink { color: var(--console-pink); }
.console-kpi-head { display: flex; align-items: center; justify-content: space-between; gap: 0.6rem; }
.console-kpi-label { color: var(--console-muted); }
.console-kpi-icon { display: inline-flex; width: 1.9rem; height: 1.9rem; align-items: center; justify-content: center; border: 1px solid currentColor; border-radius: 6px; background: rgba(255, 255, 255, 0.38); }
.console-kpi-value { display: block; margin-top: 1rem; overflow: hidden; color: var(--console-ink); font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace; font-size: 1.7rem; font-variant-numeric: tabular-nums; line-height: 1; text-overflow: ellipsis; white-space: nowrap; }
.console-kpi-note { min-height: 1.9rem; margin-top: 0.55rem; color: var(--console-muted); font-size: 0.68rem; line-height: 1.35; }
.console-kpi-index { position: absolute; right: 1rem; bottom: 0.75rem; color: currentColor; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace; font-size: 0.58rem; opacity: 0.6; }

.console-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(17rem, 21rem);
  align-items: start;
  gap: 1rem;
  padding: 1.2rem 1.5rem;
}

.console-main-column,
.console-side-column { display: flex; min-width: 0; flex-direction: column; gap: 1rem; }
.console-panel { min-width: 0; overflow: hidden; animation: console-rise 560ms 160ms cubic-bezier(0.22, 1, 0.36, 1) both; }
.console-panel:nth-child(2) { animation-delay: 220ms; }
.console-panel:nth-child(3) { animation-delay: 280ms; }
.console-side-column .console-panel { animation-delay: 320ms; }

.console-panel-heading,
.console-side-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px solid var(--console-line);
  padding: 1rem 1.05rem 0.9rem;
}

.console-panel-heading-tight { align-items: center; }
.console-panel-heading h2,
.console-side-heading h2 { margin-top: 0.28rem; color: var(--console-ink); font-family: Georgia, "Times New Roman", serif; font-size: 1.22rem; font-weight: 400; line-height: 1.1; }
.console-panel-heading p { margin-top: 0.35rem; color: var(--console-muted); font-size: 0.69rem; }
.console-text-link { display: inline-flex; align-items: center; gap: 0.25rem; flex-shrink: 0; color: var(--console-indigo); font-size: 0.7rem; font-weight: 700; transition: transform 160ms ease, color 160ms ease; }
.console-text-link:hover { color: var(--console-cyan); transform: translateX(2px); }
.console-total { flex-shrink: 0; text-align: right; }
.console-total strong { display: block; margin-top: 0.25rem; color: var(--console-ink); font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace; font-size: 1.45rem; font-variant-numeric: tabular-nums; }

.console-chart {
  position: relative;
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  align-items: end;
  gap: 0.7rem;
  min-height: 15.5rem;
  padding: 1.2rem 1.05rem 1rem;
}

.console-chart-grid { position: absolute; inset: 1.25rem 1.05rem 2.55rem; display: flex; flex-direction: column; justify-content: space-between; pointer-events: none; }
.console-chart-grid span { border-top: 1px dashed var(--console-line); }
.console-chart-column { position: relative; z-index: 1; display: flex; min-width: 0; height: 12.5rem; align-items: center; flex-direction: column; justify-content: flex-end; gap: 0.42rem; }
.console-chart-value,.console-chart-label { color: var(--console-muted); font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace; font-size: 0.6rem; font-variant-numeric: tabular-nums; }
.console-chart-track { display: flex; width: min(100%, 2.3rem); height: 10rem; align-items: flex-end; border: 1px solid var(--console-line); border-radius: 5px 5px 3px 3px; background: rgba(218, 213, 200, 0.3); overflow: hidden; }
.console-chart-fill { width: 100%; min-height: 0.25rem; border-radius: 4px 4px 0 0; background: linear-gradient(180deg, var(--console-indigo), var(--console-cyan)); box-shadow: 0 -8px 20px rgba(30, 92, 66, 0.2); transition: height 260ms ease; }

.console-route-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0.7rem; padding: 0.9rem; }
.console-route-card { min-width: 0; border: 1px solid var(--console-line); border-radius: 6px; background: rgba(255, 255, 255, 0.38); transition: transform 160ms ease, background-color 160ms ease, border-color 160ms ease; }
.console-route-card:hover { border-color: var(--console-line-strong); background: var(--console-surface-hover); transform: translateY(-2px); }
.console-route-top { display: flex; align-items: center; gap: 0.48rem; border-bottom: 1px solid var(--console-line); padding: 0.72rem 0.8rem; }
.console-route-number { color: var(--console-soft); font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace; font-size: 0.62rem; }
.console-route-platform { border-radius: 999px; padding: 0.2rem 0.45rem; color: #fff; font-size: 0.6rem; font-weight: 700; }
.console-route-platform.tone-green { background: var(--console-cyan); }
.console-route-platform.tone-blue { background: var(--console-indigo); }
.console-route-platform.tone-amber { background: var(--console-amber); }
.console-route-platform.tone-ink { background: var(--console-pink); }
.console-route-platform.tone-muted { background: var(--console-soft); }
.console-route-state { margin-left: auto; color: var(--console-cyan); font-size: 0.62rem; }
.console-route-state i { display: inline-block; width: 0.35rem; height: 0.35rem; margin-right: 0.2rem; border-radius: 999px; background: currentColor; vertical-align: middle; }
.console-route-body { padding: 0.85rem 0.8rem; }
.console-route-body h3 { overflow: hidden; color: var(--console-ink); font-size: 0.86rem; font-weight: 700; text-overflow: ellipsis; white-space: nowrap; }
.console-route-flow { display: flex; align-items: center; gap: 0.35rem; margin-top: 0.7rem; color: var(--console-muted); font-size: 0.62rem; }
.console-route-flow span { display: inline-flex; min-width: 0; align-items: center; gap: 0.22rem; }
.console-route-flow > svg { flex-shrink: 0; color: var(--console-soft); }
.console-route-foot { display: flex; justify-content: space-between; gap: 0.5rem; border-top: 1px solid var(--console-line); padding: 0.65rem 0.8rem; letter-spacing: 0; }
.console-route-foot strong { color: var(--console-indigo); font-size: 0.62rem; }

.console-table-wrap { overflow-x: auto; }
.console-table { width: 100%; min-width: 38rem; border-collapse: collapse; text-align: left; font-size: 0.72rem; }
.console-table th { border-bottom: 1px solid var(--console-line); padding: 0.62rem 1rem; font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase; }
.console-table td { border-bottom: 1px solid var(--console-line); padding: 0.78rem 1rem; vertical-align: middle; }
.console-table tbody tr { transition: background-color 160ms ease; }
.console-table tbody tr:hover { background: rgba(30, 92, 66, 0.06); }
.console-table td:first-child { width: 31%; }
.console-table-model { display: block; max-width: 12rem; overflow: hidden; color: var(--console-ink); text-overflow: ellipsis; white-space: nowrap; }
.console-table td:first-child span { display: block; margin-top: 0.25rem; color: var(--console-muted); font-size: 0.64rem; }
.console-table code { display: inline-block; max-width: 15rem; overflow: hidden; border: 1px solid var(--console-line); border-radius: 4px; padding: 0.2rem 0.35rem; color: var(--console-indigo); font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace; font-size: 0.62rem; text-overflow: ellipsis; white-space: nowrap; }
.console-table-mono { color: var(--console-muted); font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace; font-size: 0.65rem; font-variant-numeric: tabular-nums; }

.console-side-heading { align-items: center; padding: 0.9rem 0.95rem 0.78rem; }
.console-panel-icon { color: var(--console-indigo); }
.console-model-list,.console-node-list,.console-quota-list,.console-error-list { padding: 0.35rem 0.95rem 0.75rem; }
.console-model-row { display: flex; align-items: center; gap: 0.55rem; border-bottom: 1px solid var(--console-line); padding: 0.72rem 0; }
.console-model-row:last-child,.console-node-row:last-child,.console-error-row:last-child { border-bottom: 0; }
.console-model-rank { width: 1.35rem; color: var(--console-soft); font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace; font-size: 0.62rem; }
.console-model-name { overflow: hidden; color: var(--console-ink); font-size: 0.7rem; font-weight: 700; text-overflow: ellipsis; white-space: nowrap; }
.console-model-row > strong { color: var(--console-ink); font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace; font-size: 0.68rem; }
.console-progress { height: 0.32rem; margin-top: 0.42rem; overflow: hidden; border-radius: 999px; background: rgba(218, 213, 200, 0.68); }
.console-progress span { display: block; height: 100%; border-radius: inherit; background: linear-gradient(90deg, var(--console-indigo), var(--console-cyan)); transition: width 240ms ease; }
.console-node-row,.console-error-row { display: flex; align-items: center; gap: 0.5rem; border-bottom: 1px solid var(--console-line); padding: 0.72rem 0; }
.console-node-status,.console-error-dot { width: 0.46rem; height: 0.46rem; flex-shrink: 0; border-radius: 999px; background: var(--console-cyan); box-shadow: 0 0 0 0.24rem rgba(30, 92, 66, 0.12); }
.console-node-row strong,.console-error-row strong { display: block; overflow: hidden; color: var(--console-ink); font-size: 0.68rem; text-overflow: ellipsis; white-space: nowrap; }
.console-node-row span,.console-error-row span { display: block; margin-top: 0.22rem; color: var(--console-muted); font-size: 0.61rem; }
.console-node-row em { color: var(--console-indigo); font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace; font-size: 0.6rem; font-style: normal; white-space: nowrap; }
.console-side-note { padding: 0.65rem 0 0.1rem; font-size: 0.6rem; letter-spacing: 0; line-height: 1.45; }
.console-quota-row { border-bottom: 1px solid var(--console-line); padding: 0.72rem 0; }
.console-quota-row:last-child { border-bottom: 0; }
.console-quota-row > div:first-child { display: flex; justify-content: space-between; gap: 0.5rem; }
.console-quota-row strong { color: var(--console-ink); font-size: 0.68rem; }
.console-quota-row span,.console-quota-row p { color: var(--console-muted); font-size: 0.6rem; }
.console-quota-row p { margin-top: 0.35rem; }
.console-progress-quota { margin-top: 0.55rem; }
.console-progress-quota span { background: linear-gradient(90deg, var(--console-cyan), var(--console-amber)); }
.console-error-count { color: var(--console-red); font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace; font-size: 0.82rem; font-weight: 700; }
.console-error-dot { background: var(--console-red); box-shadow: 0 0 0 0.24rem rgba(158, 77, 61, 0.12); }
.console-error-row time { color: var(--console-muted); font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace; font-size: 0.58rem; white-space: nowrap; }

.console-empty { display: flex; min-height: 8rem; align-items: center; justify-content: center; gap: 0.5rem; color: var(--console-muted); font-size: 0.72rem; }
.console-empty-small { min-height: 5.5rem; padding: 0.8rem; text-align: center; }
.console-success-icon { color: var(--console-cyan); }
.console-footer { display: flex; justify-content: space-between; gap: 1rem; border-top: 1px solid var(--console-line); padding: 0.85rem 1.5rem 1rem; letter-spacing: 0; }
.console-footer .console-live-dot { width: 0.34rem; height: 0.34rem; margin-right: 0.32rem; box-shadow: none; }

@keyframes console-rise {
  from { opacity: 0; transform: translateY(9px); }
  to { opacity: 1; transform: translateY(0); }
}

:global(.dark .console-shell) {
  --console-ink: #f4f2ec;
  --console-muted: #aaa69a;
  --console-soft: #827e72;
  --console-line: #47443a;
  --console-line-strong: #615d50;
  --console-surface: rgba(36, 35, 31, 0.9);
  --console-surface-strong: rgba(36, 35, 31, 0.97);
  --console-surface-hover: #302f29;
  --console-bg: #1b1b18;
  --console-indigo: #8fc2a5;
  --console-cyan: #a7cfb3;
  --console-pink: #c99e86;
  --console-amber: #d3a55a;
  --console-red: #d78a78;
  background-image:
    linear-gradient(135deg, rgba(36, 35, 31, 0.96), rgba(27, 27, 24, 0.9)),
    linear-gradient(rgba(143, 194, 165, 0.045) 1px, transparent 1px),
    linear-gradient(90deg, rgba(143, 194, 165, 0.045) 1px, transparent 1px);
  box-shadow: 0 24px 70px rgba(0, 0, 0, 0.35);
}

:global(.dark .console-shell::before) { background: linear-gradient(105deg, rgba(143, 194, 165, 0.08), transparent 38%, rgba(211, 165, 90, 0.06)); }
:global(.dark .console-route-card) { background: rgba(36, 35, 31, 0.88); }
:global(.dark .console-kpi-icon) { background: rgba(255, 255, 255, 0.07); }
:global(.dark .console-chart-track),
:global(.dark .console-progress) { background: rgba(71, 68, 58, 0.74); }

@media (max-width: 1023px) {
  .console-layout { grid-template-columns: minmax(0, 1fr); }
  .console-side-column { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); align-items: start; }
}

@media (max-width: 767px) {
  .console-shell { border-right: 0; border-left: 0; border-radius: 0; }
  .console-hero { align-items: flex-start; flex-direction: column; padding: 1.15rem 0.9rem 1rem; }
  .console-title { font-size: 1.8rem; }
  .console-hero-actions { width: 100%; align-items: stretch; justify-content: space-between; }
  .console-sync-state { text-align: left; }
  .console-refresh { min-width: 7.2rem; }
  .console-alert { margin-right: 0.9rem; margin-left: 0.9rem; align-items: flex-start; }
  .console-alert-action { align-self: center; }
  .console-kpi-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0.6rem; padding: 0.9rem 0.9rem 0; }
  .console-kpi { padding: 0.8rem; }
  .console-kpi-value { font-size: 1.32rem; }
  .console-layout { gap: 0.8rem; padding: 0.9rem; }
  .console-main-column,.console-side-column { gap: 0.8rem; }
  .console-side-column { display: flex; }
  .console-panel-heading { padding: 0.85rem 0.85rem 0.75rem; }
  .console-panel-heading h2,.console-side-heading h2 { font-size: 1.05rem; }
  .console-panel-heading p { max-width: 17rem; }
  .console-total strong { font-size: 1.1rem; }
  .console-chart { min-height: 13.5rem; gap: 0.28rem; padding-right: 0.75rem; padding-left: 0.75rem; }
  .console-chart-grid { right: 0.75rem; left: 0.75rem; }
  .console-chart-column { height: 10.8rem; }
  .console-chart-track { height: 8.4rem; width: min(100%, 1.9rem); }
  .console-route-grid { grid-template-columns: minmax(0, 1fr); padding: 0.75rem; }
  .console-table { min-width: 35rem; }
  .console-footer { align-items: flex-start; flex-direction: column; padding-right: 0.9rem; padding-left: 0.9rem; }
}

@media (max-width: 430px) {
  .console-kpi-grid { grid-template-columns: minmax(0, 1fr); }
  .console-hero-actions { align-items: flex-start; flex-direction: column; }
  .console-refresh { width: 100%; }
  .console-sync-state { flex-direction: row; align-items: baseline; gap: 0.55rem; }
  .console-alert { flex-wrap: wrap; }
  .console-alert-action { width: 100%; }
}

@media (prefers-reduced-motion: reduce) {
  .console-shell *,
  .console-shell *::before,
  .console-shell *::after {
    animation-duration: 0.001ms !important;
    animation-iteration-count: 1 !important;
    scroll-behavior: auto !important;
    transition-duration: 0.001ms !important;
  }
}
</style>

<template>
  <AppLayout>
    <section class="scheme3-channel-status">
      <header class="scheme3-channel-status-header">
        <div class="scheme3-channel-status-heading">
          <p class="scheme3-channel-status-kicker">运行观测 / 渠道账本</p>
          <h1>渠道状态</h1>
          <p class="scheme3-channel-status-subtitle">查看可用渠道的连通性、延迟和近期稳定度。</p>
        </div>
        <div class="scheme3-channel-status-ledger" aria-label="渠道状态概览">
          <span>
            <strong>{{ items.length }}</strong>
            <small>监测渠道</small>
          </span>
          <span>
            <strong class="is-positive">{{ operationalCount }}</strong>
            <small>运行正常</small>
          </span>
          <span>
            <strong :class="degradedCount > 0 ? 'is-warning' : 'is-positive'">{{ degradedCount }}</strong>
            <small>需要留意</small>
          </span>
          <span>
            <strong>{{ currentWindowLabel }}</strong>
            <small>观测窗口</small>
          </span>
        </div>
      </header>

      <div class="scheme3-channel-status-toolbar">
        <MonitorHero
          :overall-status="overallStatus"
          :interval-seconds="DEFAULT_INTERVAL_SECONDS"
          :window="currentWindow"
          :loading="loading"
          :auto-refresh="autoRefresh"
          @update:window="handleWindowChange"
          @refresh="manualReload"
        />
      </div>

      <div class="scheme3-channel-status-grid">
        <MonitorCardGrid
          :items="items"
          :window="currentWindow"
          :countdown-seconds="countdown"
          :loading="loading"
          :detail-cache="detailCache"
          @card-click="openDetail"
        />
      </div>

      <MonitorDetailDialog
        :show="showDetail"
        :monitor-id="detailTarget?.id ?? null"
        :title="detailTitle"
        @close="closeDetail"
      />
    </section>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import {
  list as listChannelMonitorViews,
  status as fetchChannelMonitorDetail,
  type UserMonitorView,
  type UserMonitorDetail,
} from '@/api/channelMonitor'
import AppLayout from '@/components/layout/AppLayout.vue'
import MonitorHero, {
  type MonitorWindow,
  type OverallStatus,
} from '@/components/user/monitor/MonitorHero.vue'
import MonitorCardGrid from '@/components/user/monitor/MonitorCardGrid.vue'
import MonitorDetailDialog from '@/components/user/MonitorDetailDialog.vue'
import { DEFAULT_INTERVAL_SECONDS, STATUS_OPERATIONAL } from '@/constants/channelMonitor'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const { t } = useI18n()
const appStore = useAppStore()

// ── State ──
const items = ref<UserMonitorView[]>([])
const loading = ref(false)
const currentWindow = ref<MonitorWindow>('7d')
const detailCache = reactive<Record<number, UserMonitorDetail>>({})
const showDetail = ref(false)
const detailTarget = ref<UserMonitorView | null>(null)

let abortController: AbortController | null = null

const autoRefresh = useAutoRefresh({
  storageKey: 'channel-status-auto-refresh',
  intervals: [30, 60, 120] as const,
  defaultInterval: DEFAULT_INTERVAL_SECONDS,
  onRefresh: () => reload(true),
  shouldPause: () => document.hidden || loading.value,
})
const countdown = autoRefresh.countdown

// ── Computed ──
const overallStatus = computed<OverallStatus>(() => {
  if (items.value.length === 0) return 'operational'
  for (const it of items.value) {
    if (it.primary_status === 'failed' || it.primary_status === 'error') return 'degraded'
    if (it.primary_status !== STATUS_OPERATIONAL) return 'degraded'
  }
  return 'operational'
})

const detailTitle = computed(() => {
  return detailTarget.value?.name || t('channelStatus.detailTitle')
})

const operationalCount = computed(() =>
  items.value.filter((item) => item.primary_status === STATUS_OPERATIONAL).length,
)

const degradedCount = computed(() =>
  items.value.filter((item) => item.primary_status !== STATUS_OPERATIONAL).length,
)

const currentWindowLabel = computed(() => t(`channelStatus.windowTab.${currentWindow.value}`))

// ── Loaders ──
async function reload(silent = false) {
  if (abortController) abortController.abort()
  const ctrl = new AbortController()
  abortController = ctrl
  if (!silent) loading.value = true
  try {
    const res = await listChannelMonitorViews({ signal: ctrl.signal })
    if (ctrl.signal.aborted || abortController !== ctrl) return
    items.value = res.items || []
  } catch (err: unknown) {
    const e = err as { name?: string; code?: string }
    if (e?.name === 'AbortError' || e?.code === 'ERR_CANCELED') return
    appStore.showError(extractApiErrorMessage(err, t('channelStatus.loadError')))
  } finally {
    if (abortController === ctrl) {
      if (!silent) loading.value = false
      countdown.value = DEFAULT_INTERVAL_SECONDS
      abortController = null
    }
  }
}

async function manualReload() {
  await reload(false)
  // After base reload, refresh any cached detail records so non-7d availability
  // values stay in sync without forcing the user to switch tabs again.
  if (currentWindow.value !== '7d') {
    await Promise.all(items.value.map(it => loadDetail(it.id, true)))
  }
}

async function loadDetail(id: number, force = false) {
  if (!force && detailCache[id]) return
  try {
    detailCache[id] = await fetchChannelMonitorDetail(id)
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('channelStatus.detailLoadError')))
  }
}

async function ensureDetailsForWindow() {
  if (currentWindow.value === '7d') return
  await Promise.all(items.value.map(it => loadDetail(it.id)))
}

// ── Handlers ──
async function handleWindowChange(value: MonitorWindow) {
  currentWindow.value = value
  await ensureDetailsForWindow()
}

function openDetail(row: UserMonitorView) {
  detailTarget.value = row
  showDetail.value = true
}

function closeDetail() {
  showDetail.value = false
  detailTarget.value = null
}

watch(items, () => {
  void ensureDetailsForWindow()
})

watch(
  () => appStore.cachedPublicSettings?.channel_monitor_enabled,
  (enabled) => {
    if (enabled === false) autoRefresh.stop()
    else if (autoRefresh.enabled.value) autoRefresh.start()
  },
)

onMounted(() => {
  void reload(false)
  if (appStore.cachedPublicSettings?.channel_monitor_enabled !== false) {
    autoRefresh.setEnabled(true)
  }
})

onBeforeUnmount(() => {
  if (abortController) abortController.abort()
})
</script>

<style scoped>
.scheme3-channel-status {
  --monitor-paper: #f4f2ec;
  --monitor-card: #fffefa;
  --monitor-ink: #27251f;
  --monitor-muted: #777266;
  --monitor-soft: #a49e90;
  --monitor-line: #d8d2c3;
  --monitor-accent: #1e5c42;
  --monitor-amber: #b7791f;
  --monitor-danger: #9e4d3d;
  color: var(--monitor-ink);
}

.scheme3-channel-status-header {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 1.25rem;
  margin-bottom: 0.9rem;
  border-bottom: 1px solid var(--monitor-line);
  padding: 0.1rem 0 1rem;
}

.scheme3-channel-status-kicker {
  margin: 0;
  color: var(--monitor-muted);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.61rem;
  font-weight: 800;
  letter-spacing: 0.1em;
}

.scheme3-channel-status-heading h1 {
  margin: 0.34rem 0 0;
  color: var(--monitor-ink);
  font-family: Georgia, 'Times New Roman', serif;
  font-size: clamp(1.55rem, 2.6vw, 2.1rem);
  font-weight: 500;
  letter-spacing: 0;
}

.scheme3-channel-status-subtitle {
  max-width: 31rem;
  margin: 0.42rem 0 0;
  color: var(--monitor-muted);
  font-size: 0.74rem;
  line-height: 1.55;
}

.scheme3-channel-status-ledger {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  border: 1px solid var(--monitor-line);
  border-radius: 7px;
  background: var(--monitor-card);
}

.scheme3-channel-status-ledger span {
  display: grid;
  min-width: 4.8rem;
  gap: 0.08rem;
  border-right: 1px solid var(--monitor-line);
  padding: 0.48rem 0.68rem;
  text-align: right;
}

.scheme3-channel-status-ledger span:last-child { border-right: 0; }

.scheme3-channel-status-ledger strong {
  color: var(--monitor-accent);
  font-family: Georgia, 'Times New Roman', serif;
  font-size: 1rem;
  font-weight: 600;
  line-height: 1.1;
}

.scheme3-channel-status-ledger strong.is-positive { color: var(--monitor-accent); }
.scheme3-channel-status-ledger strong.is-warning { color: var(--monitor-amber); }

.scheme3-channel-status-ledger small {
  color: var(--monitor-muted);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.52rem;
  font-weight: 700;
  letter-spacing: 0.04em;
}

.scheme3-channel-status-toolbar {
  border-bottom: 1px solid var(--monitor-line);
  padding-bottom: 0.75rem;
}

.scheme3-channel-status-grid { padding-top: 0.9rem; }

.scheme3-channel-status :deep(.scheme3-monitor-empty) {
  border: 1px dashed var(--monitor-line);
  border-radius: 8px;
  background: rgba(255, 254, 250, 0.72);
}

:global(.dark .scheme3-channel-status) {
  --monitor-paper: #1b1b18;
  --monitor-card: #24231f;
  --monitor-ink: #f4f2ec;
  --monitor-muted: #aaa69a;
  --monitor-soft: #827e72;
  --monitor-line: #47443a;
  --monitor-accent: #8fc2a5;
  --monitor-amber: #d3a55a;
  --monitor-danger: #d38b79;
}

:global(html.dark .scheme3-channel-status .scheme3-monitor-empty) {
  background: rgba(36, 35, 31, 0.72);
}

@media (max-width: 767px) {
  .scheme3-channel-status-header {
    align-items: stretch;
    flex-direction: column;
    gap: 0.8rem;
    margin-bottom: 0.7rem;
  }

  .scheme3-channel-status-ledger {
    width: 100%;
    justify-content: stretch;
  }

  .scheme3-channel-status-ledger span {
    min-width: 0;
    flex: 1 1 45%;
    padding: 0.48rem 0.42rem;
  }

  .scheme3-channel-status-toolbar { padding-bottom: 0.55rem; }
  .scheme3-channel-status-grid { padding-top: 0.7rem; }
}
</style>

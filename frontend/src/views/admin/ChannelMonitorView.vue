<template>
  <AppLayout>
    <div class="scheme3-admin-monitor w-full min-w-0 space-y-6 pb-8">
      <header
        class="scheme3-admin-monitor-header"
      >
        <div class="scheme3-admin-monitor-heading">
          <p class="scheme3-admin-monitor-kicker">运维台 / 渠道观测</p>
          <h1 class="scheme3-admin-monitor-title flex items-center gap-2">
          <span class="scheme3-admin-monitor-mark inline-flex h-7 w-7 items-center justify-center">
            <Icon name="chart" size="sm" />
          </span>
          {{ t('admin.channelMonitor.title') }}
          </h1>
        <p class="scheme3-admin-monitor-description mt-1.5 text-xs">
          {{
            isV1Mode
              ? t('channelMonitorV2.admin.descriptionV1')
              : t('channelMonitorV2.admin.descriptionV2')
          }}
        </p>
        </div>
        <div class="scheme3-admin-monitor-tabs">
          <div
            class="scheme3-admin-monitor-segments inline-flex w-full max-w-xl flex-wrap sm:w-auto"
            role="tablist"
            :aria-label="t('channelMonitorV2.admin.tabAria')"
          >
            <button
              type="button"
              role="tab"
              class="scheme3-admin-monitor-segment flex-1 sm:flex-none"
              :class="adminMonitorTab === 'v2' ? 'is-active' : ''"
              :aria-selected="adminMonitorTab === 'v2'"
              @click="adminMonitorTab = 'v2'"
            >
              {{ t('channelMonitorV2.admin.tabV2') }}
            </button>
            <button
              type="button"
              role="tab"
              class="scheme3-admin-monitor-segment flex-1 sm:flex-none"
              :class="adminMonitorTab === 'legacy' ? 'is-active' : ''"
              :aria-selected="adminMonitorTab === 'legacy'"
              @click="adminMonitorTab = 'legacy'"
            >
              {{ isV1Mode ? t('channelMonitorV2.admin.tabV1Active') : t('channelMonitorV2.admin.tabV1History') }}
            </button>
          </div>
        </div>
      </header>

      <MonitorSettingsPanel v-if="adminMonitorTab === 'v2'" />

      <section v-else class="scheme3-admin-monitor-legacy">
        <div class="scheme3-admin-monitor-legacy-toolbar">
          <MonitorFiltersBar
            v-model:search="searchQuery"
            v-model:provider="providerFilter"
            v-model:enabled="enabledFilter"
            :loading="loading"
            @reload="reload"
            @create="openCreateDialog"
            @manage-templates="showTemplateManager = true"
            @search-input="handleSearch"
          />
        </div>

        <div class="scheme3-admin-monitor-table-frame">
          <DataTable
            :columns="columns"
            :data="monitors"
            :loading="loading"
            scheme3
            mobile-horizontal-scroll
          >
          <template #cell-name="{ row, value }">
            <div class="flex items-center gap-1.5">
              <span class="scheme3-monitor-table-primary font-medium">{{ value }}</span>
              <HelpTooltip v-if="row.api_key_decrypt_failed" :content="t('admin.channelMonitor.apiKeyDecryptFailed')" teleport-class="scheme3-monitor-tooltip">
                <template #trigger>
                  <Icon name="exclamationTriangle" size="sm" class="scheme3-monitor-warning" />
                </template>
              </HelpTooltip>
            </div>
          </template>

          <template #cell-provider="{ row }">
            <span class="scheme3-monitor-provider-badge inline-flex items-center px-2 py-0.5 text-xs font-medium" :class="providerClass(row.provider)">
              {{ providerLabel(row.provider) }}
            </span>
          </template>

          <template #cell-primary_model="{ row }">
            <MonitorPrimaryModelCell :row="row" />
          </template>

          <template #cell-availability_7d="{ row }">
            <span class="scheme3-monitor-table-primary text-sm">{{ formatAvailability(row) }}</span>
          </template>

          <template #cell-latency="{ row }">
            <span class="scheme3-monitor-table-primary text-sm">{{ formatLatency(row.primary_latency_ms) }}</span>
          </template>

            <template #cell-enabled="{ row }">
              <Scheme3V2Toggle :model-value="row.enabled" @update:model-value="toggleEnabled(row)" />
            </template>

          <template #cell-actions="{ row }">
            <MonitorActionsCell
              :row="row"
              :running="runningId === row.id"
              :duplicating="duplicatingIds.has(row.id)"
              @run="handleRunNow"
              @duplicate="handleDuplicate"
              @edit="openEditDialog"
              @delete="handleDelete"
            />
          </template>

            <template #empty>
              <div class="scheme3-admin-monitor-empty">
                <span class="scheme3-admin-monitor-empty-mark"><Icon name="server" size="lg" /></span>
                <h3>{{ t('admin.channelMonitor.noMonitorsYet') }}</h3>
                <p>{{ t('admin.channelMonitor.createFirstMonitor') }}</p>
                <button type="button" class="scheme3-admin-monitor-empty-action" @click="openCreateDialog">
                  <Icon name="plus" size="sm" />
                  {{ t('admin.channelMonitor.createButton') }}
                </button>
              </div>
            </template>
          </DataTable>
        </div>

        <footer v-if="pagination.total > 0" class="scheme3-admin-monitor-pagination">
          <Pagination
            scheme3
            :page="pagination.page"
            :total="pagination.total"
            :page-size="pagination.page_size"
            @update:page="onPageChange"
            @update:pageSize="onPageSizeChange"
          />
        </footer>
      </section>
    </div>

    <MonitorFormDialog
      :show="showDialog"
      :monitor="editing"
      @close="closeDialog"
      @saved="reload"
    />

    <MonitorTemplateManagerDialog
      :show="showTemplateManager"
      @close="showTemplateManager = false"
      @updated="reload"
    />

    <MonitorRunResultDialog
      :show="showRunResult"
      :results="runResults"
      @close="showRunResult = false"
    />

    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('common.delete')"
      :message="deleteConfirmMessage"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      :danger="true"
      content-class="scheme3-monitor-dialog"
      @confirm="confirmDelete"
      @cancel="showDeleteDialog = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { adminAPI } from '@/api/admin'
import type {
  ChannelMonitor,
  CheckResult,
  ListParams,
  Provider,
} from '@/api/admin/channelMonitor'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import HelpTooltip from '@/components/common/HelpTooltip.vue'
import Icon from '@/components/icons/Icon.vue'
import MonitorFiltersBar from '@/components/admin/monitor/MonitorFiltersBar.vue'
import MonitorFormDialog from '@/components/admin/monitor/MonitorFormDialog.vue'
import MonitorTemplateManagerDialog from '@/components/admin/monitor/MonitorTemplateManagerDialog.vue'
import MonitorRunResultDialog from '@/components/admin/monitor/MonitorRunResultDialog.vue'
import MonitorPrimaryModelCell from '@/components/admin/monitor/MonitorPrimaryModelCell.vue'
import MonitorActionsCell from '@/components/admin/monitor/MonitorActionsCell.vue'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import { useChannelMonitorFormat } from '@/composables/useChannelMonitorFormat'
import MonitorSettingsPanel from '@/features/channel-monitor-v2/MonitorSettingsPanel.vue'
import Scheme3V2Toggle from '@/features/channel-monitor-v2/Scheme3V2Toggle.vue'
import { isChannelMonitorV1Mode } from '@/utils/featureFlags'

const { t } = useI18n()
const appStore = useAppStore()
const isV1Mode = computed(() => isChannelMonitorV1Mode())
const adminMonitorTab = ref<'v2' | 'legacy'>(isChannelMonitorV1Mode() ? 'legacy' : 'v2')
const {
  providerLabel,
  formatLatency,
  formatAvailability,
} = useChannelMonitorFormat()

function providerClass(provider: Provider | string): string {
  switch (provider) {
    case 'openai': return 'is-openai'
    case 'anthropic': return 'is-anthropic'
    case 'gemini': return 'is-gemini'
    case 'grok': return 'is-grok'
    default: return 'is-unknown'
  }
}

const monitors = ref<ChannelMonitor[]>([])
const loading = ref(false)
const runningId = ref<number | null>(null)
const searchQuery = ref('')
const providerFilter = ref<Provider | ''>('')
const enabledFilter = ref<'' | 'true' | 'false'>('')
const pagination = reactive({ page: 1, page_size: getPersistedPageSize(), total: 0 })

const showDialog = ref(false)
const showTemplateManager = ref(false)
const editing = ref<ChannelMonitor | null>(null)
const showDeleteDialog = ref(false)
const deleting = ref<ChannelMonitor | null>(null)
const showRunResult = ref(false)
const runResults = ref<CheckResult[]>([])
const duplicatingIds = reactive(new Set<number>())

let abortController: AbortController | null = null
let searchTimeout: ReturnType<typeof setTimeout> | null = null

const columns = computed<Column[]>(() => [
  { key: 'name', label: t('admin.channelMonitor.columns.name'), sortable: false },
  { key: 'provider', label: t('admin.channelMonitor.columns.provider'), sortable: false },
  { key: 'primary_model', label: t('admin.channelMonitor.columns.primaryModel'), sortable: false },
  { key: 'availability_7d', label: t('admin.channelMonitor.columns.availability7d'), sortable: false },
  { key: 'latency', label: t('admin.channelMonitor.columns.latency'), sortable: false },
  { key: 'enabled', label: t('admin.channelMonitor.columns.enabled'), sortable: false },
  { key: 'actions', label: t('admin.channelMonitor.columns.actions'), sortable: false },
])

const deleteConfirmMessage = computed(() => {
  const name = deleting.value?.name || ''
  return t('admin.channelMonitor.deleteConfirm', { name })
})

async function reload() {
  if (abortController) abortController.abort()
  const ctrl = new AbortController()
  abortController = ctrl
  loading.value = true
  try {
    const params: ListParams = {
      page: pagination.page,
      page_size: pagination.page_size,
    }
    if (providerFilter.value) params.provider = providerFilter.value
    if (enabledFilter.value === 'true') params.enabled = true
    if (enabledFilter.value === 'false') params.enabled = false
    if (searchQuery.value.trim()) params.search = searchQuery.value.trim()

    const res = await adminAPI.channelMonitor.list(params, { signal: ctrl.signal })
    if (ctrl.signal.aborted || abortController !== ctrl) return
    monitors.value = res.items || []
    pagination.total = res.total
  } catch (err: unknown) {
    const e = err as { name?: string; code?: string }
    if (e?.name === 'AbortError' || e?.code === 'ERR_CANCELED') return
    appStore.showError(extractApiErrorMessage(err, t('admin.channelMonitor.loadError')))
  } finally {
    if (abortController === ctrl) {
      loading.value = false
      abortController = null
    }
  }
}

function handleSearch() {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    pagination.page = 1
    reload()
  }, 300)
}

function onPageChange(page: number) {
  pagination.page = page
  reload()
}

function onPageSizeChange(size: number) {
  pagination.page_size = size
  pagination.page = 1
  reload()
}

function openCreateDialog() {
  editing.value = null
  showDialog.value = true
}

function openEditDialog(row: ChannelMonitor) {
  editing.value = row
  showDialog.value = true
}

function closeDialog() {
  showDialog.value = false
  editing.value = null
}

async function toggleEnabled(row: ChannelMonitor) {
  const next = !row.enabled
  try {
    await adminAPI.channelMonitor.update(row.id, { enabled: next })
    row.enabled = next
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  }
}

async function handleRunNow(row: ChannelMonitor) {
  if (!isV1Mode.value) {
    appStore.showError(t('admin.channelMonitor.runFailed'))
    return
  }
  if (runningId.value != null) return
  runningId.value = row.id
  try {
    const res = await adminAPI.channelMonitor.runNow(row.id)
    runResults.value = res.results || []
    showRunResult.value = true
    appStore.showSuccess(t('admin.channelMonitor.runSuccess'))
    // Refresh row to get latest status from backend
    void reload()
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('admin.channelMonitor.runFailed')))
  } finally {
    runningId.value = null
  }
}

async function handleDuplicate(row: ChannelMonitor) {
  if (row.api_key_decrypt_failed) {
    appStore.showError(t('admin.channelMonitor.duplicateKeyUnavailable'))
    return
  }
  if (duplicatingIds.has(row.id)) return

  duplicatingIds.add(row.id)
  try {
    const duplicate = await adminAPI.channelMonitor.duplicate(row.id)
    appStore.showSuccess(t('admin.channelMonitor.duplicateSuccess', { name: duplicate.name }))
    await reload()
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('admin.channelMonitor.duplicateFailed')))
  } finally {
    duplicatingIds.delete(row.id)
  }
}

function handleDelete(row: ChannelMonitor) {
  deleting.value = row
  showDeleteDialog.value = true
}

async function confirmDelete() {
  if (!deleting.value) return
  try {
    await adminAPI.channelMonitor.del(deleting.value.id)
    appStore.showSuccess(t('admin.channelMonitor.deleteSuccess'))
    showDeleteDialog.value = false
    deleting.value = null
    reload()
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  }
}

watch(adminMonitorTab, (tab) => {
  if (tab === 'legacy' && monitors.value.length === 0) void reload()
})
onMounted(() => {
  if (adminMonitorTab.value === 'legacy') void reload()
})
onUnmounted(() => {
  if (searchTimeout) clearTimeout(searchTimeout)
  abortController?.abort()
})
</script>

<style scoped>
.scheme3-admin-monitor {
  --admin-paper: #f4f2ec;
  --admin-surface: #fffefa;
  --admin-subtle: #f1eee6;
  --admin-ink: #27251f;
  --admin-muted: #777266;
  --admin-line: #d8d2c3;
  --admin-accent: #1e5c42;
  color: var(--admin-ink);
}
.scheme3-admin-monitor-header {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: end;
  gap: 1rem;
  border-bottom: 1px solid var(--admin-line);
  padding: .1rem 0 1rem;
}
.scheme3-admin-monitor-kicker {
  margin: 0;
  color: var(--admin-muted);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: .61rem;
  font-weight: 800;
  letter-spacing: .1em;
}
.scheme3-admin-monitor-title {
  margin: .34rem 0 0;
  color: var(--admin-ink);
  font-family: Georgia, 'Times New Roman', serif;
  font-size: 1.65rem;
  font-weight: 500;
}
.scheme3-admin-monitor-mark {
  border: 1px solid rgba(30, 92, 66, .28);
  border-radius: 7px;
  background: rgba(30, 92, 66, .08);
  color: var(--admin-accent);
}
.scheme3-admin-monitor-description { max-width: 42rem; color: var(--admin-muted); line-height: 1.5; }
.scheme3-admin-monitor-tabs { align-self: end; }
.scheme3-admin-monitor-segments {
  gap: 0;
  border: 1px solid var(--admin-line);
  border-radius: 7px;
  background: var(--admin-subtle);
  padding: 2px;
}
.scheme3-admin-monitor-segment {
  min-height: 1.85rem;
  border-radius: 5px;
  color: var(--admin-muted);
  padding: .3rem .68rem;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: .62rem;
  font-weight: 800;
  transition: background-color 150ms ease, color 150ms ease;
}
.scheme3-admin-monitor-segment:hover { color: var(--admin-ink); }
.scheme3-admin-monitor-segment.is-active { background: var(--admin-surface); color: var(--admin-accent); box-shadow: 0 2px 6px rgba(54, 48, 34, .08); }
.scheme3-admin-monitor-legacy {
  display: grid;
  min-width: 0;
  gap: .85rem;
}
.scheme3-admin-monitor-legacy-toolbar {
  border-bottom: 1px solid var(--admin-line);
  padding: .1rem 0 .85rem;
}
.scheme3-admin-monitor-table-frame {
  min-width: 0;
  overflow: hidden;
  border: 1px solid var(--admin-line);
  border-radius: 8px;
  background: var(--admin-surface);
  box-shadow: 0 10px 24px rgba(54, 48, 34, .05);
}
.scheme3-admin-monitor-table-frame :deep(.table-wrapper) {
  min-height: 21rem;
  max-height: min(58vh, 40rem);
  overflow: auto;
  scrollbar-color: var(--admin-line) transparent;
  scrollbar-width: thin;
}
.scheme3-admin-monitor-table-frame :deep(table) {
  width: 100%;
  min-width: 48rem;
  border-collapse: collapse;
}
.scheme3-admin-monitor-table-frame :deep(.table-header) {
  position: sticky;
  top: 0;
  z-index: 2;
  background: var(--admin-subtle) !important;
  border-color: var(--admin-line) !important;
}
.scheme3-admin-monitor-table-frame :deep(.sticky-header-cell),
.scheme3-admin-monitor-table-frame :deep(.sticky-col) { background: var(--admin-subtle) !important; }
.scheme3-admin-monitor-table-frame :deep(.table-body) { background: var(--admin-surface) !important; border-color: var(--admin-line) !important; }
.scheme3-admin-monitor-table-frame :deep(th) {
  border-bottom: 1px solid var(--admin-line);
  background: var(--admin-subtle) !important;
  color: var(--admin-muted) !important;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: .59rem !important;
  font-weight: 800 !important;
  letter-spacing: .06em;
}
.scheme3-admin-monitor-table-frame :deep(td) {
  border-bottom: 1px solid var(--admin-line);
  background: var(--admin-surface) !important;
  color: var(--admin-ink) !important;
  font-size: .72rem !important;
}
.scheme3-admin-monitor-table-frame :deep(tbody tr:hover),
.scheme3-admin-monitor-table-frame :deep(tbody tr:hover td),
.scheme3-admin-monitor-table-frame :deep(tbody tr:hover .sticky-col) { background: var(--admin-subtle) !important; }
.scheme3-admin-monitor-table-frame :deep(tbody tr),
.scheme3-admin-monitor-table-frame :deep(tbody tr > td),
.scheme3-admin-monitor-table-frame :deep(thead),
.scheme3-admin-monitor-table-frame :deep(tbody) { border-color: var(--admin-line) !important; }
.scheme3-admin-monitor-empty {
  display: flex;
  min-height: 17rem;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  color: var(--admin-muted);
  text-align: center;
}
.scheme3-admin-monitor-empty-mark {
  display: inline-flex;
  width: 3rem;
  height: 3rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(30, 92, 66, .28);
  border-radius: 7px;
  background: rgba(30, 92, 66, .08);
  color: var(--admin-accent);
}
.scheme3-admin-monitor-empty h3 {
  margin: 1rem 0 .25rem;
  color: var(--admin-ink);
  font-family: Georgia, 'Times New Roman', serif;
  font-size: 1.25rem;
  font-weight: 500;
}
.scheme3-admin-monitor-empty p { margin: 0; font-size: .72rem; }
.scheme3-admin-monitor-empty-action {
  display: inline-flex;
  min-height: 2.2rem;
  align-items: center;
  gap: .4rem;
  margin-top: 1rem;
  border: 1px solid var(--admin-accent);
  border-radius: 6px;
  padding: .45rem .72rem;
  background: var(--admin-accent);
  color: #fffefa;
  font-size: .68rem;
  font-weight: 800;
}
.scheme3-admin-monitor-empty-action:hover { background: #174a35; }
.scheme3-admin-monitor-empty-action:focus-visible { outline: 2px solid rgba(30, 92, 66, .3); outline-offset: 2px; }
.scheme3-admin-monitor-pagination {
  overflow: hidden;
  border: 1px solid var(--admin-line);
  border-radius: 7px;
  background: var(--admin-surface);
}
.scheme3-admin-monitor-pagination :deep(> div) { border-color: var(--admin-line); background: transparent !important; }
.scheme3-admin-monitor-pagination :deep(button) {
  border-color: var(--admin-line) !important;
  border-radius: 5px !important;
  background: var(--admin-surface) !important;
  color: var(--admin-muted) !important;
  box-shadow: none !important;
}
.scheme3-admin-monitor-pagination :deep(button:hover:not(:disabled)) { border-color: var(--admin-accent) !important; color: var(--admin-accent) !important; }
.scheme3-admin-monitor-pagination :deep(button[aria-current='page']) { border-color: var(--admin-accent) !important; background: rgba(30, 92, 66, .1) !important; color: var(--admin-accent) !important; }
.scheme3-admin-monitor-pagination :deep(.scheme3-select-trigger) { border-color: var(--admin-line); border-radius: 5px; background: var(--admin-surface); color: var(--admin-ink); box-shadow: none; }
:global(.dark .scheme3-admin-monitor) {
  --admin-paper: #1b1b18;
  --admin-surface: #24231f;
  --admin-subtle: #2b2924;
  --admin-ink: #f4f2ec;
  --admin-muted: #aaa69a;
  --admin-line: #47443a;
  --admin-accent: #8fc2a5;
}
:global(.dark .scheme3-admin-monitor-segment.is-active) { background: var(--admin-surface); color: var(--admin-accent); }
:global(.dark .scheme3-admin-monitor-empty-action) { color: #1b1b18; }

@media (max-width: 767px) {
  .scheme3-admin-monitor-header { grid-template-columns: minmax(0, 1fr); align-items: start; gap: .7rem; }
  .scheme3-admin-monitor-tabs { width: 100%; }
  .scheme3-admin-monitor-segments { width: 100%; max-width: none; }
  .scheme3-admin-monitor-table-frame :deep(.table-wrapper) { min-height: 18rem; max-height: 60vh; }
}
</style>

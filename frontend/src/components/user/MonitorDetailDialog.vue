<template>
  <BaseDialog
    :show="show"
    :title="title"
    width="wide"
    content-class="scheme3-monitor-dialog"
    @close="$emit('close')"
  >
    <div v-if="loading" class="scheme3-monitor-detail-state py-8 text-center text-sm">
      {{ t('common.loading') }}
    </div>
    <div v-else-if="!detail" class="scheme3-monitor-detail-state py-8 text-center text-sm">
      {{ t('channelStatus.detailLoadError') }}
    </div>
    <div v-else class="scheme3-monitor-detail-table-wrap overflow-x-auto">
      <table class="scheme3-monitor-detail-table w-full text-left text-sm">
        <thead class="scheme3-monitor-detail-table-head">
          <tr class="text-xs uppercase tracking-wider">
            <th class="py-2 pr-3">{{ t('channelStatus.detailColumns.model') }}</th>
            <th class="py-2 pr-3">{{ t('channelStatus.detailColumns.latestStatus') }}</th>
            <th class="py-2 pr-3">{{ t('channelStatus.detailColumns.latestLatency') }}</th>
            <th class="py-2 pr-3">{{ t('channelStatus.detailColumns.availability7d') }}</th>
            <th class="py-2 pr-3">{{ t('channelStatus.detailColumns.availability15d') }}</th>
            <th class="py-2 pr-3">{{ t('channelStatus.detailColumns.availability30d') }}</th>
            <th class="py-2 pr-3">{{ t('channelStatus.detailColumns.avgLatency7d') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="m in detail.models"
            :key="m.model"
            class="scheme3-monitor-detail-row"
          >
            <td class="scheme3-monitor-detail-primary py-2 pr-3 font-medium">{{ m.model }}</td>
            <td class="py-2 pr-3">
              <span
                class="scheme3-monitor-status-badge inline-flex items-center px-2 py-0.5 text-[11px]"
                :class="statusClass(m.latest_status)"
              >
                {{ statusLabel(m.latest_status) }}
              </span>
            </td>
            <td class="scheme3-monitor-detail-muted py-2 pr-3">{{ formatLatency(m.latest_latency_ms) }}</td>
            <td class="scheme3-monitor-detail-muted py-2 pr-3">{{ formatPercent(m.availability_7d) }}</td>
            <td class="scheme3-monitor-detail-muted py-2 pr-3">{{ formatPercent(m.availability_15d) }}</td>
            <td class="scheme3-monitor-detail-muted py-2 pr-3">{{ formatPercent(m.availability_30d) }}</td>
            <td class="scheme3-monitor-detail-muted py-2 pr-3">{{ formatLatency(m.avg_latency_7d_ms) }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <template #footer>
      <div class="scheme3-monitor-detail-footer flex justify-end">
        <button @click="$emit('close')" class="scheme3-monitor-button">
          {{ t('channelStatus.closeDetail') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import {
  status as fetchChannelMonitorDetail,
  type MonitorStatus,
  type UserMonitorDetail,
} from '@/api/channelMonitor'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { useChannelMonitorFormat } from '@/composables/useChannelMonitorFormat'

const props = defineProps<{
  show: boolean
  monitorId: number | null
  title: string
}>()

defineEmits<{
  (e: 'close'): void
}>()

const { t } = useI18n()
const appStore = useAppStore()
const { statusLabel, formatLatency, formatPercent } = useChannelMonitorFormat()

function statusClass(status: MonitorStatus | ''): string {
  switch (status) {
    case 'operational': return 'is-operational'
    case 'degraded': return 'is-degraded'
    case 'failed': return 'is-failed'
    case 'error': return 'is-error'
    default: return 'is-unknown'
  }
}

const detail = ref<UserMonitorDetail | null>(null)
const loading = ref(false)

async function load(id: number) {
  detail.value = null
  loading.value = true
  try {
    detail.value = await fetchChannelMonitorDetail(id)
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('channelStatus.detailLoadError')))
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.show, props.monitorId] as const,
  ([show, id]) => {
    if (!show) {
      detail.value = null
      return
    }
    if (id != null) void load(id)
  },
  { immediate: true },
)
</script>

<style scoped>
:global(.scheme3-monitor-dialog) {
  border: 1px solid #d8d2c3 !important;
  border-radius: 8px !important;
  background: #fffefa !important;
  box-shadow: 0 20px 42px rgba(54,48,34,.16) !important;
}

:global(.scheme3-monitor-dialog .modal-header),
:global(.scheme3-monitor-dialog .modal-footer) {
  border-color: #d8d2c3 !important;
  background: #f8f6ef !important;
}

:global(.scheme3-monitor-dialog .modal-title) {
  color: #27251f !important;
  font-family: Georgia, 'Times New Roman', serif;
  font-weight: 500;
}

:global(.scheme3-monitor-dialog .modal-body) { background: #fffefa !important; }
.scheme3-monitor-detail-state { color: #777266 !important; }
.scheme3-monitor-detail-table-wrap { border: 1px solid #d8d2c3; border-radius: 7px; }
.scheme3-monitor-detail-table { min-width: 43rem; color: #27251f; }
.scheme3-monitor-detail-table-head { background: #f1eee6; }
.scheme3-monitor-detail-table th { border-color: #d8d2c3 !important; padding-top: .72rem; padding-bottom: .72rem; color: #777266 !important; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: .59rem; font-weight: 800; letter-spacing: .055em; }
.scheme3-monitor-detail-row { border-bottom: 1px solid rgba(216,210,195,.78); }
.scheme3-monitor-detail-table td { border-color: rgba(216,210,195,.78) !important; padding-top: .76rem; padding-bottom: .76rem; }
.scheme3-monitor-detail-primary { color: #27251f !important; }
.scheme3-monitor-detail-muted { color: #655f53 !important; }
.scheme3-monitor-status-badge { border: 1px solid currentColor; border-radius: 999px; font-weight: 800; }
.scheme3-monitor-status-badge.is-operational { border-color: rgba(30,92,66,.3); background: rgba(30,92,66,.08); color: #1e5c42; }
.scheme3-monitor-status-badge.is-degraded { border-color: rgba(183,121,31,.3); background: rgba(183,121,31,.08); color: #8b5d14; }
.scheme3-monitor-status-badge.is-failed,.scheme3-monitor-status-badge.is-error { border-color: rgba(158,77,61,.3); background: rgba(158,77,61,.08); color: #9e4d3d; }
.scheme3-monitor-status-badge.is-unknown { border-color: #d8d2c3; background: #f1eee6; color: #777266; }
.scheme3-monitor-detail-footer .scheme3-monitor-button { border: 1px solid #d8d2c3; border-radius: 7px; background: #fffefa; color: #27251f; padding: .45rem .72rem; font-size: .68rem; font-weight: 800; }

:global(.dark .scheme3-monitor-dialog) { border-color: #47443a !important; background: #24231f !important; box-shadow: 0 20px 42px rgba(0,0,0,.34) !important; }
:global(.dark .scheme3-monitor-dialog .modal-header),
:global(.dark .scheme3-monitor-dialog .modal-footer) { border-color: #47443a !important; background: #1b1b18 !important; }
:global(.dark .scheme3-monitor-dialog .modal-title),
:global(.dark .scheme3-monitor-dialog .modal-body) { color: #f4f2ec !important; }
:global(.dark .scheme3-monitor-dialog .modal-body) { background: #24231f !important; }
:global(.dark .scheme3-monitor-detail-state) { color: #aaa69a !important; }
:global(.dark .scheme3-monitor-detail-table-wrap) { border-color: #47443a; }
:global(.dark .scheme3-monitor-detail-table) { color: #f4f2ec; }
:global(.dark .scheme3-monitor-detail-table thead) { background: #2b2924; }
:global(.dark .scheme3-monitor-detail-table th) { border-color: #47443a !important; color: #aaa69a !important; }
:global(.dark .scheme3-monitor-detail-row) { border-color: rgba(71,68,58,.86); }
:global(.dark .scheme3-monitor-detail-table td) { border-color: rgba(71,68,58,.86) !important; }
:global(.dark .scheme3-monitor-detail-primary) { color: #f4f2ec !important; }
:global(.dark .scheme3-monitor-detail-muted) { color: #d4d0c6 !important; }
:global(.dark .scheme3-monitor-status-badge.is-operational) { border-color: rgba(143,194,165,.3); background: rgba(143,194,165,.1); color: #8fc2a5; }
:global(.dark .scheme3-monitor-status-badge.is-degraded) { border-color: rgba(211,164,92,.3); background: rgba(211,164,92,.1); color: #d3a45c; }
:global(.dark .scheme3-monitor-status-badge.is-failed),:global(.dark .scheme3-monitor-status-badge.is-error) { border-color: rgba(211,139,121,.3); background: rgba(211,139,121,.1); color: #d38b79; }
:global(.dark .scheme3-monitor-status-badge.is-unknown) { border-color: #47443a; background: #2b2924; color: #aaa69a; }
:global(.dark .scheme3-monitor-detail-footer .scheme3-monitor-button) { border-color: #47443a; background: #24231f; color: #f4f2ec; }

@media (max-width: 767px) {
  .scheme3-monitor-detail-table-wrap { margin-right: -.2rem; margin-left: -.2rem; }
}
</style>

<template>
  <BaseDialog
    :show="show"
    :title="title"
    width="wide"
    content-class="scheme3-monitor-dialog"
    @close="$emit('close')"
  >
    <div v-if="loading" class="scheme3-monitor-detail-state py-8 text-center text-sm text-gray-500">
      {{ t('common.loading') }}
    </div>
    <div v-else-if="!detail" class="scheme3-monitor-detail-state py-8 text-center text-sm text-gray-500">
      {{ t('channelStatus.detailLoadError') }}
    </div>
    <div v-else class="scheme3-monitor-detail-table-wrap overflow-x-auto">
      <table class="scheme3-monitor-detail-table w-full text-left text-sm">
        <thead class="border-b border-gray-200 dark:border-dark-700">
          <tr class="text-xs uppercase tracking-wider text-gray-500 dark:text-gray-400">
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
            class="border-b border-gray-100 dark:border-dark-800"
          >
            <td class="py-2 pr-3 font-medium text-gray-900 dark:text-gray-100">{{ m.model }}</td>
            <td class="py-2 pr-3">
              <span
                class="inline-flex items-center rounded-full px-2 py-0.5 text-[11px]"
                :class="statusBadgeClass(m.latest_status)"
              >
                {{ statusLabel(m.latest_status) }}
              </span>
            </td>
            <td class="py-2 pr-3 text-gray-700 dark:text-gray-300">{{ formatLatency(m.latest_latency_ms) }}</td>
            <td class="py-2 pr-3 text-gray-700 dark:text-gray-300">{{ formatPercent(m.availability_7d) }}</td>
            <td class="py-2 pr-3 text-gray-700 dark:text-gray-300">{{ formatPercent(m.availability_15d) }}</td>
            <td class="py-2 pr-3 text-gray-700 dark:text-gray-300">{{ formatPercent(m.availability_30d) }}</td>
            <td class="py-2 pr-3 text-gray-700 dark:text-gray-300">{{ formatLatency(m.avg_latency_7d_ms) }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <template #footer>
      <div class="scheme3-monitor-detail-footer flex justify-end">
        <button @click="$emit('close')" class="btn btn-secondary">
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
const { statusLabel, statusBadgeClass, formatLatency, formatPercent } = useChannelMonitorFormat()

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
.scheme3-monitor-detail-table thead { background: #f1eee6; }
.scheme3-monitor-detail-table th { border-color: #d8d2c3 !important; padding-top: .72rem; padding-bottom: .72rem; color: #777266 !important; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: .59rem; font-weight: 800; letter-spacing: .055em; }
.scheme3-monitor-detail-table td { border-color: rgba(216,210,195,.78) !important; padding-top: .76rem; padding-bottom: .76rem; color: #655f53 !important; }
.scheme3-monitor-detail-table td:first-child { color: #27251f !important; }
.scheme3-monitor-detail-table :deep(.rounded-full) { border: 1px solid currentColor; background: transparent !important; }
.scheme3-monitor-detail-footer :deep(.btn-secondary) { border-color: #d8d2c3; border-radius: 7px; background: #fffefa; color: #27251f; }

:global(.dark .scheme3-monitor-dialog) { border-color: #47443a !important; background: #24231f !important; box-shadow: 0 20px 42px rgba(0,0,0,.34) !important; }
:global(.dark .scheme3-monitor-dialog .modal-header),
:global(.dark .scheme3-monitor-dialog .modal-footer) { border-color: #47443a !important; background: #1b1b18 !important; }
:global(.dark .scheme3-monitor-dialog .modal-title),
:global(.dark .scheme3-monitor-dialog .modal-body) { color: #f4f2ec !important; }
:global(.dark .scheme3-monitor-dialog .modal-body) { background: #24231f !important; }
:global(.dark .scheme3-monitor-dialog .hover\\:bg-gray-100:hover) { background: #2b2924 !important; }
:global(.dark .scheme3-monitor-detail-state) { color: #aaa69a !important; }
:global(.dark .scheme3-monitor-detail-table-wrap) { border-color: #47443a; }
:global(.dark .scheme3-monitor-detail-table) { color: #f4f2ec; }
:global(.dark .scheme3-monitor-detail-table thead) { background: #2b2924; }
:global(.dark .scheme3-monitor-detail-table th) { border-color: #47443a !important; color: #aaa69a !important; }
:global(.dark .scheme3-monitor-detail-table td) { border-color: rgba(71,68,58,.86) !important; color: #d4d0c6 !important; }
:global(.dark .scheme3-monitor-detail-table td:first-child) { color: #f4f2ec !important; }
:global(.dark .scheme3-monitor-detail-footer :deep(.btn-secondary)) { border-color: #47443a; background: #24231f; color: #f4f2ec; }

@media (max-width: 767px) {
  .scheme3-monitor-detail-table-wrap { margin-right: -.2rem; margin-left: -.2rem; }
}
</style>

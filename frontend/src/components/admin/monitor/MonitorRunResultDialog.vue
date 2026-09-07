<template>
  <BaseDialog
    :show="show"
    :title="t('admin.channelMonitor.runResultTitle')"
    width="normal"
    content-class="scheme3-monitor-dialog"
    @close="$emit('close')"
  >
    <div class="space-y-2">
      <div
        v-for="r in results"
        :key="r.model"
        class="scheme3-monitor-result-row flex flex-col items-stretch gap-2 px-3 py-2 text-sm sm:flex-row sm:items-center sm:justify-between"
      >
        <div class="min-w-0 flex flex-col">
          <span class="scheme3-monitor-table-primary font-medium">{{ formatMonitorModel(r.model) }}</span>
          <span v-if="r.message" class="scheme3-monitor-table-muted text-xs">{{ r.message }}</span>
          <MonitorQuotaView :snapshot="r.quota" class="mt-1" />
        </div>
        <div class="flex shrink-0 items-center justify-end gap-2 sm:justify-start">
          <span
            class="scheme3-monitor-status-badge inline-flex items-center px-2 py-0.5 text-[11px]"
            :class="statusClass(r.status)"
          >
            {{ statusLabel(r.status) }}
          </span>
          <span class="scheme3-monitor-table-muted text-xs">{{ formatLatency(r.latency_ms) }} ms</span>
        </div>
      </div>
    </div>
    <template #footer>
      <div class="flex justify-end">
        <button @click="$emit('close')" class="scheme3-monitor-button is-primary">
          {{ t('common.close') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { CheckResult, MonitorStatus } from '@/api/admin/channelMonitor'
import BaseDialog from '@/components/common/BaseDialog.vue'
import MonitorQuotaView from '@/components/common/MonitorQuotaView.vue'
import { useChannelMonitorFormat } from '@/composables/useChannelMonitorFormat'

defineProps<{
  show: boolean
  results: CheckResult[]
}>()

defineEmits<{
  (e: 'close'): void
}>()

const { t } = useI18n()
const { statusLabel, formatLatency, formatMonitorModel } = useChannelMonitorFormat()

function statusClass(status: MonitorStatus): string {
  switch (status) {
    case 'operational': return 'is-operational'
    case 'degraded': return 'is-degraded'
    case 'failed': return 'is-failed'
    default: return 'is-unknown'
  }
}
</script>

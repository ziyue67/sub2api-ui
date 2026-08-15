<template>
  <div class="scheme3-monitor-model-cell flex items-center gap-2">
    <span class="scheme3-monitor-model-name text-sm">{{ row.primary_model }}</span>
      <HelpTooltip teleport-class="scheme3-monitor-tooltip">
      <template #trigger>
        <span
          class="scheme3-monitor-status-badge inline-flex items-center px-2 py-0.5 text-[11px] font-medium"
          :class="statusClass(row.primary_status)"
        >
          {{ statusLabel(row.primary_status) }}
        </span>
      </template>
      <div class="scheme3-monitor-tooltip-content space-y-2">
        <div class="scheme3-monitor-tooltip-title text-xs font-semibold">
          {{ row.primary_model }}
          <span
            class="scheme3-monitor-status-badge ml-1 inline-flex items-center px-1.5 py-0.5 text-[10px] font-medium"
            :class="statusClass(row.primary_status)"
          >
            {{ statusLabel(row.primary_status) }}
          </span>
        </div>
        <div v-if="(row.extra_models?.length ?? 0) === 0" class="scheme3-monitor-tooltip-muted text-[11px]">
          {{ t('monitorCommon.extraModelsEmpty') }}
        </div>
        <div v-else class="space-y-1">
          <div class="scheme3-monitor-tooltip-muted text-[11px] font-semibold uppercase tracking-wide">
            {{ t('monitorCommon.extraModelsHeader') }}
          </div>
          <table class="w-full text-left text-[11px]">
            <thead>
              <tr class="scheme3-monitor-tooltip-muted">
                <th class="py-0.5 pr-2 font-medium">{{ t('admin.channelMonitor.columns.primaryModel') }}</th>
                <th class="py-0.5 pr-2 font-medium">{{ t('admin.channelMonitor.columns.actions') }}</th>
                <th class="py-0.5 font-medium">{{ t('admin.channelMonitor.columns.latency') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="m in (row.extra_models_status || [])" :key="m.model">
                <td class="scheme3-monitor-tooltip-primary py-0.5 pr-2">{{ m.model }}</td>
                <td class="py-0.5 pr-2">
                  <span
                    class="scheme3-monitor-status-badge inline-flex items-center px-1.5 py-0.5 text-[10px]"
                    :class="statusClass(m.status)"
                  >
                    {{ statusLabel(m.status) }}
                  </span>
                </td>
                <td class="scheme3-monitor-tooltip-primary py-0.5">{{ formatLatency(m.latency_ms) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </HelpTooltip>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { ChannelMonitor, MonitorStatus } from '@/api/admin/channelMonitor'
import HelpTooltip from '@/components/common/HelpTooltip.vue'
import { useChannelMonitorFormat } from '@/composables/useChannelMonitorFormat'

defineProps<{
  row: ChannelMonitor
}>()

const { t } = useI18n()
const { statusLabel, formatLatency } = useChannelMonitorFormat()

function statusClass(status: MonitorStatus | ''): string {
  switch (status) {
    case 'operational': return 'is-operational'
    case 'degraded': return 'is-degraded'
    case 'failed': return 'is-failed'
    case 'error': return 'is-error'
    default: return 'is-unknown'
  }
}
</script>

<template>
  <div class="scheme3-monitor-model-cell flex flex-col gap-0.5">
    <div class="flex items-center gap-2">
      <!-- Pure quota monitors use "quota" as a data-source placeholder. -->
      <span class="scheme3-monitor-model-name text-sm">{{ formatMonitorModel(row.primary_model) }}</span>
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
          {{ formatMonitorModel(row.primary_model) }}
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
    <!-- 配额模式监控：主模型行内联展示最新用量/余额快照（管理端不受用户端开关限制） -->
    <MonitorQuotaView :snapshot="row.latest_quota" />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { ChannelMonitor, MonitorStatus } from '@/api/admin/channelMonitor'
import HelpTooltip from '@/components/common/HelpTooltip.vue'
import MonitorQuotaView from '@/components/common/MonitorQuotaView.vue'
import { useChannelMonitorFormat } from '@/composables/useChannelMonitorFormat'

defineProps<{
  row: ChannelMonitor
}>()

const { t } = useI18n()
const { statusLabel, formatLatency, formatMonitorModel } = useChannelMonitorFormat()

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

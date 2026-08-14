<template>
  <button
    type="button"
    class="scheme3-monitor-card group text-left p-5 rounded-2xl min-h-[280px] w-full bg-white/70 backdrop-blur-xl border border-gray-200/80 shadow-card dark:bg-dark-800/60 dark:border-dark-700/70 hover:-translate-y-1 hover:shadow-card-hover dark:hover:border-primary-500/30 hover:border-gray-300 transition-all duration-300 ease-out flex flex-col"
    @click="emit('click')"
  >
    <!-- Header: icon + name/model + status chip -->
    <div class="scheme3-monitor-card-header flex items-start gap-3">
      <span
        class="scheme3-monitor-provider-tile w-9 h-9 rounded-xl ring-1 ring-black/5 dark:ring-white/10 grid place-items-center flex-shrink-0"
        :class="[providerGradient(item.provider), providerTintClass]"
      >
        <ProviderIcon :provider="item.provider" :size="20" />
      </span>
      <div class="flex-1 min-w-0">
        <div class="scheme3-monitor-card-name text-base font-semibold truncate text-gray-900 dark:text-gray-100">
          {{ item.name }}
        </div>
        <div class="mt-0.5 flex items-center gap-1.5 min-w-0">
          <span
            class="scheme3-monitor-provider-badge inline-flex items-center rounded-md px-1.5 py-0.5 text-[10px] font-medium flex-shrink-0"
            :class="providerBadgeClass(item.provider)"
          >
            {{ providerLabel(item.provider) }}
          </span>
          <span class="scheme3-monitor-model font-mono text-xs truncate text-gray-500 dark:text-gray-400">
            {{ item.primary_model }}
          </span>
          <span
            v-if="item.group_name"
            class="scheme3-monitor-group inline-flex items-center rounded-md px-1.5 py-0.5 text-[10px] font-medium bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300 flex-shrink-0"
          >
            {{ item.group_name }}
          </span>
        </div>
      </div>
      <span
        class="scheme3-monitor-status px-2.5 py-1 rounded-full text-xs font-semibold flex-shrink-0"
        :class="statusBadgeClass(item.primary_status)"
      >
        {{ statusLabel(item.primary_status) }}
      </span>
    </div>

    <!-- Metrics -->
    <MonitorMetricPair
      class="scheme3-monitor-card-metrics"
      primary-icon="bolt"
      :primary-label="t('monitorCommon.dialogLatency')"
      :primary-value="formatLatency(item.primary_latency_ms)"
      primary-unit="ms"
      secondary-icon="globe"
      :secondary-label="t('monitorCommon.endpointPing')"
      :secondary-value="formatLatency(item.primary_ping_latency_ms)"
      secondary-unit="ms"
    />

    <!-- Divider -->
    <div class="scheme3-monitor-divider mt-4 border-t border-gray-100 dark:border-dark-700/60"></div>

    <!-- Availability row -->
    <MonitorAvailabilityRow
      :window-label="availabilityLabel"
      :value="availabilityValue"
      :samples-label="extraModelsCountLabel"
    />

    <!-- Timeline -->
    <MonitorTimeline
      :buckets="item.timeline"
      :countdown-seconds="countdownSeconds"
    />
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { UserMonitorView } from '@/api/channelMonitor'
import {
  useChannelMonitorFormat,
  providerGradient,
} from '@/composables/useChannelMonitorFormat'
import ProviderIcon from './ProviderIcon.vue'
import MonitorMetricPair from './MonitorMetricPair.vue'
import MonitorAvailabilityRow from './MonitorAvailabilityRow.vue'
import MonitorTimeline from './MonitorTimeline.vue'

const PROVIDER_TINT: Record<string, string> = {
  openai: 'text-emerald-600 dark:text-emerald-300',
  anthropic: 'text-orange-600 dark:text-orange-300',
  gemini: 'text-sky-600 dark:text-sky-300',
  grok: 'text-zinc-700 dark:text-zinc-200',
}

const props = defineProps<{
  item: UserMonitorView
  window: '7d' | '15d' | '30d'
  availabilityValue: number | null
  countdownSeconds: number
}>()

const emit = defineEmits<{
  (e: 'click'): void
}>()

const { t } = useI18n()
const {
  statusLabel,
  statusBadgeClass,
  providerLabel,
  providerBadgeClass,
  formatLatency,
} = useChannelMonitorFormat()

const providerTintClass = computed(() =>
  PROVIDER_TINT[props.item.provider] ?? 'text-gray-500 dark:text-gray-300'
)

const availabilityLabel = computed(() => {
  const win = t(`channelStatus.windowTab.${props.window}`)
  return `${t('monitorCommon.availabilityPrefix')} · ${win}`
})

const extraModelsCountLabel = computed(() => {
  const count = props.item.extra_models?.length ?? 0
  if (count === 0) return undefined
  return t('monitorCommon.extraModelsCount', { n: count })
})
</script>

<style scoped>
.scheme3-monitor-card {
  border: 1px solid #d8d2c3 !important;
  border-radius: 8px !important;
  background: #fffefa !important;
  box-shadow: 0 10px 24px rgba(54,48,34,.07) !important;
  color: #27251f;
  transition: transform 160ms ease, box-shadow 160ms ease, border-color 160ms ease, background-color 160ms ease;
}

.scheme3-monitor-card:hover {
  border-color: rgba(30,92,66,.32) !important;
  background: #fffefa !important;
  box-shadow: 0 15px 28px rgba(54,48,34,.11) !important;
  transform: translateY(-2px);
}

.scheme3-monitor-card:active { transform: translateY(0) scale(.995); }
.scheme3-monitor-provider-tile { border-radius: 7px !important; background: #f1eee6 !important; box-shadow: inset 0 0 0 1px #d8d2c3; }
.scheme3-monitor-card-name { color: #27251f !important; }
.scheme3-monitor-model { color: #777266 !important; }
.scheme3-monitor-group { border: 1px solid #d8d2c3; background: #f8f6ef !important; color: #655f53 !important; }
.scheme3-monitor-provider-badge { border: 1px solid currentColor; border-radius: 999px !important; background: transparent !important; }
.scheme3-monitor-status { border: 1px solid currentColor; border-radius: 999px !important; background: transparent !important; }
.scheme3-monitor-divider { border-color: #d8d2c3 !important; }

:global(.dark .scheme3-monitor-card) {
  border-color: #47443a !important;
  background: #24231f !important;
  color: #f4f2ec;
  box-shadow: 0 14px 30px rgba(0,0,0,.24) !important;
}

:global(.dark .scheme3-monitor-card:hover) {
  border-color: rgba(143,194,165,.35) !important;
  background: #24231f !important;
  box-shadow: 0 18px 34px rgba(0,0,0,.3) !important;
}

:global(.dark .scheme3-monitor-provider-tile) { background: #2b2924 !important; box-shadow: inset 0 0 0 1px #47443a; }
:global(.dark .scheme3-monitor-card-name) { color: #f4f2ec !important; }
:global(.dark .scheme3-monitor-model) { color: #aaa69a !important; }
:global(.dark .scheme3-monitor-group) { border-color: #47443a; background: #2b2924 !important; color: #d4d0c6 !important; }
:global(.dark .scheme3-monitor-divider) { border-color: #47443a !important; }
:global(.dark .scheme3-monitor-status.bg-emerald-100) { border-color: #8fc2a5; color: #8fc2a5; }
:global(.dark .scheme3-monitor-status.bg-amber-100) { border-color: #d3a55a; color: #d3a55a; }
:global(.dark .scheme3-monitor-status.bg-red-100) { border-color: #d38b79; color: #d38b79; }

@media (max-width: 560px) {
  .scheme3-monitor-card { min-height: 0; padding: 1rem !important; }
  .scheme3-monitor-status { padding-right: .45rem !important; padding-left: .45rem !important; font-size: .61rem !important; }
}
</style>

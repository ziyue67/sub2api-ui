<template>
  <section class="scheme3-monitor-hero py-3 md:py-4">
    <div class="scheme3-monitor-controls flex items-center justify-end gap-3 flex-wrap">
      <div
        role="tablist"
        class="scheme3-monitor-window-tabs inline-flex p-0.5 rounded-xl bg-gray-100 dark:bg-dark-800 border border-gray-200/60 dark:border-dark-700/60 text-xs"
      >
        <button
          v-for="opt in windowOptions"
          :key="opt.value"
          type="button"
          role="tab"
          :aria-selected="window === opt.value"
          class="scheme3-monitor-window-tab px-3 py-1 rounded-lg transition-colors"
          :class="window === opt.value
            ? 'bg-white dark:bg-dark-700 shadow-sm text-gray-900 dark:text-white font-semibold'
            : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'"
          @click="emit('update:window', opt.value)"
        >
          {{ opt.label }}
        </button>
      </div>

      <span
        class="scheme3-monitor-overall inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold tracking-wider uppercase"
        :class="overallChipClass"
      >
        <span
          class="w-1.5 h-1.5 rounded-full mr-1.5"
          :class="overallDotClass"
        ></span>
        {{ overallLabel }}
      </span>

      <button
        type="button"
        class="scheme3-monitor-refresh h-8 w-8 rounded-lg flex items-center justify-center text-gray-500 hover:text-gray-700 hover:bg-gray-100 dark:text-gray-400 dark:hover:text-gray-200 dark:hover:bg-dark-700 transition-colors disabled:opacity-50"
        :disabled="loading"
        :title="t('common.refresh')"
        @click="emit('refresh')"
      >
        <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
      </button>

      <div v-if="autoRefresh" class="scheme3-monitor-auto-refresh">
        <AutoRefreshButton
          :enabled="autoRefresh.enabled.value"
          :interval-seconds="autoRefresh.intervalSeconds.value"
          :countdown="autoRefresh.countdown.value"
          :intervals="autoRefresh.intervals"
          @update:enabled="autoRefresh.setEnabled"
          @update:interval="autoRefresh.setInterval"
        />
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import AutoRefreshButton from '@/components/common/AutoRefreshButton.vue'
export type MonitorWindow = '7d' | '15d' | '30d'
export type OverallStatus = 'operational' | 'degraded'

const props = defineProps<{
  overallStatus: OverallStatus
  intervalSeconds: number
  window: MonitorWindow
  loading: boolean
  autoRefresh?: {
    enabled: { value: boolean }
    intervalSeconds: { value: number }
    countdown: { value: number }
    intervals: readonly number[]
    setEnabled: (v: boolean) => void
    setInterval: (v: number) => void
  }
}>()

const emit = defineEmits<{
  (e: 'update:window', value: MonitorWindow): void
  (e: 'refresh'): void
}>()

const { t } = useI18n()

const windowOptions = computed<{ value: MonitorWindow; label: string }[]>(() => [
  { value: '7d', label: t('channelStatus.windowTab.7d') },
  { value: '15d', label: t('channelStatus.windowTab.15d') },
  { value: '30d', label: t('channelStatus.windowTab.30d') },
])

const overallLabel = computed(() => t(`channelStatus.overall.${props.overallStatus}`))

const overallChipClass = computed(() => {
  switch (props.overallStatus) {
    case 'operational':
      return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300'
    case 'degraded':
    default:
      return 'bg-amber-100 text-amber-700 dark:bg-amber-500/15 dark:text-amber-300'
  }
})

const overallDotClass = computed(() => {
  switch (props.overallStatus) {
    case 'operational':
      return 'bg-emerald-500 animate-pulse'
    case 'degraded':
    default:
      return 'bg-amber-500 animate-pulse'
  }
})

</script>

<style scoped>
.scheme3-monitor-hero { color: #27251f; }
.scheme3-monitor-controls { min-height: 2.5rem; }
.scheme3-monitor-window-tabs { border-color: #d8d2c3; background: #f1eee6; }
.scheme3-monitor-window-tab { min-height: 1.95rem; color: #777266; font-weight: 700; }
.scheme3-monitor-window-tab:hover { color: #27251f; }
.scheme3-monitor-window-tab[aria-selected='true'] { background: #fffefa; color: #1e5c42; box-shadow: 0 2px 6px rgba(54,48,34,.08); }
.scheme3-monitor-overall { border: 1px solid rgba(30,92,66,.2); background: rgba(30,92,66,.07); color: #1e5c42; letter-spacing: .06em; }
.scheme3-monitor-refresh { border: 1px solid #d8d2c3; background: #fffefa; color: #777266; }
.scheme3-monitor-refresh:hover { border-color: rgba(30,92,66,.25); background: #f1eee6; color: #1e5c42; }
.scheme3-monitor-auto-refresh :deep(button) { min-height: 2rem; border-color: #d8d2c3; border-radius: 7px; background: #fffefa; color: #777266; box-shadow: none; }
.scheme3-monitor-auto-refresh :deep(button:hover) { background: #f1eee6; color: #27251f; }
.scheme3-monitor-auto-refresh :deep(.absolute) { border-color: #d8d2c3; border-radius: 7px; background: #fffefa; box-shadow: 0 14px 28px rgba(54,48,34,.14); }
.scheme3-monitor-auto-refresh :deep(.absolute button:hover) { background: #f1eee6; }

:global(.dark .scheme3-monitor-hero) { color: #f4f2ec; }
:global(.dark .scheme3-monitor-window-tabs) { border-color: #47443a; background: #2b2924; }
:global(.dark .scheme3-monitor-window-tab) { color: #aaa69a; }
:global(.dark .scheme3-monitor-window-tab:hover) { color: #f4f2ec; }
:global(.dark .scheme3-monitor-window-tab[aria-selected='true']) { background: #24231f; color: #8fc2a5; box-shadow: 0 2px 7px rgba(0,0,0,.18); }
:global(.dark .scheme3-monitor-overall) { border-color: rgba(143,194,165,.28); background: rgba(143,194,165,.1); color: #8fc2a5; }
:global(.dark .scheme3-monitor-refresh) { border-color: #47443a; background: #24231f; color: #aaa69a; }
:global(.dark .scheme3-monitor-refresh:hover) { border-color: rgba(143,194,165,.3); background: #2b2924; color: #8fc2a5; }
:global(.dark .scheme3-monitor-auto-refresh :deep(button)) { border-color: #47443a; background: #24231f; color: #aaa69a; }
:global(.dark .scheme3-monitor-auto-refresh :deep(button:hover)) { background: #2b2924; color: #f4f2ec; }
:global(.dark .scheme3-monitor-auto-refresh :deep(.absolute)) { border-color: #47443a; background: #24231f; box-shadow: 0 16px 30px rgba(0,0,0,.28); }
:global(.dark .scheme3-monitor-auto-refresh :deep(.absolute button:hover)) { background: #2b2924; }

@media (max-width: 560px) {
  .scheme3-monitor-controls { align-items: stretch; justify-content: stretch; gap: .45rem; }
  .scheme3-monitor-window-tabs { flex: 1 1 auto; justify-content: space-between; }
  .scheme3-monitor-window-tab { flex: 1 1 0; padding-right: .45rem; padding-left: .45rem; }
  .scheme3-monitor-overall { justify-content: center; }
  .scheme3-monitor-auto-refresh { flex: 1 1 auto; }
  .scheme3-monitor-auto-refresh :deep(button) { width: 100%; justify-content: center; }
}
</style>

<template>
  <section class="scheme3-monitor-hero">
    <div class="scheme3-monitor-controls">
      <div role="tablist" class="scheme3-monitor-window-tabs">
        <button
          v-for="opt in windowOptions"
          :key="opt.value"
          type="button"
          role="tab"
          :aria-selected="window === opt.value"
          class="scheme3-monitor-window-tab"
          :class="{ 'is-active': window === opt.value }"
          @click="emit('update:window', opt.value)"
        >
          {{ opt.label }}
        </button>
      </div>

      <span
        class="scheme3-monitor-overall"
        :class="overallToneClass"
      >
        <span class="scheme3-monitor-overall-dot" aria-hidden="true"></span>
        {{ overallLabel }}
      </span>

      <button
        type="button"
        class="scheme3-monitor-refresh"
        :disabled="loading"
        :title="t('common.refresh')"
        @click="emit('refresh')"
      >
        <Icon name="refresh" size="md" :class="{ 'is-spinning': loading }" />
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

const overallToneClass = computed(() => `is-${props.overallStatus === 'operational' ? 'operational' : 'degraded'}`)

</script>

<style scoped>
.scheme3-monitor-hero { color: var(--monitor-ink, #27251f); padding: .75rem 0 1rem; }
.scheme3-monitor-controls { display: flex; min-height: 2.5rem; align-items: center; justify-content: flex-end; flex-wrap: wrap; gap: .7rem; }
.scheme3-monitor-window-tabs { display: inline-flex; gap: 2px; border: 1px solid var(--monitor-line, #d8d2c3); border-radius: 7px; padding: 2px; background: var(--monitor-subtle, #f1eee6); }
.scheme3-monitor-window-tab { min-height: 1.95rem; border: 1px solid transparent; border-radius: 5px; padding: .28rem .72rem; background: transparent; color: var(--monitor-muted, #777266); font-size: .66rem; font-weight: 800; transition: color 150ms ease, background-color 150ms ease, border-color 150ms ease; }
.scheme3-monitor-window-tab:hover { color: var(--monitor-ink, #27251f); }
.scheme3-monitor-window-tab.is-active { border-color: rgba(30, 92, 66, .16); background: var(--monitor-card, #fffefa); color: var(--monitor-accent, #1e5c42); box-shadow: 0 2px 6px rgba(54, 48, 34, .08); }
.scheme3-monitor-overall { display: inline-flex; align-items: center; gap: .4rem; border: 1px solid rgba(30, 92, 66, .2); border-radius: 999px; padding: .42rem .65rem; background: rgba(30, 92, 66, .07); color: var(--monitor-accent, #1e5c42); font-size: .62rem; font-weight: 800; letter-spacing: .06em; text-transform: uppercase; }
.scheme3-monitor-overall.is-degraded { border-color: rgba(183, 121, 31, .28); background: rgba(183, 121, 31, .08); color: var(--monitor-amber, #b7791f); }
.scheme3-monitor-overall-dot { width: .4rem; height: .4rem; border-radius: 999px; background: currentColor; animation: scheme3-monitor-status-pulse 1.8s ease-in-out infinite; }
.scheme3-monitor-refresh { display: inline-flex; width: 2rem; height: 2rem; align-items: center; justify-content: center; border: 1px solid var(--monitor-line, #d8d2c3); border-radius: 7px; background: var(--monitor-card, #fffefa); color: var(--monitor-muted, #777266); transition: color 150ms ease, background-color 150ms ease, border-color 150ms ease; }
.scheme3-monitor-refresh:hover:not(:disabled) { border-color: rgba(30, 92, 66, .25); background: var(--monitor-subtle, #f1eee6); color: var(--monitor-accent, #1e5c42); }
.scheme3-monitor-refresh:disabled { cursor: not-allowed; opacity: .5; }
.scheme3-monitor-refresh .is-spinning { animation: scheme3-monitor-refresh-spin 1.2s linear infinite; }
.scheme3-monitor-auto-refresh { flex: 0 0 auto; }

@keyframes scheme3-monitor-refresh-spin { to { transform: rotate(360deg); } }
@keyframes scheme3-monitor-status-pulse { 0%, 100% { opacity: .55; box-shadow: 0 0 0 0 currentColor; } 50% { opacity: 1; box-shadow: 0 0 0 .22rem transparent; } }

:global(.dark .scheme3-monitor-hero) { color: #f4f2ec; }
:global(.dark .scheme3-monitor-window-tabs) { border-color: #47443a; background: #2b2924; }
:global(.dark .scheme3-monitor-window-tab) { color: #aaa69a; }
:global(.dark .scheme3-monitor-window-tab:hover) { color: #f4f2ec; }
:global(.dark .scheme3-monitor-window-tab.is-active) { border-color: rgba(143, 194, 165, .2); background: #24231f; color: #8fc2a5; box-shadow: 0 2px 7px rgba(0, 0, 0, .18); }
:global(.dark .scheme3-monitor-overall) { border-color: rgba(143, 194, 165, .28); background: rgba(143, 194, 165, .1); color: #8fc2a5; }
:global(.dark .scheme3-monitor-overall.is-degraded) { border-color: rgba(211, 165, 90, .3); background: rgba(211, 165, 90, .1); color: #d3a55a; }
:global(.dark .scheme3-monitor-refresh) { border-color: #47443a; background: #24231f; color: #aaa69a; }
:global(.dark .scheme3-monitor-refresh:hover:not(:disabled)) { border-color: rgba(143, 194, 165, .3); background: #2b2924; color: #8fc2a5; }

@media (max-width: 560px) {
  .scheme3-monitor-hero { padding-top: .55rem; }
  .scheme3-monitor-controls { align-items: stretch; justify-content: stretch; gap: .45rem; }
  .scheme3-monitor-window-tabs { flex: 1 1 auto; justify-content: space-between; }
  .scheme3-monitor-window-tab { flex: 1 1 0; padding-right: .45rem; padding-left: .45rem; }
  .scheme3-monitor-overall { justify-content: center; }
  .scheme3-monitor-auto-refresh { flex: 1 1 auto; }
  .scheme3-monitor-auto-refresh :deep(.scheme3-auto-refresh-trigger) { width: 100%; justify-content: center; }
}
</style>

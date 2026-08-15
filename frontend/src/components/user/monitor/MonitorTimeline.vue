<template>
  <div class="scheme3-monitor-timeline">
    <div class="scheme3-monitor-timeline-label">
      <span>{{ t('monitorCommon.history60pts', { n: length }) }}</span>
      <span>{{ t('monitorCommon.nextUpdateIn', { n: countdownSeconds }) }}</span>
    </div>

    <div v-if="maintenance" class="scheme3-monitor-maintenance">
      {{ t('monitorCommon.maintenancePaused') }}
    </div>
    <div v-else class="scheme3-monitor-timeline-bars">
      <div
        v-for="(bar, idx) in displayBars"
        :key="idx"
        class="scheme3-monitor-timeline-bar"
        :class="bar.toneClass"
        :style="{ height: bar.heightPct + '%' }"
        :title="bar.title"
      ></div>
    </div>

    <div class="scheme3-monitor-timeline-axis">
      <span>{{ t('monitorCommon.past') }}</span>
      <span>{{ t('monitorCommon.now') }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { MonitorTimelinePoint } from '@/api/channelMonitor'
import { useChannelMonitorFormat } from '@/composables/useChannelMonitorFormat'

const props = withDefaults(defineProps<{
  buckets?: MonitorTimelinePoint[]
  countdownSeconds: number
  length?: number
  maintenance?: boolean
}>(), {
  buckets: () => [],
  length: 60,
  maintenance: false,
})

const { t } = useI18n()
const { statusLabel, formatLatency, formatRelativeTime } = useChannelMonitorFormat()

interface Bar {
  toneClass: string
  heightPct: number
  title: string
}

// 4 级高度 + 颜色双重编码：高=好+绿，短=坏+红，灰=未测试。
// 长绿(正常) > 中黄(降级) > 短红(失败/系统错误) > 很短灰(未测试)。
const STATUS_HEIGHT: Record<string, number> = {
  operational: 100,
  degraded: 65,
  failed: 35,
  error: 35,
  empty: 15,
}

const STATUS_TONE: Record<string, string> = {
  operational: 'is-operational',
  degraded: 'is-degraded',
  failed: 'is-failed',
  error: 'is-error',
  empty: 'is-empty',
}

const displayBars = computed<Bar[]>(() => {
  // Real points come newest-first; convert to oldest-first so the rightmost
  // bar represents "now". Pad the left with empty placeholders to keep the
  // bar count stable at `length`.
  const real = [...(props.buckets ?? [])]
    .slice(0, props.length)
    .reverse()

  const padCount = Math.max(0, props.length - real.length)
  const bars: Bar[] = []

  for (let i = 0; i < padCount; i += 1) {
    bars.push({
      toneClass: STATUS_TONE.empty,
      heightPct: STATUS_HEIGHT.empty,
      title: '',
    })
  }

  for (const point of real) {
    const status = point.status as keyof typeof STATUS_HEIGHT
    const toneClass = STATUS_TONE[status] ?? STATUS_TONE.empty
    const heightPct = STATUS_HEIGHT[status] ?? STATUS_HEIGHT.empty
    const latency = formatLatency(point.latency_ms)
    const relative = formatRelativeTime(point.checked_at)
    const label = statusLabel(point.status)
    bars.push({
      toneClass,
      heightPct,
      title: `${relative} · ${label} · ${latency}ms`,
    })
  }

  return bars
})
</script>

<style scoped>
.scheme3-monitor-timeline { margin-top: 1rem; border-top: 1px solid var(--monitor-line, #d8d2c3); padding-top: .75rem; }
.scheme3-monitor-timeline-label,
.scheme3-monitor-timeline-axis { display: flex; align-items: center; justify-content: space-between; color: var(--monitor-soft, #a49e90); font-size: .58rem; font-weight: 800; letter-spacing: .07em; line-height: 1.2; text-transform: uppercase; }
.scheme3-monitor-timeline-bars { display: flex; height: 1.25rem; align-items: flex-end; gap: 2px; width: 100%; margin-top: .48rem; }
.scheme3-monitor-timeline-bar { min-width: 0; flex: 1 1 0; border-radius: 2px 2px 1px 1px; transition: opacity 120ms ease, transform 120ms ease; }
.scheme3-monitor-timeline-bar:hover { opacity: .78; transform: translateY(-1px); }
.scheme3-monitor-timeline-bar.is-operational { background: var(--monitor-accent, #1e5c42); }
.scheme3-monitor-timeline-bar.is-degraded { background: var(--monitor-amber, #b7791f); }
.scheme3-monitor-timeline-bar.is-failed,
.scheme3-monitor-timeline-bar.is-error { background: var(--monitor-danger, #9e4d3d); }
.scheme3-monitor-timeline-bar.is-empty { background: #ddd8ca; }
.scheme3-monitor-maintenance { display: flex; height: 1.25rem; align-items: center; justify-content: center; width: 100%; margin-top: .48rem; border: 1px dashed var(--monitor-line, #d8d2c3); border-radius: 5px; color: var(--monitor-muted, #777266); font-size: .58rem; font-weight: 800; letter-spacing: .07em; text-transform: uppercase; }
.scheme3-monitor-timeline-axis { margin-top: .3rem; font-size: .52rem; }

:global(.dark .scheme3-monitor-timeline) { border-color: #47443a; }
:global(.dark .scheme3-monitor-timeline-label),
:global(.dark .scheme3-monitor-timeline-axis) { color: #827e72; }
:global(.dark .scheme3-monitor-maintenance) { border-color: #47443a; color: #aaa69a; }
:global(.dark .scheme3-monitor-timeline-bar.is-operational) { background: #8fc2a5; }
:global(.dark .scheme3-monitor-timeline-bar.is-degraded) { background: #d3a55a; }
:global(.dark .scheme3-monitor-timeline-bar.is-failed),
:global(.dark .scheme3-monitor-timeline-bar.is-error) { background: #d38b79; }
:global(.dark .scheme3-monitor-timeline-bar.is-empty) { background: #47443a; }
</style>

<template>
  <section
    class="scheme3-v2-panel scheme3-v2-trend-panel flex min-h-[360px] flex-col overflow-hidden"
  >
    <div class="scheme3-v2-panel-header mb-4 flex shrink-0 flex-wrap items-start justify-between gap-3">
      <div class="min-w-0">
        <h2 class="scheme3-v2-panel-title flex items-center gap-2">
          <span class="scheme3-v2-panel-icon inline-flex h-4 w-4" aria-hidden="true">
            <Icon name="chart" size="sm" />
          </span>
          {{ t('channelMonitorV2.chart.title') }}
        </h2>
        <p class="scheme3-v2-panel-description mt-0.5 text-xs">
          {{ t('channelMonitorV2.chart.description') }}
        </p>
      </div>
      <div class="scheme3-v2-panel-tools flex w-full min-w-0 flex-wrap items-center justify-end gap-2 text-xs sm:w-auto">
        <span class="flex shrink-0 items-center gap-1">
          <span class="scheme3-v2-legend-dot is-danger"></span>{{ t('channelMonitorV2.chart.errorLegend') }}
        </span>
        <span class="flex shrink-0 items-center gap-1">
          <span class="scheme3-v2-legend-dot is-accent"></span>{{ t('channelMonitorV2.chart.cacheLegend') }}
        </span>
        <span class="flex shrink-0 items-center gap-1">
          <span class="scheme3-v2-legend-dot is-amber"></span>{{ t('channelMonitorV2.chart.ttftLegend') }}
        </span>
        <span class="scheme3-v2-panel-badge shrink-0">{{ bucketLabel }}</span>
        <button
          type="button"
          class="scheme3-v2-panel-action inline-flex shrink-0 items-center px-2 py-1 text-[11px] font-semibold disabled:opacity-50"
          :disabled="!zoomed"
          @click="resetChartZoom"
        >
          {{ t('channelMonitorV2.chart.resetZoom') }}
        </button>
      </div>
    </div>
    <div class="scheme3-v2-panel-body min-h-0 flex-1">
      <div v-if="loading" class="flex h-[280px] items-center justify-center sm:h-[300px]">
        <div class="scheme3-v2-trend-loading animate-pulse text-sm">{{ t('common.loading') }}</div>
      </div>
      <div
        v-else-if="chartData"
        ref="chartRef"
        class="h-[280px] sm:h-[300px]"
        @wheel="onChartWheel"
      >
        <Line :data="chartData" :options="chartOptions" />
      </div>
      <div v-else class="scheme3-v2-empty flex h-[280px] items-center justify-center sm:h-[300px]" role="status">
        <span class="scheme3-v2-empty-mark" aria-hidden="true">
          <Icon name="chart" size="md" />
        </span>
        <div class="scheme3-v2-empty-copy">
          <strong>{{ t('channelMonitorV2.chart.emptyTitle') }}</strong>
          <p>{{ t('channelMonitorV2.empty.description') }}</p>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { computed, ref, watch } from 'vue'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler,
} from 'chart.js'
import { Line } from 'vue-chartjs'
import Icon from '@/components/icons/Icon.vue'
import type { MonitorCoverage, MonitorMetric, MonitorHealth } from '@/api/channelMonitorV2'
import { formatMonitorMs, formatMonitorPercent } from '@/features/channel-monitor-v2/monitorFormat'
import {
  applyWheelZoom,
  clientXRatio,
  isZoomed,
  resetZoom,
  sliceByZoom,
  type ZoomState,
} from '@/features/channel-monitor-v2/monitorZoom'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend, Filler)
const { t, locale } = useI18n()

const props = defineProps<{
  trend: Array<{ bucket_start: string; metrics: MonitorMetric; health: MonitorHealth }>
  coverage: MonitorCoverage | null
  loading?: boolean
}>()

const chartRef = ref<HTMLElement | null>(null)
const zoom = ref<ZoomState>(resetZoom())
const zoomed = computed(() => isZoomed(zoom.value))

const isDark = computed(() =>
  typeof document !== 'undefined' && document.documentElement.classList.contains('dark')
)

const bucketLabel = computed(() => {
  const seconds = props.coverage?.bucket_seconds || 60
  const minutes = seconds / 60
  if (minutes < 60) return t('channelMonitorV2.bucket.minutes', { count: minutes })
  const hours = minutes / 60
  if (hours < 24) return t('channelMonitorV2.bucket.hours', { count: hours })
  return t('channelMonitorV2.bucket.days', { count: hours / 24 })
})

const chartData = computed(() => {
  const points = visibleTrend.value
  if (!points.length) return null
  const labels = points.map((p) =>
    new Intl.DateTimeFormat(locale.value || undefined, {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    }).format(new Date(p.bucket_start))
  )
  const errorRates = smoothTrend(points.map((p) => (p.metrics.error_rate || 0) * 100))
  const cacheRates = smoothTrend(points.map((p) => (p.metrics.cache_rate || 0) * 100))
  const ttftP50 = smoothTrend(points.map((p) => p.metrics.ttft?.p50_ms ?? null))
  return {
    labels,
    datasets: [
      {
        label: t('channelMonitorV2.chart.errorDataset'),
        data: errorRates,
        borderColor: '#9e4d3d',
        backgroundColor: 'rgba(158, 77, 61, 0.10)',
        yAxisID: 'yPct',
        tension: 0.4,
        cubicInterpolationMode: 'monotone' as const,
        fill: 'origin' as const,
        pointRadius: 0,
        pointHoverRadius: 4,
        pointHitRadius: 10,
        borderWidth: 2,
      },
      {
        label: t('channelMonitorV2.chart.cacheDataset'),
        data: cacheRates,
        borderColor: '#1e5c42',
        backgroundColor: 'rgba(30, 92, 66, 0.08)',
        yAxisID: 'yPct',
        tension: 0.4,
        cubicInterpolationMode: 'monotone' as const,
        fill: false,
        pointRadius: 0,
        pointHoverRadius: 4,
        pointHitRadius: 10,
        borderWidth: 2,
      },
      {
        label: t('channelMonitorV2.chart.ttftDataset'),
        data: ttftP50,
        borderColor: '#b7791f',
        backgroundColor: 'rgba(183, 121, 31, 0.08)',
        yAxisID: 'yTtft',
        tension: 0.4,
        cubicInterpolationMode: 'monotone' as const,
        fill: false,
        pointRadius: 0,
        pointHoverRadius: 4,
        pointHitRadius: 10,
        borderWidth: 2,
        spanGaps: true,
      },
    ],
  }
})

/** Window the series by zoom state around the cursor — not always the last N points. */
const visibleTrend = computed(() => sliceByZoom(props.trend || [], zoom.value))

function onChartWheel(event: WheelEvent) {
  // Plain vertical wheel zooms X (narrower time range); shift/horizontal pans.
  event.preventDefault()
  const ratio = clientXRatio(event.clientX, chartRef.value)
  zoom.value = applyWheelZoom(zoom.value, event, ratio)
}

function resetChartZoom() {
  zoom.value = resetZoom()
}

watch(() => props.trend, () => {
  zoom.value = resetZoom()
})

function smoothTrend(values: Array<number | null>): Array<number | null> {
  if (values.length <= 2) return values
  return values.map((value, index) => {
    if (value == null) return null
    const neighbors = values.slice(Math.max(0, index - 1), Math.min(values.length, index + 2))
      .filter((item): item is number => item != null)
    if (!neighbors.length) return value
    return neighbors.reduce((sum, item) => sum + item, 0) / neighbors.length
  })
}

const chartOptions = computed(() => {
  const text = isDark.value ? '#aaa69a' : '#777266'
  const grid = isDark.value ? '#47443a' : '#d8d2c3'
  const tooltipBg = isDark.value ? '#24231f' : '#fffefa'
  const tooltipTitle = isDark.value ? '#f4f2ec' : '#27251f'
  const tooltipBody = isDark.value ? '#aaa69a' : '#777266'
  return {
    responsive: true,
    maintainAspectRatio: false,
    interaction: { mode: 'index' as const, intersect: false },
    plugins: {
      legend: { display: false },
      tooltip: {
        backgroundColor: tooltipBg,
        titleColor: tooltipTitle,
        bodyColor: tooltipBody,
        borderColor: grid,
        borderWidth: 1,
        padding: 10,
        displayColors: true,
        callbacks: {
          label(ctx: { dataset: { label?: string }; parsed: { y: number | null } }) {
            const label = ctx.dataset.label || ''
            const y = ctx.parsed.y
            if (y == null) return `${label}: -`
            if (label === t('channelMonitorV2.chart.errorDataset') || label === t('channelMonitorV2.chart.cacheDataset')) {
              return `${label}: ${formatMonitorPercent(y / 100)}`
            }
            return `${label}: ${formatMonitorMs(y)}`
          },
        },
      },
    },
    scales: {
      x: {
        ticks: { color: text, maxRotation: 0, autoSkip: true, maxTicksLimit: 8, autoSkipPadding: 10, font: { size: 10 } },
        grid: { display: false },
      },
      yPct: {
        type: 'linear' as const,
        position: 'left' as const,
        min: 0,
        suggestedMax: 100,
        ticks: {
          color: text,
          font: { size: 10 },
          callback: (v: string | number) => `${v}%`,
        },
        grid: { color: grid, borderDash: [4, 4] },
        title: { display: true, text: t('channelMonitorV2.chart.percentAxis'), color: text, font: { size: 11 } },
      },
      yTtft: {
        type: 'linear' as const,
        position: 'right' as const,
        min: 0,
        ticks: {
          color: '#b7791f',
          font: { size: 10 },
          callback: (v: string | number) => formatMonitorMs(Number(v)),
        },
        grid: { display: false },
        title: { display: true, text: t('channelMonitorV2.metrics.ttftP50'), color: '#b7791f', font: { size: 11 } },
      },
    },
  }
})
</script>

<style scoped>
.scheme3-v2-panel {
  border: 1px solid #d8d2c3;
  border-radius: 8px;
  background: #fffefa;
  padding: 1rem 1.15rem;
  color: #27251f;
  box-shadow: 0 10px 24px rgba(54, 48, 34, .06);
}
.scheme3-v2-panel-header { border-bottom: 1px solid #d8d2c3; padding-bottom: .75rem; }
.scheme3-v2-panel-title { color: #27251f; font-family: Georgia, 'Times New Roman', serif; font-size: 1rem; font-weight: 600; }
.scheme3-v2-panel-icon { color: #1e5c42; }
.scheme3-v2-panel-description, .scheme3-v2-panel-tools, .scheme3-v2-trend-loading { color: #777266; }
.scheme3-v2-panel-badge { border: 1px solid #d8d2c3; border-radius: 999px; background: #f1eee6; color: #777266; padding: .18rem .42rem; font-size: .58rem; font-weight: 800; }
.scheme3-v2-panel-action { border: 1px solid #d8d2c3; border-radius: 6px; background: #fffefa; color: #777266; }
.scheme3-v2-panel-action:hover { background: #f1eee6; color: #27251f; }
.scheme3-v2-empty { gap: .8rem; padding: 1.5rem; text-align: left; }
.scheme3-v2-empty-mark { display: inline-flex; width: 2.3rem; height: 2.3rem; flex: none; align-items: center; justify-content: center; border: 1px solid rgba(30,92,66,.28); border-radius: 6px; background: rgba(30,92,66,.08); color: #1e5c42; }
.scheme3-v2-empty-copy strong { display: block; color: #27251f; font-family: Georgia, 'Times New Roman', serif; font-size: .92rem; font-weight: 600; }
.scheme3-v2-empty-copy p { margin-top: .25rem; max-width: 26rem; color: #777266; font-size: .68rem; line-height: 1.45; }
.scheme3-v2-legend-dot { width: .42rem; height: .42rem; border-radius: 999px; }
.scheme3-v2-legend-dot.is-danger { background: #9e4d3d; }
.scheme3-v2-legend-dot.is-accent { background: #1e5c42; }
.scheme3-v2-legend-dot.is-amber { background: #b7791f; }
:global(.dark .scheme3-v2-panel) { border-color: #47443a; background: #24231f; color: #f4f2ec; box-shadow: 0 14px 28px rgba(0, 0, 0, .22); }
:global(.dark .scheme3-v2-panel-header) { border-color: #47443a; }
:global(.dark .scheme3-v2-panel-title) { color: #f4f2ec; }
:global(.dark .scheme3-v2-panel-icon) { color: #8fc2a5; }
:global(.dark .scheme3-v2-panel-description), :global(.dark .scheme3-v2-panel-tools), :global(.dark .scheme3-v2-trend-loading) { color: #aaa69a; }
:global(.dark .scheme3-v2-panel-badge) { border-color: #47443a; background: #2b2924; color: #aaa69a; }
:global(.dark .scheme3-v2-panel-action) { border-color: #47443a; background: #24231f; color: #aaa69a; }
:global(.dark .scheme3-v2-panel-action:hover) { background: #2b2924; color: #f4f2ec; }
:global(.dark .scheme3-v2-empty-mark) { border-color: rgba(143,194,165,.28); background: rgba(143,194,165,.1); color: #8fc2a5; }
:global(.dark .scheme3-v2-empty-copy strong) { color: #f4f2ec; }
:global(.dark .scheme3-v2-empty-copy p) { color: #aaa69a; }
</style>

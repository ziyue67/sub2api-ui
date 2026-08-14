<template>
  <div class="scheme3-monitor-availability mt-3 flex items-end justify-between">
    <div class="scheme3-monitor-availability-label text-[11px] uppercase tracking-widest text-gray-400">
      {{ windowLabel }}
    </div>
    <div class="flex items-baseline gap-0.5">
      <span
        class="scheme3-monitor-availability-value text-3xl font-bold tabular-nums leading-none"
        :style="colorStyle"
      >
        {{ displayValue }}
      </span>
      <span
        class="scheme3-monitor-availability-unit text-base font-semibold leading-none"
        :style="colorStyle"
      >%</span>
    </div>
  </div>
  <div
    v-if="samplesLabel"
    class="scheme3-monitor-availability-extra mt-1 text-[11px] text-gray-400 text-right"
  >
    {{ samplesLabel }}
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { hslForPct } from '@/composables/useChannelMonitorFormat'

const props = defineProps<{
  windowLabel: string
  value: number | null
  samplesLabel?: string
}>()

const { t } = useI18n()

const displayValue = computed(() => {
  if (props.value === null || Number.isNaN(props.value)) return t('monitorCommon.latencyEmpty')
  return props.value.toFixed(2)
})

const colorStyle = computed(() => {
  const colour = hslForPct(props.value)
  return colour ? { color: colour } : { color: 'rgb(156 163 175)' }
})
</script>

<style scoped>
.scheme3-monitor-availability-label { color: #777266 !important; letter-spacing: .08em; }
.scheme3-monitor-availability-extra { color: #a49e90 !important; }
.scheme3-monitor-availability-value { font-family: Georgia, 'Times New Roman', serif; font-size: 1.75rem; font-weight: 600; }
.scheme3-monitor-availability-unit { font-family: Georgia, 'Times New Roman', serif; }

:global(.dark .scheme3-monitor-availability-label) { color: #aaa69a !important; }
:global(.dark .scheme3-monitor-availability-extra) { color: #827e72 !important; }
</style>

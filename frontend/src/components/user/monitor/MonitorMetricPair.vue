<template>
  <div class="scheme3-monitor-metric-pair mt-5 grid grid-cols-2 gap-2">
    <div
      class="scheme3-monitor-metric rounded-xl p-3 bg-gray-50/80 dark:bg-dark-900/40 border border-gray-100 dark:border-dark-700/50"
    >
      <div
        class="scheme3-monitor-metric-label flex items-center gap-1.5 text-[10px] font-semibold uppercase tracking-wider text-gray-400"
      >
        <Icon :name="primaryIcon" size="xs" />
        <span>{{ primaryLabel }}</span>
      </div>
      <div class="scheme3-monitor-metric-value mt-1.5 text-lg font-bold font-mono tabular-nums text-gray-900 dark:text-gray-100">
        {{ primaryValue }}<span class="scheme3-monitor-metric-unit text-xs font-normal text-gray-400 ml-0.5">{{ primaryUnit }}</span>
      </div>
    </div>
    <div
      class="scheme3-monitor-metric rounded-xl p-3 bg-gray-50/80 dark:bg-dark-900/40 border border-gray-100 dark:border-dark-700/50"
    >
      <div
        class="scheme3-monitor-metric-label flex items-center gap-1.5 text-[10px] font-semibold uppercase tracking-wider text-gray-400"
      >
        <Icon :name="secondaryIcon" size="xs" />
        <span>{{ secondaryLabel }}</span>
      </div>
      <div class="scheme3-monitor-metric-value mt-1.5 text-lg font-bold font-mono tabular-nums text-gray-900 dark:text-gray-100">
        {{ secondaryValue }}<span class="scheme3-monitor-metric-unit text-xs font-normal text-gray-400 ml-0.5">{{ secondaryUnit }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  primaryLabel: string
  primaryValue: string
  primaryUnit: string
  primaryIcon: 'bolt' | 'globe' | 'clock' | 'link'
  secondaryLabel: string
  secondaryValue: string
  secondaryUnit: string
  secondaryIcon: 'bolt' | 'globe' | 'clock' | 'link'
}>()
</script>

<style scoped>
.scheme3-monitor-metric { border-color: #d8d2c3 !important; border-radius: 7px !important; background: #f8f6ef !important; }
.scheme3-monitor-metric-label { color: #777266 !important; letter-spacing: .055em; }
.scheme3-monitor-metric-value { color: #27251f !important; font-size: 1.05rem; }
.scheme3-monitor-metric-unit { color: #a49e90 !important; }

:global(.dark .scheme3-monitor-metric) { border-color: #47443a !important; background: #2b2924 !important; }
:global(.dark .scheme3-monitor-metric-label) { color: #aaa69a !important; }
:global(.dark .scheme3-monitor-metric-value) { color: #f4f2ec !important; }
:global(.dark .scheme3-monitor-metric-unit) { color: #827e72 !important; }
</style>

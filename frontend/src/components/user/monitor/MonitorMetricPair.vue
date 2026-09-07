<template>
  <div class="scheme3-monitor-metric-pair">
    <div class="scheme3-monitor-metric">
      <div class="scheme3-monitor-metric-label">
        <Icon :name="primaryIcon" size="xs" />
        <span>{{ primaryLabel }}</span>
      </div>
      <div class="scheme3-monitor-metric-value">
        {{ primaryValue }}<span class="scheme3-monitor-metric-unit">{{ primaryUnit }}</span>
      </div>
    </div>
    <div class="scheme3-monitor-metric">
      <div class="scheme3-monitor-metric-label">
        <Icon :name="secondaryIcon" size="xs" />
        <span>{{ secondaryLabel }}</span>
      </div>
      <div class="scheme3-monitor-metric-value">
        {{ secondaryValue }}<span class="scheme3-monitor-metric-unit">{{ secondaryUnit }}</span>
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
.scheme3-monitor-metric-pair { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: .5rem; margin-top: 1.25rem; }
.scheme3-monitor-metric { min-width: 0; border: 1px solid var(--monitor-line, #d8d2c3); border-radius: 7px; padding: .75rem; background: var(--monitor-subtle, #f8f6ef); }
.scheme3-monitor-metric-label { display: flex; min-width: 0; align-items: center; gap: .38rem; overflow: hidden; color: var(--monitor-muted, #777266); font-size: .62rem; font-weight: 800; letter-spacing: .055em; line-height: 1.2; text-overflow: ellipsis; text-transform: uppercase; white-space: nowrap; }
.scheme3-monitor-metric-label span { overflow: hidden; text-overflow: ellipsis; }
.scheme3-monitor-metric-value { margin-top: .38rem; overflow: hidden; color: var(--monitor-ink, #27251f); font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: 1.05rem; font-variant-numeric: tabular-nums; font-weight: 800; line-height: 1.2; text-overflow: ellipsis; white-space: nowrap; }
.scheme3-monitor-metric-unit { margin-left: .12rem; color: var(--monitor-soft, #a49e90); font-size: .68rem; font-weight: 500; }

:global(.dark .scheme3-monitor-metric) { border-color: #47443a; background: #2b2924; }
:global(.dark .scheme3-monitor-metric-label) { color: #aaa69a; }
:global(.dark .scheme3-monitor-metric-value) { color: #f4f2ec; }
:global(.dark .scheme3-monitor-metric-unit) { color: #827e72; }

@media (max-width: 380px) {
  .scheme3-monitor-metric { padding: .62rem; }
  .scheme3-monitor-metric-value { font-size: .94rem; }
}
</style>

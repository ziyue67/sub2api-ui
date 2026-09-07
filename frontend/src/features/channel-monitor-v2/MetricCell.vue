<template>
  <div
    class="scheme3-v2-metric-cell"
    :title="title || undefined"
  >
    <div
      v-if="state"
      class="scheme3-v2-metric-dot"
      :class="dotClass"
      aria-hidden="true"
    ></div>
    <div class="min-w-0 flex-1">
      <span class="scheme3-v2-metric-label">{{ label }}</span>
      <strong
        class="scheme3-v2-metric-value"
        :class="stateClass"
      >{{ value }}</strong>
      <div
        v-if="detailParts.length > 1"
        class="scheme3-v2-metric-detail mt-1.5 flex flex-wrap gap-x-2 gap-y-0.5"
      >
        <span
          v-for="(part, index) in detailParts"
          :key="`${index}:${part}`"
          class="whitespace-nowrap tabular-nums"
        >{{ part }}</span>
      </div>
      <small
        v-else-if="detail"
        class="scheme3-v2-metric-detail mt-1.5 block"
      >{{ detail }}</small>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { HealthState } from '@/api/channelMonitorV2'

const props = defineProps<{
  label: string
  value: string
  detail: string
  state?: HealthState
  /** Exact numeric tooltip (e.g. uncompacted RPM/TPM). */
  title?: string
}>()

/** Split "AVG 475ms · P90 800ms" into chips so nothing is ellipsized. */
const detailParts = computed(() => {
  const raw = (props.detail || '').trim()
  if (!raw || raw === '-') return []
  return raw
    .split(/\s*[·|]\s*/)
    .map((part) => part.trim())
    .filter(Boolean)
})

const stateClass = computed(() => {
  if (!props.state) return 'is-default'
  if (props.state === 'healthy') return 'is-healthy'
  if (props.state === 'warning') return 'is-warning'
  if (props.state === 'critical') return 'is-critical'
  return 'is-unknown'
})

const dotClass = computed(() => {
  if (props.state === 'healthy') return 'is-healthy'
  if (props.state === 'warning') return 'is-warning'
  if (props.state === 'critical') return 'is-critical'
  return 'is-unknown'
})
</script>

<style scoped>
.scheme3-v2-metric-cell {
  position: relative;
  display: flex;
  min-height: 6.5rem;
  gap: .72rem;
  overflow: hidden;
  border: 1px solid #d8d2c3;
  border-radius: 8px;
  background: #fffefa;
  padding: .9rem 1rem;
  box-shadow: 0 8px 18px rgba(54, 48, 34, .05);
}
.scheme3-v2-metric-cell::after {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  height: 2px;
  background: #1e5c42;
  content: '';
  opacity: .18;
}
.scheme3-v2-metric-dot { width: .42rem; height: .42rem; flex: none; margin-top: .25rem; border-radius: 999px; background: #a49e90; }
.scheme3-v2-metric-dot.is-healthy { background: #1e5c42 !important; }
.scheme3-v2-metric-dot.is-warning { background: #b7791f !important; }
.scheme3-v2-metric-dot.is-critical { background: #9e4d3d !important; }
.scheme3-v2-metric-label {
  display: block;
  color: #777266;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: .56rem;
  font-weight: 800;
  letter-spacing: .08em;
  text-transform: uppercase;
}
.scheme3-v2-metric-value {
  display: block;
  margin-top: .28rem;
  overflow: visible;
  color: #27251f !important;
  font-family: Georgia, 'Times New Roman', serif;
  font-size: 1.45rem;
  font-weight: 600;
  line-height: 1.1;
  white-space: normal;
}
.scheme3-v2-metric-value.is-healthy { color: #1e5c42 !important; }
.scheme3-v2-metric-value.is-warning { color: #b7791f !important; }
.scheme3-v2-metric-value.is-critical { color: #9e4d3d !important; }
.scheme3-v2-metric-detail { color: #a49e90 !important; font-size: .62rem; line-height: 1.35; }
:global(.dark .scheme3-v2-metric-cell) { border-color: #47443a; background: #24231f; box-shadow: 0 12px 24px rgba(0, 0, 0, .2); }
:global(.dark .scheme3-v2-metric-cell::after) { background: #8fc2a5; }
:global(.dark .scheme3-v2-metric-label), :global(.dark .scheme3-v2-metric-detail) { color: #aaa69a !important; }
:global(.dark .scheme3-v2-metric-value) { color: #f4f2ec !important; }
:global(.dark .scheme3-v2-metric-value.is-healthy) { color: #8fc2a5 !important; }
:global(.dark .scheme3-v2-metric-value.is-warning) { color: #d3a55a !important; }
:global(.dark .scheme3-v2-metric-value.is-critical) { color: #d38b79 !important; }
</style>

<template>
  <button
    type="button"
    class="scheme3-monitor-card"
    @click="emit('click')"
  >
    <div class="scheme3-monitor-card-header">
      <span
        class="scheme3-monitor-provider-tile"
        :class="providerToneClass"
      >
        <ProviderIcon :provider="item.provider" :size="20" />
      </span>
      <div class="scheme3-monitor-card-copy">
        <div class="scheme3-monitor-card-name">
          {{ item.name }}
        </div>
        <div class="scheme3-monitor-card-meta">
          <span
            class="scheme3-monitor-provider-badge"
            :class="providerToneClass"
          >
            {{ providerLabel(item.provider) }}
          </span>
          <span class="scheme3-monitor-model">
            {{ item.primary_model }}
          </span>
          <span
            v-if="item.group_name"
            class="scheme3-monitor-group"
          >
            {{ item.group_name }}
          </span>
        </div>
      </div>
      <span
        class="scheme3-monitor-status"
        :class="statusToneClass"
      >
        {{ statusLabel(item.primary_status) }}
      </span>
    </div>

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

    <div class="scheme3-monitor-divider"></div>

    <MonitorAvailabilityRow
      :window-label="availabilityLabel"
      :value="availabilityValue"
      :samples-label="extraModelsCountLabel"
    />

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
import { useChannelMonitorFormat } from '@/composables/useChannelMonitorFormat'
import ProviderIcon from './ProviderIcon.vue'
import MonitorMetricPair from './MonitorMetricPair.vue'
import MonitorAvailabilityRow from './MonitorAvailabilityRow.vue'
import MonitorTimeline from './MonitorTimeline.vue'

const PROVIDER_TONES = new Set(['openai', 'anthropic', 'gemini', 'grok'])
const STATUS_TONES = new Set(['operational', 'degraded', 'failed', 'error'])

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
  providerLabel,
  formatLatency,
} = useChannelMonitorFormat()

const providerToneClass = computed(() => {
  const provider = String(props.item.provider || '').toLowerCase()
  return `is-${PROVIDER_TONES.has(provider) ? provider : 'unknown'}`
})

const statusToneClass = computed(() => {
  const status = String(props.item.primary_status || '').toLowerCase()
  return `is-${STATUS_TONES.has(status) ? status : 'unknown'}`
})

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
  display: flex;
  min-height: 17.5rem;
  width: 100%;
  flex-direction: column;
  border: 1px solid var(--monitor-line, #d8d2c3);
  border-radius: 8px;
  padding: 1.25rem;
  background: var(--monitor-card, #fffefa);
  box-shadow: 0 10px 24px rgba(54, 48, 34, .07);
  color: var(--monitor-ink, #27251f);
  text-align: left;
  transition: transform 160ms ease, box-shadow 160ms ease, border-color 160ms ease, background-color 160ms ease;
}

.scheme3-monitor-card:hover {
  border-color: rgba(30, 92, 66, .32);
  box-shadow: 0 15px 28px rgba(54, 48, 34, .11);
  transform: translateY(-2px);
}

.scheme3-monitor-card:active { transform: translateY(0) scale(.995); }
.scheme3-monitor-card:focus-visible { outline: 2px solid rgba(30, 92, 66, .32); outline-offset: 3px; }
.scheme3-monitor-card-header { display: flex; align-items: flex-start; gap: .75rem; }
.scheme3-monitor-provider-tile { display: grid; width: 2.25rem; height: 2.25rem; flex: 0 0 2.25rem; place-items: center; border: 1px solid var(--monitor-line, #d8d2c3); border-radius: 7px; background: var(--monitor-subtle, #f1eee6); color: var(--monitor-muted, #777266); }
.scheme3-monitor-provider-tile.is-openai { color: var(--monitor-accent, #1e5c42); }
.scheme3-monitor-provider-tile.is-anthropic { color: var(--monitor-danger, #9e4d3d); }
.scheme3-monitor-provider-tile.is-gemini { color: var(--monitor-amber, #b7791f); }
.scheme3-monitor-provider-tile.is-grok { color: var(--monitor-ink, #27251f); }
.scheme3-monitor-card-copy { min-width: 0; flex: 1 1 auto; }
.scheme3-monitor-card-name { overflow: hidden; color: var(--monitor-ink, #27251f); font-size: 1rem; font-weight: 700; text-overflow: ellipsis; white-space: nowrap; }
.scheme3-monitor-card-meta { display: flex; min-width: 0; align-items: center; gap: .38rem; margin-top: .15rem; }
.scheme3-monitor-model { min-width: 0; overflow: hidden; color: var(--monitor-muted, #777266); font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: .7rem; text-overflow: ellipsis; white-space: nowrap; }
.scheme3-monitor-group,
.scheme3-monitor-provider-badge,
.scheme3-monitor-status { display: inline-flex; flex: 0 0 auto; align-items: center; border: 1px solid currentColor; font-size: .58rem; font-weight: 800; line-height: 1; }
.scheme3-monitor-group { max-width: 7rem; overflow: hidden; border-color: var(--monitor-line, #d8d2c3); border-radius: 5px; padding: .22rem .38rem; background: var(--monitor-subtle, #f8f6ef); color: var(--monitor-muted, #655f53); text-overflow: ellipsis; white-space: nowrap; }
.scheme3-monitor-provider-badge { border-radius: 5px; padding: .22rem .38rem; }
.scheme3-monitor-provider-badge.is-openai { border-color: rgba(30, 92, 66, .28); background: rgba(30, 92, 66, .08); color: var(--monitor-accent, #1e5c42); }
.scheme3-monitor-provider-badge.is-anthropic { border-color: rgba(158, 77, 61, .28); background: rgba(158, 77, 61, .07); color: var(--monitor-danger, #9e4d3d); }
.scheme3-monitor-provider-badge.is-gemini { border-color: rgba(183, 121, 31, .3); background: rgba(183, 121, 31, .08); color: #8b5d14; }
.scheme3-monitor-provider-badge.is-grok,
.scheme3-monitor-provider-badge.is-unknown { border-color: var(--monitor-line, #d8d2c3); background: var(--monitor-subtle, #f1eee6); color: var(--monitor-muted, #777266); }
.scheme3-monitor-status { border-radius: 999px; padding: .38rem .58rem; font-size: .65rem; }
.scheme3-monitor-status.is-operational { border-color: rgba(30, 92, 66, .3); background: rgba(30, 92, 66, .08); color: var(--monitor-accent, #1e5c42); }
.scheme3-monitor-status.is-degraded { border-color: rgba(183, 121, 31, .34); background: rgba(183, 121, 31, .09); color: #8b5d14; }
.scheme3-monitor-status.is-failed,
.scheme3-monitor-status.is-error { border-color: rgba(158, 77, 61, .34); background: rgba(158, 77, 61, .08); color: var(--monitor-danger, #9e4d3d); }
.scheme3-monitor-status.is-unknown { border-color: var(--monitor-line, #d8d2c3); background: var(--monitor-subtle, #f1eee6); color: var(--monitor-muted, #777266); }
.scheme3-monitor-divider { margin-top: 1rem; border-top: 1px solid var(--monitor-line, #d8d2c3); }

:global(.dark .scheme3-monitor-card) {
  border-color: #47443a;
  background: #24231f;
  color: #f4f2ec;
  box-shadow: 0 14px 30px rgba(0, 0, 0, .24);
}

:global(.dark .scheme3-monitor-card:hover) {
  border-color: rgba(143, 194, 165, .35);
  box-shadow: 0 18px 34px rgba(0, 0, 0, .3);
}

:global(.dark .scheme3-monitor-card:focus-visible) { outline-color: rgba(143, 194, 165, .38); }
:global(.dark .scheme3-monitor-provider-tile) { border-color: #47443a; background: #2b2924; color: #aaa69a; }
:global(.dark .scheme3-monitor-provider-tile.is-openai) { color: #8fc2a5; }
:global(.dark .scheme3-monitor-provider-tile.is-anthropic) { color: #d38b79; }
:global(.dark .scheme3-monitor-provider-tile.is-gemini) { color: #d3a55a; }
:global(.dark .scheme3-monitor-provider-tile.is-grok) { color: #f4f2ec; }
:global(.dark .scheme3-monitor-card-name) { color: #f4f2ec; }
:global(.dark .scheme3-monitor-model) { color: #aaa69a; }
:global(.dark .scheme3-monitor-group) { border-color: #47443a; background: #2b2924; color: #d4d0c6; }
:global(.dark .scheme3-monitor-provider-badge.is-openai),
:global(.dark .scheme3-monitor-status.is-operational) { border-color: rgba(143, 194, 165, .32); background: rgba(143, 194, 165, .1); color: #8fc2a5; }
:global(.dark .scheme3-monitor-provider-badge.is-anthropic),
:global(.dark .scheme3-monitor-status.is-failed),
:global(.dark .scheme3-monitor-status.is-error) { border-color: rgba(211, 139, 121, .34); background: rgba(211, 139, 121, .09); color: #d38b79; }
:global(.dark .scheme3-monitor-provider-badge.is-gemini),
:global(.dark .scheme3-monitor-status.is-degraded) { border-color: rgba(211, 165, 90, .34); background: rgba(211, 165, 90, .09); color: #d3a55a; }
:global(.dark .scheme3-monitor-provider-badge.is-grok),
:global(.dark .scheme3-monitor-provider-badge.is-unknown),
:global(.dark .scheme3-monitor-status.is-unknown) { border-color: #47443a; background: #2b2924; color: #aaa69a; }
:global(.dark .scheme3-monitor-divider) { border-color: #47443a; }

@media (max-width: 560px) {
  .scheme3-monitor-card { min-height: 0; padding: 1rem; }
  .scheme3-monitor-status { padding-right: .45rem; padding-left: .45rem; font-size: .61rem; }
}
</style>

<template>
  <div class="scheme3-monitor-card-grid-state">
    <div
      v-if="loading && items.length === 0"
      class="scheme3-monitor-card-grid"
      aria-hidden="true"
    >
      <div
        v-for="i in 6"
        :key="i"
        class="scheme3-monitor-skeleton"
      >
        <div class="scheme3-monitor-skeleton-header">
          <div class="scheme3-monitor-skeleton-tile"></div>
          <div class="scheme3-monitor-skeleton-copy">
            <div class="scheme3-monitor-skeleton-line is-title"></div>
            <div class="scheme3-monitor-skeleton-line is-meta"></div>
          </div>
          <div class="scheme3-monitor-skeleton-chip"></div>
        </div>
        <div class="scheme3-monitor-skeleton-metrics">
          <div class="scheme3-monitor-skeleton-metric"></div>
          <div class="scheme3-monitor-skeleton-metric"></div>
        </div>
        <div class="scheme3-monitor-skeleton-timeline"></div>
      </div>
    </div>

    <section v-else-if="items.length === 0" class="scheme3-monitor-empty" role="status">
      <span class="scheme3-monitor-empty-mark" aria-hidden="true">
        <Icon name="server" size="lg" />
      </span>
      <strong class="scheme3-monitor-empty-title">{{ t('channelStatus.empty.title') }}</strong>
      <p class="scheme3-monitor-empty-description">{{ t('channelStatus.empty.description') }}</p>
    </section>

    <div
      v-else
      class="scheme3-monitor-card-grid"
    >
      <MonitorCard
        v-for="item in items"
        :key="item.id"
        :item="item"
        :window="window"
        :availability-value="resolveAvailability(item)"
        :countdown-seconds="countdownSeconds"
        @click="emit('cardClick', item)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { UserMonitorView, UserMonitorDetail } from '@/api/channelMonitor'
import Icon from '@/components/icons/Icon.vue'
import MonitorCard from './MonitorCard.vue'

const props = defineProps<{
  items: UserMonitorView[]
  window: '7d' | '15d' | '30d'
  countdownSeconds: number
  loading: boolean
  detailCache: Record<number, UserMonitorDetail>
}>()

const emit = defineEmits<{
  (e: 'cardClick', item: UserMonitorView): void
}>()

const { t } = useI18n()

function resolveAvailability(item: UserMonitorView): number | null {
  if (props.window === '7d') {
    return item.availability_7d ?? null
  }
  const detail = props.detailCache[item.id]
  if (!detail) return null
  const primary = detail.models.find(m => m.model === item.primary_model)
  if (!primary) return null
  return props.window === '15d' ? primary.availability_15d ?? null : primary.availability_30d ?? null
}
</script>

<style scoped>
.scheme3-monitor-card-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: .85rem;
}

.scheme3-monitor-skeleton {
  display: flex;
  min-height: 17.5rem;
  flex-direction: column;
  border: 1px solid var(--monitor-line, #d8d2c3);
  border-radius: 8px;
  padding: 1.25rem;
  background: var(--monitor-card, #fffefa);
  box-shadow: 0 10px 24px rgba(54, 48, 34, .05);
  animation: scheme3-monitor-skeleton-pulse 1.6s ease-in-out infinite;
}

.scheme3-monitor-skeleton-header { display: flex; align-items: flex-start; gap: .75rem; }
.scheme3-monitor-skeleton-tile { width: 2.25rem; height: 2.25rem; flex: 0 0 2.25rem; border-radius: 7px; background: #e5e0d4; }
.scheme3-monitor-skeleton-copy { min-width: 0; flex: 1 1 auto; }
.scheme3-monitor-skeleton-line { height: .72rem; border-radius: 3px; background: #e5e0d4; }
.scheme3-monitor-skeleton-line.is-title { width: 68%; }
.scheme3-monitor-skeleton-line.is-meta { width: 48%; margin-top: .52rem; }
.scheme3-monitor-skeleton-chip { width: 4rem; height: 1.35rem; flex: 0 0 4rem; border-radius: 999px; background: #e5e0d4; }
.scheme3-monitor-skeleton-metrics { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: .5rem; margin-top: 1.25rem; }
.scheme3-monitor-skeleton-metric { height: 4rem; border-radius: 7px; background: #ece8dd; }
.scheme3-monitor-skeleton-timeline { height: 1.25rem; margin-top: auto; border-radius: 4px; background: #ece8dd; }

.scheme3-monitor-empty {
  display: flex;
  min-height: 15rem;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  border: 1px dashed var(--monitor-line, #d8d2c3);
  border-radius: 8px;
  padding: 2rem 1.25rem;
  background: rgba(255, 254, 250, .68);
  color: var(--monitor-muted, #777266);
  text-align: center;
}
.scheme3-monitor-empty-mark { display: inline-flex; width: 2.65rem; height: 2.65rem; align-items: center; justify-content: center; border: 1px solid rgba(30, 92, 66, .24); border-radius: 7px; background: rgba(30, 92, 66, .07); color: var(--monitor-accent, #1e5c42); }
.scheme3-monitor-empty-title { margin-top: .8rem; color: var(--monitor-ink, #27251f); font-family: Georgia, 'Times New Roman', serif; font-size: 1rem; font-weight: 600; }
.scheme3-monitor-empty-description { max-width: 25rem; margin: .35rem 0 0; color: var(--monitor-muted, #777266); font-size: .72rem; line-height: 1.55; }

:global(.dark .scheme3-monitor-skeleton) {
  border-color: #47443a;
  background: #24231f;
  box-shadow: 0 14px 28px rgba(0, 0, 0, .22);
}
:global(.dark .scheme3-monitor-skeleton-tile),
:global(.dark .scheme3-monitor-skeleton-line),
:global(.dark .scheme3-monitor-skeleton-chip) { background: #3a3830; }
:global(.dark .scheme3-monitor-skeleton-metric),
:global(.dark .scheme3-monitor-skeleton-timeline) { background: #2b2924; }
:global(html.dark .scheme3-monitor-empty) { border-color: #47443a; background: rgba(36, 35, 31, .68); color: #aaa69a; }
:global(html.dark .scheme3-monitor-empty-mark) { border-color: rgba(143, 194, 165, .28); background: rgba(143, 194, 165, .1); color: #8fc2a5; }
:global(html.dark .scheme3-monitor-empty-title) { color: #f4f2ec; }
:global(html.dark .scheme3-monitor-empty-description) { color: #aaa69a; }

@keyframes scheme3-monitor-skeleton-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: .58; }
}

@media (min-width: 768px) {
  .scheme3-monitor-card-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}

@media (min-width: 1280px) {
  .scheme3-monitor-card-grid { grid-template-columns: repeat(3, minmax(0, 1fr)); }
}

@media (min-width: 1536px) {
  .scheme3-monitor-card-grid { grid-template-columns: repeat(4, minmax(0, 1fr)); }
}

@media (max-width: 767px) {
  .scheme3-monitor-card-grid { gap: .7rem; }
  .scheme3-monitor-empty { min-height: 12rem; }
}
</style>

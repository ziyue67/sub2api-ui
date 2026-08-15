<template>
  <div class="scheme3-auto-refresh" ref="dropdownRef">
    <button
      type="button"
      class="scheme3-auto-refresh-trigger"
      :aria-expanded="showDropdown"
      aria-haspopup="menu"
      :title="t('common.autoRefresh.title')"
      @click="showDropdown = !showDropdown"
    >
      <Icon name="refresh" size="sm" :class="{ 'is-spinning': enabled }" />
      <span class="scheme3-auto-refresh-label">
        {{ enabled
          ? t('common.autoRefresh.countdown', { seconds: countdown })
          : t('common.autoRefresh.title')
        }}
      </span>
    </button>

    <div
      v-if="showDropdown"
      class="scheme3-auto-refresh-menu"
      role="menu"
    >
      <div class="scheme3-auto-refresh-menu-inner">
        <button
          type="button"
          class="scheme3-auto-refresh-option"
          role="menuitemcheckbox"
          :aria-checked="enabled"
          @click="$emit('update:enabled', !enabled)"
        >
          <span>{{ t('common.autoRefresh.enable') }}</span>
          <Icon v-if="enabled" name="check" size="sm" class="scheme3-auto-refresh-check" />
        </button>
        <div class="scheme3-auto-refresh-divider" role="separator"></div>
        <button
          v-for="sec in intervals"
          :key="sec"
          type="button"
          class="scheme3-auto-refresh-option"
          role="menuitemradio"
          :aria-checked="intervalSeconds === sec"
          @click="$emit('update:interval', sec)"
        >
          <span>{{ t('common.autoRefresh.seconds', { n: sec }) }}</span>
          <Icon v-if="intervalSeconds === sec" name="check" size="sm" class="scheme3-auto-refresh-check" />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  enabled: boolean
  intervalSeconds: number
  countdown: number
  intervals: readonly number[]
}>()

defineEmits<{
  (e: 'update:enabled', value: boolean): void
  (e: 'update:interval', value: number): void
}>()

const { t } = useI18n()
const showDropdown = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)

function handleClickOutside(event: MouseEvent) {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target as Node)) {
    showDropdown.value = false
  }
}

onMounted(() => document.addEventListener('click', handleClickOutside))
onBeforeUnmount(() => document.removeEventListener('click', handleClickOutside))
</script>

<style scoped>
.scheme3-auto-refresh { position: relative; min-width: 0; }
.scheme3-auto-refresh-trigger { display: inline-flex; min-height: 2rem; align-items: center; justify-content: center; gap: .42rem; border: 1px solid var(--monitor-line, #d8d2c3); border-radius: 7px; padding: .42rem .65rem; background: var(--monitor-card, #fffefa); color: var(--monitor-muted, #777266); font-size: .65rem; font-weight: 800; white-space: nowrap; transition: border-color 150ms ease, background-color 150ms ease, color 150ms ease; }
.scheme3-auto-refresh-trigger:hover { border-color: rgba(30, 92, 66, .28); background: var(--monitor-subtle, #f1eee6); color: var(--monitor-ink, #27251f); }
.scheme3-auto-refresh-trigger:focus-visible,
.scheme3-auto-refresh-option:focus-visible { outline: 2px solid rgba(30, 92, 66, .28); outline-offset: 2px; }
.scheme3-auto-refresh-trigger .is-spinning { animation: scheme3-auto-refresh-spin 1.3s linear infinite; }
.scheme3-auto-refresh-menu { position: absolute; top: calc(100% + .28rem); right: 0; z-index: 20; width: 11rem; border: 1px solid var(--monitor-line, #d8d2c3); border-radius: 7px; background: var(--monitor-card, #fffefa); box-shadow: 0 16px 34px rgba(54, 48, 34, .16); }
.scheme3-auto-refresh-menu-inner { padding: .38rem; }
.scheme3-auto-refresh-option { display: flex; width: 100%; min-height: 2rem; align-items: center; justify-content: space-between; gap: .7rem; border: 1px solid transparent; border-radius: 5px; padding: .42rem .62rem; background: transparent; color: var(--monitor-muted, #777266); font-size: .68rem; text-align: left; transition: background-color 120ms ease, border-color 120ms ease, color 120ms ease; }
.scheme3-auto-refresh-option:hover { border-color: rgba(30, 92, 66, .16); background: var(--monitor-subtle, #f1eee6); color: var(--monitor-ink, #27251f); }
.scheme3-auto-refresh-check { flex: 0 0 auto; color: var(--monitor-accent, #1e5c42); }
.scheme3-auto-refresh-divider { height: 1px; margin: .28rem .15rem; background: var(--monitor-line, #d8d2c3); }

@keyframes scheme3-auto-refresh-spin { to { transform: rotate(360deg); } }

:global(.dark .scheme3-auto-refresh-trigger) { border-color: #47443a; background: #24231f; color: #aaa69a; }
:global(.dark .scheme3-auto-refresh-trigger:hover) { border-color: rgba(143, 194, 165, .3); background: #2b2924; color: #f4f2ec; }
:global(.dark .scheme3-auto-refresh-trigger:focus-visible),
:global(.dark .scheme3-auto-refresh-option:focus-visible) { outline-color: rgba(143, 194, 165, .38); }
:global(.dark .scheme3-auto-refresh-menu) { border-color: #47443a; background: #24231f; box-shadow: 0 18px 38px rgba(0, 0, 0, .28); }
:global(.dark .scheme3-auto-refresh-option) { color: #aaa69a; }
:global(.dark .scheme3-auto-refresh-option:hover) { border-color: rgba(143, 194, 165, .18); background: #2b2924; color: #f4f2ec; }
:global(.dark .scheme3-auto-refresh-check) { color: #8fc2a5; }
:global(.dark .scheme3-auto-refresh-divider) { background: #47443a; }
</style>

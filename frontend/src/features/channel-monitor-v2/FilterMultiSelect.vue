<template>
  <div
    ref="containerRef"
    class="scheme3-v2-filter-menu relative"
    :class="compact ? 'min-w-[6.5rem] sm:min-w-[7.25rem]' : 'min-w-[150px] sm:min-w-[160px]'"
  >
    <button
      ref="triggerRef"
      type="button"
      class="scheme3-v2-select-trigger flex cursor-pointer list-none items-center justify-between gap-1.5 text-left"
      :class="[
        isOpen ? 'is-open' : '',
        compact ? 'h-8 !px-2 !py-1 text-xs' : 'h-[42px]',
      ]"
      :aria-expanded="isOpen"
      aria-haspopup="listbox"
      :aria-label="label"
      @click="toggleOpen"
    >
      <span
        class="scheme3-v2-select-value min-w-0 truncate"
        :class="compact ? 'max-w-[5.5rem] sm:max-w-[6.5rem]' : 'max-w-[11rem]'"
      >
        {{ t('channelMonitorV2.filters.labelValue', { label, value: selectionLabel }) }}
      </span>
      <span class="scheme3-v2-select-icon shrink-0 transition-transform" :class="isOpen ? 'rotate-180' : ''">
        <Icon name="chevronDown" size="sm" />
      </span>
    </button>

    <Teleport to="body">
      <Transition name="scheme3-v2-dropdown">
        <div
          v-if="isOpen"
          ref="dropdownRef"
          class="scheme3-v2-filter-dropdown"
          :class="[instanceId]"
          :style="dropdownStyle"
          role="listbox"
          aria-multiselectable="true"
          @click.stop
          @mousedown.stop
        >
          <button
            type="button"
            class="scheme3-v2-filter-option scheme3-v2-filter-option-group flex w-full items-center justify-between px-4 py-2 text-left text-sm font-semibold"
            @click="clear"
          >
            <span>{{ allLabel }}</span>
            <Icon v-if="modelValue.length === 0" name="check" size="sm" class="scheme3-v2-filter-check" />
          </button>

          <button
            v-for="option in options"
            :key="option.value"
            type="button"
            role="option"
            class="scheme3-v2-filter-option flex w-full items-center justify-between gap-3 px-4 py-2 text-left text-sm"
            :class="modelValue.includes(option.value) ? 'is-selected' : ''"
            :aria-selected="modelValue.includes(option.value)"
            @click="toggle(option.value)"
          >
            <span class="flex min-w-0 flex-1 items-center gap-2">
              <span
                class="scheme3-v2-checkbox flex h-4 w-4 items-center justify-center"
                :class="modelValue.includes(option.value) ? 'is-checked' : ''"
              >
                <Icon v-if="modelValue.includes(option.value)" name="check" size="sm" class="scheme3-v2-filter-check" />
              </span>
              <span class="min-w-0 flex-1 truncate">{{ option.label }}</span>
            </span>
            <small v-if="option.count != null" class="scheme3-v2-filter-count text-xs">{{ formatCount(option.count) }}</small>
          </button>
          <p v-if="options.length === 0" class="scheme3-v2-filter-empty px-4 py-3 text-center text-xs">{{ t('channelMonitorV2.filters.empty') }}</p>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import type { CSSProperties } from 'vue'
import { useI18n } from 'vue-i18n'
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import { monitorIntlLocale } from '@/features/channel-monitor-v2/monitorFormat'

interface FilterOption {
  value: string
  label: string
  count?: number
}

const props = withDefaults(
  defineProps<{
    label: string
    allLabel: string
    modelValue: string[]
    /** Options for this picker (parent may cascade by platform). */
    options: FilterOption[]
    /** Compact trigger for single-row toolbars. */
    compact?: boolean
  }>(),
  { compact: false },
)
const emit = defineEmits<{ 'update:modelValue': [value: string[]] }>()
const { t, locale } = useI18n()

const containerRef = ref<HTMLElement | null>(null)
const triggerRef = ref<HTMLButtonElement | null>(null)
const dropdownRef = ref<HTMLElement | null>(null)
const isOpen = ref(false)
const instanceId = `filter-select-${Math.random().toString(36).slice(2, 9)}`

const selectionLabel = computed(() => {
  if (props.modelValue.length === 0) return props.allLabel
  if (props.modelValue.length === 1) {
    return props.options.find((item) => item.value === props.modelValue[0])?.label || props.modelValue[0]
  }
  return t('channelMonitorV2.filters.selectedCount', { count: props.modelValue.length })
})

const dropdownStyle = computed<CSSProperties>(() => {
  const trigger = triggerRef.value
  if (!trigger) return {}
  const rect = trigger.getBoundingClientRect()
  const padding = 8
  const viewportRight = Math.max(padding, window.innerWidth - padding)
  const left = Math.min(Math.max(padding, rect.left), viewportRight)
  const availableWidth = Math.max(0, viewportRight - left)
  const preferredMinWidth = Math.max(200, rect.width)
  const minWidth = Math.min(preferredMinWidth, availableWidth)
  return {
    position: 'fixed' as const,
    left: `${left}px`,
    top: `${rect.bottom + 4}px`,
    minWidth: `${minWidth}px`,
    maxWidth: `${availableWidth}px`,
    zIndex: '100000020',
  }
})

function clear() {
  emit('update:modelValue', [])
  // Stay open so users can re-select without reopening.
}

function toggle(value: string) {
  const selected = new Set(props.modelValue)
  if (selected.has(value)) selected.delete(value)
  else selected.add(value)
  emit('update:modelValue', [...selected])
  // Stay open on toggle (multi-select).
}

function toggleOpen() {
  isOpen.value ? close() : open()
}

function open() {
  isOpen.value = true
  void nextTick(() => positionDropdown())
}

function close() {
  isOpen.value = false
}

function positionDropdown() {
  const trigger = triggerRef.value
  const dropdown = dropdownRef.value
  if (!trigger || !dropdown) return
  const rect = trigger.getBoundingClientRect()
  const padding = 8
  const viewportRight = Math.max(padding, window.innerWidth - padding)
  const left = Math.min(Math.max(padding, rect.left), viewportRight)
  const availableWidth = Math.max(0, viewportRight - left)
  const preferredMinWidth = Math.max(200, rect.width)
  const minWidth = Math.min(preferredMinWidth, availableWidth)
  dropdown.style.left = `${left}px`
  dropdown.style.top = `${rect.bottom + 4}px`
  dropdown.style.minWidth = `${minWidth}px`
  dropdown.style.maxWidth = `${availableWidth}px`
}

function formatCount(value: number) {
  return Intl.NumberFormat(locale.value || monitorIntlLocale(), {
    notation: value >= 10000 ? 'compact' : 'standard',
  }).format(value)
}

function onDocumentMouseDown(event: MouseEvent) {
  if (!isOpen.value) return
  const target = event.target as Node | null
  if (!target) return
  if (containerRef.value?.contains(target)) return
  if (dropdownRef.value?.contains(target)) return
  close()
}

function onDocumentKeyDown(event: KeyboardEvent) {
  if (event.key === 'Escape') close()
}

function onWindowChange() {
  if (isOpen.value) positionDropdown()
}

watch(isOpen, async (open) => {
  if (open) {
    await nextTick()
    positionDropdown()
    document.addEventListener('mousedown', onDocumentMouseDown)
    document.addEventListener('keydown', onDocumentKeyDown)
    window.addEventListener('resize', onWindowChange)
    window.addEventListener('scroll', onWindowChange, true)
  } else {
    document.removeEventListener('mousedown', onDocumentMouseDown)
    document.removeEventListener('keydown', onDocumentKeyDown)
    window.removeEventListener('resize', onWindowChange)
    window.removeEventListener('scroll', onWindowChange, true)
  }
})

onBeforeUnmount(() => {
  document.removeEventListener('mousedown', onDocumentMouseDown)
  document.removeEventListener('keydown', onDocumentKeyDown)
  window.removeEventListener('resize', onWindowChange)
  window.removeEventListener('scroll', onWindowChange, true)
})
</script>

<style scoped>
.scheme3-v2-select-trigger {
  width: 100%;
  min-width: 0;
  border: 1px solid #d8d2c3;
  border-radius: 7px;
  background: #fffefa;
  color: #27251f;
  padding: .55rem .68rem;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: .62rem;
  font-weight: 700;
  line-height: 1.2;
  transition: border-color 150ms ease, background-color 150ms ease, color 150ms ease;
}
.scheme3-v2-select-trigger:hover,
.scheme3-v2-select-trigger.is-open {
  border-color: #1e5c42;
  background: #f1eee6;
  color: #1e5c42;
}
.scheme3-v2-select-trigger:focus-visible {
  outline: 2px solid rgba(30, 92, 66, .3);
  outline-offset: 2px;
}
.scheme3-v2-select-icon { color: #777266; }
.scheme3-v2-select-trigger:hover .scheme3-v2-select-icon,
.scheme3-v2-select-trigger.is-open .scheme3-v2-select-icon { color: #1e5c42; }

:global(.scheme3-v2-filter-dropdown) {
  width: max-content;
  min-width: 200px;
  max-width: calc(100vw - 1rem);
  max-height: min(50vh, 360px);
  overflow-y: auto;
  border: 1px solid #d8d2c3;
  border-radius: 7px;
  background: #fffefa;
  color: #27251f;
  box-shadow: 0 18px 38px rgba(54, 48, 34, .18);
  padding: .25rem;
}
:global(.scheme3-v2-filter-option) {
  min-height: 2.1rem;
  cursor: pointer;
  border-radius: 5px;
  color: #27251f;
  transition: background-color 120ms ease, color 120ms ease;
}
:global(.scheme3-v2-filter-option:hover),
:global(.scheme3-v2-filter-option:focus-visible),
:global(.scheme3-v2-filter-option.is-selected) {
  background: #f1eee6;
  color: #1e5c42;
  outline: none;
}
:global(.scheme3-v2-filter-option-group) {
  border-bottom: 1px solid #d8d2c3;
  margin-bottom: .2rem;
}
:global(.scheme3-v2-checkbox) {
  width: 1rem;
  height: 1rem;
  flex: none;
  border: 1px solid #a49e90;
  border-radius: 4px;
  background: #fffefa;
}
:global(.scheme3-v2-checkbox.is-checked) {
  border-color: #1e5c42;
  background: rgba(30, 92, 66, .1);
}
:global(.scheme3-v2-filter-check) { color: #1e5c42; }
:global(.scheme3-v2-filter-count) { color: #a49e90 !important; }
:global(.scheme3-v2-filter-empty) { color: #777266; }

:global(.dark .scheme3-v2-select-trigger) {
  border-color: #47443a;
  background: #24231f;
  color: #f4f2ec;
}
:global(.dark .scheme3-v2-select-trigger:hover),
:global(.dark .scheme3-v2-select-trigger.is-open) {
  border-color: #8fc2a5;
  background: #2b2924;
  color: #8fc2a5;
}
:global(.dark .scheme3-v2-select-icon) { color: #aaa69a; }
:global(.dark .scheme3-v2-select-trigger:hover .scheme3-v2-select-icon),
:global(.dark .scheme3-v2-select-trigger.is-open .scheme3-v2-select-icon) { color: #8fc2a5; }
:global(.dark .scheme3-v2-filter-dropdown) {
  border-color: #47443a;
  background: #24231f;
  color: #f4f2ec;
  box-shadow: 0 18px 38px rgba(0, 0, 0, .34);
}
:global(.dark .scheme3-v2-filter-option) { color: #f4f2ec; }
:global(.dark .scheme3-v2-filter-option:hover),
:global(.dark .scheme3-v2-filter-option:focus-visible),
:global(.dark .scheme3-v2-filter-option.is-selected) { background: #2b2924; color: #8fc2a5; }
:global(.dark .scheme3-v2-filter-option-group) { border-color: #47443a; }
:global(.dark .scheme3-v2-checkbox) { border-color: #827e72; background: #2b2924; }
:global(.dark .scheme3-v2-checkbox.is-checked) { border-color: #8fc2a5; background: rgba(143, 194, 165, .12); }
:global(.dark .scheme3-v2-filter-check) { color: #8fc2a5; }
:global(.dark .scheme3-v2-filter-count) { color: #827e72 !important; }
:global(.dark .scheme3-v2-filter-empty) { color: #aaa69a; }

:global(.scheme3-v2-dropdown-enter-active),
:global(.scheme3-v2-dropdown-leave-active) { transition: opacity 140ms ease, transform 140ms ease; }
:global(.scheme3-v2-dropdown-enter-from),
:global(.scheme3-v2-dropdown-leave-to) { opacity: 0; transform: translateY(-4px); }

@media (max-width: 640px) {
  .scheme3-v2-filter-menu {
    min-width: 100%;
  }
}
</style>

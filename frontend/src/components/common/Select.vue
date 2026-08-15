<template>
  <div class="relative" :class="scheme3 && 'scheme3-select'" ref="containerRef">
    <button
      ref="triggerRef"
      type="button"
      @click="toggle"
      :disabled="disabled"
      :aria-expanded="isOpen"
      :aria-haspopup="true"
      :id="id"
      :aria-label="ariaLabel ?? 'Select option'"
      :aria-describedby="ariaDescribedby"
      :class="[
        scheme3 ? 'scheme3-select-trigger' : 'select-trigger',
        isOpen && (scheme3 ? 'scheme3-select-trigger-open' : 'select-trigger-open'),
        error && (scheme3 ? 'scheme3-select-trigger-error' : 'select-trigger-error'),
        disabled && (scheme3 ? 'scheme3-select-trigger-disabled' : 'select-trigger-disabled'),
        theme === 'dark' && (scheme3 ? 'scheme3-select-theme-dark' : 'select-theme-dark')
      ]"
      @keydown.down.prevent="onTriggerKeyDown"
      @keydown.up.prevent="onTriggerKeyDown"
    >
      <span :class="scheme3 ? 'scheme3-select-value' : 'select-value'">
        <slot name="selected" :option="selectedOption">
          {{ selectedLabel }}
        </slot>
      </span>
      <span
        v-if="clearable && hasValue && !disabled"
        :class="scheme3 ? 'scheme3-select-clear' : 'select-clear'"
        role="button"
        tabindex="-1"
        aria-label="Clear selection"
        @click.stop="clearSelection"
        @mousedown.stop
        @keydown.enter.stop.prevent="clearSelection"
      >
        <Icon name="x" size="sm" />
      </span>
      <span :class="scheme3 ? 'scheme3-select-icon' : 'select-icon'">
        <Icon
          name="chevronDown"
          size="md"
          :class="['transition-transform duration-200', isOpen && 'rotate-180']"
        />
      </span>
    </button>

    <!-- Teleport dropdown to body to escape stacking context -->
    <Teleport to="body">
      <Transition :name="scheme3 ? 'scheme3-select-dropdown' : 'select-dropdown'">
        <div
          v-if="isOpen"
          ref="dropdownRef"
          :class="[
            scheme3 ? 'scheme3-select-dropdown-portal' : 'select-dropdown-portal',
            instanceId,
            theme === 'dark' && (scheme3 ? 'scheme3-select-theme-dark' : 'select-theme-dark')
          ]"
          :style="dropdownStyle"
          role="listbox"
          @click.stop
          @mousedown.stop
          @keydown="onDropdownKeyDown"
        >
          <!-- Search input -->
          <div v-if="isSearchable" :class="scheme3 ? 'scheme3-select-search' : 'select-search'">
            <Icon name="search" size="sm" :class="scheme3 ? 'scheme3-select-search-icon' : 'text-gray-400'" />
            <input
              ref="searchInputRef"
              v-model="searchQuery"
              type="text"
              :placeholder="searchPlaceholderText"
              :aria-label="searchPlaceholderText"
              :class="scheme3 ? 'scheme3-select-search-input' : 'select-search-input'"
              @click.stop
            />
          </div>

          <!-- Options list -->
          <div :class="scheme3 ? 'scheme3-select-options' : 'select-options'" ref="optionsListRef">
            <div
              v-for="(option, index) in filteredOptions"
              :key="`${typeof getOptionValue(option)}:${String(getOptionValue(option) ?? '')}`"
              role="option"
              :aria-selected="isSelected(option)"
              :aria-disabled="isOptionDisabled(option)"
              @click.stop="!isOptionDisabled(option) && selectOption(option)"
              @mouseenter="handleOptionMouseEnter(option, index)"
              :class="[
                scheme3 ? 'scheme3-select-option' : 'select-option',
                isGroupHeaderOption(option) && (scheme3 ? 'scheme3-select-option-group' : 'select-option-group'),
                isSelected(option) && (scheme3 ? 'scheme3-select-option-selected' : 'select-option-selected'),
                isOptionDisabled(option) && !isGroupHeaderOption(option) && (scheme3 ? 'scheme3-select-option-disabled' : 'select-option-disabled'),
                focusedIndex === index && !isGroupHeaderOption(option) && (scheme3 ? 'scheme3-select-option-focused' : 'select-option-focused')
              ]"
            >
              <slot name="option" :option="option" :selected="isSelected(option)">
                <Icon
                  v-if="option._creatable"
                  name="search"
                  size="sm"
                  :class="scheme3 ? 'scheme3-select-option-creatable-icon' : 'flex-shrink-0 text-gray-400'"
                />
                <span :class="scheme3 ? 'scheme3-select-option-label' : 'select-option-label'">
                  <span :class="option._creatable && (scheme3 ? 'scheme3-select-option-creatable' : 'italic text-gray-500 dark:text-dark-300')">{{ getOptionLabel(option) }}</span>
                </span>
                <Icon
                  v-if="isSelected(option)"
                  name="check"
                  size="sm"
                  :class="scheme3 ? 'scheme3-select-option-check' : 'select-option-check'"
                  :stroke-width="2"
                />
              </slot>
            </div>

            <!-- Empty state -->
            <div v-if="filteredOptions.length === 0" :class="scheme3 ? 'scheme3-select-empty' : 'select-empty'">
              {{ emptyTextDisplay }}
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

// Instance ID for unique click-outside detection
const instanceId = `select-${Math.random().toString(36).substring(2, 9)}`

export interface SelectOption {
  value: string | number | boolean | null
  label: string
  disabled?: boolean
  [key: string]: unknown
}

interface Props {
  modelValue: string | number | boolean | null | undefined
  options: SelectOption[] | Array<Record<string, unknown>>
  placeholder?: string
  disabled?: boolean
  error?: boolean
  searchable?: boolean | 'auto'
  searchPlaceholder?: string
  emptyText?: string
  valueKey?: string
  labelKey?: string
  creatable?: boolean
  creatablePrefix?: string
  clearable?: boolean
  id?: string
  ariaLabel?: string
  ariaDescribedby?: string
  /** Visual theme for an individual Select instance; defaults to the app theme. */
  theme?: 'light' | 'dark'
  /** Third-version visual surface used by monitor routes without legacy utility classes. */
  scheme3?: boolean
}

interface Emits {
  (e: 'update:modelValue', value: string | number | boolean | null): void
  (e: 'change', value: string | number | boolean | null, option: SelectOption | null): void
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  error: false,
  searchable: 'auto',
  creatable: false,
  creatablePrefix: '',
  clearable: false,
  valueKey: 'value',
  labelKey: 'label',
  theme: 'light',
  scheme3: false
})

const emit = defineEmits<Emits>()

const isOpen = ref(false)
const searchQuery = ref('')
const focusedIndex = ref(-1)
const containerRef = ref<HTMLElement | null>(null)
const triggerRef = ref<HTMLButtonElement | null>(null)
const searchInputRef = ref<HTMLInputElement | null>(null)
const dropdownRef = ref<HTMLElement | null>(null)
const optionsListRef = ref<HTMLElement | null>(null)
const dropdownPosition = ref<'bottom' | 'top'>('bottom')
const triggerRect = ref<DOMRect | null>(null)
const dropdownViewportPadding = 8
const dropdownMinimumWidth = 200

// i18n placeholders
const placeholderText = computed(() => props.placeholder ?? t('common.selectOption'))
const searchPlaceholderText = computed(() => props.searchPlaceholder ?? t('common.searchPlaceholder'))
const emptyTextDisplay = computed(() => props.emptyText ?? t('common.noOptionsFound'))

const isSearchable = computed(() => {
  if (props.searchable === 'auto') return props.options.length > 5
  return props.searchable
})

// Computed style for teleported dropdown
const dropdownStyle = computed(() => {
  if (!triggerRect.value) return {}

  const rect = triggerRect.value
  const viewportRight = Math.max(dropdownViewportPadding, window.innerWidth - dropdownViewportPadding)
  const left = Math.min(
    Math.max(dropdownViewportPadding, rect.left),
    viewportRight
  )
  const availableWidth = Math.max(0, viewportRight - left)
  const preferredMinWidth = Math.max(dropdownMinimumWidth, rect.width)
  const minWidth = Math.min(preferredMinWidth, availableWidth)
  const style: Record<string, string> = {
    position: 'fixed',
    left: `${left}px`,
    minWidth: `${minWidth}px`,
    maxWidth: `${availableWidth}px`,
    zIndex: '100000020'
  }

  if (dropdownPosition.value === 'top') {
    style.bottom = `${window.innerHeight - rect.top + 4}px`
  } else {
    style.top = `${rect.bottom + 4}px`
  }

  return style
})

const getOptionValue = (option: any): any => {
  if (typeof option === 'object' && option !== null) {
    return option[props.valueKey]
  }
  return option
}

const getOptionLabel = (option: any): string => {
  if (typeof option === 'object' && option !== null) {
    return String(option[props.labelKey] ?? '')
  }
  return String(option ?? '')
}

const isOptionDisabled = (option: any): boolean => {
  if (typeof option === 'object' && option !== null) {
    return !!option.disabled
  }
  return false
}

const isGroupHeaderOption = (option: any): boolean => {
  if (typeof option === 'object' && option !== null) {
    return option.kind === 'group'
  }
  return false
}

const selectedOption = computed(() => {
  return props.options.find((opt) => getOptionValue(opt) === props.modelValue) || null
})

const selectedLabel = computed(() => {
  if (selectedOption.value) {
    return getOptionLabel(selectedOption.value)
  }
  // In creatable mode, show the raw value if no matching option
  if (props.creatable && props.modelValue) {
    return String(props.modelValue)
  }
  return placeholderText.value
})

const hasValue = computed(
  () => props.modelValue !== null && props.modelValue !== undefined && props.modelValue !== ''
)

const filteredOptions = computed(() => {
  let opts = props.options as any[]
  if (isSearchable.value && searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    opts = opts.filter((opt) => {
      // Match label
      if (getOptionLabel(opt).toLowerCase().includes(query)) return true
      // Also match description if present
      if (opt.description && String(opt.description).toLowerCase().includes(query)) return true
      return false
    })
    // In creatable mode, always prepend a fuzzy search option
    if (props.creatable && searchQuery.value.trim()) {
      const trimmed = searchQuery.value.trim()
      const prefix = props.creatablePrefix || t('common.search')
      opts = [{ [props.valueKey]: trimmed, [props.labelKey]: `${prefix} "${trimmed}"`, _creatable: true }, ...opts]
    }
  }
  return opts
})

const isSelected = (option: any): boolean => {
  return getOptionValue(option) === props.modelValue
}

const findNextEnabledIndex = (startIndex: number): number => {
  const opts = filteredOptions.value
  if (opts.length === 0) return -1
  for (let offset = 0; offset < opts.length; offset++) {
    const idx = (startIndex + offset) % opts.length
    if (!isOptionDisabled(opts[idx])) return idx
  }
  return -1
}

const findPrevEnabledIndex = (startIndex: number): number => {
  const opts = filteredOptions.value
  if (opts.length === 0) return -1
  for (let offset = 0; offset < opts.length; offset++) {
    const idx = (startIndex - offset + opts.length) % opts.length
    if (!isOptionDisabled(opts[idx])) return idx
  }
  return -1
}

const handleOptionMouseEnter = (option: any, index: number) => {
  if (isOptionDisabled(option) || isGroupHeaderOption(option)) return
  focusedIndex.value = index
}

// Update trigger rect periodically while open to follow scroll/resize
const updateTriggerRect = () => {
  if (containerRef.value) {
    triggerRect.value = containerRef.value.getBoundingClientRect()
  }
}

const calculateDropdownPosition = () => {
  if (!containerRef.value) return
  updateTriggerRect()

  nextTick(() => {
    if (!dropdownRef.value || !triggerRect.value) return
    const dropdownHeight = dropdownRef.value.offsetHeight || 240
    const spaceBelow = window.innerHeight - triggerRect.value.bottom
    const spaceAbove = triggerRect.value.top

    if (spaceBelow < dropdownHeight && spaceAbove > dropdownHeight) {
      dropdownPosition.value = 'top'
    } else {
      dropdownPosition.value = 'bottom'
    }
  })
}

const toggle = () => {
  if (props.disabled) return
  isOpen.value = !isOpen.value
}

watch(isOpen, (open) => {
  if (open) {
    calculateDropdownPosition()
    // Reset focused index to current selection or first item
    if (filteredOptions.value.length === 0) {
      focusedIndex.value = -1
    } else {
      const selectedIdx = filteredOptions.value.findIndex(isSelected)
      const initialIdx = selectedIdx >= 0 ? selectedIdx : 0
      focusedIndex.value = isOptionDisabled(filteredOptions.value[initialIdx])
        ? findNextEnabledIndex(initialIdx + 1)
        : initialIdx
    }

    if (isSearchable.value) {
      nextTick(() => searchInputRef.value?.focus())
    }
    // Add scroll listener to update position
    window.addEventListener('scroll', updateTriggerRect, { capture: true, passive: true })
    window.addEventListener('resize', calculateDropdownPosition)
  } else {
    searchQuery.value = ''
    focusedIndex.value = -1
    window.removeEventListener('scroll', updateTriggerRect, { capture: true })
    window.removeEventListener('resize', calculateDropdownPosition)
  }
})

const selectOption = (option: any) => {
  const value = getOptionValue(option) ?? null
  emit('update:modelValue', value)
  emit('change', value, option)
  isOpen.value = false
  triggerRef.value?.focus()
}

const clearSelection = () => {
  if (props.disabled) return
  emit('update:modelValue', null)
  emit('change', null, null)
}

// Keyboards
const onTriggerKeyDown = () => {
  if (!isOpen.value) {
    isOpen.value = true
  }
}

const onDropdownKeyDown = (e: KeyboardEvent) => {
  switch (e.key) {
    case 'ArrowDown':
      e.preventDefault()
      focusedIndex.value = findNextEnabledIndex(focusedIndex.value + 1)
      if (focusedIndex.value >= 0) scrollToFocused()
      break
    case 'ArrowUp':
      e.preventDefault()
      focusedIndex.value = findPrevEnabledIndex(focusedIndex.value - 1)
      if (focusedIndex.value >= 0) scrollToFocused()
      break
    case 'Enter':
      e.preventDefault()
      if (focusedIndex.value >= 0 && focusedIndex.value < filteredOptions.value.length) {
        const opt = filteredOptions.value[focusedIndex.value]
        if (!isOptionDisabled(opt)) selectOption(opt)
      }
      break
    case 'Escape':
      e.preventDefault()
      isOpen.value = false
      triggerRef.value?.focus()
      break
    case 'Tab':
      isOpen.value = false
      break
  }
}

const scrollToFocused = () => {
  nextTick(() => {
    const list = optionsListRef.value
    if (!list) return
    const focusedEl = list.children[focusedIndex.value] as HTMLElement
    if (!focusedEl) return

    if (focusedEl.offsetTop < list.scrollTop) {
      list.scrollTop = focusedEl.offsetTop
    } else if (focusedEl.offsetTop + focusedEl.offsetHeight > list.scrollTop + list.offsetHeight) {
      list.scrollTop = focusedEl.offsetTop + focusedEl.offsetHeight - list.offsetHeight
    }
  })
}

const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as HTMLElement
  // Check if click is inside THIS specific instance's dropdown or trigger
  const isInDropdown = !!target.closest(`.${instanceId}`)
  const isInTrigger = containerRef.value?.contains(target)

  if (!isInDropdown && !isInTrigger && isOpen.value) {
    isOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  window.removeEventListener('scroll', updateTriggerRect, { capture: true })
  window.removeEventListener('resize', calculateDropdownPosition)
})
</script>

<style scoped>
.select-trigger {
  @apply flex w-full items-center justify-between gap-2;
  @apply rounded-xl px-4 py-2.5 text-sm;
  @apply bg-white dark:bg-dark-800;
  @apply border border-gray-200 dark:border-dark-600;
  @apply text-gray-900 dark:text-gray-100;
  @apply transition-all duration-200;
  @apply focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/30;
  @apply hover:border-gray-300 dark:hover:border-dark-500;
  @apply cursor-pointer;
}

/* Per-instance dark styling. This does not depend on html.dark so pages can opt in independently. */
.select-trigger.select-theme-dark {
  @apply bg-dark-800 border-dark-600 text-gray-100;
}

.select-trigger.select-theme-dark:hover {
  @apply border-dark-500;
}

.select-trigger.select-theme-dark .select-icon {
  @apply text-dark-400;
}

.select-trigger.select-theme-dark.select-trigger-disabled {
  @apply bg-dark-900;
}

.select-trigger-open {
  @apply border-primary-500 ring-2 ring-primary-500/30;
}

.select-trigger-error {
  @apply border-red-500 focus:border-red-500 focus:ring-red-500/30;
}

.select-trigger-disabled {
  @apply cursor-not-allowed bg-gray-100 opacity-60 dark:bg-dark-900;
}

.select-value {
  @apply min-w-0 flex-1 truncate text-left;
}

.select-icon {
  @apply flex-shrink-0 text-gray-400 dark:text-dark-400;
}

.select-clear {
  @apply flex flex-shrink-0 cursor-pointer items-center justify-center;
  @apply rounded text-gray-400 transition-colors;
  @apply hover:text-gray-600 dark:hover:text-gray-200;
}
</style>

<style>
.select-dropdown-portal {
  @apply w-max min-w-[200px];
  @apply bg-white dark:bg-dark-800;
  @apply rounded-xl;
  @apply border border-gray-200 dark:border-dark-700;
  @apply shadow-lg shadow-black/10 dark:shadow-black/30;
  @apply overflow-hidden;
  pointer-events: auto !important;
}

.select-dropdown-portal.select-theme-dark {
  @apply bg-dark-800 border-dark-700 shadow-black/30;
}

.select-dropdown-portal.select-theme-dark .select-search {
  @apply border-dark-700;
}

.select-dropdown-portal.select-theme-dark .select-search-input {
  @apply text-gray-100 placeholder:text-dark-400;
}

.select-dropdown-portal.select-theme-dark .select-option {
  @apply text-gray-300;
}

.select-dropdown-portal.select-theme-dark .select-option:hover,
.select-dropdown-portal.select-theme-dark .select-option-focused {
  @apply bg-dark-700;
}

.select-dropdown-portal.select-theme-dark .select-option-selected {
  @apply bg-primary-900/20 text-primary-300;
}

.select-dropdown-portal.select-theme-dark .select-option-group,
.select-dropdown-portal.select-theme-dark .select-option-group:hover {
  @apply bg-dark-900 text-gray-400;
}

.select-dropdown-portal.select-theme-dark .select-empty {
  @apply text-dark-400;
}

.select-dropdown-portal .select-search {
  @apply flex items-center gap-2 px-3 py-2;
  @apply border-b border-gray-100 dark:border-dark-700;
}

.select-dropdown-portal .select-search-input {
  @apply flex-1 bg-transparent text-sm;
  @apply text-gray-900 dark:text-gray-100;
  @apply placeholder:text-gray-400 dark:placeholder:text-dark-400;
  @apply focus:outline-none;
}

.select-dropdown-portal .select-options {
  @apply max-h-80 overflow-y-auto py-1 outline-none;
}

.select-dropdown-portal .select-option {
  @apply flex items-center justify-between gap-2;
  @apply px-4 py-2.5 text-sm;
  @apply text-gray-700 dark:text-gray-300;
  @apply cursor-pointer transition-colors duration-150;
  @apply hover:bg-gray-50 dark:hover:bg-dark-700;
  pointer-events: auto !important;
}

.select-dropdown-portal .select-option-selected {
  @apply bg-primary-50 dark:bg-primary-900/20;
  @apply text-primary-700 dark:text-primary-300;
}

.select-dropdown-portal .select-option-focused {
  @apply bg-gray-100 dark:bg-dark-700;
}

.select-dropdown-portal .select-option-disabled {
  @apply cursor-not-allowed opacity-40;
}

.select-dropdown-portal .select-option-group {
  @apply cursor-default select-none;
  @apply bg-gray-50 dark:bg-dark-900;
  @apply text-[11px] font-bold uppercase tracking-wider;
  @apply text-gray-500 dark:text-gray-400;
}

.select-dropdown-portal .select-option-group:hover {
  @apply bg-gray-50 dark:bg-dark-900;
}

.select-dropdown-portal .select-option-label {
  @apply flex-1 min-w-0 truncate text-left;
}

.select-dropdown-portal .select-option-check {
  color: #14b8a6;
}

.select-dropdown-portal .select-empty {
  @apply px-4 py-8 text-center text-sm;
  @apply text-gray-500 dark:text-dark-400;
}

.select-dropdown-enter-active,
.select-dropdown-leave-active {
  transition: all 0.2s ease;
}

.select-dropdown-enter-from,
.select-dropdown-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

.scheme3-select-trigger {
  display: flex;
  width: 100%;
  min-height: 2.25rem;
  cursor: pointer;
  align-items: center;
  justify-content: space-between;
  gap: .5rem;
  border: 1px solid #dad5c8;
  border-radius: 7px;
  padding: .52rem .72rem;
  background: #fbfaf6;
  color: #27251f;
  font-size: .72rem;
  line-height: 1.25;
  transition: border-color 150ms ease, background-color 150ms ease, color 150ms ease, box-shadow 150ms ease;
}

.scheme3-select-trigger:hover,
.scheme3-select-trigger-open {
  border-color: #1e5c42;
  background: #fffefa;
  box-shadow: 0 0 0 2px rgba(30, 92, 66, .1);
}

.scheme3-select-trigger:focus-visible {
  outline: 2px solid rgba(30, 92, 66, .28);
  outline-offset: 2px;
}

.scheme3-select-trigger-error { border-color: #9e4d3d; }
.scheme3-select-trigger-disabled { cursor: not-allowed; background: #f1eee6; opacity: .58; }
.scheme3-select-value { min-width: 0; flex: 1 1 auto; overflow: hidden; text-align: left; text-overflow: ellipsis; white-space: nowrap; }
.scheme3-select-icon { display: inline-flex; flex: 0 0 auto; color: #777266; }
.scheme3-select-clear { display: inline-flex; flex: 0 0 auto; cursor: pointer; align-items: center; justify-content: center; border-radius: 4px; color: #777266; transition: color 120ms ease, background-color 120ms ease; }
.scheme3-select-clear:hover { background: #f1eee6; color: #27251f; }

.scheme3-select-dropdown-portal {
  box-sizing: border-box;
  width: max-content;
  min-width: 200px;
  overflow: hidden;
  border: 1px solid #dad5c8;
  border-radius: 7px;
  background: #fbfaf6;
  color: #27251f;
  box-shadow: 0 16px 34px rgba(54, 48, 34, .16);
  pointer-events: auto !important;
}

.scheme3-select-search { display: flex; align-items: center; gap: .5rem; border-bottom: 1px solid #dad5c8; padding: .55rem .7rem; }
.scheme3-select-search-icon { flex: 0 0 auto; color: #777266; }
.scheme3-select-search-input { min-width: 0; flex: 1 1 auto; border: 0; background: transparent; color: #27251f; font-size: .72rem; outline: none; }
.scheme3-select-search-input::placeholder { color: #a49e90; }
.scheme3-select-options { max-height: 20rem; overflow-y: auto; padding: .3rem; outline: none; }
.scheme3-select-option { display: flex; min-height: 2.2rem; cursor: pointer; align-items: center; justify-content: space-between; gap: .55rem; border: 1px solid transparent; border-radius: 5px; padding: .48rem .65rem; color: #655f53; font-size: .72rem; transition: border-color 120ms ease, background-color 120ms ease, color 120ms ease; pointer-events: auto !important; }
.scheme3-select-option:hover,
.scheme3-select-option-focused { border-color: rgba(30, 92, 66, .14); background: #f1eee6; color: #27251f; }
.scheme3-select-option-selected { border-color: rgba(30, 92, 66, .2); background: rgba(30, 92, 66, .09); color: #1e5c42; font-weight: 700; }
.scheme3-select-option-disabled { cursor: not-allowed; opacity: .42; }
.scheme3-select-option-group,
.scheme3-select-option-group:hover { cursor: default; border-color: transparent; background: #f1eee6; color: #777266; font-size: .58rem; font-weight: 800; letter-spacing: .055em; text-transform: uppercase; }
.scheme3-select-option-label { min-width: 0; flex: 1 1 auto; overflow: hidden; text-align: left; text-overflow: ellipsis; white-space: nowrap; }
.scheme3-select-option-creatable { color: #777266; font-style: italic; }
.scheme3-select-option-creatable-icon { flex: 0 0 auto; color: #777266; }
.scheme3-select-option-check { flex: 0 0 auto; color: #1e5c42; }
.scheme3-select-empty { padding: 1.4rem .8rem; color: #777266; font-size: .7rem; text-align: center; }

.scheme3-select-dropdown-enter-active,
.scheme3-select-dropdown-leave-active { transition: opacity 150ms ease, transform 150ms ease; }
.scheme3-select-dropdown-enter-from,
.scheme3-select-dropdown-leave-to { opacity: 0; transform: translateY(-5px); }

html.dark .scheme3-select-trigger,
.scheme3-select-trigger.scheme3-select-theme-dark { border-color: #47443a; background: #24231f; color: #f4f2ec; }
html.dark .scheme3-select-trigger:hover,
html.dark .scheme3-select-trigger-open,
.scheme3-select-trigger.scheme3-select-theme-dark:hover,
.scheme3-select-trigger-open.scheme3-select-theme-dark { border-color: #8fc2a5; background: #2b2924; box-shadow: 0 0 0 2px rgba(143, 194, 165, .12); }
html.dark .scheme3-select-trigger:focus-visible,
.scheme3-select-trigger.scheme3-select-theme-dark:focus-visible { outline-color: rgba(143, 194, 165, .38); }
html.dark .scheme3-select-trigger-disabled,
.scheme3-select-trigger-disabled.scheme3-select-theme-dark { background: #1b1b18; }
html.dark .scheme3-select-icon,
.scheme3-select-trigger.scheme3-select-theme-dark .scheme3-select-icon { color: #aaa69a; }
html.dark .scheme3-select-clear,
.scheme3-select-trigger.scheme3-select-theme-dark .scheme3-select-clear { color: #aaa69a; }
html.dark .scheme3-select-clear:hover,
.scheme3-select-trigger.scheme3-select-theme-dark .scheme3-select-clear:hover { background: #2b2924; color: #f4f2ec; }

html.dark .scheme3-select-dropdown-portal,
.scheme3-select-dropdown-portal.scheme3-select-theme-dark { border-color: #47443a; background: #24231f; color: #f4f2ec; box-shadow: 0 18px 38px rgba(0, 0, 0, .28); }
html.dark .scheme3-select-search,
.scheme3-select-dropdown-portal.scheme3-select-theme-dark .scheme3-select-search { border-bottom-color: #47443a; }
html.dark .scheme3-select-search-icon,
.scheme3-select-dropdown-portal.scheme3-select-theme-dark .scheme3-select-search-icon { color: #aaa69a; }
html.dark .scheme3-select-search-input,
.scheme3-select-dropdown-portal.scheme3-select-theme-dark .scheme3-select-search-input { color: #f4f2ec; }
html.dark .scheme3-select-search-input::placeholder,
.scheme3-select-dropdown-portal.scheme3-select-theme-dark .scheme3-select-search-input::placeholder { color: #827e72; }
html.dark .scheme3-select-option,
.scheme3-select-dropdown-portal.scheme3-select-theme-dark .scheme3-select-option { color: #aaa69a; }
html.dark .scheme3-select-option:hover,
html.dark .scheme3-select-option-focused,
.scheme3-select-dropdown-portal.scheme3-select-theme-dark .scheme3-select-option:hover,
.scheme3-select-dropdown-portal.scheme3-select-theme-dark .scheme3-select-option-focused { border-color: rgba(143, 194, 165, .16); background: #2b2924; color: #f4f2ec; }
html.dark .scheme3-select-option-selected,
.scheme3-select-dropdown-portal.scheme3-select-theme-dark .scheme3-select-option-selected { border-color: rgba(143, 194, 165, .22); background: rgba(143, 194, 165, .1); color: #8fc2a5; }
html.dark .scheme3-select-option-group,
html.dark .scheme3-select-option-group:hover,
.scheme3-select-dropdown-portal.scheme3-select-theme-dark .scheme3-select-option-group,
.scheme3-select-dropdown-portal.scheme3-select-theme-dark .scheme3-select-option-group:hover { background: #2b2924; color: #aaa69a; }
html.dark .scheme3-select-option-check,
.scheme3-select-dropdown-portal.scheme3-select-theme-dark .scheme3-select-option-check { color: #8fc2a5; }
html.dark .scheme3-select-empty,
.scheme3-select-dropdown-portal.scheme3-select-theme-dark .scheme3-select-empty { color: #aaa69a; }
</style>

<template>
  <div class="locale-switcher" ref="dropdownRef" @click.stop>
    <button
      type="button"
      @click="toggleDropdown"
      :disabled="switching"
      class="locale-trigger"
      :class="{ 'is-open': isOpen, 'is-switching': switching }"
      aria-haspopup="menu"
      :aria-expanded="isOpen"
      :title="currentLocale?.name"
    >
      <span class="locale-trigger-flag" aria-hidden="true">{{ currentLocale?.flag }}</span>
      <span class="locale-trigger-code">{{ currentLocale?.code.toUpperCase() }}</span>
      <Icon name="chevronDown" size="xs" class="locale-trigger-chevron" />
    </button>

    <transition name="dropdown">
      <div
        v-if="isOpen"
        class="locale-menu"
        role="menu"
        :aria-label="locale === 'zh' ? '选择语言' : 'Choose language'"
      >
        <button
          v-for="locale in availableLocales"
          :key="locale.code"
          type="button"
          :disabled="switching"
          @click="selectLocale(locale.code)"
          class="locale-option"
          role="menuitemradio"
          :aria-checked="locale.code === currentLocaleCode"
          :class="{
            'is-current':
              locale.code === currentLocaleCode
          }"
        >
          <span class="locale-option-flag" aria-hidden="true">{{ locale.flag }}</span>
          <span class="locale-option-name">{{ locale.name }}</span>
          <Icon v-if="locale.code === currentLocaleCode" name="check" size="sm" class="locale-option-check" />
        </button>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { setLocale, availableLocales } from '@/i18n'

const { locale } = useI18n()

const isOpen = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)
const switching = ref(false)

const currentLocaleCode = computed(() => locale.value)
const currentLocale = computed(() => availableLocales.find((l) => l.code === locale.value))

function toggleDropdown() {
  if (switching.value) return
  isOpen.value = !isOpen.value
}

async function selectLocale(code: string) {
  if (switching.value || code === currentLocaleCode.value) {
    isOpen.value = false
    return
  }
  switching.value = true
  try {
    await setLocale(code)
    isOpen.value = false
  } finally {
    switching.value = false
  }
}

function handleClickOutside(event: MouseEvent) {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target as Node)) {
    isOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('keydown', handleEscape)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleEscape)
})

function handleEscape(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    isOpen.value = false
  }
}
</script>

<style scoped>
.locale-switcher {
  position: relative;
  display: inline-flex;
  min-width: 0;
  color: var(--home-ink, #16150f);
}

.locale-trigger {
  position: relative;
  display: inline-flex;
  min-height: 2.25rem;
  align-items: center;
  gap: 0.42rem;
  border: 1px solid var(--home-line, #dad5c8);
  background: color-mix(in srgb, var(--home-card, #fbfaf6) 82%, transparent);
  color: var(--home-ink, #16150f);
  padding: 0.45rem 0.6rem;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 0.7rem;
  font-weight: 600;
  letter-spacing: 0.08em;
  line-height: 1;
  cursor: pointer;
  transition: border-color 160ms ease, background 160ms ease, color 160ms ease, transform 160ms ease, box-shadow 160ms ease;
}

.locale-trigger:hover:not(:disabled),
.locale-trigger.is-open {
  border-color: var(--home-green, #1e5c42);
  background: color-mix(in srgb, var(--home-green, #1e5c42) 10%, var(--home-card, #fbfaf6));
  color: var(--home-green, #1e5c42);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--home-green, #1e5c42) 12%, transparent);
}

.locale-trigger:active:not(:disabled) {
  transform: translateY(1px) scale(0.97);
}

.locale-trigger:focus-visible,
.locale-option:focus-visible {
  outline: 2px solid var(--home-green, #1e5c42);
  outline-offset: 2px;
}

.locale-trigger:disabled,
.locale-option:disabled {
  cursor: wait;
  opacity: 0.58;
}

.locale-trigger-flag,
.locale-option-flag {
  font-family: "Segoe UI Emoji", "Apple Color Emoji", sans-serif;
  letter-spacing: 0;
}

.locale-trigger-code {
  display: inline-block;
}

.locale-trigger-chevron {
  color: var(--home-muted, #6b695f);
  transition: transform 180ms ease, color 160ms ease;
}

.locale-trigger.is-open .locale-trigger-chevron {
  color: var(--home-green, #1e5c42);
  transform: rotate(180deg);
}

.locale-trigger.is-switching::after {
  position: absolute;
  right: 0.25rem;
  bottom: 0.18rem;
  left: 0.25rem;
  height: 1px;
  background: var(--home-amber, #b7791f);
  content: "";
  animation: locale-progress 800ms ease-in-out infinite;
}

.locale-menu {
  position: absolute;
  top: calc(100% + 0.45rem);
  right: 0;
  z-index: 60;
  display: grid;
  width: 9.5rem;
  overflow: hidden;
  border: 1px solid var(--home-line, #dad5c8);
  background: color-mix(in srgb, var(--home-card, #fbfaf6) 96%, transparent);
  box-shadow: 0 1rem 2.5rem color-mix(in srgb, var(--home-ink, #16150f) 16%, transparent);
  backdrop-filter: blur(16px);
}

.locale-menu::before {
  position: absolute;
  top: 0;
  left: 0;
  width: 2.6rem;
  height: 2px;
  background: var(--home-amber, #b7791f);
  content: "";
  animation: locale-signal 2.8s ease-in-out infinite;
}

.locale-option {
  position: relative;
  display: flex;
  min-height: 2.6rem;
  align-items: center;
  gap: 0.6rem;
  border: 0;
  border-bottom: 1px solid color-mix(in srgb, var(--home-line, #dad5c8) 68%, transparent);
  background: transparent;
  color: var(--home-ink, #16150f);
  padding: 0.6rem 0.75rem;
  font-size: 0.78rem;
  text-align: left;
  cursor: pointer;
  transition: background 160ms ease, color 160ms ease, padding-left 160ms ease, transform 160ms ease;
}

.locale-option:last-child { border-bottom: 0; }

.locale-option:hover:not(:disabled),
.locale-option.is-current {
  background: color-mix(in srgb, var(--home-green, #1e5c42) 12%, var(--home-card, #fbfaf6));
  color: var(--home-green, #1e5c42);
  padding-left: 0.95rem;
}

.locale-option:active:not(:disabled) {
  transform: scale(0.98);
}

.locale-option-check {
  margin-left: auto;
  color: var(--home-green, #1e5c42);
  animation: locale-check-in 240ms cubic-bezier(0.22, 1, 0.36, 1) both;
}

@keyframes locale-progress {
  0% { transform: scaleX(0.18); transform-origin: left; opacity: 0.2; }
  50% { transform: scaleX(0.65); transform-origin: left; opacity: 1; }
  100% { transform: scaleX(1); transform-origin: left; opacity: 0.2; }
}

@keyframes locale-signal {
  0%, 100% { transform: translateX(0); opacity: 0.35; }
  50% { transform: translateX(5.8rem); opacity: 1; }
}

@keyframes locale-check-in {
  from { opacity: 0; transform: scale(0.6) rotate(-12deg); }
  to { opacity: 1; transform: scale(1) rotate(0); }
}

.dropdown-enter-active,
.dropdown-leave-active {
  transform-origin: top right;
  transition: opacity 180ms ease, transform 220ms cubic-bezier(0.22, 1, 0.36, 1);
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: scale(0.94) translateY(-0.35rem);
}

@media (max-width: 480px) {
  .locale-trigger-code { display: none; }
  .locale-trigger { gap: 0.25rem; padding-inline: 0.55rem; }
  .locale-menu {
    right: auto;
    left: 0;
    transform-origin: top left;
  }
}

@media (prefers-reduced-motion: reduce) {
  .locale-trigger,
  .locale-trigger-chevron,
  .locale-option,
  .locale-option-check,
  .dropdown-enter-active,
  .dropdown-leave-active,
  .locale-menu::before,
  .locale-trigger.is-switching::after {
    animation-duration: 1ms !important;
    transition-duration: 1ms !important;
  }
}
</style>

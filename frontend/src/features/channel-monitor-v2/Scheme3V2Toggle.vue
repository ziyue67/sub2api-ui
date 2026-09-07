<template>
  <button
    type="button"
    class="scheme3-v2-toggle"
    :class="{ 'is-on': modelValue }"
    role="switch"
    :aria-checked="modelValue"
    @click="emit('update:modelValue', !modelValue)"
  >
    <span class="scheme3-v2-toggle-track" aria-hidden="true">
      <span class="scheme3-v2-toggle-thumb"></span>
    </span>
  </button>
</template>

<script setup lang="ts">
defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  (event: 'update:modelValue', value: boolean): void
}>()
</script>

<style scoped>
.scheme3-v2-toggle {
  --toggle-line: #b9b2a4;
  --toggle-off: #e5e1d8;
  --toggle-on: #1e5c42;
  --toggle-thumb: #fffefa;
  display: inline-flex;
  width: 2.75rem;
  height: 1.5rem;
  flex: none;
  align-items: center;
  border: 1px solid var(--toggle-line);
  border-radius: 999px;
  padding: .15rem;
  background: var(--toggle-off);
  cursor: pointer;
  transition: border-color 150ms ease, background-color 150ms ease, box-shadow 150ms ease;
}
.scheme3-v2-toggle:hover { border-color: var(--toggle-on); }
.scheme3-v2-toggle:focus-visible { outline: 2px solid rgba(30, 92, 66, .32); outline-offset: 2px; }
.scheme3-v2-toggle-track { display: block; width: 100%; height: 100%; }
.scheme3-v2-toggle-thumb {
  display: block;
  width: 1.1rem;
  height: 1.1rem;
  border-radius: 50%;
  background: var(--toggle-thumb);
  box-shadow: 0 1px 4px rgba(34, 31, 24, .18);
  transform: translateX(0);
  transition: transform 150ms ease, background-color 150ms ease;
}
.scheme3-v2-toggle.is-on { border-color: var(--toggle-on); background: var(--toggle-on); }
.scheme3-v2-toggle.is-on .scheme3-v2-toggle-thumb { transform: translateX(1.2rem); }
:global(.dark .scheme3-v2-toggle) {
  --toggle-line: #5c584d;
  --toggle-off: #343129;
  --toggle-on: #8fc2a5;
  --toggle-thumb: #1b1b18;
}
:global(.dark .scheme3-v2-toggle:focus-visible) { outline-color: rgba(143, 194, 165, .46); }
</style>

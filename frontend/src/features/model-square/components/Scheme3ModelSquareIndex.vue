<template>
  <aside
    ref="indexRef"
    class="scheme3-model-square-index"
    aria-label="模型索引"
    @keydown="onKeydown"
  >
    <header class="scheme3-model-square-index-header">
      <div>
        <span class="scheme3-model-square-index-kicker">目录导航</span>
        <h2>模型索引</h2>
      </div>
      <span class="scheme3-model-square-index-total">{{ models.length }} 个</span>
    </header>

    <div class="scheme3-model-square-index-help">
      <span>按平台分组</span>
      <span class="scheme3-model-square-index-keys">↑ ↓ 选择 · Enter 跳转</span>
    </div>

    <div class="scheme3-model-square-index-list">
      <template v-for="group in groupedModels" :key="group.platform">
        <div class="scheme3-model-square-index-platform">
          <span
            class="scheme3-model-square-index-platform-mark"
            :style="{ '--index-platform-color': platformColor(group.platform) }"
            aria-hidden="true"
          ></span>
          <PlatformIcon :platform="group.platform" size="xs" />
          <span>{{ platformLabel(group.platform) }}</span>
          <small>{{ group.items.length }}</small>
        </div>

        <button
          v-for="item in group.items"
          :key="item.model.key"
          type="button"
          class="scheme3-model-square-index-item"
          :class="{ 'scheme3-model-square-index-item-active': modelValue === item.model.key }"
          :data-model-index="item.index"
          :data-model-key="item.model.key"
          :tabindex="selectedIndex === item.index ? 0 : -1"
          :aria-current="modelValue === item.model.key ? 'true' : undefined"
          :title="item.model.name"
          @click="selectModel(item.model.key, item.index)"
        >
          <span class="scheme3-model-square-index-item-name">{{ item.model.name }}</span>
          <span class="scheme3-model-square-index-item-meta">
            {{ item.model.channels.length }} 渠道
          </span>
        </button>
      </template>

      <p v-if="models.length === 0" class="scheme3-model-square-index-empty">没有匹配的模型</p>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import type { ModelSquareModel } from '../types'
import { platformAccentColor, platformLabel } from '@/utils/platformColors'

interface Props {
  models: ModelSquareModel[]
  modelValue: string
}

const props = defineProps<Props>()
const emit = defineEmits<{
  'update:modelValue': [value: string]
  select: [value: string]
}>()

const indexRef = ref<HTMLElement | null>(null)
const selectedIndex = ref(0)

const groupedModels = computed(() => {
  const groups: { platform: string; items: { model: ModelSquareModel; index: number }[] }[] = []
  const map = new Map<string, { platform: string; items: { model: ModelSquareModel; index: number }[] }>()
  props.models.forEach((model) => {
    let group = map.get(model.platform)
    if (!group) {
      group = { platform: model.platform, items: [] }
      map.set(model.platform, group)
      groups.push(group)
    }
    group.items.push({ model, index: 0 })
  })
  let orderedIndex = 0
  groups.forEach((group) => {
    group.items.forEach((item) => {
      item.index = orderedIndex
      orderedIndex += 1
    })
  })
  return groups
})

const orderedModels = computed(() => groupedModels.value.flatMap((group) => group.items.map((item) => item.model)))

function platformColor(platform: string) {
  return platformAccentColor(platform)
}

function syncSelectedIndex() {
  const activeIndex = orderedModels.value.findIndex((model) => model.key === props.modelValue)
  if (activeIndex >= 0) {
    selectedIndex.value = activeIndex
    return
  }
  selectedIndex.value = Math.min(selectedIndex.value, Math.max(orderedModels.value.length - 1, 0))
}

function selectModel(key: string, index?: number) {
  if (index != null) selectedIndex.value = index
  emit('update:modelValue', key)
  emit('select', key)
}

function focusSelected() {
  void nextTick(() => {
    const button = indexRef.value?.querySelector<HTMLButtonElement>(
      `[data-model-index="${selectedIndex.value}"]`,
    )
    button?.focus({ preventScroll: true })
    button?.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
  })
}

function onKeydown(event: KeyboardEvent) {
  if (orderedModels.value.length === 0) return
  if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
    event.preventDefault()
    const step = event.key === 'ArrowDown' ? 1 : -1
    selectedIndex.value = (selectedIndex.value + step + orderedModels.value.length) % orderedModels.value.length
    focusSelected()
    return
  }
  if (event.key === 'Home' || event.key === 'End') {
    event.preventDefault()
    selectedIndex.value = event.key === 'Home' ? 0 : orderedModels.value.length - 1
    focusSelected()
    return
  }
  if (event.key === 'Enter') {
    event.preventDefault()
    const model = orderedModels.value[selectedIndex.value]
    if (model) selectModel(model.key, selectedIndex.value)
  }
}

watch(() => props.modelValue, syncSelectedIndex, { immediate: true })
watch(() => props.models, syncSelectedIndex)
</script>

<style scoped>
.scheme3-model-square-index {
  position: sticky;
  top: calc(var(--scheme3-console-topbar-height, 4.25rem) + .75rem);
  display: flex;
  min-width: 0;
  max-height: calc(100vh - var(--scheme3-console-topbar-height, 4.25rem) - 1.5rem);
  flex-direction: column;
  overflow: hidden;
  border: 1px solid var(--square-line, #dad5c8);
  border-radius: 8px;
  background: var(--square-card, #fbfaf6);
  color: var(--square-ink, #16150f);
  box-shadow: 0 8px 20px rgba(54, 48, 34, .05);
}

.scheme3-model-square-index-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: .7rem;
  border-bottom: 1px solid var(--square-line, #dad5c8);
  padding: .85rem .85rem .7rem;
}

.scheme3-model-square-index-kicker {
  display: block;
  color: var(--square-muted, #6b695f);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: .625rem;
  font-weight: 800;
  letter-spacing: .1em;
  text-transform: uppercase;
}

.scheme3-model-square-index h2 {
  margin: .22rem 0 0;
  font-family: Georgia, 'Times New Roman', serif;
  font-size: 1.08rem;
  font-weight: 600;
}

.scheme3-model-square-index-total {
  color: var(--square-muted, #6b695f);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: .625rem;
  white-space: nowrap;
}

.scheme3-model-square-index-help {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .5rem;
  border-bottom: 1px solid var(--square-line, #dad5c8);
  padding: .55rem .85rem;
  color: var(--square-muted, #6b695f);
  font-size: .625rem;
}

.scheme3-model-square-index-keys {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  white-space: nowrap;
}

.scheme3-model-square-index-list {
  min-height: 0;
  overflow-y: auto;
  padding: .35rem;
  scrollbar-width: thin;
  scrollbar-color: rgba(30, 92, 66, .48) transparent;
}

.scheme3-model-square-index-platform {
  display: flex;
  align-items: center;
  gap: .38rem;
  margin: .45rem .35rem .2rem;
  color: var(--square-muted, #6b695f);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: .625rem;
  font-weight: 800;
  letter-spacing: .06em;
  text-transform: uppercase;
}

.scheme3-model-square-index-platform:first-child { margin-top: .2rem; }

.scheme3-model-square-index-platform-mark {
  width: .35rem;
  height: .35rem;
  flex: 0 0 auto;
  border-radius: 50%;
  background: var(--index-platform-color, #1e5c42);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--index-platform-color, #1e5c42) 16%, transparent);
}

.scheme3-model-square-index-platform small {
  margin-left: auto;
  font-size: .625rem;
  font-weight: 700;
  letter-spacing: 0;
}

.scheme3-model-square-index-item {
  display: flex;
  width: 100%;
  min-width: 0;
  align-items: baseline;
  justify-content: space-between;
  gap: .5rem;
  border: 1px solid transparent;
  border-radius: 5px;
  padding: .48rem .52rem;
  background: transparent;
  color: var(--square-ink, #16150f);
  text-align: left;
  transition: background-color 140ms ease, border-color 140ms ease, color 140ms ease;
}

.scheme3-model-square-index-item:hover {
  border-color: rgba(30, 92, 66, .32);
  background: rgba(30, 92, 66, .06);
}

.scheme3-model-square-index-item:focus-visible {
  outline: 2px solid rgba(30, 92, 66, .32);
  outline-offset: -1px;
}

.scheme3-model-square-index-item-active {
  border-color: rgba(30, 92, 66, .42);
  background: rgba(30, 92, 66, .1);
  color: #174a35;
}

.scheme3-model-square-index-item-name {
  min-width: 0;
  overflow: hidden;
  font-size: .68rem;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.scheme3-model-square-index-item-meta {
  flex: 0 0 auto;
  color: var(--square-muted, #6b695f);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: .625rem;
  white-space: nowrap;
}

.scheme3-model-square-index-empty {
  margin: 1rem .5rem;
  color: var(--square-muted, #6b695f);
  font-size: .68rem;
  text-align: center;
}

:global(html.dark .scheme3-model-square-index) {
  border-color: #47443a;
  background: #24231f;
  color: #f4f2ec;
  box-shadow: 0 8px 20px rgba(0, 0, 0, .18);
}

:global(html.dark .scheme3-model-square-index-header),
:global(html.dark .scheme3-model-square-index-help) {
  border-color: #47443a;
}

:global(html.dark .scheme3-model-square-index-item-active) {
  border-color: rgba(143, 194, 165, .5);
  background: rgba(143, 194, 165, .13);
  color: #a7d0b8;
}

:global(html.dark .scheme3-model-square-index-item:hover) {
  border-color: rgba(143, 194, 165, .35);
  background: rgba(143, 194, 165, .08);
}

@media (max-width: 900px) {
  .scheme3-model-square-index { display: none; }
}
</style>

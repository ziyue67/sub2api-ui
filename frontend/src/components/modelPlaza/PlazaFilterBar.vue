<template>
  <div class="scheme3-model-plaza-filters space-y-5 p-1">
    <!-- 一级:平台 -->
    <div class="flex flex-wrap items-center gap-3">
      <span class="w-12 shrink-0 text-[10px] font-bold uppercase tracking-widest text-gray-400 dark:text-dark-500">
        {{ t('modelPlaza.filters.platformLabel') }}
      </span>
      <button
        v-for="p in ['all', ...platforms]"
        :key="`platform-${p}`"
        type="button"
        class="inline-flex items-center gap-2 rounded-xl px-4 py-2 text-sm font-semibold transition-all duration-200 disabled:cursor-not-allowed disabled:opacity-30"
        :class="chipClass(platform === p)"
        :disabled="p !== 'all' && !platformEnabled(p)"
        @click="$emit('update:platform', p)"
      >
        <PlatformIcon v-if="p !== 'all'" :platform="p as GroupPlatform" size="xs" class="scale-110" />
        {{ p === 'all' ? t('modelPlaza.filters.all') : p }}
      </button>
    </div>

    <!-- 二级:分组 -->
    <div class="flex flex-wrap items-center gap-3">
      <span class="w-12 shrink-0 text-[10px] font-bold uppercase tracking-widest text-gray-400 dark:text-dark-500">
        {{ t('modelPlaza.filters.groupLabel') }}
      </span>
      <button
        type="button"
        class="rounded-xl px-4 py-2 text-sm font-semibold transition-all duration-200"
        :class="chipClass(groupId === 'all')"
        @click="$emit('update:groupId', 'all')"
      >
        {{ t('modelPlaza.filters.all') }}
      </button>
      <button
        v-for="g in groups"
        :key="`group-${g.id}`"
        type="button"
        class="rounded-xl px-4 py-2 text-sm font-semibold transition-all duration-200 disabled:opacity-30"
        :class="chipClass(groupId === g.id)"
        :disabled="!groupEnabled(g)"
        @click="$emit('update:groupId', g.id)"
      >
        {{ g.name }}
      </button>
    </div>

    <!-- 三级:倍率 -->
    <div class="flex flex-wrap items-center gap-3">
      <span class="w-12 shrink-0 text-[10px] font-bold uppercase tracking-widest text-gray-400 dark:text-dark-500">
        {{ t('modelPlaza.filters.rateLabel') }}
      </span>
      <button
        type="button"
        class="rounded-xl px-4 py-2 text-sm font-semibold transition-all duration-200"
        :class="chipClass(rate === 'all')"
        @click="$emit('update:rate', 'all')"
      >
        {{ t('modelPlaza.filters.all') }}
      </button>
      <button
        v-for="r in rates"
        :key="`rate-${r}`"
        type="button"
        class="rounded-xl px-4 py-2 font-mono text-sm font-bold transition-all duration-200 disabled:opacity-30"
        :class="chipClass(rate === r)"
        :disabled="!rateEnabled(r)"
        @click="$emit('update:rate', r)"
      >
        {{ r }}x
      </button>
    </div>

    <!-- 四级:搜索 -->
    <div class="flex flex-wrap items-center gap-3 pt-2">
      <span class="w-12 shrink-0 text-[10px] font-bold uppercase tracking-widest text-gray-400 dark:text-dark-500">
        {{ t('modelPlaza.filters.modelLabel') }}
      </span>
      <div class="relative w-full sm:w-80 group">
        <Icon
          name="search"
          size="sm"
          class="absolute left-3.5 top-1/2 -translate-y-1/2 text-gray-400 transition-colors group-focus-within:text-primary-500"
        />
        <input
          :value="search"
          type="text"
          :placeholder="t('modelPlaza.filters.searchPlaceholder')"
          class="scheme3-model-plaza-search w-full"
          @input="$emit('update:search', ($event.target as HTMLInputElement).value)"
        />
        <button
          v-if="search"
          type="button"
          class="absolute right-2.5 top-1/2 -translate-y-1/2 text-gray-400 transition-colors hover:text-gray-600 dark:text-dark-500 dark:hover:text-gray-300"
          @click="$emit('update:search', '')"
        >
          <Icon name="x" size="xs" class="h-3.5 w-3.5" />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import type { GroupPlatform } from '@/types'

const props = defineProps<{
  /** 数据中出现的平台(去重排序后)。 */
  platforms: string[]
  /** 全量分组(含平台与生效倍率),三个维度的置灰联动由此推导。 */
  groups: Array<{ id: number; name: string; platform: string; rate: number }>
  /** 全量生效倍率去重升序。 */
  rates: number[]
  platform: string
  groupId: number | 'all'
  rate: number | 'all'
  /** 模型名搜索词(纯前端过滤)。 */
  search: string
}>()

defineEmits<{
  'update:platform': [value: string]
  'update:groupId': [value: number | 'all']
  'update:rate': [value: number | 'all']
  'update:search': [value: string]
}>()

const { t } = useI18n()

/**
 * 三个维度互为约束(faceted):某选项可点 ⟺ 在「其他两维」当前选择下仍有分组命中。
 * 「全部」永远可点,作为解除本维约束的出口;可点项组合恒有结果,无需选择修正。
 */
function platformEnabled(p: string): boolean {
  return props.groups.some(
    (g) =>
      g.platform === p &&
      (props.groupId === 'all' || g.id === props.groupId) &&
      (props.rate === 'all' || g.rate === props.rate)
  )
}

function groupEnabled(g: { platform: string; rate: number }): boolean {
  return (
    (props.platform === 'all' || g.platform === props.platform) &&
    (props.rate === 'all' || g.rate === props.rate)
  )
}

function rateEnabled(r: number): boolean {
  return props.groups.some(
    (g) =>
      g.rate === r &&
      (props.platform === 'all' || g.platform === props.platform) &&
      (props.groupId === 'all' || g.id === props.groupId)
  )
}

function chipClass(active: boolean): string {
  return active
    ? 'scheme3-model-plaza-chip-active'
    : 'scheme3-model-plaza-chip'
}
</script>

<style scoped>
.scheme3-model-plaza-filters { --filter-ink: var(--scheme3-ink,#16150f); --filter-muted: var(--scheme3-muted,#6b695f); --filter-line: var(--scheme3-line,#dad5c8); --filter-card: var(--scheme3-card,#fbfaf6); color: var(--filter-ink); }
.scheme3-model-plaza-filters > div { gap: .4rem; }
.scheme3-model-plaza-filters > div > span { width: 3.5rem; color: var(--filter-muted); font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; font-size: .56rem; letter-spacing: 0; }
.scheme3-model-plaza-chip,.scheme3-model-plaza-chip-active { border-radius: 5px; padding: .35rem .65rem; font-size: .64rem; font-weight: 800; transition: background-color 150ms ease,border-color 150ms ease,color 150ms ease; }
.scheme3-model-plaza-chip { border: 1px solid var(--filter-line); background: var(--filter-card); color: var(--filter-muted); box-shadow: none; }
.scheme3-model-plaza-chip:not(:disabled):hover { border-color: rgba(30,92,66,.5); background: #eef2ec; color: #1e5c42; }
.scheme3-model-plaza-chip-active { border: 1px solid #1e5c42; background: #1e5c42; color: #f4f2ec; box-shadow: none; }
.scheme3-model-plaza-chip-active:not(:disabled):hover { background: #174a35; }
.scheme3-model-plaza-search { min-height: 2.35rem; border: 1px solid var(--filter-line); border-radius: 6px; padding: .55rem 2.2rem .55rem 2.6rem; background: var(--filter-card); color: var(--filter-ink); outline: 0; font-size: .7rem; }.scheme3-model-plaza-search:focus { border-color: #1e5c42; box-shadow: 0 0 0 3px rgba(30,92,66,.12); }.scheme3-model-plaza-search::placeholder { color: #979286; }
.scheme3-model-plaza-filters :deep(.text-primary-500) { color: #1e5c42; }
:global(html.dark .scheme3-model-plaza-filters) { --filter-ink: #f4f2ec; --filter-muted: #aaa69a; --filter-line: #47443a; --filter-card: #24231f; color: #f4f2ec; }
:global(html.dark .scheme3-model-plaza-chip),:global(html.dark .scheme3-model-plaza-search) { border-color: #47443a; background: #24231f; color: #dedbd1; }
:global(html.dark .scheme3-model-plaza-chip:not(:disabled):hover) { border-color: rgba(143,194,165,.55); background: #2b3028; color: #8fc2a5; }
:global(html.dark .scheme3-model-plaza-chip-active) { border-color: #8fc2a5; background: #1e5c42; color: #f4f2ec; }
:global(html.dark .scheme3-model-plaza-chip-active:not(:disabled):hover) { background: #286a4e; }
:global(html.dark .scheme3-model-plaza-filters .text-primary-500) { color: #8fc2a5; }
</style>

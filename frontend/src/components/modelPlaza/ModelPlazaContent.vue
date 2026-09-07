<template>
  <div class="scheme3-model-plaza space-y-5">
    <!-- 页头(独立形态下展示标题;后台形态 AppHeader 已有页面标题) -->
    <div v-if="!embedded">
      <h1 class="text-2xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-3xl">{{ t('modelPlaza.title') }}</h1>
      <p class="mt-1.5 text-sm text-gray-500 dark:text-dark-400">{{ t('modelPlaza.description') }}</p>
    </div>

    <!-- 全局价格说明(管理员配置,Markdown) -->
    <div
      v-if="descriptionHtml"
      class="scheme3-model-plaza-description plaza-description"
      v-html="descriptionHtml"
    ></div>

    <!-- 未登录提示 -->
    <p
      v-if="!isAuthenticated"
      class="flex items-center gap-1.5 text-xs text-gray-400 dark:text-dark-500"
    >
      <Icon name="infoCircle" size="xs" class="h-3.5 w-3.5" />
      {{ t('modelPlaza.anonymousHint') }}
    </p>

    <!-- 加载/错误/空 -->
    <div v-if="loading" class="flex min-h-[240px] items-center justify-center">
      <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-600/25 border-t-primary-600 dark:border-primary-400/25 dark:border-t-primary-400"></div>
    </div>
    <div
      v-else-if="error"
      class="scheme3-model-plaza-error"
    >
      {{ t('modelPlaza.loadFailed') }}
    </div>
    <template v-else>
      <!-- 筛选区:平台 → 分组 → 倍率 -->
      <PlazaFilterBar
        :platforms="platforms"
        :groups="groupOptions"
        :rates="rates"
        :platform="selectedPlatform"
        :group-id="selectedGroupId"
        :rate="selectedRate"
        :search="searchQuery"
        @update:platform="selectedPlatform = $event"
        @update:group-id="selectedGroupId = $event"
        @update:rate="selectedRate = $event"
        @update:search="searchQuery = $event"
      />

      <!-- 分组分节的模型清单(默认按生效倍率升序) -->
      <div v-if="filteredGroups.length > 0" class="space-y-5">
        <PlazaGroupSection v-for="g in filteredGroups" :key="g.id" :group="g" />
      </div>
      <div
        v-else
        class="scheme3-model-plaza-empty"
      >
        {{ searchActive ? t('modelPlaza.noSearchResult') : t('modelPlaza.empty') }}
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import Icon from '@/components/icons/Icon.vue'
import PlazaFilterBar from './PlazaFilterBar.vue'
import PlazaGroupSection from './PlazaGroupSection.vue'
import type { ModelPlazaGroup, ModelPlazaResponse } from '@/api/modelPlaza'
import { useAuthStore } from '@/stores/auth'

const props = defineProps<{
  response: ModelPlazaResponse | null
  loading: boolean
  error?: boolean
  /** 后台内嵌形态(AppLayout 内):隐藏页头。 */
  embedded?: boolean
}>()

const { t } = useI18n()
const authStore = useAuthStore()
const isAuthenticated = computed(() => authStore.isAuthenticated)

const selectedPlatform = ref<string>('all')
const selectedGroupId = ref<number | 'all'>('all')
const selectedRate = ref<number | 'all'>('all')
const searchQuery = ref('')

const searchActive = computed(() => searchQuery.value.trim() !== '')

const descriptionHtml = computed(() => {
  const md = props.response?.description?.trim()
  if (!md) return ''
  return DOMPurify.sanitize(marked.parse(md) as string)
})

/** 生效倍率 = 用户专属倍率 ?? 分组默认倍率。 */
function effectiveRate(g: ModelPlazaGroup): number {
  return g.user_rate_multiplier ?? g.rate_multiplier
}

const platforms = computed(() =>
  [...new Set((props.response?.groups ?? []).map((g) => g.platform).filter(Boolean))].sort()
)

const groupOptions = computed(() =>
  (props.response?.groups ?? []).map((g) => ({
    id: g.id,
    name: g.name,
    platform: g.platform,
    rate: effectiveRate(g)
  }))
)

/** 全量生效倍率;当前组合下不可用的项由 FilterBar 置灰而非隐藏。 */
const rates = computed(() =>
  [...new Set((props.response?.groups ?? []).map(effectiveRate))].sort((a, b) => a - b)
)

/** 数据刷新后选中的倍率可能不复存在,重置为全部。 */
watch(rates, (list) => {
  if (selectedRate.value !== 'all' && !list.includes(selectedRate.value)) {
    selectedRate.value = 'all'
  }
})

const filteredGroups = computed(() => {
  let groups = props.response?.groups ?? []
  if (selectedPlatform.value !== 'all') {
    groups = groups.filter((g) => g.platform === selectedPlatform.value)
  }
  if (selectedGroupId.value !== 'all') {
    groups = groups.filter((g) => g.id === selectedGroupId.value)
  }
  if (selectedRate.value !== 'all') {
    groups = groups.filter((g) => effectiveRate(g) === selectedRate.value)
  }
  // 模型名搜索:分组内只留命中的模型,整组无命中则隐藏该分组。
  const q = searchQuery.value.trim().toLowerCase()
  if (q) {
    groups = groups
      .map((g) => ({ ...g, models: g.models.filter((m) => m.name.toLowerCase().includes(q)) }))
      .filter((g) => g.models.length > 0)
  }
  // 专属倍率会改变生效值,不能只依赖后端按默认倍率的排序。
  return [...groups].sort(
    (a, b) => effectiveRate(a) - effectiveRate(b) || a.name.localeCompare(b.name)
  )
})
</script>

<style scoped>
.scheme3-model-plaza { --plaza-ink: var(--scheme3-ink,#16150f); --plaza-muted: var(--scheme3-muted,#6b695f); --plaza-line: var(--scheme3-line,#dad5c8); --plaza-card: var(--scheme3-card,#fbfaf6); color: var(--plaza-ink); }
.scheme3-model-plaza-description { border: 1px solid var(--plaza-line); border-radius: 7px; padding: .8rem 1rem; background: #f4f2ec; box-shadow: none; font-size: .7rem; }
.scheme3-model-plaza-error { border: 1px solid rgba(158,77,61,.3); border-radius: 7px; padding: 1.5rem; background: rgba(158,77,61,.07); color: #9e4d3d; font-size: .72rem; text-align: center; }
.scheme3-model-plaza-empty { border: 1px dashed var(--plaza-line); border-radius: 7px; padding: 3rem 1.25rem; color: var(--plaza-muted); font-size: .72rem; text-align: center; }
.scheme3-model-plaza :deep(.scheme3-model-plaza-filters) { border-bottom: 1px solid var(--plaza-line); padding: .2rem 0 1rem; }
.scheme3-model-plaza :deep(.scheme3-model-plaza-group) { overflow: hidden; border: 1px solid var(--plaza-line) !important; border-radius: 7px; background: var(--plaza-card); box-shadow: 0 8px 18px rgba(54,48,34,.045); }
:global(html.dark .scheme3-model-plaza-description),:global(html.dark .scheme3-model-plaza-group) { border-color: #47443a; background: #24231f; color: #f4f2ec; }
:global(html.dark .scheme3-model-plaza-description) { background: #1b1b18; }
.plaza-description {
  line-height: 1.7;
  overflow-wrap: anywhere;
}

.plaza-description :deep(h1),
.plaza-description :deep(h2),
.plaza-description :deep(h3) {
  margin-top: .75rem;
  margin-bottom: .5rem;
  color: var(--plaza-ink);
  font-weight: 600;
}

.plaza-description :deep(h1:first-child),
.plaza-description :deep(h2:first-child),
.plaza-description :deep(h3:first-child) {
  margin-top: 0;
}

.plaza-description :deep(p) {
  margin-bottom: .5rem;
  color: var(--plaza-muted);
}

.plaza-description :deep(p:last-child) {
  margin-bottom: 0;
}

.plaza-description :deep(a) {
  color: #1e5c42;
  text-decoration: underline;
  text-underline-offset: 4px;
}

.plaza-description :deep(a:hover) {
  color: #174a35;
}

.plaza-description :deep(ul) {
  margin-bottom: .5rem;
  padding-left: 1.25rem;
  list-style: disc;
}

.plaza-description :deep(ol) {
  margin-bottom: .5rem;
  padding-left: 1.25rem;
  list-style: decimal;
}

.plaza-description :deep(li) {
  margin-bottom: .125rem;
  color: var(--plaza-muted);
}

.plaza-description :deep(code) {
  border-radius: 4px;
  padding: .125rem .375rem;
  background: var(--plaza-subtle, #f1eee6);
  color: var(--plaza-ink);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: .75rem;
}

.plaza-description :deep(blockquote) {
  margin: .5rem 0;
  border-left: 4px solid var(--plaza-line);
  padding-left: .75rem;
  color: var(--plaza-muted);
}

:global(html.dark .plaza-description a) {
  color: #8fc2a5;
}

:global(html.dark .plaza-description a:hover) {
  color: #a7d0b8;
}

:global(html.dark .plaza-description code) {
  background: var(--plaza-subtle, #2b2924);
}
</style>

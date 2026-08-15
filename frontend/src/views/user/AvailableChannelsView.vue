<template>
  <AppLayout>
    <section class="scheme3-available-channels">
      <header class="scheme3-available-header">
        <div>
          <p class="scheme3-available-kicker">公开目录 / 节点账本</p>
          <h1>可用节点</h1>
        </div>
        <div class="scheme3-available-ledger" aria-label="可用节点概览">
          <span><strong>{{ filteredChannels.length }}</strong><small>渠道</small></span>
          <span><strong>{{ visiblePlatformCount }}</strong><small>平台</small></span>
          <span><strong>{{ visibleModelCount }}</strong><small>模型</small></span>
          <span><strong>{{ visibleGroupCount }}</strong><small>分组</small></span>
        </div>
      </header>

    <TablePageLayout>
      <template #filters>
        <div class="scheme3-available-filters flex flex-col justify-between gap-4 lg:flex-row lg:items-start">
          <div class="flex flex-1 flex-wrap items-center gap-3">
            <div class="relative w-full sm:w-80">
              <Icon
                name="search"
                size="md"
                class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-gray-500"
              />
              <input
                v-model="searchQuery"
                type="text"
                :placeholder="t('availableChannels.searchPlaceholder')"
                class="input pl-10"
              />
            </div>
          </div>

          <div class="flex w-full flex-shrink-0 flex-wrap items-center justify-end gap-3 lg:w-auto">
            <button
              @click="loadChannels"
              :disabled="loading"
              class="btn btn-secondary"
              :title="t('common.refresh')"
            >
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <div class="scheme3-available-table">
          <AvailableChannelsTable
            :columns="columnLabels"
            :rows="filteredChannels"
            :loading="loading"
            :user-group-rates="userGroupRates"
            pricing-key-prefix="availableChannels.pricing"
            :no-pricing-label="t('availableChannels.noPricing')"
            :no-models-label="t('availableChannels.noModels')"
            :empty-label="t('availableChannels.empty')"
          />
        </div>
      </template>
    </TablePageLayout>
    </section>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import AvailableChannelsTable from '@/components/channels/AvailableChannelsTable.vue'
import userChannelsAPI, { type UserAvailableChannel } from '@/api/channels'
import userGroupsAPI from '@/api/groups'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n()
const appStore = useAppStore()

const channels = ref<UserAvailableChannel[]>([])
const userGroupRates = ref<Record<number, number>>({})
const loading = ref(false)
const searchQuery = ref('')

const columnLabels = computed(() => ({
  name: t('availableChannels.columns.name'),
  description: t('availableChannels.columns.description'),
  platform: t('availableChannels.columns.platform'),
  groups: t('availableChannels.columns.groups'),
  supportedModels: t('availableChannels.columns.supportedModels'),
}))

/**
 * 搜索过滤：
 * - 命中渠道名/描述 → 整个渠道（所有 platforms）都保留
 * - 否则按 platform/group/model 维度在 sections 里过滤，保留有匹配的 section
 * - 所有 sections 都不匹配时，渠道本身被过滤掉
 */
const filteredChannels = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return channels.value
  return channels.value
    .map((ch) => {
      const nameHit = ch.name.toLowerCase().includes(q)
      const descHit = (ch.description || '').toLowerCase().includes(q)
      if (nameHit || descHit) return ch
      const matchingSections = ch.platforms.filter(
        (p) =>
          p.platform.toLowerCase().includes(q) ||
          p.groups.some((g) => g.name.toLowerCase().includes(q)) ||
          p.supported_models.some((m) => m.name.toLowerCase().includes(q)),
      )
      if (matchingSections.length === 0) return null
      return { ...ch, platforms: matchingSections }
    })
    .filter((ch): ch is UserAvailableChannel => ch !== null)
})

const visiblePlatformCount = computed(() => filteredChannels.value.reduce((total, channel) => total + channel.platforms.length, 0))
const visibleModelCount = computed(() => {
  const models = new Set<string>()
  filteredChannels.value.forEach((channel) => channel.platforms.forEach((section) => section.supported_models.forEach((model) => models.add(`${section.platform}:${model.name}`))))
  return models.size
})
const visibleGroupCount = computed(() => {
  const groups = new Set<number>()
  filteredChannels.value.forEach((channel) => channel.platforms.forEach((section) => section.groups.forEach((group) => groups.add(group.id))))
  return groups.size
})

async function loadChannels() {
  loading.value = true
  try {
    // 渠道列表和用户专属倍率并发拉取。专属倍率失败不阻塞渠道展示——
    // 失败时只是无法渲染专属倍率角标，降级为仅显示默认倍率。
    const [list, rates] = await Promise.all([
      userChannelsAPI.getAvailable(),
      userGroupsAPI.getUserGroupRates().catch((err: unknown) => {
        console.error('Failed to load user group rates:', err)
        return {} as Record<number, number>
      }),
    ])
    channels.value = list
    userGroupRates.value = rates
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    loading.value = false
  }
}

onMounted(loadChannels)
</script>

<style scoped>
.scheme3-available-channels { --available-card: #fffefa; --available-ink: #27251f; --available-muted: #777266; --available-line: #d8d2c3; color: var(--available-ink); }
.scheme3-available-header { display: flex; align-items: end; justify-content: space-between; gap: 1.25rem; margin-bottom: 1.05rem; border-bottom: 1px solid var(--available-line); padding: .1rem 0 1rem; }
.scheme3-available-kicker { margin: 0; color: var(--available-muted); font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: .61rem; font-weight: 800; letter-spacing: .1em; }
.scheme3-available-channels h1 { margin: .34rem 0 0; color: var(--available-ink); font-family: Georgia, 'Times New Roman', serif; font-size: clamp(1.55rem, 2.6vw, 2.1rem); font-weight: 500; letter-spacing: 0; }
.scheme3-available-ledger { display: flex; flex-wrap: wrap; justify-content: flex-end; border: 1px solid var(--available-line); border-radius: 7px; background: var(--available-card); }
.scheme3-available-ledger span { display: grid; min-width: 4.5rem; gap: .06rem; border-right: 1px solid var(--available-line); padding: .48rem .68rem; text-align: right; }
.scheme3-available-ledger span:last-child { border-right: 0; }
.scheme3-available-ledger strong { color: #1e5c42; font-family: Georgia, 'Times New Roman', serif; font-size: 1.02rem; font-weight: 600; line-height: 1.1; }
.scheme3-available-ledger small { color: var(--available-muted); font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: .52rem; font-weight: 700; letter-spacing: .04em; }
.scheme3-available-channels :deep(.table-page-layout) { height: calc(100vh - 12rem); min-height: 24rem; gap: .78rem; }
.scheme3-available-filters :deep(.input) { min-height: 2.45rem; border-radius: 7px; }
.scheme3-available-filters :deep(.btn) { min-height: 2.35rem; border-radius: 7px; }
.scheme3-available-table :deep(.table-wrapper) { overflow: auto; }
.scheme3-available-channels :deep(.table-scroll-container) { border: 1px solid var(--available-line); border-radius: 8px; background: var(--available-card); box-shadow: 0 11px 24px rgba(54,48,34,.06); }
.scheme3-available-table :deep(table thead),.scheme3-available-table :deep(table thead tr),.scheme3-available-table :deep(table thead th) { background: #f1eee6 !important; }
.scheme3-available-table :deep(table th) { padding-top: .72rem; padding-bottom: .72rem; border-color: var(--available-line); color: var(--available-muted); font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: .59rem; font-weight: 800; letter-spacing: .055em; }
.scheme3-available-table :deep(table td) { border-color: rgba(216,210,195,.78); color: var(--available-ink); }
.scheme3-available-table :deep(table tbody tr) { transition: background-color 150ms ease; }
.scheme3-available-table :deep(table tbody tr:hover) { background: rgba(30,92,66,.045); }
.scheme3-available-table :deep(table > tbody) { border-bottom-color: var(--available-line); }
.scheme3-available-table :deep([data-testid='mobile-channels']) { background: var(--available-card); }
.scheme3-available-table :deep([data-testid='mobile-channels'] > section) { border-color: var(--available-line); }
.scheme3-available-table :deep([data-testid='mobile-channels'] dl > div) { border-color: rgba(216,210,195,.78); }

:global(.dark .scheme3-available-channels) { --available-card: #24231f; --available-ink: #f4f2ec; --available-muted: #aaa69a; --available-line: #47443a; }
:global(.dark .scheme3-available-ledger strong) { color: #8fc2a5; }
:global(.dark .scheme3-available-channels .table-scroll-container) { background: #24231f; }
:global(.dark .scheme3-available-table table thead),:global(.dark .scheme3-available-table table thead tr),:global(.dark .scheme3-available-table table thead th) { background: #2b2924 !important; }
:global(.dark .scheme3-available-table table th),:global(.dark .scheme3-available-table table td) { border-color: rgba(71,68,58,.86); }
:global(.dark .scheme3-available-table table tbody tr:hover) { background: rgba(143,194,165,.07); }
:global(.dark .scheme3-available-table [data-testid='mobile-channels']) { background: #24231f; }
:global(.dark .scheme3-available-table [data-testid='mobile-channels'] > section) { border-color: #47443a; }
:global(.dark .scheme3-available-table [data-testid='mobile-channels'] dl > div) { border-color: rgba(71,68,58,.86); }

@media (max-width: 767px) {
  .scheme3-available-header { align-items: stretch; flex-direction: column; gap: .8rem; margin-bottom: .85rem; }
  .scheme3-available-ledger { width: 100%; justify-content: stretch; }
  .scheme3-available-ledger span { flex: 1 1 45%; min-width: 0; padding: .48rem .42rem; }
  .scheme3-available-channels :deep(.table-page-layout) { height: auto; min-height: 0; gap: .7rem; }
  .scheme3-available-channels :deep(.table-scroll-container) { border: 0; border-radius: 0; background: transparent; box-shadow: none; }
  .scheme3-available-table :deep([data-testid='mobile-channels']) { border: 1px solid var(--available-line); border-radius: 8px; box-shadow: 0 8px 18px rgba(54,48,34,.05); }
}
</style>

<template>
  <AppLayout>
    <section class="scheme3-model-square">
      <header class="scheme3-model-square-header">
        <div>
          <p class="scheme3-model-square-kicker">模型目录 / 定价账本</p>
          <h1>模型广场</h1>
          <p class="scheme3-model-square-subtitle">按模型核对可用渠道、分组与基础定价。</p>
        </div>
        <div class="scheme3-model-square-tools">
          <label class="scheme3-model-square-search">
            <Icon name="search" size="sm" />
            <input v-model="search" aria-label="搜索模型、渠道、平台或分组" placeholder="搜索模型、渠道、平台或分组..." />
          </label>
          <button type="button" class="scheme3-model-square-refresh" :disabled="loading" title="刷新模型目录" @click="loadModels">
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
          </button>
        </div>
      </header>

      <div class="scheme3-model-square-filter" aria-label="平台筛选">
        <button
          v-for="item in platforms"
          :key="item"
          type="button"
          class="scheme3-model-square-filter-button"
          :class="{ 'scheme3-model-square-filter-button-active': platform === item }"
          @click="platform = item"
        >
          {{ item === 'all' ? '全部平台' : item.toUpperCase() }}
        </button>
      </div>

      <div class="scheme3-model-square-note">
        <Icon name="infoCircle" size="xs" />
        提示：下方价格为渠道基础价，实际扣费按所选分组倍率计算；专属倍率和高峰倍率会直接显示在分组标签中。
      </div>

      <div v-if="loading" class="scheme3-model-square-state">
        <div class="scheme3-model-square-spinner"></div>
      </div>
      <div v-else-if="filteredModels.length === 0" class="scheme3-model-square-state scheme3-model-square-empty">
        <Icon name="inbox" size="lg" />
        <span>没有可展示的模型</span>
      </div>

      <div v-else class="scheme3-model-square-scroll-region" tabindex="0">
        <div class="scheme3-model-square-grid scheme3-model-square-scroll-content">
          <article v-for="model in filteredModels" :key="model.key" class="scheme3-model-square-card">
            <header class="scheme3-model-square-card-header">
              <div>
                <h2>{{ model.name }}</h2>
                <span>{{ model.platform }}</span>
              </div>
              <div class="scheme3-model-square-card-count">
                {{ channelCount(model) }} {{ t('modelPlaza.detail.channelCount', channelCount(model)) }}
              </div>
            </header>

            <div class="scheme3-model-square-card-body">
              <section v-for="channel in model.channels" :key="channel.key" class="scheme3-model-square-channel">
                <div class="scheme3-model-square-channel-meta">
                  <span>渠道</span>
                  <strong>{{ channel.name }}</strong>
                  <span>分组</span>
                  <div class="scheme3-model-square-groups">
                    <GroupBadge
                      v-for="entry in channel.entries"
                      :key="entryKey(entry)"
                      :name="entry.group.name"
                      :platform="entry.group.platform as GroupPlatform"
                      :subscription-type="entry.group.subscription_type as SubscriptionType"
                      :rate-multiplier="entry.group.rate_multiplier"
                      :user-rate-multiplier="userGroupRates[entry.group.id] ?? null"
                      :peak-rate-enabled="entry.group.peak_rate_enabled"
                      :peak-start="entry.group.peak_start"
                      :peak-end="entry.group.peak_end"
                      :peak-rate-multiplier="entry.group.peak_rate_multiplier"
                      always-show-rate
                      class="model-square-group-badge"
                    />
                  </div>
                </div>

                <div class="scheme3-model-square-pricing">
                  <div class="scheme3-model-square-pricing-header">
                    <h3>渠道基础定价</h3>
                    <span>
                      {{ billingModeLabel(channel.pricing) }}
                    </span>
                  </div>

                  <div class="scheme3-model-square-pricing-grid">
                    <div v-for="item in fullPriceItems(channel.pricing)" :key="item.label">
                      <p>{{ item.label }}</p>
                      <strong>
                        {{ formatTokenPrice(item.value) }}
                      </strong>
                    </div>
                  </div>

                  <div v-if="isRequestBilling(channel.pricing) || (channel.pricing?.intervals?.length)" class="scheme3-model-square-pricing-extra">
                    <div v-if="isRequestBilling(channel.pricing)">
                      <p>{{ channel.pricing?.billing_mode === 'image' ? '每张价格' : '每次价格' }}</p>
                      <strong>
                        {{ formatRequestPrice(channel.pricing?.per_request_price, channel.pricing?.billing_mode) }}
                      </strong>
                    </div>
                    <div v-if="channel.pricing?.intervals?.length" class="scheme3-model-square-tier-note">
                      <Icon name="shield" size="xs" />
                      <span>已配置 {{ channel.pricing.intervals.length }} 档阶梯价格</span>
                    </div>
                  </div>
                </div>
              </section>
            </div>
          </article>
        </div>
      </div>

      <button
        v-show="showBackToTop"
        type="button"
        class="scheme3-model-square-back-to-top"
        aria-label="返回顶部"
        title="返回顶部"
        @click="scrollToTop"
      >
        <Icon name="arrowUp" size="sm" />
      </button>
    </section>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import modelSquareAPI, { type ModelSquareEntry } from '@/api/modelSquare'
import type { UserSupportedModelPricing } from '@/api/channels'
import userGroupsAPI from '@/api/groups'
import type { GroupPlatform, SubscriptionType } from '@/types'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { useI18n } from 'vue-i18n'

interface ModelSquareModel {
  key: string
  name: string
  platform: string
  entries: ModelSquareEntry[]
  channels: ModelSquareChannel[]
}

interface ModelSquareChannel {
  key: string
  name: string
  entries: ModelSquareEntry[]
  pricing: UserSupportedModelPricing | null
}

const { t } = useI18n()
const appStore = useAppStore()
const loading = ref(false)
const search = ref('')
const platform = ref('all')
const models = ref<ModelSquareEntry[]>([])
const userGroupRates = ref<Record<number, number>>({})
const showBackToTop = ref(false)

const platforms = computed(() => ['all', ...Array.from(new Set(models.value.map((item) => item.platform))).sort()])
const modelGroups = computed<ModelSquareModel[]>(() => {
  const modelsByKey = new Map<string, ModelSquareModel>()
  for (const entry of models.value) {
    const key = `${entry.platform}:${entry.name.toLowerCase()}`
    const existing = modelsByKey.get(key)
    if (existing) existing.entries.push(entry)
    else modelsByKey.set(key, { key, name: entry.name, platform: entry.platform, entries: [entry], channels: [] })
  }
  return Array.from(modelsByKey.values())
    .map((model) => ({ ...model, channels: groupChannels(model.entries) }))
    .sort((a, b) => a.platform.localeCompare(b.platform) || a.name.localeCompare(b.name))
})
const filteredModels = computed(() => {
  const query = search.value.trim().toLowerCase()
  return modelGroups.value.filter((model) => {
    if (platform.value !== 'all' && model.platform !== platform.value) return false
    if (!query) return true
    return [model.name, model.platform, ...model.entries.flatMap((entry) => [entry.channel_name, entry.group.name])]
      .join(' ')
      .toLowerCase()
      .includes(query)
  })
})

const PER_MILLION_TOKENS = 1_000_000

function entryKey(entry: ModelSquareEntry) {
  return `${entry.channel_id}:${entry.group.id}:${entry.name}`
}

function channelCount(model: ModelSquareModel) {
  return model.channels.length
}

function groupChannels(entries: ModelSquareEntry[]): ModelSquareChannel[] {
  const channelsByKey = new Map<string, ModelSquareChannel>()
  for (const entry of entries) {
    const key = entry.channel_id > 0 ? `channel:${entry.channel_id}` : 'account-only'
    const existing = channelsByKey.get(key)
    if (existing) existing.entries.push(entry)
    else channelsByKey.set(key, {
      key,
      name: entry.channel_name || '未关联渠道',
      entries: [entry],
      pricing: entry.pricing,
    })
  }
  return Array.from(channelsByKey.values())
}

function formatTokenPrice(value: number | null | undefined) {
  if (value == null) return '未配置'
  const perMillion = value * PER_MILLION_TOKENS
  return `$${perMillion.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: perMillion < 1 ? 6 : 2 })}/M`
}

function formatRequestPrice(value: number | null | undefined, billingMode?: string) {
  if (value == null) return '未配置'
  return `$${value.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 6 })}/${billingMode === 'image' ? '张' : '次'}`
}

function billingModeLabel(pricing: UserSupportedModelPricing | null) {
  switch (pricing?.billing_mode) {
    case 'image': return '图片计费'
    case 'per_request': return '按次计费'
    default: return 'Token 计费'
  }
}

function isRequestBilling(pricing: UserSupportedModelPricing | null) {
  return pricing?.billing_mode === 'image' || pricing?.billing_mode === 'per_request'
}

function fullPriceItems(pricing: UserSupportedModelPricing | null) {
  return [
    { label: '输入', value: pricing?.input_price },
    { label: '输出', value: pricing?.output_price },
    { label: '缓存写入', value: pricing?.cache_write_price },
    { label: '缓存读取', value: pricing?.cache_read_price },
    { label: '图片输入', value: pricing?.image_input_price },
    { label: '图片输出', value: pricing?.image_output_price },
  ]
}

async function loadModels() {
  loading.value = true
  try {
    models.value = await modelSquareAPI.list()
    userGroupRates.value = await userGroupsAPI.getUserGroupRates().catch(() => ({}))
  } catch (error: unknown) {
    appStore.showError(extractApiErrorMessage(error, '加载模型广场失败'))
  } finally {
    loading.value = false
  }
}

function onScroll() {
  showBackToTop.value = (window.scrollY || document.documentElement.scrollTop) > 300
}

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

onMounted(() => {
  void loadModels()
  onScroll()
  window.addEventListener('scroll', onScroll, { passive: true })
})

onBeforeUnmount(() => {
  window.removeEventListener('scroll', onScroll)
})
</script>

<style scoped>
.scheme3-model-square { --square-ink: var(--scheme3-ink,#16150f); --square-muted: var(--scheme3-muted,#6b695f); --square-line: var(--scheme3-line,#dad5c8); --square-paper: var(--scheme3-paper,#f4f2ec); --square-card: var(--scheme3-card,#fbfaf6); color: var(--square-ink); }
.scheme3-model-square-header { display: flex; align-items: end; justify-content: space-between; gap: 1.25rem; margin-bottom: 1.2rem; border-bottom: 1px solid var(--square-line); padding: .15rem 0 1.1rem; }
.scheme3-model-square-kicker { margin: 0; color: var(--square-muted); font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; font-size: .61rem; font-weight: 800; letter-spacing: .11em; }
.scheme3-model-square h1 { margin: .38rem 0 0; font-family: Georgia,'Times New Roman',serif; font-size: clamp(1.55rem,2.6vw,2.1rem); font-weight: 500; letter-spacing: 0; }
.scheme3-model-square-subtitle { margin: .36rem 0 0; color: var(--square-muted); font-size: .78rem; }
.scheme3-model-square-tools { display: flex; align-items: center; gap: .55rem; }
.scheme3-model-square-search { display: flex; width: min(23rem,42vw); min-height: 2.55rem; align-items: center; gap: .52rem; border: 1px solid var(--square-line); border-radius: 7px; padding: 0 .75rem; background: var(--square-card); color: var(--square-muted); transition: border-color 150ms ease,box-shadow 150ms ease; }
.scheme3-model-square-search:focus-within { border-color: #1e5c42; box-shadow: 0 0 0 3px rgba(30,92,66,.12); color: #1e5c42; }
.scheme3-model-square-search input { width: 100%; border: 0; outline: 0; background: transparent; color: var(--square-ink); font-size: .75rem; }
.scheme3-model-square-search input::placeholder { color: #979286; }
.scheme3-model-square-refresh { display: inline-flex; width: 2.55rem; height: 2.55rem; align-items: center; justify-content: center; border: 1px solid var(--square-line); border-radius: 7px; background: var(--square-card); color: var(--square-ink); transition: background-color 150ms ease,transform 150ms ease; }
.scheme3-model-square-refresh:hover { background: #ebe8de; }.scheme3-model-square-refresh:active { transform: scale(.96); }
.scheme3-model-square-filter { display: flex; flex-wrap: wrap; gap: .45rem; margin-bottom: .95rem; }
.scheme3-model-square-filter-button { min-height: 2rem; border: 1px solid var(--square-line); border-radius: 5px; padding: .35rem .7rem; background: var(--square-card); color: var(--square-muted); font-size: .65rem; font-weight: 800; transition: background-color 150ms ease,border-color 150ms ease,color 150ms ease,transform 150ms ease; }
.scheme3-model-square-filter-button:hover { border-color: rgba(30,92,66,.42); color: #1e5c42; }.scheme3-model-square-filter-button:active { transform: scale(.97); }.scheme3-model-square-filter-button-active { border-color: #1e5c42; background: #1e5c42; color: #f4f2ec; }
.scheme3-model-square-note { display: flex; align-items: flex-start; gap: .45rem; margin-bottom: 1.15rem; border-left: 3px solid #b7791f; padding: .52rem .7rem; background: rgba(183,121,31,.08); color: #765213; font-size: .7rem; line-height: 1.55; }
.scheme3-model-square-note svg { flex: 0 0 auto; margin-top: .15rem; }
.scheme3-model-square-state { display: flex; min-height: 19rem; align-items: center; justify-content: center; }.scheme3-model-square-spinner { width: 1.85rem; height: 1.85rem; border: 2px solid rgba(30,92,66,.18); border-top-color: #1e5c42; border-radius: 50%; animation: square-spin .7s linear infinite; }.scheme3-model-square-empty { flex-direction: column; gap: .6rem; border: 1px dashed var(--square-line); color: var(--square-muted); font-size: .75rem; }.scheme3-model-square-empty svg { color: #b7791f; }
.scheme3-model-square-scroll-region { min-width: 0; overflow-x: auto; overscroll-behavior-x: contain; }
.scheme3-model-square-scroll-content { min-width: 75rem; padding: .125rem .25rem 1rem; }
.scheme3-model-square-grid { display: grid; grid-template-columns: repeat(2,minmax(0,1fr)); gap: 1rem; }
.scheme3-model-square-card { overflow: hidden; border: 1px solid var(--square-line); border-radius: 8px; background: var(--square-card); box-shadow: 0 9px 20px rgba(54,48,34,.05); }
.scheme3-model-square-card-header { display: flex; align-items: start; justify-content: space-between; gap: 1rem; border-bottom: 1px solid var(--square-line); padding: 1rem 1.05rem .85rem; background: #f8f6ef; }.scheme3-model-square-card-header h2 { overflow: hidden; margin: 0; color: var(--square-ink); font-family: Georgia,'Times New Roman',serif; font-size: 1.04rem; font-weight: 600; text-overflow: ellipsis; white-space: nowrap; }.scheme3-model-square-card-header span,.scheme3-model-square-card-count { display: block; margin-top: .35rem; color: var(--square-muted); font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; font-size: .56rem; font-weight: 700; letter-spacing: .05em; }.scheme3-model-square-card-count { flex: 0 0 auto; text-align: right; }
.scheme3-model-square-channel { padding: 1rem 1.05rem; border-bottom: 1px solid var(--square-line); }.scheme3-model-square-channel:last-child { border-bottom: 0; }.scheme3-model-square-channel-meta { display: grid; grid-template-columns: 2.65rem minmax(0,1fr); gap: .46rem .68rem; align-items: start; }.scheme3-model-square-channel-meta > span { padding-top: .13rem; color: var(--square-muted); font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; font-size: .56rem; font-weight: 800; letter-spacing: .05em; }.scheme3-model-square-channel-meta > strong { overflow-wrap: anywhere; font-size: .74rem; }
.scheme3-model-square-groups { display: flex; flex-wrap: wrap; gap: .35rem; min-width: 0; }.model-square-group-badge { max-width: 100%; transition: transform 150ms ease; }.model-square-group-badge:hover { transform: translateY(-1px); }
.scheme3-model-square-pricing { margin-top: .85rem; border: 1px solid var(--square-line); border-radius: 6px; padding: .75rem; background: rgba(244,242,236,.62); }.scheme3-model-square-pricing-header { display: flex; justify-content: space-between; gap: .8rem; margin-bottom: .7rem; }.scheme3-model-square-pricing h3 { margin: 0; font-size: .64rem; font-weight: 800; }.scheme3-model-square-pricing-header span { color: var(--square-muted); font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; font-size: .54rem; }.scheme3-model-square-pricing-grid { display: grid; grid-template-columns: repeat(3,minmax(0,1fr)); gap: .7rem .5rem; }.scheme3-model-square-pricing-grid p,.scheme3-model-square-pricing-extra p { margin: 0 0 .18rem; color: var(--square-muted); font-size: .55rem; }.scheme3-model-square-pricing-grid strong,.scheme3-model-square-pricing-extra strong { overflow-wrap: anywhere; font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; font-size: .65rem; }.scheme3-model-square-pricing-extra { display: flex; flex-wrap: wrap; align-items: end; gap: .8rem; margin-top: .8rem; border-top: 1px solid var(--square-line); padding-top: .7rem; }.scheme3-model-square-pricing-extra > div:first-child strong { color: #1e5c42; }.scheme3-model-square-tier-note { display: inline-flex; align-items: center; gap: .28rem; border: 1px solid rgba(183,121,31,.28); border-radius: 4px; padding: .25rem .4rem; background: rgba(183,121,31,.08); color: #765213; font-size: .56rem; font-weight: 800; }
.scheme3-model-square-back-to-top { position: fixed; right: 1.25rem; bottom: calc(1.25rem + env(safe-area-inset-bottom)); z-index: 40; display: inline-flex; width: 2.6rem; height: 2.6rem; align-items: center; justify-content: center; border: 1px solid rgba(30,92,66,.45); border-radius: 6px; background: #1e5c42; color: #f4f2ec; box-shadow: 0 8px 18px rgba(22,21,15,.16); transition: background-color 150ms ease,transform 150ms ease,opacity 150ms ease; }.scheme3-model-square-back-to-top:hover { background: #174a35; }.scheme3-model-square-back-to-top:active { transform: scale(.96); }.scheme3-model-square-back-to-top:focus-visible { outline: 3px solid rgba(30,92,66,.22); outline-offset: 2px; }
@keyframes square-spin { to { transform: rotate(360deg); } }
:global(html.dark) .scheme3-model-square-card,
:global(html.dark) .scheme3-model-square-search,
:global(html.dark) .scheme3-model-square-refresh {
  border-color: #47443a;
  background: #24231f;
}
:global(html.dark) .scheme3-model-square-refresh:hover { background: #2b2924; }
:global(html.dark .scheme3-model-square-card-header) {
  border-color: #47443a;
  background: #1b1b18;
}
:global(html.dark .scheme3-model-square-card-header h2) { color: #f4f2ec; }
:global(html.dark .scheme3-model-square-card-header span),
:global(html.dark .scheme3-model-square-card-count) { color: #aaa69a; }
:global(html.dark .scheme3-model-square-search input) { color: #f4f2ec; }
:global(html.dark .scheme3-model-square-search input::placeholder) { color: #aaa69a; }
:global(html.dark .scheme3-model-square-pricing) {
  border-color: #47443a;
  background: #2b2924;
  color: #f4f2ec;
}
:global(html.dark .scheme3-model-square-pricing-header span),
:global(html.dark .scheme3-model-square-pricing-grid p),
:global(html.dark .scheme3-model-square-pricing-extra p) { color: #aaa69a; }
:global(html.dark .scheme3-model-square-pricing-grid strong),
:global(html.dark .scheme3-model-square-pricing-extra strong),
:global(html.dark .scheme3-model-square-pricing h3) { color: #f4f2ec; }
:global(html.dark .scheme3-model-square-pricing-extra) { border-color: #47443a; }
:global(html.dark .scheme3-model-square-pricing-extra > div:first-child strong) { color: #8fc2a5; }
:global(html.dark .scheme3-model-square-tier-note) {
  border-color: rgba(214,166,93,.35);
  background: rgba(183,121,31,.12);
  color: #e8c878;
}
:global(html.dark .scheme3-model-square-note) { color: #e8c878; background: rgba(183,121,31,.11); }
:global(html.dark .scheme3-model-square-back-to-top) { border-color: #8fc2a5; background: #8fc2a5; color: #16150f; }
:global(html.dark .scheme3-model-square-back-to-top:hover) { background: #a7d0b8; }
@media (max-width: 900px) { .scheme3-model-square-grid { grid-template-columns: 1fr; } }
@media (max-width: 640px) { .scheme3-model-square-header { align-items: stretch; flex-direction: column; }.scheme3-model-square-tools,.scheme3-model-square-search { width: 100%; }.scheme3-model-square-refresh { flex: 0 0 auto; }.scheme3-model-square-scroll-content { min-width: 37.5rem; }.scheme3-model-square-pricing-grid { grid-template-columns: repeat(2,minmax(0,1fr)); }.scheme3-model-square-card-header,.scheme3-model-square-channel { padding-right: .8rem; padding-left: .8rem; }.scheme3-model-square-card-count { max-width: 5.5rem; }.scheme3-model-square-note { font-size: .65rem; }.scheme3-model-square-back-to-top { right: .8rem; bottom: calc(.8rem + env(safe-area-inset-bottom)); } }
</style>

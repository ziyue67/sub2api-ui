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
            <input v-model="searchInput" aria-label="搜索模型、渠道、平台或分组" placeholder="搜索模型、渠道、平台或分组..." />
          </label>
          <button
            type="button"
            class="scheme3-model-square-refresh"
            :disabled="loading"
            :aria-busy="loading"
            aria-label="刷新模型目录"
            title="刷新模型目录"
            @click="loadModels"
          >
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
          </button>
          <button
            type="button"
            class="scheme3-model-square-refresh"
            :class="{ 'scheme3-model-square-tool-active': showAdvancedFilters }"
            :aria-expanded="showAdvancedFilters"
            aria-controls="scheme3-model-square-advanced-filters"
            aria-label="筛选模型"
            title="筛选模型"
            @click="showAdvancedFilters = !showAdvancedFilters"
          >
            <Icon name="filter" size="md" />
            <span v-if="activeAdvancedFilterCount" class="scheme3-model-square-filter-count">
              {{ activeAdvancedFilterCount }}
            </span>
          </button>
        </div>
      </header>

      <div class="scheme3-model-square-filter" role="group" aria-label="平台筛选">
        <button
          v-for="item in platforms"
          :key="item"
          type="button"
          class="scheme3-model-square-filter-button"
          :class="{ 'scheme3-model-square-filter-button-active': platform === item }"
          @click="setPlatform(item)"
        >
          {{ item === 'all' ? '全部平台' : item.toUpperCase() }}
        </button>
      </div>

      <div class="scheme3-model-square-commandbar">
        <label>
          <span>排序</span>
          <select v-model="sortOption" aria-label="模型排序">
            <option value="default">默认排序</option>
            <option value="price_asc">输入价从低到高</option>
            <option value="price_desc">输入价从高到低</option>
            <option value="output_price_asc">输出价从低到高</option>
            <option value="output_price_desc">输出价从高到低</option>
            <option value="channels_desc">渠道数量最多</option>
            <option value="name_asc">模型名称 A-Z</option>
          </select>
        </label>
        <div class="scheme3-model-square-view-switch" role="group" aria-label="视图模式">
          <button
            type="button"
            :class="{ active: viewMode === 'grid' }"
            aria-label="卡片视图"
            title="卡片视图"
            @click="viewMode = 'grid'"
          >
            <Icon name="grid" size="sm" />
          </button>
          <button
            type="button"
            :class="{ active: viewMode === 'table' }"
            aria-label="表格视图"
            title="表格视图"
            @click="viewMode = 'table'"
          >
            <Icon name="table" size="sm" />
          </button>
        </div>
        <span class="scheme3-model-square-result-count">当前 {{ filteredModels.length }} / {{ modelGroups.length }} 个模型</span>
      </div>

      <section
        v-show="showAdvancedFilters"
        id="scheme3-model-square-advanced-filters"
        class="scheme3-model-square-advanced-filters"
        aria-label="高级筛选"
      >
        <div class="scheme3-model-square-filter-section">
          <div class="scheme3-model-square-filter-heading">
            <strong>能力特性</strong>
            <button v-if="selectedCapabilities.length" type="button" @click="selectedCapabilities = []">清空</button>
          </div>
          <div class="scheme3-model-square-capabilities">
            <button
              v-for="capability in capabilityOptions"
              :key="capability"
              type="button"
              :class="{ active: selectedCapabilities.includes(capability) }"
              @click="toggleCapability(capability)"
            >
              <Icon :name="capabilityMeta(capability).icon as any" size="xs" />
              <span>{{ capabilityMeta(capability).label }}</span>
            </button>
          </div>
        </div>
        <label class="scheme3-model-square-filter-select">
          <span>输入价 / 1M tokens</span>
          <select v-model="priceRange">
            <option value="all">全部价格</option>
            <option value="free">免费</option>
            <option value="lt1">&lt; $1</option>
            <option value="1to5">$1 - $5</option>
            <option value="5to15">$5 - $15</option>
            <option value="gt15">&gt; $15</option>
          </select>
        </label>
        <label class="scheme3-model-square-filter-select">
          <span>上下文</span>
          <select v-model="contextRange">
            <option value="all">全部长度</option>
            <option value="lt32k">&lt; 32K</option>
            <option value="32kTo128k">32K - 128K</option>
            <option value="128kTo256k">128K - 256K</option>
            <option value="gt256k">&gt; 256K</option>
          </select>
        </label>
        <label class="scheme3-model-square-filter-select">
          <span>计费类型</span>
          <select v-model="billingType">
            <option value="all">全部类型</option>
            <option value="tokens">Token 计费</option>
            <option value="request">按次 / 张计费</option>
            <option value="intervals">阶梯价格</option>
          </select>
        </label>
        <button
          v-if="activeAdvancedFilterCount"
          type="button"
          class="scheme3-model-square-reset"
          @click="resetFilters"
        >
          <Icon name="refresh" size="xs" />
          重置筛选
        </button>
      </section>

      <div class="scheme3-model-square-note">
        <Icon name="infoCircle" size="xs" />
        提示：下方价格为渠道基础价，实际扣费按所选分组倍率计算；专属倍率和高峰倍率会直接显示在分组标签中。
      </div>

      <div v-if="loading" class="scheme3-model-square-state" role="status" aria-live="polite">
        <div class="scheme3-model-square-spinner"></div>
      </div>
      <div v-else-if="filteredModels.length === 0" class="scheme3-model-square-state scheme3-model-square-empty">
        <Icon name="inbox" size="lg" />
        <span>没有可展示的模型</span>
      </div>

      <ModelSquareTable
        v-else-if="viewMode === 'table'"
        :models="filteredModels"
        :user-group-rates="userGroupRates"
        @view-details="selectedModel = $event"
      />

      <div v-else class="scheme3-model-square-layout">
        <Scheme3ModelSquareIndex
          v-model="activeModelKey"
          :models="filteredModels"
          @select="handleIndexModelSelect"
        />

        <div class="scheme3-model-square-results">
          <div class="scheme3-model-square-index-bar">
            <label for="scheme3-model-square-index">模型索引</label>
            <select
              id="scheme3-model-square-index"
              v-model="activeModelKey"
              @change="handleMobileIndexChange"
            >
              <option value="">选择模型...</option>
              <option v-for="model in filteredModels" :key="model.key" :value="model.key">
                {{ platformLabel(model.platform) }} / {{ model.name }}
              </option>
            </select>
            <span>{{ filteredModels.length }} 个模型</span>
          </div>

          <div
            ref="scrollRegionRef"
            class="scheme3-model-square-scroll-region"
            role="region"
            aria-label="模型目录列表"
            tabindex="0"
          >
            <div
              class="scheme3-model-square-scroll-content scheme3-model-square-virtual-content"
              role="list"
              :style="{ height: `${virtualizerTotalSize}px` }"
            >
              <div
                v-for="virtualRow in virtualRows"
                :key="String(virtualRow.key)"
                :ref="measureVirtualRow"
                class="scheme3-model-square-virtual-row"
                :data-index="virtualRow.index"
                :style="{ transform: `translateY(${virtualRow.start - virtualScrollMargin}px)` }"
              >
                <template v-for="model in [modelRows[virtualRow.index]]" :key="model?.key">
                  <article
                    v-if="model"
                    :id="modelCardId(model.key)"
                    data-model-card="true"
                    :data-model-index="virtualRow.index"
                    :data-model-key="model.key"
                    role="listitem"
                    :aria-posinset="virtualRow.index + 1"
                    :aria-setsize="filteredModels.length"
                    :class="{ 'scheme3-model-square-card-active': activeModelKey === model.key }"
                    :aria-current="activeModelKey === model.key ? 'true' : undefined"
                    class="scheme3-model-square-card"
                  >
              <header class="scheme3-model-square-card-header">
                <div class="scheme3-model-square-card-heading">
                  <span
                    class="scheme3-model-square-platform-icon"
                    :style="{ '--square-platform-color': platformAccentColor(model.platform) }"
                  >
                    <PlatformIcon :platform="model.platform" size="lg" />
                  </span>
                  <div>
                    <span class="scheme3-model-square-card-platform">{{ platformLabel(model.platform) }}</span>
                    <h2 :title="model.name">{{ model.name }}</h2>
                    <span class="scheme3-model-square-card-count">
                      {{ t('modelPlaza.detail.channelCount', { count: channelCount(model) }) }}
                    </span>
                    <span class="scheme3-model-square-card-facts">
                      <span>{{ model.contextWindow || '128K' }}</span>
                      <span v-for="capability in (model.capabilities || []).slice(0, 3)" :key="capability">
                        {{ capabilityMeta(capability).label }}
                      </span>
                    </span>
                  </div>
                </div>
                <div class="scheme3-model-square-card-actions">
                  <button
                    type="button"
                    class="scheme3-model-square-config-toggle"
                    title="查看定价与调用示例"
                    @click="selectedModel = model"
                  >
                    <Icon name="externalLink" size="xs" />
                    详情
                  </button>
                  <button
                    type="button"
                    class="scheme3-model-square-config-toggle"
                    :aria-expanded="isModelConfigExpanded(model.key)"
                    :aria-controls="modelConfigId(model.key)"
                    @click="toggleModelConfig(model.key)"
                  >
                    <Icon
                      name="chevronDown"
                      size="xs"
                      :class="{ 'scheme3-model-square-config-chevron-open': isModelConfigExpanded(model.key) }"
                    />
                    {{ isModelConfigExpanded(model.key) ? '收起配置' : '查看配置' }}
                  </button>
                </div>
              </header>

              <div class="scheme3-model-square-card-body">
                <section v-for="channel in model.channels" :key="channel.key" class="scheme3-model-square-channel">
                  <div class="scheme3-model-square-channel-meta">
                    <div class="scheme3-model-square-channel-heading">
                      <span class="scheme3-model-square-section-label">渠道</span>
                      <strong>{{ channel.name }}</strong>
                    </div>
                    <div class="scheme3-model-square-channel-groups">
                      <span class="scheme3-model-square-section-label">可用分组</span>
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
                        scheme3
                        class="model-square-group-badge"
                      />
                      </div>
                    </div>
                  </div>

                  <div class="scheme3-model-square-pricing">
                    <div class="scheme3-model-square-pricing-header">
                      <h3>基础定价</h3>
                      <span>
                        {{ billingModeLabel(channel.pricing) }}
                      </span>
                      <em v-if="channel.isOfficialFallback" title="当前渠道未配置独立价格，显示官方参考价">官方参考价</em>
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
                        <p>{{ requestPriceLabel(channel.pricing?.billing_mode) }}</p>
                        <strong>
                          {{ formatRequestPrice(channel.pricing?.per_request_price, channel.pricing?.billing_mode) }}
                        </strong>
                      </div>
                    </div>

                    <details
                      v-if="channel.pricing?.intervals?.length"
                      class="scheme3-model-square-tier-details"
                      @toggle="remeasureVirtualRows"
                    >
                      <summary>
                        <span>
                          <Icon name="shield" size="xs" />
                          已配置 {{ channel.pricing.intervals.length }} 档阶梯价格
                        </span>
                        <Icon name="chevronDown" size="xs" class="scheme3-model-square-tier-chevron" />
                      </summary>
                      <div class="scheme3-model-square-tier-list">
                        <div
                          v-for="(interval, intervalIndex) in channel.pricing.intervals"
                          :key="`${interval.min_tokens}:${interval.max_tokens ?? 'max'}:${interval.tier_label ?? intervalIndex}`"
                          class="scheme3-model-square-tier-row"
                        >
                          <strong>{{ intervalLabel(interval, intervalIndex, channel.pricing.billing_mode) }}</strong>
                          <div class="scheme3-model-square-tier-prices">
                            <span v-for="item in intervalPriceItems(interval, channel.pricing.billing_mode)" :key="item.label">
                              <small>{{ item.label }}</small>
                              {{ item.value }}
                            </span>
                          </div>
                        </div>
                      </div>
                    </details>
                  </div>
                </section>
              </div>

              <section
                v-if="isModelConfigExpanded(model.key)"
                :id="modelConfigId(model.key)"
                class="scheme3-model-square-config"
                role="region"
                :aria-label="`${model.name} 模型配置`"
              >
                <div class="scheme3-model-square-config-title">
                  <span>模型配置</span>
                  <code>{{ model.key }}</code>
                </div>
                <div v-for="channel in model.channels" :key="channel.key" class="scheme3-model-square-config-channel">
                  <div>
                    <strong>{{ channel.name }}</strong>
                    <code>{{ channel.key }}</code>
                  </div>
                  <div class="scheme3-model-square-config-entries">
                    <div v-for="entry in channel.entries" :key="entryKey(entry)">
                      <span>{{ entry.group.name }}</span>
                      <code>
                        id={{ entry.group.id }} · 倍率 {{ entry.group.rate_multiplier }}
                        <template v-if="userGroupRates[entry.group.id] != null">
                          · 专属 {{ userGroupRates[entry.group.id] }}
                        </template>
                      </code>
                    </div>
                  </div>
                </div>
              </section>
                  </article>
                </template>
              </div>
            </div>
          </div>
        </div>
      </div>

      <ModelSquareDetailModal
        :model="selectedModel"
        :user-group-rates="userGroupRates"
        @close="selectedModel = null"
      />

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
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { defaultRangeExtractor, useWindowVirtualizer, type Range } from '@tanstack/vue-virtual'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import type { ModelSquareEntry } from '@/api/modelSquare'
import type { UserPricingInterval } from '@/api/channels'
import type { GroupPlatform, SubscriptionType } from '@/types'
import Scheme3ModelSquareIndex from '@/features/model-square/components/Scheme3ModelSquareIndex.vue'
import ModelSquareDetailModal from '@/features/model-square/components/ModelSquareDetailModal.vue'
import ModelSquareTable from '@/features/model-square/components/ModelSquareTable.vue'
import { useModelSquare } from '@/features/model-square/composables/useModelSquare'
import { useModelSquareFilters } from '@/features/model-square/composables/useModelSquareFilters'
import { useModelSquareSearch } from '@/features/model-square/composables/useModelSquareSearch'
import type { ModelCapability, ModelSquareModel } from '@/features/model-square/types'
import { capabilityMeta } from '@/features/model-square/utils/capabilities'
import { platformAccentColor, platformLabel } from '@/utils/platformColors'
import {
  billingModeLabel,
  formatRequestPrice,
  formatTokenPrice,
  fullPriceItems,
  isRequestBilling,
  requestPriceLabel,
} from '@/features/model-square/utils/pricing'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const { loading, userGroupRates, platforms, modelGroups, loadModels } = useModelSquare()
const { search, debouncedSearch, setSearch } = useModelSquareSearch()
const {
  platform,
  selectedCapabilities,
  priceRange,
  contextRange,
  billingType,
  sortOption,
  viewMode,
  filteredModels,
  setPlatform,
  toggleCapability,
  resetFilters,
} = useModelSquareFilters({
  modelGroups,
  search: debouncedSearch,
})
const searchInput = computed({
  get: () => search.value,
  set: (value: string) => setSearch(value),
})
const showBackToTop = ref(false)
const showAdvancedFilters = ref(false)
const selectedModel = ref<ModelSquareModel | null>(null)
const activeModelKey = ref('')
const capabilityOptions: ModelCapability[] = ['vision', 'tool_call', 'reasoning', 'audio', 'image_gen', 'embedding']
const activeAdvancedFilterCount = computed(() =>
  selectedCapabilities.value.length
  + (priceRange.value === 'all' ? 0 : 1)
  + (contextRange.value === 'all' ? 0 : 1)
  + (billingType.value === 'all' ? 0 : 1),
)

const MODEL_SQUARE_TOPBAR_FALLBACK = 68
const MODEL_SQUARE_SCROLL_GAP = 12
const MODEL_JUMP_FALLBACK_ATTEMPT = 180
const MODEL_JUMP_MAX_ATTEMPTS = 196
const scrollRegionRef = ref<HTMLElement | null>(null)
const virtualScrollMargin = ref(0)
const modelSquareScrollPadding = ref(MODEL_SQUARE_TOPBAR_FALLBACK + MODEL_SQUARE_SCROLL_GAP)
const pinnedModelIndex = ref<number | null>(null)
type DomFrameHandle = { kind: 'raf' | 'timeout'; id: number }
let pendingJumpFrame: DomFrameHandle | null = null
let pendingJumpKey: string | null = null
let pendingJumpGeneration = 0
let ignoreCardIntersectionUntil = 0
let modelCardObserver: IntersectionObserver | null = null
let modelCardMutationObserver: MutationObserver | null = null
const observedModelCards = new Set<HTMLElement>()

const modelRows = computed<ModelSquareModel[]>(() => filteredModels.value)
const MODEL_SQUARE_FULL_RENDER_LIMIT = 32

const modelRowVirtualizer = useWindowVirtualizer(computed(() => {
  const pinned = pinnedModelIndex.value
  return {
    count: modelRows.value.length,
    estimateSize: () => 510,
    overscan: 3,
    getItemKey: (index: number) => modelRows.value[index]?.key ?? index,
    rangeExtractor: (range: Range) => {
      if (modelRows.value.length <= MODEL_SQUARE_FULL_RENDER_LIMIT) {
        return Array.from({ length: modelRows.value.length }, (_, index) => index)
      }
      const indexes = defaultRangeExtractor(range)
      if (pinned != null && pinned >= 0 && pinned < modelRows.value.length && !indexes.includes(pinned)) {
        indexes.push(pinned)
        indexes.sort((a, b) => a - b)
      }
      return indexes
    },
    scrollMargin: virtualScrollMargin.value,
    scrollPaddingStart: modelSquareScrollPadding.value,
    initialRect: {
      width: typeof window === 'undefined' ? 1024 : Math.max(window.innerWidth, 1),
      height: typeof window === 'undefined' ? 600 : Math.max(window.innerHeight - 220, 420),
    },
    useAnimationFrameWithResizeObserver: true,
  }
}))

const virtualRows = computed(() => modelRowVirtualizer.value.getVirtualItems())
const virtualizerTotalSize = computed(() => modelRowVirtualizer.value.getTotalSize())
const expandedModelKeys = ref<Set<string>>(new Set())

function getModelSquareScrollPadding() {
  const topbar = document.querySelector<HTMLElement>('.scheme3-console-topbar')
  const topbarHeight = topbar?.getBoundingClientRect().height || MODEL_SQUARE_TOPBAR_FALLBACK
  return topbarHeight + MODEL_SQUARE_SCROLL_GAP
}

function updateVirtualLayoutMetrics() {
  const element = scrollRegionRef.value
  if (!element) return
  virtualScrollMargin.value = element.getBoundingClientRect().top + window.scrollY
  modelSquareScrollPadding.value = getModelSquareScrollPadding()
}

function measureVirtualRow(element: unknown) {
  if (element instanceof Element) modelRowVirtualizer.value.measureElement(element)
}

function isModelConfigExpanded(key: string) {
  return expandedModelKeys.value.has(key)
}

function toggleModelConfig(key: string) {
  const next = new Set(expandedModelKeys.value)
  if (next.has(key)) next.delete(key)
  else next.add(key)
  expandedModelKeys.value = next
  void nextTick(() => modelRowVirtualizer.value.measure())
}

function remeasureVirtualRows() {
  void nextTick(() => modelRowVirtualizer.value.measure())
}

function modelCardId(key: string) {
  return `model-${encodeURIComponent(key)}`
}

function modelConfigId(key: string) {
  return `${modelCardId(key)}-config`
}

function scheduleDomFrame(callback: () => void): DomFrameHandle {
  if (typeof window.requestAnimationFrame === 'function') {
    return { kind: 'raf', id: window.requestAnimationFrame(callback) }
  }
  return { kind: 'timeout', id: window.setTimeout(callback, 0) }
}

function cancelDomFrame(handle: DomFrameHandle | null) {
  if (!handle) return
  if (handle.kind === 'raf') window.cancelAnimationFrame(handle.id)
  else window.clearTimeout(handle.id)
}

function cancelPendingModelJump() {
  pendingJumpGeneration += 1
  cancelDomFrame(pendingJumpFrame)
  pendingJumpFrame = null
  pendingJumpKey = null
  ignoreCardIntersectionUntil = 0
  pinnedModelIndex.value = null
  modelRowVirtualizer.value.shouldAdjustScrollPositionOnItemSizeChange = undefined
}

function finishModelJump(key: string, generation: number) {
  if (pendingJumpKey !== key || pendingJumpGeneration !== generation) return
  pendingJumpFrame = null
  pendingJumpKey = null
  pinnedModelIndex.value = null
  modelRowVirtualizer.value.shouldAdjustScrollPositionOnItemSizeChange = undefined
  activeModelKey.value = key
  ignoreCardIntersectionUntil = window.performance.now() + 800
}

function queueModelCardAlignment(
  key: string,
  modelIndex: number,
  generation: number,
  attempt = 0,
  lastDocumentHeight = -1,
  stableFrames = 0,
) {
  pendingJumpFrame = scheduleDomFrame(() => {
    pendingJumpFrame = null
    if (pendingJumpKey !== key || pendingJumpGeneration !== generation) return
    const target = document.getElementById(modelCardId(key))
    if (target) {
      const targetOffset = target.getBoundingClientRect().top + window.scrollY - getModelSquareScrollPadding()
      window.scrollTo({ top: Math.max(0, targetOffset), behavior: 'instant' })
      activeModelKey.value = key
    } else if (attempt === MODEL_JUMP_FALLBACK_ATTEMPT) {
      modelSquareScrollPadding.value = getModelSquareScrollPadding()
      modelRowVirtualizer.value.scrollToIndex(modelIndex, { align: 'start', behavior: 'instant' })
    }

    const documentHeight = document.documentElement.scrollHeight
    const targetRect = target?.getBoundingClientRect()
    const targetVisible = Boolean(targetRect && targetRect.bottom > 0 && targetRect.top < window.innerHeight)
    const nextStableFrames = targetVisible && Math.abs(documentHeight - lastDocumentHeight) <= 1
      ? stableFrames + 1
      : 0

    if (nextStableFrames >= 4) {
      finishModelJump(key, generation)
      return
    }

    if (attempt >= MODEL_JUMP_MAX_ATTEMPTS) {
      finishModelJump(key, generation)
      return
    }

    queueModelCardAlignment(key, modelIndex, generation, attempt + 1, documentHeight, nextStableFrames)
  })
}

function jumpToModel(key: string) {
  const modelIndex = filteredModels.value.findIndex((model) => model.key === key)
  if (modelIndex < 0) return
  cancelPendingModelJump()
  const generation = pendingJumpGeneration
  pendingJumpKey = key
  ignoreCardIntersectionUntil = Number.POSITIVE_INFINITY
  pinnedModelIndex.value = modelIndex
  modelRowVirtualizer.value.shouldAdjustScrollPositionOnItemSizeChange = () => false
  void nextTick(() => {
    if (pendingJumpKey === key && pendingJumpGeneration === generation) {
      queueModelCardAlignment(key, modelIndex, generation)
    }
  })
}

function handleIndexModelSelect(key: string) {
  if (!key) return
  activeModelKey.value = key
  jumpToModel(key)
}

function handleMobileIndexChange(event: Event) {
  const target = event.target
  if (target instanceof HTMLSelectElement) handleIndexModelSelect(target.value)
}

function observeRenderedModelCards() {
  const observer = modelCardObserver
  const region = scrollRegionRef.value
  if (!observer) return
  if (!region) {
    observedModelCards.forEach((element) => observer.unobserve(element))
    observedModelCards.clear()
    return
  }

  const renderedCards = new Set(
    region.querySelectorAll<HTMLElement>('[data-model-card="true"]'),
  )
  observedModelCards.forEach((element) => {
    if (renderedCards.has(element) && region.contains(element)) return
    observer.unobserve(element)
    observedModelCards.delete(element)
  })
  renderedCards.forEach((element) => {
    if (observedModelCards.has(element)) return
    observer.observe(element)
    observedModelCards.add(element)
  })
}

function reconnectModelCardMutationObserver() {
  modelCardMutationObserver?.disconnect()
  modelCardMutationObserver = null
  const region = scrollRegionRef.value
  if (typeof MutationObserver === 'undefined' || !region) return
  modelCardMutationObserver = new MutationObserver(observeRenderedModelCards)
  modelCardMutationObserver.observe(region, { childList: true, subtree: true })
}

function setupModelCardObserver() {
  if (typeof IntersectionObserver === 'undefined') return
  modelCardObserver = new IntersectionObserver((entries) => {
    if (pendingJumpKey || window.performance.now() < ignoreCardIntersectionUntil) return
    const region = scrollRegionRef.value
    const visible = entries
      .filter((entry) => entry.isIntersecting && Boolean(region?.contains(entry.target)))
      .sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top)
    const key = visible[0]?.target.getAttribute('data-model-key')
    if (key && filteredModels.value.some((model) => model.key === key)) activeModelKey.value = key
  }, {
    rootMargin: '-12% 0px -62% 0px',
    threshold: 0,
  })
  observeRenderedModelCards()
  reconnectModelCardMutationObserver()
}

function entryKey(entry: ModelSquareEntry) {
  return `${entry.channel_id}:${entry.group.id}:${entry.name}`
}

function channelCount(model: ModelSquareModel) {
  return model.channels.length
}

function intervalLabel(interval: UserPricingInterval, index: number, billingMode?: string) {
  if (interval.tier_label?.trim()) return interval.tier_label
  if (billingMode === 'image' || billingMode === 'video') return `档位 ${index + 1}`
  const min = interval.min_tokens.toLocaleString()
  const max = interval.max_tokens == null ? '以上' : interval.max_tokens.toLocaleString()
  return `${min}–${max} tokens`
}

function intervalPriceItems(interval: UserPricingInterval, billingMode?: string) {
  if (billingMode === 'image' || billingMode === 'video' || billingMode === 'per_request') {
    return [{ label: requestPriceLabel(billingMode), value: formatRequestPrice(interval.per_request_price, billingMode) }]
  }
  return [
    { label: '输入', value: formatTokenPrice(interval.input_price) },
    { label: '输出', value: formatTokenPrice(interval.output_price) },
    { label: '缓存写入 (5m)', value: formatTokenPrice(interval.cache_write_price) },
    { label: '缓存写入 (1h)', value: formatTokenPrice(interval.cache_write_1h_price) },
    { label: '缓存读取', value: formatTokenPrice(interval.cache_read_price) },
  ]
}

function onScroll() {
  showBackToTop.value = (window.scrollY || document.documentElement.scrollTop) > 300
}

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

onMounted(() => {
  onScroll()
  void nextTick(() => {
    updateVirtualLayoutMetrics()
    setupModelCardObserver()
  })
  window.addEventListener('scroll', onScroll, { passive: true })
  window.addEventListener('resize', updateVirtualLayoutMetrics, { passive: true })
})

onBeforeUnmount(() => {
  cancelPendingModelJump()
  modelCardObserver?.disconnect()
  modelCardMutationObserver?.disconnect()
  observedModelCards.clear()
  modelCardObserver = null
  modelCardMutationObserver = null
  window.removeEventListener('scroll', onScroll)
  window.removeEventListener('resize', updateVirtualLayoutMetrics)
})

watch(scrollRegionRef, async () => {
  reconnectModelCardMutationObserver()
  await nextTick()
  updateVirtualLayoutMetrics()
  observeRenderedModelCards()
}, { flush: 'post' })

watch(
  filteredModels,
  async () => {
    if (pendingJumpKey) cancelPendingModelJump()
    if (!filteredModels.value.some((model) => model.key === activeModelKey.value)) activeModelKey.value = ''
    await nextTick()
    updateVirtualLayoutMetrics()
    modelRowVirtualizer.value.measure()
    observeRenderedModelCards()
  },
  { flush: 'post' },
)

watch(virtualRows, async () => {
  await nextTick()
  observeRenderedModelCards()
}, { flush: 'post' })
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
.scheme3-model-square-tool-active { border-color: #1e5c42; color: #1e5c42; }
.scheme3-model-square-filter-count { display: inline-flex; min-width: 1rem; height: 1rem; align-items: center; justify-content: center; margin-left: -.3rem; border-radius: 999px; padding: 0 .22rem; background: #b7791f; color: #fffefa; font-size: .55rem; font-weight: 900; }
.scheme3-model-square-commandbar { display: flex; min-width: 0; align-items: center; gap: .6rem; margin: -.25rem 0 .95rem; }
.scheme3-model-square-commandbar > label { display: flex; min-width: 0; align-items: center; gap: .45rem; color: var(--square-muted); font-size: .625rem; font-weight: 800; }
.scheme3-model-square-commandbar select,
.scheme3-model-square-filter-select select { min-height: 2rem; border: 1px solid var(--square-line); border-radius: 5px; padding: .35rem 1.8rem .35rem .55rem; background: var(--square-card); color: var(--square-ink); font-size: .6875rem; }
.scheme3-model-square-view-switch { display: inline-grid; grid-template-columns: repeat(2,2rem); border: 1px solid var(--square-line); border-radius: 5px; background: var(--square-card); }
.scheme3-model-square-view-switch button { display: inline-flex; width: 2rem; height: 2rem; align-items: center; justify-content: center; color: var(--square-muted); }
.scheme3-model-square-view-switch button + button { border-left: 1px solid var(--square-line); }
.scheme3-model-square-view-switch button.active { background: #1e5c42; color: #f4f2ec; }
.scheme3-model-square-result-count { min-width: 0; margin-left: auto; color: var(--square-muted); font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; font-size: .625rem; font-weight: 700; }
.scheme3-model-square-advanced-filters { display: grid; min-width: 0; grid-template-columns: minmax(15rem,1.4fr) repeat(3,minmax(8.5rem,.75fr)) auto; align-items: end; gap: .85rem; margin-bottom: .95rem; border: 1px solid var(--square-line); border-left: 3px solid #1e5c42; border-radius: 6px; padding: .85rem; background: color-mix(in srgb,var(--square-card) 82%,#f4f2ec); box-shadow: 0 7px 16px rgba(54,48,34,.04); }
.scheme3-model-square-filter-section,
.scheme3-model-square-filter-select { min-width: 0; }
.scheme3-model-square-filter-heading { display: flex; align-items: center; justify-content: space-between; gap: .5rem; margin-bottom: .42rem; }
.scheme3-model-square-filter-heading strong,
.scheme3-model-square-filter-select > span { display: block; color: var(--square-muted); font-size: .625rem; font-weight: 800; }
.scheme3-model-square-filter-heading button { color: #1e5c42; font-size: .625rem; font-weight: 800; }
.scheme3-model-square-capabilities { display: flex; min-width: 0; flex-wrap: wrap; gap: .3rem; }
.scheme3-model-square-capabilities button { display: inline-flex; min-height: 1.75rem; align-items: center; gap: .25rem; border: 1px solid var(--square-line); border-radius: 4px; padding: .25rem .4rem; background: var(--square-card); color: var(--square-muted); font-size: .625rem; font-weight: 700; }
.scheme3-model-square-capabilities button.active { border-color: #1e5c42; background: rgba(30,92,66,.1); color: #1e5c42; }
.scheme3-model-square-filter-select { display: grid; gap: .42rem; }
.scheme3-model-square-filter-select select { width: 100%; }
.scheme3-model-square-reset { display: inline-flex; min-height: 2rem; align-items: center; justify-content: center; gap: .32rem; border: 1px solid rgba(183,121,31,.38); border-radius: 5px; padding: .35rem .58rem; background: rgba(183,121,31,.08); color: #765213; font-size: .625rem; font-weight: 800; white-space: nowrap; }
.scheme3-model-square-note { display: flex; align-items: flex-start; gap: .45rem; margin-bottom: 1.15rem; border-left: 3px solid #b7791f; padding: .52rem .7rem; background: rgba(183,121,31,.08); color: #765213; font-size: .7rem; line-height: 1.55; }
.scheme3-model-square-note svg { flex: 0 0 auto; margin-top: .15rem; }
.scheme3-model-square-layout { display: grid; grid-template-columns: minmax(13.5rem,16rem) minmax(0,1fr); align-items: start; gap: 1.15rem; }
.scheme3-model-square-results { min-width: 0; }
.scheme3-model-square-index-bar { display: none; align-items: center; gap: .55rem; margin-bottom: .85rem; border-bottom: 1px solid var(--square-line); padding: 0 0 .75rem; color: var(--square-muted); font-size: .625rem; }.scheme3-model-square-index-bar label { flex: 0 0 auto; font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; font-weight: 800; letter-spacing: .05em; }.scheme3-model-square-index-bar select { min-width: 0; max-width: 100%; border: 1px solid var(--square-line); border-radius: 5px; padding: .45rem .55rem; background: var(--square-card); color: var(--square-ink); font-size: .6875rem; }.scheme3-model-square-index-bar select:focus-visible { outline: 2px solid rgba(30,92,66,.28); outline-offset: 1px; }.scheme3-model-square-index-bar span { margin-left: auto; font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; white-space: nowrap; }
.scheme3-model-square-state { display: flex; min-height: 19rem; align-items: center; justify-content: center; }.scheme3-model-square-spinner { width: 1.85rem; height: 1.85rem; border: 2px solid rgba(30,92,66,.18); border-top-color: #1e5c42; border-radius: 50%; animation: square-spin .7s linear infinite; }.scheme3-model-square-empty { flex-direction: column; gap: .6rem; border: 1px dashed var(--square-line); color: var(--square-muted); font-size: .75rem; }.scheme3-model-square-empty svg { color: #b7791f; }
.scheme3-model-square-scroll-region {
  min-width: 0;
  overflow-x: auto;
  overscroll-behavior-x: contain;
  padding-bottom: .75rem;
  scrollbar-width: thin;
  scrollbar-color: rgba(30,92,66,.48) rgba(218,213,200,.58);
}

.scheme3-model-square-scroll-region:hover,
.scheme3-model-square-scroll-region:focus-visible {
  scrollbar-color: rgba(30,92,66,.72) rgba(218,213,200,.78);
}

.scheme3-model-square-scroll-region::-webkit-scrollbar {
  width: .625rem;
  height: .625rem;
}

.scheme3-model-square-scroll-region::-webkit-scrollbar-track {
  border-radius: 4px;
  background: rgba(218,213,200,.58);
}

.scheme3-model-square-scroll-region::-webkit-scrollbar-thumb {
  border: 2px solid transparent;
  border-radius: 4px;
  background: rgba(30,92,66,.48);
  background-clip: padding-box;
}

.scheme3-model-square-scroll-region::-webkit-scrollbar-thumb:hover {
  background: rgba(30,92,66,.72);
  background-clip: padding-box;
}
.scheme3-model-square-scroll-region:focus-visible { outline: 2px solid rgba(30,92,66,.24); outline-offset: 3px; }
.scheme3-model-square-scroll-content { position: relative; width: 100%; min-width: 0; padding: .125rem .2rem 1rem; }
.scheme3-model-square-virtual-content { width: 100%; }
.scheme3-model-square-virtual-row {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  box-sizing: border-box;
  padding-bottom: 1rem;
  will-change: transform;
}
.scheme3-model-square-card { min-width: 0; overflow: hidden; scroll-margin-top: calc(var(--scheme3-console-topbar-height,4.25rem) + .75rem); border: 1px solid var(--square-line); border-radius: 8px; background: var(--square-card); box-shadow: 0 9px 20px rgba(54,48,34,.05); }
.scheme3-model-square-card-active { border-color: rgba(30,92,66,.72); box-shadow: 0 0 0 2px rgba(30,92,66,.12), 0 9px 20px rgba(54,48,34,.06); }
.scheme3-model-square-card-header { display: flex; min-width: 0; align-items: center; justify-content: space-between; gap: 1rem; border-bottom: 1px solid var(--square-line); padding: 1rem 1.1rem .9rem; background: #f8f6ef; }
.scheme3-model-square-card-heading { display: flex; min-width: 0; align-items: center; gap: .72rem; }
.scheme3-model-square-card-heading > div { min-width: 0; }
.scheme3-model-square-platform-icon { display: inline-flex; width: 2.35rem; height: 2.35rem; flex: 0 0 auto; align-items: center; justify-content: center; border: 1px solid color-mix(in srgb,var(--square-platform-color,#1e5c42) 35%,var(--square-line)); border-radius: 6px; background: color-mix(in srgb,var(--square-platform-color,#1e5c42) 8%,var(--square-card)); color: var(--square-platform-color,#1e5c42); }
.scheme3-model-square-card-header h2 { display: -webkit-box; max-width: 100%; overflow: hidden; margin: .14rem 0 0; color: var(--square-ink); font-family: Georgia,'Times New Roman',serif; font-size: 1.14rem; font-weight: 600; line-height: 1.18; overflow-wrap: anywhere; -webkit-box-orient: vertical; -webkit-line-clamp: 2; white-space: normal; }
.scheme3-model-square-card-platform,.scheme3-model-square-card-count { display: block; color: var(--square-muted); font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; font-size: .625rem; font-weight: 700; letter-spacing: .05em; }.scheme3-model-square-card-platform { margin: 0; text-transform: uppercase; }.scheme3-model-square-card-count { margin-top: .28rem; }
.scheme3-model-square-card-facts { display: flex; min-width: 0; flex-wrap: wrap; gap: .25rem; margin-top: .38rem; }
.scheme3-model-square-card-facts > span { border: 1px solid var(--square-line); border-radius: 3px; padding: .16rem .3rem; background: var(--square-card); color: var(--square-muted); font-size: .5625rem; font-weight: 700; }
.scheme3-model-square-card-actions { display: flex; flex: 0 0 auto; align-items: center; gap: .45rem; }
.scheme3-model-square-config-toggle { display: inline-flex; min-height: 1.8rem; align-items: center; gap: .25rem; border: 1px solid var(--square-line); border-radius: 5px; padding: .28rem .48rem; background: transparent; color: var(--square-muted); font-size: .6875rem; font-weight: 800; white-space: nowrap; }.scheme3-model-square-config-toggle:hover { border-color: rgba(30,92,66,.45); color: #1e5c42; }.scheme3-model-square-config-toggle svg { transition: transform 150ms ease; }.scheme3-model-square-config-chevron-open { transform: rotate(180deg); }
.scheme3-model-square-channel { display: grid; min-width: 0; grid-template-columns: minmax(12rem,.82fr) minmax(18rem,1.18fr); align-items: start; gap: 1rem; padding: 1rem 1.1rem; border-bottom: 1px solid var(--square-line); }.scheme3-model-square-channel:last-child { border-bottom: 0; }
.scheme3-model-square-channel-meta { min-width: 0; padding-top: .1rem; }
.scheme3-model-square-channel-heading,.scheme3-model-square-channel-groups { min-width: 0; }
.scheme3-model-square-channel-heading { display: flex; align-items: baseline; gap: .55rem; }
.scheme3-model-square-channel-heading strong { min-width: 0; overflow-wrap: anywhere; font-size: .75rem; }
.scheme3-model-square-channel-groups { margin-top: .75rem; }
.scheme3-model-square-section-label { display: block; flex: 0 0 auto; color: var(--square-muted); font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; font-size: .625rem; font-weight: 800; letter-spacing: .06em; text-transform: uppercase; }
.scheme3-model-square-channel-groups > .scheme3-model-square-section-label { margin-bottom: .42rem; }
.scheme3-model-square-groups { display: flex; flex-wrap: wrap; gap: .35rem; min-width: 0; }.model-square-group-badge { max-width: 100%; transition: transform 150ms ease; }.model-square-group-badge:hover { transform: translateY(-1px); }
.scheme3-model-square-pricing { min-width: 0; border-left: 1px solid var(--square-line); padding-left: 1rem; }.scheme3-model-square-pricing-header { display: flex; min-width: 0; justify-content: space-between; gap: .8rem; margin-bottom: .7rem; }.scheme3-model-square-pricing h3 { margin: 0; font-size: .6875rem; font-weight: 800; }.scheme3-model-square-pricing-header span { min-width: 0; color: var(--square-muted); font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; font-size: .625rem; overflow-wrap: anywhere; text-align: right; }.scheme3-model-square-pricing-grid { display: grid; min-width: 0; grid-template-columns: repeat(3,minmax(0,1fr)); gap: .7rem .5rem; }.scheme3-model-square-pricing-grid > div { min-width: 0; }.scheme3-model-square-pricing-grid p,.scheme3-model-square-pricing-extra p { margin: 0 0 .18rem; color: var(--square-muted); font-size: .625rem; }.scheme3-model-square-pricing-grid strong,.scheme3-model-square-pricing-extra strong { display: block; min-width: 0; overflow-wrap: anywhere; font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; font-size: .6875rem; }.scheme3-model-square-pricing-extra { display: flex; flex-wrap: wrap; align-items: end; gap: .8rem; margin-top: .8rem; border-top: 1px solid var(--square-line); padding-top: .7rem; }.scheme3-model-square-pricing-extra > div:first-child strong { color: #1e5c42; }.scheme3-model-square-tier-note { display: inline-flex; align-items: center; gap: .28rem; border: 1px solid rgba(183,121,31,.28); border-radius: 4px; padding: .25rem .4rem; background: rgba(183,121,31,.08); color: #765213; font-size: .625rem; font-weight: 800; }
.scheme3-model-square-pricing-header em { flex: 0 0 auto; border: 1px solid rgba(183,121,31,.3); border-radius: 3px; padding: .13rem .3rem; background: rgba(183,121,31,.08); color: #765213; font-size: .5625rem; font-style: normal; font-weight: 800; white-space: nowrap; }
.scheme3-model-square-tier-details { min-width: 0; margin-top: .7rem; border-top: 1px solid var(--square-line); padding-top: .55rem; }.scheme3-model-square-tier-details summary { display: flex; min-width: 0; align-items: center; justify-content: space-between; gap: .6rem; cursor: pointer; list-style: none; color: #765213; font-size: .625rem; font-weight: 800; }.scheme3-model-square-tier-details summary::-webkit-details-marker { display: none; }.scheme3-model-square-tier-details summary > span { display: inline-flex; min-width: 0; align-items: center; gap: .28rem; overflow-wrap: anywhere; }.scheme3-model-square-tier-chevron { flex: 0 0 auto; transition: transform 150ms ease; }.scheme3-model-square-tier-details[open] .scheme3-model-square-tier-chevron { transform: rotate(180deg); }.scheme3-model-square-tier-list { display: grid; min-width: 0; gap: .35rem; margin-top: .5rem; }.scheme3-model-square-tier-row { display: flex; min-width: 0; align-items: start; justify-content: space-between; gap: .65rem; border-left: 2px solid rgba(183,121,31,.35); padding: .35rem .45rem; background: rgba(183,121,31,.05); font-size: .625rem; }.scheme3-model-square-tier-row > strong { min-width: 0; overflow-wrap: anywhere; font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; }.scheme3-model-square-tier-prices { display: flex; min-width: 0; flex-wrap: wrap; justify-content: end; gap: .3rem .6rem; color: var(--square-ink); font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; }.scheme3-model-square-tier-prices span { max-width: 100%; overflow-wrap: anywhere; }.scheme3-model-square-tier-prices small { margin-right: .2rem; color: var(--square-muted); font-family: inherit; }
.scheme3-model-square-config { min-width: 0; border-top: 1px solid var(--square-line); padding: .85rem 1.1rem 1rem; background: rgba(244,242,236,.42); }.scheme3-model-square-config-title { display: flex; min-width: 0; align-items: center; justify-content: space-between; gap: .7rem; margin-bottom: .6rem; color: var(--square-muted); font-size: .625rem; font-weight: 800; }.scheme3-model-square-config code,.scheme3-model-square-config-channel code { min-width: 0; overflow-wrap: anywhere; color: var(--square-muted); font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; font-size: .625rem; text-align: right; }.scheme3-model-square-config-channel { min-width: 0; border-top: 1px solid var(--square-line); padding: .55rem 0 0; }.scheme3-model-square-config-channel > div:first-child { display: flex; min-width: 0; align-items: baseline; justify-content: space-between; gap: .65rem; }.scheme3-model-square-config-channel strong { min-width: 0; overflow-wrap: anywhere; font-size: .6875rem; }.scheme3-model-square-config-entries { display: grid; min-width: 0; gap: .28rem; margin-top: .4rem; }.scheme3-model-square-config-entries > div { display: flex; min-width: 0; align-items: baseline; justify-content: space-between; gap: .65rem; font-size: .625rem; }
.scheme3-model-square-config-title code,
.scheme3-model-square-config-channel > div:first-child > code,
.scheme3-model-square-config-entries > div > code {
  max-width: 100%;
  white-space: normal;
  word-break: break-word;
}
.scheme3-model-square-config-entries > div > span,
.scheme3-model-square-config-entries > div > code {
  min-width: 0;
  flex: 1 1 0;
  overflow-wrap: anywhere;
}
.scheme3-model-square-config-entries > div > span { max-width: 100%; }
.scheme3-model-square-back-to-top { position: fixed; right: 1.25rem; bottom: calc(1.25rem + env(safe-area-inset-bottom)); z-index: 40; display: inline-flex; width: 2.6rem; height: 2.6rem; align-items: center; justify-content: center; border: 1px solid rgba(30,92,66,.45); border-radius: 6px; background: #1e5c42; color: #f4f2ec; box-shadow: 0 8px 18px rgba(22,21,15,.16); transition: background-color 150ms ease,transform 150ms ease,opacity 150ms ease; }.scheme3-model-square-back-to-top:hover { background: #174a35; }.scheme3-model-square-back-to-top:active { transform: scale(.96); }.scheme3-model-square-back-to-top:focus-visible { outline: 3px solid rgba(30,92,66,.22); outline-offset: 2px; }
@keyframes square-spin { to { transform: rotate(360deg); } }
:global(html.dark .scheme3-model-square-card),
:global(html.dark .scheme3-model-square-search),
:global(html.dark .scheme3-model-square-refresh) {
  border-color: #47443a;
  background: #24231f;
}
:global(html.dark .scheme3-model-square-refresh:hover) { background: #2b2924; }
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
  background: transparent;
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
:global(html.dark .scheme3-model-square-scroll-region) { scrollbar-color: rgba(143,194,165,.65) rgba(71,68,58,.8) !important; }
:global(html.dark .scheme3-model-square-scroll-region::-webkit-scrollbar-track) { background: rgba(71,68,58,.8); }
:global(html.dark .scheme3-model-square-scroll-region::-webkit-scrollbar-thumb) { background: rgba(143,194,165,.65); }
:global(html.dark .scheme3-model-square-scroll-region::-webkit-scrollbar-thumb:hover) { background: rgba(167,208,184,.9); }
:global(html.dark .scheme3-model-square-config-toggle:hover) { border-color: #8fc2a5; color: #a7d0b8; }
:global(html.dark .scheme3-model-square-config) { background: rgba(43,41,36,.55); }
:global(html.dark .scheme3-model-square-tier-details summary) { color: #e8c878; }
:global(html.dark .scheme3-model-square-tier-row) { background: rgba(183,121,31,.12); }
:global(html.dark .scheme3-model-square-index-bar select) { border-color: #47443a; background: #24231f; color: #f4f2ec; }
:global(html.dark .scheme3-model-square-card-active) { border-color: #8fc2a5; box-shadow: 0 0 0 2px rgba(143,194,165,.16), 0 9px 20px rgba(0,0,0,.2); }
:global(html.dark .scheme3-model-square-note) { color: #e8c878; background: rgba(183,121,31,.11); }
:global(html.dark .scheme3-model-square-commandbar select),
:global(html.dark .scheme3-model-square-filter-select select),
:global(html.dark .scheme3-model-square-view-switch),
:global(html.dark .scheme3-model-square-capabilities button),
:global(html.dark .scheme3-model-square-card-facts > span) { border-color: #47443a; background: #24231f; color: #f4f2ec; }
:global(html.dark .scheme3-model-square-advanced-filters) { border-color: #47443a; border-left-color: #8fc2a5; background: #1b1b18; }
:global(html.dark .scheme3-model-square-capabilities button.active) { border-color: #8fc2a5; background: rgba(143,194,165,.12); color: #a7d0b8; }
:global(html.dark .scheme3-model-square-reset),
:global(html.dark .scheme3-model-square-pricing-header em) { border-color: rgba(214,166,93,.35); background: rgba(183,121,31,.12); color: #e8c878; }
:global(html.dark .scheme3-model-square-back-to-top) { border-color: #8fc2a5; background: #8fc2a5; color: #16150f; }
:global(html.dark .scheme3-model-square-back-to-top:hover) { background: #a7d0b8; }
@media (max-width: 900px) {
  .scheme3-model-square-advanced-filters { grid-template-columns: repeat(2,minmax(0,1fr)); }
  .scheme3-model-square-filter-section { grid-column: 1 / -1; }
  .scheme3-model-square-layout { display: block; }
  .scheme3-model-square-index-bar { display: flex; }
  .scheme3-model-square-channel { grid-template-columns: 1fr; gap: .8rem; }
  .scheme3-model-square-pricing { border-top: 1px solid var(--square-line); border-left: 0; padding-top: .8rem; padding-left: 0; }
}
@media (max-width: 640px) {
  .scheme3-model-square-commandbar { flex-wrap: wrap; }
  .scheme3-model-square-commandbar > label { flex: 1 1 12rem; }
  .scheme3-model-square-commandbar select { width: 100%; }
  .scheme3-model-square-result-count { width: 100%; margin-left: 0; }
  .scheme3-model-square-advanced-filters { grid-template-columns: 1fr; }
  .scheme3-model-square-filter-section { grid-column: auto; }
  .scheme3-model-square-filter { display: grid; grid-template-columns: repeat(3,minmax(0,1fr)); }
  .scheme3-model-square-filter-button { width: 100%; padding-right: .35rem; padding-left: .35rem; }
  .scheme3-model-square-index-bar { flex-wrap: wrap; }
  .scheme3-model-square-index-bar label { width: 100%; }
  .scheme3-model-square-index-bar select { order: 3; width: 100%; }
  .scheme3-model-square-header { align-items: stretch; flex-direction: column; }
  .scheme3-model-square-tools,.scheme3-model-square-search { width: 100%; }
  .scheme3-model-square-refresh { flex: 0 0 auto; }
  .scheme3-model-square-pricing-grid { grid-template-columns: repeat(2,minmax(0,1fr)); }
  .scheme3-model-square-card-header { align-items: flex-start; flex-wrap: wrap; padding-right: .8rem; padding-left: .8rem; }
  .scheme3-model-square-card-heading { flex: 1 1 13rem; }
  .scheme3-model-square-card-actions { width: auto; align-self: center; justify-content: flex-end; margin-left: auto; }
  .scheme3-model-square-channel { padding-right: .8rem; padding-left: .8rem; }
  .scheme3-model-square-config-channel > div:first-child,.scheme3-model-square-config-entries > div,.scheme3-model-square-tier-row { align-items: flex-start; flex-direction: column; }
  .scheme3-model-square-config-entries > div > span,.scheme3-model-square-config-entries > div > code { width: 100%; flex: none; text-align: left; }
  .scheme3-model-square-tier-prices { justify-content: flex-start; }
  .scheme3-model-square-note { font-size: .65rem; }
  .scheme3-model-square-back-to-top { right: .8rem; bottom: calc(.8rem + env(safe-area-inset-bottom)); }
}
</style>

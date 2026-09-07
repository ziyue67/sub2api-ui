<template>
  <div
    v-if="model"
    class="scheme3-model-detail-backdrop fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4 backdrop-blur-sm"
    @click.self="$emit('close')"
  >
    <div
      class="scheme3-model-detail-shell relative flex max-h-[90vh] w-full max-w-3xl flex-col overflow-hidden rounded-3xl border border-gray-200 bg-white shadow-2xl dark:border-dark-700 dark:bg-dark-900"
    >
      <!-- 弹窗顶栏 -->
      <div class="scheme3-model-detail-header flex items-center justify-between border-b border-gray-100 px-6 py-4 dark:border-dark-800">
        <div class="flex items-center gap-3">
          <div
            class="flex h-10 w-10 items-center justify-center rounded-xl bg-gray-100 dark:bg-dark-800"
            :class="platformTextClass(model.platform)"
          >
            <PlatformIcon :platform="model.platform" size="md" />
          </div>
          <div>
            <h2 class="text-lg font-bold text-gray-900 dark:text-white">{{ model.name }}</h2>
            <p class="text-xs text-gray-500 dark:text-dark-400">
              {{ platformLabel(model.platform) }} · {{ model.contextWindow || '128K' }} 上下文
            </p>
          </div>
        </div>

        <button
          type="button"
          class="rounded-xl p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-dark-800 dark:hover:text-gray-200"
          aria-label="关闭详情"
          title="关闭详情"
          @click="$emit('close')"
        >
          <Icon name="x" size="sm" />
        </button>
      </div>

      <!-- 选项卡切换：渠道定价对比 vs API 调用代码 -->
      <div class="scheme3-model-detail-tabs flex border-b border-gray-100 px-6 dark:border-dark-800 bg-gray-50/50 dark:bg-dark-950/40">
        <button
          type="button"
          class="border-b-2 px-4 py-3 text-xs font-bold transition-all"
          :class="
            activeTab === 'pricing'
              ? 'border-indigo-600 text-indigo-600 dark:border-indigo-400 dark:text-indigo-400'
              : 'border-transparent text-gray-500 hover:text-gray-800 dark:text-dark-400 dark:hover:text-gray-200'
          "
          @click="activeTab = 'pricing'"
        >
          全渠道定价与倍率
        </button>
        <button
          type="button"
          class="border-b-2 px-4 py-3 text-xs font-bold transition-all"
          :class="
            activeTab === 'curl'
              ? 'border-indigo-600 text-indigo-600 dark:border-indigo-400 dark:text-indigo-400'
              : 'border-transparent text-gray-500 hover:text-gray-800 dark:text-dark-400 dark:hover:text-gray-200'
          "
          @click="activeTab = 'curl'"
        >
          API 快速调用示例
        </button>
      </div>

      <!-- 弹窗内容区域 -->
      <div class="scheme3-model-detail-content overflow-y-auto p-6 space-y-6 flex-1">
        <!-- 渠道定价面板 -->
        <div v-if="activeTab === 'pricing'" class="space-y-4">
          <div
            v-for="channel in model.channels"
            :key="channel.key"
            class="scheme3-model-detail-channel rounded-2xl border border-gray-200/90 bg-gray-50/50 p-4 dark:border-dark-700/80 dark:bg-dark-950/30"
          >
            <!-- 渠道名称与计费模式 -->
            <div class="flex items-center justify-between mb-3">
              <div class="flex items-center gap-2">
                <span class="text-sm font-bold text-gray-900 dark:text-white">{{ channel.name }}</span>
                <span class="rounded bg-indigo-50 px-2 py-0.5 text-[10px] font-bold text-indigo-600 dark:bg-indigo-950/50 dark:text-indigo-400">
                  {{ billingModeLabel(channel.pricing) }}
                </span>
                <span
                  v-if="channel.pricing?.max_reasoning_effort_multiplier"
                  class="scheme3-model-detail-reasoning rounded border px-2 py-0.5 text-[10px] font-bold"
                  :title="`最终推理强度为 max 时，计费与额度消耗乘以 ${channel.pricing.max_reasoning_effort_multiplier}`"
                >
                  Max ×{{ channel.pricing.max_reasoning_effort_multiplier }}
                </span>
                <span v-if="channel.isOfficialFallback" class="rounded bg-amber-500/10 px-2 py-0.5 text-[10px] font-bold text-amber-600 dark:text-amber-400 border border-amber-500/20" title="该渠道未配置独立价格，已自动采用官方参考基准定价">
                  官方参考价
                </span>
              </div>
              <span class="font-mono text-xs text-gray-400">{{ channel.key }}</span>
            </div>

            <!-- 价格网格 -->
            <div class="scheme3-model-detail-price-grid grid grid-cols-2 gap-2 text-xs mb-3 sm:grid-cols-3">
              <div
                v-for="item in channelPriceItems(channel.pricing)"
                :key="item.label"
                class="rounded-lg bg-white p-2 border border-gray-100 dark:bg-dark-900 dark:border-dark-800"
              >
                <span class="text-[10px] text-gray-400 block font-bold uppercase">{{ item.label }}</span>
                <span class="font-mono font-bold text-gray-700 dark:text-gray-300">{{ item.value }}</span>
              </div>
            </div>

            <details
              v-if="channel.pricing?.intervals?.length"
              class="scheme3-model-detail-tiers mb-3 border-t border-gray-200/60 pt-2 dark:border-dark-800"
            >
              <summary class="cursor-pointer text-[11px] font-bold text-amber-700 dark:text-amber-300">
                阶梯价格 · {{ channel.pricing.intervals.length }} 档
              </summary>
              <div class="mt-2 space-y-2">
                <div
                  v-for="(interval, index) in channel.pricing.intervals"
                  :key="`${interval.min_tokens}:${interval.max_tokens ?? 'max'}:${index}`"
                  class="scheme3-model-detail-tier-row"
                >
                  <strong>{{ intervalLabel(interval) }}</strong>
                  <span v-for="item in intervalPriceItems(interval, channel.pricing.billing_mode)" :key="item.label">
                    <small>{{ item.label }}</small>{{ item.value }}
                  </span>
                </div>
              </div>
            </details>

            <!-- 支持分组与用户专属倍率 -->
            <div class="space-y-1.5 pt-2 border-t border-gray-200/60 dark:border-dark-800">
              <span class="text-[11px] font-bold text-gray-400 uppercase tracking-wider block">接入分组与倍率</span>
              <div class="flex flex-wrap gap-2">
                <div
                  v-for="entry in channel.entries"
                  :key="entryKey(entry)"
                  class="flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-2.5 py-1 text-xs dark:border-dark-700 dark:bg-dark-900"
                >
                  <span class="font-medium text-gray-800 dark:text-gray-200">{{ entry.group.name }}</span>
                  <span class="font-mono text-[11px] text-gray-400">
                    倍率 {{ entry.group.rate_multiplier }}
                    <template v-if="userGroupRates[entry.group.id] != null">
                      · 专属 <span class="text-amber-500 font-bold">{{ userGroupRates[entry.group.id] }}</span>
                    </template>
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- API 代码调用示例 -->
        <div v-else class="space-y-4">
          <div class="flex items-center justify-between">
            <span class="text-xs text-gray-500 dark:text-dark-400">使用标准 OpenAI SDK 或 cURL 直接接入该模型：</span>
            <button
              type="button"
              class="scheme3-model-detail-copy inline-flex items-center gap-1 rounded-lg bg-indigo-50 px-2.5 py-1 text-xs font-bold text-indigo-600 hover:bg-indigo-100 dark:bg-indigo-950/50 dark:text-indigo-400"
              @click="copySnippet"
            >
              <Icon :name="copied ? 'check' : 'copy'" size="xs" />
              <span>{{ copied ? '已复制' : '复制命令' }}</span>
            </button>
          </div>

          <pre
            class="scheme3-model-detail-code overflow-x-auto rounded-2xl bg-gray-900 p-4 font-mono text-xs text-gray-200 dark:bg-dark-950 border border-gray-800"
          ><code>{{ curlSnippet }}</code></pre>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import { platformLabel, platformTextClass } from '@/utils/platformColors'
import { entryKey } from '../utils/key'
import {
  billingModeLabel,
  formatRequestPrice,
  formatTokenPrice,
  fullPriceItems,
  isRequestBilling,
  requestPriceLabel,
} from '../utils/pricing'
import type { ModelSquareModel } from '../types'
import type { UserPricingInterval, UserSupportedModelPricing } from '@/api/channels'

interface Props {
  model: ModelSquareModel | null
  userGroupRates: Record<number, number>
}

const props = defineProps<Props>()
defineEmits<{
  close: []
}>()

const activeTab = ref<'pricing' | 'curl'>('pricing')
const copied = ref(false)

function channelPriceItems(pricing: UserSupportedModelPricing | null) {
  const items = fullPriceItems(pricing)
    .filter((item) => item.value != null)
    .map((item) => ({ label: item.label, value: formatTokenPrice(item.value) }))
  if (pricing && isRequestBilling(pricing) && pricing.per_request_price != null) {
    items.push({
      label: requestPriceLabel(pricing.billing_mode),
      value: formatRequestPrice(pricing.per_request_price, pricing.billing_mode),
    })
  }
  return items.length ? items : [{ label: '价格', value: '未配置' }]
}

function intervalLabel(interval: UserPricingInterval) {
  if (interval.tier_label?.trim()) return interval.tier_label
  return interval.max_tokens == null
    ? `>${interval.min_tokens.toLocaleString()} tokens`
    : `${interval.min_tokens.toLocaleString()}–${interval.max_tokens.toLocaleString()} tokens`
}

function intervalPriceItems(interval: UserPricingInterval, billingMode?: string) {
  if (billingMode === 'image' || billingMode === 'video' || billingMode === 'per_request') {
    return [{
      label: requestPriceLabel(billingMode),
      value: formatRequestPrice(interval.per_request_price, billingMode),
    }]
  }
  return [
    { label: '输入', value: interval.input_price },
    { label: '输出', value: interval.output_price },
    { label: '缓存写入 (5m)', value: interval.cache_write_price },
    { label: '缓存写入 (1h)', value: interval.cache_write_1h_price },
    { label: '缓存读取', value: interval.cache_read_price },
  ]
    .filter((item) => item.value != null)
    .map((item) => ({ label: item.label, value: formatTokenPrice(item.value) }))
}

const curlSnippet = computed(() => {
  const modelName = props.model?.name || 'gpt-4o'
  const origin = window.location.origin
  return `curl ${origin}/v1/chat/completions \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -d '{
    "model": "${modelName}",
    "messages": [
      {
        "role": "user",
        "content": "Hello world!"
      }
    ]
  }'`
})

function copySnippet() {
  navigator.clipboard.writeText(curlSnippet.value)
  copied.value = true
  setTimeout(() => {
    copied.value = false
  }, 2000)
}
</script>

<style scoped>
.scheme3-model-detail-backdrop { background: rgba(22,21,15,.68); backdrop-filter: blur(4px); }
.scheme3-model-detail-shell { border-color: var(--scheme3-line,#dad5c8); border-radius: 8px; background: var(--scheme3-card,#fbfaf6); color: var(--scheme3-ink,#16150f); box-shadow: 0 24px 70px rgba(22,21,15,.28); }
.scheme3-model-detail-header { border-color: var(--scheme3-line,#dad5c8); background: #f8f6ef; }
.scheme3-model-detail-header h2 { font-family: Georgia,'Times New Roman',serif; font-weight: 600; letter-spacing: 0; }
.scheme3-model-detail-header button { border-radius: 4px; }
.scheme3-model-detail-tabs { border-color: var(--scheme3-line,#dad5c8); background: var(--scheme3-paper,#f4f2ec); }
.scheme3-model-detail-tabs button { border-color: transparent; color: var(--scheme3-muted,#6b695f); }
.scheme3-model-detail-tabs button[class*='border-indigo'] { border-color: #1e5c42; color: #1e5c42; }
.scheme3-model-detail-content { scrollbar-color: rgba(30,92,66,.48) rgba(218,213,200,.58); }
.scheme3-model-detail-channel { border-color: var(--scheme3-line,#dad5c8); border-radius: 6px; background: rgba(244,242,236,.52); }
.scheme3-model-detail-channel [class*='bg-indigo'] { border-radius: 3px; background: rgba(30,92,66,.09); color: #1e5c42; }
.scheme3-model-detail-channel [class*='bg-amber'] { border-radius: 3px; }
.scheme3-model-detail-reasoning { border-color: rgba(30,92,66,.28); background: rgba(30,92,66,.08); color: #1e5c42; }
.scheme3-model-detail-channel [class*='rounded-lg'] { border-radius: 4px; }
.scheme3-model-detail-tier-row { display: flex; min-width: 0; flex-wrap: wrap; align-items: baseline; gap: .35rem .75rem; border-left: 2px solid rgba(183,121,31,.35); padding: .45rem .55rem; background: rgba(183,121,31,.05); font-size: .65rem; }
.scheme3-model-detail-tier-row > strong { margin-right: auto; overflow-wrap: anywhere; }
.scheme3-model-detail-tier-row > span { font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; }
.scheme3-model-detail-tier-row small { margin-right: .2rem; color: var(--scheme3-muted,#6b695f); }
.scheme3-model-detail-copy { border: 1px solid rgba(30,92,66,.3); border-radius: 4px; background: rgba(30,92,66,.08); color: #1e5c42; }
.scheme3-model-detail-code { border-color: #47443a; border-radius: 5px; background: #1b1b18; }
:global(html.dark .scheme3-model-detail-shell) { border-color: #47443a; background: #24231f; color: #f4f2ec; }
:global(html.dark .scheme3-model-detail-header),
:global(html.dark .scheme3-model-detail-tabs) { border-color: #47443a; background: #1b1b18; }
:global(html.dark .scheme3-model-detail-tabs button[class*='border-indigo']) { border-color: #8fc2a5; color: #a7d0b8; }
:global(html.dark .scheme3-model-detail-channel) { border-color: #47443a; background: rgba(27,27,24,.72); }
:global(html.dark .scheme3-model-detail-channel [class*='bg-indigo']) { background: rgba(143,194,165,.12); color: #a7d0b8; }
:global(html.dark .scheme3-model-detail-reasoning) { border-color: rgba(143,194,165,.32); background: rgba(143,194,165,.12); color: #a7d0b8; }
:global(html.dark .scheme3-model-detail-tier-row) { background: rgba(183,121,31,.12); }
:global(html.dark .scheme3-model-detail-copy) { border-color: #8fc2a5; background: rgba(143,194,165,.12); color: #a7d0b8; }
@media (max-width: 640px) {
  .scheme3-model-detail-backdrop { align-items: flex-end; padding: .5rem; }
  .scheme3-model-detail-shell { max-height: calc(100dvh - 1rem); }
  .scheme3-model-detail-header,
  .scheme3-model-detail-content { padding-right: 1rem; padding-left: 1rem; }
  .scheme3-model-detail-tabs { padding-right: 1rem; padding-left: 1rem; overflow-x: auto; }
}
</style>

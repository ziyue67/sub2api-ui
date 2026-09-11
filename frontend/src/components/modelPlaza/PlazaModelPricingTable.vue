<template>
  <div class="scheme3-plaza-pricing-table overflow-x-auto" :style="accentStyle">
    <table class="w-full min-w-[860px] table-auto border-collapse text-sm tabular-nums">
      <colgroup>
        <col class="w-[22%]" />
        <col class="w-[10%]" />
        <col class="w-[10%]" />
        <col class="w-[14%]" />
        <col class="w-[10%]" />
        <col class="w-[10%]" />
        <col class="w-[14%]" />
        <col class="w-[10%]" />
      </colgroup>
      <thead>
        <tr
          class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-dark-400"
        >
          <th
            rowspan="2"
            class="border-r border-gray-100 py-2.5 pl-5 pr-4 text-left align-middle dark:border-dark-700/60"
          >
            {{ t('modelPlaza.table.model') }}
          </th>
          <th colspan="3" class="pz-bg pt-2 text-center">
            <div class="pz-title border-b pb-2 font-semibold">
              {{ t('modelPlaza.table.paidPrice') }}
              <span class="pz-unit ml-1 normal-case font-normal">{{ t('modelPlaza.table.unitPerMillion') }}</span>
            </div>
          </th>
          <th
            colspan="3"
            class="border-l border-gray-100 pt-2 text-center dark:border-dark-700/60"
          >
            <div class="border-b border-gray-200 pb-2 text-gray-400 dark:border-dark-600 dark:text-dark-500">
              {{ t('modelPlaza.table.officialPrice') }}
              <span class="ml-1 normal-case font-normal text-gray-400 dark:text-dark-500">{{ t('modelPlaza.table.unitPerMillion') }}</span>
            </div>
          </th>
          <th
            rowspan="2"
            class="border-l border-gray-100 py-2.5 pl-3 pr-5 text-right align-middle dark:border-dark-700/60"
          >
            {{ t('modelPlaza.table.rate') }}
          </th>
        </tr>
        <tr
          class="border-b border-gray-200 text-left text-[11px] font-medium uppercase leading-4 tracking-wide text-gray-400 dark:border-dark-700 dark:text-dark-500"
        >
          <th class="pz-bg px-3 py-2 font-medium">{{ t('modelPlaza.table.input') }}</th>
          <th class="pz-bg px-3 py-2 font-medium">{{ t('modelPlaza.table.output') }}</th>
          <th class="pz-bg px-3 py-2 font-medium">{{ t('modelPlaza.table.cache') }}</th>
          <th class="border-l border-gray-100 px-3 py-2 font-medium dark:border-dark-700/60">
            {{ t('modelPlaza.table.input') }}
          </th>
          <th class="px-3 py-2 font-medium">{{ t('modelPlaza.table.output') }}</th>
          <th class="px-3 py-2 font-medium">{{ t('modelPlaza.table.cache') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="{ model: m, period, key } in rows"
          :key="key"
          class="border-b border-gray-100 transition-colors last:border-b-0 hover:bg-gray-50/70 dark:border-dark-800 dark:hover:bg-dark-800/50"
        >
          <!-- 模型名 + 非 token 计费模式徽章 -->
          <td class="border-r border-gray-100 py-2.5 pl-5 pr-4 align-middle dark:border-dark-700/60">
            <div class="flex flex-wrap items-center gap-1.5">
              <span class="font-medium text-gray-900 dark:text-white">{{ m.name }}</span>
              <span
                v-if="period"
                class="inline-flex items-center whitespace-nowrap rounded-md bg-gray-100 px-1 py-0.5 font-mono text-[10px] font-medium text-gray-500 dark:bg-dark-700/70 dark:text-dark-300"
                :title="timePricingRowHint(m)"
              >
                <span v-if="m.time_pricing?.weekdays_only" class="mr-1 font-sans">{{
                  t('modelPlaza.table.timePricingWeekdays')
                }}</span>
                {{ formatTimeWindow(period) }}
              </span>
              <span
                v-if="platform && m.platform !== platform"
                class="pz-platform-badge inline-flex items-center rounded-md px-1.5 py-0.5 text-[10px] font-medium"
              >
                {{ platformLabel(m.platform) }}
              </span>
              <span
                v-if="billingMode(m) !== BILLING_MODE_TOKEN"
                class="rounded-md bg-gray-100 px-1.5 py-0.5 text-[10px] font-medium text-gray-500 dark:bg-dark-700/70 dark:text-dark-300"
              >
                {{ billingModeLabel(m) }}
              </span>
              <span
                v-if="m.long_context_basis === 'marginal'"
                class="rounded-md bg-gray-100 px-1.5 py-0.5 text-[10px] font-medium text-gray-500 dark:bg-dark-700/70 dark:text-dark-300"
                :title="t('modelPlaza.table.tierHintMarginal')"
              >
                {{ t('modelPlaza.table.marginalBadge') }}
              </span>
              <span
                v-if="m.pricing?.max_reasoning_effort_multiplier"
                class="pz-reasoning-badge inline-flex items-center rounded-md border px-1.5 py-0.5 text-[10px] font-semibold"
                :title="
                  t('modelPlaza.table.maxReasoningMultiplierHint', {
                    multiplier: m.pricing.max_reasoning_effort_multiplier
                  })
                "
              >
                {{
                  t('modelPlaza.table.maxReasoningMultiplierBadge', {
                    multiplier: m.pricing.max_reasoning_effort_multiplier
                  })
                }}
              </span>
            </div>
          </td>

          <!-- token 计费:输入 / 输出(阶梯内联)/ 缓存(写/读) -->
          <template v-if="billingMode(m) === BILLING_MODE_TOKEN">
            <td class="pz-cell px-3 py-2.5 align-middle font-mono font-semibold text-gray-900 dark:text-gray-50">
              <template v-if="tokenIntervals(m).length">
                <div
                  v-for="(iv, idx) in tokenIntervals(m)"
                  :key="idx"
                  class="pz-tier-entry text-xs leading-5"
                >
                  <span class="pz-tier-label mr-1 font-sans font-normal text-gray-400 dark:text-dark-500" :title="tierHint(m)">{{ tierLabel(iv) }}</span>
                  {{ ' ' }}
                  <span class="pz-tier-value">{{ paidPerMillion(iv.input_price, period) }}</span>
                </div>
              </template>
              <template v-else>{{ paidPerMillion(m.pricing?.input_price, period) }}</template>
            </td>
            <td class="pz-cell px-3 py-2.5 align-middle font-mono font-semibold text-gray-900 dark:text-gray-50">
              <template v-if="tokenIntervals(m).length">
                <div
                  v-for="(iv, idx) in tokenIntervals(m)"
                  :key="idx"
                  class="pz-tier-entry text-xs leading-5"
                  :title="tierHint(m)"
                >
                  <template v-if="tokenIntervals(m).length === 1">
                    <span class="pz-tier-label mr-1 font-sans font-normal text-gray-400 dark:text-dark-500">{{ tierLabel(iv) }}</span>
                    {{ ' ' }}
                  </template>
                  <span class="pz-tier-value">{{ paidPerMillion(iv.output_price, period) }}</span>
                </div>
              </template>
              <template v-else>{{ paidPerMillion(m.pricing?.output_price, period) }}</template>
            </td>
            <td class="pz-cell px-3 py-2.5 align-middle">
              <template v-if="hasTierCachePricing(tokenIntervals(m))">
                <div
                  v-for="(iv, idx) in tokenIntervals(m)"
                  :key="idx"
                  class="pz-wrap-value font-mono text-xs leading-5 text-gray-800 dark:text-gray-200"
                  :title="tierHint(m)"
                >
                  <template v-if="iv.cache_write_price != null || iv.cache_write_1h_price != null || iv.cache_read_price != null">
                    <span class="font-sans font-normal text-gray-400 dark:text-dark-500">{{ t('modelPlaza.table.cacheWriteShort') }}</span>
                    {{ paidPerMillion(iv.cache_write_price, period) }}
                    <template v-if="iv.cache_write_1h_price != null">
                      <span class="font-sans font-normal text-gray-400 dark:text-dark-500"> (1h </span>{{ paidPerMillion(iv.cache_write_1h_price, period) }}<span class="font-sans font-normal text-gray-400 dark:text-dark-500">)</span>
                    </template>
                    <span class="ml-1 font-sans font-normal text-gray-400 dark:text-dark-500">{{ t('modelPlaza.table.cacheReadShort') }}</span>
                    {{ paidPerMillion(iv.cache_read_price, period) }}
                  </template>
                  <span v-else class="text-gray-400 dark:text-dark-500">-</span>
                </div>
              </template>
              <div
                v-else-if="hasCachePricing(m)"
                class="pz-price-stack font-mono text-xs text-gray-800 dark:text-gray-200"
              >
                <div class="pz-price-entry">
                  <span class="pz-price-label font-sans font-normal text-gray-400 dark:text-dark-500">{{ t('modelPlaza.table.cacheWrite') }}</span>
                  <span class="pz-price-value">
                    {{ paidPerMillion(m.pricing?.cache_write_price, period) }}
                    <template v-if="m.pricing?.cache_write_1h_price != null">
                      <span class="pz-price-note">(1h {{ paidPerMillion(m.pricing.cache_write_1h_price, period) }})</span>
                    </template>
                  </span>
                </div>
                <div class="pz-price-entry">
                  <span class="pz-price-label font-sans font-normal text-gray-400 dark:text-dark-500">{{ t('modelPlaza.table.cacheRead') }}</span>
                  <span class="pz-price-value">{{ paidPerMillion(m.pricing?.cache_read_price, period) }}</span>
                </div>
              </div>
              <span v-else class="text-gray-400 dark:text-dark-500">-</span>
            </td>
          </template>

          <!-- 按次 / 按图片计费:实付区整体合并,阶梯芯片或单一按次价 -->
          <template v-else>
            <td colspan="3" class="pz-cell px-3 py-2.5 align-middle">
              <div
                v-if="requestIntervals(m).length"
                class="flex flex-wrap items-start gap-1.5"
              >
                <span
                  v-for="(iv, idx) in requestIntervals(m)"
                  :key="idx"
                  class="pz-request-tier inline-flex rounded-md bg-gray-100 px-2 py-0.5 font-mono text-xs text-gray-800 dark:bg-dark-700/60 dark:text-gray-200"
                >
                  <span class="pz-tier-label font-sans text-gray-400 dark:text-dark-500">{{ tierLabel(iv) }}</span>
                  <span class="pz-tier-value"
                    >{{ paidRequestPrice(m, iv.per_request_price)
                    }}<span class="ml-1 font-sans font-normal text-gray-400 dark:text-dark-500">{{ perUnitSuffix(m) }}</span></span
                  >
                </span>
              </div>
              <template v-else-if="m.pricing?.per_request_price != null">
                <span class="font-mono font-semibold text-gray-900 dark:text-gray-50">
                  {{ paidRequestPrice(m, m.pricing.per_request_price) }}
                </span>
                <span class="ml-1 text-xs text-gray-400 dark:text-dark-500">{{ perUnitSuffix(m) }}</span>
              </template>
              <span v-else class="text-gray-400 dark:text-dark-500">-</span>
            </td>
          </template>

          <!-- 官方价格(参考价,不乘倍率;官方有阶梯时每档一行) -->
          <td
            class="border-l border-gray-100 px-3 py-2.5 align-middle font-mono text-xs text-gray-500 dark:border-dark-700/60 dark:text-dark-400"
          >
            <template v-if="officialIntervals(m).length">
              <div
                v-for="(iv, idx) in officialIntervals(m)"
                :key="idx"
                class="pz-wrap-value leading-5"
              >
                <span class="mr-1 font-sans text-gray-400 dark:text-dark-500" :title="t('modelPlaza.table.tierHint')">{{ tierLabel(iv) }}</span>
                {{ official(iv.input_price) }}
              </div>
            </template>
            <template v-else>{{ official(m.official_pricing?.input_price) }}</template>
          </td>
          <td class="px-3 py-2.5 align-middle font-mono text-xs text-gray-500 dark:text-dark-400">
            <template v-if="officialIntervals(m).length">
              <div
                v-for="(iv, idx) in officialIntervals(m)"
                :key="idx"
                class="pz-wrap-value leading-5"
                :title="t('modelPlaza.table.tierHint')"
              >
                {{ official(iv.output_price) }}
              </div>
            </template>
            <template v-else>{{ official(m.official_pricing?.output_price) }}</template>
          </td>
          <td class="px-3 py-2.5 align-middle">
            <div
              v-if="hasTierCachePricing(officialIntervals(m))"
              class="font-mono text-xs text-gray-500 dark:text-dark-400"
            >
              <div
                v-for="(iv, idx) in officialIntervals(m)"
                :key="idx"
                class="pz-wrap-value leading-5"
                :title="t('modelPlaza.table.tierHint')"
              >
                <template v-if="iv.cache_write_price != null || iv.cache_write_1h_price != null || iv.cache_read_price != null">
                  <span class="font-sans text-gray-400 dark:text-dark-500">{{ t('modelPlaza.table.cacheWriteShort') }}</span>
                  {{ official(iv.cache_write_price) }}
                  <template v-if="iv.cache_write_1h_price != null">
                    <span class="font-sans text-gray-400 dark:text-dark-500"> (1h </span>{{ official(iv.cache_write_1h_price) }}<span class="font-sans text-gray-400 dark:text-dark-500">)</span>
                  </template>
                  <span class="ml-1 font-sans text-gray-400 dark:text-dark-500">{{ t('modelPlaza.table.cacheReadShort') }}</span>
                  {{ official(iv.cache_read_price) }}
                </template>
                <span v-else class="text-gray-400 dark:text-dark-500">-</span>
              </div>
            </div>
            <div
              v-else-if="m.official_pricing && hasOfficialCache(m.official_pricing)"
              class="pz-price-stack font-mono text-xs text-gray-500 dark:text-dark-400"
            >
              <div class="pz-price-entry">
                <span class="pz-price-label font-sans font-normal text-gray-400 dark:text-dark-500">{{ t('modelPlaza.table.cacheWrite') }}</span>
                <span class="pz-price-value">{{ official(m.official_pricing.cache_write_price) }}</span>
                <span
                  v-if="m.official_pricing.cache_write_1h_price != null"
                  class="pz-price-value"
                  >(<span class="font-sans font-normal text-gray-400 dark:text-dark-500">1h </span>{{ official(m.official_pricing.cache_write_1h_price) }})</span
                >
              </div>
              <div class="pz-price-entry">
                <span class="pz-price-label font-sans font-normal text-gray-400 dark:text-dark-500">{{ t('modelPlaza.table.cacheRead') }}</span>
                <span class="pz-price-value">{{ official(m.official_pricing.cache_read_price) }}</span>
              </div>
            </div>
            <span v-else class="text-gray-400 dark:text-dark-500">-</span>
          </td>

          <!-- 折扣倍率(分时时段行展示 生效倍率×时段倍率;生图独立倍率行展示独立倍率;专属倍率划线展示原倍率) -->
          <td
            class="border-l border-gray-100 py-2.5 pl-3 pr-5 text-right align-middle font-mono text-xs dark:border-dark-700/60"
          >
            <span
              v-if="period"
              class="font-bold text-primary-600 dark:text-primary-400"
              :title="t('modelPlaza.table.timePricingRateHint', { rate: effectiveRate, multiplier: period.multiplier })"
              >{{ periodRate(period) }}x</span
            >
            <span
              v-else-if="usesIndependentImageRate(m)"
              class="font-bold text-gray-700 dark:text-gray-300"
              >{{ requestRate(m) }}x</span
            >
            <template v-else-if="hasCustomRate">
              <span class="mr-1 text-gray-400 line-through dark:text-dark-500">{{ rateMultiplier }}x</span>
              <span class="font-bold text-primary-600 dark:text-primary-400">{{ effectiveRate }}x</span>
            </template>
            <span v-else class="font-bold text-gray-700 dark:text-gray-300">{{ effectiveRate }}x</span>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { formatScaled, resolveIntervalPrices } from '@/utils/pricing'
import { platformAccentColor, platformLabel } from '@/utils/platformColors'
import {
  BILLING_MODE_TOKEN,
  BILLING_MODE_IMAGE,
  type BillingMode
} from '@/constants/channel'
import type { PlazaModel, PlazaTimePricingPeriod } from '@/api/modelPlaza'
import type { UserPricingInterval } from '@/api/channels'

const props = defineProps<{
  models: PlazaModel[]
  /** 分组平台;实付分区底色随平台着色,未知平台回退品牌青。 */
  platform?: string
  /** 分组默认倍率。 */
  rateMultiplier: number
  /** 用户专属倍率;与默认不同,实付价按此计算并划线展示原倍率。 */
  userRateMultiplier?: number | null
  /** 生图独立倍率:true 时图片计费模型的实付倍率取 imageRateMultiplier,不取分组/专属倍率。 */
  imageRateIndependent?: boolean
  imageRateMultiplier?: number | null
  /** 高峰窗口说明,仅用于分时时段行 tooltip。 */
  peakWindow?: string
  peakRateMultiplier?: number | null
}>()

const { t } = useI18n()

/** 实付分区保留 UI1 配色;同时暴露平台色供上游主题/扩展样式使用。 */
const accentStyle = computed(() => ({ '--plaza-accent': platformAccentColor(props.platform ?? '') }))

const PER_MILLION = 1_000_000

/**
 * 展示顺序:
 * 1. token 计费的排在前,按图/按次计费的沉到末尾——它们的官方 token 价与实付的按张/按次价不同量纲,混排无意义;
 * 2. 组内按官方输出价从高到低,无官方价的排最后;
 * 3. 同价按名称降序(新版本号在前,如 gpt-5.6 先于 gpt-5.5)。
 */
const sortedModels = computed(() => {
  return [...props.models].sort((a, b) => {
    const ta = billingMode(a) === BILLING_MODE_TOKEN
    const tb = billingMode(b) === BILLING_MODE_TOKEN
    if (ta !== tb) return ta ? -1 : 1
    const pa = a.official_pricing?.output_price ?? null
    const pb = b.official_pricing?.output_price ?? null
    if (pa != null && pb != null && pa !== pb) return pb - pa
    if (pa != null && pb == null) return -1
    if (pa == null && pb != null) return 1
    return b.name.localeCompare(a.name)
  })
})

const effectiveRate = computed(() => props.userRateMultiplier ?? props.rateMultiplier)
const hasCustomRate = computed(
  () => props.userRateMultiplier != null && props.userRateMultiplier !== props.rateMultiplier
)

function billingMode(m: PlazaModel): BillingMode {
  return (m.pricing?.billing_mode || BILLING_MODE_TOKEN) as BillingMode
}

function billingModeLabel(m: PlazaModel): string {
  return billingMode(m) === BILLING_MODE_IMAGE
    ? t('modelPlaza.table.perImage')
    : t('modelPlaza.table.perRequest')
}

/** 价格统一保底 2 位小数,更长的有效小数原样保留。 */
const MIN_DECIMALS = 2

/** 表格行:每个模型一行标准价;配置分时倍率时追加时段行。 */
interface PlazaRow {
  model: PlazaModel
  period: PlazaTimePricingPeriod | null
  key: string
}

const rows = computed<PlazaRow[]>(() =>
  sortedModels.value.flatMap((m) => {
    const base: PlazaRow = { model: m, period: null, key: `${m.platform}:${m.name}` }
    const periodRows = timePeriods(m).map<PlazaRow>((p, idx) => ({
      model: m,
      period: p,
      key: `${m.platform}:${m.name}:${idx}`
    }))
    return [base, ...periodRows]
  })
)

/** 时段行的生效倍率 = 生效倍率 × 时段倍率。 */
function periodRate(period: PlazaTimePricingPeriod): number {
  return Math.round(effectiveRate.value * period.multiplier * 1000) / 1000
}

/** 实付价 = 渠道单价 × 生效倍率(时段行再乘时段倍率),按 $/1M token 展示。 */
function paidPerMillion(value: number | null | undefined, period: PlazaTimePricingPeriod | null = null): string {
  if (value == null) return '-'
  const rate = period ? periodRate(period) : effectiveRate.value
  return formatScaled(value * rate, PER_MILLION, MIN_DECIMALS)
}

/** 图片计费模型且分组开启生图独立倍率:实付倍率取独立倍率,与计费口径一致。 */
function usesIndependentImageRate(m: PlazaModel): boolean {
  return billingMode(m) === BILLING_MODE_IMAGE && props.imageRateIndependent === true
}

/** 按次/按图片行的生效倍率。 */
function requestRate(m: PlazaModel): number {
  return usesIndependentImageRate(m) ? (props.imageRateMultiplier ?? 1) : effectiveRate.value
}

/** 按次 / 按图片单价(乘该行生效倍率,不换算 1M)。 */
function paidRequestPrice(m: PlazaModel, value: number | null | undefined): string {
  if (value == null) return '-'
  return formatScaled(value * requestRate(m), 1, MIN_DECIMALS)
}

/** 官方参考价不乘倍率。 */
function official(value: number | null | undefined): string {
  if (value == null) return '-'
  return formatScaled(value, PER_MILLION, MIN_DECIMALS)
}

/** 非 token 计费的单位后缀:按图片 → “/ 张”,按次 → “/ 次”。 */
function perUnitSuffix(m: PlazaModel): string {
  return billingMode(m) === BILLING_MODE_IMAGE
    ? t('modelPlaza.table.perUnitImage')
    : t('modelPlaza.table.perUnitRequest')
}

function hasCachePricing(m: PlazaModel): boolean {
  return m.pricing?.cache_write_price != null || m.pricing?.cache_write_1h_price != null || m.pricing?.cache_read_price != null
}

function hasOfficialCache(o: NonNullable<PlazaModel['official_pricing']>): boolean {
  return o.cache_write_price != null || o.cache_read_price != null || o.cache_write_1h_price != null
}

/** token 模式的阶梯定价(内联进输入/输出列)。 */
/** 分时倍率时段（后端仅返回倍率不等于 1 的时段）。 */
function timePeriods(m: PlazaModel): PlazaTimePricingPeriod[] {
  return m.time_pricing?.periods ?? []
}

function timePricingRowHint(m: PlazaModel): string {
  const key = m.time_pricing?.weekdays_only
    ? 'modelPlaza.table.timePricingRowHintWeekdays'
    : 'modelPlaza.table.timePricingRowHint'
  let hint = t(key, { timezone: m.time_pricing?.timezone })
  if (props.peakWindow) {
    hint += t('modelPlaza.table.timePricingRowHintPeak', {
      window: props.peakWindow,
      multiplier: props.peakRateMultiplier ?? 1
    })
  }
  return hint
}

function formatTimeWindow(p: PlazaTimePricingPeriod): string {
  const clock = (v: string) => v.replace(/^(\d{2}:\d{2}):00$/, '$1')
  return `${clock(p.start_time)}–${clock(p.end_time)}`
}

function sortByContext(intervals: UserPricingInterval[]): UserPricingInterval[] {
  return [...intervals].sort((a, b) => a.min_tokens - b.min_tokens)
}

function tokenIntervals(m: PlazaModel): UserPricingInterval[] {
  return sortByContext(m.pricing?.intervals ?? []).map(iv => resolveIntervalPrices(iv, m.pricing!))
}

function officialIntervals(m: PlazaModel): UserPricingInterval[] {
  return sortByContext(m.official_pricing?.intervals ?? [])
}

function hasTierCachePricing(intervals: UserPricingInterval[]): boolean {
  return intervals.some((iv) =>
    iv.cache_write_price != null || iv.cache_write_1h_price != null || iv.cache_read_price != null ||
    iv.cache_write_multiplier != null || iv.cache_read_multiplier != null
  )
}

function tierHint(m: PlazaModel): string {
  return m.long_context_basis === 'marginal'
    ? t('modelPlaza.table.tierHintMarginal')
    : t('modelPlaza.table.tierHint')
}

/** 按次/按图模式的阶梯定价(仅保留配了按次价的档位)。 */
function requestIntervals(m: PlazaModel): UserPricingInterval[] {
  return (m.pricing?.intervals ?? []).filter((iv) => iv.per_request_price != null)
}

/** 档位标签:优先后端/管理员给出的 tier_label,否则按区间生成统一形态。 */
function tierLabel(iv: UserPricingInterval): string {
  if (iv.tier_label) return iv.tier_label
  const { min_tokens: min, max_tokens: max } = iv
  return max == null ? `>${formatTokenCount(min)}` : `≤${formatTokenCount(max)}`
}

function formatTokenCount(n: number): string {
  if (n >= 1_000_000) return `${trimZero(n / 1_000_000)}M`
  if (n >= 1_000) return `${trimZero(n / 1_000)}K`
  return String(n)
}

function trimZero(n: number): string {
  return String(Math.round(n * 100) / 100)
}
</script>

<style scoped>
.scheme3-plaza-pricing-table { --pz-title: #1e5c42; --pz-bg: #f1eee6; --pz-bg-hover: #e8eee9; border: 1px solid #d8d2c3; border-radius: 8px; background: #fffefa; }
.scheme3-plaza-pricing-table table { color: #27251f; }
.scheme3-plaza-pricing-table thead { background: #f8f5ed; }
.scheme3-plaza-pricing-table th, .scheme3-plaza-pricing-table td { border-color: #d8d2c3 !important; }
.scheme3-plaza-pricing-table tbody tr:hover { background: #f8f5ed !important; }
.pz-platform-badge { border: 1px solid rgba(30,92,66,.22); background: rgba(30,92,66,.09); color: #1e5c42; }
.pz-reasoning-badge { border-color: rgba(30,92,66,.24); background: rgba(30,92,66,.08); color: #1e5c42; }
.pz-bg, .pz-cell { background-color: var(--pz-bg); }
.pz-cell { transition: background-color 150ms ease; }
tbody tr:hover .pz-cell { background-color: var(--pz-bg-hover); }
.pz-title { color: var(--pz-title); border-color: rgba(30,92,66,.22); }
.pz-unit { color: #777266; }
.pz-tier-entry, .pz-price-entry, .pz-request-tier { min-width: 0; max-width: 100%; }
.pz-tier-entry, .pz-price-entry { display: flex; flex-direction: column; align-items: flex-start; gap: .125rem; line-height: 1.25; }
.pz-tier-entry + .pz-tier-entry { margin-top: .375rem; }
.pz-tier-label, .pz-price-label { display: block; max-width: 100%; overflow-wrap: anywhere; white-space: normal; line-height: 1.25; }
.pz-tier-value, .pz-price-value, .pz-wrap-value { display: block; max-width: 100%; overflow-wrap: anywhere; word-break: break-word; white-space: normal; }
.pz-price-stack { display: flex; min-width: 0; flex-direction: column; gap: .375rem; }
.pz-request-tier { flex-direction: column; align-items: flex-start; gap: .125rem; line-height: 1.25; overflow: hidden; }
.scheme3-plaza-pricing-table th, .scheme3-plaza-pricing-table td { min-width: 0; overflow: hidden; }

:global(.dark .scheme3-plaza-pricing-table) { --pz-title: #8fc2a5; --pz-bg: #2b2924; --pz-bg-hover: #33322c; border-color: #47443a; background: #24231f; }
:global(.dark .scheme3-plaza-pricing-table table) { color: #f4f2ec; }
:global(.dark .scheme3-plaza-pricing-table thead) { background: #2b2924; }
:global(.dark .scheme3-plaza-pricing-table th), :global(.dark .scheme3-plaza-pricing-table td) { border-color: #47443a !important; }
:global(.dark .scheme3-plaza-pricing-table tbody tr:hover) { background: #2b2924 !important; }
:global(html.dark .pz-platform-badge) { border-color: rgba(143,194,165,.3); background: rgba(143,194,165,.12); color: #8fc2a5; }
:global(html.dark .pz-reasoning-badge) { border-color: rgba(143,194,165,.32); background: rgba(143,194,165,.12); color: #8fc2a5; }
:global(.dark .pz-unit) { color: #aaa69a; }
</style>

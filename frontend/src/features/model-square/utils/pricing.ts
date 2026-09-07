import type { UserSupportedModelPricing } from '@/api/channels'

export const PER_MILLION_TOKENS = 1_000_000

const tokenPriceFormatter = new Intl.NumberFormat(undefined, {
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
})
const smallTokenPriceFormatter = new Intl.NumberFormat(undefined, {
  minimumFractionDigits: 2,
  maximumFractionDigits: 6,
})
const requestPriceFormatter = new Intl.NumberFormat(undefined, {
  minimumFractionDigits: 2,
  maximumFractionDigits: 6,
})

export function formatTokenPrice(value: number | null | undefined) {
  if (value == null) return '未配置'
  const perMillion = value * PER_MILLION_TOKENS
  const formatted = perMillion < 1
    ? smallTokenPriceFormatter.format(perMillion)
    : tokenPriceFormatter.format(perMillion)
  return '$' + formatted + '/M'
}

export function formatRequestPrice(value: number | null | undefined, billingMode?: string) {
  if (value == null) return '未配置'
  return '$' + requestPriceFormatter.format(value) + '/' + requestBillingUnit(billingMode)
}

export function requestBillingUnit(billingMode?: string) {
  if (billingMode === 'image') return '张'
  if (billingMode === 'video') return '个视频'
  return '次'
}

export function requestPriceLabel(billingMode?: string) {
  if (billingMode === 'image') return '每张价格'
  if (billingMode === 'video') return '每个视频价格'
  return '每次价格'
}

export function billingModeLabel(pricing: UserSupportedModelPricing | null) {
  switch (pricing?.billing_mode) {
    case 'image':
      return '图片计费'
    case 'per_request':
      return '按次计费'
    case 'video':
      return '视频计费'
    default:
      return 'Token 计费'
  }
}

export function isRequestBilling(pricing: UserSupportedModelPricing | null) {
  return pricing?.billing_mode === 'image' || pricing?.billing_mode === 'video' || pricing?.billing_mode === 'per_request'
}

export function fullPriceItems(pricing: UserSupportedModelPricing | null) {
  return [
    { label: '输入', value: pricing?.input_price },
    { label: '输出', value: pricing?.output_price },
    { label: '缓存写入 (5m)', value: pricing?.cache_write_price },
    { label: '缓存写入 (1h)', value: pricing?.cache_write_1h_price },
    { label: '缓存读取', value: pricing?.cache_read_price },
    { label: '图片输入', value: pricing?.image_input_price },
    { label: '图片输出', value: pricing?.image_output_price },
  ]
}

export function getPerMillionTokensPrice(value: number | null | undefined): number | null {
  if (value == null) return null
  return value * PER_MILLION_TOKENS
}

export function formatMultiplier(multiplier: number | null | undefined): string {
  if (multiplier == null) return ''
  if (multiplier === 1) return '原价'
  if (multiplier < 1) {
    const discount = (multiplier * 10).toFixed(1).replace(/\.0$/, '')
    return `${discount} 折`
  }
  return `${multiplier}x`
}

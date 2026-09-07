import { describe, expect, it } from 'vitest'
import {
  billingModeLabel,
  formatRequestPrice,
  fullPriceItems,
  requestPriceLabel,
  requestBillingUnit,
} from '../pricing'

describe('model square pricing labels', () => {
  it('keeps video pricing distinct from per-request pricing', () => {
    expect(billingModeLabel({ billing_mode: 'video' } as never)).toBe('视频计费')
    expect(requestPriceLabel('video')).toBe('每个视频价格')
    expect(requestBillingUnit('video')).toBe('个视频')
    expect(formatRequestPrice(0.125, 'video')).toContain('/个视频')
  })

  it('keeps the 5m and 1h cache write prices as separate display items', () => {
    const items = fullPriceItems({
      cache_write_price: 3e-6,
      cache_write_1h_price: 6e-6,
    } as never)

    expect(items).toEqual(expect.arrayContaining([
      { label: '缓存写入 (5m)', value: 3e-6 },
      { label: '缓存写入 (1h)', value: 6e-6 },
    ]))
  })
})

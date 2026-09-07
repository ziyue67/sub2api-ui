import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const frontendRoot = resolve(__dirname, '../../../..')
const stripeConsumers = [
  'src/views/user/StripePaymentView.vue',
  'src/views/user/StripePopupView.vue',
  'src/components/payment/StripePaymentInline.vue',
]

function readFrontendFile(path: string): string {
  return readFileSync(resolve(frontendRoot, path), 'utf8')
}

describe('Payment SDK lazy-loading contract', () => {
  it.each(stripeConsumers)('%s uses the side-effect-free Stripe loader', (path) => {
    const source = readFrontendFile(path)

    expect(source).toContain("await import('@stripe/stripe-js/pure')")
    expect(source).not.toMatch(/await import\(['"]@stripe\/stripe-js['"]\)/)
  })

  it('keeps Stripe out of the shared vendor chunk', () => {
    const viteConfig = readFrontendFile('vite.config.ts')
    const stripeRule = viteConfig.indexOf("id.includes('/@stripe/stripe-js/')")
    const miscFallback = viteConfig.indexOf("return 'vendor-misc'")

    expect(stripeRule).toBeGreaterThan(-1)
    expect(viteConfig.slice(stripeRule, miscFallback)).toContain("return 'vendor-stripe'")
    expect(stripeRule).toBeLessThan(miscFallback)
  })

  it('keeps the side-effectful Airwallex SDK out of the preloaded shared vendor chunk', () => {
    const paymentView = readFrontendFile('src/views/user/AirwallexPaymentView.vue')
    const viteConfig = readFrontendFile('vite.config.ts')
    const airwallexRule = viteConfig.indexOf("id.includes('/@airwallex/components-sdk/')")
    const miscFallback = viteConfig.indexOf("return 'vendor-misc'")

    expect(paymentView).toContain("await import('@airwallex/components-sdk')")
    expect(paymentView).not.toMatch(/^\s*import\s+.*from\s+['"]@airwallex\/components-sdk['"]/m)
    expect(airwallexRule).toBeGreaterThan(-1)
    expect(viteConfig.slice(airwallexRule, miscFallback)).toContain("return 'vendor-airwallex'")
    expect(airwallexRule).toBeLessThan(miscFallback)
    expect(viteConfig).toContain("id.includes('/@airwallex/airtracker/')")
  })
})

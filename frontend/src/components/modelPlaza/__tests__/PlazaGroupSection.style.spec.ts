import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { describe, expect, it, vi } from 'vitest'
import { compileStyle, parse } from 'vue/compiler-sfc'
import PlazaGroupSection from '../PlazaGroupSection.vue'
import type { ModelPlazaGroup } from '@/api/modelPlaza'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key })
  }
})

const filename = resolve(__dirname, '../PlazaGroupSection.vue')
const source = readFileSync(filename, 'utf8')

const badgeGroup: ModelPlazaGroup = {
  id: 1,
  name: 'UI1 Test',
  description: '',
  platform: 'openai',
  subscription_type: 'subscription',
  rate_multiplier: 1,
  peak_rate_enabled: false,
  peak_start: '',
  peak_end: '',
  peak_rate_multiplier: 1,
  is_exclusive: true,
  image_rate_independent: false,
  image_rate_multiplier: 1,
  models: []
}

function compileScopedStyle() {
  const { descriptor, errors: parseErrors } = parse(source, { filename })
  expect(parseErrors).toEqual([])

  const style = descriptor.styles.find((block) => block.scoped)
  expect(style).toBeDefined()

  const result = compileStyle({
    filename,
    id: 'data-v-plaza-group',
    source: style!.content,
    scoped: true
  })
  expect(result.errors).toEqual([])
  return result.code
}

describe('PlazaGroupSection third-version badges', () => {
  it('renders UI1 semantic badges without upstream purple and violet utilities', () => {
    setActivePinia(createPinia())
    const wrapper = mount(PlazaGroupSection, {
      props: { group: badgeGroup },
      global: {
        stubs: {
          GroupBadge: true,
          Icon: true,
          PlazaModelPricingTable: true
        }
      }
    })

    expect(wrapper.find('.scheme3-model-plaza-exclusive-badge').exists()).toBe(true)
    expect(wrapper.find('.scheme3-model-plaza-subscription-badge').exists()).toBe(true)
    expect(source).not.toMatch(/(?:bg|text|border)-(?:purple|violet)-/)
  })

  it('compiles dark theme rules onto the badges instead of the root html element', () => {
    const css = compileScopedStyle()

    expect(css).toContain('html.dark .scheme3-model-plaza-exclusive-badge')
    expect(css).toContain('html.dark .scheme3-model-plaza-subscription-badge')
    expect(css).not.toMatch(/html\.dark\s*\{/)
    expect(css).toContain('background: rgba(211, 164, 92, .12)')
    expect(css).toContain('color: #d3a45c')
    expect(css).toContain('background: rgba(143, 194, 165, .12)')
    expect(css).toContain('color: #8fc2a5')
  })

  it('uses the UI1 group badge variant instead of the shared upstream-style badge', () => {
    setActivePinia(createPinia())
    const wrapper = mount(PlazaGroupSection, {
      props: { group: badgeGroup },
      global: {
        stubs: {
          Icon: true,
          PlazaModelPricingTable: true
        }
      }
    })

    expect(wrapper.find('.scheme3-monitor-group-badge').exists()).toBe(true)
  })
})

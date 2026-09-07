import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import { compileStyle, parse } from 'vue/compiler-sfc'
import PlazaFilterBar from '../PlazaFilterBar.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key })
  }
})

const filename = resolve(__dirname, '../PlazaFilterBar.vue')
const source = readFileSync(filename, 'utf8')

function compileScopedStyle() {
  const { descriptor, errors: parseErrors } = parse(source, { filename })
  expect(parseErrors).toEqual([])

  const style = descriptor.styles.find((block) => block.scoped)
  expect(style).toBeDefined()

  const result = compileStyle({
    filename,
    id: 'data-v-plaza-filters',
    source: style!.content,
    scoped: true
  })
  expect(result.errors).toEqual([])
  return result.code
}

describe('PlazaFilterBar UI1 styles', () => {
  it('renders every platform and group filter with the UI1 chip variants', () => {
    const wrapper = mount(PlazaFilterBar, {
      props: {
        platforms: ['antigravity', 'gemini'],
        groups: [
          { id: 1, name: 'Antigravity Group', platform: 'antigravity', rate: 1 },
          { id: 2, name: 'Gemini Group', platform: 'gemini', rate: 0.9 }
        ],
        rates: [0.9, 1],
        platform: 'all',
        groupId: 'all',
        rate: 'all',
        search: ''
      },
      global: {
        stubs: {
          Icon: true,
          PlatformIcon: true
        }
      }
    })

    const buttons = wrapper.findAll('button')
    expect(buttons.length).toBeGreaterThan(0)
    expect(buttons.every((button) => /scheme3-model-plaza-chip/.test(button.attributes('class')))).toBe(true)
    expect(buttons.every((button) => !/chip-tinted/.test(button.attributes('class')))).toBe(true)
    expect(buttons.every((button) => !button.attributes('style')?.includes('--chip-accent'))).toBe(true)
  })

  it('compiles dark chip rules onto the controls without legacy platform colors', () => {
    const css = compileScopedStyle()

    expect(source).not.toContain('chip-tinted')
    expect(source).not.toContain('--chip-accent')
    expect(source).not.toContain('platformAccentColor')
    expect(css).toContain('html.dark .scheme3-model-plaza-chip')
    expect(css).toContain('html.dark .scheme3-model-plaza-chip-active')
    expect(css).not.toMatch(/html\.dark\s*\{/)
    expect(css).not.toContain('color-mix')
  })
})

import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import Scheme3ModelSquareIndex from '../Scheme3ModelSquareIndex.vue'
import type { ModelSquareModel } from '../../types'

function model(key: string, name: string, platform: string, channels = 1): ModelSquareModel {
  return {
    key,
    name,
    platform,
    entries: [],
    channels: Array.from({ length: channels }, (_, index) => ({
      key: `${key}:channel-${index}`,
      name: `channel-${index}`,
      entries: [],
      pricing: null,
    })),
  }
}

const models = [
  model('openai:gpt-4o', 'gpt-4o', 'openai', 2),
  model('anthropic:claude', 'claude', 'anthropic'),
  model('openai:gpt-mini', 'gpt-mini', 'openai'),
]

describe('Scheme3ModelSquareIndex', () => {
  beforeEach(() => {
    Element.prototype.scrollIntoView = vi.fn()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('groups models by platform and keeps every model as a safe text node', () => {
    const unsafeName = '<img src=x onerror=alert(1)>'
    const wrapper = mount(Scheme3ModelSquareIndex, {
      props: {
        models: [...models, model('custom:unsafe', unsafeName, 'custom')],
        modelValue: '',
      },
    })

    expect(wrapper.findAll('.scheme3-model-square-index-platform')).toHaveLength(3)
    expect(wrapper.findAll('.scheme3-model-square-index-item')).toHaveLength(4)
    expect(wrapper.find('.scheme3-model-square-index-list img').exists()).toBe(false)
    expect(wrapper.text()).toContain(unsafeName)
  })

  it('emits selection without depending on a rendered model card', async () => {
    const wrapper = mount(Scheme3ModelSquareIndex, {
      props: { models, modelValue: '' },
    })

    await wrapper.findAll('.scheme3-model-square-index-item')[1].trigger('click')

    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual(['openai:gpt-mini'])
    expect(wrapper.emitted('select')?.[0]).toEqual(['openai:gpt-mini'])
  })

  it('uses the grouped DOM order for arrows and Enter', async () => {
    const wrapper = mount(Scheme3ModelSquareIndex, {
      props: { models, modelValue: 'openai:gpt-4o' },
    })
    const firstButton = wrapper.findAll('.scheme3-model-square-index-item')[0]

    await firstButton.trigger('keydown', { key: 'ArrowDown' })
    await firstButton.trigger('keydown', { key: 'Enter' })

    expect(wrapper.emitted('select')?.at(-1)).toEqual(['openai:gpt-mini'])
    const buttons = wrapper.findAll<HTMLButtonElement>('.scheme3-model-square-index-item')
    expect(buttons.map((button) => button.attributes('tabindex'))).toEqual(['-1', '0', '-1'])
  })

  it('keeps a single roving Tab stop while keyboard events bubble from the buttons', async () => {
    const wrapper = mount(Scheme3ModelSquareIndex, {
      props: { models, modelValue: 'openai:gpt-4o' },
    })
    const aside = wrapper.get('aside')
    const buttons = wrapper.findAll<HTMLButtonElement>('.scheme3-model-square-index-item')

    expect(aside.attributes('tabindex')).toBeUndefined()
    expect(buttons.filter((button) => button.attributes('tabindex') === '0')).toHaveLength(1)

    await buttons[0].trigger('keydown', { key: 'End' })
    await buttons[0].trigger('keydown', { key: 'Enter' })
    expect(wrapper.emitted('select')?.at(-1)).toEqual(['anthropic:claude'])
  })

  it('synchronizes with modelValue and supports Home and End', async () => {
    const wrapper = mount(Scheme3ModelSquareIndex, {
      props: { models, modelValue: 'anthropic:claude' },
    })
    const aside = wrapper.find('aside')

    await aside.trigger('keydown', { key: 'Home' })
    await aside.trigger('keydown', { key: 'Enter' })
    expect(wrapper.emitted('select')?.at(-1)).toEqual(['openai:gpt-4o'])

    await aside.trigger('keydown', { key: 'End' })
    await aside.trigger('keydown', { key: 'Enter' })
    expect(wrapper.emitted('select')?.at(-1)).toEqual(['anthropic:claude'])
  })

  it('clamps keyboard selection when filtering removes the active model', async () => {
    const wrapper = mount(Scheme3ModelSquareIndex, {
      props: { models, modelValue: 'anthropic:claude' },
    })

    await wrapper.setProps({
      models: [models[0]],
      modelValue: '',
    })
    await wrapper.find('aside').trigger('keydown', { key: 'Enter' })

    expect(wrapper.emitted('select')?.at(-1)).toEqual(['openai:gpt-4o'])
  })
})

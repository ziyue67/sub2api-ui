import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import Scheme3V2Toggle from '../Scheme3V2Toggle.vue'

describe('Scheme3V2Toggle', () => {
  it('uses third-version classes and emits the next state', async () => {
    const wrapper = mount(Scheme3V2Toggle, { props: { modelValue: true } })
    const button = wrapper.get('button[role="switch"]')

    expect(button.attributes('aria-checked')).toBe('true')
    expect(button.classes()).toContain('scheme3-v2-toggle')
    expect(button.classes()).toContain('is-on')
    expect(wrapper.html()).not.toContain('bg-primary-600')
    expect(wrapper.html()).not.toContain('ring-0')

    await button.trigger('click')
    expect(wrapper.emitted('update:modelValue')).toEqual([[false]])
  })
})

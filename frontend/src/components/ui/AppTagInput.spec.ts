import { describe, it, expect } from 'vitest'
import { mount, type VueWrapper } from '@vue/test-utils'
import AppTagInput from './AppTagInput.vue'

function lastEmit(wrapper: VueWrapper): unknown[] | undefined {
    const events = wrapper.emitted('update:modelValue')
    return events ? events[events.length - 1] : undefined
}

describe('AppTagInput', () => {
    it('renders existing tags', () => {
        const wrapper = mount(AppTagInput, { props: { modelValue: ['vip', 'beta'] } })
        expect(wrapper.text()).toContain('vip')
        expect(wrapper.text()).toContain('beta')
    })

    it('adds a tag on Enter', async () => {
        const wrapper = mount(AppTagInput, { props: { modelValue: [] } })
        const input = wrapper.get('input')
        await input.setValue('vip')
        await input.trigger('keydown', { key: 'Enter' })
        expect(lastEmit(wrapper)).toEqual([['vip']])
    })

    it('does not add duplicate tags', async () => {
        const wrapper = mount(AppTagInput, { props: { modelValue: ['vip'] } })
        const input = wrapper.get('input')
        await input.setValue('vip')
        await input.trigger('keydown', { key: 'Enter' })
        // no update emitted because the tag already exists
        expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    })

    it('removes the last tag on Backspace when the input is empty', async () => {
        const wrapper = mount(AppTagInput, { props: { modelValue: ['vip', 'beta'] } })
        const input = wrapper.get('input')
        await input.trigger('keydown', { key: 'Backspace' })
        expect(lastEmit(wrapper)).toEqual([['vip']])
    })

    it('removes a tag when its × is clicked', async () => {
        const wrapper = mount(AppTagInput, { props: { modelValue: ['vip', 'beta'] } })
        await wrapper.get('button').trigger('click')
        expect(lastEmit(wrapper)).toEqual([['beta']])
    })
})

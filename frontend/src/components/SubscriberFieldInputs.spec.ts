import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import SubscriberFieldInputs from './SubscriberFieldInputs.vue'
import type { FieldDefinition, FieldType } from '@/stores/fields'

function def(field_type: FieldType, key: string, extra: Partial<FieldDefinition> = {}): FieldDefinition {
    return { id: key, project_id: 'p', key, label: key, field_type, options: extra.options ?? [], required: extra.required ?? false, created_at: '', ...extra }
}

describe('SubscriberFieldInputs', () => {
    it('renders a control per definition type', () => {
        const definitions = [
            def('string', 'name'),
            def('number', 'age'),
            def('date', 'dob'),
            def('enum', 'plan', { options: ['free', 'pro'] }),
        ]
        const wrapper = mount(SubscriberFieldInputs, { props: { definitions, modelValue: {} } })
        expect(wrapper.find('input[type="text"]').exists()).toBe(true)
        expect(wrapper.find('input[type="number"]').exists()).toBe(true)
        expect(wrapper.find('input[type="date"]').exists()).toBe(true)
        const options = wrapper.find('select').findAll('option').map(o => o.text())
        expect(options).toContain('free')
        expect(options).toContain('pro')
    })

    it('emits the typed value on input', async () => {
        const wrapper = mount(SubscriberFieldInputs, {
            props: { definitions: [def('string', 'name')], modelValue: {} },
        })
        await wrapper.get('input[type="text"]').setValue('Ada')
        const events = wrapper.emitted('update:modelValue')
        expect(events?.[events.length - 1]).toEqual([{ name: 'Ada' }])
    })

    it('shows a per-field error message', () => {
        const wrapper = mount(SubscriberFieldInputs, {
            props: {
                definitions: [def('number', 'age', { label: 'Age' })],
                modelValue: {},
                errors: { age: 'Age must be a number' },
            },
        })
        expect(wrapper.text()).toContain('Age must be a number')
    })

    it('renders nothing when there are no definitions', () => {
        const wrapper = mount(SubscriberFieldInputs, { props: { definitions: [], modelValue: {} } })
        expect(wrapper.find('input').exists()).toBe(false)
    })
})

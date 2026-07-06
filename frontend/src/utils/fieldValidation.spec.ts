import { describe, it, expect } from 'vitest'
import { validateFieldValues } from './fieldValidation'
import type { FieldDefinition, FieldType } from '@/stores/fields'

function def(field_type: FieldType, extra: Partial<FieldDefinition> = {}): FieldDefinition {
    return {
        id: 'id',
        project_id: 'p',
        key: extra.key ?? field_type,
        label: extra.label ?? field_type,
        field_type,
        options: extra.options ?? [],
        required: extra.required ?? false,
        created_at: '',
        ...extra,
    }
}

describe('validateFieldValues', () => {
    it('flags a missing required field', () => {
        const errors = validateFieldValues([def('string', { key: 'name', label: 'Name', required: true })], {})
        expect(errors.name).toBe('Name is required')
    })

    it('accepts a present required field', () => {
        const errors = validateFieldValues([def('string', { key: 'name', required: true })], { name: 'Ada' })
        expect(errors).toEqual({})
    })

    it('skips empty optional fields', () => {
        const errors = validateFieldValues([def('number', { key: 'age' })], { age: '' })
        expect(errors).toEqual({})
    })

    it('rejects a non-numeric number field', () => {
        const errors = validateFieldValues([def('number', { key: 'age', label: 'Age' })], { age: 'abc' })
        expect(errors.age).toContain('must be a number')
    })

    it('accepts a numeric number field', () => {
        expect(validateFieldValues([def('number', { key: 'age' })], { age: 42 })).toEqual({})
        expect(validateFieldValues([def('number', { key: 'age' })], { age: '42' })).toEqual({})
    })

    it('rejects an invalid date', () => {
        const errors = validateFieldValues([def('date', { key: 'dob', label: 'DOB' })], { dob: 'not-a-date' })
        expect(errors.dob).toContain('valid date')
    })

    it('accepts a valid date', () => {
        expect(validateFieldValues([def('date', { key: 'dob' })], { dob: '1990-05-12' })).toEqual({})
    })

    it('rejects an enum value outside its options', () => {
        const d = def('enum', { key: 'plan', label: 'Plan', options: ['free', 'pro'] })
        const errors = validateFieldValues([d], { plan: 'team' })
        expect(errors.plan).toContain('allowed options')
    })

    it('accepts an enum value within its options', () => {
        const d = def('enum', { key: 'plan', options: ['free', 'pro'] })
        expect(validateFieldValues([d], { plan: 'pro' })).toEqual({})
    })

    it('collects errors across multiple fields', () => {
        const defs = [
            def('string', { key: 'name', label: 'Name', required: true }),
            def('number', { key: 'age', label: 'Age' }),
        ]
        const errors = validateFieldValues(defs, { age: 'x' })
        expect(Object.keys(errors).sort()).toEqual(['age', 'name'])
    })
})

import type { FieldDefinition } from '@/stores/fields'

// Validates subscriber custom-field values against their definitions, mirroring
// the server-side rules so the user gets immediate, per-field feedback instead
// of a generic backend error. Returns a map of field key -> error message.
export function validateFieldValues(
    defs: FieldDefinition[],
    values: Record<string, any>,
): Record<string, string> {
    const errors: Record<string, string> = {}
    for (const def of defs) {
        const value = values[def.key]
        const empty = value === undefined || value === null || value === ''

        if (def.required && empty) {
            errors[def.key] = `${def.label} is required`
            continue
        }
        if (empty) continue

        if (def.field_type === 'number' && Number.isNaN(Number(value))) {
            errors[def.key] = `${def.label} must be a number`
        } else if (def.field_type === 'date' && Number.isNaN(Date.parse(String(value)))) {
            errors[def.key] = `${def.label} must be a valid date`
        } else if (def.field_type === 'enum' && !def.options.includes(String(value))) {
            errors[def.key] = `${def.label} must be one of its allowed options`
        }
    }
    return errors
}

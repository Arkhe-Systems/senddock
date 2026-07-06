import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/api/client'

export type FieldType = 'string' | 'number' | 'date' | 'boolean' | 'enum'

export interface FieldDefinition {
    id: string
    project_id: string
    key: string
    label: string
    field_type: FieldType
    options: string[]
    required: boolean
    created_at: string
}

export interface FieldDefinitionInput {
    key: string
    label: string
    field_type: FieldType
    options: string[]
    required: boolean
}

export const useFieldStore = defineStore('fields', () => {
    const byProject = ref<Record<string, FieldDefinition[]>>({})

    function fields(projectId: string): FieldDefinition[] {
        return byProject.value[projectId] || []
    }

    async function fetchFields(projectId: string): Promise<FieldDefinition[]> {
        try {
            const res = await api<FieldDefinition[]>(`/projects/${projectId}/fields`)
            byProject.value[projectId] = res || []
        } catch {
            byProject.value[projectId] = []
        }
        return byProject.value[projectId]!
    }

    async function createField(projectId: string, input: FieldDefinitionInput): Promise<FieldDefinition> {
        const created = await api<FieldDefinition>(`/projects/${projectId}/fields`, {
            method: 'POST',
            body: input,
        })
        await fetchFields(projectId)
        return created
    }

    async function updateField(projectId: string, fieldId: string, input: Omit<FieldDefinitionInput, 'key' | 'field_type'>): Promise<void> {
        await api(`/projects/${projectId}/fields/${fieldId}`, {
            method: 'PATCH',
            body: input,
        })
        await fetchFields(projectId)
    }

    async function deleteField(projectId: string, fieldId: string): Promise<void> {
        await api(`/projects/${projectId}/fields/${fieldId}`, { method: 'DELETE' })
        await fetchFields(projectId)
    }

    return { byProject, fields, fetchFields, createField, updateField, deleteField }
})

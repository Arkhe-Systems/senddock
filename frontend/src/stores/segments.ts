import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/api/client'

export type SegmentMatch = 'all' | 'any'

export interface SegmentRule {
    field: string
    op: string
    value: any
}

export interface SegmentPredicate {
    match: SegmentMatch
    rules: SegmentRule[]
}

export interface Segment {
    id: string
    project_id: string
    name: string
    predicate: SegmentPredicate
    created_at: string
    updated_at: string
}

export const useSegmentStore = defineStore('segments', () => {
    const byProject = ref<Record<string, Segment[]>>({})

    function segments(projectId: string): Segment[] {
        return byProject.value[projectId] || []
    }

    async function fetchSegments(projectId: string): Promise<Segment[]> {
        try {
            const res = await api<Segment[]>(`/projects/${projectId}/segments`)
            byProject.value[projectId] = res || []
        } catch {
            byProject.value[projectId] = []
        }
        return byProject.value[projectId]!
    }

    async function createSegment(projectId: string, name: string, predicate: SegmentPredicate): Promise<Segment> {
        const created = await api<Segment>(`/projects/${projectId}/segments`, {
            method: 'POST',
            body: { name, predicate },
        })
        await fetchSegments(projectId)
        return created
    }

    async function updateSegment(projectId: string, id: string, name: string, predicate: SegmentPredicate): Promise<void> {
        await api(`/projects/${projectId}/segments/${id}`, {
            method: 'PATCH',
            body: { name, predicate },
        })
        await fetchSegments(projectId)
    }

    async function deleteSegment(projectId: string, id: string): Promise<void> {
        await api(`/projects/${projectId}/segments/${id}`, { method: 'DELETE' })
        await fetchSegments(projectId)
    }

    async function previewSegment(projectId: string, predicate: SegmentPredicate): Promise<number> {
        const res = await api<{ count: number }>(`/projects/${projectId}/segments/preview`, {
            method: 'POST',
            body: { predicate },
        })
        return res.count
    }

    return { byProject, segments, fetchSegments, createSegment, updateSegment, deleteSegment, previewSegment }
})

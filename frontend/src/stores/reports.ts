import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/api/client'
import type { SegmentPredicate } from '@/stores/segments'

export type DimKind = 'column' | 'time' | 'tag' | 'custom'
export interface SchemaDim { key: string; label: string; kind: DimKind }
export interface SchemaMeasure { key: string; label: string }
export interface SchemaDataset {
    key: string; label: string
    dimensions: SchemaDim[]; measures: SchemaMeasure[]
}
export interface ReportSchema {
    datasets: SchemaDataset[]
    viz_types: string[]
    granularities: string[]
}

export interface ReportWindow { from: string; to: string; granularity: string }
export interface ReportConfig {
    dataset: string
    measure: string
    dimensions: string[]
    filters?: SegmentPredicate
    window?: ReportWindow
    viz: string
}

export interface RunRow { label: string; value: number }
export interface PivotRow { label: string; cells: Record<string, number> }
export interface PivotResult { columns: string[]; rows: PivotRow[] }
export interface RunResult {
    dataset: string; measure: string; dimensions: string[]
    rows?: RunRow[]
    pivot?: PivotResult
}

export interface SavedReport {
    id: string; project_id: string; name: string
    config: ReportConfig; created_at: string; updated_at: string
}

export const useReportsStore = defineStore('reports', () => {
    const base = (projectId: string) => `/projects/${projectId}/reports`
    const byProject = ref<Record<string, SavedReport[]>>({})

    function schema(projectId: string) {
        return api<ReportSchema>(`${base(projectId)}/schema`)
    }
    function run(projectId: string, config: ReportConfig) {
        return api<RunResult>(`${base(projectId)}/run`, { method: 'POST', body: config })
    }

    function reports(projectId: string): SavedReport[] {
        return byProject.value[projectId] ?? []
    }
    async function fetchReports(projectId: string) {
        try {
            const res = await api<{ reports: SavedReport[] }>(base(projectId))
            byProject.value[projectId] = res.reports
        } catch {
            byProject.value[projectId] = []
        }
    }
    async function createReport(projectId: string, name: string, config: ReportConfig) {
        const r = await api<SavedReport>(base(projectId), { method: 'POST', body: { name, config } })
        await fetchReports(projectId)
        return r
    }
    async function updateReport(projectId: string, id: string, name: string, config: ReportConfig) {
        const r = await api<SavedReport>(`${base(projectId)}/${id}`, { method: 'PATCH', body: { name, config } })
        await fetchReports(projectId)
        return r
    }
    async function deleteReport(projectId: string, id: string) {
        await api(`${base(projectId)}/${id}`, { method: 'DELETE' })
        await fetchReports(projectId)
    }

    return { schema, run, reports, fetchReports, createReport, updateReport, deleteReport }
})

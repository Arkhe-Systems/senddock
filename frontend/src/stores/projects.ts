import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/api/client'

export interface Project {
    id: string
    name: string
    description: string | null
    from_name: string | null
    from_email: string | null
    smtp_host: string | null
    smtp_port: number | null
    smtp_user: string | null
    created_at: string
    updated_at: string
}

export const useProjectStore = defineStore('projects', () => {

    const projects = ref<Project[]>([])

    async function fetchProjects(workspaceId?: string | null) {
        try {
            const url = workspaceId
                ? `/workspaces/${workspaceId}/projects`
                : '/projects'
            const response = await api<Project[]>(url)
            projects.value = response || []
        } catch {
            projects.value = []
        }
    }

    async function deleteProject(id: string) {
        await api(`/projects/${id}`, { method: 'DELETE' })
        await fetchProjects()
    }

    return { projects, fetchProjects, deleteProject }
})

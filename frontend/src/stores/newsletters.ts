import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/api/client'

export interface Newsletter {
    id: string
    project_id: string
    name: string
    description: string
    active_count: number
    created_at: string
    updated_at: string
}

export interface SubscriberNewsletter {
    id: string
    name: string
    unsubscribed_at: string | null
}

export const useNewsletterStore = defineStore('newsletters', () => {
    const byProject = ref<Record<string, Newsletter[]>>({})

    function newsletters(projectId: string): Newsletter[] {
        return byProject.value[projectId] || []
    }

    async function fetchNewsletters(projectId: string): Promise<Newsletter[]> {
        try {
            const res = await api<Newsletter[]>(`/projects/${projectId}/newsletters`)
            byProject.value[projectId] = res || []
        } catch {
            byProject.value[projectId] = []
        }
        return byProject.value[projectId]!
    }

    async function createNewsletter(projectId: string, name: string, description: string): Promise<Newsletter> {
        const created = await api<Newsletter>(`/projects/${projectId}/newsletters`, {
            method: 'POST',
            body: { name, description },
        })
        await fetchNewsletters(projectId)
        return created
    }

    async function updateNewsletter(projectId: string, id: string, name: string, description: string): Promise<void> {
        await api(`/projects/${projectId}/newsletters/${id}`, {
            method: 'PATCH',
            body: { name, description },
        })
        await fetchNewsletters(projectId)
    }

    async function deleteNewsletter(projectId: string, id: string): Promise<void> {
        await api(`/projects/${projectId}/newsletters/${id}`, { method: 'DELETE' })
        await fetchNewsletters(projectId)
    }

    async function fetchSubscriberNewsletters(projectId: string, subscriberId: string): Promise<SubscriberNewsletter[]> {
        return (await api<SubscriberNewsletter[]>(`/projects/${projectId}/subscribers/${subscriberId}/newsletters`)) || []
    }

    async function setSubscriberNewsletters(projectId: string, subscriberId: string, newsletterIds: string[]): Promise<void> {
        await api(`/projects/${projectId}/subscribers/${subscriberId}/newsletters`, {
            method: 'PUT',
            body: { newsletter_ids: newsletterIds },
        })
    }

    return {
        byProject,
        newsletters,
        fetchNewsletters,
        createNewsletter,
        updateNewsletter,
        deleteNewsletter,
        fetchSubscriberNewsletters,
        setSubscriberNewsletters,
    }
})

import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { api } from '@/api/client'

export interface Workspace {
    id: string
    name: string
    created_by: string
    created_at: string
    updated_at: string
    role?: string
}

export interface WorkspaceMember {
    user_id: string
    email: string
    name: string
    role: 'owner' | 'member'
    joined_at: string
}

const ACTIVE_KEY = 'senddock.activeWorkspaceId'

export const useWorkspaceStore = defineStore('workspaces', () => {
    const workspaces = ref<Workspace[]>([])
    const activeId = ref<string | null>(localStorage.getItem(ACTIVE_KEY))
    const loading = ref(false)

    const active = computed<Workspace | null>(() =>
        workspaces.value.find(w => w.id === activeId.value) || workspaces.value[0] || null
    )

    function setActive(id: string) {
        activeId.value = id
        localStorage.setItem(ACTIVE_KEY, id)
    }

    async function fetch() {
        loading.value = true
        try {
            const res = await api<{ workspaces: Workspace[] }>('/workspaces')
            workspaces.value = res.workspaces || []
            if (!activeId.value || !workspaces.value.some(w => w.id === activeId.value)) {
                if (workspaces.value[0]) setActive(workspaces.value[0].id)
            }
        } finally {
            loading.value = false
        }
    }

    async function create(name: string): Promise<Workspace> {
        const ws = await api<Workspace>('/workspaces', { method: 'POST', body: { name } })
        workspaces.value = [...workspaces.value, ws]
        setActive(ws.id)
        return ws
    }

    async function rename(id: string, name: string) {
        const ws = await api<Workspace>(`/workspaces/${id}`, { method: 'PATCH', body: { name } })
        workspaces.value = workspaces.value.map(w => (w.id === id ? { ...w, ...ws } : w))
    }

    async function remove(id: string) {
        await api(`/workspaces/${id}`, { method: 'DELETE' })
        workspaces.value = workspaces.value.filter(w => w.id !== id)
        if (activeId.value === id) {
            const next = workspaces.value[0]?.id || null
            activeId.value = next
            if (next) localStorage.setItem(ACTIVE_KEY, next)
            else localStorage.removeItem(ACTIVE_KEY)
        }
    }

    async function listMembers(id: string): Promise<WorkspaceMember[]> {
        const res = await api<{ members: WorkspaceMember[] }>(`/workspaces/${id}/members`)
        return res.members || []
    }

    async function addMember(id: string, email: string, role: 'owner' | 'member' = 'member') {
        return api<WorkspaceMember>(`/workspaces/${id}/members`, {
            method: 'POST',
            body: { email, role },
        })
    }

    async function updateMemberRole(id: string, userId: string, role: 'owner' | 'member') {
        await api(`/workspaces/${id}/members/${userId}`, { method: 'PATCH', body: { role } })
    }

    async function removeMember(id: string, userId: string) {
        await api(`/workspaces/${id}/members/${userId}`, { method: 'DELETE' })
    }

    function reset() {
        workspaces.value = []
        activeId.value = null
        localStorage.removeItem(ACTIVE_KEY)
    }

    return {
        workspaces,
        activeId,
        active,
        loading,
        fetch,
        create,
        rename,
        remove,
        setActive,
        listMembers,
        addMember,
        updateMemberRole,
        removeMember,
        reset,
    }
})

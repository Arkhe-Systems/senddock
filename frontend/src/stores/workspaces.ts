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

export type WorkspaceRole = 'owner' | 'admin' | 'developer' | 'member' | 'viewer'

export interface WorkspaceMember {
    user_id: string
    email: string
    name: string
    role: WorkspaceRole
    joined_at: string
}

export const ROLE_LABEL: Record<WorkspaceRole, string> = {
    owner: 'Owner',
    admin: 'Admin',
    developer: 'Developer',
    member: 'Member',
    viewer: 'Viewer',
}

export const ROLE_DESCRIPTION: Record<WorkspaceRole, string> = {
    owner: 'Full access — can manage members, delete the workspace, and do everything an admin can.',
    admin: 'Everything project-related: settings, templates, subscribers, sends, broadcasts, API keys.',
    developer: 'Send transactional email only (`/send`). Read-only on templates, subscribers, logs.',
    member: 'Legacy role kept for backward compatibility. Same access as admin.',
    viewer: 'Read-only — view templates, subscribers, logs, analytics. Cannot send or modify anything.',
}

export const ASSIGNABLE_ROLES: WorkspaceRole[] = ['owner', 'admin', 'developer', 'viewer']

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

    async function addMember(id: string, email: string, role: WorkspaceRole = 'member') {
        return api<WorkspaceMember>(`/workspaces/${id}/members`, {
            method: 'POST',
            body: { email, role },
        })
    }

    async function createUser(id: string, payload: { email: string; name: string; password: string; role: WorkspaceRole }) {
        return api<WorkspaceMember>(`/workspaces/${id}/users`, {
            method: 'POST',
            body: payload,
        })
    }

    async function updateMemberRole(id: string, userId: string, role: WorkspaceRole) {
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
        createUser,
        updateMemberRole,
        removeMember,
        reset,
    }
})

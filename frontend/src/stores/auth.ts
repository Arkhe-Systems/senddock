import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api, ApiError } from '@/api/client'

interface MessageResponse {
    message: string
}

export const useAuthStore = defineStore('auth', () => {

    const isAuthenticated = ref(false)
    const sessionExpired = ref(false)
    const userId = ref<string | null>(null)

    async function checkAuth() {
        const wasAuthenticated = isAuthenticated.value
        try {
            const me = await api<{ user_id: string }>('/me', { silent: true })
            isAuthenticated.value = true
            sessionExpired.value = false
            userId.value = me.user_id
        } catch (e) {
            if (e instanceof ApiError && (e.status === 0 || e.status === 429 || e.status >= 500)) {
                return
            }
            isAuthenticated.value = false
            userId.value = null
            if (wasAuthenticated) {
                sessionExpired.value = true
            }
        }
    }

    async function refreshSession(): Promise<boolean> {
        try {
            await api<MessageResponse>('/auth/refresh', { method: 'POST' })
            isAuthenticated.value = true
            sessionExpired.value = false
            return true
        } catch {
            isAuthenticated.value = false
            sessionExpired.value = true
            return false
        }
    }

    async function login(email: string, password: string) {
        await api<MessageResponse>('/auth/login', {
            method: 'POST',
            body: { email, password },
        })
        isAuthenticated.value = true
        sessionExpired.value = false
        api('/license/status').catch(() => {})
    }

    async function register(email: string, password: string, name: string) {
        await api<MessageResponse>('/auth/register', {
            method: 'POST',
            body: { email, password, name },
        })
        isAuthenticated.value = true
        sessionExpired.value = false
    }

    async function logout() {
        try {
            await api<MessageResponse>('/auth/logout', { method: 'POST' })
        } catch {
        }
        isAuthenticated.value = false
        sessionExpired.value = false
        userId.value = null
    }

    return { isAuthenticated, sessionExpired, userId, login, register, logout, checkAuth, refreshSession }
})

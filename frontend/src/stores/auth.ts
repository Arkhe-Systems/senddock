import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/api/client'

interface MessageResponse {
    message: string
}

export const useAuthStore = defineStore('auth', () => {

    const isAuthenticated = ref(false)
    const sessionExpired = ref(false)

    async function checkAuth() {
        const wasAuthenticated = isAuthenticated.value
        try {
            await api<any>('/me')
            isAuthenticated.value = true
            sessionExpired.value = false
        } catch {
            isAuthenticated.value = false
            if (wasAuthenticated) {
                // Was logged in, now the session is gone
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
            // Ignore errors on logout
        }
        isAuthenticated.value = false
        sessionExpired.value = false
    }

    return { isAuthenticated, sessionExpired, login, register, logout, checkAuth, refreshSession }
})

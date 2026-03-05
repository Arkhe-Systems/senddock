import {defineStore} from 'pinia'
import {ref} from 'vue'
import {api} from '@/api/client'

interface MessageResponse {
    message: string
}

export const useAuthStore = defineStore('auth', () => {

    const isAuthenticated = ref(false);

    async function checkAuth() {
        try {
            await api<any>('/me')
            isAuthenticated.value = true
        }catch {
            isAuthenticated.value = false
        }
    }

    async function login(email: string, password: string) {
        const response = await api<MessageResponse>('/auth/login', {
            method: 'POST',
            body: {email, password},
        })
        isAuthenticated.value = true
    }

    async function register(email: string, password: string, name: string) {
        const response = await api<MessageResponse>('/auth/register', {
            method: 'POST',
            body: {email, password, name},
        })
        isAuthenticated.value = true
    }

    async function logout() {
        await api<MessageResponse>('/auth/logout', {method: 'POST'})
        isAuthenticated.value = false
    }

    return { isAuthenticated, login, register, logout, checkAuth }
})
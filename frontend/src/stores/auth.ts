import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api, ApiError } from '@/api/client'

interface MessageResponse {
    message: string
}

interface MeResponse {
    user_id: string
    email: string
    name: string
    plan: string
    created_at: string
    totp_enabled: boolean
}

export const useAuthStore = defineStore('auth', () => {

    const isAuthenticated = ref(false)
    const sessionExpired = ref(false)
    const userId = ref<string | null>(null)
    const email = ref<string | null>(null)
    const name = ref<string | null>(null)
    const plan = ref<string | null>(null)
    const createdAt = ref<string | null>(null)
    const totpEnabled = ref(false)

    async function checkAuth() {
        const wasAuthenticated = isAuthenticated.value
        try {
            const me = await api<MeResponse>('/me', { silent: true })
            isAuthenticated.value = true
            sessionExpired.value = false
            userId.value = me.user_id
            email.value = me.email
            name.value = me.name
            plan.value = me.plan
            createdAt.value = me.created_at
            totpEnabled.value = me.totp_enabled
        } catch (e) {
            if (e instanceof ApiError && (e.status === 0 || e.status === 429 || e.status >= 500)) {
                return
            }
            isAuthenticated.value = false
            userId.value = null
            email.value = null
            name.value = null
            plan.value = null
            createdAt.value = null
            totpEnabled.value = false
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

    interface LoginResponse {
        requires_2fa?: boolean
        two_factor_token?: string
        message?: string
    }

    async function login(email: string, password: string): Promise<{ requires_2fa: boolean; two_factor_token?: string }> {
        const res = await api<LoginResponse>('/auth/login', {
            method: 'POST',
            body: { email, password },
        })
        if (res.requires_2fa) {
            return { requires_2fa: true, two_factor_token: res.two_factor_token }
        }
        isAuthenticated.value = true
        sessionExpired.value = false
        api('/license/status').catch(() => {})
        return { requires_2fa: false }
    }

    async function verifyTwoFactor(twoFactorToken: string, code: string) {
        await api<MessageResponse>('/auth/2fa', {
            method: 'POST',
            body: { two_factor_token: twoFactorToken, code },
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

    async function signup(email: string, password: string, name: string): Promise<string> {
        const res = await api<MessageResponse>('/auth/signup', {
            method: 'POST',
            body: { email, password, name },
        })
        return res.message
    }

    async function verifyEmail(token: string) {
        await api<MessageResponse>('/auth/verify', {
            method: 'POST',
            body: { token },
        })
        isAuthenticated.value = true
        sessionExpired.value = false
        api('/license/status').catch(() => {})
    }

    async function resendVerification(email: string): Promise<string> {
        const res = await api<MessageResponse>('/auth/resend-verification', {
            method: 'POST',
            body: { email },
        })
        return res.message
    }

    async function logout() {
        try {
            await api<MessageResponse>('/auth/logout', { method: 'POST' })
        } catch {
        }
        isAuthenticated.value = false
        sessionExpired.value = false
        userId.value = null
        email.value = null
        name.value = null
        plan.value = null
        createdAt.value = null
        totpEnabled.value = false
    }

    return { isAuthenticated, sessionExpired, userId, email, name, plan, createdAt, totpEnabled, login, register, signup, verifyEmail, resendVerification, logout, checkAuth, refreshSession, verifyTwoFactor }
})

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent } from 'vue'

const push = vi.fn()
const currentRoute = { value: { name: 'dashboard' } }

vi.mock('vue-router', () => ({
    useRouter: () => ({ push, currentRoute }),
}))

vi.mock('@/api/client', () => ({
    api: vi.fn(() => Promise.resolve({})),
    ApiError: class extends Error {
        status: number
        constructor(status: number, message: string) {
            super(message)
            this.status = status
        }
    },
}))

import { useSessionWatch } from './useSessionWatch'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'

const MINUTE = 60 * 1000

function installLocalStorage() {
    const store = new Map<string, string>()
    Object.defineProperty(window, 'localStorage', {
        configurable: true,
        value: {
            getItem: (k: string) => store.get(k) ?? null,
            setItem: (k: string, v: string) => void store.set(k, String(v)),
            removeItem: (k: string) => void store.delete(k),
            clear: () => store.clear(),
        },
    })
}

const Host = defineComponent({
    setup() {
        useSessionWatch()
        return () => null
    },
})

function startWatching(mode: 'cloud' | 'self-hosted', idleMinutes?: number) {
    const wrapper = mount(Host)
    const auth = useAuthStore()
    const app = useAppStore()
    app.deploymentMode = mode
    if (idleMinutes !== undefined) app.sessionIdleTimeoutMinutes = idleMinutes
    auth.isAuthenticated = true
    return { wrapper, auth }
}

describe('useSessionWatch', () => {
    beforeEach(() => {
        setActivePinia(createPinia())
        installLocalStorage()
        push.mockClear()
        currentRoute.value = { name: 'dashboard' }
        vi.useFakeTimers()
    })

    afterEach(() => {
        vi.useRealTimers()
    })

    it('signs the user out after 20 idle minutes on cloud', async () => {
        const { auth } = startWatching('cloud')

        await vi.advanceTimersByTimeAsync(20 * MINUTE + 1000)

        expect(auth.isAuthenticated).toBe(false)
        expect(push).toHaveBeenCalledWith({ name: 'login', query: { reason: 'idle_timeout' } })
    })

    it('keeps the session while the user is actually interacting', async () => {
        const { auth } = startWatching('cloud')

        for (let i = 0; i < 5; i++) {
            await vi.advanceTimersByTimeAsync(15 * MINUTE)
            document.dispatchEvent(new Event('mousedown'))
        }

        expect(auth.isAuthenticated).toBe(true)
        expect(push).not.toHaveBeenCalled()
    })

    it('does not apply the 20-minute cloud window to self-hosted', async () => {
        const { auth } = startWatching('self-hosted')

        await vi.advanceTimersByTimeAsync(25 * MINUTE)

        expect(auth.isAuthenticated).toBe(true)
        expect(push).not.toHaveBeenCalled()
    })

    it('honours the configured self-hosted timeout', async () => {
        const { auth } = startWatching('self-hosted', 30)

        await vi.advanceTimersByTimeAsync(29 * MINUTE)
        expect(auth.isAuthenticated).toBe(true)

        await vi.advanceTimersByTimeAsync(2 * MINUTE)
        expect(auth.isAuthenticated).toBe(false)
        expect(push).toHaveBeenCalledWith({ name: 'login', query: { reason: 'idle_timeout' } })
    })

    it('ignores the self-hosted setting on cloud', async () => {
        const { auth } = startWatching('cloud', 480)

        await vi.advanceTimersByTimeAsync(21 * MINUTE)

        expect(auth.isAuthenticated).toBe(false)
    })

    it('ends the session when only background polling happened', async () => {
        const { auth } = startWatching('cloud')

        await vi.advanceTimersByTimeAsync(21 * MINUTE)

        expect(auth.isAuthenticated).toBe(false)
        expect(push).toHaveBeenCalledTimes(1)
    })

    it('follows another tab that already ended the session', async () => {
        const { auth } = startWatching('cloud')

        window.dispatchEvent(
            new StorageEvent('storage', { key: 'senddock:session_ended', newValue: String(Date.now()) }),
        )
        await vi.advanceTimersByTimeAsync(0)

        expect(auth.isAuthenticated).toBe(false)
        expect(push).toHaveBeenCalledWith({ name: 'login', query: { reason: 'idle_timeout' } })
    })
})

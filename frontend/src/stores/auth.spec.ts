import { describe, it, expect, vi, beforeEach } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

const apiMock = vi.fn()
vi.mock('@/api/client', () => ({
    api: (...args: unknown[]) => apiMock(...args),
    ApiError: class extends Error {
        status: number
        constructor(status: number, message: string) {
            super(message)
            this.status = status
        }
    },
}))

import { useAuthStore } from './auth'

describe('auth store — device confirmation flow', () => {
    beforeEach(() => {
        setActivePinia(createPinia())
        apiMock.mockReset()
    })

    it('verifyEmail marks the session authenticated when not pending', async () => {
        apiMock.mockResolvedValueOnce({ message: 'verified', pending: false })
        apiMock.mockResolvedValueOnce({})

        const auth = useAuthStore()
        const result = await auth.verifyEmail('tok')

        expect(result).toEqual({ pending: false })
        expect(auth.isAuthenticated).toBe(true)
        expect(auth.sessionExpired).toBe(false)
        expect(apiMock).toHaveBeenCalledWith('/auth/verify', { method: 'POST', body: { token: 'tok' } })
        expect(apiMock).toHaveBeenCalledWith('/license/status')
    })

    it('verifyEmail stays unauthenticated when pending approval', async () => {
        apiMock.mockResolvedValueOnce({ message: 'check your other device', pending: true })

        const auth = useAuthStore()
        const result = await auth.verifyEmail('tok')

        expect(result).toEqual({ pending: true })
        expect(auth.isAuthenticated).toBe(false)
        expect(apiMock).not.toHaveBeenCalledWith('/license/status')
    })

    it('pendingStatus returns the approval flag from the server', async () => {
        apiMock.mockResolvedValueOnce({ approved: true })

        const auth = useAuthStore()
        await expect(auth.pendingStatus()).resolves.toBe(true)
        expect(apiMock).toHaveBeenCalledWith('/auth/pending-login/status')
    })

    it('pendingComplete authenticates the session', async () => {
        apiMock.mockResolvedValueOnce({ message: 'approved' })
        apiMock.mockResolvedValueOnce({})

        const auth = useAuthStore()
        await auth.pendingComplete()

        expect(auth.isAuthenticated).toBe(true)
        expect(auth.sessionExpired).toBe(false)
        expect(apiMock).toHaveBeenCalledWith('/auth/pending-login/complete', { method: 'POST' })
    })
})

const API_URL = import.meta.env.VITE_API_URL || '/api/v1'

// Default per-request timeout. Without this, fetch() never times out and a
// hung backend (SMTP test against an unreachable host, slow tunnel, dead
// connection) leaves the calling component stuck in its loading state
// forever — which manifests as "buttons stop working" because the UI
// still thinks something is in flight.
const DEFAULT_TIMEOUT_MS = 30_000

export function getApiBase(): string {
    return API_URL
}

interface ApiOptions {
    method?: string
    body?: unknown
    silent?: boolean
    timeoutMs?: number
    _retry?: boolean
}

let onSessionExpired: (() => void) | null = null
let onRateLimited: (() => void) | null = null

export function setSessionExpiredHandler(fn: () => void) {
    onSessionExpired = fn
}

export function setRateLimitedHandler(fn: () => void) {
    onRateLimited = fn
}

export class ApiError extends Error {
    status: number
    code?: string
    constructor(status: number, message: string, code?: string) {
        super(message)
        this.status = status
        this.code = code
        this.name = 'ApiError'
    }
}

export async function api<T>(endpoint: string, options: ApiOptions = {}): Promise<T> {
    const { method = 'GET', body, silent = false, timeoutMs = DEFAULT_TIMEOUT_MS, _retry = false } = options

    const headers: Record<string, string> = {
        'Content-Type': 'application/json',
    }

    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), timeoutMs)

    let response: Response
    try {
        response = await fetch(`${API_URL}${endpoint}`, {
            method,
            headers,
            body: body ? JSON.stringify(body) : undefined,
            credentials: 'include',
            signal: controller.signal,
        })
    } catch (e: any) {
        clearTimeout(timeoutId)
        if (e?.name === 'AbortError') {
            throw new ApiError(0, `request timed out after ${timeoutMs / 1000}s`)
        }
        throw new ApiError(0, 'network error')
    }
    clearTimeout(timeoutId)

    if (response.status === 401 && !_retry && !endpoint.includes('/auth/')) {
        const refreshController = new AbortController()
        const refreshTimeout = setTimeout(() => refreshController.abort(), 10_000)
        let refreshRes: Response
        try {
            refreshRes = await fetch(`${API_URL}/auth/refresh`, {
                method: 'POST',
                credentials: 'include',
                signal: refreshController.signal,
            })
        } catch {
            clearTimeout(refreshTimeout)
            if (!silent && onSessionExpired) onSessionExpired()
            throw new ApiError(401, 'session_expired')
        }
        clearTimeout(refreshTimeout)

        if (refreshRes.ok) {
            return api<T>(endpoint, { ...options, _retry: true })
        }

        if (!silent && onSessionExpired) onSessionExpired()
        throw new ApiError(401, 'session_expired')
    }

    if (response.status === 429) {
        if (!silent && onRateLimited) onRateLimited()
        throw new ApiError(429, 'rate limit exceeded')
    }

    if (!response.ok) {
        let message = 'something went wrong'
        let code: string | undefined
        try {
            const error = await response.json()
            if (error?.error) message = error.error
            if (error?.code) code = error.code
        } catch {}
        throw new ApiError(response.status, message, code)
    }

    if (response.status === 204) {
        return {} as T
    }

    return response.json()
}

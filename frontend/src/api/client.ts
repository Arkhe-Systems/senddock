const API_URL = import.meta.env.VITE_API_URL || '/api/v1'

export function getApiBase(): string {
    return API_URL
}

interface ApiOptions {
    method?: string
    body?: unknown
    silent?: boolean
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
    const { method = 'GET', body, silent = false, _retry = false } = options

    const headers: Record<string, string> = {
        'Content-Type': 'application/json',
    }

    let response: Response
    try {
        response = await fetch(`${API_URL}${endpoint}`, {
            method,
            headers,
            body: body ? JSON.stringify(body) : undefined,
            credentials: 'include',
        })
    } catch {
        throw new ApiError(0, 'network error')
    }

    if (response.status === 401 && !_retry && !endpoint.includes('/auth/')) {
        const refreshRes = await fetch(`${API_URL}/auth/refresh`, {
            method: 'POST',
            credentials: 'include',
        })

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

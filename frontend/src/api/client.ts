const API_URL = 'http://localhost:8080/api/v1'

interface ApiOptions {
    method?: string
    body?: unknown
    _retry?: boolean
}

export async function api<T>(endpoint: string, options: ApiOptions = {}): Promise<T> {
    const { method = 'GET', body, _retry = false } = options

    const headers: Record<string, string> = {
        'Content-Type': 'application/json',
    }

    const response = await fetch(`${API_URL}${endpoint}`, {
        method,
        headers,
        body: body ? JSON.stringify(body) : undefined,
        credentials: 'include',
    })

    // If 401 and not already a retry, try refreshing the token
    if (response.status === 401 && !_retry && !endpoint.includes('/auth/')) {
        const refreshRes = await fetch(`${API_URL}/auth/refresh`, {
            method: 'POST',
            credentials: 'include',
        })

        if (refreshRes.ok) {
            // Retry the original request with new token
            return api<T>(endpoint, { ...options, _retry: true })
        }

        // Refresh failed — session is dead
        throw new Error('session_expired')
    }

    if (!response.ok) {
        const error = await response.json()
        throw new Error(error.error || 'something went wrong')
    }

    if (response.status === 204) {
        return {} as T
    }

    return response.json()
}

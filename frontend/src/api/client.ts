const API_URL = 'http://localhost:8080/api/v1'

interface ApiOptions {
    method?: string
    body?: unknown
}

export async function api<T>(endpoint: string, options: ApiOptions = {}): Promise<T> {
    const {method = 'GET', body } = options

    const headers : Record<string, string> = {
        'Content-Type' : 'application/json',
    }

    const response = await fetch(`${API_URL}${endpoint}`,{
        method,
        headers,
        body: body ? JSON.stringify(body) : undefined,
        credentials: 'include'
    })

    if (!response.ok) {
        const error = await response.json()
        throw new Error(error.error || 'something went wrong')
    }

    if (response.status === 204) {
        return {} as T
    }

    return response.json()
}
const LOOPBACK_HOSTS = ['localhost', '::1', '0.0.0.0']

export function hasHttpScheme(raw: string): boolean {
    return /^https?:\/\//i.test(raw.trim())
}

export function publicUrlHost(raw: string): string | null {
    try {
        const host = new URL(raw.trim()).hostname.toLowerCase()
        if (!host) return null
        return host.startsWith('[') && host.endsWith(']') ? host.slice(1, -1) : host
    } catch {
        return null
    }
}

export function isLoopbackUrl(raw: string): boolean {
    const host = publicUrlHost(raw)
    if (host === null) return false
    return LOOPBACK_HOSTS.includes(host) || host.startsWith('127.')
}

export function isPubliclyReachableUrl(raw: string): boolean {
    if (!raw.trim()) return false
    if (!hasHttpScheme(raw)) return false
    if (publicUrlHost(raw) === null) return false
    return !isLoopbackUrl(raw)
}

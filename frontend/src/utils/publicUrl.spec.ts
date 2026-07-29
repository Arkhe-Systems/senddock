import { describe, it, expect } from 'vitest'
import { hasHttpScheme, isLoopbackUrl, isPubliclyReachableUrl, publicUrlHost } from './publicUrl'

describe('hasHttpScheme', () => {
    it('accepts http and https in any case', () => {
        expect(hasHttpScheme('http://mail.example.com')).toBe(true)
        expect(hasHttpScheme('HTTPS://mail.example.com')).toBe(true)
        expect(hasHttpScheme('  https://mail.example.com  ')).toBe(true)
    })

    it('rejects a bare host or another scheme', () => {
        expect(hasHttpScheme('mail.example.com')).toBe(false)
        expect(hasHttpScheme('ftp://mail.example.com')).toBe(false)
        expect(hasHttpScheme('')).toBe(false)
    })
})

describe('isLoopbackUrl', () => {
    it('spots addresses that only work on this machine', () => {
        expect(isLoopbackUrl('http://localhost:8080')).toBe(true)
        expect(isLoopbackUrl('http://127.0.0.1:8080')).toBe(true)
        expect(isLoopbackUrl('http://[::1]:8080')).toBe(true)
        expect(isLoopbackUrl('http://0.0.0.0:8080')).toBe(true)
    })

    it('does not match a public host that merely contains the word', () => {
        expect(isLoopbackUrl('https://not-localhost.example.com')).toBe(false)
        expect(isLoopbackUrl('https://localhost.example.com')).toBe(false)
        expect(isLoopbackUrl('https://mail.127.example.com')).toBe(false)
    })
})

describe('publicUrlHost', () => {
    it('returns null when the value is not a URL', () => {
        expect(publicUrlHost('not a url')).toBeNull()
        expect(publicUrlHost('')).toBeNull()
    })
})

describe('isPubliclyReachableUrl', () => {
    it('is true only for an absolute http(s) URL on a routable host', () => {
        expect(isPubliclyReachableUrl('https://mail.example.com')).toBe(true)
        expect(isPubliclyReachableUrl('http://mail.example.com')).toBe(true)
    })

    it('is false for empty, schemeless and loopback values', () => {
        expect(isPubliclyReachableUrl('')).toBe(false)
        expect(isPubliclyReachableUrl('mail.example.com')).toBe(false)
        expect(isPubliclyReachableUrl('http://localhost:8080')).toBe(false)
    })
})

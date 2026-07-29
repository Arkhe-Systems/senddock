import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { api, ApiError } from '@/api/client'

export type LicenseTier = 'free' | 'pro' | 'team'

export interface LicenseStatus {
    tier: LicenseTier
    plan?: string
    has_license: boolean
    reason?: string
    checked_at?: string
    expires_at?: string
}

export interface LicenseActivation {
    valid: boolean
    plan?: string
    masked_key?: string
    reason?: string
    unverified?: boolean
}

export const useLicenseStore = defineStore('license', () => {
    const status = ref<LicenseStatus | null>(null)
    const loading = ref(false)
    const fetched = ref(false)
    const available = ref(true)

    async function fetch(force = false) {
        if (fetched.value && !force) return
        loading.value = true
        try {
            status.value = await api<LicenseStatus>('/license/status')
            available.value = true
        } catch (e) {
            if (e instanceof ApiError && e.status === 404) available.value = false
            status.value = { tier: 'free', has_license: false }
        } finally {
            loading.value = false
            fetched.value = true
        }
    }

    async function activate(licenseKey: string): Promise<LicenseActivation> {
        const result = await api<LicenseActivation>('/instance/license', {
            method: 'POST',
            body: { license_key: licenseKey },
        })
        await fetch(true)
        return result
    }

    const tier = computed<LicenseTier>(() => status.value?.tier || 'free')
    const allowsPro = computed(() => tier.value === 'pro' || tier.value === 'team')
    const allowsTeam = computed(() => tier.value === 'team')

    function reset() {
        status.value = null
        fetched.value = false
    }

    return { status, loading, available, tier, allowsPro, allowsTeam, fetch, activate, reset }
})

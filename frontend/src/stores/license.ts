import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { api } from '@/api/client'

export type LicenseTier = 'free' | 'pro' | 'team'

export interface LicenseStatus {
    tier: LicenseTier
    plan?: string
    has_license: boolean
    reason?: string
    checked_at?: string
    expires_at?: string
}

export const useLicenseStore = defineStore('license', () => {
    const status = ref<LicenseStatus | null>(null)
    const loading = ref(false)
    const fetched = ref(false)

    async function fetch(force = false) {
        if (fetched.value && !force) return
        loading.value = true
        try {
            status.value = await api<LicenseStatus>('/license/status')
        } catch {
            status.value = { tier: 'free', has_license: false }
        } finally {
            loading.value = false
            fetched.value = true
        }
    }

    const tier = computed<LicenseTier>(() => status.value?.tier || 'free')
    const allowsPro = computed(() => tier.value === 'pro' || tier.value === 'team')
    const allowsTeam = computed(() => tier.value === 'team')

    function reset() {
        status.value = null
        fetched.value = false
    }

    return { status, loading, tier, allowsPro, allowsTeam, fetch, reset }
})

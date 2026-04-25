import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAppStore = defineStore('app', () => {
    const deploymentMode = ref('self-hosted')
    const setupRequired = ref(false)
    const publicUrl = ref('')
    const checked = ref(false)

    const publicUrlIsReachable = computed(() => {
        const raw = publicUrl.value
        if (!raw) return false
        try {
            const parsed = new URL(raw)
            const host = parsed.hostname.toLowerCase()
            if (!host) return false
            return host !== 'localhost' && !host.startsWith('127.') && host !== '::1' && host !== '0.0.0.0'
        } catch {
            return false
        }
    })

    return { deploymentMode, setupRequired, publicUrl, checked, publicUrlIsReachable }
})

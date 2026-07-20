import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { isPubliclyReachableUrl } from '@/utils/publicUrl'

export const useAppStore = defineStore('app', () => {
    const deploymentMode = ref('self-hosted')
    const setupRequired = ref(false)
    const publicUrl = ref('')
    const sessionIdleTimeoutMinutes = ref(120)
    const checked = ref(false)

    const publicUrlIsReachable = computed(() => isPubliclyReachableUrl(publicUrl.value))

    return { deploymentMode, setupRequired, publicUrl, sessionIdleTimeoutMinutes, checked, publicUrlIsReachable }
})

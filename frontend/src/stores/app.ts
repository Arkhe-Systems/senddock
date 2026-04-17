import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAppStore = defineStore('app', () => {
    const deploymentMode = ref('self-hosted')
    const setupRequired = ref(false)
    const checked = ref(false)

    return { deploymentMode, setupRequired, checked }
})

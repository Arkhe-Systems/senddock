import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAppStore = defineStore('app', () => {
    const deploymentMode = ref('self-hosted')
    const setupRequired = ref(false)
    const publicUrl = ref('')
    const checked = ref(false)

    return { deploymentMode, setupRequired, publicUrl, checked }
})

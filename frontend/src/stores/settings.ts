import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/api/client'
import { useAppStore } from '@/stores/app'

export interface InstanceSettings {
    public_url: string
    session_idle_timeout_minutes: number
}

export const useSettingsStore = defineStore('settings', () => {
    const settings = ref<InstanceSettings | null>(null)
    const loading = ref(false)

    function sync(next: InstanceSettings) {
        settings.value = next
        const app = useAppStore()
        app.publicUrl = next.public_url
        app.sessionIdleTimeoutMinutes = next.session_idle_timeout_minutes
    }

    async function fetchSettings(): Promise<InstanceSettings | null> {
        loading.value = true
        try {
            sync(await api<InstanceSettings>('/instance/settings'))
        } catch {
            settings.value = null
        } finally {
            loading.value = false
        }
        return settings.value
    }

    async function updateSettings(input: Partial<InstanceSettings>): Promise<InstanceSettings> {
        const updated = await api<InstanceSettings>('/instance/settings', {
            method: 'PATCH',
            body: input,
        })
        sync(updated)
        return updated
    }

    return { settings, loading, fetchSettings, updateSettings }
})

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '@/api/client'
import { useAppStore } from '@/stores/app'

interface Dimension {
    used: number
    limit: number
}

interface Usage {
    plan: string
    subscribers: Dimension
    projects: Dimension
    seats: Dimension
    retention_days: number
}

interface UsageAlert {
    key: string
    label: string
    used: number
    limit: number
    level: 'warn' | 'over'
}

export const useUsageStore = defineStore('usage', () => {
    const usage = ref<Usage | null>(null)

    async function fetch() {
        if (useAppStore().deploymentMode !== 'cloud') return
        try {
            usage.value = await api<Usage>('/cloud/usage', { silent: true })
        } catch {
            usage.value = null
        }
    }

    const alerts = computed<UsageAlert[]>(() => {
        const u = usage.value
        if (!u) return []
        const dims = [
            { key: 'subscribers', label: 'subscribers', dim: u.subscribers },
            { key: 'projects', label: 'projects', dim: u.projects },
            { key: 'seats', label: 'team members', dim: u.seats },
        ]
        const out: UsageAlert[] = []
        for (const { key, label, dim } of dims) {
            if (dim.limit <= 1) continue
            const pct = dim.used / dim.limit
            if (dim.used >= dim.limit) {
                out.push({ key, label, used: dim.used, limit: dim.limit, level: 'over' })
            } else if (pct >= 0.9) {
                out.push({ key, label, used: dim.used, limit: dim.limit, level: 'warn' })
            }
        }
        return out
    })

    function reset() {
        usage.value = null
    }

    return { usage, fetch, alerts, reset }
})

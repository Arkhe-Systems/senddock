<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api/client'
import type { Project } from '@/stores/projects'

interface EmailStats {
    total: number
    sent: number
    failed: number
}

interface EmailLog {
    id: string
    to_email: string
    subject: string
    status: string
    error: string | null
    sent_at: string
}

const props = defineProps<{ project: Project }>()

const stats = ref<EmailStats>({ total: 0, sent: 0, failed: 0 })
const recentLogs = ref<EmailLog[]>([])
const loading = ref(true)

onMounted(async () => {
    try {
        const [statsRes, logsRes] = await Promise.all([
            api<EmailStats>(`/projects/${props.project.id}/stats`),
            api<{ logs: EmailLog[] | null, total: number }>(`/projects/${props.project.id}/logs?limit=10`),
        ])
        stats.value = statsRes
        recentLogs.value = logsRes.logs || []
    } catch {
    } finally {
        loading.value = false
    }
})
</script>

<template>
    <div>
        <h1 class="text-2xl font-bold text-white mb-6">Overview</h1>

        <div v-if="loading" class="text-zinc-500 py-8 text-center">Loading...</div>

        <div v-else>
            <div class="grid grid-cols-3 gap-4 mb-8">
                <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                    <p class="text-sm text-zinc-400">Total Emails</p>
                    <p class="text-2xl font-bold text-white mt-1">{{ stats.total }}</p>
                </div>
                <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                    <p class="text-sm text-zinc-400">Sent</p>
                    <p class="text-2xl font-bold text-green-400 mt-1">{{ stats.sent }}</p>
                </div>
                <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                    <p class="text-sm text-zinc-400">Failed</p>
                    <p class="text-2xl font-bold mt-1" :class="stats.failed > 0 ? 'text-red-400' : 'text-white'">{{ stats.failed }}</p>
                </div>
            </div>

            <div v-if="!project.smtp_host" class="bg-zinc-900 border border-zinc-800 rounded-lg p-6 mb-8">
                <p class="text-zinc-400 text-sm">Configure your SMTP settings to start sending emails.</p>
            </div>

            <div v-if="recentLogs.length > 0">
                <h2 class="text-lg font-semibold text-white mb-4">Recent Activity</h2>
                <div class="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
                    <table class="w-full">
                        <thead>
                            <tr class="border-b border-zinc-800">
                                <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">To</th>
                                <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Subject</th>
                                <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Status</th>
                                <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Date</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="log in recentLogs" :key="log.id" class="border-b border-zinc-800 last:border-0">
                                <td class="px-4 py-3 text-sm text-white">{{ log.to_email }}</td>
                                <td class="px-4 py-3 text-sm text-zinc-400">{{ log.subject }}</td>
                                <td class="px-4 py-3">
                                    <span :class="[
                                        'text-xs px-2 py-1 rounded-full',
                                        log.status === 'sent' && 'bg-green-500/10 text-green-400',
                                        log.status === 'failed' && 'bg-red-500/10 text-red-400',
                                    ]">
                                        {{ log.status }}
                                    </span>
                                </td>
                                <td class="px-4 py-3 text-sm text-zinc-500">{{ new Date(log.sent_at).toLocaleString() }}</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    </div>
</template>

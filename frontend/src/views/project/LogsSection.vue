<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api/client'
import type { Project } from '@/stores/projects'

interface EmailLog {
    id: string
    to_email: string
    subject: string
    status: string
    error: string | null
    sent_at: string
}

const props = defineProps<{ project: Project }>()

const logs = ref<EmailLog[]>([])
const total = ref(0)
const loading = ref(true)
const page = ref(0)
const limit = 25

const filterStatus = ref('')
const filterFrom = ref('')
const filterTo = ref('')

async function fetchLogs() {
    loading.value = true
    try {
        let url = `/projects/${props.project.id}/logs?limit=${limit}&offset=${page.value * limit}`
        if (filterStatus.value) url += `&status=${filterStatus.value}`
        if (filterFrom.value) url += `&from=${new Date(filterFrom.value).toISOString()}`
        if (filterTo.value) url += `&to=${new Date(filterTo.value + 'T23:59:59').toISOString()}`

        const res = await api<{ logs: EmailLog[] | null, total: number }>(url)
        logs.value = res.logs || []
        total.value = res.total
    } catch {
        logs.value = []
    } finally {
        loading.value = false
    }
}

function applyFilters() {
    page.value = 0
    fetchLogs()
}

function clearFilters() {
    filterStatus.value = ''
    filterFrom.value = ''
    filterTo.value = ''
    page.value = 0
    fetchLogs()
}

onMounted(fetchLogs)
</script>

<template>
    <div>
        <div class="flex items-center justify-between mb-6">
            <div>
                <h1 class="text-2xl font-bold text-white">Email Logs</h1>
                <p class="text-sm text-zinc-500 mt-1">{{ total }} total</p>
            </div>
        </div>

        <div class="flex items-end gap-3 mb-6">
            <div>
                <label class="block text-xs text-zinc-500 mb-1">Status</label>
                <select v-model="filterStatus" @change="applyFilters"
                    class="px-3 py-1.5 text-sm bg-zinc-900 border border-zinc-800 rounded-lg text-white">
                    <option value="">All</option>
                    <option value="sent">Sent</option>
                    <option value="failed">Failed</option>
                </select>
            </div>
            <div>
                <label class="block text-xs text-zinc-500 mb-1">From</label>
                <input v-model="filterFrom" type="date" @change="applyFilters"
                    class="px-3 py-1.5 text-sm bg-zinc-900 border border-zinc-800 rounded-lg text-white" />
            </div>
            <div>
                <label class="block text-xs text-zinc-500 mb-1">To</label>
                <input v-model="filterTo" type="date" @change="applyFilters"
                    class="px-3 py-1.5 text-sm bg-zinc-900 border border-zinc-800 rounded-lg text-white" />
            </div>
            <button v-if="filterStatus || filterFrom || filterTo" @click="clearFilters"
                class="px-3 py-1.5 text-sm text-zinc-500 hover:text-white transition cursor-pointer">
                Clear
            </button>
        </div>

        <div v-if="loading" class="text-zinc-500 py-8 text-center">Loading...</div>

        <div v-else-if="logs.length > 0" class="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
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
                    <template v-for="log in logs" :key="log.id">
                        <tr class="border-b border-zinc-800 last:border-0">
                            <td class="px-4 py-3 text-sm text-white">{{ log.to_email }}</td>
                            <td class="px-4 py-3 text-sm text-zinc-400">{{ log.subject || '(no subject)' }}</td>
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
                        <tr v-if="log.status === 'failed' && log.error">
                            <td colspan="4" class="px-4 py-2 bg-red-500/5">
                                <p class="text-xs text-red-400">{{ log.error }}</p>
                            </td>
                        </tr>
                    </template>
                </tbody>
            </table>
        </div>

        <div v-else class="bg-zinc-900 border border-zinc-800 rounded-lg p-8 text-center">
            <p class="text-zinc-400">No logs found.</p>
        </div>

        <div v-if="total > limit" class="flex items-center justify-between mt-4">
            <button @click="page--; fetchLogs()" :disabled="page === 0"
                class="text-sm text-zinc-400 hover:text-white disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer">
                Previous
            </button>
            <span class="text-sm text-zinc-500">Page {{ page + 1 }} of {{ Math.ceil(total / limit) }}</span>
            <button @click="page++; fetchLogs()" :disabled="(page + 1) * limit >= total"
                class="text-sm text-zinc-400 hover:text-white disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer">
                Next
            </button>
        </div>
    </div>
</template>

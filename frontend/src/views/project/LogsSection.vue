<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api/client'
import type { Project } from '@/stores/projects'
import AppPagination from '@/components/ui/AppPagination.vue'

interface EmailLog {
    id: string
    project_id: string
    subscriber_id: string | null
    template_id: string | null
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
const limit = ref(25)

const filterStatus = ref('')
const filterFrom = ref('')
const filterTo = ref('')
const filterSearch = ref('')
const expandedId = ref<string | null>(null)

let searchTimer: ReturnType<typeof setTimeout> | null = null

async function fetchLogs() {
    loading.value = true
    try {
        const params = new URLSearchParams({
            limit: String(limit.value),
            offset: String(page.value * limit.value),
        })
        if (filterStatus.value) params.set('status', filterStatus.value)
        if (filterFrom.value) params.set('from', new Date(filterFrom.value).toISOString())
        if (filterTo.value) params.set('to', new Date(filterTo.value + 'T23:59:59').toISOString())
        if (filterSearch.value.trim()) params.set('q', filterSearch.value.trim())

        const res = await api<{ logs: EmailLog[] | null, total: number }>(
            `/projects/${props.project.id}/logs?${params.toString()}`,
        )
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
    expandedId.value = null
    fetchLogs()
}

function onSearchInput() {
    if (searchTimer) clearTimeout(searchTimer)
    searchTimer = setTimeout(applyFilters, 300)
}

function clearFilters() {
    filterStatus.value = ''
    filterFrom.value = ''
    filterTo.value = ''
    filterSearch.value = ''
    page.value = 0
    expandedId.value = null
    fetchLogs()
}

function toggleExpand(id: string) {
    expandedId.value = expandedId.value === id ? null : id
}

const hasFilters = () =>
    filterStatus.value !== '' ||
    filterFrom.value !== '' ||
    filterTo.value !== '' ||
    filterSearch.value.trim() !== ''

onMounted(fetchLogs)
</script>

<template>
    <div>
        <div class="flex flex-wrap items-center justify-between gap-3 mb-6">
            <div>
                <h1 class="text-2xl font-bold text-white">Email Logs</h1>
                <p class="text-sm text-zinc-500 mt-1">{{ total }} total</p>
            </div>
        </div>

        <div class="flex flex-wrap items-end gap-3 mb-6">
            <div class="flex-1 min-w-[220px] max-w-md">
                <label class="block text-xs text-zinc-500 mb-1">Search</label>
                <input
                    v-model="filterSearch"
                    @input="onSearchInput"
                    type="text"
                    placeholder="email or subject..."
                    class="w-full px-3 py-1.5 text-sm bg-zinc-900 border border-zinc-800 rounded-lg text-white placeholder-zinc-600 focus:outline-none focus:border-zinc-600" />
            </div>
            <div>
                <label class="block text-xs text-zinc-500 mb-1">Status</label>
                <select v-model="filterStatus" @change="applyFilters"
                    class="px-3 py-1.5 text-sm bg-zinc-900 border border-zinc-800 rounded-lg text-white">
                    <option value="">All</option>
                    <option value="sent">Sent</option>
                    <option value="failed">Failed</option>
                    <option value="bounced">Bounced</option>
                    <option value="suppressed">Suppressed</option>
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
            <button v-if="hasFilters()" @click="clearFilters"
                class="px-3 py-1.5 text-sm text-zinc-500 hover:text-white transition cursor-pointer">
                Clear
            </button>
        </div>

        <div v-if="loading" class="text-zinc-500 py-8 text-center">Loading...</div>

        <div v-else-if="logs.length > 0" class="bg-zinc-900 border border-zinc-800 rounded-lg overflow-x-auto">
            <table class="w-full min-w-[640px]">
                <thead>
                    <tr class="border-b border-zinc-800">
                        <th class="w-8"></th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">To</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Subject</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Status</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Date</th>
                    </tr>
                </thead>
                <tbody>
                    <template v-for="log in logs" :key="log.id">
                        <tr
                            @click="toggleExpand(log.id)"
                            class="border-b border-zinc-800 last:border-0 hover:bg-zinc-800/40 cursor-pointer transition-colors">
                            <td class="px-3 py-3 text-zinc-500">
                                <span class="inline-block transition-transform" :class="expandedId === log.id ? 'rotate-90' : ''">›</span>
                            </td>
                            <td class="px-4 py-3 text-sm text-white">{{ log.to_email }}</td>
                            <td class="px-4 py-3 text-sm text-zinc-400">{{ log.subject || '(no subject)' }}</td>
                            <td class="px-4 py-3">
                                <span :class="[
                                    'text-xs px-2 py-1 rounded-full',
                                    log.status === 'sent' && 'bg-green-500/10 text-green-400',
                                    log.status === 'failed' && 'bg-red-500/10 text-red-400',
                                    log.status === 'bounced' && 'bg-orange-500/10 text-orange-400',
                                    log.status === 'suppressed' && 'bg-zinc-500/10 text-zinc-400',
                                ]">
                                    {{ log.status }}
                                </span>
                            </td>
                            <td class="px-4 py-3 text-sm text-zinc-500">{{ new Date(log.sent_at).toLocaleString() }}</td>
                        </tr>
                        <tr v-if="expandedId === log.id" class="border-b border-zinc-800 last:border-0 bg-zinc-950/60">
                            <td></td>
                            <td colspan="4" class="px-4 py-4">
                                <dl class="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-2 text-xs">
                                    <div>
                                        <dt class="text-zinc-500 uppercase tracking-wide font-medium">Log ID</dt>
                                        <dd class="text-zinc-300 font-mono break-all">{{ log.id }}</dd>
                                    </div>
                                    <div>
                                        <dt class="text-zinc-500 uppercase tracking-wide font-medium">Sent at</dt>
                                        <dd class="text-zinc-300 font-mono">{{ new Date(log.sent_at).toISOString() }}</dd>
                                    </div>
                                    <div v-if="log.subscriber_id">
                                        <dt class="text-zinc-500 uppercase tracking-wide font-medium">Subscriber</dt>
                                        <dd class="text-zinc-300 font-mono break-all">{{ log.subscriber_id }}</dd>
                                    </div>
                                    <div v-if="log.template_id">
                                        <dt class="text-zinc-500 uppercase tracking-wide font-medium">Template</dt>
                                        <dd class="text-zinc-300 font-mono break-all">{{ log.template_id }}</dd>
                                    </div>
                                </dl>
                                <div v-if="log.error" class="mt-3 pt-3 border-t border-zinc-800">
                                    <p class="text-xs text-zinc-500 uppercase tracking-wide font-medium mb-1">
                                        {{ log.status === 'suppressed' ? 'Reason' : 'Error' }}
                                    </p>
                                    <p :class="[
                                        'text-xs whitespace-pre-wrap break-words',
                                        log.status === 'failed' && 'text-red-400',
                                        log.status === 'bounced' && 'text-orange-400',
                                        log.status === 'suppressed' && 'text-zinc-400',
                                    ]">
                                        {{ log.error }}
                                    </p>
                                </div>
                            </td>
                        </tr>
                    </template>
                </tbody>
            </table>
        </div>

        <div v-else class="bg-zinc-900 border border-zinc-800 rounded-lg p-8 text-center">
            <p class="text-zinc-400">No logs found.</p>
        </div>

        <AppPagination
            v-model:page="page"
            v-model:limit="limit"
            :total="total"
            @change="fetchLogs" />
    </div>
</template>

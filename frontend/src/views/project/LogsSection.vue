<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api, getApiBase } from '@/api/client'
import type { Project } from '@/stores/projects'
import { useNewsletterStore } from '@/stores/newsletters'
import AppPagination from '@/components/ui/AppPagination.vue'
import AppFilterChip from '@/components/ui/AppFilterChip.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppSelect from '@/components/ui/AppSelect.vue'
import AppStatusPill, { type PillTone } from '@/components/ui/AppStatusPill.vue'
import LogDetailDrawer from './LogDetailDrawer.vue'

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
    opened_at: string | null
    clicked_at: string | null
}

interface Template { id: string; name: string }

const STATUS_CHIPS: { value: string; label: string }[] = [
    { value: '', label: 'All' },
    { value: 'sent', label: 'Sent' },
    { value: 'failed', label: 'Failed' },
    { value: 'bounced', label: 'Bounced' },
    { value: 'suppressed', label: 'Suppressed' },
]

const props = defineProps<{ project: Project }>()
const newsletterStore = useNewsletterStore()
const newsletters = computed(() => newsletterStore.newsletters(props.project.id))

const logs = ref<EmailLog[]>([])
const templates = ref<Template[]>([])
const total = ref(0)
const loading = ref(true)
const exporting = ref(false)
const page = ref(0)
const limit = ref(25)

const filterStatus = ref('')
const filterTemplateId = ref('')
const filterNewsletterId = ref('')
const filterFrom = ref('')
const filterTo = ref('')
const filterSearch = ref('')
const selectedLogId = ref<string | null>(null)

let searchTimer: ReturnType<typeof setTimeout> | null = null

function buildFilterParams(): URLSearchParams {
    const params = new URLSearchParams()
    if (filterStatus.value) params.set('status', filterStatus.value)
    if (filterTemplateId.value) params.set('template_id', filterTemplateId.value)
    if (filterNewsletterId.value) params.set('newsletter_id', filterNewsletterId.value)
    if (filterFrom.value) params.set('from', new Date(filterFrom.value).toISOString())
    if (filterTo.value) params.set('to', new Date(filterTo.value + 'T23:59:59').toISOString())
    if (filterSearch.value.trim()) params.set('q', filterSearch.value.trim())
    return params
}

async function fetchLogs() {
    loading.value = true
    try {
        const params = buildFilterParams()
        params.set('limit', String(limit.value))
        params.set('offset', String(page.value * limit.value))

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

async function fetchTemplates() {
    try {
        const res = await api<Template[] | null>(`/projects/${props.project.id}/templates`)
        templates.value = res || []
    } catch {
        templates.value = []
    }
}

function applyFilters() {
    page.value = 0
    selectedLogId.value = null
    fetchLogs()
}

function setStatus(v: string) {
    filterStatus.value = v
    applyFilters()
}

function statusTone(status: string): PillTone {
    switch (status) {
        case 'sent': return 'emerald'
        case 'failed': return 'red'
        case 'bounced': return 'orange'
        default: return 'zinc'
    }
}

const templateOptions = computed(() => [
    { value: '', label: 'All templates' },
    ...templates.value.map(t => ({ value: t.id, label: t.name })),
])

const newsletterOptions = computed(() => [
    { value: '', label: 'All newsletters' },
    ...newsletters.value.map(n => ({ value: n.id, label: n.name })),
])

function onSearchInput() {
    if (searchTimer) clearTimeout(searchTimer)
    searchTimer = setTimeout(applyFilters, 300)
}

function clearFilters() {
    filterStatus.value = ''
    filterTemplateId.value = ''
    filterNewsletterId.value = ''
    filterFrom.value = ''
    filterTo.value = ''
    filterSearch.value = ''
    page.value = 0
    selectedLogId.value = null
    fetchLogs()
}

async function exportCSV() {
    exporting.value = true
    try {
        const params = buildFilterParams()
        const url = `${getApiBase()}/projects/${props.project.id}/logs/export.csv?${params.toString()}`
        const res = await fetch(url, { credentials: 'include' })
        if (!res.ok) throw new Error('export failed')
        const blob = await res.blob()
        const link = document.createElement('a')
        link.href = URL.createObjectURL(blob)
        link.download = res.headers.get('Content-Disposition')?.match(/filename="([^"]+)"/)?.[1]
            ?? `email-logs-${new Date().toISOString().slice(0, 10)}.csv`
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        URL.revokeObjectURL(link.href)
    } catch {
        // silent: button shows loading state, user retries
    } finally {
        exporting.value = false
    }
}

function openDetail(id: string) {
    selectedLogId.value = id
}

function closeDetail() {
    selectedLogId.value = null
}

const hasFilters = () =>
    filterStatus.value !== '' ||
    filterTemplateId.value !== '' ||
    filterNewsletterId.value !== '' ||
    filterFrom.value !== '' ||
    filterTo.value !== '' ||
    filterSearch.value.trim() !== ''

onMounted(() => {
    fetchLogs()
    fetchTemplates()
    newsletterStore.fetchNewsletters(props.project.id)
})
</script>

<template>
    <div>
        <div class="flex flex-wrap items-center justify-between gap-3 mb-6">
            <div>
                <h1 class="text-xl font-semibold text-white">Email Logs</h1>
                <p class="text-sm text-zinc-400 mt-1">{{ total }} total</p>
            </div>
            <AppButton size="md" :loading="exporting" :disabled="exporting" @click="exportCSV">
                {{ exporting ? 'Exporting…' : 'Export CSV' }}
            </AppButton>
        </div>

        <div class="flex flex-wrap items-center gap-2 mb-3">
            <AppFilterChip v-for="chip in STATUS_CHIPS" :key="chip.value"
                :active="filterStatus === chip.value"
                @click="setStatus(chip.value)">
                {{ chip.label }}
            </AppFilterChip>
        </div>

        <div class="flex flex-wrap items-end gap-3 mb-6">
            <div class="flex-1 min-w-[220px] max-w-md">
                <label class="block text-xs text-zinc-400 mb-1">Search</label>
                <input
                    v-model="filterSearch"
                    @input="onSearchInput"
                    type="text"
                    placeholder="email or subject..."
                    class="w-full px-3 py-1.5 text-sm bg-zinc-900 border border-zinc-800 rounded-lg text-white placeholder-zinc-600 focus:outline-none focus:border-emerald-500" />
            </div>
            <div>
                <label class="block text-xs text-zinc-400 mb-1">Template</label>
                <AppSelect v-model="filterTemplateId" size="sm" class="max-w-[200px] truncate"
                    :options="templateOptions" @change="applyFilters" />
            </div>
            <div v-if="newsletters.length">
                <label class="block text-xs text-zinc-400 mb-1">Newsletter</label>
                <AppSelect v-model="filterNewsletterId" size="sm" class="max-w-[200px] truncate"
                    :options="newsletterOptions" @change="applyFilters" />
            </div>
            <div>
                <label class="block text-xs text-zinc-400 mb-1">From</label>
                <input v-model="filterFrom" type="date" @change="applyFilters"
                    class="px-3 py-1.5 text-sm bg-zinc-900 border border-zinc-800 rounded-lg text-white [color-scheme:dark] cursor-pointer" />
            </div>
            <div>
                <label class="block text-xs text-zinc-400 mb-1">To</label>
                <input v-model="filterTo" type="date" @change="applyFilters"
                    class="px-3 py-1.5 text-sm bg-zinc-900 border border-zinc-800 rounded-lg text-white [color-scheme:dark] cursor-pointer" />
            </div>
            <button v-if="hasFilters()" @click="clearFilters"
                class="px-3 py-1.5 text-sm text-zinc-400 hover:text-white transition cursor-pointer">
                Clear
            </button>
        </div>

        <div v-if="loading" class="text-zinc-400 py-8 text-center">Loading...</div>

        <div v-else-if="logs.length > 0" class="bg-zinc-900 border border-zinc-800 rounded-lg overflow-x-auto">
            <table class="w-full min-w-[640px]">
                <thead>
                    <tr class="border-b border-zinc-800">
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">To</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">Subject</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">Status</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">Engagement</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">Date</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="log in logs" :key="log.id"
                        @click="openDetail(log.id)"
                        :class="[
                            'border-b border-zinc-800 last:border-0 hover:bg-zinc-850/40 cursor-pointer transition-colors',
                            selectedLogId === log.id && 'bg-zinc-850/40'
                        ]">
                        <td class="px-4 py-3 text-sm text-white">{{ log.to_email }}</td>
                        <td class="px-4 py-3 text-sm text-zinc-300 max-w-md truncate">{{ log.subject || '(no subject)' }}</td>
                        <td class="px-4 py-3">
                            <AppStatusPill :tone="statusTone(log.status)" :label="log.status" />
                        </td>
                        <td class="px-4 py-3">
                            <div class="flex items-center gap-1.5">
                                <span v-if="log.opened_at" title="Opened"
                                    class="inline-flex items-center justify-center w-5 h-5 rounded bg-blue-500/15 text-blue-400 text-[10px] font-semibold">O</span>
                                <span v-if="log.clicked_at" title="Clicked"
                                    class="inline-flex items-center justify-center w-5 h-5 rounded bg-amber-500/15 text-amber-400 text-[10px] font-semibold">C</span>
                                <span v-if="!log.opened_at && !log.clicked_at" class="text-xs text-zinc-600">—</span>
                            </div>
                        </td>
                        <td class="px-4 py-3 text-sm text-zinc-400 whitespace-nowrap">{{ new Date(log.sent_at).toLocaleString() }}</td>
                    </tr>
                </tbody>
            </table>
        </div>

        <div v-else class="bg-zinc-900 border border-zinc-800 rounded-lg p-8 text-center">
            <p class="text-zinc-300">No logs found.</p>
        </div>

        <AppPagination
            v-model:page="page"
            v-model:limit="limit"
            :total="total"
            @change="fetchLogs" />

        <LogDetailDrawer
            :project-id="project.id"
            :log-id="selectedLogId"
            @close="closeDetail" />
    </div>
</template>

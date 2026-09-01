<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { api } from '@/api/client'
import { useLicenseStore } from '@/stores/license'
import type { Project } from '@/stores/projects'
import AppProPaywall from '@/components/ui/AppProPaywall.vue'
import AppPagination from '@/components/ui/AppPagination.vue'
import AppSelect from '@/components/ui/AppSelect.vue'
import AppButton from '@/components/ui/AppButton.vue'

const licenseStore = useLicenseStore()

interface AuditEntry {
    id: string
    user_id: string
    action: string
    target_type: string
    target_id: string
    metadata: Record<string, any>
    ip_address: string
    user_agent: string
    created_at: string
}

const ACTIONS = [
    { value: '', label: 'All actions' },
    { value: 'project.create', label: 'Project created' },
    { value: 'project.update', label: 'Project updated' },
    { value: 'project.delete', label: 'Project deleted' },
    { value: 'smtp.update', label: 'SMTP updated' },
    { value: 'broadcast.send', label: 'Broadcast sent' },
    { value: 'api_key.create', label: 'API key created' },
    { value: 'api_key.revoke', label: 'API key revoked' },
    { value: 'webhook.create', label: 'Webhook created' },
    { value: 'webhook.delete', label: 'Webhook deleted' },
    { value: 'suppression.add', label: 'Suppression added' },
    { value: 'suppression.delete', label: 'Suppression removed' },
    { value: 'subscriber.bulk_delete', label: 'Subscribers bulk-deleted' },
    { value: 'subscriber.bulk_update_status', label: 'Subscribers status changed' },
]

const props = defineProps<{ project: Project }>()

const entries = ref<AuditEntry[]>([])
const total = ref(0)
const loading = ref(true)
const errorState = ref<'none' | 'paywall' | 'generic'>('none')

const filterAction = ref('')
const filterFrom = ref('')
const filterTo = ref('')
const page = ref(0)
const limit = ref(50)

const summary = computed(() => `${total.value} ${total.value === 1 ? 'entry' : 'entries'}`)

const actionLabel = (a: string) => ACTIONS.find(x => x.value === a)?.label ?? a

async function fetchList() {
    loading.value = true
    errorState.value = 'none'
    await licenseStore.fetch()
    if (!licenseStore.allowsPro) {
        errorState.value = 'paywall'
        loading.value = false
        return
    }
    try {
        const params = new URLSearchParams({ limit: String(limit.value), offset: String(page.value * limit.value) })
        if (filterAction.value) params.set('action', filterAction.value)
        if (filterFrom.value) params.set('from', new Date(filterFrom.value).toISOString())
        if (filterTo.value) params.set('to', new Date(filterTo.value + 'T23:59:59').toISOString())
        const res = await api<{ entries: AuditEntry[] | null, total: number }>(
            `/projects/${props.project.id}/audit-log?${params.toString()}`,
        )
        entries.value = res.entries ?? []
        total.value = res.total
    } catch (e) {
        const msg = e instanceof Error ? e.message : ''
        if (msg.includes('license required')) errorState.value = 'paywall'
        else errorState.value = 'generic'
    } finally {
        loading.value = false
    }
}

function applyFilters() {
    page.value = 0
    fetchList()
}

function clearFilters() {
    filterAction.value = ''
    filterFrom.value = ''
    filterTo.value = ''
    applyFilters()
}

function fmtDate(iso: string) {
    return new Date(iso).toLocaleString('en-US', { dateStyle: 'medium', timeStyle: 'short' })
}

function fmtMetadata(m: Record<string, any> | undefined) {
    if (!m || Object.keys(m).length === 0) return ''
    return Object.entries(m)
        .map(([k, v]) => `${k}=${typeof v === 'object' ? JSON.stringify(v) : v}`)
        .join(' · ')
}

onMounted(fetchList)
</script>

<template>
    <div>
        <div class="flex flex-wrap items-start justify-between gap-3 mb-6">
            <div>
                <div class="flex items-center gap-2">
                    <h1 class="text-xl font-semibold text-white">Audit log</h1>
                    <span class="text-[10px] font-semibold tracking-wider uppercase px-1.5 py-0.5 rounded bg-amber-500/15 text-amber-400 border border-amber-500/30">Pro</span>
                </div>
                <p class="text-sm text-zinc-400 mt-1">
                    Every important action against this project. {{ summary }}.
                </p>
            </div>
        </div>

        <div v-if="loading" class="text-zinc-400 py-8 text-center">Loading...</div>

        <AppProPaywall v-else-if="errorState === 'paywall'"
            title="Audit log"
            description="A timestamped trail of every project change — who changed what, from which IP, with what payload." />

        <div v-else-if="errorState === 'generic'"
            class="bg-zinc-900 border border-zinc-800 rounded-lg p-6 text-center">
            <p class="text-zinc-300 text-sm">Couldn't load the audit log. Try again later.</p>
            <AppButton variant="outline" size="sm" class="mt-3" @click="fetchList">Retry</AppButton>
        </div>

        <template v-else>
            <div class="flex items-end gap-3 mb-4 flex-wrap">
                <div>
                    <label class="block text-xs text-zinc-400 mb-1">Action</label>
                    <AppSelect v-model="filterAction" size="sm" :options="ACTIONS" @change="applyFilters" />
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
                <button v-if="filterAction || filterFrom || filterTo" @click="clearFilters"
                    class="px-3 py-1.5 text-sm text-zinc-400 hover:text-white transition cursor-pointer">
                    Clear
                </button>
            </div>

            <div v-if="entries.length === 0" class="bg-zinc-900 border border-zinc-800 rounded-lg p-10 text-center">
                <h2 class="text-base font-semibold text-white mb-1">No entries match these filters</h2>
                <p class="text-sm text-zinc-400">As soon as you create projects, send broadcasts, manage webhooks or change settings, they'll show up here.</p>
            </div>

            <div v-else class="bg-zinc-900 border border-zinc-800 rounded-lg overflow-x-auto">
                <table class="w-full min-w-[768px]">
                    <thead>
                        <tr class="border-b border-zinc-800">
                            <th class="text-left px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">When</th>
                            <th class="text-left px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">Action</th>
                            <th class="text-left px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">Target</th>
                            <th class="text-left px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">Details</th>
                            <th class="text-left px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">IP</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="entry in entries" :key="entry.id" class="border-b border-zinc-800 last:border-0 hover:bg-zinc-850/40 transition">
                            <td class="px-4 py-3 text-sm text-zinc-300 whitespace-nowrap">{{ fmtDate(entry.created_at) }}</td>
                            <td class="px-4 py-3 text-sm">
                                <span class="px-2 py-0.5 text-xs rounded-md bg-zinc-850 text-zinc-300 font-mono">{{ entry.action }}</span>
                                <span class="block text-xs text-zinc-400 mt-1">{{ actionLabel(entry.action) }}</span>
                            </td>
                            <td class="px-4 py-3 text-sm text-zinc-300">
                                <span v-if="entry.target_type">{{ entry.target_type }}</span>
                                <span v-if="entry.target_id" class="block text-xs font-mono text-zinc-400 truncate max-w-xs">{{ entry.target_id }}</span>
                            </td>
                            <td class="px-4 py-3 text-xs text-zinc-300 font-mono break-words max-w-md">{{ fmtMetadata(entry.metadata) }}</td>
                            <td class="px-4 py-3 text-xs text-zinc-400 font-mono whitespace-nowrap">{{ entry.ip_address || '—' }}</td>
                        </tr>
                    </tbody>
                </table>
            </div>

            <AppPagination
                v-model:page="page"
                v-model:limit="limit"
                :total="total"
                @change="fetchList" />
        </template>
    </div>
</template>

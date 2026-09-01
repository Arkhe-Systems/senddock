<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { api } from '@/api/client'
import { useToastStore } from '@/stores/toast'
import type { Project } from '@/stores/projects'
import AppButton from '@/components/ui/AppButton.vue'
import AppModal from '@/components/ui/AppModal.vue'
import AppConfirmModal from '@/components/ui/AppConfirmModal.vue'
import AppPagination from '@/components/ui/AppPagination.vue'
import AppFilterChip from '@/components/ui/AppFilterChip.vue'

interface Suppression {
    id: string
    project_id: string
    email: string
    reason: string
    source: string
    created_at: string
    last_seen_at: string
}

const REASONS = [
    { value: '', label: 'All reasons' },
    { value: 'unsubscribe', label: 'Unsubscribed' },
    { value: 'bounce', label: 'Bounced' },
    { value: 'complaint', label: 'Complaint' },
    { value: 'manual', label: 'Manually added' },
]

const props = defineProps<{ project: Project }>()
const toast = useToastStore()

const suppressions = ref<Suppression[]>([])
const total = ref(0)
const loading = ref(true)

const reasonFilter = ref('')
const page = ref(0)
const limit = ref(50)

const showAddModal = ref(false)
const newEmails = ref('')
const newReason = ref('manual')
const adding = ref(false)

const itemToDelete = ref<Suppression | null>(null)
const deleting = ref(false)

const reasonLabel = (reason: string) => REASONS.find(r => r.value === reason)?.label ?? reason

const summary = computed(() => `${total.value} ${total.value === 1 ? 'address' : 'addresses'} suppressed`)

async function fetchList() {
    loading.value = true
    try {
        const params = new URLSearchParams({
            limit: String(limit.value),
            offset: String(page.value * limit.value),
        })
        if (reasonFilter.value) params.set('reason', reasonFilter.value)
        const res = await api<{ suppressions: Suppression[] | null, total: number }>(
            `/projects/${props.project.id}/suppressions?${params.toString()}`,
        )
        suppressions.value = res.suppressions ?? []
        total.value = res.total
    } catch (e) {
        toast.error(e instanceof Error ? e.message : 'Failed to load suppressions')
    } finally {
        loading.value = false
    }
}

function applyReason(value: string) {
    reasonFilter.value = value
    page.value = 0
    fetchList()
}

function parseEmails(text: string): string[] {
    return text
        .split(/[\n,;]+/)
        .map(s => s.trim())
        .filter(Boolean)
}

async function submitAdd() {
    const emails = parseEmails(newEmails.value)
    if (emails.length === 0) {
        toast.error('Paste at least one email')
        return
    }
    adding.value = true
    try {
        const res = await api<{ added: number, skipped: number }>(
            `/projects/${props.project.id}/suppressions`,
            {
                method: 'POST',
                body: { emails, reason: newReason.value, source: 'added from dashboard' },
            },
        )
        if (res.added > 0) {
            toast.success(`${res.added} address${res.added === 1 ? '' : 'es'} suppressed`)
        }
        if (res.skipped > 0) {
            toast.error(`${res.skipped} skipped (invalid format)`)
        }
        showAddModal.value = false
        newEmails.value = ''
        newReason.value = 'manual'
        page.value = 0
        fetchList()
    } catch (e) {
        toast.error(e instanceof Error ? e.message : 'Failed to add')
    } finally {
        adding.value = false
    }
}

function readFile(file: File) {
    if (!/\.csv$|text\/csv|text\/plain/i.test(file.name + ' ' + file.type)) {
        toast.error('Pick a .csv or .txt file')
        return
    }
    const reader = new FileReader()
    reader.onload = e => { newEmails.value = String(e.target?.result ?? '') }
    reader.onerror = () => toast.error('Could not read file')
    reader.readAsText(file)
}

function handleFileUpload(event: Event) {
    const input = event.target as HTMLInputElement
    const file = input.files?.[0]
    if (file) readFile(file)
    input.value = ''
}

async function deleteSuppression() {
    if (!itemToDelete.value) return
    deleting.value = true
    try {
        await api(`/projects/${props.project.id}/suppressions/${itemToDelete.value.id}`, { method: 'DELETE' })
        toast.success('Removed from suppression list')
        itemToDelete.value = null
        fetchList()
    } catch (e) {
        toast.error(e instanceof Error ? e.message : 'Failed to remove')
    } finally {
        deleting.value = false
    }
}

function fmtDate(iso: string) {
    return new Date(iso).toLocaleString('en-US', { dateStyle: 'medium', timeStyle: 'short' })
}

onMounted(fetchList)
</script>

<template>
    <div>
        <div class="flex flex-wrap items-start justify-between gap-3 mb-6">
            <div>
                <h1 class="text-xl font-semibold text-white">Suppressions</h1>
                <p class="text-sm text-zinc-400 mt-1">Addresses that will be skipped on every send. {{ summary }}.</p>
            </div>
            <AppButton size="sm" @click="showAddModal = true">+ Add address</AppButton>
        </div>

        <div class="flex flex-wrap gap-2 mb-4">
            <AppFilterChip v-for="r in REASONS" :key="r.value"
                :active="reasonFilter === r.value"
                @click="applyReason(r.value)">
                {{ r.label }}
            </AppFilterChip>
        </div>

        <div v-if="loading" class="text-zinc-400 py-8 text-center">Loading...</div>

        <div v-else-if="suppressions.length === 0" class="bg-zinc-900 border border-zinc-800 rounded-lg p-10 text-center">
            <h2 class="text-base font-semibold text-white mb-1">No suppressed addresses</h2>
            <p class="text-sm text-zinc-400 mb-5 max-w-md mx-auto">
                When subscribers unsubscribe or you add addresses manually, they'll show up here. SendDock will skip every send that targets them.
            </p>
            <AppButton size="sm" @click="showAddModal = true">Add your first address</AppButton>
        </div>

        <div v-else class="bg-zinc-900 border border-zinc-800 rounded-lg overflow-x-auto">
            <table class="w-full min-w-[640px]">
                <thead>
                    <tr class="border-b border-zinc-800">
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">Email</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">Reason</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">Source</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">Last seen</th>
                        <th class="text-right px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">Actions</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="s in suppressions" :key="s.id" class="border-b border-zinc-800 last:border-0 hover:bg-zinc-850/40 transition">
                        <td class="px-4 py-3 text-sm text-white font-mono break-all">{{ s.email }}</td>
                        <td class="px-4 py-3 text-sm">
                            <span class="px-2 py-0.5 text-xs rounded-md bg-zinc-850 text-zinc-300">{{ reasonLabel(s.reason) }}</span>
                        </td>
                        <td class="px-4 py-3 text-sm text-zinc-300">{{ s.source || '—' }}</td>
                        <td class="px-4 py-3 text-sm text-zinc-300">{{ fmtDate(s.last_seen_at) }}</td>
                        <td class="px-4 py-3 text-right">
                            <button @click="itemToDelete = s"
                                class="px-3 py-1 text-xs text-red-400 border border-red-900/50 rounded-md hover:bg-red-950/40 transition cursor-pointer">
                                Remove
                            </button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <AppPagination
            v-if="!loading"
            v-model:page="page"
            v-model:limit="limit"
            :total="total"
            @change="fetchList" />

        <AppModal :show="showAddModal" title="Add to suppression list" size="lg" @close="showAddModal = false">
            <form @submit.prevent="submitAdd" class="space-y-4">
                <div>
                    <div class="flex items-center justify-between mb-1">
                        <label class="text-sm font-medium text-zinc-300">Emails (one per line, comma or semicolon-separated)</label>
                        <label class="text-xs text-zinc-300 hover:text-white border border-zinc-700 rounded-md px-2 py-1 cursor-pointer transition hover:bg-zinc-850">
                            Choose file
                            <input type="file" accept=".csv,.txt,text/csv,text/plain" class="hidden" @change="handleFileUpload" />
                        </label>
                    </div>
                    <textarea v-model="newEmails" rows="6"
                        placeholder="ada@example.com&#10;alan@example.com&#10;…or paste a CSV with one column"
                        class="w-full px-3 py-2 bg-zinc-950 border border-zinc-800 rounded-lg text-sm text-white font-mono placeholder-zinc-600 focus:outline-none focus:ring-2 focus:ring-emerald-500 transition resize-y" />
                </div>

                <div>
                    <label class="block text-sm font-medium text-zinc-300 mb-1">Reason</label>
                    <select v-model="newReason"
                        class="w-full px-3 py-2 bg-zinc-950 border border-zinc-800 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-emerald-500 transition cursor-pointer">
                        <option value="manual">Manually added</option>
                        <option value="unsubscribe">Unsubscribed</option>
                        <option value="bounce">Bounced</option>
                        <option value="complaint">Complaint</option>
                    </select>
                </div>

                <div class="flex gap-2 pt-2">
                    <AppButton type="button" variant="ghost" size="sm" class="flex-1" @click="showAddModal = false">Cancel</AppButton>
                    <AppButton type="submit" size="sm" :loading="adding" :disabled="adding" class="flex-1">
                        {{ adding ? 'Adding...' : 'Add to list' }}
                    </AppButton>
                </div>
            </form>
        </AppModal>

        <AppConfirmModal
            :show="!!itemToDelete"
            title="Remove from suppression list"
            :message="itemToDelete ? `Remove ${itemToDelete.email} from the suppression list? Future sends to this address will go through again.` : ''"
            confirmLabel="Remove"
            danger
            :loading="deleting"
            @confirm="deleteSuppression"
            @cancel="itemToDelete = null" />
    </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '@/api/client'
import { useToastStore } from '@/stores/toast'
import type { Project } from '@/stores/projects'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import AppModal from '@/components/ui/AppModal.vue'
import AppCheckbox from '@/components/ui/AppCheckbox.vue'
import AppConfirmModal from '@/components/ui/AppConfirmModal.vue'

interface Subscriber {
    id: string
    email: string
    name: string
    status: string
    created_at: string
}

const props = defineProps<{ project: Project }>()
const toast = useToastStore()

const subscribers = ref<Subscriber[]>([])
const total = ref(0)
const loading = ref(true)

const showAddModal = ref(false)
const newEmail = ref('')
const newName = ref('')
const addLoading = ref(false)

const showImportModal = ref(false)
const importText = ref('')
const importLoading = ref(false)
const validateMX = ref(true)
const validateDisposable = ref(true)

interface RejectedRow { email: string; name: string; reason: string }
interface ImportResult {
    imported: number
    duplicates: number
    syntax_invalid: number
    no_mx: number
    disposable: number
    suppressed: number
    rejected: RejectedRow[]
}
const importResult = ref<ImportResult | null>(null)

const page = ref(0)
const limit = 50

const selectedIds = ref<string[]>([])

const allSelected = computed(() => {
    return subscribers.value.length > 0 && selectedIds.value.length === subscribers.value.length
})

function toggleSelectAll(checked: boolean) {
    if (checked) {
        selectedIds.value = subscribers.value.map(s => s.id)
    } else {
        selectedIds.value = []
    }
}

function toggleSelected(id: string, checked: boolean) {
    if (checked) {
        if (!selectedIds.value.includes(id)) selectedIds.value.push(id)
    } else {
        selectedIds.value = selectedIds.value.filter(s => s !== id)
    }
}

const bulkLoading = ref(false)
const showBulkDeleteConfirm = ref(false)

function confirmBulkDelete() {
    showBulkDeleteConfirm.value = true
}

async function handleBulkAction(action: 'delete' | 'update_status', status?: string) {
    bulkLoading.value = true
    try {
        await api(`/projects/${props.project.id}/subscribers/bulk`, {
            method: 'POST',
            body: {
                action,
                status,
                subscriber_ids: selectedIds.value
            }
        })
        toast.success(`Bulk action completed`)
        selectedIds.value = []
        showBulkDeleteConfirm.value = false
        fetchSubscribers()
    } catch (e: any) {
        toast.error(e.message || 'Failed to perform bulk action')
    } finally {
        bulkLoading.value = false
    }
}

async function fetchSubscribers() {
    loading.value = true
    try {
        const res = await api<{ subscribers: Subscriber[] | null, total: number }>(
            `/projects/${props.project.id}/subscribers?limit=${limit}&offset=${page.value * limit}`
        )
        subscribers.value = res.subscribers || []
        total.value = res.total
        selectedIds.value = [] // clear selection on page change
    } catch {
        subscribers.value = []
    } finally {
        loading.value = false
    }
}

async function handleAdd() {
    if (!newEmail.value) {
        toast.error('Email is required')
        return
    }
    addLoading.value = true
    try {
        await api(`/projects/${props.project.id}/subscribers`, {
            method: 'POST',
            body: { email: newEmail.value, name: newName.value },
        })
        showAddModal.value = false
        newEmail.value = ''
        newName.value = ''
        toast.success('Subscriber added')
        fetchSubscribers()
    } catch (e: any) {
        toast.error(e.message || 'Failed to add subscriber')
    } finally {
        addLoading.value = false
    }
}

async function toggleStatus(sub: Subscriber) {
    const newStatus = sub.status === 'active' ? 'unsubscribed' : 'active'
    try {
        await api(`/projects/${props.project.id}/subscribers/${sub.id}`, {
            method: 'PATCH',
            body: { status: newStatus },
        })
        toast.success(`Subscriber ${newStatus === 'active' ? 'activated' : 'unsubscribed'}`)
        fetchSubscribers()
    } catch (e: any) {
        toast.error(e.message || 'Failed to update status')
    }
}

const showDeleteModal = ref(false)
const subscriberToDelete = ref<Subscriber | null>(null)
const deleteLoading = ref(false)

function openDeleteModal(sub: Subscriber) {
    subscriberToDelete.value = sub
    showDeleteModal.value = true
}

async function handleDelete() {
    if (!subscriberToDelete.value) return
    deleteLoading.value = true
    try {
        await api(`/projects/${props.project.id}/subscribers/${subscriberToDelete.value.id}`, {
            method: 'DELETE',
        })
        toast.success('Subscriber removed')
        showDeleteModal.value = false
        subscriberToDelete.value = null
        fetchSubscribers()
    } catch (e: any) {
        toast.error(e.message || 'Failed to delete subscriber')
    } finally {
        deleteLoading.value = false
    }
}

function parseImportText(text: string): { email: string; name: string }[] {
    const rows: { email: string; name: string }[] = []
    const lines = text.split(/\r?\n/).map(l => l.trim()).filter(Boolean)
    if (lines.length === 0) return rows

    let startIdx = 0
    const first = lines[0]!.toLowerCase()
    if (first.startsWith('email,') || first === 'email' || first.startsWith('email ')) {
        startIdx = 1
    }

    for (let i = startIdx; i < lines.length; i++) {
        const line = lines[i]!
        const parts = line.split(',').map(p => p.trim().replace(/^"|"$/g, ''))
        const email = parts[0] ?? ''
        const name = parts[1] ?? ''
        if (email) rows.push({ email, name })
    }
    return rows
}

async function handleImport() {
    const rows = parseImportText(importText.value)
    if (rows.length === 0) {
        toast.error('Paste at least one email')
        return
    }
    importLoading.value = true
    importResult.value = null
    try {
        const params = new URLSearchParams({
            validate_mx: String(validateMX.value),
            validate_disposable: String(validateDisposable.value),
        })
        const res = await api<ImportResult>(
            `/projects/${props.project.id}/subscribers/import?${params.toString()}`,
            { method: 'POST', body: rows },
        )
        importResult.value = res
        if (res.imported > 0) {
            toast.success(`${res.imported} imported`)
            fetchSubscribers()
        } else {
            toast.error('No subscribers were imported')
        }
    } catch (e: any) {
        toast.error(e.message || 'Import failed')
    } finally {
        importLoading.value = false
    }
}

function downloadRejected() {
    if (!importResult.value || importResult.value.rejected.length === 0) return
    const header = 'email,name,reason\n'
    const body = importResult.value.rejected.map(r => {
        const safe = (s: string) => `"${s.replace(/"/g, '""')}"`
        return `${safe(r.email)},${safe(r.name)},${r.reason}`
    }).join('\n')
    const blob = new Blob([header + body], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `rejected-subscribers-${new Date().toISOString().slice(0, 10)}.csv`
    a.click()
    URL.revokeObjectURL(url)
}

function resetImport() {
    importText.value = ''
    importResult.value = null
    showImportModal.value = false
}

function readFile(file: File) {
    if (!/\.csv$|text\/csv/i.test(file.name + ' ' + file.type)) {
        toast.error('Pick a .csv file')
        return
    }
    const reader = new FileReader()
    reader.onload = e => { importText.value = String(e.target?.result ?? '') }
    reader.onerror = () => toast.error('Could not read file')
    reader.readAsText(file)
}

function handleFileUpload(event: Event) {
    const input = event.target as HTMLInputElement
    const file = input.files?.[0]
    if (file) readFile(file)
    input.value = ''
}

const isDragging = ref(false)

function handleFileDrop(event: DragEvent) {
    event.preventDefault()
    isDragging.value = false
    const file = event.dataTransfer?.files?.[0]
    if (file) readFile(file)
}

onMounted(fetchSubscribers)
</script>

<template>
    <div>
        <div class="flex flex-wrap items-center justify-between gap-3 mb-6">
            <div>
                <h1 class="text-2xl font-bold text-white">Subscribers</h1>
                <p class="text-sm text-zinc-500 mt-1">{{ total }} total</p>
            </div>
            <div class="flex flex-wrap items-center gap-2">
                <AppButton variant="ghost" size="sm" @click="showImportModal = true">Import CSV</AppButton>
                <AppButton size="sm" @click="showAddModal = true">+ Add Subscriber</AppButton>
            </div>
        </div>

        <div v-if="selectedIds.length > 0" class="bg-zinc-800 border border-zinc-700 rounded-lg p-3 mb-6 flex items-center justify-between shadow-lg">
            <span class="text-sm font-medium text-white px-2">{{ selectedIds.length }} selected</span>
            <div class="flex items-center gap-2">
                <select @change="(e) => handleBulkAction('update_status', (e.target as HTMLSelectElement).value)" class="text-sm bg-zinc-900 border border-zinc-700 rounded-md px-3 py-1.5 text-white focus:outline-none focus:ring-1 focus:ring-zinc-500">
                    <option value="" disabled selected>Change Status...</option>
                    <option value="active">Mark Active</option>
                    <option value="pending">Mark Pending</option>
                    <option value="unsubscribed">Mark Unsubscribed</option>
                </select>
                <button @click="confirmBulkDelete" :disabled="bulkLoading" class="text-sm bg-red-500/10 text-red-400 hover:bg-red-500/20 border border-red-500/20 rounded-md px-3 py-1.5 transition cursor-pointer disabled:opacity-50">
                    Delete
                </button>
            </div>
        </div>

        <div v-if="loading" class="text-zinc-500 py-8 text-center">Loading...</div>

        <div v-else-if="subscribers.length > 0" class="bg-zinc-900 border border-zinc-800 rounded-lg overflow-x-auto">
            <table class="w-full min-w-[640px]">
                <thead>
                    <tr class="border-b border-zinc-800">
                        <th class="px-4 py-3 w-10">
                            <AppCheckbox :modelValue="allSelected" @update:modelValue="toggleSelectAll" />
                        </th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Email</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Name</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Status</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Added</th>
                        <th class="text-right px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Actions</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="sub in subscribers" :key="sub.id" class="border-b border-zinc-800 last:border-0 hover:bg-zinc-800/50 transition" :class="{'bg-zinc-800/30': selectedIds.includes(sub.id)}">
                        <td class="px-4 py-3">
                            <AppCheckbox :modelValue="selectedIds.includes(sub.id)" @update:modelValue="(v: boolean) => toggleSelected(sub.id, v)" />
                        </td>
                        <td class="px-4 py-3 text-sm text-white">{{ sub.email }}</td>
                        <td class="px-4 py-3 text-sm text-zinc-400">{{ sub.name || '-' }}</td>
                        <td class="px-4 py-3">
                            <span :class="[
                                'text-xs px-2 py-1 rounded-full',
                                sub.status === 'active' && 'bg-green-500/10 text-green-400',
                                sub.status === 'unsubscribed' && 'bg-red-500/10 text-red-400',
                                sub.status === 'pending' && 'bg-yellow-500/10 text-yellow-400',
                            ]">
                                {{ sub.status }}
                            </span>
                        </td>
                        <td class="px-4 py-3 text-sm text-zinc-500">{{ new Date(sub.created_at).toLocaleDateString() }}</td>
                        <td class="px-4 py-3 text-right space-x-3">
                            <button @click="toggleStatus(sub)"
                                class="text-xs text-zinc-500 hover:text-white transition cursor-pointer">
                                {{ sub.status === 'active' ? 'Unsubscribe' : 'Activate' }}
                            </button>
                            <button @click="openDeleteModal(sub)"
                                class="text-xs text-zinc-500 hover:text-red-400 transition cursor-pointer">
                                Delete
                            </button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <div v-else class="bg-zinc-900 border border-zinc-800 rounded-lg p-8 text-center">
            <p class="text-zinc-400 mb-2">No subscribers yet.</p>
            <p class="text-zinc-500 text-sm">Add subscribers manually or collect them via the API.</p>
        </div>

        <div v-if="total > limit" class="flex items-center justify-between mt-4">
            <button @click="page--; fetchSubscribers()" :disabled="page === 0"
                class="text-sm text-zinc-400 hover:text-white disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer">
                Previous
            </button>
            <span class="text-sm text-zinc-500">Page {{ page + 1 }} of {{ Math.ceil(total / limit) }}</span>
            <button @click="page++; fetchSubscribers()" :disabled="(page + 1) * limit >= total"
                class="text-sm text-zinc-400 hover:text-white disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer">
                Next
            </button>
        </div>

        <AppModal :show="showAddModal" title="Add Subscriber" @close="showAddModal = false">
            <form @submit.prevent="handleAdd" class="space-y-4">
                <AppInput v-model="newEmail" label="Email" type="email" placeholder="subscriber@example.com" required />
                <AppInput v-model="newName" label="Name" placeholder="John Doe" />
                <AppButton :loading="addLoading">
                    {{ addLoading ? 'Adding...' : 'Add Subscriber' }}
                </AppButton>
            </form>
        </AppModal>

        <AppConfirmModal
            :show="showDeleteModal"
            title="Remove subscriber"
            :message="subscriberToDelete ? `Remove ${subscriberToDelete.email} from the list? This cannot be undone.` : ''"
            confirmLabel="Remove"
            danger
            :loading="deleteLoading"
            @confirm="handleDelete"
            @cancel="showDeleteModal = false" />

        <AppConfirmModal
            :show="showBulkDeleteConfirm"
            title="Delete subscribers"
            :message="`Delete ${selectedIds.length} subscriber${selectedIds.length === 1 ? '' : 's'} from the list? This cannot be undone.`"
            confirmLabel="Delete"
            danger
            :loading="bulkLoading"
            @confirm="handleBulkAction('delete')"
            @cancel="showBulkDeleteConfirm = false" />

        <AppModal :show="showImportModal" title="Import subscribers" size="lg" @close="resetImport">
            <div v-if="!importResult" class="space-y-4">
                <div>
                    <div class="flex items-center justify-between mb-1">
                        <label class="text-sm font-medium text-zinc-300">CSV (email, name)</label>
                        <label class="text-xs text-zinc-300 hover:text-white border border-zinc-700 rounded-md px-2 py-1 cursor-pointer transition hover:bg-zinc-800">
                            Choose file
                            <input type="file" accept=".csv,text/csv" class="hidden" @change="handleFileUpload" />
                        </label>
                    </div>
                    <div @dragover.prevent="isDragging = true"
                        @dragleave.prevent="isDragging = false"
                        @drop="handleFileDrop"
                        :class="[
                            'rounded-lg border transition',
                            isDragging ? 'border-white border-dashed bg-zinc-900' : 'border-zinc-800',
                        ]">
                        <textarea v-model="importText" rows="8" placeholder="email,name&#10;ada@example.com,Ada Lovelace&#10;alan@example.com,Alan Turing&#10;&#10;…or drop a .csv file here"
                            class="w-full px-3 py-2 bg-zinc-950 rounded-lg text-sm text-white font-mono placeholder-zinc-600 focus:outline-none focus:ring-2 focus:ring-zinc-500 transition resize-y" />
                    </div>
                    <p class="text-xs text-zinc-500 mt-1">First line can be a header (<code class="text-zinc-400">email,name</code>). Name column is optional. Drop a .csv file or pick one above.</p>
                </div>

                <div class="space-y-2">
                    <label class="flex items-start gap-2.5 p-2.5 rounded-lg border border-zinc-800 hover:border-zinc-700 cursor-pointer transition">
                        <span class="mt-0.5">
                            <AppCheckbox v-model="validateMX" />
                        </span>
                        <div class="min-w-0">
                            <p class="text-sm text-white">Reject addresses without MX records</p>
                            <p class="text-xs text-zinc-500 mt-0.5">DNS lookup per unique domain. Skips dead inboxes that would bounce on first send.</p>
                        </div>
                    </label>
                    <label class="flex items-start gap-2.5 p-2.5 rounded-lg border border-zinc-800 hover:border-zinc-700 cursor-pointer transition">
                        <span class="mt-0.5">
                            <AppCheckbox v-model="validateDisposable" />
                        </span>
                        <div class="min-w-0">
                            <p class="text-sm text-white">Reject disposable domains</p>
                            <p class="text-xs text-zinc-500 mt-0.5">Blocks Mailinator, 10MinuteMail, YopMail and similar single-use mailbox services.</p>
                        </div>
                    </label>
                </div>

                <div class="flex gap-2 pt-2">
                    <AppButton type="button" variant="ghost" size="sm" class="flex-1" @click="resetImport">Cancel</AppButton>
                    <AppButton size="sm" :loading="importLoading" :disabled="importLoading" class="flex-1" @click="handleImport">
                        {{ importLoading ? 'Importing...' : 'Import' }}
                    </AppButton>
                </div>
            </div>

            <div v-else class="space-y-4">
                <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-2">
                    <div class="bg-emerald-500/10 border border-emerald-500/30 rounded-lg p-3">
                        <p class="text-[11px] text-emerald-400/80 uppercase tracking-wide">Imported</p>
                        <p class="text-2xl font-bold text-emerald-300 tabular-nums">{{ importResult.imported }}</p>
                    </div>
                    <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-3">
                        <p class="text-[11px] text-zinc-500 uppercase tracking-wide">Duplicates</p>
                        <p class="text-2xl font-bold text-zinc-300 tabular-nums">{{ importResult.duplicates }}</p>
                    </div>
                    <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-3">
                        <p class="text-[11px] text-zinc-500 uppercase tracking-wide">Bad syntax</p>
                        <p class="text-2xl font-bold text-zinc-300 tabular-nums">{{ importResult.syntax_invalid }}</p>
                    </div>
                    <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-3">
                        <p class="text-[11px] text-zinc-500 uppercase tracking-wide">No MX</p>
                        <p class="text-2xl font-bold text-zinc-300 tabular-nums">{{ importResult.no_mx }}</p>
                    </div>
                    <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-3">
                        <p class="text-[11px] text-zinc-500 uppercase tracking-wide">Disposable</p>
                        <p class="text-2xl font-bold text-zinc-300 tabular-nums">{{ importResult.disposable }}</p>
                    </div>
                    <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-3">
                        <p class="text-[11px] text-zinc-500 uppercase tracking-wide">Suppressed</p>
                        <p class="text-2xl font-bold text-zinc-300 tabular-nums">{{ importResult.suppressed }}</p>
                    </div>
                </div>

                <div v-if="importResult.rejected.length > 0">
                    <div class="flex items-center justify-between mb-2">
                        <p class="text-sm font-medium text-white">{{ importResult.rejected.length }} rejected</p>
                        <button @click="downloadRejected"
                            class="px-3 py-1.5 text-xs text-zinc-300 border border-zinc-700 rounded-md hover:bg-zinc-800 transition cursor-pointer">
                            Download CSV
                        </button>
                    </div>
                    <div class="bg-zinc-950 border border-zinc-800 rounded-lg max-h-56 overflow-auto">
                        <table class="w-full text-xs">
                            <thead class="sticky top-0 bg-zinc-900 border-b border-zinc-800">
                                <tr>
                                    <th class="text-left px-3 py-2 font-medium text-zinc-400">Email</th>
                                    <th class="text-left px-3 py-2 font-medium text-zinc-400">Reason</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr v-for="row in importResult.rejected.slice(0, 100)" :key="row.email" class="border-b border-zinc-800/50 last:border-0">
                                    <td class="px-3 py-1.5 font-mono text-zinc-300 truncate max-w-xs">{{ row.email }}</td>
                                    <td class="px-3 py-1.5 text-zinc-500">{{ row.reason }}</td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                    <p v-if="importResult.rejected.length > 100" class="text-xs text-zinc-500 mt-1">
                        Showing first 100. Download CSV for the full list.
                    </p>
                </div>

                <AppButton @click="resetImport">Done</AppButton>
            </div>
        </AppModal>
    </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '@/api/client'
import { useToastStore } from '@/stores/toast'
import { useNewsletterStore } from '@/stores/newsletters'
import { useFieldStore } from '@/stores/fields'
import type { Project } from '@/stores/projects'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import AppModal from '@/components/ui/AppModal.vue'
import AppCheckbox from '@/components/ui/AppCheckbox.vue'
import AppConfirmModal from '@/components/ui/AppConfirmModal.vue'
import AppPagination from '@/components/ui/AppPagination.vue'
import SubscriberFieldInputs from '@/components/SubscriberFieldInputs.vue'
import AppTagInput from '@/components/ui/AppTagInput.vue'
import AppSelect from '@/components/ui/AppSelect.vue'
import AppStatusPill, { type PillTone } from '@/components/ui/AppStatusPill.vue'
import { validateFieldValues } from '@/utils/fieldValidation'

interface Subscriber {
    id: string
    email: string
    name: string
    status: string
    fields: Record<string, any>
    tags: string[]
    created_at: string
}

const props = defineProps<{ project: Project }>()
const toast = useToastStore()
const newsletterStore = useNewsletterStore()
const projectNewsletters = computed(() => newsletterStore.newsletters(props.project.id))

const STATUS_OPTIONS = [
    { value: '', label: 'All statuses' },
    { value: 'active', label: 'Active' },
    { value: 'pending', label: 'Pending' },
    { value: 'unsubscribed', label: 'Unsubscribed' },
]

const tagOptions = computed(() => [
    { value: '', label: 'All tags' },
    ...tagSuggestions.value.map(tag => ({ value: tag, label: tag })),
])

const newsletterOptions = computed(() => [
    { value: '', label: 'All newsletters' },
    ...projectNewsletters.value.map(n => ({ value: n.id, label: n.name })),
])

const bulkNewsletterOptions = computed(() => [
    { value: '', label: 'Pick a newsletter…' },
    ...projectNewsletters.value.map(n => ({ value: n.id, label: n.name })),
])

function statusTone(status: string): PillTone {
    switch (status) {
        case 'active': return 'emerald'
        case 'pending': return 'amber'
        case 'unsubscribed': return 'red'
        default: return 'zinc'
    }
}
const editNewsletters = ref<string[]>([])
const fieldStore = useFieldStore()

const fieldDefinitions = computed(() => fieldStore.fields(props.project.id))

const subscribers = ref<Subscriber[]>([])
const total = ref(0)
const loading = ref(true)

const filterStatus = ref('')
const filterTag = ref('')
const filterNewsletterId = ref('')
const hasFilters = computed(() => filterStatus.value !== '' || filterTag.value !== '' || filterNewsletterId.value !== '')

const tagSuggestions = ref<string[]>([])

const showAddModal = ref(false)
const newEmail = ref('')
const newName = ref('')
const newFields = ref<Record<string, any>>({})
const newTags = ref<string[]>([])
const addLoading = ref(false)
const addEmailError = ref('')
const addFieldErrors = ref<Record<string, string>>({})

const showEditModal = ref(false)
const editingSubscriber = ref<Subscriber | null>(null)
const editFields = ref<Record<string, any>>({})
const editTags = ref<string[]>([])
const editLoading = ref(false)
const editFieldErrors = ref<Record<string, string>>({})

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

function openAddModal() {
    newEmail.value = ''
    newName.value = ''
    newFields.value = {}
    newTags.value = []
    addEmailError.value = ''
    addFieldErrors.value = {}
    showAddModal.value = true
}

const showBulkTagModal = ref(false)
const bulkTags = ref<string[]>([])
const showBulkNewsletterModal = ref(false)
const bulkNewsletter = ref('')
const bulkNewsletterLoading = ref(false)
const bulkTagLoading = ref(false)

async function fetchTagSuggestions() {
    try {
        tagSuggestions.value = await api<string[]>(`/projects/${props.project.id}/tags`) || []
    } catch {
        tagSuggestions.value = []
    }
}

async function handleBulkTags(action: 'add_tags' | 'remove_tags') {
    if (bulkTags.value.length === 0) {
        toast.error('Add at least one tag')
        return
    }
    bulkTagLoading.value = true
    try {
        await api(`/projects/${props.project.id}/subscribers/bulk`, {
            method: 'POST',
            body: { action, tags: bulkTags.value, subscriber_ids: selectedIds.value },
        })
        toast.success('Tags updated')
        showBulkTagModal.value = false
        bulkTags.value = []
        selectedIds.value = []
        fetchSubscribers()
        fetchTagSuggestions()
    } catch (e: any) {
        toast.error(e.message || 'Failed to update tags')
    } finally {
        bulkTagLoading.value = false
    }
}

async function handleBulkNewsletter(action: 'add_newsletter' | 'remove_newsletter') {
    if (!bulkNewsletter.value) {
        toast.error('Pick a newsletter')
        return
    }
    bulkNewsletterLoading.value = true
    try {
        await api(`/projects/${props.project.id}/subscribers/bulk`, {
            method: 'POST',
            body: { action, newsletter_id: bulkNewsletter.value, subscriber_ids: selectedIds.value },
        })
        toast.success(action === 'add_newsletter' ? 'Added to newsletter' : 'Removed from newsletter')
        showBulkNewsletterModal.value = false
        bulkNewsletter.value = ''
        selectedIds.value = []
        fetchSubscribers()
        newsletterStore.fetchNewsletters(props.project.id)
    } catch (e: any) {
        toast.error(e.message || 'Failed to update newsletter membership')
    } finally {
        bulkNewsletterLoading.value = false
    }
}

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
const limit = ref(50)

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

function applyFilters() {
    page.value = 0
    fetchSubscribers()
}

function clearFilters() {
    filterStatus.value = ''
    filterTag.value = ''
    filterNewsletterId.value = ''
    applyFilters()
}

async function fetchSubscribers() {
    loading.value = true
    try {
        const params = new URLSearchParams({ limit: String(limit.value), offset: String(page.value * limit.value) })
        if (filterStatus.value) params.set('status', filterStatus.value)
        if (filterTag.value) params.set('tag', filterTag.value)
        if (filterNewsletterId.value) params.set('newsletter_id', filterNewsletterId.value)
        const res = await api<{ subscribers: Subscriber[] | null, total: number }>(
            `/projects/${props.project.id}/subscribers?${params.toString()}`
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
    addEmailError.value = ''
    if (!newEmail.value.trim()) {
        addEmailError.value = 'Email is required'
    } else if (!EMAIL_RE.test(newEmail.value.trim())) {
        addEmailError.value = 'Enter a valid email address'
    }
    addFieldErrors.value = validateFieldValues(fieldDefinitions.value, newFields.value)
    if (addEmailError.value || Object.keys(addFieldErrors.value).length > 0) {
        return
    }

    addLoading.value = true
    try {
        await api(`/projects/${props.project.id}/subscribers`, {
            method: 'POST',
            body: { email: newEmail.value.trim(), name: newName.value, fields: newFields.value, tags: newTags.value },
        })
        showAddModal.value = false
        toast.success('Subscriber added')
        fetchSubscribers()
        fetchTagSuggestions()
    } catch (e: any) {
        toast.error(e.message || 'Failed to add subscriber')
    } finally {
        addLoading.value = false
    }
}

function openEditModal(sub: Subscriber) {
    editingSubscriber.value = sub
    editFields.value = { ...sub.fields }
    editTags.value = [...(sub.tags ?? [])]
    editFieldErrors.value = {}
    editNewsletters.value = []
    newsletterStore.fetchNewsletters(props.project.id)
    newsletterStore.fetchSubscriberNewsletters(props.project.id, sub.id).then((memberships) => {
        editNewsletters.value = memberships.filter(m => !m.unsubscribed_at).map(m => m.id)
    })
    showEditModal.value = true
}

async function handleEditFields() {
    if (!editingSubscriber.value) return
    editFieldErrors.value = validateFieldValues(fieldDefinitions.value, editFields.value)
    if (Object.keys(editFieldErrors.value).length > 0) return
    editLoading.value = true
    const id = editingSubscriber.value.id
    try {
        await api(`/projects/${props.project.id}/subscribers/${id}`, {
            method: 'PATCH',
            body: { fields: editFields.value },
        })
        await api(`/projects/${props.project.id}/subscribers/${id}/tags`, {
            method: 'PUT',
            body: { tags: editTags.value },
        })
        if (projectNewsletters.value.length > 0) {
            await newsletterStore.setSubscriberNewsletters(props.project.id, id, editNewsletters.value)
        }
        showEditModal.value = false
        editingSubscriber.value = null
        toast.success('Subscriber updated')
        fetchSubscribers()
        fetchTagSuggestions()
    } catch (e: any) {
        toast.error(e.message || 'Failed to update subscriber')
    } finally {
        editLoading.value = false
    }
}

function formatFieldValue(value: any): string {
    if (value === undefined || value === null || value === '') return '-'
    if (typeof value === 'boolean') return value ? 'Yes' : 'No'
    return String(value)
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

interface ImportRow { email: string; name: string; fields: Record<string, any> }

function splitCsvLine(line: string): string[] {
    return line.split(',').map(p => p.trim().replace(/^"|"$/g, ''))
}

function resolveFieldKey(header: string): string | null {
    const normalized = header.trim().toLowerCase()
    for (const def of fieldDefinitions.value) {
        if (def.key.toLowerCase() === normalized || def.label.toLowerCase() === normalized) {
            return def.key
        }
    }
    return null
}

function coerceFieldValue(key: string, raw: string): any {
    const def = fieldDefinitions.value.find(d => d.key === key)
    if (!def) return raw
    if (def.field_type === 'number') {
        const parsed = Number(raw)
        return Number.isNaN(parsed) ? raw : parsed
    }
    if (def.field_type === 'boolean') {
        return ['true', '1', 'yes', 'y'].includes(raw.toLowerCase())
    }
    return raw
}

function parseImportText(text: string): ImportRow[] {
    const rows: ImportRow[] = []
    const lines = text.split(/\r?\n/).map(l => l.trim()).filter(Boolean)
    if (lines.length === 0) return rows

    let startIdx = 0
    const columnKeys: (string | null)[] = []
    const first = lines[0]!.toLowerCase()
    const hasHeader = first.startsWith('email,') || first === 'email' || first.startsWith('email ')
    if (hasHeader) {
        startIdx = 1
        const headers = splitCsvLine(lines[0]!)
        headers.forEach((header, idx) => {
            if (idx === 0 || idx === 1) {
                columnKeys.push(null)
            } else {
                columnKeys.push(resolveFieldKey(header))
            }
        })
    }

    for (let i = startIdx; i < lines.length; i++) {
        const parts = splitCsvLine(lines[i]!)
        const email = parts[0] ?? ''
        const name = parts[1] ?? ''
        if (!email) continue
        const fields: Record<string, any> = {}
        for (let c = 2; c < parts.length; c++) {
            const key = columnKeys[c]
            const value = parts[c]
            if (key && value !== undefined && value !== '') {
                fields[key] = coerceFieldValue(key, value)
            }
        }
        rows.push({ email, name, fields })
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

onMounted(() => {
    newsletterStore.fetchNewsletters(props.project.id)
    fetchSubscribers()
    fieldStore.fetchFields(props.project.id)
    fetchTagSuggestions()
})
</script>

<template>
    <div>
        <div class="flex flex-wrap items-center justify-between gap-3 mb-6">
            <div>
                <h1 class="text-xl font-semibold text-white">Subscribers</h1>
                <p class="text-sm text-zinc-400 mt-1">{{ total }} total</p>
            </div>
            <div class="flex flex-wrap items-center gap-2">
                <AppButton variant="ghost" size="md" @click="showImportModal = true">Import CSV</AppButton>
                <AppButton size="md" @click="openAddModal">+ Add Subscriber</AppButton>
            </div>
        </div>

        <div class="flex flex-wrap items-end gap-3 mb-6">
            <div>
                <label class="block text-xs text-zinc-400 mb-1">Status</label>
                <AppSelect v-model="filterStatus" size="sm" :options="STATUS_OPTIONS" @change="applyFilters" />
            </div>
            <div v-if="tagSuggestions.length">
                <label class="block text-xs text-zinc-400 mb-1">Tag</label>
                <AppSelect v-model="filterTag" size="sm" class="max-w-[200px] truncate" :options="tagOptions" @change="applyFilters" />
            </div>
            <div v-if="projectNewsletters.length">
                <label class="block text-xs text-zinc-400 mb-1">Newsletter</label>
                <AppSelect v-model="filterNewsletterId" size="sm" class="max-w-[200px] truncate"
                    :options="newsletterOptions" @change="applyFilters" />
            </div>
            <button v-if="hasFilters" @click="clearFilters"
                class="px-3 py-1.5 text-sm text-zinc-400 hover:text-white transition cursor-pointer">
                Clear
            </button>
        </div>

        <div v-if="selectedIds.length > 0" class="bg-zinc-850 border border-zinc-700 rounded-lg p-3 mb-6 flex items-center justify-between shadow-lg">
            <span class="text-sm font-medium text-white px-2">{{ selectedIds.length }} selected</span>
            <div class="flex items-center gap-2">
                <select @change="(e) => handleBulkAction('update_status', (e.target as HTMLSelectElement).value)" class="text-sm bg-zinc-900 border border-zinc-700 rounded-md px-3 py-1.5 text-white focus:outline-none focus:ring-1 focus:ring-emerald-500">
                    <option value="" disabled selected>Change Status...</option>
                    <option value="active">Mark Active</option>
                    <option value="pending">Mark Pending</option>
                    <option value="unsubscribed">Mark Unsubscribed</option>
                </select>
                <AppButton variant="outline" size="sm" @click="showBulkTagModal = true">Tags</AppButton>
                <AppButton v-if="projectNewsletters.length > 0" variant="outline" size="sm"
                    @click="showBulkNewsletterModal = true; newsletterStore.fetchNewsletters(props.project.id)">Newsletter</AppButton>
                <AppButton variant="danger-outline" size="sm" :loading="bulkLoading" :disabled="bulkLoading" @click="confirmBulkDelete">
                    Delete
                </AppButton>
            </div>
        </div>

        <div v-if="loading" class="text-zinc-400 py-8 text-center">Loading...</div>

        <div v-else-if="subscribers.length > 0" class="bg-zinc-900 border border-zinc-800 rounded-lg overflow-x-auto">
            <table class="w-full min-w-[640px]">
                <thead>
                    <tr class="border-b border-zinc-800">
                        <th class="px-4 py-3 w-10">
                            <AppCheckbox :modelValue="allSelected" @update:modelValue="toggleSelectAll" />
                        </th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">Email</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">Name</th>
                        <th v-for="def in fieldDefinitions" :key="def.id" class="text-left px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">{{ def.label }}</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">Tags</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">Status</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">Added</th>
                        <th class="text-right px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">Actions</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="sub in subscribers" :key="sub.id" class="border-b border-zinc-800 last:border-0 hover:bg-zinc-850/50 transition" :class="{'bg-emerald-500/8 shadow-[inset_2px_0_0_var(--color-emerald-500)]': selectedIds.includes(sub.id)}">
                        <td class="px-4 py-3">
                            <AppCheckbox :modelValue="selectedIds.includes(sub.id)" @update:modelValue="(v: boolean) => toggleSelected(sub.id, v)" />
                        </td>
                        <td class="px-4 py-3 text-sm text-white">{{ sub.email }}</td>
                        <td class="px-4 py-3 text-sm text-zinc-300">{{ sub.name || '-' }}</td>
                        <td v-for="def in fieldDefinitions" :key="def.id" class="px-4 py-3 text-sm text-zinc-300">{{ formatFieldValue(sub.fields?.[def.key]) }}</td>
                        <td class="px-4 py-3">
                            <div v-if="sub.tags?.length" class="flex flex-wrap gap-1">
                                <span v-for="tag in sub.tags" :key="tag" class="text-xs bg-zinc-850 text-zinc-300 px-2 py-0.5 rounded border border-zinc-700">{{ tag }}</span>
                            </div>
                            <span v-else class="text-sm text-zinc-600">-</span>
                        </td>
                        <td class="px-4 py-3">
                            <AppStatusPill :tone="statusTone(sub.status)" :label="sub.status" />
                        </td>
                        <td class="px-4 py-3 text-sm text-zinc-400">{{ new Date(sub.created_at).toLocaleDateString() }}</td>
                        <td class="px-4 py-3 text-right space-x-3">
                            <button @click="openEditModal(sub)"
                                class="text-xs text-zinc-400 hover:text-white transition cursor-pointer">
                                Edit
                            </button>
                            <button @click="toggleStatus(sub)"
                                class="text-xs text-zinc-400 hover:text-white transition cursor-pointer">
                                {{ sub.status === 'active' ? 'Unsubscribe' : 'Activate' }}
                            </button>
                            <button @click="openDeleteModal(sub)"
                                class="text-xs text-zinc-400 hover:text-red-400 transition cursor-pointer">
                                Delete
                            </button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <div v-else-if="hasFilters" class="bg-zinc-900 border border-zinc-800 rounded-lg p-8 text-center">
            <p class="text-zinc-300 mb-2">No subscribers match these filters.</p>
            <button @click="clearFilters" class="text-sm text-zinc-400 hover:text-white transition cursor-pointer underline">Clear filters</button>
        </div>

        <div v-else class="bg-zinc-900 border border-zinc-800 rounded-lg p-8 text-center">
            <p class="text-zinc-300 mb-2">No subscribers yet.</p>
            <p class="text-zinc-400 text-sm">Add subscribers manually or collect them via the API.</p>
        </div>

        <AppPagination
            v-model:page="page"
            v-model:limit="limit"
            :total="total"
            @change="fetchSubscribers" />

        <AppModal :show="showAddModal" title="Add Subscriber" @close="showAddModal = false">
            <form @submit.prevent="handleAdd" class="space-y-4" novalidate>
                <AppInput v-model="newEmail" label="Email" type="email" placeholder="subscriber@example.com" :error="addEmailError" />
                <AppInput v-model="newName" label="Name" placeholder="John Doe" />
                <SubscriberFieldInputs v-model="newFields" :definitions="fieldDefinitions" :errors="addFieldErrors" />
                <AppTagInput v-model="newTags" label="Tags" :suggestions="tagSuggestions" />
                <AppButton :loading="addLoading">
                    {{ addLoading ? 'Adding...' : 'Add Subscriber' }}
                </AppButton>
            </form>
        </AppModal>

        <AppModal :show="showEditModal" :title="editingSubscriber ? `Edit ${editingSubscriber.email}` : 'Edit subscriber'" @close="showEditModal = false">
            <form @submit.prevent="handleEditFields" class="space-y-4" novalidate>
                <SubscriberFieldInputs v-model="editFields" :definitions="fieldDefinitions" :errors="editFieldErrors" />
                <AppTagInput v-model="editTags" label="Tags" :suggestions="tagSuggestions" />
                <div v-if="projectNewsletters.length > 0">
                    <label class="block text-sm font-medium text-zinc-300 mb-2">Newsletters</label>
                    <div class="space-y-2">
                        <label v-for="newsletter in projectNewsletters" :key="newsletter.id" class="flex items-center gap-2 text-sm text-zinc-300 cursor-pointer">
                            <input type="checkbox" :value="newsletter.id" v-model="editNewsletters"
                                class="rounded border-zinc-700 bg-zinc-900 text-white focus:ring-emerald-500" />
                            {{ newsletter.name }}
                        </label>
                    </div>
                    <p class="text-xs text-zinc-400 mt-1">Checking a newsletter re-subscribes the reader if they had opted out of it.</p>
                </div>
                <AppButton :loading="editLoading">
                    {{ editLoading ? 'Saving...' : 'Save' }}
                </AppButton>
            </form>
        </AppModal>

        <AppModal :show="showBulkTagModal" title="Tag selected subscribers" @close="showBulkTagModal = false">
            <div class="space-y-4">
                <p class="text-sm text-zinc-300">{{ selectedIds.length }} subscriber(s) selected.</p>
                <AppTagInput v-model="bulkTags" label="Tags" :suggestions="tagSuggestions" />
                <div class="flex gap-2">
                    <AppButton variant="ghost" size="sm" class="flex-1" :loading="bulkTagLoading" @click="handleBulkTags('remove_tags')">Remove</AppButton>
                    <AppButton size="sm" class="flex-1" :loading="bulkTagLoading" @click="handleBulkTags('add_tags')">Add</AppButton>
                </div>
            </div>
        </AppModal>

        <AppModal :show="showBulkNewsletterModal" title="Newsletter membership" @close="showBulkNewsletterModal = false">
            <div class="space-y-4">
                <p class="text-sm text-zinc-300">{{ selectedIds.length }} subscriber(s) selected.</p>
                <div>
                    <label class="block text-sm font-medium text-zinc-300 mb-1">Newsletter</label>
                    <AppSelect v-model="bulkNewsletter" size="md" class="w-full" :options="bulkNewsletterOptions" />
                </div>
                <div class="flex gap-2">
                    <AppButton variant="ghost" size="sm" class="flex-1" :loading="bulkNewsletterLoading" @click="handleBulkNewsletter('remove_newsletter')">Remove</AppButton>
                    <AppButton size="sm" class="flex-1" :loading="bulkNewsletterLoading" @click="handleBulkNewsletter('add_newsletter')">Add</AppButton>
                </div>
            </div>
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
                        <label class="text-xs text-zinc-300 hover:text-white border border-zinc-700 rounded-md px-2 py-1 cursor-pointer transition hover:bg-zinc-850">
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
                            class="w-full px-3 py-2 bg-zinc-900 rounded-lg text-sm text-white font-mono placeholder-zinc-600 focus:outline-none focus:ring-2 focus:ring-emerald-500 transition resize-y" />
                    </div>
                    <p class="text-xs text-zinc-400 mt-1">First line can be a header (<code class="text-zinc-300">email,name</code>). Name column is optional. Extra columns whose header matches a custom field key or label are imported into that field. Drop a .csv file or pick one above.</p>
                </div>

                <div class="space-y-2">
                    <label class="flex items-start gap-2.5 p-2.5 rounded-lg border border-zinc-800 hover:border-zinc-700 cursor-pointer transition">
                        <span class="mt-0.5">
                            <AppCheckbox v-model="validateMX" />
                        </span>
                        <div class="min-w-0">
                            <p class="text-sm text-white">Reject addresses without MX records</p>
                            <p class="text-xs text-zinc-400 mt-0.5">DNS lookup per unique domain. Skips dead inboxes that would bounce on first send.</p>
                        </div>
                    </label>
                    <label class="flex items-start gap-2.5 p-2.5 rounded-lg border border-zinc-800 hover:border-zinc-700 cursor-pointer transition">
                        <span class="mt-0.5">
                            <AppCheckbox v-model="validateDisposable" />
                        </span>
                        <div class="min-w-0">
                            <p class="text-sm text-white">Reject disposable domains</p>
                            <p class="text-xs text-zinc-400 mt-0.5">Blocks Mailinator, 10MinuteMail, YopMail and similar single-use mailbox services.</p>
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
                        <p class="text-[11px] text-zinc-400 uppercase tracking-wide">Duplicates</p>
                        <p class="text-2xl font-bold text-zinc-300 tabular-nums">{{ importResult.duplicates }}</p>
                    </div>
                    <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-3">
                        <p class="text-[11px] text-zinc-400 uppercase tracking-wide">Bad syntax</p>
                        <p class="text-2xl font-bold text-zinc-300 tabular-nums">{{ importResult.syntax_invalid }}</p>
                    </div>
                    <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-3">
                        <p class="text-[11px] text-zinc-400 uppercase tracking-wide">No MX</p>
                        <p class="text-2xl font-bold text-zinc-300 tabular-nums">{{ importResult.no_mx }}</p>
                    </div>
                    <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-3">
                        <p class="text-[11px] text-zinc-400 uppercase tracking-wide">Disposable</p>
                        <p class="text-2xl font-bold text-zinc-300 tabular-nums">{{ importResult.disposable }}</p>
                    </div>
                    <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-3">
                        <p class="text-[11px] text-zinc-400 uppercase tracking-wide">Suppressed</p>
                        <p class="text-2xl font-bold text-zinc-300 tabular-nums">{{ importResult.suppressed }}</p>
                    </div>
                </div>

                <div v-if="importResult.rejected.length > 0">
                    <div class="flex items-center justify-between mb-2">
                        <p class="text-sm font-medium text-white">{{ importResult.rejected.length }} rejected</p>
                        <button @click="downloadRejected"
                            class="px-3 py-1.5 text-xs text-zinc-300 border border-zinc-700 rounded-md hover:bg-zinc-850 transition cursor-pointer">
                            Download CSV
                        </button>
                    </div>
                    <div class="bg-zinc-950 border border-zinc-800 rounded-lg max-h-56 overflow-auto">
                        <table class="w-full text-xs">
                            <thead class="sticky top-0 bg-zinc-900 border-b border-zinc-800">
                                <tr>
                                    <th class="text-left px-3 py-2 font-medium text-zinc-300">Email</th>
                                    <th class="text-left px-3 py-2 font-medium text-zinc-300">Reason</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr v-for="row in importResult.rejected.slice(0, 100)" :key="row.email" class="border-b border-zinc-800/50 last:border-0">
                                    <td class="px-3 py-1.5 font-mono text-zinc-300 truncate max-w-xs">{{ row.email }}</td>
                                    <td class="px-3 py-1.5 text-zinc-400">{{ row.reason }}</td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                    <p v-if="importResult.rejected.length > 100" class="text-xs text-zinc-400 mt-1">
                        Showing first 100. Download CSV for the full list.
                    </p>
                </div>

                <AppButton @click="resetImport">Done</AppButton>
            </div>
        </AppModal>
    </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '@/api/client'
import { useToastStore } from '@/stores/toast'
import type { Project } from '@/stores/projects'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import AppModal from '@/components/ui/AppModal.vue'

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

const page = ref(0)
const limit = 50

const selectedIds = ref<string[]>([])

const allSelected = computed(() => {
    return subscribers.value.length > 0 && selectedIds.value.length === subscribers.value.length
})

function toggleSelectAll(event: Event) {
    const checked = (event.target as HTMLInputElement).checked
    if (checked) {
        selectedIds.value = subscribers.value.map(s => s.id)
    } else {
        selectedIds.value = []
    }
}

const bulkLoading = ref(false)

async function handleBulkAction(action: 'delete' | 'update_status', status?: string) {
    if (action === 'delete' && !confirm(`Are you sure you want to delete ${selectedIds.value.length} subscribers?`)) return
    
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

onMounted(fetchSubscribers)
</script>

<template>
    <div>
        <div class="flex items-center justify-between mb-6">
            <div>
                <h1 class="text-2xl font-bold text-white">Subscribers</h1>
                <p class="text-sm text-zinc-500 mt-1">{{ total }} total</p>
            </div>
            <AppButton @click="showAddModal = true" class="w-auto! px-4">+ Add Subscriber</AppButton>
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
                <button @click="handleBulkAction('delete')" :disabled="bulkLoading" class="text-sm bg-red-500/10 text-red-400 hover:bg-red-500/20 border border-red-500/20 rounded-md px-3 py-1.5 transition cursor-pointer disabled:opacity-50">
                    Delete
                </button>
            </div>
        </div>

        <div v-if="loading" class="text-zinc-500 py-8 text-center">Loading...</div>

        <div v-else-if="subscribers.length > 0" class="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
            <table class="w-full">
                <thead>
                    <tr class="border-b border-zinc-800">
                        <th class="px-4 py-3 w-10">
                            <input type="checkbox" :checked="allSelected" @change="toggleSelectAll" class="appearance-none w-[18px] h-[18px] border-2 border-zinc-600 rounded bg-transparent checked:border-white relative cursor-pointer focus:outline-none transition-colors checked:after:content-[''] checked:after:absolute checked:after:inset-[3px] checked:after:bg-white checked:after:rounded-sm hover:border-zinc-400" />
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
                            <input type="checkbox" :value="sub.id" v-model="selectedIds" class="appearance-none w-[18px] h-[18px] border-2 border-zinc-600 rounded bg-transparent checked:border-white relative cursor-pointer focus:outline-none transition-colors checked:after:content-[''] checked:after:absolute checked:after:inset-[3px] checked:after:bg-white checked:after:rounded-sm hover:border-zinc-400" />
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

        <AppModal :show="showDeleteModal" title="Remove Subscriber" @close="showDeleteModal = false">
            <div class="space-y-4">
                <p class="text-zinc-400 text-sm">
                    Are you sure you want to remove
                    <span class="font-semibold text-white">{{ subscriberToDelete?.email }}</span>?
                </p>
                <div class="flex gap-3">
                    <AppButton variant="secondary" @click="showDeleteModal = false">Cancel</AppButton>
                    <AppButton variant="danger" :loading="deleteLoading" @click="handleDelete">
                        {{ deleteLoading ? 'Removing...' : 'Remove' }}
                    </AppButton>
                </div>
            </div>
        </AppModal>
    </div>
</template>

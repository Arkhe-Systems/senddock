<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api/client'
import { useToastStore } from '@/stores/toast'
import type { Project } from '@/stores/projects'
import AppButton from '@/components/ui/AppButton.vue'
import AppModal from '@/components/ui/AppModal.vue'
import AppInput from '@/components/ui/AppInput.vue'

interface Template {
    id: string
    name: string
    subject: string
}

interface Campaign {
    id: string
    name: string
    template_id: string
    status: string
    scheduled_at: string
    sent_at: string | null
    sent_count: number
    failed_count: number
}

const props = defineProps<{ project: Project }>()
const toast = useToastStore()

const campaigns = ref<Campaign[]>([])
const templates = ref<Template[]>([])
const loading = ref(true)

const showCreateModal = ref(false)
const createLoading = ref(false)

const newName = ref('')
const selectedTemplate = ref('')
const sendType = ref<'now' | 'scheduled'>('now')
const scheduledDate = ref('')
const scheduledTime = ref('')

async function loadData() {
    loading.value = true
    try {
        const [campRes, tempRes] = await Promise.all([
            api<Campaign[] | null>(`/projects/${props.project.id}/campaigns`),
            api<Template[] | null>(`/projects/${props.project.id}/templates`),
        ])
        campaigns.value = campRes || []
        templates.value = tempRes || []
    } catch {
        toast.error('Failed to load campaigns')
    } finally {
        loading.value = false
    }
}

function openCreateModal() {
    if (templates.value.length === 0) {
        toast.error('You need to create a template first')
        return
    }
    newName.value = ''
    selectedTemplate.value = templates.value[0]?.id ?? ''
    sendType.value = 'now'
    
    const tomorrow = new Date()
    tomorrow.setDate(tomorrow.getDate() + 1)
    scheduledDate.value = tomorrow.toISOString().split('T')[0] ?? ''
    scheduledTime.value = '09:00'
    
    showCreateModal.value = true
}

async function handleCreate() {
    if (!newName.value) {
        toast.error('Name is required')
        return
    }
    
    let finalScheduledAt: string
    if (sendType.value === 'now') {
        finalScheduledAt = new Date().toISOString()
    } else {
        if (!scheduledDate.value || !scheduledTime.value) {
            toast.error('Please select a date and time')
            return
        }
        const combined = new Date(`${scheduledDate.value}T${scheduledTime.value}:00`)
        if (combined < new Date()) {
            toast.error('Scheduled time must be in the future')
            return
        }
        finalScheduledAt = combined.toISOString()
    }

    createLoading.value = true
    try {
        await api(`/projects/${props.project.id}/campaigns`, {
            method: 'POST',
            body: {
                name: newName.value,
                template_id: selectedTemplate.value,
                scheduled_at: finalScheduledAt
            }
        })
        toast.success('Campaign scheduled successfully')
        showCreateModal.value = false
        loadData()
    } catch (e: any) {
        toast.error(e.message || 'Failed to schedule campaign')
    } finally {
        createLoading.value = false
    }
}

async function handleDelete(id: string) {
    if (!confirm('Are you sure you want to delete this scheduled campaign?')) return
    try {
        await api(`/projects/${props.project.id}/campaigns/${id}`, {
            method: 'DELETE'
        })
        toast.success('Campaign deleted')
        loadData()
    } catch (e: any) {
        toast.error(e.message || 'Failed to delete campaign. It might have already started sending.')
    }
}

onMounted(loadData)
</script>

<template>
    <div>
        <div class="flex items-center justify-between mb-6">
            <div>
                <h1 class="text-2xl font-bold text-white">Newsletters</h1>
                <p class="text-sm text-zinc-400 mt-1">Schedule and send email campaigns to your subscribers.</p>
            </div>
            <button @click="openCreateModal"
                class="px-4 py-2 text-sm font-medium bg-white text-zinc-950 rounded-lg hover:bg-zinc-200 transition cursor-pointer">
                + New Campaign
            </button>
        </div>

        <div v-if="loading" class="text-zinc-500 py-8 text-center">Loading...</div>

        <div v-else-if="campaigns.length > 0" class="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
            <table class="w-full">
                <thead>
                    <tr class="border-b border-zinc-800">
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Name</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Status</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Scheduled For</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Sent / Failed</th>
                        <th class="text-right px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Actions</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="c in campaigns" :key="c.id" class="border-b border-zinc-800 last:border-0 group">
                        <td class="px-4 py-3 text-sm text-white font-medium">{{ c.name }}</td>
                        <td class="px-4 py-3">
                            <span :class="[
                                'text-xs px-2 py-1 rounded-full',
                                c.status === 'scheduled' && 'bg-blue-500/10 text-blue-400',
                                c.status === 'sending' && 'bg-yellow-500/10 text-yellow-400',
                                c.status === 'sent' && 'bg-green-500/10 text-green-400',
                                c.status === 'failed' && 'bg-red-500/10 text-red-400',
                            ]">
                                {{ c.status }}
                            </span>
                        </td>
                        <td class="px-4 py-3 text-sm text-zinc-400">
                            {{ new Date(c.scheduled_at).toLocaleString() }}
                        </td>
                        <td class="px-4 py-3 text-sm text-zinc-400">
                            <span class="text-green-400">{{ c.sent_count }}</span> / 
                            <span :class="c.failed_count > 0 ? 'text-red-400' : 'text-zinc-500'">{{ c.failed_count }}</span>
                        </td>
                        <td class="px-4 py-3 text-right">
                            <button v-if="c.status === 'scheduled'" @click="handleDelete(c.id)"
                                class="text-xs text-zinc-500 hover:text-red-400 transition cursor-pointer opacity-0 group-hover:opacity-100">
                                Delete
                            </button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <div v-else class="text-center py-20 border border-dashed border-zinc-800 rounded-lg">
            <p class="text-zinc-500 mb-4">No newsletters scheduled yet.</p>
            <button @click="openCreateModal"
                class="px-6 py-2 text-sm font-medium bg-white text-zinc-950 rounded-lg hover:bg-zinc-200 transition cursor-pointer">
                Create your first campaign
            </button>
        </div>

        <AppModal :show="showCreateModal" title="New Newsletter Campaign" @close="showCreateModal = false">
            <form @submit.prevent="handleCreate" class="space-y-4">
                <AppInput v-model="newName" label="Campaign Name" placeholder="Monthly Update - May" required />
                
                <div>
                    <label class="block text-sm font-medium text-zinc-300 mb-1">Template</label>
                    <select v-model="selectedTemplate"
                        class="w-full px-3 py-2 bg-zinc-900 border border-zinc-800 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-zinc-500 focus:border-transparent">
                        <option v-for="t in templates" :key="t.id" :value="t.id">{{ t.name }}</option>
                    </select>
                </div>

                <div>
                    <label class="block text-sm font-medium text-zinc-300 mb-2">When to send?</label>
                    <div class="flex gap-2 mb-3">
                        <button type="button" @click="sendType = 'now'"
                            :class="['px-3 py-1.5 text-sm rounded-lg transition cursor-pointer', sendType === 'now' ? 'bg-zinc-800 text-white' : 'text-zinc-500 hover:text-white']">
                            Send Now
                        </button>
                        <button type="button" @click="sendType = 'scheduled'"
                            :class="['px-3 py-1.5 text-sm rounded-lg transition cursor-pointer', sendType === 'scheduled' ? 'bg-zinc-800 text-white' : 'text-zinc-500 hover:text-white']">
                            Schedule for later
                        </button>
                    </div>

                    <div v-if="sendType === 'scheduled'" class="flex gap-2 mt-2">
                        <div class="flex-1">
                            <input type="date" v-model="scheduledDate" required
                                class="w-full px-3 py-2 bg-zinc-900 border border-zinc-800 rounded-lg text-white placeholder-zinc-500 focus:outline-none focus:ring-2 focus:ring-zinc-500 focus:border-transparent" />
                        </div>
                        <div class="w-32">
                            <input type="time" v-model="scheduledTime" required
                                class="w-full px-3 py-2 bg-zinc-900 border border-zinc-800 rounded-lg text-white placeholder-zinc-500 focus:outline-none focus:ring-2 focus:ring-zinc-500 focus:border-transparent" />
                        </div>
                    </div>
                </div>

                <div class="pt-2">
                    <AppButton :loading="createLoading" class="w-full">
                        {{ createLoading ? 'Scheduling...' : (sendType === 'now' ? 'Send Campaign Now' : 'Schedule Campaign') }}
                    </AppButton>
                </div>
            </form>
        </AppModal>
    </div>
</template>

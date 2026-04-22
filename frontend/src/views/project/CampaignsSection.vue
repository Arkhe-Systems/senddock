<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
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
    html_body: string
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
    variables: Record<string, string>
}

const props = defineProps<{ project: Project }>()
const toast = useToastStore()

const campaigns = ref<Campaign[]>([])
const templates = ref<Template[]>([])
const loading = ref(true)

const showCreateModal = ref(false)
const createLoading = ref(false)
const editingCampaign = ref<Campaign | null>(null)

const newName = ref('')
const selectedTemplate = ref('')
const sendType = ref<'now' | 'scheduled'>('now')
const scheduledDate = ref('')
const scheduledTime = ref('')
const campaignVars = ref<Record<string, string>>({})

const SYSTEM_VARS = new Set(['name', 'email', 'subscriber_id', 'unsubscribe_url'])

const selectedTemplateVars = computed(() => {
    const tmpl = templates.value.find(t => t.id === selectedTemplate.value)
    if (!tmpl) return []
    const text = tmpl.html_body + ' ' + tmpl.subject
    const regex = /\{\{\s*([a-zA-Z0-9_]+)\s*\}\}/g
    const matches = Array.from(text.matchAll(regex)).map(m => m[1] as string).filter(Boolean)
    return [...new Set(matches)]
})

const customTemplateVars = computed(() =>
    selectedTemplateVars.value.filter(v => !SYSTEM_VARS.has(v))
)

watch(selectedTemplateVars, (newVars) => {
    if (!editingCampaign.value) {
        campaignVars.value = {}
        newVars.filter(v => !SYSTEM_VARS.has(v)).forEach(v => { campaignVars.value[v] = '' })
    }
}, { deep: true, immediate: true })

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
    editingCampaign.value = null
    newName.value = ''
    selectedTemplate.value = templates.value[0]?.id ?? ''
    sendType.value = 'now'
    
    const tomorrow = new Date()
    tomorrow.setDate(tomorrow.getDate() + 1)
    scheduledDate.value = tomorrow.toISOString().split('T')[0] ?? ''
    scheduledTime.value = '09:00'
    campaignVars.value = {}
    
    showCreateModal.value = true
}

function openEditModal(c: Campaign) {
    editingCampaign.value = c
    newName.value = c.name
    selectedTemplate.value = c.template_id
    sendType.value = 'scheduled'
    
    const date = new Date(c.scheduled_at)
    scheduledDate.value = date.toISOString().split('T')[0] ?? ''
    scheduledTime.value = date.toTimeString().split(' ')[0]?.slice(0, 5) ?? '09:00'
    
    campaignVars.value = c.variables ? { ...c.variables } : {}
    
    showCreateModal.value = true
}

async function handleSave() {
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
        if (combined < new Date() && !editingCampaign.value) {
            toast.error('Scheduled time must be in the future')
            return
        }
        finalScheduledAt = combined.toISOString()
    }

    createLoading.value = true
    try {
        if (editingCampaign.value) {
            await api(`/projects/${props.project.id}/campaigns/${editingCampaign.value.id}`, {
                method: 'PATCH',
                body: {
                    name: newName.value,
                    template_id: selectedTemplate.value,
                    scheduled_at: finalScheduledAt,
                    variables: campaignVars.value
                }
            })
            toast.success('Campaign updated')
        } else {
            await api(`/projects/${props.project.id}/campaigns`, {
                method: 'POST',
                body: {
                    name: newName.value,
                    template_id: selectedTemplate.value,
                    scheduled_at: finalScheduledAt,
                    variables: campaignVars.value
                }
            })
            toast.success('Campaign scheduled')
        }
        showCreateModal.value = false
        loadData()
    } catch (e: any) {
        toast.error(e.message || 'Failed to save campaign')
    } finally {
        createLoading.value = false
    }
}

const showDeleteModal = ref(false)
const campaignToDelete = ref<Campaign | null>(null)
const deleteLoading = ref(false)

function confirmDelete(c: Campaign) {
    campaignToDelete.value = c
    showDeleteModal.value = true
}

async function handleDelete() {
    if (!campaignToDelete.value) return
    deleteLoading.value = true
    try {
        await api(`/projects/${props.project.id}/campaigns/${campaignToDelete.value.id}`, {
            method: 'DELETE'
        })
        toast.success('Campaign deleted')
        showDeleteModal.value = false
        loadData()
    } catch (e: any) {
        toast.error(e.message || 'Failed to delete campaign.')
    } finally {
        deleteLoading.value = false
    }
}

function varLabel(v: string | undefined) {
    return '{{' + (v ?? '') + '}}'
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
                        <td class="px-4 py-3 text-right space-x-3">
                            <button v-if="c.status === 'scheduled'" @click="openEditModal(c)"
                                class="text-xs text-zinc-500 hover:text-white transition cursor-pointer opacity-0 group-hover:opacity-100">
                                Edit
                            </button>
                            <button v-if="c.status === 'scheduled'" @click="confirmDelete(c)"
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

        <AppModal :show="showCreateModal" :title="editingCampaign ? 'Edit Campaign' : 'New Newsletter Campaign'" @close="showCreateModal = false">
            <form @submit.prevent="handleSave" class="space-y-4">
                <AppInput v-model="newName" label="Campaign Name" placeholder="Monthly Update - May" required />
                
                <div>
                    <label class="block text-sm font-medium text-zinc-300 mb-1">Template</label>
                    <select v-model="selectedTemplate"
                        class="w-full px-3 py-2 bg-zinc-900 border border-zinc-800 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-zinc-500 focus:border-transparent">
                        <option v-for="t in templates" :key="t.id" :value="t.id">{{ t.name }}</option>
                    </select>
                </div>

                <div v-if="selectedTemplateVars.length > 0" class="p-3 bg-zinc-900 border border-zinc-800 rounded-lg space-y-3">
                    <p class="text-xs font-medium text-zinc-400">Template Variables</p>

                    <div v-if="customTemplateVars.length > 0" class="space-y-2">
                        <p class="text-xs text-zinc-500">Fill in the custom values for this send:</p>
                        <div v-for="v in customTemplateVars" :key="v">
                            <AppInput v-model="campaignVars[v]" :label="v" :placeholder="'Value for {{' + v + '}}'" />
                        </div>
                    </div>

                    <div>
                        <p class="text-xs text-zinc-500 mb-1">Auto-filled per subscriber:</p>
                        <div class="flex flex-wrap gap-1">
                            <span v-for="v in selectedTemplateVars.filter(v => SYSTEM_VARS.has(v))" :key="v"
                                class="text-xs bg-zinc-800 text-zinc-400 px-2 py-0.5 rounded border border-zinc-700 font-mono">
                                {{ varLabel(v) }}
                            </span>
                        </div>
                    </div>
                </div>

                <div v-else-if="selectedTemplate" class="p-3 bg-zinc-900 border border-zinc-800 rounded-lg">
                    <p class="text-xs text-zinc-500">This template has no variables detected.</p>
                </div>

                <div>
                    <label class="block text-sm font-medium text-zinc-300 mb-2">When to send?</label>
                    <div class="flex gap-2 mb-3">
                        <button v-if="!editingCampaign" type="button" @click="sendType = 'now'"
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
                        {{ createLoading ? 'Saving...' : (editingCampaign ? 'Save Changes' : (sendType === 'now' ? 'Send Campaign Now' : 'Schedule Campaign')) }}
                    </AppButton>
                </div>
            </form>
        </AppModal>

        <AppModal :show="showDeleteModal" title="Delete Campaign" @close="showDeleteModal = false">
            <div class="space-y-4">
                <p class="text-zinc-400 text-sm">
                    Are you sure you want to delete <span class="font-semibold text-white">{{ campaignToDelete?.name }}</span>?
                </p>
                <p class="text-zinc-500 text-xs">
                    This will permanently cancel the scheduled send.
                </p>
                <div class="flex gap-2 justify-end mt-4">
                    <AppButton variant="secondary" @click="showDeleteModal = false">Cancel</AppButton>
                    <AppButton variant="danger" :loading="deleteLoading" @click="handleDelete">
                        {{ deleteLoading ? 'Deleting...' : 'Delete Campaign' }}
                    </AppButton>
                </div>
            </div>
        </AppModal>
    </div>
</template>

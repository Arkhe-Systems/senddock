<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, computed, watch } from 'vue'
import { api } from '@/api/client'
import { useToastStore } from '@/stores/toast'
import { useAppStore } from '@/stores/app'
import { useNewsletterStore } from '@/stores/newsletters'
import type { Project } from '@/stores/projects'
import AppButton from '@/components/ui/AppButton.vue'
import AppFilterChip from '@/components/ui/AppFilterChip.vue'
import AppModal from '@/components/ui/AppModal.vue'
import AppInput from '@/components/ui/AppInput.vue'
import AppConfirmModal from '@/components/ui/AppConfirmModal.vue'

interface Template {
    id: string
    name: string
    subject: string
    html_body: string
    type?: string
}

interface Campaign {
    id: string
    name: string
    template_id: string
    subject: string
    status: string
    scheduled_at: string
    sent_at: string | null
    sent_count: number
    failed_count: number
    variables: Record<string, string>
    newsletter_id: string | null
}

const props = defineProps<{ project: Project }>()
const toast = useToastStore()
const appStore = useAppStore()
const newsletterStore = useNewsletterStore()

const campaigns = ref<Campaign[]>([])
const templates = ref<Template[]>([])
const loading = ref(true)

const showCreateModal = ref(false)
const createLoading = ref(false)
const editingCampaign = ref<Campaign | null>(null)

const newName = ref('')
const selectedTemplate = ref('')
const subjectOverride = ref('')
const sendType = ref<'now' | 'scheduled'>('now')
const scheduledDate = ref('')
const scheduledTime = ref('')
const campaignVars = ref<Record<string, string>>({})
const selectedNewsletter = ref('')
const projectNewsletters = computed(() => newsletterStore.newsletters(props.project.id))

const selectedTemplateData = computed(() => templates.value.find(t => t.id === selectedTemplate.value))
const templateHasSubject = computed(() => !!(selectedTemplateData.value?.subject?.trim()))

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
        templates.value = (tempRes || []).filter(t => (t.type ?? 'email') === 'email')
    } catch {
        toast.error('Failed to load campaigns')
    } finally {
        loading.value = false
    }
}

function openCreateModal() {
    if (!appStore.publicUrlIsReachable) {
        toast.error('Configure PUBLIC_URL with your public domain before scheduling campaigns. Recipients need a working unsubscribe link.')
        return
    }
    if (templates.value.length === 0) {
        toast.error('You need to create a template first')
        return
    }
    editingCampaign.value = null
    newName.value = ''
    selectedTemplate.value = templates.value[0]?.id ?? ''
    subjectOverride.value = ''
    sendType.value = 'now'

    const tomorrow = new Date()
    tomorrow.setDate(tomorrow.getDate() + 1)
    scheduledDate.value = tomorrow.toISOString().split('T')[0] ?? ''
    scheduledTime.value = '09:00'
    campaignVars.value = {}
    selectedNewsletter.value = ''
    newsletterStore.fetchNewsletters(props.project.id)

    showCreateModal.value = true
}

function openEditModal(c: Campaign) {
    editingCampaign.value = c
    newName.value = c.name
    selectedTemplate.value = c.template_id
    subjectOverride.value = c.subject || ''
    sendType.value = 'scheduled'

    const date = new Date(c.scheduled_at)
    scheduledDate.value = date.toISOString().split('T')[0] ?? ''
    scheduledTime.value = date.toTimeString().split(' ')[0]?.slice(0, 5) ?? '09:00'

    campaignVars.value = c.variables ? { ...c.variables } : {}
    selectedNewsletter.value = c.newsletter_id || ''
    newsletterStore.fetchNewsletters(props.project.id)

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
                    subject: subjectOverride.value,
                    scheduled_at: finalScheduledAt,
                    variables: campaignVars.value,
                    newsletter_id: selectedNewsletter.value
                }
            })
            toast.success('Campaign updated')
        } else {
            await api(`/projects/${props.project.id}/campaigns`, {
                method: 'POST',
                body: {
                    name: newName.value,
                    template_id: selectedTemplate.value,
                    subject: subjectOverride.value,
                    scheduled_at: finalScheduledAt,
                    variables: campaignVars.value,
                    newsletter_id: selectedNewsletter.value
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

const deleteMessage = computed(() => {
    const c = campaignToDelete.value
    if (!c) return ''
    switch (c.status) {
        case 'scheduled':
            return `Delete "${c.name}"? This permanently cancels the scheduled send.`
        case 'sending':
            return `Delete "${c.name}"? This row is marked as sending — emails already in flight will continue to be delivered, but the campaign will disappear from this list.`
        case 'sent':
            return `Delete "${c.name}"? Removes it from this list. The emails already sent and the broadcast history are kept.`
        case 'failed':
            return `Delete "${c.name}"? Removes the failed campaign from this list.`
        default:
            return `Delete "${c.name}"?`
    }
})

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

let pollTimer: ReturnType<typeof setInterval> | null = null

function startPollingIfNeeded() {
    if (pollTimer) return
    if (!campaigns.value.some(c => c.status === 'sending')) return
    pollTimer = setInterval(async () => {
        if (!campaigns.value.some(c => c.status === 'sending')) {
            stopPolling()
            return
        }
        try {
            const res = await api<Campaign[] | null>(`/projects/${props.project.id}/campaigns`)
            campaigns.value = res || []
        } catch {
            // poll keeps running; next tick will retry
        }
    }, 5000)
}

function stopPolling() {
    if (pollTimer) {
        clearInterval(pollTimer)
        pollTimer = null
    }
}

watch(campaigns, () => startPollingIfNeeded(), { deep: false })
onBeforeUnmount(() => stopPolling())

onMounted(loadData)
</script>

<template>
    <div>
        <div class="flex flex-wrap items-center justify-between gap-3 mb-6">
            <div>
                <h1 class="text-xl font-semibold text-white">Campaigns</h1>
                <p class="text-sm text-zinc-300 mt-1">Schedule and send email campaigns to your subscribers.</p>
            </div>
            <AppButton size="md" @click="openCreateModal" :disabled="!appStore.publicUrlIsReachable"
                :title="!appStore.publicUrlIsReachable ? 'Set PUBLIC_URL to a public domain in your .env to enable campaigns' : ''"
                class="disabled:opacity-40">
                + New Campaign
            </AppButton>
        </div>

        <div v-if="!appStore.publicUrlIsReachable" class="bg-amber-500/5 border border-amber-500/30 text-amber-300 rounded-lg p-4 mb-6 text-sm">
            <p class="font-medium mb-1">Campaigns are disabled</p>
            <p class="text-amber-300/80">Your instance is reachable only at localhost, so unsubscribe links inside emails would not work for recipients. Set <code class="font-mono">PUBLIC_URL</code> to your public domain in <code class="font-mono">.env</code> and restart the server to enable campaigns.</p>
        </div>

        <div v-if="loading" class="text-zinc-400 py-8 text-center">Loading...</div>

        <div v-else-if="campaigns.length > 0" class="bg-zinc-900 border border-zinc-800 rounded-lg overflow-x-auto">
            <table class="w-full min-w-[640px]">
                <thead>
                    <tr class="border-b border-zinc-800">
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">Name</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">Status</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">Scheduled For</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">Sent / Failed</th>
                        <th class="text-right px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">Actions</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="c in campaigns" :key="c.id" class="border-b border-zinc-800 last:border-0 group">
                        <td class="px-4 py-3 text-sm text-white font-medium">{{ c.name }}</td>
                        <td class="px-4 py-3">
                            <span :class="[
                                'text-xs px-2 py-1 rounded-full',
                                c.status === 'scheduled' && 'bg-blue-500/10 text-blue-400',
                                c.status === 'sending' && 'bg-amber-500/10 text-amber-400',
                                c.status === 'sent' && 'bg-emerald-500/10 text-emerald-400',
                                c.status === 'failed' && 'bg-red-500/10 text-red-400',
                            ]">
                                {{ c.status }}
                            </span>
                        </td>
                        <td class="px-4 py-3 text-sm text-zinc-300">
                            {{ new Date(c.scheduled_at).toLocaleString() }}
                        </td>
                        <td class="px-4 py-3 text-sm text-zinc-300">
                            <span class="text-emerald-400">{{ c.sent_count }}</span> / 
                            <span :class="c.failed_count > 0 ? 'text-red-400' : 'text-zinc-400'">{{ c.failed_count }}</span>
                        </td>
                        <td class="px-4 py-3 text-right space-x-3 whitespace-nowrap">
                            <button v-if="c.status === 'scheduled'" @click="openEditModal(c)"
                                class="text-xs text-zinc-300 hover:text-white transition cursor-pointer">
                                Edit
                            </button>
                            <button @click="confirmDelete(c)"
                                class="text-xs text-zinc-300 hover:text-red-400 transition cursor-pointer">
                                Delete
                            </button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <div v-else class="text-center py-20 border border-dashed border-zinc-800 rounded-lg">
            <p class="text-zinc-400 mb-4">No campaigns scheduled yet.</p>
            <AppButton size="lg" @click="openCreateModal"
                class="">
                Create your first campaign
            </AppButton>
        </div>

        <AppModal :show="showCreateModal" :title="editingCampaign ? 'Edit Campaign' : 'New Campaign'" @close="showCreateModal = false">
            <form @submit.prevent="handleSave" class="space-y-4">
                <AppInput v-model="newName" label="Campaign Name" placeholder="Monthly Update - May" required />

                <div>
                    <label class="block text-sm font-medium text-zinc-300 mb-1">Template</label>
                    <select v-model="selectedTemplate"
                        class="w-full px-3 py-2 bg-zinc-900 border border-zinc-800 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent">
                        <option v-for="t in templates" :key="t.id" :value="t.id">{{ t.name }}</option>
                    </select>
                </div>

                <div v-if="projectNewsletters.length > 0">
                    <label class="block text-sm font-medium text-zinc-300 mb-1">Newsletter <span class="text-zinc-400 font-normal">(optional)</span></label>
                    <select v-model="selectedNewsletter"
                        class="w-full px-3 py-2 bg-zinc-900 border border-zinc-800 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent">
                        <option value="">No newsletter — whole audience</option>
                        <option v-for="newsletter in projectNewsletters" :key="newsletter.id" :value="newsletter.id">{{ newsletter.name }} ({{ newsletter.active_count }})</option>
                    </select>
                    <p class="text-xs text-zinc-400 mt-1">Targets active members of the newsletter; their unsubscribe link leaves only this newsletter.</p>
                </div>

                <div>
                    <label class="block text-sm font-medium text-zinc-300 mb-1">
                        Email Subject
                        <span v-if="templateHasSubject" class="text-zinc-400 font-normal">(override the template's subject — optional)</span>
                    </label>
                    <input v-model="subjectOverride" type="text" :placeholder="templateHasSubject ? selectedTemplateData?.subject : 'Set a subject for this campaign'"
                        class="w-full px-3 py-2 bg-zinc-900 border border-zinc-800 rounded-lg text-white placeholder-zinc-500 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent" />
                    <p v-if="!templateHasSubject && !subjectOverride.trim()" class="text-xs text-amber-400 mt-1">
                        This template has no subject. Sending campaigns without a subject is a strong spam signal — most providers will route them to the spam folder.
                    </p>
                </div>

                <div v-if="selectedTemplateVars.length > 0" class="p-3 bg-zinc-900 border border-zinc-800 rounded-lg space-y-3">
                    <p class="text-xs font-medium text-zinc-300">Template Variables</p>

                    <div v-if="customTemplateVars.length > 0" class="space-y-2">
                        <p class="text-xs text-zinc-400">Fill in the custom values for this send:</p>
                        <div v-for="v in customTemplateVars" :key="v">
                            <AppInput v-model="campaignVars[v]" :label="v" :placeholder="'Value for {{' + v + '}}'" />
                        </div>
                    </div>

                    <div>
                        <p class="text-xs text-zinc-400 mb-1">Auto-filled per subscriber:</p>
                        <div class="flex flex-wrap gap-1">
                            <span v-for="v in selectedTemplateVars.filter(v => SYSTEM_VARS.has(v))" :key="v"
                                class="text-xs bg-zinc-850 text-zinc-300 px-2 py-0.5 rounded border border-zinc-700 font-mono">
                                {{ varLabel(v) }}
                            </span>
                        </div>
                    </div>
                </div>

                <div v-else-if="selectedTemplate" class="p-3 bg-zinc-900 border border-zinc-800 rounded-lg">
                    <p class="text-xs text-zinc-400">This template has no variables detected.</p>
                </div>

                <div>
                    <label class="block text-sm font-medium text-zinc-300 mb-2">When to send?</label>
                    <div class="flex gap-2 mb-3">
                        <AppFilterChip v-if="!editingCampaign" size="sm" :active="sendType === 'now'" @click="sendType = 'now'">
                            Send Now
                        </AppFilterChip>
                        <AppFilterChip size="sm" :active="sendType === 'scheduled'" @click="sendType = 'scheduled'">
                            Schedule for later
                        </AppFilterChip>
                    </div>

                    <div v-if="sendType === 'scheduled'" class="flex gap-2 mt-2">
                        <div class="flex-1">
                            <input type="date" v-model="scheduledDate" required
                                class="w-full px-3 py-2 bg-zinc-900 border border-zinc-800 rounded-lg text-white placeholder-zinc-500 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent" />
                        </div>
                        <div class="w-32">
                            <input type="time" v-model="scheduledTime" required
                                class="w-full px-3 py-2 bg-zinc-900 border border-zinc-800 rounded-lg text-white placeholder-zinc-500 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent" />
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

        <AppConfirmModal
            :show="showDeleteModal"
            title="Delete campaign"
            :message="deleteMessage"
            confirmLabel="Delete"
            danger
            :loading="deleteLoading"
            @confirm="handleDelete"
            @cancel="showDeleteModal = false" />
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import AppInput from '@/components/ui/AppInput.vue'
import RichTextEditor from '@/components/ui/RichTextEditor.vue'
import { Type, WandSparkles } from 'lucide-vue-next'
import { api } from '@/api/client'
import { useToastStore } from '@/stores/toast'
import { useAppStore } from '@/stores/app'
import { useSegmentStore } from '@/stores/segments'
import type { Project } from '@/stores/projects'
import AppButton from '@/components/ui/AppButton.vue'
import AppModal from '@/components/ui/AppModal.vue'

interface EmailStats {
    total: number
    sent: number
    failed: number
}

interface EmailLog {
    id: string
    to_email: string
    subject: string
    status: string
    error: string | null
    sent_at: string
}

interface Template {
    id: string
    name: string
    subject: string
    html_body: string
}

const props = defineProps<{ project: Project }>()
const toast = useToastStore()
const appStore = useAppStore()
const segmentStore = useSegmentStore()

const segments = computed(() => segmentStore.segments(props.project.id))
const selectedSegment = ref('')

const stats = ref<EmailStats>({ total: 0, sent: 0, failed: 0 })
const recentLogs = ref<EmailLog[]>([])
const loading = ref(true)

const showSendModal = ref(false)
const templates = ref<Template[]>([])
const selectedTemplate = ref('')
const sendMode = ref<'broadcast' | 'direct'>('broadcast')
const directEmail = ref('')
const subjectOverride = ref('')
const sendLoading = ref(false)
const openSendLoading = ref(false)
const templateVars = ref<Record<string, string>>({})
const richFields = ref<Record<string, boolean>>({})

const selectedTemplateData = computed(() => templates.value.find(t => t.id === selectedTemplate.value))
const templateHasSubject = computed(() => !!(selectedTemplateData.value?.subject?.trim()))
const effectiveSubject = computed(() => subjectOverride.value.trim() || selectedTemplateData.value?.subject?.trim() || '')

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

watch(selectedTemplate, () => {
    templateVars.value = {}
    richFields.value = {}
    selectedTemplateVars.value.filter(v => !SYSTEM_VARS.has(v)).forEach(v => { templateVars.value[v] = '' })
})

const htmlFields = computed(() => customTemplateVars.value.filter(v => richFields.value[v]))

async function loadData() {
    try {
        const [statsRes, logsRes] = await Promise.all([
            api<EmailStats>(`/projects/${props.project.id}/stats`),
            api<{ logs: EmailLog[] | null, total: number }>(`/projects/${props.project.id}/logs?limit=10`),
        ])
        stats.value = statsRes
        recentLogs.value = logsRes.logs || []
    } catch {
    } finally {
        loading.value = false
    }
}

async function openSendModal() {
    if (openSendLoading.value) return
    openSendLoading.value = true
    try {
        const res = await api<Template[] | null>(`/projects/${props.project.id}/templates`)
        templates.value = res || []
        if (templates.value.length === 0) {
            toast.error('Create a template first')
            return
        }
        selectedTemplate.value = templates.value[0]?.id ?? ''
        sendMode.value = appStore.publicUrlIsReachable ? 'broadcast' : 'direct'
        directEmail.value = ''
        subjectOverride.value = ''
        selectedSegment.value = ''
        segmentStore.fetchSegments(props.project.id)
        showSendModal.value = true
    } catch {
        toast.error('Failed to load templates')
    } finally {
        openSendLoading.value = false
    }
}

function varLabel(v: string | undefined) {
    return '{{' + (v ?? '') + '}}'
}

async function handleSend() {
    if (!selectedTemplate.value) {
        toast.error('Select a template')
        return
    }

    if (sendMode.value === 'broadcast' && !appStore.publicUrlIsReachable) {
        toast.error('Configure PUBLIC_URL with your public domain before sending broadcasts. Recipients need a working unsubscribe link.')
        return
    }

    sendLoading.value = true
    try {
        if (sendMode.value === 'broadcast') {
            const result = await api<{ sent: number, failed: number }>(`/projects/${props.project.id}/broadcast`, {
                method: 'POST',
                body: { template_id: selectedTemplate.value, subject: subjectOverride.value, variables: templateVars.value, html_fields: htmlFields.value, segment_id: selectedSegment.value },
            })
            toast.success(`Broadcast complete: ${result.sent} sent, ${result.failed} failed`)
        } else {
            if (!directEmail.value) {
                toast.error('Enter an email address')
                sendLoading.value = false
                return
            }
            await api(`/projects/${props.project.id}/send`, {
                method: 'POST',
                body: { template_id: selectedTemplate.value, to: directEmail.value, subject: subjectOverride.value, data: templateVars.value, html_fields: htmlFields.value },
            })
            toast.success(`Email sent to ${directEmail.value}`)
        }
        showSendModal.value = false
        loadData()
    } catch (e: any) {
        toast.error(e.message || 'Failed to send')
    } finally {
        sendLoading.value = false
    }
}

onMounted(loadData)
</script>

<template>
    <div>
        <div class="flex flex-wrap items-center justify-between gap-3 mb-6">
            <h1 class="text-2xl font-bold text-white">Overview</h1>
            <button v-if="project.smtp_host" @click="openSendModal" :disabled="openSendLoading"
                class="px-4 py-2 text-sm font-medium bg-white text-zinc-950 rounded-lg hover:bg-zinc-200 transition cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed">
                {{ openSendLoading ? 'Loading…' : 'Send Email' }}
            </button>
        </div>

        <div v-if="loading" class="text-zinc-500 py-8 text-center">Loading...</div>

        <div v-else>
            <div class="grid grid-cols-3 gap-4 mb-8">
                <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                    <p class="text-sm text-zinc-400">Total Emails</p>
                    <p class="text-2xl font-bold text-white mt-1">{{ stats.total }}</p>
                </div>
                <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                    <p class="text-sm text-zinc-400">Sent</p>
                    <p class="text-2xl font-bold text-green-400 mt-1">{{ stats.sent }}</p>
                </div>
                <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                    <p class="text-sm text-zinc-400">Failed</p>
                    <p class="text-2xl font-bold mt-1" :class="stats.failed > 0 ? 'text-red-400' : 'text-white'">{{ stats.failed }}</p>
                </div>
            </div>

            <div v-if="!project.smtp_host" class="bg-zinc-900 border border-zinc-800 rounded-lg p-6 mb-8">
                <p class="text-zinc-400 text-sm">Configure your SMTP settings to start sending emails.</p>
            </div>

            <div v-if="recentLogs.length > 0">
                <h2 class="text-lg font-semibold text-white mb-4">Recent Activity</h2>
                <div class="bg-zinc-900 border border-zinc-800 rounded-lg overflow-x-auto">
                    <table class="w-full min-w-[640px]">
                        <thead>
                            <tr class="border-b border-zinc-800">
                                <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">To</th>
                                <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Subject</th>
                                <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Status</th>
                                <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Date</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="log in recentLogs" :key="log.id" class="border-b border-zinc-800 last:border-0 group">
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
                            <tr v-for="log in recentLogs" v-if="false" :key="'err-'+log.id"></tr>
                        </tbody>
                    </table>
                    <div v-for="log in recentLogs.filter(l => l.status === 'failed' && l.error)" :key="'detail-'+log.id"
                        class="px-4 py-2 border-t border-zinc-800 bg-red-500/5">
                        <p class="text-xs text-red-400">
                            <span class="text-zinc-500">{{ log.to_email }}:</span> {{ log.error }}
                        </p>
                    </div>
                </div>
            </div>
        </div>

        <AppModal :show="showSendModal" title="Send Email" @close="showSendModal = false">
            <div class="space-y-4">
                <div>
                    <label class="block text-sm font-medium text-zinc-300 mb-1">Template</label>
                    <select v-model="selectedTemplate"
                        class="w-full px-3 py-2 bg-zinc-900 border border-zinc-800 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-zinc-500 focus:border-transparent">
                        <option v-for="t in templates" :key="t.id" :value="t.id">{{ t.name }}</option>
                    </select>
                </div>

                <div>
                    <label class="block text-sm font-medium text-zinc-300 mb-1">
                        Subject
                        <span v-if="templateHasSubject" class="text-zinc-500 font-normal">(override the template's subject — optional)</span>
                    </label>
                    <input v-model="subjectOverride" type="text" :placeholder="templateHasSubject ? selectedTemplateData?.subject : 'Set a subject'"
                        class="w-full px-3 py-2 bg-zinc-900 border border-zinc-800 rounded-lg text-white placeholder-zinc-500 focus:outline-none focus:ring-2 focus:ring-zinc-500 focus:border-transparent" />
                    <p v-if="!templateHasSubject && !subjectOverride.trim()" class="text-xs text-yellow-400 mt-1">
                        This template has no subject. Sending without one is a strong spam signal — most providers will route it to the spam folder.
                    </p>
                </div>

                <div v-if="selectedTemplateVars.length > 0" class="p-3 bg-zinc-900 border border-zinc-800 rounded-lg space-y-3">
                    <p class="text-xs font-medium text-zinc-400">Template Variables</p>

                    <div v-if="customTemplateVars.length > 0" class="space-y-3">
                        <p class="text-xs text-zinc-500">Fill in the custom values for this send:</p>
                        <div v-for="v in customTemplateVars" :key="v" class="space-y-1">
                            <div class="flex items-center justify-between gap-2">
                                <label class="text-sm font-medium text-zinc-300 font-mono">{{ varLabel(v) }}</label>
                                <div class="flex gap-0.5 bg-zinc-950/60 rounded-md p-0.5 border border-zinc-800">
                                    <button type="button" @click="richFields[v] = false" title="Plain text"
                                        :class="['flex items-center gap-1 px-2 py-0.5 text-xs rounded cursor-pointer transition', !richFields[v] ? 'bg-zinc-700 text-white' : 'text-zinc-500 hover:text-white']">
                                        <Type class="w-3 h-3" /> Text
                                    </button>
                                    <button type="button" @click="richFields[v] = true" title="Rich text (bold, lists, links…)"
                                        :class="['flex items-center gap-1 px-2 py-0.5 text-xs rounded cursor-pointer transition', richFields[v] ? 'bg-zinc-700 text-white' : 'text-zinc-500 hover:text-white']">
                                        <WandSparkles class="w-3 h-3" /> Rich
                                    </button>
                                </div>
                            </div>
                            <RichTextEditor v-if="richFields[v]" v-model="templateVars[v]" />
                            <AppInput v-else v-model="templateVars[v]" :placeholder="'Value for {{' + v + '}}'" />
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
                    <label class="block text-sm font-medium text-zinc-300 mb-2">Send to</label>
                    <div class="flex gap-2 mb-3">
                        <button @click="sendMode = 'broadcast'" :disabled="!appStore.publicUrlIsReachable"
                            :title="!appStore.publicUrlIsReachable ? 'Set PUBLIC_URL to a public domain to enable broadcasts' : ''"
                            :class="['px-3 py-1.5 text-sm rounded-lg transition cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed', sendMode === 'broadcast' ? 'bg-zinc-800 text-white' : 'text-zinc-500 hover:text-white']">
                            All subscribers
                        </button>
                        <button @click="sendMode = 'direct'"
                            :class="['px-3 py-1.5 text-sm rounded-lg transition cursor-pointer', sendMode === 'direct' ? 'bg-zinc-800 text-white' : 'text-zinc-500 hover:text-white']">
                            Specific email
                        </button>
                    </div>
                    <input v-if="sendMode === 'direct'" v-model="directEmail" type="email" placeholder="user@example.com"
                        class="w-full px-3 py-2 bg-zinc-900 border border-zinc-800 rounded-lg text-white placeholder-zinc-500 focus:outline-none focus:ring-2 focus:ring-zinc-500 focus:border-transparent" />
                    <div v-else>
                        <select v-model="selectedSegment"
                            class="w-full px-3 py-2 bg-zinc-900 border border-zinc-800 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-zinc-500 transition">
                            <option value="">All active subscribers</option>
                            <option v-for="segment in segments" :key="segment.id" :value="segment.id">{{ segment.name }}</option>
                        </select>
                        <p class="text-xs text-zinc-500 mt-1">
                            {{ selectedSegment ? 'Sends to active subscribers matching this segment.' : 'Sends to all active subscribers in this project.' }}
                        </p>
                    </div>
                </div>

                <AppButton :loading="sendLoading" @click="handleSend">
                    {{ sendLoading ? 'Sending...' : sendMode === 'broadcast' ? (selectedSegment ? 'Send to Segment' : 'Send to All') : 'Send' }}
                </AppButton>
            </div>
        </AppModal>
    </div>
</template>

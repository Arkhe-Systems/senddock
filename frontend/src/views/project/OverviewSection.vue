<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api/client'
import { useToastStore } from '@/stores/toast'
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
}

const props = defineProps<{ project: Project }>()
const toast = useToastStore()

const stats = ref<EmailStats>({ total: 0, sent: 0, failed: 0 })
const recentLogs = ref<EmailLog[]>([])
const loading = ref(true)

const showSendModal = ref(false)
const templates = ref<Template[]>([])
const selectedTemplate = ref('')
const sendMode = ref<'broadcast' | 'direct'>('broadcast')
const directEmail = ref('')
const sendLoading = ref(false)

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
    try {
        const res = await api<Template[] | null>(`/projects/${props.project.id}/templates`)
        templates.value = res || []
        if (templates.value.length === 0) {
            toast.error('Create a template first')
            return
        }
        selectedTemplate.value = templates.value[0].id
        sendMode.value = 'broadcast'
        directEmail.value = ''
        showSendModal.value = true
    } catch {
        toast.error('Failed to load templates')
    }
}

async function handleSend() {
    if (!selectedTemplate.value) {
        toast.error('Select a template')
        return
    }

    sendLoading.value = true
    try {
        if (sendMode.value === 'broadcast') {
            const result = await api<{ sent: number, failed: number }>(`/projects/${props.project.id}/broadcast`, {
                method: 'POST',
                body: { template_id: selectedTemplate.value },
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
                body: { template_id: selectedTemplate.value, to: directEmail.value },
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
        <div class="flex items-center justify-between mb-6">
            <h1 class="text-2xl font-bold text-white">Overview</h1>
            <button v-if="project.smtp_host" @click="openSendModal"
                class="px-4 py-2 text-sm font-medium bg-white text-zinc-950 rounded-lg hover:bg-zinc-200 transition cursor-pointer">
                Send Email
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
                <div class="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
                    <table class="w-full">
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
                    <label class="block text-sm font-medium text-zinc-300 mb-2">Send to</label>
                    <div class="flex gap-2 mb-3">
                        <button @click="sendMode = 'broadcast'"
                            :class="['px-3 py-1.5 text-sm rounded-lg transition cursor-pointer', sendMode === 'broadcast' ? 'bg-zinc-800 text-white' : 'text-zinc-500 hover:text-white']">
                            All subscribers
                        </button>
                        <button @click="sendMode = 'direct'"
                            :class="['px-3 py-1.5 text-sm rounded-lg transition cursor-pointer', sendMode === 'direct' ? 'bg-zinc-800 text-white' : 'text-zinc-500 hover:text-white']">
                            Specific email
                        </button>
                    </div>
                    <input v-if="sendMode === 'direct'" v-model="directEmail" type="email" placeholder="user@example.com"
                        class="w-full px-3 py-2 bg-zinc-900 border border-zinc-800 rounded-lg text-white placeholder-zinc-500 focus:outline-none focus:ring-2 focus:ring-zinc-500 focus:border-transparent" />
                    <p v-else class="text-xs text-zinc-500">Sends to all active subscribers in this project.</p>
                </div>

                <AppButton :loading="sendLoading" @click="handleSend">
                    {{ sendLoading ? 'Sending...' : sendMode === 'broadcast' ? 'Send to All' : 'Send' }}
                </AppButton>
            </div>
        </AppModal>
    </div>
</template>

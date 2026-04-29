<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { api } from '@/api/client'
import { useToastStore } from '@/stores/toast'
import type { Project } from '@/stores/projects'
import AppInput from '@/components/ui/AppInput.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppModal from '@/components/ui/AppModal.vue'

interface Webhook {
    id: string
    url: string
    secret: string
    events: string[]
    active: boolean
    created_at: string
}

interface Delivery {
    id: string
    event_type: string
    status: string
    attempts: number
    last_status_code?: number
    last_error?: string
    next_attempt_at?: string
    delivered_at?: string
    created_at: string
}

const ALL_EVENTS = [
    { value: 'email.sent', label: 'Email sent', description: 'Fired when an email is successfully delivered to the SMTP relay' },
    { value: 'email.failed', label: 'Email failed', description: 'Fired when sending an email returns an error' },
    { value: 'email.opened', label: 'Email opened', description: 'Fired the first time a recipient opens an email' },
    { value: 'email.clicked', label: 'Email clicked', description: 'Fired the first time a recipient clicks a tracked link' },
    { value: 'subscriber.created', label: 'Subscriber created', description: 'Fired when a new subscriber is added' },
    { value: 'subscriber.unsubscribed', label: 'Subscriber unsubscribed', description: 'Fired when a subscriber opts out' },
]

const props = defineProps<{ project: Project }>()
const toast = useToastStore()

const webhooks = ref<Webhook[]>([])
const loading = ref(true)
const errorState = ref<'none' | 'paywall' | 'generic'>('none')

const showCreate = ref(false)
const creating = ref(false)
const newUrl = ref('')
const newEvents = ref<string[]>(ALL_EVENTS.map(e => e.value))

const createdWebhook = ref<Webhook | null>(null)
const secretCopied = ref(false)

const deliveriesFor = ref<Webhook | null>(null)
const deliveries = ref<Delivery[]>([])
const deliveriesLoading = ref(false)

const togglingId = ref<string | null>(null)
const deletingId = ref<string | null>(null)

const sortedWebhooks = computed(() =>
    [...webhooks.value].sort((a, b) => b.created_at.localeCompare(a.created_at))
)

async function load() {
    loading.value = true
    errorState.value = 'none'
    try {
        const res = await api<{ webhooks: Webhook[] }>(`/projects/${props.project.id}/webhooks`)
        webhooks.value = res.webhooks ?? []
    } catch (e) {
        const msg = e instanceof Error ? e.message : ''
        if (msg.includes('license required')) errorState.value = 'paywall'
        else errorState.value = 'generic'
    } finally {
        loading.value = false
    }
}

function openCreate() {
    newUrl.value = ''
    newEvents.value = ALL_EVENTS.map(e => e.value)
    showCreate.value = true
}

function toggleEvent(value: string) {
    const idx = newEvents.value.indexOf(value)
    if (idx >= 0) newEvents.value.splice(idx, 1)
    else newEvents.value.push(value)
}

async function submitCreate() {
    if (!newUrl.value) {
        toast.error('URL is required')
        return
    }
    if (newEvents.value.length === 0) {
        toast.error('Select at least one event')
        return
    }
    creating.value = true
    try {
        const hook = await api<Webhook>(`/projects/${props.project.id}/webhooks`, {
            method: 'POST',
            body: { url: newUrl.value, events: newEvents.value },
        })
        webhooks.value.unshift(hook)
        showCreate.value = false
        createdWebhook.value = hook
        secretCopied.value = false
        toast.success('Webhook created')
    } catch (e) {
        toast.error(e instanceof Error ? e.message : 'Failed to create webhook')
    } finally {
        creating.value = false
    }
}

async function toggleActive(hook: Webhook) {
    togglingId.value = hook.id
    try {
        const updated = await api<Webhook>(`/projects/${props.project.id}/webhooks/${hook.id}`, {
            method: 'PATCH',
            body: { active: !hook.active },
        })
        const i = webhooks.value.findIndex(w => w.id === hook.id)
        if (i >= 0) webhooks.value[i] = updated
    } catch (e) {
        toast.error(e instanceof Error ? e.message : 'Failed to update webhook')
    } finally {
        togglingId.value = null
    }
}

async function deleteHook(hook: Webhook) {
    if (!confirm(`Delete webhook for ${hook.url}? This cannot be undone.`)) return
    deletingId.value = hook.id
    try {
        await api(`/projects/${props.project.id}/webhooks/${hook.id}`, { method: 'DELETE' })
        webhooks.value = webhooks.value.filter(w => w.id !== hook.id)
        toast.success('Webhook deleted')
    } catch (e) {
        toast.error(e instanceof Error ? e.message : 'Failed to delete webhook')
    } finally {
        deletingId.value = null
    }
}

async function openDeliveries(hook: Webhook) {
    deliveriesFor.value = hook
    deliveries.value = []
    deliveriesLoading.value = true
    try {
        const res = await api<{ deliveries: Delivery[] }>(`/projects/${props.project.id}/webhooks/${hook.id}/deliveries`)
        deliveries.value = res.deliveries ?? []
    } catch (e) {
        toast.error(e instanceof Error ? e.message : 'Failed to load deliveries')
    } finally {
        deliveriesLoading.value = false
    }
}

async function copySecret() {
    if (!createdWebhook.value) return
    try {
        await navigator.clipboard.writeText(createdWebhook.value.secret)
        secretCopied.value = true
        setTimeout(() => { secretCopied.value = false }, 2000)
    } catch {
        toast.error('Could not copy to clipboard')
    }
}

function eventLabel(value: string) {
    return ALL_EVENTS.find(e => e.value === value)?.label ?? value
}

function fmtDate(iso: string) {
    if (!iso) return ''
    return new Date(iso).toLocaleString('en-US', { dateStyle: 'medium', timeStyle: 'short' })
}

function statusClasses(status: string) {
    if (status === 'delivered') return 'bg-emerald-500/15 text-emerald-400'
    if (status === 'failed') return 'bg-red-500/15 text-red-400'
    if (status === 'inflight') return 'bg-sky-500/15 text-sky-400'
    return 'bg-amber-500/15 text-amber-400'
}

onMounted(load)
</script>

<template>
    <div>
        <div class="flex flex-wrap items-center justify-between gap-3 mb-6">
            <div>
                <div class="flex items-center gap-2">
                    <h1 class="text-2xl font-bold text-white">Webhooks</h1>
                    <span class="text-[10px] font-semibold tracking-wider uppercase px-1.5 py-0.5 rounded bg-amber-500/15 text-amber-400 border border-amber-500/30">Pro</span>
                </div>
                <p class="text-sm text-zinc-500 mt-1">Receive HTTP notifications when events happen in this project</p>
            </div>
            <button v-if="errorState === 'none' && !loading" @click="openCreate"
                class="px-4 py-2 text-sm font-medium bg-white text-zinc-950 rounded-lg hover:bg-zinc-200 transition cursor-pointer">
                New webhook
            </button>
        </div>

        <div v-if="loading" class="text-zinc-500 py-8 text-center">Loading...</div>

        <div v-else-if="errorState === 'paywall'"
            class="bg-zinc-900 border border-zinc-800 rounded-lg p-8 text-center">
            <h2 class="text-lg font-semibold text-white mb-2">Webhooks are a Pro feature</h2>
            <p class="text-sm text-zinc-400 mb-6 max-w-md mx-auto">
                Webhook delivery, signed payloads and retries require a SendDock Pro license.
                Activate one to receive real-time event notifications in your own systems.
            </p>
            <a href="https://senddock.com/pricing" target="_blank" rel="noopener"
                class="inline-block px-4 py-2 text-sm font-medium bg-white text-zinc-950 rounded-lg hover:bg-zinc-200 transition">
                See Pro plans
            </a>
        </div>

        <div v-else-if="errorState === 'generic'"
            class="bg-zinc-900 border border-zinc-800 rounded-lg p-6 text-center">
            <p class="text-zinc-400 text-sm">Couldn't load webhooks. Try again later.</p>
            <button @click="load" class="mt-3 px-3 py-1 text-sm text-white border border-zinc-700 rounded-md hover:bg-zinc-800 cursor-pointer">Retry</button>
        </div>

        <div v-else-if="sortedWebhooks.length === 0"
            class="bg-zinc-900 border border-zinc-800 rounded-lg p-10 text-center">
            <h2 class="text-base font-semibold text-white mb-1">No webhooks yet</h2>
            <p class="text-sm text-zinc-500 mb-5 max-w-md mx-auto">
                Create one to receive a signed HTTP POST every time an event you care about happens.
            </p>
            <button @click="openCreate"
                class="px-4 py-2 text-sm font-medium bg-white text-zinc-950 rounded-lg hover:bg-zinc-200 transition cursor-pointer">
                Create your first webhook
            </button>
        </div>

        <div v-else class="space-y-3">
            <div v-for="hook in sortedWebhooks" :key="hook.id"
                class="bg-zinc-900 border border-zinc-800 rounded-xl p-5">
                <div class="flex items-start justify-between gap-4">
                    <div class="min-w-0 flex-1">
                        <div class="flex items-center gap-2 mb-2">
                            <span :class="[
                                'inline-flex items-center gap-1 text-[10px] font-semibold tracking-wider uppercase px-1.5 py-0.5 rounded',
                                hook.active ? 'bg-emerald-500/15 text-emerald-400' : 'bg-zinc-800 text-zinc-500'
                            ]">
                                <span class="w-1.5 h-1.5 rounded-full" :class="hook.active ? 'bg-emerald-400' : 'bg-zinc-500'"></span>
                                {{ hook.active ? 'Active' : 'Paused' }}
                            </span>
                            <span class="text-xs text-zinc-500">Created {{ fmtDate(hook.created_at) }}</span>
                        </div>
                        <p class="text-sm text-white font-mono truncate" :title="hook.url">{{ hook.url }}</p>
                        <div class="flex flex-wrap gap-1.5 mt-3">
                            <span v-for="ev in hook.events" :key="ev"
                                class="text-[11px] px-2 py-0.5 rounded-md bg-zinc-800 text-zinc-300">
                                {{ ev }}
                            </span>
                        </div>
                    </div>
                    <div class="flex items-center gap-2 flex-shrink-0">
                        <button @click="openDeliveries(hook)"
                            class="px-3 py-1.5 text-xs text-zinc-300 border border-zinc-700 rounded-md hover:bg-zinc-800 transition cursor-pointer">
                            Deliveries
                        </button>
                        <button @click="toggleActive(hook)" :disabled="togglingId === hook.id"
                            class="px-3 py-1.5 text-xs text-zinc-300 border border-zinc-700 rounded-md hover:bg-zinc-800 transition cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed">
                            {{ hook.active ? 'Pause' : 'Resume' }}
                        </button>
                        <button @click="deleteHook(hook)" :disabled="deletingId === hook.id"
                            class="px-3 py-1.5 text-xs text-red-400 border border-red-900/50 rounded-md hover:bg-red-950/40 transition cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed">
                            Delete
                        </button>
                    </div>
                </div>
            </div>
        </div>

        <AppModal :show="showCreate" title="New webhook" @close="showCreate = false">
            <form @submit.prevent="submitCreate" class="space-y-5">
                <AppInput v-model="newUrl" label="Endpoint URL" type="url"
                    placeholder="https://example.com/webhooks/senddock" required />

                <div>
                    <label class="block text-sm font-medium text-zinc-300 mb-2">Events</label>
                    <div class="space-y-2">
                        <label v-for="ev in ALL_EVENTS" :key="ev.value"
                            class="flex items-start gap-2.5 p-2.5 rounded-lg border border-zinc-800 hover:border-zinc-700 cursor-pointer transition group">
                            <span class="relative flex-shrink-0 mt-0.5 w-[18px] h-[18px]">
                                <input type="checkbox" :checked="newEvents.includes(ev.value)"
                                    @change="toggleEvent(ev.value)"
                                    class="peer sr-only" />
                                <span class="absolute inset-0 rounded border-2 border-zinc-600 bg-transparent transition-colors peer-checked:border-white peer-checked:bg-white group-hover:border-zinc-400"></span>
                                <svg v-if="newEvents.includes(ev.value)"
                                    class="absolute inset-0 m-auto w-3 h-3 text-zinc-950 pointer-events-none"
                                    viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
                                    <polyline points="3 8 7 12 13 4" />
                                </svg>
                            </span>
                            <div class="min-w-0">
                                <p class="text-sm text-white font-mono">{{ ev.value }}</p>
                                <p class="text-xs text-zinc-500 mt-0.5">{{ ev.description }}</p>
                            </div>
                        </label>
                    </div>
                </div>

                <div class="flex gap-3 pt-2">
                    <button type="button" @click="showCreate = false"
                        class="flex-1 py-2 text-sm font-medium border border-zinc-700 text-zinc-300 rounded-lg hover:bg-zinc-800 transition cursor-pointer">
                        Cancel
                    </button>
                    <AppButton type="submit" :loading="creating" :disabled="creating">
                        {{ creating ? 'Creating...' : 'Create webhook' }}
                    </AppButton>
                </div>
            </form>
        </AppModal>

        <AppModal :show="!!createdWebhook" title="Save your signing secret"
            @close="createdWebhook = null">
            <div v-if="createdWebhook" class="space-y-4">
                <div class="bg-amber-500/10 border border-amber-500/30 rounded-lg p-3">
                    <p class="text-xs text-amber-300">
                        This is the only time you'll see this secret. Store it somewhere safe — you'll need it to verify webhook signatures.
                    </p>
                </div>

                <div>
                    <label class="block text-sm font-medium text-zinc-300 mb-1">Signing secret</label>
                    <div class="flex gap-2">
                        <code class="flex-1 px-3 py-2 bg-zinc-950 border border-zinc-800 rounded-lg text-xs text-emerald-400 font-mono break-all">{{ createdWebhook.secret }}</code>
                        <button @click="copySecret"
                            class="px-3 py-2 text-xs font-medium border border-zinc-700 text-zinc-300 rounded-lg hover:bg-zinc-800 transition cursor-pointer flex-shrink-0">
                            {{ secretCopied ? 'Copied' : 'Copy' }}
                        </button>
                    </div>
                </div>

                <div>
                    <label class="block text-sm font-medium text-zinc-300 mb-1">Endpoint</label>
                    <code class="block px-3 py-2 bg-zinc-950 border border-zinc-800 rounded-lg text-xs text-zinc-300 font-mono break-all">{{ createdWebhook.url }}</code>
                </div>

                <p class="text-xs text-zinc-500">
                    Each delivery includes a <code class="text-zinc-400">X-SendDock-Signature</code> header in the form
                    <code class="text-zinc-400">t=&lt;timestamp&gt;,v1=&lt;hex&gt;</code>, where the hex is HMAC-SHA256 of
                    <code class="text-zinc-400">timestamp.body</code> using this secret.
                </p>

                <button @click="createdWebhook = null"
                    class="w-full py-2 font-medium bg-white text-zinc-950 rounded-lg hover:bg-zinc-200 transition cursor-pointer">
                    Done
                </button>
            </div>
        </AppModal>

        <AppModal :show="!!deliveriesFor" :title="deliveriesFor ? 'Recent deliveries' : ''"
            @close="deliveriesFor = null">
            <div v-if="deliveriesFor">
                <p class="text-xs text-zinc-500 font-mono break-all mb-4">{{ deliveriesFor.url }}</p>

                <div v-if="deliveriesLoading" class="text-center text-sm text-zinc-500 py-6">Loading...</div>

                <div v-else-if="deliveries.length === 0" class="text-center text-sm text-zinc-500 py-6">
                    No deliveries yet. Trigger an event to see attempts here.
                </div>

                <ul v-else class="space-y-2">
                    <li v-for="d in deliveries" :key="d.id"
                        class="bg-zinc-950 border border-zinc-800 rounded-lg p-3">
                        <div class="flex items-center justify-between gap-2 mb-1.5">
                            <span class="text-xs font-mono text-zinc-300">{{ d.event_type }}</span>
                            <span :class="['text-[10px] font-semibold uppercase tracking-wider px-1.5 py-0.5 rounded', statusClasses(d.status)]">
                                {{ d.status }}
                            </span>
                        </div>
                        <div class="flex items-center justify-between text-xs text-zinc-500">
                            <span>{{ fmtDate(d.created_at) }}</span>
                            <span>{{ d.attempts }} attempt{{ d.attempts === 1 ? '' : 's' }}<span v-if="d.last_status_code"> · HTTP {{ d.last_status_code }}</span></span>
                        </div>
                        <p v-if="d.last_error" class="mt-1.5 text-xs text-red-400 break-words">{{ d.last_error }}</p>
                        <p v-else-if="d.next_attempt_at && d.status === 'pending'" class="mt-1.5 text-xs text-amber-400">
                            Next attempt {{ fmtDate(d.next_attempt_at) }}
                        </p>
                    </li>
                </ul>
            </div>
        </AppModal>
    </div>
</template>

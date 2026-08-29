<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { api } from '@/api/client'

interface EmailLog {
    id: string
    project_id: string
    subscriber_id: string | null
    template_id: string | null
    to_email: string
    subject: string
    status: string
    error: string | null
    sent_at: string
    opened_at: string | null
    clicked_at: string | null
}

interface EmailClick {
    id: string
    url: string
    clicked_at: string
    user_agent: string | null
}

interface DetailResponse {
    log: EmailLog
    clicks: EmailClick[] | null
}

const props = defineProps<{
    projectId: string
    logId: string | null
}>()

const emit = defineEmits<{
    close: []
}>()

const detail = ref<DetailResponse | null>(null)
const loading = ref(false)
const errored = ref(false)

const log = computed(() => detail.value?.log ?? null)
const clicks = computed(() => detail.value?.clicks ?? [])

const statusClass = computed(() => {
    if (!log.value) return ''
    switch (log.value.status) {
        case 'sent': return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30'
        case 'failed': return 'bg-red-500/10 text-red-400 border-red-500/30'
        case 'bounced': return 'bg-orange-500/10 text-orange-400 border-orange-500/30'
        case 'suppressed': return 'bg-zinc-500/10 text-zinc-300 border-zinc-600'
        default: return 'bg-zinc-700/30 text-zinc-300 border-zinc-700'
    }
})

interface TimelineEvent {
    label: string
    at: string
    dotClass: string
    detail?: string
}

const timeline = computed<TimelineEvent[]>(() => {
    if (!log.value) return []
    const out: TimelineEvent[] = []
    const l = log.value

    out.push({ label: 'Sent', at: l.sent_at, dotClass: 'bg-zinc-400' })

    if (l.opened_at) {
        out.push({ label: 'Opened', at: l.opened_at, dotClass: 'bg-blue-400' })
    }
    if (l.clicked_at) {
        out.push({ label: 'First click', at: l.clicked_at, dotClass: 'bg-amber-400', detail: clicks.value.length > 1 ? `+${clicks.value.length - 1} more` : undefined })
    }
    if (l.status === 'bounced') {
        out.push({ label: 'Bounced', at: l.sent_at, dotClass: 'bg-orange-400', detail: l.error || undefined })
    }
    if (l.status === 'failed') {
        out.push({ label: 'Send failed', at: l.sent_at, dotClass: 'bg-red-400', detail: l.error || undefined })
    }
    if (l.status === 'suppressed') {
        out.push({ label: 'Suppressed', at: l.sent_at, dotClass: 'bg-zinc-500', detail: l.error || 'Recipient is on the suppression list' })
    }
    return out
})

async function load(id: string) {
    loading.value = true
    errored.value = false
    detail.value = null
    try {
        detail.value = await api<DetailResponse>(`/projects/${props.projectId}/logs/${id}`)
    } catch {
        errored.value = true
    } finally {
        loading.value = false
    }
}

watch(() => props.logId, (id) => {
    if (id) load(id)
    else detail.value = null
}, { immediate: true })

function formatDate(iso: string): string {
    return new Date(iso).toLocaleString()
}

function userAgentSummary(ua: string | null): string {
    if (!ua) return '—'
    if (ua.length <= 60) return ua
    return ua.slice(0, 60) + '…'
}
</script>

<template>
    <Teleport to="body">
        <div v-if="logId" class="fixed inset-0 z-40 flex justify-end">
            <div class="absolute inset-0 bg-black/50" @click="emit('close')"></div>

            <aside class="relative bg-zinc-900 border-l border-zinc-800 w-full sm:max-w-md md:max-w-lg h-full overflow-y-auto shadow-2xl">
                <div class="sticky top-0 bg-zinc-900/95 backdrop-blur border-b border-zinc-800 px-5 py-3 flex items-start justify-between gap-3 z-10">
                    <div class="min-w-0 flex-1">
                        <p class="text-xs uppercase tracking-wide text-zinc-400">Email log</p>
                        <h2 class="text-base font-semibold text-white truncate">
                            {{ log?.subject || '(no subject)' }}
                        </h2>
                    </div>
                    <button @click="emit('close')" type="button" aria-label="Close"
                        class="w-8 h-8 flex items-center justify-center rounded-lg text-zinc-300 hover:text-white hover:bg-zinc-850 transition cursor-pointer shrink-0">
                        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
                            <line x1="18" y1="6" x2="6" y2="18"></line>
                            <line x1="6" y1="6" x2="18" y2="18"></line>
                        </svg>
                    </button>
                </div>

                <div class="p-5 space-y-6">
                    <div v-if="loading" class="text-zinc-400 text-sm">Loading...</div>

                    <div v-else-if="errored" class="text-zinc-300 text-sm">
                        Failed to load this log. It may have been deleted.
                    </div>

                    <template v-else-if="log">
                        <section>
                            <div class="flex items-center gap-2 mb-3">
                                <span :class="['text-xs px-2 py-1 rounded-full border whitespace-nowrap', statusClass]">
                                    {{ log.status }}
                                </span>
                                <span v-if="log.opened_at" class="text-xs px-2 py-1 rounded-full bg-blue-500/10 text-blue-400 border border-blue-500/30">opened</span>
                                <span v-if="log.clicked_at" class="text-xs px-2 py-1 rounded-full bg-amber-500/10 text-amber-400 border border-amber-500/30">clicked</span>
                            </div>
                            <dl class="grid grid-cols-1 gap-3 text-sm">
                                <div>
                                    <dt class="text-xs uppercase tracking-wide text-zinc-400 mb-0.5">To</dt>
                                    <dd class="text-white break-all">{{ log.to_email }}</dd>
                                </div>
                                <div>
                                    <dt class="text-xs uppercase tracking-wide text-zinc-400 mb-0.5">Sent at</dt>
                                    <dd class="text-zinc-300 font-mono text-xs">{{ new Date(log.sent_at).toISOString() }}</dd>
                                </div>
                            </dl>
                        </section>

                        <section>
                            <h3 class="text-sm font-semibold text-white mb-3">Timeline</h3>
                            <ol class="space-y-3">
                                <li v-for="(ev, i) in timeline" :key="i" class="flex gap-3">
                                    <div class="flex flex-col items-center pt-1">
                                        <span :class="['w-2.5 h-2.5 rounded-full shrink-0', ev.dotClass]"></span>
                                        <span v-if="i < timeline.length - 1" class="w-px flex-1 bg-zinc-850 mt-1"></span>
                                    </div>
                                    <div class="flex-1 min-w-0 pb-3">
                                        <p class="text-sm text-white">{{ ev.label }}</p>
                                        <p class="text-xs text-zinc-400">{{ formatDate(ev.at) }}</p>
                                        <p v-if="ev.detail" class="text-xs text-zinc-300 mt-1 break-words">{{ ev.detail }}</p>
                                    </div>
                                </li>
                            </ol>
                        </section>

                        <section v-if="clicks.length > 0">
                            <h3 class="text-sm font-semibold text-white mb-3">
                                Click events
                                <span class="text-xs font-normal text-zinc-400 ml-1">({{ clicks.length }})</span>
                            </h3>
                            <ul class="space-y-2">
                                <li v-for="c in clicks" :key="c.id"
                                    class="bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs">
                                    <a :href="c.url" target="_blank" rel="noopener"
                                        class="text-amber-400 hover:text-amber-300 break-all underline decoration-zinc-700 hover:decoration-amber-400">
                                        {{ c.url }}
                                    </a>
                                    <div class="flex items-center justify-between mt-1.5 text-zinc-400">
                                        <span>{{ formatDate(c.clicked_at) }}</span>
                                        <span class="truncate ml-3" :title="c.user_agent ?? ''">{{ userAgentSummary(c.user_agent) }}</span>
                                    </div>
                                </li>
                            </ul>
                        </section>

                        <section v-if="log.error">
                            <h3 class="text-sm font-semibold text-white mb-2">
                                {{ log.status === 'suppressed' ? 'Suppression reason' : 'Error' }}
                            </h3>
                            <p :class="[
                                'text-xs whitespace-pre-wrap break-words bg-zinc-950 border rounded-lg p-3',
                                log.status === 'failed' && 'text-red-400 border-red-500/30',
                                log.status === 'bounced' && 'text-orange-400 border-orange-500/30',
                                log.status === 'suppressed' && 'text-zinc-300 border-zinc-800',
                            ]">{{ log.error }}</p>
                        </section>

                        <section>
                            <h3 class="text-sm font-semibold text-white mb-3">References</h3>
                            <dl class="grid grid-cols-1 gap-2 text-xs">
                                <div>
                                    <dt class="text-zinc-400 uppercase tracking-wide font-medium">Log ID</dt>
                                    <dd class="text-zinc-300 font-mono break-all">{{ log.id }}</dd>
                                </div>
                                <div v-if="log.subscriber_id">
                                    <dt class="text-zinc-400 uppercase tracking-wide font-medium">Subscriber ID</dt>
                                    <dd class="text-zinc-300 font-mono break-all">{{ log.subscriber_id }}</dd>
                                </div>
                                <div v-if="log.template_id">
                                    <dt class="text-zinc-400 uppercase tracking-wide font-medium">Template ID</dt>
                                    <dd class="text-zinc-300 font-mono break-all">{{ log.template_id }}</dd>
                                </div>
                            </dl>
                        </section>
                    </template>
                </div>
            </aside>
        </div>
    </Teleport>
</template>

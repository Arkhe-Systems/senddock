<script setup lang="ts">
import { ref, onMounted, computed, onBeforeUnmount } from 'vue'
import { marked } from 'marked'
import { api, ApiError } from '@/api/client'
import { useToastStore } from '@/stores/toast'
import AppModal from '@/components/ui/AppModal.vue'

interface ReleaseInfo {
    current: string
    latest: string
    outdated: boolean
    release_url: string
    notes: string
    checked_at: string
    available: boolean
    enabled: boolean
}

interface WatchtowerStatus {
    configured: boolean
    healthy: boolean
    url?: string
    last_check?: string
    last_error?: string
}

const toast = useToastStore()
const release = ref<ReleaseInfo | null>(null)
const watchtower = ref<WatchtowerStatus | null>(null)
const showModal = ref(false)
const triggering = ref(false)
const polling = ref(false)
const showManualCommand = ref(false)

const updateCommand = 'docker compose pull && docker compose up -d'

marked.setOptions({ gfm: true, breaks: false })

const renderedNotes = computed(() => {
    if (!release.value?.notes) return ''
    return marked.parse(release.value.notes) as string
})

const modalTitle = computed(() => {
    if (!release.value) return ''
    if (polling.value) return 'Updating SendDock...'
    return release.value.outdated
        ? 'Update available'
        : `What's in v${release.value.current}`
})

const canOneClick = computed(() => Boolean(watchtower.value?.configured && watchtower.value?.healthy))

let pollTimer: ReturnType<typeof setInterval> | null = null

async function loadStatus() {
    try {
        release.value = await api<ReleaseInfo>('/version')
    } catch {
        release.value = null
    }
    try {
        watchtower.value = await api<WatchtowerStatus>('/update/auto-status', { silent: true })
    } catch {
        watchtower.value = { configured: false, healthy: false }
    }
}

onMounted(loadStatus)
onBeforeUnmount(() => {
    if (pollTimer) clearInterval(pollTimer)
})

async function copyCommand() {
    await navigator.clipboard.writeText(updateCommand)
    toast.success('Update command copied')
}

async function triggerUpdate() {
    if (!release.value || triggering.value || polling.value) return
    triggering.value = true
    const originalVersion = release.value.current
    try {
        await api('/update/trigger', { method: 'POST' })
        polling.value = true
        toast.success('Update triggered — SendDock will restart shortly')

        const startedAt = Date.now()
        const timeoutMs = 5 * 60 * 1000
        pollTimer = setInterval(async () => {
            if (Date.now() - startedAt > timeoutMs) {
                stopPolling()
                toast.error('Update timed out — refresh manually to check')
                return
            }
            try {
                const fresh = await api<ReleaseInfo>('/version', { silent: true })
                if (fresh.current && fresh.current !== originalVersion) {
                    stopPolling()
                    release.value = fresh
                    toast.success(`Updated to v${fresh.current}`)
                }
            } catch {
                // network errors are expected while the backend restarts
            }
        }, 2000)
    } catch (e) {
        const msg = e instanceof ApiError ? e.message : 'failed to trigger update'
        toast.error(msg)
    } finally {
        triggering.value = false
    }
}

function stopPolling() {
    polling.value = false
    if (pollTimer) {
        clearInterval(pollTimer)
        pollTimer = null
    }
}
</script>

<template>
    <div v-if="release && release.enabled">
        <button v-if="release.outdated" @click="showModal = true"
            class="flex items-center gap-2 px-3 py-1.5 text-xs rounded-lg bg-yellow-500/10 border border-yellow-500/30 text-yellow-300 hover:bg-yellow-500/20 transition cursor-pointer">
            <span class="relative flex h-2 w-2">
                <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-yellow-400 opacity-75"></span>
                <span class="relative inline-flex rounded-full h-2 w-2 bg-yellow-400"></span>
            </span>
            Update available · v{{ release.latest }}
        </button>
        <button v-else @click="showModal = true"
            class="flex items-center gap-1.5 px-2 py-1 text-xs font-mono text-zinc-500 hover:text-white hover:bg-zinc-800 rounded-md transition cursor-pointer">
            v{{ release.current }}
        </button>

        <AppModal :show="showModal" :title="modalTitle" size="lg" @close="polling ? null : (showModal = false)">
            <div class="space-y-5">
                <div v-if="release.outdated && !polling" class="flex items-center gap-3">
                    <div class="px-3 py-2 bg-zinc-900 border border-zinc-800 rounded-lg">
                        <p class="text-[11px] text-zinc-500 uppercase tracking-wide">Current</p>
                        <p class="text-sm text-white font-mono">v{{ release.current }}</p>
                    </div>
                    <span class="text-zinc-600">&rarr;</span>
                    <div class="px-3 py-2 bg-yellow-500/10 border border-yellow-500/30 rounded-lg">
                        <p class="text-[11px] text-yellow-400/80 uppercase tracking-wide">Latest</p>
                        <p class="text-sm text-yellow-300 font-mono">v{{ release.latest }}</p>
                    </div>
                </div>

                <div v-if="polling" class="space-y-3 text-sm text-zinc-300">
                    <div class="flex items-center gap-3">
                        <span class="inline-block w-4 h-4 border-2 border-zinc-700 border-t-yellow-400 rounded-full animate-spin"></span>
                        <span>Watchtower is pulling v{{ release.latest }} and recreating the container. This usually takes 30–90 seconds.</span>
                    </div>
                    <p class="text-xs text-zinc-500">
                        Your session may briefly disconnect. When the new version is live, this dialog will update automatically.
                    </p>
                </div>

                <template v-else-if="release.outdated">
                    <div v-if="canOneClick" class="space-y-3">
                        <button @click="triggerUpdate" :disabled="triggering"
                            class="w-full px-4 py-3 text-sm font-semibold bg-yellow-500 text-zinc-950 rounded-lg hover:bg-yellow-400 transition cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed">
                            {{ triggering ? 'Triggering...' : `Update now to v${release.latest}` }}
                        </button>
                        <p class="text-xs text-zinc-500 text-center">
                            Watchtower at <code class="text-zinc-400">{{ watchtower?.url }}</code> will pull the new image and recreate the container. Postgres and Redis volumes are preserved.
                        </p>
                        <details class="text-xs">
                            <summary class="text-zinc-500 hover:text-zinc-300 cursor-pointer">Or run it manually</summary>
                            <div class="mt-2 flex items-center gap-2">
                                <code class="flex-1 px-3 py-2 bg-zinc-950 border border-zinc-800 rounded-lg text-xs text-white font-mono select-all">{{ updateCommand }}</code>
                                <button @click="copyCommand"
                                    class="px-3 py-2 text-xs bg-zinc-800 hover:bg-zinc-700 text-white rounded-lg transition cursor-pointer">
                                    Copy
                                </button>
                            </div>
                        </details>
                    </div>

                    <div v-else>
                        <div v-if="watchtower?.configured && !watchtower?.healthy" class="mb-3 p-3 bg-red-500/10 border border-red-500/30 rounded-lg text-xs text-red-300">
                            One-click update unavailable — Watchtower at
                            <code class="font-mono text-red-200">{{ watchtower.url }}</code>
                            is unreachable. <span v-if="watchtower.last_error" class="text-red-400">({{ watchtower.last_error }})</span>
                        </div>
                        <p v-else class="text-xs text-zinc-500 mb-2">
                            Tip: configure Watchtower with the
                            <code class="text-zinc-400">SENDDOCK_WATCHTOWER_URL</code>
                            env var to enable one-click updates from this page.
                        </p>
                        <p class="text-sm font-medium text-white mb-2">Run this from the SendDock folder on your host:</p>
                        <div class="flex items-center gap-2">
                            <code class="flex-1 px-3 py-2 bg-zinc-950 border border-zinc-800 rounded-lg text-xs text-white font-mono select-all">{{ updateCommand }}</code>
                            <button @click="copyCommand"
                                class="px-3 py-2 text-xs bg-zinc-800 hover:bg-zinc-700 text-white rounded-lg transition cursor-pointer">
                                Copy
                            </button>
                        </div>
                        <p class="text-xs text-zinc-500 mt-2">
                            Pulls the new image, recreates the container and runs migrations on first start. Postgres and Redis volumes are preserved, so subscribers, templates and history stay intact. If you built from source, run <code class="text-zinc-400">git pull && docker compose -f docker-compose.prod.yml up -d --build</code> instead.
                        </p>
                    </div>
                </template>

                <div v-if="renderedNotes && !polling">
                    <p v-if="release.outdated" class="text-sm font-medium text-white mb-2">Release notes</p>
                    <div class="release-notes text-sm text-zinc-300 bg-zinc-950 border border-zinc-800 rounded-lg p-4 max-h-96 overflow-auto" v-html="renderedNotes" />
                </div>

                <div class="flex justify-between items-center text-xs">
                    <a v-if="release.release_url" :href="release.release_url" target="_blank" rel="noopener"
                        class="text-zinc-400 hover:text-white transition underline decoration-zinc-700 underline-offset-2">
                        Full release on GitHub &rarr;
                    </a>
                    <span class="text-zinc-600">Checked {{ new Date(release.checked_at).toLocaleString() }}</span>
                </div>
            </div>
        </AppModal>
    </div>
</template>

<style scoped>
.release-notes :deep(h1),
.release-notes :deep(h2) {
    font-size: 0.95rem;
    font-weight: 600;
    color: #fafafa;
    margin: 1rem 0 0.5rem;
}
.release-notes :deep(h2):first-child,
.release-notes :deep(h1):first-child {
    margin-top: 0;
}
.release-notes :deep(h3) {
    font-size: 0.85rem;
    font-weight: 600;
    color: #fafafa;
    margin: 0.75rem 0 0.4rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
}
.release-notes :deep(p) {
    margin: 0.4rem 0;
    line-height: 1.55;
}
.release-notes :deep(ul),
.release-notes :deep(ol) {
    margin: 0.4rem 0;
    padding-left: 1.2rem;
    line-height: 1.55;
}
.release-notes :deep(li) {
    margin: 0.2rem 0;
}
.release-notes :deep(li)::marker {
    color: #71717a;
}
.release-notes :deep(code) {
    background: #18181b;
    border-radius: 4px;
    padding: 0.05rem 0.35rem;
    font-size: 0.8em;
    color: #e4e4e7;
    word-break: break-word;
}
.release-notes :deep(pre) {
    background: #09090b;
    border: 1px solid #27272a;
    border-radius: 6px;
    padding: 0.6rem 0.8rem;
    font-size: 0.78rem;
    margin: 0.6rem 0;
    white-space: pre-wrap;
    word-break: break-word;
    overflow-wrap: anywhere;
}
.release-notes :deep(pre code) {
    background: transparent;
    border: none;
    padding: 0;
    font-size: inherit;
    white-space: inherit;
    word-break: inherit;
    overflow-wrap: inherit;
}
.release-notes :deep(a) {
    color: #fbbf24;
    text-decoration: underline;
    text-decoration-color: #52525b;
    text-underline-offset: 2px;
}
.release-notes :deep(a):hover {
    text-decoration-color: #fbbf24;
}
.release-notes :deep(strong) {
    color: #fafafa;
    font-weight: 600;
}
.release-notes :deep(hr) {
    border: none;
    border-top: 1px solid #27272a;
    margin: 0.8rem 0;
}
</style>

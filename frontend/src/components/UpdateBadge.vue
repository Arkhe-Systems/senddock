<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { marked } from 'marked'
import { api } from '@/api/client'
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

const toast = useToastStore()
const release = ref<ReleaseInfo | null>(null)
const showModal = ref(false)

const updateCommand = 'docker compose pull && docker compose up -d'

marked.setOptions({ gfm: true, breaks: false })

const renderedNotes = computed(() => {
    if (!release.value?.notes) return ''
    return marked.parse(release.value.notes) as string
})

onMounted(async () => {
    try {
        release.value = await api<ReleaseInfo>('/version')
    } catch {
        release.value = null
    }
})

async function copyCommand() {
    await navigator.clipboard.writeText(updateCommand)
    toast.success('Update command copied')
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
        <span v-else class="text-xs text-zinc-500 font-mono">v{{ release.current }}</span>

        <AppModal :show="showModal" title="Update available" @close="showModal = false">
            <div class="space-y-5">
                <div class="flex items-center gap-3">
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

                <div>
                    <p class="text-sm font-medium text-white mb-2">Run this from the SendDock folder on your host:</p>
                    <div class="flex items-center gap-2">
                        <code class="flex-1 px-3 py-2 bg-zinc-950 border border-zinc-800 rounded-lg text-xs text-white font-mono select-all">{{ updateCommand }}</code>
                        <button @click="copyCommand"
                            class="px-3 py-2 text-xs bg-zinc-800 hover:bg-zinc-700 text-white rounded-lg transition cursor-pointer">
                            Copy
                        </button>
                    </div>
                    <p class="text-xs text-zinc-500 mt-2">
                        Pulls the new image, recreates the container and runs migrations on first start. Postgres and Redis volumes are preserved, so subscribers, templates and history stay intact. If you built from source, run <code class="text-zinc-400">git pull && ./setup.sh</code> instead.
                    </p>
                </div>

                <div v-if="renderedNotes">
                    <p class="text-sm font-medium text-white mb-2">Release notes</p>
                    <div class="release-notes text-sm text-zinc-300 bg-zinc-950 border border-zinc-800 rounded-lg p-4 max-h-72 overflow-auto" v-html="renderedNotes" />
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
    border: 1px solid #27272a;
    border-radius: 4px;
    padding: 0.05rem 0.35rem;
    font-size: 0.8em;
    color: #e4e4e7;
}
.release-notes :deep(pre) {
    background: #09090b;
    border: 1px solid #27272a;
    border-radius: 6px;
    padding: 0.6rem 0.8rem;
    overflow-x: auto;
    font-size: 0.78rem;
    margin: 0.6rem 0;
}
.release-notes :deep(pre code) {
    background: transparent;
    border: none;
    padding: 0;
    font-size: inherit;
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

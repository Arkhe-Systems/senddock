<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
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
}

const toast = useToastStore()
const release = ref<ReleaseInfo | null>(null)
const showModal = ref(false)

const updateCommand = 'git pull && ./setup.sh'

const truncatedNotes = computed(() => {
    if (!release.value?.notes) return ''
    const notes = release.value.notes
    if (notes.length <= 600) return notes
    return notes.slice(0, 600) + '\n\n…'
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
    <div v-if="release">
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
                        The script keeps your <code>.env</code> and database, rebuilds the image with the new code, and restarts services. Your subscribers, templates, and history are preserved.
                    </p>
                </div>

                <div v-if="truncatedNotes">
                    <p class="text-sm font-medium text-white mb-2">Release notes</p>
                    <pre class="text-xs text-zinc-400 bg-zinc-950 border border-zinc-800 rounded-lg p-3 max-h-60 overflow-auto whitespace-pre-wrap font-mono">{{ truncatedNotes }}</pre>
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

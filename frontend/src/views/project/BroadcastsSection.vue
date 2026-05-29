<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { api } from '@/api/client'
import type { Project } from '@/stores/projects'
import AppPagination from '@/components/ui/AppPagination.vue'

interface Broadcast {
    id: string
    project_id: string
    template_id: string
    subject: string
    status: 'sending' | 'completed' | string
    total_recipients: number
    sent_count: number
    failed_count: number
    suppressed_count: number
    started_at: string
    finished_at: string | null
}

const props = defineProps<{ project: Project }>()

const broadcasts = ref<Broadcast[]>([])
const total = ref(0)
const loading = ref(true)
const page = ref(0)
const limit = ref(25)
const expandedId = ref<string | null>(null)

let pollTimer: ReturnType<typeof setInterval> | null = null

const hasInProgress = computed(() => broadcasts.value.some(b => b.status === 'sending'))

async function fetchBroadcasts() {
    try {
        const res = await api<{ broadcasts: Broadcast[] | null; total: number }>(
            `/projects/${props.project.id}/broadcasts?limit=${limit.value}&offset=${page.value * limit.value}`,
        )
        broadcasts.value = res.broadcasts ?? []
        total.value = res.total
    } catch {
        broadcasts.value = []
    } finally {
        loading.value = false
    }
}

function progressFor(b: Broadcast): number {
    if (b.total_recipients === 0) return 0
    const done = b.sent_count + b.failed_count + b.suppressed_count
    return Math.min(100, Math.round((done / b.total_recipients) * 100))
}

function durationFor(b: Broadcast): string {
    const start = new Date(b.started_at).getTime()
    const end = b.finished_at ? new Date(b.finished_at).getTime() : Date.now()
    const seconds = Math.max(0, Math.round((end - start) / 1000))
    if (seconds < 60) return `${seconds}s`
    const minutes = Math.floor(seconds / 60)
    const remainder = seconds % 60
    if (minutes < 60) return `${minutes}m ${remainder}s`
    const hours = Math.floor(minutes / 60)
    return `${hours}h ${minutes % 60}m`
}

function toggleExpand(id: string) {
    expandedId.value = expandedId.value === id ? null : id
}

onMounted(() => {
    fetchBroadcasts()
    pollTimer = setInterval(() => {
        if (hasInProgress.value) fetchBroadcasts()
    }, 3000)
})

import { onBeforeUnmount } from 'vue'
onBeforeUnmount(() => {
    if (pollTimer) clearInterval(pollTimer)
})
</script>

<template>
    <div>
        <div class="flex flex-wrap items-center justify-between gap-3 mb-6">
            <div>
                <h1 class="text-2xl font-bold text-white">Broadcasts</h1>
                <p class="text-sm text-zinc-500 mt-1">
                    History of every send to all active subscribers, with per-broadcast progress and recovery info.
                </p>
            </div>
        </div>

        <div v-if="loading" class="text-zinc-500 py-8 text-center">Loading...</div>

        <div v-else-if="broadcasts.length > 0" class="bg-zinc-900 border border-zinc-800 rounded-lg overflow-x-auto">
            <table class="w-full min-w-[720px]">
                <thead>
                    <tr class="border-b border-zinc-800">
                        <th class="w-8"></th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Subject</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Status</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Progress</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Started</th>
                    </tr>
                </thead>
                <tbody>
                    <template v-for="b in broadcasts" :key="b.id">
                        <tr
                            @click="toggleExpand(b.id)"
                            class="border-b border-zinc-800 last:border-0 hover:bg-zinc-800/40 cursor-pointer transition-colors">
                            <td class="px-3 py-3 text-zinc-500">
                                <span class="inline-block transition-transform" :class="expandedId === b.id ? 'rotate-90' : ''">›</span>
                            </td>
                            <td class="px-4 py-3 text-sm text-white max-w-md truncate">{{ b.subject || '(no subject)' }}</td>
                            <td class="px-4 py-3">
                                <span :class="[
                                    'text-xs px-2 py-1 rounded-full whitespace-nowrap',
                                    b.status === 'sending' && 'bg-blue-500/10 text-blue-400',
                                    b.status === 'completed' && 'bg-green-500/10 text-green-400',
                                ]">
                                    {{ b.status }}
                                </span>
                            </td>
                            <td class="px-4 py-3 text-sm text-zinc-400">
                                <div class="flex items-center gap-2 min-w-[180px]">
                                    <div class="flex-1 h-1.5 bg-zinc-800 rounded-full overflow-hidden">
                                        <div
                                            :class="[
                                                'h-full transition-all',
                                                b.status === 'completed' && 'bg-green-500',
                                                b.status === 'sending' && 'bg-blue-500',
                                            ]"
                                            :style="{ width: progressFor(b) + '%' }">
                                        </div>
                                    </div>
                                    <span class="text-xs text-zinc-500 tabular-nums whitespace-nowrap">
                                        {{ b.sent_count + b.failed_count + b.suppressed_count }}/{{ b.total_recipients }}
                                    </span>
                                </div>
                            </td>
                            <td class="px-4 py-3 text-sm text-zinc-500 whitespace-nowrap">{{ new Date(b.started_at).toLocaleString() }}</td>
                        </tr>
                        <tr v-if="expandedId === b.id" class="border-b border-zinc-800 last:border-0 bg-zinc-950/60">
                            <td></td>
                            <td colspan="4" class="px-4 py-4">
                                <dl class="grid grid-cols-2 sm:grid-cols-4 gap-x-6 gap-y-3 text-xs">
                                    <div>
                                        <dt class="text-zinc-500 uppercase tracking-wide font-medium">Sent</dt>
                                        <dd class="text-green-400 text-base font-semibold tabular-nums">{{ b.sent_count }}</dd>
                                    </div>
                                    <div>
                                        <dt class="text-zinc-500 uppercase tracking-wide font-medium">Failed</dt>
                                        <dd class="text-red-400 text-base font-semibold tabular-nums">{{ b.failed_count }}</dd>
                                    </div>
                                    <div>
                                        <dt class="text-zinc-500 uppercase tracking-wide font-medium">Suppressed</dt>
                                        <dd class="text-zinc-300 text-base font-semibold tabular-nums">{{ b.suppressed_count }}</dd>
                                    </div>
                                    <div>
                                        <dt class="text-zinc-500 uppercase tracking-wide font-medium">Total recipients</dt>
                                        <dd class="text-zinc-300 text-base font-semibold tabular-nums">{{ b.total_recipients }}</dd>
                                    </div>
                                </dl>
                                <div class="mt-4 pt-3 border-t border-zinc-800 grid grid-cols-1 sm:grid-cols-3 gap-x-6 gap-y-2 text-xs">
                                    <div>
                                        <dt class="text-zinc-500 uppercase tracking-wide font-medium">Started</dt>
                                        <dd class="text-zinc-300 font-mono">{{ new Date(b.started_at).toISOString() }}</dd>
                                    </div>
                                    <div>
                                        <dt class="text-zinc-500 uppercase tracking-wide font-medium">{{ b.finished_at ? 'Finished' : 'Running for' }}</dt>
                                        <dd class="text-zinc-300 font-mono">
                                            {{ b.finished_at ? new Date(b.finished_at).toISOString() : durationFor(b) }}
                                        </dd>
                                    </div>
                                    <div>
                                        <dt class="text-zinc-500 uppercase tracking-wide font-medium">Duration</dt>
                                        <dd class="text-zinc-300 font-mono">{{ durationFor(b) }}</dd>
                                    </div>
                                </div>
                                <div class="mt-3 pt-3 border-t border-zinc-800">
                                    <dt class="text-xs text-zinc-500 uppercase tracking-wide font-medium mb-1">Broadcast ID</dt>
                                    <dd class="text-xs text-zinc-400 font-mono break-all">{{ b.id }}</dd>
                                </div>
                            </td>
                        </tr>
                    </template>
                </tbody>
            </table>
        </div>

        <div v-else class="bg-zinc-900 border border-zinc-800 rounded-lg p-8 text-center">
            <p class="text-zinc-400">No broadcasts yet.</p>
            <p class="text-zinc-500 text-sm mt-1">
                Sends from <code class="text-zinc-400">POST /broadcast</code> show up here as they run, with live progress.
            </p>
        </div>

        <AppPagination
            v-model:page="page"
            v-model:limit="limit"
            :total="total"
            @change="fetchBroadcasts" />
    </div>
</template>

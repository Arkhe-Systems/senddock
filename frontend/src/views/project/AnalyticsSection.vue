<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, computed } from 'vue'
import { api } from '@/api/client'
import type { Project } from '@/stores/projects'
import AppProPaywall from '@/components/ui/AppProPaywall.vue'

interface OpenBucket { bucket: string; opens: number }
interface TemplateStat { template_id: string; name: string; sends: number }
interface StatusBucket { status: string; count: number }
interface LinkStat { url: string; clicks: number }
interface BroadcastInFlight {
    id: string
    subject: string
    total: number
    sent: number
    failed: number
    suppressed: number
    remaining: number
    started_at: string
}
interface PeriodMetrics {
    total_sent: number
    total_failed: number
    total_opened: number
    total_clicked: number
    deliverability_pct: number
    open_rate_pct: number
    click_rate_pct: number
}
interface Overview {
    from: string
    to: string
    granularity: 'hour' | 'day' | 'week' | 'month'
    range_days: number
    total_sent: number
    total_failed: number
    total_opened: number
    total_clicked: number
    deliverability_pct: number
    open_rate_pct: number
    click_rate_pct: number
    click_to_open_pct: number
    opens_series: OpenBucket[] | null
    top_templates: TemplateStat[] | null
    top_clicked_links: LinkStat[] | null
    active_subscribers: number
    sends_by_status: StatusBucket[] | null
    previous: PeriodMetrics
    broadcasts_in_flight: BroadcastInFlight[] | null
}

const props = defineProps<{ project: Project }>()

const overview = ref<Overview | null>(null)
const loading = ref(true)
const errorState = ref<'none' | 'paywall' | 'generic'>('none')
type Preset = '24h' | '7d' | '30d' | '90d' | '1y' | 'custom'

const preset = ref<Preset>('30d')
const customFrom = ref('')
const customTo = ref('')
const showCustomPopover = ref(false)
const fromISO = ref('')
const toISO = ref('')

const PRESETS: { value: Preset; label: string }[] = [
    { value: '24h', label: '24h' },
    { value: '7d', label: '7d' },
    { value: '30d', label: '30d' },
    { value: '90d', label: '90d' },
    { value: '1y', label: '1y' },
]

function presetWindow(p: Preset): { from: Date; to: Date } | null {
    const now = new Date()
    const to = now
    let from: Date
    switch (p) {
        case '24h': from = new Date(now.getTime() - 24 * 60 * 60 * 1000); break
        case '7d':  from = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000); break
        case '30d': from = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000); break
        case '90d': from = new Date(now.getTime() - 90 * 24 * 60 * 60 * 1000); break
        case '1y':  from = new Date(now.getTime() - 365 * 24 * 60 * 60 * 1000); break
        default: return null
    }
    return { from, to }
}

const opensByDay = computed(() => overview.value?.opens_series ?? [])
const topTemplates = computed(() => overview.value?.top_templates ?? [])
const topClickedLinks = computed(() => overview.value?.top_clicked_links ?? [])
const broadcastsInFlight = computed(() => overview.value?.broadcasts_in_flight ?? [])

function progressPct(b: BroadcastInFlight): number {
    if (b.total <= 0) return 0
    const done = b.sent + b.failed + b.suppressed
    return Math.min(100, Math.round((done / b.total) * 100))
}

function fmtElapsed(startedAt: string): string {
    const s = Math.floor((Date.now() - new Date(startedAt).getTime()) / 1000)
    if (s < 60) return `${s}s`
    if (s < 3600) return `${Math.floor(s / 60)}m ${s % 60}s`
    const h = Math.floor(s / 3600)
    return `${h}h ${Math.floor((s % 3600) / 60)}m`
}

let pollTimer: ReturnType<typeof setInterval> | null = null

function startPollingIfNeeded() {
    if (pollTimer) return
    if (broadcastsInFlight.value.length === 0) return
    pollTimer = setInterval(async () => {
        if (broadcastsInFlight.value.length === 0) {
            stopPolling()
            return
        }
        await load()
    }, 5000)
}

function stopPolling() {
    if (pollTimer) {
        clearInterval(pollTimer)
        pollTimer = null
    }
}

onBeforeUnmount(() => stopPolling())
const maxLinkClicks = computed(() => topClickedLinks.value.reduce((m, l) => Math.max(m, l.clicks), 0) || 1)

const totalSendsCurrent = computed(() => {
    const o = overview.value
    if (!o) return 0
    return o.total_sent + o.total_failed
})

const maxTopSends = computed(() => topTemplates.value.reduce((m, t) => Math.max(m, t.sends), 0) || 1)

interface ChartGeometry {
    width: number
    height: number
    points: { x: number; y: number; bucket: string; opens: number }[]
    linePath: string
    areaPath: string
}

const chart = computed<ChartGeometry>(() => {
    const data = opensByDay.value
    const w = 600
    const h = 180
    if (data.length === 0) return { width: w, height: h, points: [], linePath: '', areaPath: '' }

    const max = Math.max(...data.map(d => d.opens), 1)
    const xStep = data.length === 1 ? 0 : w / (data.length - 1)
    const points = data.map((d, i) => ({
        x: data.length === 1 ? w / 2 : i * xStep,
        y: h - (d.opens / max) * (h - 24) - 12,
        bucket: d.bucket,
        opens: d.opens,
    }))

    const first = points[0]!
    const last = points[points.length - 1]!
    let line = `M ${first.x} ${first.y}`
    for (let i = 1; i < points.length; i++) {
        const p1 = points[i - 1]!
        const p2 = points[i]!
        const p0 = points[i - 2] ?? p1
        const p3 = points[i + 1] ?? p2
        const cp1x = p1.x + (p2.x - p0.x) / 6
        const cp1y = p1.y + (p2.y - p0.y) / 6
        const cp2x = p2.x - (p3.x - p1.x) / 6
        const cp2y = p2.y - (p3.y - p1.y) / 6
        line += ` C ${cp1x} ${cp1y}, ${cp2x} ${cp2y}, ${p2.x} ${p2.y}`
    }
    const area = `${line} L ${last.x} ${h} L ${first.x} ${h} Z`

    return { width: w, height: h, points, linePath: line, areaPath: area }
})

const statusDonut = computed(() => {
    const o = overview.value
    if (!o) return { sent: 0, failed: 0, total: 0, sentPct: 0 }
    const total = o.total_sent + o.total_failed
    return {
        sent: o.total_sent,
        failed: o.total_failed,
        total,
        sentPct: total === 0 ? 0 : (o.total_sent / total) * 100,
    }
})

interface Trend {
    direction: 'up' | 'down' | 'flat'
    pct: number
    label: string
}

function trend(current: number, previous: number, invertGood = false): Trend {
    if (current === 0 && previous === 0) {
        return { direction: 'flat', pct: 0, label: 'no change' }
    }
    if (previous === 0) {
        return { direction: 'up', pct: 100, label: 'new' }
    }
    const delta = ((current - previous) / previous) * 100
    if (Math.abs(delta) < 0.5) {
        return { direction: 'flat', pct: 0, label: 'flat' }
    }
    const dir = delta > 0 ? 'up' : 'down'
    return {
        direction: dir,
        pct: Math.abs(delta),
        label: invertGood
            ? (dir === 'up' ? 'worse' : 'better')
            : (dir === 'up' ? 'better' : 'worse'),
    }
}

const sentTrend = computed(() => overview.value ? trend(overview.value.total_sent, overview.value.previous.total_sent) : null)
const failedTrend = computed(() => overview.value ? trend(overview.value.total_failed, overview.value.previous.total_failed, true) : null)
const openedTrend = computed(() => overview.value ? trend(overview.value.total_opened, overview.value.previous.total_opened) : null)
const clickedTrend = computed(() => overview.value ? trend(overview.value.total_clicked, overview.value.previous.total_clicked) : null)
const openRateTrend = computed(() => overview.value ? trend(overview.value.open_rate_pct, overview.value.previous.open_rate_pct) : null)
const clickRateTrend = computed(() => overview.value ? trend(overview.value.click_rate_pct, overview.value.previous.click_rate_pct) : null)

const insights = computed(() => {
    const o = overview.value
    if (!o) return []
    const items: { tone: 'good' | 'warn' | 'info'; text: string }[] = []

    if (opensByDay.value.length > 0) {
        const top = [...opensByDay.value].sort((a, b) => b.opens - a.opens)[0]
        if (top && top.opens > 0) {
            const date = new Date(top.bucket)
            const weekday = date.toLocaleDateString('en-US', { weekday: 'long', timeZone: 'UTC' })
            const day = date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', timeZone: 'UTC' })
            const noun = o.granularity === 'hour' ? 'hour' : (o.granularity === 'week' ? 'week' : (o.granularity === 'month' ? 'month' : 'day'))
            items.push({
                tone: 'info',
                text: `Best ${noun} this period: ${weekday} ${day} with ${top.opens} open${top.opens === 1 ? '' : 's'}.`,
            })
        }
    }

    if (o.open_rate_pct > 0) {
        const benchmark = 21
        if (o.open_rate_pct >= benchmark) {
            items.push({
                tone: 'good',
                text: `Open rate of ${o.open_rate_pct.toFixed(1)}% is above the industry average (~21%). Keep it up.`,
            })
        } else {
            items.push({
                tone: 'warn',
                text: `Open rate of ${o.open_rate_pct.toFixed(1)}% is below the industry average (~21%). Try shorter subjects or A/B test send times.`,
            })
        }
    }

    if (o.total_failed > 0 && totalSendsCurrent.value > 0) {
        const failRate = (o.total_failed / totalSendsCurrent.value) * 100
        if (failRate >= 5) {
            items.push({
                tone: 'warn',
                text: `Failure rate of ${failRate.toFixed(1)}% is high — check SMTP credentials and recipient list quality.`,
            })
        }
    }

    const topTpl = topTemplates.value[0]
    if (topTpl) {
        items.push({
            tone: 'info',
            text: `Most-used template: ${topTpl.name} with ${topTpl.sends} send${topTpl.sends === 1 ? '' : 's'}.`,
        })
    }

    if (totalSendsCurrent.value === 0) {
        items.push({
            tone: 'info',
            text: 'No sends recorded yet in this range. Send a broadcast or transactional email to populate this view.',
        })
    }

    return items.slice(0, 4)
})

async function load() {
    if (!pollTimer) loading.value = true
    errorState.value = 'none'
    try {
        const params = new URLSearchParams({ from: fromISO.value, to: toISO.value })
        const url = `/projects/${props.project.id}/analytics/overview?${params.toString()}`
        overview.value = await api<Overview>(url)
        startPollingIfNeeded()
    } catch (e) {
        const msg = e instanceof Error ? e.message : ''
        if (msg.includes('license required')) errorState.value = 'paywall'
        else errorState.value = 'generic'
    } finally {
        loading.value = false
    }
}

function applyPreset(p: Preset) {
    if (p === 'custom') {
        showCustomPopover.value = true
        return
    }
    preset.value = p
    showCustomPopover.value = false
    const w = presetWindow(p)
    if (!w) return
    fromISO.value = w.from.toISOString()
    toISO.value = w.to.toISOString()
    load()
}

function applyCustom() {
    if (!customFrom.value || !customTo.value) return
    const from = new Date(customFrom.value + 'T00:00:00Z')
    const to = new Date(customTo.value + 'T23:59:59Z')
    if (!(to > from)) return
    preset.value = 'custom'
    fromISO.value = from.toISOString()
    toISO.value = to.toISOString()
    showCustomPopover.value = false
    load()
}

function fmtPct(n: number | undefined) {
    if (n === undefined || Number.isNaN(n)) return '0.0%'
    return `${n.toFixed(1)}%`
}

function fmtBucket(iso: string, granularity: string) {
    const d = new Date(iso)
    if (granularity === 'hour') {
        return d.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', hour12: false, timeZone: 'UTC' })
    }
    if (granularity === 'month') {
        return d.toLocaleDateString('en-US', { month: 'short', year: '2-digit', timeZone: 'UTC' })
    }
    return d.toLocaleDateString('en-US', { day: '2-digit', month: 'short', timeZone: 'UTC' })
}

function fmtRange() {
    if (!overview.value) return ''
    const f = new Date(overview.value.from)
    const t = new Date(overview.value.to)
    const fStr = f.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
    const tStr = t.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
    if (fStr === tStr) return fStr
    return `${fStr} → ${tStr}`
}

function rangeLabel() {
    return `vs previous ${preset.value === 'custom' ? 'period' : preset.value}`
}

onMounted(() => applyPreset(preset.value))
</script>

<template>
    <div>
        <div class="flex flex-wrap items-center justify-between gap-3 mb-6">
            <div>
                <div class="flex items-center gap-2">
                    <h1 class="text-2xl font-bold text-white">Analytics</h1>
                    <span class="text-[10px] font-semibold tracking-wider uppercase px-1.5 py-0.5 rounded bg-amber-500/15 text-amber-400 border border-amber-500/30">Pro</span>
                </div>
                <p class="text-sm text-zinc-500 mt-1">Send performance and engagement trends</p>
            </div>
            <div class="relative">
                <div class="flex gap-1 bg-zinc-900 border border-zinc-800 rounded-lg p-1">
                    <button v-for="p in PRESETS" :key="p.value" @click="applyPreset(p.value)" :class="[
                        'px-3 py-1 text-sm rounded-md transition cursor-pointer',
                        preset === p.value ? 'bg-zinc-800 text-white' : 'text-zinc-400 hover:text-white'
                    ]">{{ p.label }}</button>
                    <button @click="applyPreset('custom')" :class="[
                        'px-3 py-1 text-sm rounded-md transition cursor-pointer',
                        preset === 'custom' ? 'bg-zinc-800 text-white' : 'text-zinc-400 hover:text-white'
                    ]">Custom</button>
                </div>
                <div v-if="showCustomPopover"
                    class="absolute right-0 top-full mt-2 z-20 bg-zinc-900 border border-zinc-800 rounded-lg shadow-xl p-4 w-72">
                    <p class="text-xs font-semibold text-zinc-400 uppercase tracking-wide mb-3">Custom range</p>
                    <label class="block text-xs text-zinc-500 mb-1">From</label>
                    <input v-model="customFrom" type="date"
                        class="w-full mb-3 px-3 py-2 text-sm bg-zinc-950 border border-zinc-800 rounded-md text-white focus:border-zinc-600 focus:outline-none" />
                    <label class="block text-xs text-zinc-500 mb-1">To</label>
                    <input v-model="customTo" type="date"
                        class="w-full mb-4 px-3 py-2 text-sm bg-zinc-950 border border-zinc-800 rounded-md text-white focus:border-zinc-600 focus:outline-none" />
                    <div class="flex gap-2">
                        <button @click="showCustomPopover = false"
                            class="flex-1 px-3 py-1.5 text-sm rounded-md text-zinc-300 hover:bg-zinc-800 transition cursor-pointer">Cancel</button>
                        <button @click="applyCustom" :disabled="!customFrom || !customTo"
                            class="flex-1 px-3 py-1.5 text-sm rounded-md bg-white text-zinc-950 hover:bg-zinc-200 transition cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed">Apply</button>
                    </div>
                </div>
            </div>
        </div>

        <div v-if="loading" class="text-zinc-500 py-8 text-center">Loading...</div>

        <AppProPaywall v-else-if="errorState === 'paywall'"
            title="Analytics is a Pro feature"
            description="Detailed deliverability metrics, open tracking trends and top-template insights require a SendDock Pro license. Activate one to unlock this section." />

        <div v-else-if="errorState === 'generic'"
            class="bg-zinc-900 border border-zinc-800 rounded-lg p-6 text-center">
            <p class="text-zinc-400 text-sm">Couldn't load analytics. Try again later.</p>
            <button @click="load" class="mt-3 px-3 py-1 text-sm text-white border border-zinc-700 rounded-md hover:bg-zinc-800 cursor-pointer">Retry</button>
        </div>

        <div v-else-if="overview">
            <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4 mb-6">
                <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-5">
                    <p class="text-xs text-zinc-500 uppercase tracking-wide">Sent</p>
                    <p class="text-3xl font-bold text-white mt-2 tabular-nums">{{ overview.total_sent }}</p>
                    <div v-if="sentTrend" class="mt-3 flex items-center gap-1.5">
                        <span :class="[
                            'inline-flex items-center gap-0.5 text-xs font-medium px-1.5 py-0.5 rounded',
                            sentTrend.direction === 'up' ? 'bg-emerald-500/15 text-emerald-400'
                                : sentTrend.direction === 'down' ? 'bg-red-500/15 text-red-400'
                                : 'bg-zinc-800 text-zinc-400'
                        ]">
                            <span v-if="sentTrend.direction === 'up'">▲</span>
                            <span v-else-if="sentTrend.direction === 'down'">▼</span>
                            <span v-else>—</span>
                            <span>{{ sentTrend.pct.toFixed(0) }}%</span>
                        </span>
                        <span class="text-xs text-zinc-500">{{ rangeLabel() }}</span>
                    </div>
                </div>

                <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-5">
                    <p class="text-xs text-zinc-500 uppercase tracking-wide">Failed</p>
                    <p class="text-3xl font-bold mt-2 tabular-nums" :class="overview.total_failed > 0 ? 'text-red-400' : 'text-white'">{{ overview.total_failed }}</p>
                    <div v-if="failedTrend" class="mt-3 flex items-center gap-1.5">
                        <span :class="[
                            'inline-flex items-center gap-0.5 text-xs font-medium px-1.5 py-0.5 rounded',
                            failedTrend.direction === 'down' ? 'bg-emerald-500/15 text-emerald-400'
                                : failedTrend.direction === 'up' ? 'bg-red-500/15 text-red-400'
                                : 'bg-zinc-800 text-zinc-400'
                        ]">
                            <span v-if="failedTrend.direction === 'up'">▲</span>
                            <span v-else-if="failedTrend.direction === 'down'">▼</span>
                            <span v-else>—</span>
                            <span>{{ failedTrend.pct.toFixed(0) }}%</span>
                        </span>
                        <span class="text-xs text-zinc-500">{{ rangeLabel() }}</span>
                    </div>
                </div>

                <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-5">
                    <p class="text-xs text-zinc-500 uppercase tracking-wide">Opened</p>
                    <p class="text-3xl font-bold text-white mt-2 tabular-nums">{{ overview.total_opened }}</p>
                    <div v-if="openedTrend" class="mt-3 flex items-center gap-1.5">
                        <span :class="[
                            'inline-flex items-center gap-0.5 text-xs font-medium px-1.5 py-0.5 rounded',
                            openedTrend.direction === 'up' ? 'bg-emerald-500/15 text-emerald-400'
                                : openedTrend.direction === 'down' ? 'bg-red-500/15 text-red-400'
                                : 'bg-zinc-800 text-zinc-400'
                        ]">
                            <span v-if="openedTrend.direction === 'up'">▲</span>
                            <span v-else-if="openedTrend.direction === 'down'">▼</span>
                            <span v-else>—</span>
                            <span>{{ openedTrend.pct.toFixed(0) }}%</span>
                        </span>
                        <span class="text-xs text-zinc-500">{{ rangeLabel() }}</span>
                    </div>
                </div>

                <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-5">
                    <p class="text-xs text-zinc-500 uppercase tracking-wide">Clicked</p>
                    <p class="text-3xl font-bold text-white mt-2 tabular-nums">{{ overview.total_clicked }}</p>
                    <div v-if="clickedTrend" class="mt-3 flex items-center gap-1.5">
                        <span :class="[
                            'inline-flex items-center gap-0.5 text-xs font-medium px-1.5 py-0.5 rounded',
                            clickedTrend.direction === 'up' ? 'bg-emerald-500/15 text-emerald-400'
                                : clickedTrend.direction === 'down' ? 'bg-red-500/15 text-red-400'
                                : 'bg-zinc-800 text-zinc-400'
                        ]">
                            <span v-if="clickedTrend.direction === 'up'">▲</span>
                            <span v-else-if="clickedTrend.direction === 'down'">▼</span>
                            <span v-else>—</span>
                            <span>{{ clickedTrend.pct.toFixed(0) }}%</span>
                        </span>
                        <span class="text-xs text-zinc-500">{{ rangeLabel() }}</span>
                    </div>
                </div>

                <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-5">
                    <p class="text-xs text-zinc-500 uppercase tracking-wide">Open rate</p>
                    <p class="text-3xl font-bold text-white mt-2 tabular-nums">{{ fmtPct(overview.open_rate_pct) }}</p>
                    <div v-if="openRateTrend" class="mt-3 flex items-center gap-1.5">
                        <span :class="[
                            'inline-flex items-center gap-0.5 text-xs font-medium px-1.5 py-0.5 rounded',
                            openRateTrend.direction === 'up' ? 'bg-emerald-500/15 text-emerald-400'
                                : openRateTrend.direction === 'down' ? 'bg-red-500/15 text-red-400'
                                : 'bg-zinc-800 text-zinc-400'
                        ]">
                            <span v-if="openRateTrend.direction === 'up'">▲</span>
                            <span v-else-if="openRateTrend.direction === 'down'">▼</span>
                            <span v-else>—</span>
                            <span>{{ openRateTrend.pct.toFixed(0) }}%</span>
                        </span>
                        <span class="text-xs text-zinc-500">{{ rangeLabel() }}</span>
                    </div>
                </div>

                <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-5">
                    <p class="text-xs text-zinc-500 uppercase tracking-wide">Click rate</p>
                    <p class="text-3xl font-bold text-white mt-2 tabular-nums">{{ fmtPct(overview.click_rate_pct) }}</p>
                    <div v-if="clickRateTrend" class="mt-3 flex items-center gap-1.5">
                        <span :class="[
                            'inline-flex items-center gap-0.5 text-xs font-medium px-1.5 py-0.5 rounded',
                            clickRateTrend.direction === 'up' ? 'bg-emerald-500/15 text-emerald-400'
                                : clickRateTrend.direction === 'down' ? 'bg-red-500/15 text-red-400'
                                : 'bg-zinc-800 text-zinc-400'
                        ]">
                            <span v-if="clickRateTrend.direction === 'up'">▲</span>
                            <span v-else-if="clickRateTrend.direction === 'down'">▼</span>
                            <span v-else>—</span>
                            <span>{{ clickRateTrend.pct.toFixed(0) }}%</span>
                        </span>
                        <span class="text-xs text-zinc-500">{{ rangeLabel() }}</span>
                    </div>
                </div>
            </div>

            <div v-if="broadcastsInFlight.length > 0" class="bg-zinc-900 border border-blue-500/30 rounded-xl p-5 mb-6">
                <div class="flex items-center justify-between mb-4">
                    <div class="flex items-center gap-2">
                        <span class="relative flex h-2.5 w-2.5">
                            <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75"></span>
                            <span class="relative inline-flex rounded-full h-2.5 w-2.5 bg-blue-500"></span>
                        </span>
                        <h2 class="text-sm font-semibold text-white">
                            {{ broadcastsInFlight.length === 1 ? '1 broadcast in flight' : `${broadcastsInFlight.length} broadcasts in flight` }}
                        </h2>
                    </div>
                    <span class="text-xs text-zinc-500">Live · refreshing every 5s</span>
                </div>
                <div class="space-y-3">
                    <div v-for="b in broadcastsInFlight" :key="b.id" class="bg-zinc-950 border border-zinc-800 rounded-lg p-4">
                        <div class="flex items-start justify-between gap-3 mb-2">
                            <div class="min-w-0 flex-1">
                                <p class="text-sm font-medium text-white truncate">{{ b.subject || '(no subject)' }}</p>
                                <p class="text-xs text-zinc-500 mt-0.5">Started {{ fmtElapsed(b.started_at) }} ago</p>
                            </div>
                            <div class="text-right shrink-0">
                                <p class="text-sm font-semibold text-white tabular-nums">{{ b.sent + b.failed + b.suppressed }} / {{ b.total }}</p>
                                <p class="text-xs text-zinc-500">{{ progressPct(b) }}%</p>
                            </div>
                        </div>
                        <div class="h-2 bg-zinc-900 rounded-full overflow-hidden">
                            <div class="h-full bg-blue-500 transition-all duration-500" :style="{ width: progressPct(b) + '%' }"></div>
                        </div>
                        <div class="flex items-center gap-4 mt-2 text-xs">
                            <span class="text-emerald-400 tabular-nums">✓ {{ b.sent }} sent</span>
                            <span v-if="b.failed > 0" class="text-red-400 tabular-nums">✗ {{ b.failed }} failed</span>
                            <span v-if="b.suppressed > 0" class="text-zinc-500 tabular-nums">⊘ {{ b.suppressed }} suppressed</span>
                            <span v-if="b.remaining > 0" class="text-zinc-500 tabular-nums ml-auto">{{ b.remaining }} pending</span>
                        </div>
                    </div>
                </div>
            </div>

            <div class="grid grid-cols-1 lg:grid-cols-3 gap-4 mb-6">
                <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-6 lg:col-span-2">
                    <h2 class="text-sm font-semibold text-white mb-1">Conversion funnel</h2>
                    <p class="text-xs text-zinc-500 mb-5">Send → delivered → opened, {{ fmtRange() }}</p>

                    <div v-if="totalSendsCurrent === 0" class="text-center text-sm text-zinc-500 py-8">
                        No sends in this range yet.
                    </div>
                    <div v-else class="space-y-3">
                        <div>
                            <div class="flex items-baseline justify-between mb-1.5">
                                <span class="text-sm text-zinc-300">Sent</span>
                                <span class="text-sm text-white tabular-nums font-medium">{{ totalSendsCurrent }}</span>
                            </div>
                            <div class="h-7 rounded-md bg-gradient-to-r from-zinc-700 to-zinc-600" style="width: 100%"></div>
                        </div>
                        <div>
                            <div class="flex items-baseline justify-between mb-1.5">
                                <span class="text-sm text-zinc-300">Delivered <span class="text-zinc-500 ml-1">{{ fmtPct(overview.deliverability_pct) }}</span></span>
                                <span class="text-sm text-white tabular-nums font-medium">{{ overview.total_sent }}</span>
                            </div>
                            <div class="h-7 rounded-md bg-gradient-to-r from-emerald-600 to-emerald-500"
                                :style="{ width: (overview.deliverability_pct) + '%' }"></div>
                        </div>
                        <div>
                            <div class="flex items-baseline justify-between mb-1.5">
                                <span class="text-sm text-zinc-300">Opened <span class="text-zinc-500 ml-1">{{ fmtPct(overview.open_rate_pct) }}</span></span>
                                <span class="text-sm text-white tabular-nums font-medium">{{ overview.total_opened }}</span>
                            </div>
                            <div class="h-7 rounded-md bg-gradient-to-r from-sky-600 to-sky-400"
                                :style="{ width: ((overview.total_opened / Math.max(totalSendsCurrent, 1)) * 100) + '%' }"></div>
                        </div>
                        <div>
                            <div class="flex items-baseline justify-between mb-1.5">
                                <span class="text-sm text-zinc-300">Clicked <span class="text-zinc-500 ml-1">{{ fmtPct(overview.click_rate_pct) }}</span><span v-if="overview.click_to_open_pct > 0" class="text-zinc-600 ml-2 text-xs">CTOR {{ fmtPct(overview.click_to_open_pct) }}</span></span>
                                <span class="text-sm text-white tabular-nums font-medium">{{ overview.total_clicked }}</span>
                            </div>
                            <div class="h-7 rounded-md bg-gradient-to-r from-fuchsia-600 to-violet-400"
                                :style="{ width: ((overview.total_clicked / Math.max(totalSendsCurrent, 1)) * 100) + '%' }"></div>
                        </div>
                    </div>
                </div>

                <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-6">
                    <h2 class="text-sm font-semibold text-white mb-1">Insights</h2>
                    <p class="text-xs text-zinc-500 mb-4">Auto-generated from your data</p>
                    <ul v-if="insights.length > 0" class="space-y-3">
                        <li v-for="(it, idx) in insights" :key="idx" class="flex gap-2.5 text-sm">
                            <span class="mt-0.5 inline-block w-1.5 h-1.5 rounded-full flex-shrink-0" :class="{
                                'bg-emerald-400': it.tone === 'good',
                                'bg-amber-400': it.tone === 'warn',
                                'bg-sky-400': it.tone === 'info'
                            }"></span>
                            <span class="text-zinc-300 leading-snug">{{ it.text }}</span>
                        </li>
                    </ul>
                    <p v-else class="text-sm text-zinc-500">No insights yet.</p>
                </div>
            </div>

            <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-6 mb-6">
                <div class="flex items-baseline justify-between mb-5">
                    <div>
                        <h2 class="text-sm font-semibold text-white">Opens over time</h2>
                        <p class="text-xs text-zinc-500 mt-0.5">{{ fmtRange() }}</p>
                    </div>
                </div>

                <div v-if="opensByDay.length === 0" class="text-center text-sm text-zinc-500 py-12">
                    No opens yet in this range.
                </div>

                <div v-else>
                    <svg :viewBox="`0 0 ${chart.width} ${chart.height}`" class="w-full h-48 overflow-visible">
                        <defs>
                            <linearGradient id="opensGradient" x1="0" y1="0" x2="0" y2="1">
                                <stop offset="0%" stop-color="rgb(16, 185, 129)" stop-opacity="0.45" />
                                <stop offset="100%" stop-color="rgb(16, 185, 129)" stop-opacity="0" />
                            </linearGradient>
                        </defs>
                        <line v-for="g in [0.25, 0.5, 0.75]" :key="g" x1="0" :x2="chart.width"
                            :y1="g * chart.height" :y2="g * chart.height" stroke="rgb(63, 63, 70)" stroke-width="0.5" stroke-dasharray="3 4" />
                        <path :d="chart.areaPath" fill="url(#opensGradient)" />
                        <path :d="chart.linePath" fill="none" stroke="rgb(16, 185, 129)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
                        <g v-for="p in chart.points" :key="p.bucket" class="group">
                            <circle :cx="p.x" :cy="p.y" r="3" fill="rgb(16, 185, 129)"
                                stroke="rgb(24, 24, 27)" stroke-width="2"
                                class="opacity-0 group-hover:opacity-100 transition" />
                            <rect :x="p.x - 18" y="0" width="36" :height="chart.height" fill="transparent"
                                class="cursor-pointer">
                                <title>{{ p.bucket }}: {{ p.opens }} opens</title>
                            </rect>
                        </g>
                    </svg>
                    <div class="flex justify-between text-[10px] text-zinc-500 mt-1 px-1">
                        <span v-for="(p, i) in chart.points" :key="p.bucket"
                            v-show="chart.points.length <= 14 || i % Math.ceil(chart.points.length / 8) === 0">
                            {{ fmtBucket(p.bucket, overview.granularity) }}
                        </span>
                    </div>
                </div>
            </div>

            <div class="grid grid-cols-1 lg:grid-cols-2 gap-4 mb-4">
                <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-6">
                    <h2 class="text-sm font-semibold text-white mb-1">Top templates</h2>
                    <p class="text-xs text-zinc-500 mb-4">Most-used templates this period</p>
                    <div v-if="topTemplates.length === 0" class="text-center text-sm text-zinc-500 py-8">
                        No template sends recorded yet.
                    </div>
                    <ul v-else class="space-y-3">
                        <li v-for="t in topTemplates" :key="t.template_id">
                            <div class="flex items-baseline justify-between mb-1">
                                <span class="text-sm text-white truncate pr-2">{{ t.name }}</span>
                                <span class="text-sm text-zinc-300 tabular-nums">{{ t.sends }}</span>
                            </div>
                            <div class="h-1.5 rounded-full bg-zinc-800 overflow-hidden">
                                <div class="h-full bg-gradient-to-r from-emerald-500 to-sky-500 rounded-full"
                                    :style="{ width: ((t.sends / maxTopSends) * 100) + '%' }"></div>
                            </div>
                        </li>
                    </ul>
                </div>

                <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-6">
                    <h2 class="text-sm font-semibold text-white mb-1">Top clicked links</h2>
                    <p class="text-xs text-zinc-500 mb-4">Most-clicked links this period</p>
                    <div v-if="topClickedLinks.length === 0" class="text-center text-sm text-zinc-500 py-8">
                        No clicks recorded yet.
                    </div>
                    <ul v-else class="space-y-3">
                        <li v-for="l in topClickedLinks" :key="l.url">
                            <div class="flex items-baseline justify-between mb-1 gap-2">
                                <a :href="l.url" target="_blank" rel="noopener noreferrer"
                                    class="text-sm text-white truncate hover:text-sky-400 transition" :title="l.url">{{ l.url }}</a>
                                <span class="text-sm text-zinc-300 tabular-nums flex-shrink-0">{{ l.clicks }}</span>
                            </div>
                            <div class="h-1.5 rounded-full bg-zinc-800 overflow-hidden">
                                <div class="h-full bg-gradient-to-r from-fuchsia-500 to-violet-500 rounded-full"
                                    :style="{ width: ((l.clicks / maxLinkClicks) * 100) + '%' }"></div>
                            </div>
                        </li>
                    </ul>
                </div>
            </div>

            <div class="grid grid-cols-1 gap-4">
                <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-6">
                    <h2 class="text-sm font-semibold text-white mb-1">Send status</h2>
                    <p class="text-xs text-zinc-500 mb-5">Sent vs failed</p>
                    <div v-if="statusDonut.total === 0" class="text-center text-sm text-zinc-500 py-8">
                        No sends yet.
                    </div>
                    <div v-else class="flex items-center gap-5">
                        <svg viewBox="0 0 100 100" class="w-28 h-28 -rotate-90">
                            <circle cx="50" cy="50" r="40" fill="none" stroke="rgb(239, 68, 68)" stroke-width="14" />
                            <circle cx="50" cy="50" r="40" fill="none" stroke="rgb(16, 185, 129)" stroke-width="14"
                                pathLength="100"
                                :stroke-dasharray="`${statusDonut.sentPct} 100`" />
                            <circle cx="50" cy="50" r="28" fill="rgb(24, 24, 27)" />
                        </svg>
                        <div class="space-y-2 flex-1">
                            <div class="flex items-center gap-2">
                                <span class="w-2.5 h-2.5 rounded-full bg-emerald-500"></span>
                                <span class="text-sm text-zinc-300 flex-1">Sent</span>
                                <span class="text-sm text-white tabular-nums">{{ statusDonut.sent }}</span>
                            </div>
                            <div class="flex items-center gap-2">
                                <span class="w-2.5 h-2.5 rounded-full bg-red-500"></span>
                                <span class="text-sm text-zinc-300 flex-1">Failed</span>
                                <span class="text-sm text-white tabular-nums">{{ statusDonut.failed }}</span>
                            </div>
                            <div class="pt-2 border-t border-zinc-800 mt-2">
                                <div class="flex items-baseline justify-between">
                                    <span class="text-xs text-zinc-500 uppercase tracking-wide">Active subs</span>
                                    <span class="text-sm text-white tabular-nums font-medium">{{ overview.active_subscribers }}</span>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import type { Project } from '@/stores/projects'
import { useLicenseStore } from '@/stores/license'
import { useSegmentStore } from '@/stores/segments'
import {
    useAnalyticsStore, type Overview, type CampaignStat, type CampaignDetail,
    type Audience, type Engagement, type DomainHealth, type BouncesByProvider,
} from '@/stores/analytics'
import { ApiError } from '@/api/client'
import { chartColors } from '@/components/charts/chartSetup'
import LineChart from '@/components/charts/LineChart.vue'
import BarChart from '@/components/charts/BarChart.vue'
import DonutChart from '@/components/charts/DonutChart.vue'
import AppHeatmap from '@/components/charts/AppHeatmap.vue'
import AppStatTile from '@/components/ui/AppStatTile.vue'
import AppStatusPill from '@/components/ui/AppStatusPill.vue'
import AppProPaywall from '@/components/ui/AppProPaywall.vue'
import AppLoader from '@/components/ui/AppLoader.vue'
import AppAlert from '@/components/ui/AppAlert.vue'

const props = defineProps<{ project: Project }>()

const licenseStore = useLicenseStore()
const segmentStore = useSegmentStore()
const analytics = useAnalyticsStore()

type Tab = 'overview' | 'campaigns' | 'audience' | 'engagement' | 'deliverability'
const tab = ref<Tab>('overview')
const TABS: { key: Tab; label: string; pro?: boolean }[] = [
    { key: 'overview', label: 'Overview' },
    { key: 'campaigns', label: 'Campaigns' },
    { key: 'audience', label: 'Audience' },
    { key: 'engagement', label: 'Engagement' },
    { key: 'deliverability', label: 'Deliverability', pro: true },
]

// --- date window + segment filter (kept from the previous implementation) ---
type Preset = '24h' | '7d' | '30d' | '90d' | '1y'
const PRESETS: { value: Preset; label: string }[] = [
    { value: '24h', label: '24h' }, { value: '7d', label: '7d' },
    { value: '30d', label: '30d' }, { value: '90d', label: '90d' }, { value: '1y', label: '1y' },
]
const preset = ref<Preset>('30d')
const fromISO = ref('')
const toISO = ref('')

function presetWindow(p: Preset): { from: Date; to: Date } {
    const to = new Date()
    const from = new Date(to)
    switch (p) {
        case '24h': from.setDate(to.getDate() - 1); break
        case '7d': from.setDate(to.getDate() - 7); break
        case '30d': from.setDate(to.getDate() - 30); break
        case '90d': from.setDate(to.getDate() - 90); break
        case '1y': from.setFullYear(to.getFullYear() - 1); break
    }
    return { from, to }
}

const selectedSegment = ref('')
const segments = computed(() => segmentStore.segments(props.project.id))

const errorState = ref<'none' | 'generic'>('none')
const loading = ref(false)

const overview = ref<Overview | null>(null)
const campaigns = ref<CampaignStat[]>([])
const openCampaign = ref<CampaignDetail | null>(null)
const audience = ref<Audience | null>(null)
const engagement = ref<Engagement | null>(null)
const domainHealth = ref<DomainHealth | null>(null)
const bounces = ref<BouncesByProvider | null>(null)

function handleError(e: unknown) {
    // The descriptive tabs are free; only deliverability can 402, and that path
    // is gated by allowsPro before we ever fetch — so anything here is a real error.
    errorState.value = 'generic'
    if (e instanceof ApiError) { /* keep last good data */ }
}

async function loadCurrentTab() {
    loading.value = true
    errorState.value = 'none'
    try {
        if (tab.value === 'overview') {
            overview.value = await analytics.overview(props.project.id, fromISO.value, toISO.value, selectedSegment.value || undefined)
        } else if (tab.value === 'campaigns') {
            const res = await analytics.campaigns(props.project.id)
            campaigns.value = res.campaigns
        } else if (tab.value === 'audience') {
            audience.value = await analytics.audience(props.project.id, fromISO.value, toISO.value)
        } else if (tab.value === 'engagement') {
            engagement.value = await analytics.engagement(props.project.id, fromISO.value, toISO.value)
        } else if (tab.value === 'deliverability') {
            if (!licenseStore.allowsPro) return
            const [dh, bp] = await Promise.all([
                analytics.domainHealth(props.project.id),
                analytics.bouncesByProvider(props.project.id, fromISO.value, toISO.value),
            ])
            domainHealth.value = dh
            bounces.value = bp
        }
    } catch (e) {
        handleError(e)
    } finally {
        loading.value = false
    }
}

function applyPreset(p: Preset) {
    preset.value = p
    const { from, to } = presetWindow(p)
    fromISO.value = from.toISOString()
    toISO.value = to.toISOString()
    loadCurrentTab()
}

function switchTab(t: Tab) {
    tab.value = t
    openCampaign.value = null
    loadCurrentTab()
}

async function viewCampaign(c: CampaignStat) {
    loading.value = true
    try {
        openCampaign.value = await analytics.campaign(props.project.id, c.broadcast_id)
    } catch (e) {
        handleError(e)
    } finally {
        loading.value = false
    }
}

function exportCsv() {
    window.open(analytics.exportUrl(props.project.id), '_blank')
}

// --- overview formatting helpers ---
function fmtBucket(iso: string): string {
    const d = new Date(iso)
    return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
}
function pct(v: number): string { return `${v.toFixed(1)}%` }
function trend(cur: number, prev: number): number | null {
    if (prev === 0) return cur === 0 ? 0 : null
    return ((cur - prev) / prev) * 100
}

const opensLine = computed(() => {
    const o = overview.value
    if (!o) return { labels: [] as string[], series: [] as { label: string; values: number[]; color: string; fill: boolean }[] }
    const labels = o.opens_series.map((b) => fmtBucket(b.bucket))
    return {
        labels,
        series: [
            { label: 'Opens', values: o.opens_series.map((b) => b.opens), color: chartColors.indigo, fill: true },
            { label: 'Clicks', values: o.clicks_series.map((b) => b.clicks), color: chartColors.emerald, fill: false },
        ],
    }
})

const statusDonut = computed(() => {
    const o = overview.value
    if (!o) return { labels: [] as string[], values: [] as number[], colors: [] as string[] }
    return {
        labels: ['Sent', 'Bounced', 'Failed'],
        values: [o.total_sent, o.total_bounced, o.total_failed],
        colors: [chartColors.emerald, chartColors.amber, chartColors.red],
    }
})

const audienceChart = computed(() => {
    const a = audience.value
    if (!a) return { labels: [] as string[], values: [] as number[] }
    return { labels: a.series.map((b) => fmtBucket(b.bucket)), values: a.series.map((b) => b.cumulative_net) }
})

const bouncesDonut = computed(() => {
    const b = bounces.value
    if (!b || !b.providers.length) return { labels: [] as string[], values: [] as number[] }
    return { labels: b.providers.map((p) => p.provider), values: b.providers.map((p) => p.total) }
})

// --- engagement views ---
const engView = ref<'donut' | 'bars'>('donut')

const funnelRows = computed(() => {
    const f = engagement.value?.funnel
    if (!f) return [] as { label: string; count: number; pct: number }[]
    const base = f.sent || 1
    return [
        { label: 'Sent', count: f.sent, pct: 100 },
        { label: 'Opened', count: f.opened, pct: (f.opened / base) * 100 },
        { label: 'Clicked', count: f.clicked, pct: (f.clicked / base) * 100 },
    ]
})

function breakdownPct(items: { label: string; count: number }[]) {
    const total = items.reduce((s, i) => s + i.count, 0) || 1
    return items.map((i) => ({ ...i, pct: (i.count / total) * 100 }))
}

const activityLine = computed(() => {
    const e = engagement.value
    if (!e) return { labels: [] as string[], series: [] as { label: string; values: number[]; color: string; fill: boolean }[] }
    return {
        labels: e.series.map((b) => fmtBucket(b.bucket)),
        series: [
            { label: 'Opens', values: e.series.map((b) => b.opens), color: chartColors.indigo, fill: true },
            { label: 'Clicks', values: e.series.map((b) => b.clicks), color: chartColors.emerald, fill: false },
        ],
    }
})

// --- broadcasts-in-flight polling (kept) ---
let pollTimer: ReturnType<typeof setInterval> | null = null
const inFlight = computed(() => overview.value?.broadcasts_in_flight ?? [])
function progressPct(b: { total: number; sent: number; failed: number; suppressed: number }): number {
    if (b.total <= 0) return 0
    return Math.min(100, Math.round(((b.sent + b.failed + b.suppressed) / b.total) * 100))
}
function startPolling() {
    if (pollTimer) return
    pollTimer = setInterval(async () => {
        if (tab.value !== 'overview') return
        try {
            overview.value = await analytics.overview(props.project.id, fromISO.value, toISO.value, selectedSegment.value || undefined)
        } catch { /* keep last good data */ }
    }, 5000)
}
onBeforeUnmount(() => { if (pollTimer) clearInterval(pollTimer) })

onMounted(async () => {
    await licenseStore.fetch()
    segmentStore.fetchSegments(props.project.id)
    applyPreset(preset.value)
    startPolling()
})
</script>

<template>
    <div>
        <div class="flex items-center justify-between mb-6 flex-wrap gap-3">
            <div>
                <h2 class="text-2xl font-bold text-white">Analytics</h2>
                <p class="text-sm text-zinc-500 mt-1">Send performance, campaigns and audience</p>
            </div>
            <div class="flex items-center gap-2 flex-wrap">
                <select v-if="segments.length" v-model="selectedSegment" @change="loadCurrentTab"
                    class="px-3 py-1.5 text-sm bg-zinc-900 border border-zinc-800 rounded-lg text-white focus:outline-none focus:ring-1 focus:ring-zinc-500 cursor-pointer">
                    <option value="">All subscribers</option>
                    <option v-for="s in segments" :key="s.id" :value="s.id">{{ s.name }}</option>
                </select>
                <div class="flex gap-1 bg-zinc-900 border border-zinc-800 rounded-lg p-1">
                    <button v-for="p in PRESETS" :key="p.value" @click="applyPreset(p.value)"
                        :class="['px-2.5 py-1 text-sm rounded-md transition', preset === p.value ? 'bg-zinc-800 text-white' : 'text-zinc-400 hover:text-white']">
                        {{ p.label }}
                    </button>
                </div>
            </div>
        </div>

        <div class="flex gap-1 border-b border-zinc-800 mb-6">
            <button v-for="t in TABS" :key="t.key" @click="switchTab(t.key)"
                :class="['px-4 py-2 text-sm font-medium border-b-2 -mb-px transition flex items-center gap-1.5',
                    tab === t.key ? 'border-indigo-400 text-white' : 'border-transparent text-zinc-400 hover:text-white']">
                {{ t.label }}
                <span v-if="t.pro" class="text-[9px] font-semibold tracking-wider uppercase px-1 py-0.5 rounded bg-amber-500/15 text-amber-400 border border-amber-500/30">Pro</span>
            </button>
        </div>

            <AppAlert v-if="errorState === 'generic'" type="error" message="Could not load analytics. Try again." class="mb-4" />
            <AppLoader v-if="loading" message="Loading…" />

            <!-- OVERVIEW -->
            <div v-else-if="tab === 'overview' && overview" class="space-y-6">
                <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-3">
                    <AppStatTile label="Sent" :value="overview.total_sent" :trend="trend(overview.total_sent, overview.previous.total_sent)" />
                    <AppStatTile label="Acceptance" :value="pct(overview.acceptance_pct)" :trend="trend(overview.acceptance_pct, overview.previous.acceptance_pct)" hint="accepted by relay" />
                    <AppStatTile label="Bounce rate" :value="pct(overview.bounce_rate_pct)" :trend="trend(overview.bounce_rate_pct, overview.previous.bounce_rate_pct)" invert-good />
                    <AppStatTile label="Open rate" :value="pct(overview.open_rate_pct)" :trend="trend(overview.open_rate_pct, overview.previous.open_rate_pct)" />
                    <AppStatTile label="Click rate" :value="pct(overview.click_rate_pct)" :trend="trend(overview.click_rate_pct, overview.previous.click_rate_pct)" />
                    <AppStatTile label="Active subs" :value="overview.active_subscribers" />
                </div>

                <div v-if="inFlight.length" class="space-y-2">
                    <p class="text-xs font-semibold text-zinc-400 uppercase tracking-wide">Sending now</p>
                    <div v-for="b in inFlight" :key="b.id" class="bg-zinc-900 border border-zinc-800 rounded-lg p-3">
                        <div class="flex justify-between text-sm mb-1">
                            <span class="text-white truncate">{{ b.subject }}</span>
                            <span class="text-zinc-400">{{ b.sent + b.failed + b.suppressed }}/{{ b.total }}</span>
                        </div>
                        <div class="h-1.5 bg-zinc-800 rounded-full overflow-hidden">
                            <div class="h-full bg-indigo-400 transition-all" :style="{ width: progressPct(b) + '%' }"></div>
                        </div>
                    </div>
                </div>

                <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
                    <div class="lg:col-span-2 bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                        <p class="text-sm font-semibold text-white mb-3">Opens &amp; clicks over time</p>
                        <LineChart :labels="opensLine.labels" :series="opensLine.series" />
                    </div>
                    <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                        <p class="text-sm font-semibold text-white mb-3">Send status</p>
                        <DonutChart :labels="statusDonut.labels" :values="statusDonut.values" :colors="statusDonut.colors" />
                    </div>
                </div>

                <div v-if="overview.top_clicked_links.length" class="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                    <p class="text-sm font-semibold text-white mb-3">Top clicked links</p>
                    <BarChart :labels="overview.top_clicked_links.map(l => l.url)"
                        :values="overview.top_clicked_links.map(l => l.clicks)" horizontal />
                </div>
            </div>

            <!-- CAMPAIGNS -->
            <div v-else-if="tab === 'campaigns'">
                <div class="flex justify-end mb-3">
                    <button @click="exportCsv" class="px-3 py-1.5 text-sm bg-zinc-900 border border-zinc-800 rounded-lg text-zinc-300 hover:text-white transition">Export CSV</button>
                </div>
                <p v-if="!campaigns.length" class="text-sm text-zinc-500 py-8 text-center">No campaigns sent yet.</p>
                <div v-else class="overflow-x-auto">
                    <table class="w-full text-sm">
                        <thead class="text-zinc-500 text-xs uppercase tracking-wide border-b border-zinc-800">
                            <tr>
                                <th class="text-left px-3 py-2">Campaign</th>
                                <th class="text-right px-3 py-2">Recipients</th>
                                <th class="text-right px-3 py-2">Acceptance</th>
                                <th class="text-right px-3 py-2">Bounce</th>
                                <th class="text-right px-3 py-2">Open</th>
                                <th class="text-right px-3 py-2">Click</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="c in campaigns" :key="c.broadcast_id" @click="viewCampaign(c)"
                                class="border-b border-zinc-800/60 hover:bg-zinc-900 cursor-pointer">
                                <td class="px-3 py-2.5 text-white truncate max-w-xs">{{ c.subject || '(no subject)' }}</td>
                                <td class="px-3 py-2.5 text-right text-zinc-300">{{ c.total_recipients }}</td>
                                <td class="px-3 py-2.5 text-right text-zinc-300">{{ pct(c.acceptance_pct) }}</td>
                                <td class="px-3 py-2.5 text-right text-zinc-300">{{ pct(c.bounce_rate_pct) }}</td>
                                <td class="px-3 py-2.5 text-right text-zinc-300">{{ pct(c.open_rate_pct) }}</td>
                                <td class="px-3 py-2.5 text-right text-zinc-300">{{ pct(c.click_rate_pct) }}</td>
                            </tr>
                        </tbody>
                    </table>
                </div>

                <div v-if="openCampaign" class="mt-6 bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                    <p class="text-sm font-semibold text-white mb-3">{{ openCampaign.subject || '(no subject)' }}</p>
                    <div class="grid grid-cols-2 md:grid-cols-5 gap-3 mb-4">
                        <AppStatTile label="Acceptance" :value="pct(openCampaign.acceptance_pct)" />
                        <AppStatTile label="Bounce rate" :value="pct(openCampaign.bounce_rate_pct)" invert-good />
                        <AppStatTile label="Open rate" :value="pct(openCampaign.open_rate_pct)" />
                        <AppStatTile label="Click rate" :value="pct(openCampaign.click_rate_pct)" />
                        <AppStatTile label="Click-to-open" :value="pct(openCampaign.click_to_open_pct)" />
                    </div>
                    <div v-if="openCampaign.top_clicked_links.length">
                        <p class="text-xs font-semibold text-zinc-400 uppercase tracking-wide mb-2">Top links</p>
                        <BarChart :labels="openCampaign.top_clicked_links.map(l => l.url)"
                            :values="openCampaign.top_clicked_links.map(l => l.clicks)" horizontal />
                    </div>
                </div>
            </div>

            <!-- AUDIENCE -->
            <div v-else-if="tab === 'audience' && audience" class="space-y-6">
                <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
                    <AppStatTile label="Active" :value="audience.active_total" />
                    <AppStatTile label="Unsubscribed" :value="audience.unsubscribed_total" />
                    <AppStatTile label="New in range" :value="audience.added_in_range" />
                    <AppStatTile label="Unsubs in range" :value="audience.unsubscribed_in_range" invert-good />
                </div>
                <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                    <p class="text-sm font-semibold text-white mb-3">Net growth over time</p>
                    <LineChart :labels="audienceChart.labels"
                        :series="[{ label: 'Cumulative net', values: audienceChart.values, color: chartColors.emerald, fill: true }]" />
                </div>
            </div>

            <!-- ENGAGEMENT -->
            <div v-else-if="tab === 'engagement' && engagement" class="space-y-6">
                <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                    <p class="text-sm font-semibold text-white mb-3">Engagement funnel</p>
                    <div class="space-y-3">
                        <div v-for="row in funnelRows" :key="row.label">
                            <div class="flex justify-between text-xs mb-1">
                                <span class="text-zinc-300">{{ row.label }}</span>
                                <span class="text-zinc-400">{{ row.count }} · {{ row.pct.toFixed(1) }}%</span>
                            </div>
                            <div class="h-2 bg-zinc-800 rounded-full overflow-hidden">
                                <div class="h-full bg-indigo-400" :style="{ width: row.pct + '%' }"></div>
                            </div>
                        </div>
                    </div>
                </div>

                <div class="flex justify-end">
                    <div class="flex gap-1 bg-zinc-900 border border-zinc-800 rounded-lg p-1">
                        <button v-for="v in (['donut', 'bars'] as const)" :key="v" @click="engView = v"
                            :class="['px-2.5 py-1 text-xs rounded-md transition capitalize', engView === v ? 'bg-zinc-800 text-white' : 'text-zinc-400 hover:text-white']">
                            {{ v }}
                        </button>
                    </div>
                </div>

                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div v-for="group in [{ title: 'Devices', items: engagement.devices }, { title: 'Mail clients', items: engagement.clients }]" :key="group.title"
                        class="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                        <p class="text-sm font-semibold text-white mb-3">{{ group.title }}</p>
                        <template v-if="group.items.length">
                            <DonutChart v-if="engView === 'donut'" :labels="group.items.map(i => i.label)" :values="group.items.map(i => i.count)" />
                            <BarChart v-else :labels="group.items.map(i => i.label)" :values="group.items.map(i => i.count)" horizontal />
                            <table class="w-full text-sm mt-3">
                                <tbody>
                                    <tr v-for="i in breakdownPct(group.items)" :key="i.label" class="border-b border-zinc-800/60 last:border-0">
                                        <td class="py-1.5 text-zinc-300">{{ i.label }}</td>
                                        <td class="py-1.5 text-right text-zinc-400">{{ i.count }}</td>
                                        <td class="py-1.5 text-right text-zinc-500 w-16">{{ i.pct.toFixed(1) }}%</td>
                                    </tr>
                                </tbody>
                            </table>
                        </template>
                        <p v-else class="text-sm text-zinc-500 py-8 text-center">No click data yet.</p>
                    </div>
                </div>

                <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                    <p class="text-sm font-semibold text-white mb-3">Opens &amp; clicks over time</p>
                    <LineChart v-if="activityLine.labels.length" :labels="activityLine.labels" :series="activityLine.series" />
                    <p v-else class="text-sm text-zinc-500 py-8 text-center">No activity in this range.</p>
                </div>

                <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                    <p class="text-sm font-semibold text-white mb-1">Clicks by day &amp; hour</p>
                    <p class="text-xs text-zinc-500 mb-3">When your audience actually clicks (UTC) — based on clicks, since Apple MPP makes open times unreliable.</p>
                    <AppHeatmap v-if="engagement.heatmap.length" :cells="engagement.heatmap" />
                    <p v-else class="text-sm text-zinc-500 py-8 text-center">No clicks in this range.</p>
                </div>
            </div>

            <!-- DELIVERABILITY (Pro) -->
            <div v-else-if="tab === 'deliverability'" class="space-y-6">
                <AppProPaywall v-if="!licenseStore.allowsPro"
                    title="Deliverability is a Pro feature"
                    description="Check your domain authentication (SPF, DKIM, DMARC) and see which providers bounce your mail." />
                <template v-else>
                    <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                        <div class="flex items-center justify-between mb-3">
                            <p class="text-sm font-semibold text-white">Domain authentication</p>
                            <span v-if="domainHealth" class="text-xs text-zinc-500">{{ domainHealth.domain }}</span>
                        </div>
                        <div v-if="domainHealth" class="space-y-3">
                            <div v-for="c in domainHealth.checks" :key="c.name"
                                class="border-b border-zinc-800/60 last:border-0 pb-3 last:pb-0">
                                <div class="flex items-center gap-2">
                                    <span class="text-sm text-white font-medium w-16">{{ c.name }}</span>
                                    <AppStatusPill :status="c.status" />
                                </div>
                                <p class="text-xs text-zinc-400 mt-1">{{ c.detail }}</p>
                                <p v-if="c.fix" class="text-xs text-zinc-500 mt-0.5">Fix: {{ c.fix }}</p>
                            </div>
                        </div>
                    </div>

                    <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
                        <div class="lg:col-span-2 bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                            <p class="text-sm font-semibold text-white mb-3">Bounces by provider</p>
                            <p v-if="!bounces || !bounces.providers.length" class="text-sm text-zinc-500 py-8 text-center">No bounces in this range.</p>
                            <div v-else class="overflow-x-auto">
                                <table class="w-full text-sm">
                                    <thead class="text-zinc-500 text-xs uppercase tracking-wide border-b border-zinc-800">
                                        <tr>
                                            <th class="text-left px-3 py-2">Provider</th>
                                            <th class="text-right px-3 py-2">Total</th>
                                            <th class="text-right px-3 py-2">Hard</th>
                                            <th class="text-right px-3 py-2">Soft</th>
                                            <th class="text-right px-3 py-2">Unknown</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        <tr v-for="p in bounces.providers" :key="p.provider" class="border-b border-zinc-800/60">
                                            <td class="px-3 py-2.5 text-white">{{ p.provider }}</td>
                                            <td class="px-3 py-2.5 text-right text-zinc-300">{{ p.total }}</td>
                                            <td class="px-3 py-2.5 text-right text-red-400">{{ p.hard }}</td>
                                            <td class="px-3 py-2.5 text-right text-amber-400">{{ p.soft }}</td>
                                            <td class="px-3 py-2.5 text-right text-zinc-500">{{ p.unknown }}</td>
                                        </tr>
                                    </tbody>
                                </table>
                            </div>
                        </div>
                        <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                            <p class="text-sm font-semibold text-white mb-3">Bounce share</p>
                            <DonutChart v-if="bouncesDonut.values.length" :labels="bouncesDonut.labels" :values="bouncesDonut.values" />
                            <p v-else class="text-sm text-zinc-500 py-8 text-center">No bounces.</p>
                        </div>
                    </div>
                </template>
            </div>
    </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import type { Project } from '@/stores/projects'
import { useLicenseStore } from '@/stores/license'
import { useSegmentStore } from '@/stores/segments'
import {
    useReportsStore, type ReportSchema, type ReportConfig, type RunResult, type SavedReport,
} from '@/stores/reports'
import { chartColors } from '@/components/charts/chartSetup'
import LineChart from '@/components/charts/LineChart.vue'
import BarChart from '@/components/charts/BarChart.vue'
import DonutChart from '@/components/charts/DonutChart.vue'
import StackedBarChart from '@/components/charts/StackedBarChart.vue'
import AppProPaywall from '@/components/ui/AppProPaywall.vue'
import AppSelect from '@/components/ui/AppSelect.vue'
import AppLoader from '@/components/ui/AppLoader.vue'
import AppAlert from '@/components/ui/AppAlert.vue'
import AppModal from '@/components/ui/AppModal.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppConfirmModal from '@/components/ui/AppConfirmModal.vue'

const props = defineProps<{ project: Project }>()

const licenseStore = useLicenseStore()
const segmentStore = useSegmentStore()
const reportsStore = useReportsStore()

const schema = ref<ReportSchema | null>(null)
const loading = ref(false)
const error = ref('')

const dataset = ref('subscribers')
const measure = ref('count')
const dim1 = ref('status')
const dim2 = ref('')
const gran = ref('month')
const viz = ref('bar')
const segmentFilter = ref('')
const fromDate = ref('')
const toDate = ref('')

const segments = computed(() => segmentStore.segments(props.project.id))
const savedReports = computed(() => reportsStore.reports(props.project.id))

const currentDataset = computed(() => schema.value?.datasets.find(d => d.key === dataset.value))
const measures = computed(() => currentDataset.value?.measures ?? [])
const dims = computed(() => currentDataset.value?.dimensions ?? [])
const dim2Options = computed(() => dims.value.filter(d => d.key !== dim1.value))

function dimKind(key: string): string {
    return dims.value.find(d => d.key === key)?.kind ?? 'column'
}
const usesTime = computed(() => dimKind(dim1.value) === 'time' || (!!dim2.value && dimKind(dim2.value) === 'time'))
const isPivot = computed(() => !!dim2.value)
const vizOptions = computed(() =>
    isPivot.value ? ['pivot', 'bar', 'line', 'area'] : ['bar', 'donut', 'pie', 'line', 'area', 'table'],
)

const config = computed<ReportConfig>(() => {
    const dimensions = dim2.value ? [dim1.value, dim2.value] : [dim1.value]
    const seg = segments.value.find(s => s.id === segmentFilter.value)
    return {
        dataset: dataset.value,
        measure: measure.value,
        dimensions,
        viz: viz.value,
        window: { from: fromDate.value, to: toDate.value, granularity: gran.value },
        filters: seg ? seg.predicate : undefined,
    }
})

const result = ref<RunResult | null>(null)

watch(dataset, () => {
    measure.value = measures.value[0]?.key ?? 'count'
    dim1.value = dims.value[0]?.key ?? 'status'
    dim2.value = ''
    segmentFilter.value = ''
})
watch(isPivot, () => {
    if (!vizOptions.value.includes(viz.value)) viz.value = vizOptions.value[0] ?? 'table'
})

let debounce: ReturnType<typeof setTimeout> | null = null
watch(config, () => {
    if (!licenseStore.allowsPro) return
    if (debounce) clearTimeout(debounce)
    debounce = setTimeout(runReport, 400)
}, { deep: true })
onBeforeUnmount(() => { if (debounce) clearTimeout(debounce) })

async function runReport() {
    loading.value = true
    error.value = ''
    try {
        result.value = await reportsStore.run(props.project.id, config.value)
    } catch {
        error.value = 'Could not run this report. Adjust the selection and try again.'
    } finally {
        loading.value = false
    }
}

const measureLabel = computed(() => measures.value.find(m => m.key === measure.value)?.label ?? measure.value)

const oneDim = computed(() => result.value?.rows ?? [])
const pivot = computed(() => result.value?.pivot ?? null)

const singleSeries = computed(() => ({
    labels: oneDim.value.map(r => r.label),
    values: oneDim.value.map(r => r.value),
}))
const singleLine = computed(() => ({
    labels: oneDim.value.map(r => r.label),
    series: [{ label: measureLabel.value, values: oneDim.value.map(r => r.value), color: chartColors.indigo, fill: viz.value === 'area' }],
}))
const pivotSeries = computed(() => {
    const p = pivot.value
    if (!p) return { labels: [] as string[], series: [] as { label: string; values: number[]; color?: string; fill?: boolean }[] }
    return {
        labels: p.rows.map(r => r.label),
        series: p.columns.map(col => ({
            label: col,
            values: p.rows.map(r => r.cells[col] ?? 0),
            fill: viz.value === 'area',
        })),
    }
})

function exportCsv() {
    let csv = ''
    if (isPivot.value && pivot.value) {
        csv = [dim1.value, ...pivot.value.columns].join(',') + '\n'
        for (const row of pivot.value.rows) {
            csv += [csvCell(row.label), ...pivot.value.columns.map(c => row.cells[c] ?? 0)].join(',') + '\n'
        }
    } else {
        csv = `${dim1.value},${measure.value}\n`
        for (const row of oneDim.value) csv += `${csvCell(row.label)},${row.value}\n`
    }
    const blob = new Blob([csv], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'report.csv'
    a.click()
    URL.revokeObjectURL(url)
}
function csvCell(s: string): string {
    return /[",\n]/.test(s) ? `"${s.replace(/"/g, '""')}"` : s
}

const saveModalOpen = ref(false)
const saveName = ref('')
const editingId = ref('')
const deleteTarget = ref<SavedReport | null>(null)

function loadReport(rep: SavedReport) {
    const c = rep.config
    dataset.value = c.dataset
    measure.value = c.measure
    dim1.value = c.dimensions[0] ?? 'status'
    dim2.value = c.dimensions[1] ?? ''
    viz.value = c.viz
    gran.value = c.window?.granularity ?? 'month'
    fromDate.value = c.window?.from?.slice(0, 10) ?? ''
    toDate.value = c.window?.to?.slice(0, 10) ?? ''
    editingId.value = rep.id
    saveName.value = rep.name
}

function openSave() {
    if (!editingId.value) saveName.value = ''
    saveModalOpen.value = true
}
async function confirmSave() {
    if (!saveName.value.trim()) return
    if (editingId.value) {
        await reportsStore.updateReport(props.project.id, editingId.value, saveName.value.trim(), config.value)
    } else {
        const r = await reportsStore.createReport(props.project.id, saveName.value.trim(), config.value)
        editingId.value = r.id
    }
    saveModalOpen.value = false
}
async function confirmDelete() {
    if (!deleteTarget.value) return
    const id = deleteTarget.value.id
    await reportsStore.deleteReport(props.project.id, id)
    if (editingId.value === id) { editingId.value = ''; saveName.value = '' }
    deleteTarget.value = null
}
function newReport() {
    editingId.value = ''
    saveName.value = ''
}

onMounted(async () => {
    await licenseStore.fetch()
    if (!licenseStore.allowsPro) return
    const to = new Date()
    const from = new Date(to)
    from.setFullYear(to.getFullYear() - 1)
    fromDate.value = from.toISOString().slice(0, 10)
    toDate.value = to.toISOString().slice(0, 10)
    segmentStore.fetchSegments(props.project.id)
    reportsStore.fetchReports(props.project.id)
    try {
        schema.value = await reportsStore.schema(props.project.id)
    } catch {
        error.value = 'Could not load the report schema.'
    }
    runReport()
})
</script>

<template>
    <div>
        <div class="mb-6">
            <div class="flex items-center gap-2">
                <h2 class="text-xl font-semibold text-white">Reports</h2>
                <span class="text-[10px] font-semibold tracking-wider uppercase px-1.5 py-0.5 rounded bg-amber-500/15 text-amber-400 border border-amber-500/30">Pro</span>
            </div>
            <p class="text-sm text-zinc-400 mt-1">Build your own breakdowns — pick a dataset, a measure, how to group it, and how to see it.</p>
        </div>

        <AppProPaywall v-if="!licenseStore.allowsPro"
            title="Reports is a Pro feature"
            description="Compose custom reports over your subscribers and email events, with pivots, filters and saved views." />

        <template v-else-if="schema">
            <div class="flex flex-wrap items-center gap-2 mb-4">
                <AppSelect v-model="editingId" size="sm"
                    :options="[{ value: '', label: 'New report…' }, ...savedReports.map(r => ({ value: r.id, label: r.name }))]"
                    @change="() => { const r = savedReports.find(s => s.id === editingId); if (r) loadReport(r) }" />
                <AppButton size="sm" variant="secondary" @click="newReport">New</AppButton>
                <AppButton size="sm" @click="openSave">{{ editingId ? 'Save' : 'Save as…' }}</AppButton>
                <AppButton v-if="editingId" size="sm" variant="danger"
                    @click="deleteTarget = savedReports.find(s => s.id === editingId) ?? null">Delete</AppButton>
                <div class="flex-1"></div>
                <AppButton size="sm" variant="primary" @click="exportCsv">Export CSV</AppButton>
            </div>

            <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4 mb-6 grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
                <label class="block">
                    <span class="text-xs text-zinc-400 uppercase tracking-wide">Dataset</span>
                    <AppSelect v-model="dataset" size="sm" class="mt-1 w-full" :options="schema.datasets.map(d => ({ value: d.key, label: d.label }))" />
                </label>
                <label class="block">
                    <span class="text-xs text-zinc-400 uppercase tracking-wide">Measure</span>
                    <AppSelect v-model="measure" size="sm" class="mt-1 w-full" :options="measures.map(m => ({ value: m.key, label: m.label }))" />
                </label>
                <label class="block">
                    <span class="text-xs text-zinc-400 uppercase tracking-wide">Group by</span>
                    <AppSelect v-model="dim1" size="sm" class="mt-1 w-full" :options="dims.map(d => ({ value: d.key, label: d.label }))" />
                </label>
                <label class="block">
                    <span class="text-xs text-zinc-400 uppercase tracking-wide">Then by (pivot)</span>
                    <AppSelect v-model="dim2" size="sm" class="mt-1 w-full" :options="[{ value: '', label: '— none —' }, ...dim2Options.map(d => ({ value: d.key, label: d.label }))]" />
                </label>
                <label v-if="usesTime" class="block">
                    <span class="text-xs text-zinc-400 uppercase tracking-wide">Granularity</span>
                    <AppSelect v-model="gran" size="sm" class="mt-1 w-full" :options="schema.granularities" />
                </label>
                <label class="block">
                    <span class="text-xs text-zinc-400 uppercase tracking-wide">Filter (segment)</span>
                    <AppSelect v-model="segmentFilter" size="sm" class="mt-1 w-full" :options="[{ value: '', label: 'All' }, ...segments.map(sg => ({ value: sg.id, label: sg.name }))]" />
                </label>
                <label v-if="dataset === 'emails'" class="block col-span-2">
                    <span class="text-xs text-zinc-400 uppercase tracking-wide">Date range</span>
                    <div class="mt-1 flex items-center gap-1">
                        <input type="date" v-model="fromDate" class="flex-1 px-2 py-1.5 text-sm bg-zinc-900 border border-zinc-800 rounded-lg text-zinc-300 [color-scheme:dark] focus:outline-none" />
                        <span class="text-zinc-600 text-xs">→</span>
                        <input type="date" v-model="toDate" class="flex-1 px-2 py-1.5 text-sm bg-zinc-900 border border-zinc-800 rounded-lg text-zinc-300 [color-scheme:dark] focus:outline-none" />
                    </div>
                </label>
                <label class="block">
                    <span class="text-xs text-zinc-400 uppercase tracking-wide">Chart</span>
                    <AppSelect v-model="viz" size="sm" class="mt-1 w-full capitalize"
                        :options="vizOptions" />
                </label>
            </div>

            <AppAlert v-if="error" type="error" :message="error" class="mb-4" />
            <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4 min-h-[16rem]">
                <AppLoader v-if="loading" message="Running…" />

                <template v-else-if="isPivot && pivot">
                    <StackedBarChart v-if="viz === 'bar'" :labels="pivotSeries.labels" :series="pivotSeries.series" />
                    <LineChart v-else-if="viz === 'line' || viz === 'area'" :labels="pivotSeries.labels" :series="pivotSeries.series" />
                    <div v-else class="overflow-x-auto">
                        <table class="w-full text-sm">
                            <thead class="text-zinc-400 text-xs uppercase tracking-wide border-b border-zinc-800">
                                <tr>
                                    <th class="text-left px-3 py-2">{{ dim1 }}</th>
                                    <th v-for="c in pivot.columns" :key="c" class="text-right px-3 py-2">{{ c }}</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr v-for="row in pivot.rows" :key="row.label" class="border-b border-zinc-800/60">
                                    <td class="px-3 py-2 text-white">{{ row.label }}</td>
                                    <td v-for="c in pivot.columns" :key="c" class="px-3 py-2 text-right text-zinc-300">{{ row.cells[c] ?? 0 }}</td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </template>

                <template v-else-if="oneDim.length">
                    <DonutChart v-if="viz === 'donut' || viz === 'pie'" :labels="singleSeries.labels" :values="singleSeries.values" :cutout="viz === 'pie' ? '0%' : undefined" />
                    <BarChart v-else-if="viz === 'bar'" :labels="singleSeries.labels" :values="singleSeries.values" horizontal />
                    <LineChart v-else-if="viz === 'line' || viz === 'area'" :labels="singleLine.labels" :series="singleLine.series" />
                    <div v-else class="overflow-x-auto">
                        <table class="w-full text-sm">
                            <thead class="text-zinc-400 text-xs uppercase tracking-wide border-b border-zinc-800">
                                <tr>
                                    <th class="text-left px-3 py-2">{{ dim1 }}</th>
                                    <th class="text-right px-3 py-2">{{ measureLabel }}</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr v-for="row in oneDim" :key="row.label" class="border-b border-zinc-800/60">
                                    <td class="px-3 py-2 text-white">{{ row.label }}</td>
                                    <td class="px-3 py-2 text-right text-zinc-300">{{ row.value }}</td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </template>

                <p v-else class="text-sm text-zinc-400 py-12 text-center">No data for this selection.</p>
            </div>
        </template>

        <AppModal :show="saveModalOpen" title="Save report" @close="saveModalOpen = false">
            <label class="block">
                <span class="text-sm text-zinc-300">Name</span>
                <input v-model="saveName" type="text" placeholder="e.g. Subscribers by plan"
                    class="mt-1 w-full px-3 py-2 text-sm bg-zinc-900 border border-zinc-800 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent transition" @keyup.enter="confirmSave" />
            </label>
            <div class="flex gap-2 justify-end mt-4">
                <AppButton variant="secondary" @click="saveModalOpen = false">Cancel</AppButton>
                <AppButton @click="confirmSave">Save</AppButton>
            </div>
        </AppModal>

        <AppConfirmModal v-if="deleteTarget" :show="true" title="Delete report"
            :message="`Delete “${deleteTarget.name}”?`" confirm-label="Delete" danger
            @confirm="confirmDelete" @cancel="deleteTarget = null" />
    </div>
</template>

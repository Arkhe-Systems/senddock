<script setup lang="ts">
import { computed } from 'vue'
import { Bar } from 'vue-chartjs'
import { baseOptions, seriesPalette } from './chartSetup'

const props = defineProps<{
    labels: string[]
    series: { label: string; values: number[]; color?: string }[]
    horizontal?: boolean
}>()

const data = computed(() => ({
    labels: props.labels,
    datasets: props.series.map((s, i) => ({
        label: s.label,
        data: s.values,
        backgroundColor: s.color ?? seriesPalette[i % seriesPalette.length],
        borderRadius: 2,
        maxBarThickness: 40,
    })),
}))

const options = computed(() => {
    const o = baseOptions() as any
    if (props.horizontal) o.indexAxis = 'y'
    o.scales.x.stacked = true
    o.scales.y.stacked = true
    return o
})
</script>

<template>
    <div class="h-64">
        <Bar :data="data" :options="options" />
    </div>
</template>

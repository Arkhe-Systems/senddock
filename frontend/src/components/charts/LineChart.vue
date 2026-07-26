<script setup lang="ts">
import { computed } from 'vue'
import { Line } from 'vue-chartjs'
import { baseOptions, chartColors } from './chartSetup'

interface Series {
    label: string
    values: number[]
    color?: string
    fill?: boolean
}

const props = defineProps<{ labels: string[]; series: Series[] }>()

const data = computed(() => ({
    labels: props.labels,
    datasets: props.series.map((s) => {
        const color = s.color ?? chartColors.indigo
        return {
            label: s.label,
            data: s.values,
            borderColor: color,
            backgroundColor: s.fill ? chartColors.indigoFill : color,
            fill: s.fill ?? false,
            tension: 0.35,
            borderWidth: 2,
            pointRadius: 0,
            pointHoverRadius: 4,
            pointHoverBackgroundColor: color,
        }
    }),
}))

const options = computed(() => {
    const o = baseOptions() as any
    o.interaction = { mode: 'index', intersect: false }
    return o
})
</script>

<template>
    <div class="h-64">
        <Line :data="data" :options="options" />
    </div>
</template>

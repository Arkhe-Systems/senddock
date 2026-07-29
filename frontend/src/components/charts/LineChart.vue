<script setup lang="ts">
import { computed } from 'vue'
import { Line } from 'vue-chartjs'
import { baseOptions, chartColors, seriesPalette } from './chartSetup'

interface Series {
    label: string
    values: number[]
    color?: string
    fill?: boolean
}

const props = defineProps<{ labels: string[]; series: Series[] }>()

function translucent(color: string, alpha = 0.15): string {
    if (color.startsWith('#') && color.length === 7) {
        const r = parseInt(color.slice(1, 3), 16)
        const g = parseInt(color.slice(3, 5), 16)
        const b = parseInt(color.slice(5, 7), 16)
        return `rgba(${r}, ${g}, ${b}, ${alpha})`
    }
    return color
}

const data = computed(() => ({
    labels: props.labels,
    datasets: props.series.map((s, i) => {
        const color = s.color ?? (props.series.length > 1 ? seriesPalette[i % seriesPalette.length] ?? chartColors.indigo : chartColors.indigo)
        return {
            label: s.label,
            data: s.values,
            borderColor: color,
            backgroundColor: s.fill ? translucent(color) : color,
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

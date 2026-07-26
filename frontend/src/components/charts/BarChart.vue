<script setup lang="ts">
import { computed } from 'vue'
import { Bar } from 'vue-chartjs'
import { baseOptions, chartColors } from './chartSetup'

const props = defineProps<{
    labels: string[]
    values: number[]
    color?: string
    horizontal?: boolean
}>()

const data = computed(() => ({
    labels: props.labels,
    datasets: [{
        data: props.values,
        backgroundColor: props.color ?? chartColors.indigo,
        borderRadius: 4,
        maxBarThickness: 28,
    }],
}))

const options = computed(() => {
    const o = baseOptions() as any
    o.plugins.legend.display = false
    if (props.horizontal) o.indexAxis = 'y'
    return o
})
</script>

<template>
    <div class="h-64">
        <Bar :data="data" :options="options" />
    </div>
</template>

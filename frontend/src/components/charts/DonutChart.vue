<script setup lang="ts">
import { computed } from 'vue'
import { Doughnut } from 'vue-chartjs'
import { baseOptions, seriesPalette } from './chartSetup'

const props = defineProps<{
    labels: string[]
    values: number[]
    colors?: string[]
    cutout?: string
}>()

const data = computed(() => ({
    labels: props.labels,
    datasets: [{
        data: props.values,
        backgroundColor: props.colors ?? seriesPalette,
        borderColor: '#18181b',
        borderWidth: 2,
    }],
}))

const options = computed(() => {
    const o = baseOptions() as any
    delete o.scales
    o.cutout = props.cutout ?? '62%'
    o.plugins.legend.position = 'right'
    return o
})
</script>

<template>
    <div class="h-64">
        <Doughnut :data="data" :options="options" />
    </div>
</template>

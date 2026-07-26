<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
    label: string
    value: string | number
    // Percent change vs the previous period. Positive is not always "good"
    // (e.g. failures), so invertGood flips the colour without flipping the sign.
    trend?: number | null
    invertGood?: boolean
    hint?: string
}>()

const hasTrend = computed(() => props.trend !== null && props.trend !== undefined && isFinite(props.trend as number))

const trendGood = computed(() => {
    const t = props.trend as number
    if (t === 0) return null
    const up = t > 0
    return props.invertGood ? !up : up
})

const trendClass = computed(() => {
    if (trendGood.value === null) return 'text-zinc-500'
    return trendGood.value ? 'text-emerald-400' : 'text-red-400'
})

const trendLabel = computed(() => {
    const t = props.trend as number
    const arrow = t > 0 ? '↑' : t < 0 ? '↓' : '·'
    return `${arrow} ${Math.abs(t).toFixed(1)}%`
})
</script>

<template>
    <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
        <p class="text-xs text-zinc-500 uppercase tracking-wide">{{ label }}</p>
        <p class="text-2xl font-bold text-white mt-1">{{ value }}</p>
        <div class="mt-1 flex items-center gap-2">
            <span v-if="hasTrend" :class="['text-xs font-medium', trendClass]">{{ trendLabel }}</span>
            <span v-if="hint" class="text-xs text-zinc-600">{{ hint }}</span>
        </div>
    </div>
</template>

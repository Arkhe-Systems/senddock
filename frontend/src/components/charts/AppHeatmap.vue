<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
    cells: { weekday: number; hour: number; count: number }[]
}>()

const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
const hours = Array.from({ length: 24 }, (_, i) => i)

const grid = computed(() => {
    const m: number[][] = Array.from({ length: 7 }, () => Array(24).fill(0) as number[])
    let max = 0
    for (const c of props.cells) {
        const row = m[c.weekday]
        if (row && c.hour >= 0 && c.hour < 24) {
            row[c.hour] = c.count
            if (c.count > max) max = c.count
        }
    }
    return { m, max }
})

function cellStyle(count: number): Record<string, string> {
    if (!count || grid.value.max <= 0) return { backgroundColor: 'rgb(39 39 42)' }
    const alpha = 0.15 + 0.85 * (count / grid.value.max)
    return { backgroundColor: `rgba(52, 211, 153, ${alpha})` }
}
</script>

<template>
    <div class="overflow-x-auto">
        <div class="inline-block">
            <div v-for="(row, d) in grid.m" :key="d" class="flex items-center gap-1 mb-1">
                <span class="w-8 text-[10px] text-zinc-500 shrink-0">{{ days[d] }}</span>
                <div v-for="h in hours" :key="h"
                    class="w-3.5 h-3.5 rounded-sm shrink-0"
                    :style="cellStyle(row[h] ?? 0)"
                    :title="`${days[d]} ${h}:00 — ${row[h] ?? 0} clicks`"></div>
            </div>
            <div class="flex items-center gap-1">
                <span class="w-8 shrink-0"></span>
                <span v-for="h in hours" :key="h" class="w-3.5 text-[8px] text-zinc-600 text-center shrink-0">{{ h % 6 === 0 ? h : '' }}</span>
            </div>
        </div>
    </div>
</template>

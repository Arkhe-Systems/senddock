<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
    page: number
    limit: number
    total: number
    pageSizes?: number[]
}>(), {
    pageSizes: () => [10, 25, 50, 100, 200],
})

const emit = defineEmits<{
    'update:page': [value: number]
    'update:limit': [value: number]
    change: []
}>()

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.limit)))
const rangeStart = computed(() => props.total === 0 ? 0 : props.page * props.limit + 1)
const rangeEnd = computed(() => Math.min((props.page + 1) * props.limit, props.total))
const isFirst = computed(() => props.page === 0)
const isLast = computed(() => props.page + 1 >= totalPages.value)

function goPrev() {
    if (isFirst.value) return
    emit('update:page', props.page - 1)
    emit('change')
}

function goNext() {
    if (isLast.value) return
    emit('update:page', props.page + 1)
    emit('change')
}

function onLimitChange(e: Event) {
    const next = Number((e.target as HTMLSelectElement).value)
    if (next === props.limit) return
    // Always reset to page 0 when page size changes — avoids "page 7 of 5" errors
    emit('update:limit', next)
    emit('update:page', 0)
    emit('change')
}

function onPageJump(e: Event) {
    const target = e.target as HTMLInputElement
    const raw = parseInt(target.value, 10)
    const clamped = Math.max(1, Math.min(totalPages.value, isNaN(raw) ? 1 : raw))
    target.value = String(clamped)
    if (clamped - 1 === props.page) return
    emit('update:page', clamped - 1)
    emit('change')
}
</script>

<template>
    <div v-if="total > 0" class="flex flex-wrap items-center justify-between gap-3 mt-4 text-sm">
        <div class="flex items-center gap-2 text-zinc-400">
            <label class="whitespace-nowrap">Per page</label>
            <select
                :value="limit"
                @change="onLimitChange"
                class="px-2 py-1 bg-zinc-900 border border-zinc-800 rounded-md text-zinc-300 hover:text-white focus:outline-none focus:border-emerald-500 cursor-pointer">
                <option v-for="size in pageSizes" :key="size" :value="size">{{ size }}</option>
            </select>
        </div>

        <div class="text-zinc-400 hidden sm:block">
            Showing {{ rangeStart }}–{{ rangeEnd }} of {{ total }}
        </div>

        <div class="flex items-center gap-3">
            <button
                @click="goPrev"
                :disabled="isFirst"
                class="text-zinc-300 hover:text-white disabled:opacity-40 disabled:cursor-not-allowed cursor-pointer transition">
                ‹ Previous
            </button>

            <div class="flex items-center gap-1.5 text-zinc-400">
                <span>Page</span>
                <input
                    type="number"
                    :value="page + 1"
                    :min="1"
                    :max="totalPages"
                    @change="onPageJump"
                    @keydown.enter.prevent="onPageJump($event)"
                    class="w-12 px-1.5 py-0.5 text-center bg-zinc-900 border border-zinc-800 rounded-md text-zinc-300 focus:outline-none focus:border-emerald-500 [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none" />
                <span>of {{ totalPages }}</span>
            </div>

            <button
                @click="goNext"
                :disabled="isLast"
                class="text-zinc-300 hover:text-white disabled:opacity-40 disabled:cursor-not-allowed cursor-pointer transition">
                Next ›
            </button>
        </div>
    </div>
</template>

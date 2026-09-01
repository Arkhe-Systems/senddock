<script setup lang="ts">
// Shared toggleable chip used for every filter row and segmented mode picker
// (status filters, reason filters, type filters, date presets, send-mode
// toggles…). One component keeps the active/selected state identical across
// the whole dashboard instead of each screen styling its own chip.
defineProps<{
    active: boolean
    size?: 'xs' | 'sm'
}>()
</script>

<template>
    <button type="button"
        :class="[
            'inline-flex items-center justify-center rounded-md border transition cursor-pointer whitespace-nowrap',
            size === 'sm' ? 'px-3 py-1.5 text-sm' : 'px-3 py-1.5 text-xs',
            active
                ? 'bg-emerald-500/12 border-emerald-500/45 text-emerald-300'
                : 'border-zinc-800 text-zinc-300 hover:text-white hover:border-zinc-700',
        ]">
        <slot />
    </button>
</template>

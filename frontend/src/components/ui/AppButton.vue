<script setup lang="ts">
defineProps<{
    variant?: 'primary' | 'secondary' | 'danger' | 'ghost'
    size?: 'sm' | 'md' | 'lg'
    disabled?: boolean
    loading?: boolean
}>()
</script>

<template>
    <button
        :disabled="disabled || loading"
        :class="[
            'font-medium transition cursor-pointer border border-transparent',
            'disabled:opacity-50 disabled:cursor-not-allowed',
            size === 'sm' && 'px-3 py-1.5 text-sm rounded-md',
            size === 'md' && 'px-4 py-2 text-sm rounded-lg',
            size === 'lg' && 'px-6 py-2.5 text-sm rounded-lg',
            !size && 'w-full py-2 rounded-lg',
            variant === 'danger' && 'bg-red-600 hover:bg-red-500 text-white',
            variant === 'secondary' && 'bg-zinc-850 hover:bg-zinc-700 text-zinc-300',
            variant === 'ghost' && 'text-zinc-300 hover:bg-zinc-850 hover:text-white',
            (!variant || variant === 'primary') && 'bg-emerald-500/12 border-emerald-500/45 text-emerald-300 hover:bg-emerald-500/20 hover:border-emerald-500/70 hover:text-emerald-200',
        ]">
        <slot />
    </button>
</template>

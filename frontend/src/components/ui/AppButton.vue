<script setup lang="ts">
defineProps<{
    variant?: 'primary' | 'secondary' | 'danger' | 'ghost'
    size?: 'sm' | 'lg'
    disabled?: boolean
    loading?: boolean
}>()
</script>

<template>
    <button
        :disabled="disabled || loading"
        :class="[
            'font-medium transition cursor-pointer',
            'disabled:opacity-50 disabled:cursor-not-allowed',
            size === 'sm'
                ? 'px-3 py-1.5 text-sm rounded-md'
                : 'w-full py-2 rounded-lg',
            variant === 'danger' && 'bg-red-600 hover:bg-red-500 text-white',
            variant === 'secondary' && 'bg-zinc-800 hover:bg-zinc-700 text-zinc-300',
            variant === 'ghost' && 'text-zinc-300 hover:bg-zinc-800 hover:text-white',
            (!variant || variant === 'primary') && 'bg-white hover:bg-zinc-200 text-zinc-950',
        ]">
        <slot />
    </button>
</template>

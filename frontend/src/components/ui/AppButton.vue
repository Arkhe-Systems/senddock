<script setup lang="ts">
defineProps<{
    variant?: 'primary' | 'secondary' | 'danger'
    disabled?: boolean
    loading?:boolean
}>()
</script>

<template>
    <button
    :disabled="disabled || loading"
    :class="['w-full py-2 font-medium rounded-lg transition',
        'disabled:opacity-50 disabled:cursor-not-allowed',
        variant === 'danger' && 'bg-red-600 hover:bg-red-500 text-white',
        variant === 'secondary' && 'bg-zinc-800 hover:bg-zinc-700 text-zinc-300',
        (!variant || variant === 'primary') && 'bg-indigo-600 hover:bg-indigo-500 text-white',

    ]">
    <slot />
    </button>
</template>
<script setup lang="ts">
defineProps<{
    message?: string
    fullscreen?: boolean
}>()
</script>

<template>
    <div :class="[
        'flex flex-col items-center justify-center gap-3',
        fullscreen ? 'min-h-screen' : 'py-16'
    ]">
        <svg class="animate-spin h-6 w-6 text-zinc-400" viewBox="0 0 24 24" fill="none">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="3" />
            <path class="opacity-75" fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
        </svg>
        <p v-if="message" class="text-sm text-zinc-500">{{ message }}</p>
    </div>
</template>

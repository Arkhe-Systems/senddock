<script setup lang="ts">
defineProps<{
    type?: 'error' | 'success' | 'info'
    message: string
}>()
</script>

<template>
    <div v-if="message"
    :class="[
        'rounded-lg px-4 py-3 text-sm border',
        type === 'success' && 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400',
        type === 'info' && 'bg-amber-500/10 border-amber-500/20 text-amber-400',
        (!type || type === 'error') && 'bg-red-500/10 border-red-500/20 text-red-400'
    ]">
        {{ message }}
    </div>

</template>
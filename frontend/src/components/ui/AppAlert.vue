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
        type === 'success' && 'bg-green-500/10 border-green-500/20 text-green-400',
        type === 'info' && 'bg-yellow-500/10 border-yellow-500/20 text-yellow-400',
        (!type || type === 'error') && 'bg-red-500/10 border-red-500/20 text-red-400'
    ]">
        {{ message }}
    </div>

</template>
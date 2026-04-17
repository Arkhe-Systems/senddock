<script setup lang="ts">

defineProps<{
    title?: string,
    show?: boolean
}>()

const emit = defineEmits<{
    close: []
}>()
</script>
<template>
    <div v-if="show" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click="emit('close')">
        <div class="bg-zinc-900 rounded-lg p-6 w-full max-w-md" @click.stop>
            <div class="flex items-center justify-between mb-4">
                <h2 class="text-lg font-semibold text-white">{{ title }}</h2>
                <button @click="emit('close')" class="text-zinc-400 hover:text-white">&times;</button>
            </div>
            <slot />
        </div>
    </div>
</template>
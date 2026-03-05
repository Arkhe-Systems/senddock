<script setup lang="ts">
defineProps<{
    title?:string
    subtitle?:string
}>()
</script>

<template>
    <div class="w-full max-w-sm">
        <div v-if="title || subtitle" class="text-center mb-8">
            <h1 v-if="title" class="text-2xl font-bold text-white">{{  title }}</h1>
            <p v-if="subtitle" class="text-zinc-400 mt-2">{{ subtitle }}</p>
        </div>
        <slot/>
    </div>
</template>
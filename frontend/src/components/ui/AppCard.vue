<script setup lang="ts">
defineProps<{
    title?:string
    subtitle?:string
    variant?: 'default' | 'card'
}>()
</script>

<template>
    <div :class="[ 'w-full', variant !== 'card' && 'max-w-sm', variant === 'card' && 'bg-zinc-900 border border-zinc-800 rounded-lg p-4 hover:border-zinc-700 transition cursor-pointer']">
        <div v-if="title || subtitle" :class="variant === 'card' ? 'mb-2' : 'text-center mb-8'">
            <h1 v-if="title" :class="variant === 'card' ? 'text-base font-semibold text-white' : 'text-2xl font-bold text-white'">{{  title }}</h1>
            <p v-if="subtitle" class="text-zinc-400 mt-2">{{ subtitle }}</p>
        </div>
        <slot/>
    </div>
</template>
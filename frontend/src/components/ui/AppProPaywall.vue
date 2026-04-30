<script setup lang="ts">
defineProps<{
    title: string
    description: string
    cta?: string
}>()

const PRICING_URL = 'https://senddock.dev/pricing'
</script>

<template>
    <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-6 sm:p-8 text-center">
        <div class="inline-flex items-center gap-2 px-2.5 py-0.5 rounded-full bg-amber-500/15 text-amber-400 border border-amber-500/30 text-[11px] font-semibold tracking-wider uppercase mb-4">
            Pro
        </div>
        <h2 class="text-base sm:text-lg font-semibold text-white mb-2">{{ title }}</h2>
        <p class="text-sm text-zinc-400 mb-6 max-w-md mx-auto">
            {{ description }}
        </p>
        <a :href="PRICING_URL" target="_blank" rel="noopener"
            class="inline-block px-4 py-2 text-sm font-medium bg-white text-zinc-950 rounded-lg hover:bg-zinc-200 transition">
            {{ cta || 'See Pro plans' }}
        </a>
    </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useUsageStore } from '@/stores/usage'

const usageStore = useUsageStore()

onMounted(() => usageStore.fetch())

const alert = computed(() => {
    const over = usageStore.alerts.find(a => a.level === 'over')
    return over || usageStore.alerts[0] || null
})
</script>

<template>
    <div v-if="alert"
        :class="['rounded-lg border px-4 py-3 text-sm flex items-center justify-between gap-4',
            alert.level === 'over'
                ? 'border-red-500/30 bg-red-500/10 text-red-200'
                : 'border-amber-500/30 bg-amber-500/10 text-amber-200']">
        <span>
            <template v-if="alert.level === 'over'">
                You've reached your {{ alert.label }} limit ({{ alert.used }}/{{ alert.limit }}). Upgrade your plan to add more.
            </template>
            <template v-else>
                You're using {{ alert.used }} of {{ alert.limit }} {{ alert.label }} — close to your plan limit.
            </template>
        </span>
        <RouterLink to="/billing" class="shrink-0 underline text-white hover:text-zinc-200">Upgrade</RouterLink>
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

// Shared sidebar navigation link. Every nav menu (project sidebar, profile
// panel, …) renders through this component so the active state stays visually
// identical everywhere — one source of truth instead of per-menu classes.
const props = defineProps<{
    to: string | Record<string, unknown>
}>()

const route = useRoute()
const router = useRouter()

const active = computed(() => {
    try {
        return route.name === router.resolve(props.to).name
    } catch {
        return false
    }
})
</script>

<template>
    <RouterLink :to="to"
        :class="[
            'flex items-center gap-2.5 px-3 py-2 text-sm rounded-lg transition',
            active
                ? 'bg-emerald-500/10 text-emerald-400'
                : 'text-zinc-300 hover:text-white hover:bg-zinc-850'
        ]">
        <slot />
    </RouterLink>
</template>

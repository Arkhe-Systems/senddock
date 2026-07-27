<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
    status: 'pass' | 'warn' | 'fail'
    label?: string
}>()

const text = computed(() => props.label ?? { pass: 'Pass', warn: 'Warn', fail: 'Fail' }[props.status])

const cls = computed(() => ({
    pass: 'bg-green-500/10 text-green-400',
    warn: 'bg-amber-500/10 text-amber-400',
    fail: 'bg-red-500/10 text-red-400',
}[props.status]))
</script>

<template>
    <span :class="['text-xs font-medium px-2 py-1 rounded-full', cls]">{{ text }}</span>
</template>

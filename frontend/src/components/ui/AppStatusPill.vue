<script setup lang="ts">
import { computed } from 'vue'

export type PillTone = 'emerald' | 'red' | 'amber' | 'blue' | 'orange' | 'zinc'

const props = withDefaults(defineProps<{
    status?: 'pass' | 'warn' | 'fail'
    tone?: PillTone
    label?: string
    bordered?: boolean
}>(), {})

const resolvedTone = computed<PillTone>(() => {
    if (props.tone) return props.tone
    return { pass: 'emerald', warn: 'amber', fail: 'red' }[props.status ?? 'pass'] as PillTone
})

const text = computed(() => props.label ?? { pass: 'Pass', warn: 'Warn', fail: 'Fail' }[props.status ?? 'pass'])

const toneCls: Record<PillTone, string> = {
    emerald: 'bg-emerald-500/10 text-emerald-400',
    amber: 'bg-amber-500/10 text-amber-400',
    red: 'bg-red-500/10 text-red-400',
    blue: 'bg-blue-500/10 text-blue-400',
    orange: 'bg-orange-500/10 text-orange-400',
    zinc: 'bg-zinc-500/10 text-zinc-300',
}

const borderCls: Record<PillTone, string> = {
    emerald: 'border-emerald-500/30',
    amber: 'border-amber-500/30',
    red: 'border-red-500/30',
    blue: 'border-blue-500/30',
    orange: 'border-orange-500/30',
    zinc: 'border-zinc-600',
}
</script>

<template>
    <span :class="['text-xs font-medium px-2 py-1 rounded-full whitespace-nowrap', toneCls[resolvedTone], bordered && ['border', borderCls[resolvedTone]]]">
        {{ text }}
    </span>
</template>

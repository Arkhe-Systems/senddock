<script setup lang="ts">
const props = defineProps<{
    modelValue?: boolean
    disabled?: boolean
}>()

const emit = defineEmits<{
    'update:modelValue': [value: boolean]
}>()

function onChange(e: Event) {
    emit('update:modelValue', (e.target as HTMLInputElement).checked)
}
</script>

<template>
    <span class="relative inline-flex items-center justify-center flex-shrink-0 w-[18px] h-[18px]">
        <input type="checkbox" :checked="!!modelValue" :disabled="disabled" @change="onChange" class="peer sr-only" />
        <span :class="[
            'absolute inset-0 rounded border-2 border-zinc-600 bg-transparent transition-colors',
            'peer-checked:border-white peer-checked:bg-white',
            'peer-disabled:opacity-50 peer-disabled:cursor-not-allowed',
            !disabled && 'cursor-pointer hover:border-zinc-400',
        ]"></span>
        <svg v-if="modelValue" class="relative w-3 h-3 text-zinc-950 pointer-events-none"
            viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="3 8 7 12 13 4" />
        </svg>
    </span>
</template>

<script setup lang="ts">
defineProps<{
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
        <input type="checkbox" :checked="!!modelValue" :disabled="disabled" @change="onChange"
            class="peer absolute inset-0 m-0 appearance-none opacity-0 cursor-pointer disabled:cursor-not-allowed z-10" />
        <span :class="[
            'pointer-events-none absolute inset-0 rounded border-2 border-zinc-600 bg-transparent transition-colors',
            'peer-checked:border-emerald-500 peer-checked:bg-emerald-500',
            'peer-disabled:opacity-50',
            !disabled && 'peer-hover:border-zinc-400',
        ]"></span>
        <svg v-if="modelValue" class="pointer-events-none relative w-3 h-3 text-white"
            viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="3 8 7 12 13 4" />
        </svg>
    </span>
</template>

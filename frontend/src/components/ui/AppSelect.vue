<script setup lang="ts">
export type SelectOption = string | { value: string; label: string }

const model = defineModel<string>()

withDefaults(defineProps<{
    options?: SelectOption[]
    size?: 'xs' | 'sm' | 'md'
    disabled?: boolean
}>(), { size: 'md' })
</script>

<template>
    <select
        v-model="model"
        :disabled="disabled"
        :class="[
            'bg-zinc-900 border border-zinc-800 text-white focus:outline-none focus:ring-emerald-500 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed',
            size === 'xs' && 'px-2 py-1 text-xs rounded-md focus:ring-1',
            size === 'sm' && 'px-3 py-1.5 text-sm rounded-lg focus:ring-1',
            size === 'md' && 'px-3 py-2 rounded-lg focus:ring-2 focus:border-transparent transition',
        ]">
        <option v-for="opt in options" :key="typeof opt === 'string' ? opt : opt.value"
            :value="typeof opt === 'string' ? opt : opt.value">
            {{ typeof opt === 'string' ? opt : opt.label }}
        </option>
    </select>
</template>

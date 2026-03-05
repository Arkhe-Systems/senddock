<script setup lang="ts">
const model = defineModel<string>()

defineProps<{
    label?: string,
    type?: string,
    placeholder?: string,
    required?: boolean
    id?: string
}>()
</script>

<template>
    <div>
        <label v-if="label" :for="id" class="block text-sm font-medium text-zinc-300 mb-1">
            {{ label }}
        </label>

        <input
            :id="id"
            v-model="model"
            :type="type ?? 'text'"
            :placeholder="placeholder"
            :required="required"
            class="w-full px-3 py-2 bg-zinc-900 border border-zinc-800 rounded-lg text-white placeholder-zinc-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition"
        />
    </div>
</template>
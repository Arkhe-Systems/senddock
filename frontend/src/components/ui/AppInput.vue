<script setup lang="ts">
const model = defineModel<string>()

const props = defineProps<{
    label?: string,
    type?: string,
    placeholder?: string,
    required?: boolean,
    id?: string,
    large?: boolean,
    autocomplete?: string,
    error?: string,
}>()

const ignoreAttrs = {
    'data-bwignore': 'true',
    'data-1p-ignore': 'true',
    'data-lpignore': 'true',
    'data-form-type': 'other',
}

const passwordManagerAttrs = props.autocomplete ? {} : ignoreAttrs
</script>

<template>
    <div>
        <label v-if="label" :for="id" class="block text-sm font-medium text-zinc-300 mb-1">
            {{ label }}
        </label>

        <textarea
            v-if="large"
            :id="id"
            v-model="model"
            :placeholder="placeholder"
            :required="required"
            autocomplete="off"
            v-bind="ignoreAttrs"
            rows="3"
            class="w-full px-3 py-2 bg-zinc-900 border border-zinc-800 rounded-lg text-white placeholder-zinc-500 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent transition resize-none"/>

        <input
            v-else
            :id="id"
            v-model="model"
            :type="type ?? 'text'"
            :placeholder="placeholder"
            :required="required"
            :autocomplete="autocomplete ?? 'off'"
            v-bind="passwordManagerAttrs"
            :class="[
                'w-full px-3 py-2 bg-zinc-900 border rounded-lg text-white placeholder-zinc-500 focus:outline-none focus:ring-2 focus:border-transparent transition',
                error ? 'border-red-500/50 focus:ring-red-500/50' : 'border-zinc-800 focus:ring-emerald-500',
            ]"
        />

        <p v-if="error" class="mt-1 text-xs text-red-400">{{ error }}</p>
    </div>
</template>

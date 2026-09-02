<script setup lang="ts">
import { ref } from 'vue'

const model = defineModel<string[]>({ required: true })

defineProps<{
    label?: string
    placeholder?: string
    suggestions?: string[]
}>()

const draft = ref('')

function addTag(raw: string) {
    const tag = raw.trim()
    if (!tag) return
    if (!model.value.includes(tag)) {
        model.value = [...model.value, tag]
    }
    draft.value = ''
}

function onKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' || e.key === ',') {
        e.preventDefault()
        addTag(draft.value)
    } else if (e.key === 'Backspace' && draft.value === '' && model.value.length > 0) {
        model.value = model.value.slice(0, -1)
    }
}

function removeTag(tag: string) {
    model.value = model.value.filter(t => t !== tag)
}
</script>

<template>
    <div>
        <label v-if="label" class="block text-sm font-medium text-zinc-300 mb-1">{{ label }}</label>
        <div class="flex flex-wrap gap-1.5 px-2 py-2 bg-zinc-900 border border-zinc-800 rounded-lg focus-within:ring-2 focus-within:ring-zinc-500 transition">
            <span v-for="tag in model" :key="tag"
                class="inline-flex items-center gap-1 text-xs bg-zinc-850 text-zinc-200 px-2 py-1 rounded border border-zinc-700">
                {{ tag }}
                <button type="button" @click="removeTag(tag)" class="text-zinc-400 hover:text-red-400 transition cursor-pointer">&times;</button>
            </span>
            <input
                v-model="draft"
                :placeholder="model.length === 0 ? (placeholder || 'Add a tag and press Enter') : ''"
                list="tag-suggestions"
                autocomplete="off"
                class="flex-1 min-w-[8rem] bg-transparent text-sm text-white placeholder-zinc-500 focus:outline-none"
                @keydown="onKeydown"
                @blur="addTag(draft)" />
            <datalist v-if="suggestions && suggestions.length" id="tag-suggestions">
                <option v-for="s in suggestions" :key="s" :value="s" />
            </datalist>
        </div>
    </div>
</template>

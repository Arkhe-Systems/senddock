<script setup lang="ts">
import AppCheckbox from '@/components/ui/AppCheckbox.vue'
import type { FieldDefinition } from '@/stores/fields'

const props = defineProps<{ definitions: FieldDefinition[]; errors?: Record<string, string> }>()
const model = defineModel<Record<string, any>>({ required: true })

const baseInput =
    'w-full px-3 py-2 bg-zinc-900 border rounded-lg text-white placeholder-zinc-500 focus:outline-none focus:ring-2 focus:border-transparent transition'

function inputClass(key: string) {
    return [baseInput, props.errors?.[key] ? 'border-red-500/50 focus:ring-red-500/50' : 'border-zinc-800 focus:ring-emerald-500']
}

function setValue(key: string, value: any) {
    model.value = { ...model.value, [key]: value }
}

function onNumber(key: string, raw: string) {
    if (raw === '') {
        const next = { ...model.value }
        delete next[key]
        model.value = next
        return
    }
    const parsed = Number(raw)
    setValue(key, Number.isNaN(parsed) ? raw : parsed)
}
</script>

<template>
    <div v-if="props.definitions.length > 0" class="space-y-4">
        <div v-for="def in props.definitions" :key="def.id">
            <label class="block text-sm font-medium text-zinc-300 mb-1">
                {{ def.label }}
                <span v-if="def.required" class="text-red-400">*</span>
            </label>

            <input
                v-if="def.field_type === 'string'"
                :value="model[def.key] ?? ''"
                type="text"
                :required="def.required"
                :class="inputClass(def.key)"
                @input="setValue(def.key, ($event.target as HTMLInputElement).value)" />

            <input
                v-else-if="def.field_type === 'number'"
                :value="model[def.key] ?? ''"
                type="number"
                :required="def.required"
                :class="inputClass(def.key)"
                @input="onNumber(def.key, ($event.target as HTMLInputElement).value)" />

            <input
                v-else-if="def.field_type === 'date'"
                :value="model[def.key] ?? ''"
                type="date"
                :required="def.required"
                :class="inputClass(def.key)"
                @input="setValue(def.key, ($event.target as HTMLInputElement).value)" />

            <select
                v-else-if="def.field_type === 'enum'"
                :value="model[def.key] ?? ''"
                :required="def.required"
                :class="inputClass(def.key)"
                @change="setValue(def.key, ($event.target as HTMLSelectElement).value)">
                <option value="" :disabled="def.required">—</option>
                <option v-for="opt in def.options" :key="opt" :value="opt">{{ opt }}</option>
            </select>

            <label v-else-if="def.field_type === 'boolean'" class="flex items-center gap-2 mt-1">
                <AppCheckbox
                    :modelValue="!!model[def.key]"
                    @update:modelValue="(v: boolean) => setValue(def.key, v)" />
                <span class="text-sm text-zinc-300">{{ def.label }}</span>
            </label>

            <p v-if="props.errors?.[def.key]" class="mt-1 text-xs text-red-400">{{ props.errors[def.key] }}</p>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useToastStore } from '@/stores/toast'
import { useFieldStore, type FieldDefinition, type FieldType } from '@/stores/fields'
import type { Project } from '@/stores/projects'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import AppModal from '@/components/ui/AppModal.vue'
import AppCheckbox from '@/components/ui/AppCheckbox.vue'
import AppConfirmModal from '@/components/ui/AppConfirmModal.vue'
import AppSelect from '@/components/ui/AppSelect.vue'

const props = defineProps<{ project: Project }>()
const toast = useToastStore()
const fieldStore = useFieldStore()

const definitions = computed(() => fieldStore.fields(props.project.id))

const customSyntaxHint = '{{custom.key}}'

const typeOptions: { value: FieldType; label: string }[] = [
    { value: 'string', label: 'Text' },
    { value: 'number', label: 'Number' },
    { value: 'date', label: 'Date' },
    { value: 'boolean', label: 'Boolean' },
    { value: 'enum', label: 'Dropdown' },
]

const showModal = ref(false)
const editing = ref<FieldDefinition | null>(null)
const saving = ref(false)
const keyError = ref('')
const optionsError = ref('')

const KEY_RE = /^[a-zA-Z][a-zA-Z0-9_]*$/

const form = ref({
    key: '',
    label: '',
    field_type: 'string' as FieldType,
    optionsText: '',
    required: false,
})

function clearErrors() {
    keyError.value = ''
    optionsError.value = ''
}

function openCreate() {
    editing.value = null
    form.value = { key: '', label: '', field_type: 'string', optionsText: '', required: false }
    clearErrors()
    showModal.value = true
}

function openEdit(def: FieldDefinition) {
    editing.value = def
    form.value = {
        key: def.key,
        label: def.label,
        field_type: def.field_type,
        optionsText: (def.options || []).join(', '),
        required: def.required,
    }
    clearErrors()
    showModal.value = true
}

function parseOptions(): string[] {
    return form.value.optionsText
        .split(',')
        .map(o => o.trim())
        .filter(Boolean)
}

async function save() {
    clearErrors()
    if (!editing.value) {
        const key = form.value.key.trim()
        if (!key) {
            keyError.value = 'Key is required'
        } else if (!KEY_RE.test(key)) {
            keyError.value = 'Use letters, numbers or underscores; must start with a letter'
        }
    }
    if (form.value.field_type === 'enum' && parseOptions().length === 0) {
        optionsError.value = 'Add at least one option'
    }
    if (keyError.value || optionsError.value) {
        return
    }
    saving.value = true
    try {
        const options = form.value.field_type === 'enum' ? parseOptions() : []
        if (editing.value) {
            await fieldStore.updateField(props.project.id, editing.value.id, {
                label: form.value.label.trim() || form.value.key,
                options,
                required: form.value.required,
            })
            toast.success('Field updated')
        } else {
            await fieldStore.createField(props.project.id, {
                key: form.value.key.trim(),
                label: form.value.label.trim() || form.value.key.trim(),
                field_type: form.value.field_type,
                options,
                required: form.value.required,
            })
            toast.success('Field created')
        }
        showModal.value = false
    } catch (e: any) {
        toast.error(e.message || 'Failed to save field')
    } finally {
        saving.value = false
    }
}

const toDelete = ref<FieldDefinition | null>(null)
const deleting = ref(false)

async function confirmDelete() {
    if (!toDelete.value) return
    deleting.value = true
    try {
        await fieldStore.deleteField(props.project.id, toDelete.value.id)
        toast.success('Field deleted')
        toDelete.value = null
    } catch (e: any) {
        toast.error(e.message || 'Failed to delete field')
    } finally {
        deleting.value = false
    }
}

function typeLabel(type: FieldType): string {
    return typeOptions.find(t => t.value === type)?.label ?? type
}

onMounted(() => fieldStore.fetchFields(props.project.id))
</script>

<template>
    <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-6 max-w-lg">
        <div class="flex items-center justify-between mb-2">
            <h2 class="text-sm font-medium text-white">Custom Fields</h2>
            <button @click="openCreate" class="text-sm text-zinc-300 hover:text-white transition cursor-pointer">
                + Add Field
            </button>
        </div>
        <p class="text-xs text-zinc-400 mb-4">
            Typed attributes stored per subscriber. Use them in templates as
            <code class="text-zinc-300">{{ customSyntaxHint }}</code> and to build segments.
        </p>

        <div v-if="definitions.length > 0" class="divide-y divide-zinc-800">
            <div v-for="def in definitions" :key="def.id" class="flex items-center justify-between py-2.5">
                <div class="min-w-0">
                    <p class="text-sm text-white truncate">
                        {{ def.label }}
                        <span v-if="def.required" class="text-red-400">*</span>
                    </p>
                    <p class="text-xs text-zinc-400 font-mono truncate">
                        {{ def.key }} · {{ typeLabel(def.field_type) }}
                    </p>
                </div>
                <div class="flex items-center gap-3 flex-shrink-0 pl-3">
                    <button @click="openEdit(def)" class="text-xs text-zinc-400 hover:text-white transition cursor-pointer">Edit</button>
                    <button @click="toDelete = def" class="text-xs text-zinc-400 hover:text-red-400 transition cursor-pointer">Delete</button>
                </div>
            </div>
        </div>
        <p v-else class="text-sm text-zinc-400">No custom fields yet.</p>

        <AppModal :show="showModal" :title="editing ? 'Edit field' : 'New field'" @close="showModal = false">
            <form @submit.prevent="save" class="space-y-4">
                <AppInput
                    v-if="!editing"
                    v-model="form.key"
                    label="Key"
                    placeholder="plan_tier"
                    :error="keyError" />
                <p v-else class="text-xs text-zinc-400">
                    Key <code class="text-zinc-300 font-mono">{{ form.key }}</code> · {{ typeLabel(form.field_type) }} (not editable)
                </p>

                <AppInput v-model="form.label" label="Label" placeholder="Plan tier" />

                <div v-if="!editing">
                    <label class="block text-sm font-medium text-zinc-300 mb-1">Type</label>
                    <AppSelect v-model="form.field_type" size="md" class="w-full" :options="typeOptions" />
                </div>

                <AppInput
                    v-if="form.field_type === 'enum'"
                    v-model="form.optionsText"
                    label="Options (comma separated)"
                    placeholder="free, pro, team"
                    :error="optionsError" />

                <label class="flex items-center gap-2">
                    <AppCheckbox v-model="form.required" />
                    <span class="text-sm text-zinc-300">Required</span>
                </label>

                <AppButton :loading="saving">{{ editing ? 'Save' : 'Create field' }}</AppButton>
            </form>
        </AppModal>

        <AppConfirmModal
            :show="!!toDelete"
            title="Delete field"
            :message="toDelete ? `Delete field '${toDelete.label}'? Existing subscriber values stay stored until their next update.` : ''"
            confirmLabel="Delete"
            danger
            :loading="deleting"
            @confirm="confirmDelete"
            @cancel="toDelete = null" />
    </div>
</template>

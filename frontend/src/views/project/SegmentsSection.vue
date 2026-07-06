<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useToastStore } from '@/stores/toast'
import { useSegmentStore, type Segment, type SegmentMatch } from '@/stores/segments'
import { useFieldStore, type FieldDefinition } from '@/stores/fields'
import type { Project } from '@/stores/projects'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import AppModal from '@/components/ui/AppModal.vue'
import AppConfirmModal from '@/components/ui/AppConfirmModal.vue'
import AppTagInput from '@/components/ui/AppTagInput.vue'

const props = defineProps<{ project: Project }>()
const toast = useToastStore()
const segmentStore = useSegmentStore()
const fieldStore = useFieldStore()

const segments = computed(() => segmentStore.segments(props.project.id))
const fields = computed(() => fieldStore.fields(props.project.id))

interface EditorRule {
    field: string
    op: string
    value: any
}

type FieldKind = 'status' | 'tags' | 'custom'

interface FieldOption {
    value: string
    label: string
    kind: FieldKind
    def?: FieldDefinition
}

const fieldOptions = computed<FieldOption[]>(() => {
    const base: FieldOption[] = [
        { value: 'status', label: 'Status', kind: 'status' },
        { value: 'tags', label: 'Tags', kind: 'tags' },
    ]
    for (const def of fields.value) {
        base.push({ value: `custom.${def.key}`, label: def.label, kind: 'custom', def })
    }
    return base
})

function fieldOption(field: string): FieldOption | undefined {
    return fieldOptions.value.find(f => f.value === field)
}

function opsFor(field: string): { value: string; label: string }[] {
    const opt = fieldOption(field)
    if (!opt) return [{ value: 'eq', label: 'is' }]
    if (opt.kind === 'status') return [{ value: 'eq', label: 'is' }, { value: 'neq', label: 'is not' }]
    if (opt.kind === 'tags') return [
        { value: 'includes_any', label: 'has any of' },
        { value: 'includes_all', label: 'has all of' },
        { value: 'excludes', label: 'has none of' },
    ]
    const type = opt.def?.field_type
    if (type === 'number' || type === 'date') return [
        { value: 'eq', label: 'is' },
        { value: 'neq', label: 'is not' },
        { value: 'gt', label: type === 'date' ? 'after' : 'greater than' },
        { value: 'lt', label: type === 'date' ? 'before' : 'less than' },
    ]
    if (type === 'string') return [
        { value: 'eq', label: 'is' },
        { value: 'neq', label: 'is not' },
        { value: 'contains', label: 'contains' },
    ]
    return [{ value: 'eq', label: 'is' }, { value: 'neq', label: 'is not' }]
}

function defaultValueFor(field: string): any {
    const opt = fieldOption(field)
    if (opt?.kind === 'tags') return []
    if (opt?.kind === 'status') return 'active'
    if (opt?.def?.field_type === 'boolean') return 'true'
    if (opt?.def?.field_type === 'enum') return opt.def.options[0] ?? ''
    return ''
}

function newRule(): EditorRule {
    return { field: 'status', op: 'eq', value: 'active' }
}

function onFieldChange(rule: EditorRule) {
    rule.op = opsFor(rule.field)[0]?.value ?? 'eq'
    rule.value = defaultValueFor(rule.field)
}

const showModal = ref(false)
const editing = ref<Segment | null>(null)
const saving = ref(false)
const name = ref('')
const nameError = ref('')
const match = ref<SegmentMatch>('all')
const rules = ref<EditorRule[]>([])
const previewCount = ref<number | null>(null)
const previewing = ref(false)

function buildPredicate() {
    return {
        match: match.value,
        rules: rules.value.map(r => {
            const opt = fieldOption(r.field)
            let value: any = r.value
            if (opt?.def?.field_type === 'number' && typeof value === 'string' && value !== '') {
                const parsed = Number(value)
                if (!Number.isNaN(parsed)) value = parsed
            } else if (opt?.def?.field_type === 'boolean') {
                value = value === 'true' || value === true
            }
            return { field: r.field, op: r.op, value }
        }),
    }
}

function openCreate() {
    editing.value = null
    name.value = ''
    nameError.value = ''
    match.value = 'all'
    rules.value = [newRule()]
    previewCount.value = null
    showModal.value = true
}

function openEdit(segment: Segment) {
    editing.value = segment
    nameError.value = ''
    name.value = segment.name
    match.value = segment.predicate?.match || 'all'
    rules.value = (segment.predicate?.rules || []).map(r => ({
        field: r.field,
        op: r.op,
        value: fieldOption(r.field)?.def?.field_type === 'boolean' ? String(r.value) : r.value,
    }))
    if (rules.value.length === 0) rules.value = [newRule()]
    previewCount.value = null
    showModal.value = true
}

function addRule() {
    rules.value.push(newRule())
}

function removeRule(index: number) {
    rules.value.splice(index, 1)
}

let previewTimer: ReturnType<typeof setTimeout> | undefined
watch([rules, match], () => {
    if (!showModal.value) return
    clearTimeout(previewTimer)
    previewTimer = setTimeout(runPreview, 400)
}, { deep: true })

async function runPreview() {
    previewing.value = true
    try {
        previewCount.value = await segmentStore.previewSegment(props.project.id, buildPredicate())
    } catch {
        previewCount.value = null
    } finally {
        previewing.value = false
    }
}

async function save() {
    nameError.value = ''
    if (!name.value.trim()) {
        nameError.value = 'Name is required'
        return
    }
    saving.value = true
    try {
        const predicate = buildPredicate()
        if (editing.value) {
            await segmentStore.updateSegment(props.project.id, editing.value.id, name.value.trim(), predicate)
            toast.success('Segment updated')
        } else {
            await segmentStore.createSegment(props.project.id, name.value.trim(), predicate)
            toast.success('Segment created')
        }
        showModal.value = false
    } catch (e: any) {
        toast.error(e.message || 'Failed to save segment')
    } finally {
        saving.value = false
    }
}

const toDelete = ref<Segment | null>(null)
const deleting = ref(false)

async function confirmDelete() {
    if (!toDelete.value) return
    deleting.value = true
    try {
        await segmentStore.deleteSegment(props.project.id, toDelete.value.id)
        toast.success('Segment deleted')
        toDelete.value = null
    } catch (e: any) {
        toast.error(e.message || 'Failed to delete segment')
    } finally {
        deleting.value = false
    }
}

onMounted(() => {
    segmentStore.fetchSegments(props.project.id)
    fieldStore.fetchFields(props.project.id)
})
</script>

<template>
    <div>
        <div class="flex flex-wrap items-center justify-between gap-3 mb-6">
            <div>
                <h1 class="text-2xl font-bold text-white">Segments</h1>
                <p class="text-sm text-zinc-500 mt-1">Saved filters used to target broadcasts.</p>
            </div>
            <AppButton size="sm" @click="openCreate">+ New Segment</AppButton>
        </div>

        <div v-if="segments.length > 0" class="bg-zinc-900 border border-zinc-800 rounded-lg overflow-x-auto">
            <table class="w-full min-w-[480px]">
                <thead>
                    <tr class="border-b border-zinc-800">
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Name</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Rules</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Match</th>
                        <th class="text-right px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Actions</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="segment in segments" :key="segment.id" class="border-b border-zinc-800 last:border-0 hover:bg-zinc-800/50 transition">
                        <td class="px-4 py-3 text-sm text-white">{{ segment.name }}</td>
                        <td class="px-4 py-3 text-sm text-zinc-400">{{ (segment.predicate?.rules || []).length }} rule(s)</td>
                        <td class="px-4 py-3 text-sm text-zinc-500">{{ segment.predicate?.match === 'any' ? 'Any' : 'All' }}</td>
                        <td class="px-4 py-3 text-right space-x-3">
                            <button @click="openEdit(segment)" class="text-xs text-zinc-500 hover:text-white transition cursor-pointer">Edit</button>
                            <button @click="toDelete = segment" class="text-xs text-zinc-500 hover:text-red-400 transition cursor-pointer">Delete</button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <div v-else class="bg-zinc-900 border border-zinc-800 rounded-lg p-8 text-center">
            <p class="text-zinc-400 mb-2">No segments yet.</p>
            <p class="text-zinc-500 text-sm">Create a segment to send broadcasts to a subset of your subscribers.</p>
        </div>

        <AppModal :show="showModal" :title="editing ? 'Edit segment' : 'New segment'" size="lg" @close="showModal = false">
            <form @submit.prevent="save" class="space-y-4">
                <AppInput v-model="name" label="Name" placeholder="Active pro customers" :error="nameError" />

                <div>
                    <label class="block text-sm font-medium text-zinc-300 mb-1">Match</label>
                    <select v-model="match"
                        class="w-full px-3 py-2 bg-zinc-900 border border-zinc-800 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-zinc-500 transition">
                        <option value="all">All rules (AND)</option>
                        <option value="any">Any rule (OR)</option>
                    </select>
                </div>

                <div class="space-y-2">
                    <div v-for="(rule, index) in rules" :key="index" class="p-3 bg-zinc-950 border border-zinc-800 rounded-lg space-y-2">
                        <div class="flex gap-2">
                            <select v-model="rule.field" @change="onFieldChange(rule)"
                                class="flex-1 px-2 py-1.5 text-sm bg-zinc-900 border border-zinc-800 rounded-md text-white focus:outline-none focus:ring-1 focus:ring-zinc-500">
                                <option v-for="opt in fieldOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
                            </select>
                            <select v-model="rule.op"
                                class="flex-1 px-2 py-1.5 text-sm bg-zinc-900 border border-zinc-800 rounded-md text-white focus:outline-none focus:ring-1 focus:ring-zinc-500">
                                <option v-for="op in opsFor(rule.field)" :key="op.value" :value="op.value">{{ op.label }}</option>
                            </select>
                            <button type="button" @click="removeRule(index)" class="px-2 text-zinc-500 hover:text-red-400 transition cursor-pointer">&times;</button>
                        </div>

                        <AppTagInput v-if="fieldOption(rule.field)?.kind === 'tags'" v-model="rule.value" placeholder="Add tags…" />

                        <select v-else-if="fieldOption(rule.field)?.kind === 'status'" v-model="rule.value"
                            class="w-full px-2 py-1.5 text-sm bg-zinc-900 border border-zinc-800 rounded-md text-white focus:outline-none focus:ring-1 focus:ring-zinc-500">
                            <option value="active">active</option>
                            <option value="pending">pending</option>
                            <option value="unsubscribed">unsubscribed</option>
                        </select>

                        <select v-else-if="fieldOption(rule.field)?.def?.field_type === 'enum'" v-model="rule.value"
                            class="w-full px-2 py-1.5 text-sm bg-zinc-900 border border-zinc-800 rounded-md text-white focus:outline-none focus:ring-1 focus:ring-zinc-500">
                            <option v-for="opt in fieldOption(rule.field)?.def?.options || []" :key="opt" :value="opt">{{ opt }}</option>
                        </select>

                        <select v-else-if="fieldOption(rule.field)?.def?.field_type === 'boolean'" v-model="rule.value"
                            class="w-full px-2 py-1.5 text-sm bg-zinc-900 border border-zinc-800 rounded-md text-white focus:outline-none focus:ring-1 focus:ring-zinc-500">
                            <option value="true">true</option>
                            <option value="false">false</option>
                        </select>

                        <input v-else v-model="rule.value"
                            :type="fieldOption(rule.field)?.def?.field_type === 'number' ? 'number' : fieldOption(rule.field)?.def?.field_type === 'date' ? 'date' : 'text'"
                            placeholder="Value"
                            class="w-full px-2 py-1.5 text-sm bg-zinc-900 border border-zinc-800 rounded-md text-white placeholder-zinc-500 focus:outline-none focus:ring-1 focus:ring-zinc-500" />
                    </div>

                    <button type="button" @click="addRule" class="text-sm text-zinc-400 hover:text-white transition cursor-pointer">+ Add rule</button>
                </div>

                <div class="flex items-center justify-between pt-2 border-t border-zinc-800">
                    <p class="text-sm text-zinc-400">
                        <span v-if="previewing" class="text-zinc-500">Counting…</span>
                        <span v-else-if="previewCount !== null">{{ previewCount }} active subscriber(s) match</span>
                        <span v-else class="text-zinc-500">—</span>
                    </p>
                    <AppButton :loading="saving" class="w-auto! px-4">{{ editing ? 'Save' : 'Create segment' }}</AppButton>
                </div>
            </form>
        </AppModal>

        <AppConfirmModal
            :show="!!toDelete"
            title="Delete segment"
            :message="toDelete ? `Delete segment '${toDelete.name}'? Broadcasts already sent are unaffected.` : ''"
            confirmLabel="Delete"
            danger
            :loading="deleting"
            @confirm="confirmDelete"
            @cancel="toDelete = null" />
    </div>
</template>

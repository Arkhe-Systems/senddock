<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, nextTick, computed } from 'vue'
import { api } from '@/api/client'
import { useToastStore } from '@/stores/toast'
import { useFieldStore } from '@/stores/fields'
import type { Project } from '@/stores/projects'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import AppModal from '@/components/ui/AppModal.vue'
import AppConfirmModal from '@/components/ui/AppConfirmModal.vue'
import AppCopyId from '@/components/ui/AppCopyId.vue'
import { Codemirror } from 'vue-codemirror'
import { html } from '@codemirror/lang-html'
import { oneDark } from '@codemirror/theme-one-dark'
import { EditorView } from '@codemirror/view'
import EmailEditor from '@/components/ui/EmailEditor.vue'
import TemplateLibraryBrowser from '@/views/project/TemplateLibraryBrowser.vue'
import * as prettier from 'prettier/standalone'
import htmlPlugin from 'prettier/plugins/html'
import { detectTemplateVariables, systemVariablesFor } from '@/utils/templateVariables'

interface Template {
    id: string
    name: string
    subject: string
    html_body: string
    text_body: string
    type?: string
    created_at: string
    updated_at: string
}

const props = defineProps<{ project: Project }>()
const toast = useToastStore()
const fieldStore = useFieldStore()

const customVariables = computed(() => fieldStore.fields(props.project.id))

const templates = ref<Template[]>([])
const visibleTemplates = computed(() => typeFilter.value === 'all' ? templates.value : templates.value.filter(t => (t.type || 'email') === typeFilter.value))
const loading = ref(true)

const extensions = [html(), oneDark, EditorView.lineWrapping]

const showCreateModal = ref(false)
const newName = ref('')
const newType = ref<'email' | 'page'>('email')
const typeFilter = ref<'all' | 'email' | 'page'>('all')
const createLoading = ref(false)

const showLibraryModal = ref(false)

function handleLibraryUsed(tmpl: Template) {
    fetchTemplates()
    openEditor(tmpl)
}

const editing = ref<Template | null>(null)
const editName = ref('')
const editSubject = ref('')
const editHtml = ref('')
const saveLoading = ref(false)
const activeTab = ref<'code' | 'visual'>('code')
const emailEditorRef = ref<{ flush: () => void; refresh: () => void; insertContent: (html: string) => void } | null>(null)
const showDiscardModal = ref(false)

const originalName = ref('')
const originalSubject = ref('')
const originalHtml = ref('')

const isDirty = computed(() =>
    editing.value !== null &&
    (editName.value !== originalName.value ||
        editSubject.value !== originalSubject.value ||
        editHtml.value !== originalHtml.value)
)

const previewHtml = computed(() => {
    let html = editHtml.value
        .replace(/\{\{\s*name\s*\}\}/g, 'John Doe')
        .replace(/\{\{\s*email\s*\}\}/g, 'john@example.com')
        .replace(/\{\{\s*subscriber_id\s*\}\}/g, 'sub_1234567890')
        .replace(/\{\{\s*unsubscribe_url\s*\}\}/g, '#')
    for (const field of customVariables.value) {
        const escaped = field.key.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
        html = html.replace(new RegExp(`\\{\\{\\s*custom\\.${escaped}\\s*\\}\\}`, 'g'), field.label || field.key)
    }
    return html
})

const detectedVariables = computed(() => detectTemplateVariables(editHtml.value, editSubject.value))

async function fetchTemplates() {
    loading.value = true
    try {
        const res = await api<Template[] | null>(`/projects/${props.project.id}/templates`)
        templates.value = res || []
    } catch {
        templates.value = []
    } finally {
        loading.value = false
    }
}

async function handleCreate() {
    if (!newName.value) {
        toast.error('Name is required')
        return
    }
    createLoading.value = true
    try {
        const tmpl = await api<Template>(`/projects/${props.project.id}/templates`, {
            method: 'POST',
            body: { name: newName.value, subject: '', html_body: '', text_body: '', type: newType.value },
        })
        showCreateModal.value = false
        newName.value = ''
        newType.value = 'email'
        toast.success('Template created')
        openEditor(tmpl)
        fetchTemplates()
    } catch (e: any) {
        toast.error(e.message || 'Failed to create template')
    } finally {
        createLoading.value = false
    }
}

function openEditor(tmpl: Template) {
    editing.value = tmpl
    editName.value = tmpl.name
    editSubject.value = tmpl.subject
    editHtml.value = tmpl.html_body
    originalName.value = tmpl.name
    originalSubject.value = tmpl.subject
    originalHtml.value = tmpl.html_body
    activeTab.value = 'code'
}

function switchTab(tab: 'code' | 'visual') {
    activeTab.value = tab
    if (tab === 'visual') {
        nextTick(() => emailEditorRef.value?.refresh())
    }
}

function closeEditor() {
    if (isDirty.value) {
        showDiscardModal.value = true
        return
    }
    editing.value = null
    fetchTemplates()
}

function confirmDiscard() {
    showDiscardModal.value = false
    editing.value = null
    fetchTemplates()
}

function htmlToText(html: string): string {
    return html
        .replace(/<style[\s\S]*?<\/style>/gi, '')
        .replace(/<(br|\/p|\/div|\/tr|\/h[1-6]|\/li)[^>]*>/gi, '\n')
        .replace(/<[^>]+>/g, '')
        .replace(/&nbsp;/gi, ' ')
        .replace(/&amp;/gi, '&')
        .replace(/&lt;/gi, '<')
        .replace(/&gt;/gi, '>')
        .replace(/[ \t]+/g, ' ')
        .replace(/\n[ \t]*\n+/g, '\n')
        .trim()
}

function onBeforeUnload(e: BeforeUnloadEvent) {
    if (isDirty.value) {
        e.preventDefault()
        e.returnValue = ''
    }
}

async function formatHtml(html: string): Promise<string> {
    try {
        return await prettier.format(html, {
            parser: 'html',
            plugins: [htmlPlugin],
            printWidth: 120,
            tabWidth: 2,
        })
    } catch {
        return html
    }
}

async function handleSave() {
    if (!editing.value || !editName.value) return
    saveLoading.value = true
    try {
        emailEditorRef.value?.flush()
        const htmlBody = await formatHtml(editHtml.value)
        await api(`/projects/${props.project.id}/templates/${editing.value.id}`, {
            method: 'PUT',
            body: {
                name: editName.value,
                subject: editSubject.value,
                html_body: htmlBody,
                text_body: htmlToText(editHtml.value),
            },
        })
        editHtml.value = htmlBody
        originalName.value = editName.value
        originalSubject.value = editSubject.value
        originalHtml.value = htmlBody
        toast.success('Template saved')
    } catch (e: any) {
        toast.error(e.message || 'Failed to save template')
    } finally {
        saveLoading.value = false
    }
}

const showDeleteModal = ref(false)
const templateToDelete = ref<Template | null>(null)
const deleteConfirmName = ref('')
const deleteLoading = ref(false)

function openDeleteModal(tmpl: Template) {
    templateToDelete.value = tmpl
    deleteConfirmName.value = ''
    showDeleteModal.value = true
}

async function handleDelete() {
    if (!templateToDelete.value || deleteConfirmName.value !== templateToDelete.value.name) return
    deleteLoading.value = true
    try {
        await api(`/projects/${props.project.id}/templates/${templateToDelete.value.id}`, { method: 'DELETE' })
        toast.success('Template deleted')
        if (editing.value?.id === templateToDelete.value.id) {
            editing.value = null
        }
        showDeleteModal.value = false
        templateToDelete.value = null
        fetchTemplates()
    } catch (e: any) {
        toast.error(e.message || 'Failed to delete template')
    } finally {
        deleteLoading.value = false
    }
}

const systemVariables = computed(() => systemVariablesFor(editing.value?.type as 'email' | 'page' | undefined))

let cmView: any = null
function onCmReady(payload: any) {
    cmView = payload.view
}

function insertVariable(v: string) {
    const text = `{{${v}}}`
    if (activeTab.value === 'code' && cmView) {
        cmView.dispatch({ changes: { from: cmView.state.selection.main.head, insert: text } })
    } else if (activeTab.value === 'visual') {
        emailEditorRef.value?.insertContent(text)
    } else {
        editHtml.value += text
    }
}

function varLabel(v: string): string {
    return '{{' + v + '}}'
}

onMounted(() => {
    window.addEventListener('beforeunload', onBeforeUnload)
    fetchTemplates()
    fieldStore.fetchFields(props.project.id)
})

onBeforeUnmount(() => {
    window.removeEventListener('beforeunload', onBeforeUnload)
})
</script>

<template>
    <div>
        <div v-if="editing">
            <div class="flex items-center justify-between mb-4">
                <div class="flex items-center gap-3">
                    <button @click="closeEditor" class="text-zinc-300 hover:text-white transition cursor-pointer">&larr;</button>
                    <h1 class="text-xl font-semibold text-white">{{ editName }}</h1>
                </div>
                <AppButton size="md" :loading="saveLoading" @click="handleSave">
                    {{ saveLoading ? 'Saving...' : 'Save' }}
                </AppButton>
            </div>

            <div class="grid grid-cols-2 gap-4 mb-4 max-w-xl">
                <AppInput v-model="editName" label="Template Name" required />
                <AppInput v-model="editSubject" label="Email Subject" placeholder="Welcome!" />
            </div>

            <div class="flex gap-4" style="height: calc(100vh - 280px);">
                <div :class="['flex flex-col min-w-0', activeTab === 'visual' ? 'flex-1' : 'flex-1']">
                    <div class="flex gap-1 border-b border-zinc-800">
                        <button @click="switchTab('code')"
                            :class="[
                                'px-4 py-2 text-sm transition cursor-pointer border-b-2 -mb-px',
                                activeTab === 'code'
                                    ? 'text-white border-white'
                                    : 'text-zinc-400 border-transparent hover:text-zinc-300'
                            ]">
                            Code
                        </button>
                        <button @click="switchTab('visual')"
                            :class="[
                                'px-4 py-2 text-sm transition cursor-pointer border-b-2 -mb-px',
                                activeTab === 'visual'
                                    ? 'text-white border-white'
                                    : 'text-zinc-400 border-transparent hover:text-zinc-300'
                            ]">
                            Visual
                        </button>
                    </div>

                    <div v-show="activeTab === 'code'" class="flex-1 border border-zinc-800 border-t-0 rounded-b-lg overflow-hidden">
                        <Codemirror
                            v-model="editHtml"
                            :extensions="extensions"
                            :style="{ height: '100%', fontSize: '13px' }"
                            @ready="onCmReady"
                            placeholder="<h1>Hello {{name}}</h1>
<p>Welcome to our newsletter!</p>"
                        />
                    </div>

                    <div v-show="activeTab === 'visual'" class="flex-1 border border-zinc-800 border-t-0 rounded-b-lg overflow-hidden">
                        <EmailEditor ref="emailEditorRef" v-model="editHtml" />
                    </div>
                </div>

                <div v-show="activeTab === 'code'" class="flex-1 flex flex-col min-w-0">
                    <div class="flex items-center px-4 py-2 border-b border-zinc-800">
                        <span class="text-sm text-zinc-300">Preview</span>
                    </div>
                    <div class="flex-1 border border-zinc-800 border-t-0 rounded-b-lg overflow-hidden bg-white">
                        <iframe v-if="editHtml" :srcdoc="previewHtml" class="w-full h-full border-0" sandbox="" />
                        <div v-else class="flex items-center justify-center h-full">
                            <p class="text-zinc-300 text-sm">Write HTML to see a preview</p>
                        </div>
                    </div>
                </div>
            </div>

            <div class="mt-3 p-3 bg-zinc-900 border border-zinc-800 rounded-lg space-y-3">
                <p v-pre class="text-xs text-zinc-300 font-medium">Variables — click one to insert <code class="text-zinc-400 font-mono">{{variable}}</code></p>

                <div v-if="detectedVariables.length > 0">
                    <p class="text-[10px] font-medium text-zinc-500 uppercase tracking-wide mb-1.5">Detected in this template</p>
                    <div class="flex flex-wrap gap-1.5">
                        <button v-for="v in detectedVariables" :key="'d' + v" type="button" @click="insertVariable(v)"
                            class="text-xs bg-zinc-850 text-zinc-300 px-2 py-1 rounded border border-zinc-700 font-mono hover:border-emerald-500/60 hover:text-white transition cursor-pointer">
                            {{ varLabel(v) }}
                        </button>
                    </div>
                </div>

                <div>
                    <p class="text-[10px] font-medium text-zinc-500 uppercase tracking-wide mb-1.5">Available</p>
                    <div class="flex flex-wrap gap-1.5">
                        <button v-for="v in systemVariables" :key="'s' + v" type="button" @click="insertVariable(v)"
                            class="text-xs bg-zinc-850 text-zinc-300 px-2 py-1 rounded border border-zinc-700 font-mono hover:border-emerald-500/60 hover:text-white transition cursor-pointer">
                            {{ varLabel(v) }}
                        </button>
                    </div>
                </div>

                <div v-if="customVariables.length > 0 && editing?.type !== 'page'">
                    <p class="text-[10px] font-medium text-zinc-500 uppercase tracking-wide mb-1.5">Custom fields</p>
                    <div class="flex flex-wrap gap-1.5">
                        <button v-for="def in customVariables" :key="'c' + def.id" type="button"
                            @click="insertVariable('custom.' + def.key)"
                            :title="def.label"
                            class="text-xs bg-zinc-850 text-zinc-300 px-2 py-1 rounded border border-zinc-700 font-mono hover:border-emerald-500/60 hover:text-white transition cursor-pointer">
                            {{ varLabel('custom.' + def.key) }}
                        </button>
                    </div>
                </div>
            </div>
        </div>

        <div v-else>
            <div class="flex flex-wrap items-center justify-between gap-3 mb-6">
                <div class="flex items-center gap-4">
                    <h1 class="text-xl font-semibold text-white">Templates</h1>
                    <div class="flex gap-1 bg-zinc-900 border border-zinc-800 rounded-lg p-0.5">
                        <button v-for="t in (['all', 'email', 'page'] as const)" :key="t" @click="typeFilter = t"
                            :class="['px-2.5 py-1 text-xs rounded-md transition cursor-pointer capitalize', typeFilter === t ? 'bg-emerald-500/15 text-emerald-400' : 'text-zinc-400 hover:text-white']">
                            {{ t }}
                        </button>
                    </div>
                </div>
                <div class="flex items-center gap-2">
                    <AppButton size="md" variant="secondary" @click="showLibraryModal = true">
                        ★ Browse library
                    </AppButton>
                    <AppButton size="md" @click="showCreateModal = true">+ New Template</AppButton>
                </div>
            </div>

            <div v-if="loading" class="text-zinc-400 py-8 text-center">Loading...</div>

            <div v-else-if="visibleTemplates.length > 0" class="space-y-2">
                <div v-for="tmpl in visibleTemplates" :key="tmpl.id"
                    class="bg-zinc-900 border border-zinc-800 rounded-lg p-4 flex items-center justify-between hover:border-zinc-700 transition cursor-pointer"
                    @click="openEditor(tmpl)">
                    <div>
                        <p class="text-white font-medium">
                            {{ tmpl.name }}
                            <span v-if="tmpl.type === 'page'" class="ml-2 align-middle text-[10px] font-semibold uppercase tracking-wide bg-zinc-850 border border-zinc-700 text-zinc-300 rounded px-1.5 py-0.5">Page</span>
                        </p>
                        <p class="text-sm text-zinc-400 mt-1">
                            {{ tmpl.subject || 'No subject set' }}
                        </p>
                    </div>
                    <div class="flex items-center gap-3">
                        <AppCopyId :value="tmpl.id" />
                        <span class="text-xs text-zinc-400">{{ new Date(tmpl.updated_at).toLocaleDateString() }}</span>
                        <button @click.stop="openDeleteModal(tmpl)"
                            class="text-xs text-zinc-400 hover:text-red-400 transition cursor-pointer">
                            Delete
                        </button>
                    </div>
                </div>
            </div>

            <div v-else class="bg-zinc-900 border border-zinc-800 rounded-lg p-8 text-center">
                <p class="text-zinc-300 mb-2">No templates yet.</p>
                <p class="text-zinc-400 text-sm">Create email templates to use when sending to your subscribers.</p>
            </div>
        </div>

        <AppModal :show="showCreateModal" title="New Template" @close="showCreateModal = false">
            <form @submit.prevent="handleCreate" class="space-y-4">
                <AppInput v-model="newName" label="Template Name" placeholder="Welcome Email" required />
                <div>
                    <label class="block text-sm font-medium text-zinc-300 mb-2">Type</label>
                    <div class="grid grid-cols-2 gap-2">
                        <button type="button" @click="newType = 'email'"
                            :class="['px-3 py-2.5 text-sm rounded-lg border text-left transition cursor-pointer', newType === 'email' ? 'bg-zinc-850 border-zinc-600 text-white' : 'bg-zinc-900 border-zinc-800 text-zinc-300 hover:text-white']">
                            <span class="block font-medium">Email</span>
                            <span class="block text-xs text-zinc-400 mt-0.5">Sent to subscribers</span>
                        </button>
                        <button type="button" @click="newType = 'page'"
                            :class="['px-3 py-2.5 text-sm rounded-lg border text-left transition cursor-pointer', newType === 'page' ? 'bg-zinc-850 border-zinc-600 text-white' : 'bg-zinc-900 border-zinc-800 text-zinc-300 hover:text-white']">
                            <span class="block font-medium">Unsubscribe page</span>
                            <span class="block text-xs text-zinc-400 mt-0.5">Branded public page</span>
                        </button>
                    </div>
                    <p v-if="newType === 'page'" class="text-xs text-zinc-400 mt-2">Placeholders: <code class="font-mono" v-pre>{{project_name}}</code> <code class="font-mono" v-pre>{{email}}</code> <code class="font-mono" v-pre>{{newsletter_name}}</code> <code class="font-mono" v-pre>{{confirm_button}}</code>. Select it under Settings → Unsubscribe page.</p>
                </div>
                <AppButton :loading="createLoading">
                    {{ createLoading ? 'Creating...' : 'Create Template' }}
                </AppButton>
            </form>
        </AppModal>

        <TemplateLibraryBrowser
            :show="showLibraryModal"
            :project-id="props.project.id"
            @close="showLibraryModal = false"
            @used="handleLibraryUsed" />


        <AppConfirmModal
            :show="showDiscardModal"
            title="Discard changes?"
            message="You have unsaved changes to this template. Discard them and leave?"
            confirm-label="Discard"
            cancel-label="Keep editing"
            @confirm="confirmDiscard"
            @cancel="showDiscardModal = false"
        />


        <AppModal :show="showDeleteModal" title="Delete Template" @close="showDeleteModal = false">
            <div class="space-y-4">
                <p class="text-zinc-300 text-sm">
                    This action cannot be undone. Type
                    <span class="font-semibold text-white">{{ templateToDelete?.name }}</span>
                    to confirm.
                </p>
                <AppInput v-model="deleteConfirmName" placeholder="Type template name to confirm" />
                <AppButton variant="danger" :disabled="deleteConfirmName !== templateToDelete?.name" :loading="deleteLoading"
                    @click="handleDelete">
                    {{ deleteLoading ? 'Deleting...' : 'Delete Template' }}
                </AppButton>
            </div>
        </AppModal>
    </div>
</template>

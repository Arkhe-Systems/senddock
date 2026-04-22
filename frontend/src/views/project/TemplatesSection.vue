<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { api } from '@/api/client'
import { useToastStore } from '@/stores/toast'
import type { Project } from '@/stores/projects'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import AppModal from '@/components/ui/AppModal.vue'
import { Codemirror } from 'vue-codemirror'
import { html } from '@codemirror/lang-html'
import { oneDark } from '@codemirror/theme-one-dark'
import { EditorView } from '@codemirror/view'
import EmailEditor from '@/components/ui/EmailEditor.vue'

interface Template {
    id: string
    name: string
    subject: string
    html_body: string
    text_body: string
    created_at: string
    updated_at: string
}

const props = defineProps<{ project: Project }>()
const toast = useToastStore()

const templates = ref<Template[]>([])
const loading = ref(true)

const extensions = [html(), oneDark, EditorView.lineWrapping]

const showCreateModal = ref(false)
const newName = ref('')
const createLoading = ref(false)

const editing = ref<Template | null>(null)
const editName = ref('')
const editSubject = ref('')
const editHtml = ref('')
const saveLoading = ref(false)
const activeTab = ref<'code' | 'visual'>('code')

const previewHtml = computed(() => {
    return editHtml.value
        .replace(/\{\{name\}\}/g, 'John Doe')
        .replace(/\{\{email\}\}/g, 'john@example.com')
})

const detectedVariables = computed(() => {
    const text = editHtml.value + ' ' + editSubject.value
    const regex = /\{\{\s*([a-zA-Z0-9_]+)\s*\}\}/g
    const matches = Array.from(text.matchAll(regex)).map(m => m[1])
    return [...new Set(matches)]
})

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
            body: { name: newName.value, subject: '', html_body: '', text_body: '' },
        })
        showCreateModal.value = false
        newName.value = ''
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
    activeTab.value = 'code'
}

function closeEditor() {
    editing.value = null
    fetchTemplates()
}

async function handleSave() {
    if (!editing.value || !editName.value) return
    saveLoading.value = true
    try {
        await api(`/projects/${props.project.id}/templates/${editing.value.id}`, {
            method: 'PUT',
            body: {
                name: editName.value,
                subject: editSubject.value,
                html_body: editHtml.value,
                text_body: '',
            },
        })
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

async function copyId(id: string) {
    await navigator.clipboard.writeText(id)
    toast.success('ID copied')
}

onMounted(fetchTemplates)
</script>

<template>
    <div>
        <div v-if="editing">
            <div class="flex items-center justify-between mb-4">
                <div class="flex items-center gap-3">
                    <button @click="closeEditor" class="text-zinc-400 hover:text-white transition cursor-pointer">&larr;</button>
                    <h1 class="text-2xl font-bold text-white">{{ editName }}</h1>
                </div>
                <AppButton :loading="saveLoading" @click="handleSave" class="w-auto! px-4">
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
                        <button @click="activeTab = 'code'"
                            :class="[
                                'px-4 py-2 text-sm transition cursor-pointer border-b-2 -mb-px',
                                activeTab === 'code'
                                    ? 'text-white border-white'
                                    : 'text-zinc-500 border-transparent hover:text-zinc-300'
                            ]">
                            Code
                        </button>
                        <button @click="activeTab = 'visual'"
                            :class="[
                                'px-4 py-2 text-sm transition cursor-pointer border-b-2 -mb-px',
                                activeTab === 'visual'
                                    ? 'text-white border-white'
                                    : 'text-zinc-500 border-transparent hover:text-zinc-300'
                            ]">
                            Visual
                        </button>
                    </div>

                    <div v-show="activeTab === 'code'" class="flex-1 border border-zinc-800 border-t-0 rounded-b-lg overflow-hidden">
                        <Codemirror
                            v-model="editHtml"
                            :extensions="extensions"
                            :style="{ height: '100%', fontSize: '13px' }"
                            placeholder="<h1>Hello {{name}}</h1>
<p>Welcome to our newsletter!</p>"
                        />
                    </div>

                    <div v-show="activeTab === 'visual'" class="flex-1 border border-zinc-800 border-t-0 rounded-b-lg overflow-hidden">
                        <EmailEditor v-model="editHtml" />
                    </div>
                </div>

                <div v-show="activeTab === 'code'" class="flex-1 flex flex-col min-w-0">
                    <div class="flex items-center px-4 py-2 border-b border-zinc-800">
                        <span class="text-sm text-zinc-400">Preview</span>
                    </div>
                    <div class="flex-1 border border-zinc-800 border-t-0 rounded-b-lg overflow-hidden bg-white">
                        <iframe v-if="editHtml" :srcdoc="previewHtml" class="w-full h-full border-0" sandbox="" />
                        <div v-else class="flex items-center justify-center h-full">
                            <p class="text-zinc-400 text-sm">Write HTML to see a preview</p>
                        </div>
                    </div>
                </div>
            </div>

            <div class="mt-3 p-3 bg-zinc-900 border border-zinc-800 rounded-lg">
                <p class="text-xs text-zinc-400 font-medium mb-2">Variables in this template:</p>
                <div class="flex flex-wrap gap-2">
                    <span v-for="v in detectedVariables" :key="v" class="text-xs bg-zinc-800 text-zinc-300 px-2 py-1 rounded border border-zinc-700 font-mono">
                        {{ `{{${v}}}` }}
                    </span>
                    <span v-if="detectedVariables.length === 0" class="text-xs text-zinc-500">None detected</span>
                </div>
                <p class="text-xs text-zinc-500 mt-2">
                    System variables: <code class="text-zinc-400">name</code>,
                    <code class="text-zinc-400">email</code>,
                    <code class="text-zinc-400">subscriber_id</code>,
                    <code class="text-zinc-400">unsubscribe_url</code>
                </p>
            </div>
        </div>

        <div v-else>
            <div class="flex items-center justify-between mb-6">
                <h1 class="text-2xl font-bold text-white">Templates</h1>
                <AppButton @click="showCreateModal = true" class="w-auto! px-4">+ New Template</AppButton>
            </div>

            <div v-if="loading" class="text-zinc-500 py-8 text-center">Loading...</div>

            <div v-else-if="templates.length > 0" class="space-y-2">
                <div v-for="tmpl in templates" :key="tmpl.id"
                    class="bg-zinc-900 border border-zinc-800 rounded-lg p-4 flex items-center justify-between hover:border-zinc-700 transition cursor-pointer"
                    @click="openEditor(tmpl)">
                    <div>
                        <p class="text-white font-medium">{{ tmpl.name }}</p>
                        <p class="text-sm text-zinc-500 mt-1">
                            {{ tmpl.subject || 'No subject set' }}
                        </p>
                    </div>
                    <div class="flex items-center gap-3">
                        <button @click.stop="copyId(tmpl.id)"
                            class="text-xs text-zinc-400 hover:text-white transition cursor-pointer font-mono">
                            Copy ID
                        </button>
                        <span class="text-xs text-zinc-500">{{ new Date(tmpl.updated_at).toLocaleDateString() }}</span>
                        <button @click.stop="openDeleteModal(tmpl)"
                            class="text-xs text-zinc-500 hover:text-red-400 transition cursor-pointer">
                            Delete
                        </button>
                    </div>
                </div>
            </div>

            <div v-else class="bg-zinc-900 border border-zinc-800 rounded-lg p-8 text-center">
                <p class="text-zinc-400 mb-2">No templates yet.</p>
                <p class="text-zinc-500 text-sm">Create email templates to use when sending to your subscribers.</p>
            </div>
        </div>

        <AppModal :show="showCreateModal" title="New Template" @close="showCreateModal = false">
            <form @submit.prevent="handleCreate" class="space-y-4">
                <AppInput v-model="newName" label="Template Name" placeholder="Welcome Email" required />
                <AppButton :loading="createLoading">
                    {{ createLoading ? 'Creating...' : 'Create Template' }}
                </AppButton>
            </form>
        </AppModal>

        <AppModal :show="showDeleteModal" title="Delete Template" @close="showDeleteModal = false">
            <div class="space-y-4">
                <p class="text-zinc-400 text-sm">
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

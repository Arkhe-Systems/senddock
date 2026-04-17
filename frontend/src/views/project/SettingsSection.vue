<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api/client'
import { useToastStore } from '@/stores/toast'
import type { Project } from '@/stores/projects'
import AppInput from '@/components/ui/AppInput.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppModal from '@/components/ui/AppModal.vue'

interface APIKey {
    id: string
    name: string
    key_prefix: string
    last_used_at: string | null
    created_at: string
}

interface APIKeyCreateResult {
    key: string
    api_key: APIKey
}

const props = defineProps<{ project: Project }>()
const emit = defineEmits<{ updated: [] }>()

const router = useRouter()
const toast = useToastStore()

// General
const projectName = ref('')
const projectDescription = ref('')
const generalLoading = ref(false)

// API Keys
const apiKeys = ref<APIKey[]>([])
const showKeyModal = ref(false)
const newKeyName = ref('')
const keyLoading = ref(false)
const createdKey = ref('')  // shown once after creation

// Delete project
const deleteConfirmName = ref('')
const deleteLoading = ref(false)

onMounted(() => {
    projectName.value = props.project.name
    projectDescription.value = props.project.description ?? ''
    fetchAPIKeys()
})

async function copyProjectId() {
    await navigator.clipboard.writeText(props.project.id)
    toast.success('Project ID copied')
}

async function handleSaveGeneral() {
    if (!projectName.value) {
        toast.error('Name is required')
        return
    }
    generalLoading.value = true
    try {
        await api(`/projects/${props.project.id}`, {
            method: 'PUT',
            body: { name: projectName.value, description: projectDescription.value },
        })
        toast.success('Project updated')
        emit('updated')
    } catch (e: any) {
        toast.error(e.message || 'Failed to update project')
    } finally {
        generalLoading.value = false
    }
}

// API Keys
async function fetchAPIKeys() {
    try {
        const res = await api<APIKey[] | null>(`/projects/${props.project.id}/keys`)
        apiKeys.value = res || []
    } catch {
        apiKeys.value = []
    }
}

async function handleCreateKey() {
    if (!newKeyName.value) {
        toast.error('Name is required')
        return
    }
    keyLoading.value = true
    try {
        const result = await api<APIKeyCreateResult>(`/projects/${props.project.id}/keys`, {
            method: 'POST',
            body: { name: newKeyName.value },
        })
        createdKey.value = result.key
        newKeyName.value = ''
        toast.success('API key created')
        fetchAPIKeys()
    } catch (e: any) {
        toast.error(e.message || 'Failed to create API key')
    } finally {
        keyLoading.value = false
    }
}

function closeKeyModal() {
    showKeyModal.value = false
    createdKey.value = ''
    newKeyName.value = ''
}

async function copyKey() {
    await navigator.clipboard.writeText(createdKey.value)
    toast.success('Key copied to clipboard')
}

async function handleRevokeKey(key: APIKey) {
    try {
        await api(`/projects/${props.project.id}/keys/${key.id}`, { method: 'DELETE' })
        toast.success('API key revoked')
        fetchAPIKeys()
    } catch (e: any) {
        toast.error(e.message || 'Failed to revoke key')
    }
}

// Delete project
async function handleDelete() {
    if (deleteConfirmName.value !== props.project.name) return
    deleteLoading.value = true
    try {
        await api(`/projects/${props.project.id}`, { method: 'DELETE' })
        toast.success('Project deleted')
        router.push('/dashboard')
    } catch (e: any) {
        toast.error(e.message || 'Failed to delete project')
    } finally {
        deleteLoading.value = false
    }
}
</script>

<template>
    <div class="space-y-8">
        <h1 class="text-2xl font-bold text-white">Settings</h1>

        <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-6 max-w-lg">
            <h2 class="text-sm font-medium text-white mb-4">General</h2>
            <form @submit.prevent="handleSaveGeneral" class="space-y-4">
                <AppInput v-model="projectName" label="Project Name" required />
                <AppInput v-model="projectDescription" large label="Description" placeholder="What is this project about?" />
                <div class="pt-2">
                    <div class="flex items-center gap-2 mb-3">
                        <p class="text-xs text-zinc-500">Project ID:</p>
                        <code class="text-xs text-zinc-400 font-mono">{{ project.id }}</code>
                        <button type="button" @click="copyProjectId"
                            class="text-xs text-zinc-600 hover:text-zinc-400 transition cursor-pointer">
                            Copy
                        </button>
                    </div>
                    <AppButton :loading="generalLoading" class="w-auto! px-4">
                        {{ generalLoading ? 'Saving...' : 'Save Changes' }}
                    </AppButton>
                </div>
            </form>
        </div>

        <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-6 max-w-lg">
            <div class="flex items-center justify-between mb-4">
                <h2 class="text-sm font-medium text-white">API Keys</h2>
                <button @click="showKeyModal = true"
                    class="text-sm text-zinc-400 hover:text-white transition cursor-pointer">
                    + Create Key
                </button>
            </div>

            <p class="text-xs text-zinc-500 mb-4">
                Use API keys to authenticate requests from external applications. Keys are shown only once when created.
            </p>

            <div v-if="apiKeys.length > 0" class="space-y-2">
                <div v-for="key in apiKeys" :key="key.id"
                    class="flex items-center justify-between py-2 border-b border-zinc-800 last:border-0">
                    <div>
                        <p class="text-sm text-white">{{ key.name }}</p>
                        <p class="text-xs text-zinc-500 mt-0.5">
                            <code>{{ key.key_prefix }}...</code>
                            <span class="ml-2">Created {{ new Date(key.created_at).toLocaleDateString() }}</span>
                            <span v-if="key.last_used_at" class="ml-2">Last used {{ new Date(key.last_used_at).toLocaleDateString() }}</span>
                        </p>
                    </div>
                    <button @click="handleRevokeKey(key)"
                        class="text-xs text-zinc-500 hover:text-red-400 transition cursor-pointer">
                        Revoke
                    </button>
                </div>
            </div>

            <p v-else class="text-sm text-zinc-500">No API keys yet.</p>
        </div>

        <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-6 max-w-lg opacity-60">
            <div class="flex items-center gap-2 mb-2">
                <h2 class="text-sm font-medium text-white">Team Members</h2>
                <span class="text-xs px-2 py-0.5 bg-zinc-800 text-zinc-400 rounded">PRO</span>
            </div>
            <p class="text-xs text-zinc-500">Invite team members and manage roles. Available in the Pro edition.</p>
        </div>

        <div class="bg-zinc-900 border border-red-500/20 rounded-lg p-6 max-w-lg">
            <h2 class="text-sm font-medium text-red-400 mb-4">Danger Zone</h2>
            <p class="text-zinc-400 text-sm mb-4">
                Deleting this project will permanently remove all its data including subscribers, templates, and email history.
            </p>
            <p class="text-zinc-400 text-sm mb-4">
                Type <span class="font-semibold text-white">{{ project.name }}</span> to confirm.
            </p>
            <AppInput v-model="deleteConfirmName" placeholder="Type project name to confirm" class="mb-4" />
            <AppButton variant="danger" :disabled="deleteConfirmName !== project.name" :loading="deleteLoading"
                @click="handleDelete">
                {{ deleteLoading ? 'Deleting...' : 'Delete Project' }}
            </AppButton>
        </div>

        <AppModal :show="showKeyModal" title="Create API Key" @close="closeKeyModal">
            <div v-if="createdKey" class="space-y-4">
                <p class="text-sm text-zinc-400">
                    Your API key has been created. Copy it now — you won't be able to see it again.
                </p>
                <div class="flex items-center gap-2">
                    <code class="flex-1 px-3 py-2 bg-zinc-950 border border-zinc-800 rounded-lg text-sm text-white break-all select-all">
                        {{ createdKey }}
                    </code>
                    <button @click="copyKey"
                        class="px-3 py-2 text-sm bg-zinc-800 hover:bg-zinc-700 text-white rounded-lg transition cursor-pointer">
                        Copy
                    </button>
                </div>
                <AppButton @click="closeKeyModal">Done</AppButton>
            </div>

            <form v-else @submit.prevent="handleCreateKey" class="space-y-4">
                <AppInput v-model="newKeyName" label="Key Name" placeholder="Production API" required />
                <p class="text-xs text-zinc-500">Give your key a descriptive name so you can identify it later.</p>
                <AppButton :loading="keyLoading">
                    {{ keyLoading ? 'Creating...' : 'Create Key' }}
                </AppButton>
            </form>
        </AppModal>
    </div>
</template>

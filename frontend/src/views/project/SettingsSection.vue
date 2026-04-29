<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api/client'
import { useToastStore } from '@/stores/toast'
import { useAppStore } from '@/stores/app'
import type { Project } from '@/stores/projects'
import AppInput from '@/components/ui/AppInput.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppModal from '@/components/ui/AppModal.vue'
import AppConfirmModal from '@/components/ui/AppConfirmModal.vue'

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
const appStore = useAppStore()

const instanceUrl = computed(() => appStore.publicUrl || window.location.origin)
const isLocalhost = computed(() => instanceUrl.value.startsWith('http://localhost'))

async function copyInstanceUrl() {
    await navigator.clipboard.writeText(instanceUrl.value)
    toast.success('Instance URL copied')
}

const bounceWebhookPath = ref('')
const bounceToken = ref('')
const showRotateBounceConfirm = ref(false)
const rotateLoading = ref(false)

const bounceWebhookUrl = computed(() => bounceWebhookPath.value ? instanceUrl.value.replace(/\/$/, '') + bounceWebhookPath.value : '')

async function loadBounceWebhook() {
    try {
        const res = await api<{ project_id: string, bounce_token: string, path: string }>(
            `/projects/${props.project.id}/bounce-webhook`,
        )
        bounceToken.value = res.bounce_token
        bounceWebhookPath.value = res.path
    } catch {
        bounceWebhookPath.value = ''
    }
}

async function copyBounceWebhook() {
    if (!bounceWebhookUrl.value) return
    await navigator.clipboard.writeText(bounceWebhookUrl.value)
    toast.success('Webhook URL copied')
}

interface BounceIMAPConfig {
    enabled: boolean
    folder: string
    host?: string
    port?: number
    user?: string
    password_set: boolean
}

const imapConfig = ref<BounceIMAPConfig>({ enabled: false, folder: 'INBOX', password_set: false })
const imapForm = ref({ host: '', port: 993, user: '', password: '', folder: 'INBOX', enabled: false })
const imapLoading = ref(false)

async function loadBounceIMAP() {
    try {
        const res = await api<BounceIMAPConfig>(`/projects/${props.project.id}/bounce-imap`)
        imapConfig.value = res
        imapForm.value = {
            host: res.host ?? '',
            port: res.port ?? 993,
            user: res.user ?? '',
            password: '',
            folder: res.folder || 'INBOX',
            enabled: res.enabled,
        }
    } catch {
        /* noop */
    }
}

async function saveBounceIMAP() {
    imapLoading.value = true
    try {
        await api(`/projects/${props.project.id}/bounce-imap`, {
            method: 'PUT',
            body: imapForm.value,
        })
        toast.success('Bounce mailbox settings saved')
        imapForm.value.password = ''
        loadBounceIMAP()
    } catch (e: any) {
        toast.error(e.message || 'Failed to save IMAP settings')
    } finally {
        imapLoading.value = false
    }
}

async function rotateBounceToken() {
    rotateLoading.value = true
    try {
        const res = await api<{ project_id: string, bounce_token: string, path: string }>(
            `/projects/${props.project.id}/bounce-webhook/rotate`,
            { method: 'POST' },
        )
        bounceToken.value = res.bounce_token
        bounceWebhookPath.value = res.path
        toast.success('New bounce token issued')
        showRotateBounceConfirm.value = false
    } catch (e: any) {
        toast.error(e.message || 'Failed to rotate token')
    } finally {
        rotateLoading.value = false
    }
}

const projectName = ref('')
const projectDescription = ref('')
const generalLoading = ref(false)

const apiKeys = ref<APIKey[]>([])
const showKeyModal = ref(false)
const newKeyName = ref('')
const keyLoading = ref(false)
const createdKey = ref('')

const deleteConfirmName = ref('')
const deleteLoading = ref(false)

onMounted(() => {
    projectName.value = props.project.name
    projectDescription.value = props.project.description ?? ''
    fetchAPIKeys()
    loadBounceWebhook()
    loadBounceIMAP()
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
                            class="text-xs text-zinc-400 hover:text-white transition cursor-pointer">
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
            <h2 class="text-sm font-medium text-white mb-2">Instance URL</h2>
            <p class="text-xs text-zinc-500 mb-4">
                Public URL used to build unsubscribe links and tracking pixels in outgoing emails. Set <code class="text-zinc-400">PUBLIC_URL</code> in your environment to change it.
            </p>
            <div class="flex items-center gap-2">
                <code class="flex-1 px-3 py-2 bg-zinc-950 border border-zinc-800 rounded-lg text-xs text-white break-all">{{ instanceUrl }}</code>
                <button type="button" @click="copyInstanceUrl"
                    class="px-3 py-2 text-xs bg-zinc-800 hover:bg-zinc-700 text-white rounded-lg transition cursor-pointer">
                    Copy
                </button>
            </div>
            <p v-if="isLocalhost" class="text-xs text-yellow-400 mt-3">
                Heads up: this URL points to localhost. Unsubscribe links and tracking pixels in outgoing emails won't work outside this machine. Set <code>PUBLIC_URL</code> to your public domain in production.
            </p>
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

        <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-6 max-w-lg">
            <h2 class="text-sm font-medium text-white mb-2">Bounce ingestion webhook</h2>
            <p class="text-xs text-zinc-500 mb-4">
                Public endpoint your email provider can POST bounce notifications to. Each delivered payload adds the affected address to the suppression list with reason <code class="text-zinc-400">bounce</code>. Accepts a generic <code class="text-zinc-400">{ "email": "...", "reason": "..." }</code> JSON body or a Mailgun event payload.
            </p>
            <div class="flex items-center gap-2">
                <code class="flex-1 px-3 py-2 bg-zinc-950 border border-zinc-800 rounded-lg text-xs text-white break-all">{{ bounceWebhookUrl || 'Loading…' }}</code>
                <button type="button" @click="copyBounceWebhook" :disabled="!bounceWebhookUrl"
                    class="px-3 py-2 text-xs bg-zinc-800 hover:bg-zinc-700 text-white rounded-lg transition cursor-pointer disabled:opacity-50">
                    Copy
                </button>
            </div>
            <p class="text-xs text-zinc-500 mt-3">
                The token in the query string authenticates the call. If it leaks, rotate it — the new token immediately invalidates the old one.
            </p>
            <button type="button" @click="showRotateBounceConfirm = true"
                class="mt-3 px-3 py-1.5 text-xs text-red-400 border border-red-900/50 rounded-md hover:bg-red-950/40 transition cursor-pointer">
                Rotate token
            </button>
        </div>

        <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-6 max-w-lg">
            <h2 class="text-sm font-medium text-white mb-2">Bounce mailbox (IMAP)</h2>
            <p class="text-xs text-zinc-500 mb-4">
                Poll a mailbox where your SMTP relay forwards Delivery Status Notifications. Every 5 minutes SendDock fetches unseen messages, extracts hard-bounce recipients (RFC 3464 <code>Final-Recipient</code> first, then 5xx lines as fallback) and adds them to the suppression list.
            </p>
            <form @submit.prevent="saveBounceIMAP" class="space-y-3">
                <AppInput v-model="imapForm.host" label="Host" placeholder="imap.your-mailbox.com" />
                <div class="grid grid-cols-2 gap-3">
                    <div>
                        <label class="block text-sm font-medium text-zinc-300 mb-1">Port</label>
                        <input v-model.number="imapForm.port" type="number" min="1" max="65535"
                            class="w-full px-3 py-2 bg-zinc-900 border border-zinc-800 rounded-lg text-white placeholder-zinc-500 focus:outline-none focus:ring-2 focus:ring-zinc-500 transition" />
                    </div>
                    <div>
                        <label class="block text-sm font-medium text-zinc-300 mb-1">Folder</label>
                        <input v-model="imapForm.folder"
                            class="w-full px-3 py-2 bg-zinc-900 border border-zinc-800 rounded-lg text-white placeholder-zinc-500 focus:outline-none focus:ring-2 focus:ring-zinc-500 transition" />
                    </div>
                </div>
                <AppInput v-model="imapForm.user" label="Username" placeholder="bounces@your-domain.com" />
                <AppInput v-model="imapForm.password" type="password" label="Password"
                    :placeholder="imapConfig.password_set ? 'Leave empty to keep current' : 'Mailbox password'" />
                <label class="flex items-center gap-2 text-sm text-zinc-300">
                    <input type="checkbox" v-model="imapForm.enabled" class="accent-white" />
                    Enable polling
                </label>
                <AppButton type="submit" :loading="imapLoading" :disabled="imapLoading">
                    {{ imapLoading ? 'Saving...' : 'Save IMAP settings' }}
                </AppButton>
            </form>
        </div>

        <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-6 max-w-lg opacity-60">
            <div class="flex items-center gap-2 mb-2">
                <h2 class="text-sm font-medium text-white">Team Members</h2>
                <span class="text-xs px-2 py-0.5 bg-zinc-800 text-zinc-400 rounded">PRO</span>
            </div>
            <p class="text-xs text-zinc-500">Invite team members and manage roles. Available in the Pro edition.</p>
        </div>

        <AppConfirmModal
            :show="showRotateBounceConfirm"
            title="Rotate bounce webhook token"
            message="Generate a new token? The current URL stops working immediately, so update your provider configuration to point at the new URL right after this."
            confirmLabel="Rotate"
            danger
            :loading="rotateLoading"
            @confirm="rotateBounceToken"
            @cancel="showRotateBounceConfirm = false" />

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

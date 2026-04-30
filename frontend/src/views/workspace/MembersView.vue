<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useWorkspaceStore, type WorkspaceMember } from '@/stores/workspaces'
import { useToastStore } from '@/stores/toast'
import { ApiError } from '@/api/client'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import AppAlert from '@/components/ui/AppAlert.vue'
import AppLoader from '@/components/ui/AppLoader.vue'
import AppModal from '@/components/ui/AppModal.vue'
import AppConfirmModal from '@/components/ui/AppConfirmModal.vue'

const route = useRoute()
const router = useRouter()
const workspaceStore = useWorkspaceStore()
const toast = useToastStore()

const workspaceId = computed(() => route.params.id as string)
const members = ref<WorkspaceMember[]>([])
const loading = ref(true)

const showInvite = ref(false)
const inviteEmail = ref('')
const inviteRole = ref<'owner' | 'member'>('member')
const inviteError = ref('')
const inviteLoading = ref(false)

const showRemove = ref(false)
const memberToRemove = ref<WorkspaceMember | null>(null)
const removeLoading = ref(false)

const showRename = ref(false)
const renameValue = ref('')
const renameLoading = ref(false)
const renameError = ref('')

const workspace = computed(() =>
    workspaceStore.workspaces.find(w => w.id === workspaceId.value) || workspaceStore.active
)

const myRole = computed(() => {
    const ws = workspace.value
    if (!ws) return 'member'
    const me = members.value.find(m => m.user_id === ws.created_by)
    return me?.role || 'member'
})

const ownerCount = computed(() => members.value.filter(m => m.role === 'owner').length)

async function load() {
    loading.value = true
    try {
        members.value = await workspaceStore.listMembers(workspaceId.value)
    } catch (e: any) {
        toast.error(e.message || 'failed to load members')
    } finally {
        loading.value = false
    }
}

async function handleInvite() {
    inviteError.value = ''
    if (!inviteEmail.value.trim()) {
        inviteError.value = 'email is required'
        return
    }
    inviteLoading.value = true
    try {
        await workspaceStore.addMember(workspaceId.value, inviteEmail.value.trim(), inviteRole.value)
        showInvite.value = false
        inviteEmail.value = ''
        inviteRole.value = 'member'
        await load()
        toast.success('Member added')
    } catch (e: any) {
        if (e instanceof ApiError && e.status === 404) {
            inviteError.value = 'no SendDock account uses that email yet'
        } else {
            inviteError.value = e.message || 'failed to add member'
        }
    } finally {
        inviteLoading.value = false
    }
}

function openRemove(m: WorkspaceMember) {
    memberToRemove.value = m
    showRemove.value = true
}

async function confirmRemove() {
    if (!memberToRemove.value) return
    removeLoading.value = true
    try {
        await workspaceStore.removeMember(workspaceId.value, memberToRemove.value.user_id)
        showRemove.value = false
        memberToRemove.value = null
        await load()
        toast.success('Member removed')
    } catch (e: any) {
        toast.error(e.message || 'failed to remove member')
    } finally {
        removeLoading.value = false
    }
}

async function changeRole(m: WorkspaceMember, role: 'owner' | 'member') {
    if (m.role === role) return
    if (role === 'member' && m.role === 'owner' && ownerCount.value <= 1) {
        toast.error('a workspace needs at least one owner')
        return
    }
    try {
        await workspaceStore.updateMemberRole(workspaceId.value, m.user_id, role)
        await load()
        toast.success(`Role updated to ${role}`)
    } catch (e: any) {
        toast.error(e.message || 'failed to update role')
    }
}

function openRename() {
    renameValue.value = workspace.value?.name || ''
    renameError.value = ''
    showRename.value = true
}

async function confirmRename() {
    renameError.value = ''
    if (!renameValue.value.trim()) {
        renameError.value = 'name is required'
        return
    }
    renameLoading.value = true
    try {
        await workspaceStore.rename(workspaceId.value, renameValue.value.trim())
        showRename.value = false
        toast.success('Workspace renamed')
    } catch (e: any) {
        renameError.value = e.message || 'failed to rename'
    } finally {
        renameLoading.value = false
    }
}

function fmtDate(iso: string) {
    return new Date(iso).toLocaleDateString()
}

onMounted(async () => {
    if (workspaceStore.workspaces.length === 0) {
        await workspaceStore.fetch()
    }
    await load()
})
</script>

<template>
    <div class="min-h-screen bg-zinc-950 p-4 sm:p-6 md:p-8">
        <div class="max-w-3xl mx-auto">
            <button @click="router.push('/dashboard')"
                class="text-sm text-zinc-400 hover:text-white transition mb-6 inline-flex items-center gap-1 cursor-pointer">
                &larr; Dashboard
            </button>

            <AppLoader v-if="loading" message="Loading members..." />

            <template v-else>
                <div class="flex flex-wrap items-center justify-between gap-3 mb-2">
                    <div>
                        <h1 class="text-2xl font-bold text-white">{{ workspace?.name || 'Workspace' }}</h1>
                        <p class="text-sm text-zinc-500 mt-1">{{ members.length }} {{ members.length === 1 ? 'member' : 'members' }}</p>
                    </div>
                    <div class="flex items-center gap-2">
                        <AppButton variant="ghost" size="sm" v-if="myRole === 'owner'" @click="openRename">Rename</AppButton>
                        <AppButton size="sm" @click="showInvite = true" v-if="myRole === 'owner'">+ Add member</AppButton>
                    </div>
                </div>

                <div class="bg-zinc-900 border border-zinc-800 rounded-lg overflow-x-auto">
                    <table class="w-full min-w-[640px]">
                        <thead>
                            <tr class="border-b border-zinc-800">
                                <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Member</th>
                                <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Role</th>
                                <th class="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase tracking-wide">Joined</th>
                                <th class="px-4 py-3"></th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="m in members" :key="m.user_id" class="border-b border-zinc-800 last:border-0">
                                <td class="px-4 py-3">
                                    <div class="text-sm text-white font-medium">{{ m.name || m.email }}</div>
                                    <div class="text-xs text-zinc-500">{{ m.email }}</div>
                                </td>
                                <td class="px-4 py-3">
                                    <select v-if="myRole === 'owner'" :value="m.role"
                                        @change="changeRole(m, ($event.target as HTMLSelectElement).value as 'owner' | 'member')"
                                        class="px-2 py-1 text-xs bg-zinc-950 border border-zinc-800 rounded-md text-white cursor-pointer">
                                        <option value="owner">Owner</option>
                                        <option value="member">Member</option>
                                    </select>
                                    <span v-else class="text-xs text-zinc-300 capitalize">{{ m.role }}</span>
                                </td>
                                <td class="px-4 py-3 text-sm text-zinc-400">{{ fmtDate(m.joined_at) }}</td>
                                <td class="px-4 py-3 text-right">
                                    <button v-if="myRole === 'owner'" @click="openRemove(m)"
                                        class="text-xs text-zinc-500 hover:text-red-400 transition cursor-pointer">
                                        Remove
                                    </button>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </template>

            <AppModal :show="showInvite" title="Add member" @close="showInvite = false">
                <form @submit.prevent="handleInvite" class="space-y-4">
                    <p class="text-xs text-zinc-500">The user must already have a SendDock account on this instance.</p>
                    <AppInput v-model="inviteEmail" label="Email" placeholder="user@example.com" required />
                    <div>
                        <label class="block text-xs text-zinc-400 mb-1">Role</label>
                        <select v-model="inviteRole"
                            class="w-full px-3 py-2 text-sm bg-zinc-950 border border-zinc-800 rounded-lg text-white cursor-pointer">
                            <option value="member">Member</option>
                            <option value="owner">Owner</option>
                        </select>
                    </div>
                    <AppAlert :message="inviteError" />
                    <AppButton :loading="inviteLoading">
                        {{ inviteLoading ? 'Adding...' : 'Add member' }}
                    </AppButton>
                </form>
            </AppModal>

            <AppModal :show="showRename" title="Rename workspace" @close="showRename = false">
                <form @submit.prevent="confirmRename" class="space-y-4">
                    <AppInput v-model="renameValue" label="Workspace name" required />
                    <AppAlert :message="renameError" />
                    <AppButton :loading="renameLoading">
                        {{ renameLoading ? 'Saving...' : 'Save' }}
                    </AppButton>
                </form>
            </AppModal>

            <AppConfirmModal :show="showRemove" title="Remove member"
                :message="`Remove ${memberToRemove?.email} from this workspace? They will lose access to every project in it.`"
                confirm-label="Remove" danger :loading="removeLoading"
                @confirm="confirmRemove" @cancel="showRemove = false" />
        </div>
    </div>
</template>

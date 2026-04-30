<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import {
    useWorkspaceStore,
    type WorkspaceMember,
    type WorkspaceRole,
    ROLE_LABEL,
    ROLE_DESCRIPTION,
    ASSIGNABLE_ROLES,
} from '@/stores/workspaces'
import { useToastStore } from '@/stores/toast'
import { ApiError } from '@/api/client'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import AppAlert from '@/components/ui/AppAlert.vue'
import AppLoader from '@/components/ui/AppLoader.vue'
import AppModal from '@/components/ui/AppModal.vue'
import AppConfirmModal from '@/components/ui/AppConfirmModal.vue'
import AppProPaywall from '@/components/ui/AppProPaywall.vue'

const route = useRoute()
const router = useRouter()
const workspaceStore = useWorkspaceStore()
const toast = useToastStore()
const auth = useAuthStore()

const workspaceId = computed(() => route.params.id as string)
const members = ref<WorkspaceMember[]>([])
const loading = ref(true)
const paywall = ref(false)

const showInvite = ref(false)
const inviteMode = ref<'existing' | 'new'>('existing')
const inviteEmail = ref('')
const newUserName = ref('')
const newUserPassword = ref('')
const inviteRole = ref<WorkspaceRole>('developer')
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
    const me = members.value.find(m => m.user_id === auth.userId)
    return me?.role || 'viewer'
})

const isOwner = computed(() => myRole.value === 'owner')
const ownerCount = computed(() => members.value.filter(m => m.role === 'owner').length)

async function load() {
    loading.value = true
    paywall.value = false
    try {
        members.value = await workspaceStore.listMembers(workspaceId.value)
    } catch (e: any) {
        if (e instanceof ApiError && e.status === 402) {
            paywall.value = true
        } else {
            toast.error(e.message || 'failed to load members')
        }
    } finally {
        loading.value = false
    }
}

function openInvite(mode: 'existing' | 'new') {
    inviteMode.value = mode
    inviteEmail.value = ''
    newUserName.value = ''
    newUserPassword.value = ''
    inviteRole.value = 'developer'
    inviteError.value = ''
    showInvite.value = true
}

async function handleInvite() {
    inviteError.value = ''
    if (!inviteEmail.value.trim()) {
        inviteError.value = 'email is required'
        return
    }
    if (inviteMode.value === 'new') {
        if (!newUserName.value.trim()) {
            inviteError.value = 'name is required'
            return
        }
        if (newUserPassword.value.length < 8) {
            inviteError.value = 'password must be at least 8 characters'
            return
        }
    }
    inviteLoading.value = true
    try {
        if (inviteMode.value === 'existing') {
            await workspaceStore.addMember(workspaceId.value, inviteEmail.value.trim(), inviteRole.value)
        } else {
            await workspaceStore.createUser(workspaceId.value, {
                email: inviteEmail.value.trim(),
                name: newUserName.value.trim(),
                password: newUserPassword.value,
                role: inviteRole.value,
            })
        }
        showInvite.value = false
        await load()
        toast.success(inviteMode.value === 'new' ? 'User created and added' : 'Member added')
    } catch (e: any) {
        if (e instanceof ApiError && e.status === 402) {
            paywall.value = true
            showInvite.value = false
        } else if (e instanceof ApiError && e.status === 404) {
            inviteError.value = 'no SendDock account uses that email yet'
        } else if (e instanceof ApiError && e.status === 409) {
            inviteError.value = e.message
        } else {
            inviteError.value = e.message || 'failed'
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
        if (e instanceof ApiError && e.status === 402) paywall.value = true
        else toast.error(e.message || 'failed')
    } finally {
        removeLoading.value = false
    }
}

async function changeRole(m: WorkspaceMember, role: WorkspaceRole) {
    if (m.role === role) return
    if (role !== 'owner' && m.role === 'owner' && ownerCount.value <= 1) {
        toast.error('a workspace needs at least one owner')
        return
    }
    try {
        await workspaceStore.updateMemberRole(workspaceId.value, m.user_id, role)
        await load()
        toast.success('Role updated')
    } catch (e: any) {
        if (e instanceof ApiError && e.status === 402) paywall.value = true
        else toast.error(e.message || 'failed')
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
        renameError.value = e.message || 'failed'
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

            <AppProPaywall v-else-if="paywall"
                title="Team management is a Team plan feature"
                description="Adding members, creating user accounts and changing roles require a SendDock Team license. Without one you can still organize your projects across multiple workspaces — you just stay the only member."
                cta="See Team plan" />

            <template v-else>
                <div class="flex flex-wrap items-center justify-between gap-3 mb-2">
                    <div>
                        <h1 class="text-2xl font-bold text-white">{{ workspace?.name || 'Workspace' }}</h1>
                        <p class="text-sm text-zinc-500 mt-1">{{ members.length }} {{ members.length === 1 ? 'member' : 'members' }}</p>
                    </div>
                    <div class="flex items-center gap-2">
                        <AppButton variant="ghost" size="sm" v-if="isOwner" @click="openRename">Rename</AppButton>
                        <AppButton variant="ghost" size="sm" v-if="isOwner" @click="openInvite('existing')">+ Add existing</AppButton>
                        <AppButton size="sm" v-if="isOwner" @click="openInvite('new')">+ Create user</AppButton>
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
                                    <select v-if="isOwner" :value="m.role"
                                        @change="changeRole(m, ($event.target as HTMLSelectElement).value as WorkspaceRole)"
                                        class="px-2 py-1 text-xs bg-zinc-950 border border-zinc-800 rounded-md text-white cursor-pointer">
                                        <option v-for="r in ASSIGNABLE_ROLES" :key="r" :value="r">{{ ROLE_LABEL[r] }}</option>
                                    </select>
                                    <span v-else class="text-xs text-zinc-300">{{ ROLE_LABEL[m.role] }}</span>
                                </td>
                                <td class="px-4 py-3 text-sm text-zinc-400">{{ fmtDate(m.joined_at) }}</td>
                                <td class="px-4 py-3 text-right">
                                    <button v-if="isOwner && m.user_id !== auth.userId" @click="openRemove(m)"
                                        class="text-xs text-zinc-500 hover:text-red-400 transition cursor-pointer">
                                        Remove
                                    </button>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>

                <div class="mt-6 grid grid-cols-1 sm:grid-cols-2 gap-3">
                    <div v-for="r in ASSIGNABLE_ROLES" :key="r"
                        class="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                        <h3 class="text-sm font-semibold text-white mb-1">{{ ROLE_LABEL[r] }}</h3>
                        <p class="text-xs text-zinc-500 leading-relaxed">{{ ROLE_DESCRIPTION[r] }}</p>
                    </div>
                </div>
            </template>

            <AppModal :show="showInvite" :title="inviteMode === 'new' ? 'Create user' : 'Add existing user'" @close="showInvite = false">
                <form @submit.prevent="handleInvite" class="space-y-4">
                    <div class="flex bg-zinc-950 border border-zinc-800 rounded-lg p-1">
                        <button type="button" @click="inviteMode = 'existing'"
                            :class="['flex-1 px-3 py-1.5 text-xs rounded-md transition cursor-pointer', inviteMode === 'existing' ? 'bg-zinc-800 text-white' : 'text-zinc-400']">
                            Existing user
                        </button>
                        <button type="button" @click="inviteMode = 'new'"
                            :class="['flex-1 px-3 py-1.5 text-xs rounded-md transition cursor-pointer', inviteMode === 'new' ? 'bg-zinc-800 text-white' : 'text-zinc-400']">
                            New user
                        </button>
                    </div>
                    <p class="text-xs text-zinc-500">
                        <template v-if="inviteMode === 'existing'">The user must already have a SendDock account on this instance.</template>
                        <template v-else>Create a new user account and add them to this workspace. Pass the password to them out of band.</template>
                    </p>
                    <AppInput v-model="inviteEmail" label="Email" placeholder="user@example.com" required type="email" />
                    <template v-if="inviteMode === 'new'">
                        <AppInput v-model="newUserName" label="Name" placeholder="Jane Doe" required />
                        <div>
                            <label class="block text-xs text-zinc-400 mb-1">Temporary password</label>
                            <input v-model="newUserPassword" type="password" minlength="8" required
                                class="w-full px-3 py-2 text-sm bg-zinc-950 border border-zinc-800 rounded-lg text-white" />
                            <p class="text-xs text-zinc-500 mt-1">At least 8 characters. The user can change it after first login.</p>
                        </div>
                    </template>
                    <div>
                        <label class="block text-xs text-zinc-400 mb-1">Role</label>
                        <select v-model="inviteRole"
                            class="w-full px-3 py-2 text-sm bg-zinc-950 border border-zinc-800 rounded-lg text-white cursor-pointer">
                            <option v-for="r in ASSIGNABLE_ROLES" :key="r" :value="r">{{ ROLE_LABEL[r] }}</option>
                        </select>
                        <p class="text-xs text-zinc-500 mt-1">{{ ROLE_DESCRIPTION[inviteRole] }}</p>
                    </div>
                    <AppAlert :message="inviteError" />
                    <AppButton :loading="inviteLoading">
                        {{ inviteLoading ? 'Saving...' : (inviteMode === 'new' ? 'Create user' : 'Add member') }}
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

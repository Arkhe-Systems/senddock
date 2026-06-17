<script setup lang="ts">
import { useProjectStore, type Project } from '@/stores/projects';
import { useWorkspaceStore } from '@/stores/workspaces';
import { useRouter } from 'vue-router';
import AppButton from '@/components/ui/AppButton.vue';
import { onMounted, ref, computed, watch } from 'vue';
import AppModal from '@/components/ui/AppModal.vue';
import AppLoader from '@/components/ui/AppLoader.vue';
import AppInput from '@/components/ui/AppInput.vue';
import AppAlert from '@/components/ui/AppAlert.vue';
import UpdateBadge from '@/components/UpdateBadge.vue';
import UserProfilePanel from '@/components/UserProfilePanel.vue';
import UsageBanner from '@/components/UsageBanner.vue';
import { api } from '@/api/client';
import { useToastStore } from '@/stores/toast';
import { useAppStore } from '@/stores/app';

const toast = useToastStore()
const router = useRouter()
const projectStore = useProjectStore()
const workspaceStore = useWorkspaceStore()
const appStore = useAppStore()
const pageLoading = ref(true)
const mobileNavOpen = ref(false)

const showCreateModal = ref(false)
const createError = ref('')
const createLoading = ref(false)
const projectName = ref('')
const projectDescription = ref('')

const showWorkspaceMenu = ref(false)
const showCreateWorkspaceModal = ref(false)
const newWorkspaceName = ref('')
const newWorkspaceLoading = ref(false)
const newWorkspaceError = ref('')

const activeWorkspace = computed(() => workspaceStore.active)

async function handleCreateProject() {
    createError.value = ''
    if (!activeWorkspace.value) {
        createError.value = 'no workspace selected'
        return
    }
    createLoading.value = true
    try {
        await api('/projects', {
            method: 'POST',
            body: {
                workspace_id: activeWorkspace.value.id,
                name: projectName.value,
                description: projectDescription.value,
            },
        })
        showCreateModal.value = false
        projectName.value = ''
        projectDescription.value = ''
        await projectStore.fetchProjects(activeWorkspace.value.id)
        toast.success('Project created')
    } catch (e: any) {
        createError.value = e.message
    } finally {
        createLoading.value = false
    }
}

const showDeleteModal = ref(false)
const projectToDelete = ref<Project | null>(null)
const deleteConfirmName = ref('')
const deleteLoading = ref(false)

const canDelete = computed(() =>
    deleteConfirmName.value === projectToDelete.value?.name
)

function openDeleteModal(project: Project) {
    projectToDelete.value = project
    deleteConfirmName.value = ''
    showDeleteModal.value = true
}

async function handleDeleteProject() {
    if (!projectToDelete.value || !canDelete.value) return
    deleteLoading.value = true
    try {
        await projectStore.deleteProject(projectToDelete.value.id)
        await projectStore.fetchProjects(activeWorkspace.value?.id)
        showDeleteModal.value = false
        projectToDelete.value = null
        deleteConfirmName.value = ''
        toast.success('Project deleted')
    } finally {
        deleteLoading.value = false
    }
}

async function selectWorkspace(id: string) {
    workspaceStore.setActive(id)
    showWorkspaceMenu.value = false
    mobileNavOpen.value = false
    await projectStore.fetchProjects(id)
}

async function handleCreateWorkspace() {
    newWorkspaceError.value = ''
    if (!newWorkspaceName.value.trim()) {
        newWorkspaceError.value = 'name is required'
        return
    }
    newWorkspaceLoading.value = true
    try {
        const ws = await workspaceStore.create(newWorkspaceName.value.trim())
        showCreateWorkspaceModal.value = false
        newWorkspaceName.value = ''
        await projectStore.fetchProjects(ws.id)
        toast.success('Workspace created')
    } catch (e: any) {
        newWorkspaceError.value = e.message
    } finally {
        newWorkspaceLoading.value = false
    }
}

watch(() => activeWorkspace.value?.id, (id) => {
    if (id) projectStore.fetchProjects(id)
})

onMounted(async () => {
    await workspaceStore.fetch()
    if (activeWorkspace.value) {
        await projectStore.fetchProjects(activeWorkspace.value.id)
    }
    pageLoading.value = false
})
</script>

<template>
    <div class="min-h-screen bg-zinc-950">
        <AppLoader v-if="pageLoading" message="Loading projects..." fullscreen />

        <div v-else class="md:flex md:min-h-screen">
            <header class="md:hidden sticky top-0 z-30 bg-zinc-900 border-b border-zinc-800 px-4 py-3 flex items-center justify-between gap-3">
                <h1 class="text-base font-semibold text-white">SendDock</h1>
                <button type="button" @click="mobileNavOpen = !mobileNavOpen"
                    class="shrink-0 inline-flex items-center justify-center w-9 h-9 rounded-lg border border-zinc-800 text-zinc-300 hover:text-white hover:bg-zinc-800 transition cursor-pointer"
                    :aria-expanded="mobileNavOpen" aria-label="Toggle navigation">
                    <svg v-if="!mobileNavOpen" xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h16M4 18h16" />
                    </svg>
                    <svg v-else xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M6 6l12 12M18 6L6 18" />
                    </svg>
                </button>
            </header>

            <div v-if="mobileNavOpen" class="md:hidden fixed inset-0 z-20 bg-black/60" @click="mobileNavOpen = false"></div>

            <aside :class="[
                'bg-zinc-900 border-zinc-800 flex flex-col',
                'md:w-64 md:shrink-0 md:border-r md:p-4 md:block md:sticky md:top-0 md:h-screen md:overflow-y-auto',
                mobileNavOpen
                    ? 'fixed top-[57px] left-0 right-0 bottom-0 z-30 p-4 border-t overflow-y-auto'
                    : 'hidden'
            ]">
                <div class="hidden md:flex items-center justify-between mb-6">
                    <h1 class="text-lg font-semibold text-white">SendDock</h1>
                    <UpdateBadge v-if="appStore.deploymentMode !== 'cloud'" />
                </div>

                <div class="mb-4">
                    <p class="text-xs uppercase tracking-wide text-zinc-500 mb-2 px-1">Workspace</p>
                    <div class="relative">
                        <button type="button" @click="showWorkspaceMenu = !showWorkspaceMenu"
                            class="w-full flex items-center justify-between gap-2 px-3 py-2 text-sm bg-zinc-950 border border-zinc-800 rounded-lg text-white hover:border-zinc-700 transition cursor-pointer">
                            <span class="truncate font-medium">{{ activeWorkspace?.name || 'None' }}</span>
                            <svg class="w-4 h-4 text-zinc-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                                <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
                            </svg>
                        </button>
                        <div v-if="showWorkspaceMenu"
                            class="absolute z-20 mt-2 left-0 right-0 bg-zinc-900 border border-zinc-800 rounded-lg shadow-lg py-1">
                            <button v-for="ws in workspaceStore.workspaces" :key="ws.id"
                                @click="selectWorkspace(ws.id)"
                                :class="[
                                    'w-full text-left px-3 py-2 text-sm flex items-center justify-between hover:bg-zinc-800 transition cursor-pointer',
                                    ws.id === activeWorkspace?.id ? 'text-white' : 'text-zinc-400'
                                ]">
                                <span class="truncate">{{ ws.name }}</span>
                                <span v-if="ws.id === activeWorkspace?.id" class="text-xs text-zinc-500">active</span>
                            </button>
                            <div class="border-t border-zinc-800 my-1"></div>
                            <button v-if="activeWorkspace" @click="router.push(`/workspaces/${activeWorkspace.id}/members`); showWorkspaceMenu = false; mobileNavOpen = false"
                                class="w-full text-left px-3 py-2 text-sm text-zinc-300 hover:bg-zinc-800 transition cursor-pointer">
                                Manage members
                            </button>
                            <button @click="showCreateWorkspaceModal = true; showWorkspaceMenu = false"
                                class="w-full text-left px-3 py-2 text-sm text-zinc-300 hover:bg-zinc-800 transition cursor-pointer">
                                + New workspace
                            </button>
                        </div>
                    </div>
                </div>

                <nav class="space-y-1">
                    <div class="px-3 py-2 text-sm rounded-lg bg-zinc-800 text-white">
                        Projects
                    </div>
                </nav>

                <div class="mt-auto">
                    <UserProfilePanel />
                </div>
            </aside>

            <main class="flex-1 min-w-0 p-4 sm:p-6 md:p-8">
                <div class="max-w-5xl mx-auto">
                    <UsageBanner class="mb-6" />
                    <div class="flex flex-wrap items-center justify-between gap-3 mb-6">
                        <div>
                            <h2 class="text-2xl font-bold text-white">Projects</h2>
                            <p v-if="activeWorkspace" class="text-sm text-zinc-500 mt-1">
                                In workspace <span class="text-zinc-300">{{ activeWorkspace.name }}</span>
                            </p>
                        </div>
                        <button @click="showCreateModal = true" :disabled="!activeWorkspace"
                            class="px-4 py-2 text-sm font-medium bg-white text-zinc-950 rounded-lg hover:bg-zinc-200 transition cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed">
                            + New Project
                        </button>
                    </div>

                    <div v-if="projectStore.projects.length > 0" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                        <RouterLink v-for="project in projectStore.projects" :key="project.id"
                            :to="`/projects/${project.id}`"
                            class="block bg-zinc-900 border border-zinc-800 rounded-lg p-5 hover:border-zinc-600 hover:bg-zinc-800/50 transition cursor-pointer group">
                            <h3 class="text-white font-semibold group-hover:text-white">{{ project.name }}</h3>
                            <p v-if="project.description" class="text-sm text-zinc-500 mt-1 line-clamp-2">{{ project.description }}</p>
                            <div class="flex items-center justify-between mt-4 pt-3 border-t border-zinc-800">
                                <span class="text-xs text-zinc-600">
                                    {{ new Date(project.created_at).toLocaleDateString() }}
                                </span>
                                <button @click.prevent.stop="openDeleteModal(project)"
                                    class="text-xs text-zinc-600 hover:text-red-400 transition cursor-pointer opacity-0 group-hover:opacity-100">
                                    Delete
                                </button>
                            </div>
                        </RouterLink>
                    </div>

                    <div v-else class="text-center py-20 border border-dashed border-zinc-800 rounded-lg">
                        <p class="text-zinc-500 mb-4">No projects in this workspace yet.</p>
                        <button @click="showCreateModal = true" :disabled="!activeWorkspace"
                            class="px-6 py-2 text-sm font-medium bg-white text-zinc-950 rounded-lg hover:bg-zinc-200 transition cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed">
                            Create Project
                        </button>
                    </div>
                </div>
            </main>
        </div>

        <AppModal :show="showCreateModal" title="New Project" @close="showCreateModal = false">
            <form @submit.prevent="handleCreateProject" class="space-y-4">
                <p class="text-xs text-zinc-500">In workspace <span class="text-white font-medium">{{ activeWorkspace?.name }}</span></p>
                <AppInput v-model="projectName" label="Project Name" placeholder="My awesome project" required />
                <AppInput v-model="projectDescription" large label="Description" placeholder="What is this project about?" />
                <AppAlert :message="createError" />
                <AppButton :loading="createLoading">
                    {{ createLoading ? 'Creating...' : 'Create Project' }}
                </AppButton>
            </form>
        </AppModal>

        <AppModal :show="showCreateWorkspaceModal" title="New Workspace" @close="showCreateWorkspaceModal = false">
            <form @submit.prevent="handleCreateWorkspace" class="space-y-4">
                <AppInput v-model="newWorkspaceName" label="Workspace name" placeholder="Acme Marketing" required />
                <AppAlert :message="newWorkspaceError" />
                <AppButton :loading="newWorkspaceLoading">
                    {{ newWorkspaceLoading ? 'Creating...' : 'Create Workspace' }}
                </AppButton>
            </form>
        </AppModal>

        <AppModal :show="showDeleteModal" title="Delete Project" @close="showDeleteModal = false">
            <div class="space-y-4">
                <p class="text-zinc-400 text-sm">
                    This action cannot be undone. Type
                    <span class="font-semibold text-white">{{ projectToDelete?.name }}</span>
                    to confirm.
                </p>
                <AppInput v-model="deleteConfirmName" placeholder="Type project name to confirm" />
                <AppButton variant="danger" :disabled="!canDelete" :loading="deleteLoading"
                    @click="handleDeleteProject">
                    {{ deleteLoading ? 'Deleting...' : 'Delete Project' }}
                </AppButton>
            </div>
        </AppModal>
    </div>
</template>

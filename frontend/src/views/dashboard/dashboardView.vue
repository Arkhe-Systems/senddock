<script setup lang="ts">
import { useProjectStore, type Project } from '@/stores/projects';
import { useAuthStore } from '@/stores/auth';
import { useRouter } from 'vue-router';
import AppButton from '@/components/ui/AppButton.vue';
import { onMounted, ref, computed } from 'vue';
import AppModal from '@/components/ui/AppModal.vue';
import AppLoader from '@/components/ui/AppLoader.vue';
import AppInput from '@/components/ui/AppInput.vue';
import AppAlert from '@/components/ui/AppAlert.vue';
import { api } from '@/api/client';
import { useToastStore } from '@/stores/toast';

const auth = useAuthStore()
const toast = useToastStore()
const router = useRouter()
const projectStore = useProjectStore()
const pageLoading = ref(true)

// Create project
const showCreateModal = ref(false)
const createError = ref('')
const createLoading = ref(false)
const projectName = ref('')
const projectDescription = ref('')

async function handleCreateProject() {
    createError.value = ''
    createLoading.value = true
    try {
        await api('/projects', {
            method: 'POST',
            body: {
                name: projectName.value,
                description: projectDescription.value
            },
        })
        showCreateModal.value = false
        projectName.value = ''
        projectDescription.value = ''
        projectStore.fetchProjects()
        toast.success('Project created')
    } catch (e: any) {
        createError.value = e.message
    } finally {
        createLoading.value = false
    }
}

// Delete project
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
        showDeleteModal.value = false
        projectToDelete.value = null
        deleteConfirmName.value = ''
        toast.success('Project deleted')
    } finally {
        deleteLoading.value = false
    }
}

// Logout
async function handleLogout() {
    await auth.logout()
    router.push('/login')
}

onMounted(async () => {
    await projectStore.fetchProjects()
    pageLoading.value = false
})
</script>

<template>
    <div class="min-h-screen bg-zinc-950 p-8">
        <div class="max-w-5xl mx-auto">
            <div class="flex items-center justify-between mb-10">
                <h1 class="text-2xl font-bold text-white">SendDock</h1>
                <button @click="handleLogout"
                    class="px-4 py-2 text-sm text-zinc-400 hover:text-white border border-zinc-800 rounded-lg transition cursor-pointer">
                    Logout
                </button>
            </div>

            <AppLoader v-if="pageLoading" message="Loading projects..." />

            <template v-else>
            <div class="flex items-center justify-between mb-6">
                <h2 class="text-xl font-semibold text-white">Your Projects</h2>
                <button @click="showCreateModal = true"
                    class="px-4 py-2 text-sm font-medium bg-white text-zinc-950 rounded-lg hover:bg-zinc-200 transition cursor-pointer">
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
                <p class="text-zinc-500 mb-4">No projects yet. Create your first one.</p>
                <button @click="showCreateModal = true"
                    class="px-6 py-2 text-sm font-medium bg-white text-zinc-950 rounded-lg hover:bg-zinc-200 transition cursor-pointer">
                    Create Project
                </button>
            </div>

            </template>

            <AppModal :show="showCreateModal" title="New Project" @close="showCreateModal = false">
                <form @submit.prevent="handleCreateProject" class="space-y-4">
                    <AppInput v-model="projectName" label="Project Name" placeholder="My awesome project" required />
                    <AppInput v-model="projectDescription" large label="Description" placeholder="What is this project about?" />
                    <AppAlert :message="createError" />
                    <AppButton :loading="createLoading">
                        {{ createLoading ? 'Creating...' : 'Create Project' }}
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
    </div>
</template>

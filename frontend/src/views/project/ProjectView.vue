<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { onMounted, ref, computed } from 'vue'
import { api } from '@/api/client'
import type { Project } from '@/stores/projects'
import AppLoader from '@/components/ui/AppLoader.vue'

const route = useRoute()
const router = useRouter()
const project = ref<Project | null>(null)
const loading = ref(true)

const projectId = computed(() => route.params.id as string)

async function loadProject() {
    try {
        project.value = await api<Project>(`/projects/${projectId.value}`)
    } catch {
        router.push('/dashboard')
    } finally {
        loading.value = false
    }
}

const navItems = [
    { name: 'Overview', route: 'project-overview' },
    { name: 'Subscribers', route: 'project-subscribers' },
    { name: 'Templates', route: 'project-templates' },
    { name: 'SMTP Settings', route: 'project-smtp' },
    { name: 'Settings', route: 'project-settings' },
]

const proItems = [
    { name: 'Analytics' },
    { name: 'Webhooks' },
]

onMounted(loadProject)
</script>

<template>
    <div class="min-h-screen bg-zinc-950">
        <AppLoader v-if="loading" message="Loading project..." fullscreen />

        <div v-else-if="project" class="flex min-h-screen">
            <aside class="w-64 bg-zinc-900 border-r border-zinc-800 p-4 flex flex-col">
                <RouterLink to="/dashboard" class="text-sm text-zinc-400 hover:text-white transition mb-6 inline-flex items-center gap-1">
                    &larr; Projects
                </RouterLink>

                <h2 class="text-lg font-semibold text-white mb-1">{{ project.name }}</h2>
                <p v-if="project.description" class="text-xs text-zinc-500 mb-6">{{ project.description }}</p>
                <div v-else class="mb-6"></div>

                <nav class="space-y-1 flex-1">
                    <RouterLink v-for="item in navItems" :key="item.route"
                        :to="{ name: item.route, params: { id: projectId } }"
                        :class="[
                            'block px-3 py-2 text-sm rounded-lg transition',
                            route.name === item.route
                                ? 'bg-zinc-800 text-white'
                                : 'text-zinc-400 hover:text-white hover:bg-zinc-800'
                        ]">
                        {{ item.name }}
                    </RouterLink>

                    <div class="pt-4 mt-4 border-t border-zinc-800">
                        <p class="px-3 text-xs text-zinc-500 uppercase tracking-wide mb-2">Pro</p>
                        <span v-for="item in proItems" :key="item.name"
                            class="block px-3 py-2 text-sm rounded-lg text-zinc-500 cursor-not-allowed">
                            {{ item.name }}
                        </span>
                    </div>
                </nav>
            </aside>

            <main class="flex-1 p-8">
                <RouterView :project="project" @updated="loadProject" />
            </main>
        </div>
    </div>
</template>

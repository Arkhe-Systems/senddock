<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { onMounted, ref, computed, watch } from 'vue'
import { api } from '@/api/client'
import type { Project } from '@/stores/projects'
import AppLoader from '@/components/ui/AppLoader.vue'
import UserProfilePanel from '@/components/UserProfilePanel.vue'
import AppNavLink from '@/components/ui/AppNavLink.vue'
import type { Component } from 'vue'
import {
    LayoutDashboard, Users, Filter, Ban, LayoutTemplate, ScrollText,
    Server, Megaphone, Newspaper, Send, BarChart3, Webhook, Settings, ShieldCheck, Table2,
} from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const project = ref<Project | null>(null)
const loading = ref(true)
const mobileNavOpen = ref(false)

const projectId = computed(() => route.params.id as string)

watch(() => route.fullPath, () => {
    mobileNavOpen.value = false
})

async function loadProject() {
    try {
        project.value = await api<Project>(`/projects/${projectId.value}`)
    } catch {
        router.push('/dashboard')
    } finally {
        loading.value = false
    }
}

const navGroups: { label: string; items: { name: string; route: string; icon: Component }[] }[] = [
    {
        label: 'Audience',
        items: [
            { name: 'Subscribers', route: 'project-subscribers', icon: Users },
            { name: 'Segments', route: 'project-segments', icon: Filter },
            { name: 'Newsletters', route: 'project-newsletters', icon: Newspaper },
            { name: 'Suppressions', route: 'project-suppressions', icon: Ban },
        ],
    },
    {
        label: 'Content & sending',
        items: [
            { name: 'Templates', route: 'project-templates', icon: LayoutTemplate },
            { name: 'Campaigns', route: 'project-campaigns', icon: Megaphone },
            { name: 'Broadcasts', route: 'project-broadcasts', icon: Send },
        ],
    },
    {
        label: 'Monitoring',
        items: [
            { name: 'Analytics', route: 'project-analytics', icon: BarChart3 },
            { name: 'Logs', route: 'project-logs', icon: ScrollText },
        ],
    },
    {
        label: 'Configuration',
        items: [
            { name: 'SMTP Settings', route: 'project-smtp', icon: Server },
            { name: 'Webhooks', route: 'project-webhooks', icon: Webhook },
            { name: 'Settings', route: 'project-settings', icon: Settings },
        ],
    },
]

const overviewItem = { name: 'Overview', route: 'project-overview', icon: LayoutDashboard }

const proItems: { name: string; route: string; icon: Component }[] = [
    { name: 'Reports', route: 'project-reports', icon: Table2 },
    { name: 'Audit log', route: 'project-audit-log', icon: ShieldCheck },
]

onMounted(loadProject)
</script>

<template>
    <div class="min-h-screen bg-zinc-950">
        <AppLoader v-if="loading" message="Loading project..." fullscreen />

        <div v-else-if="project" class="md:flex md:min-h-screen">
            <header class="md:hidden sticky top-0 z-30 bg-zinc-900 border-b border-zinc-800 px-4 py-3 flex items-center justify-between gap-3">
                <div class="min-w-0 flex-1">
                    <RouterLink to="/dashboard" class="text-xs text-zinc-300 hover:text-white transition inline-flex items-center gap-1">
                        &larr; Projects
                    </RouterLink>
                    <h2 class="text-base font-semibold text-white truncate">{{ project.name }}</h2>
                </div>
                <button type="button" @click="mobileNavOpen = !mobileNavOpen"
                    class="shrink-0 inline-flex items-center justify-center w-9 h-9 rounded-lg border border-zinc-800 text-zinc-300 hover:text-white hover:bg-zinc-850 transition cursor-pointer"
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
                <RouterLink to="/dashboard" class="hidden md:inline-flex text-sm text-zinc-300 hover:text-white transition mb-6 items-center gap-1">
                    &larr; Projects
                </RouterLink>

                <h2 class="hidden md:block text-lg font-semibold text-white mb-1">{{ project.name }}</h2>
                <p v-if="project.description" class="hidden md:block text-xs text-zinc-400 mb-6">{{ project.description }}</p>
                <div v-else class="hidden md:block mb-6"></div>

                <nav class="space-y-1 flex-1">
                    <AppNavLink :to="{ name: overviewItem.route, params: { id: projectId } }">
                        <component :is="overviewItem.icon" class="w-4 h-4 shrink-0" />
                        <span>{{ overviewItem.name }}</span>
                    </AppNavLink>

                    <div v-for="group in navGroups" :key="group.label" class="pt-4 mt-4 border-t border-zinc-800">
                        <p class="px-3 text-xs text-zinc-300 uppercase tracking-wide mb-2">{{ group.label }}</p>
                        <AppNavLink v-for="item in group.items" :key="item.route"
                            :to="{ name: item.route, params: { id: projectId } }">
                            <component :is="item.icon" class="w-4 h-4 shrink-0" />
                            <span>{{ item.name }}</span>
                        </AppNavLink>
                    </div>

                    <div class="pt-4 mt-4 border-t border-zinc-800">
                        <p class="px-3 text-xs text-zinc-300 uppercase tracking-wide mb-2">Pro</p>
                        <template v-for="item in proItems" :key="item.name">
                            <AppNavLink v-if="item.route"
                                :to="{ name: item.route, params: { id: projectId } }">
                                <span class="flex items-center gap-2.5 flex-1 min-w-0">
                                    <component :is="item.icon" class="w-4 h-4 shrink-0" />
                                    {{ item.name }}
                                </span>
                                <span class="text-[10px] font-semibold tracking-wider uppercase px-1.5 py-0.5 rounded bg-amber-500/15 text-amber-400 border border-amber-500/30">Pro</span>
                            </AppNavLink>
                            <span v-else
                                class="flex items-center justify-between px-3 py-2 text-sm rounded-lg text-zinc-400 cursor-not-allowed">
                                <span>{{ item.name }}</span>
                                <span class="text-[10px] tracking-wider uppercase text-zinc-600">soon</span>
                            </span>
                        </template>
                    </div>
                </nav>

                <UserProfilePanel />
            </aside>

            <main class="flex-1 min-w-0 p-4 sm:p-6 md:p-8">
                <RouterView :project="project" @updated="loadProject" />
            </main>
        </div>
    </div>
</template>

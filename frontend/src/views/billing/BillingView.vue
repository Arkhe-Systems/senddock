<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useLicenseStore } from '@/stores/license'
import AppProPaywall from '@/components/ui/AppProPaywall.vue'
import UserProfilePanel from '@/components/UserProfilePanel.vue'

const auth = useAuthStore()
const licenseStore = useLicenseStore()
const mobileNavOpen = ref(false)

const tier = computed(() => licenseStore.tier)
const status = computed(() => licenseStore.status)

const tierLabel = computed(() => {
    if (tier.value === 'team') return 'Team'
    if (tier.value === 'pro') return 'Pro'
    return 'Free'
})

const tierBadgeClass = computed(() => {
    if (tier.value === 'team') return 'bg-purple-500/15 text-purple-400 border-purple-500/30'
    if (tier.value === 'pro') return 'bg-amber-500/15 text-amber-400 border-amber-500/30'
    return 'bg-zinc-700/30 text-zinc-300 border-zinc-700'
})

const expiresLabel = computed(() => {
    if (!status.value?.expires_at) return null
    return new Date(status.value.expires_at).toLocaleString()
})

const checkedLabel = computed(() => {
    if (!status.value?.checked_at) return null
    return new Date(status.value.checked_at).toLocaleString()
})

const hasLicenseProblem = computed(() => {
    return status.value?.has_license && tier.value === 'free' && status.value?.reason
})

onMounted(async () => {
    if (!auth.email) await auth.checkAuth()
    await licenseStore.fetch(true)
})
</script>

<template>
    <div class="min-h-screen bg-zinc-950">
        <div class="md:flex md:min-h-screen">
            <header class="md:hidden sticky top-0 z-30 bg-zinc-900 border-b border-zinc-800 px-4 py-3 flex items-center justify-between gap-3">
                <h1 class="text-base font-semibold text-white">Billing</h1>
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
                <RouterLink to="/dashboard" class="hidden md:inline-flex text-sm text-zinc-400 hover:text-white transition mb-6 items-center gap-1">
                    &larr; Projects
                </RouterLink>

                <nav class="space-y-1 flex-1">
                    <RouterLink to="/account" class="block px-3 py-2 text-sm rounded-lg text-zinc-400 hover:text-white hover:bg-zinc-800 transition">
                        Account
                    </RouterLink>
                    <RouterLink to="/billing" class="block px-3 py-2 text-sm rounded-lg bg-zinc-800 text-white">
                        Billing
                    </RouterLink>
                </nav>

                <UserProfilePanel />
            </aside>

            <main class="flex-1 min-w-0 p-4 sm:p-6 md:p-8">
                <div class="max-w-2xl mx-auto space-y-6">
                    <div>
                        <h2 class="text-2xl font-bold text-white">Billing</h2>
                        <p class="text-sm text-zinc-500 mt-1">Your SendDock license and plan.</p>
                    </div>

                    <section class="bg-zinc-900 border border-zinc-800 rounded-xl p-5">
                        <div class="flex items-start justify-between gap-4 mb-4">
                            <div>
                                <p class="text-xs uppercase tracking-wide text-zinc-500 mb-1">Current plan</p>
                                <p class="text-2xl font-bold text-white">{{ tierLabel }}</p>
                            </div>
                            <span :class="['text-[11px] uppercase tracking-wider px-2 py-1 rounded border whitespace-nowrap', tierBadgeClass]">
                                {{ tierLabel }}
                            </span>
                        </div>

                        <dl v-if="tier !== 'free'" class="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-3 text-sm border-t border-zinc-800 pt-4">
                            <div v-if="expiresLabel">
                                <dt class="text-xs uppercase tracking-wide text-zinc-500 mb-1">Expires</dt>
                                <dd class="text-white">{{ expiresLabel }}</dd>
                            </div>
                            <div v-if="checkedLabel">
                                <dt class="text-xs uppercase tracking-wide text-zinc-500 mb-1">Last checked</dt>
                                <dd class="text-white">{{ checkedLabel }}</dd>
                            </div>
                        </dl>

                        <div v-if="hasLicenseProblem" class="mt-4 p-3 bg-red-500/10 border border-red-500/30 rounded-lg text-xs text-red-300">
                            License validation issue: {{ status?.reason }}
                        </div>

                        <p v-if="tier !== 'free'" class="text-xs text-zinc-500 mt-4">
                            Your license key is set via the <code class="text-zinc-400">SENDDOCK_LICENSE_KEY</code> environment variable on the server. To change it, update the env var and restart SendDock.
                        </p>
                    </section>

                    <div v-if="tier === 'free'" class="space-y-4">
                        <AppProPaywall
                            tier="pro"
                            title="Unlock SendDock Pro"
                            description="Analytics dashboard, webhooks, audit log, custom tracking domain, advanced deliverability checks, and more." />
                        <AppProPaywall
                            tier="team"
                            title="SendDock Team"
                            description="Everything in Pro plus workspaces, member roles, approval workflows, A/B testing, segmentation and broadcast cancellation." />
                    </div>

                    <p class="text-xs text-zinc-500 text-center">
                        SendDock is Bring Your Own SMTP — you only pay for SendDock features, not for emails sent.
                    </p>
                </div>
            </main>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useLicenseStore } from '@/stores/license'
import { useAppStore } from '@/stores/app'
import { useUsageStore } from '@/stores/usage'
import { api } from '@/api/client'
import AppProPaywall from '@/components/ui/AppProPaywall.vue'
import UserProfilePanel from '@/components/UserProfilePanel.vue'

const auth = useAuthStore()
const licenseStore = useLicenseStore()
const appStore = useAppStore()
const usageStore = useUsageStore()
const mobileNavOpen = ref(false)

const isCloud = computed(() => appStore.deploymentMode === 'cloud')
const tier = computed(() => licenseStore.tier)
const status = computed(() => licenseStore.status)

const cloudPlan = computed(() => usageStore.usage?.plan || 'free')

const planLabels: Record<string, string> = { free: 'Free', starter: 'Starter', growth: 'Growth', scale: 'Scale' }
const selfHostedLabel = computed(() => (tier.value === 'team' ? 'Team' : tier.value === 'pro' ? 'Pro' : 'Free'))
const displayPlan = computed(() => (isCloud.value ? planLabels[cloudPlan.value] || 'Free' : selfHostedLabel.value))

const tierBadgeClass = computed(() => {
    const p = isCloud.value ? cloudPlan.value : tier.value
    if (p === 'scale' || p === 'team') return 'bg-purple-500/15 text-purple-400 border-purple-500/30'
    if (p === 'growth') return 'bg-blue-500/15 text-blue-400 border-blue-500/30'
    if (p === 'starter' || p === 'pro') return 'bg-amber-500/15 text-amber-400 border-amber-500/30'
    return 'bg-zinc-700/30 text-zinc-300 border-zinc-700'
})

const cloudPlans = [
    { tier: 'starter', name: 'Starter', price: 19, subs: 'Up to 10,000 subscribers', features: ['Pro Analytics', 'Webhooks UI', 'Audit log', '90-day event history'] },
    { tier: 'growth', name: 'Growth', price: 49, subs: 'Up to 50,000 subscribers', features: ['Multi-user & roles', '1-year event history', 'Priority email support'] },
    { tier: 'scale', name: 'Scale', price: 129, subs: 'Up to 250,000 subscribers', features: ['Highest limits', 'All features', 'Priority support'] },
]

const expiresLabel = computed(() => {
    if (!status.value?.expires_at) return null
    return new Date(status.value.expires_at).toLocaleString()
})
const checkedLabel = computed(() => {
    if (!status.value?.checked_at) return null
    return new Date(status.value.checked_at).toLocaleString()
})
const hasLicenseProblem = computed(() => status.value?.has_license && tier.value === 'free' && status.value?.reason)

const checkoutLoading = ref('')
const portalLoading = ref(false)
const billingError = ref('')

async function upgrade(planTier: string) {
    billingError.value = ''
    checkoutLoading.value = planTier
    try {
        const res = await api<{ url: string }>('/billing/checkout/' + planTier)
        window.location.href = res.url
    } catch (e: any) {
        billingError.value = e.message || 'Could not start checkout'
        checkoutLoading.value = ''
    }
}

async function manageSubscription() {
    billingError.value = ''
    portalLoading.value = true
    try {
        const res = await api<{ url: string }>('/billing/portal')
        window.location.href = res.url
    } catch (e: any) {
        billingError.value = 'No active subscription to manage yet.'
        portalLoading.value = false
    }
}

onMounted(async () => {
    if (!auth.email) await auth.checkAuth()
    await licenseStore.fetch(true)
    if (isCloud.value) await usageStore.fetch()
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
                        <p class="text-sm text-zinc-500 mt-1">Your SendDock plan.</p>
                    </div>

                    <p v-if="billingError" class="rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-200">
                        {{ billingError }}
                    </p>

                    <section class="bg-zinc-900 border border-zinc-800 rounded-xl p-5">
                        <div class="flex items-start justify-between gap-4 mb-4">
                            <div>
                                <p class="text-xs uppercase tracking-wide text-zinc-500 mb-1">Current plan</p>
                                <p class="text-2xl font-bold text-white">{{ displayPlan }}</p>
                            </div>
                            <span :class="['text-[11px] uppercase tracking-wider px-2 py-1 rounded border whitespace-nowrap', tierBadgeClass]">
                                {{ displayPlan }}
                            </span>
                        </div>

                        <button v-if="isCloud && cloudPlan !== 'free'" @click="manageSubscription" :disabled="portalLoading"
                            class="px-4 py-2 text-sm font-medium border border-zinc-700 text-white rounded-lg hover:bg-zinc-800 transition cursor-pointer disabled:opacity-50">
                            {{ portalLoading ? 'Opening…' : 'Manage subscription' }}
                        </button>

                        <dl v-if="!isCloud && tier !== 'free'" class="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-3 text-sm border-t border-zinc-800 pt-4">
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

                        <p v-if="!isCloud && tier !== 'free'" class="text-xs text-zinc-500 mt-4">
                            Your license key is set via the <code class="text-zinc-400">SENDDOCK_LICENSE_KEY</code> environment variable on the server. To change it, update the env var and restart SendDock.
                        </p>
                    </section>

                    <div v-if="isCloud" class="grid grid-cols-1 sm:grid-cols-3 gap-4">
                        <div v-for="p in cloudPlans" :key="p.tier"
                            :class="['rounded-xl border p-5 flex flex-col', p.tier === cloudPlan ? 'border-white/40 bg-zinc-900' : 'border-zinc-800 bg-zinc-900']">
                            <p class="text-sm font-semibold text-white">{{ p.name }}</p>
                            <p class="mt-1"><span class="text-2xl font-bold text-white">${{ p.price }}</span><span class="text-sm text-zinc-500">/mo</span></p>
                            <p class="text-xs text-zinc-500 mt-1">{{ p.subs }}</p>
                            <ul class="mt-4 space-y-1.5 flex-1">
                                <li v-for="f in p.features" :key="f" class="text-xs text-zinc-400 flex items-start gap-1.5">
                                    <span class="text-emerald-400">✓</span> {{ f }}
                                </li>
                            </ul>
                            <button v-if="p.tier === cloudPlan" disabled
                                class="mt-5 px-4 py-2 text-sm font-medium rounded-lg bg-zinc-800 text-zinc-400 cursor-default">
                                Current plan
                            </button>
                            <button v-else @click="upgrade(p.tier)" :disabled="checkoutLoading === p.tier"
                                class="mt-5 px-4 py-2 text-sm font-medium rounded-lg bg-white text-zinc-950 hover:bg-zinc-200 transition cursor-pointer disabled:opacity-50">
                                {{ checkoutLoading === p.tier ? 'Redirecting…' : 'Choose ' + p.name }}
                            </button>
                        </div>
                    </div>

                    <div v-else-if="tier === 'free'" class="space-y-4">
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

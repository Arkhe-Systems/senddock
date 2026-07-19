<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { useSettingsStore } from '@/stores/settings'
import { useAppStore } from '@/stores/app'
import { useLicenseStore } from '@/stores/license'
import { useToastStore } from '@/stores/toast'
import { ApiError } from '@/api/client'
import AppInput from '@/components/ui/AppInput.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppAlert from '@/components/ui/AppAlert.vue'
import UserProfilePanel from '@/components/UserProfilePanel.vue'

const settingsStore = useSettingsStore()
const appStore = useAppStore()
const licenseStore = useLicenseStore()
const toast = useToastStore()

const licenseKey = ref('')
const licenseError = ref('')
const licenseNotice = ref('')
const activating = ref(false)

const tierLabel = computed(() => {
    const tier = licenseStore.status?.tier ?? 'free'
    return tier === 'free' ? 'Community (free)' : tier === 'team' ? 'Team' : 'Pro'
})

const mobileNavOpen = ref(false)
const publicUrl = ref('')
const idleTimeout = ref('120')
const formError = ref('')
const saving = ref(false)
const forbidden = ref(false)

const publicUrlWarning = computed(() => {
    const raw = publicUrl.value.trim()
    if (!raw) return 'Newsletters cannot be sent until this is set to a public address.'
    if (/localhost|127\.|0\.0\.0\.0|\[::1\]/.test(raw)) {
        return 'This points at your own machine. Unsubscribe and tracking links will not work in outgoing emails.'
    }
    if (!/^https?:\/\//.test(raw)) return 'Include the scheme, for example https://mail.example.com'
    return ''
})

onMounted(async () => {
    licenseStore.fetch(true)
    const current = await settingsStore.fetchSettings()
    if (!current) {
        forbidden.value = true
        return
    }
    publicUrl.value = current.public_url
    idleTimeout.value = String(current.session_idle_timeout_minutes)
})

async function activateLicense() {
    licenseError.value = ''
    licenseNotice.value = ''

    const key = licenseKey.value.trim()
    if (!key) {
        licenseError.value = 'Paste the license key you received by email.'
        return
    }

    activating.value = true
    try {
        const result = await licenseStore.activate(key)
        if (result.unverified) {
            licenseNotice.value = 'Key saved, but it could not be verified right now. SendDock will retry on its own.'
        } else {
            toast.success(`License active — ${result.plan || 'Pro'}`)
        }
        licenseKey.value = ''
    } catch (e) {
        licenseError.value = e instanceof ApiError ? e.message : 'Could not activate the license'
    } finally {
        activating.value = false
    }
}

async function save() {
    formError.value = ''
    const raw = publicUrl.value.trim()
    if (raw && !/^https?:\/\//.test(raw)) {
        formError.value = 'The public URL must start with http:// or https://'
        return
    }
    const minutes = Number(idleTimeout.value)
    if (!Number.isInteger(minutes) || minutes < 5 || minutes > 1440) {
        formError.value = 'The session timeout must be a whole number between 5 and 1440 minutes.'
        return
    }

    saving.value = true
    try {
        await settingsStore.updateSettings({
            public_url: raw,
            session_idle_timeout_minutes: minutes,
        })
        toast.success('Instance settings saved')
    } catch (e) {
        formError.value = e instanceof ApiError ? e.message : 'Could not save instance settings'
    } finally {
        saving.value = false
    }
}
</script>

<template>
    <div class="min-h-screen bg-zinc-950">
        <div class="md:flex md:min-h-screen">
            <header class="md:hidden sticky top-0 z-30 bg-zinc-900 border-b border-zinc-800 px-4 py-3 flex items-center justify-between gap-3">
                <h1 class="text-base font-semibold text-white">Instance</h1>
                <button type="button" @click="mobileNavOpen = !mobileNavOpen"
                    class="shrink-0 inline-flex items-center justify-center w-9 h-9 rounded-lg border border-zinc-800 text-zinc-300 hover:text-white hover:bg-zinc-800 transition cursor-pointer"
                    :aria-expanded="mobileNavOpen" aria-label="Toggle navigation">
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h16M4 18h16" />
                    </svg>
                </button>
            </header>

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

                <div class="flex-1"></div>

                <UserProfilePanel />
            </aside>

            <main class="flex-1 min-w-0 p-4 sm:p-6 md:p-8">
                <div class="max-w-2xl mx-auto space-y-8">
                    <div>
                        <h2 class="text-2xl font-bold text-white">Instance</h2>
                        <p class="text-sm text-zinc-500 mt-1">Settings that apply to this SendDock installation.</p>
                    </div>

                    <AppAlert v-if="forbidden" type="error"
                        message="Only a workspace owner can manage instance settings." />

                    <template v-else>
                        <section class="bg-zinc-900 border border-zinc-800 rounded-xl p-5 space-y-4">
                            <div>
                                <h3 class="text-sm font-semibold text-white">Public URL</h3>
                                <p class="text-xs text-zinc-500 mt-1">
                                    The address where this instance is reachable from the internet. Used to build
                                    unsubscribe and tracking links inside outgoing emails.
                                </p>
                            </div>

                            <AppInput v-model="publicUrl" label="Public URL" placeholder="https://mail.example.com" />

                            <p v-if="publicUrlWarning" class="text-xs text-amber-300">{{ publicUrlWarning }}</p>
                        </section>

                        <section class="bg-zinc-900 border border-zinc-800 rounded-xl p-5 space-y-4">
                            <div>
                                <h3 class="text-sm font-semibold text-white">Session timeout</h3>
                                <p class="text-xs text-zinc-500 mt-1">
                                    Sign users out after this many minutes without any activity. Between 5 and 1440.
                                </p>
                            </div>

                            <AppInput v-model="idleTimeout" label="Minutes of inactivity" type="number" />
                        </section>

                        <AppAlert v-if="formError" :message="formError" type="error" />

                        <div class="flex justify-end">
                            <AppButton :loading="saving" :disabled="settingsStore.loading" @click="save">
                                Save changes
                            </AppButton>
                        </div>

                        <section v-if="licenseStore.available" class="bg-zinc-900 border border-zinc-800 rounded-xl p-5 space-y-4">
                            <div class="flex items-start justify-between gap-4">
                                <div>
                                    <h3 class="text-sm font-semibold text-white">License</h3>
                                    <p class="text-xs text-zinc-500 mt-1">
                                        Paste the key you received by email after purchasing. It takes effect
                                        immediately — no restart, no editing files on the server.
                                    </p>
                                </div>
                                <span class="shrink-0 text-xs px-2 py-1 rounded-lg border border-zinc-700 text-zinc-300">
                                    {{ tierLabel }}
                                </span>
                            </div>

                            <p v-if="licenseStore.status?.has_license" class="text-xs text-zinc-400">
                                A key is already stored. Pasting a new one replaces it.
                            </p>
                            <p v-if="licenseStore.status?.has_license && licenseStore.status?.reason" class="text-xs text-amber-300">
                                {{ licenseStore.status.reason }}
                            </p>

                            <AppInput v-model="licenseKey" label="License key" placeholder="Paste your key" />

                            <AppAlert v-if="licenseError" :message="licenseError" type="error" />
                            <AppAlert v-if="licenseNotice" :message="licenseNotice" type="info" />

                            <div class="flex justify-end">
                                <AppButton :loading="activating" @click="activateLicense">
                                    Activate license
                                </AppButton>
                            </div>
                        </section>
                    </template>
                </div>
            </main>
        </div>
    </div>
</template>

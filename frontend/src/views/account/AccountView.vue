<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { api, ApiError } from '@/api/client'
import { useToastStore } from '@/stores/toast'
import AppInput from '@/components/ui/AppInput.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppAlert from '@/components/ui/AppAlert.vue'
import AppModal from '@/components/ui/AppModal.vue'
import UserProfilePanel from '@/components/UserProfilePanel.vue'
import TwoFactorSetupModal from '@/components/TwoFactorSetupModal.vue'

const auth = useAuthStore()
const toast = useToastStore()
const router = useRouter()
const mobileNavOpen = ref(false)

const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const formError = ref('')
const submitting = ref(false)

const showSetup2FA = ref(false)
const showDisable2FA = ref(false)
const disablePassword = ref('')
const disableCode = ref('')
const disableError = ref('')
const disabling = ref(false)

const showRegenerate = ref(false)
const regenCode = ref('')
const regenError = ref('')
const regenerating = ref(false)
const newRecoveryCodes = ref<string[] | null>(null)

const memberSince = computed(() => {
    if (!auth.createdAt) return '—'
    return new Date(auth.createdAt).toLocaleDateString(undefined, {
        year: 'numeric', month: 'long', day: 'numeric',
    })
})

const planLabel = computed(() => {
    const p = auth.plan?.toLowerCase() ?? ''
    if (p === 'pro') return 'Pro'
    if (p === 'team') return 'Team'
    return 'Free'
})

const passwordsMatch = computed(() => newPassword.value === confirmPassword.value)

async function handleSubmit() {
    formError.value = ''
    if (!currentPassword.value || !newPassword.value) {
        formError.value = 'fill in all password fields'
        return
    }
    if (!passwordsMatch.value) {
        formError.value = 'new password and confirmation do not match'
        return
    }
    if (newPassword.value === currentPassword.value) {
        formError.value = 'new password must be different from current'
        return
    }

    submitting.value = true
    try {
        await api('/me/password', {
            method: 'POST',
            body: {
                current_password: currentPassword.value,
                new_password: newPassword.value,
            },
        })
        currentPassword.value = ''
        newPassword.value = ''
        confirmPassword.value = ''
        toast.success('Password updated')
    } catch (e) {
        formError.value = e instanceof ApiError ? e.message : 'failed to update password'
    } finally {
        submitting.value = false
    }
}

async function on2FAEnabled() {
    await auth.checkAuth()
    toast.success('Two-factor authentication enabled')
}

async function handleDisable2FA() {
    disableError.value = ''
    if (!disablePassword.value || !disableCode.value) {
        disableError.value = 'password and code are required'
        return
    }
    disabling.value = true
    try {
        await api('/me/2fa/disable', {
            method: 'POST',
            body: { password: disablePassword.value, code: disableCode.value },
        })
        showDisable2FA.value = false
        disablePassword.value = ''
        disableCode.value = ''
        await auth.checkAuth()
        toast.success('Two-factor authentication disabled')
    } catch (e) {
        disableError.value = e instanceof ApiError ? e.message : 'failed to disable 2FA'
    } finally {
        disabling.value = false
    }
}

async function handleRegenerate() {
    regenError.value = ''
    if (!regenCode.value) {
        regenError.value = 'authentication code required'
        return
    }
    regenerating.value = true
    try {
        const res = await api<{ recovery_codes: string[] }>('/me/2fa/recovery-codes', {
            method: 'POST',
            body: { code: regenCode.value },
        })
        newRecoveryCodes.value = res.recovery_codes
        regenCode.value = ''
        toast.success('Recovery codes regenerated')
    } catch (e) {
        regenError.value = e instanceof ApiError ? e.message : 'failed to regenerate'
    } finally {
        regenerating.value = false
    }
}

function closeRegenerate() {
    showRegenerate.value = false
    regenCode.value = ''
    regenError.value = ''
    newRecoveryCodes.value = null
}

async function copyNewCodes() {
    if (!newRecoveryCodes.value) return
    await navigator.clipboard.writeText(newRecoveryCodes.value.join('\n'))
    toast.success('Recovery codes copied')
}

onMounted(async () => {
    if (!auth.email) await auth.checkAuth()
})
</script>

<template>
    <div class="min-h-screen bg-zinc-950">
        <div class="md:flex md:min-h-screen">
            <header class="md:hidden sticky top-0 z-30 bg-zinc-900 border-b border-zinc-800 px-4 py-3 flex items-center justify-between gap-3">
                <h1 class="text-base font-semibold text-white">Account</h1>
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

                <div class="flex-1"></div>

                <UserProfilePanel />
            </aside>

            <main class="flex-1 min-w-0 p-4 sm:p-6 md:p-8">
                <div class="max-w-2xl mx-auto space-y-8">
                    <div>
                        <h2 class="text-2xl font-bold text-white">Account</h2>
                        <p class="text-sm text-zinc-500 mt-1">Manage your profile and password.</p>
                    </div>

                    <section class="bg-zinc-900 border border-zinc-800 rounded-xl p-5">
                        <h3 class="text-sm font-semibold text-white mb-4">Profile</h3>
                        <dl class="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-4 text-sm">
                            <div>
                                <dt class="text-xs uppercase tracking-wide text-zinc-500 mb-1">Name</dt>
                                <dd class="text-white">{{ auth.name || '—' }}</dd>
                            </div>
                            <div>
                                <dt class="text-xs uppercase tracking-wide text-zinc-500 mb-1">Email</dt>
                                <dd class="text-white break-all">{{ auth.email || '—' }}</dd>
                            </div>
                            <div>
                                <dt class="text-xs uppercase tracking-wide text-zinc-500 mb-1">Plan</dt>
                                <dd class="text-white">{{ planLabel }}</dd>
                            </div>
                            <div>
                                <dt class="text-xs uppercase tracking-wide text-zinc-500 mb-1">Member since</dt>
                                <dd class="text-white">{{ memberSince }}</dd>
                            </div>
                        </dl>
                        <p class="text-xs text-zinc-600 mt-4">
                            Changing name or email isn't supported yet. If you need a change, contact support.
                        </p>
                    </section>

                    <section class="bg-zinc-900 border border-zinc-800 rounded-xl p-5">
                        <div class="flex items-start gap-4 mb-4">
                            <div :class="[
                                'w-10 h-10 rounded-lg flex items-center justify-center shrink-0',
                                auth.totpEnabled ? 'bg-green-500/10 text-green-400' : 'bg-zinc-800 text-zinc-400'
                            ]">
                                <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                    <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
                                </svg>
                            </div>
                            <div class="flex-1 min-w-0">
                                <div class="flex items-center gap-2 flex-wrap">
                                    <h3 class="text-sm font-semibold text-white">Two-factor authentication</h3>
                                    <span :class="[
                                        'inline-flex items-center gap-1.5 text-[10px] uppercase tracking-wider px-1.5 py-0.5 rounded border',
                                        auth.totpEnabled
                                            ? 'bg-green-500/10 text-green-400 border-green-500/30'
                                            : 'bg-zinc-800 text-zinc-400 border-zinc-700'
                                    ]">
                                        <span :class="['w-1.5 h-1.5 rounded-full', auth.totpEnabled ? 'bg-green-400' : 'bg-zinc-500']"></span>
                                        {{ auth.totpEnabled ? 'enabled' : 'disabled' }}
                                    </span>
                                </div>
                                <p class="text-xs text-zinc-500 mt-1 leading-relaxed">
                                    {{ auth.totpEnabled
                                        ? 'You\'ll be asked for a 6-digit code from your authenticator app on every sign-in. Recovery codes can be used if you lose access to your device.'
                                        : 'Add an extra layer of security using a TOTP app (Google Authenticator, Authy, 1Password, Bitwarden, etc.). Strongly recommended for any account exposed to the internet.' }}
                                </p>
                            </div>
                        </div>

                        <div v-if="!auth.totpEnabled" class="pt-1">
                            <AppButton @click="showSetup2FA = true">Enable two-factor authentication</AppButton>
                        </div>

                        <div v-else class="flex flex-wrap gap-2 pt-1">
                            <button @click="showRegenerate = true"
                                class="inline-flex items-center gap-2 px-3 py-2 text-xs font-medium text-zinc-200 bg-zinc-800 hover:bg-zinc-700 border border-zinc-700 rounded-lg transition cursor-pointer">
                                <svg xmlns="http://www.w3.org/2000/svg" class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                    <polyline points="23 4 23 10 17 10"/><polyline points="1 20 1 14 7 14"/><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/>
                                </svg>
                                Regenerate recovery codes
                            </button>
                            <button @click="showDisable2FA = true"
                                class="inline-flex items-center gap-2 px-3 py-2 text-xs font-medium text-red-400 hover:text-red-300 bg-red-500/10 hover:bg-red-500/15 border border-red-500/30 rounded-lg transition cursor-pointer">
                                <svg xmlns="http://www.w3.org/2000/svg" class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                    <circle cx="12" cy="12" r="10"/><line x1="4.93" y1="4.93" x2="19.07" y2="19.07"/>
                                </svg>
                                Disable
                            </button>
                        </div>
                    </section>

                    <section class="bg-zinc-900 border border-zinc-800 rounded-xl p-5">
                        <h3 class="text-sm font-semibold text-white mb-4">Change password</h3>
                        <form @submit.prevent="handleSubmit" class="space-y-4">
                            <AppInput v-model="currentPassword" type="password" label="Current password" required />
                            <AppInput v-model="newPassword" type="password" label="New password" required />
                            <AppInput v-model="confirmPassword" type="password" label="Confirm new password" required />
                            <p class="text-xs text-zinc-500">
                                At least 8 characters, with one uppercase letter, one number and one special character.
                            </p>
                            <AppAlert :message="formError" />
                            <AppButton :loading="submitting" :disabled="submitting">
                                {{ submitting ? 'Updating...' : 'Update password' }}
                            </AppButton>
                        </form>
                    </section>
                </div>
            </main>
        </div>

        <TwoFactorSetupModal :show="showSetup2FA" @close="showSetup2FA = false" @enabled="on2FAEnabled" />

        <AppModal :show="showDisable2FA" title="Disable two-factor authentication" @close="showDisable2FA = false">
            <form @submit.prevent="handleDisable2FA" class="space-y-4">
                <p class="text-sm text-zinc-300">
                    Confirm your password and enter a current authenticator code (or one of your recovery codes) to disable 2FA.
                </p>
                <AppInput v-model="disablePassword" type="password" label="Current password" required />
                <AppInput v-model="disableCode" label="Authentication code or recovery code" placeholder="123456" required />
                <AppAlert :message="disableError" />
                <div class="flex gap-2 justify-end">
                    <button type="button" @click="showDisable2FA = false"
                        class="px-3 py-2 text-sm text-zinc-400 hover:text-white transition cursor-pointer">Cancel</button>
                    <AppButton variant="danger" :loading="disabling" :disabled="disabling">
                        {{ disabling ? 'Disabling...' : 'Disable 2FA' }}
                    </AppButton>
                </div>
            </form>
        </AppModal>

        <AppModal :show="showRegenerate" title="Regenerate recovery codes" size="lg" @close="closeRegenerate">
            <div v-if="!newRecoveryCodes" class="space-y-4">
                <p class="text-sm text-zinc-300">
                    Enter your current 6-digit authenticator code. This will <strong>invalidate every existing recovery code</strong> and generate 10 new ones.
                </p>
                <form @submit.prevent="handleRegenerate" class="space-y-4">
                    <AppInput v-model="regenCode" label="Authentication code" placeholder="123456" required />
                    <AppAlert :message="regenError" />
                    <div class="flex gap-2 justify-end">
                        <button type="button" @click="closeRegenerate"
                            class="px-3 py-2 text-sm text-zinc-400 hover:text-white transition cursor-pointer">Cancel</button>
                        <AppButton :loading="regenerating" :disabled="regenerating">
                            {{ regenerating ? 'Generating...' : 'Generate new codes' }}
                        </AppButton>
                    </div>
                </form>
            </div>
            <div v-else class="space-y-4">
                <div class="bg-green-500/10 border border-green-500/30 rounded-lg p-3 text-sm text-green-300">
                    ✓ New recovery codes generated. Previous codes no longer work.
                </div>
                <p class="text-xs text-zinc-400">
                    Each code can be used <strong>once</strong>. Store them somewhere safe — they won't be shown again.
                </p>
                <div class="grid grid-cols-2 gap-2 bg-zinc-950 border border-zinc-800 rounded-lg p-4 font-mono text-sm">
                    <code v-for="code in newRecoveryCodes" :key="code" class="text-zinc-200 select-all">{{ code }}</code>
                </div>
                <div class="flex justify-between">
                    <button @click="copyNewCodes" class="px-3 py-2 text-xs bg-zinc-800 hover:bg-zinc-700 text-white rounded-lg transition cursor-pointer">Copy all</button>
                    <AppButton @click="closeRegenerate">I've saved them — done</AppButton>
                </div>
            </div>
        </AppModal>
    </div>
</template>

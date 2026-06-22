<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import AppInput from '@/components/ui/AppInput.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppAlert from '@/components/ui/AppAlert.vue'
import { ApiError } from '@/api/client'

const route = useRoute()

const reasonMessages: Record<string, string> = {
    session_expired: "Your session has expired. Please sign in again."
}
const reason = route.query.reason as string | undefined

const router = useRouter()
const auth = useAuthStore()

const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)
const appStore = useAppStore()
const isCloud = computed(() => appStore.deploymentMode === 'cloud')

const twoFactorToken = ref<string | null>(null)
const twoFactorCode = ref('')
const twoFactorError = ref('')
const verifying = ref(false)

const needsVerification = ref(false)
const resent = ref(false)
const deviceConfirmation = ref(false)

async function handleLogin() {
    error.value = ''
    needsVerification.value = false
    resent.value = false
    deviceConfirmation.value = false
    loading.value = true

    try {
        const result = await auth.login(email.value, password.value)
        if (result.requires_2fa && result.two_factor_token) {
            twoFactorToken.value = result.two_factor_token
            twoFactorCode.value = ''
            twoFactorError.value = ''
            return
        }
        if (result.requires_device_confirmation) {
            deviceConfirmation.value = true
            return
        }
        router.push('/dashboard')
    } catch (e: any) {
        if (e instanceof ApiError && e.code === 'email_not_verified') {
            needsVerification.value = true
        } else {
            error.value = e.message
        }
    } finally {
        loading.value = false
    }
}

async function handleResend() {
    try {
        await auth.resendVerification(email.value)
        resent.value = true
    } catch {}
}

async function handleVerify() {
    twoFactorError.value = ''
    if (!twoFactorToken.value || !twoFactorCode.value) {
        twoFactorError.value = 'enter your authentication or recovery code'
        return
    }
    verifying.value = true
    try {
        await auth.verifyTwoFactor(twoFactorToken.value, twoFactorCode.value)
        router.push('/dashboard')
    } catch (e: any) {
        twoFactorError.value = e.message
    } finally {
        verifying.value = false
    }
}

function backToLogin() {
    twoFactorToken.value = null
    twoFactorCode.value = ''
    twoFactorError.value = ''
    deviceConfirmation.value = false
    needsVerification.value = false
    password.value = ''
}
</script>

<template>
    <div class="min-h-screen bg-zinc-950 flex items-center justify-center px-4">
        <div class="w-full max-w-sm">
            <div class="text-center mb-8">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" class="w-12 h-12 mx-auto mb-4">
                    <rect width="512" height="512" rx="112" fill="#000000"/>
                    <g stroke="#ffffff" stroke-width="8" stroke-linejoin="round" stroke-linecap="round" fill="#000000">
                        <polyline points="106,236 206,176 306,236" fill="none"/>
                        <polygon points="106,236 66,216 166,156 206,176"/>
                        <polygon points="206,176 246,156 346,216 306,236"/>
                        <path d="M 206 266 Q 260 180 355 155" fill="none" stroke-dasharray="12 16"/>
                        <polygon points="106,236 106,356 206,416 206,296"/>
                        <polygon points="306,236 306,356 206,416 206,296"/>
                        <polygon points="106,236 66,256 166,316 206,296"/>
                        <polygon points="206,296 246,316 346,256 306,236"/>
                        <polygon points="336,136 371,146 446,96"/>
                        <polygon points="406,126 371,146 446,96"/>
                        <polygon points="371,146 381,176 446,96"/>
                    </g>
                </svg>
                <h1 class="text-2xl font-bold text-white">SendDock</h1>
                <p class="text-zinc-400 mt-2">Sign in to your account</p>
            </div>

            <AppAlert v-if="reason && !twoFactorToken" :message="reasonMessages[reason] ?? ''" type="info" class="mb-4" />

            <div v-if="needsVerification" class="mb-4 rounded-lg border border-amber-500/30 bg-amber-500/10 px-4 py-3 text-sm text-amber-200">
                Your account isn't activated yet. Check your email for the activation link.
                <button v-if="!resent" type="button" @click="handleResend" class="block mt-2 text-white underline hover:text-zinc-200 cursor-pointer">
                    Resend activation email
                </button>
                <span v-else class="block mt-2 text-emerald-300">A new link is on its way.</span>
            </div>

            <div v-if="deviceConfirmation" class="text-center space-y-4">
                <div class="inline-flex items-center justify-center w-12 h-12 rounded-full bg-zinc-900 border border-zinc-800 mx-auto">
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-zinc-300" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <rect x="2" y="4" width="20" height="16" rx="2"/><path d="m22 7-10 5L2 7"/>
                    </svg>
                </div>
                <h2 class="text-base font-semibold text-white">New device detected</h2>
                <p class="text-sm text-zinc-400">
                    For your security, we sent a confirmation link to <span class="text-white">{{ email }}</span>. Click it to finish signing in on this device.
                </p>
                <button type="button" @click="backToLogin" class="text-xs text-zinc-500 hover:text-white transition cursor-pointer">
                    &larr; Use a different account
                </button>
            </div>

            <form v-if="!twoFactorToken && !deviceConfirmation" @submit.prevent="handleLogin" class="space-y-4">
                <AppAlert :message="error" />
                <AppInput v-model="email" label="Email" type="email" autocomplete="email" placeholder="your@example.com" required />
                <AppInput v-model="password" label="Password" type="password" autocomplete="current-password" placeholder="••••••••" required />
                <AppButton :loading="loading">
                    {{ loading ? 'Signing in...' : 'Sign in' }}
                </AppButton>
            </form>

            <form v-if="twoFactorToken" @submit.prevent="handleVerify" class="space-y-5">
                <div class="text-center">
                    <div class="inline-flex items-center justify-center w-12 h-12 rounded-full bg-zinc-900 border border-zinc-800 mb-3">
                        <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-zinc-300" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
                        </svg>
                    </div>
                    <h2 class="text-base font-semibold text-white">Two-factor authentication</h2>
                    <p class="text-sm text-zinc-400 mt-1">
                        Enter the 6-digit code from your authenticator app, or a recovery code.
                    </p>
                </div>

                <input v-model="twoFactorCode" type="text" inputmode="text" autocomplete="one-time-code"
                    maxlength="20" autofocus
                    class="w-full px-4 py-4 bg-zinc-900 border border-zinc-800 rounded-lg text-white text-center text-2xl font-mono tracking-[0.3em] focus:outline-none focus:border-zinc-600 transition" />

                <AppAlert :message="twoFactorError" />

                <AppButton :loading="verifying">
                    {{ verifying ? 'Verifying...' : 'Verify and sign in' }}
                </AppButton>

                <p class="text-center">
                    <button type="button" @click="backToLogin" class="text-xs text-zinc-500 hover:text-white transition cursor-pointer">
                        &larr; Use a different account
                    </button>
                </p>
            </form>

            <p v-if="isCloud && !twoFactorToken && !deviceConfirmation" class="text-center text-sm text-zinc-400 mt-6">
                Don't have an account?
                <RouterLink to="/register" class="text-white hover:text-zinc-300 underline">Create one</RouterLink>
            </p>
        </div>
    </div>
</template>

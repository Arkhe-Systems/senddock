<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import QRCode from 'qrcode'
import { api, ApiError } from '@/api/client'
import { useToastStore } from '@/stores/toast'
import AppModal from '@/components/ui/AppModal.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import AppAlert from '@/components/ui/AppAlert.vue'

interface SetupResponse {
    secret: string
    otpauth_url: string
    recovery_codes: string[]
}

const props = defineProps<{ show: boolean }>()
const emit = defineEmits<{ close: []; enabled: [] }>()

const toast = useToastStore()
const step = ref<'loading' | 'scan' | 'verify' | 'codes' | 'error'>('loading')
const setup = ref<SetupResponse | null>(null)
const qrSvg = ref('')
const verifyCode = ref('')
const verifyError = ref('')
const verifying = ref(false)
const initError = ref('')

watch(() => props.show, async (open) => {
    if (!open) return
    step.value = 'loading'
    setup.value = null
    verifyCode.value = ''
    verifyError.value = ''
    initError.value = ''
    try {
        const res = await api<SetupResponse>('/me/2fa/setup', { method: 'POST' })
        setup.value = res
        qrSvg.value = await QRCode.toString(res.otpauth_url, {
            type: 'svg',
            margin: 1,
            width: 192,
            color: { dark: '#111113', light: '#ffffff' },
        })
        await nextTick()
        step.value = 'scan'
    } catch (e) {
        initError.value = e instanceof ApiError ? e.message : 'failed to start 2FA setup'
        step.value = 'error'
    }
})

async function handleVerify() {
    verifyError.value = ''
    if (!verifyCode.value || verifyCode.value.length < 6) {
        verifyError.value = 'enter the 6-digit code from your authenticator app'
        return
    }
    verifying.value = true
    try {
        await api('/me/2fa/verify', { method: 'POST', body: { code: verifyCode.value } })
        step.value = 'codes'
    } catch (e) {
        verifyError.value = e instanceof ApiError ? e.message : 'verification failed'
    } finally {
        verifying.value = false
    }
}

async function copySecret() {
    if (!setup.value) return
    await navigator.clipboard.writeText(setup.value.secret)
    toast.success('Secret copied')
}

async function copyAllCodes() {
    if (!setup.value) return
    await navigator.clipboard.writeText(setup.value.recovery_codes.join('\n'))
    toast.success('Recovery codes copied')
}

function downloadCodes() {
    if (!setup.value) return
    const body = [
        'SendDock recovery codes',
        'Each code can be used once to log in if you lose access to your authenticator app.',
        '',
        ...setup.value.recovery_codes,
    ].join('\n')
    const blob = new Blob([body], { type: 'text/plain' })
    const link = document.createElement('a')
    link.href = URL.createObjectURL(blob)
    link.download = `senddock-recovery-codes-${new Date().toISOString().slice(0, 10)}.txt`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(link.href)
}

function finish() {
    emit('enabled')
    emit('close')
}
</script>

<template>
    <AppModal :show="show" title="Enable two-factor authentication" size="lg" @close="step === 'codes' ? finish() : emit('close')">
        <div v-if="step === 'loading'" class="text-sm text-zinc-300 py-8 text-center">Preparing your TOTP secret...</div>

        <div v-else-if="step === 'error'" class="space-y-4">
            <AppAlert :message="initError" />
            <p class="text-xs text-zinc-400">Close this dialog and try again. If the problem persists, the server may not have generated a secret successfully.</p>
        </div>

        <div v-else-if="step === 'scan'" class="space-y-6">
            <div>
                <h3 class="text-base font-semibold text-white mb-1">Scan with your authenticator</h3>
                <p class="text-sm text-zinc-300">
                    Open Google Authenticator, Authy, 1Password, Bitwarden, or your password manager and scan this QR. Then click <span class="text-zinc-200">Continue</span> to confirm with your first code.
                </p>
            </div>

            <div class="flex flex-col sm:flex-row gap-5">
                <div class="bg-white rounded-xl p-3 shrink-0 self-center sm:self-start" v-html="qrSvg" />

                <div class="flex-1 min-w-0 space-y-4">
                    <div>
                        <p class="text-[11px] uppercase tracking-wider text-zinc-400 mb-1.5">Can't scan? Enter the secret manually</p>
                        <div class="flex items-center gap-2">
                            <code class="flex-1 px-3 py-2 bg-zinc-950 border border-zinc-800 rounded-lg text-xs text-zinc-200 font-mono tracking-wide break-all">{{ setup?.secret }}</code>
                            <button @click="copySecret" class="px-3 py-2 text-xs bg-zinc-850 hover:bg-zinc-700 text-white rounded-lg transition cursor-pointer shrink-0">Copy</button>
                        </div>
                    </div>
                    <dl class="grid grid-cols-2 gap-3 text-xs">
                        <div>
                            <dt class="text-zinc-400 mb-0.5">Issuer</dt>
                            <dd class="text-zinc-200 font-mono">SendDock</dd>
                        </div>
                        <div>
                            <dt class="text-zinc-400 mb-0.5">Algorithm</dt>
                            <dd class="text-zinc-200 font-mono">SHA1 · 6 digits · 30s</dd>
                        </div>
                    </dl>
                </div>
            </div>

            <div class="flex justify-end pt-2 border-t border-zinc-800">
                <AppButton @click="step = 'verify'">Continue &rarr;</AppButton>
            </div>
        </div>

        <form v-else-if="step === 'verify'" @submit.prevent="handleVerify" class="space-y-5">
            <div>
                <h3 class="text-base font-semibold text-white mb-1">Verify your authenticator</h3>
                <p class="text-sm text-zinc-300">
                    Enter the 6-digit code currently shown for <span class="text-zinc-200">SendDock</span> in your authenticator app.
                </p>
            </div>

            <input v-model="verifyCode" type="text" inputmode="numeric" autocomplete="one-time-code"
                pattern="[0-9]*" maxlength="6" autofocus
                class="w-full px-4 py-4 bg-zinc-950 border border-zinc-800 rounded-lg text-white text-center text-3xl font-mono tracking-[0.3em] focus:outline-none focus:border-emerald-500 transition" />

            <AppAlert :message="verifyError" />

            <div class="flex justify-between items-center">
                <button type="button" @click="step = 'scan'" class="text-xs text-zinc-400 hover:text-white transition cursor-pointer">&larr; Back to QR</button>
                <AppButton :loading="verifying" :disabled="verifying">
                    {{ verifying ? 'Verifying...' : 'Verify and enable' }}
                </AppButton>
            </div>
        </form>

        <div v-else-if="step === 'codes'" class="space-y-5">
            <div class="bg-emerald-500/10 border border-emerald-500/30 rounded-lg p-3 text-sm text-emerald-300 flex items-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                    <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
                <span>Two-factor authentication is now enabled on your account.</span>
            </div>

            <div>
                <h3 class="text-base font-semibold text-white mb-1">Save your recovery codes</h3>
                <p class="text-xs text-zinc-300 mb-4">
                    Each code can be used <strong class="text-zinc-200">once</strong> to sign in if you lose access to your authenticator app. Store them somewhere safe — a password manager is ideal. <strong class="text-zinc-200">They won't be shown again.</strong>
                </p>
                <div class="bg-zinc-950 border border-zinc-800 rounded-lg p-4">
                    <ol class="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-2 font-mono text-sm">
                        <li v-for="(code, i) in setup?.recovery_codes" :key="code" class="flex items-baseline gap-3">
                            <span class="text-zinc-600 text-xs tabular-nums w-5 text-right shrink-0">{{ String(i + 1).padStart(2, '0') }}.</span>
                            <code class="text-zinc-100 select-all tracking-wider">{{ code }}</code>
                        </li>
                    </ol>
                </div>
            </div>

            <div class="flex flex-wrap gap-2 items-center justify-between pt-2 border-t border-zinc-800">
                <div class="flex gap-2">
                    <button @click="copyAllCodes" class="inline-flex items-center gap-1.5 px-3 py-2 text-xs bg-zinc-850 hover:bg-zinc-700 text-white rounded-lg transition cursor-pointer">
                        <svg xmlns="http://www.w3.org/2000/svg" class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
                        </svg>
                        Copy all
                    </button>
                    <button @click="downloadCodes" class="inline-flex items-center gap-1.5 px-3 py-2 text-xs bg-zinc-850 hover:bg-zinc-700 text-white rounded-lg transition cursor-pointer">
                        <svg xmlns="http://www.w3.org/2000/svg" class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/>
                        </svg>
                        Download .txt
                    </button>
                </div>
                <AppButton @click="finish">I've saved them — done</AppButton>
            </div>
        </div>
    </AppModal>
</template>

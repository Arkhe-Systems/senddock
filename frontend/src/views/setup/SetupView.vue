<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { useToastStore } from '@/stores/toast'
import { hasHttpScheme, isLoopbackUrl } from '@/utils/publicUrl'
import AppInput from '@/components/ui/AppInput.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppAlert from '@/components/ui/AppAlert.vue'

const router = useRouter()
const auth = useAuthStore()
const app = useAppStore()
const toast = useToastStore()

const name = ref('')
const email = ref('')
const password = ref('')
const passwordConfirm = ref('')
const publicUrl = ref(app.publicUrl || window.location.origin)
const error = ref('')
const loading = ref(false)

const publicUrlWarning = computed(() => {
    const raw = publicUrl.value.trim()
    if (!raw) return ''
    if (isLoopbackUrl(raw)) {
        return 'This points at your own machine, so unsubscribe and tracking links will not work in outgoing emails. Fine for a local test — you can change it later under Instance.'
    }
    return ''
})

function validatePassword(pw: string): string | null {
    if (pw.length < 8) return 'Password must be at least 8 characters'
    if (!/[A-Z]/.test(pw)) return 'Password must contain at least one uppercase letter'
    if (!/[0-9]/.test(pw)) return 'Password must contain at least one number'
    if (!/[^A-Za-z0-9]/.test(pw)) return 'Password must contain at least one special character'
    return null
}

const passwordStrength = computed(() => {
    const pw = password.value
    if (!pw) return { label: '', color: '', width: 0 }
    let score = 0
    if (pw.length >= 8) score++
    if (pw.length >= 14) score++
    if (/[A-Z]/.test(pw)) score++
    if (/[0-9]/.test(pw)) score++
    if (/[^A-Za-z0-9]/.test(pw)) score++
    if (score <= 1) return { label: 'Weak', color: 'bg-red-500 text-red-400', width: 25 }
    if (score === 2) return { label: 'Fair', color: 'bg-yellow-500 text-yellow-400', width: 50 }
    if (score === 3) return { label: 'Good', color: 'bg-blue-500 text-blue-400', width: 75 }
    return { label: 'Strong', color: 'bg-green-500 text-green-400', width: 100 }
})

async function handleSetup() {
    error.value = ''

    if (!name.value || !email.value || !password.value) {
        error.value = 'All fields are required'
        return
    }

    const pwErr = validatePassword(password.value)
    if (pwErr) {
        error.value = pwErr
        return
    }

    if (password.value !== passwordConfirm.value) {
        error.value = 'Passwords do not match'
        return
    }

    const url = publicUrl.value.trim()
    if (url && !hasHttpScheme(url)) {
        error.value = 'The public URL must start with http:// or https://'
        return
    }

    loading.value = true
    try {
        await api('/setup', {
            method: 'POST',
            body: { name: name.value, email: email.value, password: password.value, public_url: publicUrl.value.trim() },
        })
        auth.isAuthenticated = true
        app.setupRequired = false
        app.publicUrl = publicUrl.value.trim()
        toast.success('Welcome to SendDock! Your admin account is ready.')
        router.push('/dashboard')
    } catch (e: any) {
        error.value = e.message || 'Setup failed'
    } finally {
        loading.value = false
    }
}
</script>

<template>
    <div class="min-h-screen bg-zinc-950 flex items-center justify-center px-4">
        <div class="w-full max-w-sm">
            <div class="text-center mb-8">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" class="w-16 h-16 mx-auto mb-4">
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
                <h1 class="text-2xl font-bold text-white">Welcome to SendDock</h1>
                <p class="text-zinc-400 mt-2">Create your admin account to get started.</p>
            </div>

            <form @submit.prevent="handleSetup" class="space-y-4">
                <AppAlert :message="error" />
                <AppInput v-model="name" label="Full Name" placeholder="John Doe" required />
                <AppInput v-model="email" label="Email" type="email" placeholder="admin@example.com" required />
                <div>
                    <AppInput v-model="password" label="Password" type="password" placeholder="Min 8 chars with uppercase, number and symbol" required />
                    <div v-if="password" class="mt-2">
                        <div class="flex items-center justify-between mb-1">
                            <span class="text-xs" :class="passwordStrength.color.split(' ')[1]">{{ passwordStrength.label }}</span>
                        </div>
                        <div class="h-1 bg-zinc-800 rounded-full overflow-hidden">
                            <div class="h-full transition-all" :class="passwordStrength.color.split(' ')[0]" :style="{ width: passwordStrength.width + '%' }"></div>
                        </div>
                    </div>
                </div>
                <AppInput v-model="passwordConfirm" label="Confirm Password" type="password" placeholder="Repeat your password" required />

                <div class="pt-2 border-t border-zinc-800">
                    <AppInput v-model="publicUrl" label="Public URL" placeholder="https://mail.example.com" />
                    <p class="text-xs text-zinc-500 mt-2">
                        Where this instance is reachable from the internet. Used for unsubscribe and tracking links.
                        You can change it later under Instance.
                    </p>
                    <p v-if="publicUrlWarning" class="text-xs text-amber-300 mt-2">{{ publicUrlWarning }}</p>
                </div>

                <AppButton :loading="loading">
                    {{ loading ? 'Setting up...' : 'Complete Setup' }}
                </AppButton>
            </form>
        </div>
    </div>
</template>

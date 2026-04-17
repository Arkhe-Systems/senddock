<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { deploymentMode } from '@/router'
import AppInput from '@/components/ui/AppInput.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppAlert from '@/components/ui/AppAlert.vue'

const route = useRoute()

const reasonMessages: Record<string, string> = {
    auth_required: "You need to sign in to access that page.",
    session_expired: "Your session has expired. Please sign in again."
}
const reason = route.query.reason as string | undefined

const router = useRouter()
const auth = useAuthStore()

const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)
const isCloud = computed(() => deploymentMode.value === 'cloud')

async function handleLogin() {
    error.value = ''
    loading.value = true

    try {
        await auth.login(email.value, password.value)
        router.push('/dashboard')
    } catch (e: any) {
        error.value = e.message
    } finally {
        loading.value = false
    }
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

            <AppAlert v-if="reason" :message="reasonMessages[reason] ?? ''" type="info" class="mb-4" />

            <form @submit.prevent="handleLogin" class="space-y-4">
                <AppAlert :message="error" />
                <AppInput v-model="email" label="Email" type="email" placeholder="your@example.com" required />
                <AppInput v-model="password" label="Password" type="password" placeholder="••••••••" required />
                <AppButton :loading="loading">
                    {{ loading ? 'Signing in...' : 'Sign in' }}
                </AppButton>
            </form>

            <p v-if="isCloud" class="text-center text-sm text-zinc-400 mt-6">
                Don't have an account?
                <RouterLink to="/register" class="text-white hover:text-zinc-300 underline">Create one</RouterLink>
            </p>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import AppInput from '@/components/ui/AppInput.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppAlert from '@/components/ui/AppAlert.vue'

const router = useRouter()
const auth = useAuthStore()

const name = ref('')
const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function handleSetup() {
    error.value = ''

    if (!name.value || !email.value || !password.value) {
        error.value = 'All fields are required'
        return
    }

    if (password.value.length < 8) {
        error.value = 'Password must be at least 8 characters'
        return
    }

    loading.value = true
    try {
        await api('/setup', {
            method: 'POST',
            body: { name: name.value, email: email.value, password: password.value },
        })
        auth.isAuthenticated = true
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
                <AppInput v-model="password" label="Password" type="password" placeholder="Minimum 8 characters" required />
                <AppButton :loading="loading">
                    {{ loading ? 'Setting up...' : 'Complete Setup' }}
                </AppButton>
            </form>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import AppInput from '@/components/ui/AppInput.vue';
import AppButton from '@/components/ui/AppButton.vue';
import AppAlert from '@/components/ui/AppAlert.vue';
import AppCard from '@/components/ui/AppCard.vue';

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

async function handleLogin(){
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
        <AppCard title="Senddock" subtitle="Sign in to your account">
            <AppAlert v-if="reason" :message="reasonMessages[reason] ?? ''" type="info"/>
            <form @submit.prevent="handleLogin" class="space-y-4">
                <AppAlert :message="error" />
                <AppInput id="email" v-model="email" label="Email" type="email" placeholder="your@example.com" required/>
                <AppInput id="password" v-model="password" label="Password" type="password" placeholder="••••••••" required/>
                <AppButton :loading="loading">
                    {{ loading ? 'Signing in...' : 'Sign in' }}
                </AppButton>
            </form>

            <p class="text-center text-sm text-zinc-400 mt-6">
                Don't have an account?
                <RouterLink to="/register" class="text-indigo-400 hover:text-indigo-300">Create one</RouterLink>
            </p>
        </AppCard>
    </div>
</template>
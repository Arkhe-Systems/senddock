<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import AppInput from '@/components/ui/AppInput.vue';
import AppButton from '@/components/ui/AppButton.vue';
import AppAlert from '@/components/ui/AppAlert.vue';
import AppCard from '@/components/ui/AppCard.vue';

const router = useRouter()
const auth = useAuthStore()

const name = ref('')
const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function handleRegister(){
    error.value = ''
    loading.value = true

    try {
        await auth.register(email.value, password.value, name.value)
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
        <AppCard title="Senddock" subtitle="Create your account">
            <form @submit.prevent="handleRegister" class="space-y-4">
                <AppAlert :message="error" />
                <AppInput id="name" v-model="name" label="Full Name" type="text" placeholder="Jhon Doe" required/>
                <AppInput id="email" v-model="email" label="Email" type="email" placeholder="your@example.com" required/>
                <AppInput id="password" v-model="password" label="Password" type="password" placeholder="••••••••" required/>
                <AppButton :loading="loading">
                    {{ loading ? 'Signing up...' : 'Sign up' }}
                </AppButton>
            </form>

            <p class="text-center text-sm text-zinc-400 mt-6">
                Already have an account?
                <RouterLink to="/login" class="text-indigo-400 hover:text-indigo-300">Sign in</RouterLink>
            </p>
        </AppCard>
    </div>
</template>
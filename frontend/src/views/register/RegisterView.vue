<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth';
import AppInput from '@/components/ui/AppInput.vue';
import AppButton from '@/components/ui/AppButton.vue';
import AppAlert from '@/components/ui/AppAlert.vue';
import AppCard from '@/components/ui/AppCard.vue';

const auth = useAuthStore()

const name = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const error = ref('')
const loading = ref(false)
const submitted = ref(false)
const resent = ref(false)

async function handleRegister(){
    error.value = ''
    if (password.value !== confirmPassword.value) {
        error.value = "Passwords don't match"
        return
    }
    loading.value = true
    try {
        await auth.signup(email.value, password.value, name.value)
        submitted.value = true
    } catch (e: any) {
        error.value = e.message
    } finally {
        loading.value = false
    }
}

async function handleResend(){
    try {
        await auth.resendVerification(email.value)
        resent.value = true
    } catch {}
}
</script>

<template>
    <div class="min-h-screen bg-zinc-950 flex items-center justify-center px-4">
        <AppCard v-if="!submitted" title="Senddock" subtitle="Create your account">
            <form @submit.prevent="handleRegister" class="space-y-4">
                <AppAlert :message="error" />
                <AppInput id="name" v-model="name" label="Full Name" type="text" autocomplete="name" placeholder="Jhon Doe" required/>
                <AppInput id="email" v-model="email" label="Email" type="email" autocomplete="email" placeholder="your@example.com" required/>
                <AppInput id="password" v-model="password" label="Password" type="password" autocomplete="new-password" placeholder="••••••••" required/>
                <AppInput id="confirmPassword" v-model="confirmPassword" label="Confirm password" type="password" autocomplete="new-password" placeholder="••••••••" required/>
                <AppButton :loading="loading">
                    {{ loading ? 'Creating account...' : 'Sign up' }}
                </AppButton>
            </form>

            <p class="text-center text-sm text-zinc-300 mt-6">
                Already have an account?
                <RouterLink to="/login" class="text-white hover:text-zinc-300 underline">Sign in</RouterLink>
            </p>
        </AppCard>

        <AppCard v-else title="Check your email" subtitle="One more step to activate your account">
            <div class="space-y-4 text-center">
                <div class="inline-flex items-center justify-center w-12 h-12 rounded-full bg-zinc-900 border border-zinc-800 mx-auto">
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-zinc-300" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <rect x="2" y="4" width="20" height="16" rx="2"/><path d="m22 7-10 5L2 7"/>
                    </svg>
                </div>
                <p class="text-sm text-zinc-300">
                    We sent an activation link to <span class="text-white">{{ email }}</span>. Click it to activate your account — the link expires in 24 hours.
                </p>
                <p v-if="resent" class="text-sm text-emerald-400">A new link is on its way.</p>
                <button v-else type="button" @click="handleResend" class="text-sm text-zinc-300 hover:text-white underline cursor-pointer">
                    Didn't get it? Resend
                </button>
                <p class="text-sm text-zinc-300 pt-2 border-t border-zinc-800">
                    <RouterLink to="/login" class="text-white hover:text-zinc-300 underline">Back to sign in</RouterLink>
                </p>
            </div>
        </AppCard>
    </div>
</template>

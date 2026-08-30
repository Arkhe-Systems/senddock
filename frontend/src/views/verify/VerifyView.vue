<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import AppCard from '@/components/ui/AppCard.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const state = ref<'verifying' | 'success' | 'pending' | 'error'>('verifying')
const message = ref('')

onMounted(async () => {
    const token = route.query.token as string | undefined
    if (!token) {
        state.value = 'error'
        message.value = 'This link is missing its verification token.'
        return
    }
    try {
        const res = await auth.verifyEmail(token)
        if (res.pending) {
            state.value = 'pending'
            return
        }
        state.value = 'success'
        setTimeout(() => router.push('/dashboard'), 1200)
    } catch (e: any) {
        state.value = 'error'
        message.value = e.message || 'This link is invalid or has expired.'
    }
})
</script>

<template>
    <div class="min-h-screen bg-zinc-950 flex items-center justify-center px-4">
        <AppCard title="SendDock">
            <div class="text-center space-y-4 py-2">
                <template v-if="state === 'verifying'">
                    <p class="text-zinc-300">Activating your account…</p>
                </template>

                <template v-else-if="state === 'pending'">
                    <div class="inline-flex items-center justify-center w-12 h-12 rounded-full bg-emerald-500/10 border border-emerald-500/30 mx-auto">
                        <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-emerald-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
                    </div>
                    <p class="text-white font-medium">Email verified</p>
                    <p class="text-sm text-zinc-300">Return to your other device — it will finish signing you in.</p>
                </template>

                <template v-else-if="state === 'success'">
                    <div class="inline-flex items-center justify-center w-12 h-12 rounded-full bg-emerald-500/10 border border-emerald-500/30 mx-auto">
                        <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-emerald-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
                    </div>
                    <p class="text-white font-medium">Account activated!</p>
                    <p class="text-sm text-zinc-300">Taking you to your dashboard…</p>
                </template>

                <template v-else>
                    <div class="inline-flex items-center justify-center w-12 h-12 rounded-full bg-red-500/10 border border-red-500/30 mx-auto">
                        <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-red-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                    </div>
                    <p class="text-white font-medium">Couldn't activate</p>
                    <p class="text-sm text-zinc-300">{{ message }}</p>
                    <p class="text-sm pt-2">
                        <RouterLink to="/login" class="text-white hover:text-zinc-300 underline">Back to sign in</RouterLink>
                    </p>
                </template>
            </div>
        </AppCard>
    </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { api, ApiError } from '@/api/client'
import { useToastStore } from '@/stores/toast'
import AppInput from '@/components/ui/AppInput.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppAlert from '@/components/ui/AppAlert.vue'
import UserProfilePanel from '@/components/UserProfilePanel.vue'

const auth = useAuthStore()
const toast = useToastStore()
const router = useRouter()
const mobileNavOpen = ref(false)

const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const formError = ref('')
const submitting = ref(false)

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

                <nav class="space-y-1 flex-1">
                    <RouterLink to="/account" class="block px-3 py-2 text-sm rounded-lg bg-zinc-800 text-white">
                        Account
                    </RouterLink>
                    <RouterLink to="/billing" class="block px-3 py-2 text-sm rounded-lg text-zinc-400 hover:text-white hover:bg-zinc-800 transition">
                        Billing
                    </RouterLink>
                </nav>

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
    </div>
</template>

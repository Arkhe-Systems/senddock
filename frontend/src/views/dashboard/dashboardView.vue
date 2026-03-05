<script setup lang="ts">
import { useAuthStore } from '@/stores/auth';
import { useRouter } from 'vue-router';
import AppButton from '@/components/ui/AppButton.vue';

const auth = useAuthStore()
const router = useRouter()

async function handleLogout(){
    await auth.logout()
    router.push('/login')
}
</script>

<template>
    <div class="min-h-screen bg-zinc-950 p-8">
        <div class="max-w-4xl mx-auto">
            <div class="flex items-center justify-between mb-8">
                <h1 class="text-2xl font-bold text-white">dashboard</h1>
                <button @click="handleLogout" class="px-4 py-2 text-sm text-zinc-400 hover:text-white border border-zinc-800 rounded-lg transition"
                >
                    Logout
                </button>
            </div>
            <p class="text-zinc-400">Welcome to SendDock. Projects will appear here</p>
        </div>
    </div>
</template>
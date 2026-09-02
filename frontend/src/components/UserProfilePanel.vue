<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { useWorkspaceStore } from '@/stores/workspaces'
import { useLicenseStore } from '@/stores/license'
import { useToastStore } from '@/stores/toast'
import { User, CreditCard, HardDrive, LogOut } from 'lucide-vue-next'
import AppNavLink from '@/components/ui/AppNavLink.vue'

const auth = useAuthStore()
const appStore = useAppStore()
const workspaceStore = useWorkspaceStore()
const licenseStore = useLicenseStore()
const toast = useToastStore()
const router = useRouter()

const isSelfHosted = computed(() => appStore.deploymentMode !== 'cloud')

const initials = computed(() => {
    const source = (auth.name || auth.email || '').trim()
    if (!source) return '?'
    const parts = source.split(/\s+/).filter(Boolean)
    const first = parts[0] ?? ''
    const second = parts[1] ?? ''
    if (first && second) {
        return ((first[0] ?? '') + (second[0] ?? '')).toUpperCase()
    }
    return source.slice(0, 2).toUpperCase()
})

const planLabel = computed(() => {
    const p = auth.plan?.toLowerCase() ?? ''
    if (p === 'pro') return 'Pro'
    if (p === 'team') return 'Team'
    return 'Free'
})

const planClass = computed(() => {
    const p = auth.plan?.toLowerCase() ?? ''
    if (p === 'pro') return 'bg-amber-500/15 text-amber-400 border-amber-500/30'
    if (p === 'team') return 'bg-purple-500/15 text-purple-400 border-purple-500/30'
    return 'bg-zinc-700/30 text-zinc-300 border-zinc-700'
})

async function handleLogout() {
    workspaceStore.reset()
    licenseStore.reset()
    await auth.logout()
    toast.success('Signed out')
    router.push('/login')
}
</script>

<template>
    <div class="border-t border-zinc-800 pt-4 mt-4">
        <div class="flex items-center gap-3 mb-3">
            <div class="w-9 h-9 rounded-full bg-zinc-850 text-zinc-200 flex items-center justify-center text-xs font-semibold shrink-0">
                {{ initials }}
            </div>
            <div class="min-w-0 flex-1">
                <p class="text-sm font-medium text-white truncate">{{ auth.name || 'User' }}</p>
                <p class="text-xs text-zinc-400 truncate">{{ auth.email || '—' }}</p>
            </div>
            <span :class="['text-[10px] uppercase tracking-wider px-1.5 py-0.5 rounded border whitespace-nowrap', planClass]">
                {{ planLabel }}
            </span>
        </div>

        <div class="space-y-1">
            <AppNavLink to="/account">
                <User class="w-4 h-4 shrink-0" />
                <span>Account</span>
            </AppNavLink>
            <AppNavLink to="/billing">
                <CreditCard class="w-4 h-4 shrink-0" />
                <span>Billing</span>
            </AppNavLink>
            <AppNavLink v-if="isSelfHosted" to="/instance">
                <HardDrive class="w-4 h-4 shrink-0" />
                <span>Instance</span>
            </AppNavLink>
            <button type="button" @click="handleLogout"
                class="w-full flex items-center gap-2.5 px-3 py-2 text-sm rounded-lg text-zinc-300 hover:text-white hover:bg-zinc-850 transition cursor-pointer">
                <LogOut class="w-4 h-4 shrink-0" />
                <span>Logout</span>
            </button>
        </div>
    </div>
</template>

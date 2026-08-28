<script setup lang="ts">
import AppButton from '@/components/ui/AppButton.vue'
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { checkoutUrl, isLaunchPromoActive, LAUNCH_DISCOUNT_CODE, type Tier } from '@/config/checkout'

const props = withDefaults(defineProps<{
    title: string
    description: string
    tier?: Tier
}>(), {
    tier: 'pro',
})

const router = useRouter()
const appStore = useAppStore()
const isCloud = computed(() => appStore.deploymentMode === 'cloud')

const tierLabel = computed(() => (props.tier === 'team' ? 'Team' : 'Pro'))
const tierPrice = computed(() => (props.tier === 'team' ? '$29/mo · $290/yr' : '$9/mo · $90/yr'))

const cloudPlan = computed(() => (props.tier === 'team' ? 'Growth' : 'Starter'))
const badgeLabel = computed(() => (isCloud.value ? cloudPlan.value : tierLabel.value))

const buyHref = computed(() => {
    return checkoutUrl(props.tier, props.tier === 'pro' && isLaunchPromoActive()
        ? { discount: LAUNCH_DISCOUNT_CODE }
        : undefined)
})

const promoVisible = computed(() => props.tier === 'pro' && isLaunchPromoActive())
</script>

<template>
    <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-6 sm:p-8 text-center">
        <div class="inline-flex items-center gap-2 px-2.5 py-0.5 rounded-full bg-amber-500/15 text-amber-400 border border-amber-500/30 text-[11px] font-semibold tracking-wider uppercase mb-4">
            {{ badgeLabel }}
        </div>
        <h2 class="text-base sm:text-lg font-semibold text-white mb-2">{{ title }}</h2>
        <p class="text-sm text-zinc-300 mb-6 max-w-md mx-auto">{{ description }}</p>

        <template v-if="isCloud">
            <AppButton size="lg" @click="router.push('/billing')"
                class="mb-4">
                Upgrade to {{ cloudPlan }}
            </AppButton>
            <p class="text-xs text-zinc-400 max-w-md mx-auto">
                Change your plan anytime from
                <button @click="router.push('/billing')" class="underline decoration-zinc-700 hover:text-zinc-300 cursor-pointer">Billing</button>.
            </p>
        </template>

        <template v-else>
            <div class="flex flex-col sm:flex-row items-center justify-center gap-2 mb-4">
                <a :href="buyHref" target="_blank" rel="noopener"
                    class="inline-block px-6 py-2.5 text-sm font-medium rounded-lg border border-emerald-500/45 bg-emerald-500/12 text-emerald-300 hover:bg-emerald-500/20 hover:border-emerald-500/70 hover:text-emerald-200 transition cursor-pointer">
                    Subscribe to {{ tierLabel }} — {{ tierPrice }}
                </a>
                <a href="https://senddock.dev/#pricing" target="_blank" rel="noopener"
                    class="inline-block px-3 py-2 text-xs text-zinc-300 hover:text-white transition">
                    Compare plans →
                </a>
            </div>

            <p v-if="promoVisible" class="text-xs text-amber-400 mb-3">
                🎁 Code <span class="font-mono">{{ LAUNCH_DISCOUNT_CODE }}</span> applied · 3 months free for the first 50 customers
            </p>

            <p class="text-xs text-zinc-400 max-w-md mx-auto">
                After payment you'll get a license key by email. Paste it under <span class="text-zinc-300">Instance → License</span> and it activates right away.
                <a href="https://docs.senddock.dev/self-hosting/configuration#plans-licensing" target="_blank" rel="noopener" class="underline decoration-zinc-700 hover:text-zinc-300">Read the docs</a>.
            </p>
        </template>
    </div>
</template>

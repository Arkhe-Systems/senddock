<script setup lang="ts">
import { useToastStore } from '@/stores/toast'

const toast = useToastStore()
</script>

<template>
    <div class="fixed top-4 right-4 z-[100] space-y-2">
        <TransitionGroup name="toast">
            <div v-for="item in toast.toasts" :key="item.id"
                :class="[
                    'px-4 py-3 rounded-lg text-sm shadow-lg border cursor-pointer min-w-72 backdrop-blur-none',
                    item.type === 'success' && 'bg-zinc-950 border-green-500/30 text-green-400',
                    item.type === 'error' && 'bg-zinc-950 border-red-500/30 text-red-400',
                    item.type === 'info' && 'bg-zinc-950 border-yellow-500/30 text-yellow-400',
                ]"
                @click="toast.remove(item.id)">
                {{ item.message }}
            </div>
        </TransitionGroup>
    </div>
</template>

<style scoped>
.toast-enter-active {
    transition: all 0.3s ease;
}
.toast-leave-active {
    transition: all 0.2s ease;
}
.toast-enter-from {
    opacity: 0;
    transform: translateX(100%);
}
.toast-leave-to {
    opacity: 0;
    transform: translateX(100%);
}
</style>

<script setup lang="ts">
import { useToastStore } from '@/stores/toast'

const toast = useToastStore()
</script>

<template>
    <div class="fixed top-4 right-4 z-50 space-y-2">
        <TransitionGroup name="toast">
            <div v-for="item in toast.toasts" :key="item.id"
                :class="[
                    'px-4 py-3 rounded-lg text-sm shadow-lg border cursor-pointer min-w-72',
                    item.type === 'success' && 'bg-green-500/10 border-green-500/20 text-green-400',
                    item.type === 'error' && 'bg-red-500/10 border-red-500/20 text-red-400',
                    item.type === 'info' && 'bg-yellow-500/10 border-yellow-500/20 text-yellow-400',
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

<script setup lang="ts">
import { onMounted, onUnmounted, watch, computed } from 'vue'

const props = defineProps<{
    title?: string,
    show?: boolean,
    size?: 'sm' | 'md' | 'lg' | 'xl',
}>()

const emit = defineEmits<{
    close: []
}>()

const widthClass = computed(() => {
    switch (props.size) {
        case 'sm': return 'max-w-sm'
        case 'lg': return 'max-w-2xl'
        case 'xl': return 'max-w-4xl'
        default: return 'max-w-md'
    }
})

function onKey(e: KeyboardEvent) {
    if (e.key === 'Escape' && props.show) emit('close')
}

onMounted(() => window.addEventListener('keydown', onKey))
onUnmounted(() => window.removeEventListener('keydown', onKey))

watch(() => props.show, (v) => {
    document.body.style.overflow = v ? 'hidden' : ''
})
</script>
<template>
    <Teleport to="body">
        <div v-if="show" class="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
            <div :class="['bg-zinc-900 border border-zinc-800 rounded-xl shadow-2xl w-full max-h-[90vh] flex flex-col', widthClass]">
                <div class="flex items-center justify-between px-6 py-4 border-b border-zinc-800">
                    <h2 class="text-base font-semibold text-white">{{ title }}</h2>
                    <button @click="emit('close')" type="button" aria-label="Close"
                        class="w-8 h-8 flex items-center justify-center rounded-lg text-zinc-400 hover:text-white hover:bg-zinc-800 transition cursor-pointer">
                        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
                            <line x1="18" y1="6" x2="6" y2="18"></line>
                            <line x1="6" y1="6" x2="18" y2="18"></line>
                        </svg>
                    </button>
                </div>
                <div class="px-6 py-5 overflow-y-auto">
                    <slot />
                </div>
            </div>
        </div>
    </Teleport>
</template>

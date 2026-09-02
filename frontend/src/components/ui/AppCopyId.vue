<script setup lang="ts">
import { ref, computed } from 'vue'
import { Copy, Check } from 'lucide-vue-next'

const props = defineProps<{ value: string; label?: string; full?: boolean }>()

const copied = ref(false)

const display = computed(() => {
    if (props.full || props.value.length <= 14) return props.value
    return props.value.slice(0, 8) + '…' + props.value.slice(-4)
})

async function copy() {
    try {
        await navigator.clipboard.writeText(props.value)
    } catch {
        const el = document.createElement('textarea')
        el.value = props.value
        document.body.appendChild(el)
        el.select()
        document.execCommand('copy')
        document.body.removeChild(el)
    }
    copied.value = true
    setTimeout(() => (copied.value = false), 1600)
}
</script>

<template>
    <button type="button" @click.stop="copy" :title="`Copy ${value}`"
        class="group inline-flex items-center gap-2 bg-zinc-950 border border-zinc-800 hover:border-zinc-600 rounded-lg px-2.5 py-1.5 transition cursor-pointer max-w-full">
        <span v-if="label" class="text-xs text-zinc-400 shrink-0">{{ label }}</span>
        <code class="font-mono text-xs text-zinc-300 truncate">{{ display }}</code>
        <span class="inline-flex items-center gap-1 text-xs shrink-0"
            :class="copied ? 'text-emerald-400' : 'text-zinc-400 group-hover:text-white'">
            <Check v-if="copied" class="w-3.5 h-3.5" />
            <Copy v-else class="w-3.5 h-3.5" />
            {{ copied ? 'Copied' : 'Copy' }}
        </span>
    </button>
</template>

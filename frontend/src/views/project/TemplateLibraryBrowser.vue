<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { api } from '@/api/client'
import { useToastStore } from '@/stores/toast'
import AppModal from '@/components/ui/AppModal.vue'
import AppButton from '@/components/ui/AppButton.vue'

interface LibraryEntry {
    id: string
    name: string
    category: string
    description: string
    thumbnail_url: string
    html_url: string
    variables: string[]
}

interface Template {
    id: string
    name: string
    subject: string
    html_body: string
    text_body: string
    created_at: string
    updated_at: string
}

const props = defineProps<{
    show: boolean
    projectId: string
}>()

const emit = defineEmits<{
    close: []
    used: [tmpl: Template]
}>()

const toast = useToastStore()

const entries = ref<LibraryEntry[]>([])
const loading = ref(false)
const errorMessage = ref('')
const activeCategory = ref('all')
const useLoading = ref('')
const thumbnailFailed = ref<Record<string, boolean>>({})

const categories = computed(() => {
    const cats = new Set<string>(['all'])
    entries.value.forEach(e => cats.add(e.category))
    return Array.from(cats)
})

const filteredEntries = computed(() =>
    activeCategory.value === 'all'
        ? entries.value
        : entries.value.filter(e => e.category === activeCategory.value)
)

async function fetchLibrary() {
    loading.value = true
    errorMessage.value = ''
    try {
        const res = await api<LibraryEntry[] | null>(`/projects/${props.projectId}/templates/library`)
        entries.value = res || []
        activeCategory.value = 'all'
        thumbnailFailed.value = {}
    } catch (e: any) {
        errorMessage.value = e?.message || 'Failed to load library'
    } finally {
        loading.value = false
    }
}

async function useTemplate(entry: LibraryEntry) {
    if (useLoading.value) return
    useLoading.value = entry.id
    try {
        const tmpl = await api<Template>(
            `/projects/${props.projectId}/templates/library/${entry.id}/use`,
            { method: 'POST' }
        )
        toast.success(`Template "${entry.name}" added to your project`)
        emit('used', tmpl)
        emit('close')
    } catch (e: any) {
        toast.error(e?.message || 'Failed to clone template')
    } finally {
        useLoading.value = ''
    }
}

function categoryLabel(c: string) {
    if (c === 'all') return 'All'
    return c.charAt(0).toUpperCase() + c.slice(1)
}

watch(() => props.show, (v) => {
    if (v) fetchLibrary()
})
</script>

<template>
    <AppModal :show="show" size="xl" title="Template Library" @close="emit('close')">
        <div v-if="loading" class="py-12 text-center text-sm text-zinc-500">Loading library…</div>

        <div v-else-if="errorMessage" class="py-12 text-center">
            <p class="text-sm text-red-400 mb-3">{{ errorMessage }}</p>
            <button @click="fetchLibrary" type="button"
                class="text-sm text-zinc-400 hover:text-white underline cursor-pointer">
                Retry
            </button>
        </div>

        <div v-else-if="entries.length === 0" class="py-12 text-center max-w-md mx-auto">
            <p class="text-zinc-300 font-medium mb-2">The library is empty right now.</p>
            <p class="text-zinc-500 text-sm mb-5">
                Templates live in a separate community repo. Be the first to contribute a starter — every SendDock instance picks it up within an hour.
            </p>
            <a href="https://github.com/Arkhe-Systems/senddock-templates" target="_blank" rel="noopener"
                class="inline-flex items-center gap-1.5 px-4 py-2 text-sm bg-zinc-800 hover:bg-zinc-700 text-white rounded-lg transition">
                Open the templates repo
                <span aria-hidden="true">↗</span>
            </a>
        </div>

        <div v-else class="space-y-4">
            <div class="flex gap-1.5 border-b border-zinc-800 pb-3 overflow-x-auto">
                <button v-for="cat in categories" :key="cat" @click="activeCategory = cat" type="button"
                    :class="[
                        'px-3 py-1.5 text-sm rounded-md transition cursor-pointer whitespace-nowrap',
                        activeCategory === cat
                            ? 'bg-zinc-800 text-white'
                            : 'text-zinc-500 hover:text-white'
                    ]">
                    {{ categoryLabel(cat) }}
                </button>
            </div>

            <div v-if="filteredEntries.length === 0" class="py-8 text-center text-sm text-zinc-500">
                No templates in this category yet.
            </div>

            <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                <div v-for="entry in filteredEntries" :key="entry.id"
                    class="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden flex flex-col hover:border-zinc-700 transition">
                    <div class="aspect-[3/2] bg-zinc-950 border-b border-zinc-800 overflow-hidden flex items-center justify-center">
                        <img v-if="!thumbnailFailed[entry.id]" :src="entry.thumbnail_url" :alt="entry.name"
                            loading="lazy"
                            class="w-full h-full object-cover"
                            @error="thumbnailFailed[entry.id] = true" />
                        <span v-else class="text-xs text-zinc-600">no preview</span>
                    </div>
                    <div class="p-3 flex-1 flex flex-col">
                        <div class="flex items-start justify-between gap-2 mb-1.5">
                            <h3 :title="entry.name" class="text-sm font-medium text-white truncate">{{ entry.name }}</h3>
                            <span class="text-[10px] uppercase tracking-wide px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-400 shrink-0">
                                {{ entry.category }}
                            </span>
                        </div>
                        <p class="text-xs text-zinc-500 mb-3 flex-1 line-clamp-3">{{ entry.description }}</p>
                        <AppButton size="sm" :loading="useLoading === entry.id"
                            :disabled="!!useLoading && useLoading !== entry.id"
                            @click="useTemplate(entry)">
                            {{ useLoading === entry.id ? 'Cloning…' : 'Use template' }}
                        </AppButton>
                    </div>
                </div>
            </div>
        </div>
    </AppModal>
</template>

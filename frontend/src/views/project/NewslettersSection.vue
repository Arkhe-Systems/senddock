<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useToastStore } from '@/stores/toast'
import { useNewsletterStore, type Newsletter } from '@/stores/newsletters'
import type { Project } from '@/stores/projects'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import AppModal from '@/components/ui/AppModal.vue'
import AppConfirmModal from '@/components/ui/AppConfirmModal.vue'
import AppCopyId from '@/components/ui/AppCopyId.vue'

const props = defineProps<{ project: Project }>()
const toast = useToastStore()
const newsletterStore = useNewsletterStore()

const newsletters = computed(() => newsletterStore.newsletters(props.project.id))

const showModal = ref(false)
const editing = ref<Newsletter | null>(null)
const saving = ref(false)
const name = ref('')
const description = ref('')
const nameError = ref('')

const toDelete = ref<Newsletter | null>(null)
const deleting = ref(false)

function openCreate() {
    editing.value = null
    name.value = ''
    description.value = ''
    nameError.value = ''
    showModal.value = true
}

function openEdit(newsletter: Newsletter) {
    editing.value = newsletter
    name.value = newsletter.name
    description.value = newsletter.description
    nameError.value = ''
    showModal.value = true
}

async function save() {
    if (!name.value.trim()) {
        nameError.value = 'Name is required'
        return
    }
    nameError.value = ''
    saving.value = true
    try {
        if (editing.value) {
            await newsletterStore.updateNewsletter(props.project.id, editing.value.id, name.value.trim(), description.value.trim())
            toast.success('Newsletter updated')
        } else {
            await newsletterStore.createNewsletter(props.project.id, name.value.trim(), description.value.trim())
            toast.success('Newsletter created')
        }
        showModal.value = false
    } catch (e: any) {
        toast.error(e.message || 'Failed to save newsletter')
    } finally {
        saving.value = false
    }
}

async function confirmDelete() {
    if (!toDelete.value) return
    deleting.value = true
    try {
        await newsletterStore.deleteNewsletter(props.project.id, toDelete.value.id)
        toast.success('Newsletter deleted')
        toDelete.value = null
    } catch (e: any) {
        toast.error(e.message || 'Failed to delete newsletter')
    } finally {
        deleting.value = false
    }
}

onMounted(() => {
    newsletterStore.fetchNewsletters(props.project.id)
})
</script>

<template>
    <div>
        <div class="flex flex-wrap items-center justify-between gap-3 mb-6">
            <div>
                <h1 class="text-xl font-semibold text-white">Newsletters</h1>
                <p class="text-sm text-zinc-400 mt-1">Publications your subscribers can join and leave individually.</p>
            </div>
            <AppButton size="sm" @click="openCreate">+ New Newsletter</AppButton>
        </div>

        <div v-if="newsletters.length > 0" class="bg-zinc-900 border border-zinc-800 rounded-lg overflow-x-auto">
            <table class="w-full min-w-[480px]">
                <thead>
                    <tr class="border-b border-zinc-800">
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">Name</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">Description</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">Active members</th>
                        <th class="text-left px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">ID</th>
                        <th class="text-right px-4 py-3 text-xs font-medium text-zinc-300 uppercase tracking-wide">Actions</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="newsletter in newsletters" :key="newsletter.id" class="border-b border-zinc-800 last:border-0 hover:bg-zinc-850/50 transition">
                        <td class="px-4 py-3 text-sm text-white">{{ newsletter.name }}</td>
                        <td class="px-4 py-3 text-sm text-zinc-300 max-w-xs truncate">{{ newsletter.description || '—' }}</td>
                        <td class="px-4 py-3 text-sm text-zinc-300">{{ newsletter.active_count }}</td>
                        <td class="px-4 py-3"><AppCopyId :value="newsletter.id" /></td>
                        <td class="px-4 py-3 text-right space-x-3">
                            <button @click="openEdit(newsletter)" class="text-xs text-zinc-400 hover:text-white transition cursor-pointer">Edit</button>
                            <button @click="toDelete = newsletter" class="text-xs text-zinc-400 hover:text-red-400 transition cursor-pointer">Delete</button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <div v-else class="bg-zinc-900 border border-zinc-800 rounded-lg p-8 text-center">
            <p class="text-zinc-300 mb-2">No newsletters yet.</p>
            <p class="text-zinc-400 text-sm">Create one to let subscribers join specific publications — and leave one without leaving your whole list.</p>
        </div>

        <AppModal :show="showModal" :title="editing ? 'Edit newsletter' : 'New newsletter'" @close="showModal = false">
            <form @submit.prevent="save" class="space-y-4">
                <AppInput v-model="name" label="Name" placeholder="Dev Tips" :error="nameError" />
                <AppInput v-model="description" label="Description" placeholder="Weekly software development digest" />
                <div class="flex justify-end pt-2">
                    <AppButton size="md" :loading="saving">{{ editing ? 'Save' : 'Create newsletter' }}</AppButton>
                </div>
            </form>
        </AppModal>

        <AppConfirmModal
            :show="!!toDelete"
            title="Delete newsletter"
            :message="toDelete ? `Delete newsletter '${toDelete.name}'? Memberships are removed; subscribers stay in the project. Unsubscribe links already sent for it will fall back to a project-wide unsubscribe.` : ''"
            confirmLabel="Delete"
            danger
            :loading="deleting"
            @confirm="confirmDelete"
            @cancel="toDelete = null" />
    </div>
</template>

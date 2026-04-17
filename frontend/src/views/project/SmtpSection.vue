<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { api } from '@/api/client'
import { useToastStore } from '@/stores/toast'
import type { Project } from '@/stores/projects'
import AppInput from '@/components/ui/AppInput.vue'
import AppButton from '@/components/ui/AppButton.vue'

const props = defineProps<{ project: Project }>()
const emit = defineEmits<{ updated: [] }>()

const toast = useToastStore()
const loading = ref(false)
const testLoading = ref(false)

const smtpHost = ref('')
const smtpPort = ref('')
const smtpUser = ref('')
const smtpPassword = ref('')
const fromName = ref('')
const fromEmail = ref('')

const hasSmtpConfig = computed(() => !!props.project.smtp_host)

onMounted(() => {
    smtpHost.value = props.project.smtp_host ?? ''
    smtpPort.value = props.project.smtp_port?.toString() ?? ''
    smtpUser.value = props.project.smtp_user ?? ''
    fromName.value = props.project.from_name ?? ''
    fromEmail.value = props.project.from_email ?? ''
})

async function handleSave() {
    if (!smtpHost.value || !smtpPort.value || !smtpUser.value || !smtpPassword.value) {
        toast.error('All SMTP fields are required')
        return
    }

    loading.value = true
    try {
        await api(`/projects/${props.project.id}/smtp`, {
            method: 'PUT',
            body: {
                smtp_host: smtpHost.value,
                smtp_port: parseInt(smtpPort.value),
                smtp_user: smtpUser.value,
                smtp_password: smtpPassword.value,
                from_name: fromName.value,
                from_email: fromEmail.value,
            },
        })
        toast.success('SMTP settings saved')
        emit('updated')
    } catch (e: any) {
        toast.error(e.message || 'Failed to save SMTP settings')
    } finally {
        loading.value = false
    }
}

async function handleTest() {
    testLoading.value = true
    try {
        const res = await api<{ message: string }>(`/projects/${props.project.id}/smtp/test`, {
            method: 'POST',
        })
        toast.success(res.message)
    } catch (e: any) {
        toast.error(e.message || 'SMTP test failed')
    } finally {
        testLoading.value = false
    }
}
</script>

<template>
    <div>
        <h1 class="text-2xl font-bold text-white mb-6">SMTP Settings</h1>

        <form @submit.prevent="handleSave" class="max-w-lg space-y-4">
            <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-6 space-y-4">
                <h2 class="text-sm font-medium text-white mb-2">Server Configuration</h2>

                <div class="grid grid-cols-3 gap-4">
                    <div class="col-span-2">
                        <AppInput v-model="smtpHost" label="SMTP Host" placeholder="smtp.gmail.com" required />
                    </div>
                    <div>
                        <AppInput v-model="smtpPort" label="Port" placeholder="587" type="number" required />
                    </div>
                </div>

                <AppInput v-model="smtpUser" label="Username" placeholder="you@gmail.com" required />
                <AppInput v-model="smtpPassword" label="Password" type="password" placeholder="App password or SMTP key" required />
            </div>

            <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-6 space-y-4">
                <h2 class="text-sm font-medium text-white mb-2">Sender Identity</h2>
                <AppInput v-model="fromName" label="From Name" placeholder="My Newsletter" />
                <AppInput v-model="fromEmail" label="From Email" placeholder="noreply@mydomain.com" />
                <p class="text-xs text-zinc-500">The email address recipients will see as the sender.</p>
            </div>

            <div class="flex gap-3">
                <AppButton :loading="loading">
                    {{ loading ? 'Saving...' : 'Save Settings' }}
                </AppButton>
                <button v-if="hasSmtpConfig" type="button" @click="handleTest" :disabled="testLoading"
                    class="flex-1 py-2 text-sm font-medium border border-zinc-700 text-zinc-300 rounded-lg hover:bg-zinc-800 transition cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed">
                    {{ testLoading ? 'Testing...' : 'Test Connection' }}
                </button>
            </div>
        </form>
    </div>
</template>

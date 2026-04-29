<script setup lang="ts">
import AppModal from './AppModal.vue'
import AppButton from './AppButton.vue'

withDefaults(defineProps<{
    show?: boolean
    title?: string
    message?: string
    confirmLabel?: string
    cancelLabel?: string
    danger?: boolean
    loading?: boolean
}>(), {
    title: 'Are you sure?',
    confirmLabel: 'Confirm',
    cancelLabel: 'Cancel',
})

const emit = defineEmits<{
    confirm: []
    cancel: []
}>()
</script>

<template>
    <AppModal :show="show" :title="title" @close="emit('cancel')">
        <div class="space-y-5">
            <p v-if="message" class="text-sm text-zinc-300 leading-relaxed">{{ message }}</p>
            <slot />
            <div class="flex gap-2 pt-1">
                <AppButton type="button" variant="ghost" size="sm" class="flex-1" :disabled="loading" @click="emit('cancel')">
                    {{ cancelLabel }}
                </AppButton>
                <AppButton type="button" :variant="danger ? 'danger' : 'primary'" size="sm" :loading="loading" :disabled="loading" class="flex-1" @click="emit('confirm')">
                    {{ confirmLabel }}
                </AppButton>
            </div>
        </div>
    </AppModal>
</template>

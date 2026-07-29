<script setup lang="ts">
import AppToast from '@/components/ui/AppToast.vue'
import { useSessionWatch } from '@/composables/useSessionWatch'

useSessionWatch()
</script>

<template>
  <RouterView />
  <AppToast />
</template>

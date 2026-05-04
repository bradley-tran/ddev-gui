<script setup lang="ts">
import { useAppStore } from '@/stores/app'
import ToastItem from './ToastItem.vue'

const appStore = useAppStore()
</script>

<template>
  <div id="toastContainer">
    <ToastItem
      v-for="toast in appStore.toasts"
      :id="toast.id"
      :key="toast.id"
      :message="toast.message"
      :type="toast.type"
      :duration="toast.duration"
      @dismiss="appStore.removeToast"
    />
  </div>
</template>
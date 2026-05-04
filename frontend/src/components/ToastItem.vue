<script setup lang="ts">
import { CircleCheckIcon, CircleXIcon, InfoIcon } from '@lucide/vue'
import { onBeforeUnmount, onMounted, ref } from 'vue'
import type { ToastType } from '@/lib/types'

const props = defineProps<{
  id: string
  message: string
  type: ToastType
  duration: number
}>()

const emit = defineEmits<{
  dismiss: [id: string]
}>()

const toastRef = ref<HTMLDivElement | null>(null)
let timerId: number | null = null
let dismissed = false

function dismiss() {
  if (dismissed) return
  dismissed = true

  const element = toastRef.value
  if (!element) {
    emit('dismiss', props.id)
    return
  }

  element.classList.add('toast-exit')
  element.addEventListener(
    'animationend',
    () => {
      emit('dismiss', props.id)
    },
    { once: true },
  )
}

onMounted(() => {
  timerId = window.setTimeout(dismiss, props.duration)
})

onBeforeUnmount(() => {
  if (timerId !== null) {
    clearTimeout(timerId)
  }
})
</script>

<template>
  <div ref="toastRef" class="toast" :class="`toast-${props.type}`" role="alert" aria-live="polite" @click="dismiss">
    <span class="toast-icon">
      <CircleCheckIcon v-if="props.type === 'success'" :size="16" :stroke-width="2.5" />
      <CircleXIcon v-else-if="props.type === 'error'" :size="16" :stroke-width="2.5" />
      <InfoIcon v-else :size="16" :stroke-width="2.5" />
    </span>
    <span class="toast-msg">{{ props.message }}</span>
  </div>
</template>
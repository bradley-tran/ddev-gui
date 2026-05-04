<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

const props = withDefaults(defineProps<{
  title: string
  wide?: boolean
}>(), {
  wide: false,
})

const emit = defineEmits<{
  close: []
}>()

const overlayRef = ref<HTMLDivElement | null>(null)
const titleId = `app-modal-title-${Math.random().toString(36).slice(2)}`

function closeModal() {
  emit('close')
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    closeModal()
  }
}

function handleOverlayClick(event: MouseEvent) {
  if (event.target === overlayRef.value) {
    closeModal()
  }
}

onMounted(() => {
  document.addEventListener('keydown', onKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <div
    ref="overlayRef"
    class="flu-modal-overlay show"
    role="dialog"
    aria-modal="true"
    :aria-labelledby="titleId"
    @click="handleOverlayClick"
  >
    <div class="flu-modal" :class="{ 'flu-modal-wide': props.wide }">
      <div class="flu-modal-header">
        <h2 :id="titleId">{{ props.title }}</h2>
        <button
          class="flu-modal-close"
          aria-label="Close"
          title="Close"
          type="button"
          @click="closeModal"
        >
          &times;
        </button>
      </div>
      <div class="flu-modal-body">
        <slot />
      </div>
      <div v-if="$slots.footer" class="flu-modal-footer">
        <slot name="footer" />
      </div>
    </div>
  </div>
</template>
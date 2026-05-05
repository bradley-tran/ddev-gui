<script setup lang="ts">
import Modal from '@/components/Modal.vue'
import Spinner from '@/components/Spinner.vue'
import { useTranslation } from '@/lib/i18n'

const props = withDefaults(defineProps<{
  title: string
  message: string
  confirmText?: string
  pending?: boolean
}>(), {
  confirmText: '',
  pending: false,
})

const emit = defineEmits<{
  close: []
  confirm: []
}>()

const { t } = useTranslation()

function handleClose() {
  if (props.pending) return
  emit('close')
}

function handleConfirm() {
  if (props.pending) return
  emit('confirm')
}
</script>

<template>
  <Modal :title="props.title" @close="handleClose">
    <p class="confirm-delete-modal-message">{{ props.message }}</p>

    <template #footer>
      <button
        type="button"
        class="flu-btn flu-btn-ghost confirm-delete-modal-cancel"
        :disabled="props.pending"
        @click="handleClose"
      >
        {{ t('general.cancel') }}
      </button>
      <button
        type="button"
        class="flu-btn flu-btn-danger confirm-delete-modal-confirm"
        :disabled="props.pending"
        @click="handleConfirm"
      >
        <Spinner v-if="props.pending" />
        <span>{{ props.confirmText || t('general.delete') }}</span>
      </button>
    </template>
  </Modal>
</template>

<style scoped>
.confirm-delete-modal-message {
  margin: 0;
  white-space: pre-line;
}

.confirm-delete-modal-confirm {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}
</style>
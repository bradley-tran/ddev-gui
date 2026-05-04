<script setup lang="ts">
import { ref } from 'vue'
import Modal from '@/components/Modal.vue'
import Spinner from '@/components/Spinner.vue'
import { useTranslation } from '@/lib/i18n'
import { DdevService as DdevApi } from '@/lib/wails'
import { useAppStore } from '@/stores/app'

const props = defineProps<{
  projectName: string
}>()

const emit = defineEmits<{
  close: []
  created: []
}>()

const appStore = useAppStore()
const { t } = useTranslation()

const running = ref(false)
const snapshotName = ref('')

async function handleSubmit() {
  if (running.value) return

  running.value = true
  appStore.appLog(`Creating snapshot for ${props.projectName}...`, 'info')

  try {
    await DdevApi.snapshotCreate(props.projectName, snapshotName.value.trim())
    appStore.appLog(`Snapshot created for ${props.projectName}`, 'success')
    appStore.showToast('Snapshot created', 'success')
    emit('created')
    emit('close')
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error)
    appStore.appLog(`Snapshot create failed: ${message}`, 'error')
    appStore.showToast('Snapshot create failed', 'error')
  } finally {
    running.value = false
  }
}
</script>

<template>
  <Modal :title="t('detail.snapshots.createTitle')" @close="emit('close')">
    <div class="flu-field">
      <label class="flu-label" for="createSnapshotName">{{ t('general.name') }}</label>
      <input
        id="createSnapshotName"
        v-model="snapshotName"
        class="flu-input snapshot-create-modal-input"
        :disabled="running"
        @keydown.enter.prevent="handleSubmit"
      >
      <p class="text-muted">
        {{ t('detail.snapshots.nameHelp') }}
      </p>
    </div>

    <template #footer>
      <button type="button" class="flu-btn flu-btn-ghost" :disabled="running" @click="emit('close')">
        {{ t('general.cancel') }}
      </button>
      <button
        type="button"
        class="flu-btn flu-btn-accent snapshot-create-modal-submit"
        :disabled="running"
        @click="handleSubmit"
      >
        <template v-if="running">
          <Spinner />
          {{ t('detail.snapshots.create') }}
        </template>
        <template v-else>
          {{ t('detail.snapshots.create') }}
        </template>
      </button>
    </template>
  </Modal>
</template>
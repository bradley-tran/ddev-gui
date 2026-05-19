<script setup lang="ts">
import { ref, watch } from 'vue'
import Modal from '@/components/Modal.vue'
import Select from '@/components/Select.vue'
import Spinner from '@/components/Spinner.vue'
import { useTranslation } from '@/lib/i18n'
import type { DdevProject } from '@/lib/types'
import { pickProjectValue } from '@/lib/utils'
import { DdevService as DdevApi } from '@/lib/wails'
import { useAppStore } from '@/stores/app'

const PHP_VERSIONS = ['8.4', '8.3', '8.2', '8.1', '8.0', '7.4', '7.3', '7.2', '7.1', '7.0', '5.6']
const NODEJS_VERSIONS = ['22', '20', '18', '16', '14', '12']
const PROJECT_TYPES = [
  'backdrop', 'cakephp', 'codeigniter', 'craftcms', 'django4', 'drupal',
  'drupal6', 'drupal7', 'drupal8', 'drupal9', 'drupal10', 'drupal11',
  'laravel', 'magento', 'magento2', 'php', 'shopware6', 'silverstripe',
  'symfony', 'typo3', 'wordpress',
]

const props = defineProps<{
  projectName: string
  project: DdevProject | null
}>()

const emit = defineEmits<{
  close: []
  modified: []
}>()

const appStore = useAppStore()
const { t } = useTranslation()

const running = ref(false)
const phpVersion = ref('8.3')
const nodejsVersion = ref('20')
const projectType = ref('php')
const docroot = ref('')

watch(
  () => props.project,
  (project) => {
    syncForm(project)
  },
  { immediate: true },
)

function syncForm(project: DdevProject | null) {
  phpVersion.value = String(pickProjectValue(project, ['php_version', 'phpversion']) ?? '') || '8.3'
  nodejsVersion.value = String(pickProjectValue(project, ['nodejs_version']) ?? '') || '20'
  projectType.value = String(pickProjectValue(project, ['type', 'projecttype']) ?? '') || 'php'
  docroot.value = String(pickProjectValue(project, ['docroot']) ?? '')
}

async function handleSubmit() {
  running.value = true
  appStore.appLog(`Modifying project ${props.projectName}...`, 'info')

  try {
    await DdevApi.modifyProject(
      props.projectName,
      phpVersion.value,
      nodejsVersion.value,
      projectType.value,
      docroot.value,
    )
    appStore.appLog(`Project ${props.projectName} modified`, 'success')
    appStore.showToast(`Project ${props.projectName} modified`, 'success')
    emit('modified')
    emit('close')
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error)
    appStore.appLog(`Modify failed: ${message}`, 'error')
    appStore.showToast('Modify failed', 'error')
  } finally {
    running.value = false
  }
}
</script>

<template>
  <Modal :title="t('detail.modify.title')" @close="emit('close')">
    <div class="flu-field">
      <label class="flu-label" for="modifyPhpVersion">{{ t('detail.overview.phpVersion') }}</label>
      <Select
        id="modifyPhpVersion"
        :model-value="phpVersion"
        :options="PHP_VERSIONS.map((value) => ({ value, label: value }))"
        :disabled="running"
        @update:model-value="phpVersion = $event"
      />
    </div>
    <div class="flu-field">
      <label class="flu-label" for="modifyNodejsVersion">{{ t('detail.overview.nodejs') }}</label>
      <Select
        id="modifyNodejsVersion"
        :model-value="nodejsVersion"
        :options="NODEJS_VERSIONS.map((value) => ({ value, label: value }))"
        :disabled="running"
        @update:model-value="nodejsVersion = $event"
      />
    </div>
    <div class="flu-field">
      <label class="flu-label" for="modifyProjectType">{{ t('detail.overview.type') }}</label>
      <Select
        id="modifyProjectType"
        :model-value="projectType"
        :options="PROJECT_TYPES.map((value) => ({ value, label: value }))"
        :disabled="running"
        @update:model-value="projectType = $event"
      />
    </div>
    <div class="flu-field">
      <label class="flu-label" for="modifyDocroot">{{ t('detail.overview.docroot') }}</label>
      <input
        id="modifyDocroot"
        v-model="docroot"
        class="flu-input"
        :disabled="running"
      >
    </div>

    <template #footer>
      <button type="button" class="flu-btn flu-btn-ghost" :disabled="running" @click="emit('close')">
        {{ t('general.cancel') }}
      </button>
      <button type="button" class="flu-btn flu-btn-accent" :disabled="running" @click="handleSubmit">
        <template v-if="running">
          <Spinner />
          {{ t('general.saving') }}
        </template>
        <template v-else>
          {{ t('detail.modify.apply') }}
        </template>
      </button>
    </template>
  </Modal>
</template>

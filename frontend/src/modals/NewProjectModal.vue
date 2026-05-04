<script setup lang="ts">
import { ref } from 'vue'
import Modal from '@/components/Modal.vue'
import Select from '@/components/Select.vue'
import { useTranslation } from '@/lib/i18n'
import { ConfigService, DdevService } from '@/lib/wails'
import { useAppStore } from '@/stores/app'

const emit = defineEmits<{
  close: []
}>()

const PROJECT_TYPES = [
  'backdrop', 'cakephp', 'codeigniter', 'craftcms',
  'drupal', 'drupal7', 'drupal10', 'drupal11', 'drupal12',
  'generic', 'laravel',
  'php', 'shopware6', 'silverstripe', 'symfony', 'typo3', 'wordpress',
]

const DOCROOT_DEFAULTS: Record<string, string> = {
  backdrop: '',
  cakephp: 'webroot',
  codeigniter: 'public',
  craftcms: 'web',
  drupal: 'web',
  drupal7: '',
  drupal10: 'web',
  drupal11: 'web',
  drupal12: 'web',
  generic: '',
  laravel: 'public',
  php: '',
  shopware6: 'public',
  silverstripe: 'public',
  symfony: 'public',
  typo3: 'public',
  wordpress: '',
}

const PHP_VERSIONS = ['8.5', '8.4', '8.3', '8.2', '8.1', '8.0', '7.4']

function getDefaultDocroot(projectType: string): string {
  return DOCROOT_DEFAULTS[projectType] ?? ''
}

const appStore = useAppStore()
const { t } = useTranslation()

const name = ref('')
const type = ref('drupal11')
const docroot = ref(getDefaultDocroot('drupal11'))
const docrootManual = ref(false)
const phpVersion = ref('8.3')
const gitRepo = ref('')
const submitting = ref(false)
const error = ref('')

function handleTypeChange(newType: string) {
  type.value = newType
  if (!docrootManual.value) {
    docroot.value = getDefaultDocroot(newType)
  }
}

function handleDocrootChange(value: string) {
  docroot.value = value
  docrootManual.value = true
}

async function handleSubmit(event: Event) {
  event.preventDefault()
  error.value = ''

  if (!name.value.trim()) {
    error.value = t('newProject.nameRequired')
    return
  }

  submitting.value = true
  appStore.appLog(`Creating project "${name.value}"...`, 'info')

  try {
    const trimmedName = name.value.trim()
    const trimmedRepo = gitRepo.value.trim()
    if (trimmedRepo) {
      appStore.appLog(`Cloning ${trimmedRepo}...`, 'info')
      await DdevService.cloneRepo(trimmedName, trimmedRepo)
    }

    await DdevService.configureProject(trimmedName, type.value, docroot.value.trim(), phpVersion.value)
    await DdevService.start(trimmedName)
    appStore.appLog(`Project "${trimmedName}" created and started.`, 'success')
    appStore.showToast(`Project "${trimmedName}" created`, 'success')

    const isInitialized = Boolean(trimmedRepo)
    await ConfigService.setProjectConfig(trimmedName, 'initialized', isInitialized)
    appStore.patchConfig({
      projects: {
        ...(appStore.config.projects ?? {}),
        [trimmedName]: { initialized: isInitialized },
      },
    })

    appStore.setProjectsJSON(await DdevService.listJSON())
    emit('close')
  } catch (caughtError) {
    const message = caughtError instanceof Error ? caughtError.message : String(caughtError)
    error.value = message
    appStore.appLog(`Failed to create project: ${message}`, 'error')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <Modal :title="t('newProject.title')" @close="emit('close')">
    <template #footer>
      <button class="flu-btn flu-btn-ghost" type="button" :disabled="submitting" @click="emit('close')">
        {{ t('general.cancel') }}
      </button>
      <button class="flu-btn flu-btn-accent" type="submit" form="newProjectForm" :disabled="submitting">
        {{ submitting ? t('newProject.creating') : t('newProject.create') }}
      </button>
    </template>

    <form id="newProjectForm" class="flu-form" @submit="handleSubmit">
      <div v-if="error" class="form-error">{{ error }}</div>

      <div class="flu-form-group">
        <label for="projName">{{ t('newProject.fieldName') }}</label>
        <input
          id="projName"
          v-model="name"
          class="flu-input"
          type="text"
          placeholder="my-project"
          autofocus
        >
      </div>

      <div class="flu-form-group">
        <label for="projGitRepo">
          {{ t('newProject.fieldGitRepo') }}
          <span class="text-muted">{{ t('newProject.optional') }}</span>
        </label>
        <input
          id="projGitRepo"
          v-model="gitRepo"
          class="flu-input"
          type="text"
          placeholder="https://github.com/user/repo.git"
        >
      </div>

      <div class="flu-form-group">
        <label for="projType">{{ t('newProject.fieldType') }}</label>
        <Select
          id="projType"
          :model-value="type"
          :options="PROJECT_TYPES.map((projectType) => ({ value: projectType, label: projectType }))"
          @update:model-value="handleTypeChange"
        />
      </div>

      <div class="flu-form-group">
        <label for="projDocroot">
          {{ t('newProject.fieldDocroot') }}
          <span v-if="getDefaultDocroot(type)" class="text-muted">
            {{ t('newProject.docrootDefault', { val: getDefaultDocroot(type) }) }}
          </span>
        </label>
        <input
          id="projDocroot"
          :value="docroot"
          class="flu-input"
          type="text"
          :placeholder="getDefaultDocroot(type) || t('newProject.docrootPlaceholder')"
          @input="handleDocrootChange(($event.target as HTMLInputElement).value)"
        >
      </div>

      <div class="flu-form-group">
        <label for="projPhpVersion">{{ t('newProject.fieldPhp') }}</label>
        <Select
          id="projPhpVersion"
          v-model="phpVersion"
          :options="PHP_VERSIONS.map((version) => ({ value: version, label: version }))"
        />
      </div>
    </form>
  </Modal>
</template>
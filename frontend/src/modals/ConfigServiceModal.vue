<script setup lang="ts">
import { ref, watch } from 'vue'
import Modal from '@/components/Modal.vue'
import Spinner from '@/components/Spinner.vue'
import { useTranslation } from '@/lib/i18n'
import type { DdevProject } from '@/lib/types'
import { DdevService as DdevApi } from '@/lib/wails'
import { useAppStore } from '@/stores/app'

const props = defineProps<{
  projectName: string
  project: DdevProject | null
}>()

const emit = defineEmits<{
  close: []
  configured: []
}>()

const appStore = useAppStore()
const { t } = useTranslation()

const running = ref(false)
const webPort = ref('')
const dbPort = ref('')
const xdebugEnabled = ref(false)
const xhprofEnabled = ref(false)
const xhguiEnabled = ref(false)

watch(
  () => props.project,
  (project) => {
    syncForm(project)
  },
  { immediate: true },
)

watch(xhguiEnabled, (enabled) => {
  if (enabled) {
    xhprofEnabled.value = true
  }
})

watch(xhprofEnabled, (enabled) => {
  if (!enabled) {
    xhguiEnabled.value = false
  }
})

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}

function pickProjectValue(project: DdevProject | null, keys: string[]): unknown {
  if (!project) return undefined

  const projectKeys = Object.keys(project)
  for (const key of keys) {
    const match = projectKeys.find((candidate) => candidate.toLowerCase() === key.toLowerCase())
    if (match) {
      return project[match]
    }
  }

  return undefined
}

function normalizeBool(value: unknown): boolean {
  if (typeof value === 'boolean') return value
  if (typeof value === 'number') return value !== 0
  if (typeof value !== 'string') return false

  const normalized = value.trim().toLowerCase()
  return ['1', 'true', 'on', 'yes', 'enabled', 'running', 'ok', 'healthy'].includes(normalized)
}

function normalizeToggleStatus(value: unknown): boolean | null {
  if (value == null) return null

  const normalized = String(value).trim().toLowerCase()
  if (normalized === '') return null
  if (['enabled', 'on', 'true', '1', 'running', 'ok', 'healthy'].includes(normalized)) return true
  if (['disabled', 'off', 'false', '0', 'stopped', 'down', 'inactive'].includes(normalized)) return false

  return null
}

function serviceStatus(project: DdevProject | null, serviceName: string): string {
  const services = project?.services
  if (!services || typeof services !== 'object') return ''
  const service = services[serviceName]
  if (!isRecord(service)) return ''
  return String(service.status ?? '').trim().toLowerCase()
}

function syncForm(project: DdevProject | null) {
  // Keep ports empty by default. Non-empty values represent explicit overrides.
  webPort.value = ''
  dbPort.value = ''

  xdebugEnabled.value = normalizeBool(pickProjectValue(project, ['xdebug_enabled']))

  const xhprofMode = String(pickProjectValue(project, ['xhprof_mode']) ?? '').trim().toLowerCase()
  const xhguiStatus = pickProjectValue(project, ['xhgui_status'])
  const xhguiRunning = serviceStatus(project, 'xhgui') === 'running'
  const xhguiStatusToggle = normalizeToggleStatus(xhguiStatus)

  // Prefer explicit xhgui_status when available, because xhprof_mode can be stale.
  if (xhguiStatusToggle !== null) {
    xhguiEnabled.value = xhguiStatusToggle
  } else {
    xhguiEnabled.value = xhguiRunning || xhprofMode === 'xhgui'
  }

  xhprofEnabled.value = xhguiEnabled.value || xhprofMode === 'prepend'

  if (xhguiEnabled.value) {
    xhprofEnabled.value = true
  }
}

function isValidPort(portValue: string): boolean {
  if (portValue === '') return true
  if (!/^[0-9]+$/.test(portValue)) return false

  const portNum = Number(portValue)
  return Number.isInteger(portNum) && portNum >= 1 && portNum <= 65535
}

async function handleSave() {
  const normalizedWebPort = webPort.value.trim()
  const normalizedDbPort = dbPort.value.trim()

  if (!isValidPort(normalizedWebPort) || !isValidPort(normalizedDbPort)) {
    appStore.appLog(t('detail.services.invalidPort'), 'error')
    appStore.showToast(t('detail.services.invalidPort'), 'error')
    return
  }

  running.value = true
  try {
    await DdevApi.configureServices(
      props.projectName,
      normalizedWebPort,
      normalizedDbPort,
      xdebugEnabled.value,
      xhprofEnabled.value,
      xhguiEnabled.value,
    )

    appStore.appLog(`Service configuration saved for ${props.projectName}`, 'success')
    appStore.showToast(t('detail.services.configSaved'), 'success')
    emit('configured')
    emit('close')
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error)
    appStore.appLog(`Service config save failed: ${message}`, 'error')
    appStore.showToast(t('detail.services.configSaveFailed'), 'error')
  } finally {
    running.value = false
  }
}
</script>

<template>
  <Modal :title="t('detail.services.configTitle')" @close="emit('close')">
    <div class="flu-form">
      <div class="flu-form-group">
        <label class="flu-field-label" for="svcWebPort">{{ t('detail.services.webPort') }}</label>
        <input
          id="svcWebPort"
          v-model="webPort"
          type="text"
          inputmode="numeric"
          class="flu-input"
          :disabled="running"
          placeholder="(unchanged)"
          autocomplete="off"
        >
      </div>

      <div class="flu-form-group">
        <label class="flu-field-label" for="svcDbPort">{{ t('detail.services.dbPort') }}</label>
        <input
          id="svcDbPort"
          v-model="dbPort"
          type="text"
          inputmode="numeric"
          class="flu-input"
          :disabled="running"
          placeholder="(unchanged)"
          autocomplete="off"
        >
      </div>

      <hr style="border: none; border-top: 1px solid var(--surface-stroke); margin: 0">

      <div class="flu-form-group">
        <label class="flu-toggle-label" for="svcXdebug">
          <input
            id="svcXdebug"
            v-model="xdebugEnabled"
            type="checkbox"
            class="flu-toggle"
            :disabled="running"
          >
          <span>{{ t('detail.services.xdebug') }}</span>
        </label>
      </div>

      <div class="flu-form-group">
        <label class="flu-toggle-label" for="svcXhprof">
          <input
            id="svcXhprof"
            v-model="xhprofEnabled"
            type="checkbox"
            class="flu-toggle"
            :disabled="running"
          >
          <span>{{ t('detail.services.xhprof') }}</span>
        </label>
      </div>

      <div class="flu-form-group">
        <label class="flu-toggle-label" for="svcXhgui">
          <input
            id="svcXhgui"
            v-model="xhguiEnabled"
            type="checkbox"
            class="flu-toggle"
            :disabled="running"
          >
          <span>{{ t('detail.services.xhgui') }}</span>
        </label>
        <p class="text-muted" style="margin-top: 0.25rem; font-size: 0.85em">
          {{ t('detail.services.xhguiHint') }}
        </p>
      </div>
    </div>

    <template #footer>
      <button type="button" class="flu-btn flu-btn-ghost" :disabled="running" @click="emit('close')">
        {{ t('general.cancel') }}
      </button>
      <button
        type="button"
        class="flu-btn flu-btn-accent service-config-modal-submit"
        :disabled="running"
        @click="handleSave"
      >
        <template v-if="running">
          <Spinner />
          {{ t('general.saving') }}
        </template>
        <template v-else>
          {{ t('general.save') }}
        </template>
      </button>
    </template>
  </Modal>
</template>

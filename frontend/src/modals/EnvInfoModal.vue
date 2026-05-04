<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import Modal from '@/components/Modal.vue'
import Spinner from '@/components/Spinner.vue'
import { useTranslation } from '@/lib/i18n'
import { ConfigService, DdevService, Runtime } from '@/lib/wails'
import { useAppStore } from '@/stores/app'

const emit = defineEmits<{
  close: []
  openSettings: []
}>()

const appStore = useAppStore()
const { t } = useTranslation()

const versionText = ref('')
const loading = ref(true)
const error = ref('')
const installing = ref(false)
const installProgress = ref('')
const telemetryOptIn = ref(appStore.config.ddevTelemetryOptIn ?? true)
const devMode = computed(() => appStore.config.devMode ?? false)
const isWslError = computed(() => Boolean(error.value) && /wslshell|wsl\.exe|wsl\b|pipe|distro/i.test(error.value))

const installProgressHandler = (...args: unknown[]) => {
  const message = typeof args[0] === 'string' ? args[0] : String(args[0])
  installProgress.value = message
  appStore.appLog(message, 'info')
}

onMounted(async () => {
  try {
    versionText.value = await DdevService.ddevInstalledVersion()
  } catch (caughtError) {
    const message = caughtError instanceof Error ? caughtError.message : String(caughtError)
    error.value = message
    appStore.appLog(`Failed to load environment info: ${message}`, 'error')
  } finally {
    loading.value = false
  }
})

onBeforeUnmount(() => {
  Runtime.off('ddev:output', installProgressHandler)
})

async function handleInstall() {
  installing.value = true
  installProgress.value = 'Connecting to GitHub…'
  appStore.appLog('Installing/updating DDEV...', 'info')
  Runtime.off('ddev:output', installProgressHandler)
  Runtime.on('ddev:output', installProgressHandler)

  try {
    await DdevService.installDdev()
    appStore.appLog('DDEV installer launched - please restart this app after installation finishes.', 'success')
    window.setTimeout(() => Runtime.quit(), 2000)
  } catch (caughtError) {
    const message = caughtError instanceof Error ? caughtError.message : String(caughtError)
    appStore.appLog(`DDEV install failed: ${message}`, 'error')
  } finally {
    Runtime.off('ddev:output', installProgressHandler)
    installing.value = false
    installProgress.value = ''
  }
}

async function handleDevModeToggle(enabled: boolean) {
  appStore.patchConfig({ devMode: enabled })
  try {
    await ConfigService.set('devMode', enabled)
  } catch {
    appStore.patchConfig({ devMode: !enabled })
  }
}

async function persistTelemetryOptIn(enabled: boolean) {
  try {
    await appStore.saveConfigValue('ddevTelemetryOptIn', enabled)
  } catch {
    telemetryOptIn.value = appStore.config.ddevTelemetryOptIn ?? true
  }
}

async function handleTelemetryOptInToggle(enabled: boolean) {
  telemetryOptIn.value = enabled
  await persistTelemetryOptIn(enabled)
}

async function handleClose() {
  if (typeof appStore.config.ddevTelemetryOptIn === 'undefined') {
    await persistTelemetryOptIn(telemetryOptIn.value)
  }

  emit('close')
}

function openSettings() {
  emit('openSettings')
}
</script>

<template>
  <Modal :title="t('env.title')" @close="handleClose">
    <template #footer>
      <button class="flu-btn flu-btn-ghost" type="button" :disabled="installing" @click="handleClose">
        {{ t('general.close') }}
      </button>
      <button
        v-if="error && isWslError"
        class="flu-btn flu-btn-accent"
        type="button"
        :disabled="installing"
        @click="openSettings"
      >
        {{ t('env.openSettings') }}
      </button>
      <button
        v-else-if="error && !isWslError"
        class="flu-btn flu-btn-accent"
        type="button"
        :disabled="installing"
        @click="handleInstall"
      >
        <template v-if="installing"><Spinner /> {{ t('env.installing') }}</template>
        <template v-else>{{ t('env.installDdev') }}</template>
      </button>
    </template>

    <div v-if="loading" style="text-align: center; padding: 2rem">
      <Spinner /> {{ t('general.loading') }}
    </div>
    <template v-else-if="error">
      <div class="form-error">{{ isWslError ? error : t('env.ddevMissing') }}</div>
      <div
        v-if="installing && installProgress"
        style="margin-top: 0.75rem; font-size: 0.85em; color: var(--text-secondary); display: flex; align-items: center; gap: 0.5rem"
      >
        <Spinner /><span>{{ installProgress }}</span>
      </div>
    </template>
    <template v-else>
      <div class="detail-kv">
        <div class="detail-kv-item">
          <span class="kv-label">{{ t('env.ddevVersion') }}</span>
          <span class="kv-value">{{ versionText || t('env.notInstalled') }}</span>
        </div>
      </div>
      <div
        v-if="installing && installProgress"
        style="margin-top: 0.75rem; font-size: 0.85em; color: var(--text-secondary); display: flex; align-items: center; gap: 0.5rem"
      >
        <Spinner /><span>{{ installProgress }}</span>
      </div>
    </template>

    <hr style="border: none; border-top: 1px solid var(--surface-stroke); margin: var(--space-md) 0">

    <div class="flu-form-group">
      <label class="flu-toggle-label">
        <input
          type="checkbox"
          name="ddev-telemetry-opt-in"
          class="flu-toggle"
          :checked="telemetryOptIn"
          @change="handleTelemetryOptInToggle(($event.target as HTMLInputElement).checked)"
        >
        <span>{{ t('env.telemetryOptIn') }}</span>
      </label>
      <p class="text-muted" style="margin-top: 0.25rem; font-size: 0.85em">
        {{ t('env.telemetryOptInDesc') }}
      </p>
    </div>

    <div class="flu-form-group">
      <label class="flu-toggle-label">
        <input
          type="checkbox"
          name="dev-mode"
          class="flu-toggle"
          :checked="devMode"
          @change="handleDevModeToggle(($event.target as HTMLInputElement).checked)"
        >
        <span>{{ t('env.developerMode') }}</span>
      </label>
      <p class="text-muted" style="margin-top: 0.25rem; font-size: 0.85em">
        {{ t('env.developerModeDesc') }}
      </p>
    </div>
  </Modal>
</template>

<style scoped>
.flu-form-group {
  margin-bottom: var(--space-md);
}
</style>
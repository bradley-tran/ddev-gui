<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import Modal from '@/components/Modal.vue'
import Select from '@/components/Select.vue'
import { LOCALE_LABELS, useTranslation } from '@/lib/i18n'
import type { Locale } from '@/lib/i18n'
import type { AppConfig, BackendType, PreferredEditorType, SshConfig, ThemeType } from '@/lib/types'
import { coerceToBool } from '@/lib/utils'
import { ConfigService, DdevService } from '@/lib/wails'
import { useAppStore } from '@/stores/app'

const emit = defineEmits<{
  close: []
}>()

const ALL_BACKEND_OPTIONS: Array<{ value: BackendType; label: string }> = [
  { value: 'wsl', label: 'WSL (Windows Subsystem for Linux)' },
  { value: 'ssh', label: 'SSH (Remote Host)' },
  { value: 'local', label: 'Local (direct execution)' },
]

const THEME_OPTIONS: Array<{ value: ThemeType; label: string }> = [
  { value: 'default', label: 'Default' },
  { value: 'acrylic', label: 'Windows Acrylic' },
  { value: 'tabbed', label: 'Windows Tabbed' },
]

const EDITOR_OPTIONS: Array<{ value: PreferredEditorType; label: string }> = [
  { value: 'vscode', label: 'VS Code' },
  { value: 'phpstorm', label: 'PhpStorm' },
  { value: 'neovim', label: 'Neovim' },
  { value: 'sublime', label: 'Sublime Text' },
  { value: 'antigravity', label: 'Antigravity' },
]

const appStore = useAppStore()
const { t } = useTranslation()
const cfg = appStore.config

const devMode = computed(() => appStore.config.devMode ?? false)
const backendOptions = computed(() =>
  devMode.value ? ALL_BACKEND_OPTIONS : ALL_BACKEND_OPTIONS.filter((option) => option.value !== 'ssh'),
)
const isLinuxPlatform = document.body.classList.contains('platform-linux')

const openLinksInBrowser = ref(coerceToBool(cfg.openLinksInBrowser))
const preferredEditor = ref<PreferredEditorType>(cfg.preferredEditor ?? 'vscode')
const backend = ref<BackendType>(cfg.backend ?? 'wsl')
const theme = ref<ThemeType>(cfg.theme ?? 'default')
const locale = ref<Locale>((cfg.locale as Locale) ?? 'en')
const sshHost = ref(cfg.ssh?.host ?? '')
const sshPort = ref(cfg.ssh?.port ?? '22')
const sshUser = ref(cfg.ssh?.user ?? '')
const sshKey = ref(cfg.ssh?.keyPath ?? '')
const wslDistro = ref(cfg.wslDistro ?? '')
const wslDistros = ref<string[]>([])
const saving = ref(false)

watch(
  backend,
  async (nextBackend) => {
    if (nextBackend !== 'wsl') return
    try {
      const list = await DdevService.listWSLDistros()
      if (list?.length) {
        wslDistros.value = list
        const fallbackDistro = list[0]
        if (!fallbackDistro) return
        if (!wslDistro.value || !list.includes(wslDistro.value)) {
          wslDistro.value = list.find((distro) => distro.toUpperCase() === 'DDEV') ?? fallbackDistro
        }
      }
    } catch {
      return
    }
  },
  { immediate: true },
)

async function handleSave() {
  saving.value = true
  try {
    const sshCfg: SshConfig = {
      host: sshHost.value.trim(),
      port: sshPort.value.trim() || '22',
      user: sshUser.value.trim(),
      keyPath: sshKey.value.trim(),
    }

    const patch: Partial<AppConfig> = {
      openLinksInBrowser: openLinksInBrowser.value,
      preferredEditor: preferredEditor.value,
      backend: backend.value,
      theme: theme.value,
      locale: locale.value,
      ...(backend.value === 'ssh' ? { ssh: sshCfg } : {}),
    }

    await ConfigService.set('openLinksInBrowser', openLinksInBrowser.value)
    await ConfigService.set('preferredEditor', preferredEditor.value)
    await ConfigService.set('backend', backend.value)
    await ConfigService.set('theme', theme.value)
    await ConfigService.set('locale', locale.value)

    if (backend.value === 'ssh') {
      await ConfigService.set('ssh', sshCfg)
    }

    if (backend.value === 'wsl') {
      const distroValue = wslDistro.value.trim()
      await ConfigService.set('wslDistro', distroValue)
      patch.wslDistro = distroValue
    }

    appStore.patchConfig(patch)
    appStore.appLog(t('settings.saved'), 'success')
    appStore.showToast(t('settings.savedToast'), 'success')

    await DdevService.reloadBackend()
    await appStore.refreshProjects()
    emit('close')
  } catch (caughtError) {
    const message = caughtError instanceof Error ? caughtError.message : String(caughtError)
    appStore.appLog(`Failed to save settings: ${message}`, 'error')
    appStore.showToast(t('settings.saveFailed'), 'error')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <Modal :title="t('settings.title')" @close="emit('close')">
    <template #footer>
      <button class="flu-btn flu-btn-ghost" type="button" :disabled="saving" @click="emit('close')">
        {{ t('general.cancel') }}
      </button>
      <button class="flu-btn flu-btn-accent" type="button" :disabled="saving" @click="handleSave">
        {{ saving ? t('general.saving') : t('general.save') }}
      </button>
    </template>

    <div class="flu-form">
      <div class="flu-form-group">
        <label class="flu-field-label" for="settLanguage">{{ t('settings.language') }}</label>
        <Select
          id="settLanguage"
          v-model="locale"
          :options="Object.entries(LOCALE_LABELS).map(([value, label]) => ({ value, label }))"
        />
        <p class="text-muted" style="margin-top: 0.25rem; font-size: 0.85em">
          {{ t('settings.languageDesc') }}
        </p>
      </div>

      <hr style="border: none; border-top: 1px solid var(--surface-stroke); margin: 0">

      <div class="flu-form-group">
        <label class="flu-toggle-label">
          <input
            v-model="openLinksInBrowser"
            type="checkbox"
            class="flu-toggle"
          >
          <span>{{ t('settings.openInBrowser') }}</span>
        </label>
        <p class="text-muted" style="margin-top: 0.25rem; font-size: 0.85em">
          {{ t('settings.openInBrowserDesc') }}
        </p>
      </div>

      <div class="flu-form-group">
        <label class="flu-field-label" for="settPreferredEditor">{{ t('settings.preferredEditor') }}</label>
        <p class="text-muted" style="margin-top: 0.1rem; margin-bottom: 0.4rem; font-size: 0.85em">
          {{ t('settings.preferredEditorDesc') }}
        </p>
        <Select id="settPreferredEditor" v-model="preferredEditor" :options="EDITOR_OPTIONS" />
      </div>

      <template v-if="!isLinuxPlatform">
        <hr style="border: none; border-top: 1px solid var(--surface-stroke); margin: 0">

        <div class="flu-form-group">
          <label class="flu-field-label" for="settTheme">{{ t('settings.theme') }}</label>
          <Select id="settTheme" v-model="theme" :options="THEME_OPTIONS" />
          <p class="text-muted" style="margin-top: 0.25rem; font-size: 0.85em">
            {{ t('settings.themeDesc') }}
          </p>
        </div>
      </template>

      <div class="flu-form-group">
        <label class="flu-field-label" for="settBackend">{{ t('settings.backend') }}</label>
        <p class="text-muted" style="margin-top: 0.1rem; margin-bottom: 0.4rem; font-size: 0.85em">
          {{ t('settings.backendDesc') }}
        </p>
        <Select id="settBackend" v-model="backend" :options="backendOptions" />
      </div>

      <div v-if="backend === 'wsl'" class="flu-form-group">
        <label class="flu-field-label" for="settWSLDistro">{{ t('settings.wslDistro') }}</label>
        <p class="text-muted" style="margin-top: 0.1rem; margin-bottom: 0.4rem; font-size: 0.85em">
          {{ t('settings.wslDistroDesc') }}
        </p>
        <div style="max-width: 16rem">
          <Select
            id="settWSLDistro"
            v-model="wslDistro"
            :options="[
              ...(wslDistros.length > 0 ? wslDistros : [wslDistro]).map((distro) => ({ value: distro, label: distro })),
              ...(wslDistros.length > 0 && !wslDistros.includes(wslDistro) ? [{ value: wslDistro, label: wslDistro }] : []),
            ]"
          />
        </div>
      </div>

      <div v-if="devMode && backend === 'ssh'" style="display: flex; flex-direction: column; gap: var(--space-sm)">
        <div class="flu-form-group">
          <label class="flu-field-label" for="settSSHHost">{{ t('settings.sshHost') }}</label>
          <input
            id="settSSHHost"
            v-model="sshHost"
            type="text"
            class="flu-input"
            placeholder="192.168.1.100 or myserver.example.com"
            autocomplete="off"
            spellcheck="false"
          >
        </div>
        <div class="flu-form-group">
          <label class="flu-field-label" for="settSSHPort">{{ t('settings.sshPort') }}</label>
          <input
            id="settSSHPort"
            v-model="sshPort"
            type="text"
            class="flu-input"
            placeholder="22"
            autocomplete="off"
            style="max-width: 8rem"
          >
        </div>
        <div class="flu-form-group">
          <label class="flu-field-label" for="settSSHUser">{{ t('settings.sshUser') }}</label>
          <input
            id="settSSHUser"
            v-model="sshUser"
            type="text"
            class="flu-input"
            placeholder="username"
            autocomplete="off"
            spellcheck="false"
          >
        </div>
        <div class="flu-form-group">
          <label class="flu-field-label" for="settSSHKey">{{ t('settings.sshKey') }}</label>
          <p class="text-muted" style="margin-top: 0.1rem; margin-bottom: 0.4rem; font-size: 0.85em">
            {{ t('settings.sshKeyDesc') }}
          </p>
          <input
            id="settSSHKey"
            v-model="sshKey"
            type="text"
            class="flu-input"
            placeholder="~/.ssh/id_rsa"
            autocomplete="off"
            spellcheck="false"
          >
        </div>
      </div>
    </div>
  </Modal>
</template>
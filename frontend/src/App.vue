<script setup lang="ts">
import { onMounted, watch } from 'vue'
import { RouterView, useRoute } from 'vue-router'
import LogPanel from '@/components/LogPanel.vue'
import Titlebar from '@/components/Titlebar.vue'
import ToastContainer from '@/components/ToastContainer.vue'
import { useRuntimeEvents } from '@/composables/useRuntimeEvents'
import { useTranslation } from '@/lib/i18n'
import type { AppModal } from '@/lib/types'
import { DdevService } from '@/lib/wails'
import { useAppStore } from '@/stores/app'
import AboutModal from '@/modals/AboutModal.vue'
import EnvInfoModal from '@/modals/EnvInfoModal.vue'
import NewProjectModal from '@/modals/NewProjectModal.vue'
import SettingsModal from '@/modals/SettingsModal.vue'

const themeClasses = ['theme-acrylic', 'theme-default', 'theme-tabbed']

const route = useRoute()
const appStore = useAppStore()
const { setLocale } = useTranslation()

useRuntimeEvents()

function closeModal(modal: AppModal) {
  appStore.closeModal(modal)
}

function openSettingsFromEnvInfo() {
  appStore.closeModal('envInfo')
  appStore.openModal('settings')
}

async function shouldShowEnvInfoForMissingWsl(): Promise<boolean> {
  try {
    return !(await DdevService.wslExists())
  } catch (caughtError) {
    const message = caughtError instanceof Error ? caughtError.message : String(caughtError)
    appStore.appLog(`Failed to verify WSL availability: ${message}`, 'error')
    return false
  }
}

function isDdevMissingMessage(message: string): boolean {
  return /ddev.*(not found|not installed)|command not found|executable file not found|no such file|not recognized as an internal or external command|exit status\s+(127|9009)/i.test(
    message,
  )
}

async function shouldShowEnvInfoForMissingDdev(): Promise<boolean> {
  try {
    await DdevService.ddevInstalledVersion()
    return false
  } catch (caughtError) {
    const message = caughtError instanceof Error ? caughtError.message : String(caughtError)
    if (isDdevMissingMessage(message)) {
      return true
    }

    appStore.appLog(`Failed to verify DDEV availability: ${message}`, 'error')
    return false
  }
}

watch(
  () => route.params.name,
  () => {
    if (route.name === 'project-detail' && typeof route.params.name === 'string') {
      appStore.navigateToDetail(route.params.name)
      return
    }

    appStore.navigateToList()
  },
  { immediate: true },
)

watch(
  () => appStore.config.theme,
  (theme) => {
    document.body.classList.remove(...themeClasses)
    document.body.classList.add(`theme-${theme ?? 'default'}`)
  },
  { immediate: true },
)

watch(
  () => appStore.config.locale,
  async (locale) => {
    await setLocale(locale ?? 'en')
  },
  { immediate: true },
)

onMounted(async () => {
  const refreshProjectsPromise = Promise.allSettled([appStore.refreshProjects()])

  await appStore.loadConfig()
  const missingWsl = await shouldShowEnvInfoForMissingWsl()
  const missingDdev = missingWsl ? false : await shouldShowEnvInfoForMissingDdev()

  if (typeof appStore.config.ddevTelemetryOptIn === 'undefined' || missingWsl || missingDdev) {
    appStore.openModal('envInfo')
  }

  await refreshProjectsPromise
})
</script>

<template>
  <div id="appRoot" class="app-root">
    <Titlebar />

    <main id="mainContent" class="main-content">
      <RouterView />
    </main>

    <LogPanel />
    <ToastContainer />

    <NewProjectModal
      v-if="appStore.modals.newProject"
      @close="closeModal('newProject')"
    />
    <EnvInfoModal
      v-if="appStore.modals.envInfo"
      @close="closeModal('envInfo')"
      @open-settings="openSettingsFromEnvInfo"
    />
    <SettingsModal
      v-if="appStore.modals.settings"
      @close="closeModal('settings')"
    />
    <AboutModal
      v-if="appStore.modals.about"
      @close="closeModal('about')"
    />
  </div>
</template>

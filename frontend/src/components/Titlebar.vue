<script setup lang="ts">
import {
  CircleCheckIcon,
  CopyIcon,
  GlobeIcon,
  InfoIcon,
  LogsIcon,
  MinusIcon,
  PlusIcon,
  RefreshCwIcon as RefreshIcon,
  SettingsIcon,
  SquareIcon,
  XIcon,
} from '@lucide/vue'
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useTranslation } from '@/lib/i18n'
import type { AppModal } from '@/lib/types'
import { getProjectName, getProjectType } from '@/lib/utils'
import { DdevService, Runtime } from '@/lib/wails'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()
const router = useRouter()
const { t } = useTranslation()

const isMaximised = ref(false)
const openMenu = ref<'projects' | 'view' | 'help' | null>(null)

const projectName = computed(() =>
  appStore.currentView !== 'list' && appStore.selectedProject ? appStore.selectedProject : null,
)

const projectType = computed(() => {
  if (!projectName.value) return null
  const match = appStore.projects.find((project) => getProjectName(project) === projectName.value)
  return match ? getProjectType(match) || null : null
})

const centerLabel = computed(() => {
  if (!projectName.value) return t('general.home')
  return projectType.value ? `${projectName.value} | ${projectType.value}` : projectName.value
})

function syncMaximisedState() {
  Runtime.isMaximised().then((value) => {
    isMaximised.value = value
  }).catch(() => {})
}

function toggleMaximise() {
  Runtime.toggleMaximise()
  window.setTimeout(syncMaximisedState, 100)
}

function handleTitlebarDblClick(event: MouseEvent) {
  const target = event.target as HTMLElement | null
  if (target?.closest('button, a, .menubar, .window-controls, .titlebar-brand')) return
  toggleMaximise()
}

function closeMenus() {
  openMenu.value = null
}

function toggleMenu(menu: 'projects' | 'view' | 'help') {
  openMenu.value = openMenu.value === menu ? null : menu
}

function openModal(modal: AppModal) {
  closeMenus()
  appStore.openModal(modal)
}

async function refreshProjects() {
  closeMenus()
  await appStore.refreshProjects()
}

async function stopAllProjects() {
  closeMenus()
  appStore.appLog('Stopping all projects...', 'info')
  try {
    await DdevService.powerOff()
    appStore.appLog('All projects stopped.', 'success')
    appStore.showToast('All projects stopped', 'success')
    await appStore.refreshProjects()
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error)
    appStore.appLog(`Stop all failed: ${message}`, 'error')
  }
}

async function toggleBooleanConfig(key: 'openLinksInBrowser' | 'showLog') {
  const next = !(appStore.config[key] ?? true)
  await appStore.saveConfigValue(key, next)
}

function navigateHome() {
  closeMenus()
  appStore.navigateToList()
  void router.push({ name: 'project-list' })
}

onMounted(() => {
  syncMaximisedState()
  window.addEventListener('resize', syncMaximisedState)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', syncMaximisedState)
})
</script>

<template>
  <div class="titlebar" @dblclick="handleTitlebarDblClick">
    <div class="titlebar-left">
      <div
        class="titlebar-brand"
        role="button"
        tabindex="0"
        :title="t('titlebar.backToList')"
        @click="navigateHome"
        @keydown.enter.prevent="navigateHome"
        @keydown.space.prevent="navigateHome"
      >
        <h1>
          <span class="titlebar-brand-accent">DDEV</span>
          <span class="titlebar-brand-light">GUI</span>
        </h1>
      </div>

      <div id="menubar" class="menubar" @mouseleave="closeMenus">
        <div class="menubar-item" @mouseenter="openMenu = 'projects'">
          <button class="menubar-label" type="button" @click="toggleMenu('projects')">
            {{ t('menu.projects') }}
          </button>
          <div v-if="openMenu === 'projects'" class="menubar-dropdown">
            <button class="menubar-dropdown-item" type="button" @click="openModal('newProject')">
              <PlusIcon :size="14" :stroke-width="2" />
              {{ t('menu.newProject') }}
            </button>
            <button class="menubar-dropdown-item" type="button" @click="refreshProjects">
              <RefreshIcon :size="14" :stroke-width="2" />
              {{ t('menu.refresh') }}
              <span class="menu-shortcut">F5</span>
            </button>
            <div class="menubar-sep" />
            <button class="menubar-dropdown-item" type="button" @click="stopAllProjects">
              <SquareIcon :size="14" :stroke-width="2" />
              {{ t('menu.stopAll') }}
            </button>
          </div>
        </div>

        <div class="menubar-item" @mouseenter="openMenu = 'view'">
          <button class="menubar-label" type="button" @click="toggleMenu('view')">
            {{ t('menu.view') }}
          </button>
          <div v-if="openMenu === 'view'" class="menubar-dropdown">
            <button class="menubar-dropdown-item" type="button" @click="toggleBooleanConfig('showLog')">
              <CircleCheckIcon v-if="appStore.config.showLog !== false" :size="14" :stroke-width="2" />
              <LogsIcon v-else :size="14" :stroke-width="2" />
              {{ t('menu.toggleLog') }}
              <span class="menu-shortcut">Ctrl+L</span>
            </button>
            <div class="menubar-sep" />
            <button class="menubar-dropdown-item" type="button" @click="openModal('settings')">
              <SettingsIcon :size="14" :stroke-width="2" />
              {{ t('menu.settings') }}
            </button>
          </div>
        </div>

        <div class="menubar-item" @mouseenter="openMenu = 'help'">
          <button class="menubar-label" type="button" @click="toggleMenu('help')">
            {{ t('menu.help') }}
          </button>
          <div v-if="openMenu === 'help'" class="menubar-dropdown">
            <button class="menubar-dropdown-item" type="button" @click="openModal('about')">
              <InfoIcon :size="14" :stroke-width="2" />
              {{ t('menu.about') }}
            </button>
            <button class="menubar-dropdown-item" type="button" @click="openModal('envInfo')">
              <CircleCheckIcon :size="14" :stroke-width="2" />
              {{ t('menu.checkEnvironment') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <span id="titlebarCenter" class="titlebar-center">{{ centerLabel }}</span>

    <div class="titlebar-actions">
      <div class="view-toggle" style="align-self: center; margin-right: 4px">
        <button
          id="browserToggleBtn"
          type="button"
          class="view-toggle-btn"
          :class="{ active: appStore.config.openLinksInBrowser }"
          :title="appStore.config.openLinksInBrowser ? t('titlebar.useEmbedded') : t('titlebar.useBrowser')"
          @click="toggleBooleanConfig('openLinksInBrowser')"
        >
          <GlobeIcon :size="14" :stroke-width="2" />
        </button>
      </div>

      <div class="view-toggle" style="align-self: center; margin-right: 8px">
        <button
          id="logToggleBtn"
          type="button"
          class="view-toggle-btn"
          :class="{ active: appStore.config.showLog !== false }"
          :title="appStore.config.showLog !== false ? t('titlebar.hideLog') : t('titlebar.showLog')"
          @click="toggleBooleanConfig('showLog')"
        >
          <LogsIcon :size="14" :stroke-width="2" />
        </button>
      </div>

      <div class="window-controls">
        <button
          id="winMinimize"
          type="button"
          class="win-ctrl win-minimize"
          :title="t('titlebar.minimize')"
          @click="Runtime.minimise()"
        >
          <MinusIcon :size="15" :stroke-width="2" />
        </button>
        <button
          id="winMaximize"
          type="button"
          class="win-ctrl win-maximize"
          :title="isMaximised ? t('titlebar.restore') : t('titlebar.maximize')"
          @click="toggleMaximise"
        >
          <CopyIcon v-if="isMaximised" :size="15" :stroke-width="2" />
          <SquareIcon v-else :size="13" :stroke-width="2" />
        </button>
        <button
          id="winClose"
          type="button"
          class="win-ctrl win-close"
          :title="t('titlebar.close')"
          @click="Runtime.quit()"
        >
          <XIcon :size="15" :stroke-width="2" />
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.titlebar-brand {
  cursor: pointer;
}

.titlebar-brand-accent {
  font-weight: 800;
  color: #38bdf8;
}

.titlebar-brand-light {
  margin-left: 0.35rem;
  font-weight: 300;
  opacity: 0.9;
}
</style>
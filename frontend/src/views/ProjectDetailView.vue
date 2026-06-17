<script setup lang="ts">
import {
  ArrowLeftIcon,
  Code2Icon,
  CopyIcon,
  DownloadIcon,
  DropletIcon,
  EllipsisVerticalIcon,
  EraserIcon,
  ExternalLinkIcon,
  FileIcon,
  FolderOpenIcon,
  InfoIcon,
  LayersIcon,
  LogsIcon,
  MailIcon,
  PlayIcon,
  RefreshCwIcon as RefreshIcon,
  SquareIcon,
  TerminalIcon,
  Trash2Icon,
  UploadIcon,
  VenetianMaskIcon,
  ZapIcon,
} from '@lucide/vue'
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import EmbeddedTerminal from '@/components/EmbeddedTerminal.vue'
import FileExplorer from '@/components/FileExplorer.vue'
import ProjectLogs from '@/components/ProjectLogs.vue'
import ProjectOverview from '@/components/ProjectOverview.vue'
import ProjectSnapshots from '@/components/ProjectSnapshots.vue'
import ProjectTypeLogo from '@/components/ProjectTypeLogo.vue'
import ProjectAddons from '@/components/ProjectAddons.vue'
import ConfigServiceModal from '@/modals/ConfigServiceModal.vue'
import ConfirmDeleteModal from '@/modals/ConfirmDeleteModal.vue'
import CreateSnapshotModal from '@/modals/CreateSnapshotModal.vue'
import MasqueradeModal from '@/modals/MasqueradeModal.vue'
import ModifyProjectModal from '@/modals/ModifyProjectModal.vue'
import { useTranslation } from '@/lib/i18n'
import type { DdevProject, DdevService, DdevSnapshot } from '@/lib/types'
import { ConfigService, DdevService as DdevApi, Runtime } from '@/lib/wails'
import {
  coerceToBool,
  getMailpitUrl,
  getPrimaryUrl,
  getProjectName,
  getProjectStatus,
  getProjectType,
  isProjectStopped,
  openUrl,
  pickProjectValue,
} from '@/lib/utils'
import { useAppStore } from '@/stores/app'

type DetailTab = 'overview' | 'files' | 'addons' | 'snapshots' | 'logs' | 'terminal'
type ToolbarMenu = 'drupal' | 'more' | null

type DeleteTarget = { kind: 'project' } | { kind: 'snapshot'; snapshotName: string }

interface OverviewItem {
  label: string
  value: string
  isStatus?: boolean
}

const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const { t } = useTranslation()

const describeData = ref<DdevProject | null>(null)
const snapshots = ref<DdevSnapshot[]>([])
const activeTab = ref<DetailTab>('overview')
const loadingDesc = ref(false)
const loadingSnaps = ref(false)
const initRunning = ref(false)
const openToolbarMenu = ref<ToolbarMenu>(null)
const showMasquerade = ref(false)
const showModify = ref(false)
const showSnapshotCreate = ref(false)
const showServiceConfig = ref(false)
const deleteTarget = ref<DeleteTarget | null>(null)
const deleteRunning = ref(false)

const routeProjectName = computed(() => String(route.params.name ?? ''))
const cachedProject = computed(() => appStore.projectsMap.get(routeProjectName.value) ?? null)
const displayProject = computed(() => describeData.value ?? cachedProject.value)
const projectType = computed(() =>
  displayProject.value ? getProjectType(displayProject.value) : '',
)
const projectStatus = computed(() =>
  displayProject.value ? getProjectStatus(displayProject.value) : '',
)
const primaryUrl = computed(() => (displayProject.value ? getPrimaryUrl(displayProject.value) : ''))
const mailpitUrl = computed(() => (displayProject.value ? getMailpitUrl(displayProject.value) : ''))
const isStopped = computed(() =>
  displayProject.value ? isProjectStopped(displayProject.value) : false,
)
const isInitialized = computed(
  () => !!appStore.config.projects?.[routeProjectName.value]?.initialized,
)
const isDrupal = computed(() => projectType.value.toLowerCase().startsWith('drupal'))
const isSshBackend = computed(() => appStore.config.backend === 'ssh')
const projectLocation = computed(() =>
  String(
    displayProject.value?.approot ?? displayProject.value?.shortroot ?? routeProjectName.value,
  ),
)
const tabs = computed(() => [
  { id: 'overview' as const, label: t('detail.tabOverview') },
  { id: 'addons' as const, label: t('detail.tabAddons') },
  { id: 'snapshots' as const, label: t('detail.tabSnapshots') },
  { id: 'files' as const, label: t('detail.tabFiles') },
  { id: 'logs' as const, label: t('detail.tabLogs') },
  { id: 'terminal' as const, label: t('detail.tabTerminal') },
])
const overviewItems = computed<OverviewItem[]>(() => {
  const project = displayProject.value

  return [
    {
      label: t('detail.overview.name'),
      value: project ? getProjectName(project) : routeProjectName.value,
    },
    { label: t('detail.overview.status'), value: projectStatus.value || 'n/a', isStatus: true },
    { label: t('detail.overview.type'), value: projectType.value || 'n/a' },
    {
      label: t('detail.overview.phpVersion'),
      value: String(pickProjectValue(project, ['php_version', 'phpversion']) ?? '') || 'n/a',
    },
    {
      label: t('detail.overview.docroot'),
      value: String(pickProjectValue(project, ['docroot']) ?? '') || 'n/a',
    },
    {
      label: t('detail.overview.location'),
      value: String(pickProjectValue(project, ['approot', 'shortroot']) ?? '') || 'n/a',
    },
    {
      label: t('detail.overview.router'),
      value: String(pickProjectValue(project, ['router']) ?? '') || 'n/a',
    },
    {
      label: t('detail.overview.nodejs'),
      value: String(pickProjectValue(project, ['nodejs_version']) ?? '') || 'n/a',
    },
  ]
})
const services = computed(() => {
  const raw = displayProject.value?.services
  if (!raw || typeof raw !== 'object') return []

  return Object.entries(raw as Record<string, DdevService>)
})
const logServiceNames = computed(() => services.value.map(([serviceName]) => serviceName))
const normalizedSnapshots = computed(() =>
  snapshots.value.map(normalizeSnapshotName).filter((name): name is string => Boolean(name)),
)
const deleteMessage = computed(() => {
  if (!deleteTarget.value) return ''

  if (deleteTarget.value.kind === 'project') {
    return t('detail.delete.confirm', { name: routeProjectName.value })
  }

  return t('detail.snapshots.deleteConfirm', { snap: deleteTarget.value.snapshotName })
})

watch(
  routeProjectName,
  async (projectName) => {
    activeTab.value = 'overview'

    if (!projectName) {
      describeData.value = null
      snapshots.value = []
      return
    }

    if (!appStore.isProjectsLoaded && !appStore.isLoadingProjects) {
      await appStore.refreshProjects()
    }

    await Promise.allSettled([loadDescribe(projectName), loadSnapshots(projectName)])
  },
  { immediate: true },
)

watch(
  activeTab,
  (tab) => {
    appStore.setTerminalActive(tab === 'terminal' || tab === 'files' || tab === 'logs')
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  appStore.setTerminalActive(false)
})

function setToolbarMenu(menu: Exclude<ToolbarMenu, null> | null) {
  openToolbarMenu.value = menu
}

function toggleToolbarMenu(menu: Exclude<ToolbarMenu, null>) {
  openToolbarMenu.value = openToolbarMenu.value === menu ? null : menu
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}

function normalizeSnapshotName(snapshot: DdevSnapshot | string): string | null {
  if (typeof snapshot === 'string') return snapshot.trim() || null

  const value =
    snapshot.Name ?? snapshot.name ?? snapshot.snapshot_name ?? snapshot.snapshotName ?? ''
  const normalized = String(value).trim()
  return normalized || null
}

async function loadDescribe(projectName = routeProjectName.value) {
  if (!projectName) return

  loadingDesc.value = true
  try {
    const json = await DdevApi.describeJSON(projectName)
    let data: unknown = JSON.parse(json)

    if (Array.isArray(data)) data = data[0]
    if (isRecord(data) && isRecord(data.raw)) data = data.raw

    describeData.value = isRecord(data) ? (data as DdevProject) : null
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error)
    appStore.appLog(`Failed to load project details: ${message}`, 'error')
  } finally {
    loadingDesc.value = false
  }
}

async function loadSnapshots(projectName = routeProjectName.value) {
  if (!projectName) return

  loadingSnaps.value = true
  try {
    const json = await DdevApi.snapshotListJSON(projectName)
    const data = JSON.parse(json) as unknown

    if (Array.isArray(data)) {
      snapshots.value = data as DdevSnapshot[]
      return
    }

    if (isRecord(data) && Array.isArray(data.raw)) {
      snapshots.value = data.raw as DdevSnapshot[]
      return
    }

    if (isRecord(data) && isRecord(data.raw)) {
      let list: unknown[] | undefined
      for (const key in data.raw) {
        if (Array.isArray(data.raw[key])) {
          list = data.raw[key]
          break
        }
      }
      snapshots.value = list ? (list as DdevSnapshot[]) : []
      return
    }

    snapshots.value = []
  } catch {
    snapshots.value = []
  } finally {
    loadingSnaps.value = false
  }
}

async function refreshProjectState() {
  await Promise.allSettled([loadDescribe(), appStore.refreshProjects()])
}

function openProjectUrl(url: string) {
  openUrl(url, coerceToBool(appStore.config.openLinksInBrowser))
}

function openLocation(eventName: 'open:terminal' | 'open:folder' | 'open:editor') {
  Runtime.emit(eventName, { location: projectLocation.value })
}

async function runAction(action: 'start' | 'stop' | 'restart') {
  appStore.appLog(`Running ${action} on ${routeProjectName.value}...`, 'info')

  try {
    if (action === 'start') await DdevApi.start(routeProjectName.value)
    else if (action === 'stop') await DdevApi.stop(routeProjectName.value)
    else await DdevApi.restart(routeProjectName.value)

    appStore.appLog(`${routeProjectName.value} ${action}ed`, 'success')
    appStore.showToast(`${routeProjectName.value} ${action}ed`, 'success')
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error)
    appStore.appLog(`${action} failed: ${message}`, 'error')
    appStore.showToast(`${action} failed`, 'error')
  } finally {
    await refreshProjectState()
  }
}

function openProjectDeleteModal() {
  deleteTarget.value = { kind: 'project' }
}

function closeDeleteModal() {
  if (deleteRunning.value) return
  deleteTarget.value = null
}

async function deleteProject() {
  appStore.appLog(`Deleting project ${routeProjectName.value}...`, 'info')
  try {
    await DdevApi.deleteProject(routeProjectName.value)
    await ConfigService.setProjectConfig(routeProjectName.value, 'initialized', false)
    appStore.patchConfig({
      projects: {
        ...appStore.config.projects,
        [routeProjectName.value]: { initialized: false },
      },
    })
    appStore.appLog(`Project ${routeProjectName.value} deleted.`, 'success')
    appStore.showToast(`Project ${routeProjectName.value} deleted`, 'success')
    await appStore.refreshProjects()
    await router.push({ name: 'project-list' })
    return true
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error)
    appStore.appLog(`Error deleting ${routeProjectName.value}: ${message}`, 'error')
    appStore.showToast(`Error deleting ${routeProjectName.value}`, 'error')
    return false
  }
}

async function handleExportDB() {
  appStore.appLog(`Exporting database for ${routeProjectName.value}...`, 'info')

  try {
    await DdevApi.exportDB(routeProjectName.value)
    appStore.appLog(`Database exported for ${routeProjectName.value}`, 'success')
    appStore.showToast(`Database exported for ${routeProjectName.value}`, 'success')
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error)
    if (message.includes('cancelled')) return
    appStore.appLog(`Export DB failed: ${message}`, 'error')
    appStore.showToast('Export DB failed', 'error')
  }
}

async function handleImportDB() {
  try {
    const filePath = await DdevApi.importDBSelectFile(routeProjectName.value)
    if (!filePath) return

    const fileName = filePath.split('/').pop() || filePath
    if (
      !window.confirm(
        t('detail.importDb.confirm', { file: fileName, project: routeProjectName.value }),
      )
    ) {
      return
    }

    appStore.appLog(`Importing database for ${routeProjectName.value} from ${fileName}...`, 'info')
    await DdevApi.importDBFromFile(routeProjectName.value, filePath)
    await ConfigService.setProjectConfig(routeProjectName.value, 'initialized', true)
    appStore.patchConfig({
      projects: {
        ...appStore.config.projects,
        [routeProjectName.value]: { initialized: true },
      },
    })
    appStore.appLog(`Database imported for ${routeProjectName.value}`, 'success')
    appStore.showToast(`Database imported for ${routeProjectName.value}`, 'success')
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error)
    if (message.includes('cancelled')) return
    appStore.appLog(`Import DB failed: ${message}`, 'error')
    appStore.showToast('Import DB failed', 'error')
  }
}

async function handleDrushUli() {
  appStore.appLog(`Getting admin login URL for ${routeProjectName.value}...`, 'info')

  try {
    let uliUrl = await DdevApi.drushUli(routeProjectName.value)
    if (uliUrl && !/^https?:\/\//i.test(uliUrl)) {
      const base = primaryUrl.value.replace(/\/+$/, '')
      const path = uliUrl.startsWith('/') ? uliUrl : `/${uliUrl}`
      uliUrl = `${base}${path}`
    }

    if (uliUrl) {
      openProjectUrl(uliUrl)
    }
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error)
    appStore.appLog(`Drush ULI failed: ${message}`, 'error')
  }
}

async function handleClearCache() {
  appStore.appLog(`Clearing cache for ${routeProjectName.value}...`, 'info')

  try {
    await DdevApi.drushCacheRebuild(routeProjectName.value)
    appStore.appLog(`Cache cleared for ${routeProjectName.value}`, 'success')
    appStore.showToast(`Cache cleared for ${routeProjectName.value}`, 'success')
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error)
    appStore.appLog(`Clear cache failed: ${message}`, 'error')
    appStore.showToast('Clear cache failed', 'error')
  }
}

async function openMasqueradeModal() {
  showMasquerade.value = true
}

function openModifyModal() {
  showModify.value = true
}

function openServiceConfigModal() {
  showServiceConfig.value = true
}

function openSnapshotCreateModal() {
  showSnapshotCreate.value = true
}

async function handleServicesConfigured() {
  await refreshProjectState()
}

async function handleInitSite(isReinit = false) {
  const type = projectType.value.toLowerCase()
  const isDrupalSite = type.startsWith('drupal')

  initRunning.value = true
  appStore.appLog(
    `${isReinit ? 'Re-initializing' : 'Initializing'} site for ${routeProjectName.value} (${type})...`,
    'info',
  )

  try {
    await DdevApi.start(routeProjectName.value)
    if (type === 'wordpress') {
      await DdevApi.wpCoreInstall(routeProjectName.value)
    } else {
      await DdevApi.composerInstall(routeProjectName.value, type)
      if (isDrupalSite) {
        await DdevApi.drushSiteInstall(routeProjectName.value)
      }
    }

    await ConfigService.setProjectConfig(routeProjectName.value, 'initialized', true)
    appStore.patchConfig({
      projects: {
        ...appStore.config.projects,
        [routeProjectName.value]: { initialized: true },
      },
    })
    appStore.appLog(`Site initialized for ${routeProjectName.value}`, 'success')
    appStore.showToast(`Site initialized for ${routeProjectName.value}`, 'success')
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error)
    appStore.appLog(`Init site failed: ${message}`, 'error')
    appStore.showToast(`Init failed for ${routeProjectName.value}`, 'error')
  } finally {
    initRunning.value = false
    await refreshProjectState()
  }
}

async function handleSnapshotCreated() {
  await loadSnapshots()
}

async function handleSnapshotRestore(snapshotName: string) {
  if (!window.confirm(t('detail.snapshots.restoreConfirm', { snap: snapshotName }))) return

  appStore.appLog(`Restoring snapshot ${snapshotName} for ${routeProjectName.value}...`, 'info')
  try {
    await DdevApi.snapshotRestore(routeProjectName.value, snapshotName)
    appStore.appLog(`Snapshot ${snapshotName} restored`, 'success')
    appStore.showToast('Snapshot restored', 'success')
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error)
    appStore.appLog(`Snapshot restore failed: ${message}`, 'error')
    appStore.showToast('Snapshot restore failed', 'error')
  }
}

function handleSnapshotDelete(snapshotName: string) {
  deleteTarget.value = { kind: 'snapshot', snapshotName }
}

async function deleteSnapshot(snapshotName: string) {
  appStore.appLog(`Deleting snapshot ${snapshotName} from ${routeProjectName.value}...`, 'info')
  try {
    await DdevApi.snapshotDelete(routeProjectName.value, snapshotName)
    appStore.appLog(`Snapshot ${snapshotName} deleted`, 'success')
    appStore.showToast('Snapshot deleted', 'success')
    await loadSnapshots()
    return true
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error)
    appStore.appLog(`Snapshot delete failed: ${message}`, 'error')
    appStore.showToast('Snapshot delete failed', 'error')
    return false
  }
}

async function handleDeleteConfirm() {
  if (!deleteTarget.value || deleteRunning.value) return

  deleteRunning.value = true

  try {
    const success =
      deleteTarget.value.kind === 'project'
        ? await deleteProject()
        : await deleteSnapshot(deleteTarget.value.snapshotName)

    if (success) {
      deleteTarget.value = null
    }
  } finally {
    deleteRunning.value = false
  }
}
</script>

<template>
  <section class="detail-root">
    <div class="detail-sticky-header">
      <div class="content-toolbar">
        <h2 id="sectionTitle" class="project-detail-view__title">
          <ProjectTypeLogo
            class="project-detail-view__title-mark"
            :type="projectType || routeProjectName"
            :style="{ marginRight: 0 }"
          />
          {{ routeProjectName }}
        </h2>
        <div class="header-controls" id="headerControls">
          <div class="detail-toolbar">
            <button
              v-if="!isInitialized"
              type="button"
              class="flu-btn flu-btn-sm flu-btn-accent proj-action"
              :disabled="initRunning"
              @click="handleInitSite()"
            >
              <ZapIcon :size="14" :stroke-width="2" />
              {{ initRunning ? t('detail.initializing') : t('detail.initSite') }}
            </button>
            <button
              type="button"
              class="flu-btn flu-btn-sm proj-action"
              :class="isStopped ? 'flu-btn-accent' : 'flu-btn-danger'"
              @click="runAction(isStopped ? 'start' : 'stop')"
            >
              <component :is="isStopped ? PlayIcon : SquareIcon" :size="14" :stroke-width="2" />
              {{ isStopped ? t('detail.start') : t('detail.stop') }}
            </button>
            <button
              type="button"
              class="flu-btn flu-btn-sm flu-btn-ghost proj-action"
              @click="runAction('restart')"
            >
              <RefreshIcon :size="14" :stroke-width="2" />
              {{ t('detail.restart') }}
            </button>
            <button
              type="button"
              class="flu-btn flu-btn-sm flu-btn-ghost proj-action"
              @click="openLocation('open:terminal')"
            >
              <TerminalIcon :size="14" :stroke-width="2" />
              {{ t('detail.openCli') }}
            </button>
            <div
              v-if="isDrupal"
              class="toolbar-dropdown"
              @mouseenter="setToolbarMenu('drupal')"
              @mouseleave="setToolbarMenu(null);"
            >
              <button
                type="button"
                class="flu-btn flu-btn-sm flu-btn-ghost toolbar-dropdown-toggle"
                @click="toggleToolbarMenu('drupal')"
              >
                <DropletIcon :size="14" :stroke-width="2" />
                {{ t('detail.drupal.label') }}
              </button>
              <div v-if="openToolbarMenu === 'drupal'" class="toolbar-dropdown-menu">
                <div class="toolbar-dropdown-menu-inner">
                  <button
                    type="button"
                    class="toolbar-dropdown-item proj-action"
                    :disabled="isSshBackend"
                    :title="isSshBackend ? t('detail.drupal.notAvailableSsh') : undefined"
                    :style="isSshBackend ? { opacity: 0.45, cursor: 'not-allowed' } : undefined"
                    @click="
                      setToolbarMenu(null);
                      handleExportDB()
                    "
                  >
                    <DownloadIcon :size="12" :stroke-width="2" />
                    {{ t('detail.drupal.exportDb') }}
                  </button>
                  <button
                    type="button"
                    class="toolbar-dropdown-item proj-action"
                    :disabled="isSshBackend"
                    :title="isSshBackend ? t('detail.drupal.notAvailableSsh') : undefined"
                    :style="isSshBackend ? { opacity: 0.45, cursor: 'not-allowed' } : undefined"
                    @click="
                      setToolbarMenu(null);
                      handleImportDB()
                    "
                  >
                    <UploadIcon :size="12" :stroke-width="2" />
                    {{ t('detail.drupal.importDb') }}
                  </button>
                  <button
                    type="button"
                    class="toolbar-dropdown-item proj-action"
                    :disabled="isStopped"
                    :title="isStopped ? t('detail.drupal.mustBeRunning') : undefined"
                    :style="isStopped ? { opacity: 0.45, cursor: 'not-allowed' } : undefined"
                    @click="
                      setToolbarMenu(null);
                      openMasqueradeModal()
                    "
                  >
                    <VenetianMaskIcon :size="12" :stroke-width="2" />
                    {{ t('detail.drupal.masquerade') }}
                  </button>
                  <button
                    type="button"
                    class="toolbar-dropdown-item proj-action"
                    :disabled="isStopped"
                    :title="isStopped ? t('detail.drupal.mustBeRunning') : undefined"
                    :style="isStopped ? { opacity: 0.45, cursor: 'not-allowed' } : undefined"
                    @click="
                      setToolbarMenu(null);
                      handleClearCache()
                    "
                  >
                    <EraserIcon :size="12" :stroke-width="2" />
                    {{ t('detail.drupal.clearCache') }}
                  </button>
                </div>
              </div>
            </div>
            <div
              class="toolbar-dropdown"
              @mouseenter="setToolbarMenu('more')"
              @mouseleave="setToolbarMenu(null);"
            >
              <button
                type="button"
                class="flu-btn flu-btn-sm flu-btn-ghost toolbar-dropdown-toggle"
                @click="toggleToolbarMenu('more')"
              >
                <EllipsisVerticalIcon :size="14" :stroke-width="2" />
                {{ t('detail.more.label') }}
              </button>
              <div v-if="openToolbarMenu === 'more'" class="toolbar-dropdown-menu">
                <div class="toolbar-dropdown-menu-inner">
                  <button
                    type="button"
                    class="toolbar-dropdown-item proj-action"
                    @click="
                      setToolbarMenu(null);
                      openLocation('open:folder')
                    "
                  >
                    <FolderOpenIcon :size="14" :stroke-width="2" />
                    {{ t('detail.more.openFolder') }}
                  </button>
                  <button
                    type="button"
                    class="toolbar-dropdown-item proj-action"
                    @click="
                      setToolbarMenu(null);
                      openLocation('open:editor')
                    "
                  >
                    <Code2Icon :size="14" :stroke-width="2" />
                    {{ t('detail.more.openEditor') }}
                  </button>
                  <button
                    v-if="isInitialized"
                    type="button"
                    class="toolbar-dropdown-item proj-action"
                    :disabled="initRunning"
                    @click="
                      setToolbarMenu(null);
                      handleInitSite(true)
                    "
                  >
                    <ZapIcon :size="12" :stroke-width="2" />
                    {{ t('detail.more.reinitSite') }}
                  </button>
                  <button
                    type="button"
                    class="toolbar-dropdown-item proj-action toolbar-dropdown-item-danger"
                    @click="
                      setToolbarMenu(null);
                      openProjectDeleteModal()
                    "
                  >
                    <Trash2Icon :size="12" :stroke-width="2" />
                    {{ t('detail.more.delete') }}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div id="detailUrls" class="url-pills" :class="{ stopped: isStopped }">
        <RouterLink id="backToList" class="flu-btn flu-btn-ghost" :to="{ name: 'project-list' }">
          <ArrowLeftIcon :size="12" :stroke-width="2" />
          {{ t('detail.back') }}
        </RouterLink>
        <div class="url-pills-right">
          <button
            v-if="primaryUrl"
            type="button"
            class="url-pill proj-action"
            @click="openProjectUrl(primaryUrl)"
          >
            <ExternalLinkIcon :size="12" :stroke-width="2" />
            {{ t('detail.openSite') }}
          </button>
          <button
            v-if="mailpitUrl"
            type="button"
            class="url-pill proj-action"
            @click="openProjectUrl(mailpitUrl)"
          >
            <MailIcon :size="12" :stroke-width="2" />
            Mailpit
          </button>
          <button
            v-if="primaryUrl && isDrupal"
            type="button"
            class="url-pill proj-action"
            @click="handleDrushUli"
          >
            <ExternalLinkIcon :size="12" :stroke-width="2" />
            {{ t('detail.openSiteAdmin') }}
          </button>
        </div>
      </div>
    </div>

    <div class="detail-scroll">
      <nav class="detail-sidebar" aria-label="Project detail sections">
        <button
          type="button"
          class="detail-sidebar-btn"
          :class="{ active: activeTab === 'overview' }"
          :title="t('detail.tabOverview')"
          @click="activeTab = 'overview'"
        >
          <InfoIcon :size="18" :stroke-width="2" />
          <span>{{ t('detail.tabOverview') }}</span>
        </button>
        <button
          type="button"
          class="detail-sidebar-btn"
          :class="{ active: activeTab === 'addons' }"
          :title="t('detail.tabAddons')"
          @click="activeTab = 'addons'"
        >
          <LayersIcon :size="18" :stroke-width="2" />
          <span>{{ t('detail.tabAddons') }}</span>
        </button>
        <button
          type="button"
          class="detail-sidebar-btn"
          :class="{ active: activeTab === 'snapshots' }"
          :title="t('detail.tabSnapshots')"
          @click="activeTab = 'snapshots'"
        >
          <CopyIcon :size="18" :stroke-width="2" />
          <span>{{ t('detail.tabSnapshots') }}</span>
        </button>
        <button
          type="button"
          class="detail-sidebar-btn"
          :class="{ active: activeTab === 'files' }"
          :title="t('detail.tabFiles')"
          @click="activeTab = 'files'"
        >
          <FileIcon :size="18" :stroke-width="2" />
          <span>{{ t('detail.tabFiles') }}</span>
        </button>
        <button
          type="button"
          class="detail-sidebar-btn"
          :class="{ active: activeTab === 'logs' }"
          :title="t('detail.tabLogs')"
          @click="activeTab = 'logs'"
        >
          <LogsIcon :size="18" :stroke-width="2" />
          <span>{{ t('detail.tabLogs') }}</span>
        </button>
        <button
          type="button"
          class="detail-sidebar-btn"
          :class="{ active: activeTab === 'terminal' }"
          :title="t('detail.tabTerminal')"
          @click="activeTab = 'terminal'"
        >
          <TerminalIcon :size="18" :stroke-width="2" />
          <span>{{ t('detail.tabTerminal') }}</span>
        </button>
      </nav>

      <div class="detail-content">
        <ProjectOverview
          v-if="activeTab === 'overview'"
          :loading="loadingDesc"
          :has-project="Boolean(displayProject)"
          :overview-items="overviewItems"
          :services="services"
          @modify="openModifyModal"
          @config-services="openServiceConfigModal"
          @open-url="openProjectUrl"
        />

        <div v-else-if="activeTab === 'files'" id="detailFiles" class="detail-surface-fill">
          <FileExplorer :project-name="routeProjectName" :project-root="projectLocation" />
        </div>

        <div v-else-if="activeTab === 'addons'" id="detailAddons" class="detail-surface-fill">
          <ProjectAddons :project-name="routeProjectName" />
        </div>

        <div v-else-if="activeTab === 'terminal'" id="detailTerminal" class="detail-surface-fill">
          <EmbeddedTerminal :project-name="routeProjectName" />
        </div>

        <div v-else-if="activeTab === 'logs'" id="detailLogsPanel" class="detail-surface-fill">
          <ProjectLogs :project-name="routeProjectName" :service-names="logServiceNames" />
        </div>

        <ProjectSnapshots
          v-else-if="activeTab === 'snapshots'"
          :loading="loadingSnaps"
          :snapshots="normalizedSnapshots"
          @create="openSnapshotCreateModal"
          @restore="handleSnapshotRestore"
          @delete="handleSnapshotDelete"
        />

        <section v-else class="detail-section">
          <div class="detail-section-title">
            {{ tabs.find((tab) => tab.id === activeTab)?.label }}
          </div>
          <div class="detail-section-body text-muted detail-placeholder-body">
            This section is not available for the current project view.
          </div>
        </section>
      </div>
    </div>

    <MasqueradeModal
      v-if="showMasquerade"
      :project-name="routeProjectName"
      :primary-url="primaryUrl"
      @close="showMasquerade = false"
    />

    <ModifyProjectModal
      v-if="showModify"
      :project-name="routeProjectName"
      :project="displayProject"
      @close="showModify = false"
      @modified="refreshProjectState()"
    />

    <ConfigServiceModal
      v-if="showServiceConfig"
      :project-name="routeProjectName"
      :project="displayProject"
      @close="showServiceConfig = false"
      @configured="handleServicesConfigured"
    />

    <CreateSnapshotModal
      v-if="showSnapshotCreate"
      :project-name="routeProjectName"
      @close="showSnapshotCreate = false"
      @created="handleSnapshotCreated"
    />

    <ConfirmDeleteModal
      v-if="deleteTarget"
      :title="t('general.delete')"
      :message="deleteMessage"
      :confirm-text="t('general.delete')"
      :pending="deleteRunning"
      @close="closeDeleteModal"
      @confirm="handleDeleteConfirm"
    />
  </section>
</template>

<style scoped>
.project-detail-view__title {
  display: inline-flex;
  align-items: center;
  gap: 0.75rem;
}

.project-detail-view__title-mark {
  display: block;
  flex-shrink: 0;
}

.detail-surface-fill {
  min-height: 0;
  height: 100%;
}

.detail-placeholder-body {
  max-width: 64ch;
}

@media (max-width: 900px) {
  .detail-scroll {
    grid-template-columns: 1fr;
  }

  .detail-toolbar-actions {
    justify-content: flex-start;
  }
}
</style>

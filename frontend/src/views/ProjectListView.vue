<script setup lang="ts">
import {
  FileBracesCornerIcon,
  FolderOpenIcon,
  LayoutGridIcon,
  ListIcon,
  PlayIcon,
  RotateCwIcon,
  SquareIcon,
  TerminalIcon,
} from '@lucide/vue'
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import ProjectTypeLogo from '@/components/ProjectTypeLogo.vue'
import Spinner from '@/components/Spinner.vue'
import { useTranslation } from '@/lib/i18n'
import type { DdevProject, ViewMode } from '@/lib/types'
import { getPrimaryUrl, getProjectName, getProjectStatus, getProjectType, isProjectStopped, openUrl } from '@/lib/utils'
import { DdevService, Runtime } from '@/lib/wails'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()
const { t } = useTranslation()

const errorMessage = ref('')

const projects = computed(() => appStore.projects)
const isLoading = computed(() => appStore.isLoadingProjects && !appStore.isProjectsLoaded)
const viewMode = computed(() => appStore.config.viewMode ?? 'list')

async function handleMenuRefresh() {
  await refreshProjects()
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key !== 'F5') return
  event.preventDefault()
  void refreshProjects()
}

onMounted(() => {
  if (!appStore.isProjectsLoaded && !appStore.isLoadingProjects) {
    void refreshProjects()
  }

  Runtime.on('menu:refresh', handleMenuRefresh)
  Runtime.on('menu:start', handleMenuRefresh)
  Runtime.on('menu:stop', handleMenuRefresh)
  document.addEventListener('keydown', handleKeydown)
})

onBeforeUnmount(() => {
  Runtime.off('menu:refresh', handleMenuRefresh)
  Runtime.off('menu:start', handleMenuRefresh)
  Runtime.off('menu:stop', handleMenuRefresh)
  document.removeEventListener('keydown', handleKeydown)
})

async function refreshProjects() {
  errorMessage.value = ''

  try {
    await appStore.refreshProjects()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : String(error)
  }
}

async function setViewMode(mode: ViewMode) {
  if (mode === viewMode.value) return

  try {
    await appStore.saveConfigValue('viewMode', mode)
  } catch {
    return
  }
}

async function runProjectAction(projectName: string, action: 'start' | 'stop' | 'restart') {
  try {
    const actionVerb = action === 'start' ? t('projectList.starting') : action === 'stop' ? t('projectList.stopping') : t('projectList.restarting')
    appStore.appLog(`${actionVerb}...`, 'info')

    if (action === 'start') await DdevService.start(projectName)
    else if (action === 'stop') await DdevService.stop(projectName)
    else await DdevService.restart(projectName)

    appStore.showToast(`${projectName}: ${action}`, 'success')
    await refreshProjects()
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error)
    appStore.appLog(`${projectName}: ${message}`, 'error')
  }
}

function openProjectUrl(url: string) {
  openUrl(url, appStore.config.openLinksInBrowser)
}

function openProjectFolder(project: DdevProject) {
  Runtime.emit('open:folder', { location: project.approot || project.shortroot || getProjectName(project) })
}

function openProjectCli(project: DdevProject) {
  Runtime.emit('open:terminal', { location: project.approot || project.shortroot || getProjectName(project) })
}

function openProjectEditor(project: DdevProject) {
  Runtime.emit('open:editor', { location: project.approot || project.shortroot || getProjectName(project) })
}

function statusClass(status: string): string {
  const normalized = status.toLowerCase()
  if (normalized.includes('run')) return 'status-badge running'
  if (normalized.includes('pause')) return 'status-badge paused'
  return 'status-badge stopped'
}
</script>

<template>
  <section class="detail-root">
    <div class="detail-sticky-header">
      <div class="content-toolbar">
        <h2 id="sectionTitle">{{ t('projectList.title') }}</h2>
        <div class="header-controls" id="headerControls">
          <div class="view-toggle" id="viewToggle">
            <button
              type="button"
              class="view-toggle-btn"
              :class="{ active: viewMode === 'list' }"
              data-view="list"
              :title="t('projectList.listView')"
              @click="setViewMode('list')"
            >
              <ListIcon :size="14" :stroke-width="2" />
            </button>
            <button
              type="button"
              class="view-toggle-btn"
              :class="{ active: viewMode === 'grid' }"
              data-view="grid"
              :title="t('projectList.gridView')"
              @click="setViewMode('grid')"
            >
              <LayoutGridIcon :size="14" :stroke-width="2" />
            </button>
          </div>
        </div>
      </div>
    </div>

    <div class="detail-scroll">
      <div id="listOutput" class="list-output">
        <div v-if="isLoading && !projects.length && !errorMessage" class="loading-state">
          <Spinner />
          {{ t('projectList.loading') }}
        </div>

        <div v-else-if="errorMessage" class="empty-state">
          <span class="project-list-view__state-icon project-list-view__state-icon--error">!</span>
          <p>
            {{ t('projectList.failedToLoad') }}<br>
            <span class="text-muted project-list-view__error-copy">{{ errorMessage }}</span>
          </p>
          <div class="project-list-view__state-actions">
            <button type="button" class="flu-btn flu-btn-sm flu-btn-accent" @click="refreshProjects">
              {{ t('general.retry') }}
            </button>
            <button type="button" class="flu-btn flu-btn-sm flu-btn-ghost" @click="Runtime.emit('menu:env')">
              {{ t('projectList.checkEnv') }}
            </button>
          </div>
        </div>

        <div v-else-if="!projects.length" class="empty-state">
          <span class="project-list-view__state-icon">i</span>
          <p>
            {{ t('projectList.empty') }}<br>
            <span v-html="t('projectList.emptyHint')" />
          </p>
        </div>

        <div v-else-if="viewMode === 'grid'" class="project-grid">
          <article
            v-for="project in projects"
            :key="getProjectName(project)"
            class="project-card"
            :class="{ stopped: isProjectStopped(project) }"
          >
            <RouterLink
              class="project-card-link proj-name"
              :to="{ name: 'project-detail', params: { name: getProjectName(project) } }"
            >
              <div class="project-card-logo">
                <ProjectTypeLogo
                  :type="getProjectType(project) || getProjectName(project)"
                  :size="36"
                  :style="{ marginRight: 0 }"
                />
              </div>
              <div class="project-card-name">{{ getProjectName(project) }}</div>
            </RouterLink>
            <div class="project-card-type">{{ getProjectType(project) || 'unknown' }}</div>
            <span :class="statusClass(getProjectStatus(project) || 'unknown')">
              {{ getProjectStatus(project) || 'unknown' }}
            </span>
            <div class="project-card-actions">
              <button
                type="button"
                class="flu-btn flu-btn-xs proj-action"
                :class="isProjectStopped(project) ? 'flu-btn-accent' : 'flu-btn-danger'"
                :title="isProjectStopped(project) ? t('projectList.start') : t('projectList.stop')"
                @click="runProjectAction(getProjectName(project), isProjectStopped(project) ? 'start' : 'stop')"
              >
                <PlayIcon v-if="isProjectStopped(project)" :size="14" :stroke-width="2" />
                <SquareIcon v-else :size="14" :stroke-width="2" />
              </button>
              <button
                type="button"
                class="flu-btn flu-btn-xs flu-btn-ghost proj-action"
                :title="t('projectList.restart')"
                @click="runProjectAction(getProjectName(project), 'restart')"
              >
                <RotateCwIcon :size="14" :stroke-width="2" />
              </button>
              <button
                type="button"
                class="flu-btn flu-btn-sm flu-btn-ghost proj-action"
                :title="t('detail.more.openEditor')"
                @click="openProjectEditor(project)"
              >
                <FileBracesCornerIcon :size="14" :stroke-width="2" />
              </button>
              <button
                type="button"
                class="flu-btn flu-btn-xs flu-btn-ghost proj-action"
                :title="t('projectList.openFolder')"
                @click="openProjectFolder(project)"
              >
                <FolderOpenIcon :size="14" :stroke-width="2" />
              </button>
              <button
                type="button"
                class="flu-btn flu-btn-xs flu-btn-ghost proj-action project-list-view__cli-button"
                :title="t('projectList.openCli')"
                @click="openProjectCli(project)"
              >
                <TerminalIcon :size="14" :stroke-width="2" />
              </button>
            </div>
          </article>
        </div>

        <div v-else class="flu-table-wrap">
          <table class="flu-table">
            <thead>
              <tr>
                <th>{{ t('projectList.colName') }}</th>
                <th>{{ t('projectList.colStatus') }}</th>
                <th>{{ t('projectList.colType') }}</th>
                <th>{{ t('projectList.colUrl') }}</th>
                <th>{{ t('projectList.colActions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="project in projects"
                :key="getProjectName(project)"
                :class="{ stopped: isProjectStopped(project) }"
              >
                <td>
                  <RouterLink
                    class="proj-name"
                    :to="{ name: 'project-detail', params: { name: getProjectName(project) } }"
                  >
                    {{ getProjectName(project) }}
                  </RouterLink>
                </td>
                <td>
                  <span :class="statusClass(getProjectStatus(project) || 'unknown')">
                    {{ getProjectStatus(project) || 'unknown' }}
                  </span>
                </td>
                <td>{{ getProjectType(project) || 'unknown' }}</td>
                <td>
                  <button
                    v-if="getPrimaryUrl(project)"
                    type="button"
                    class="proj-link project-list-view__link-button"
                    @click="openProjectUrl(getPrimaryUrl(project))"
                  >
                    {{ getPrimaryUrl(project).replace(/^https?:\/\//, '') }}
                  </button>
                </td>
                <td>
                  <div class="action-group project-list-view__table-actions">
                    <button
                      type="button"
                      class="flu-btn flu-btn-sm proj-action"
                      :class="isProjectStopped(project) ? 'flu-btn-accent' : 'flu-btn-danger'"
                      @click="runProjectAction(getProjectName(project), isProjectStopped(project) ? 'start' : 'stop')"
                    >
                      {{ isProjectStopped(project) ? t('projectList.start') : t('projectList.stop') }}
                    </button>
                    <button
                      type="button"
                      class="flu-btn flu-btn-sm flu-btn-ghost proj-action"
                      @click="runProjectAction(getProjectName(project), 'restart')"
                    >
                      {{ t('projectList.restart') }}
                    </button>
                    <button
                      type="button"
                      class="flu-btn flu-btn-sm flu-btn-ghost proj-action"
                      :title="t('detail.more.openEditor')"
                      @click="openProjectEditor(project)"
                    >
                      <FileBracesCornerIcon :size="14" :stroke-width="2" />
                    </button>
                    <button
                      type="button"
                      class="flu-btn flu-btn-sm flu-btn-ghost proj-action"
                      :title="t('projectList.openFolder')"
                      @click="openProjectFolder(project)"
                    >
                      <FolderOpenIcon :size="14" :stroke-width="2" />
                    </button>
                    <button
                      type="button"
                      class="flu-btn flu-btn-sm flu-btn-ghost proj-action project-list-view__cli-button"
                      :title="t('projectList.openCli')"
                      @click="openProjectCli(project)"
                    >
                      <TerminalIcon :size="14" :stroke-width="2" />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.project-list-view__state-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  border: 2px solid var(--accent);
  color: var(--accent);
  font-size: 1.3rem;
  font-weight: 700;
  line-height: 1;
  user-select: none;
}

.project-list-view__state-icon--error {
  border-color: var(--danger, #e74c3c);
  color: var(--danger, #e74c3c);
}

.project-list-view__state-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: 0.5rem;
}

.project-list-view__error-copy {
  font-size: 0.85em;
}

.project-list-view__link-button {
  border: 0;
  padding: 0;
  background: transparent;
  cursor: pointer;
  font: inherit;
  text-align: left;
}

.project-list-view__table-actions {
  justify-content: flex-end;
}

.project-list-view__cli-button {
  background: #111;
  border: 1px solid var(--border-strong);
}
</style>
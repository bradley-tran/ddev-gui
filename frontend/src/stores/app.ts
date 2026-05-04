import { defineStore } from 'pinia'
import { DEFAULT_APP_CONFIG, normalizeAppConfig } from '@/lib/config'
import type { AppConfig, AppModal, AppModals, CurrentView, DdevProject, LogEntry, LogLevel, ToastEntry, ToastType, ViewMode } from '@/lib/types'
import { parseProjectsJSON, uid } from '@/lib/utils'
import { ConfigService, DdevService } from '@/lib/wails'

const MAX_LOG_ENTRIES = 200

interface AppState {
  projectsJSON: string
  config: AppConfig
  logEntries: LogEntry[]
  toasts: ToastEntry[]
  currentView: CurrentView
  selectedProject: string | null
  terminalActive: boolean
  isConfigLoaded: boolean
  isProjectsLoaded: boolean
  isLoadingProjects: boolean
  modals: AppModals
}

export function defaultModals(): AppModals {
  return {
    newProject: false,
    envInfo: false,
    settings: false,
    about: false,
  }
}

export const useAppStore = defineStore('app', {
  state: (): AppState => ({
    projectsJSON: '',
    config: { ...DEFAULT_APP_CONFIG },
    logEntries: [],
    toasts: [],
    currentView: 'list',
    selectedProject: null,
    terminalActive: false,
    isConfigLoaded: false,
    isProjectsLoaded: false,
    isLoadingProjects: false,
    modals: defaultModals(),
  }),

  getters: {
    projects: (state): DdevProject[] => parseProjectsJSON(state.projectsJSON),
  },

  actions: {
    setProjectsJSON(payload: string) {
      this.projectsJSON = payload
    },

    setConfig(payload: AppConfig) {
      this.config = payload
    },

    patchConfig(payload: Partial<AppConfig>) {
      this.config = { ...this.config, ...payload }
    },

    addLog(entry: LogEntry) {
      const last = this.logEntries[this.logEntries.length - 1]
      if (last && last.message === entry.message && last.level === entry.level) {
        return
      }

      this.logEntries = [...this.logEntries, entry].slice(-MAX_LOG_ENTRIES)
    },

    appLog(message: string, level: LogLevel = 'info') {
      const now = new Date()
      this.addLog({
        id: uid(),
        timestamp: now.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' }),
        message,
        level,
      })
    },

    clearLog() {
      this.logEntries = []
    },

    addToast(entry: ToastEntry) {
      this.toasts = [...this.toasts, entry]
    },

    removeToast(id: string) {
      this.toasts = this.toasts.filter((toast) => toast.id !== id)
    },

    showToast(message: string, type: ToastType = 'info', duration = 4000) {
      const id = uid()
      this.addToast({ id, message, type, duration })
    },

    async saveConfigValue<K extends keyof AppConfig>(key: K, value: AppConfig[K]) {
      const previousValue = this.config[key]
      this.patchConfig({ [key]: value } as Partial<AppConfig>)

      try {
        await ConfigService.set(String(key), value)
      } catch (error) {
        this.patchConfig({ [key]: previousValue } as Partial<AppConfig>)
        const message = error instanceof Error ? error.message : String(error)
        this.appLog(`Failed to save ${String(key)}: ${message}`, 'error')
        throw error
      }
    },

    navigateToDetail(projectName: string) {
      this.currentView = 'detail'
      this.selectedProject = projectName
    },

    navigateToList() {
      this.currentView = 'list'
      this.selectedProject = null
    },

    setViewMode(viewMode: ViewMode) {
      this.config = { ...this.config, viewMode }
    },

    setTerminalActive(active: boolean) {
      this.terminalActive = active
    },

    openModal(modal: AppModal) {
      this.modals[modal] = true
    },

    closeModal(modal: AppModal) {
      this.modals[modal] = false
    },

    closeAllModals() {
      this.modals = defaultModals()
    },

    async loadConfig() {
      try {
        const json = await ConfigService.getAll()
        const raw = JSON.parse(json)
        this.setConfig(normalizeAppConfig(raw))
      } catch {
        this.setConfig(normalizeAppConfig(undefined))
      } finally {
        this.isConfigLoaded = true
      }
    },

    async refreshProjects() {
      this.isLoadingProjects = true
      try {
        this.setProjectsJSON(await DdevService.listJSON())
        this.isProjectsLoaded = true
      } catch (error) {
        const message = error instanceof Error ? error.message : String(error)
        this.appLog(`Failed to refresh projects: ${message}`, 'error')
        throw error
      } finally {
        this.isLoadingProjects = false
      }
    },
  },
})
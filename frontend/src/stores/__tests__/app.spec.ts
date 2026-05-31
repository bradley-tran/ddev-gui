import { describe, expect, it, beforeEach, vi, afterEach } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useAppStore, defaultModals } from '../app'
import { ConfigService, DdevService } from '@/lib/wails'
import { DEFAULT_APP_CONFIG } from '@/lib/config'
import * as utils from '@/lib/utils'
import type { LogEntry, ToastEntry } from '@/lib/types'

vi.mock('@/lib/wails', () => ({
  ConfigService: {
    getAll: vi.fn(),
    set: vi.fn(),
  },
  DdevService: {
    listJSON: vi.fn(),
  },
}))

vi.mock('@/lib/utils', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/lib/utils')>()
  return {
    ...actual,
    uid: vi.fn(() => 'mock-uid'),
  }
})

describe('app store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  describe('defaultModals', () => {
    it('should return initial state for all modals', () => {
      const modals = defaultModals()
      expect(modals).toEqual({
        newProject: false,
        envInfo: false,
        settings: false,
        about: false,
      })
    })
  })

  describe('initial state', () => {
    it('should initialize with default states', () => {
      const store = useAppStore()
      expect(store.projectsJSON).toBe('')
      expect(store.config).toEqual(DEFAULT_APP_CONFIG)
      expect(store.logEntries).toEqual([])
      expect(store.toasts).toEqual([])
      expect(store.currentView).toBe('list')
      expect(store.selectedProject).toBeNull()
      expect(store.terminalActive).toBe(false)
      expect(store.isConfigLoaded).toBe(false)
      expect(store.isProjectsLoaded).toBe(false)
      expect(store.isLoadingProjects).toBe(false)
      expect(store.modals).toEqual(defaultModals())
    })
  })

  describe('getters', () => {
    it('projects should parse projectsJSON', () => {
      const store = useAppStore()
      store.projectsJSON = JSON.stringify([{ name: 'test-project' }])
      expect(store.projects).toEqual([{ name: 'test-project' }])
    })

    it('projects should handle invalid JSON', () => {
      const store = useAppStore()
      store.projectsJSON = 'invalid json'
      expect(store.projects).toEqual([])
    })
  })

  describe('actions - state changes', () => {
    it('setProjectsJSON should update projectsJSON', () => {
      const store = useAppStore()
      store.setProjectsJSON('{"test": true}')
      expect(store.projectsJSON).toBe('{"test": true}')
    })

    it('setConfig should update config', () => {
      const store = useAppStore()
      const newConfig = { ...DEFAULT_APP_CONFIG, showLog: false }
      store.setConfig(newConfig)
      expect(store.config).toEqual(newConfig)
    })

    it('patchConfig should partially update config', () => {
      const store = useAppStore()
      store.patchConfig({ showLog: false })
      expect(store.config.showLog).toBe(false)
      expect(store.config.theme).toBe(DEFAULT_APP_CONFIG.theme)
    })

    it('setViewMode should update viewMode', () => {
      const store = useAppStore()
      store.setViewMode('grid')
      expect(store.config.viewMode).toBe('grid')
    })

    it('setTerminalActive should update terminalActive', () => {
      const store = useAppStore()
      store.setTerminalActive(true)
      expect(store.terminalActive).toBe(true)
    })
  })

  describe('actions - navigation', () => {
    it('navigateToDetail should set view to detail and selected project', () => {
      const store = useAppStore()
      store.navigateToDetail('my-project')
      expect(store.currentView).toBe('detail')
      expect(store.selectedProject).toBe('my-project')
    })

    it('navigateToList should set view to list and clear selected project', () => {
      const store = useAppStore()
      store.currentView = 'detail'
      store.selectedProject = 'my-project'

      store.navigateToList()

      expect(store.currentView).toBe('list')
      expect(store.selectedProject).toBeNull()
    })
  })

  describe('actions - logs', () => {
    it('addLog should add entry and limit to MAX_LOG_ENTRIES', () => {
      const store = useAppStore()
      const entry1: LogEntry = { id: '1', message: 'test 1', timestamp: '12:00:00', level: 'info' }

      store.addLog(entry1)
      expect(store.logEntries).toEqual([entry1])

      // Add 200 more entries
      for (let i = 0; i < 200; i++) {
        store.addLog({ id: String(i + 2), message: `test ${i + 2}`, timestamp: '12:00:00', level: 'info' })
      }

      expect(store.logEntries.length).toBe(200)
      expect(store.logEntries[0]?.id).toBe('2') // The first entry should be dropped
      expect(store.logEntries[199]?.id).toBe('201')
    })

    it('addLog should ignore duplicate consecutive entries', () => {
      const store = useAppStore()
      const entry1: LogEntry = { id: '1', message: 'test msg', timestamp: '12:00:00', level: 'info' }
      const entry2: LogEntry = { id: '2', message: 'test msg', timestamp: '12:00:01', level: 'info' }

      store.addLog(entry1)
      store.addLog(entry2)

      expect(store.logEntries).toEqual([entry1])
    })

    it('appLog should create and add a log entry with current time and uid', () => {
      const date = new Date('2024-01-01T14:30:45')
      vi.setSystemTime(date)

      const store = useAppStore()
      store.appLog('test message', 'error')

      expect(store.logEntries.length).toBe(1)
      expect(store.logEntries[0]).toEqual({
        id: 'mock-uid',
        timestamp: date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' }),
        message: 'test message',
        level: 'error',
      })
    })

    it('appLog should default to info level', () => {
      const store = useAppStore()
      store.appLog('test message')
      expect(store.logEntries[0]?.level).toBe('info')
    })

    it('clearLog should remove all log entries', () => {
      const store = useAppStore()
      store.appLog('test message')
      expect(store.logEntries.length).toBe(1)

      store.clearLog()
      expect(store.logEntries.length).toBe(0)
    })
  })

  describe('actions - toasts', () => {
    it('addToast should append a toast entry', () => {
      const store = useAppStore()
      const toast: ToastEntry = { id: '1', message: 'test', type: 'info', duration: 1000 }
      store.addToast(toast)
      expect(store.toasts).toEqual([toast])
    })

    it('removeToast should remove a toast entry by id', () => {
      const store = useAppStore()
      const toast1: ToastEntry = { id: '1', message: 'test 1', type: 'info', duration: 1000 }
      const toast2: ToastEntry = { id: '2', message: 'test 2', type: 'info', duration: 1000 }
      store.toasts = [toast1, toast2]

      store.removeToast('1')
      expect(store.toasts).toEqual([toast2])
    })

    it('showToast should create a toast entry with default values', () => {
      const store = useAppStore()
      store.showToast('test message')

      expect(store.toasts.length).toBe(1)
      expect(store.toasts[0]).toEqual({
        id: 'mock-uid',
        message: 'test message',
        type: 'info',
        duration: 4000,
      })
    })

    it('showToast should accept custom type and duration', () => {
      const store = useAppStore()
      store.showToast('test error', 'error', 5000)

      expect(store.toasts.length).toBe(1)
      expect(store.toasts[0]).toEqual({
        id: 'mock-uid',
        message: 'test error',
        type: 'error',
        duration: 5000,
      })
    })
  })

  describe('actions - modals', () => {
    it('openModal should set modal state to true', () => {
      const store = useAppStore()
      store.openModal('settings')
      expect(store.modals.settings).toBe(true)
    })

    it('closeModal should set modal state to false', () => {
      const store = useAppStore()
      store.modals.settings = true
      store.closeModal('settings')
      expect(store.modals.settings).toBe(false)
    })

    it('closeAllModals should reset all modal states to false', () => {
      const store = useAppStore()
      store.modals.settings = true
      store.modals.about = true

      store.closeAllModals()

      expect(store.modals).toEqual(defaultModals())
      expect(store.modals.settings).toBe(false)
      expect(store.modals.about).toBe(false)
    })
  })

  describe('actions - async wails interactions', () => {
    describe('saveConfigValue', () => {
      it('should save config value successfully', async () => {
        const store = useAppStore()
        vi.mocked(ConfigService.set).mockResolvedValueOnce('')

        await store.saveConfigValue('showLog', false)

        expect(store.config.showLog).toBe(false)
        expect(ConfigService.set).toHaveBeenCalledWith('showLog', false)
      })

      it('should revert config value and log error if save fails', async () => {
        const store = useAppStore()
        const error = new Error('Save failed')
        vi.mocked(ConfigService.set).mockRejectedValueOnce(error)

        const initialShowLog = store.config.showLog

        await expect(store.saveConfigValue('showLog', false)).rejects.toThrow('Save failed')

        expect(store.config.showLog).toBe(initialShowLog)
        expect(store.logEntries[0]?.message).toContain('Failed to save showLog: Save failed')
        expect(store.logEntries[0]?.level).toBe('error')
      })
    })

    describe('loadConfig', () => {
      it('should load config successfully', async () => {
        const store = useAppStore()
        const rawConfig = { showLog: false, theme: 'dark' }
        vi.mocked(ConfigService.getAll).mockResolvedValueOnce(JSON.stringify(rawConfig))

        await store.loadConfig()

        expect(store.isConfigLoaded).toBe(true)
        expect(store.config.showLog).toBe(false) // from rawConfig
        // normalizeAppConfig handles invalid enums, 'dark' should fallback to 'default'
        expect(store.config.theme).toBe('default')
      })

      it('should use fallback config if load fails', async () => {
        const store = useAppStore()
        vi.mocked(ConfigService.getAll).mockRejectedValueOnce(new Error('Load failed'))

        await store.loadConfig()

        expect(store.isConfigLoaded).toBe(true)
        expect(store.config).toEqual(DEFAULT_APP_CONFIG)
      })
    })

    describe('refreshProjects', () => {
      it('should refresh projects successfully', async () => {
        const store = useAppStore()
        const projectsData = '[{"name":"proj1"}]'
        vi.mocked(DdevService.listJSON).mockResolvedValueOnce(projectsData)

        expect(store.isLoadingProjects).toBe(false)

        const promise = store.refreshProjects()

        expect(store.isLoadingProjects).toBe(true)

        await promise

        expect(store.projectsJSON).toBe(projectsData)
        expect(store.isProjectsLoaded).toBe(true)
        expect(store.isLoadingProjects).toBe(false)
      })

      it('should log error and throw if refresh fails', async () => {
        const store = useAppStore()
        const error = new Error('Refresh failed')
        vi.mocked(DdevService.listJSON).mockRejectedValueOnce(error)

        await expect(store.refreshProjects()).rejects.toThrow('Refresh failed')

        expect(store.isLoadingProjects).toBe(false)
        expect(store.logEntries[0]?.message).toContain('Failed to refresh projects: Refresh failed')
        expect(store.logEntries[0]?.level).toBe('error')
      })
    })
  })
})

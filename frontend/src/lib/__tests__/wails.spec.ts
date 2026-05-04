import { describe, expect, it, vi, beforeEach } from 'vitest'
import { DdevService, ConfigService, Runtime } from '../wails'

describe('wails bridge', () => {
  const getMockDdevService = () => window.go!.backend!.DdevService! as any
  const getMockConfigService = () => window.go!.backend!.ConfigService! as any
  const getMockRuntime = () => window.runtime! as any

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('DdevService', () => {
    it('should call ListJSON', async () => {
      await DdevService.listJSON()
      expect(getMockDdevService().ListJSON).toHaveBeenCalled()
    })

    it('should call DescribeJSON', async () => {
      await DdevService.describeJSON('proj')
      expect(getMockDdevService().DescribeJSON).toHaveBeenCalledWith('proj')
    })

    it('should call Status', async () => {
      await DdevService.status('proj')
      expect(getMockDdevService().Status).toHaveBeenCalledWith('proj')
    })

    it('should call Start', async () => {
      await DdevService.start('proj')
      expect(getMockDdevService().Start).toHaveBeenCalledWith('proj')
    })

    it('should call Stop', async () => {
      await DdevService.stop('proj')
      expect(getMockDdevService().Stop).toHaveBeenCalledWith('proj')
    })

    it('should call Restart', async () => {
      await DdevService.restart('proj')
      expect(getMockDdevService().Restart).toHaveBeenCalledWith('proj')
    })

    it('should call PowerOff', async () => {
      await DdevService.powerOff()
      expect(getMockDdevService().PowerOff).toHaveBeenCalled()
    })

    it('should call AddonsJSON', async () => {
      await DdevService.addonsJSON('proj')
      expect(getMockDdevService().AddonsJSON).toHaveBeenCalledWith('proj')
    })

    it('should call AddonsAvailableJSON', async () => {
      await DdevService.addonsAvailableJSON('proj')
      expect(getMockDdevService().AddonsAvailableJSON).toHaveBeenCalledWith('proj')
    })

    it('should call AddonInstall', async () => {
      await DdevService.addonInstall('proj', 'addon')
      expect(getMockDdevService().AddonInstall).toHaveBeenCalledWith('proj', 'addon')
    })

    it('should call AddonRemove', async () => {
      await DdevService.addonRemove('proj', 'addon')
      expect(getMockDdevService().AddonRemove).toHaveBeenCalledWith('proj', 'addon')
    })

    it('should call ComposerInstall', async () => {
      await DdevService.composerInstall('proj', 'php')
      expect(getMockDdevService().ComposerInstall).toHaveBeenCalledWith('proj', 'php')
    })

    it('should call ConfigureProject', async () => {
      await DdevService.configureProject('name', 'type', 'docroot', 'php')
      expect(getMockDdevService().ConfigureProject).toHaveBeenCalledWith('~', 'name', 'type', 'docroot', 'php')
    })

    it('should call ModifyProject', async () => {
      await DdevService.modifyProject('name', 'php', 'node', 'type', 'docroot')
      expect(getMockDdevService().ModifyProject).toHaveBeenCalledWith('name', 'php', 'node', 'type', 'docroot')
    })

    it('should call ConfigureServices', async () => {
      await DdevService.configureServices('name', '8080', '3307', true, true, false)
      expect(getMockDdevService().ConfigureServices).toHaveBeenCalledWith('name', '8080', '3307', true, true, false)
    })

    it('should call CloneRepo', async () => {
      await DdevService.cloneRepo('name', 'url')
      expect(getMockDdevService().CloneRepo).toHaveBeenCalledWith('name', 'url')
    })

    it('should call DdevInstalledVersion', async () => {
      await DdevService.ddevInstalledVersion()
      expect(getMockDdevService().DdevInstalledVersion).toHaveBeenCalled()
    })

    it('should call InstallDdev', async () => {
      await DdevService.installDdev()
      expect(getMockDdevService().InstallDdev).toHaveBeenCalled()
    })

    it('should call DeleteProject', async () => {
      await DdevService.deleteProject('name')
      expect(getMockDdevService().DeleteProject).toHaveBeenCalledWith('name')
    })

    it('should call ExportDB', async () => {
      await DdevService.exportDB('name')
      expect(getMockDdevService().ExportDB).toHaveBeenCalledWith('name')
    })

    it('should call ImportDBSelectFile', async () => {
      await DdevService.importDBSelectFile('name')
      expect(getMockDdevService().ImportDBSelectFile).toHaveBeenCalledWith('name')
    })

    it('should call ImportDBFromFile', async () => {
      await DdevService.importDBFromFile('name', 'path')
      expect(getMockDdevService().ImportDBFromFile).toHaveBeenCalledWith('name', 'path')
    })

    it('should call DrushUli', async () => {
      await DdevService.drushUli('name')
      expect(getMockDdevService().DrushUli).toHaveBeenCalledWith('name')
    })

    it('should call DrushUliAsUser', async () => {
      await DdevService.drushUliAsUser('name', '1')
      expect(getMockDdevService().DrushUliAsUser).toHaveBeenCalledWith('name', '1')
    })

    it('should call DrushRecentUsers', async () => {
      await DdevService.drushRecentUsers('name')
      expect(getMockDdevService().DrushRecentUsers).toHaveBeenCalledWith('name')
    })

    it('should call DrushSiteInstall', async () => {
      await DdevService.drushSiteInstall('name')
      expect(getMockDdevService().DrushSiteInstall).toHaveBeenCalledWith('name')
    })

    it('should call DrushCacheRebuild', async () => {
      await DdevService.drushCacheRebuild('name')
      expect(getMockDdevService().DrushCacheRebuild).toHaveBeenCalledWith('name')
    })

    it('should call WpCoreInstall', async () => {
      await DdevService.wpCoreInstall('name')
      expect(getMockDdevService().WpCoreInstall).toHaveBeenCalledWith('name')
    })

    it('should call LaravelInit', async () => {
      await DdevService.laravelInit('name')
      expect(getMockDdevService().LaravelInit).toHaveBeenCalledWith('name')
    })

    it('should call SnapshotListJSON', async () => {
      await DdevService.snapshotListJSON('name')
      expect(getMockDdevService().SnapshotListJSON).toHaveBeenCalledWith('name')
    })

    it('should call SnapshotCreate', async () => {
      await DdevService.snapshotCreate('name', 'snap')
      expect(getMockDdevService().SnapshotCreate).toHaveBeenCalledWith('name', 'snap')
    })

    it('should call SnapshotRestore', async () => {
      await DdevService.snapshotRestore('name', 'snap')
      expect(getMockDdevService().SnapshotRestore).toHaveBeenCalledWith('name', 'snap')
    })

    it('should call SnapshotDelete', async () => {
      await DdevService.snapshotDelete('name', 'snap')
      expect(getMockDdevService().SnapshotDelete).toHaveBeenCalledWith('name', 'snap')
    })

    it('should call ProjectLogs', async () => {
      await DdevService.ProjectLogs('name', 'db')
      expect(getMockDdevService().ProjectLogs).toHaveBeenCalledWith('name', 'db')
    })

    it('should call ActiveBackend', async () => {
      await DdevService.activeBackend()
      expect(getMockDdevService().ActiveBackend).toHaveBeenCalled()
    })

    it('should call ListWSLDistros', async () => {
      await DdevService.listWSLDistros()
      expect(getMockDdevService().ListWSLDistros).toHaveBeenCalled()
    })

    it('should call ReloadBackend', async () => {
      await DdevService.reloadBackend()
      expect(getMockDdevService().ReloadBackend).toHaveBeenCalled()
    })

    it('should call AppVersion', async () => {
      await DdevService.appVersion()
      expect(getMockDdevService().AppVersion).toHaveBeenCalled()
    })

    it('should call ListDir', async () => {
      await DdevService.listDir('proj', 'path')
      expect(getMockDdevService().ListDir).toHaveBeenCalledWith('proj', 'path')
    })

    it('should call ReadFile', async () => {
      await DdevService.readFile('proj', 'path')
      expect(getMockDdevService().ReadFile).toHaveBeenCalledWith('proj', 'path')
    })

    it('should call ReadFileBase64', async () => {
      await DdevService.readFileBase64('proj', 'path')
      expect(getMockDdevService().ReadFileBase64).toHaveBeenCalledWith('proj', 'path')
    })

    it('should call ExecCommand', async () => {
      await DdevService.execCommand('proj', 'cmd')
      expect(getMockDdevService().ExecCommand).toHaveBeenCalledWith('proj', 'cmd')
    })
  })

  describe('ConfigService', () => {
    it('should call GetAll', async () => {
      await ConfigService.getAll()
      expect(getMockConfigService().GetAll).toHaveBeenCalled()
    })

    it('should call Set', async () => {
      await ConfigService.set('key', 'val')
      expect(getMockConfigService().Set).toHaveBeenCalledWith('key', 'val')
    })

    it('should call SetProjectConfig', async () => {
      await ConfigService.setProjectConfig('proj', 'key', 'val')
      expect(getMockConfigService().SetProjectConfig).toHaveBeenCalledWith('proj', 'key', 'val')
    })
  })

  describe('ensureBinding', () => {
    it('should throw error if DdevService is missing', async () => {
      const originalGo = window.go
      try {
        // @ts-ignore
        window.go = undefined
        expect(() => DdevService.listJSON()).toThrow('window.go.backend.DdevService is not available')
      } finally {
        window.go = originalGo
      }
    })

    it('should throw error if ConfigService is missing', async () => {
      const originalGo = window.go
      try {
        // @ts-ignore
        window.go = { backend: {} }
        expect(() => ConfigService.getAll()).toThrow('window.go.backend.ConfigService is not available')
      } finally {
        window.go = originalGo
      }
    })
  })

  describe('Runtime', () => {
    it('should call on and trigger callback', () => {
      const cb = vi.fn()
      Runtime.on('event1', cb)
      expect(getMockRuntime().EventsOn).toHaveBeenCalledWith('event1', expect.any(Function))

      const bridge = getMockRuntime().EventsOn.mock.calls.find((c: any) => c[0] === 'event1')[1]
      bridge('data')
      expect(cb).toHaveBeenCalledWith('data')
      Runtime.off('event1')
    })

    it('should handle off with specific callback', () => {
      const cb1 = vi.fn()
      const cb2 = vi.fn()
      Runtime.on('event2', cb1)
      Runtime.on('event2', cb2)

      // Find the bridge for event2
      const call = getMockRuntime().EventsOn.mock.calls.find((c: any) => c[0] === 'event2')
      const bridge = call[1]

      Runtime.off('event2', cb1)
      expect(getMockRuntime().EventsOff).not.toHaveBeenCalled()

      bridge('data')
      expect(cb1).not.toHaveBeenCalled()
      expect(cb2).toHaveBeenCalledWith('data')

      Runtime.off('event2', cb2)
      expect(getMockRuntime().EventsOff).toHaveBeenCalledWith('event2')
    })

    it('should handle off without callback (all listeners)', () => {
      const cb = vi.fn()
      Runtime.on('event3', cb)
      Runtime.off('event3')
      expect(getMockRuntime().EventsOff).toHaveBeenCalledWith('event3')
    })

    it('should call emit', () => {
      Runtime.emit('event-emit', 'data')
      expect(getMockRuntime().EventsEmit).toHaveBeenCalledWith('event-emit', 'data')
    })

    it('should call minimise', () => {
      Runtime.minimise()
      expect(getMockRuntime().WindowMinimise).toHaveBeenCalled()
    })

    it('should call toggleMaximise', () => {
      Runtime.toggleMaximise()
      expect(getMockRuntime().WindowToggleMaximise).toHaveBeenCalled()
    })

    it('should call isMaximised', async () => {
      getMockRuntime().WindowIsMaximised.mockResolvedValue(true)
      const result = await Runtime.isMaximised()
      expect(result).toBe(true)
      expect(getMockRuntime().WindowIsMaximised).toHaveBeenCalled()
    })

    it('should call quit', () => {
      Runtime.quit()
      expect(getMockRuntime().Quit).toHaveBeenCalled()
    })
  })
})

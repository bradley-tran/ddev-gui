import { beforeEach, vi } from 'vitest'

window.HTMLElement.prototype.scrollIntoView = vi.fn()

const mockRuntime = {
  EventsOn: vi.fn(),
  EventsOff: vi.fn(),
  EventsEmit: vi.fn(),
  WindowMinimise: vi.fn(),
  WindowToggleMaximise: vi.fn(),
  WindowIsMaximised: vi.fn().mockResolvedValue(false),
  Quit: vi.fn(),
}

const mockDdevService = {
  ListJSON: vi.fn().mockResolvedValue('[]'),
  ListDir: vi.fn().mockResolvedValue('[]'),
  ReadFile: vi.fn().mockResolvedValue(''),
  ReadFileBase64: vi.fn().mockResolvedValue(''),
  DescribeJSON: vi.fn().mockResolvedValue('{}'),
  Status: vi.fn().mockResolvedValue(''),
  Start: vi.fn().mockResolvedValue(''),
  Stop: vi.fn().mockResolvedValue(''),
  Restart: vi.fn().mockResolvedValue(''),
  PowerOff: vi.fn().mockResolvedValue(''),
  AddonsJSON: vi.fn().mockResolvedValue('[]'),
  AddonsAvailableJSON: vi.fn().mockResolvedValue('[]'),
  AddonInstall: vi.fn().mockResolvedValue(''),
  AddonRemove: vi.fn().mockResolvedValue(''),
  ComposerInstall: vi.fn().mockResolvedValue(''),
  ConfigureProject: vi.fn().mockResolvedValue(''),
  ModifyProject: vi.fn().mockResolvedValue(''),
  ConfigureServices: vi.fn().mockResolvedValue(''),
  CloneRepo: vi.fn().mockResolvedValue(''),
  DdevInstalledVersion: vi.fn().mockResolvedValue(''),
  InstallDdev: vi.fn().mockResolvedValue(''),
  DeleteProject: vi.fn().mockResolvedValue(''),
  ExportDB: vi.fn().mockResolvedValue(''),
  ImportDBSelectFile: vi.fn().mockResolvedValue(''),
  ImportDBFromFile: vi.fn().mockResolvedValue(''),
  DrushUli: vi.fn().mockResolvedValue(''),
  DrushUliAsUser: vi.fn().mockResolvedValue(''),
  DrushRecentUsers: vi.fn().mockResolvedValue('[]'),
  DrushSiteInstall: vi.fn().mockResolvedValue(''),
  DrushCacheRebuild: vi.fn().mockResolvedValue(''),
  WpCoreInstall: vi.fn().mockResolvedValue(''),
  LaravelInit: vi.fn().mockResolvedValue(''),
  SnapshotListJSON: vi.fn().mockResolvedValue('[]'),
  SnapshotCreate: vi.fn().mockResolvedValue(''),
  SnapshotRestore: vi.fn().mockResolvedValue(''),
  SnapshotDelete: vi.fn().mockResolvedValue(''),
  ProjectLogs: vi.fn().mockResolvedValue(''),
  ActiveBackend: vi.fn().mockResolvedValue('wsl'),
  WSLExists: vi.fn().mockResolvedValue(true),
  ListWSLDistros: vi.fn().mockResolvedValue([]),
  ReloadBackend: vi.fn().mockResolvedValue(undefined),
  AppVersion: vi.fn().mockResolvedValue({ version: 'test', commitHash: 'abc1234' }),
  ExecCommand: vi.fn().mockResolvedValue(''),
}

const mockConfigService = {
  GetAll: vi.fn().mockResolvedValue('{}'),
  Set: vi.fn().mockResolvedValue(''),
  SetProjectConfig: vi.fn().mockResolvedValue(''),
}

Object.defineProperty(window, 'runtime', {
  value: mockRuntime,
  writable: true,
})

Object.defineProperty(window, 'go', {
  value: {
    backend: {
      DdevService: mockDdevService,
      ConfigService: mockConfigService,
    },
  },
  writable: true,
})

Object.defineProperty(window, 'open', {
  value: vi.fn(),
  writable: true,
})

Object.defineProperty(window, 'scrollTo', {
  value: vi.fn(),
  writable: true,
})

beforeEach(() => {
  vi.clearAllMocks()
  mockDdevService.ListJSON.mockResolvedValue('[]')
  mockDdevService.DescribeJSON.mockResolvedValue('{}')
  mockDdevService.SnapshotListJSON.mockResolvedValue('[]')
  mockDdevService.ProjectLogs.mockResolvedValue('')
  mockDdevService.AddonsJSON.mockResolvedValue('[]')
  mockConfigService.GetAll.mockResolvedValue('{}')
  mockRuntime.WindowIsMaximised.mockResolvedValue(false)
})
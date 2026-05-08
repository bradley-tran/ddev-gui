declare global {
  interface Window {
    go?: {
      backend?: {
        DdevService?: WailsDdevService
        ConfigService?: WailsConfigService
      }
    }
    runtime?: WailsRuntime
  }
}

interface WailsDdevService {
  ListJSON(): Promise<string>
  ListDir(project: string, relPath: string): Promise<string>
  ReadFile(project: string, relPath: string): Promise<string>
  ReadFileBase64(project: string, relPath: string): Promise<string>
  DescribeJSON(name: string): Promise<string>
  Status(name: string): Promise<string>
  Start(name: string): Promise<string>
  Stop(name: string): Promise<string>
  Restart(name: string): Promise<string>
  PowerOff(): Promise<string>
  AddonsJSON(name: string): Promise<string>
  AddonsAvailableJSON(name: string): Promise<string>
  AddonInstall(name: string, addon: string): Promise<string>
  AddonRemove(name: string, addon: string): Promise<string>
  ComposerInstall(name: string, projType: string): Promise<string>
  ConfigureProject(dir: string, name: string, projType: string, docroot: string, phpVersion: string): Promise<string>
  ModifyProject(name: string, phpVersion: string, nodejsVersion: string, projectType: string, docroot: string): Promise<string>
  ConfigureServices(
    name: string,
    webPort: string,
    dbPort: string,
    xdebugEnabled: boolean,
    xhprofEnabled: boolean,
    xhguiEnabled: boolean,
  ): Promise<string>
  CloneRepo(name: string, repoURL: string): Promise<string>
  DdevInstalledVersion(): Promise<string>
  InstallDdev(): Promise<string>
  DeleteProject(name: string): Promise<string>
  ExportDB(name: string): Promise<string>
  ImportDBSelectFile(name: string): Promise<string>
  ImportDBFromFile(name: string, filePath: string): Promise<string>
  DrushUli(name: string): Promise<string>
  DrushUliAsUser(name: string, uid: string): Promise<string>
  DrushRecentUsers(name: string): Promise<string>
  DrushSiteInstall(name: string): Promise<string>
  DrushCacheRebuild(name: string): Promise<string>
  WpCoreInstall(name: string): Promise<string>
  LaravelInit(name: string): Promise<string>
  SnapshotListJSON(name: string): Promise<string>
  SnapshotCreate(name: string, snapName: string): Promise<string>
  SnapshotRestore(name: string, snapName: string): Promise<string>
  SnapshotDelete(name: string, snapName: string): Promise<string>
  ProjectLogs(name: string, service?: string): Promise<string>
  ActiveBackend(): Promise<string>
  WSLExists(): Promise<boolean>
  ListWSLDistros(): Promise<string[]>
  ReloadBackend(): Promise<void>
  AppVersion(): Promise<{ version: string; commitHash: string }>
  ExecCommand(project: string, command: string): Promise<string>
}

interface WailsConfigService {
  GetAll(): Promise<string>
  Set(key: string, value: unknown): Promise<string>
  SetProjectConfig(project: string, key: string, value: unknown): Promise<string>
}

interface WailsRuntime {
  EventsOn(event: string, callback: (...args: unknown[]) => void): void
  EventsOff(event: string): void
  EventsEmit(event: string, data?: unknown): void
  WindowMinimise(): void
  WindowToggleMaximise(): void
  WindowIsMaximised(): Promise<boolean>
  Quit(): void
}

function ensureBinding<T>(binding: T | undefined, name: string): T {
  if (!binding) {
    throw new Error(`${name} is not available. Run the Vue frontend through Wails or provide a test mock.`)
  }

  return binding
}

function getDdevService(): WailsDdevService {
  return ensureBinding(window.go?.backend?.DdevService, 'window.go.backend.DdevService')
}

function getConfigService(): WailsConfigService {
  return ensureBinding(window.go?.backend?.ConfigService, 'window.go.backend.ConfigService')
}

type RuntimeCallback = (...args: unknown[]) => void

const runtimeListeners = new Map<string, Set<RuntimeCallback>>()
const runtimeBridges = new Map<string, RuntimeCallback>()

function ensureRuntimeBridge(event: string): RuntimeCallback {
  const existingBridge = runtimeBridges.get(event)
  if (existingBridge) return existingBridge

  const bridge: RuntimeCallback = (...args) => {
    const listeners = runtimeListeners.get(event)
    if (!listeners) return
    for (const listener of Array.from(listeners)) {
      listener(...args)
    }
  }

  runtimeBridges.set(event, bridge)
  window.runtime?.EventsOn(event, bridge)
  return bridge
}

export const DdevService = {
  listJSON: () => getDdevService().ListJSON(),
  describeJSON: (name: string) => getDdevService().DescribeJSON(name),
  status: (name: string) => getDdevService().Status(name),
  start: (name: string) => getDdevService().Start(name),
  stop: (name: string) => getDdevService().Stop(name),
  restart: (name: string) => getDdevService().Restart(name),
  powerOff: () => getDdevService().PowerOff(),
  addonsJSON: (name: string) => getDdevService().AddonsJSON(name),
  addonsAvailableJSON: (name: string) => getDdevService().AddonsAvailableJSON(name),
  addonInstall: (name: string, addon: string) => getDdevService().AddonInstall(name, addon),
  addonRemove: (name: string, addon: string) => getDdevService().AddonRemove(name, addon),
  composerInstall: (name: string, projType: string) => getDdevService().ComposerInstall(name, projType),
  configureProject: (name: string, projType: string, docroot: string, phpVersion: string) =>
    getDdevService().ConfigureProject('~', name, projType, docroot, phpVersion),
  modifyProject: (name: string, phpVersion: string, nodejsVersion: string, projectType: string, docroot: string) =>
    getDdevService().ModifyProject(name, phpVersion, nodejsVersion, projectType, docroot),
  configureServices: (
    name: string,
    webPort: string,
    dbPort: string,
    xdebugEnabled: boolean,
    xhprofEnabled: boolean,
    xhguiEnabled: boolean,
  ) => getDdevService().ConfigureServices(name, webPort, dbPort, xdebugEnabled, xhprofEnabled, xhguiEnabled),
  cloneRepo: (name: string, repoURL: string) => getDdevService().CloneRepo(name, repoURL),
  ddevInstalledVersion: () => getDdevService().DdevInstalledVersion(),
  installDdev: () => getDdevService().InstallDdev(),
  deleteProject: (name: string) => getDdevService().DeleteProject(name),
  exportDB: (name: string) => getDdevService().ExportDB(name),
  importDBSelectFile: (name: string) => getDdevService().ImportDBSelectFile(name),
  importDBFromFile: (name: string, filePath: string) => getDdevService().ImportDBFromFile(name, filePath),
  drushUli: (name: string) => getDdevService().DrushUli(name),
  drushUliAsUser: (name: string, uid: string) => getDdevService().DrushUliAsUser(name, uid),
  drushRecentUsers: (name: string) => getDdevService().DrushRecentUsers(name),
  drushSiteInstall: (name: string) => getDdevService().DrushSiteInstall(name),
  drushCacheRebuild: (name: string) => getDdevService().DrushCacheRebuild(name),
  wpCoreInstall: (name: string) => getDdevService().WpCoreInstall(name),
  laravelInit: (name: string) => getDdevService().LaravelInit(name),
  snapshotListJSON: (name: string) => getDdevService().SnapshotListJSON(name),
  snapshotCreate: (name: string, snapName: string) => getDdevService().SnapshotCreate(name, snapName),
  snapshotRestore: (name: string, snapName: string) => getDdevService().SnapshotRestore(name, snapName),
  snapshotDelete: (name: string, snapName: string) => getDdevService().SnapshotDelete(name, snapName),
  ProjectLogs: (name: string, service = 'web') => getDdevService().ProjectLogs(name, service),
  activeBackend: () => getDdevService().ActiveBackend(),
  wslExists: () => getDdevService().WSLExists(),
  listWSLDistros: () => getDdevService().ListWSLDistros(),
  reloadBackend: () => getDdevService().ReloadBackend(),
  appVersion: () => getDdevService().AppVersion(),
  listDir: (project: string, relPath: string) => getDdevService().ListDir(project, relPath),
  readFile: (project: string, relPath: string) => getDdevService().ReadFile(project, relPath),
  readFileBase64: (project: string, relPath: string) => getDdevService().ReadFileBase64(project, relPath),
  execCommand: (project: string, command: string) => getDdevService().ExecCommand(project, command),
}

export const ConfigService = {
  getAll: () => getConfigService().GetAll(),
  set: (key: string, value: unknown) => getConfigService().Set(key, value),
  setProjectConfig: (project: string, key: string, value: unknown) =>
    getConfigService().SetProjectConfig(project, key, value),
}

export const Runtime = {
  on: (event: string, cb: RuntimeCallback) => {
    const listeners = runtimeListeners.get(event) ?? new Set<RuntimeCallback>()
    runtimeListeners.set(event, listeners)
    listeners.add(cb)
    ensureRuntimeBridge(event)
  },
  off: (event: string, cb?: RuntimeCallback) => {
    if (!cb) {
      runtimeListeners.delete(event)
      runtimeBridges.delete(event)
      window.runtime?.EventsOff(event)
      return
    }

    const listeners = runtimeListeners.get(event)
    if (!listeners) return

    listeners.delete(cb)
    if (listeners.size === 0) {
      runtimeListeners.delete(event)
      runtimeBridges.delete(event)
      window.runtime?.EventsOff(event)
    }
  },
  emit: (event: string, data?: unknown) => window.runtime?.EventsEmit(event, data),
  minimise: () => window.runtime?.WindowMinimise(),
  toggleMaximise: () => window.runtime?.WindowToggleMaximise(),
  isMaximised: () => window.runtime?.WindowIsMaximised() ?? Promise.resolve(false),
  quit: () => window.runtime?.Quit(),
}
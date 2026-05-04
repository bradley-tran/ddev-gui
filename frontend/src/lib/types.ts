export type ViewMode = 'list' | 'grid'
export type BackendType = 'wsl' | 'ssh' | 'local'
export type ThemeType = 'default' | 'acrylic' | 'tabbed'
export type LocaleType = 'en' | 'zh' | 'vi'
export type PreferredEditorType = 'vscode' | 'phpstorm' | 'neovim' | 'sublime' | 'antigravity'
export type LogLevel = 'info' | 'success' | 'error' | 'output'
export type ToastType = 'success' | 'error' | 'info'
export type CurrentView = 'list' | 'detail'
export type AppModal = 'newProject' | 'envInfo' | 'settings' | 'about'

export interface SshConfig {
  host: string
  port: string
  user: string
  keyPath: string
}

export interface LogEntry {
  id: string
  timestamp: string
  message: string
  level: LogLevel
}

export interface ToastEntry {
  id: string
  message: string
  type: ToastType
  duration: number
}

export interface AppConfig {
  viewMode?: ViewMode
  showLog?: boolean
  openLinksInBrowser?: boolean
  ddevTelemetryOptIn?: boolean
  devMode?: boolean
  preferredEditor?: PreferredEditorType
  backend?: BackendType
  wslDistro?: string
  ssh?: SshConfig
  projects?: Record<string, ProjectConfig>
  theme?: ThemeType
  locale?: LocaleType
}

export interface AppModals {
  newProject: boolean
  envInfo: boolean
  settings: boolean
  about: boolean
}

export interface ProjectConfig {
  initialized?: boolean
}

export interface DdevProject {
  name?: string
  project?: string
  projectname?: string
  status?: string
  status_desc?: string
  state?: string
  type?: string
  projecttype?: string
  docroot?: string
  approot?: string
  shortroot?: string
  httpurl?: string
  httpsurl?: string
  url?: string
  primary_url?: string
  mailpit_url?: string
  mailpit_https_url?: string
  router?: string
  php_version?: string
  nodejs_version?: string
  services?: Record<string, DdevService>
  [key: string]: unknown
}

export interface DdevService {
  status?: string
  http_url?: string
  https_url?: string
  host_http_url?: string
  host_https_url?: string
  exposed_ports?: string
  host_ports?: string
  [key: string]: unknown
}

export interface DdevAddon {
  name?: string
  Name?: string
  version?: string
  Version?: string
  repository?: string
  Repository?: string
  full_name?: string
  FullName?: string
  repo?: string
  source?: string
  user?: string
  User?: string
  installed_date?: string
  InstalledDate?: string
  installedDate?: string
  installed?: string
  date?: string
  [key: string]: unknown
}

export interface DdevSnapshot {
  name?: string
  Name?: string
  snapshot_name?: string
  snapshotName?: string
  [key: string]: unknown
}
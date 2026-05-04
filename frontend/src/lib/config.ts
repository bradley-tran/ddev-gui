import type { AppConfig, BackendType, LocaleType, PreferredEditorType, ThemeType, ViewMode } from './types'
import { coerceToBool } from './utils'

const VIEW_MODES = ['list', 'grid'] as const satisfies readonly ViewMode[]
const THEMES = ['default', 'acrylic', 'tabbed'] as const satisfies readonly ThemeType[]
const LOCALES = ['en', 'zh', 'vi'] as const satisfies readonly LocaleType[]
const BACKENDS = ['wsl', 'ssh', 'local'] as const satisfies readonly BackendType[]
const EDITORS = ['vscode', 'phpstorm', 'neovim', 'sublime', 'antigravity'] as const satisfies readonly PreferredEditorType[]

export const DEFAULT_APP_CONFIG: AppConfig = {
  viewMode: 'list',
  showLog: true,
  openLinksInBrowser: true,
  preferredEditor: 'vscode',
  theme: 'default',
  locale: 'en',
}

function normalizeOptionalBool(value: unknown): boolean | undefined {
  if (value === true || value === 'true') return true
  if (value === false || value === 'false') return false
  return undefined
}

function normalizeEnum<T extends string>(value: unknown, allowed: readonly T[], fallback: T): T {
  if (typeof value !== 'string') return fallback
  return allowed.includes(value as T) ? (value as T) : fallback
}

export function normalizeAppConfig(input: unknown): AppConfig {
  if (!input || typeof input !== 'object' || Array.isArray(input)) {
    return { ...DEFAULT_APP_CONFIG }
  }

  const raw = input as Record<string, unknown>

  return {
    ...DEFAULT_APP_CONFIG,
    ...raw,
    viewMode: normalizeEnum(raw.viewMode, VIEW_MODES, DEFAULT_APP_CONFIG.viewMode ?? 'list'),
    showLog: coerceToBool(raw.showLog, DEFAULT_APP_CONFIG.showLog ?? true),
    openLinksInBrowser: coerceToBool(raw.openLinksInBrowser, DEFAULT_APP_CONFIG.openLinksInBrowser ?? true),
    ddevTelemetryOptIn: normalizeOptionalBool(raw.ddevTelemetryOptIn),
    devMode: coerceToBool(raw.devMode, false),
    preferredEditor: normalizeEnum(raw.preferredEditor, EDITORS, DEFAULT_APP_CONFIG.preferredEditor ?? 'vscode'),
    theme: normalizeEnum(raw.theme, THEMES, DEFAULT_APP_CONFIG.theme ?? 'default'),
    locale: normalizeEnum(raw.locale, LOCALES, DEFAULT_APP_CONFIG.locale ?? 'en'),
    backend:
      typeof raw.backend === 'undefined'
        ? undefined
        : normalizeEnum(raw.backend, BACKENDS, BACKENDS[0]),
  }
}
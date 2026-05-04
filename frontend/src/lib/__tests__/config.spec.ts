import { describe, expect, it } from 'vitest'
import { DEFAULT_APP_CONFIG, normalizeAppConfig } from '../config'
import type { AppConfig } from '../types'

describe('config', () => {
  describe('DEFAULT_APP_CONFIG', () => {
    it('should have the expected default values', () => {
      expect(DEFAULT_APP_CONFIG).toEqual({
        viewMode: 'list',
        showLog: true,
        openLinksInBrowser: true,
        preferredEditor: 'vscode',
        theme: 'default',
        locale: 'en',
      })
    })
  })

  describe('normalizeAppConfig', () => {
    it('should return DEFAULT_APP_CONFIG for non-object inputs', () => {
      expect(normalizeAppConfig(null)).toEqual(DEFAULT_APP_CONFIG)
      expect(normalizeAppConfig(undefined)).toEqual(DEFAULT_APP_CONFIG)
      expect(normalizeAppConfig('string')).toEqual(DEFAULT_APP_CONFIG)
      expect(normalizeAppConfig(123)).toEqual(DEFAULT_APP_CONFIG)
      expect(normalizeAppConfig([])).toEqual(DEFAULT_APP_CONFIG)
    })

    it('should return DEFAULT_APP_CONFIG for an empty object', () => {
      const result = normalizeAppConfig({})
      expect(result).toEqual({
        ...DEFAULT_APP_CONFIG,
        ddevTelemetryOptIn: undefined,
        devMode: false,
        backend: undefined,
      })
    })

    it('should preserve valid configuration values', () => {
      const input: Partial<AppConfig> = {
        viewMode: 'grid',
        showLog: false,
        openLinksInBrowser: false,
        ddevTelemetryOptIn: true,
        preferredEditor: 'phpstorm',
        theme: 'acrylic',
        locale: 'zh',
        devMode: true,
        backend: 'ssh',
      }
      const result = normalizeAppConfig(input)
      expect(result).toEqual(expect.objectContaining(input))
    })

    it('should fall back to defaults for invalid enum values', () => {
      const input = {
        viewMode: 'invalid',
        preferredEditor: 'invalid',
        theme: 'invalid',
        locale: 'invalid',
        backend: 'invalid',
      }
      const result = normalizeAppConfig(input)
      expect(result.viewMode).toBe(DEFAULT_APP_CONFIG.viewMode)
      expect(result.preferredEditor).toBe(DEFAULT_APP_CONFIG.preferredEditor)
      expect(result.theme).toBe(DEFAULT_APP_CONFIG.theme)
      expect(result.locale).toBe(DEFAULT_APP_CONFIG.locale)
      expect(result.backend).toBe('wsl') // BACKENDS[0]
    })

    it('should correctly coerce boolean values', () => {
      const input = {
        showLog: 'false',
        openLinksInBrowser: 'true',
        devMode: 'true',
      }
      const result = normalizeAppConfig(input)
      expect(result.showLog).toBe(false)
      expect(result.openLinksInBrowser).toBe(true)
      expect(result.devMode).toBe(true)
    })

    it('should preserve telemetry opt-in only for valid boolean values', () => {
      expect(normalizeAppConfig({ ddevTelemetryOptIn: true }).ddevTelemetryOptIn).toBe(true)
      expect(normalizeAppConfig({ ddevTelemetryOptIn: 'false' }).ddevTelemetryOptIn).toBe(false)
      expect(normalizeAppConfig({ ddevTelemetryOptIn: null }).ddevTelemetryOptIn).toBeUndefined()
    })

    it('should handle backend field correctly', () => {
      // Undefined backend remains undefined
      expect(normalizeAppConfig({}).backend).toBeUndefined()

      // Valid backend is preserved
      expect(normalizeAppConfig({ backend: 'local' }).backend).toBe('local')

      // Invalid backend falls back to 'wsl' (BACKENDS[0])
      expect(normalizeAppConfig({ backend: 'invalid' }).backend).toBe('wsl')
    })

    it('should preserve extra fields from input', () => {
      const input = {
        viewMode: 'grid' as const,
        wslDistro: 'Ubuntu',
        customField: 'customValue',
      }
      const result = normalizeAppConfig(input) as any
      expect(result.viewMode).toBe('grid')
      expect(result.wslDistro).toBe('Ubuntu')
      expect(result.customField).toBe('customValue')
    })
  })
})

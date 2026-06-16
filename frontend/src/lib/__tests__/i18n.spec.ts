import { describe, expect, it, vi } from 'vitest'
import { createApp } from 'vue'
import {
  createI18nState,
  installI18n,
  useTranslation,
  loadLocale,
  en,
  LOCALE_LABELS,
} from '../i18n'

// Mock the dynamic imports for locales
vi.mock('../locales/zh.po?raw', () => ({
  default: 'msgid "general.loading"\nmsgstr "加载中..."',
}))

vi.mock('../locales/vi.po?raw', () => ({
  default: 'msgid "general.loading"\nmsgstr "Đang tải..."',
}))

describe('i18n', () => {
  describe('LOCALE_LABELS', () => {
    it('should define labels for supported locales', () => {
      expect(LOCALE_LABELS).toEqual({
        en: 'English',
        zh: '简体中文',
        vi: 'Tiếng Việt',
      })
    })
  })

  describe('createI18nState', () => {
    it('should initialize with default locale (en)', () => {
      const state = createI18nState()
      expect(state.locale.value).toBe('en')
      expect(state.messages.value).toEqual(en)
    })

    it('should initialize with a specific locale', () => {
      const state = createI18nState('zh')
      expect(state.locale.value).toBe('zh')
    })
  })

  describe('t function', () => {
    it('should translate a known key', () => {
      const state = createI18nState('en')
      expect(state.t('general.loading')).toBe('Loading…')
    })

    it('should fallback to English if key is missing in current locale', async () => {
      // Create a state with 'zh'
      const state = createI18nState('zh')
      // For this test, we might need to manually set messages to simulate a missing key
      state.messages.value = { 'only.in.zh': 'Only in ZH' }

      // 'general.cancel' is in en.po but not in our mocked zh messages
      expect(state.t('general.cancel')).toBe('Cancel')
    })

    it('should return the key itself if not found in current or English locale', () => {
      const state = createI18nState('en')
      expect(state.t('non.existent.key')).toBe('non.existent.key')
    })

    it('should replace variables in the translation', () => {
      const state = createI18nState('en')
      // detail.snapshots.deleteConfirm is "Delete snapshot \"{snap}\"?"
      expect(state.t('detail.snapshots.deleteConfirm', { snap: 'my-snap' })).toBe('Delete snapshot "my-snap"?')
    })

    it('should handle multiple occurrences of the same variable', () => {
      const state = createI18nState('en')
      // Manually add a message with multiple tokens
      state.messages.value = { 'test.multiple': '{val} - {val}' }
      expect(state.t('test.multiple', { val: 'foo' })).toBe('foo - foo')
    })
  })

  describe('setLocale', () => {
    it('should change the locale and load messages', async () => {
      const state = createI18nState('en')
      await state.setLocale('zh')
      expect(state.locale.value).toBe('zh')
      expect(state.t('general.loading')).toBe('加载中...')
    })

    it('should switch back to English', async () => {
      const state = createI18nState('zh')
      await state.setLocale('en')
      expect(state.locale.value).toBe('en')
      expect(state.messages.value).toEqual(en)
    })
  })

  describe('loadLocale', () => {
    it('should load and parse zh locale', async () => {
      const messages = await loadLocale('zh')
      expect(messages['general.loading']).toBe('加载中...')
    })

    it('should load and parse vi locale', async () => {
      const messages = await loadLocale('vi')
      expect(messages['general.loading']).toBe('Đang tải...')
    })

    it('should use cache for subsequent loads', async () => {
      const firstLoad = await loadLocale('zh')
      const secondLoad = await loadLocale('zh')
      expect(firstLoad).toBe(secondLoad)
    })
  })

  describe('Vue integration', () => {
    it('should provide and inject i18n state', () => {
      const app = createApp({
        setup() {
          const i18n = useTranslation()
          return { i18n }
        },
        template: '<div>{{ i18n.locale.value }}</div>',
      })

      const i18nState = createI18nState()
      installI18n(app, i18nState)

      // We don't necessarily need to mount it to test provide/inject if we use the app context
      // but let's test it via useTranslation in a component-like way
      app.runWithContext(() => {
        const injected = useTranslation()
        expect(injected).toBe(i18nState)
      })
    })

    it('should throw error if useTranslation is called without installI18n', () => {
      const app = createApp({
        setup() {
          useTranslation()
          return {}
        },
      })

      expect(() => {
        app.runWithContext(() => {
          useTranslation()
        })
      }).toThrow('useTranslation must be used after installI18n()')
    })
  })
})

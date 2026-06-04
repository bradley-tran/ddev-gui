import type { App, InjectionKey, Ref } from 'vue'
import { inject, ref } from 'vue'
import enRaw from './locales/en.po?raw'

function extractQuoted(line: string): string {
  const match = line.match(/^"(.*)"$/)
  return match?.[1]
    ? match[1]
        .replace(/\\n/g, '\n')
        .replace(/\\"/g, '"')
        .replace(/\\\\/g, '\\')
    : ''
}

function parsePO(raw: string): Record<string, string> {
  const result: Record<string, string> = {}
  let msgid = ''
  let msgstr = ''
  let section: 'id' | 'str' | '' = ''

  function flush() {
    if (msgid) result[msgid] = msgstr
    msgid = ''
    msgstr = ''
    section = ''
  }

  let start = 0
  const len = raw.length
  while (start <= len) {
    let end = raw.indexOf('\n', start)
    if (end === -1) end = len
    const line = raw.slice(start, end)
    start = end + 1

    const trimmed = line.trim()
    if (!trimmed || trimmed.startsWith('#')) {
      flush()
      continue
    }

    if (trimmed.startsWith('msgid ')) {
      flush()
      msgid = extractQuoted(trimmed.slice(6))
      section = 'id'
      continue
    }

    if (trimmed.startsWith('msgstr ')) {
      msgstr = extractQuoted(trimmed.slice(7))
      section = 'str'
      continue
    }

    if (trimmed.startsWith('"')) {
      const continuation = extractQuoted(trimmed)
      if (section === 'id') msgid += continuation
      if (section === 'str') msgstr += continuation
    }
  }

  flush()
  return result
}

export const en = parsePO(enRaw)
export type TranslationKey = keyof typeof en
export type Locale = 'en' | 'zh' | 'vi'

export const LOCALE_LABELS: Record<Locale, string> = {
  en: 'English',
  zh: '简体中文',
  vi: 'Tiếng Việt',
}

const localeCache: Partial<Record<Locale, Record<string, string>>> = { en }

export async function loadLocale(locale: Locale): Promise<Record<string, string>> {
  if (localeCache[locale]) return localeCache[locale] as Record<string, string>

  const raw =
    locale === 'zh'
      ? (await import('./locales/zh.po?raw')).default
      : (await import('./locales/vi.po?raw')).default

  localeCache[locale] = parsePO(raw)
  return localeCache[locale] as Record<string, string>
}

export interface I18nState {
  locale: Ref<Locale>
  messages: Ref<Record<string, string>>
  t: (key: string, vars?: Record<string, string>) => string
  setLocale: (locale: Locale) => Promise<void>
}

const I18N_KEY: InjectionKey<I18nState> = Symbol('i18n')

export function createI18nState(initialLocale: Locale = 'en'): I18nState {
  const locale = ref<Locale>(initialLocale)
  const messages = ref<Record<string, string>>(en)

  async function setLocale(nextLocale: Locale): Promise<void> {
    if (nextLocale === 'en') {
      locale.value = 'en'
      messages.value = en
      return
    }

    locale.value = nextLocale
    messages.value = await loadLocale(nextLocale)
  }

  function t(key: string, vars?: Record<string, string>): string {
    let value = messages.value[key] ?? en[key] ?? key
    if (vars) {
      // Use split/join instead of replaceAll to avoid TS target library issues
      for (const [token, replacement] of Object.entries(vars)) {
        value = value.split(`{${token}}`).join(replacement)
      }
    }
    return value
  }

  return {
    locale,
    messages,
    t,
    setLocale,
  }
}

export function installI18n(app: App, i18n = createI18nState()): I18nState {
  app.provide(I18N_KEY, i18n)
  return i18n
}

export function useTranslation(): I18nState {
  const i18n = inject(I18N_KEY)
  if (!i18n) {
    throw new Error('useTranslation must be used after installI18n()')
  }
  return i18n
}
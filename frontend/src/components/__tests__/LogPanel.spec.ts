import type { App as VueApp } from 'vue'
import { describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'
import { installI18n } from '@/lib/i18n'
import { useAppStore } from '@/stores/app'
import LogPanel from '../LogPanel.vue'

const i18nPlugin = {
  install(app: VueApp) {
    installI18n(app)
  },
}

describe('LogPanel', () => {
  const setup = () => {
    const pinia = createPinia()
    setActivePinia(pinia)
    const appStore = useAppStore()

    const wrapper = mount(LogPanel, {
      global: {
        plugins: [pinia, i18nPlugin],
      },
    })

    return { wrapper, appStore }
  }

  it('renders correctly with log entries', async () => {
    const { wrapper, appStore } = setup()

    appStore.addLog({
      id: '1',
      timestamp: '10:00:00',
      message: 'Hello World',
      level: 'info'
    })

    await flushPromises()

    expect(wrapper.find('.log-card').exists()).toBe(true)
    expect(wrapper.find('.log-entry').exists()).toBe(true)
    expect(wrapper.find('.log-time').text()).toBe('10:00:00')
    expect(wrapper.find('.log-msg').text()).toBe('Hello World')
    expect(wrapper.find('.log-entry').classes()).toContain('log-info')
  })

  it('is hidden when showLog is false', async () => {
    const { wrapper, appStore } = setup()

    appStore.config.showLog = false
    await flushPromises()

    expect(wrapper.find('.log-card').exists()).toBe(false)
  })

  it('is hidden when terminal is active', async () => {
    const { wrapper, appStore } = setup()

    appStore.terminalActive = true
    await flushPromises()

    expect(wrapper.find('.log-card').exists()).toBe(false)
  })

  it('clears log entries when clear button is clicked', async () => {
    const { wrapper, appStore } = setup()

    appStore.addLog({
      id: '1',
      timestamp: '10:00:00',
      message: 'Hello World',
      level: 'info'
    })
    await flushPromises()

    const clearBtn = wrapper.find('#logClearBtn')
    await clearBtn.trigger('click')

    expect(appStore.logEntries.length).toBe(0)
    await flushPromises()
    expect(wrapper.find('.log-entry').exists()).toBe(false)
  })

  it('renders ANSI codes as HTML for output level', async () => {
    const { wrapper, appStore } = setup()

    appStore.addLog({
      id: '1',
      timestamp: '10:00:00',
      message: '\x1b[31mRed Text\x1b[m',
      level: 'output'
    })

    await flushPromises()

    const msg = wrapper.find('.log-msg')
    expect(msg.html()).toContain('style="color:#e06c75"')
    expect(msg.text()).toBe('Red Text')
  })

  it('escapes HTML for non-output levels', async () => {
    const { wrapper, appStore } = setup()

    appStore.addLog({
      id: '1',
      timestamp: '10:00:00',
      message: '<b>Bold</b>',
      level: 'info'
    })

    await flushPromises()

    const msg = wrapper.find('.log-msg')
    expect(msg.html()).toContain('&lt;b&gt;Bold&lt;/b&gt;')
    expect(msg.text()).toBe('<b>Bold</b>')
  })

  it('scrolls to bottom when new log entry is added', async () => {
    const { wrapper, appStore } = setup()

    // The scrollIntoView is mocked in setup.ts
    const scrollSpy = window.HTMLElement.prototype.scrollIntoView

    appStore.addLog({
      id: '1',
      timestamp: '10:00:00',
      message: 'New Entry',
      level: 'info'
    })

    await flushPromises()

    expect(scrollSpy).toHaveBeenCalled()
  })
})

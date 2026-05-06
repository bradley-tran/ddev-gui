import type { App as VueApp } from 'vue'
import { describe, expect, it, type Mock } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import EmbeddedTerminal from '@/components/EmbeddedTerminal.vue'
import { installI18n } from '@/lib/i18n'

const i18nPlugin = {
  install(app: VueApp) {
    installI18n(app)
  },
}

describe('EmbeddedTerminal', () => {
  it('runs a command, renders streamed output, and clears with the built-in command', async () => {
    if (!window.go?.backend || !window.runtime) {
      throw new Error('Wails mocks are not available')
    }

    const ddevService = window.go.backend.DdevService as unknown as {
      ExecCommand: Mock
    }
    const runtime = window.runtime as unknown as {
      EventsOn: Mock
      EventsOff: Mock
    }

    const listeners = new Map<string, (...args: unknown[]) => void>()
    runtime.EventsOn.mockImplementation((event: string, callback: (...args: unknown[]) => void) => {
      listeners.set(event, callback)
    })

    const wrapper = mount(EmbeddedTerminal, {
      props: {
        projectName: 'demo',
      },
      global: {
        plugins: [i18nPlugin],
      },
      attachTo: document.body,
    })

    await flushPromises()

    await wrapper.get('[data-testid="embedded-terminal-input"]').setValue('php -v')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(ddevService.ExecCommand).toHaveBeenCalledWith('demo', 'php -v')
    expect(wrapper.text()).toContain('$ php -v')

    listeners.get('terminal:output:demo')?.('PHP 8.3.0')
    listeners.get('terminal:done:demo')?.(0)
    await flushPromises()

    expect(wrapper.text()).toContain('PHP 8.3.0')

    await wrapper.get('[data-testid="embedded-terminal-input"]').setValue('clear')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('ddev ssh - demo')
    expect(wrapper.text()).not.toContain('PHP 8.3.0')

    wrapper.unmount()
  })

  it('opens URLs from output and keeps output clicks from stealing input focus', async () => {
    if (!window.runtime) {
      throw new Error('Wails mocks are not available')
    }

    const runtime = window.runtime as unknown as {
      EventsOn: Mock
      EventsOff: Mock
      EventsEmit: Mock
    }

    const listeners = new Map<string, (...args: unknown[]) => void>()
    runtime.EventsOn.mockImplementation((event: string, callback: (...args: unknown[]) => void) => {
      listeners.set(event, callback)
    })

    const wrapper = mount(EmbeddedTerminal, {
      props: {
        projectName: 'demo',
      },
      global: {
        plugins: [i18nPlugin],
      },
      attachTo: document.body,
    })

    await flushPromises()

    listeners.get('terminal:output:demo')?.('Read https://example.com/docs.')
    await flushPromises()

    const input = wrapper.get('[data-testid="embedded-terminal-input"]').element as HTMLInputElement
    input.focus()
    input.blur()

    await wrapper.get('.terminal-line-output').trigger('click')
    expect(document.activeElement).not.toBe(input)

    const link = wrapper.get('.terminal-line-output a.terminal-link')
    expect(link.text()).toBe('https://example.com/docs')

    await link.trigger('click')
    expect(runtime.EventsEmit).toHaveBeenCalledWith('open:url', { url: 'https://example.com/docs' })

    wrapper.unmount()
  })
})
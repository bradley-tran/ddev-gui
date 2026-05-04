import type { App as VueApp } from 'vue'
import { describe, expect, it, type Mock } from 'vitest'
import { createPinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

import { installI18n } from '@/lib/i18n'
import router from '@/router'
import ProjectListView from '../ProjectListView.vue'

const i18nPlugin = {
  install(app: VueApp) {
    installI18n(app)
  },
}

describe('ProjectListView', () => {
  it('refreshes on menu events and F5', async () => {
    if (!window.go?.backend || !window.runtime) {
      throw new Error('Wails mocks are not available')
    }

    const ddevService = window.go.backend.DdevService as unknown as {
      ListJSON: Mock
    }
    const runtime = window.runtime as unknown as {
      EventsOn: Mock
      EventsOff: Mock
    }

    ddevService.ListJSON.mockResolvedValue(JSON.stringify([
      {
        name: 'demo',
        type: 'drupal10',
        status_desc: 'running',
        approot: '/workspace/demo',
      },
    ]))

    const pinia = createPinia()

    await router.push('/')
    await router.isReady()

    const wrapper = mount(ProjectListView, {
      global: {
        plugins: [pinia, router, i18nPlugin],
      },
    })

    await flushPromises()

    expect(ddevService.ListJSON).toHaveBeenCalledTimes(1)

    const refreshRegistration = runtime.EventsOn.mock.calls.find(([eventName]) => eventName === 'menu:refresh')
    expect(refreshRegistration).toBeDefined()
    if (!refreshRegistration) {
      throw new Error('menu:refresh listener was not registered')
    }

    const refreshCallback = refreshRegistration[1] as (() => void) | undefined
    expect(refreshCallback).toBeDefined()
    if (!refreshCallback) {
      throw new Error('menu:refresh callback was not registered')
    }

    refreshCallback()
    await flushPromises()

    expect(ddevService.ListJSON).toHaveBeenCalledTimes(2)

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'F5', bubbles: true, cancelable: true }))
    await flushPromises()

    expect(ddevService.ListJSON).toHaveBeenCalledTimes(3)
    expect(wrapper.text()).toContain('demo')
    wrapper.unmount()
    expect(runtime.EventsOff).toHaveBeenCalledWith('menu:refresh')
    expect(runtime.EventsOff).toHaveBeenCalledWith('menu:start')
    expect(runtime.EventsOff).toHaveBeenCalledWith('menu:stop')
  })
})
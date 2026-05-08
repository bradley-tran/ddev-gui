import type { App as VueApp } from 'vue'
import { describe, it, expect, type Mock } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

import App from '../App.vue'
import { installI18n } from '../lib/i18n'
import router from '../router'
import { useAppStore } from '../stores/app'

const i18nPlugin = {
  install(app: VueApp) {
    installI18n(app)
  },
}

const getConfigService = () => window.go!.backend!.ConfigService as unknown as {
  GetAll: Mock
}

const getDdevService = () => window.go!.backend!.DdevService as unknown as {
  WSLExists: Mock
  DdevInstalledVersion: Mock
}

describe('App', () => {
  it('mounts the ported Vue shell', async () => {
    getConfigService().GetAll.mockResolvedValueOnce(JSON.stringify({ ddevTelemetryOptIn: true }))

    await router.push('/')
    const pinia = createPinia()
    setActivePinia(pinia)

    const wrapper = mount(App, {
      global: {
        plugins: [pinia, router, i18nPlugin],
      },
    })

    await router.isReady()
    await flushPromises()

    expect(wrapper.find('.titlebar').exists()).toBe(true)
    expect(wrapper.find('.log-card').exists()).toBe(true)
    expect(wrapper.text()).toContain('Projects')
  })

  it('opens EnvInfoModal on startup when telemetry preference is unset', async () => {
    getConfigService().GetAll.mockResolvedValueOnce('{}')

    await router.push('/')
    const pinia = createPinia()
    setActivePinia(pinia)

    mount(App, {
      global: {
        plugins: [pinia, router, i18nPlugin],
      },
    })

    await router.isReady()
    await flushPromises()

    expect(useAppStore().modals.envInfo).toBe(true)
  })

  it('does not open EnvInfoModal on startup when telemetry preference is set', async () => {
    getConfigService().GetAll.mockResolvedValueOnce(JSON.stringify({ ddevTelemetryOptIn: false }))

    await router.push('/')
    const pinia = createPinia()
    setActivePinia(pinia)

    mount(App, {
      global: {
        plugins: [pinia, router, i18nPlugin],
      },
    })

    await router.isReady()
    await flushPromises()

    expect(useAppStore().modals.envInfo).toBe(false)
  })

  it('opens EnvInfoModal on startup when WSL is missing', async () => {
    getConfigService().GetAll.mockResolvedValueOnce(JSON.stringify({ ddevTelemetryOptIn: false }))
    getDdevService().WSLExists.mockResolvedValueOnce(false)

    await router.push('/')
    const pinia = createPinia()
    setActivePinia(pinia)

    mount(App, {
      global: {
        plugins: [pinia, router, i18nPlugin],
      },
    })

    await router.isReady()
    await flushPromises()

    expect(useAppStore().modals.envInfo).toBe(true)
  })

  it('opens EnvInfoModal on startup when DDEV is missing', async () => {
    getConfigService().GetAll.mockResolvedValueOnce(JSON.stringify({ ddevTelemetryOptIn: false }))
    getDdevService().WSLExists.mockResolvedValueOnce(true)
    getDdevService().DdevInstalledVersion.mockRejectedValueOnce(new Error('ddev not found'))

    await router.push('/')
    const pinia = createPinia()
    setActivePinia(pinia)

    mount(App, {
      global: {
        plugins: [pinia, router, i18nPlugin],
      },
    })

    await router.isReady()
    await flushPromises()

    expect(useAppStore().modals.envInfo).toBe(true)
  })

  it('opens EnvInfoModal on startup when DDEV missing returns exit status 127', async () => {
    getConfigService().GetAll.mockResolvedValueOnce(JSON.stringify({ ddevTelemetryOptIn: false }))
    getDdevService().WSLExists.mockResolvedValueOnce(true)
    getDdevService().DdevInstalledVersion.mockRejectedValueOnce(new Error('exit status 127'))

    await router.push('/')
    const pinia = createPinia()
    setActivePinia(pinia)

    mount(App, {
      global: {
        plugins: [pinia, router, i18nPlugin],
      },
    })

    await router.isReady()
    await flushPromises()

    expect(useAppStore().modals.envInfo).toBe(true)
  })
})

import type { App as VueApp } from 'vue'
import { describe, expect, it, vi, beforeEach, type Mock } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

import { installI18n } from '@/lib/i18n'
import EnvInfoModal from '../EnvInfoModal.vue'
import { useAppStore } from '@/stores/app'

const i18nPlugin = {
  install(app: VueApp) {
    installI18n(app)
  },
}

describe('EnvInfoModal.vue', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    getDdevService().WSLExists.mockResolvedValue(true)
  })

  const getDdevService = () => window.go!.backend!.DdevService as unknown as {
    DdevInstalledVersion: Mock
    InstallDdev: Mock
    WSLExists: Mock
  }

  const getConfigService = () => window.go!.backend!.ConfigService as unknown as {
    Set: Mock
  }

  it('renders loading state initially', async () => {
    // Delay the resolution of ddevInstalledVersion to see the loading state
    let resolveVersion: (value: string) => void
    const promise = new Promise<string>((resolve) => {
      resolveVersion = resolve
    })
    getDdevService().DdevInstalledVersion.mockReturnValue(promise)

    const wrapper = mount(EnvInfoModal, {
      global: {
        plugins: [i18nPlugin],
      },
    })

    expect(wrapper.text()).toContain('Loading…')

    // @ts-ignore
    resolveVersion!('v1.23.0')
    await flushPromises()
    expect(wrapper.text()).not.toContain('Loading…')
  })

  it('displays DDEV version when installed', async () => {
    getDdevService().DdevInstalledVersion.mockResolvedValue('v1.23.0')

    const wrapper = mount(EnvInfoModal, {
      global: {
        plugins: [i18nPlugin],
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('DDEV Version')
    expect(wrapper.text()).toContain('v1.23.0')
  })

  it('displays error and install button when DDEV is missing', async () => {
    getDdevService().DdevInstalledVersion.mockRejectedValue(new Error('ddev not found'))

    const wrapper = mount(EnvInfoModal, {
      global: {
        plugins: [i18nPlugin],
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('DDEV was not found')
    expect(wrapper.text()).toContain('ddev not found')
    expect(wrapper.find('button.flu-btn-accent').text()).toBe('Install DDEV')
  })

  it('surfaces WSL availability check errors in the modal', async () => {
    getDdevService().WSLExists.mockRejectedValue(new Error('wsl check failed'))

    const wrapper = mount(EnvInfoModal, {
      global: {
        plugins: [i18nPlugin],
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('wsl check failed')
    expect(wrapper.find('button.flu-btn-accent').text()).toBe('Open Settings')
    expect(getDdevService().DdevInstalledVersion).not.toHaveBeenCalled()
  })

  it('displays WSL error and open settings button when WSL error occurs', async () => {
    getDdevService().WSLExists.mockResolvedValue(true)
    getDdevService().DdevInstalledVersion.mockRejectedValue(new Error('wsl.exe not found'))

    const wrapper = mount(EnvInfoModal, {
      global: {
        plugins: [i18nPlugin],
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('wsl.exe not found')
    expect(wrapper.find('button.flu-btn-accent').text()).toBe('Open Settings')
  })

  it('displays the WSL install guide when WSL is missing', async () => {
    getDdevService().WSLExists.mockResolvedValue(false)

    const wrapper = mount(EnvInfoModal, {
      global: {
        plugins: [i18nPlugin],
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('WSL was not found.')
    expect(wrapper.text()).toContain('wsl --install')
    expect(wrapper.text()).toContain('wsl --update')
    expect(wrapper.text()).toContain('wsl --install Ubuntu --name DDEV')
    expect(wrapper.find('button.flu-btn-accent').exists()).toBe(false)
    expect(getDdevService().DdevInstalledVersion).not.toHaveBeenCalled()
  })

  it('toggles developer mode', async () => {
    getDdevService().DdevInstalledVersion.mockResolvedValue('v1.23.0')
    const appStore = useAppStore()
    appStore.config.devMode = false
    appStore.config.ddevTelemetryOptIn = true

    const wrapper = mount(EnvInfoModal, {
      global: {
        plugins: [i18nPlugin],
      },
    })

    await flushPromises()

    const checkbox = wrapper.find('input[name="dev-mode"]')
    expect((checkbox.element as HTMLInputElement).checked).toBe(false)

    await checkbox.setValue(true)
    expect(appStore.config.devMode).toBe(true)
    expect(getConfigService().Set).toHaveBeenCalledWith('devMode', true)
  })

  it('toggles DDEV telemetry opt-in', async () => {
    getDdevService().DdevInstalledVersion.mockResolvedValue('v1.23.0')
    const appStore = useAppStore()
    appStore.config.ddevTelemetryOptIn = true

    const wrapper = mount(EnvInfoModal, {
      global: {
        plugins: [i18nPlugin],
      },
    })

    await flushPromises()

    const checkbox = wrapper.find('input[name="ddev-telemetry-opt-in"]')
    expect((checkbox.element as HTMLInputElement).checked).toBe(true)

    await checkbox.setValue(false)
    expect(appStore.config.ddevTelemetryOptIn).toBe(false)
    expect(getConfigService().Set).toHaveBeenCalledWith('ddevTelemetryOptIn', false)
  })

  it('defaults telemetry opt-in to true and saves it when closed', async () => {
    getDdevService().DdevInstalledVersion.mockResolvedValue('v1.23.0')
    const appStore = useAppStore()
    appStore.config.ddevTelemetryOptIn = undefined

    const wrapper = mount(EnvInfoModal, {
      global: {
        plugins: [i18nPlugin],
      },
    })

    await flushPromises()

    const checkbox = wrapper.find('input[name="ddev-telemetry-opt-in"]')
    expect((checkbox.element as HTMLInputElement).checked).toBe(true)

    await wrapper.find('button.flu-btn-ghost').trigger('click')
    await flushPromises()

    expect(appStore.config.ddevTelemetryOptIn).toBe(true)
    expect(getConfigService().Set).toHaveBeenCalledWith('ddevTelemetryOptIn', true)
    expect(wrapper.emitted()).toHaveProperty('close')
  })

  it('handles DDEV installation and progress updates', async () => {
    getDdevService().DdevInstalledVersion.mockRejectedValue(new Error('ddev not found'))

    let resolveInstall: (value: string) => void
    const installPromise = new Promise<string>((resolve) => {
      resolveInstall = resolve
    })
    getDdevService().InstallDdev.mockReturnValue(installPromise)

    const wrapper = mount(EnvInfoModal, {
      global: {
        plugins: [i18nPlugin],
      },
    })

    await flushPromises()

    const installBtn = wrapper.find('button.flu-btn-accent')
    await installBtn.trigger('click')

    expect(getDdevService().InstallDdev).toHaveBeenCalled()
    expect(wrapper.text()).toContain('Connecting to GitHub…')

    // Simulate event from backend
    const runtime = window.runtime as any
    const installProgressHandler = runtime.EventsOn.mock.calls.find((call: any) => call[0] === 'ddev:output')?.[1]
    expect(installProgressHandler).toBeDefined()

    installProgressHandler('Downloading…')
    await flushPromises()
    expect(wrapper.text()).toContain('Downloading…')

    // @ts-ignore
    resolveInstall!('ok')
    await flushPromises()
    expect(wrapper.text()).not.toContain('Downloading…')
  })

  it('surfaces DDEV installation errors in the modal', async () => {
    getDdevService().DdevInstalledVersion.mockRejectedValue(new Error('ddev not found'))
    getDdevService().InstallDdev.mockRejectedValue(new Error('download failed'))

    const wrapper = mount(EnvInfoModal, {
      global: {
        plugins: [i18nPlugin],
      },
    })

    await flushPromises()

    const installBtn = wrapper.find('button.flu-btn-accent')
    await installBtn.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('download failed')
  })

  it('emits close event when close button is clicked', async () => {
    getDdevService().DdevInstalledVersion.mockResolvedValue('v1.23.0')
    const appStore = useAppStore()
    appStore.config.ddevTelemetryOptIn = true

    const wrapper = mount(EnvInfoModal, {
      global: {
        plugins: [i18nPlugin],
      },
    })

    await flushPromises()

    await wrapper.find('button.flu-btn-ghost').trigger('click')
    await flushPromises()
    expect(wrapper.emitted()).toHaveProperty('close')
  })

  it('emits openSettings event when Open Settings button is clicked', async () => {
    getDdevService().WSLExists.mockResolvedValue(true)
    getDdevService().DdevInstalledVersion.mockRejectedValue(new Error('wsl.exe not found'))

    const wrapper = mount(EnvInfoModal, {
      global: {
        plugins: [i18nPlugin],
      },
    })

    await flushPromises()

    await wrapper.find('button.flu-btn-accent').trigger('click')
    expect(wrapper.emitted()).toHaveProperty('openSettings')
  })
})

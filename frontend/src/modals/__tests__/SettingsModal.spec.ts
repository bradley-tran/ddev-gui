import { type App as VueApp } from 'vue'
import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { installI18n } from '@/lib/i18n'
import SettingsModal from '../SettingsModal.vue'
import Select from '@/components/Select.vue'
import { useAppStore } from '@/stores/app'

const i18nPlugin = {
  install(app: VueApp) {
    installI18n(app)
  },
}

describe('SettingsModal.vue', () => {
  let pinia: any
  let appStore: any

  beforeEach(() => {
    vi.clearAllMocks()
    pinia = createPinia()
    setActivePinia(pinia)
    appStore = useAppStore()

    // Default config
    appStore.setConfig({
      locale: 'en',
      openLinksInBrowser: true,
      preferredEditor: 'vscode',
      theme: 'default',
      backend: 'wsl',
      wslDistro: 'Ubuntu',
      devMode: false
    })

    const ddevService = window.go!.backend!.DdevService as any
    ddevService.ListWSLDistros.mockResolvedValue(['Ubuntu', 'Debian'])
    ddevService.ReloadBackend.mockResolvedValue(undefined)

    const configService = window.go!.backend!.ConfigService as any
    configService.Set.mockResolvedValue(undefined)

    document.body.classList.remove('platform-linux')
  })

  const mountSettingsModal = () => {
    return mount(SettingsModal, {
      global: {
        plugins: [pinia, i18nPlugin],
      },
    })
  }

  it('renders initial settings correctly', async () => {
    const wrapper = mountSettingsModal()
    await flushPromises()

    expect(wrapper.find('h2').text()).toBe('Settings')

    // Check if Select components are rendered with correct values
    // Using findComponent(Select) might be tricky if there are multiple.
    // We can check the v-model values or internal state if needed,
    // but checking the DOM is better for integration.

    const selects = wrapper.findAllComponents(Select)
    // Language, Preferred Editor, Theme, Backend, WSL Distro
    expect(selects.length).toBe(5)
  })

  it('hides theme selector on Linux platform', async () => {
    document.body.classList.add('platform-linux')
    const wrapper = mountSettingsModal()
    await flushPromises()

    const selects = wrapper.findAllComponents(Select)
    // Theme selector should be hidden
    expect(selects.length).toBe(4)
    expect(wrapper.text()).not.toContain('Theme')
  })

  it('shows SSH options only when devMode is enabled and backend is SSH', async () => {
    appStore.patchConfig({ devMode: true, backend: 'ssh', ssh: { host: 'testhost', port: '22', user: 'testuser', keyPath: '/path' } })
    const wrapper = mountSettingsModal()
    await flushPromises()

    expect(wrapper.find('#settSSHHost').exists()).toBe(true)
    expect(wrapper.find('#settSSHPort').exists()).toBe(true)
    expect(wrapper.find('#settSSHUser').exists()).toBe(true)
    expect(wrapper.find('#settSSHKey').exists()).toBe(true)
  })

  it('hides SSH options when devMode is disabled even if backend is SSH', async () => {
    appStore.patchConfig({ devMode: false, backend: 'ssh' })
    const wrapper = mountSettingsModal()
    await flushPromises()

    expect(wrapper.find('#settSSHHost').exists()).toBe(false)
  })

  it('calls ConfigService.set and reloads backend on save', async () => {
    const wrapper = mountSettingsModal()
    await flushPromises()

    const saveButton = wrapper.find('button.flu-btn-accent')
    await saveButton.trigger('click')
    await flushPromises()

    const configService = window.go!.backend!.ConfigService as any
    expect(configService.Set).toHaveBeenCalledWith('locale', 'en')
    expect(configService.Set).toHaveBeenCalledWith('openLinksInBrowser', true)

    const ddevService = window.go!.backend!.DdevService as any
    expect(ddevService.ReloadBackend).toHaveBeenCalled()
    expect(wrapper.emitted()).toHaveProperty('close')
  })

  it('handles save errors gracefully', async () => {
    const configService = window.go!.backend!.ConfigService as any
    configService.Set.mockRejectedValue(new Error('Save failed'))

    const spyLog = vi.spyOn(appStore, 'appLog')
    const spyToast = vi.spyOn(appStore, 'showToast')

    const wrapper = mountSettingsModal()
    await flushPromises()

    const saveButton = wrapper.find('button.flu-btn-accent')
    await saveButton.trigger('click')
    await flushPromises()

    expect(spyLog).toHaveBeenCalledWith(expect.stringContaining('Failed to save settings: Save failed'), 'error')
    expect(spyToast).toHaveBeenCalledWith('Failed to save settings', 'error')
    expect(wrapper.emitted()).not.toHaveProperty('close')
  })

  it('updates WSL distro options when backend is changed to wsl', async () => {
    appStore.patchConfig({ backend: 'local' })
    const ddevService = window.go!.backend!.DdevService as any
    ddevService.ListWSLDistros.mockResolvedValue(['Alpine', 'Fedora'])

    const wrapper = mountSettingsModal()
    await flushPromises()

    // Change backend to WSL
    const selects = wrapper.findAllComponents(Select)
    // Find the backend select and trigger change.
    // In SettingsModal, backend is the 3rd or 4th select depending on platform.
    // Let's find it by looking for the one with ALL_BACKEND_OPTIONS.
    const backendSelect = selects.find(s => s.props('options').some((o: any) => o.value === 'wsl'))
    expect(backendSelect).toBeDefined()

    await backendSelect!.setValue('wsl')
    await flushPromises()

    expect(ddevService.ListWSLDistros).toHaveBeenCalled()

    // Re-query selects because DOM might have changed with v-if
    const updatedSelects = wrapper.findAllComponents(Select)

    // Check if the distro select now contains the new distros
    const distroSelect = updatedSelects.find(s => s.props('options').some((o: any) => o.value === 'Alpine'))
    expect(distroSelect).toBeDefined()
    expect(distroSelect!.props('options')).toContainEqual({ value: 'Alpine', label: 'Alpine' })
    expect(distroSelect!.props('options')).toContainEqual({ value: 'Fedora', label: 'Fedora' })
  })

  it('saves SSH configuration when backend is SSH', async () => {
    appStore.patchConfig({ devMode: true, backend: 'ssh' })
    const wrapper = mountSettingsModal()
    await flushPromises()

    await wrapper.find('#settSSHHost').setValue('remote.host')
    await wrapper.find('#settSSHPort').setValue('2222')
    await wrapper.find('#settSSHUser').setValue('remoteuser')
    await wrapper.find('#settSSHKey').setValue('/home/user/.ssh/id_rsa')

    const saveButton = wrapper.find('button.flu-btn-accent')
    await saveButton.trigger('click')
    await flushPromises()

    const configService = window.go!.backend!.ConfigService as any
    expect(configService.Set).toHaveBeenCalledWith('ssh', {
      host: 'remote.host',
      port: '2222',
      user: 'remoteuser',
      keyPath: '/home/user/.ssh/id_rsa'
    })
  })
})

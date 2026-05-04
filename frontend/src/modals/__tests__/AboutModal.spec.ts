import { type App as VueApp } from 'vue'
import { describe, expect, it, vi, type Mock, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { installI18n } from '@/lib/i18n'
import AboutModal from '../AboutModal.vue'
import Spinner from '@/components/Spinner.vue'

const i18nPlugin = {
  install(app: VueApp) {
    installI18n(app)
  },
}

describe('AboutModal.vue', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    // Reset DdevInstalledVersion and AppVersion to default successful states
    const ddevService = window.go!.backend!.DdevService as any
    ddevService.DdevInstalledVersion.mockResolvedValue('v1.23.0')
    ddevService.AppVersion.mockResolvedValue({ version: '1.0.0', commitHash: 'abcdef1' })
  })

  const mountAboutModal = () => {
    const pinia = createPinia()
    return mount(AboutModal, {
      global: {
        plugins: [pinia, i18nPlugin],
      },
    })
  }

  it('shows loading spinners initially', async () => {
    const ddevService = window.go!.backend!.DdevService as unknown as {
      DdevInstalledVersion: Mock
      AppVersion: Mock
    }

    // Create a promise that doesn't resolve yet
    let resolveDdev: (v: string) => void
    ddevService.DdevInstalledVersion.mockReturnValue(new Promise((resolve) => { resolveDdev = resolve }))

    const wrapper = mountAboutModal()

    expect(wrapper.findComponent(Spinner).exists()).toBe(true)

    // Cleanup
    resolveDdev!('v1.23.0')
    await flushPromises()
  })

  it('renders app and DDEV versions correctly when loaded', async () => {
    const ddevService = window.go!.backend!.DdevService as unknown as {
      DdevInstalledVersion: Mock
      AppVersion: Mock
    }

    ddevService.DdevInstalledVersion.mockResolvedValue('v1.24.0')
    ddevService.AppVersion.mockResolvedValue({ version: '1.0.0', commitHash: 'abcdef1' })

    const wrapper = mountAboutModal()
    await flushPromises()

    expect(wrapper.text()).toContain('1.0.0 (abcdef1)')
    expect(wrapper.text()).toContain('v1.24.0')
    expect(wrapper.findComponent(Spinner).exists()).toBe(false)
  })

  it('handles "not installed" and "unknown" DDEV version states', async () => {
    const ddevService = window.go!.backend!.DdevService as unknown as {
      DdevInstalledVersion: Mock
    }

    // Case: empty version returned
    ddevService.DdevInstalledVersion.mockResolvedValue('')
    let wrapper = mountAboutModal()
    await flushPromises()
    expect(wrapper.text()).toContain('not installed')

    // Case: promise rejected
    ddevService.DdevInstalledVersion.mockRejectedValue(new Error('fail'))
    wrapper = mountAboutModal()
    await flushPromises()
    expect(wrapper.text()).toContain('unknown')
  })

  it('handles various app version states', async () => {
    const ddevService = window.go!.backend!.DdevService as unknown as {
      AppVersion: Mock
      DdevInstalledVersion: Mock
    }

    // Ensure DdevInstalledVersion is successful so it doesn't contain "unknown"
    ddevService.DdevInstalledVersion.mockResolvedValue('v1.23.0')

    // Case: no commit hash
    ddevService.AppVersion.mockResolvedValue({ version: '1.2.3', commitHash: '' })
    let wrapper = mountAboutModal()
    await flushPromises()
    expect(wrapper.text()).toContain('1.2.3')
    expect(wrapper.text()).not.toContain('()')

    // Case: "unknown" commit hash
    ddevService.AppVersion.mockResolvedValue({ version: '1.2.3', commitHash: 'unknown' })
    wrapper = mountAboutModal()
    await flushPromises()
    expect(wrapper.text()).toContain('1.2.3')
    // We want to make sure the "unknown" from commitHash is not shown.
    // DDEV version might be "unknown" if it failed, but we mocked it to succeed above.
    // However, looking at the previous failure, the word "unknown" was present because DdevInstalledVersion failed.
    expect(wrapper.text()).not.toMatch(/1\.2\.3.*unknown/)

    // Case: promise rejected
    ddevService.AppVersion.mockRejectedValue(new Error('fail'))
    wrapper = mountAboutModal()
    await flushPromises()
    expect(wrapper.text()).toContain('dev')
  })

  it('emits close event when close button is clicked', async () => {
    const wrapper = mountAboutModal()
    await flushPromises()

    const closeButton = wrapper.find('button.flu-btn-ghost')
    await closeButton.trigger('click')

    expect(wrapper.emitted()).toHaveProperty('close')
  })
})

import { type App as VueApp } from 'vue'
import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia, type Pinia } from 'pinia'
import { installI18n } from '@/lib/i18n'
import MasqueradeModal from '../MasqueradeModal.vue'
import Spinner from '@/components/Spinner.vue'
import { openUrl } from '@/lib/utils'
import { useAppStore } from '@/stores/app'

vi.mock('@/lib/utils', async () => {
  const actual = await vi.importActual('@/lib/utils') as any
  return {
    ...actual,
    openUrl: vi.fn(),
  }
})

const i18nPlugin = {
  install(app: VueApp) {
    installI18n(app)
  },
}

describe('MasqueradeModal.vue', () => {
  let ddevService: any
  let pinia: Pinia

  beforeEach(() => {
    vi.clearAllMocks()
    pinia = createPinia()
    setActivePinia(pinia)

    ddevService = window.go!.backend!.DdevService
    ddevService.DrushRecentUsers.mockResolvedValue('[]')
    ddevService.DrushUliAsUser.mockResolvedValue('/user/reset/1/abc/login')
  })

  const mountMasqueradeModal = (props = {}) => {
    return mount(MasqueradeModal, {
      props: {
        projectName: 'test-project',
        primaryUrl: 'https://test-project.ddev.site',
        ...props,
      },
      global: {
        plugins: [pinia, i18nPlugin],
        stubs: {
          Teleport: true,
        },
      },
    })
  }

  it('shows loading state initially', async () => {
    let resolveUsers: (v: string) => void
    const promise = new Promise<string>((resolve) => { resolveUsers = resolve })
    ddevService.DrushRecentUsers.mockReturnValue(promise)

    const wrapper = mountMasqueradeModal()

    // We need to wait a tick for onMounted to call loadUsers and it to reach the first await
    await new Promise(resolve => setTimeout(resolve, 0))

    expect(wrapper.find('.loading-state').exists()).toBe(true)
    expect(wrapper.findComponent(Spinner).exists()).toBe(true)

    resolveUsers!('[]')
    await flushPromises()
    expect(wrapper.find('.loading-state').exists()).toBe(false)
  })

  it('loads and displays users on mount', async () => {
    const mockUsers = [
      { uid: '1', name: 'admin', mail: 'admin@example.com' },
      { uid: '2', name: 'user', mail: 'user@example.com' },
    ]
    ddevService.DrushRecentUsers.mockResolvedValue(JSON.stringify(mockUsers))

    const wrapper = mountMasqueradeModal()
    await flushPromises()

    expect(ddevService.DrushRecentUsers).toHaveBeenCalledWith('test-project')

    const rows = wrapper.findAll('tr.masq-row')
    expect(rows).toHaveLength(2)
    expect(rows[0]!.text()).toContain('admin')
    expect(rows[0]!.text()).toContain('admin@example.com')
    expect(rows[1]!.text()).toContain('user')
    expect(rows[1]!.text()).toContain('user@example.com')
  })

  it('shows "no users found" message when list is empty', async () => {
    ddevService.DrushRecentUsers.mockResolvedValue('[]')
    const wrapper = mountMasqueradeModal()
    await flushPromises()

    expect(wrapper.text()).toContain('No users found')
  })

  it('handles error when loading users', async () => {
    ddevService.DrushRecentUsers.mockRejectedValue(new Error('Drush failed'))
    const wrapper = mountMasqueradeModal()
    await flushPromises()

    expect(wrapper.text()).toContain('No users found')
  })

  it('performs masquerade when a user row is clicked', async () => {
    const mockUsers = [{ uid: '42', name: 'target', mail: 'target@example.com' }]
    ddevService.DrushRecentUsers.mockResolvedValue(JSON.stringify(mockUsers))
    ddevService.DrushUliAsUser.mockResolvedValue('/user/reset/42/token')

    const wrapper = mountMasqueradeModal()
    await flushPromises()

    await wrapper.find('tr.masq-row').trigger('click')

    expect(ddevService.DrushUliAsUser).toHaveBeenCalledWith('test-project', '42')
    expect(openUrl).toHaveBeenCalledWith('https://test-project.ddev.site/user/reset/42/token', expect.any(Boolean))
    expect(wrapper.emitted()).toHaveProperty('close')
  })

  it('performs masquerade when the Login button in a row is clicked', async () => {
    const mockUsers = [{ uid: '42', name: 'target', mail: 'target@example.com' }]
    ddevService.DrushRecentUsers.mockResolvedValue(JSON.stringify(mockUsers))
    ddevService.DrushUliAsUser.mockResolvedValue('/user/reset/42/token')

    const wrapper = mountMasqueradeModal()
    await flushPromises()

    // Find the login button in the row
    await wrapper.find('tr.masq-row button').trigger('click')

    expect(ddevService.DrushUliAsUser).toHaveBeenCalledWith('test-project', '42')
    expect(wrapper.emitted()).toHaveProperty('close')
  })

  it('performs masquerade via manual UID entry', async () => {
    const wrapper = mountMasqueradeModal()
    await flushPromises()

    const input = wrapper.find('input.flu-input')
    await input.setValue('123')
    await wrapper.find('.masq-input-row button').trigger('click')

    expect(ddevService.DrushUliAsUser).toHaveBeenCalledWith('test-project', '123')
    expect(wrapper.emitted()).toHaveProperty('close')
  })

  it('handles absolute ULI URLs', async () => {
    ddevService.DrushUliAsUser.mockResolvedValue('https://external.site/login')

    const wrapper = mountMasqueradeModal()
    await flushPromises()

    await wrapper.find('input.flu-input').setValue('1')
    await wrapper.find('.masq-input-row button').trigger('click')

    expect(openUrl).toHaveBeenCalledWith('https://external.site/login', expect.any(Boolean))
  })

  it('handles relative ULI URLs that do not start with slash', async () => {
    ddevService.DrushUliAsUser.mockResolvedValue('user/reset/1/abc')

    const wrapper = mountMasqueradeModal()
    await flushPromises()

    await wrapper.find('input.flu-input').setValue('1')
    await wrapper.find('.masq-input-row button').trigger('click')

    expect(openUrl).toHaveBeenCalledWith('https://test-project.ddev.site/user/reset/1/abc', expect.any(Boolean))
  })

  it('shows error toast and logs when masquerade fails', async () => {
    const appStore = useAppStore()
    const logSpy = vi.spyOn(appStore, 'appLog')
    const toastSpy = vi.spyOn(appStore, 'showToast')

    ddevService.DrushUliAsUser.mockRejectedValue(new Error('Uli failed'))

    const wrapper = mountMasqueradeModal()
    await flushPromises()

    await wrapper.find('input.flu-input').setValue('1')
    await wrapper.find('.masq-input-row button').trigger('click')

    expect(logSpy).toHaveBeenCalledWith(expect.stringContaining('Masquerade failed: Uli failed'), 'error')
    expect(toastSpy).toHaveBeenCalledWith('Masquerade failed', 'error')
    expect(wrapper.emitted()).not.toHaveProperty('close')
  })

  it('disables interactions while masquerading is in progress', async () => {
    let resolveUli: (v: string) => void
    ddevService.DrushUliAsUser.mockImplementation(() => {
      return new Promise((resolve) => { resolveUli = resolve })
    })

    const wrapper = mountMasqueradeModal()
    await flushPromises()

    await wrapper.find('input.flu-input').setValue('1')
    await wrapper.find('.masq-input-row button').trigger('click')

    expect(wrapper.find('.masq-input-row button').attributes('disabled')).toBeDefined()
    expect(wrapper.find('input.flu-input').attributes('disabled')).toBeDefined()
    expect(wrapper.findComponent(Spinner).exists()).toBe(true)

    resolveUli!('/ok')
    await flushPromises()

    expect(wrapper.emitted()).toHaveProperty('close')
  })
})

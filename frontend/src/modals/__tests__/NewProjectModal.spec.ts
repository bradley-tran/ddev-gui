import type { App as VueApp } from 'vue'
import { describe, expect, it, vi, beforeEach, type Mock } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { type VueWrapper, flushPromises, mount } from '@vue/test-utils'

import { installI18n } from '@/lib/i18n'
import NewProjectModal from '../NewProjectModal.vue'
import { useAppStore } from '@/stores/app'

const i18nPlugin = {
  install(app: VueApp) {
    installI18n(app)
  },
}

describe('NewProjectModal.vue', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  const getDdevService = () => window.go!.backend!.DdevService as unknown as {
    CloneRepo: Mock
    ConfigureProject: Mock
    Start: Mock
    ListJSON: Mock
  }

  const getConfigService = () => window.go!.backend!.ConfigService as unknown as {
    SetProjectConfig: Mock
  }

  const mountModal = () => {
    return mount(NewProjectModal, {
      global: {
        plugins: [i18nPlugin],
        stubs: {
          Teleport: true,
        },
      },
    })
  }

  it('renders with initial default values', async () => {
    const wrapper = mountModal()

    expect(wrapper.find('h2').text()).toBe('New Project')
    expect((wrapper.find('#projName').element as HTMLInputElement).value).toBe('')
    expect((wrapper.find('#projGitRepo').element as HTMLInputElement).value).toBe('')

    const typeSelect = wrapper.findComponent('#projType') as VueWrapper<any>
    expect(typeSelect.props('modelValue')).toBe('drupal11')

    const docrootInput = wrapper.find('#projDocroot')
    expect((docrootInput.element as HTMLInputElement).value).toBe('web')

    const phpSelect = wrapper.findComponent('#projPhpVersion') as VueWrapper<any>
    expect(phpSelect.props('modelValue')).toBe('8.3')
  })

  it('shows error if project name is missing on submit', async () => {
    const wrapper = mountModal()
    const form = wrapper.find('form')

    await form.trigger('submit')

    expect(wrapper.find('.form-error').text()).toBe('Project name is required.')
  })

  it('updates docroot automatically when type changes', async () => {
    const wrapper = mountModal()
    const typeSelect = wrapper.findComponent('#projType') as VueWrapper<any>

    await typeSelect.vm.$emit('update:modelValue', 'laravel')
    expect((wrapper.find('#projDocroot').element as HTMLInputElement).value).toBe('public')

    await typeSelect.vm.$emit('update:modelValue', 'php')
    expect((wrapper.find('#projDocroot').element as HTMLInputElement).value).toBe('')
  })

  it('stops auto-updating docroot if manually edited', async () => {
    const wrapper = mountModal()
    const docrootInput = wrapper.find('#projDocroot')
    const typeSelect = wrapper.findComponent('#projType') as VueWrapper<any>

    await docrootInput.setValue('custom-folder')
    await typeSelect.vm.$emit('update:modelValue', 'laravel')

    expect((docrootInput.element as HTMLInputElement).value).toBe('custom-folder')
  })

  it('handles successful project creation without git repo', async () => {
    const ddevService = getDdevService()
    const configService = getConfigService()
    const appStore = useAppStore()
    const logSpy = vi.spyOn(appStore, 'appLog')
    const toastSpy = vi.spyOn(appStore, 'showToast')
    const patchConfigSpy = vi.spyOn(appStore, 'patchConfig')
    const setProjectsJSONSpy = vi.spyOn(appStore, 'setProjectsJSON')

    ddevService.ListJSON.mockResolvedValue('[{"name": "new-site"}]')

    const wrapper = mountModal()
    await wrapper.find('#projName').setValue('new-site')

    const form = wrapper.find('form')
    await form.trigger('submit')

    expect(ddevService.CloneRepo).not.toHaveBeenCalled()
    expect(ddevService.ConfigureProject).toHaveBeenCalledWith('~', 'new-site', 'drupal11', 'web', '8.3')
    expect(ddevService.Start).toHaveBeenCalledWith('new-site')
    expect(configService.SetProjectConfig).toHaveBeenCalledWith('new-site', 'initialized', false)

    await flushPromises()

    expect(logSpy).toHaveBeenCalledWith('Project "new-site" created and started.', 'success')
    expect(toastSpy).toHaveBeenCalledWith('Project "new-site" created', 'success')
    expect(patchConfigSpy).toHaveBeenCalledWith({
      projects: {
        'new-site': { initialized: false }
      }
    })
    expect(setProjectsJSONSpy).toHaveBeenCalledWith('[{"name": "new-site"}]')
    expect(wrapper.emitted()).toHaveProperty('close')
  })

  it('handles successful project creation with git repo', async () => {
    const ddevService = getDdevService()
    const configService = getConfigService()

    const wrapper = mountModal()
    await wrapper.find('#projName').setValue('cloned-site')
    await wrapper.find('#projGitRepo').setValue('https://github.com/example/repo.git')

    const form = wrapper.find('form')
    await form.trigger('submit')

    expect(ddevService.CloneRepo).toHaveBeenCalledWith('cloned-site', 'https://github.com/example/repo.git')
    expect(ddevService.ConfigureProject).toHaveBeenCalledWith('~', 'cloned-site', 'drupal11', 'web', '8.3')
    expect(configService.SetProjectConfig).toHaveBeenCalledWith('cloned-site', 'initialized', true)

    await flushPromises()
    expect(wrapper.emitted()).toHaveProperty('close')
  })

  it('handles backend errors during creation', async () => {
    const ddevService = getDdevService()
    ddevService.ConfigureProject.mockRejectedValue(new Error('Config failed'))

    const appStore = useAppStore()
    const logSpy = vi.spyOn(appStore, 'appLog')

    const wrapper = mountModal()
    await wrapper.find('#projName').setValue('fail-site')

    const form = wrapper.find('form')
    await form.trigger('submit')
    await flushPromises()

    expect(wrapper.find('.form-error').text()).toBe('Config failed')
    expect(logSpy).toHaveBeenCalledWith('Failed to create project: Config failed', 'error')

    // Ensure button is re-enabled
    expect(wrapper.find('button.flu-btn-accent').attributes('disabled')).toBeUndefined()
  })

  it('shows loading state during submission', async () => {
    const ddevService = getDdevService()
    // Mock a slow operation
    let resolveStart: (v: string) => void
    const startPromise = new Promise<string>(resolve => { resolveStart = resolve })
    ddevService.ConfigureProject.mockResolvedValue('')
    ddevService.Start.mockReturnValue(startPromise)

    const wrapper = mountModal()
    await wrapper.find('#projName').setValue('slow-site')

    const form = wrapper.find('form')
    await form.trigger('submit')

    // Allow the microtask for ConfigureProject to complete so it gets to Start
    await flushPromises()

    const submitBtn = wrapper.find('button.flu-btn-accent')
    const cancelBtn = wrapper.find('button.flu-btn-ghost')

    expect(submitBtn.attributes('disabled')).toBeDefined()
    expect(submitBtn.text()).toBe('Creating…')
    expect(cancelBtn.attributes('disabled')).toBeDefined()

    // @ts-ignore
    resolveStart!('')
    await flushPromises()

    expect(wrapper.emitted()).toHaveProperty('close')
  })

  it('emits close when cancel button is clicked', async () => {
    const wrapper = mountModal()
    await wrapper.find('button.flu-btn-ghost').trigger('click')
    expect(wrapper.emitted()).toHaveProperty('close')
  })
})

import type { App as VueApp } from 'vue'
import { describe, expect, it, vi, beforeEach, type Mock } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { type VueWrapper, flushPromises, mount } from '@vue/test-utils'

import { installI18n } from '@/lib/i18n'
import ModifyProjectModal from '../ModifyProjectModal.vue'
import Select from '@/components/Select.vue'
import Spinner from '@/components/Spinner.vue'
import { useAppStore } from '@/stores/app'
import type { DdevProject } from '@/lib/types'

const i18nPlugin = {
  install(app: VueApp) {
    installI18n(app)
  },
}

describe('ModifyProjectModal.vue', () => {
  const mockProject: DdevProject = {
    name: 'test-project',
    php_version: '8.2',
    nodejs_version: '18',
    type: 'laravel',
    docroot: 'public',
  }

  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  const getDdevService = () => window.go!.backend!.DdevService as unknown as {
    ModifyProject: Mock
  }

  const mountModal = (project: DdevProject | null = mockProject) => {
    return mount(ModifyProjectModal, {
      props: {
        projectName: 'test-project',
        project,
      },
      global: {
        plugins: [i18nPlugin],
        stubs: {
          Teleport: true,
        },
      },
    })
  }

  it('renders with initial project values', async () => {
    const wrapper = mountModal()

    const phpSelect = wrapper.findComponent('#modifyPhpVersion') as VueWrapper<any>
    expect(phpSelect.props('modelValue')).toBe('8.2')

    const nodeSelect = wrapper.findComponent('#modifyNodejsVersion') as VueWrapper<any>
    expect(nodeSelect.props('modelValue')).toBe('18')

    const typeSelect = wrapper.findComponent('#modifyProjectType') as VueWrapper<any>
    expect(typeSelect.props('modelValue')).toBe('laravel')

    const docrootInput = wrapper.find('#modifyDocroot')
    expect((docrootInput.element as HTMLInputElement).value).toBe('public')
  })

  it('updates form when project prop changes', async () => {
    const wrapper = mountModal()

    const newProject: DdevProject = {
      name: 'test-project',
      php_version: '8.3',
      nodejs_version: '20',
      type: 'wordpress',
      docroot: '',
    }

    await wrapper.setProps({ project: newProject })

    expect((wrapper.findComponent('#modifyPhpVersion') as VueWrapper<any>).props('modelValue')).toBe('8.3')
    expect((wrapper.findComponent('#modifyNodejsVersion') as VueWrapper<any>).props('modelValue')).toBe('20')
    expect((wrapper.findComponent('#modifyProjectType') as VueWrapper<any>).props('modelValue')).toBe('wordpress')
    expect((wrapper.find('#modifyDocroot').element as HTMLInputElement).value).toBe('')
  })

  it('handles successful submission', async () => {
    const ddevService = getDdevService()
    ddevService.ModifyProject.mockResolvedValue('ok')
    const appStore = useAppStore()
    const logSpy = vi.spyOn(appStore, 'appLog')
    const toastSpy = vi.spyOn(appStore, 'showToast')

    const wrapper = mountModal()

    // Change some values
    await (wrapper.findComponent('#modifyPhpVersion') as VueWrapper<any>).vm.$emit('update:modelValue', '8.4')
    await wrapper.find('#modifyDocroot').setValue('web')

    const applyBtn = wrapper.find('button.flu-btn-accent')
    await applyBtn.trigger('click')

    expect(ddevService.ModifyProject).toHaveBeenCalledWith(
      'test-project',
      '8.4',
      '18',
      'laravel',
      'web',
    )

    await flushPromises()

    expect(logSpy).toHaveBeenCalledWith('Project test-project modified', 'success')
    expect(toastSpy).toHaveBeenCalledWith('Project test-project modified', 'success')
    expect(wrapper.emitted()).toHaveProperty('modified')
    expect(wrapper.emitted()).toHaveProperty('close')
  })

  it('handles submission failure', async () => {
    const ddevService = getDdevService()
    ddevService.ModifyProject.mockRejectedValue(new Error('Modification failed'))
    const appStore = useAppStore()
    const logSpy = vi.spyOn(appStore, 'appLog')
    const toastSpy = vi.spyOn(appStore, 'showToast')

    const wrapper = mountModal()

    const applyBtn = wrapper.find('button.flu-btn-accent')
    await applyBtn.trigger('click')

    await flushPromises()

    expect(logSpy).toHaveBeenCalledWith('Modify failed: Modification failed', 'error')
    expect(toastSpy).toHaveBeenCalledWith('Modify failed', 'error')
    expect(wrapper.emitted()).not.toHaveProperty('modified')
    expect(wrapper.emitted()).not.toHaveProperty('close')

    // Check if it's not running anymore
    expect(applyBtn.attributes('disabled')).toBeUndefined()
  })

  it('emits close event when cancel button is clicked', async () => {
    const wrapper = mountModal()
    await wrapper.find('button.flu-btn-ghost').trigger('click')
    expect(wrapper.emitted()).toHaveProperty('close')
  })

  it('disables inputs and shows spinner while running', async () => {
    const ddevService = getDdevService()
    let resolveModify: (v: string) => void
    const promise = new Promise<string>((resolve) => { resolveModify = resolve })
    ddevService.ModifyProject.mockReturnValue(promise)

    const wrapper = mountModal()

    await wrapper.find('button.flu-btn-accent').trigger('click')

    expect((wrapper.findComponent('#modifyPhpVersion') as VueWrapper<any>).props('disabled')).toBe(true)
    expect((wrapper.findComponent('#modifyNodejsVersion') as VueWrapper<any>).props('disabled')).toBe(true)
    expect((wrapper.findComponent('#modifyProjectType') as VueWrapper<any>).props('disabled')).toBe(true)
    expect(wrapper.find('#modifyDocroot').attributes('disabled')).toBeDefined()
    expect(wrapper.find('button.flu-btn-ghost').attributes('disabled')).toBeDefined()

    const applyBtn = wrapper.find('button.flu-btn-accent')
    expect(applyBtn.attributes('disabled')).toBeDefined()
    expect(applyBtn.text()).toContain('Saving…')
    expect(wrapper.findComponent(Spinner).exists()).toBe(true)

    // @ts-ignore
    resolveModify!('ok')
    await flushPromises()

    expect(applyBtn.text()).not.toContain('Saving…')
  })

  it('uses default values when project fields are missing', async () => {
    const wrapper = mountModal({} as DdevProject)

    expect((wrapper.findComponent('#modifyPhpVersion') as VueWrapper<any>).props('modelValue')).toBe('8.3')
    expect((wrapper.findComponent('#modifyNodejsVersion') as VueWrapper<any>).props('modelValue')).toBe('20')
    expect((wrapper.findComponent('#modifyProjectType') as VueWrapper<any>).props('modelValue')).toBe('php')
    expect((wrapper.find('#modifyDocroot').element as HTMLInputElement).value).toBe('')
  })
})

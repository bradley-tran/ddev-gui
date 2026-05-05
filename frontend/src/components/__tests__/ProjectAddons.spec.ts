import type { App as VueApp } from 'vue'
import { describe, expect, it, type Mock } from 'vitest'
import { createPinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

import { installI18n } from '@/lib/i18n'
import ProjectAddons from '../ProjectAddons.vue'

const i18nPlugin = {
  install(app: VueApp) {
    installI18n(app)
  },
}

describe('ProjectAddons', () => {
  it('loads installed add-ons and filters the picker list', async () => {
    if (!window.go?.backend) {
      throw new Error('Wails backend mock is not available')
    }

    const ddevService = window.go.backend.DdevService as unknown as {
      AddonsJSON: Mock
      AddonsAvailableJSON: Mock
    }

    ddevService.AddonsJSON.mockResolvedValue(
      JSON.stringify({
        installed: [
          {
            Name: 'Redis',
            Version: '1.2.3',
            Repository: 'ddev/ddev-redis',
            InstalledDate: 'today',
          },
        ],
      }),
    )
    ddevService.AddonsAvailableJSON.mockResolvedValue(
      JSON.stringify([
        { Repository: 'ddev/ddev-redis', Description: 'Redis integration' },
        { Repository: 'ddev/ddev-solr', Description: 'Solr search' },
      ]),
    )

    const wrapper = mount(ProjectAddons, {
      props: {
        projectName: 'demo',
      },
      global: {
        plugins: [createPinia(), i18nPlugin],
      },
    })

    await flushPromises()

    expect(ddevService.AddonsJSON).toHaveBeenCalledWith('demo')
    expect(wrapper.text()).toContain('Redis')
    expect(wrapper.text()).toContain('1.2.3')

    await wrapper.get('[data-testid="project-addons-open-picker"]').trigger('click')
    await flushPromises()

    expect(ddevService.AddonsAvailableJSON).toHaveBeenCalledWith('demo')
    expect(wrapper.text()).toContain('ddev/ddev-redis')
    expect(wrapper.text()).toContain('ddev/ddev-solr')

    await wrapper.get('[data-testid="project-addons-search"]').setValue('redis')
    await flushPromises()

    expect(wrapper.text()).toContain('ddev/ddev-redis')
    expect(wrapper.text()).not.toContain('ddev/ddev-solr')
  })

  it('confirms before removing an installed add-on', async () => {
    if (!window.go?.backend) {
      throw new Error('Wails backend mock is not available')
    }

    const ddevService = window.go.backend.DdevService as unknown as {
      AddonsJSON: Mock
      AddonRemove: Mock
    }

    ddevService.AddonsJSON.mockResolvedValue(
      JSON.stringify({
        installed: [
          {
            Name: 'Redis',
            Repository: 'ddev/ddev-redis',
          },
        ],
      }),
    )

    const wrapper = mount(ProjectAddons, {
      props: {
        projectName: 'demo',
      },
      global: {
        plugins: [createPinia(), i18nPlugin],
      },
    })

    await flushPromises()

    await wrapper.get('.flu-btn-danger').trigger('click')
    await flushPromises()

    expect(ddevService.AddonRemove).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Remove add-on "ddev/ddev-redis"?')

    await wrapper.get('.confirm-delete-modal-confirm').trigger('click')
    await flushPromises()

    expect(ddevService.AddonRemove).toHaveBeenCalledWith('demo', 'ddev/ddev-redis')
  })
})
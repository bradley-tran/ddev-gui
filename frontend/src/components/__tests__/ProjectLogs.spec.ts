import type { App as VueApp } from 'vue'
import { describe, expect, it, type Mock } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import { installI18n } from '@/lib/i18n'
import ProjectLogs from '../ProjectLogs.vue'

const i18nPlugin = {
  install(app: VueApp) {
    installI18n(app)
  },
}

describe('ProjectLogs', () => {
  it('loads the default web logs, switches services, and refreshes the selected service', async () => {
    if (!window.go?.backend) {
      throw new Error('Wails backend mock is not available')
    }

    const ddevService = window.go.backend.DdevService as unknown as {
      ProjectLogs: Mock
    }

    ddevService.ProjectLogs.mockResolvedValueOnce('web | [notice] ready')
    ddevService.ProjectLogs.mockResolvedValueOnce('db  | [ok] healthy')
    ddevService.ProjectLogs.mockResolvedValueOnce('db  | [info] refreshed')

    const wrapper = mount(ProjectLogs, {
      props: {
        projectName: 'demo',
        serviceNames: ['web', 'db'],
      },
      global: {
        plugins: [i18nPlugin],
      },
    })

    await flushPromises()

    expect(ddevService.ProjectLogs).toHaveBeenCalledWith('demo', 'web')
    expect(wrapper.text()).toContain('web | [notice] ready')

    await wrapper.get('[data-testid="project-logs-service-toggle"]').trigger('click')
    await flushPromises()

    await wrapper.get('[data-testid="project-logs-service-option-db"]').trigger('click')
    await flushPromises()

    expect(ddevService.ProjectLogs).toHaveBeenNthCalledWith(2, 'demo', 'db')
    expect(wrapper.text()).toContain('db  | [ok] healthy')

    await wrapper.get('[data-testid="project-logs-refresh"]').trigger('click')
    await flushPromises()

    expect(ddevService.ProjectLogs).toHaveBeenCalledTimes(3)
    expect(ddevService.ProjectLogs).toHaveBeenNthCalledWith(3, 'demo', 'db')
    expect(wrapper.text()).toContain('db  | [info] refreshed')
  })
})

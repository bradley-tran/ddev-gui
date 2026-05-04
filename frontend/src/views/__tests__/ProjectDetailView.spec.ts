import type { App as VueApp } from 'vue'
import { describe, expect, it, type Mock } from 'vitest'
import { createPinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

import { installI18n } from '@/lib/i18n'
import router from '@/router'
import { useAppStore } from '@/stores/app'
import ProjectDetailView from '../ProjectDetailView.vue'

const i18nPlugin = {
  install(app: VueApp) {
    installI18n(app)
  },
}

describe('ProjectDetailView', () => {
  it('loads project detail data and renders snapshots', async () => {
    if (!window.go?.backend) {
      throw new Error('Wails backend mock is not available')
    }

    const ddevService = window.go.backend.DdevService as unknown as {
      DescribeJSON: Mock
      DrushRecentUsers: Mock
      SnapshotCreate: Mock
      SnapshotListJSON: Mock
      ProjectLogs: Mock
    }
    const pinia = createPinia()
    const appStore = useAppStore(pinia)

    appStore.setProjectsJSON(
      JSON.stringify([
        {
          name: 'demo',
          type: 'drupal10',
          status_desc: 'running',
          docroot: 'web',
          approot: '/workspace/demo',
          httpsurl: 'https://demo.ddev.site',
        },
      ]),
    )
    appStore.isProjectsLoaded = true

    ddevService.DescribeJSON.mockResolvedValue(
      JSON.stringify({
        raw: {
          name: 'demo',
          type: 'drupal10',
          status_desc: 'running',
          docroot: 'web',
          approot: '/workspace/demo',
          router: 'https',
          php_version: '8.3',
          nodejs_version: '20',
          services: {
            web: {
              status: 'running',
              https_url: 'https://demo.ddev.site',
            },
          },
        },
      }),
    )
    ddevService.SnapshotListJSON.mockResolvedValue(JSON.stringify([{ name: 'pre-deploy' }]))
    ddevService.ProjectLogs.mockResolvedValue('web | [notice] ready')

    await router.push('/projects/demo')
    await router.isReady()

    const wrapper = mount(ProjectDetailView, {
      global: {
        plugins: [pinia, router, i18nPlugin],
      },
    })

    await flushPromises()

    expect(ddevService.DescribeJSON).toHaveBeenCalledWith('demo')
    expect(ddevService.SnapshotListJSON).toHaveBeenCalledWith('demo')
    expect(wrapper.text()).toContain('drupal10')
    expect(wrapper.text()).toContain('8.3')
    expect(wrapper.findAll('.toolbar-dropdown')).toHaveLength(2)
    expect(appStore.terminalActive).toBe(false)

    const toolbarToggles = wrapper.findAll('.toolbar-dropdown-toggle')
    expect(toolbarToggles).toHaveLength(2)

    const drupalToggle = toolbarToggles[0]
    const moreToggle = toolbarToggles[1]
    expect(drupalToggle).toBeDefined()
    expect(moreToggle).toBeDefined()

    if (!drupalToggle || !moreToggle) {
      throw new Error('Toolbar dropdown toggles were not rendered')
    }

    await drupalToggle.trigger('click')
    await flushPromises()
    expect(wrapper.findAll('.toolbar-dropdown-item')).toHaveLength(4)

    const drupalMenuItems = wrapper.findAll('.toolbar-dropdown-item')
    const masqueradeItem = drupalMenuItems[2]
    expect(masqueradeItem).toBeDefined()
    if (!masqueradeItem) {
      throw new Error('Masquerade dropdown item was not rendered')
    }

    await masqueradeItem.trigger('click')
    await flushPromises()
    expect(ddevService.DrushRecentUsers).toHaveBeenCalledWith('demo')
    expect(wrapper.text()).toContain('Masquerade')

    const closeButton = wrapper.find('.flu-modal-close')
    expect(closeButton.exists()).toBe(true)
    await closeButton.trigger('click')
    await flushPromises()

    await moreToggle.trigger('click')
    await flushPromises()
    expect(wrapper.findAll('.toolbar-dropdown-item')).toHaveLength(3)
    expect(wrapper.text()).not.toContain('Actions')

    const modifyButton = wrapper.find('.detail-section-title .flu-btn')
    expect(modifyButton.exists()).toBe(true)
    await modifyButton.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Modify Project')
    expect(wrapper.find('#modifyPhpVersion').exists()).toBe(true)

    const closeModifyButton = wrapper.find('.flu-modal-close')
    expect(closeModifyButton.exists()).toBe(true)
    await closeModifyButton.trigger('click')
    await flushPromises()

    const configButton = wrapper
      .findAll('.detail-section-title .flu-btn')
      .find((button) => button.text().includes('Config'))
    expect(configButton).toBeDefined()
    if (!configButton) {
      throw new Error('Service config button was not rendered')
    }

    await configButton.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Configure Services')
    expect(wrapper.find('#svcWebPort').exists()).toBe(true)

    const findSidebarTab = (title: string) =>
      wrapper.findAll('.detail-sidebar-btn').find((tab) => tab.attributes('title') === title)

    const snapshotTab = findSidebarTab('Snapshots')
    expect(snapshotTab).toBeDefined()
    if (!snapshotTab) {
      throw new Error('Snapshot tab was not rendered')
    }

    const filesTab = findSidebarTab('Files')
    expect(filesTab).toBeDefined()
    if (!filesTab) {
      throw new Error('Files tab was not rendered')
    }

    await filesTab.trigger('click')
    await flushPromises()
    expect(appStore.terminalActive).toBe(true)

    await snapshotTab.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('pre-deploy')
    expect(appStore.terminalActive).toBe(false)

    const createSnapshotButton = wrapper.find('.snapshot-create-btn')
    expect(createSnapshotButton.exists()).toBe(true)
    await createSnapshotButton.trigger('click')
    await flushPromises()

    const createSnapshotName = wrapper.find('#createSnapshotName')
    expect(createSnapshotName.exists()).toBe(true)
    await createSnapshotName.setValue('pre-deploy-2')

    const createSnapshotSubmit = wrapper.find('.snapshot-create-modal-submit')
    expect(createSnapshotSubmit.exists()).toBe(true)
    await createSnapshotSubmit.trigger('click')
    await flushPromises()

    expect(ddevService.SnapshotCreate).toHaveBeenCalledWith('demo', 'pre-deploy-2')
    expect(wrapper.find('#createSnapshotName').exists()).toBe(false)

    const logsTab = findSidebarTab('Logs')
    expect(logsTab).toBeDefined()
    if (!logsTab) {
      throw new Error('Logs tab was not rendered')
    }

    await logsTab.trigger('click')
    await flushPromises()

    expect(ddevService.ProjectLogs).toHaveBeenCalledWith('demo', 'web')
    expect(wrapper.text()).toContain('web | [notice] ready')
    expect(appStore.terminalActive).toBe(true)
  })
})
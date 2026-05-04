import type { App as VueApp } from 'vue'
import { describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'
import { installI18n } from '@/lib/i18n'
import { useAppStore } from '@/stores/app'
import { Runtime, DdevService } from '@/lib/wails'
import Titlebar from '../Titlebar.vue'

const pushMock = vi.fn()
vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: pushMock,
  }),
}))

const i18nPlugin = {
  install(app: VueApp) {
    installI18n(app)
  },
}

describe('Titlebar', () => {
  const setup = () => {
    const pinia = createPinia()
    setActivePinia(pinia)
    const appStore = useAppStore()
    appStore.navigateToList = vi.fn()
    appStore.refreshProjects = vi.fn().mockResolvedValue(undefined)
    appStore.appLog = vi.fn()
    appStore.showToast = vi.fn()
    appStore.saveConfigValue = vi.fn().mockResolvedValue(undefined)
    appStore.openModal = vi.fn()

    const wrapper = mount(Titlebar, {
      global: {
        plugins: [pinia, i18nPlugin],
      },
    })

    return { wrapper, appStore }
  }

  it('renders correctly', () => {
    const { wrapper } = setup()
    expect(wrapper.find('.titlebar-brand').exists()).toBe(true)
    expect(wrapper.find('.titlebar-center').exists()).toBe(true)
  })

  it('navigates home when clicking brand', async () => {
    const { wrapper, appStore } = setup()

    await wrapper.find('.titlebar-brand').trigger('click')

    expect(appStore.navigateToList).toHaveBeenCalled()
    expect(pushMock).toHaveBeenCalledWith({ name: 'project-list' })
  })

  it('displays home title in center when no project is selected', () => {
    const { wrapper } = setup()
    // en.po has "Home" for "general.home"
    expect(wrapper.find('.titlebar-center').text()).toBe('Home')
  })

  it('displays project name and type in center when project is selected', async () => {
    const { wrapper, appStore } = setup()

    appStore.selectedProject = 'my-project'
    appStore.currentView = 'detail'
    appStore.setProjectsJSON(JSON.stringify([
      { name: 'my-project', type: 'drupal10' }
    ]))

    await flushPromises()

    expect(wrapper.find('.titlebar-center').text()).toBe('my-project | drupal10')
  })

  it('displays only project name in center when type is unknown', async () => {
    const { wrapper, appStore } = setup()

    appStore.selectedProject = 'other-project'
    appStore.currentView = 'detail'
    appStore.setProjectsJSON(JSON.stringify([
      { name: 'other-project' }
    ]))

    await flushPromises()

    expect(wrapper.find('.titlebar-center').text()).toBe('other-project')
  })

  it('opens and closes menus', async () => {
    const { wrapper } = setup()

    const projectsMenu = wrapper.findAll('.menubar-item')[0]!
    await projectsMenu.trigger('mouseenter')
    expect(wrapper.find('.menubar-dropdown').exists()).toBe(true)
    // en.po has "New Project" for "menu.newProject"
    expect(wrapper.find('.menubar-dropdown').text()).toContain('New Project')

    await wrapper.find('.menubar').trigger('mouseleave')
    expect(wrapper.find('.menubar-dropdown').exists()).toBe(false)
  })

  it('handles "Stop All" functionality', async () => {
    const { wrapper, appStore } = setup()
    const powerOffSpy = vi.spyOn(DdevService, 'powerOff').mockResolvedValue('OK')

    // Open projects menu
    await wrapper.findAll('.menubar-item')[0]!.trigger('mouseenter')

    // Find by translated text "Stop All"
    const stopAllBtn = wrapper.findAll('.menubar-dropdown-item').find((b) => b.text().includes('Stop All'))!
    await stopAllBtn.trigger('click')

    expect(appStore.appLog).toHaveBeenCalledWith('Stopping all projects...', 'info')
    expect(powerOffSpy).toHaveBeenCalled()
    await flushPromises()
    expect(appStore.appLog).toHaveBeenCalledWith('All projects stopped.', 'success')
    expect(appStore.showToast).toHaveBeenCalledWith('All projects stopped', 'success')
    expect(appStore.refreshProjects).toHaveBeenCalled()
  })

  it('handles "Refresh" functionality', async () => {
    const { wrapper, appStore } = setup()

    await wrapper.findAll('.menubar-item')[0]!.trigger('mouseenter')
    const refreshBtn = wrapper.findAll('.menubar-dropdown-item').find((b) => b.text().includes('Refresh'))!
    await refreshBtn.trigger('click')

    expect(appStore.refreshProjects).toHaveBeenCalled()
  })

  it('toggles settings via buttons', async () => {
    const { wrapper, appStore } = setup()

    const browserToggle = wrapper.find('#browserToggleBtn')
    // config.openLinksInBrowser defaults to true
    await browserToggle.trigger('click')
    expect(appStore.saveConfigValue).toHaveBeenCalledWith('openLinksInBrowser', false)

    const logToggle = wrapper.find('#logToggleBtn')
    // config.showLog defaults to true
    await logToggle.trigger('click')
    expect(appStore.saveConfigValue).toHaveBeenCalledWith('showLog', false)
  })

  it('handles window controls', async () => {
    const { wrapper } = setup()

    const minimiseSpy = vi.spyOn(Runtime, 'minimise')
    const toggleMaximiseSpy = vi.spyOn(Runtime, 'toggleMaximise')
    const quitSpy = vi.spyOn(Runtime, 'quit')

    await wrapper.find('#winMinimize').trigger('click')
    expect(minimiseSpy).toHaveBeenCalled()

    await wrapper.find('#winMaximize').trigger('click')
    expect(toggleMaximiseSpy).toHaveBeenCalled()

    await wrapper.find('#winClose').trigger('click')
    expect(quitSpy).toHaveBeenCalled()
  })

  it('toggles maximise on double click', async () => {
    const { wrapper } = setup()
    const toggleMaximiseSpy = vi.spyOn(Runtime, 'toggleMaximise')

    await wrapper.find('.titlebar').trigger('dblclick')
    expect(toggleMaximiseSpy).toHaveBeenCalled()
  })

  it('manages resize event listener', () => {
    const addSpy = vi.spyOn(window, 'addEventListener')
    const removeSpy = vi.spyOn(window, 'removeEventListener')

    const { wrapper } = setup()
    expect(addSpy).toHaveBeenCalledWith('resize', expect.any(Function))

    wrapper.unmount()
    expect(removeSpy).toHaveBeenCalledWith('resize', expect.any(Function))
  })
})

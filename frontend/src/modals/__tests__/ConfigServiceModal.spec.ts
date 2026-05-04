import type { App as VueApp } from 'vue'
import { beforeEach, describe, expect, it, vi, type Mock } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

import { installI18n } from '@/lib/i18n'
import ConfigServiceModal from '../ConfigServiceModal.vue'
import { useAppStore } from '@/stores/app'
import type { DdevProject } from '@/lib/types'

const i18nPlugin = {
  install(app: VueApp) {
    installI18n(app)
  },
}

describe('ConfigServiceModal.vue', () => {
  const getDdevService = () => window.go!.backend!.DdevService as unknown as {
    ConfigureServices: Mock
  }

  const mountModal = (project: DdevProject | null = null) => {
    return mount(ConfigServiceModal, {
      props: {
        projectName: 'demo',
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

  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('starts with empty ports and loads toggle values from project data', () => {
    const wrapper = mountModal({
      router_http_port: '8080',
      dbinfo: { published_port: 3307 },
      xdebug_enabled: true,
      xhprof_mode: 'prepend',
      xhgui_status: 'disabled',
    })

    expect((wrapper.find('#svcWebPort').element as HTMLInputElement).value).toBe('')
    expect((wrapper.find('#svcDbPort').element as HTMLInputElement).value).toBe('')
    expect((wrapper.find('#svcXdebug').element as HTMLInputElement).checked).toBe(true)
    expect((wrapper.find('#svcXhprof').element as HTMLInputElement).checked).toBe(true)
    expect((wrapper.find('#svcXhgui').element as HTMLInputElement).checked).toBe(false)
    expect(wrapper.find('#svcWebPort').attributes('placeholder')).toBe('(unchanged)')
    expect(wrapper.find('#svcDbPort').attributes('placeholder')).toBe('(unchanged)')
  })

  it('uses xhgui_status over stale xhprof_mode when loading xhgui toggle', () => {
    const wrapper = mountModal({
      xhprof_mode: 'xhgui',
      xhgui_status: 'disabled',
    })

    expect((wrapper.find('#svcXhgui').element as HTMLInputElement).checked).toBe(false)
    expect((wrapper.find('#svcXhprof').element as HTMLInputElement).checked).toBe(false)
  })

  it('keeps xhgui and xhprof states consistent', async () => {
    const wrapper = mountModal()

    await wrapper.find('#svcXhgui').setValue(true)
    expect((wrapper.find('#svcXhprof').element as HTMLInputElement).checked).toBe(true)

    await wrapper.find('#svcXhprof').setValue(false)
    expect((wrapper.find('#svcXhgui').element as HTMLInputElement).checked).toBe(false)
  })

  it('saves service config without persisting app project config', async () => {
    const ddevService = getDdevService()
    const configService = window.go!.backend!.ConfigService as unknown as {
      SetProjectConfig: Mock
    }
    ddevService.ConfigureServices.mockResolvedValue('ok')

    const appStore = useAppStore()
    const logSpy = vi.spyOn(appStore, 'appLog')
    const toastSpy = vi.spyOn(appStore, 'showToast')

    const wrapper = mountModal()

    await wrapper.find('#svcWebPort').setValue('8080')
    await wrapper.find('#svcDbPort').setValue('3307')
    await wrapper.find('#svcXdebug').setValue(true)
    await wrapper.find('#svcXhprof').setValue(true)
    await wrapper.find('#svcXhgui').setValue(true)

    await wrapper.find('.service-config-modal-submit').trigger('click')
    await flushPromises()

    expect(ddevService.ConfigureServices).toHaveBeenCalledWith('demo', '8080', '3307', true, true, true)
    expect(configService.SetProjectConfig).not.toHaveBeenCalled()
    expect(logSpy).toHaveBeenCalledWith('Service configuration saved for demo', 'success')
    expect(toastSpy).toHaveBeenCalledWith('Service configuration saved', 'success')
    expect(wrapper.emitted()).toHaveProperty('configured')
    expect(wrapper.emitted()).toHaveProperty('close')
  })

  it('rejects invalid ports before calling backend', async () => {
    const ddevService = getDdevService()
    const appStore = useAppStore()
    const logSpy = vi.spyOn(appStore, 'appLog')

    const wrapper = mountModal()

    await wrapper.find('#svcWebPort').setValue('abc')
    await wrapper.find('.service-config-modal-submit').trigger('click')
    await flushPromises()

    expect(ddevService.ConfigureServices).not.toHaveBeenCalled()
    expect(logSpy).toHaveBeenCalledWith('Ports must be numbers between 1 and 65535', 'error')
    expect(wrapper.emitted()).not.toHaveProperty('configured')
  })
})

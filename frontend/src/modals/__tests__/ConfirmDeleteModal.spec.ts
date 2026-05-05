import type { App as VueApp } from 'vue'
import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import { installI18n } from '@/lib/i18n'
import ConfirmDeleteModal from '../ConfirmDeleteModal.vue'

const i18nPlugin = {
  install(app: VueApp) {
    installI18n(app)
  },
}

describe('ConfirmDeleteModal', () => {
  it('renders the message and emits confirm', async () => {
    const wrapper = mount(ConfirmDeleteModal, {
      props: {
        title: 'Delete',
        message: 'Delete project "demo"?',
      },
      global: {
        plugins: [i18nPlugin],
      },
    })

    expect(wrapper.text()).toContain('Delete project "demo"?')

    await wrapper.get('.confirm-delete-modal-confirm').trigger('click')

    expect(wrapper.emitted('confirm')).toHaveLength(1)
  })

  it('prevents close and confirm actions while pending', async () => {
    const wrapper = mount(ConfirmDeleteModal, {
      props: {
        title: 'Delete',
        message: 'Delete project "demo"?',
        pending: true,
      },
      global: {
        plugins: [i18nPlugin],
      },
    })

    await wrapper.find('.flu-modal-close').trigger('click')
    await wrapper.get('.confirm-delete-modal-confirm').trigger('click')

    expect(wrapper.emitted('close')).toBeUndefined()
    expect(wrapper.emitted('confirm')).toBeUndefined()
  })
})
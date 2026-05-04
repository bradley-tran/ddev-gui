import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import Modal from '../Modal.vue'

describe('Modal.vue', () => {
  const title = 'Test Modal'
  const defaultSlotContent = '<p>Default slot content</p>'
  const footerSlotContent = '<button>Footer button</button>'

  it('renders the title prop', () => {
    const wrapper = mount(Modal, {
      props: { title },
    })
    expect(wrapper.find('h2').text()).toBe(title)
  })

  it('renders default slot content', () => {
    const wrapper = mount(Modal, {
      props: { title },
      slots: {
        default: defaultSlotContent,
      },
    })
    expect(wrapper.find('.flu-modal-body').html()).toContain(defaultSlotContent)
  })

  it('renders footer slot content when provided', () => {
    const wrapper = mount(Modal, {
      props: { title },
      slots: {
        footer: footerSlotContent,
      },
    })
    expect(wrapper.find('.flu-modal-footer').exists()).toBe(true)
    expect(wrapper.find('.flu-modal-footer').html()).toContain(footerSlotContent)
  })

  it('does not render footer container when footer slot is not provided', () => {
    const wrapper = mount(Modal, {
      props: { title },
    })
    expect(wrapper.find('.flu-modal-footer').exists()).toBe(false)
  })

  it('applies flu-modal-wide class when wide prop is true', () => {
    const wrapper = mount(Modal, {
      props: { title, wide: true },
    })
    expect(wrapper.find('.flu-modal').classes()).toContain('flu-modal-wide')
  })

  it('does not apply flu-modal-wide class when wide prop is false', () => {
    const wrapper = mount(Modal, {
      props: { title, wide: false },
    })
    expect(wrapper.find('.flu-modal').classes()).not.toContain('flu-modal-wide')
  })

  it('emits close event when close button is clicked', async () => {
    const wrapper = mount(Modal, {
      props: { title },
    })
    await wrapper.find('.flu-modal-close').trigger('click')
    expect(wrapper.emitted()).toHaveProperty('close')
  })

  it('emits close event when overlay is clicked', async () => {
    const wrapper = mount(Modal, {
      props: { title },
    })
    // The overlay is the root element in the template
    await wrapper.trigger('click')
    expect(wrapper.emitted()).toHaveProperty('close')
  })

  it('does not emit close event when modal content is clicked', async () => {
    const wrapper = mount(Modal, {
      props: { title },
    })
    await wrapper.find('.flu-modal').trigger('click')
    expect(wrapper.emitted('close')).toBeUndefined()
  })

  it('emits close event when Escape key is pressed', async () => {
    const wrapper = mount(Modal, {
      props: { title },
    })
    const event = new KeyboardEvent('keydown', { key: 'Escape' })
    document.dispatchEvent(event)
    expect(wrapper.emitted()).toHaveProperty('close')
  })

  it('removes keydown listener on unmount', () => {
    const addSpy = vi.spyOn(document, 'addEventListener')
    const removeSpy = vi.spyOn(document, 'removeEventListener')

    const wrapper = mount(Modal, {
      props: { title },
    })

    expect(addSpy).toHaveBeenCalledWith('keydown', expect.any(Function))

    const handler = addSpy.mock.calls.find(call => call[0] === 'keydown')?.[1]

    wrapper.unmount()

    expect(removeSpy).toHaveBeenCalledWith('keydown', handler)

    addSpy.mockRestore()
    removeSpy.mockRestore()
  })
})

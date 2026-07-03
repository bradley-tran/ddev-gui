import { describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { mount } from '@vue/test-utils'
import { useAppStore } from '@/stores/app'
import ToastContainer from '../ToastContainer.vue'
import ToastItem from '../ToastItem.vue'

describe('ToastContainer.vue', () => {
  const setup = () => {
    const pinia = createPinia()
    setActivePinia(pinia)
    const appStore = useAppStore()

    // We mock removeToast to check if it's called
    appStore.removeToast = vi.fn()

    const wrapper = mount(ToastContainer, {
      global: {
        plugins: [pinia],
      },
    })

    return { wrapper, appStore }
  }

  it('renders an empty container when there are no toasts', () => {
    const { wrapper } = setup()
    expect(wrapper.find('#toastContainer').exists()).toBe(true)
    expect(wrapper.findAllComponents(ToastItem).length).toBe(0)
  })

  it('renders ToastItem components based on appStore.toasts', async () => {
    const { wrapper, appStore } = setup()

    appStore.toasts = [
      { id: '1', message: 'First toast', type: 'info', duration: 3000 },
      { id: '2', message: 'Second toast', type: 'error', duration: 4000 }
    ]

    // Wait for Vue to update the DOM
    await wrapper.vm.$nextTick()

    const toastItems = wrapper.findAllComponents(ToastItem)
    expect(toastItems.length).toBe(2)

    // Check props for the first toast
    expect(toastItems[0]!.props('id')).toBe('1')
    expect(toastItems[0]!.props('message')).toBe('First toast')
    expect(toastItems[0]!.props('type')).toBe('info')
    expect(toastItems[0]!.props('duration')).toBe(3000)

    // Check props for the second toast
    expect(toastItems[1]!.props('id')).toBe('2')
    expect(toastItems[1]!.props('message')).toBe('Second toast')
    expect(toastItems[1]!.props('type')).toBe('error')
    expect(toastItems[1]!.props('duration')).toBe(4000)
  })

  it('calls appStore.removeToast when ToastItem emits dismiss', async () => {
    const { wrapper, appStore } = setup()

    appStore.toasts = [
      { id: '1', message: 'Dismiss me', type: 'info', duration: 3000 }
    ]

    await wrapper.vm.$nextTick()

    const toastItem = wrapper.findComponent(ToastItem)
    await toastItem.vm.$emit('dismiss', '1')

    expect(appStore.removeToast).toHaveBeenCalledWith('1')
  })
})

import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import ImageViewer from '../ImageViewer.vue'

describe('ImageViewer', () => {
  const defaultProps = {
    data: 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==',
    fileName: 'test.png',
  }

  it('renders the component with basic props', () => {
    const wrapper = mount(ImageViewer, {
      props: defaultProps,
    })

    expect(wrapper.find('[data-testid="image-viewer"]').exists()).toBe(true)
    const img = wrapper.find('img')
    expect(img.exists()).toBe(true)
    expect(img.attributes('src')).toBe(`data:image/png;base64,${defaultProps.data}`)
    expect(img.attributes('alt')).toBe(defaultProps.fileName)
  })

  it('computes src correctly for different extensions', () => {
    const testCases = [
      { fileName: 'image.jpg', expectedMime: 'image/jpeg' },
      { fileName: 'image.jpeg', expectedMime: 'image/jpeg' },
      { fileName: 'image.svg', expectedMime: 'image/svg+xml' },
      { fileName: 'image.webp', expectedMime: 'image/webp' },
      { fileName: 'image.gif', expectedMime: 'image/gif' },
    ]

    testCases.forEach(({ fileName, expectedMime }) => {
      const wrapper = mount(ImageViewer, {
        props: { ...defaultProps, fileName },
      })
      expect(wrapper.find('img').attributes('src')).toBe(`data:${expectedMime};base64,${defaultProps.data}`)
    })
  })

  it('applies custom className', () => {
    const wrapper = mount(ImageViewer, {
      props: { ...defaultProps, className: 'custom-class' },
    })

    expect(wrapper.find('.img-viewer').classes()).toContain('custom-class')
  })

  it('handles zooming behavior correctly', async () => {
    const wrapper = mount(ImageViewer, {
      props: { ...defaultProps, zoom: 100 },
    })

    const img = wrapper.find('img')
    expect(wrapper.classes()).not.toContain('img-viewer-zoomed')
    expect(img.element.style.transform).toBe('scale(1)')

    await wrapper.setProps({ zoom: 200 })
    expect(wrapper.classes()).toContain('img-viewer-zoomed')
    expect(img.element.style.transform).toBe('scale(2) translate(0px, 0px)')

    await wrapper.setProps({ zoom: 50 })
    expect(wrapper.classes()).not.toContain('img-viewer-zoomed')
    expect(img.element.style.transform).toBe('scale(0.5)')
  })

  it('resets offset when zoom is reset to 100 or less', async () => {
    const wrapper = mount(ImageViewer, {
      props: { ...defaultProps, zoom: 200 },
    })

    // Simulate dragging to change offset
    await wrapper.trigger('mousedown', { clientX: 100, clientY: 100 })
    await wrapper.trigger('mousemove', { clientX: 150, clientY: 150 })
    await wrapper.trigger('mouseup')

    expect(wrapper.find('img').element.style.transform).toBe('scale(2) translate(25px, 25px)')

    await wrapper.setProps({ zoom: 100 })
    expect(wrapper.find('img').element.style.transform).toBe('scale(1)')

    // Zoom back in and verify offset is 0
    await wrapper.setProps({ zoom: 200 })
    expect(wrapper.find('img').element.style.transform).toBe('scale(2) translate(0px, 0px)')
  })

  it('handles dragging (panning) correctly when zoomed', async () => {
    const wrapper = mount(ImageViewer, {
      props: { ...defaultProps, zoom: 200 },
    })

    const img = wrapper.find('img')

    // Initial state
    expect(img.element.style.transform).toBe('scale(2) translate(0px, 0px)')

    // Start dragging
    await wrapper.trigger('mousedown', { clientX: 100, clientY: 100 })
    await wrapper.trigger('mousemove', { clientX: 120, clientY: 130 })

    // dx = 20, dy = 30
    // transform: scale(2) translate(20/2, 30/2) => translate(10px, 15px)
    expect(img.element.style.transform).toBe('scale(2) translate(10px, 15px)')

    // Continue dragging
    await wrapper.trigger('mousemove', { clientX: 150, clientY: 180 })
    // dx = 50, dy = 80
    // transform: scale(2) translate(50/2, 80/2) => translate(25px, 40px)
    expect(img.element.style.transform).toBe('scale(2) translate(25px, 40px)')

    // Stop dragging
    await wrapper.trigger('mouseup')
    await wrapper.trigger('mousemove', { clientX: 200, clientY: 200 })
    // Should not change after mouseup
    expect(img.element.style.transform).toBe('scale(2) translate(25px, 40px)')
  })

  it('stops dragging on mouseleave', async () => {
    const wrapper = mount(ImageViewer, {
      props: { ...defaultProps, zoom: 200 },
    })

    await wrapper.trigger('mousedown', { clientX: 100, clientY: 100 })
    await wrapper.trigger('mousemove', { clientX: 120, clientY: 130 })
    expect(wrapper.find('img').element.style.transform).toBe('scale(2) translate(10px, 15px)')

    await wrapper.trigger('mouseleave')
    await wrapper.trigger('mousemove', { clientX: 150, clientY: 180 })
    // Should not change after mouseleave
    expect(wrapper.find('img').element.style.transform).toBe('scale(2) translate(10px, 15px)')
  })

  it('does not drag when not zoomed', async () => {
    const wrapper = mount(ImageViewer, {
      props: { ...defaultProps, zoom: 100 },
    })

    await wrapper.trigger('mousedown', { clientX: 100, clientY: 100 })
    await wrapper.trigger('mousemove', { clientX: 150, clientY: 150 })

    expect(wrapper.find('img').element.style.transform).toBe('scale(1)')
  })
})

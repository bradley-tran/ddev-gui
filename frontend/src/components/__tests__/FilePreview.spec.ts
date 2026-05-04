import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { FileIcon, FileTextIcon } from '@lucide/vue'
import FilePreview from '../FilePreview.vue'

describe('FilePreview', () => {
  it('renders empty state when fileName is null', () => {
    const wrapper = mount(FilePreview, {
      props: {
        fileName: null,
        content: '',
      },
    })

    expect(wrapper.text()).toContain('Select a file to preview')
    expect(wrapper.findComponent(FileTextIcon).exists()).toBe(true)
  })

  it('renders loading state when loading prop is true', () => {
    const wrapper = mount(FilePreview, {
      props: {
        fileName: 'test.txt',
        content: '',
        loading: true,
      },
    })

    expect(wrapper.text()).toContain('Loading…')
    expect(wrapper.find('.flu-spinner').exists()).toBe(true)
  })

  it('renders markdown preview for .md files', () => {
    const content = '# Hello Markdown'
    const wrapper = mount(FilePreview, {
      props: {
        fileName: 'README.md',
        content,
      },
    })

    expect(wrapper.findComponent({ name: 'MarkdownViewer' }).exists()).toBe(true)
    expect(wrapper.find('[data-testid="markdown-viewer"]').exists()).toBe(true)
  })

  it('renders image preview and handles zoom', async () => {
    const wrapper = mount(FilePreview, {
      props: {
        fileName: 'image.png',
        content: 'base64data',
      },
    })

    expect(wrapper.findComponent({ name: 'ImageViewer' }).exists()).toBe(true)
    const zoomLabel = wrapper.find('.fe-zoom-label')
    expect(zoomLabel.text()).toBe('100%')

    const zoomInBtn = wrapper.find('button[title="Zoom in"]')
    await zoomInBtn.trigger('click')
    expect(zoomLabel.text()).toBe('150%')

    const zoomOutBtn = wrapper.find('button[title="Zoom out"]')
    await zoomOutBtn.trigger('click')
    expect(zoomLabel.text()).toBe('100%')

    await zoomOutBtn.trigger('click')
    expect(zoomLabel.text()).toBe('75%')

    await zoomLabel.trigger('click') // Reset zoom
    expect(zoomLabel.text()).toBe('100%')
  })

  it('renders binary file message for binary files', () => {
    const wrapper = mount(FilePreview, {
      props: {
        fileName: 'test.exe',
        content: '',
      },
    })

    expect(wrapper.text()).toContain('Binary file - preview not available')
    expect(wrapper.findComponent(FileIcon).exists()).toBe(true)
  })

  it('renders code viewer for supported code files', () => {
    const content = 'console.log("hello");'
    const wrapper = mount(FilePreview, {
      props: {
        fileName: 'script.ts',
        content,
      },
    })

    expect(wrapper.findComponent({ name: 'CodeViewer' }).exists()).toBe(true)
    expect(wrapper.find('[data-testid="code-viewer"]').exists()).toBe(true)
  })

  it('falls back to pre tag for unknown text files', () => {
    const content = 'Some plain text'
    const wrapper = mount(FilePreview, {
      props: {
        fileName: 'unknown.foo',
        content,
      },
    })

    const pre = wrapper.find('pre.fe-preview-code')
    expect(pre.exists()).toBe(true)
    expect(pre.text()).toBe(content)
  })

  it('calls onRefresh when refresh button is clicked', async () => {
    const onRefresh = vi.fn()
    const wrapper = mount(FilePreview, {
      props: {
        fileName: 'test.txt',
        content: 'hello',
        onRefresh,
      },
    })

    const refreshBtn = wrapper.find('button[title="Refresh"]')
    expect(refreshBtn.exists()).toBe(true)
    await refreshBtn.trigger('click')
    expect(onRefresh).toHaveBeenCalled()
  })

  it('resets zoom when fileName changes', async () => {
    const wrapper = mount(FilePreview, {
      props: {
        fileName: 'image1.png',
        content: 'data1',
      },
    })

    const zoomInBtn = wrapper.find('button[title="Zoom in"]')
    await zoomInBtn.trigger('click')
    expect(wrapper.find('.fe-zoom-label').text()).toBe('150%')

    await wrapper.setProps({ fileName: 'image2.png', content: 'data2' })
    expect(wrapper.find('.fe-zoom-label').text()).toBe('100%')
  })
})

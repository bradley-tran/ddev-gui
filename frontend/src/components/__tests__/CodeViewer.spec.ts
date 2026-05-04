import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import CodeViewer from '../CodeViewer.vue'

describe('CodeViewer', () => {
  it('renders the component with basic props', () => {
    const wrapper = mount(CodeViewer, {
      props: {
        content: 'console.log("hello")',
        fileName: 'test.js',
      },
    })

    expect(wrapper.find('[data-testid="code-viewer"]').exists()).toBe(true)
    expect(wrapper.find('.code-viewer-lang').text()).toBe('javascript')
  })

  it('detects language correctly for typescript', () => {
    const wrapper = mount(CodeViewer, {
      props: {
        content: 'const x: number = 1',
        fileName: 'test.ts',
      },
    })

    expect(wrapper.find('.code-viewer-lang').text()).toBe('typescript')
    expect(wrapper.find('code').classes()).toContain('language-typescript')
  })

  it('falls back to plaintext for unknown extensions', () => {
    const wrapper = mount(CodeViewer, {
      props: {
        content: 'some random text',
        fileName: 'test.unknown',
      },
    })

    expect(wrapper.find('.code-viewer-lang').text()).toBe('plaintext')
    expect(wrapper.find('code').classes()).toContain('language-plaintext')
  })

  it('renders correct number of lines', () => {
    const content = 'line 1\nline 2\nline 3'
    const wrapper = mount(CodeViewer, {
      props: {
        content,
        fileName: 'test.txt',
      },
    })

    const lines = wrapper.find('.code-viewer-lines').text()
    expect(lines).toBe('1\n2\n3')
  })

  it('applies custom className', () => {
    const wrapper = mount(CodeViewer, {
      props: {
        content: 'test',
        fileName: 'test.txt',
        className: 'custom-class',
      },
    })

    expect(wrapper.find('.code-viewer').classes()).toContain('custom-class')
  })

  it('renders highlighted code HTML', () => {
    const content = 'const a = 1;'
    const wrapper = mount(CodeViewer, {
      props: {
        content,
        fileName: 'test.js',
      },
    })

    const codeElement = wrapper.find('code')
    expect(codeElement.exists()).toBe(true)
    // highlight.js should wrap 'const' in a span with 'hljs-keyword' class
    expect(codeElement.html()).toContain('hljs-keyword')
    expect(codeElement.text()).toBe(content)
  })

  it('handles empty content', () => {
    const wrapper = mount(CodeViewer, {
      props: {
        content: '',
        fileName: 'test.js',
      },
    })

    expect(wrapper.find('.code-viewer-lines').text()).toBe('1')
    expect(wrapper.find('code').text()).toBe('')
  })
})

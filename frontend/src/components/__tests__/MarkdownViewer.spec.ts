import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import MarkdownViewer from '../MarkdownViewer.vue'

describe('MarkdownViewer', () => {
  it('renders basic markdown correctly', () => {
    const content = '# Header 1\n\nThis is a paragraph.'
    const wrapper = mount(MarkdownViewer, {
      props: { content }
    })

    expect(wrapper.find('h1').text()).toBe('Header 1')
    expect(wrapper.find('p').text()).toBe('This is a paragraph.')
  })

  it('renders code blocks with syntax highlighting', () => {
    const content = '```javascript\nconst x = 1;\n```'
    const wrapper = mount(MarkdownViewer, {
      props: { content }
    })

    const pre = wrapper.find('pre')
    expect(pre.classes()).toContain('md-code-block')

    const code = pre.find('code')
    expect(code.classes()).toContain('hljs')
    expect(code.classes()).toContain('language-javascript')
    // highlight.js should wrap keywords
    expect(code.html()).toContain('hljs-keyword')
  })

  it('renders inline code with custom class', () => {
    const content = 'This is `inline code`.'
    const wrapper = mount(MarkdownViewer, {
      props: { content }
    })

    const code = wrapper.find('code')
    expect(code.classes()).toContain('md-inline-code')
    expect(code.text()).toBe('inline code')
  })

  it('renders links with custom class and attributes', () => {
    const content = '[DDEV](https://ddev.com)'
    const wrapper = mount(MarkdownViewer, {
      props: { content }
    })

    const link = wrapper.find('a')
    expect(link.attributes('href')).toBe('https://ddev.com')
    expect(link.classes()).toContain('md-link')
    expect(link.attributes('target')).toBe('_blank')
    expect(link.attributes('rel')).toBe('noopener noreferrer')
    expect(link.text()).toBe('DDEV')
  })

  it('sanitizes potentially malicious HTML', () => {
    const content = 'Hello <script>alert("xss")</script><img src=x onerror=alert(1)>'
    const wrapper = mount(MarkdownViewer, {
      props: { content }
    })

    expect(wrapper.html()).not.toContain('<script>')
    expect(wrapper.html()).not.toContain('onerror')
  })

  it('handles empty content', () => {
    const wrapper = mount(MarkdownViewer, {
      props: { content: '' }
    })

    expect(wrapper.text()).toBe('')
  })

  it('applies custom className prop', () => {
    const wrapper = mount(MarkdownViewer, {
      props: {
        content: 'test',
        className: 'my-custom-markdown'
      }
    })

    expect(wrapper.find('[data-testid="markdown-viewer"]').classes()).toContain('my-custom-markdown')
  })
})

<script setup lang="ts">
import { computed } from 'vue'
import DOMPurify from 'dompurify'
import hljs from 'highlight.js'
import { marked } from 'marked'

const props = withDefaults(
  defineProps<{
    content: string
    className?: string
  }>(),
  {
    className: '',
  },
)

marked.setOptions({
  gfm: true,
  breaks: true,
})

const renderer = new marked.Renderer()

renderer.code = function ({ text, lang }) {
  const language = lang && hljs.getLanguage(lang) ? lang : 'plaintext'
  const highlighted = hljs.highlight(text, { language }).value
  return `<pre class="md-code-block"><code class="hljs language-${language}">${highlighted}</code></pre>`
}

renderer.codespan = function ({ text }) {
  return `<code class="md-inline-code">${text}</code>`
}

renderer.link = function ({ href, text }) {
  return `<a href="${href || '#'}" class="md-link" target="_blank" rel="noopener noreferrer">${text}</a>`
}

const html = computed(() => {
  if (!props.content) return ''
  const rawHtml = marked.parse(props.content, { renderer }) as string
  return DOMPurify.sanitize(rawHtml, {
    ADD_ATTR: ['target'],
  })
})
</script>

<template>
  <div data-testid="markdown-viewer" :class="['md-viewer', props.className]" v-html="html" />
</template>
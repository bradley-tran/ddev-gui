<script setup lang="ts">
import { computed } from 'vue'
import DOMPurify from 'dompurify'
import hljs from 'highlight.js'
import { detectLanguage } from '@/lib/filePreview'

const props = withDefaults(
  defineProps<{
    content: string
    fileName: string
    className?: string
  }>(),
  {
    className: '',
  },
)

const language = computed(() => {
  const detected = detectLanguage(props.fileName)
  return hljs.getLanguage(detected) ? detected : 'plaintext'
})

const html = computed(() => {
  const result = hljs.highlight(props.content, { language: language.value })
  return DOMPurify.sanitize(result.value)
})

const lineCount = computed(() => {
  let count = 1
  let pos = props.content.indexOf('\n')
  while (pos !== -1) {
    count++
    pos = props.content.indexOf('\n', pos + 1)
  }
  return count
})

const lineNumbers = computed(() =>
  Array.from({ length: lineCount.value }, (_, index) => String(index + 1)).join('\n'),
)
</script>

<template>
  <div data-testid="code-viewer" :class="['code-viewer', props.className]">
    <div class="code-viewer-lang">{{ language }}</div>
    <div class="code-viewer-scroll">
      <pre class="code-viewer-lines" aria-hidden="true">{{ lineNumbers }}</pre>
      <pre class="code-viewer-code"><code :class="['hljs', `language-${language}`]" v-html="html" /></pre>
    </div>
  </div>
</template>
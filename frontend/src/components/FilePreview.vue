<script setup lang="ts">
import {
  FileIcon,
  FileTextIcon,
  RefreshCwIcon as RefreshIcon,
  ZoomInIcon,
  ZoomOutIcon,
} from '@lucide/vue'
import { computed, ref, watch } from 'vue'
import CodeViewer from '@/components/CodeViewer.vue'
import ImageViewer from '@/components/ImageViewer.vue'
import MarkdownViewer from '@/components/MarkdownViewer.vue'
import {
  isBinaryFile,
  isCodeFile,
  isImageFile,
  isMarkdownFile,
} from '@/lib/filePreview'

const ZOOM_STEPS = [25, 50, 75, 100, 150, 200, 300, 400]
const DEFAULT_ZOOM = 100
const MIN_ZOOM = ZOOM_STEPS[0] ?? DEFAULT_ZOOM
const MAX_ZOOM = ZOOM_STEPS[ZOOM_STEPS.length - 1] ?? DEFAULT_ZOOM

const props = withDefaults(
  defineProps<{
    fileName: string | null
    content: string
    loading?: boolean
    onRefresh?: (() => void) | undefined
  }>(),
  {
    loading: false,
    onRefresh: undefined,
  },
)

const zoom = ref(DEFAULT_ZOOM)

const markdown = computed(() => (props.fileName ? isMarkdownFile(props.fileName) : false))
const image = computed(() => (props.fileName ? isImageFile(props.fileName) : false))
const binary = computed(() => (props.fileName ? isBinaryFile(props.fileName) : false))
const code = computed(() => (props.fileName ? isCodeFile(props.fileName) : false))

watch(
  () => props.fileName,
  () => {
    zoom.value = DEFAULT_ZOOM
  },
  { immediate: true },
)

function handleZoomIn() {
  const next = ZOOM_STEPS.find((step) => step > zoom.value)
  if (next) zoom.value = next
}

function handleZoomOut() {
  const prev = [...ZOOM_STEPS].reverse().find((step) => step < zoom.value)
  if (prev) zoom.value = prev
}

function handleZoomReset() {
  zoom.value = DEFAULT_ZOOM
}

function triggerRefresh() {
  props.onRefresh?.()
}
</script>

<template>
  <div v-if="!fileName" class="fe-preview-empty">
    <FileTextIcon :size="32" :stroke-width="1.5" style="opacity: 0.3" />
    <p>Select a file to preview</p>
  </div>

  <div v-else-if="loading" class="fe-preview-empty">
    <span class="flu-spinner" />
    <p>Loading…</p>
  </div>

  <div v-else class="fe-preview-content">
    <div class="fe-preview-header">
      <span class="fe-preview-title">{{ fileName }}</span>
      <div v-if="image" class="fe-zoom-bar">
        <button
          type="button"
          class="fe-zoom-btn"
          :disabled="zoom <= MIN_ZOOM"
          title="Zoom out"
          @click="handleZoomOut"
        >
          <ZoomOutIcon :size="14" />
        </button>
        <button type="button" class="fe-zoom-label" title="Reset zoom" @click="handleZoomReset">
          {{ zoom }}%
        </button>
        <button
          type="button"
          class="fe-zoom-btn"
          :disabled="zoom >= MAX_ZOOM"
          title="Zoom in"
          @click="handleZoomIn"
        >
          <ZoomInIcon :size="14" />
        </button>
      </div>
      <button
        v-if="onRefresh && !binary"
        type="button"
        class="fe-zoom-btn"
        title="Refresh"
        :style="{ marginLeft: image ? '0' : 'auto' }"
        @click="triggerRefresh"
      >
        <RefreshIcon :size="14" />
      </button>
    </div>

    <div class="fe-preview-body">
      <MarkdownViewer v-if="markdown" :content="content" />
      <ImageViewer v-else-if="image" :data="content" :file-name="fileName" :zoom="zoom" />
      <div v-else-if="binary" class="fe-preview-empty">
        <FileIcon :size="32" :stroke-width="1.5" style="opacity: 0.3" />
        <p>Binary file - preview not available</p>
      </div>
      <CodeViewer v-else-if="code" :content="content" :file-name="fileName" />
      <pre v-else class="fe-preview-code">{{ content }}</pre>
    </div>
  </div>
</template>
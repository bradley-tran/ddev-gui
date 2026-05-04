<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { mimeFromExt } from '@/lib/filePreview'

const props = withDefaults(
  defineProps<{
    data: string
    fileName: string
    zoom?: number
    className?: string
  }>(),
  {
    zoom: 100,
    className: '',
  },
)

const offset = ref({ x: 0, y: 0 })

const dragState = {
  dragging: false,
  startX: 0,
  startY: 0,
  origX: 0,
  origY: 0,
}

const isZoomed = computed(() => props.zoom > 100)
const scale = computed(() => props.zoom / 100)
const src = computed(() => `data:${mimeFromExt(props.fileName)};base64,${props.data}`)

const transform = computed(() =>
  isZoomed.value
    ? `scale(${scale.value}) translate(${offset.value.x / scale.value}px, ${offset.value.y / scale.value}px)`
    : `scale(${scale.value})`,
)

watch(
  () => props.zoom,
  (zoom) => {
    if (zoom <= 100) {
      offset.value = { x: 0, y: 0 }
    }
  },
)

function handleMouseDown(event: MouseEvent) {
  if (!isZoomed.value) return

  event.preventDefault()
  dragState.dragging = true
  dragState.startX = event.clientX
  dragState.startY = event.clientY
  dragState.origX = offset.value.x
  dragState.origY = offset.value.y
}

function handleMouseMove(event: MouseEvent) {
  if (!dragState.dragging) return

  const dx = event.clientX - dragState.startX
  const dy = event.clientY - dragState.startY

  offset.value = {
    x: dragState.origX + dx,
    y: dragState.origY + dy,
  }
}

function stopDragging() {
  dragState.dragging = false
}
</script>

<template>
  <div
    data-testid="image-viewer"
    :class="['img-viewer', { 'img-viewer-zoomed': isZoomed }, props.className]"
    @mousedown="handleMouseDown"
    @mousemove="handleMouseMove"
    @mouseup="stopDragging"
    @mouseleave="stopDragging"
  >
    <img class="img-viewer-img" :src="src" :alt="fileName" :style="{ transform }" draggable="false" />
  </div>
</template>
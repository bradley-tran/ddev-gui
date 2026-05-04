<script setup lang="ts">
import {
  ChevronLeftIcon,
  FileIcon,
  FolderIcon,
  FolderOpenIcon,
  RefreshCwIcon as RefreshIcon,
} from '@lucide/vue'
import { computed, ref, watch } from 'vue'
import FilePreview from '@/components/FilePreview.vue'
import Spinner from '@/components/Spinner.vue'
import { cacheKey, dirCache, fileCache } from '@/lib/cache'
import { isBinaryFile, isImageFile } from '@/lib/filePreview'
import { DdevService } from '@/lib/wails'

interface FileEntry {
  name: string
  isDir: boolean
  size: string
  modified: string
}

const props = defineProps<{
  projectName: string
  projectRoot: string
}>()

const currentPath = ref('.')
const entries = ref<FileEntry[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const selectedFile = ref<string | null>(null)
const fileContent = ref('')
const loadingFile = ref(false)
const fileRequestId = ref(0)

const breadcrumbParts = computed(() => (currentPath.value === '.' ? [] : currentPath.value.split('/')))
const selectedFileName = computed(() => selectedFile.value?.split('/').pop() ?? null)

watch(
  () => props.projectName,
  async () => {
    clearSelection()
    await fetchDir('.')
  },
  { immediate: true },
)

function sortEntries(list: FileEntry[]): FileEntry[] {
  return [...list].sort((left, right) => {
    if (left.isDir !== right.isDir) return left.isDir ? -1 : 1
    return left.name.localeCompare(right.name)
  })
}

function clearSelection() {
  selectedFile.value = null
  fileContent.value = ''
  loadingFile.value = false
}

function filePathFor(entry: FileEntry): string {
  return currentPath.value === '.' ? entry.name : `${currentPath.value}/${entry.name}`
}

async function fetchDir(relPath: string) {
  const key = cacheKey(props.projectName, relPath)
  const cached = dirCache.get(key)

  if (cached) {
    entries.value = cached
    currentPath.value = relPath
    loading.value = false
    error.value = null
    return
  }

  loading.value = true
  error.value = null

  try {
    const raw = await DdevService.listDir(props.projectName, relPath)
    const parsed = JSON.parse(raw) as FileEntry[]
    const sorted = sortEntries(parsed)

    dirCache.set(key, sorted)
    entries.value = sorted
    currentPath.value = relPath
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err)
    entries.value = []
  } finally {
    loading.value = false
  }
}

async function goToPath(relPath: string) {
  clearSelection()
  await fetchDir(relPath)
}

async function handleNavigate(entry: FileEntry) {
  if (entry.isDir) {
    await goToPath(filePathFor(entry))
    return
  }

  await handleFileSelect(entry)
}

async function handleFileSelect(entry: FileEntry) {
  const filePath = filePathFor(entry)
  selectedFile.value = filePath
  fileContent.value = ''

  if (isBinaryFile(entry.name) && !isImageFile(entry.name)) {
    loadingFile.value = false
    return
  }

  const key = cacheKey(props.projectName, filePath)
  const cached = fileCache.get(key)
  if (cached !== undefined) {
    fileContent.value = cached
    loadingFile.value = false
    return
  }

  const requestId = ++fileRequestId.value
  loadingFile.value = true

  try {
    const content = isImageFile(entry.name)
      ? await DdevService.readFileBase64(props.projectName, filePath)
      : await DdevService.readFile(props.projectName, filePath)

    if (fileRequestId.value !== requestId) return

    fileCache.set(key, content)
    fileContent.value = content
  } catch (err) {
    if (fileRequestId.value !== requestId) return
    fileContent.value = `Error reading file: ${err instanceof Error ? err.message : String(err)}`
  } finally {
    if (fileRequestId.value === requestId) {
      loadingFile.value = false
    }
  }
}

async function handleGoUp() {
  if (currentPath.value === '.') return

  const parts = currentPath.value.split('/')
  parts.pop()
  await goToPath(parts.length === 0 ? '.' : parts.join('/'))
}

async function handleRefresh() {
  dirCache.delete(cacheKey(props.projectName, currentPath.value))

  const selectedPath = selectedFile.value
  if (selectedPath) {
    fileCache.delete(cacheKey(props.projectName, selectedPath))
  }

  await fetchDir(currentPath.value)

  if (!selectedPath) return

  const fileName = selectedPath.split('/').pop()
  if (!fileName) return

  await handleFileSelect({ name: fileName, isDir: false, size: '', modified: '' })
}
</script>

<template>
  <div class="fe-grid" data-testid="file-explorer">
    <div class="fe-panel-tree">
      <div class="fe-breadcrumb">
        <button
          type="button"
          class="fe-breadcrumb-item fe-breadcrumb-root"
          :title="projectRoot"
          @click="goToPath('.')"
        >
          <FolderOpenIcon :size="14" :stroke-width="2" />
          <span>{{ projectName }}</span>
        </button>

        <span v-for="(part, index) in breadcrumbParts" :key="`${part}-${index}`" class="fe-breadcrumb-sep-wrap">
          <span class="fe-breadcrumb-sep">/</span>
          <button
            type="button"
            class="fe-breadcrumb-item"
            @click="goToPath(breadcrumbParts.slice(0, index + 1).join('/'))"
          >
            {{ part }}
          </button>
        </span>

        <button type="button" class="fe-refresh-btn" title="Refresh" @click="handleRefresh">
          <RefreshIcon :size="13" :stroke-width="2" />
        </button>
      </div>

      <div class="fe-list">
        <div v-if="loading" class="fe-loading">
          <Spinner /> Loading…
        </div>
        <div v-else-if="error" class="fe-error">{{ error }}</div>
        <div v-else-if="entries.length === 0" class="fe-empty">Empty directory</div>
        <template v-else>
          <button
            v-if="currentPath !== '.'"
            type="button"
            class="fe-row fe-row-up"
            data-up-directory="true"
            @click="handleGoUp"
          >
            <span class="fe-icon"><ChevronLeftIcon :size="14" :stroke-width="2" /></span>
            <span class="fe-name">..</span>
          </button>

          <button
            v-for="entry in entries"
            :key="entry.name"
            type="button"
            class="fe-row"
            :class="[
              entry.isDir ? 'fe-row-dir' : 'fe-row-file',
              filePathFor(entry) === selectedFile ? 'fe-row-selected' : '',
            ]"
            :data-entry-path="filePathFor(entry)"
            @click="handleNavigate(entry)"
          >
            <span class="fe-icon">
              <FolderIcon v-if="entry.isDir" :size="14" :stroke-width="1.5" style="color: var(--accent)" />
              <FileIcon v-else :size="14" :stroke-width="1.5" />
            </span>
            <span class="fe-name">{{ entry.name }}</span>
          </button>
        </template>
      </div>
    </div>

    <div class="fe-panel-preview">
      <FilePreview :file-name="selectedFileName" :content="fileContent" :loading="loadingFile" />
    </div>
  </div>
</template>
<script setup lang="ts">
import { ChevronsUpDownIcon, RefreshCwIcon as RefreshIcon } from '@lucide/vue'
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import Spinner from '@/components/Spinner.vue'
import { ansiToHtml } from '@/lib/ansi'
import { useTranslation } from '@/lib/i18n'
import { DdevService } from '@/lib/wails'

const props = defineProps<{
  projectName: string
  serviceNames?: string[]
}>()

const { t } = useTranslation()
const DEFAULT_LOG_SERVICE = 'web'

const loading = ref(false)
const logs = ref('')
const errorMessage = ref('')
const selectedService = ref(DEFAULT_LOG_SERVICE)
const serviceMenuOpen = ref(false)
const serviceMenuRef = ref<HTMLElement | null>(null)
let requestId = 0

const logsHtml = computed(() => ansiToHtml(logs.value))
const serviceOptions = computed(() => {
  const orderedNames = [DEFAULT_LOG_SERVICE]

  for (const rawServiceName of props.serviceNames ?? []) {
    const serviceName = String(rawServiceName ?? '').trim()
    if (!serviceName || orderedNames.includes(serviceName)) continue
    orderedNames.push(serviceName)
  }

  return orderedNames.map((serviceName) => ({
    value: serviceName,
    label: serviceName,
  }))
})

watch(
  () => props.projectName,
  (projectName, previousProjectName) => {
    if (projectName !== previousProjectName) {
      selectedService.value = DEFAULT_LOG_SERVICE
      closeServiceMenu()
    }

    if (!projectName) {
      logs.value = ''
      errorMessage.value = ''
      closeServiceMenu()
    }
  },
)

watch(serviceOptions, (options) => {
  if (!options.some((option) => option.value === selectedService.value)) {
    selectedService.value = DEFAULT_LOG_SERVICE
    closeServiceMenu()
  }
})

watch(serviceMenuOpen, (isOpen, _previous, onCleanup) => {
  if (!isOpen) return

  const handlePointerDownOutside = (event: MouseEvent) => {
    const target = event.target as Node | null
    if (!target || serviceMenuRef.value?.contains(target)) return
    closeServiceMenu()
  }

  document.addEventListener('mousedown', handlePointerDownOutside)

  onCleanup(() => {
    document.removeEventListener('mousedown', handlePointerDownOutside)
  })
})

watch(
  [() => props.projectName, selectedService],
  async ([projectName]) => {
    if (!projectName) return
    await loadLogs(projectName, selectedService.value)
  },
  { immediate: true },
)

async function loadLogs(projectName = props.projectName, service = selectedService.value) {
  const currentRequestId = ++requestId
  loading.value = true
  errorMessage.value = ''

  try {
    const output = await DdevService.ProjectLogs(projectName, service)
    if (currentRequestId !== requestId) return
    logs.value = String(output ?? '').trim()
  } catch (error) {
    if (currentRequestId !== requestId) return
    const message = error instanceof Error ? error.message : String(error)
    errorMessage.value = message || t('detail.logs.loadError')
    logs.value = ''
  } finally {
    if (currentRequestId === requestId) {
      loading.value = false
    }
  }
}

function toggleServiceMenu() {
  if (loading.value) return
  serviceMenuOpen.value = !serviceMenuOpen.value
}

function closeServiceMenu() {
  serviceMenuOpen.value = false
}

function selectService(service: string) {
  selectedService.value = service
  closeServiceMenu()
}

onBeforeUnmount(() => {
  closeServiceMenu()
})
</script>

<template>
  <div id="detailLogs" data-testid="project-logs">
    <section class="detail-section">
      <div class="detail-section-title">
        <div class="project-logs-heading">
          <span>{{ t('detail.logs.title') }}</span>
          <div
            ref="serviceMenuRef"
            class="toolbar-dropdown project-logs-service-picker"
            @keydown.escape.prevent="closeServiceMenu"
          >
            <span class="project-logs-service-label">{{ t('detail.services.title') }}</span>
            <button
              type="button"
              class="flu-btn flu-btn-sm flu-btn-ghost toolbar-dropdown-toggle project-logs-service-toggle"
              :disabled="loading"
              :aria-expanded="serviceMenuOpen"
              :aria-label="`${t('detail.services.title')}: ${selectedService}`"
              aria-haspopup="menu"
              data-testid="project-logs-service-toggle"
              @click="toggleServiceMenu"
            >
              <span class="project-logs-service-current">{{ selectedService }}</span>
              <ChevronsUpDownIcon class="project-logs-service-icon" :size="13" :stroke-width="2.1" />
            </button>

            <div v-if="serviceMenuOpen" class="toolbar-dropdown-menu project-logs-service-menu">
              <div class="toolbar-dropdown-menu-inner project-logs-service-menu-inner">
                <button
                  v-for="option in serviceOptions"
                  :key="option.value"
                  type="button"
                  class="toolbar-dropdown-item project-logs-service-item"
                  :class="{
                    'project-logs-service-item-selected': option.value === selectedService,
                  }"
                  :data-testid="`project-logs-service-option-${option.value}`"
                  @click="selectService(option.value)"
                >
                  {{ option.label }}
                </button>
              </div>
            </div>
          </div>
        </div>
        <button
          type="button"
          class="flu-btn flu-btn-sm flu-btn-ghost project-logs-refresh"
          :disabled="loading"
          data-testid="project-logs-refresh"
          @click="loadLogs()"
        >
          <RefreshIcon :size="12" :stroke-width="2" />
          {{ t('detail.logs.refresh') }}
        </button>
      </div>
      <div class="detail-section-body">
        <div v-if="loading" class="detail-loading-row">
          <Spinner />
          <span>{{ t('general.loading') }}</span>
        </div>

        <div v-else-if="errorMessage" class="project-logs-error" role="alert">
          {{ t('detail.logs.loadError') }} {{ errorMessage }}
        </div>

        <div v-else-if="!logs" class="text-muted">
          {{ t('detail.logs.empty') }}
        </div>

        <div v-else class="project-logs-output" v-html="logsHtml" />
      </div>
    </section>
  </div>
</template>

<style scoped>
.detail-loading-row {
  display: inline-flex;
  align-items: center;
  gap: 0.65rem;
}

.project-logs-heading {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  min-width: 0;
  flex-wrap: wrap;
}

.project-logs-service-picker {
  display: inline-flex;
  align-items: center;
  position: relative;
}

.project-logs-service-label {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.project-logs-service-toggle {
  gap: 0.35rem;
  padding: 0.2rem 0.5rem;
  text-transform: none;
}

.project-logs-service-current {
  min-width: 2.25rem;
  text-align: left;
}

.project-logs-service-icon {
  opacity: 0.75;
}

.project-logs-service-menu {
  left: 0;
  right: auto;
}

.project-logs-service-menu-inner {
  min-width: 6.5rem;
  backdrop-filter: blur(5px);
}

.project-logs-service-item-selected {
  color: var(--accent);
  font-weight: 600;
}

.project-logs-refresh {
  gap: 0.4rem;
}

.project-logs-error {
  color: var(--danger, #d92d20);
  white-space: pre-wrap;
}

.project-logs-output {
  max-height: min(62vh, 44rem);
  overflow: auto;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background: var(--bg-embed);
  color: #c9d1d9;
  padding: 0.85rem;
  font-family: 'Cascadia Code', 'Fira Code', 'JetBrains Mono', 'Consolas', 'Monaco', monospace;
  font-size: 12.5px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
}
</style>

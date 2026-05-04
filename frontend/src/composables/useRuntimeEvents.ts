import { onBeforeUnmount, onMounted } from 'vue'
import type { AppModal } from '@/lib/types'
import { Runtime } from '@/lib/wails'
import { useAppStore } from '@/stores/app'

const MENU_MODAL_EVENTS: Record<string, AppModal> = {
  'menu:new': 'newProject',
  'menu:env': 'envInfo',
  'menu:settings': 'settings',
}

export function useRuntimeEvents() {
  const appStore = useAppStore()

  const projectStatusHandler = (data: unknown) => {
    const detail = data as { name?: string; status?: string }
    if (detail?.name && detail?.status) {
      appStore.appLog(`${detail.name}: ${detail.status}`, 'info')
    }
  }

  const infoHandler = (data: unknown) => {
    const detail = data as { message?: string }
    if (detail?.message) {
      appStore.showToast(detail.message, 'info', 6000)
    }
  }

  const errorHandler = (data: unknown) => {
    const detail = data as { message?: string }
    if (detail?.message) {
      appStore.showToast(detail.message, 'error', 6000)
    }
  }

  const outputHandler = (line: unknown) => {
    if (typeof line === 'string' && line.trim()) {
      appStore.appLog(line, 'output')
    }
  }

  onMounted(() => {
    Runtime.on('project:status', projectStatusHandler)
    Runtime.on('ui:info', infoHandler)
    Runtime.on('ui:error', errorHandler)
    Runtime.on('ddev:output', outputHandler)

    for (const [eventName, modal] of Object.entries(MENU_MODAL_EVENTS)) {
      Runtime.on(eventName, () => appStore.openModal(modal))
    }
  })

  onBeforeUnmount(() => {
    Runtime.off('project:status')
    Runtime.off('ui:info')
    Runtime.off('ui:error')
    Runtime.off('ddev:output')

    for (const eventName of Object.keys(MENU_MODAL_EVENTS)) {
      Runtime.off(eventName)
    }
  })
}
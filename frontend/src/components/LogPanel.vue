<script setup lang="ts">
import { nextTick, watch } from 'vue'
import { ansiToHtml, escapeHtml } from '@/lib/ansi'
import type { LogEntry } from '@/lib/types'
import { useTranslation } from '@/lib/i18n'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()
const { t } = useTranslation()

const bottomRef = defineModel<HTMLDivElement | null>('bottomRef', { default: null })

const isHidden = () => appStore.config.showLog === false || appStore.terminalActive

watch(
  () => appStore.logEntries.length,
  async () => {
    await nextTick()
    bottomRef.value?.scrollIntoView({ behavior: 'smooth' })
  },
)

function messageHtml(entry: LogEntry): string {
  return entry.level === 'output' ? ansiToHtml(entry.message) : escapeHtml(entry.message)
}
</script>

<template>
  <div v-if="!isHidden()" class="flu-card log-card">
    <div class="flu-card-header">
      <h2>{{ t('log.title') }}</h2>
      <div class="header-controls">
        <button id="logClearBtn" type="button" class="flu-btn flu-btn-sm flu-btn-ghost" @click="appStore.clearLog()">
          {{ t('log.clear') }}
        </button>
      </div>
    </div>
    <div class="flu-card-body" style="padding: 0">
      <div id="logOutput" class="log-area">
        <div v-for="entry in appStore.logEntries" :key="entry.id" class="log-entry" :class="`log-${entry.level}`">
          <span class="log-time">{{ entry.timestamp }}</span>
          <span class="log-msg" v-html="messageHtml(entry)" />
        </div>
        <div ref="bottomRef" />
      </div>
    </div>
  </div>
</template>
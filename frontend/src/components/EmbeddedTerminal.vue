<script setup lang="ts">
import { EraserIcon, SendIcon, TerminalIcon } from '@lucide/vue'
import { nextTick, ref, watch } from 'vue'
import { ansiToHtml, escapeHtml, linkifyHtmlUrls } from '@/lib/ansi'
import { useTranslation } from '@/lib/i18n'
import { openUrl } from '@/lib/utils'
import { DdevService, Runtime } from '@/lib/wails'

interface TerminalLine {
  id: number
  type: 'input' | 'output' | 'error' | 'info'
  text: string
}

const props = defineProps<{
  projectName: string
}>()

const { t } = useTranslation()

let lineIdCounter = 0

const lines = ref<TerminalLine[]>([])
const input = ref('')
const running = ref(false)
const history = ref<string[]>([])
const historyIdx = ref(-1)
const outputRef = ref<HTMLDivElement | null>(null)
const inputRef = ref<HTMLInputElement | null>(null)
const userScrolled = ref(false)

watch(
  () => props.projectName,
  (projectName, _previousProject, onCleanup) => {
    lines.value = [{ id: lineIdCounter++, type: 'info', text: `ddev ssh - ${projectName}` }]
    input.value = ''
    history.value = []
    historyIdx.value = -1
    running.value = false
    userScrolled.value = false

    const outputEvent = `terminal:output:${projectName}`
    const doneEvent = `terminal:done:${projectName}`

    const handleOutput = (...args: unknown[]) => {
      lines.value = [...lines.value, { id: lineIdCounter++, type: 'output', text: String(args[0] ?? '') }]
    }

    const handleDone = (...args: unknown[]) => {
      const exitCode = Number(args[0] ?? 0)
      if (exitCode !== 0) {
        lines.value = [
          ...lines.value,
          { id: lineIdCounter++, type: 'error', text: `Process exited with code ${exitCode}` },
        ]
      }
      running.value = false
    }

    Runtime.on(outputEvent, handleOutput)
    Runtime.on(doneEvent, handleDone)

    onCleanup(() => {
      Runtime.off(outputEvent, handleOutput)
      Runtime.off(doneEvent, handleDone)
    })
  },
  { immediate: true },
)

watch(
  lines,
  async () => {
    await nextTick()
    scrollToBottom()
  },
  { deep: true },
)

watch(
  running,
  async (isRunning) => {
    if (isRunning) return
    await nextTick()
    inputRef.value?.focus()
  },
  { immediate: true },
)

function scrollToBottom() {
  if (!outputRef.value || userScrolled.value) return
  outputRef.value.scrollTop = outputRef.value.scrollHeight
}

function handleScroll() {
  if (!outputRef.value) return
  const element = outputRef.value
  const atBottom = element.scrollHeight - element.scrollTop - element.clientHeight < 30
  userScrolled.value = !atBottom
}

async function handleSubmit() {
  const command = input.value.trim()
  if (!command || running.value) return

  if (command === 'clear' || command === 'cls') {
    lines.value = [{ id: lineIdCounter++, type: 'info', text: `ddev ssh - ${props.projectName}` }]
    input.value = ''
    historyIdx.value = -1
    return
  }

  history.value = [command, ...history.value.filter((entry) => entry !== command)].slice(0, 100)
  historyIdx.value = -1
  lines.value = [...lines.value, { id: lineIdCounter++, type: 'input', text: `$ ${command}` }]
  input.value = ''
  running.value = true
  userScrolled.value = false

  try {
    await DdevService.execCommand(props.projectName, command)
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error)
    if (!message.includes('exit status')) {
      lines.value = [...lines.value, { id: lineIdCounter++, type: 'error', text: message }]
    }
    running.value = false
  }
}

function handleKeyDown(event: KeyboardEvent) {
  if (event.key === 'ArrowUp') {
    event.preventDefault()
    if (history.value.length === 0) return

    const nextIndex = historyIdx.value < history.value.length - 1 ? historyIdx.value + 1 : historyIdx.value
    const nextEntry = history.value[nextIndex]
    if (nextEntry === undefined) return

    historyIdx.value = nextIndex
    input.value = nextEntry
    return
  }

  if (event.key === 'ArrowDown') {
    event.preventDefault()
    if (historyIdx.value <= 0) {
      historyIdx.value = -1
      input.value = ''
      return
    }

    const nextIndex = historyIdx.value - 1
    const nextEntry = history.value[nextIndex]
    if (nextEntry === undefined) return

    historyIdx.value = nextIndex
    input.value = nextEntry
  }
}

function handleClear() {
  lines.value = [{ id: lineIdCounter++, type: 'info', text: `ddev ssh - ${props.projectName}` }]
  inputRef.value?.focus()
}

function handleContainerClick(event: MouseEvent) {
  if (!(event.target instanceof Element)) return
  if (event.target.closest('.terminal-output')) return
  inputRef.value?.focus()
}

function handleOutputClick(event: MouseEvent) {
  if (!(event.target instanceof Element)) return

  const link = event.target.closest('a[data-terminal-url]')
  if (!(link instanceof HTMLAnchorElement)) return

  event.preventDefault()
  event.stopPropagation()
  openUrl(link.dataset.terminalUrl ?? link.getAttribute('href') ?? '')
}

function lineHtml(line: TerminalLine): string {
  if (line.type === 'output') return linkifyHtmlUrls(ansiToHtml(line.text))
  if (line.type === 'input') return `<span class="terminal-prompt">${escapeHtml(line.text)}</span>`
  return escapeHtml(line.text)
}
</script>

<template>
  <div class="embedded-terminal" data-testid="embedded-terminal" @click="handleContainerClick">
    <div class="terminal-toolbar">
      <span class="terminal-title">
        <TerminalIcon :size="14" :stroke-width="2" />
        {{ t('detail.tabTerminal') }}
      </span>
      <button
        type="button"
        class="flu-btn flu-btn-xs flu-btn-ghost terminal-clear-btn"
        title="Clear terminal"
        data-testid="embedded-terminal-clear"
        @click.stop="handleClear"
      >
        <EraserIcon :size="12" :stroke-width="2" />
        Clear
      </button>
    </div>

    <div ref="outputRef" class="terminal-output" @scroll="handleScroll" @click="handleOutputClick">
      <div
        v-for="line in lines"
        :key="line.id"
        :class="['terminal-line', `terminal-line-${line.type}`]"
        v-html="lineHtml(line)"
      />
      <div v-if="running" class="terminal-line terminal-running-indicator">
        <span class="terminal-spinner" />
      </div>
    </div>

    <form class="terminal-input-row" @submit.prevent="handleSubmit">
      <span class="terminal-prompt-char">$</span>
      <input
        ref="inputRef"
        v-model="input"
        class="terminal-input"
        type="text"
        :disabled="running"
        :placeholder="running ? 'Running…' : 'Type a command to run in container'"
        autocomplete="off"
        spellcheck="false"
        data-testid="embedded-terminal-input"
        @keydown="handleKeyDown"
      />
      <button
        type="submit"
        class="terminal-send-btn"
        :disabled="running || !input.trim()"
        title="Send command"
      >
        <SendIcon :size="14" :stroke-width="2" />
      </button>
    </form>
  </div>
</template>
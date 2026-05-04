<script setup lang="ts">
import { ChevronDownIcon } from '@lucide/vue'
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'

interface SelectOption {
  value: string
  label: string
}

const props = withDefaults(defineProps<{
  id?: string
  className?: string
  modelValue: string
  options: SelectOption[]
  disabled?: boolean
}>(), {
  id: undefined,
  className: '',
  disabled: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const open = ref(false)
const focusedIndex = ref(-1)
const triggerRef = ref<HTMLButtonElement | null>(null)
const menuRef = ref<HTMLUListElement | null>(null)
const menuStyle = ref<Record<string, string | number>>({})
const searchBuffer = ref('')
let searchTimer: number | undefined

const MENU_GAP = 2
const VIEWPORT_MARGIN = 8
const DEFAULT_MENU_MAX_HEIGHT = 220

const selectedLabel = computed(
  () => props.options.find((option) => option.value === props.modelValue)?.label ?? props.modelValue,
)

const selectedIndex = computed(
  () => props.options.findIndex((option) => option.value === props.modelValue),
)

function updatePosition() {
  const trigger = triggerRef.value
  if (!trigger) return

  const rect = trigger.getBoundingClientRect()
  const menu = menuRef.value
  const viewportWidth = window.innerWidth
  const viewportHeight = window.innerHeight

  const menuWidth = menu?.offsetWidth ?? rect.width
  const measuredMenuHeight = menu?.offsetHeight ?? 0
  const menuHeight = measuredMenuHeight > 0 ? measuredMenuHeight : DEFAULT_MENU_MAX_HEIGHT

  const minLeft = VIEWPORT_MARGIN
  const maxLeft = Math.max(VIEWPORT_MARGIN, viewportWidth - menuWidth - VIEWPORT_MARGIN)
  const left = Math.min(Math.max(rect.left, minLeft), maxLeft)

  const spaceBelow = viewportHeight - rect.bottom - VIEWPORT_MARGIN - MENU_GAP
  const spaceAbove = rect.top - VIEWPORT_MARGIN - MENU_GAP
  const openUpward = spaceBelow < Math.min(menuHeight, DEFAULT_MENU_MAX_HEIGHT) && spaceAbove > spaceBelow

  const availableHeight = openUpward ? spaceAbove : spaceBelow
  const maxHeight = Math.max(96, Math.floor(availableHeight))
  const top = openUpward
    ? Math.max(VIEWPORT_MARGIN, rect.top - Math.min(menuHeight, maxHeight) - MENU_GAP)
    : rect.bottom + MENU_GAP

  menuStyle.value = {
    position: 'fixed',
    top: `${Math.round(top)}px`,
    left: `${Math.round(left)}px`,
    width: `${Math.round(rect.width)}px`,
    maxHeight: `${maxHeight}px`,
  }
}

function openMenu() {
  focusedIndex.value = selectedIndex.value >= 0 ? selectedIndex.value : 0
  open.value = true
}

function closeMenu() {
  open.value = false
}

function commitValue(value: string) {
  emit('update:modelValue', value)
}

function focusTrigger() {
  triggerRef.value?.focus()
}

function typeAheadJump(char: string) {
  window.clearTimeout(searchTimer)
  searchBuffer.value += char.toLowerCase()
  searchTimer = window.setTimeout(() => {
    searchBuffer.value = ''
  }, 500)

  const query = searchBuffer.value
  const startFrom = focusedIndex.value >= 0 ? focusedIndex.value + 1 : 0
  for (let index = 0; index < props.options.length; index += 1) {
    const optionIndex = (startFrom + index) % props.options.length
    const option = props.options[optionIndex]
    if (!option) continue

    if (option.label.toLowerCase().startsWith(query)) {
      if (open.value) {
        focusedIndex.value = optionIndex
      } else {
        commitValue(option.value)
      }
      return
    }
  }
}

function handleKeyDown(event: KeyboardEvent) {
  if (props.disabled) return

  if (!open.value) {
    if (['ArrowDown', 'ArrowUp', 'Enter', ' '].includes(event.key)) {
      event.preventDefault()
      openMenu()
      return
    }
    if (event.key.length === 1 && !event.ctrlKey && !event.metaKey && !event.altKey) {
      typeAheadJump(event.key)
    }
    return
  }

  switch (event.key) {
    case 'ArrowDown':
      event.preventDefault()
      focusedIndex.value = (focusedIndex.value + 1) % props.options.length
      break
    case 'ArrowUp':
      event.preventDefault()
      focusedIndex.value = (focusedIndex.value - 1 + props.options.length) % props.options.length
      break
    case 'Enter':
    case ' ': {
      event.preventDefault()
      const option = props.options[focusedIndex.value]
      if (option) {
        commitValue(option.value)
        closeMenu()
        focusTrigger()
      }
      break
    }
    case 'Escape':
      event.preventDefault()
      closeMenu()
      focusTrigger()
      break
    case 'Home':
      event.preventDefault()
      focusedIndex.value = 0
      break
    case 'End':
      event.preventDefault()
      focusedIndex.value = props.options.length - 1
      break
    default:
      if (event.key.length === 1 && !event.ctrlKey && !event.metaKey && !event.altKey) {
        typeAheadJump(event.key)
      }
      break
  }
}

watch(
  open,
  (isOpen, _previous, onCleanup) => {
    if (!isOpen) return

    updatePosition()
    void nextTick(updatePosition)

    const frameId = window.requestAnimationFrame(updatePosition)
    const delayedUpdateId = window.setTimeout(updatePosition, 220)

    const reposition = () => updatePosition()
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Node
      if (
        triggerRef.value && !triggerRef.value.contains(target) &&
        menuRef.value && !menuRef.value.contains(target)
      ) {
        closeMenu()
      }
    }

    window.addEventListener('scroll', reposition, true)
    window.addEventListener('resize', reposition)
    document.addEventListener('mousedown', handleClickOutside)

    onCleanup(() => {
      window.cancelAnimationFrame(frameId)
      window.clearTimeout(delayedUpdateId)
      window.removeEventListener('scroll', reposition, true)
      window.removeEventListener('resize', reposition)
      document.removeEventListener('mousedown', handleClickOutside)
    })
  },
)

watch(
  () => props.disabled,
  (isDisabled) => {
    if (isDisabled) {
      closeMenu()
    }
  },
)

watch(
  [open, focusedIndex],
  async ([isOpen, currentFocusedIndex]) => {
    if (!isOpen || currentFocusedIndex < 0) return
    await nextTick()
    const item = menuRef.value?.children[currentFocusedIndex] as HTMLElement | undefined
    item?.scrollIntoView({ block: 'nearest' })
  },
)

onBeforeUnmount(() => {
  window.clearTimeout(searchTimer)
})
</script>

<template>
  <div :id="props.id" :class="['custom-select', props.className]">
    <button
      ref="triggerRef"
      type="button"
      class="custom-select-trigger"
      :disabled="props.disabled"
      :aria-disabled="props.disabled"
      :aria-expanded="open"
      aria-haspopup="listbox"
      @click="props.disabled ? undefined : (open ? closeMenu() : openMenu())"
      @keydown="handleKeyDown"
    >
      <span class="custom-select-value">{{ selectedLabel }}</span>
      <ChevronDownIcon class="custom-select-arrow" :size="12" :stroke-width="2.5" />
    </button>

    <Teleport to="body">
      <ul
        v-if="open"
        ref="menuRef"
        class="custom-select-menu"
        role="listbox"
        :style="menuStyle"
      >
        <li
          v-for="(option, index) in props.options"
          :key="option.value"
          class="custom-select-option"
          :class="{
            selected: option.value === props.modelValue,
            focused: index === focusedIndex,
          }"
          role="option"
          :aria-selected="option.value === props.modelValue"
          @click="commitValue(option.value); closeMenu(); focusTrigger()"
          @mouseenter="focusedIndex = index"
        >
          {{ option.label }}
        </li>
      </ul>
    </Teleport>
  </div>
</template>
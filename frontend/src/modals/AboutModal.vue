<script setup lang="ts">
import { onMounted, ref } from 'vue'
import Modal from '@/components/Modal.vue'
import Spinner from '@/components/Spinner.vue'
import { useTranslation } from '@/lib/i18n'
import { DdevService } from '@/lib/wails'

const emit = defineEmits<{
  close: []
}>()

const { t } = useTranslation()
const ddevVersion = ref('')
const appVersion = ref('')
const loading = ref(true)

onMounted(() => {
  Promise.all([
    DdevService.ddevInstalledVersion()
      .then((version) => { ddevVersion.value = version || 'not installed' })
      .catch(() => { ddevVersion.value = 'unknown' }),
    DdevService.appVersion()
      .then((version) => {
        const semanticVersion = version.version || 'dev'
        const hash = version.commitHash && version.commitHash !== 'unknown' ? version.commitHash : ''
        appVersion.value = hash ? `${semanticVersion} (${hash})` : semanticVersion
      })
      .catch(() => { appVersion.value = 'dev' }),
  ]).finally(() => {
    loading.value = false
  })
})
</script>

<template>
  <Modal :title="t('about.title')" @close="emit('close')">
    <template #footer>
      <button class="flu-btn flu-btn-ghost" type="button" @click="emit('close')">
        {{ t('general.close') }}
      </button>
    </template>

    <div style="text-align: center; padding: var(--space-lg) 0">
      <div style="margin-bottom: var(--space-lg)">
        <span
          style="display: inline-flex; align-items: center; justify-content: center; width: 56px; height: 56px; border-radius: 50%; border: 2.5px solid var(--accent); font-family: Georgia, 'Times New Roman', serif; font-style: italic; font-weight: 700; font-size: 32px; color: var(--accent); line-height: 1; user-select: none"
        >i</span>
      </div>

      <h2 style="font-size: 20px; font-weight: 800; margin: 0 0 4px 0">
        <span style="color: #38bdf8">DDEV</span>
        <span style="font-weight: 300; opacity: 0.9"> GUI</span>
      </h2>

      <p style="color: var(--text-secondary); font-size: 12px; margin: 0 0 var(--space-lg) 0">
        {{ t('about.tagline') }}
      </p>

      <div
        style="display: inline-flex; flex-direction: column; gap: var(--space-sm); text-align: left; background: var(--bg-subtle); border-radius: var(--radius-md); padding: var(--space-md) var(--space-xl); border: 1px solid var(--border-subtle)"
      >
        <div style="display: flex; gap: var(--space-xl); justify-content: space-between">
          <span style="font-size: 11px; font-weight: 600; color: var(--text-secondary); text-transform: uppercase; letter-spacing: 0.04em">
            {{ t('about.appVersion') }}
          </span>
          <span style="font-size: 13px; color: var(--text-primary)">
            <Spinner v-if="loading" />
            <template v-else>{{ appVersion }}</template>
          </span>
        </div>
        <div style="display: flex; gap: var(--space-xl); justify-content: space-between">
          <span style="font-size: 11px; font-weight: 600; color: var(--text-secondary); text-transform: uppercase; letter-spacing: 0.04em">
            {{ t('about.ddevVersion') }}
          </span>
          <span style="font-size: 13px; color: var(--text-primary)">
            <Spinner v-if="loading" />
            <template v-else>{{ ddevVersion }}</template>
          </span>
        </div>
        <div style="display: flex; gap: var(--space-xl); justify-content: space-between">
          <span style="font-size: 11px; font-weight: 600; color: var(--text-secondary); text-transform: uppercase; letter-spacing: 0.04em">
            {{ t('about.builtWith') }}
          </span>
          <span style="font-size: 13px; color: var(--text-primary)">Wails v2</span>
        </div>
      </div>
    </div>
  </Modal>
</template>
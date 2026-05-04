<script setup lang="ts">
import { onMounted, ref } from 'vue'
import Modal from '@/components/Modal.vue'
import Spinner from '@/components/Spinner.vue'
import { useTranslation } from '@/lib/i18n'
import { coerceToBool, openUrl } from '@/lib/utils'
import { DdevService as DdevApi } from '@/lib/wails'
import { useAppStore } from '@/stores/app'

interface MasqueradeUser {
  uid: string
  name: string
  mail: string
}

const props = defineProps<{
  projectName: string
  primaryUrl: string
}>()

const emit = defineEmits<{
  close: []
}>()

const appStore = useAppStore()
const { t } = useTranslation()

const users = ref<MasqueradeUser[]>([])
const loading = ref(false)
const uid = ref('')
const running = ref(false)

onMounted(() => {
  void loadUsers()
})

async function loadUsers() {
  uid.value = ''
  loading.value = true

  try {
    const json = await DdevApi.drushRecentUsers(props.projectName)
    const parsed = JSON.parse(json) as unknown
    users.value = Array.isArray(parsed) ? (parsed as MasqueradeUser[]) : []
  } catch {
    users.value = []
  } finally {
    loading.value = false
  }
}

async function doMasquerade(rawUid: string) {
  const normalizedUid = rawUid.trim()
  if (!normalizedUid || running.value) return

  running.value = true
  appStore.appLog(`Masquerading as user ${normalizedUid} on ${props.projectName}...`, 'info')

  try {
    let uliUrl = await DdevApi.drushUliAsUser(props.projectName, normalizedUid)
    if (uliUrl && !/^https?:\/\//i.test(uliUrl)) {
      const base = props.primaryUrl.replace(/\/+$/, '')
      const path = uliUrl.startsWith('/') ? uliUrl : `/${uliUrl}`
      uliUrl = `${base}${path}`
    }

    if (uliUrl) {
      openUrl(uliUrl, coerceToBool(appStore.config.openLinksInBrowser))
    }

    emit('close')
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error)
    appStore.appLog(`Masquerade failed: ${message}`, 'error')
    appStore.showToast('Masquerade failed', 'error')
  } finally {
    running.value = false
  }
}
</script>

<template>
  <Modal :title="t('detail.drupal.masqTitle')" wide @close="emit('close')">
    <div class="flu-field" style="margin-bottom: var(--space-sm)">
      <label class="flu-label">{{ t('detail.drupal.masqSelectUser') }}</label>
      <div v-if="loading" class="loading-state" style="padding: var(--space-sm) 0">
        <Spinner />
        {{ t('detail.drupal.masqLoadingUsers') }}
      </div>
      <div v-else-if="users.length === 0" class="text-muted" style="padding: var(--space-sm) 0">
        {{ t('detail.drupal.masqNoUsers') }}
      </div>
      <div v-else class="flu-table-wrap" style="max-height: 260px; overflow-y: auto">
        <table class="flu-table">
          <thead>
            <tr>
              <th>{{ t('detail.drupal.masqColUid') }}</th>
              <th>{{ t('detail.drupal.masqColName') }}</th>
              <th>{{ t('detail.drupal.masqColEmail') }}</th>
              <th />
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="user in users"
              :key="user.uid"
              class="addon-pick-row masq-row"
              @click="!running && doMasquerade(user.uid)"
            >
              <td>{{ user.uid }}</td>
              <td>{{ user.name }}</td>
              <td class="masq-email">{{ user.mail }}</td>
              <td>
                <button
                  type="button"
                  class="flu-btn flu-btn-xs flu-btn-accent"
                  :disabled="running"
                  @click.stop="doMasquerade(user.uid)"
                >
                  <template v-if="running">
                    <Spinner />
                  </template>
                  <template v-else>
                    {{ t('general.login') }}
                  </template>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="flu-field">
      <label class="flu-label">{{ t('detail.drupal.masqEnterUid') }}</label>
      <div class="masq-input-row">
        <input
          v-model="uid"
          class="flu-input"
          :placeholder="t('detail.drupal.masqUidPlaceholder')"
          :disabled="running"
          style="flex: 1"
          @keydown.enter="uid.trim() ? doMasquerade(uid) : undefined"
        >
        <button
          type="button"
          class="flu-btn flu-btn-accent"
          :disabled="!uid.trim() || running"
          @click="doMasquerade(uid)"
        >
          <template v-if="running">
            <Spinner />
          </template>
          <template v-else>
            {{ t('general.go') }}
          </template>
        </button>
      </div>
    </div>
  </Modal>
</template>

<style scoped>
.masq-row {
  cursor: pointer;
}

.masq-email {
  color: var(--text-secondary);
  font-size: 12px;
}

.masq-input-row {
  display: flex;
  gap: var(--space-xs);
}
</style>

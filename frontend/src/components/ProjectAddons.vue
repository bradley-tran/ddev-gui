<script setup lang="ts">
import { LayersPlusIcon } from '@lucide/vue'
import { computed, ref, watch } from 'vue'
import Modal from '@/components/Modal.vue'
import ConfirmDeleteModal from '@/modals/ConfirmDeleteModal.vue'
import Spinner from '@/components/Spinner.vue'
import { filterAddons } from '@/lib/addonFilter'
import { useTranslation } from '@/lib/i18n'
import type { DdevAddon } from '@/lib/types'
import { coerceToBool, openUrl } from '@/lib/utils'
import { DdevService } from '@/lib/wails'
import { useAppStore } from '@/stores/app'

interface NormalizedAddon {
  name: string
  version: string
  repo: string
  installed: string
}

interface AvailableAddon {
  repo: string
  description: string
}

const props = defineProps<{
  projectName: string
}>()

const appStore = useAppStore()
const { t } = useTranslation()

const addons = ref<DdevAddon[]>([])
const available = ref<DdevAddon[]>([])
const loadingAddons = ref(true)
const loadingAvailable = ref(false)
const showPicker = ref(false)
const search = ref('')
const installing = ref('')
const removeCandidate = ref<string | null>(null)
const removing = ref(false)

const normalizedAddons = computed<NormalizedAddon[]>(() =>
  addons.value
    .map((item) => ({
      name: String(item.Name ?? item.name ?? '').trim(),
      version: String(item.Version ?? item.version ?? '').trim(),
      repo: String(
        item.Repository ?? item.repository ?? item.full_name ?? item.FullName ?? item.repo ?? item.source ?? '',
      ).trim(),
      installed: String(
        item.InstalledDate ?? item.installed_date ?? item.installedDate ?? item.installed ?? item.date ?? '',
      ).trim(),
    }))
    .filter((item) => item.name),
)

const filteredAvailable = computed<AvailableAddon[]>(() =>
  filterAddons(available.value, search.value)
    .map((item) => normalizeAvailableAddon(item))
    .filter((item): item is AvailableAddon => Boolean(item)),
)

const removeMessage = computed(() =>
  removeCandidate.value ? t('detail.addons.removeConfirm', { repo: removeCandidate.value }) : '',
)

watch(
  () => props.projectName,
  async (projectName) => {
    if (!projectName) {
      addons.value = []
      available.value = []
      return
    }

    await loadAddons(projectName)
  },
  { immediate: true },
)

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}

function parseAddonItems(json: string, fallbackKeys: string[] = []): DdevAddon[] {
  const data = JSON.parse(json) as unknown
  if (Array.isArray(data)) return data as DdevAddon[]

  if (isRecord(data)) {
    if (Array.isArray(data.raw)) return data.raw as DdevAddon[]
    if (Array.isArray(data.items)) return data.items as DdevAddon[]

    for (const key of fallbackKeys) {
      const candidate = data[key]
      if (Array.isArray(candidate)) {
        return candidate as DdevAddon[]
      }
    }
  }

  return []
}

function normalizeAvailableAddon(item: DdevAddon): AvailableAddon | null {
  const rawRepo = String(item.Repository ?? item.repository ?? item.repo ?? item.full_name ?? item.FullName ?? '').trim()
  const user = String(item.user ?? item.User ?? '').trim()
  const repo = rawRepo && !rawRepo.includes('/') && user ? `${user}/${rawRepo}` : rawRepo
  const description = String(item.Description ?? item.description ?? '').trim()

  if (!repo) return null

  return {
    repo,
    description,
  }
}

function closePicker() {
  showPicker.value = false
}

function openAddonLink(repo: string) {
  openUrl(`https://addons.ddev.com/addons/${repo}`, coerceToBool(appStore.config.openLinksInBrowser))
}

async function loadAddons(projectName = props.projectName) {
  if (!projectName) return

  loadingAddons.value = true
  try {
    const json = await DdevService.addonsJSON(projectName)
    addons.value = parseAddonItems(json, ['installed'])
  } catch (error) {
    addons.value = []
    const message = error instanceof Error ? error.message : String(error)
    appStore.appLog(`Failed to load add-ons: ${message}`, 'error')
  } finally {
    loadingAddons.value = false
  }
}

async function openPicker() {
  showPicker.value = true
  search.value = ''
  loadingAvailable.value = true

  try {
    const json = await DdevService.addonsAvailableJSON(props.projectName)
    available.value = parseAddonItems(json)
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error)
    appStore.appLog(`Failed to load available add-ons: ${message}`, 'error')
    available.value = []
  } finally {
    loadingAvailable.value = false
  }
}

async function handleInstall(repo: string) {
  installing.value = repo
  appStore.appLog(`Installing add-on ${repo} for ${props.projectName}...`, 'info')

  try {
    await DdevService.addonInstall(props.projectName, repo)
    appStore.appLog(`Add-on ${repo} installed`, 'success')
    appStore.showToast('Add-on installed', 'success')
    await loadAddons()
    closePicker()
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error)
    appStore.appLog(`Failed to install ${repo}: ${message}`, 'error')
    appStore.showToast(`Failed to install ${repo}`, 'error')
  } finally {
    installing.value = ''
  }
}

async function handleRemove(repo: string) {
  removeCandidate.value = repo
}

function closeRemoveModal() {
  if (removing.value) return
  removeCandidate.value = null
}

async function handleRemoveConfirm() {
  const repo = removeCandidate.value
  if (!repo || removing.value) return

  removing.value = true

  appStore.appLog(`Removing add-on ${repo} from ${props.projectName}...`, 'info')

  try {
    await DdevService.addonRemove(props.projectName, repo)
    appStore.appLog(`Add-on ${repo} removed`, 'success')
    appStore.showToast(`Add-on ${repo} removed`, 'success')
    await loadAddons()
    removeCandidate.value = null
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error)
    appStore.appLog(`Failed to remove ${repo}: ${message}`, 'error')
    appStore.showToast(`Failed to remove ${repo}`, 'error')
  } finally {
    removing.value = false
  }
}
</script>

<template>
  <section v-if="loadingAddons" class="detail-section" data-testid="project-addons-loading">
    <div class="detail-section-title">{{ t('detail.tabAddons') }}</div>
    <div class="detail-section-body detail-loading-row">
      <Spinner />
      <span class="loading-text">{{ t('general.loading') }}</span>
    </div>
  </section>

  <div v-else id="detailAddons" class="detail-section" data-testid="project-addons">
    <div class="flu-table-wrap">
      <table class="flu-table">
        <thead>
          <tr>
            <th>{{ t('detail.addons.colName') }}</th>
            <th>{{ t('detail.addons.colVersion') }}</th>
            <th>{{ t('detail.addons.colRepo') }}</th>
            <th>
              <button
                type="button"
                class="flu-btn flu-btn-sm flu-btn-ghost"
                :title="t('detail.addons.installTitle')"
                data-testid="project-addons-open-picker"
                @click="openPicker"
              >
                <LayersPlusIcon :size="14" :stroke-width="2" />
                {{ t('general.add') }}
              </button>
            </th>
          </tr>
        </thead>
        <div v-if="normalizedAddons.length === 0" class="empty-table">
          <div class="text-muted">
            {{ t('detail.addons.none') }}
          </div>
        </div>
        <tbody v-else>
          <tr v-for="item in normalizedAddons" :key="`${item.name}-${item.repo}`">
            <td>{{ item.name }}</td>
            <td>{{ item.version }}</td>
            <td>
              <button
                v-if="item.repo"
                type="button"
                class="proj-link addon-link-button"
                @click="openAddonLink(item.repo)"
              >
                {{ item.repo }}
              </button>
            </td>
            <td>
              <button
                v-if="item.repo"
                type="button"
                class="flu-btn flu-btn-xs flu-btn-danger"
                @click="handleRemove(item.repo)"
              >
                {{ t('general.remove') }}
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>

  <Modal v-if="showPicker" :title="t('detail.addons.installTitle')" wide @close="closePicker">
    <div class="flu-field">
      <input
        v-model="search"
        class="flu-input"
        :placeholder="t('detail.addons.searchPlaceholder')"
        data-testid="project-addons-search"
        autofocus
      />
    </div>

    <div v-if="loadingAvailable" class="loading-state detail-loading-row">
      <Spinner />
      <span>{{ t('detail.addons.loadingAvailable') }}</span>
    </div>

    <div v-else-if="filteredAvailable.length === 0" class="text-muted">
      {{ search ? t('detail.addons.noMatching') : t('detail.addons.noAvailable') }}
    </div>

    <div v-else class="addon-table-wrap">
      <table class="flu-table addon-pick-table">
        <thead>
          <tr>
            <th>{{ t('detail.addons.colRepo') }}</th>
            <th>{{ t('detail.addons.colDescription') }}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in filteredAvailable" :key="item.repo" class="addon-pick-row">
            <td class="addon-pick-name">{{ item.repo }}</td>
            <td class="addon-pick-desc">{{ item.description }}</td>
            <td>
              <button
                type="button"
                class="flu-btn flu-btn-xs flu-btn-accent"
                :disabled="Boolean(installing)"
                @click="handleInstall(item.repo)"
              >
                <Spinner v-if="installing === item.repo" />
                <template v-else>{{ t('general.install') }}</template>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <template #footer>
      <button type="button" class="flu-btn flu-btn-ghost" @click="closePicker">
        {{ t('general.close') }}
      </button>
    </template>
  </Modal>

  <ConfirmDeleteModal
    v-if="removeCandidate"
    :title="t('general.remove')"
    :message="removeMessage"
    :confirm-text="t('general.remove')"
    :pending="removing"
    @close="closeRemoveModal"
    @confirm="handleRemoveConfirm"
  />
</template>

<style scoped>
.loading-text {
  margin-left: 0.5rem;
}

.empty-table {
  padding: 1rem;
}

.addon-link-button {
  border: 0;
  padding: 0;
  background: transparent;
  cursor: pointer;
}
</style>
<script setup lang="ts">
import { EditIcon, SlidersHorizontalIcon } from '@lucide/vue'
import Spinner from '@/components/Spinner.vue'
import { useTranslation } from '@/lib/i18n'
import type { DdevService } from '@/lib/types'

interface OverviewItem {
  label: string
  value: string
  isStatus?: boolean
}

defineProps<{
  loading: boolean
  hasProject: boolean
  overviewItems: OverviewItem[]
  services: Array<[string, DdevService]>
}>()

const emit = defineEmits<{
  modify: []
  'config-services': []
  'open-url': [url: string]
}>()

const { t } = useTranslation()

function statusClass(status: string): string {
  return /run/i.test(status) ? 'status-badge running' : 'status-badge stopped'
}
</script>

<template>
  <div id="detailInfo">
    <section v-if="loading" class="detail-section">
      <div class="detail-section-title">{{ t('detail.tabOverview') }}</div>
      <div class="detail-section-body detail-loading-row">
        <Spinner />
        <span>{{ t('general.loading') }}</span>
      </div>
    </section>

    <template v-else-if="hasProject">
      <section class="detail-section">
        <div class="detail-section-title">
          <span>{{ t('detail.overview.title') }}</span>
          <button type="button" class="flu-btn flu-btn-sm flu-btn-ghost" @click="emit('modify')">
            <EditIcon :size="12" :stroke-width="2" />
            {{ t('detail.overview.modify') }}
          </button>
        </div>
        <div class="detail-section-body">
          <div class="detail-kv">
            <div v-for="item in overviewItems" :key="item.label" class="detail-kv-item">
              <span class="kv-label">{{ item.label }}</span>
              <span class="kv-value">
                <span v-if="item.isStatus" :class="statusClass(item.value)">
                  {{ item.value }}
                </span>
                <template v-else>
                  {{ item.value }}
                </template>
              </span>
            </div>
          </div>
        </div>
      </section>

      <section class="detail-section">
        <div class="detail-section-title">
          <span>{{ t('detail.services.title') }}</span>
          <button type="button" class="flu-btn flu-btn-sm flu-btn-ghost" @click="emit('config-services')">
            <SlidersHorizontalIcon :size="12" :stroke-width="2" />
            {{ t('detail.services.config') }}
          </button>
        </div>
        <div class="detail-section-body">
          <div v-if="services.length === 0" class="text-muted">
            {{ t('detail.services.none') }}
          </div>
          <div v-else class="detail-kv">
            <div v-for="[serviceName, service] in services" :key="serviceName" class="detail-kv-item">
              <span class="kv-label">{{ serviceName }}</span>
              <span class="kv-value detail-service-links">
                <span :class="statusClass(String(service.status || 'n/a'))">
                  {{ String(service.status || 'n/a') }}
                </span>
                <button
                  v-if="service.https_url || service.host_https_url"
                  type="button"
                  class="detail-link-button"
                  @click="emit('open-url', String(service.https_url || service.host_https_url || ''))"
                >
                  https
                </button>
                <button
                  v-if="service.http_url || service.host_http_url"
                  type="button"
                  class="detail-link-button"
                  @click="emit('open-url', String(service.http_url || service.host_http_url || ''))"
                >
                  http
                </button>
                <span v-if="service.exposed_ports || service.host_ports" class="text-muted">
                  {{ t('detail.services.ports') }}:
                  {{ String(service.exposed_ports || service.host_ports || '') }}
                </span>
              </span>
            </div>
          </div>
        </div>
      </section>
    </template>

    <section v-else class="detail-section">
      <div class="detail-section-title">{{ t('detail.overview.title') }}</div>
      <div class="detail-section-body text-muted">
        Project data is not available yet. Refresh the list or reopen the project.
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

.detail-service-links {
  display: inline-flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
}

.detail-link-button {
  border: 0;
  padding: 0;
  background: transparent;
  color: var(--accent-primary);
  cursor: pointer;
  text-decoration: underline;
}
</style>

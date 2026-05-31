<script setup lang="ts">
import Spinner from '@/components/Spinner.vue'
import { CameraIcon } from '@lucide/vue'
import { useTranslation } from '@/lib/i18n'

defineProps<{
  loading: boolean
  snapshots: string[]
}>()

const emit = defineEmits<{
  create: []
  restore: [snapshotName: string]
  delete: [snapshotName: string]
}>()

const { t } = useTranslation()
</script>

<template>
  <div id="detailSnapshots">
    <section class="detail-section">
      <div class="detail-section-title">
        <span>{{ t('detail.snapshots.title') }}</span>
        <button
          type="button"
          class="flu-btn flu-btn-sm flu-btn-ghost snapshot-create-btn"
          @click="emit('create')"
        >
          <CameraIcon :size="14" :stroke-width="2" />
          {{ t('detail.snapshots.create') }}
        </button>
      </div>
      <div class="detail-section-body">
        <div v-if="loading" class="detail-loading-row">
          <Spinner />
          <span>{{ t('general.loading') }}</span>
        </div>
        <div v-else-if="snapshots.length === 0" class="text-muted">
          {{ t('detail.snapshots.none') }}
        </div>
        <div v-else class="detail-snapshot-list">
          <div v-for="snapshotName in snapshots" :key="snapshotName" class="detail-snapshot-item">
            <span class="detail-snapshot-name">{{ snapshotName }}</span>
            <div class="detail-snapshot-actions">
              <button
                type="button"
                class="flu-btn flu-btn-xs flu-btn-accent snapshot-restore-btn"
                @click="emit('restore', snapshotName)"
              >
                {{ t('detail.snapshots.restore') }}
              </button>
              <button
                type="button"
                class="flu-btn flu-btn-xs flu-btn-danger snapshot-delete-btn"
                @click="emit('delete', snapshotName)"
              >
                {{ t('detail.snapshots.delete') }}
              </button>
            </div>
          </div>
        </div>
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

.detail-snapshot-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 0.5rem;
  padding: 0.25rem 0 0.5rem;
}

.detail-snapshot-item:not(:last-child) {
  border-bottom: 1px solid var(--border-subtle);
}

.detail-snapshot-name {
  font-weight: 500;
}

.detail-snapshot-actions {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  margin: 0;
}

.detail-snapshot-actions .flu-btn {
  margin: 0;
  height: 22px;
  padding-top: 0;
  padding-bottom: 0;
  line-height: 1;
}

.snapshot-restore-btn,
.snapshot-delete-btn {
  align-self: center;
  transform: translateY(0);
}
</style>

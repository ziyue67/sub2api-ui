<template>
  <div class="scheme3-admin-monitor-filterbar">
    <div class="scheme3-admin-monitor-filter-fields">
      <label class="scheme3-admin-monitor-search">
        <span class="sr-only">{{ t('admin.channelMonitor.searchPlaceholder') }}</span>
        <Icon
          name="search"
          size="md"
          class="scheme3-admin-monitor-search-icon"
        />
        <input
          v-model="search"
          type="text"
          :placeholder="t('admin.channelMonitor.searchPlaceholder')"
          class="scheme3-admin-monitor-search-input"
          @input="$emit('search-input')"
        />
      </label>

      <Select
        v-model="provider"
        scheme3
        :options="providerFilterOptions"
        :placeholder="t('admin.channelMonitor.allProviders')"
        :aria-label="t('admin.channelMonitor.allProviders')"
        class="scheme3-admin-monitor-filter-select"
        @change="$emit('reload')"
      />

      <Select
        v-model="enabled"
        scheme3
        :options="enabledFilterOptions"
        :placeholder="t('admin.channelMonitor.enabledFilter')"
        :aria-label="t('admin.channelMonitor.enabledFilter')"
        class="scheme3-admin-monitor-filter-select"
        @change="$emit('reload')"
      />
    </div>

    <div class="scheme3-admin-monitor-filter-actions">
      <button
        type="button"
        class="scheme3-admin-monitor-filter-action scheme3-admin-monitor-filter-icon"
        :disabled="loading"
        :title="t('common.refresh')"
        :aria-label="t('common.refresh')"
        @click="$emit('reload')"
      >
        <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
      </button>
      <button
        type="button"
        class="scheme3-admin-monitor-filter-action"
        @click="$emit('manage-templates')"
        :aria-label="t('admin.channelMonitor.template.manageButton')"
      >
        <Icon name="cog" size="sm" />
        <span>{{ t('admin.channelMonitor.template.manageButton') }}</span>
      </button>
      <button
        type="button"
        class="scheme3-admin-monitor-filter-action is-primary"
        @click="$emit('create')"
      >
        <Icon name="plus" size="sm" />
        <span>{{ t('admin.channelMonitor.createButton') }}</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Provider } from '@/api/admin/channelMonitor'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import {
  PROVIDER_OPENAI,
  PROVIDER_ANTHROPIC,
  PROVIDER_GEMINI,
  PROVIDER_GROK,
} from '@/constants/channelMonitor'

defineProps<{
  loading: boolean
}>()

defineEmits<{
  (e: 'reload'): void
  (e: 'create'): void
  (e: 'manage-templates'): void
  (e: 'search-input'): void
}>()

const search = defineModel<string>('search', { required: true })
const provider = defineModel<Provider | ''>('provider', { required: true })
const enabled = defineModel<'' | 'true' | 'false'>('enabled', { required: true })

const { t } = useI18n()

const providerFilterOptions = computed(() => [
  { value: '', label: t('admin.channelMonitor.allProviders') },
  { value: PROVIDER_OPENAI, label: t('monitorCommon.providers.openai') },
  { value: PROVIDER_ANTHROPIC, label: t('monitorCommon.providers.anthropic') },
  { value: PROVIDER_GEMINI, label: t('monitorCommon.providers.gemini') },
  { value: PROVIDER_GROK, label: t('monitorCommon.providers.grok') },
])

const enabledFilterOptions = computed(() => [
  { value: '', label: t('admin.channelMonitor.allStatus') },
  { value: 'true', label: t('admin.channelMonitor.onlyEnabled') },
  { value: 'false', label: t('admin.channelMonitor.onlyDisabled') },
])
</script>

<style scoped>
.scheme3-admin-monitor-filterbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .8rem;
}
.scheme3-admin-monitor-filter-fields,
.scheme3-admin-monitor-filter-actions {
  display: flex;
  min-width: 0;
  flex-wrap: wrap;
  align-items: center;
  gap: .55rem;
}
.scheme3-admin-monitor-filter-fields { flex: 1; }
.scheme3-admin-monitor-search {
  position: relative;
  width: min(100%, 16rem);
}
.scheme3-admin-monitor-search-icon {
  position: absolute;
  left: .72rem;
  top: 50%;
  z-index: 1;
  color: var(--admin-muted, #777266);
  transform: translateY(-50%);
  pointer-events: none;
}
.scheme3-admin-monitor-search-input {
  width: 100%;
  min-height: 2.35rem;
  border: 1px solid var(--admin-line, #d8d2c3);
  border-radius: 6px;
  padding: .5rem .72rem .5rem 2.25rem;
  background: var(--admin-surface, #fffefa);
  color: var(--admin-ink, #27251f);
  font-size: .72rem;
  outline: none;
  transition: border-color 150ms ease, box-shadow 150ms ease;
}
.scheme3-admin-monitor-search-input::placeholder { color: var(--admin-muted, #777266); }
.scheme3-admin-monitor-search-input:focus { border-color: var(--admin-accent, #1e5c42); box-shadow: 0 0 0 2px rgba(30, 92, 66, .13); }
.scheme3-admin-monitor-filter-select { width: 10.5rem; }
.scheme3-admin-monitor-filter-select :deep(.scheme3-select-trigger) {
  min-height: 2.35rem;
  border-color: var(--admin-line, #d8d2c3);
  border-radius: 6px;
  padding: .5rem .72rem;
  background: var(--admin-surface, #fffefa);
  color: var(--admin-ink, #27251f);
  box-shadow: none;
}
.scheme3-admin-monitor-filter-select :deep(.scheme3-select-trigger:hover),
.scheme3-admin-monitor-filter-select :deep(.scheme3-select-trigger-open) { border-color: var(--admin-accent, #1e5c42); box-shadow: 0 0 0 2px rgba(30, 92, 66, .12); }
.scheme3-admin-monitor-filter-action {
  display: inline-flex;
  min-height: 2.35rem;
  align-items: center;
  justify-content: center;
  gap: .4rem;
  border: 1px solid var(--admin-line, #d8d2c3);
  border-radius: 6px;
  padding: .48rem .68rem;
  background: var(--admin-surface, #fffefa);
  color: var(--admin-muted, #777266);
  font-size: .66rem;
  font-weight: 800;
  transition: border-color 150ms ease, background-color 150ms ease, color 150ms ease, transform 120ms ease;
}
.scheme3-admin-monitor-filter-action:hover:not(:disabled) { border-color: var(--admin-accent, #1e5c42); color: var(--admin-accent, #1e5c42); }
.scheme3-admin-monitor-filter-action:active:not(:disabled) { transform: scale(.98); }
.scheme3-admin-monitor-filter-action:focus-visible { outline: 2px solid rgba(30, 92, 66, .28); outline-offset: 2px; }
.scheme3-admin-monitor-filter-action:disabled { cursor: not-allowed; opacity: .45; }
.scheme3-admin-monitor-filter-action.is-primary { border-color: var(--admin-accent, #1e5c42); background: var(--admin-accent, #1e5c42); color: #fffefa; }
.scheme3-admin-monitor-filter-action.is-primary:hover:not(:disabled) { background: #174a35; color: #fffefa; }
.scheme3-admin-monitor-filter-icon { width: 2.35rem; padding: 0; }

@media (max-width: 900px) {
  .scheme3-admin-monitor-filterbar { align-items: stretch; flex-direction: column; }
  .scheme3-admin-monitor-filter-fields,
  .scheme3-admin-monitor-filter-actions { width: 100%; }
  .scheme3-admin-monitor-filter-actions { justify-content: flex-end; }
}
@media (max-width: 560px) {
  .scheme3-admin-monitor-search,
  .scheme3-admin-monitor-filter-select { width: 100%; }
  .scheme3-admin-monitor-filter-actions { display: grid; grid-template-columns: 2.35rem minmax(0, 1fr) minmax(0, 1fr); }
}
</style>

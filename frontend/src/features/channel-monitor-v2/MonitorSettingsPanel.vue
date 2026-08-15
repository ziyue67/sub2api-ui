<template>
  <section class="scheme3-v2-settings mx-auto w-full max-w-6xl space-y-5 px-1 py-2 sm:px-2">
    <header
      class="scheme3-v2-settings-header flex flex-wrap items-center justify-between gap-3 p-5 sm:p-6"
    >
      <div class="min-w-0">
        <h2 class="scheme3-v2-settings-title flex items-center gap-2">
          <span class="scheme3-v2-settings-mark inline-flex h-7 w-7 items-center justify-center">
            <Icon name="chart" size="sm" />
          </span>
          {{ t('channelMonitorV2.settings.title') }}
        </h2>
        <p class="scheme3-v2-settings-description mt-1.5 text-xs">
          {{ t('channelMonitorV2.settings.description') }}
        </p>
      </div>
      <button
        type="button"
        class="scheme3-v2-settings-save inline-flex items-center gap-1.5"
        :disabled="saving || !dirty"
        @click="save"
      >
        <Icon name="check" size="sm" />
        {{ t('channelMonitorV2.settings.save') }}
      </button>
    </header>

    <div
      v-if="!systemModeV2"
      class="scheme3-v2-settings-banner px-4 py-3 text-sm"
      role="status"
    >
      {{
        t('channelMonitorV2.settings.modeBanner', {
          mode: systemModeLabel,
          modeV2: t('channelMonitorV2.settings.modeV2'),
        })
      }}
      <router-link class="scheme3-v2-settings-link ml-1 font-medium underline" to="/admin/settings">{{ t('admin.settings.tabs.features') }}</router-link>
    </div>

    <div
      v-if="loading"
      class="scheme3-v2-settings-loading flex min-h-[200px] items-center justify-center text-sm"
    >
      <span class="animate-pulse">{{ t('channelMonitorV2.settings.loading') }}</span>
    </div>

    <template v-else-if="draft">
      <div class="scheme3-v2-settings-panel scheme3-v2-settings-basics">
        <div class="scheme3-v2-settings-row flex flex-wrap items-center justify-between gap-4 px-5 py-4">
          <div>
            <strong class="scheme3-v2-settings-label text-sm font-semibold">{{ t('channelMonitorV2.settings.enableTitle') }}</strong>
            <p class="scheme3-v2-settings-hint mt-0.5 text-xs">
              {{ t('channelMonitorV2.settings.enableHint') }}
            </p>
          </div>
          <Scheme3V2Toggle v-model="draft.enabled" />
        </div>
        <div class="scheme3-v2-settings-row flex flex-wrap items-center justify-between gap-4 px-5 py-4">
          <div>
            <strong class="scheme3-v2-settings-label text-sm font-semibold">{{ t('channelMonitorV2.settings.refreshTitle') }}</strong>
            <p class="scheme3-v2-settings-hint mt-0.5 text-xs">{{ t('channelMonitorV2.settings.refreshHint') }}</p>
          </div>
          <div class="scheme3-v2-settings-segments inline-flex w-auto" role="group" :aria-label="t('channelMonitorV2.settings.refreshAria')">
            <button
              type="button"
              class="scheme3-v2-settings-segment"
              :class="draft.refresh_interval_seconds === 60 ? 'is-active' : ''"
              @click="draft.refresh_interval_seconds = 60"
            >
              1 min
            </button>
            <button
              type="button"
              class="scheme3-v2-settings-segment"
              :class="draft.refresh_interval_seconds === 300 ? 'is-active' : ''"
              @click="draft.refresh_interval_seconds = 300"
            >
              5 min
            </button>
          </div>
        </div>
      </div>

      <div class="scheme3-v2-settings-panel overflow-hidden">
        <div class="scheme3-v2-settings-panel-header !py-3">
          <h3 class="scheme3-v2-settings-label text-sm font-semibold">{{ t('channelMonitorV2.settings.platformsTitle') }}</h3>
          <p class="scheme3-v2-settings-hint mt-0.5 text-xs">
            {{ t('channelMonitorV2.settings.platformsHint') }}
          </p>
        </div>
        <div class="scheme3-v2-settings-list">
          <div
            v-for="platform in draft.platforms"
            :key="platform.platform"
            class="scheme3-v2-settings-row grid grid-cols-1 items-center gap-3 px-5 py-3 sm:grid-cols-[auto_7rem_minmax(0,1fr)_auto]"
          >
            <Scheme3V2Toggle v-model="platform.enabled" />
            <strong class="scheme3-v2-settings-label text-sm font-medium">{{ platformLabel(platform.platform) }}</strong>
            <input
              class="scheme3-v2-settings-input"
              :value="platform.models.join(', ')"
              type="text"
              :placeholder="t('channelMonitorV2.settings.modelsPlaceholder')"
              @change="setModels(platform, $event)"
            />
            <span
              class="scheme3-v2-settings-badge justify-self-start sm:justify-self-end"
              :class="platform.models.length ? 'is-muted' : 'is-accent'"
            >
              {{ platform.models.length ? t('channelMonitorV2.settings.badgeOther') : t('channelMonitorV2.settings.badgeAllModels') }}
            </span>
          </div>
        </div>
      </div>

      <div class="scheme3-v2-settings-panel overflow-hidden">
        <div class="scheme3-v2-settings-panel-header flex flex-wrap items-center justify-between gap-2 !py-3">
          <div>
            <h3 class="scheme3-v2-settings-label text-sm font-semibold">{{ t('channelMonitorV2.settings.groupsTitle') }}</h3>
            <p class="scheme3-v2-settings-hint mt-0.5 text-xs">
              {{
                draft.group_ids.length
                  ? t('channelMonitorV2.settings.groupsSelected', { count: draft.group_ids.length })
                  : t('channelMonitorV2.settings.groupsAll')
              }}
            </p>
          </div>
          <button
            v-if="draft.group_ids.length"
            type="button"
            class="scheme3-v2-settings-inline-action"
            @click="draft.group_ids = []"
          >
            {{ t('channelMonitorV2.settings.groupsAll') }}
          </button>
        </div>
        <div class="max-h-[min(40vh,280px)] overflow-y-auto px-3 py-2 sm:px-4">
          <div class="grid grid-cols-1 gap-1 sm:grid-cols-2">
            <label
              v-for="group in groups"
              :key="group.id"
              class="scheme3-v2-settings-choice flex cursor-pointer items-center gap-3 px-3 py-2.5 text-sm transition"
            >
              <input
                type="checkbox"
                class="scheme3-v2-settings-checkbox"
                :checked="draft.group_ids.includes(group.id)"
                @change="toggleGroup(group.id)"
              />
              <span class="scheme3-v2-settings-choice-label min-w-0 flex-1 truncate font-medium">{{ group.name }}</span>
              <small class="scheme3-v2-settings-meta shrink-0 text-xs">{{ platformLabel(group.platform) }} · #{{ group.id }}</small>
            </label>
          </div>
          <p v-if="groups.length === 0" class="scheme3-v2-settings-empty py-8 text-sm">{{ t('channelMonitorV2.settings.groupsEmpty') }}</p>
        </div>
      </div>

      <div class="scheme3-v2-settings-panel overflow-hidden">
        <div class="scheme3-v2-settings-panel-header !py-3">
          <h3 class="scheme3-v2-settings-label text-sm font-semibold">{{ t('channelMonitorV2.settings.errorsTitle') }}</h3>
          <p class="scheme3-v2-settings-hint mt-0.5 text-xs">
            {{ t('channelMonitorV2.settings.errorsHint') }}
          </p>
        </div>
        <div class="max-h-[min(40vh,320px)] overflow-y-auto px-3 py-2 sm:px-4">
          <div class="grid grid-cols-1 gap-1 sm:grid-cols-2">
            <label
              v-for="category in errorCategories"
              :key="category"
              class="scheme3-v2-settings-choice flex cursor-pointer items-center gap-3 px-3 py-2.5 text-sm transition"
            >
              <input
                type="checkbox"
                class="scheme3-v2-settings-checkbox"
                :checked="isCategoryIgnored(category)"
                @change="toggleIgnoredCategory(category)"
              />
              <span class="scheme3-v2-settings-choice-label min-w-0 flex-1 truncate font-medium">
                {{ categoryLabel(category) }}
              </span>
              <small class="scheme3-v2-settings-meta shrink-0 font-mono text-[10px]">{{ category }}</small>
            </label>
          </div>
        </div>
        <div class="scheme3-v2-settings-footer px-5 py-3 text-xs">
          {{
            t('channelMonitorV2.settings.ignoredSummary', {
              ignored: draft.ignored_error_categories?.length || 0,
              counted: countedErrorCategoryCount,
            })
          }}
        </div>
      </div>

      <div class="scheme3-v2-settings-panel overflow-hidden">
        <div class="scheme3-v2-settings-panel-header !py-3">
          <h3 class="scheme3-v2-settings-label text-sm font-semibold">{{ t('channelMonitorV2.settings.healthTitle') }}</h3>
          <p class="scheme3-v2-settings-hint mt-0.5 text-xs">
            {{ t('channelMonitorV2.settings.healthHint') }}
          </p>
        </div>
        <div class="scheme3-v2-settings-fields grid grid-cols-1 gap-4 px-5 py-4 sm:grid-cols-2 lg:grid-cols-4">
          <label class="block">
            <span class="scheme3-v2-settings-field-label">{{ t('channelMonitorV2.settings.fields.minimumSample') }}</span>
            <input v-model.number="draft.health_thresholds.minimum_sample" class="scheme3-v2-settings-input" type="number" min="1" max="10000" />
          </label>
          <label class="block">
            <span class="scheme3-v2-settings-field-label">{{ t('channelMonitorV2.settings.fields.warningError') }}</span>
            <input v-model.number="warningErrorPercent" class="scheme3-v2-settings-input" type="number" min="0" max="100" step="0.1" />
          </label>
          <label class="block">
            <span class="scheme3-v2-settings-field-label">{{ t('channelMonitorV2.settings.fields.criticalError') }}</span>
            <input v-model.number="criticalErrorPercent" class="scheme3-v2-settings-input" type="number" min="0" max="100" step="0.1" />
          </label>
          <label class="block">
            <span class="scheme3-v2-settings-field-label">{{ t('channelMonitorV2.settings.fields.targetTtft') }}</span>
            <input v-model.number="draft.health_thresholds.target_ttft_ms" class="scheme3-v2-settings-input" type="number" min="1" step="100" />
          </label>
          <label class="block">
            <span class="scheme3-v2-settings-field-label">{{ t('channelMonitorV2.settings.fields.warningTtft') }}</span>
            <input v-model.number="draft.health_thresholds.warning_ttft_ms" class="scheme3-v2-settings-input" type="number" min="1" step="100" />
          </label>
          <label class="block">
            <span class="scheme3-v2-settings-field-label">{{ t('channelMonitorV2.settings.fields.criticalTtft') }}</span>
            <input v-model.number="draft.health_thresholds.critical_ttft_ms" class="scheme3-v2-settings-input" type="number" min="1" step="100" />
          </label>
          <label class="block">
            <span class="scheme3-v2-settings-field-label">{{ t('channelMonitorV2.settings.fields.warningCache') }}</span>
            <input v-model.number="warningCachePercent" class="scheme3-v2-settings-input" type="number" min="0" max="100" step="0.1" />
          </label>
          <label class="block">
            <span class="scheme3-v2-settings-field-label">{{ t('channelMonitorV2.settings.fields.criticalCache') }}</span>
            <input v-model.number="criticalCachePercent" class="scheme3-v2-settings-input" type="number" min="0" max="100" step="0.1" />
          </label>
        </div>
      </div>

      <div class="scheme3-v2-settings-notes space-y-2">
        <div class="scheme3-v2-settings-note scheme3-v2-settings-note-accent px-4 py-3 text-sm">
          <template v-if="namedModelCount === 0">
            {{ t('channelMonitorV2.settings.namedModelsEmpty') }}
          </template>
          <template v-else>
            {{ t('channelMonitorV2.settings.namedModelsCount', { count: namedModelCount }) }}
          </template>
        </div>
        <div class="scheme3-v2-settings-note scheme3-v2-settings-note-neutral px-4 py-3 text-xs">
          <p class="scheme3-v2-settings-label font-medium">{{ t('channelMonitorV2.settings.userContractTitle') }}</p>
          <ul class="mt-1.5 list-disc space-y-0.5 pl-4">
            <li>{{ t('channelMonitorV2.settings.userContract.health') }}</li>
            <li>{{ t('channelMonitorV2.settings.userContract.trend') }}</li>
            <li>{{ t('channelMonitorV2.settings.userContract.latency') }}</li>
            <li>{{ t('channelMonitorV2.settings.userContract.models') }}</li>
          </ul>
        </div>
      </div>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Scheme3V2Toggle from './Scheme3V2Toggle.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { getChannelMonitorMode, isChannelMonitorV2Mode } from '@/utils/featureFlags'
import {
  getConfig,
  updateConfig,
  MONITOR_ERROR_CATEGORIES,
  type MonitorConfig,
} from '@/api/channelMonitorV2'
import { adminAPI } from '@/api/admin'
import type { AdminGroup } from '@/types'

const { t, te } = useI18n()
const appStore = useAppStore()
const loading = ref(true)
const saving = ref(false)
const draft = ref<MonitorConfig | null>(null)
const original = ref('')
const groups = ref<AdminGroup[]>([])

const dirty = computed(() => (draft.value ? JSON.stringify(draft.value) !== original.value : false))
const namedModelCount = computed(
  () => draft.value?.platforms.filter((p) => p.enabled).reduce((sum, p) => sum + p.models.length, 0) || 0
)
const errorCategories = MONITOR_ERROR_CATEGORIES
const countedErrorCategoryCount = computed(
  () => errorCategories.length - (draft.value?.ignored_error_categories?.length || 0)
)
/** System settings mode must be v2 for aggregation to run; config remains editable for prep. */
const systemModeV2 = computed(() => isChannelMonitorV2Mode())
const systemModeLabel = computed(() => {
  if (!appStore.cachedPublicSettings?.channel_monitor_enabled) {
    return t('channelMonitorV2.settings.modeClosed')
  }
  return getChannelMonitorMode() === 'v1'
    ? t('channelMonitorV2.settings.modeV1')
    : t('channelMonitorV2.settings.modeV2')
})
const defaultThresholds = {
  minimum_sample: 50,
  warning_error_rate: 0.05,
  critical_error_rate: 0.20,
  target_ttft_ms: 3000,
  warning_ttft_ms: 3000,
  critical_ttft_ms: 10000,
  // Higher is better: below 85% watch, below 60% critical.
  warning_cache_rate: 0.85,
  critical_cache_rate: 0.60,
  error_weight: 0.60,
  ttft_weight: 0.20,
  cache_weight: 0.20,
}

/** Factory ignored categories (matches backend DefaultChannelMonitorV2IgnoredErrorCategories). */
const defaultIgnoredErrorCategories = [
  'authentication',
  'client_cancelled',
  'content_policy',
  'context_limit',
  'group_access',
  'model_unsupported',
  'not_found',
  'quota_or_balance',
] as const
function percentModel(key: 'warning_error_rate' | 'critical_error_rate' | 'warning_cache_rate' | 'critical_cache_rate') {
  return computed({
    get: () => ((draft.value?.health_thresholds?.[key] ?? defaultThresholds[key]) * 100),
    set: (value: number) => {
      if (!draft.value) return
      draft.value.health_thresholds[key] = Math.max(0, Math.min(100, Number(value) || 0)) / 100
    },
  })
}
const warningErrorPercent = percentModel('warning_error_rate')
const criticalErrorPercent = percentModel('critical_error_rate')
const warningCachePercent = percentModel('warning_cache_rate')
const criticalCachePercent = percentModel('critical_cache_rate')

function setModels(platform: MonitorConfig['platforms'][number], event: Event) {
  platform.models = [
    ...new Set(
      (event.target as HTMLInputElement).value
        .split(',')
        .map((v) => v.trim())
        .filter(Boolean)
    ),
  ].sort()
}

function toggleGroup(id: number) {
  if (!draft.value) return
  draft.value.group_ids = draft.value.group_ids.includes(id)
    ? draft.value.group_ids.filter((value) => value !== id)
    : [...draft.value.group_ids, id].sort((a, b) => a - b)
}

function isCategoryIgnored(category: string): boolean {
  return Boolean(draft.value?.ignored_error_categories?.includes(category))
}

function toggleIgnoredCategory(category: string) {
  if (!draft.value) return
  const current = new Set(draft.value.ignored_error_categories || [])
  if (current.has(category)) current.delete(category)
  else current.add(category)
  draft.value.ignored_error_categories = [...current].sort()
}

function categoryLabel(category: string) {
  const key = `channelMonitorV2.errorCategories.${category}`
  return te(key) ? t(key) : category
}

function platformLabel(value: string) {
  return (
    {
      anthropic: 'Claude',
      openai: 'OpenAI',
      grok: 'Grok',
      kiro: 'Kiro',
      gemini: 'Gemini',
      antigravity: 'Antigravity',
      composite: 'Composite',
    } as Record<string, string>
  )[value] || value
}

function normalizeConfig(value: MonitorConfig): MonitorConfig {
  const ignored = value.ignored_error_categories
  return {
    ...value,
    health_thresholds: { ...defaultThresholds, ...(value.health_thresholds || {}) },
    // Preserve explicit empty arrays from the server (operator cleared all).
    ignored_error_categories: [
      ...(ignored == null ? [...defaultIgnoredErrorCategories] : ignored),
    ].sort(),
  }
}

async function load() {
  loading.value = true
  try {
    const [value, groupRows] = await Promise.all([getConfig(), adminAPI.groups.getAllIncludingInactive()])
    const normalized = normalizeConfig(value)
    draft.value = structuredClone(normalized)
    groups.value = groupRows
    original.value = JSON.stringify(normalized)
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('channelMonitorV2.settings.loadFailed')))
  } finally {
    loading.value = false
  }
}

async function save() {
  if (!draft.value) return
  saving.value = true
  try {
    const payload = normalizeConfig(draft.value)
    const value = await updateConfig(payload)
    const normalized = normalizeConfig(value)
    draft.value = structuredClone(normalized)
    original.value = JSON.stringify(normalized)
    appStore.showSuccess(t('channelMonitorV2.settings.saveSuccess'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('channelMonitorV2.settings.saveFailed')))
    await load()
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.scheme3-v2-settings {
  --settings-paper: #f4f2ec;
  --settings-card: #fffefa;
  --settings-subtle: #f1eee6;
  --settings-ink: #27251f;
  --settings-muted: #777266;
  --settings-soft: #a49e90;
  --settings-line: #d8d2c3;
  --settings-accent: #1e5c42;
  --settings-amber: #b7791f;
  --settings-danger: #9e4d3d;
  color: var(--settings-ink);
}
.scheme3-v2-settings-header {
  border-bottom: 1px solid var(--settings-line);
  padding-bottom: 1rem;
}
.scheme3-v2-settings-title {
  margin: 0;
  color: var(--settings-ink);
  font-family: Georgia, 'Times New Roman', serif;
  font-size: 1.55rem;
  font-weight: 500;
}
.scheme3-v2-settings-mark {
  border: 1px solid rgba(30, 92, 66, .28);
  border-radius: 7px;
  background: rgba(30, 92, 66, .08);
  color: var(--settings-accent);
}
.scheme3-v2-settings-description,
.scheme3-v2-settings-hint,
.scheme3-v2-settings-meta,
.scheme3-v2-settings-empty,
.scheme3-v2-settings-footer { color: var(--settings-muted); }
.scheme3-v2-settings-save,
.scheme3-v2-settings-inline-action {
  border: 1px solid var(--settings-line);
  border-radius: 6px;
  background: var(--settings-card);
  color: var(--settings-muted);
  padding: .48rem .72rem;
  font-size: .68rem;
  font-weight: 800;
  transition: border-color 150ms ease, background-color 150ms ease, color 150ms ease;
}
.scheme3-v2-settings-save { border-color: var(--settings-accent); background: var(--settings-accent); color: #fffefa; }
.scheme3-v2-settings-save:hover:not(:disabled),
.scheme3-v2-settings-inline-action:hover { border-color: var(--settings-accent); background: var(--settings-subtle); color: var(--settings-accent); }
.scheme3-v2-settings-save:disabled { cursor: not-allowed; opacity: .45; }
.scheme3-v2-settings-banner,
.scheme3-v2-settings-note {
  border: 1px solid var(--settings-line);
  border-radius: 7px;
}
.scheme3-v2-settings-banner { border-color: rgba(183, 121, 31, .35); background: rgba(183, 121, 31, .09); color: var(--settings-amber); }
.scheme3-v2-settings-link { color: var(--settings-accent); }
.scheme3-v2-settings-loading {
  min-height: 12rem;
  border: 1px dashed var(--settings-line);
  border-radius: 8px;
  color: var(--settings-muted);
}
.scheme3-v2-settings-panel {
  overflow: hidden;
  border: 1px solid var(--settings-line);
  border-radius: 8px;
  background: var(--settings-card);
  box-shadow: 0 10px 24px rgba(54, 48, 34, .05);
}
.scheme3-v2-settings-row { border-bottom: 1px solid var(--settings-line); }
.scheme3-v2-settings-row:last-child { border-bottom: 0; }
.scheme3-v2-settings-panel-header {
  border-bottom: 1px solid var(--settings-line);
  padding: .8rem 1.25rem;
}
.scheme3-v2-settings-label { color: var(--settings-ink); }
.scheme3-v2-settings-list { background: var(--settings-card); }
.scheme3-v2-settings-segments {
  gap: 0;
  border: 1px solid var(--settings-line);
  border-radius: 7px;
  background: var(--settings-subtle);
  padding: 2px;
}
.scheme3-v2-settings-segment {
  min-height: 1.8rem;
  border-radius: 5px;
  color: var(--settings-muted);
  padding: .3rem .58rem;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: .62rem;
  font-weight: 800;
}
.scheme3-v2-settings-segment:hover { color: var(--settings-ink); }
.scheme3-v2-settings-segment.is-active { background: var(--settings-card); color: var(--settings-accent); box-shadow: 0 2px 6px rgba(54, 48, 34, .08); }
.scheme3-v2-settings-input {
  width: 100%;
  min-height: 2.2rem;
  border: 1px solid var(--settings-line);
  border-radius: 6px;
  background: var(--settings-card);
  color: var(--settings-ink);
  padding: .45rem .6rem;
  font-size: .72rem;
  transition: border-color 150ms ease, background-color 150ms ease, box-shadow 150ms ease;
}
.scheme3-v2-settings-input:focus { border-color: var(--settings-accent); outline: none; box-shadow: 0 0 0 2px rgba(30, 92, 66, .14); }
.scheme3-v2-settings-badge {
  border: 1px solid var(--settings-line);
  border-radius: 999px;
  padding: .2rem .48rem;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: .56rem;
  font-weight: 800;
}
.scheme3-v2-settings-badge.is-muted { color: var(--settings-muted); background: var(--settings-subtle); }
.scheme3-v2-settings-badge.is-accent { border-color: rgba(30, 92, 66, .25); color: var(--settings-accent); background: rgba(30, 92, 66, .08); }
.scheme3-v2-settings-choice { border-radius: 6px; }
.scheme3-v2-settings-choice:hover { background: var(--settings-subtle); }
.scheme3-v2-settings-choice-label { color: var(--settings-ink); }
.scheme3-v2-settings-checkbox {
  width: 1rem;
  height: 1rem;
  flex: none;
  appearance: none;
  border: 1px solid var(--settings-soft);
  border-radius: 4px;
  background: var(--settings-card);
  transition: border-color 120ms ease, background-color 120ms ease, box-shadow 120ms ease;
}
.scheme3-v2-settings-checkbox:checked { border-color: var(--settings-accent); background: var(--settings-accent); box-shadow: inset 0 0 0 3px var(--settings-card); }
.scheme3-v2-settings-checkbox:focus-visible { outline: 2px solid rgba(30, 92, 66, .28); outline-offset: 2px; }
.scheme3-v2-settings-footer { border-top: 1px solid var(--settings-line); }
.scheme3-v2-settings-field-label {
  display: block;
  margin-bottom: .35rem;
  color: var(--settings-muted);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: .58rem;
  font-weight: 800;
  letter-spacing: .04em;
  text-transform: uppercase;
}
.scheme3-v2-settings-note-accent { border-color: rgba(30, 92, 66, .25); background: rgba(30, 92, 66, .07); color: var(--settings-accent); }
.scheme3-v2-settings-note-neutral { background: var(--settings-subtle); color: var(--settings-muted); }
:global(.dark .scheme3-v2-settings) {
  --settings-paper: #1b1b18;
  --settings-card: #24231f;
  --settings-subtle: #2b2924;
  --settings-ink: #f4f2ec;
  --settings-muted: #aaa69a;
  --settings-soft: #827e72;
  --settings-line: #47443a;
  --settings-accent: #8fc2a5;
  --settings-amber: #d3a55a;
  --settings-danger: #d38b79;
}
:global(.dark .scheme3-v2-settings-banner) { border-color: rgba(211, 165, 90, .35); background: rgba(211, 165, 90, .1); }
:global(.dark .scheme3-v2-settings-save) { color: #1b1b18; }
:global(.dark .scheme3-v2-settings-note-accent) { border-color: rgba(143, 194, 165, .28); background: rgba(143, 194, 165, .1); }
:global(.dark .scheme3-v2-settings-input),
:global(.dark .scheme3-v2-settings-checkbox) { background: var(--settings-card); color: var(--settings-ink); }
:global(.dark .scheme3-v2-settings-checkbox:checked) { box-shadow: inset 0 0 0 3px var(--settings-card); }

@media (max-width: 640px) {
  .scheme3-v2-settings { padding-left: 0; padding-right: 0; }
  .scheme3-v2-settings-header { padding-left: .2rem; padding-right: .2rem; }
  .scheme3-v2-settings-title { font-size: 1.35rem; }
  .scheme3-v2-settings-panel-header { padding-left: .9rem; padding-right: .9rem; }
}
</style>

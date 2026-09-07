<template>
  <div class="scheme3-monitor-advanced space-y-4">
    <!-- Headers key-value rows -->
    <div>
      <label class="scheme3-monitor-label">{{ t('admin.channelMonitor.advanced.headers') }}</label>
      <div class="space-y-1.5">
        <div
          v-for="(row, i) in headerRows"
          :key="i"
          class="flex items-center gap-2"
        >
          <input
            v-model="row.name"
            type="text"
            spellcheck="false"
            :placeholder="t('admin.channelMonitor.advanced.headerNamePlaceholder')"
            class="scheme3-monitor-input w-52 flex-none font-mono text-xs"
            @blur="commitHeaders"
          />
          <input
            v-model="row.value"
            type="text"
            spellcheck="false"
            :placeholder="t('admin.channelMonitor.advanced.headerValuePlaceholder')"
            class="scheme3-monitor-input flex-1 font-mono text-xs"
            @blur="commitHeaders"
          />
          <button
            type="button"
            class="scheme3-monitor-icon-button is-danger flex-none"
            :title="t('common.delete')"
            @click="removeRow(i)"
          >
            <Icon name="x" size="sm" />
          </button>
        </div>
        <button
          type="button"
          class="scheme3-monitor-button scheme3-monitor-button-dashed"
          @click="addRow"
        >
          <Icon name="plus" size="sm" />
          {{ t('admin.channelMonitor.advanced.headerAddRow') }}
        </button>
      </div>
      <p v-if="headersError" class="scheme3-monitor-error mt-1 text-xs">{{ headersError }}</p>
      <p v-else class="scheme3-monitor-hint mt-1 text-xs">
        {{ t('admin.channelMonitor.advanced.headersHint') }}
      </p>
    </div>

    <!-- Body mode radio -->
    <div>
      <label class="scheme3-monitor-label">{{ t('admin.channelMonitor.advanced.bodyMode') }}</label>
      <div class="grid grid-cols-3 gap-3">
        <button
          v-for="opt in bodyModeOptions"
          :key="opt.value"
          type="button"
          class="scheme3-monitor-choice"
          :class="bodyModeButtonClass(opt.value)"
          @click="updateBodyMode(opt.value)"
        >
          {{ opt.label }}
        </button>
      </div>
      <p class="scheme3-monitor-hint mt-1 text-xs">
        {{ bodyModeHint }}
      </p>
    </div>

    <!-- Body JSON (仅当 mode != off) -->
    <div v-if="bodyOverrideMode !== 'off'">
      <div class="mb-1 flex items-center justify-between">
        <label class="scheme3-monitor-label !mb-0">{{ t('admin.channelMonitor.advanced.bodyJson') }}</label>
        <button
          type="button"
          class="scheme3-monitor-inline-action text-xs"
          :disabled="!bodyText.trim()"
          @click="formatBody"
        >
          {{ t('admin.channelMonitor.advanced.bodyJsonFormat') }}
        </button>
      </div>
      <textarea
        v-model="bodyText"
        rows="10"
        :placeholder="bodyPlaceholder"
        class="scheme3-monitor-input font-mono text-xs"
        style="white-space: pre; overflow-wrap: normal; overflow-x: auto;"
        spellcheck="false"
        @blur="commitBody"
      />
      <p v-if="bodyError" class="scheme3-monitor-error mt-1 text-xs">{{ bodyError }}</p>
      <p v-else class="scheme3-monitor-hint mt-1 text-xs">
        {{ t('admin.channelMonitor.advanced.bodyJsonHint') }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { APIMode, BodyOverrideMode, Provider } from '@/api/admin/channelMonitor'
import {
  API_MODE_RESPONSES,
  DEFAULT_GROK_MODEL,
  PROVIDER_GROK,
  PROVIDER_OPENAI,
} from '@/constants/channelMonitor'

const props = defineProps<{
  provider?: Provider
  apiMode?: APIMode
  extraHeaders: Record<string, string>
  bodyOverrideMode: BodyOverrideMode
  bodyOverride: Record<string, unknown> | null
}>()

const emit = defineEmits<{
  (e: 'update:extraHeaders', value: Record<string, string>): void
  (e: 'update:bodyOverrideMode', value: BodyOverrideMode): void
  (e: 'update:bodyOverride', value: Record<string, unknown> | null): void
}>()

const { t } = useI18n()

// ---- Headers key-value rows ----
interface HeaderRow {
  name: string
  value: string
}

const headerRows = ref<HeaderRow[]>(toRows(props.extraHeaders))
const headersError = ref('')

watch(
  () => props.extraHeaders,
  (v) => {
    // 外部重置时（切换平台 / 应用模板）同步行。
    // 同值不回写，避免每次 commit 都把行重排。
    if (!isSameHeaderMap(toMap(headerRows.value), v)) {
      headerRows.value = toRows(v)
    }
    headersError.value = ''
  },
)

function toRows(h: Record<string, string>): HeaderRow[] {
  const entries = Object.entries(h || {})
  if (entries.length === 0) return [{ name: '', value: '' }]
  return entries.map(([name, value]) => ({ name, value }))
}

function toMap(rows: HeaderRow[]): Record<string, string> {
  const out: Record<string, string> = {}
  for (const row of rows) {
    const name = row.name.trim()
    if (name === '') continue
    out[name] = row.value
  }
  return out
}

function isSameHeaderMap(a: Record<string, string>, b: Record<string, string>): boolean {
  const ak = Object.keys(a)
  const bk = Object.keys(b || {})
  if (ak.length !== bk.length) return false
  for (const k of ak) {
    if (a[k] !== b[k]) return false
  }
  return true
}

function commitHeaders() {
  // 空白 name + 空白 value 的行允许保留作为"占位新行"，不报错；
  // name 非空但 value 为空（或反之）都视为用户正在编辑，同样不报错。
  // 只在 name 里含冒号这种明显不合法时兜一下。
  for (const row of headerRows.value) {
    const name = row.name.trim()
    if (name === '') continue
    if (name.includes(':') || /\s/.test(name)) {
      headersError.value = t('admin.channelMonitor.advanced.headerNameInvalid', { name })
      return
    }
  }
  headersError.value = ''
  emit('update:extraHeaders', toMap(headerRows.value))
}

function addRow() {
  headerRows.value.push({ name: '', value: '' })
}

function removeRow(index: number) {
  headerRows.value.splice(index, 1)
  if (headerRows.value.length === 0) {
    headerRows.value.push({ name: '', value: '' })
  }
  commitHeaders()
}

// ---- Body mode + JSON ----
const bodyText = ref(serializeBody(props.bodyOverride))
const bodyError = ref('')

watch(
  () => props.bodyOverride,
  (v) => {
    bodyText.value = serializeBody(v)
    bodyError.value = ''
  },
)

function commitBody() {
  if (props.bodyOverrideMode === 'off') {
    return
  }
  const trimmed = bodyText.value.trim()
  if (trimmed === '') {
    emit('update:bodyOverride', null)
    bodyError.value = ''
    return
  }
  try {
    const parsed = JSON.parse(trimmed)
    if (parsed === null || typeof parsed !== 'object' || Array.isArray(parsed)) {
      bodyError.value = t('admin.channelMonitor.advanced.bodyJsonObjectError')
      return
    }
    emit('update:bodyOverride', parsed as Record<string, unknown>)
    bodyError.value = ''
  } catch (e) {
    bodyError.value =
      t('admin.channelMonitor.advanced.bodyJsonError') +
      ': ' +
      (e instanceof Error ? e.message : String(e))
  }
}

function formatBody() {
  const trimmed = bodyText.value.trim()
  if (trimmed === '') return
  try {
    const parsed = JSON.parse(trimmed)
    bodyText.value = JSON.stringify(parsed, null, 2)
    bodyError.value = ''
    // 同步把校验过的对象提交，避免格式化后焦点未移走时父组件读到旧值
    if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
      emit('update:bodyOverride', parsed as Record<string, unknown>)
    }
  } catch (e) {
    bodyError.value =
      t('admin.channelMonitor.advanced.bodyJsonError') +
      ': ' +
      (e instanceof Error ? e.message : String(e))
  }
}

function serializeBody(body: Record<string, unknown> | null): string {
  if (!body || Object.keys(body).length === 0) return ''
  return JSON.stringify(body, null, 2)
}

function updateBodyMode(mode: BodyOverrideMode) {
  emit('update:bodyOverrideMode', mode)
  // 切换到 off 时清掉 body（提示用户）
  if (mode === 'off') {
    emit('update:bodyOverride', null)
  }
}

const bodyModeOptions = computed<{ value: BodyOverrideMode; label: string }[]>(() => [
  { value: 'off', label: t('admin.channelMonitor.advanced.bodyModeOff') },
  { value: 'merge', label: t('admin.channelMonitor.advanced.bodyModeMerge') },
  { value: 'replace', label: t('admin.channelMonitor.advanced.bodyModeReplace') },
])

function bodyModeButtonClass(mode: BodyOverrideMode): string {
  return props.bodyOverrideMode === mode ? 'is-active' : ''
}

const bodyModeHint = computed(() => {
  switch (props.bodyOverrideMode) {
    case 'merge':
      return t('admin.channelMonitor.advanced.bodyModeHintMerge')
    case 'replace':
      return t('admin.channelMonitor.advanced.bodyModeHintReplace')
    default:
      return t('admin.channelMonitor.advanced.bodyModeHintOff')
  }
})

const bodyPlaceholder = computed(() => {
  if (props.provider === PROVIDER_OPENAI && props.apiMode === API_MODE_RESPONSES) {
    if (props.bodyOverrideMode === 'merge') {
      return '{\n  "max_output_tokens": 20\n}'
    }
    return '{\n  "model": "gpt-4o-mini",\n  "instructions": "You are a health check endpoint. Reply briefly.",\n  "input": "Reply with exactly: ok",\n  "max_output_tokens": 20,\n  "stream": false\n}'
  }
  if (props.provider === PROVIDER_OPENAI || props.provider === PROVIDER_GROK) {
    if (props.bodyOverrideMode === 'merge') {
      return '{\n  "max_tokens": 20\n}'
    }
    const model = props.provider === PROVIDER_GROK ? DEFAULT_GROK_MODEL : 'gpt-4o-mini'
    return `{\n  "model": "${model}",\n  "messages": [{"role":"user","content":"Reply with exactly: ok"}],\n  "max_tokens": 20,\n  "stream": false\n}`
  }
  if (props.bodyOverrideMode === 'merge') {
    return '{\n  "system": "You are Claude Code..."\n}'
  }
  return '{\n  "model": "claude-x",\n  "messages": [{"role":"user","content":"hi"}],\n  "max_tokens": 10\n}'
})
</script>

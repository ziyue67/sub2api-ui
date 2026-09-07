<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useClipboard } from '@/composables/useClipboard'
import type { CustomEndpoint } from '@/types'

const props = defineProps<{
  apiBaseUrl: string
  customEndpoints: CustomEndpoint[]
}>()

const { t } = useI18n()
const { copyToClipboard } = useClipboard()
const copiedEndpoint = ref<string | null>(null)

let copiedResetTimer: number | undefined

const allEndpoints = computed(() => {
  const items: Array<{ name: string; endpoint: string; description: string; isDefault: boolean }> = []
  if (props.apiBaseUrl) {
    items.push({
      name: t('keys.endpoints.title'),
      endpoint: props.apiBaseUrl,
      description: '',
      isDefault: true,
    })
  }
  for (const ep of props.customEndpoints) {
    items.push({ ...ep, isDefault: false })
  }
  return items
})

async function copy(url: string) {
  const success = await copyToClipboard(url, t('keys.endpoints.copied'))
  if (!success) return

  copiedEndpoint.value = url
  if (copiedResetTimer !== undefined) {
    window.clearTimeout(copiedResetTimer)
  }
  copiedResetTimer = window.setTimeout(() => {
    if (copiedEndpoint.value === url) {
      copiedEndpoint.value = null
    }
  }, 1800)
}

function tooltipHint(endpoint: string): string {
  return copiedEndpoint.value === endpoint
    ? t('keys.endpoints.copiedHint')
    : t('keys.endpoints.clickToCopy')
}

function speedTestUrl(endpoint: string): string {
  return `https://www.tcptest.cn/http/${encodeURIComponent(endpoint)}`
}

onBeforeUnmount(() => {
  if (copiedResetTimer !== undefined) {
    window.clearTimeout(copiedResetTimer)
  }
})
</script>

<template>
  <div v-if="allEndpoints.length > 0" class="flex flex-wrap gap-2">
    <div
      v-for="(item, index) in allEndpoints"
      :key="index"
      class="flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-2.5 py-1.5 text-xs transition-colors hover:border-primary-200 dark:border-dark-600 dark:bg-dark-800 dark:hover:border-primary-700"
    >
      <span class="font-medium text-gray-600 dark:text-gray-300">{{ item.name }}</span>
      <span
        v-if="item.isDefault"
        class="rounded bg-primary-50 px-1 py-px text-[10px] font-medium leading-tight text-primary-600 dark:bg-primary-900/30 dark:text-primary-400"
      >{{ t('keys.endpoints.default') }}</span>

      <span class="text-gray-300 dark:text-dark-500">|</span>

      <div class="group/endpoint relative flex items-center gap-1.5">
        <div
          class="scheme3-endpoint-tooltip pointer-events-none absolute bottom-full left-1/2 z-20 mb-2 w-max max-w-[24rem] -translate-x-1/2 translate-y-1 rounded-xl px-3 py-2.5 text-left opacity-0 transition-all duration-150 group-hover/endpoint:translate-y-0 group-hover/endpoint:opacity-100 group-focus-within/endpoint:translate-y-0 group-focus-within/endpoint:opacity-100"
        >
          <p
            v-if="item.description"
            class="scheme3-endpoint-description max-w-[24rem] break-words text-xs leading-5"
          >
            {{ item.description }}
          </p>
          <p
            class="scheme3-endpoint-hint flex items-center gap-1.5 text-[11px] leading-4"
            :class="item.description ? 'mt-1.5' : ''"
          >
            <span class="scheme3-endpoint-hint-dot h-1.5 w-1.5 rounded-full"></span>
            {{ tooltipHint(item.endpoint) }}
          </p>
          <div class="scheme3-endpoint-tooltip-arrow absolute left-1/2 top-full h-3 w-3 -translate-x-1/2 -translate-y-1/2 rotate-45"></div>
        </div>

        <code
          class="cursor-pointer font-mono text-gray-500 decoration-gray-400 decoration-dashed underline-offset-2 hover:text-primary-600 hover:underline focus:text-primary-600 focus:underline focus:outline-none dark:text-gray-400 dark:decoration-gray-500 dark:hover:text-primary-400 dark:focus:text-primary-400"
          role="button"
          tabindex="0"
          @click="copy(item.endpoint)"
          @keydown.enter.prevent="copy(item.endpoint)"
          @keydown.space.prevent="copy(item.endpoint)"
        >{{ item.endpoint }}</code>

        <button
          type="button"
          class="rounded p-0.5 transition-colors"
          :class="copiedEndpoint === item.endpoint
            ? 'text-emerald-500 dark:text-emerald-400'
            : 'text-gray-400 hover:text-primary-500 dark:text-gray-500 dark:hover:text-primary-400'"
          :aria-label="tooltipHint(item.endpoint)"
          @click="copy(item.endpoint)"
        >
          <svg v-if="copiedEndpoint === item.endpoint" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
          </svg>
          <svg v-else class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
          </svg>
        </button>

        <a
          :href="speedTestUrl(item.endpoint)"
          target="_blank"
          rel="noopener noreferrer"
          class="rounded p-0.5 text-gray-400 transition-colors hover:text-amber-500 dark:text-gray-500 dark:hover:text-amber-400"
          :title="t('keys.endpoints.speedTest')"
        >
          <svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
        </a>
      </div>
    </div>
  </div>
</template>

<style scoped>
.scheme3-endpoint-tooltip {
  border: 1px solid #dad5c8;
  background: #fbfaf6;
  color: #27251f;
  box-shadow: 0 14px 36px rgba(54, 48, 34, .16);
}
.scheme3-endpoint-description { color: #6b695f; }
.scheme3-endpoint-hint { color: #1e5c42; }
.scheme3-endpoint-hint-dot { background: #1e5c42; }
.scheme3-endpoint-tooltip-arrow { border-right: 1px solid #dad5c8; border-bottom: 1px solid #dad5c8; background: #fbfaf6; }
:global(html.dark .scheme3-endpoint-tooltip) { border-color: #47443a; background: #24231f; color: #f4f2ec; box-shadow: 0 16px 38px rgba(0, 0, 0, .28); }
:global(html.dark .scheme3-endpoint-description) { color: #aaa69a; }
:global(html.dark .scheme3-endpoint-hint) { color: #8fc2a5; }
:global(html.dark .scheme3-endpoint-hint-dot) { background: #8fc2a5; }
:global(html.dark .scheme3-endpoint-tooltip-arrow) { border-color: #47443a; background: #24231f; }
</style>

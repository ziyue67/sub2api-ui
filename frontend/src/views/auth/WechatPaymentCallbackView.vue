<template>
  <AuthLayout>
    <section class="scheme3-wechat-payment-callback">
        <h1 class="text-lg font-semibold text-gray-900 dark:text-white">
          {{ callbackTitleText }}
        </h1>
        <p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
          {{ errorMessage || callbackProcessingText }}
        </p>

        <div
          v-if="!errorMessage"
          class="mt-6 flex items-center justify-center py-10"
        >
          <div
            class="h-8 w-8 animate-spin rounded-full border-4 border-primary-500 border-t-transparent"
          ></div>
        </div>

        <div v-else class="scheme3-wechat-payment-callback-error">
          <p class="text-sm text-gray-700 dark:text-gray-300">
            {{ errorMessage }}
          </p>
          <button
            class="btn btn-primary mt-4"
            type="button"
            @click="goBackToPayment"
          >
            {{ backToPaymentText }}
          </button>
        </div>
    </section>
  </AuthLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { AuthLayout } from '@/components/layout'
import { useAppStore } from '@/stores'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()

const errorMessage = ref('')

watch(errorMessage, (message) => {
  if (message) {
    appStore.showError(message)
  }
})

const callbackProcessingText = computed(() => t('auth.wechatPayment.callbackProcessing'))
const callbackTitleText = computed(() => t('auth.wechatPayment.callbackTitle'))
const backToPaymentText = computed(() => t('auth.wechatPayment.backToPayment'))

function readQueryString(key: string): string {
  const value = route.query[key]
  if (Array.isArray(value)) {
    return typeof value[0] === 'string' ? value[0] : ''
  }
  return typeof value === 'string' ? value : ''
}

function parseFragmentParams(): URLSearchParams {
  const raw = typeof window !== 'undefined' ? window.location.hash : ''
  const hash = raw.startsWith('#') ? raw.slice(1) : raw
  return new URLSearchParams(hash)
}

function normalizeRedirectPath(path: string | null | undefined): string {
  const value = (path || '').trim()
  if (!value) return '/purchase'
  if (!value.startsWith('/')) return '/purchase'
  if (value.startsWith('//') || value.includes('://')) return '/purchase'
  if (value === '/payment') return '/purchase'
  if (value.startsWith('/payment?')) return '/purchase' + value.slice('/payment'.length)
  return value
}

function appendQueryParam(query: Record<string, string>, key: string, value: string) {
  if (value) {
    query[key] = value
  }
}

function goBackToPayment() {
  void router.replace('/purchase')
}

onMounted(async () => {
  const fragment = parseFragmentParams()
  const readParam = (key: string) => fragment.get(key) || readQueryString(key)

  const error = readParam('error') || readParam('err_msg') || readParam('errmsg')
  const errorDescription = readParam('error_description') || readParam('message')

  if (error) {
    errorMessage.value = errorDescription || error
    return
  }

  const resumeToken = readParam('wechat_resume_token')
  const openid = readParam('openid')
  const state = readParam('state')
  const scope = readParam('scope')
  const paymentType = readParam('payment_type')
  const amount = readParam('amount')
  const orderType = readParam('order_type')
  const planId = readParam('plan_id')
  const redirectURL = new URL(
    normalizeRedirectPath(readParam('redirect')),
    window.location.origin,
  )

  if (!resumeToken && !openid) {
    errorMessage.value = t('auth.wechatPayment.callbackMissingResumeToken')
    return
  }

  const query: Record<string, string> = {
    ...Object.fromEntries(redirectURL.searchParams.entries()),
    wechat_resume: '1',
  }

  if (resumeToken) {
    query.wechat_resume_token = resumeToken
  } else {
    query.openid = openid
    appendQueryParam(query, 'state', state)
    appendQueryParam(query, 'scope', scope)
    appendQueryParam(query, 'payment_type', paymentType)
    appendQueryParam(query, 'amount', amount)
    appendQueryParam(query, 'order_type', orderType)
    appendQueryParam(query, 'plan_id', planId)
  }

  await router.replace({
    path: redirectURL.pathname,
    query,
  })
})
</script>

<style scoped>
.scheme3-wechat-payment-callback {
  --wechat-callback-ink: #24231f;
  --wechat-callback-muted: #777266;
  --wechat-callback-line: #d8d2c3;
  --wechat-callback-green: #1e5c42;
  min-width: 0;
}

.scheme3-wechat-payment-callback :deep(.text-gray-900),
.scheme3-wechat-payment-callback :deep(.text-gray-800),
.scheme3-wechat-payment-callback :deep(.text-gray-700) { color: var(--wechat-callback-ink) !important; }
.scheme3-wechat-payment-callback :deep(.text-gray-600),
.scheme3-wechat-payment-callback :deep(.text-gray-500),
.scheme3-wechat-payment-callback :deep(.text-gray-400) { color: var(--wechat-callback-muted) !important; }
.scheme3-wechat-payment-callback :deep(.border-primary-500) { border-color: var(--wechat-callback-green) !important; }
.scheme3-wechat-payment-callback :deep(.btn) { min-height: 2.45rem; border-radius: 6px; font-size: .82rem; font-weight: 700; }
.scheme3-wechat-payment-callback :deep(.btn-primary) { border-color: var(--wechat-callback-green); background: var(--wechat-callback-green); color: #fffefa; }
.scheme3-wechat-payment-callback :deep(.btn-primary:hover) { background: #287052; }
.scheme3-wechat-payment-callback-error { margin-top: 1.5rem; border: 1px solid #c98772; border-radius: 6px; background: #fff8f4; padding: 1rem; }

:global(html.dark .scheme3-wechat-payment-callback) { --wechat-callback-ink: #f4f2ec; --wechat-callback-muted: #aaa69a; --wechat-callback-line: #47443a; --wechat-callback-green: #8fc2a5; }
:global(html.dark .scheme3-wechat-payment-callback .btn-primary) { border-color: var(--wechat-callback-green); background: var(--wechat-callback-green); color: #1b1b18; }
:global(html.dark .scheme3-wechat-payment-callback .btn-primary:hover) { background: #a7d2b7; }
:global(html.dark .scheme3-wechat-payment-callback-error) { border-color: #8f5c52; background: #321f1b; }
</style>

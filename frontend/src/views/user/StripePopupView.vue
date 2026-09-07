<template>
  <div class="scheme3-stripe-popup flex min-h-screen items-center justify-center p-4">
    <div
      class="scheme3-stripe-popup-panel w-full max-w-md space-y-4 p-6"
    >
      <!-- Amount + Order ID -->
      <div v-if="amount" class="text-center">
        <p class="scheme3-stripe-popup-amount text-3xl font-bold">¥{{ amount }}</p>
        <p v-if="orderId" class="scheme3-stripe-popup-order mt-1 text-sm">
          {{ t('payment.orders.orderId') }}: {{ orderId }}
        </p>
      </div>

      <!-- Error -->
      <div v-if="error" class="space-y-3">
        <div
          class="scheme3-stripe-popup-error p-3 text-sm"
        >
          {{ error }}
        </div>
        <button
          class="scheme3-stripe-popup-close w-full text-sm underline"
          @click="closeWindow"
        >
          {{ t('common.close') }}
        </button>
      </div>

      <!-- Success -->
      <div v-else-if="success" class="space-y-3 py-4 text-center">
        <div class="scheme3-stripe-popup-success-mark text-5xl">✓</div>
        <p class="scheme3-stripe-popup-order text-sm">{{ t('payment.result.success') }}</p>
        <button
          class="scheme3-stripe-popup-close text-sm underline"
          @click="closeWindow"
        >
          {{ t('common.close') }}
        </button>
      </div>

      <!-- Loading / Redirecting -->
      <div v-else class="flex items-center justify-center py-8">
        <div
          class="scheme3-stripe-popup-spinner h-8 w-8 animate-spin rounded-full border-2 border-t-transparent"
        />
        <span class="scheme3-stripe-popup-order ml-3 text-sm">{{ hint }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { isMobileDevice } from '@/utils/device'
import { buildApiUrl } from '@/api/client'

interface StripeWithWechatPay {
  confirmWechatPayPayment(clientSecret: string, options: Record<string, unknown>): Promise<{ error?: { message?: string }; paymentIntent?: { status: string } }>
}

const { t } = useI18n()
const route = useRoute()

const orderId = String(route.query.order_id || '')
const method = String(route.query.method || 'alipay')
const amount = String(route.query.amount || '')

const error = ref('')
const success = ref(false)
const hint = ref(t('payment.stripePopup.redirecting'))

let pollTimer: ReturnType<typeof setInterval> | null = null
let initTimeoutTimer: ReturnType<typeof setTimeout> | null = null
let messageHandler: ((event: MessageEvent) => void) | null = null

function closeWindow() { window.close() }

function clearInitTimeout() {
  if (initTimeoutTimer) {
    clearTimeout(initTimeoutTimer)
    initTimeoutTimer = null
  }
}

onMounted(() => {
  messageHandler = (event: MessageEvent) => {
    if (event.origin !== window.location.origin) return
    if (event.data?.type !== 'STRIPE_POPUP_INIT') return
    // INIT 已到达，取消兜底超时，避免长时间的扫码支付被误判为超时。
    clearInitTimeout()
    if (messageHandler) {
      window.removeEventListener('message', messageHandler)
      messageHandler = null
    }
    initStripe(event.data.clientSecret, event.data.publishableKey)
  }
  window.addEventListener('message', messageHandler)

  if (window.opener) {
    window.opener.postMessage({ type: 'STRIPE_POPUP_READY' }, window.location.origin)
  }

  // 仅兜底“父窗口始终未发 STRIPE_POPUP_INIT”的场景。
  initTimeoutTimer = setTimeout(() => {
    if (!error.value && !success.value) {
      error.value = t('payment.stripePopup.timeout')
    }
  }, 15000)
})

onUnmounted(() => {
  if (pollTimer) { clearInterval(pollTimer); pollTimer = null }
  clearInitTimeout()
  if (messageHandler) {
    window.removeEventListener('message', messageHandler)
    messageHandler = null
  }
})

async function initStripe(clientSecret: string, publishableKey: string) {
  if (!clientSecret || !publishableKey) {
    error.value = t('payment.stripeMissingParams')
    return
  }
  try {
    const { loadStripe } = await import('@stripe/stripe-js/pure')
    const stripe = await loadStripe(publishableKey)
    if (!stripe) { error.value = t('payment.stripeLoadFailed'); return }

    const returnUrl = window.location.origin + '/payment/result?order_id=' + orderId + '&status=success'

    if (method === 'alipay') {
      // Alipay: redirect this popup to Alipay payment page
      const { error: err } = await stripe.confirmAlipayPayment(clientSecret, { return_url: returnUrl })
      if (err) error.value = err.message || t('payment.result.failed')
    } else if (method === 'wechat_pay') {
      // WeChat: Stripe shows its built-in QR dialog, user scans, promise resolves
      hint.value = t('payment.stripePopup.loadingQr')
      const result = await (stripe as unknown as StripeWithWechatPay).confirmWechatPayPayment(clientSecret, {
        payment_method_options: { wechat_pay: { client: isMobileDevice() ? 'mobile_web' : 'web' } },
      })
      if (result.error) {
        error.value = result.error.message || t('payment.result.failed')
      } else if (result.paymentIntent?.status === 'succeeded') {
        success.value = true
        setTimeout(closeWindow, 2000)
      } else {
        // Payment not completed (user closed QR dialog)
        startPolling()
      }
    }
  } catch (err: unknown) {
    error.value = extractI18nErrorMessage(err, t, 'payment.errors', t('payment.stripeLoadFailed'))
  }
}

function startPolling() {
  let inFlight = false
  pollTimer = setInterval(async () => {
    // 防重入：接口响应慢于轮询间隔时避免并发重叠请求。
    if (inFlight) return
    inFlight = true
    try {
      // access token 存储在 localStorage 的 'auth_token' 键下（见 api/client.ts），
      // 之前误读 'token' 导致轮询请求不带认证、永远 401，支付成功无法被检测到。
      const token = localStorage.getItem('auth_token') || ''
      const res = await fetch(buildApiUrl(`/payment/orders/${orderId}`), {
        headers: token ? { Authorization: 'Bearer ' + token } : {},
        credentials: 'include',
      })
      if (!res.ok) return
      const data = await res.json()
      const status = data?.data?.status
      if (status === 'COMPLETED' || status === 'PAID') {
        if (pollTimer) { clearInterval(pollTimer); pollTimer = null }
        success.value = true
        setTimeout(closeWindow, 2000)
      }
    } catch { /* ignore */ } finally {
      inFlight = false
    }
  }, 3000)
}
</script>

<style scoped>
.scheme3-stripe-popup { background: #f4f2ec; color: #27251f; }
.scheme3-stripe-popup-panel { border: 1px solid #d8d2c3; border-radius: 8px; background: #fffefa; box-shadow: 0 18px 38px rgba(54,48,34,.12); }
.scheme3-stripe-popup-amount { color: #1e5c42; font-family: Georgia, 'Times New Roman', serif; }
.scheme3-stripe-popup-order { color: #777266; }
.scheme3-stripe-popup-error { border: 1px solid rgba(158,77,61,.3); border-radius: 7px; background: rgba(158,77,61,.08); color: #9e4d3d; }
.scheme3-stripe-popup-close { color: #1e5c42; text-underline-offset: 3px; transition: color 150ms ease; }
.scheme3-stripe-popup-close:hover { color: #174a35; }
.scheme3-stripe-popup-success-mark { color: #1e5c42; }
.scheme3-stripe-popup-spinner { border-color: #1e5c42; border-top-color: transparent; }

:global(.dark .scheme3-stripe-popup) { background: #1b1b18; color: #f4f2ec; }
:global(.dark .scheme3-stripe-popup-panel) { border-color: #47443a; background: #24231f; box-shadow: 0 18px 38px rgba(0,0,0,.26); }
:global(.dark .scheme3-stripe-popup-amount),:global(.dark .scheme3-stripe-popup-success-mark),:global(.dark .scheme3-stripe-popup-close) { color: #8fc2a5; }
:global(.dark .scheme3-stripe-popup-order) { color: #aaa69a; }
:global(.dark .scheme3-stripe-popup-error) { border-color: rgba(211,139,121,.32); background: rgba(211,139,121,.12); color: #d38b79; }
:global(.dark .scheme3-stripe-popup-spinner) { border-color: #8fc2a5; border-top-color: transparent; }
</style>

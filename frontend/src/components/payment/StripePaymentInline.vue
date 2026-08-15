<template>
  <div class="scheme3-stripe-inline space-y-4">
    <div v-if="loading" class="flex items-center justify-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-primary-500 border-t-transparent"></div>
    </div>
    <div v-else-if="initError" class="card p-6 text-center">
      <p class="text-sm text-red-600 dark:text-red-400">{{ initError }}</p>
      <button class="btn btn-secondary mt-4" @click="$emit('back')">{{ t('payment.result.backToRecharge') }}</button>
    </div>
    <!-- Success -->
    <template v-else-if="success">
      <div class="card scheme3-stripe-inline-card p-6">
        <div class="flex flex-col items-center space-y-4 py-4">
          <div class="flex h-16 w-16 items-center justify-center rounded-full bg-green-100 dark:bg-green-900/30">
            <Icon name="check" size="lg" class="text-green-500" />
          </div>
          <p class="text-lg font-bold text-gray-900 dark:text-white">{{ t('payment.result.success') }}</p>
          <div class="scheme3-stripe-inline-summary w-full p-4">
            <div class="space-y-2 text-sm">
              <div class="flex justify-between">
                <span class="text-gray-500 dark:text-gray-400">{{ t('payment.orders.orderId') }}</span>
                <span class="font-medium text-gray-900 dark:text-white">#{{ orderId }}</span>
              </div>
              <div v-if="amount > 0" class="flex justify-between">
                <span class="text-gray-500 dark:text-gray-400">{{ t('payment.orders.amount') }}</span>
                <span class="font-medium text-gray-900 dark:text-white">{{ creditedAmountSymbol }}{{ amount.toFixed(2) }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-500 dark:text-gray-400">{{ t('payment.orders.payAmount') }}</span>
                <span class="font-medium text-gray-900 dark:text-white">{{ paymentAmountSymbol }}{{ payAmount.toFixed(2) }}</span>
              </div>
            </div>
          </div>
          <button class="btn btn-primary" @click="$emit('done')">{{ t('common.confirm') }}</button>
        </div>
      </div>
    </template>
    <template v-else>
      <!-- Amount -->
      <div class="card scheme3-stripe-inline-card overflow-hidden">
        <div class="scheme3-stripe-inline-amount px-6 py-5 text-center">
          <p class="text-sm font-medium">{{ t('payment.actualPay') }}</p>
          <p class="mt-1 text-3xl font-bold">{{ paymentAmountSymbol }}{{ payAmount.toFixed(2) }}</p>
        </div>
      </div>
      <!-- Stripe Payment Element -->
      <div class="card scheme3-stripe-inline-card p-6">
        <div ref="stripeMount" class="min-h-[200px]"></div>
        <p v-if="error" class="mt-4 text-sm text-red-600 dark:text-red-400">{{ error }}</p>
        <button class="btn scheme3-stripe-inline-pay mt-6 w-full py-3 text-base" :disabled="submitting || !ready" @click="handlePay">
          <span v-if="submitting" class="flex items-center justify-center gap-2">
            <span class="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"></span>
            {{ t('common.processing') }}
          </span>
          <span v-else>{{ t('payment.stripePay') }}</span>
        </button>
      </div>
      <!-- Cancel order -->
      <button class="btn btn-secondary w-full" :disabled="cancelling" @click="handleCancel">
        {{ cancelling ? t('common.processing') : t('payment.qr.cancelOrder') }}
      </button>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { paymentAPI } from '@/api/payment'
import { useAppStore } from '@/stores'
import { getPaymentPopupFeatures } from '@/components/payment/providerConfig'
import { currencySymbol } from '@/components/payment/currency'
import type { Stripe, StripeElements } from '@stripe/stripe-js'
import Icon from '@/components/icons/Icon.vue'

// Stripe payment methods that open a popup (redirect or QR code)
const POPUP_METHODS = new Set(['alipay', 'wechat_pay'])

const props = defineProps<{
  orderId: number
  amount: number
  clientSecret: string
  orderType?: 'balance' | 'subscription'
  publishableKey: string
  payAmount: number
  currency?: string
}>()

const emit = defineEmits<{ success: []; done: []; back: []; redirect: [orderId: number, payUrl: string] }>()

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()

const stripeMount = ref<HTMLElement | null>(null)
const loading = ref(true)
const initError = ref('')
const error = ref('')
const submitting = ref(false)
const cancelling = ref(false)
const success = ref(false)
const ready = ref(false)
const selectedType = ref('')
const creditedAmountSymbol = currencySymbol('USD')
const paymentAmountSymbol = computed(() => currencySymbol(props.currency))

let stripeInstance: Stripe | null = null
let elementsInstance: StripeElements | null = null

onMounted(async () => {
  try {
    const { loadStripe } = await import('@stripe/stripe-js/pure')
    const stripe = await loadStripe(props.publishableKey)
    if (!stripe) { initError.value = t('payment.stripeLoadFailed'); return }

    stripeInstance = stripe
    loading.value = false
    await nextTick()
    if (!stripeMount.value) return

    const isDark = document.documentElement.classList.contains('dark')
    const elements = stripe.elements({
      clientSecret: props.clientSecret,
      appearance: { theme: isDark ? 'night' : 'stripe', variables: { borderRadius: '8px' } },
    })
    elementsInstance = elements
    const paymentElement = elements.create('payment', {
      layout: 'tabs',
      paymentMethodOrder: ['alipay', 'wechat_pay', 'card', 'link'],
    } as Record<string, unknown>)
    paymentElement.mount(stripeMount.value)
    paymentElement.on('ready', () => { ready.value = true })
    paymentElement.on('change', (event: { value: { type: string } }) => {
      selectedType.value = event.value.type
    })
  } catch (err: unknown) {
    initError.value = extractI18nErrorMessage(err, t, 'payment.errors', t('payment.stripeLoadFailed'))
  } finally {
    loading.value = false
  }
})

async function handlePay() {
  if (!stripeInstance || !elementsInstance || submitting.value) return

  // Alipay / WeChat Pay: open popup for redirect or QR display
  if (POPUP_METHODS.has(selectedType.value)) {
    const popupUrl = router.resolve({
      path: '/payment/stripe-popup',
      query: {
        order_id: String(props.orderId),
        method: selectedType.value,
        amount: String(props.payAmount),
      },
    }).href
    const popup = window.open(popupUrl, 'paymentPopup', getPaymentPopupFeatures())

    const onReady = (event: MessageEvent) => {
      if (event.source !== popup || event.data?.type !== 'STRIPE_POPUP_READY') return
      window.removeEventListener('message', onReady)
      popup?.postMessage({
        type: 'STRIPE_POPUP_INIT',
        clientSecret: props.clientSecret,
        publishableKey: props.publishableKey,
      }, window.location.origin)
    }
    window.addEventListener('message', onReady)

    emit('redirect', props.orderId, popupUrl)
    return
  }

  // Card / Link: confirm inline
  submitting.value = true
  error.value = ''
  try {
    const { error: stripeError } = await stripeInstance.confirmPayment({
      elements: elementsInstance,
      confirmParams: {
        return_url: window.location.origin + '/payment/result?order_id=' + props.orderId + '&status=success',
      },
      redirect: 'if_required',
    })
    if (stripeError) {
      error.value = stripeError.message || t('payment.result.failed')
    } else {
      success.value = true
      emit('success')
    }
  } catch (err: unknown) {
    error.value = extractI18nErrorMessage(err, t, 'payment.errors', t('payment.result.failed'))
  } finally {
    submitting.value = false
  }
}

async function handleCancel() {
  if (!props.orderId || cancelling.value) return
  cancelling.value = true
  try {
    await paymentAPI.cancelOrder(props.orderId)
    emit('back')
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    cancelling.value = false
  }
}
</script>

<style scoped>
.scheme3-stripe-inline { --scheme3-stripe-inline-card: #fffefa; --scheme3-stripe-inline-line: #d8d2c3; --scheme3-stripe-inline-ink: #27251f; }
.scheme3-stripe-inline :deep(.card),.scheme3-stripe-inline-card { border: 1px solid var(--scheme3-stripe-inline-line); border-radius: 8px; background: var(--scheme3-stripe-inline-card); box-shadow: 0 10px 24px rgba(54,48,34,.06); }
.scheme3-stripe-inline-amount { border-left: 3px solid #b7791f; background: #f1eee6; color: var(--scheme3-stripe-inline-ink); }
.scheme3-stripe-inline-amount p:first-child { color: #777266; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: .66rem; letter-spacing: .06em; }
.scheme3-stripe-inline-amount p:last-child { color: #1e5c42; font-family: Georgia, 'Times New Roman', serif; }
.scheme3-stripe-inline-summary { border: 1px solid var(--scheme3-stripe-inline-line); border-radius: 7px; background: #f8f6ef; }
.scheme3-stripe-inline :deep(.btn-secondary) { border-color: var(--scheme3-stripe-inline-line); border-radius: 7px; background: var(--scheme3-stripe-inline-card); color: var(--scheme3-stripe-inline-ink); }
.scheme3-stripe-inline-pay { border-radius: 7px; background: #1e5c42; color: #fffefa; box-shadow: 0 8px 16px rgba(30,92,66,.15); }
.scheme3-stripe-inline-pay:hover { background: #174a35; }
.scheme3-stripe-inline-pay:disabled { background: #8e887b; box-shadow: none; }

:global(.dark .scheme3-stripe-inline) { --scheme3-stripe-inline-card: #24231f; --scheme3-stripe-inline-line: #47443a; --scheme3-stripe-inline-ink: #f4f2ec; }
:global(.dark .scheme3-stripe-inline-amount) { border-color: #d6a65d; background: #2b2924; }
:global(.dark .scheme3-stripe-inline-amount p:first-child) { color: #aaa69a; }
:global(.dark .scheme3-stripe-inline-amount p:last-child) { color: #8fc2a5; }
:global(.dark .scheme3-stripe-inline-summary) { border-color: #47443a; background: #2b2924; }
:global(.dark .scheme3-stripe-inline-pay) { background: #8fc2a5; color: #1b1b18; }
:global(.dark .scheme3-stripe-inline-pay:hover) { background: #a7d2b7; }
:global(.dark .scheme3-stripe-inline-pay:disabled) { background: #5b625a; color: #d4d0c6; }
</style>

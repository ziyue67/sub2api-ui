<template>
  <div class="scheme3-payment-method-selector">
    <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
      {{ t('payment.paymentMethod') }}
    </label>
    <div
      data-testid="payment-method-grid"
      class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4"
    >
      <button
        v-for="method in sortedMethods"
        :key="method.type"
        type="button"
        :title="methodLabel(method)"
        :disabled="!method.available"
        :class="[
          'scheme3-payment-method relative flex h-[60px] min-w-0 flex-col items-center justify-center border px-3 transition-all',
          !method.available
            ? 'scheme3-payment-method-disabled cursor-not-allowed opacity-50'
            : selected === method.type
              ? methodSelectedClass(method.type)
              : 'scheme3-payment-method-idle',
        ]"
        @click="method.available && emit('select', method.type)"
      >
        <span class="flex w-full min-w-0 items-center justify-center gap-2">
          <img :src="methodIcon(method.type)" :alt="methodLabel(method)" class="h-7 w-7 shrink-0 object-contain" />
          <span class="flex min-w-0 flex-col items-start leading-none">
            <span data-testid="payment-method-label" class="block w-full truncate text-base font-semibold">
              {{ methodLabel(method) }}
            </span>
            <span
              v-if="method.fee_rate > 0"
              class="text-[10px] tracking-wide text-gray-500 dark:text-dark-400"
            >
              {{ t('payment.fee') }} {{ method.fee_rate }}%
            </span>
          </span>
        </span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { METHOD_ORDER, isBuiltInAlipayMethod, isBuiltInWxpayMethod } from './providerConfig'
import alipayIcon from '@/assets/icons/alipay.svg'
import wxpayIcon from '@/assets/icons/wxpay.svg'
import stripeIcon from '@/assets/icons/stripe.svg'
import airwallexIcon from '@/assets/icons/airwallex.svg'
import paymentIcon from '@/assets/icons/payment.svg'

export interface PaymentMethodOption {
  type: string
  display_name?: string
  fee_rate: number
  available: boolean
}

const props = defineProps<{
  methods: PaymentMethodOption[]
  selected: string
}>()

const emit = defineEmits<{
  select: [type: string]
}>()

const { t } = useI18n()

const METHOD_ICONS: Record<string, string> = {
  alipay: alipayIcon,
  wxpay: wxpayIcon,
  stripe: stripeIcon,
  airwallex: airwallexIcon,
  credit_card: paymentIcon,
}

const sortedMethods = computed(() => {
  const order: readonly string[] = METHOD_ORDER
  return [...props.methods].sort((a, b) => {
    const ai = order.indexOf(a.type)
    const bi = order.indexOf(b.type)
    return (ai === -1 ? 999 : ai) - (bi === -1 ? 999 : bi)
  })
})

function methodIcon(type: string): string {
  if (isBuiltInAlipayMethod(type)) return METHOD_ICONS.alipay
  if (isBuiltInWxpayMethod(type)) return METHOD_ICONS.wxpay
  if (type === 'airwallex') return METHOD_ICONS.airwallex
  return METHOD_ICONS[type] || paymentIcon
}

function methodLabel(method: PaymentMethodOption): string {
  return method.display_name || t(`payment.methods.${method.type}`, method.type)
}

function methodSelectedClass(_type: string): string {
  return 'scheme3-payment-method-selected'
}
</script>

<style scoped>
.scheme3-payment-method-selector { color: #27251f; }
.scheme3-payment-method { border-color: #d8d2c3; border-radius: 7px; background: #fffefa; color: #655f53; transition: border-color 150ms ease, background 150ms ease, color 150ms ease, transform 150ms ease, box-shadow 150ms ease; }
.scheme3-payment-method-idle:hover { border-color: rgba(30,92,66,.35); background: #f8f5ed; color: #27251f; transform: translateY(-1px); }
.scheme3-payment-method-selected { border-color: rgba(30,92,66,.45); background: rgba(30,92,66,.075); color: #1e5c42; box-shadow: inset 3px 0 0 #1e5c42, 0 4px 10px rgba(54,48,34,.06); }
.scheme3-payment-method-disabled { border-color: #e4dfd4; background: #f1eee6; color: #aaa69a; }
.scheme3-payment-method:focus-visible { outline: 3px solid rgba(30,92,66,.18); outline-offset: 2px; }

:global(.dark .scheme3-payment-method-selector) { color: #f4f2ec; }
:global(.dark .scheme3-payment-method) { border-color: #47443a; background: #24231f; color: #c4c0b6; }
:global(.dark .scheme3-payment-method-idle:hover) { border-color: rgba(143,194,165,.38); background: #2b2924; color: #f4f2ec; }
:global(.dark .scheme3-payment-method-selected) { border-color: rgba(143,194,165,.42); background: rgba(143,194,165,.1); color: #8fc2a5; box-shadow: inset 3px 0 0 #8fc2a5, 0 4px 10px rgba(0,0,0,.14); }
:global(.dark .scheme3-payment-method-disabled) { border-color: #3b3932; background: #2b2924; color: #777266; }
</style>

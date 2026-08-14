<template>
  <AppLayout>
    <section class="scheme3-orders space-y-4">
      <header class="scheme3-orders-header">
        <div>
          <p class="scheme3-orders-kicker">结算记录 / 交易账本</p>
          <h1>订单记录</h1>
        </div>
        <div class="scheme3-orders-ledger" aria-label="订单概览">
          <span><strong>{{ orders.length }}</strong><small>本页订单</small></span>
          <span><strong>{{ completedOrderCount }}</strong><small>已完成</small></span>
          <span><strong>{{ pendingOrderCount }}</strong><small>处理中</small></span>
          <span><strong>${{ creditedTotal.toFixed(2) }}</strong><small>本页入账</small></span>
        </div>
      </header>

      <!-- Filters -->
      <div class="scheme3-orders-filter p-4">
        <div class="flex flex-wrap items-center gap-3">
          <Select v-model="currentFilter" :options="statusFilters" class="w-36" @change="fetchOrders" />
          <div class="flex flex-1 items-center justify-end gap-2">
            <button @click="fetchOrders" :disabled="loading" class="btn btn-secondary" :title="t('common.refresh')">
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
            <button class="btn btn-primary" @click="router.push('/purchase')">{{ t('payment.result.backToRecharge') }}</button>
          </div>
        </div>
      </div>

      <!-- Table -->
      <div class="scheme3-orders-table">
      <OrderTable :orders="orders" :loading="loading">
        <template #actions="{ row }">
          <div class="flex items-center gap-2">
            <button v-if="row.status === 'PENDING'" @click="handleCancel(row.id)" class="scheme3-order-row-action scheme3-order-row-cancel inline-flex items-center gap-1 px-2 py-1 text-xs font-medium">
              <Icon name="x" size="sm" />
              <span>{{ t('payment.orders.cancel') }}</span>
            </button>
            <button v-if="canRequestRefund(row)" @click="openRefundDialog(row)" class="scheme3-order-row-action scheme3-order-row-refund inline-flex items-center gap-1 px-2 py-1 text-xs font-medium">
              <Icon name="dollar" size="sm" />
              <span>{{ t('payment.orders.requestRefund') }}</span>
            </button>
          </div>
        </template>
      </OrderTable>
      </div>

      <!-- Pagination -->
      <Pagination
        v-if="pagination.total > 0"
        :page="pagination.page"
        :total="pagination.total"
        :page-size="pagination.page_size"
        @update:page="handlePageChange"
        @update:pageSize="handlePageSizeChange"
      />
    </section>

    <!-- Cancel Confirm Dialog -->
    <BaseDialog :show="!!cancelTargetId" :title="t('payment.orders.cancel')" width="narrow" content-class="scheme3-orders-dialog" @close="cancelTargetId = null">
      <p class="text-sm text-gray-600 dark:text-gray-300">{{ t('payment.confirmCancel') }}</p>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button class="btn btn-secondary" @click="cancelTargetId = null">{{ t('common.cancel') }}</button>
          <button class="btn btn-danger" :disabled="actionLoading" @click="confirmCancel">{{ actionLoading ? t('common.processing') : t('payment.orders.cancel') }}</button>
        </div>
      </template>
    </BaseDialog>

    <!-- Refund Dialog -->
    <BaseDialog :show="!!refundTarget" :title="t('payment.orders.requestRefund')" content-class="scheme3-orders-dialog" @close="refundTarget = null">
      <div v-if="refundTarget" class="space-y-4">
        <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-800">
          <div class="flex justify-between text-sm">
            <span class="text-gray-500 dark:text-gray-400">{{ t('payment.orders.orderId') }}</span>
            <span class="font-mono text-gray-900 dark:text-white">#{{ refundTarget.id }}</span>
          </div>
          <div class="mt-2 flex justify-between text-sm">
            <span class="text-gray-500 dark:text-gray-400">{{ t('payment.orders.amount') }}</span>
            <span class="text-gray-900 dark:text-white">${{ refundTarget.amount.toFixed(2) }}</span>
          </div>
        </div>
        <div>
          <label class="input-label">{{ t('payment.refundReason') }}</label>
          <textarea v-model="refundReason" rows="3" class="input mt-1 w-full" :placeholder="t('payment.refundReasonPlaceholder')" />
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button class="btn btn-secondary" @click="refundTarget = null">{{ t('common.cancel') }}</button>
          <button class="btn btn-primary" :disabled="actionLoading || !refundReason.trim()" @click="confirmRefund">{{ actionLoading ? t('common.processing') : t('payment.orders.requestRefund') }}</button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores'
import { paymentAPI } from '@/api/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import type { PaymentOrder } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import OrderTable from '@/components/payment/OrderTable.vue'

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()

const loading = ref(false)
const actionLoading = ref(false)
const orders = ref<PaymentOrder[]>([])
const refundEligibleProviders = ref<Set<string>>(new Set())
const currentFilter = ref('')
const cancelTargetId = ref<number | null>(null)
const refundTarget = ref<PaymentOrder | null>(null)
const refundReason = ref('')
const pagination = reactive({ page: 1, page_size: 20, total: 0 })
const completedOrderCount = computed(() => orders.value.filter((order) => ['COMPLETED', 'PAID'].includes(order.status)).length)
const pendingOrderCount = computed(() => orders.value.filter((order) => ['PENDING', 'RECHARGING'].includes(order.status)).length)
const creditedTotal = computed(() =>
  orders.value
    .filter((order) => ['COMPLETED', 'PAID'].includes(order.status))
    .reduce((total, order) => total + Number(order.amount || 0), 0)
)

const statusFilters = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'PENDING', label: t('payment.status.pending') },
  { value: 'COMPLETED', label: t('payment.status.completed') },
  { value: 'FAILED', label: t('payment.status.failed') },
  { value: 'REFUNDED', label: t('payment.status.refunded') },
])

async function fetchOrders() {
  loading.value = true
  try {
    const res = await paymentAPI.getMyOrders({
      page: pagination.page,
      page_size: pagination.page_size,
      status: currentFilter.value || undefined,
    })
    orders.value = res.data.items || []
    pagination.total = res.data.total || 0
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

function handlePageChange(page: number) { pagination.page = page; fetchOrders() }
function handlePageSizeChange(size: number) { pagination.page_size = size; pagination.page = 1; fetchOrders() }

function handleCancel(orderId: number) { cancelTargetId.value = orderId }

async function confirmCancel() {
  if (!cancelTargetId.value) return
  actionLoading.value = true
  try {
    await paymentAPI.cancelOrder(cancelTargetId.value)
    appStore.showSuccess(t('common.success'))
    cancelTargetId.value = null
    await fetchOrders()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    actionLoading.value = false
  }
}

function openRefundDialog(order: PaymentOrder) { refundTarget.value = order; refundReason.value = '' }

async function confirmRefund() {
  if (!refundTarget.value || !refundReason.value.trim()) return
  actionLoading.value = true
  try {
    await paymentAPI.requestRefund(refundTarget.value.id, { reason: refundReason.value.trim() })
    appStore.showSuccess(t('common.success'))
    refundTarget.value = null
    refundReason.value = ''
    await fetchOrders()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    actionLoading.value = false
  }
}

function canRequestRefund(order: PaymentOrder): boolean {
  if (order.status !== 'COMPLETED') return false
  if (!order.provider_instance_id) return false
  return refundEligibleProviders.value.has(order.provider_instance_id)
}

async function loadRefundEligibility() {
  try {
    const res = await paymentAPI.getRefundEligibleProviders()
    refundEligibleProviders.value = new Set(res.data.provider_instance_ids || [])
  } catch { /* ignore — default to hiding refund button */ }
}

onMounted(() => { fetchOrders(); loadRefundEligibility() })
</script>

<style scoped>
.scheme3-orders { --orders-card: #fffefa; --orders-ink: #27251f; --orders-muted: #777266; --orders-line: #d8d2c3; color: var(--orders-ink); }
.scheme3-orders-header { display: flex; align-items: end; justify-content: space-between; gap: 1.25rem; border-bottom: 1px solid var(--orders-line); padding: .1rem 0 1rem; }
.scheme3-orders-kicker { margin: 0; color: var(--orders-muted); font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: .61rem; font-weight: 800; letter-spacing: .1em; }
.scheme3-orders h1 { margin: .34rem 0 0; color: var(--orders-ink); font-family: Georgia, 'Times New Roman', serif; font-size: clamp(1.55rem, 2.6vw, 2.1rem); font-weight: 500; letter-spacing: 0; }
.scheme3-orders-ledger { display: flex; flex-wrap: wrap; justify-content: flex-end; border: 1px solid var(--orders-line); border-radius: 7px; background: var(--orders-card); }
.scheme3-orders-ledger span { display: grid; min-width: 4.7rem; gap: .06rem; border-right: 1px solid var(--orders-line); padding: .48rem .7rem; text-align: right; }
.scheme3-orders-ledger span:last-child { border-right: 0; }
.scheme3-orders-ledger strong { color: #1e5c42; font-family: Georgia, 'Times New Roman', serif; font-size: 1.02rem; font-weight: 600; line-height: 1.1; }
.scheme3-orders-ledger small { color: var(--orders-muted); font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: .52rem; font-weight: 700; letter-spacing: .04em; }
.scheme3-orders-filter { border: 1px solid var(--orders-line); border-radius: 8px; background: var(--orders-card); box-shadow: 0 10px 22px rgba(54,48,34,.05); }
.scheme3-orders-filter :deep(.btn) { min-height: 2.35rem; border-radius: 7px; }
.scheme3-orders-filter :deep(.btn-primary) { color: #fffefa; }
.scheme3-orders-table { overflow: hidden; border: 1px solid var(--orders-line); border-radius: 8px; background: var(--orders-card); box-shadow: 0 11px 24px rgba(54,48,34,.06); }
.scheme3-orders-table :deep(.table-wrapper) { overflow: auto; }
.scheme3-orders-table :deep(thead),.scheme3-orders-table :deep(.table-header),.scheme3-orders-table :deep(.sticky-header-cell) { background: #f1eee6; }
.scheme3-orders-table :deep(th) { padding-top: .7rem; padding-bottom: .7rem; color: var(--orders-muted); font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: .59rem; font-weight: 800; letter-spacing: .055em; }
.scheme3-orders-table :deep(td) { padding-top: .78rem; padding-bottom: .78rem; border-color: rgba(216,210,195,.78); }
.scheme3-orders-table :deep(tbody tr) { transition: background-color 150ms ease; }
.scheme3-orders-table :deep(tbody tr:hover) { background: rgba(30,92,66,.045) !important; }
.scheme3-orders-table :deep(tbody .sticky-col) { background: #fffefa; }
.scheme3-orders-table :deep(tbody tr:hover .sticky-col) { background: #f6f4ed; }
.scheme3-orders-table :deep(.bg-green-100) { border: 1px solid rgba(30,92,66,.22); background: rgba(30,92,66,.1) !important; color: #1e5c42 !important; }
.scheme3-orders-table :deep(.bg-yellow-100) { border: 1px solid rgba(183,121,31,.25); background: rgba(183,121,31,.1) !important; color: #a56613 !important; }
.scheme3-orders-table :deep(.bg-blue-100),.scheme3-orders-table :deep(.bg-purple-100),.scheme3-orders-table :deep(.bg-gray-100) { border: 1px solid var(--orders-line); background: #f1eee6 !important; color: #655f53 !important; }
.scheme3-orders-table :deep(.bg-orange-100) { border: 1px solid rgba(183,121,31,.25); background: rgba(183,121,31,.1) !important; color: #a56613 !important; }
.scheme3-orders-table :deep(.bg-red-100) { border: 1px solid rgba(158,77,61,.25); background: rgba(158,77,61,.1) !important; color: #9e4d3d !important; }
.scheme3-orders-table :deep(td .rounded-full) { border-radius: 999px; }
.scheme3-order-row-action { border: 1px solid transparent; border-radius: 6px; transition: background-color 150ms ease, border-color 150ms ease, color 150ms ease, transform 150ms ease; }
.scheme3-order-row-action:active { transform: scale(.96); }
.scheme3-order-row-cancel { color: #a56613; }.scheme3-order-row-cancel:hover { border-color: rgba(183,121,31,.25); background: rgba(183,121,31,.1); }
.scheme3-order-row-refund { color: #655f53; }.scheme3-order-row-refund:hover { border-color: rgba(101,95,83,.22); background: #f1eee6; }

:global(.scheme3-orders-dialog) { border: 1px solid #d8d2c3 !important; border-radius: 8px !important; background: #fffefa !important; box-shadow: 0 20px 42px rgba(54,48,34,.16) !important; }
:global(.scheme3-orders-dialog .modal-header),:global(.scheme3-orders-dialog .modal-footer) { border-color: #d8d2c3 !important; background: #f8f6ef !important; }
:global(.scheme3-orders-dialog .modal-title) { color: #27251f !important; font-family: Georgia, 'Times New Roman', serif; font-weight: 500; }
:global(.scheme3-orders-dialog .input),:global(.scheme3-orders-dialog textarea) { border-color: #d8d2c3 !important; border-radius: 7px !important; background: #fffefa !important; color: #27251f !important; }
:global(.scheme3-orders-dialog .input:focus),:global(.scheme3-orders-dialog textarea:focus) { border-color: #1e5c42 !important; box-shadow: 0 0 0 3px rgba(30,92,66,.12) !important; }
:global(.scheme3-orders-dialog .input-label) { color: #27251f !important; }
:global(.scheme3-orders-dialog .btn-primary) { border-radius: 7px; background: #1e5c42; color: #fffefa; }
:global(.scheme3-orders-dialog .btn-secondary) { border-color: #d8d2c3; border-radius: 7px; background: #fffefa; color: #27251f; }
:global(.scheme3-orders-dialog .btn-danger) { border-radius: 7px; background: #9e4d3d; color: #fffefa; }
:global(.scheme3-orders-dialog .bg-gray-50) { border: 1px solid #d8d2c3; border-radius: 7px; background: #f1eee6 !important; }

:global(.dark .scheme3-orders) { --orders-card: #24231f; --orders-ink: #f4f2ec; --orders-muted: #aaa69a; --orders-line: #47443a; }
:global(.dark .scheme3-orders-ledger strong) { color: #8fc2a5; }
:global(.dark .scheme3-orders-table),:global(.dark .scheme3-orders-table .table-wrapper),:global(.dark .scheme3-orders-table tbody) { background: #24231f !important; }
:global(.dark .scheme3-orders-table thead),:global(.dark .scheme3-orders-table .table-header),:global(.dark .scheme3-orders-table .sticky-header-cell) { background: #2b2924 !important; }
:global(.dark .scheme3-orders-table td) { border-color: rgba(71,68,58,.86); }
:global(.dark .scheme3-orders-table tbody tr:hover) { background: rgba(143,194,165,.07) !important; }
:global(.dark .scheme3-orders-table tbody .sticky-col) { background: #24231f !important; }
:global(.dark .scheme3-orders-table tbody tr:hover .sticky-col) { background: #2b2924 !important; }
:global(.dark .scheme3-orders-table .bg-green-100) { border-color: rgba(143,194,165,.25); background: rgba(143,194,165,.1) !important; color: #8fc2a5 !important; }
:global(.dark .scheme3-orders-table .bg-yellow-100),:global(.dark .scheme3-orders-table .bg-orange-100) { border-color: rgba(214,166,93,.28); background: rgba(214,166,93,.11) !important; color: #d6a65d !important; }
:global(.dark .scheme3-orders-table .bg-blue-100),:global(.dark .scheme3-orders-table .bg-purple-100),:global(.dark .scheme3-orders-table .bg-gray-100) { border-color: #47443a; background: #2b2924 !important; color: #c4c0b6 !important; }
:global(.dark .scheme3-orders-table .bg-red-100) { border-color: rgba(211,139,121,.3); background: rgba(211,139,121,.11) !important; color: #d38b79 !important; }
:global(.dark .scheme3-order-row-cancel) { color: #d6a65d; }.scheme3-order-row-cancel:hover { background: rgba(214,166,93,.11); }
:global(.dark .scheme3-order-row-refund) { color: #c4c0b6; }.scheme3-order-row-refund:hover { border-color: #47443a; background: #2b2924; }
:global(.dark .scheme3-orders-dialog) { border-color: #47443a !important; background: #24231f !important; box-shadow: 0 20px 42px rgba(0,0,0,.3) !important; }
:global(.dark .scheme3-orders-dialog .modal-header),:global(.dark .scheme3-orders-dialog .modal-footer) { border-color: #47443a !important; background: #1b1b18 !important; }
:global(.dark .scheme3-orders-dialog .modal-title),:global(.dark .scheme3-orders-dialog .input-label) { color: #f4f2ec !important; }
:global(.dark .scheme3-orders-dialog .input),:global(.dark .scheme3-orders-dialog textarea) { border-color: #47443a !important; background: #24231f !important; color: #f4f2ec !important; }
:global(.dark .scheme3-orders-dialog .btn-primary) { background: #8fc2a5; color: #1b1b18; }
:global(.dark .scheme3-orders-dialog .btn-secondary) { border-color: #47443a; background: #24231f; color: #f4f2ec; }
:global(.dark .scheme3-orders-dialog .btn-danger) { background: #d38b79; color: #1b1b18; }
:global(.dark .scheme3-orders-dialog .bg-gray-50) { border-color: #47443a; background: #2b2924 !important; }

@media (max-width: 767px) {
  .scheme3-orders-header { align-items: stretch; flex-direction: column; gap: .8rem; }
  .scheme3-orders-ledger { width: 100%; justify-content: stretch; }
  .scheme3-orders-ledger span { flex: 1 1 45%; min-width: 0; padding: .48rem .42rem; }
  .scheme3-orders-filter > div { align-items: stretch; }
  .scheme3-orders-filter :deep(.btn) { flex: 1 1 auto; }
  .scheme3-orders-table { overflow: visible; border: 0; border-radius: 0; background: transparent; box-shadow: none; }
  .scheme3-orders-table :deep(> .space-y-3 > div) { border: 1px solid var(--orders-line); border-radius: 8px; background: var(--orders-card); box-shadow: 0 8px 18px rgba(54,48,34,.05); }
  .scheme3-orders-table :deep([data-field]) { border-bottom: 1px solid rgba(216,210,195,.7); padding-bottom: .55rem; }
  .scheme3-orders-table :deep([data-field]:last-of-type) { border-bottom: 0; padding-bottom: 0; }
}
</style>

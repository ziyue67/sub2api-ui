<template>
  <AppLayout>
    <div class="scheme3-subscriptions space-y-6">
      <!-- Loading State -->
      <div v-if="loading" class="flex justify-center py-12">
        <div
          class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"
        ></div>
      </div>

      <!-- Empty State -->
      <div v-else-if="subscriptions.length === 0" class="scheme3-subscriptions-empty">
        <div
          class="scheme3-subscriptions-empty-icon mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full"
        >
          <Icon name="creditCard" size="xl" class="text-gray-400" />
        </div>
        <h3 class="mb-2 text-lg font-semibold text-gray-900 dark:text-white">
          {{ t('userSubscriptions.noActiveSubscriptions') }}
        </h3>
        <p class="text-gray-500 dark:text-dark-400">
          {{ t('userSubscriptions.noActiveSubscriptionsDesc') }}
        </p>
      </div>

      <!-- Subscriptions Grid -->
      <div v-else class="scheme3-subscriptions-grid grid gap-6 lg:grid-cols-2">
        <div
          v-for="subscription in subscriptions"
          :key="subscription.id"
          class="scheme3-subscription-card"
        >
          <!-- Header -->
          <div class="scheme3-subscription-card-header">
            <div class="flex items-center gap-3">
              <div :class="['scheme3-subscription-platform-dot', platformAccentDotClass(subscription.group?.platform || '')]" />
              <div>
                <div class="flex items-center gap-2">
                  <h3 class="font-semibold text-gray-900 dark:text-white">
                    {{ subscription.group?.name || `Group #${subscription.group_id}` }}
                  </h3>
                  <span class="scheme3-subscription-platform">
                    {{ platformLabel(subscription.group?.platform || '') }}
                  </span>
                </div>
                <p v-if="subscription.group?.description" class="mt-0.5 text-xs text-gray-500 dark:text-dark-400">
                  {{ subscription.group.description }}
                </p>
                <div class="mt-1 flex flex-wrap gap-x-3 gap-y-1 text-[11px] text-gray-400 dark:text-gray-500">
                  <span>{{ t('payment.planCard.rate') }}: ×{{ subscription.group?.rate_multiplier ?? 1 }}</span>
                  <span v-if="subscriptionHasPeakRate(subscription)" class="text-amber-700 dark:text-amber-300">
                    {{ t('payment.planCard.peakRate') }}: {{ subscriptionPeakRateLabel(subscription) }}
                  </span>
                </div>
              </div>
            </div>
            <div class="flex items-center gap-2">
              <span
                :class="[
                  'scheme3-subscription-status',
                  subscription.status === 'active'
                    ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300'
                    : subscription.status === 'expired'
                      ? 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-400'
                      : 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300'
                ]"
              >
                {{ t(`userSubscriptions.status.${subscription.status}`) }}
              </span>
              <button
                v-if="subscription.status === 'active'"
                class="scheme3-subscription-renew"
                @click="router.push({ path: '/purchase', query: { tab: 'subscription', group: String(subscription.group_id) } })"
              >
                {{ t('payment.renewNow') }}
              </button>
            </div>
          </div>

          <!-- Usage Progress -->
          <div class="scheme3-subscription-card-body">
            <!-- Expiration Info -->
            <div v-if="subscription.expires_at" class="flex items-center justify-between text-sm">
              <span class="text-gray-500 dark:text-dark-400">{{
                t('userSubscriptions.expires')
              }}</span>
              <span :class="getExpirationClass(subscription.expires_at)">
                {{ formatExpirationDate(subscription.expires_at) }}
              </span>
            </div>
            <div v-else class="flex items-center justify-between text-sm">
              <span class="text-gray-500 dark:text-dark-400">{{
                t('userSubscriptions.expires')
              }}</span>
              <span class="text-gray-700 dark:text-gray-300">{{
                t('userSubscriptions.noExpiration')
              }}</span>
            </div>

            <!-- Daily Usage -->
            <div v-if="subscription.group?.daily_limit_usd" class="space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('userSubscriptions.daily') }}
                </span>
                <span class="text-sm text-gray-500 dark:text-dark-400">
                  ${{ (subscription.daily_usage_usd || 0).toFixed(2) }} / ${{
                    subscription.group.daily_limit_usd.toFixed(2)
                  }}
                </span>
              </div>
              <div class="scheme3-subscription-progress-track">
                <div
                  class="scheme3-subscription-progress-bar"
                  :class="
                    getProgressBarClass(
                      subscription.daily_usage_usd,
                      subscription.group.daily_limit_usd
                    )
                  "
                  :style="{
                    width: getProgressWidth(
                      subscription.daily_usage_usd,
                      subscription.group.daily_limit_usd
                    )
                  }"
                ></div>
              </div>
              <p
                v-if="subscription.daily_window_start"
                class="text-xs text-gray-500 dark:text-dark-400"
              >
                {{ formatDailyUsageWindow(subscription) }}
              </p>
            </div>

            <!-- Weekly Usage -->
            <div v-if="subscription.group?.weekly_limit_usd" class="space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('userSubscriptions.weekly') }}
                </span>
                <span class="text-sm text-gray-500 dark:text-dark-400">
                  ${{ (subscription.weekly_usage_usd || 0).toFixed(2) }} / ${{
                    subscription.group.weekly_limit_usd.toFixed(2)
                  }}
                </span>
              </div>
              <div class="scheme3-subscription-progress-track">
                <div
                  class="scheme3-subscription-progress-bar"
                  :class="
                    getProgressBarClass(
                      subscription.weekly_usage_usd,
                      subscription.group.weekly_limit_usd
                    )
                  "
                  :style="{
                    width: getProgressWidth(
                      subscription.weekly_usage_usd,
                      subscription.group.weekly_limit_usd
                    )
                  }"
                ></div>
              </div>
              <p
                v-if="subscription.weekly_window_start"
                class="text-xs text-gray-500 dark:text-dark-400"
              >
                {{
                  t('userSubscriptions.resetIn', {
                    time: formatResetTime(subscription.weekly_window_start, 168)
                  })
                }}
              </p>
            </div>

            <!-- Monthly Usage -->
            <div v-if="subscription.group?.monthly_limit_usd" class="space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('userSubscriptions.monthly') }}
                </span>
                <span class="text-sm text-gray-500 dark:text-dark-400">
                  ${{ (subscription.monthly_usage_usd || 0).toFixed(2) }} / ${{
                    subscription.group.monthly_limit_usd.toFixed(2)
                  }}
                </span>
              </div>
              <div class="scheme3-subscription-progress-track">
                <div
                  class="scheme3-subscription-progress-bar"
                  :class="
                    getProgressBarClass(
                      subscription.monthly_usage_usd,
                      subscription.group.monthly_limit_usd
                    )
                  "
                  :style="{
                    width: getProgressWidth(
                      subscription.monthly_usage_usd,
                      subscription.group.monthly_limit_usd
                    )
                  }"
                ></div>
              </div>
              <p
                v-if="subscription.monthly_window_start"
                class="text-xs text-gray-500 dark:text-dark-400"
              >
                {{
                  t('userSubscriptions.resetIn', {
                    time: formatResetTime(subscription.monthly_window_start, 720)
                  })
                }}
              </p>
            </div>

            <!-- No limits configured - Unlimited badge -->
            <div
              v-if="
                !subscription.group?.daily_limit_usd &&
                !subscription.group?.weekly_limit_usd &&
                !subscription.group?.monthly_limit_usd
              "
              class="scheme3-subscription-unlimited"
            >
              <div class="flex items-center gap-3">
                <span class="text-4xl text-emerald-600 dark:text-emerald-400">∞</span>
                <div>
                  <p class="text-sm font-medium text-emerald-700 dark:text-emerald-300">
                    {{ t('userSubscriptions.unlimited') }}
                  </p>
                  <p class="text-xs text-emerald-600/70 dark:text-emerald-400/70">
                    {{ t('userSubscriptions.unlimitedDesc') }}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import subscriptionsAPI from '@/api/subscriptions'
import type { UserSubscription } from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTimeToMinute } from '@/utils/format'
import { hasPeakRate, formatPeakRateWindow, serverTimezoneLabel } from '@/utils/peak-rate'
import { platformLabel } from '@/utils/platformColors'
import {
  getExpirationDateRelation,
  getRemainingDurationParts,
  isOneTimeDailyQuota,
  type RemainingDurationParts
} from '@/utils/subscriptionQuota'

function platformAccentDotClass(p: string): string {
  switch (p) {
    case 'anthropic': return 'bg-orange-500'
    case 'openai': return 'bg-emerald-500'
    case 'antigravity': return 'bg-purple-500'
    case 'gemini': return 'bg-blue-500'
    default: return 'bg-gray-400'
  }
}

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()

const subscriptions = ref<UserSubscription[]>([])
const loading = ref(true)

function subscriptionHasPeakRate(subscription: UserSubscription): boolean {
  return hasPeakRate(subscription.group)
}

function subscriptionPeakRateLabel(subscription: UserSubscription): string {
  return formatPeakRateWindow(subscription.group, serverTimezoneLabel(appStore.cachedPublicSettings?.server_utc_offset))
}

async function loadSubscriptions() {
  try {
    loading.value = true
    subscriptions.value = await subscriptionsAPI.getMySubscriptions()
  } catch (error) {
    console.error('Failed to load subscriptions:', error)
    appStore.showError(t('userSubscriptions.failedToLoad'))
  } finally {
    loading.value = false
  }
}

function getProgressWidth(used: number | undefined, limit: number | null | undefined): string {
  if (!limit || limit === 0) return '0%'
  const percentage = Math.min(((used || 0) / limit) * 100, 100)
  return `${percentage}%`
}

function getProgressBarClass(used: number | undefined, limit: number | null | undefined): string {
  if (!limit || limit === 0) return 'scheme3-progress-neutral'
  const percentage = ((used || 0) / limit) * 100
  if (percentage >= 90) return 'scheme3-progress-danger'
  if (percentage >= 70) return 'scheme3-progress-warning'
  return 'scheme3-progress-ok'
}

function formatExpirationDate(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))
  const relation = getExpirationDateRelation(expires, now)

  if (relation === null) return ''

  if (relation === 'expired') {
    return t('userSubscriptions.status.expired')
  }

  const dateStr = formatDateTimeToMinute(expires)

  if (relation === 'today') {
    return `${dateStr} (${t('common.today')})`
  }
  if (relation === 'tomorrow') {
    return `${dateStr} (${t('common.tomorrow')})`
  }

  return t('userSubscriptions.daysRemaining', { days }) + ` (${dateStr})`
}

function getExpirationClass(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))

  if (diff <= 0) return 'text-red-600 dark:text-red-400 font-medium'
  if (days <= 3) return 'text-red-600 dark:text-red-400'
  if (days <= 7) return 'text-orange-600 dark:text-orange-400'
  return 'text-gray-700 dark:text-gray-300'
}

function formatDurationParts(parts: RemainingDurationParts): string {
  if (parts.days > 0) {
    return `${parts.days}d ${parts.hours}h`
  }

  if (parts.hours > 0) {
    return `${parts.hours}h ${parts.minutes}m`
  }

  return `${parts.minutes}m`
}

function formatDailyUsageWindow(subscription: UserSubscription): string {
  if (isOneTimeDailyQuota(subscription) && subscription.expires_at) {
    const parts = getRemainingDurationParts(subscription.expires_at)
    if (!parts) return t('userSubscriptions.windowNotActive')
    return t('userSubscriptions.quotaEndsIn', { time: formatDurationParts(parts) })
  }

  return t('userSubscriptions.resetIn', {
    time: formatResetTime(subscription.daily_window_start, 24)
  })
}

function formatResetTime(windowStart: string | null, windowHours: number): string {
  if (!windowStart) return t('userSubscriptions.windowNotActive')

  const start = new Date(windowStart)
  const end = new Date(start.getTime() + windowHours * 60 * 60 * 1000)
  const parts = getRemainingDurationParts(end)

  return parts ? formatDurationParts(parts) : t('userSubscriptions.windowNotActive')
}

onMounted(() => {
  loadSubscriptions()
})
</script>

<style scoped>
.scheme3-subscriptions { --scheme3-subscriptions-card: #fffefa; --scheme3-subscriptions-ink: #27251f; --scheme3-subscriptions-muted: #777266; --scheme3-subscriptions-line: #d8d2c3; }
.scheme3-subscription-card { overflow: hidden; border: 1px solid var(--scheme3-subscriptions-line); border-radius: 8px; background: var(--scheme3-subscriptions-card); box-shadow: 0 10px 24px rgba(54,48,34,.06); }
.scheme3-subscription-card-header { display: flex; align-items: center; justify-content: space-between; gap: 1rem; border-bottom: 1px solid var(--scheme3-subscriptions-line); padding: 1rem; }
.scheme3-subscription-platform-dot { width: .42rem; height: .42rem; flex-shrink: 0; border-radius: 999px; box-shadow: 0 0 0 4px rgba(30,92,66,.08); }
.scheme3-subscription-platform { border: 1px solid rgba(30,92,66,.22); border-radius: 999px; padding: .18rem .48rem; color: #1e5c42; background: rgba(30,92,66,.07); font-size: .65rem; font-weight: 700; }
.scheme3-subscription-status { border: 1px solid transparent; border-radius: 999px; padding: .2rem .52rem; font-size: .65rem; font-weight: 800; }
.scheme3-subscription-status.bg-emerald-100 { border-color: rgba(30,92,66,.2); background: rgba(30,92,66,.1); color: #1e5c42; }
.scheme3-subscription-status.bg-gray-100 { border-color: var(--scheme3-subscriptions-line); background: #f1eee6; color: var(--scheme3-subscriptions-muted); }
.scheme3-subscription-status.bg-red-100 { border-color: rgba(158,77,61,.25); background: rgba(158,77,61,.1); color: #9e4d3d; }
.scheme3-subscription-renew { border: 0; border-radius: 7px; padding: .42rem .75rem; background: #1e5c42; color: #fffefa; font-size: .68rem; font-weight: 800; transition: background 150ms ease, transform 150ms ease; }
.scheme3-subscription-renew:hover { background: #174a35; }
.scheme3-subscription-renew:active { transform: scale(.97); }
.scheme3-subscription-card-body { display: flex; flex-direction: column; gap: 1rem; padding: 1rem; color: var(--scheme3-subscriptions-ink); }
.scheme3-subscription-card-body > div { border-bottom: 1px solid rgba(216,210,195,.65); padding-bottom: .85rem; }
.scheme3-subscription-card-body > div:last-child { border-bottom: 0; padding-bottom: 0; }
.scheme3-subscription-progress-track { position: relative; height: .38rem; overflow: hidden; border-radius: 999px; background: #e7e2d7; }
.scheme3-subscription-progress-bar { position: absolute; inset: 0 auto 0 0; width: 0; border-radius: inherit; transition: width 300ms ease; }
.scheme3-progress-neutral { background: #8e887b; }
.scheme3-progress-ok { background: #1e5c42; }
.scheme3-progress-warning { background: #b7791f; }
.scheme3-progress-danger { background: #9e4d3d; }
.scheme3-subscription-unlimited { display: flex; align-items: center; justify-content: center; border: 1px solid rgba(30,92,66,.18); border-radius: 7px; padding: 1.15rem; background: rgba(30,92,66,.055); }
.scheme3-subscription-unlimited > div { display: flex; align-items: center; gap: .75rem; }
.scheme3-subscription-unlimited > div > span { color: #1e5c42; font-family: Georgia, 'Times New Roman', serif; font-size: 2rem; line-height: 1; }
.scheme3-subscription-unlimited p:first-child { color: #1e5c42; font-size: .75rem; font-weight: 800; }
.scheme3-subscription-unlimited p:last-child { color: var(--scheme3-subscriptions-muted); font-size: .68rem; }
.scheme3-subscriptions-empty { border: 1px solid var(--scheme3-subscriptions-line); border-radius: 8px; background: var(--scheme3-subscriptions-card); padding: 3rem; text-align: center; }
.scheme3-subscriptions-empty-icon { background: #f1eee6; color: #8e887b; }

:global(.dark .scheme3-subscriptions) { --scheme3-subscriptions-card: #24231f; --scheme3-subscriptions-ink: #f4f2ec; --scheme3-subscriptions-muted: #aaa69a; --scheme3-subscriptions-line: #47443a; }
:global(.dark .scheme3-subscription-platform) { border-color: rgba(143,194,165,.28); background: rgba(143,194,165,.1); color: #8fc2a5; }
:global(.dark .scheme3-subscription-status.bg-emerald-100) { border-color: rgba(143,194,165,.25); background: rgba(143,194,165,.12); color: #8fc2a5; }
:global(.dark .scheme3-subscription-status.bg-gray-100) { background: #2b2924; color: #aaa69a; }
:global(.dark .scheme3-subscription-status.bg-red-100) { border-color: rgba(211,139,121,.3); background: rgba(211,139,121,.12); color: #d38b79; }
:global(.dark .scheme3-subscription-renew) { background: #8fc2a5; color: #1b1b18; }
:global(.dark .scheme3-subscription-renew:hover) { background: #a7d2b7; }
:global(.dark .scheme3-subscription-progress-track) { background: #3b3932; }
:global(.dark .scheme3-progress-ok) { background: #8fc2a5; }
:global(.dark .scheme3-progress-warning) { background: #d6a65d; }
:global(.dark .scheme3-progress-danger) { background: #d38b79; }
:global(.dark .scheme3-subscription-unlimited) { border-color: rgba(143,194,165,.25); background: rgba(143,194,165,.1); }
:global(.dark .scheme3-subscription-unlimited > div > span), :global(.dark .scheme3-subscription-unlimited p:first-child) { color: #8fc2a5; }
:global(.dark .scheme3-subscriptions-empty-icon) { background: #2b2924; color: #aaa69a; }

@media (max-width: 640px) {
  .scheme3-subscription-card-header { align-items: flex-start; flex-direction: column; }
  .scheme3-subscription-card-header > div:last-child { width: 100%; justify-content: space-between; }
}
</style>

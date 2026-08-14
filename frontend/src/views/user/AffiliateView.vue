<template>
  <AppLayout>
    <section class="scheme3-affiliate">
      <div v-if="loading" class="scheme3-affiliate-state" aria-live="polite">
        <div class="scheme3-affiliate-spinner" aria-hidden="true"></div>
        <span>正在整理分润账本</span>
      </div>

      <template v-else-if="detail">
        <header class="scheme3-affiliate-header">
          <div>
            <p class="scheme3-affiliate-kicker">推广中心 / 分润账本</p>
            <h1>{{ t('affiliate.title') }}</h1>
            <p class="scheme3-affiliate-subtitle">{{ t('affiliate.description') }}</p>
          </div>
          <div class="scheme3-affiliate-ledger" aria-label="分润概览">
            <span>
              <strong class="is-accent">{{ formattedRebateRate }}%</strong>
              <small>{{ t('affiliate.stats.rebateRate') }}</small>
            </span>
            <span>
              <strong>{{ formatCount(detail.aff_count) }}</strong>
              <small>{{ t('affiliate.stats.invitedUsers') }}</small>
            </span>
            <span>
              <strong class="is-accent">{{ formatCurrency(detail.aff_quota) }}</strong>
              <small>{{ t('affiliate.stats.availableQuota') }}</small>
            </span>
            <span>
              <strong>{{ formatCurrency(detail.aff_history_quota) }}</strong>
              <small>{{ t('affiliate.stats.totalQuota') }}</small>
            </span>
          </div>
        </header>

        <div class="scheme3-affiliate-layout">
          <section class="scheme3-affiliate-panel scheme3-affiliate-credential-panel">
            <div class="scheme3-affiliate-panel-head">
              <div>
                <p class="scheme3-affiliate-panel-kicker">推广凭证</p>
                <h2>邀请入口</h2>
              </div>
              <Icon name="users" size="lg" class="scheme3-affiliate-panel-icon" />
            </div>

            <div class="scheme3-affiliate-credentials">
              <div class="scheme3-affiliate-credential-row">
                <div class="scheme3-affiliate-credential-label">{{ t('affiliate.yourCode') }}</div>
                <code class="scheme3-affiliate-credential-value">{{ detail.aff_code }}</code>
                <button class="scheme3-affiliate-copy" type="button" @click="copyCode">
                  <Icon name="copy" size="sm" />
                  <span>{{ t('affiliate.copyCode') }}</span>
                </button>
              </div>

              <div class="scheme3-affiliate-credential-row">
                <div class="scheme3-affiliate-credential-label">{{ t('affiliate.inviteLink') }}</div>
                <code class="scheme3-affiliate-credential-value">{{ inviteLink }}</code>
                <button class="scheme3-affiliate-copy" type="button" @click="copyInviteLink">
                  <Icon name="copy" size="sm" />
                  <span>{{ t('affiliate.copyLink') }}</span>
                </button>
              </div>
            </div>

            <div class="scheme3-affiliate-note">
              <p>{{ t('affiliate.tips.title') }}</p>
              <ul>
                <li>1. {{ t('affiliate.tips.line1') }}</li>
                <li>2. {{ t('affiliate.tips.line2', { rate: `${formattedRebateRate}%` }) }}</li>
                <li>3. {{ t('affiliate.tips.line3') }}</li>
                <li v-if="detail.aff_frozen_quota > 0">4. {{ t('affiliate.tips.line4') }}</li>
              </ul>
            </div>
          </section>

          <section class="scheme3-affiliate-panel scheme3-affiliate-transfer-panel">
            <div class="scheme3-affiliate-panel-head">
              <div>
                <p class="scheme3-affiliate-panel-kicker">余额结转</p>
                <h2>{{ t('affiliate.transfer.title') }}</h2>
              </div>
              <Icon name="dollar" size="lg" class="scheme3-affiliate-panel-icon" />
            </div>
            <p class="scheme3-affiliate-transfer-copy">{{ t('affiliate.transfer.description') }}</p>
            <div class="scheme3-affiliate-transfer-amount">
              <span>当前可结转</span>
              <strong>{{ formatCurrency(detail.aff_quota) }}</strong>
            </div>
            <p v-if="detail.aff_frozen_quota > 0" class="scheme3-affiliate-frozen">
              {{ t('affiliate.stats.frozenQuota') }}: {{ formatCurrency(detail.aff_frozen_quota) }}
            </p>
            <button
              class="scheme3-affiliate-transfer-button"
              :disabled="transferring || detail.aff_quota <= 0"
              @click="transferQuota"
            >
              <Icon v-if="transferring" name="refresh" size="sm" class="animate-spin" />
              <Icon v-else name="dollar" size="sm" />
              <span>{{ transferring ? t('affiliate.transfer.transferring') : t('affiliate.transfer.button') }}</span>
            </button>
            <p v-if="detail.aff_quota <= 0" class="scheme3-affiliate-empty-transfer">
              {{ t('affiliate.transfer.empty') }}
            </p>
          </section>
        </div>

        <section class="scheme3-affiliate-panel scheme3-affiliate-invitees-panel">
          <div class="scheme3-affiliate-panel-head">
            <div>
              <p class="scheme3-affiliate-panel-kicker">邀请明细</p>
              <h2>{{ t('affiliate.invitees.title') }}</h2>
            </div>
            <span class="scheme3-affiliate-invitee-count">{{ formatCount(detail.invitees.length) }} 位</span>
          </div>
          <div v-if="detail.invitees.length === 0" class="scheme3-affiliate-empty-invitees">
            {{ t('affiliate.invitees.empty') }}
          </div>
          <div v-else class="scheme3-affiliate-table-wrap">
            <table class="scheme3-affiliate-table">
              <thead>
                <tr>
                  <th>{{ t('affiliate.invitees.columns.email') }}</th>
                  <th>{{ t('affiliate.invitees.columns.username') }}</th>
                  <th class="is-right">{{ t('affiliate.invitees.columns.rebate') }}</th>
                  <th>{{ t('affiliate.invitees.columns.joinedAt') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in detail.invitees" :key="item.user_id">
                  <td :data-label="t('affiliate.invitees.columns.email')">{{ item.email || '-' }}</td>
                  <td :data-label="t('affiliate.invitees.columns.username')">{{ item.username || '-' }}</td>
                  <td :data-label="t('affiliate.invitees.columns.rebate')" class="is-rebate">{{ formatCurrency(item.total_rebate) }}</td>
                  <td :data-label="t('affiliate.invitees.columns.joinedAt')">{{ formatDateTime(item.created_at) || '-' }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </template>
    </section>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import userAPI from '@/api/user'
import type { UserAffiliateDetail } from '@/types'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { useClipboard } from '@/composables/useClipboard'
import { formatCurrency, formatDateTime } from '@/utils/format'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const { copyToClipboard } = useClipboard()

const loading = ref(true)
const transferring = ref(false)
const detail = ref<UserAffiliateDetail | null>(null)

const inviteLink = computed(() => {
  if (!detail.value) return ''
  if (typeof window === 'undefined') return `/register?aff=${encodeURIComponent(detail.value.aff_code)}`
  return `${window.location.origin}/register?aff=${encodeURIComponent(detail.value.aff_code)}`
})

// Rebate rate is a percentage in the range [0, 100]; backend already clamps it.
// We trim trailing zeros (e.g. 20.00 → "20", 12.50 → "12.5") for a cleaner UI.
const formattedRebateRate = computed(() => {
  const v = detail.value?.effective_rebate_rate_percent ?? 0
  const rounded = Math.round(v * 100) / 100
  return Number.isInteger(rounded) ? String(rounded) : rounded.toString()
})

function formatCount(value: number): string {
  return value.toLocaleString()
}

async function loadAffiliateDetail(silent = false): Promise<void> {
  if (!silent) {
    loading.value = true
  }
  try {
    detail.value = await userAPI.getAffiliateDetail()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.loadFailed')))
  } finally {
    if (!silent) {
      loading.value = false
    }
  }
}

async function copyCode(): Promise<void> {
  if (!detail.value?.aff_code) return
  await copyToClipboard(detail.value.aff_code, t('affiliate.codeCopied'))
}

async function copyInviteLink(): Promise<void> {
  if (!inviteLink.value) return
  await copyToClipboard(inviteLink.value, t('affiliate.linkCopied'))
}

async function transferQuota(): Promise<void> {
  if (!detail.value || detail.value.aff_quota <= 0 || transferring.value) return
  transferring.value = true
  try {
    const resp = await userAPI.transferAffiliateQuota()
    appStore.showSuccess(t('affiliate.transfer.success', { amount: formatCurrency(resp.transferred_quota) }))
    await Promise.all([
      loadAffiliateDetail(true),
      authStore.refreshUser().catch(() => undefined),
    ])
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.transferFailed')))
  } finally {
    transferring.value = false
  }
}

onMounted(() => {
  void loadAffiliateDetail()
})
</script>

<style scoped>
.scheme3-affiliate {
  --affiliate-card: #fffefa;
  --affiliate-ink: #27251f;
  --affiliate-muted: #777266;
  --affiliate-soft: #a49e90;
  --affiliate-line: #d8d2c3;
  --affiliate-accent: #1e5c42;
  --affiliate-amber: #b7791f;
  --affiliate-danger: #9e4d3d;
  color: var(--affiliate-ink);
}

.scheme3-affiliate-state {
  display: flex;
  min-height: 22rem;
  align-items: center;
  justify-content: center;
  gap: .6rem;
  border: 1px dashed var(--affiliate-line);
  border-radius: 8px;
  background: rgba(255,254,250,.65);
  color: var(--affiliate-muted);
  font-size: .76rem;
}

.scheme3-affiliate-spinner {
  width: 1.45rem;
  height: 1.45rem;
  border: 2px solid rgba(30,92,66,.2);
  border-top-color: var(--affiliate-accent);
  border-radius: 999px;
  animation: affiliate-spin .8s linear infinite;
}

.scheme3-affiliate-header {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 1.25rem;
  margin-bottom: 1rem;
  border-bottom: 1px solid var(--affiliate-line);
  padding: .1rem 0 1rem;
}

.scheme3-affiliate-kicker,.scheme3-affiliate-panel-kicker {
  margin: 0;
  color: var(--affiliate-muted);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: .61rem;
  font-weight: 800;
  letter-spacing: .1em;
}

.scheme3-affiliate h1 {
  margin: .34rem 0 0;
  color: var(--affiliate-ink);
  font-family: Georgia, 'Times New Roman', serif;
  font-size: clamp(1.55rem, 2.6vw, 2.1rem);
  font-weight: 500;
  letter-spacing: 0;
}

.scheme3-affiliate-subtitle {
  max-width: 31rem;
  margin: .42rem 0 0;
  color: var(--affiliate-muted);
  font-size: .74rem;
  line-height: 1.55;
}

.scheme3-affiliate-ledger {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  border: 1px solid var(--affiliate-line);
  border-radius: 7px;
  background: var(--affiliate-card);
}

.scheme3-affiliate-ledger span {
  display: grid;
  min-width: 5.4rem;
  gap: .08rem;
  border-right: 1px solid var(--affiliate-line);
  padding: .48rem .68rem;
  text-align: right;
}

.scheme3-affiliate-ledger span:last-child { border-right: 0; }

.scheme3-affiliate-ledger strong {
  overflow: hidden;
  color: var(--affiliate-ink);
  font-family: Georgia, 'Times New Roman', serif;
  font-size: .98rem;
  font-weight: 600;
  line-height: 1.1;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.scheme3-affiliate-ledger strong.is-accent { color: var(--affiliate-accent); }

.scheme3-affiliate-ledger small {
  overflow: hidden;
  color: var(--affiliate-muted);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: .52rem;
  font-weight: 700;
  letter-spacing: .04em;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.scheme3-affiliate-layout {
  display: grid;
  grid-template-columns: minmax(0, 1.45fr) minmax(16rem, .7fr);
  gap: .85rem;
}

.scheme3-affiliate-panel {
  border: 1px solid var(--affiliate-line);
  border-radius: 8px;
  background: var(--affiliate-card);
  box-shadow: 0 10px 24px rgba(54,48,34,.06);
}

.scheme3-affiliate-credential-panel,.scheme3-affiliate-transfer-panel { padding: 1.05rem; }
.scheme3-affiliate-invitees-panel { margin-top: .85rem; overflow: hidden; }

.scheme3-affiliate-panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .8rem;
  border-bottom: 1px solid var(--affiliate-line);
  padding-bottom: .8rem;
}

.scheme3-affiliate-panel-head h2 {
  margin: .24rem 0 0;
  color: var(--affiliate-ink);
  font-family: Georgia, 'Times New Roman', serif;
  font-size: 1.06rem;
  font-weight: 500;
}

.scheme3-affiliate-panel-icon {
  flex-shrink: 0;
  color: var(--affiliate-accent);
}

.scheme3-affiliate-credentials { display: grid; gap: .65rem; margin-top: .9rem; }

.scheme3-affiliate-credential-row {
  display: grid;
  grid-template-columns: 5.45rem minmax(0,1fr) auto;
  align-items: center;
  gap: .55rem;
  border: 1px solid var(--affiliate-line);
  border-radius: 7px;
  background: #f8f6ef;
  padding: .55rem;
}

.scheme3-affiliate-credential-label {
  color: var(--affiliate-muted);
  font-size: .69rem;
  font-weight: 700;
}

.scheme3-affiliate-credential-value {
  min-width: 0;
  overflow: hidden;
  color: var(--affiliate-ink);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: .68rem;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.scheme3-affiliate-copy,.scheme3-affiliate-transfer-button {
  display: inline-flex;
  min-height: 2.05rem;
  align-items: center;
  justify-content: center;
  gap: .35rem;
  border: 1px solid var(--affiliate-line);
  border-radius: 6px;
  padding: .4rem .6rem;
  background: var(--affiliate-card);
  color: var(--affiliate-ink);
  font-size: .66rem;
  font-weight: 800;
  transition: transform 150ms ease, border-color 150ms ease, background-color 150ms ease, color 150ms ease;
}

.scheme3-affiliate-copy:hover {
  border-color: rgba(30,92,66,.28);
  background: rgba(30,92,66,.07);
  color: var(--affiliate-accent);
  transform: translateY(-1px);
}

.scheme3-affiliate-copy:active,.scheme3-affiliate-transfer-button:active { transform: scale(.98); }

.scheme3-affiliate-note {
  margin-top: .9rem;
  border-left: 3px solid var(--affiliate-amber);
  background: rgba(183,121,31,.07);
  padding: .7rem .8rem;
}

.scheme3-affiliate-note p {
  margin: 0;
  color: #855712;
  font-size: .71rem;
  font-weight: 800;
}

.scheme3-affiliate-note ul {
  display: grid;
  gap: .24rem;
  margin: .45rem 0 0;
  padding: 0;
  color: var(--affiliate-muted);
  font-size: .67rem;
  line-height: 1.5;
  list-style: none;
}

.scheme3-affiliate-transfer-panel { display: flex; min-height: 100%; flex-direction: column; }
.scheme3-affiliate-transfer-copy { margin: .85rem 0 0; color: var(--affiliate-muted); font-size: .72rem; line-height: 1.55; }

.scheme3-affiliate-transfer-amount {
  display: grid;
  gap: .2rem;
  margin-top: auto;
  border-top: 1px solid var(--affiliate-line);
  border-bottom: 1px solid var(--affiliate-line);
  padding: .8rem 0;
}

.scheme3-affiliate-transfer-amount span { color: var(--affiliate-muted); font-size: .64rem; font-weight: 700; }
.scheme3-affiliate-transfer-amount strong { color: var(--affiliate-accent); font-family: Georgia, 'Times New Roman', serif; font-size: 1.48rem; font-weight: 600; line-height: 1.1; }
.scheme3-affiliate-frozen { margin: .58rem 0 0; color: var(--affiliate-amber); font-size: .65rem; }

.scheme3-affiliate-transfer-button {
  width: 100%;
  margin-top: .82rem;
  border-color: var(--affiliate-accent);
  background: var(--affiliate-accent);
  color: #fffefa;
  box-shadow: 0 8px 16px rgba(30,92,66,.14);
}

.scheme3-affiliate-transfer-button:hover:not(:disabled) { background: #174a35; transform: translateY(-1px); }
.scheme3-affiliate-transfer-button:disabled { cursor: not-allowed; border-color: var(--affiliate-line); background: #e6e1d5; color: #968f80; box-shadow: none; }
.scheme3-affiliate-empty-transfer { margin: .54rem 0 0; color: var(--affiliate-amber); font-size: .65rem; line-height: 1.45; }

.scheme3-affiliate-invitees-panel .scheme3-affiliate-panel-head { padding: 1.05rem 1.05rem .8rem; }
.scheme3-affiliate-invitee-count { border: 1px solid rgba(30,92,66,.22); border-radius: 999px; padding: .22rem .46rem; background: rgba(30,92,66,.07); color: var(--affiliate-accent); font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: .58rem; font-weight: 800; }
.scheme3-affiliate-empty-invitees { margin: 1rem; border: 1px dashed var(--affiliate-line); border-radius: 7px; padding: 2.2rem 1rem; color: var(--affiliate-muted); font-size: .73rem; text-align: center; }
.scheme3-affiliate-table-wrap { overflow-x: auto; }
.scheme3-affiliate-table { width: 100%; min-width: 38rem; border-collapse: collapse; color: var(--affiliate-ink); font-size: .72rem; }
.scheme3-affiliate-table th,.scheme3-affiliate-table td { border-bottom: 1px solid rgba(216,210,195,.78); padding: .78rem 1.05rem; text-align: left; }
.scheme3-affiliate-table th { background: #f1eee6; color: var(--affiliate-muted); font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: .57rem; font-weight: 800; letter-spacing: .055em; }
.scheme3-affiliate-table th.is-right { text-align: right; }
.scheme3-affiliate-table tr:last-child td { border-bottom: 0; }
.scheme3-affiliate-table tbody tr { transition: background-color 150ms ease; }
.scheme3-affiliate-table tbody tr:hover { background: rgba(30,92,66,.045); }
.scheme3-affiliate-table td.is-rebate { color: var(--affiliate-accent); font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-weight: 800; text-align: right; }

:global(.dark .scheme3-affiliate) {
  --affiliate-card: #24231f;
  --affiliate-ink: #f4f2ec;
  --affiliate-muted: #aaa69a;
  --affiliate-soft: #827e72;
  --affiliate-line: #47443a;
  --affiliate-accent: #8fc2a5;
  --affiliate-amber: #d3a55a;
  --affiliate-danger: #d38b79;
}

:global(.dark .scheme3-affiliate-state) { background: rgba(36,35,31,.65); }
:global(.dark .scheme3-affiliate-spinner) { border-color: rgba(143,194,165,.2); border-top-color: #8fc2a5; }
:global(.dark .scheme3-affiliate-credential-row) { background: #2b2924; }
:global(.dark .scheme3-affiliate-copy) { background: #24231f; }
:global(.dark .scheme3-affiliate-copy:hover) { border-color: rgba(143,194,165,.3); background: rgba(143,194,165,.1); color: #8fc2a5; }
:global(.dark .scheme3-affiliate-note) { background: rgba(211,165,90,.09); }
:global(.dark .scheme3-affiliate-note p) { color: #d3a55a; }
:global(.dark .scheme3-affiliate-transfer-button) { border-color: #8fc2a5; background: #8fc2a5; color: #1b1b18; box-shadow: 0 8px 16px rgba(0,0,0,.2); }
:global(.dark .scheme3-affiliate-transfer-button:hover:not(:disabled)) { background: #a7cfb3; }
:global(.dark .scheme3-affiliate-transfer-button:disabled) { border-color: #47443a; background: #35332d; color: #827e72; }
:global(.dark .scheme3-affiliate-invitee-count) { border-color: rgba(143,194,165,.28); background: rgba(143,194,165,.1); color: #8fc2a5; }
:global(.dark .scheme3-affiliate-table th) { background: #2b2924; }
:global(.dark .scheme3-affiliate-table th),:global(.dark .scheme3-affiliate-table td) { border-color: rgba(71,68,58,.86); }
:global(.dark .scheme3-affiliate-table tbody tr:hover) { background: rgba(143,194,165,.07); }

@media (max-width: 930px) {
  .scheme3-affiliate-layout { grid-template-columns: minmax(0,1fr); }
  .scheme3-affiliate-transfer-panel { min-height: 0; }
  .scheme3-affiliate-transfer-amount { margin-top: 1rem; }
}

@media (max-width: 767px) {
  .scheme3-affiliate-header { align-items: stretch; flex-direction: column; gap: .8rem; margin-bottom: .8rem; }
  .scheme3-affiliate-ledger { width: 100%; justify-content: stretch; }
  .scheme3-affiliate-ledger span { flex: 1 1 45%; min-width: 0; padding: .48rem .42rem; }
  .scheme3-affiliate-credential-panel,.scheme3-affiliate-transfer-panel { padding: .9rem; }
  .scheme3-affiliate-invitees-panel .scheme3-affiliate-panel-head { padding: .9rem; }
  .scheme3-affiliate-credential-row { grid-template-columns: minmax(0,1fr) auto; gap: .48rem; }
  .scheme3-affiliate-credential-label { grid-column: 1 / -1; }
  .scheme3-affiliate-table-wrap { overflow: visible; }
  .scheme3-affiliate-table,.scheme3-affiliate-table tbody,.scheme3-affiliate-table tr,.scheme3-affiliate-table td { display: block; min-width: 0; width: 100%; }
  .scheme3-affiliate-table thead { display: none; }
  .scheme3-affiliate-table tbody { display: grid; gap: .6rem; padding: .75rem; }
  .scheme3-affiliate-table tr { border: 1px solid var(--affiliate-line); border-radius: 7px; background: #f8f6ef; padding: .1rem .75rem; }
  .scheme3-affiliate-table td { display: flex; align-items: baseline; justify-content: space-between; gap: 1rem; border-bottom: 1px solid rgba(216,210,195,.72); padding: .55rem 0; text-align: right; word-break: break-word; }
  .scheme3-affiliate-table td:last-child { border-bottom: 0; }
  .scheme3-affiliate-table td::before { content: attr(data-label); flex: 0 0 auto; color: var(--affiliate-muted); font-size: .59rem; font-weight: 800; text-align: left; }
  .scheme3-affiliate-table td.is-rebate { text-align: right; }
  :global(.dark .scheme3-affiliate-table tr) { background: #2b2924; }
  :global(.dark .scheme3-affiliate-table td) { border-color: rgba(71,68,58,.86); }
}

@keyframes affiliate-spin { to { transform: rotate(360deg); } }

@media (prefers-reduced-motion: reduce) {
  .scheme3-affiliate *, .scheme3-affiliate *::before, .scheme3-affiliate *::after { animation-duration: .001ms !important; transition-duration: .001ms !important; }
}
</style>

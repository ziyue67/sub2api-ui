<template>
  <AppLayout>
    <div class="scheme3-channel-status-v2 space-y-6 pb-12">
      <section class="scheme3-v2-control-sheet">
        <header class="scheme3-v2-control-header">
          <div class="scheme3-v2-heading min-w-0">
            <p class="scheme3-v2-kicker">运行观测 / 多维渠道账本</p>
            <h1 class="scheme3-v2-title">
              <span class="scheme3-v2-title-mark">
                <Icon name="chart" size="sm" />
              </span>
              {{ t('channelMonitorV2.title') }}
            </h1>
            <p class="scheme3-v2-subtitle">成功率、首 Token、缓存与错误路径，统一收进同一套观测秩序。</p>
            <div class="scheme3-v2-status-line">
              <span class="scheme3-v2-status-dot" :class="loading || refreshing ? 'is-loading' : 'is-live'"></span>
              <span v-if="refreshing" class="scheme3-v2-status-copy is-accent">
                <LoadingSpinner size="sm" />
                {{ t('channelMonitorV2.updating') }}
              </span>
              <span v-else-if="snapshot?.coverage.data_through">
                {{ t('channelMonitorV2.updatedTo', { time: formatTime(snapshot.coverage.data_through) }) }}
              </span>
              <span v-else class="scheme3-v2-status-copy is-muted">{{ t('common.loading') }}</span>
              <span
                v-if="snapshot && !snapshot.coverage.coverage_complete && !bootstrapActive"
                class="scheme3-v2-badge scheme3-v2-badge-warning"
              >
                {{ t('channelMonitorV2.partialCoverage') }}
              </span>
              <span
                v-if="bootstrapActive"
                class="scheme3-v2-badge scheme3-v2-badge-accent inline-flex items-center gap-1"
              >
                <LoadingSpinner size="sm" />
                {{ t('channelMonitorV2.bootstrap.progress', { percent: bootstrapPercent }) }}
              </span>
            </div>
          </div>
          <div class="scheme3-v2-ledger" aria-label="渠道状态概览">
            <span>
              <strong>{{ observedRows }}</strong>
              <small>观测维度</small>
            </span>
            <span>
              <strong class="is-positive">{{ healthyRows }}</strong>
              <small>运行正常</small>
            </span>
            <span>
              <strong :class="degradedRows > 0 ? 'is-warning' : 'is-positive'">{{ degradedRows }}</strong>
              <small>需要留意</small>
            </span>
            <span>
              <strong>{{ currentRangeLabel }}</strong>
              <small>观测窗口</small>
            </span>
          </div>
          <button
            class="scheme3-v2-refresh"
            type="button"
            :title="t('common.refresh')"
            :disabled="loading"
            @click="reload(false)"
          >
            <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
          </button>
        </header>

        <!-- First-upgrade silent backfill: show until 30d product window is covered -->
        <div
          v-if="bootstrapActive"
          class="scheme3-v2-bootstrap"
          role="status"
          aria-live="polite"
        >
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div class="min-w-0 flex-1">
              <p class="scheme3-v2-bootstrap-title">
                {{ t('channelMonitorV2.bootstrap.title') }}
              </p>
              <p class="scheme3-v2-bootstrap-copy">
                {{ t('channelMonitorV2.bootstrap.description') }}
              </p>
            </div>
            <span class="scheme3-v2-bootstrap-percent">
              {{ t('channelMonitorV2.bootstrap.progress', { percent: bootstrapPercent }) }}
            </span>
          </div>
          <div
            class="scheme3-v2-progress-track"
            role="progressbar"
            :aria-valuenow="bootstrapPercent"
            aria-valuemin="0"
            aria-valuemax="100"
            :aria-label="t('channelMonitorV2.bootstrap.working')"
          >
            <div
              class="scheme3-v2-progress-value"
              :style="{ width: `${bootstrapPercent}%` }"
            />
          </div>
        </div>

        <!-- Single compact toolbar row: range · filters · view controls -->
        <div class="scheme3-v2-toolbar monitor-toolbar">
          <div
            class="scheme3-v2-segmented inline-flex shrink-0"
            role="group"
            :aria-label="t('channelMonitorV2.timeRange')"
          >
            <button
              v-for="option in ranges"
              :key="option.value"
              type="button"
              class="scheme3-v2-segment"
              :class="filter.range === option.value ? 'is-active' : ''"
              @click="setRange(option.value)"
            >
              {{ option.label }}
            </button>
          </div>

          <span class="scheme3-v2-divider" aria-hidden="true"></span>

          <FilterMultiSelect
            v-model="filter.platforms"
            compact
            :label="t('channelMonitorV2.filters.platform')"
            :all-label="t('channelMonitorV2.filters.allPlatforms')"
            :options="platformOptions"
          />
          <FilterMultiSelect
            v-model="selectedGroupIds"
            compact
            :label="t('channelMonitorV2.filters.group')"
            :all-label="t('channelMonitorV2.filters.allGroups')"
            :options="groupOptions"
          />
          <FilterMultiSelect
            v-model="filter.models"
            compact
            :label="t('channelMonitorV2.filters.model')"
            :all-label="t('channelMonitorV2.filters.allModels')"
            :options="modelOptions"
          />
          <button
            type="button"
            class="scheme3-v2-clear shrink-0"
            :disabled="!hasDimensionFilter"
            :class="!hasDimensionFilter ? 'opacity-40' : ''"
            @click="clearDimensions"
          >
            {{ t('channelMonitorV2.clearFilters') }}
          </button>

          <span class="scheme3-v2-divider" aria-hidden="true"></span>

          <Select
            v-model="matrixGroupBy"
            scheme3
            :options="matrixGroupOptions"
            :placeholder="t('channelMonitorV2.groupBy.label')"
            class="scheme3-v2-native-select monitor-toolbar-select w-[7.5rem] shrink-0 sm:w-[8.5rem]"
          />

          <div
            class="scheme3-v2-segmented inline-flex shrink-0"
            role="group"
            :aria-label="t('channelMonitorV2.trendView.label')"
          >
            <button
              type="button"
              class="scheme3-v2-segment"
              :class="trendView === 'pulse' ? 'is-active' : ''"
              @click="trendView = 'pulse'"
            >
              {{ t('channelMonitorV2.trendView.pulse') }}
            </button>
            <button
              type="button"
              class="scheme3-v2-segment"
              :class="trendView === 'line' ? 'is-active' : ''"
              @click="trendView = 'line'"
            >
              {{ t('channelMonitorV2.trendView.line') }}
            </button>
          </div>

          <div
            v-if="trendView === 'pulse'"
            class="scheme3-v2-segmented inline-flex shrink-0"
            role="group"
            :aria-label="t('channelMonitorV2.healthMode.label')"
          >
            <button
              v-for="option in healthModeOptions"
              :key="option.value"
              type="button"
              class="scheme3-v2-segment"
              :class="healthMode === option.value ? 'is-active' : ''"
              @click="healthMode = option.value"
            >
              {{ option.label }}
            </button>
          </div>
        </div>
      </section>

      <!-- Overview KPI: success · TTFT · tokens/s(optional) · cache · (+ RPM when throughput visible) -->
      <section
        v-if="snapshot"
        class="scheme3-v2-kpi-grid grid grid-cols-2 gap-3 sm:grid-cols-3"
        :class="showThroughput ? 'xl:grid-cols-5' : 'xl:grid-cols-4'"
        :aria-label="t('channelMonitorV2.summaryAria')"
      >
        <MetricCell
          :label="t('channelMonitorV2.metrics.successRate')"
          :value="formatPercent(1 - snapshot.metrics.error_rate)"
          :detail="t('channelMonitorV2.metrics.errorRateValue', { value: formatPercent(snapshot.metrics.error_rate) })"
          :state="snapshot.health.error_rate"
        />
        <MetricCell
          :label="t('channelMonitorV2.metrics.ttftP50')"
          :value="formatMs(snapshot.metrics.ttft.p50_ms)"
          :detail="latencyKpiSecondary(snapshot.metrics.ttft)"
          :title="latencyDetail(snapshot.metrics.ttft)"
          :state="snapshot.health.ttft"
        />
        <MetricCell
          v-if="showThroughput"
          :label="t('channelMonitorV2.metrics.tps')"
          :value="formatTps(snapshot.metrics.tpm)"
          :detail="t('channelMonitorV2.metrics.tpsDetail')"
          :title="exactTps(snapshot.metrics.tpm)"
        />
        <MetricCell
          :label="t('channelMonitorV2.metrics.cacheRate')"
          :value="formatPercent(snapshot.metrics.cache_rate)"
          :detail="t('channelMonitorV2.metrics.cacheDetail')"
          :state="snapshot.health.cache || snapshot.health.overall"
        />
        <MetricCell
          v-if="showThroughput"
          :label="t('channelMonitorV2.metrics.rpm')"
          :value="formatRate(snapshot.metrics.rpm)"
          :detail="t('channelMonitorV2.metrics.rpmDetail')"
          :title="exactRate(snapshot.metrics.rpm)"
        />
      </section>
      <section
        v-else-if="loading"
        class="scheme3-v2-kpi-grid grid grid-cols-2 gap-3 sm:grid-cols-3"
        :class="showThroughput ? 'xl:grid-cols-5' : 'xl:grid-cols-4'"
        aria-hidden="true"
      >
        <div
          v-for="i in (showThroughput ? 5 : 4)"
          :key="i"
          class="scheme3-v2-skeleton h-24 animate-pulse"
        />
      </section>

      <div class="relative min-h-[320px]">
        <MonitorTrendChart
          v-if="trendView === 'line'"
          :trend="snapshot?.trend || []"
          :coverage="snapshot?.coverage || null"
          :loading="loading && !snapshot"
        />
        <RelayPulseMatrix
          v-else-if="matrix"
          :rows="matrixRows"
          :coverage="matrix.coverage"
          :health-mode="healthMode"
          :show-throughput="showThroughput"
        />
        <div
          v-else-if="loading"
          class="scheme3-v2-loading-panel flex min-h-[320px] items-center justify-center"
        >
          <span class="animate-pulse">{{ t('common.loading') }}</span>
        </div>
      </div>

      <section class="scheme3-v2-data-panel flex min-h-0 flex-col overflow-hidden">
        <div class="scheme3-v2-data-tabs">
          <nav class="scheme3-v2-segmented w-full max-w-md sm:w-auto" role="tablist" :aria-label="t('channelMonitorV2.tabs.aria')">
            <button
              v-for="item in tabs"
              :key="item.value"
              type="button"
              role="tab"
              class="scheme3-v2-segment flex-1 sm:flex-none"
              :aria-selected="activeTab === item.value"
              :class="activeTab === item.value ? 'is-active' : ''"
              @click="activeTab = item.value"
            >
              {{ item.label }}
            </button>
          </nav>
        </div>
        <div class="scheme3-v2-table-scroll min-h-0 max-h-[min(52vh,520px)] overflow-auto p-4 sm:p-5">
          <div v-if="activeTab === 'models'" class="scheme3-v2-table-wrap border-0">
            <table class="scheme3-v2-table min-w-[720px]">
              <thead>
                <tr>
                  <th>{{ t('channelMonitorV2.table.platformModel') }}</th>
                  <th>{{ t('channelMonitorV2.metrics.successRate') }}</th>
                  <th>{{ t('channelMonitorV2.metrics.ttftP50') }}</th>
                  <th v-if="showThroughput">{{ t('channelMonitorV2.metrics.tps') }}</th>
                  <th>{{ t('channelMonitorV2.metrics.cacheRate') }}</th>
                  <th v-if="showThroughput">{{ t('channelMonitorV2.metrics.rpm') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="row in modelRows"
                  :key="`${row.platform}:${row.model}`"
                  class="cursor-pointer"
                  @click="drillModel(row)"
                >
                  <td>
                    <div class="flex items-center gap-2">
                      <span :class="statusDot(row.health)" aria-hidden="true"></span>
                      <div>
                        <span class="scheme3-v2-table-meta block text-xs">{{ row.platform }}</span>
                        <strong class="scheme3-v2-table-primary font-semibold">
                          {{ row.model === '__other__' ? t('channelMonitorV2.otherModels') : row.model }}
                        </strong>
                      </div>
                    </div>
                  </td>
                  <td>
                    <span class="block">{{ formatPercent(1 - row.metrics.error_rate) }}</span>
                    <small class="scheme3-v2-table-secondary text-xs">{{ t('channelMonitorV2.metrics.errorRateValue', { value: formatPercent(row.metrics.error_rate) }) }}</small>
                  </td>
                  <td>
                    <span class="block">{{ formatMs(row.metrics.ttft.p50_ms) }}</span>
                    <small class="scheme3-v2-table-secondary text-xs">{{ latencyDetail(row.metrics.ttft) }}</small>
                  </td>
                  <td v-if="showThroughput" :title="exactTps(row.metrics.tpm)">{{ formatTps(row.metrics.tpm) }}</td>
                  <td>{{ formatPercent(row.metrics.cache_rate) }}</td>
                  <td v-if="showThroughput">{{ formatRate(row.metrics.rpm) }}</td>
                </tr>
              </tbody>
            </table>
          </div>

          <div v-else-if="activeTab === 'errors'" class="scheme3-v2-error-list space-y-3">
            <div
              v-for="row in errorRows"
              :key="row.category"
              class="scheme3-v2-error-row p-4 text-sm"
              :class="row.ignored ? 'opacity-60' : ''"
            >
              <button
                type="button"
                class="grid w-full grid-cols-[minmax(100px,200px)_1fr_auto_auto] items-center gap-3 text-left"
                @click="toggleError(row.category)"
              >
                <span class="scheme3-v2-row-label flex min-w-0 items-center gap-1.5 truncate">
                  <span class="truncate">{{ errorLabel(row.category) }}</span>
                  <span v-if="row.ignored" class="scheme3-v2-badge scheme3-v2-badge-muted shrink-0 text-[10px]">{{ t('channelMonitorV2.ignored') }}</span>
                </span>
                <span class="scheme3-v2-error-track h-2 overflow-hidden">
                  <i
                    class="scheme3-v2-error-fill block h-full"
                    :class="row.ignored ? 'is-ignored' : 'is-active'"
                    :style="{ width: `${Math.max(2, row.rate * 100)}%` }"
                  ></i>
                </span>
                <small
                  class="scheme3-v2-error-rate w-14 text-right text-xs tabular-nums"
                  :class="row.ignored ? 'is-ignored' : ''"
                >{{ formatPercent(row.rate) }}</small>
                <Icon name="chevronDown" size="sm" class="scheme3-v2-row-chevron transition-transform" :class="expandedErrors.has(row.category) ? 'rotate-180' : ''" />
              </button>
              <div v-if="expandedErrors.has(row.category)" class="scheme3-v2-error-details mt-3 space-y-2 pt-3">
                <template v-if="isAdmin && (row.details || []).length">
                  <div
                    v-for="(detail, index) in row.details || []"
                    :key="`${row.category}:${index}:${detail.message}`"
                    class="scheme3-v2-error-detail px-3 py-2 text-xs"
                  >
                    <div class="mb-1 flex flex-wrap items-center gap-2">
                      <span class="scheme3-v2-badge scheme3-v2-badge-muted text-[10px]">{{ detail.platform || '-' }}</span>
                      <span class="truncate font-medium">{{ detail.model || '-' }}</span>
                      <span v-if="detail.status_code" class="scheme3-v2-table-secondary">{{ t('channelMonitorV2.errorDetail.http', { code: detail.status_code }) }}</span>
                      <span v-if="detail.upstream_status_code" class="scheme3-v2-table-secondary">{{ t('channelMonitorV2.errorDetail.upstream', { code: detail.upstream_status_code }) }}</span>
                      <span class="scheme3-v2-table-secondary ml-auto">×{{ detail.count }}</span>
                    </div>
                    <p class="break-words leading-relaxed">{{ detail.message || detail.error_type || t('channelMonitorV2.errorDetail.noMessage') }}</p>
                  </div>
                </template>
                <p v-else class="scheme3-v2-table-secondary text-xs">{{ t('channelMonitorV2.errorDetail.empty') }}</p>
              </div>
            </div>
          </div>

          <div v-else class="scheme3-v2-table-wrap border-0">
            <table class="scheme3-v2-table min-w-[640px]">
              <thead>
                <tr>
                  <th class="w-16">{{ t('channelMonitorV2.table.rank') }}</th>
                  <th>{{ t('channelMonitorV2.table.user') }}</th>
                  <th>{{ t('channelMonitorV2.metrics.successRate') }}</th>
                  <th>{{ t('channelMonitorV2.metrics.ttftP50') }}</th>
                  <th v-if="showThroughput">{{ t('channelMonitorV2.metrics.tps') }}</th>
                  <th>{{ t('channelMonitorV2.metrics.cacheRate') }}</th>
                  <th v-if="showThroughput">{{ t('channelMonitorV2.metrics.rpm') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="row in userRows"
                  :key="row.user_id || row.display_label"
                  :class="row.is_self ? 'scheme3-v2-current-row' : ''"
                >
                  <td><MonitorRankBadge :rank="row.rank" /></td>
                  <td>
                    <strong
                      class="font-semibold"
                      :class="row.is_self ? 'scheme3-v2-current-user' : ''"
                    >
                      {{ row.display_label }}
                      <span
                        v-if="row.is_self"
                        class="scheme3-v2-current-badge ml-2 text-[10px]"
                      >{{ t('channelMonitorV2.currentUser') }}</span>
                    </strong>
                  </td>
                  <td>
                    <span class="block">{{ formatPercent(1 - row.metrics.error_rate) }}</span>
                    <small class="scheme3-v2-table-secondary text-xs">{{ t('channelMonitorV2.metrics.errorRateValue', { value: formatPercent(row.metrics.error_rate) }) }}</small>
                  </td>
                  <td>
                    <span class="block">{{ formatMs(row.metrics.ttft.p50_ms) }}</span>
                    <small class="scheme3-v2-table-secondary text-xs">{{ latencyDetail(row.metrics.ttft) }}</small>
                  </td>
                  <td v-if="showThroughput" :title="exactTps(row.metrics.tpm)">{{ formatTps(row.metrics.tpm) }}</td>
                  <td>{{ formatPercent(row.metrics.cache_rate) }}</td>
                  <td v-if="showThroughput">{{ formatRate(row.metrics.rpm) }}</td>
                </tr>
              </tbody>
            </table>
          </div>

          <div v-if="tabLoading" class="scheme3-v2-empty py-10 text-sm">{{ t('common.loading') }}</div>
          <div v-else-if="activeRowsEmpty" class="scheme3-v2-empty py-10">
            <p class="scheme3-v2-empty-title text-base">
              {{
                bootstrapActive
                  ? t('channelMonitorV2.bootstrap.title')
                  : t('channelMonitorV2.empty.title')
              }}
            </p>
            <p class="scheme3-v2-empty-description">
              {{
                bootstrapActive
                  ? t('channelMonitorV2.bootstrap.description')
                  : t('channelMonitorV2.empty.description')
              }}
            </p>
          </div>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Select from '@/components/common/Select.vue'
import FilterMultiSelect from '@/features/channel-monitor-v2/FilterMultiSelect.vue'
import MetricCell from '@/features/channel-monitor-v2/MetricCell.vue'
import MonitorRankBadge from '@/features/channel-monitor-v2/MonitorRankBadge.vue'
import MonitorTrendChart from '@/features/channel-monitor-v2/MonitorTrendChart.vue'
import RelayPulseMatrix from '@/features/channel-monitor-v2/RelayPulseMatrix.vue'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { isChannelMonitorThroughputHidden } from '@/utils/featureFlags'
import * as api from '@/api/channelMonitorV2'
import type {
  HealthState,
  MonitorDimensions,
  MonitorErrorRow,
  MonitorFilter,
  MonitorHealth,
  MonitorMatrixGroupBy,
  MonitorMatrixResponse,
  MonitorModelRow,
  MonitorRange,
  MonitorSnapshot,
  MonitorUserRow,
} from '@/api/channelMonitorV2'
import {
  formatLatencyKpiSecondary,
  formatLatencyPrivacy,
  formatMonitorMs,
  formatMonitorPercent,
  formatMonitorThroughput,
  formatMonitorTokensPerSecond,
  tokensPerSecondFromTpm,
  healthScoreClass,
  monitorErrorCategoryLabel,
} from '@/features/channel-monitor-v2/monitorFormat'

type Tab = 'models' | 'errors' | 'users'
type HealthMode = 'overall' | 'success' | 'ttft' | 'cache'
type TrendView = 'pulse' | 'line'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const appStore = useAppStore()
const { t, te, locale } = useI18n()
const isAdmin = computed(() => authStore.isAdmin)
/** Admins always see RPM/TPM; users honor the hide-throughput system setting. */
const showThroughput = computed(() => isAdmin.value || !isChannelMonitorThroughputHidden())

const ranges = computed(() => [
  { value: '90m' as MonitorRange, label: t('channelMonitorV2.ranges.90m') },
  { value: '24h' as MonitorRange, label: t('channelMonitorV2.ranges.24h') },
  { value: '7d' as MonitorRange, label: t('channelMonitorV2.ranges.7d') },
  { value: '30d' as MonitorRange, label: t('channelMonitorV2.ranges.30d') },
])
const tabs = computed(() => [
  { value: 'models' as Tab, label: t('channelMonitorV2.tabs.models') },
  { value: 'errors' as Tab, label: t('channelMonitorV2.tabs.errors') },
  { value: 'users' as Tab, label: t('channelMonitorV2.tabs.users') },
])
const matrixGroupOptions = computed(() => [
  { value: 'platform' as MonitorMatrixGroupBy, label: t('channelMonitorV2.groupBy.platform') },
  { value: 'platform_group' as MonitorMatrixGroupBy, label: t('channelMonitorV2.groupBy.platformGroup') },
  { value: 'platform_model' as MonitorMatrixGroupBy, label: t('channelMonitorV2.groupBy.platformModel') },
  { value: 'platform_group_model' as MonitorMatrixGroupBy, label: t('channelMonitorV2.groupBy.platformGroupModel') },
])
const healthModeOptions = computed(() => [
  { value: 'overall' as HealthMode, label: t('channelMonitorV2.healthMode.overall') },
  { value: 'success' as HealthMode, label: t('channelMonitorV2.healthMode.success') },
  { value: 'ttft' as HealthMode, label: t('channelMonitorV2.healthMode.ttft') },
  { value: 'cache' as HealthMode, label: t('channelMonitorV2.healthMode.cache') },
])

const filter = ref<MonitorFilter>({
  range: parseRange(route.query.range),
  platforms: csv(route.query.platform),
  groupIds: csv(route.query.group).map(Number).filter(Boolean),
  models: csv(route.query.model),
})
const activeTab = ref<Tab>(
  (['models', 'errors', 'users'].includes(String(route.query.tab)) ? route.query.tab : 'models') as Tab
)
const matrixGroupBy = ref<MonitorMatrixGroupBy>(parseMatrixGroupBy(route.query.group_by))
const healthMode = ref<HealthMode>(parseHealthMode(route.query.health_mode))
const trendView = ref<TrendView>(parseTrendView(route.query.trend_view))
const dimensions = ref<MonitorDimensions>({ platforms: [], groups: [], models: [] })
const snapshot = ref<MonitorSnapshot | null>(null)
const matrix = ref<MonitorMatrixResponse | null>(null)
const modelRows = ref<MonitorModelRow[]>([])
const errorRows = ref<MonitorErrorRow[]>([])
const userRows = ref<MonitorUserRow[]>([])
const loading = ref(false)
const tabLoading = ref(false)
const refreshing = ref(false)
const expandedErrors = ref(new Set<string>())
let controller: AbortController | null = null
let sequence = 0
let autoRefreshTimer: number | null = null

const hasDimensionFilter = computed(
  () => filter.value.platforms.length + filter.value.groupIds.length + filter.value.models.length > 0
)
// Full platform catalog (never pruned). Groups/models cascade by selected platforms
// so choosing a platform narrows the other pickers without collapsing platforms.
const platformOptions = computed(() =>
  (dimensions.value.platforms || []).map((item) => ({
    value: item.value,
    label: item.label,
  }))
)
const selectedPlatforms = computed(() => new Set(filter.value.platforms))
const groupOptions = computed(() =>
  (dimensions.value.groups || [])
    .filter(
      (item) =>
        selectedPlatforms.value.size === 0 ||
        !item.platform ||
        selectedPlatforms.value.has(item.platform),
    )
    .map((item) => ({
      value: String(item.id),
      label: item.platform ? `${item.platform} / ${item.name || `#${item.id}`}` : item.name || `#${item.id}`,
    }))
)
const modelOptions = computed(() =>
  (dimensions.value.models || [])
    .filter(
      (item) =>
        selectedPlatforms.value.size === 0 ||
        !item.platform ||
        selectedPlatforms.value.has(item.platform),
    )
    .map((item) => ({
      value: item.value,
      label:
        item.platform && !item.label.includes(item.platform)
          ? `${item.platform} / ${item.label}`
          : item.label,
    }))
)
const selectedGroupIds = computed({
  get: () => filter.value.groupIds.map(String),
  set: (value: string[]) => {
    filter.value.groupIds = value.map(Number).filter((id) => Number.isInteger(id) && id > 0)
  },
})
// Soft-prune group/model selections that fall outside the platform cascade.
// Do NOT wipe when options are temporarily empty (loading); only drop invalid ids.
watch(
  [groupOptions, modelOptions],
  () => {
    if (groupOptions.value.length > 0) {
      const allowed = new Set(groupOptions.value.map((item) => item.value))
      const next = filter.value.groupIds.filter((id) => allowed.has(String(id)))
      if (next.length !== filter.value.groupIds.length) {
        filter.value.groupIds = next
      }
    }
    if (modelOptions.value.length > 0) {
      const allowed = new Set(modelOptions.value.map((item) => item.value))
      const next = filter.value.models.filter((model) => allowed.has(model))
      if (next.length !== filter.value.models.length) {
        filter.value.models = next
      }
    }
  },
  { flush: 'post' },
)
const activeRowsEmpty = computed(() =>
  activeTab.value === 'models'
    ? modelRows.value.length === 0
    : activeTab.value === 'errors'
      ? errorRows.value.length === 0
      : userRows.value.length === 0
)
/** First-upgrade backfill toward 90m/24h/7d/30d; banner hides when backend omits bootstrap. */
const bootstrapActive = computed(() => Boolean(snapshot.value?.coverage?.bootstrap?.active))
const bootstrapPercent = computed(() => {
  const raw = snapshot.value?.coverage?.bootstrap?.progress_percent
  if (typeof raw !== 'number' || Number.isNaN(raw)) return 0
  return Math.min(100, Math.max(0, Math.round(raw)))
})
const matrixRows = computed(() => {
  const items = matrix.value?.items || []
  // platform_group views should only show real groups, never bare platform placeholders.
  if (matrixGroupBy.value === 'platform_group' || matrixGroupBy.value === 'platform_group_model') {
    return items.filter((row) => row.group_id != null && Number(row.group_id) > 0)
  }
  return items
})
const observedRows = computed(() => matrixRows.value.length)
const healthyRows = computed(() => matrixRows.value.filter((row) => row.health.overall === 'healthy').length)
const degradedRows = computed(() => Math.max(0, observedRows.value - healthyRows.value))
const currentRangeLabel = computed(() => ranges.value.find((item) => item.value === filter.value.range)?.label || filter.value.range)

function csv(value: unknown) {
  return typeof value === 'string' ? value.split(',').filter(Boolean) : []
}
function parseRange(value: unknown): MonitorRange {
  return ['90m', '24h', '7d', '30d'].includes(String(value)) ? (value as MonitorRange) : '90m'
}
function parseMatrixGroupBy(value: unknown): MonitorMatrixGroupBy {
  const allowed: MonitorMatrixGroupBy[] = [
    'platform',
    'platform_group',
    'platform_model',
    'platform_group_model',
  ]
  return allowed.includes(value as MonitorMatrixGroupBy)
    ? (value as MonitorMatrixGroupBy)
    : 'platform_group'
}
function parseHealthMode(value: unknown): HealthMode {
  const allowed: HealthMode[] = ['overall', 'success', 'ttft', 'cache']
  return allowed.includes(value as HealthMode) ? (value as HealthMode) : 'overall'
}
function parseTrendView(value: unknown): TrendView {
  return value === 'line' ? 'line' : 'pulse'
}
function syncQuery() {
  void router.replace({
    query: {
      range: filter.value.range,
      platform: filter.value.platforms.join(',') || undefined,
      group: filter.value.groupIds.join(',') || undefined,
      model: filter.value.models.join(',') || undefined,
      group_by: matrixGroupBy.value,
      health_mode: healthMode.value,
      trend_view: trendView.value === 'line' ? 'line' : undefined,
      tab: activeTab.value,
    },
  })
}
/** Dimensions catalog: range only — never re-filtered by platform/group/model selection. */
async function loadDimensions(signal?: AbortSignal, id = sequence) {
  const rangeOnly: MonitorFilter = {
    range: filter.value.range,
    platforms: [],
    groupIds: [],
    models: [],
  }
  const next = await api.getDimensions(rangeOnly, isAdmin.value, signal)
  if (id !== sequence) return
  dimensions.value = next
}

async function loadMetrics(signal?: AbortSignal, id = sequence) {
  const [nextSnapshot, nextMatrix] = await Promise.all([
    api.getSnapshot(filter.value, isAdmin.value, signal),
    api.getMatrix(filter.value, matrixGroupBy.value, isAdmin.value, signal),
  ])
  if (id !== sequence) return
  snapshot.value = nextSnapshot
  matrix.value = nextMatrix
  scheduleAutoRefresh()
  await loadTab(signal, id)
}

async function reload(silent = true) {
  controller?.abort()
  const request = new AbortController()
  controller = request
  const id = ++sequence
  refreshing.value = true
  if (!silent) loading.value = true
  try {
    // Catalog + metrics in parallel; catalog ignores dimension filters so options never shrink.
    await Promise.all([
      loadDimensions(request.signal, id),
      loadMetrics(request.signal, id),
    ])
  } catch (error) {
    if ((error as { name?: string }).name !== 'CanceledError') {
      appStore.showError(extractApiErrorMessage(error, t('channelMonitorV2.loadFailed')))
    }
  } finally {
    if (id === sequence) {
      loading.value = false
      tabLoading.value = false
      refreshing.value = false
    }
  }
}

/** When only range changes, still refresh dimensions; dimension filters only re-load metrics. */
async function reloadMetricsOnly(silent = true) {
  controller?.abort()
  const request = new AbortController()
  controller = request
  const id = ++sequence
  refreshing.value = true
  if (!silent) loading.value = true
  try {
    await loadMetrics(request.signal, id)
  } catch (error) {
    if ((error as { name?: string }).name !== 'CanceledError') {
      appStore.showError(extractApiErrorMessage(error, t('channelMonitorV2.loadFailed')))
    }
  } finally {
    if (id === sequence) {
      loading.value = false
      tabLoading.value = false
      refreshing.value = false
    }
  }
}
async function loadTab(signal?: AbortSignal, id = sequence) {
  tabLoading.value = true
  try {
    if (activeTab.value === 'models') {
      modelRows.value = (await api.getModels(filter.value, isAdmin.value, signal)).items || []
    } else if (activeTab.value === 'errors') {
      errorRows.value = (await api.getErrors(filter.value, isAdmin.value, signal)).items || []
    } else {
      userRows.value = (await api.getUsers(filter.value, isAdmin.value, signal)).items || []
    }
  } catch (error) {
    const e = error as { name?: string; code?: string }
    if (e?.name === 'AbortError' || e?.name === 'CanceledError' || e?.code === 'ERR_CANCELED') return
    appStore.showError(extractApiErrorMessage(error, t('channelMonitorV2.detailLoadFailed')))
  } finally {
    if (id === sequence) tabLoading.value = false
  }
}
function setRange(value: MonitorRange) {
  filter.value.range = value
}
function clearDimensions() {
  // Replace arrays so deep watch always fires and metrics reload full window.
  filter.value = {
    ...filter.value,
    platforms: [],
    groupIds: [],
    models: [],
  }
}
function scheduleAutoRefresh() {
  if (autoRefreshTimer) {
    window.clearInterval(autoRefreshTimer)
    autoRefreshTimer = null
  }
  // Poll faster while first-upgrade bootstrap is filling 90m→30d so the progress bar moves.
  const seconds = bootstrapActive.value
    ? 10
    : snapshot.value?.config?.refresh_interval_seconds || 300
  autoRefreshTimer = window.setInterval(() => {
    if (!loading.value && !refreshing.value) {
      void reload(true)
    }
  }, Math.max(bootstrapActive.value ? 10 : 60, seconds) * 1000)
}
function drillModel(row: MonitorModelRow) {
  filter.value.platforms = [row.platform]
  filter.value.models = [row.model]
}
function formatRate(value: number) {
  return formatMonitorThroughput(value)
}
function exactRate(value: number) {
  return Intl.NumberFormat(locale.value || undefined, { maximumFractionDigits: 2 }).format(value || 0)
}
function formatTps(tpm: number | null | undefined) {
  return formatMonitorTokensPerSecond(tpm)
}
function exactTps(tpm: number | null | undefined) {
  return Intl.NumberFormat(locale.value || undefined, { maximumFractionDigits: 3 }).format(
    tokensPerSecondFromTpm(tpm),
  )
}
function formatPercent(value: number) {
  return formatMonitorPercent(value)
}
function formatMs(value: number | null) {
  return formatMonitorMs(value)
}
function latencyDetail(metric: {
  p50_ms: number | null
  p90_ms?: number | null
  p95_ms: number | null
  avg_ms?: number | null
}) {
  return formatLatencyPrivacy(metric.p50_ms, metric.p90_ms, metric.avg_ms, metric.p95_ms)
}
/** KPI secondary: AVG · P90 under the P50 primary value. */
function latencyKpiSecondary(metric: {
  p90_ms?: number | null
  p95_ms: number | null
  avg_ms?: number | null
}) {
  return formatLatencyKpiSecondary(metric.avg_ms, metric.p90_ms, metric.p95_ms)
}
function formatTime(value: string) {
  return new Intl.DateTimeFormat(locale.value || undefined, {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value))
}
function statusDot(health?: MonitorHealth | HealthState) {
  if (!health || typeof health === 'string') {
    return `scheme3-v2-status-dot health-${health || 'unknown'}`
  }
  // Prefer multi-band score when available; otherwise fall back to the coarse
  // overall state for mixed-version/older payloads.
  const klass =
    health.score != null
      ? healthScoreClass(health, 'overall', 0)
      : `health-${health.overall || 'unknown'}`
  return `scheme3-v2-status-dot ${klass}`
}
function errorLabel(value: string) {
  const key = `channelMonitorV2.errorCategories.${value}`
  return te(key) ? t(key) : monitorErrorCategoryLabel(value)
}
function toggleError(category: string) {
  const next = new Set(expandedErrors.value)
  if (next.has(category)) next.delete(category)
  else next.add(category)
  expandedErrors.value = next
}

let lastRange: MonitorRange = filter.value.range
watch(
  filter,
  () => {
    syncQuery()
    const rangeChanged = filter.value.range !== lastRange
    lastRange = filter.value.range
    if (rangeChanged) void reload(true)
    else void reloadMetricsOnly(true)
  },
  { deep: true }
)
watch(matrixGroupBy, () => {
  syncQuery()
  void reloadMetricsOnly(true)
})
watch(healthMode, syncQuery)
watch(trendView, syncQuery)
watch(activeTab, () => {
  syncQuery()
  void loadTab()
})
onMounted(() => void reload(false))
onBeforeUnmount(() => {
  controller?.abort()
  if (autoRefreshTimer) window.clearInterval(autoRefreshTimer)
})
</script>

<style scoped>
.scheme3-channel-status-v2 {
  --v2-paper: #f4f2ec;
  --v2-surface: #fffefa;
  --v2-subtle: #f1eee6;
  --v2-ink: #27251f;
  --v2-muted: #777266;
  --v2-soft: #a49e90;
  --v2-line: #d8d2c3;
  --v2-accent: #1e5c42;
  --v2-amber: #b7791f;
  --v2-danger: #9e4d3d;
  color: var(--v2-ink);
}

.scheme3-v2-control-sheet { border-bottom: 1px solid var(--v2-line); }
.scheme3-v2-control-header {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto auto;
  align-items: end;
  gap: 1rem;
  border-bottom: 1px solid var(--v2-line);
  padding: .1rem 0 1rem;
}
.scheme3-v2-kicker {
  margin: 0;
  color: var(--v2-muted);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: .61rem;
  font-weight: 800;
  letter-spacing: .1em;
}
.scheme3-v2-title {
  display: flex;
  align-items: center;
  gap: .55rem;
  margin: .34rem 0 0;
  color: var(--v2-ink);
  font-family: Georgia, 'Times New Roman', serif;
  font-size: clamp(1.55rem, 2.6vw, 2.1rem);
  font-weight: 500;
  letter-spacing: 0;
}
.scheme3-v2-title-mark {
  display: inline-flex;
  width: 1.75rem;
  height: 1.75rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(30, 92, 66, .28);
  border-radius: 7px;
  background: rgba(30, 92, 66, .08);
  color: var(--v2-accent);
}
.scheme3-v2-subtitle {
  max-width: 38rem;
  margin: .42rem 0 0;
  color: var(--v2-muted);
  font-size: .74rem;
  line-height: 1.55;
}
.scheme3-v2-status-line {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: .45rem;
  margin-top: .55rem;
  color: var(--v2-muted);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: .6rem;
  font-weight: 700;
}
.scheme3-v2-status-dot {
  display: inline-block;
  width: .42rem;
  height: .42rem;
  border-radius: 999px;
  background: var(--v2-soft);
}
.scheme3-v2-status-dot.is-live { background: var(--v2-accent); box-shadow: 0 0 0 .2rem rgba(30, 92, 66, .12); }
.scheme3-v2-status-dot.is-loading { background: var(--v2-amber); }
.scheme3-v2-status-dot.health-score10 { background: var(--v2-accent); }
.scheme3-v2-status-dot.health-score9,
.scheme3-v2-status-dot.health-score8,
.scheme3-v2-status-dot.health-score7 { background: #4e8d68; }
.scheme3-v2-status-dot.health-score6,
.scheme3-v2-status-dot.health-score5,
.scheme3-v2-status-dot.health-score4 { background: var(--v2-amber); }
.scheme3-v2-status-dot.health-score3,
.scheme3-v2-status-dot.health-score2,
.scheme3-v2-status-dot.health-score1,
.scheme3-v2-status-dot.health-score0 { background: var(--v2-danger); }
.scheme3-v2-status-dot.health-healthy { background: var(--v2-accent); }
.scheme3-v2-status-dot.health-warning { background: var(--v2-amber); }
.scheme3-v2-status-dot.health-critical { background: var(--v2-danger); }
.scheme3-v2-status-dot.health-unknown { background: var(--v2-soft); }
.scheme3-v2-status-copy { display: inline-flex; align-items: center; gap: .3rem; }
.scheme3-v2-status-copy.is-accent { color: var(--v2-accent); }
.scheme3-v2-status-copy.is-muted { color: var(--v2-soft); }
.scheme3-v2-badge {
  display: inline-flex;
  align-items: center;
  border: 1px solid transparent;
  border-radius: 999px;
  padding: .18rem .42rem;
  font-size: .54rem;
  font-weight: 800;
  letter-spacing: .04em;
}
.scheme3-v2-badge-warning { border-color: rgba(183, 121, 31, .24); background: rgba(183, 121, 31, .1); color: var(--v2-amber); }
.scheme3-v2-badge-accent { border-color: rgba(30, 92, 66, .22); background: rgba(30, 92, 66, .08); color: var(--v2-accent); }
.scheme3-v2-ledger {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  border: 1px solid var(--v2-line);
  border-radius: 7px;
  background: var(--v2-surface);
}
.scheme3-v2-ledger span {
  display: grid;
  min-width: 4.8rem;
  gap: .08rem;
  border-right: 1px solid var(--v2-line);
  padding: .48rem .68rem;
  text-align: right;
}
.scheme3-v2-ledger span:last-child { border-right: 0; }
.scheme3-v2-ledger strong {
  color: var(--v2-accent);
  font-family: Georgia, 'Times New Roman', serif;
  font-size: 1rem;
  font-weight: 600;
  line-height: 1.1;
}
.scheme3-v2-ledger strong.is-positive { color: var(--v2-accent); }
.scheme3-v2-ledger strong.is-warning { color: var(--v2-amber); }
.scheme3-v2-ledger small {
  color: var(--v2-muted);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: .52rem;
  font-weight: 700;
  letter-spacing: .04em;
}
.scheme3-v2-refresh {
  display: inline-flex;
  width: 2rem;
  height: 2rem;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--v2-line);
  border-radius: 7px;
  background: var(--v2-surface);
  color: var(--v2-muted);
  transition: color 150ms ease, background-color 150ms ease, border-color 150ms ease;
}
.scheme3-v2-refresh:hover { border-color: rgba(30, 92, 66, .28); background: var(--v2-subtle); color: var(--v2-accent); }
.scheme3-v2-refresh:disabled { cursor: not-allowed; opacity: .5; }
.scheme3-v2-bootstrap {
  border-bottom: 1px solid rgba(30, 92, 66, .22);
  background: rgba(30, 92, 66, .06);
  padding: .7rem 0;
}
.scheme3-v2-bootstrap-title { color: var(--v2-accent); font-size: .78rem; font-weight: 800; }
.scheme3-v2-bootstrap-copy { margin-top: .15rem; color: var(--v2-muted); font-size: .66rem; }
.scheme3-v2-bootstrap-percent { color: var(--v2-accent); font-size: .66rem; font-weight: 800; font-variant-numeric: tabular-nums; }
.scheme3-v2-progress-track { height: .34rem; margin-top: .6rem; overflow: hidden; border-radius: 999px; background: rgba(30, 92, 66, .14); }
.scheme3-v2-progress-value { height: 100%; border-radius: inherit; background: var(--v2-accent); transition: width 500ms ease-out; }
.scheme3-v2-toolbar {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: .42rem;
  overflow-x: auto;
  padding: .7rem 0;
}
.scheme3-v2-divider { display: block; width: 1px; height: 1.25rem; flex: none; background: var(--v2-line); }
.scheme3-channel-status-v2 :deep(.scheme3-v2-segmented) {
  gap: 0 !important;
  border: 1px solid var(--v2-line) !important;
  border-radius: 7px !important;
  background: var(--v2-subtle) !important;
  padding: 2px !important;
}
.scheme3-channel-status-v2 :deep(.scheme3-v2-segment) {
  min-height: 1.75rem;
  border-radius: 5px !important;
  padding: .28rem .5rem !important;
  color: var(--v2-muted) !important;
  font-size: .61rem;
  font-weight: 800;
  box-shadow: none !important;
}
.scheme3-channel-status-v2 :deep(.scheme3-v2-segment:hover) { color: var(--v2-ink) !important; }
.scheme3-channel-status-v2 :deep(.scheme3-v2-segment.is-active) { background: var(--v2-surface) !important; color: var(--v2-accent) !important; box-shadow: 0 2px 6px rgba(54, 48, 34, .08) !important; }
.scheme3-v2-clear { border: 1px solid transparent; border-radius: 6px; background: transparent; color: var(--v2-muted); padding: .38rem .55rem; font-size: .62rem; font-weight: 800; }
.scheme3-v2-clear:hover { background: var(--v2-subtle) !important; color: var(--v2-ink) !important; }
.scheme3-channel-status-v2 :deep(.scheme3-v2-native-select .scheme3-select-trigger),
.scheme3-channel-status-v2 :deep(.scheme3-v2-native-select select) {
  border-color: var(--v2-line) !important;
  border-radius: 7px !important;
  background: var(--v2-surface) !important;
  color: var(--v2-ink) !important;
  box-shadow: none !important;
}
.scheme3-v2-kpi-grid { margin-top: -.15rem; }
.scheme3-v2-skeleton { border: 1px solid var(--v2-line); border-radius: 8px; background: var(--v2-subtle); }
.scheme3-v2-loading-panel { border: 1px dashed var(--v2-line); border-radius: 8px; color: var(--v2-muted); }
.scheme3-v2-data-panel { border: 1px solid var(--v2-line); border-radius: 8px; background: var(--v2-surface); }
.scheme3-v2-data-tabs { border-bottom: 1px solid var(--v2-line); padding: .7rem .9rem 0; }
.scheme3-v2-table-scroll { scrollbar-color: var(--v2-soft) transparent; }
.scheme3-channel-status-v2 :deep(.scheme3-v2-table) { width: 100%; border-collapse: collapse; color: var(--v2-ink); }
.scheme3-channel-status-v2 :deep(.scheme3-v2-table thead) { background: var(--v2-subtle); }
.scheme3-channel-status-v2 :deep(.scheme3-v2-table th) { border-color: var(--v2-line) !important; color: var(--v2-muted) !important; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: .57rem; letter-spacing: .05em; text-transform: uppercase; }
.scheme3-channel-status-v2 :deep(.scheme3-v2-table td) { border-color: var(--v2-line) !important; color: var(--v2-ink) !important; font-size: .7rem; }
.scheme3-channel-status-v2 :deep(.scheme3-v2-table tbody tr:hover) { background: var(--v2-subtle); }
.scheme3-v2-error-row { border-bottom: 1px solid var(--v2-line); background: transparent; }
.scheme3-v2-error-row:last-child { border-bottom: 0; }
.scheme3-v2-row-label { color: var(--v2-ink); }
.scheme3-v2-error-track { border-radius: 999px; background: var(--v2-subtle); }
.scheme3-v2-error-fill { border-radius: inherit; }
.scheme3-v2-error-fill.is-active { background: var(--v2-danger); }
.scheme3-v2-error-fill.is-ignored { background: var(--v2-soft); }
.scheme3-v2-error-rate { color: var(--v2-muted); }
.scheme3-v2-error-rate.is-ignored { color: var(--v2-soft); }
.scheme3-v2-row-chevron { color: var(--v2-muted); }
.scheme3-v2-error-details { border-top: 1px solid var(--v2-line); }
.scheme3-v2-error-detail { border: 1px solid var(--v2-line); border-radius: 6px; background: var(--v2-subtle); color: var(--v2-muted); }
.scheme3-v2-current-row { background: rgba(30, 92, 66, .08) !important; box-shadow: inset 3px 0 0 var(--v2-accent); }
.scheme3-v2-current-user { color: var(--v2-accent) !important; }
.scheme3-v2-current-badge { border: 1px solid rgba(30, 92, 66, .2); border-radius: 999px; background: rgba(30, 92, 66, .1); color: var(--v2-accent); padding: .14rem .35rem; font-weight: 800; }
.scheme3-v2-badge-muted { border-color: var(--v2-line); background: var(--v2-subtle); color: var(--v2-muted); }
.scheme3-v2-table-meta,
.scheme3-v2-table-secondary { color: var(--v2-muted); }
.scheme3-v2-table-primary { color: var(--v2-ink); }
.scheme3-v2-empty { color: var(--v2-muted); text-align: center; }
.scheme3-v2-empty-title { color: var(--v2-ink); font-weight: 700; }
.scheme3-v2-empty-description { color: var(--v2-muted); font-size: .72rem; }

:global(.dark .scheme3-channel-status-v2) {
  --v2-paper: #1b1b18;
  --v2-surface: #24231f;
  --v2-subtle: #2b2924;
  --v2-ink: #f4f2ec;
  --v2-muted: #aaa69a;
  --v2-soft: #827e72;
  --v2-line: #47443a;
  --v2-accent: #8fc2a5;
  --v2-amber: #d3a55a;
  --v2-danger: #d38b79;
}
:global(.dark .scheme3-channel-status-v2 .scheme3-v2-title-mark) { border-color: rgba(143, 194, 165, .3); background: rgba(143, 194, 165, .1); }
:global(.dark .scheme3-channel-status-v2 .scheme3-v2-status-dot.is-live) { box-shadow: 0 0 0 .2rem rgba(143, 194, 165, .12); }
:global(.dark .scheme3-channel-status-v2 .scheme3-v2-current-row) { background: rgba(143, 194, 165, .1) !important; }

@media (max-width: 767px) {
  .scheme3-v2-control-header { grid-template-columns: minmax(0, 1fr) auto; align-items: start; gap: .7rem; }
  .scheme3-v2-ledger { grid-column: 1 / -1; width: 100%; justify-content: stretch; }
  .scheme3-v2-ledger span { min-width: 0; flex: 1 1 45%; padding: .48rem .42rem; }
  .scheme3-v2-refresh { grid-column: 2; grid-row: 1; }
  .scheme3-v2-toolbar { padding-bottom: .55rem; }
}
@media (max-width: 520px) {
  .scheme3-v2-control-header { grid-template-columns: minmax(0, 1fr) auto; }
  .scheme3-v2-subtitle { max-width: 19rem; }
  .scheme3-v2-data-tabs { padding-left: .55rem; padding-right: .55rem; }
}
details > summary::-webkit-details-marker { display: none; }
</style>

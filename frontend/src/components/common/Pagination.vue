<template>
  <div
    :class="scheme3
      ? 'scheme3-pagination'
      : 'flex items-center justify-between border-t border-gray-200 bg-white px-4 py-3 dark:border-dark-700 dark:bg-dark-800 sm:px-6'"
  >
    <div :class="scheme3 ? 'scheme3-pagination-mobile' : 'flex flex-1 items-center justify-between sm:hidden'">
      <!-- Mobile pagination -->
      <button
        @click="goToPage(page - 1)"
        :disabled="page === 1"
        :class="scheme3 ? 'scheme3-pagination-button' : 'relative inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-200 dark:hover:bg-dark-600'"
      >
        {{ t('pagination.previous') }}
      </button>
      <span :class="scheme3 ? 'scheme3-pagination-info' : 'text-sm text-gray-700 dark:text-gray-300'">
        {{ t('pagination.pageOf', { page, total: totalPages }) }}
      </span>
      <button
        @click="goToPage(page + 1)"
        :disabled="page === totalPages"
        :class="scheme3 ? 'scheme3-pagination-button' : 'relative ml-3 inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-200 dark:hover:bg-dark-600'"
      >
        {{ t('pagination.next') }}
      </button>
    </div>

    <div :class="scheme3 ? 'scheme3-pagination-desktop' : 'hidden sm:flex sm:flex-1 sm:items-center sm:justify-between'">
      <!-- Desktop pagination info -->
      <div :class="scheme3 ? 'scheme3-pagination-summary' : 'flex items-center space-x-4'">
        <p :class="scheme3 ? 'scheme3-pagination-info' : 'text-sm text-gray-700 dark:text-gray-300'">
          {{ t('pagination.showing') }}
          <span class="font-medium">{{ fromItem }}</span>
          {{ t('pagination.to') }}
          <span class="font-medium">{{ toItem }}</span>
          {{ t('pagination.of') }}
          <span class="font-medium">{{ total }}</span>
          {{ t('pagination.results') }}
        </p>

        <!-- Page size selector -->
        <div v-if="showPageSizeSelector" :class="scheme3 ? 'scheme3-pagination-field' : 'flex items-center space-x-2'">
          <span :class="scheme3 ? 'scheme3-pagination-info' : 'text-sm text-gray-700 dark:text-gray-300'"
            >{{ t('pagination.perPage') }}:</span
          >
          <div :class="scheme3 ? 'scheme3-pagination-page-size' : 'page-size-select w-20'">
            <Select
              :model-value="pageSize"
              :options="pageSizeSelectOptions"
              :scheme3="scheme3"
              @update:model-value="handlePageSizeChange"
            />
          </div>
        </div>

        <div v-if="showJump" :class="scheme3 ? 'scheme3-pagination-field' : 'flex items-center space-x-2'">
          <span :class="scheme3 ? 'scheme3-pagination-info' : 'text-sm text-gray-700 dark:text-gray-300'">{{ t('pagination.jumpTo') }}</span>
          <input
            v-model="jumpPage"
            type="number"
            min="1"
            :max="totalPages"
            :class="scheme3 ? 'scheme3-pagination-input' : 'input w-20 text-sm'"
            :placeholder="t('pagination.jumpPlaceholder')"
            @keyup.enter="submitJump"
          />
          <button type="button" :class="scheme3 ? 'scheme3-pagination-button' : 'btn btn-ghost btn-sm'" @click="submitJump">
            {{ t('pagination.jumpAction') }}
          </button>
        </div>
      </div>

      <!-- Desktop pagination buttons -->
      <nav
        :class="scheme3 ? 'scheme3-pagination-nav' : 'relative z-0 inline-flex -space-x-px rounded-md shadow-sm'"
        aria-label="Pagination"
      >
        <!-- Previous button -->
        <button
          @click="goToPage(page - 1)"
          :disabled="page === 1"
          :class="scheme3 ? 'scheme3-pagination-button is-icon' : 'relative inline-flex items-center rounded-l-md border border-gray-300 bg-white px-2 py-2 text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-400 dark:hover:bg-dark-600'"
          :aria-label="t('pagination.previous')"
        >
          <Icon name="chevronLeft" size="md" />
        </button>

        <!-- Page numbers -->
        <button
          v-for="(pageNum, index) in visiblePages"
          :key="`${pageNum}-${index}`"
          @click="typeof pageNum === 'number' && goToPage(pageNum)"
          :disabled="typeof pageNum !== 'number'"
          :class="[
            scheme3 ? 'scheme3-pagination-button is-page' : 'relative inline-flex items-center border px-4 py-2 text-sm font-medium',
            pageNum === page
              ? (scheme3 ? 'is-current' : 'z-10 border-primary-500 bg-primary-50 text-primary-600 dark:bg-primary-900/30 dark:text-primary-400')
              : (scheme3 ? '' : 'border-gray-300 bg-white text-gray-700 hover:bg-gray-50 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-300 dark:hover:bg-dark-600'),
            typeof pageNum !== 'number' && (scheme3 ? 'is-placeholder' : 'cursor-default')
          ]"
          :aria-label="
            typeof pageNum === 'number' ? t('pagination.goToPage', { page: pageNum }) : undefined
          "
          :aria-current="pageNum === page ? 'page' : undefined"
        >
          {{ pageNum }}
        </button>

        <!-- Next button -->
        <button
          @click="goToPage(page + 1)"
          :disabled="page === totalPages"
          :class="scheme3 ? 'scheme3-pagination-button is-icon' : 'relative inline-flex items-center rounded-r-md border border-gray-300 bg-white px-2 py-2 text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-400 dark:hover:bg-dark-600'"
          :aria-label="t('pagination.next')"
        >
          <Icon name="chevronRight" size="md" />
        </button>
      </nav>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import Select from './Select.vue'
import { getConfiguredTablePageSizeOptions, normalizeTablePageSize } from '@/utils/tablePreferences'
import { setPersistedPageSize } from '@/composables/usePersistedPageSize'

const { t } = useI18n()

interface Props {
  total: number
  page: number
  pageSize: number
  pageSizeOptions?: number[]
  showPageSizeSelector?: boolean
  showJump?: boolean
  scheme3?: boolean
}

interface Emits {
  (e: 'update:page', page: number): void
  (e: 'update:pageSize', pageSize: number): void
}

const props = withDefaults(defineProps<Props>(), {
  pageSizeOptions: () => getConfiguredTablePageSizeOptions(),
  showPageSizeSelector: true,
  showJump: false,
  scheme3: false
})

const emit = defineEmits<Emits>()

const totalPages = computed(() => Math.ceil(props.total / props.pageSize))

const fromItem = computed(() => {
  if (props.total === 0) return 0
  return (props.page - 1) * props.pageSize + 1
})

const toItem = computed(() => {
  const to = props.page * props.pageSize
  return to > props.total ? props.total : to
})

const pageSizeSelectOptions = computed(() => {
  const options = Array.from(
    new Set([
      ...getConfiguredTablePageSizeOptions(),
      normalizeTablePageSize(props.pageSize)
    ])
  ).sort((a, b) => a - b)

  return options.map((size) => ({
    value: size,
    label: String(size)
  }))
})

const jumpPage = ref('')

const visiblePages = computed(() => {
  const pages: (number | string)[] = []
  const maxVisible = 7
  const total = totalPages.value

  if (total <= maxVisible) {
    // Show all pages if total is small
    for (let i = 1; i <= total; i++) {
      pages.push(i)
    }
  } else {
    // Always show first page
    pages.push(1)

    const start = Math.max(2, props.page - 2)
    const end = Math.min(total - 1, props.page + 2)

    // Add ellipsis before if needed
    if (start > 2) {
      pages.push('...')
    }

    // Add middle pages
    for (let i = start; i <= end; i++) {
      pages.push(i)
    }

    // Add ellipsis after if needed
    if (end < total - 1) {
      pages.push('...')
    }

    // Always show last page
    pages.push(total)
  }

  return pages
})

const goToPage = (newPage: number) => {
  if (newPage >= 1 && newPage <= totalPages.value && newPage !== props.page) {
    emit('update:page', newPage)
  }
}

const handlePageSizeChange = (value: string | number | boolean | null) => {
  if (value === null || typeof value === 'boolean') return
  const newPageSize = normalizeTablePageSize(typeof value === 'string' ? parseInt(value, 10) : value)
  setPersistedPageSize(newPageSize)
  emit('update:pageSize', newPageSize)
}

const submitJump = () => {
  const value = jumpPage.value.trim()
  if (!value) return
  const pageNum = Number.parseInt(value, 10)
  if (Number.isNaN(pageNum)) return
  const nextPage = Math.min(Math.max(pageNum, 1), totalPages.value)
  jumpPage.value = ''
  goToPage(nextPage)
}
</script>

<style scoped>
.page-size-select :deep(.select-trigger) {
  @apply px-3 py-1.5 text-sm;
}

.scheme3-pagination {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border-top: 1px solid #dad5c8;
  padding: .78rem 0 0;
  background: transparent;
  color: #655f53;
}
.scheme3-pagination-mobile,
.scheme3-pagination-desktop { min-width: 0; align-items: center; justify-content: space-between; }
.scheme3-pagination-mobile { display: flex; flex: 1 1 auto; gap: .6rem; }
.scheme3-pagination-desktop { display: flex; flex: 1 1 auto; gap: 1rem; }
.scheme3-pagination-summary,
.scheme3-pagination-field { display: flex; min-width: 0; align-items: center; gap: .65rem; }
.scheme3-pagination-summary { flex-wrap: wrap; }
.scheme3-pagination-info { color: #777266; font-size: .66rem; line-height: 1.35; white-space: nowrap; }
.scheme3-pagination-info .font-medium { color: #27251f; font-weight: 800; }
.scheme3-pagination-page-size { width: 4.8rem; }
.scheme3-pagination-input { width: 5rem; min-height: 2rem; border: 1px solid #dad5c8; border-radius: 5px; padding: .35rem .5rem; background: #fffefa; color: #27251f; font-size: .68rem; outline: none; }
.scheme3-pagination-input:focus { border-color: #1e5c42; box-shadow: 0 0 0 2px rgba(30, 92, 66, .1); }
.scheme3-pagination-nav { display: inline-flex; min-width: 0; }
.scheme3-pagination-button { display: inline-flex; min-height: 2rem; cursor: pointer; align-items: center; justify-content: center; gap: .35rem; border: 1px solid #dad5c8; border-radius: 5px; padding: .38rem .62rem; background: #fffefa; color: #655f53; font-size: .66rem; font-weight: 800; line-height: 1; transition: border-color 120ms ease, background-color 120ms ease, color 120ms ease; }
.scheme3-pagination-button:hover:not(:disabled) { border-color: rgba(30, 92, 66, .3); background: #f1eee6; color: #27251f; }
.scheme3-pagination-button:focus-visible { outline: 2px solid rgba(30, 92, 66, .28); outline-offset: 2px; }
.scheme3-pagination-button:disabled { cursor: not-allowed; opacity: .42; }
.scheme3-pagination-button.is-icon { min-width: 2rem; border-radius: 0; }
.scheme3-pagination-button.is-icon:first-child { border-radius: 5px 0 0 5px; }
.scheme3-pagination-button.is-icon:last-child { border-radius: 0 5px 5px 0; }
.scheme3-pagination-button.is-page { min-width: 2rem; border-left-width: 0; border-radius: 0; padding-right: .5rem; padding-left: .5rem; }
.scheme3-pagination-button.is-page.is-current { position: relative; z-index: 1; border-color: #1e5c42; background: rgba(30, 92, 66, .1); color: #1e5c42; }
.scheme3-pagination-button.is-page.is-placeholder { cursor: default; color: #a49e90; }

:global(html.dark) .scheme3-pagination { border-color: #47443a; color: #aaa69a; }
:global(html.dark) .scheme3-pagination-info { color: #aaa69a; }
:global(html.dark) .scheme3-pagination-info .font-medium { color: #f4f2ec; }
:global(html.dark) .scheme3-pagination-input,
:global(html.dark) .scheme3-pagination-button { border-color: #47443a; background: #24231f; color: #aaa69a; }
:global(html.dark) .scheme3-pagination-input:focus { border-color: #8fc2a5; box-shadow: 0 0 0 2px rgba(143, 194, 165, .12); }
:global(html.dark) .scheme3-pagination-button:hover:not(:disabled) { border-color: rgba(143, 194, 165, .3); background: #2b2924; color: #f4f2ec; }
:global(html.dark) .scheme3-pagination-button:focus-visible { outline-color: rgba(143, 194, 165, .38); }
:global(html.dark) .scheme3-pagination-button.is-page.is-current { border-color: #8fc2a5; background: rgba(143, 194, 165, .1); color: #8fc2a5; }
:global(html.dark) .scheme3-pagination-button.is-page.is-placeholder { color: #827e72; }

@media (max-width: 639px) {
  .scheme3-pagination { padding-top: .65rem; }
  .scheme3-pagination-desktop { display: none; }
  .scheme3-pagination-mobile { display: flex; }
}
@media (min-width: 640px) {
  .scheme3-pagination-mobile { display: none; }
}
</style>

<template>
  <AppLayout>
    <div class="scheme3-image-manager" data-testid="image-manager-view">
      <header class="image-manager-header">
        <div class="image-manager-title min-w-0">
          <p class="image-manager-kicker">影像档案 / 结果库</p>
          <h1>{{ t('imageManager.title') }}</h1>
          <p class="image-manager-subtitle">{{ t('imageManager.subtitle') }}</p>
        </div>
        <div class="image-manager-ledger" aria-label="图片库统计">
          <span><strong>{{ total }}</strong><small>归档结果</small></span>
          <span><strong>{{ selectedIds.length }}</strong><small>当前选中</small></span>
          <span><strong>{{ hasActiveFilters ? '筛选中' : '全部' }}</strong><small>浏览视图</small></span>
        </div>
        <div class="image-manager-actions">
          <button type="button" class="btn btn-secondary btn-sm" :disabled="loading" @click="loadImages">
            <Icon name="refresh" size="sm" />
            <span>{{ t('common.refresh') }}</span>
          </button>
          <a href="/chat-images" target="_blank" rel="noopener noreferrer" class="btn btn-primary btn-sm">
            <Icon name="sparkles" size="sm" />
            <span>{{ t('imageManager.openStudio') }}</span>
          </a>
        </div>
      </header>

      <section class="image-manager-toolbar">
        <div class="image-manager-toolbar-summary">
          <span>档案概览</span>
          <strong>{{ t('imageManager.totalImages', { count: total }) }}</strong>
        </div>
        <div v-if="selectedIds.length > 0" class="image-manager-selection">
          <span>{{ t('imageManager.selectedImages', { count: selectedIds.length }) }}</span>
          <button type="button" class="image-manager-text-button" @click="downloadSelected">
            <Icon name="download" size="xs" />
            <span>{{ t('imageManager.downloadSelected') }}</span>
          </button>
          <button type="button" class="image-manager-text-button danger" :disabled="deleting" @click="deleteSelected">
            <Icon name="trash" size="xs" />
            <span>{{ t('imageManager.deleteSelected') }}</span>
          </button>
          <button type="button" class="image-manager-text-button" @click="clearSelection">
            {{ t('imageManager.clearSelection') }}
          </button>
        </div>
      </section>

      <section class="image-manager-filters" data-testid="image-manager-filters">
        <div class="image-manager-filter-heading">
          <span>检索索引</span>
          <strong>按提示词、生成时间与画面规格整理你的创作结果。</strong>
        </div>
        <label class="image-manager-field image-manager-field-wide">
          <span>{{ t('imageManager.search') }}</span>
          <input v-model.trim="filters.q" type="search" class="input" :placeholder="t('imageManager.searchPlaceholder')" @keyup.enter="applyFilters" />
        </label>
        <label class="image-manager-field">
          <span>{{ t('imageManager.startDate') }}</span>
          <input v-model="filters.start_date" type="date" class="input" />
        </label>
        <label class="image-manager-field">
          <span>{{ t('imageManager.endDate') }}</span>
          <input v-model="filters.end_date" type="date" class="input" />
        </label>
        <label class="image-manager-field">
          <span>{{ t('imageManager.format') }}</span>
          <select v-model="filters.format" class="input">
            <option value="">{{ t('imageManager.allFormats') }}</option>
            <option value="png">PNG</option>
            <option value="jpeg">JPG</option>
            <option value="webp">WEBP</option>
            <option value="other">{{ t('imageManager.other') }}</option>
          </select>
        </label>
        <label class="image-manager-field">
          <span>{{ t('imageManager.orientation') }}</span>
          <select v-model="filters.orientation" class="input">
            <option value="">{{ t('imageManager.allOrientations') }}</option>
            <option value="landscape">{{ t('imageManager.landscape') }}</option>
            <option value="portrait">{{ t('imageManager.portrait') }}</option>
            <option value="square">{{ t('imageManager.square') }}</option>
            <option value="unknown">{{ t('imageManager.unknownSize') }}</option>
          </select>
        </label>
        <label class="image-manager-field">
          <span>{{ t('imageManager.resolution') }}</span>
          <select v-model="filters.resolution" class="input">
            <option value="">{{ t('imageManager.allResolutions') }}</option>
            <option value="1080p">1080P</option>
            <option value="2k">2K</option>
            <option value="4k">4K</option>
            <option value="unknown">{{ t('imageManager.unknownSize') }}</option>
          </select>
        </label>
        <label class="image-manager-field">
          <span>{{ t('imageManager.aspectRatio') }}</span>
          <select v-model="filters.aspect_ratio" class="input">
            <option value="">{{ t('imageManager.allAspectRatios') }}</option>
            <option value="1:1">1:1</option>
            <option value="4:3">4:3</option>
            <option value="3:4">3:4</option>
            <option value="16:9">16:9</option>
            <option value="9:16">9:16</option>
            <option value="other">{{ t('imageManager.other') }}</option>
            <option value="unknown">{{ t('imageManager.unknownSize') }}</option>
          </select>
        </label>
        <div class="image-manager-filter-actions">
          <button type="button" class="btn btn-primary btn-sm" :disabled="loading" @click="applyFilters">
            <Icon name="search" size="sm" />
            <span>{{ t('imageManager.applyFilters') }}</span>
          </button>
          <button type="button" class="btn btn-secondary btn-sm" :disabled="loading || !hasActiveFilters" @click="resetFilters">
            {{ t('imageManager.resetFilters') }}
          </button>
        </div>
      </section>

      <section v-if="loading && images.length === 0" class="image-manager-state">
        <Icon name="sync" size="xl" class="animate-spin text-primary-500" />
        <span>{{ t('imageManager.loading') }}</span>
      </section>

      <section v-else-if="images.length === 0" class="image-manager-state">
        <Icon name="image" size="xl" class="text-primary-500" />
        <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('imageManager.emptyTitle') }}</h2>
        <p class="max-w-md text-center text-sm text-gray-500 dark:text-dark-300">{{ t('imageManager.emptyDescription') }}</p>
        <a href="/chat-images" target="_blank" rel="noopener noreferrer" class="btn btn-primary btn-sm">
          <Icon name="sparkles" size="sm" />
          <span>{{ t('imageManager.openStudio') }}</span>
        </a>
      </section>

      <section v-else class="image-manager-grid">
        <article
          v-for="(item, index) in images"
          :key="item.id"
          class="image-manager-card"
          :class="{ selected: isSelected(item.id) }"
          data-testid="image-manager-card"
        >
          <button
            type="button"
            class="image-manager-select"
            :class="{ active: isSelected(item.id) }"
            :aria-label="isSelected(item.id) ? t('imageManager.deselectImage') : t('imageManager.selectImage')"
            @click.stop="toggleImage(item.id)"
          >
            <Icon v-if="isSelected(item.id)" name="check" size="xs" />
          </button>

          <button type="button" class="image-manager-preview" :aria-label="t('imageManager.previewImage')" @click="openPreview(item)">
            <img :src="displayUrl(item)" alt="" loading="lazy" />
          </button>

          <div class="image-manager-card-body">
            <div class="flex items-center justify-between gap-2">
              <span class="image-manager-format">{{ String(item.output_format || 'image').toUpperCase() }}</span>
              <span class="text-xs text-gray-400 dark:text-dark-400">{{ formatFileSize(item.byte_size) }}</span>
            </div>
            <p class="image-manager-prompt">{{ item.task_prompt || item.revised_prompt || t('imageManager.noPrompt') }}</p>
            <div class="image-manager-meta">
              <span>{{ displayModelLabel(item.task_model || 'gpt-image-2') }}</span>
              <span v-if="dimensionSummary(item)">{{ dimensionSummary(item) }}</span>
              <span>{{ formatTime(item.created_at) }}</span>
            </div>
            <div class="image-manager-card-actions">
              <button type="button" class="image-manager-icon-button" :title="t('imageManager.download')" @click="downloadImage(item, index)">
                <Icon name="download" size="sm" />
              </button>
              <button type="button" class="image-manager-icon-button" :title="t('imageManager.copyPrompt')" @click="copyPrompt(item)">
                <Icon name="copy" size="sm" />
              </button>
              <button type="button" class="image-manager-icon-button" :title="t('imageManager.reusePrompt')" @click="reusePrompt(item)">
                <Icon name="sparkles" size="sm" />
              </button>
              <button type="button" class="image-manager-icon-button" :title="t('imageManager.useAsReference')" @click="useAsReference(item)">
                <Icon name="image" size="sm" />
              </button>
              <button type="button" class="image-manager-icon-button danger" :title="t('imageManager.delete')" @click="deleteOne(item)">
                <Icon name="trash" size="sm" />
              </button>
            </div>
          </div>
        </article>
      </section>

      <footer v-if="hasMore" class="image-manager-footer">
        <button type="button" class="btn btn-secondary" :disabled="loading" @click="loadMore">
          {{ loading ? t('imageManager.loading') : t('imageManager.loadMore') }}
        </button>
      </footer>

      <div
        v-if="previewImage"
        class="image-manager-lightbox"
        data-testid="image-manager-lightbox"
        @click.self="closePreview"
      >
        <div class="mb-3 flex items-center justify-end gap-2">
          <button type="button" class="btn btn-secondary btn-sm bg-white/95 text-gray-800 hover:bg-white dark:bg-dark-800/95 dark:text-dark-100" @click.stop="downloadPreview">
            <Icon name="download" size="sm" />
            <span class="sr-only sm:not-sr-only">{{ t('imageManager.download') }}</span>
          </button>
          <button type="button" class="image-manager-lightbox-close" :aria-label="t('imageManager.closePreview')" @click.stop="closePreview">
            <Icon name="x" size="md" />
          </button>
        </div>
        <div class="flex min-h-0 flex-1 items-center justify-center">
          <img :src="displayUrl(previewImage)" alt="" class="max-h-full max-w-full rounded-lg object-contain shadow-2xl" />
        </div>
        <p v-if="previewPrompt" class="mx-auto mt-3 max-w-4xl text-center text-xs leading-5 text-white/75 sm:text-sm">
          {{ previewPrompt }}
        </p>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import {
  deleteManagedImages,
  downloadImageFile,
  listManagedImages,
  type ImageCreatorImageListParams,
  type ImageCreatorManagedImage,
} from '@/api/imageCreator'
import { useClipboard } from '@/composables/useClipboard'
import { useAppStore } from '@/stores'
import { displayModelLabel } from '@/utils/modelDisplay'

const PAGE_SIZE = 40

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const images = ref<ImageCreatorManagedImage[]>([])
const total = ref(0)
const loading = ref(false)
const deleting = ref(false)
const selectedIds = ref<number[]>([])
const previewImage = ref<ImageCreatorManagedImage | null>(null)
const imageDisplayUrls = ref<Record<number, string>>({})
const objectUrls = new Set<string>()
const filters = reactive({
  q: '',
  start_date: '',
  end_date: '',
  format: '',
  orientation: '',
  resolution: '',
  aspect_ratio: '',
})

const hasMore = computed(() => images.value.length < total.value)
const previewPrompt = computed(() => previewImage.value?.task_prompt || previewImage.value?.revised_prompt || '')
const hasActiveFilters = computed(() => Object.values(filters).some((value) => String(value || '').trim() !== ''))

onMounted(() => {
  void loadImages()
})

onUnmounted(() => {
  revokeObjectUrls()
})

async function loadImages(): Promise<void> {
  loading.value = true
  try {
    const response = await listManagedImages(buildListParams(0))
    images.value = response.items || []
    total.value = response.total || images.value.length
    selectedIds.value = selectedIds.value.filter((id) => images.value.some((image) => image.id === id))
    void hydrateImages(images.value)
  } catch (error: any) {
    appStore.showError(error?.message || t('imageManager.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function loadMore(): Promise<void> {
  if (loading.value || !hasMore.value) return
  loading.value = true
  try {
    const response = await listManagedImages(buildListParams(images.value.length))
    const nextImages = response.items || []
    images.value = [...images.value, ...nextImages]
    total.value = response.total || images.value.length
    void hydrateImages(nextImages)
  } catch (error: any) {
    appStore.showError(error?.message || t('imageManager.loadFailed'))
  } finally {
    loading.value = false
  }
}

function buildListParams(offset: number): ImageCreatorImageListParams {
  const params: ImageCreatorImageListParams = {
    limit: PAGE_SIZE,
    offset,
  }
  for (const [key, value] of Object.entries(filters)) {
    const text = String(value || '').trim()
    if (text) {
      const writableParams = params as Record<string, string | number>
      writableParams[key] = text
    }
  }
  return params
}

function applyFilters(): void {
  selectedIds.value = []
  void loadImages()
}

function resetFilters(): void {
  filters.q = ''
  filters.start_date = ''
  filters.end_date = ''
  filters.format = ''
  filters.orientation = ''
  filters.resolution = ''
  filters.aspect_ratio = ''
  applyFilters()
}

function isSelected(id: number): boolean {
  return selectedIds.value.includes(id)
}

function toggleImage(id: number): void {
  selectedIds.value = isSelected(id)
    ? selectedIds.value.filter((item) => item !== id)
    : [...selectedIds.value, id]
}

function clearSelection(): void {
  selectedIds.value = []
}

async function deleteOne(item: ImageCreatorManagedImage): Promise<void> {
  await deleteImages([item.id])
}

async function deleteSelected(): Promise<void> {
  await deleteImages(selectedIds.value)
}

async function deleteImages(ids: number[]): Promise<void> {
  const normalized = Array.from(new Set(ids.filter((id) => Number.isFinite(id) && id > 0)))
  if (normalized.length === 0 || deleting.value) return
  deleting.value = true
  try {
    const result = await deleteManagedImages(normalized)
    const deleted = result.deleted || normalized.length
    images.value = images.value.filter((image) => !normalized.includes(image.id))
    total.value = Math.max(0, total.value - deleted)
    selectedIds.value = selectedIds.value.filter((id) => !normalized.includes(id))
    if (previewImage.value && normalized.includes(previewImage.value.id)) {
      previewImage.value = null
    }
    appStore.showSuccess(t('imageManager.deleteSuccess', { count: deleted }))
  } catch (error: any) {
    appStore.showError(error?.message || t('imageManager.deleteFailed'))
  } finally {
    deleting.value = false
  }
}

async function hydrateImages(items: ImageCreatorManagedImage[]): Promise<void> {
  await Promise.all(items.map(async (item) => {
    if (!shouldFetchImageUrl(item.url) || imageDisplayUrls.value[item.id]) return
    try {
      const blob = await downloadImageFile(item.url)
      const objectUrl = URL.createObjectURL(blob)
      objectUrls.add(objectUrl)
      imageDisplayUrls.value = { ...imageDisplayUrls.value, [item.id]: objectUrl }
    } catch {
      imageDisplayUrls.value = { ...imageDisplayUrls.value, [item.id]: item.url }
    }
  }))
}

function shouldFetchImageUrl(url: string): boolean {
  return typeof url === 'string' && url.startsWith('/api/')
}

function displayUrl(item: ImageCreatorManagedImage): string {
  return imageDisplayUrls.value[item.id] || item.url
}

async function ensureDisplayUrl(item: ImageCreatorManagedImage): Promise<string> {
  if (!imageDisplayUrls.value[item.id]) {
    await hydrateImages([item])
  }
  return displayUrl(item)
}

async function downloadImage(item: ImageCreatorManagedImage, index: number): Promise<void> {
  const href = await ensureDisplayUrl(item)
  const link = document.createElement('a')
  link.href = href
  link.download = buildDownloadName(item, index)
  document.body.appendChild(link)
  link.click()
  link.remove()
}

async function downloadSelected(): Promise<void> {
  const selected = images.value
    .map((image, index) => ({ image, index }))
    .filter(({ image }) => selectedIds.value.includes(image.id))
  for (const { image, index } of selected) {
    await downloadImage(image, index)
  }
}

function downloadPreview(): void {
  if (!previewImage.value) return
  const index = images.value.findIndex((item) => item.id === previewImage.value?.id)
  void downloadImage(previewImage.value, index >= 0 ? index : 0)
}

function buildDownloadName(item: ImageCreatorManagedImage, index: number): string {
  const ext = String(item.output_format || 'png').toLowerCase() === 'jpeg' ? 'jpg' : String(item.output_format || 'png').toLowerCase()
  return `sub2api-image-${String(index + 1).padStart(2, '0')}.${ext}`
}

function copyPrompt(item: ImageCreatorManagedImage): void {
  void copyToClipboard(item.task_prompt || item.revised_prompt || '')
}

function openChatImagesInNewTab(query: Record<string, string>): void {
  const resolved = router.resolve({
    path: '/chat-images',
    query,
  })
  window.open(resolved.href, '_blank', 'noopener,noreferrer')
}

function reusePrompt(item: ImageCreatorManagedImage): void {
  const prompt = item.task_prompt || item.revised_prompt || ''
  openChatImagesInNewTab(prompt ? { prompt, mode: 'image' } : { mode: 'image' })
}

function useAsReference(item: ImageCreatorManagedImage): void {
  const prompt = item.task_prompt || item.revised_prompt || ''
  openChatImagesInNewTab({
    mode: 'image',
    reference_image_id: String(item.id),
    ...(prompt ? { prompt } : {}),
  })
}

function openPreview(item: ImageCreatorManagedImage): void {
  previewImage.value = item
  void hydrateImages([item])
}

function closePreview(): void {
  previewImage.value = null
}

function formatTime(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return new Intl.DateTimeFormat('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

function formatFileSize(value: number): string {
  const size = Number(value)
  if (!Number.isFinite(size) || size <= 0) return ''
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / 1024 / 1024).toFixed(1)} MB`
}

function dimensionSummary(item: ImageCreatorManagedImage): string {
  const width = Number(item.width)
  const height = Number(item.height)
  if (!Number.isFinite(width) || !Number.isFinite(height) || width <= 0 || height <= 0) {
    return ''
  }
  return [item.resolution || `${width}x${height}`, item.aspect_ratio || '', formatMegapixels(width, height)]
    .filter(Boolean)
    .join(' · ')
}

function formatMegapixels(width: number, height: number): string {
  const mp = width * height / 1_000_000
  if (!Number.isFinite(mp) || mp <= 0) return ''
  return `${mp >= 10 ? mp.toFixed(1) : mp.toFixed(2)}MP`
}

function revokeObjectUrls(): void {
  for (const url of objectUrls) {
    URL.revokeObjectURL(url)
  }
  objectUrls.clear()
}
</script>

<style scoped>
.scheme3-image-manager {
  --image-library-card: #fffefa;
  --image-library-subtle: #f1eee6;
  --image-library-ink: #27251f;
  --image-library-muted: #777266;
  --image-library-soft: #a49e90;
  --image-library-line: #d8d2c3;
  --image-library-accent: #1e5c42;
  --image-library-accent-hover: #174a35;
  --image-library-primary-fg: #fffefa;
  --image-library-amber: #a56613;
  --image-library-danger: #9e4d3d;
  display: flex;
  min-height: calc(100vh - 7rem);
  flex-direction: column;
  gap: .9rem;
  color: var(--image-library-ink);
}

.image-manager-header {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto auto;
  align-items: end;
  gap: .9rem;
  border-bottom: 1px solid var(--image-library-line);
  padding: .15rem 0 1rem;
}

.image-manager-kicker,
.image-manager-filter-heading > span,
.image-manager-toolbar-summary > span {
  margin: 0;
  color: var(--image-library-muted);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: .6rem;
  font-weight: 800;
}

.image-manager-title h1 {
  margin: .32rem 0 0;
  color: var(--image-library-ink);
  font-family: Georgia, 'Times New Roman', serif;
  font-size: 2rem;
  font-weight: 500;
  line-height: 1.08;
}

.image-manager-subtitle {
  max-width: 31rem;
  margin: .45rem 0 0;
  color: var(--image-library-muted);
  font-size: .74rem;
  line-height: 1.55;
}

.image-manager-ledger {
  display: flex;
  overflow: hidden;
  border: 1px solid var(--image-library-line);
  border-radius: 7px;
  background: var(--image-library-card);
}

.image-manager-ledger span {
  display: grid;
  min-width: 4.8rem;
  gap: .1rem;
  border-right: 1px solid var(--image-library-line);
  padding: .48rem .62rem;
  text-align: right;
}

.image-manager-ledger span:last-child { border-right: 0; }

.image-manager-ledger strong {
  overflow: hidden;
  color: var(--image-library-accent);
  font-family: Georgia, 'Times New Roman', serif;
  font-size: .86rem;
  font-weight: 600;
  line-height: 1.15;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.image-manager-ledger small {
  color: var(--image-library-muted);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: .52rem;
  font-weight: 700;
  white-space: nowrap;
}

.image-manager-actions,
.image-manager-selection,
.image-manager-card-actions {
  display: flex;
  align-items: center;
  gap: .45rem;
  flex-wrap: wrap;
}

.scheme3-image-manager :deep(.btn) {
  min-height: 2.18rem;
  border-radius: 7px;
  font-size: .7rem;
  font-weight: 800;
}

.scheme3-image-manager :deep(.btn-secondary) {
  border-color: var(--image-library-line);
  background: var(--image-library-card);
  color: var(--image-library-ink);
  box-shadow: none;
}

.scheme3-image-manager :deep(.btn-secondary:hover:not(:disabled)) {
  border-color: rgba(30,92,66,.3);
  background: var(--image-library-subtle);
  color: var(--image-library-accent);
}

.scheme3-image-manager :deep(.btn-primary) {
  border-color: var(--image-library-accent);
  background: var(--image-library-accent);
  color: var(--image-library-primary-fg);
  box-shadow: 0 8px 17px rgba(30,92,66,.14);
}

.scheme3-image-manager :deep(.btn-primary:hover:not(:disabled)) { background: var(--image-library-accent-hover); }

.image-manager-toolbar {
  display: flex;
  min-height: 3.55rem;
  align-items: center;
  justify-content: space-between;
  gap: .8rem;
  border: 1px solid var(--image-library-line);
  border-radius: 8px;
  background: var(--image-library-card);
  box-shadow: 0 9px 21px rgba(54,48,34,.045);
  padding: .65rem .85rem;
}

.image-manager-toolbar-summary { display: grid; gap: .12rem; }

.image-manager-toolbar-summary strong {
  color: var(--image-library-ink);
  font-family: Georgia, 'Times New Roman', serif;
  font-size: .92rem;
  font-weight: 500;
}

.image-manager-selection {
  border: 1px solid rgba(30,92,66,.2);
  border-radius: 7px;
  background: rgba(30,92,66,.075);
  padding: .32rem .38rem .32rem .58rem;
  color: var(--image-library-accent);
  font-size: .68rem;
  font-weight: 800;
}

.image-manager-text-button,
.image-manager-icon-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: .27rem;
  border: 1px solid transparent;
  border-radius: 6px;
  color: var(--image-library-muted);
  transition: color 150ms ease, background-color 150ms ease, border-color 150ms ease, transform 150ms ease;
}

.image-manager-text-button { min-height: 1.72rem; padding: .18rem .42rem; font-size: .65rem; font-weight: 800; }
.image-manager-text-button:hover:not(:disabled) { background: var(--image-library-card); color: var(--image-library-accent); }
.image-manager-text-button:active,.image-manager-icon-button:active { transform: scale(.96); }
.image-manager-text-button.danger,.image-manager-icon-button.danger { color: var(--image-library-danger); }

.image-manager-filters {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: .7rem;
  border: 1px solid var(--image-library-line);
  border-radius: 8px;
  background: var(--image-library-card);
  box-shadow: 0 9px 21px rgba(54,48,34,.045);
  padding: .9rem;
}

.image-manager-filter-heading {
  display: grid;
  grid-column: 1 / -1;
  gap: .18rem;
  border-bottom: 1px solid var(--image-library-line);
  padding: .02rem 0 .7rem;
}

.image-manager-filter-heading strong {
  color: var(--image-library-ink);
  font-family: Georgia, 'Times New Roman', serif;
  font-size: .9rem;
  font-weight: 500;
}

.image-manager-field {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: .34rem;
  color: var(--image-library-ink);
  font-size: .65rem;
  font-weight: 800;
}

.image-manager-field-wide { grid-column: span 2; }

.scheme3-image-manager :deep(.input) {
  min-height: 2.35rem;
  border-color: var(--image-library-line);
  border-radius: 7px;
  background: var(--image-library-card);
  color: var(--image-library-ink);
  box-shadow: none;
  font-size: .73rem;
}

.scheme3-image-manager :deep(.input::placeholder) { color: var(--image-library-soft); }
.scheme3-image-manager :deep(.input:focus) { border-color: var(--image-library-accent); box-shadow: 0 0 0 3px rgba(30,92,66,.11); }

.image-manager-filter-actions {
  display: flex;
  align-items: end;
  gap: .42rem;
  flex-wrap: wrap;
}

.image-manager-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: .8rem;
}

.image-manager-card {
  position: relative;
  overflow: hidden;
  border: 1px solid var(--image-library-line);
  border-radius: 8px;
  background: var(--image-library-card);
  box-shadow: 0 9px 20px rgba(54,48,34,.06);
  transition: transform 160ms ease, box-shadow 160ms ease, border-color 160ms ease;
}

.image-manager-card:hover {
  border-color: rgba(30,92,66,.34);
  box-shadow: 0 15px 26px rgba(54,48,34,.1);
  transform: translateY(-2px);
}

.image-manager-card.selected {
  border-color: var(--image-library-accent);
  box-shadow: 0 0 0 2px rgba(30,92,66,.15), 0 13px 25px rgba(54,48,34,.09);
}

.image-manager-preview {
  display: flex;
  width: 100%;
  aspect-ratio: 1 / 1;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid var(--image-library-line);
  background: var(--image-library-subtle);
}

.image-manager-preview img { height: 100%; width: 100%; object-fit: contain; }

.image-manager-select {
  position: absolute;
  left: .56rem;
  top: .56rem;
  z-index: 2;
  display: flex;
  height: 1.66rem;
  width: 1.66rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(255,254,250,.72);
  border-radius: 6px;
  background: rgba(39,37,31,.65);
  color: #fffefa;
  transition: background-color 150ms ease, transform 150ms ease;
}

.image-manager-select:hover { transform: scale(1.05); }
.image-manager-select.active { border-color: var(--image-library-accent); background: var(--image-library-accent); }

.image-manager-card-body { display: flex; flex-direction: column; gap: .58rem; padding: .76rem; }

.image-manager-format {
  border: 1px solid rgba(30,92,66,.18);
  border-radius: 999px;
  background: rgba(30,92,66,.07);
  padding: .14rem .42rem;
  color: var(--image-library-accent);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: .59rem;
  font-weight: 800;
}

.image-manager-prompt {
  min-height: 2.55rem;
  display: -webkit-box;
  overflow: hidden;
  color: var(--image-library-ink);
  font-family: Georgia, 'Times New Roman', serif;
  font-size: .86rem;
  font-weight: 500;
  line-height: 1.48;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.image-manager-meta {
  display: flex;
  justify-content: space-between;
  gap: .5rem;
  border-top: 1px dashed var(--image-library-line);
  padding-top: .48rem;
  color: var(--image-library-muted);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: .57rem;
  font-weight: 700;
}

.image-manager-meta span { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.image-manager-card-actions { justify-content: flex-end; }

.image-manager-icon-button {
  height: 1.95rem;
  width: 1.95rem;
  border-color: var(--image-library-line);
  background: var(--image-library-card);
}

.image-manager-icon-button:hover:not(:disabled) { border-color: rgba(30,92,66,.28); background: var(--image-library-subtle); color: var(--image-library-accent); }
.image-manager-icon-button.danger:hover:not(:disabled) { border-color: rgba(158,77,61,.28); background: rgba(158,77,61,.08); color: var(--image-library-danger); }

.image-manager-state {
  display: flex;
  min-height: 25rem;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: .62rem;
  border: 1px dashed var(--image-library-line);
  border-radius: 8px;
  background: var(--image-library-card);
  color: var(--image-library-muted);
}

.image-manager-state h2 { margin: .22rem 0 0; color: var(--image-library-ink); font-family: Georgia, 'Times New Roman', serif; font-size: 1.04rem; font-weight: 500; }
.image-manager-state p { margin: 0; color: var(--image-library-muted) !important; }
.image-manager-footer { display: flex; justify-content: center; padding: .15rem 0 1rem; }

.image-manager-lightbox {
  position: fixed;
  inset: 0;
  z-index: 70;
  display: flex;
  flex-direction: column;
  background: rgba(22,21,15,.88);
  padding: 1rem;
  backdrop-filter: blur(8px);
}

.image-manager-lightbox :deep(.btn-secondary),
.image-manager-lightbox-close {
  border: 1px solid rgba(255,254,250,.28);
  border-radius: 7px;
  background: #fffefa !important;
  color: #27251f !important;
}

.image-manager-lightbox-close { display: flex; height: 2.2rem; width: 2.2rem; align-items: center; justify-content: center; }
.image-manager-lightbox :deep(img) { border: 1px solid rgba(255,254,250,.2); border-radius: 8px !important; }

:global(.dark .scheme3-image-manager) {
  --image-library-card: #24231f;
  --image-library-subtle: #2b2924;
  --image-library-ink: #f4f2ec;
  --image-library-muted: #aaa69a;
  --image-library-soft: #827e72;
  --image-library-line: #47443a;
  --image-library-accent: #8fc2a5;
  --image-library-accent-hover: #a7cfb3;
  --image-library-primary-fg: #1b1b18;
  --image-library-amber: #d3a55a;
  --image-library-danger: #d38b79;
}

@media (max-width: 1100px) {
  .image-manager-header { grid-template-columns: minmax(0, 1fr) auto; }
  .image-manager-actions { grid-column: 1 / -1; justify-content: flex-end; }
  .image-manager-filters { grid-template-columns: repeat(3, minmax(0, 1fr)); }
}

@media (max-width: 720px) {
  .image-manager-header { grid-template-columns: 1fr; align-items: stretch; gap: .72rem; margin-bottom: 0; padding-bottom: .82rem; }
  .image-manager-title h1 { font-size: 1.62rem; }
  .image-manager-ledger { width: 100%; }
  .image-manager-ledger span { flex: 1 1 0; min-width: 0; padding: .46rem .38rem; }
  .image-manager-actions { grid-column: auto; justify-content: stretch; }
  .image-manager-actions :deep(.btn) { flex: 1 1 0; justify-content: center; }
  .image-manager-toolbar { align-items: stretch; flex-direction: column; }
  .image-manager-selection { justify-content: space-between; }
  .image-manager-filters { grid-template-columns: repeat(2, minmax(0, 1fr)); gap: .6rem; padding: .72rem; }
  .image-manager-field-wide,.image-manager-filter-heading { grid-column: 1 / -1; }
  .image-manager-filter-actions { grid-column: 1 / -1; }
  .image-manager-filter-actions :deep(.btn) { flex: 1 1 0; justify-content: center; }
  .image-manager-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); gap: .65rem; }
  .image-manager-card-body { padding: .62rem; }
  .image-manager-meta { display: grid; grid-template-columns: 1fr; gap: .22rem; }
  .image-manager-card-actions { gap: .32rem; }
  .image-manager-icon-button { height: 1.82rem; width: 1.82rem; }
}

@media (max-width: 380px) {
  .image-manager-grid { grid-template-columns: 1fr; }
  .image-manager-selection { align-items: stretch; flex-direction: column; }
  .image-manager-selection > span { padding: .1rem .2rem; }
}

@media (prefers-reduced-motion: reduce) {
  .scheme3-image-manager *, .scheme3-image-manager *::before, .scheme3-image-manager *::after { animation-duration: .001ms !important; transition-duration: .001ms !important; }
}
</style>

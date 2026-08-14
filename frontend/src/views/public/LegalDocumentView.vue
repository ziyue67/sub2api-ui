<template>
  <div class="scheme3-legal-document">
    <header class="scheme3-legal-header">
      <div class="scheme3-legal-header-inner">
        <RouterLink to="/home" class="scheme3-legal-brand">
          <template v-if="settings">
            <span class="scheme3-legal-logo">
              <img :src="siteLogo || '/logo.svg'" alt="Logo" class="h-full w-full object-contain" />
            </span>
            <span class="scheme3-legal-brand-copy">
              <small>规则文档 / 服务说明</small>
              <strong>{{ siteName }}</strong>
            </span>
          </template>
          <template v-else>
            <span class="scheme3-legal-logo scheme3-legal-skeleton" aria-hidden="true"></span>
            <span class="scheme3-legal-brand-skeleton" aria-hidden="true"></span>
          </template>
        </RouterLink>
        <RouterLink
          to="/login"
          class="scheme3-legal-login"
        >
          <Icon name="login" size="sm" />
          {{ t('home.login') }}
        </RouterLink>
      </div>
    </header>

    <main class="scheme3-legal-main">
      <div v-if="loading" class="scheme3-legal-loading" role="status">
        <span class="scheme3-legal-spinner" aria-hidden="true"></span>
      </div>

      <section
        v-else-if="loadError"
        class="scheme3-legal-state scheme3-legal-state-error"
      >
        <h1 class="text-lg font-semibold">{{ t('legal.loadFailed') }}</h1>
        <p class="mt-2 text-sm">{{ t('legal.retryLater') }}</p>
      </section>

      <section
        v-else-if="!currentDocument"
        class="scheme3-legal-state"
      >
        <div class="scheme3-legal-state-row">
          <span class="scheme3-legal-state-icon">
            <Icon name="document" size="sm" />
          </span>
          <div>
            <h1>{{ t('legal.notFound') }}</h1>
            <p>
              {{ t('legal.notFoundDescription') }}
            </p>
          </div>
        </div>
      </section>

      <article v-else class="scheme3-legal-sheet">
        <div class="scheme3-legal-sheet-head">
          <div class="scheme3-legal-title-row">
            <span class="scheme3-legal-document-icon">
              <Icon :name="documentIcon" size="md" />
            </span>
            <div class="scheme3-legal-title-copy">
              <p class="scheme3-legal-kicker">{{ documentTypeLabel }}</p>
              <h1>
                {{ currentDocument.title }}
              </h1>
              <p v-if="updatedAt" class="scheme3-legal-updated">
                {{ t('legal.updatedAt', { date: updatedAt }) }}
              </p>
            </div>
          </div>
        </div>

        <div
          v-if="hasContent"
          class="legal-document-content scheme3-legal-content"
          v-html="renderedHtml"
        ></div>
        <div
          v-else
          class="scheme3-legal-empty"
        >
          {{ t('legal.empty') }}
        </div>
      </article>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { getLocale } from '@/i18n'
import { sanitizeUrl } from '@/utils/url'
import { resolveDisplaySiteName } from '@/utils/branding'
import { useAppStore } from '@/stores/app'
import type { LoginAgreementDocument } from '@/types'
import zhAdminCompliance from '../../../../docs/legal/admin-compliance.zh.md?raw'
import enAdminCompliance from '../../../../docs/legal/admin-compliance.en.md?raw'

type LegalDocumentIcon = 'document' | 'shield' | 'globe' | 'cog'

const route = useRoute()
const { t } = useI18n()
const appStore = useAppStore()
const settings = computed(() => appStore.cachedPublicSettings)
const loading = ref(!settings.value)
const loadError = ref(false)

marked.setOptions({
  breaks: true,
  gfm: true,
})

const documentId = computed(() => String(route.params.documentId || ''))
const isAdminComplianceDocument = computed(() => documentId.value === 'admin-compliance')
const documents = computed(() => settings.value?.login_agreement_documents ?? [])
const siteName = computed(() => resolveDisplaySiteName(settings.value?.site_name))
const siteLogo = computed(() => sanitizeUrl(settings.value?.site_logo || '', {
  allowRelative: true,
  allowDataUrl: true,
}))
const updatedAt = computed(() =>
  isAdminComplianceDocument.value ? '' : settings.value?.login_agreement_updated_at || ''
)
const documentTypeLabel = computed(() =>
  isAdminComplianceDocument.value ? t('legal.adminCompliance') : t('legal.loginAgreement')
)

const currentDocument = computed<LoginAgreementDocument | null>(() => {
  if (isAdminComplianceDocument.value) {
    return {
      id: 'admin-compliance',
      title: t('adminCompliance.title'),
      content_md: getLocale() === 'zh' ? zhAdminCompliance : enAdminCompliance
    }
  }
  const id = documentId.value
  if (!id) {
    return null
  }
  return documents.value.find((doc) => doc.id === id) ?? null
})

const hasContent = computed(() => Boolean(currentDocument.value?.content_md?.trim()))

const renderedHtml = computed(() => {
  const content = currentDocument.value?.content_md?.trim() || ''
  if (!content) {
    return ''
  }
  const html = marked.parse(content) as string
  return DOMPurify.sanitize(html)
})

const documentIcon = computed<LegalDocumentIcon>(() => {
  const title = currentDocument.value?.title || ''
  if (title.includes('政策') || title.includes('隐私')) {
    return 'shield'
  }
  if (title.includes('国家') || title.includes('地区')) {
    return 'globe'
  }
  if (title.includes('特定')) {
    return 'cog'
  }
  return 'document'
})

onMounted(async () => {
  loadError.value = false
  const loadedSettings = await appStore.fetchPublicSettings()
  if (!loadedSettings) {
    loadError.value = true
  }
  loading.value = false
})
</script>

<style scoped>
.scheme3-legal-document {
  --legal-paper: #f4f2ec;
  --legal-sheet: #fffefa;
  --legal-ink: #24231f;
  --legal-muted: #777266;
  --legal-line: #d8d2c3;
  --legal-green: #1e5c42;
  --legal-green-soft: #e3ece5;
  min-height: 100vh;
  background: var(--legal-paper);
  color: var(--legal-ink);
}

.scheme3-legal-header {
  border-bottom: 1px solid var(--legal-line);
  background: rgba(244, 242, 236, .96);
}

.scheme3-legal-header-inner {
  display: flex;
  width: min(100% - 2rem, 68rem);
  min-height: 5.15rem;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin: 0 auto;
}

.scheme3-legal-brand {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: .7rem;
  color: inherit;
  text-decoration: none;
}

.scheme3-legal-logo {
  display: inline-flex;
  width: 2.55rem;
  height: 2.55rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border: 1px solid rgba(30, 92, 66, .32);
  border-radius: 7px;
  background: var(--legal-sheet);
  box-shadow: 0 6px 16px rgba(54, 48, 34, .08);
}

.scheme3-legal-brand-copy {
  display: grid;
  min-width: 0;
  gap: .12rem;
}

.scheme3-legal-brand-copy small,
.scheme3-legal-kicker {
  color: var(--legal-muted);
  font-size: .67rem;
  font-weight: 700;
  line-height: 1.35;
}

.scheme3-legal-brand-copy strong {
  overflow: hidden;
  color: var(--legal-ink);
  font-family: Georgia, 'Times New Roman', serif;
  font-size: 1.12rem;
  font-weight: 600;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.scheme3-legal-skeleton,
.scheme3-legal-brand-skeleton {
  animation: scheme3-legal-pulse 1.5s ease-in-out infinite;
  background: #e2ddd1;
}

.scheme3-legal-brand-skeleton {
  width: 8.5rem;
  height: 1rem;
  border-radius: 3px;
}

.scheme3-legal-login {
  display: inline-flex;
  min-height: 2.35rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  gap: .4rem;
  border: 1px solid var(--legal-green);
  border-radius: 6px;
  background: var(--legal-green);
  color: #fffefa;
  padding: .48rem .82rem;
  font-size: .78rem;
  font-weight: 700;
  line-height: 1;
  text-decoration: none;
  transition: background-color 150ms ease, transform 150ms ease;
}

.scheme3-legal-login:hover { background: #287052; }
.scheme3-legal-login:active { transform: scale(.98); }

.scheme3-legal-main {
  width: min(100% - 2rem, 58rem);
  min-height: calc(100vh - 5.15rem);
  margin: 0 auto;
  padding: clamp(2rem, 5vw, 4.5rem) 0;
}

.scheme3-legal-loading {
  display: flex;
  min-height: 20rem;
  align-items: center;
  justify-content: center;
}

.scheme3-legal-spinner {
  width: 2rem;
  height: 2rem;
  border: 2px solid var(--legal-line);
  border-top-color: var(--legal-green);
  border-radius: 50%;
  animation: scheme3-legal-spin .7s linear infinite;
}

.scheme3-legal-state,
.scheme3-legal-sheet {
  border: 1px solid var(--legal-line);
  border-radius: 8px;
  background: var(--legal-sheet);
  box-shadow: 0 18px 42px rgba(54, 48, 34, .08);
}

.scheme3-legal-state { padding: 1.5rem; }
.scheme3-legal-state-error { border-color: #c98772; background: #fff8f4; color: #7f3328; }
.scheme3-legal-state-row { display: flex; align-items: flex-start; gap: .8rem; }
.scheme3-legal-state-icon,
.scheme3-legal-document-icon {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
}
.scheme3-legal-state-icon { width: 2.3rem; height: 2.3rem; background: #ebe7dc; color: var(--legal-green); }
.scheme3-legal-state h1 { margin: 0; color: var(--legal-ink); font-family: Georgia, 'Times New Roman', serif; font-size: 1.2rem; font-weight: 600; }
.scheme3-legal-state p { margin: .45rem 0 0; color: var(--legal-muted); font-size: .85rem; line-height: 1.7; }
.scheme3-legal-state-error h1,
.scheme3-legal-state-error p { color: inherit; }

.scheme3-legal-sheet { position: relative; overflow: hidden; padding: clamp(1.35rem, 4vw, 2.8rem); }
.scheme3-legal-sheet::before { position: absolute; top: 0; bottom: 0; left: clamp(1rem, 3vw, 2rem); width: 1px; background: rgba(30, 92, 66, .13); content: ''; pointer-events: none; }
.scheme3-legal-sheet-head,
.scheme3-legal-content,
.scheme3-legal-empty { position: relative; z-index: 1; }
.scheme3-legal-sheet-head { border-bottom: 1px solid var(--legal-line); padding-bottom: 1.45rem; }
.scheme3-legal-title-row { display: flex; align-items: flex-start; gap: .9rem; }
.scheme3-legal-document-icon { width: 2.75rem; height: 2.75rem; border: 1px solid rgba(30, 92, 66, .26); background: var(--legal-green-soft); color: var(--legal-green); }
.scheme3-legal-title-copy { min-width: 0; }
.scheme3-legal-title-copy h1 { margin: .22rem 0 0; overflow-wrap: anywhere; color: var(--legal-ink); font-family: Georgia, 'Times New Roman', serif; font-size: clamp(1.75rem, 4vw, 2.55rem); font-weight: 600; line-height: 1.18; }
.scheme3-legal-updated { margin: .65rem 0 0; color: var(--legal-muted); font-size: .76rem; line-height: 1.5; }
.scheme3-legal-empty { border: 1px dashed var(--legal-line); border-radius: 6px; margin-top: 1.65rem; padding: 3.4rem 1.25rem; color: var(--legal-muted); font-size: .86rem; text-align: center; }

.legal-document-content {
  margin-top: 1.7rem;
  line-height: 1.82;
  overflow-wrap: anywhere;
  color: var(--legal-ink);
}

.legal-document-content :deep(h1) {
  margin: 2.3rem 0 1rem;
  border-bottom: 1px solid var(--legal-line);
  padding-bottom: .75rem;
  color: var(--legal-ink);
  font-family: Georgia, 'Times New Roman', serif;
  font-size: clamp(1.6rem, 3vw, 2rem);
  font-weight: 600;
  line-height: 1.25;
}

.legal-document-content :deep(h2) {
  margin: 2rem 0 .75rem;
  color: var(--legal-ink);
  font-family: Georgia, 'Times New Roman', serif;
  font-size: clamp(1.35rem, 2.5vw, 1.65rem);
  font-weight: 600;
  line-height: 1.3;
}

.legal-document-content :deep(h3) {
  margin: 1.6rem 0 .6rem;
  color: var(--legal-ink);
  font-size: 1.1rem;
  font-weight: 700;
  line-height: 1.45;
}

.legal-document-content :deep(h4) {
  margin: 1.4rem 0 .5rem;
  color: var(--legal-ink);
  font-size: 1rem;
  font-weight: 700;
}

.legal-document-content :deep(p) {
  margin: 0 0 1rem;
  color: var(--legal-ink);
  font-size: .92rem;
}

.legal-document-content :deep(a) {
  color: var(--legal-green);
  text-decoration: underline;
  text-underline-offset: 3px;
}

.legal-document-content :deep(ul) {
  margin: 0 0 1rem;
  padding-left: 1.35rem;
  list-style: disc;
}

.legal-document-content :deep(ol) {
  margin: 0 0 1rem;
  padding-left: 1.35rem;
  list-style: decimal;
}

.legal-document-content :deep(li) {
  margin-bottom: .35rem;
  color: var(--legal-ink);
  font-size: .92rem;
}

.legal-document-content :deep(blockquote) {
  margin: 1.4rem 0;
  border-left: 3px solid var(--legal-green);
  padding-left: 1rem;
  color: var(--legal-muted);
}

.legal-document-content :deep(code) {
  border-radius: 4px;
  background: #eeeae0;
  padding: .12rem .32rem;
  color: var(--legal-ink);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: .82em;
}

.legal-document-content :deep(pre) {
  margin: 1.4rem 0;
  overflow-x: auto;
  border-radius: 6px;
  background: #27261f;
  padding: 1rem;
  color: #f4f2ec;
}

.legal-document-content :deep(pre code) {
  background: transparent;
  padding: 0;
  color: inherit;
}

.legal-document-content :deep(table) {
  display: block;
  width: 100%;
  margin: 1.4rem 0;
  overflow-x: auto;
  border-collapse: collapse;
}

.legal-document-content :deep(th) {
  border: 1px solid var(--legal-line);
  background: #f1eee6;
  padding: .55rem .7rem;
  color: var(--legal-ink);
  font-size: .84rem;
  font-weight: 700;
  text-align: left;
}

.legal-document-content :deep(td) {
  border: 1px solid var(--legal-line);
  padding: .55rem .7rem;
  color: var(--legal-ink);
  font-size: .84rem;
}

.legal-document-content :deep(img) {
  display: block;
  max-width: 100%;
  height: auto;
  margin: 1.4rem 0;
  border-radius: 6px;
}

.legal-document-content :deep(hr) {
  margin: 1.8rem 0;
  border: 0;
  border-top: 1px solid var(--legal-line);
}

:global(html.dark .scheme3-legal-document) { --legal-paper: #1b1b18; --legal-sheet: #24231f; --legal-ink: #f4f2ec; --legal-muted: #aaa69a; --legal-line: #47443a; --legal-green: #8fc2a5; --legal-green-soft: #293a31; }
:global(html.dark .scheme3-legal-header) { background: rgba(27, 27, 24, .96); }
:global(html.dark .scheme3-legal-logo) { border-color: rgba(143, 194, 165, .28); box-shadow: 0 6px 16px rgba(0, 0, 0, .24); }
:global(html.dark .scheme3-legal-skeleton),
:global(html.dark .scheme3-legal-brand-skeleton) { background: #34322c; }
:global(html.dark .scheme3-legal-login) { border-color: var(--legal-green); background: var(--legal-green); color: #1b1b18; }
:global(html.dark .scheme3-legal-login:hover) { background: #a7d2b7; }
:global(html.dark .scheme3-legal-state),
:global(html.dark .scheme3-legal-sheet) { box-shadow: 0 20px 44px rgba(0, 0, 0, .28); }
:global(html.dark .scheme3-legal-state-error) { border-color: #8f5c52; background: #321f1b; color: #efb8ad; }
:global(html.dark .scheme3-legal-state-icon) { background: #2b2924; }
:global(html.dark .scheme3-legal-document .legal-document-content code) { background: #302e29; }
:global(html.dark .scheme3-legal-document .legal-document-content th) { background: #2b2924; }

@keyframes scheme3-legal-spin { to { transform: rotate(360deg); } }
@keyframes scheme3-legal-pulse { 50% { opacity: .5; } }

@media (max-width: 620px) {
  .scheme3-legal-header-inner,
  .scheme3-legal-main { width: min(100% - 1.4rem, 58rem); }
  .scheme3-legal-header-inner { min-height: 4.6rem; }
  .scheme3-legal-brand-copy small { font-size: .59rem; }
  .scheme3-legal-brand-copy strong { font-size: 1rem; }
  .scheme3-legal-login { min-height: 2.2rem; padding: .45rem .65rem; }
  .scheme3-legal-sheet { padding: 1.15rem; }
  .scheme3-legal-sheet::before { left: .62rem; }
  .scheme3-legal-title-row { gap: .65rem; }
  .scheme3-legal-document-icon { width: 2.35rem; height: 2.35rem; }
  .scheme3-legal-content { margin-top: 1.25rem; }
  .legal-document-content :deep(p),
  .legal-document-content :deep(li) { font-size: .88rem; }
}
</style>

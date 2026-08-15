<template>
  <!-- 管理员配置的首页内容保持独立渲染，避免影响外部页面或自定义 HTML。 -->
  <div v-if="hasHomeContent" class="home-custom">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <div v-else v-html="homeContent"></div>
  </div>

  <div
    v-else-if="compactHomeEnabled"
    data-testid="compact-home"
    class="home-shell home-shell-compact"
  >
    <header class="home-nav">
      <nav class="home-nav-inner">
        <a href="#top" class="home-brand" :aria-label="t('home.scheme3.enterAccount')">
          <span class="home-brand-mark">
            <img v-if="siteLogo" :src="siteLogo" alt="" />
            <span v-else class="home-brand-monogram" aria-hidden="true">ST</span>
          </span>
          <span class="home-brand-copy">
            <span class="home-brand-name">{{ siteName }}</span>
            <span class="home-brand-caption">{{ t('home.scheme3.brandCaptionCompact') }}</span>
          </span>
        </a>
        <div class="home-nav-actions">
          <LocaleSwitcher />
          <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="home-icon-button" :title="t('home.viewDocs')">
            <Icon name="book" size="sm" />
          </a>
          <button
            type="button"
            class="home-icon-button"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
            :aria-pressed="isDark"
            @click="toggleTheme"
          >
            <Icon v-if="isDark" name="sun" size="sm" />
            <Icon v-else name="moon" size="sm" />
          </button>
          <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="home-nav-link">
            <Icon :name="isAuthenticated ? 'home' : 'login'" size="sm" />
            <span>{{ isAuthenticated ? t('home.scheme3.enterWorkspace') : t('home.scheme3.login') }}</span>
          </router-link>
        </div>
      </nav>
    </header>

    <main id="top" class="home-compact-main">
      <p class="home-overline">{{ t('home.scheme3.overline') }}</p>
      <h1 class="home-title">{{ siteName }}</h1>
      <p class="home-lead">{{ t('home.scheme3.lead') }}</p>
      <p class="home-configured-subtitle">{{ siteSubtitle }}</p>
      <div class="home-actions">
        <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="home-primary-action">
          <span>{{ isAuthenticated ? t('home.scheme3.viewRecords') : t('home.scheme3.startUsing') }}</span>
          <Icon name="arrowRight" size="sm" />
        </router-link>
        <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="home-secondary-action">
          <Icon name="book" size="sm" />
          <span>{{ t('home.scheme3.readApi') }}</span>
        </a>
      </div>

      <div class="home-contact-strip" :aria-label="t('home.scheme3.contactLabel')">
        <span class="home-contact-label">{{ t('home.scheme3.contactLabel') }}</span>
        <strong>QQ群 808603209</strong>
        <span class="home-contact-divider" aria-hidden="true">·</span>
        <strong>QQ 776523718</strong>
      </div>

      <dl class="home-compact-index">
        <div>
          <dt>{{ t('home.scheme3.primaryEntry') }}</dt>
          <dd>/v1/messages</dd>
        </div>
        <div>
          <dt>{{ t('home.scheme3.routeMethod') }}</dt>
          <dd>{{ t('home.scheme3.routeFallback') }}</dd>
        </div>
        <div>
          <dt>{{ t('home.scheme3.recordStatus') }}</dt>
          <dd class="home-value-green">{{ t('home.scheme3.traceable') }}</dd>
        </div>
      </dl>
    </main>

    <footer class="home-footer">
      <span>&copy; {{ currentYear }} {{ siteName }} · {{ t('home.scheme3.publicInfo') }}</span>
      <span class="home-footer-links">
        <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">{{ t('home.scheme3.docs') }}</a>
        <a :href="githubUrl" target="_blank" rel="noopener noreferrer">{{ t('home.scheme3.code') }}</a>
      </span>
    </footer>
  </div>

  <div v-else id="top" class="home-shell" data-testid="scheme3-home">
    <header class="home-nav">
      <nav class="home-nav-inner">
        <a href="#top" class="home-brand" :aria-label="t('home.scheme3.enterAccount')">
          <span class="home-brand-mark">
            <img v-if="siteLogo" :src="siteLogo" alt="" />
            <span v-else class="home-brand-monogram" aria-hidden="true">ST</span>
          </span>
          <span class="home-brand-copy">
            <span class="home-brand-name">{{ siteName }}</span>
            <span class="home-brand-caption">{{ t('home.scheme3.brandCaption') }}</span>
          </span>
        </a>
        <div class="home-nav-actions">
          <LocaleSwitcher />
          <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="home-icon-button" :title="t('home.viewDocs')">
            <Icon name="book" size="sm" />
          </a>
          <button
            type="button"
            class="home-icon-button"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
            :aria-pressed="isDark"
            @click="toggleTheme"
          >
            <Icon v-if="isDark" name="sun" size="sm" />
            <Icon v-else name="moon" size="sm" />
          </button>
          <router-link v-if="isAuthenticated" :to="dashboardPath" class="home-user-link">
            <span class="home-user-mark">{{ userInitial }}</span>
            <span>{{ t('home.scheme3.enterWorkspace') }}</span>
            <Icon name="externalLink" size="xs" />
          </router-link>
          <router-link v-else to="/login" class="home-nav-link">
            <Icon name="login" size="sm" />
            <span>{{ t('home.scheme3.login') }}</span>
          </router-link>
        </div>
      </nav>
    </header>

    <main class="home-main">
      <section class="home-intro" aria-labelledby="home-title">
        <div>
          <p class="home-overline">{{ t('home.scheme3.overline') }}</p>
          <h1 id="home-title" class="home-title">{{ siteName }}</h1>
          <p class="home-lead">{{ t('home.scheme3.lead') }}</p>
          <p class="home-configured-subtitle">{{ siteSubtitle }}</p>
          <div class="home-actions">
            <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="home-primary-action">
              <span>{{ isAuthenticated ? t('home.scheme3.viewWorkspace') : t('home.scheme3.enterAccount') }}</span>
              <Icon name="arrowRight" size="sm" />
            </router-link>
            <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="home-secondary-action">
              <Icon name="book" size="sm" />
              <span>{{ t('home.scheme3.viewApi') }}</span>
            </a>
          </div>
        </div>

        <div class="home-intro-notes" :aria-label="t('home.scheme3.summaryLabel')">
          <div><span>{{ t('home.scheme3.entry') }}</span><strong>/v1/messages</strong></div>
          <div><span>{{ t('home.scheme3.routing') }}</span><strong>{{ t('home.scheme3.routePriority') }}</strong></div>
          <div><span>{{ t('home.scheme3.usage') }}</span><strong>{{ t('home.scheme3.records') }}</strong></div>
        </div>

        <div class="home-contact-strip" :aria-label="t('home.scheme3.contactLabel')">
          <span class="home-contact-label">{{ t('home.scheme3.contactLabel') }}</span>
          <strong>QQ群 808603209</strong>
          <span class="home-contact-divider" aria-hidden="true">·</span>
          <strong>QQ 776523718</strong>
        </div>
      </section>

        <section class="home-ledger" aria-labelledby="ledger-title">
          <header class="home-section-heading">
            <div>
            <p class="home-overline">{{ t('home.scheme3.runLog') }}</p>
            <h2 id="ledger-title" class="home-section-title">{{ t('home.scheme3.activity') }}</h2>
          </div>
          <span class="home-section-meta home-record-meta" :class="{ 'home-record-meta-active': isAuthenticated }">
            <span class="home-live-marker" aria-hidden="true"></span>
            {{ isAuthenticated ? t('home.scheme3.accountScope') : t('home.scheme3.publicScope') }}
          </span>
        </header>
        <div class="home-ledger-table" role="table" :aria-label="t('home.scheme3.activity')">
          <div class="home-ledger-row home-ledger-head" role="row">
            <span>{{ t('home.scheme3.time') }}</span><span>{{ t('home.scheme3.entry') }}</span><span>{{ t('home.scheme3.purpose') }}</span><span>{{ t('home.scheme3.status') }}</span>
          </div>
          <div
            v-for="entry in ledgerEntries"
            :key="entry.time"
            class="home-ledger-row"
            role="row"
            @animationend="clearEntryAnimation"
          >
            <span class="home-mono">{{ entry.time }}</span>
            <span class="home-mono">{{ entry.route }}</span>
            <span>{{ entry.label }}</span>
            <span class="home-status" :class="entry.tone">{{ entry.state }}</span>
          </div>
        </div>
      </section>

      <section class="home-detail-grid" :aria-label="t('home.scheme3.availableEntry')">
        <article class="home-detail-section">
          <header class="home-section-heading home-section-heading-compact">
            <div>
              <p class="home-overline">{{ t('home.scheme3.serviceEntry') }}</p>
              <h2 class="home-section-title">{{ t('home.scheme3.availableEntry') }}</h2>
            </div>
            <Icon name="link" size="sm" class="home-muted" />
          </header>
          <div class="home-route-list">
            <div v-for="route in routeRows" :key="route.path" class="home-route-row">
              <span class="home-route-mark" :class="route.tone"></span>
              <div>
                <p class="home-mono">{{ route.path }}</p>
                <p class="home-row-note">{{ route.note }}</p>
              </div>
              <span class="home-route-state">{{ route.state }}</span>
            </div>
          </div>
        </article>

        <article class="home-detail-section">
          <header class="home-section-heading home-section-heading-compact">
            <div>
              <p class="home-overline">{{ t('home.scheme3.runtime') }}</p>
              <h2 class="home-section-title">{{ t('home.scheme3.running') }}</h2>
            </div>
            <Icon name="infoCircle" size="sm" class="home-muted" />
          </header>
          <ul class="home-note-list">
            <li v-for="note in serviceNotes" :key="note.label">
              <span class="home-note-index">{{ note.index }}</span>
              <span><strong>{{ note.label }}</strong>{{ note.text }}</span>
            </li>
          </ul>
        </article>
      </section>
    </main>

    <footer class="home-footer">
      <span>&copy; {{ currentYear }} {{ siteName }} · {{ t('home.scheme3.publicInfo') }}</span>
      <span class="home-footer-links">
        <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">{{ t('home.scheme3.docs') }}</a>
        <a :href="githubUrl" target="_blank" rel="noopener noreferrer">{{ t('home.scheme3.code') }}</a>
      </span>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import { sanitizeUrl } from '@/utils/url'
import { resolveDisplaySiteName } from '@/utils/branding'

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

const siteName = computed(() => {
  const configuredName = appStore.cachedPublicSettings?.site_name || appStore.siteName
  return resolveDisplaySiteName(configuredName)
})
const siteLogo = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || t('auth.scheme3.subtitle'))
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const hasHomeContent = computed(() => homeContent.value.trim().length > 0)
const compactHomeEnabled = computed(() => appStore.cachedPublicSettings?.compact_home_enabled === true)

const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isDark = ref(document.documentElement.classList.contains('dark'))
const githubUrl = 'https://github.com/ShourGG/sub2api'
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')
const userInitial = computed(() => {
  const user = authStore.user
  if (!user || !user.email) return ''
  return user.email.charAt(0).toUpperCase()
})
const currentYear = computed(() => new Date().getFullYear())

const ledgerEntries = computed(() => [
  { time: '09:42', route: '/v1/messages', label: t('home.scheme3.messages'), state: t('home.scheme3.normal'), tone: 'home-status-green' },
  { time: '09:18', route: '/v1/chat/completions', label: t('home.scheme3.compatibleRequest'), state: t('home.scheme3.traceable'), tone: 'home-status-muted' },
  { time: '08:55', route: '/v1/models', label: t('home.scheme3.modelCatalog'), state: t('home.scheme3.readable'), tone: 'home-status-amber' },
  { time: '08:31', route: '/v1/messages', label: t('home.scheme3.poolSwitch'), state: t('home.scheme3.switched'), tone: 'home-status-green' }
])

const routeRows = computed(() => [
  { path: '/v1/messages', note: t('home.scheme3.routeMessagesNote'), state: t('home.scheme3.routeMain'), tone: 'home-route-green' },
  { path: '/v1/chat/completions', note: t('home.scheme3.routeCompletionsNote'), state: t('home.scheme3.routeCompatible'), tone: 'home-route-amber' },
  { path: '/v1/models', note: t('home.scheme3.routeModelsNote'), state: t('home.scheme3.routeCatalog'), tone: 'home-route-muted' }
])

const serviceNotes = computed(() => [
  { index: '01', label: t('home.scheme3.routeOnline'), text: ` ${t('home.scheme3.routeOnlineText')}` },
  { index: '02', label: t('home.scheme3.channelSwitch'), text: ` ${t('home.scheme3.channelSwitchText')}` },
  { index: '03', label: t('home.scheme3.records'), text: ` ${t('home.scheme3.recordsText')}` }
])

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (savedTheme === 'dark' || (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

function clearEntryAnimation(event: AnimationEvent) {
  const target = event.currentTarget
  if (target instanceof HTMLElement && event.target === target) {
    target.classList.add('home-row-ready')
    target.removeAttribute('style')
  }
}

onMounted(() => {
  initTheme()
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>

<style scoped>
.home-shell {
  --home-paper: #f4f2ec;
  --home-card: #fbfaf6;
  --home-line: #dad5c8;
  --home-ink: #16150f;
  --home-muted: #6b695f;
  --home-green: #1e5c42;
  --home-amber: #b7791f;
  min-height: 100vh;
  overflow: hidden;
  scroll-behavior: smooth;
  background: var(--home-paper);
  color: var(--home-ink);
  font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}

.home-shell::before {
  position: fixed;
  inset: 0.75rem;
  border: 1px solid color-mix(in srgb, var(--home-line) 72%, transparent);
  content: "";
  pointer-events: none;
  z-index: 0;
  animation: home-frame-breathe 6s ease-in-out infinite;
}

.home-shell::after {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 1px;
  background: var(--home-green);
  content: "";
  opacity: 0.18;
  pointer-events: none;
  transform: translateY(-8vh);
  animation: home-scan 9s linear infinite;
  z-index: 0;
}

@keyframes home-fade-down {
  from { opacity: 0; transform: translateY(-0.7rem); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes home-rise-in {
  from { opacity: 0; transform: translateY(1rem); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes home-draw-in {
  from { opacity: 0; transform: scaleX(0.97); transform-origin: left center; }
  to { opacity: 1; transform: scaleX(1); transform-origin: left center; }
}

@keyframes home-marker-pulse {
  0%, 100% { opacity: 0.45; }
  50% { opacity: 1; }
}

@keyframes home-frame-breathe {
  0%, 100% { opacity: 0.72; }
  50% { opacity: 1; }
}

@keyframes home-scan {
  from { transform: translateY(-8vh); }
  to { transform: translateY(108vh); }
}

@keyframes home-contact-sweep {
  0% { transform: translateX(-125%); }
  42%, 100% { transform: translateX(720%); }
}

@keyframes home-sheen {
  from { transform: translateX(-140%) skewX(-18deg); }
  to { transform: translateX(320%) skewX(-18deg); }
}

@keyframes home-panel-signal {
  0%, 100% { opacity: 0; transform: translateX(-110%); }
  18%, 68% { opacity: 0.65; }
  82% { opacity: 0; transform: translateX(110%); }
}

@keyframes home-brand-signal {
  0%, 100% { box-shadow: 0 0 0 0 color-mix(in srgb, var(--home-green) 0%, transparent); }
  48% { box-shadow: 0 0 0 0.35rem color-mix(in srgb, var(--home-green) 11%, transparent); }
}

@keyframes home-status-pulse {
  0%, 100% { opacity: 0.5; transform: scale(0.84); }
  50% { opacity: 1; transform: scale(1); }
}

.home-shell .home-nav {
  animation: home-fade-down 560ms cubic-bezier(0.22, 1, 0.36, 1) both;
}

.home-shell:not(.home-shell-compact) .home-intro {
  animation: home-rise-in 650ms 80ms cubic-bezier(0.22, 1, 0.36, 1) both;
}

.home-shell:not(.home-shell-compact) .home-intro-notes > div {
  opacity: 0;
  animation: home-draw-in 440ms cubic-bezier(0.22, 1, 0.36, 1) both;
}

.home-shell:not(.home-shell-compact) .home-intro-notes > div:nth-child(1) { animation-delay: 220ms; }
.home-shell:not(.home-shell-compact) .home-intro-notes > div:nth-child(2) { animation-delay: 300ms; }
.home-shell:not(.home-shell-compact) .home-intro-notes > div:nth-child(3) { animation-delay: 380ms; }

.home-shell:not(.home-shell-compact) .home-ledger {
  animation: home-rise-in 620ms 220ms cubic-bezier(0.22, 1, 0.36, 1) both;
}

.home-shell:not(.home-shell-compact) .home-detail-grid {
  animation: home-rise-in 620ms 360ms cubic-bezier(0.22, 1, 0.36, 1) both;
}

.home-shell:not(.home-shell-compact) .home-footer {
  animation: home-rise-in 520ms 480ms cubic-bezier(0.22, 1, 0.36, 1) both;
}

.home-shell-compact .home-compact-main > * {
  opacity: 0;
  animation: home-rise-in 560ms cubic-bezier(0.22, 1, 0.36, 1) both;
}

.home-shell-compact .home-compact-main > *:nth-child(1) { animation-delay: 80ms; }
.home-shell-compact .home-compact-main > *:nth-child(2) { animation-delay: 140ms; }
.home-shell-compact .home-compact-main > *:nth-child(3) { animation-delay: 210ms; }
.home-shell-compact .home-compact-main > *:nth-child(4) { animation-delay: 280ms; }
.home-shell-compact .home-compact-main > *:nth-child(5) { animation-delay: 350ms; }
.home-shell-compact .home-compact-main > *:nth-child(6) { animation-delay: 420ms; }

.home-nav,
.home-main,
.home-footer {
  position: relative;
  z-index: 1;
}

.home-nav {
  z-index: 3;
  border-bottom: 1px solid var(--home-line);
  padding: 1rem clamp(1.25rem, 5vw, 4rem);
}

.home-nav-inner {
  display: flex;
  max-width: 74rem;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin: 0 auto;
}

.home-brand {
  display: inline-flex;
  min-width: 0;
  align-items: center;
  gap: 0.7rem;
  color: inherit;
  text-decoration: none;
}

.home-brand-mark {
  position: relative;
  display: inline-flex;
  width: 2.65rem;
  height: 2.65rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border: 1px solid var(--home-line);
  background: var(--home-card);
  animation: home-brand-signal 4.8s ease-in-out 800ms infinite;
}

.home-brand-mark::after {
  position: absolute;
  inset: 0.28rem;
  border: 1px solid color-mix(in srgb, var(--home-green) 42%, transparent);
  content: "";
  opacity: 0;
  pointer-events: none;
  transform: scale(0.84);
  transition: opacity 180ms ease, transform 220ms cubic-bezier(0.22, 1, 0.36, 1);
}

.home-brand:hover .home-brand-mark::after {
  opacity: 1;
  transform: scale(1);
}

.home-brand-mark img {
  width: 100%;
  height: 100%;
  object-fit: contain;
  transition: transform 260ms cubic-bezier(0.22, 1, 0.36, 1);
}

.home-brand-monogram {
  position: relative;
  z-index: 1;
  color: var(--home-green);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: .68rem;
  font-weight: 900;
  letter-spacing: .08em;
}

.home-brand:hover .home-brand-mark img {
  transform: translateY(-2px) rotate(-2deg);
}

.home-brand-copy {
  display: grid;
  min-width: 0;
  gap: 0.22rem;
}

.home-brand-name {
  overflow: hidden;
  font-family: Georgia, "Times New Roman", serif;
  font-size: 1.25rem;
  line-height: 1;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.home-brand-caption,
.home-overline,
.home-section-meta,
.home-mono,
.home-footer,
.home-compact-index dt,
.home-compact-index dd {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
}

.home-brand-caption,
.home-overline {
  color: var(--home-muted);
  font-size: 0.64rem;
  letter-spacing: 0.14em;
  line-height: 1.35;
  text-transform: uppercase;
}

.home-nav-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.45rem;
}

.home-icon-button,
.home-nav-link,
.home-user-link {
  display: inline-flex;
  min-height: 2.25rem;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  border: 1px solid transparent;
  color: var(--home-muted);
  padding: 0.45rem 0.6rem;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 0.7rem;
  text-decoration: none;
  transition: border-color 160ms ease, color 160ms ease, background 160ms ease, transform 160ms ease;
}

.home-icon-button { width: 2.25rem; }

.home-icon-button:hover,
.home-nav-link:hover,
.home-user-link:hover {
  border-color: var(--home-line);
  background: var(--home-card);
  color: var(--home-green);
}

.home-icon-button:active,
.home-nav-link:active,
.home-user-link:active,
.home-primary-action:active,
.home-secondary-action:active {
  transform: translateY(1px);
}

.home-icon-button svg,
.home-nav-link svg,
.home-user-link svg {
  transition: transform 180ms ease;
}

.home-icon-button:hover svg { transform: translateY(-1px); }
.home-nav-link:hover svg { transform: translateX(2px); }
.home-user-link:hover svg { transform: translate(2px, -2px); }

.home-nav-link {
  border-color: var(--home-green);
  color: var(--home-green);
}

.home-user-link {
  border-color: var(--home-line);
  color: var(--home-ink);
}

.home-user-mark {
  display: inline-flex;
  width: 1.35rem;
  height: 1.35rem;
  align-items: center;
  justify-content: center;
  background: var(--home-ink);
  color: var(--home-paper);
  font-size: 0.65rem;
}

.home-main {
  max-width: 74rem;
  margin: 0 auto;
  padding: clamp(2.5rem, 8vw, 6rem) clamp(1.25rem, 5vw, 4rem) 4rem;
}

.home-intro {
  display: grid;
  grid-template-columns: minmax(0, 1.2fr) minmax(15rem, 0.8fr);
  gap: clamp(2rem, 7vw, 6rem);
  align-items: end;
  border-bottom: 1px solid var(--home-line);
  padding-bottom: clamp(2rem, 6vw, 4rem);
}

.home-title {
  max-width: none;
  margin: 0.65rem 0 0;
  font-family: Georgia, "Times New Roman", serif;
  font-size: 5.6rem;
  font-weight: 400;
  letter-spacing: 0;
  line-height: 0.98;
  white-space: nowrap;
}

.home-lead {
  max-width: 40rem;
  margin: 1.2rem 0 0;
  color: var(--home-muted);
  font-size: clamp(0.95rem, 2vw, 1.15rem);
  line-height: 1.7;
}

.home-configured-subtitle {
  max-width: 40rem;
  margin: 0.7rem 0 0;
  color: var(--home-muted);
  font-size: 0.78rem;
  line-height: 1.5;
}

.home-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.65rem;
  margin-top: 1.6rem;
}

.home-primary-action,
.home-secondary-action {
  position: relative;
  overflow: hidden;
  display: inline-flex;
  min-height: 2.8rem;
  align-items: center;
  justify-content: center;
  gap: 0.55rem;
  border: 1px solid var(--home-green);
  padding: 0.7rem 1rem;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 0.74rem;
  letter-spacing: 0.04em;
  text-decoration: none;
  transition: background 160ms ease, color 160ms ease, border-color 160ms ease, transform 160ms ease;
}

.home-primary-action::after,
.home-secondary-action::after {
  position: absolute;
  top: -35%;
  left: 0;
  width: 1.7rem;
  height: 170%;
  background: color-mix(in srgb, white 24%, transparent);
  content: "";
  opacity: 0;
  pointer-events: none;
  transform: translateX(-140%) skewX(-18deg);
}

.home-primary-action:hover,
.home-secondary-action:hover {
  transform: translateY(-2px);
}

.home-primary-action:hover::after,
.home-secondary-action:hover::after {
  opacity: 1;
  animation: home-sheen 720ms ease-out both;
}

.home-primary-action {
  background: var(--home-green);
  color: var(--home-paper);
}

.home-primary-action:hover {
  background: #174a35;
}

.home-primary-action svg {
  transition: transform 200ms cubic-bezier(0.22, 1, 0.36, 1);
}

.home-primary-action:hover svg { transform: translateX(4px); }

.home-secondary-action {
  border-color: var(--home-line);
  background: var(--home-card);
  color: var(--home-ink);
}

.home-secondary-action:hover {
  border-color: var(--home-green);
  color: var(--home-green);
}

.home-intro-notes {
  border-top: 1px solid var(--home-line);
}

.home-intro-notes > div {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px solid var(--home-line);
  padding: 0.85rem 0;
  font-size: 0.78rem;
  transition: background 160ms ease, padding-left 160ms ease;
}

.home-intro-notes > div:hover {
  background: color-mix(in srgb, var(--home-green) 5%, var(--home-card));
  padding-left: 0.45rem;
}

.home-intro-notes span { color: var(--home-muted); }
.home-intro-notes strong { font-weight: 500; text-align: right; }

.home-contact-strip {
  position: relative;
  display: flex;
  flex-wrap: wrap;
  grid-column: 1 / -1;
  align-items: baseline;
  gap: 0.55rem 1rem;
  margin-top: 1.25rem;
  border-top: 2px solid var(--home-green);
  border-bottom: 1px solid var(--home-line);
  padding: 0.95rem 0;
  color: var(--home-ink);
  overflow: hidden;
}

.home-contact-strip::after {
  position: absolute;
  bottom: -1px;
  left: 0;
  width: min(18%, 8rem);
  height: 2px;
  background: var(--home-amber);
  content: "";
  animation: home-contact-sweep 5s ease-in-out infinite;
}

.home-contact-label {
  color: var(--home-green);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 0.68rem;
  letter-spacing: 0.1em;
  text-transform: uppercase;
}

.home-contact-strip strong {
  font-size: 1rem;
  font-weight: 600;
  letter-spacing: 0.02em;
}

.home-contact-divider { color: var(--home-amber); }

.home-ledger,
.home-detail-section {
  position: relative;
  overflow: hidden;
  border: 1px solid var(--home-line);
  background: var(--home-card);
}

.home-ledger::after,
.home-detail-section::after {
  position: absolute;
  top: 0;
  bottom: 0;
  left: 0;
  width: 1px;
  background: var(--home-amber);
  content: "";
  opacity: 0;
  pointer-events: none;
  animation: home-panel-signal 7s ease-in-out 1.2s infinite;
}

.home-ledger {
  margin-top: 2.25rem;
}

.home-section-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px solid var(--home-line);
  padding: 1.15rem 1.25rem;
}

.home-section-heading-compact { padding: 1rem 1.1rem; }

.home-section-title {
  margin: 0.35rem 0 0;
  font-family: Georgia, "Times New Roman", serif;
  font-size: 1.45rem;
  font-weight: 400;
  line-height: 1.1;
}

.home-section-meta {
  color: var(--home-muted);
  font-size: 0.65rem;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.home-record-meta {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
}

.home-live-marker {
  width: 0.36rem;
  height: 0.36rem;
  background: currentColor;
}

.home-record-meta-active .home-live-marker {
  animation: home-marker-pulse 2.4s ease-in-out infinite;
}

.home-record-meta .home-live-marker {
  animation: home-marker-pulse 2.4s ease-in-out infinite;
}

.home-ledger-table { display: grid; }

.home-ledger-row {
  position: relative;
  display: grid;
  grid-template-columns: 5rem minmax(10rem, 1.2fr) minmax(7rem, 1fr) 5rem;
  gap: 1rem;
  align-items: center;
  border-bottom: 1px solid var(--home-line);
  padding: 0.85rem 1.25rem;
  font-size: 0.8rem;
  transition: background-color 180ms ease, transform 180ms ease;
}

.home-ledger-row:last-child { border-bottom: 0; }

.home-ledger-row:not(.home-ledger-head)::before {
  position: absolute;
  top: 0;
  bottom: 0;
  left: 0;
  width: 2px;
  background: var(--home-green);
  content: "";
  transform: scaleY(0);
  transform-origin: center;
  transition: transform 180ms ease;
  z-index: 2;
}

.home-ledger-row:not(.home-ledger-head):hover {
  background: color-mix(in srgb, var(--home-green) 5%, var(--home-card));
  transform: translateX(2px);
}

:global(.home-row-ready) {
  opacity: 1 !important;
  animation: none !important;
}

.home-ledger-row:not(.home-ledger-head):hover::before { transform: scaleY(1); }

.home-shell:not(.home-shell-compact) .home-ledger-row:not(.home-ledger-head) {
  opacity: 0;
  animation: home-rise-in 420ms cubic-bezier(0.22, 1, 0.36, 1) both;
}

.home-shell:not(.home-shell-compact) .home-ledger-row:nth-child(2) { animation-delay: 320ms; }
.home-shell:not(.home-shell-compact) .home-ledger-row:nth-child(3) { animation-delay: 390ms; }
.home-shell:not(.home-shell-compact) .home-ledger-row:nth-child(4) { animation-delay: 460ms; }
.home-shell:not(.home-shell-compact) .home-ledger-row:nth-child(5) { animation-delay: 530ms; }

.home-ledger-head {
  color: var(--home-muted);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 0.64rem;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.home-status { font-size: 0.7rem; font-weight: 600; }
.home-status-green { color: var(--home-green); }
.home-status-amber { color: var(--home-amber); }
.home-status-muted { color: var(--home-muted); }

.home-status-green {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
}

.home-status-green::before {
  width: 0.34rem;
  height: 0.34rem;
  background: currentColor;
  content: "";
  animation: home-status-pulse 2.2s ease-in-out infinite;
}

.home-detail-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1.25rem;
  margin-top: 1.25rem;
}

.home-route-list,
.home-note-list { display: grid; }

.home-route-row {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  gap: 0.75rem;
  align-items: start;
  border-bottom: 1px solid var(--home-line);
  padding: 0.95rem 1.1rem;
  transition: background-color 180ms ease, padding-left 180ms ease;
}

.home-route-row:last-child { border-bottom: 0; }
.home-route-row:hover {
  background: color-mix(in srgb, var(--home-green) 5%, var(--home-card));
  padding-left: 1.25rem;
}

.home-route-row:active,
.home-note-list li:active,
.home-intro-notes > div:active {
  transform: translateX(2px) scale(0.995);
}

.home-route-mark {
  width: 0.35rem;
  height: 2rem;
  background: var(--home-muted);
  transform-origin: center;
  transition: transform 180ms ease;
}

.home-route-row:hover .home-route-mark { transform: scaleY(1.18); }

.home-shell:not(.home-shell-compact) .home-route-row,
.home-shell:not(.home-shell-compact) .home-note-list li {
  opacity: 0;
  animation: home-rise-in 420ms cubic-bezier(0.22, 1, 0.36, 1) both;
}

.home-shell:not(.home-shell-compact) .home-route-row:nth-child(1),
.home-shell:not(.home-shell-compact) .home-note-list li:nth-child(1) { animation-delay: 500ms; }
.home-shell:not(.home-shell-compact) .home-route-row:nth-child(2),
.home-shell:not(.home-shell-compact) .home-note-list li:nth-child(2) { animation-delay: 570ms; }
.home-shell:not(.home-shell-compact) .home-route-row:nth-child(3),
.home-shell:not(.home-shell-compact) .home-note-list li:nth-child(3) { animation-delay: 640ms; }
.home-route-green { background: var(--home-green); }
.home-route-mark.home-route-green { animation: home-status-pulse 3s ease-in-out infinite; }
.home-route-amber { background: var(--home-amber); }
.home-route-muted { background: var(--home-line); }
.home-row-note { margin: 0.35rem 0 0; color: var(--home-muted); font-size: 0.75rem; line-height: 1.45; }
.home-route-state { color: var(--home-muted); font-size: 0.7rem; }
.home-muted { color: var(--home-muted); }

.home-note-list li {
  display: grid;
  grid-template-columns: 2rem minmax(0, 1fr);
  gap: 0.7rem;
  border-bottom: 1px solid var(--home-line);
  padding: 0.95rem 1.1rem;
  font-size: 0.8rem;
  line-height: 1.55;
  transition: background-color 180ms ease, padding-left 180ms ease;
}

.home-note-list li:last-child { border-bottom: 0; }
.home-note-index { color: var(--home-amber); font-family: ui-monospace, monospace; font-size: 0.7rem; }
.home-note-list strong { font-weight: 600; }

.home-note-list li:hover {
  background: color-mix(in srgb, var(--home-amber) 5%, var(--home-card));
  padding-left: 1.25rem;
}

.home-footer {
  display: flex;
  max-width: 74rem;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin: 0 auto;
  border-top: 1px solid var(--home-line);
  color: var(--home-muted);
  padding: 1.15rem clamp(1.25rem, 5vw, 4rem) 2rem;
  font-size: 0.66rem;
  line-height: 1.5;
}

.home-footer-links { display: inline-flex; gap: 1rem; }
.home-footer a { color: inherit; text-decoration: underline; text-underline-offset: 0.2em; }
.home-footer a:hover { color: var(--home-green); }

.home-shell-compact .home-compact-main {
  display: flex;
  min-height: calc(100vh - 13rem);
  max-width: 50rem;
  flex-direction: column;
  justify-content: center;
  margin: 0 auto;
  padding: 4rem 1.25rem;
}

.home-shell-compact .home-contact-strip {
  grid-column: auto;
  margin-top: 1.75rem;
}

.home-shell-compact .home-title { max-width: none; font-size: 5.8rem; }
.home-compact-index { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); margin: 3rem 0 0; border-top: 1px solid var(--home-line); border-bottom: 1px solid var(--home-line); }
.home-compact-index > div { min-width: 0; padding: 1rem 0.75rem; border-right: 1px solid var(--home-line); }
.home-compact-index > div:first-child { padding-left: 0; }
.home-compact-index > div:last-child { border-right: 0; }
.home-compact-index dt { color: var(--home-muted); font-size: 0.64rem; letter-spacing: 0.08em; text-transform: uppercase; }
.home-compact-index dd { margin: 0.45rem 0 0; overflow-wrap: anywhere; font-size: 0.75rem; }
.home-value-green { color: var(--home-green); }

.home-custom { min-height: 100vh; background: var(--home-paper, #f4f2ec); }

.home-shell a:focus-visible,
.home-shell button:focus-visible {
  outline: 2px solid var(--home-green);
  outline-offset: 3px;
}

:global(html.dark .home-shell) {
  --home-paper: #1b1b18;
  --home-card: #24231f;
  --home-line: #47443a;
  --home-ink: #f4f2ec;
  --home-muted: #aaa69a;
  --home-green: #8fc2a5;
  --home-amber: #d3a55a;
}

:global(html.dark .home-shell-compact .home-primary-action),
:global(html.dark .home-primary-action) { color: #1b1b18; }

@media (max-width: 760px) {
  .home-nav-inner { align-items: flex-start; }
  .home-nav-actions { flex-wrap: wrap; }
  .home-intro { grid-template-columns: 1fr; gap: 2rem; }
  .home-title { font-size: 4rem; }
  .home-shell-compact .home-title { font-size: 4.1rem; }
  .home-detail-grid { grid-template-columns: 1fr; }
  .home-ledger-row { grid-template-columns: 4rem minmax(0, 1fr) 4.5rem; }
  .home-ledger-row > span:nth-child(3) { display: none; }
  .home-ledger-head > span:nth-child(3) { display: none; }
}

@media (max-width: 480px) {
  .home-shell::before { inset: 0.45rem; }
  .home-nav { padding: 0.8rem 1rem; }
  .home-nav-inner { display: block; }
  .home-nav-actions { justify-content: flex-start; margin-top: 0.75rem; }
  .home-brand-name { max-width: 12rem; }
  .home-main { padding: 2.5rem 1rem 3rem; }
  .home-title,
  .home-shell-compact .home-title { font-size: 2.75rem; }
  .home-ledger-row { gap: 0.55rem; padding: 0.75rem 0.8rem; font-size: 0.72rem; }
  .home-section-heading { padding: 0.95rem 0.8rem; }
  .home-route-row,
  .home-note-list li { padding-left: 0.8rem; padding-right: 0.8rem; }
  .home-footer { align-items: flex-start; flex-direction: column; padding: 1rem 1rem 1.5rem; }
  .home-compact-index { grid-template-columns: 1fr; }
  .home-compact-index > div { border-right: 0; border-bottom: 1px solid var(--home-line); padding-left: 0; }
  .home-compact-index > div:last-child { border-bottom: 0; }
}

@media (max-width: 360px) {
  .home-title,
  .home-shell-compact .home-title { font-size: 2.25rem; }
}

@media (prefers-reduced-motion: reduce) {
  .home-shell {
    scroll-behavior: auto;
  }

  .home-shell *,
  .home-shell *::before,
  .home-shell *::after,
  .home-shell::before,
  .home-shell::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    scroll-behavior: auto !important;
    transition-duration: 0.01ms !important;
  }
}
</style>

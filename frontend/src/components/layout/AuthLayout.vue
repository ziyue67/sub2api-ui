<template>
  <div class="auth-shell" :data-settings-ready="settingsLoaded">
    <div class="auth-rule auth-rule-top" aria-hidden="true"></div>
    <div class="auth-wrap">
      <header class="auth-brand">
        <div class="auth-brand-mark">
          <img :src="siteLogo || '/logo.svg'" alt="Logo" />
        </div>
        <div class="min-w-0">
          <p class="auth-kicker">{{ t('auth.scheme3.kicker') }}</p>
          <h1 class="auth-brand-name">{{ siteName }}</h1>
          <p class="auth-brand-subtitle">{{ siteSubtitle }}</p>
        </div>
      </header>

      <main class="auth-card">
        <slot />
      </main>

      <footer class="auth-footer">
        <slot name="footer" />
      </footer>

      <p class="auth-copyright">&copy; {{ currentYear }} {{ siteName }} · {{ t('auth.scheme3.copyright') }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'
import { resolveDisplaySiteName } from '@/utils/branding'

const { t } = useI18n()
const appStore = useAppStore()

const siteName = computed(() => resolveDisplaySiteName(appStore.siteName))
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || t('auth.scheme3.subtitle'))
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)

const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  appStore.fetchPublicSettings()
})
</script>

<style scoped>
@keyframes auth-page-enter {
  from {
    opacity: 0;
    transform: translateY(0.65rem);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes auth-brand-enter {
  from { opacity: 0; transform: translateY(0.45rem); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes auth-card-enter {
  from { opacity: 0; transform: translateY(0.75rem); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes auth-scan {
  from { transform: translateY(-8vh); }
  to { transform: translateY(108vh); }
}

@keyframes auth-rule-sweep {
  0%, 20% { opacity: 0; transform: translateX(-260%); }
  45%, 75% { opacity: 1; transform: translateX(0); }
  100% { opacity: 0; transform: translateX(260%); }
}

@keyframes auth-card-signal {
  0%, 100% { opacity: 0; transform: translateX(-120%); }
  18%, 70% { opacity: 0.7; }
  86% { opacity: 0; transform: translateX(320%); }
}

@keyframes auth-mark-breathe {
  0%, 100% { box-shadow: 0 0 0 0 color-mix(in srgb, var(--auth-green) 0%, transparent); }
  50% { box-shadow: 0 0 0 0.35rem color-mix(in srgb, var(--auth-green) 10%, transparent); }
}

.auth-shell {
  --auth-paper: #f4f2ec;
  --auth-card: #fbfaf6;
  --auth-line: #dad5c8;
  --auth-ink: #16150f;
  --auth-muted: #6b695f;
  --auth-green: #1e5c42;
  position: relative;
  display: flex;
  min-height: 100vh;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  background: var(--auth-paper);
  color: var(--auth-ink);
  padding: 2rem 1rem;
  font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  animation: auth-page-enter 420ms cubic-bezier(0.22, 1, 0.36, 1) both;
}

.auth-shell::before {
  position: absolute;
  inset: 1rem;
  border: 1px solid color-mix(in srgb, var(--auth-line) 72%, transparent);
  content: "";
  pointer-events: none;
}

.auth-shell::after {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 1px;
  background: var(--auth-green);
  content: "";
  opacity: 0.16;
  pointer-events: none;
  transform: translateY(-8vh);
  animation: auth-scan 10s linear infinite;
  z-index: 0;
}

.auth-rule {
  position: absolute;
  left: 0;
  right: 0;
  height: 1px;
  background: var(--auth-line);
  opacity: 0.7;
}

.auth-rule-top {
  top: 4.5rem;
}

.auth-rule-top::after {
  position: absolute;
  top: -1px;
  right: 0;
  width: 4rem;
  height: 2px;
  background: var(--auth-green);
  content: "";
  opacity: 0;
  transform: translateX(-260%);
  animation: auth-rule-sweep 4.8s ease-in-out infinite;
}

.auth-wrap {
  position: relative;
  z-index: 1;
  width: min(100%, 31rem);
}

.auth-brand {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  margin: 0 auto 1.75rem;
  max-width: 27rem;
  animation: auth-brand-enter 520ms 80ms cubic-bezier(0.22, 1, 0.36, 1) both;
}

.auth-brand-mark {
  display: flex;
  width: 3.25rem;
  height: 3.25rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border: 1px solid var(--auth-line);
  background: var(--auth-card);
  animation: auth-mark-breathe 4s ease-in-out 700ms infinite;
  transition: border-color 160ms ease, transform 180ms ease;
}

.auth-brand:hover .auth-brand-mark {
  border-color: var(--auth-green);
  transform: translateY(-2px) rotate(-1deg);
}

.auth-brand-mark img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.auth-kicker,
.auth-copyright {
  margin: 0;
  color: var(--auth-muted);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 0.65rem;
  letter-spacing: 0.16em;
  line-height: 1.4;
  text-transform: uppercase;
}

.auth-brand-name {
  margin: 0.2rem 0 0;
  overflow: hidden;
  font-family: Georgia, "Times New Roman", serif;
  font-size: clamp(1.65rem, 6vw, 2.2rem);
  font-weight: 400;
  letter-spacing: 0;
  line-height: 1;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.auth-brand-subtitle {
  margin: 0.45rem 0 0;
  overflow: hidden;
  color: var(--auth-muted);
  font-size: 0.78rem;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.auth-card {
  position: relative;
  border: 1px solid var(--auth-line);
  background: var(--auth-card);
  padding: clamp(1.25rem, 5vw, 2rem);
  box-shadow: 0 1.25rem 3rem rgba(50, 46, 36, 0.08);
  animation: auth-card-enter 560ms 120ms cubic-bezier(0.22, 1, 0.36, 1) both;
}

.auth-card::after {
  position: absolute;
  top: -1px;
  left: 0;
  width: 5rem;
  height: 2px;
  background: var(--auth-green);
  content: "";
  opacity: 0;
  pointer-events: none;
  animation: auth-card-signal 6.5s ease-in-out 1.4s infinite;
}

.auth-card :deep(.auth-panel) {
  animation: auth-card-enter 520ms 220ms cubic-bezier(0.22, 1, 0.36, 1) both;
}

.auth-footer {
  min-height: 1.5rem;
  margin-top: 1.15rem;
  color: var(--auth-muted);
  font-size: 0.82rem;
  text-align: center;
  animation: auth-card-enter 520ms 300ms cubic-bezier(0.22, 1, 0.36, 1) both;
}

.auth-copyright {
  margin-top: 1.75rem;
  text-align: center;
  text-transform: none;
  animation: auth-card-enter 520ms 360ms cubic-bezier(0.22, 1, 0.36, 1) both;
}

/* Shared treatment for auth flows that use the generic input/button primitives. */
.auth-card :deep(.auth-default-panel) {
  --auth-alert-green: #e7f0e9;
  --auth-alert-amber: #fbf2e2;
  --auth-alert-red: #fff0ec;
}

.auth-card :deep(.auth-default-panel .text-gray-900),
.auth-card :deep(.auth-default-panel .text-gray-800),
.auth-card :deep(.auth-default-panel .text-gray-700) {
  color: var(--auth-ink) !important;
}

.auth-card :deep(.auth-default-panel .text-gray-600),
.auth-card :deep(.auth-default-panel .text-gray-500),
.auth-card :deep(.auth-default-panel .text-gray-400) {
  color: var(--auth-muted) !important;
}

.auth-card :deep(.auth-default-panel .input-label) {
  color: var(--auth-ink);
  font-size: .78rem;
  font-weight: 700;
}

.auth-card :deep(.auth-default-panel .input),
.auth-card :deep(.auth-default-panel .input[type='date']) {
  min-height: 2.75rem;
  border-color: var(--auth-line);
  border-radius: 6px;
  background: var(--auth-card);
  color: var(--auth-ink);
  box-shadow: none;
}

.auth-card :deep(.auth-default-panel .input:focus) {
  border-color: var(--auth-green);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--auth-green) 15%, transparent);
}

.auth-card :deep(.auth-default-panel .input:disabled) {
  background: #f1eee6;
  color: var(--auth-muted);
}

.auth-card :deep(.auth-default-panel .input-hint) { color: var(--auth-muted); }

.auth-card :deep(.auth-default-panel .btn) {
  min-height: 2.55rem;
  border-radius: 6px;
  box-shadow: none;
  font-size: .82rem;
  font-weight: 700;
}

.auth-card :deep(.auth-default-panel .btn-primary) {
  border-color: var(--auth-green);
  background: var(--auth-green);
  color: #fbfaf6;
}

.auth-card :deep(.auth-default-panel .btn-primary:hover:not(:disabled)) { background: #174a35; }
.auth-card :deep(.auth-default-panel .btn-secondary) { border-color: var(--auth-line); background: transparent; color: var(--auth-ink); }
.auth-card :deep(.auth-default-panel .btn-secondary:hover:not(:disabled)) { border-color: var(--auth-green); background: #f1eee6; color: var(--auth-green); }

.auth-card :deep(.auth-default-panel .bg-green-50) { border-color: color-mix(in srgb, var(--auth-green) 34%, var(--auth-line)) !important; background: var(--auth-alert-green) !important; }
.auth-card :deep(.auth-default-panel .bg-amber-50) { border-color: #e3bb72 !important; background: var(--auth-alert-amber) !important; }
.auth-card :deep(.auth-default-panel .bg-red-50) { border-color: #cf8c7d !important; background: var(--auth-alert-red) !important; }
.auth-card :deep(.auth-default-panel .text-green-800),
.auth-card :deep(.auth-default-panel .text-green-700),
.auth-card :deep(.auth-default-panel .text-green-600) { color: var(--auth-green) !important; }
.auth-card :deep(.auth-default-panel .text-amber-800),
.auth-card :deep(.auth-default-panel .text-amber-700) { color: #8b5b1a !important; }

.auth-footer :deep(.text-gray-500),
.auth-footer :deep(.text-gray-400) { color: var(--auth-muted) !important; }
.auth-footer :deep(.text-primary-600),
.auth-footer :deep(.text-primary-500) { color: var(--auth-green) !important; }

:global(html.dark .auth-card .auth-default-panel) {
  --auth-alert-green: #26382e;
  --auth-alert-amber: #392d1a;
  --auth-alert-red: #3c2420;
}

:global(html.dark .auth-card .auth-default-panel .input) { background: #2b2924; }
:global(html.dark .auth-card .auth-default-panel .input:disabled) { background: #2b2924; }
:global(html.dark .auth-card .auth-default-panel .btn-primary) { border-color: var(--auth-green); background: var(--auth-green); color: #1b1b18; }
:global(html.dark .auth-card .auth-default-panel .btn-primary:hover:not(:disabled)) { background: #a7d2b7; }
:global(html.dark .auth-card .auth-default-panel .btn-secondary:hover:not(:disabled)) { background: #2b2924; }
:global(html.dark .auth-card .auth-default-panel .text-green-800),
:global(html.dark .auth-card .auth-default-panel .text-green-700),
:global(html.dark .auth-card .auth-default-panel .text-green-600) { color: #8fc2a5 !important; }
:global(html.dark .auth-card .auth-default-panel .text-amber-800),
:global(html.dark .auth-card .auth-default-panel .text-amber-700) { color: #e7bf78 !important; }

:global(html.dark .auth-shell) {
  --auth-paper: #1b1b18;
  --auth-card: #24231f;
  --auth-line: #47443a;
  --auth-ink: #f4f2ec;
  --auth-muted: #aaa69a;
  --auth-green: #8fc2a5;
}

:global(html.dark .auth-card) {
  box-shadow: 0 1.25rem 3rem rgba(0, 0, 0, 0.24);
}

@media (max-width: 480px) {
  .auth-shell {
    align-items: flex-start;
    padding: 1.25rem 0.75rem 1.5rem;
  }

  .auth-shell::before {
    inset: 0.5rem;
  }

  .auth-rule-top {
    top: 3.25rem;
  }

  .auth-brand {
    margin-bottom: 1.25rem;
  }

  .auth-brand-mark {
    width: 2.75rem;
    height: 2.75rem;
  }

  .auth-brand-subtitle {
    max-width: 15rem;
  }
}

@media (prefers-reduced-motion: reduce) {
  .auth-shell,
  .auth-shell::before,
  .auth-shell::after,
  .auth-shell *,
  .auth-shell *::before,
  .auth-shell *::after {
    animation-duration: 1ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 1ms !important;
  }
}
</style>

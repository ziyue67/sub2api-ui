<template>
  <header
    class="scheme3-model-plaza-nav sticky top-0 z-30"
  >
    <div class="mx-auto flex max-w-7xl items-center justify-between gap-4 px-4 py-3.5 sm:px-6">
      <!-- 左:站点 logo + 名称 -->
      <div class="flex min-w-0 items-center gap-3">
        <template v-if="settings">
          <span
            class="flex h-9 w-9 flex-shrink-0 items-center justify-center overflow-hidden rounded-xl bg-white shadow-sm ring-1 ring-gray-200 dark:bg-dark-800 dark:ring-dark-700"
          >
            <img v-if="siteLogo" :src="siteLogo" alt="" class="h-full w-full object-contain" />
            <span v-else class="scheme3-model-plaza-monogram" aria-hidden="true">ST</span>
          </span>
          <span class="truncate text-base font-semibold text-gray-950 dark:text-white">
            {{ siteName }}
          </span>
        </template>
        <template v-else>
          <span class="h-9 w-9 flex-shrink-0 animate-pulse rounded-xl bg-gray-200 dark:bg-dark-700" aria-hidden="true"></span>
          <span class="h-5 w-28 animate-pulse rounded bg-gray-200 dark:bg-dark-700" aria-hidden="true"></span>
        </template>
      </div>

      <!-- 右:登录 / 回到后台 -->
      <RouterLink
        v-if="isAuthenticated"
        :to="backTarget"
        class="scheme3-model-plaza-nav-action"
      >
        {{ t('modelPlaza.nav.backToDashboard') }}
      </RouterLink>
      <RouterLink
        v-else
        :to="{ path: '/login', query: { redirect: '/model-plaza' } }"
        class="scheme3-model-plaza-nav-action"
      >
        {{ t('modelPlaza.nav.login') }}
      </RouterLink>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { sanitizeUrl } from '@/utils/url'
import { resolveDisplaySiteName } from '@/utils/branding'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

const settings = computed(() => appStore.cachedPublicSettings)
const siteName = computed(() => resolveDisplaySiteName(settings.value?.site_name))
const siteLogo = computed(() =>
  sanitizeUrl(settings.value?.site_logo || '', { allowRelative: true, allowDataUrl: true })
)
const isAuthenticated = computed(() => authStore.isAuthenticated)
const backTarget = computed(() => (authStore.isAdmin ? '/admin/dashboard' : '/dashboard'))
</script>

<style scoped>
.scheme3-model-plaza-nav { border-bottom: 1px solid #dad5c8; background: rgba(244,242,236,.94); backdrop-filter: blur(18px); }.scheme3-model-plaza-monogram { color: #1e5c42; font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; font-size: .62rem; font-weight: 900; letter-spacing: .08em; }.scheme3-model-plaza-nav-action { display: inline-flex; min-height: 2.25rem; align-items: center; justify-content: center; border: 1px solid #1e5c42; border-radius: 6px; padding: .4rem .7rem; background: #1e5c42; color: #f4f2ec; font-size: .68rem; font-weight: 800; transition: background-color 150ms ease,transform 150ms ease; }.scheme3-model-plaza-nav-action:hover { background: #174a35; }.scheme3-model-plaza-nav-action:active { transform: scale(.98); }
:global(html.dark) .scheme3-model-plaza-nav { border-color: #47443a; background: rgba(27,27,24,.94); }:global(html.dark) .scheme3-model-plaza-monogram { color: #8fc2a5; }:global(html.dark) .scheme3-model-plaza-nav-action { border-color: #8fc2a5; background: #1e5c42; color: #f4f2ec; }
</style>

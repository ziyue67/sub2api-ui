<template>
  <div class="scheme3-console-layout" :class="{ 'scheme3-console-layout-collapsed': navCollapsed }">
    <aside class="scheme3-console-sidebar" :class="{ 'scheme3-console-sidebar-open': mobileNavOpen }" aria-label="空间导航">
      <div class="scheme3-console-brand">
        <router-link to="/dashboard" class="scheme3-console-brand-mark" aria-label="返回控制台" @click="closeMobileNav">
          <img v-if="appStore.siteLogo" :src="appStore.siteLogo" alt="" />
          <span v-else>ST</span>
        </router-link>
        <div class="scheme3-console-brand-copy">
          <strong>Shour or ToKen</strong>
          <span>个人工作空间</span>
        </div>
        <button type="button" class="scheme3-console-close" aria-label="关闭导航" @click="closeMobileNav">
          <Icon name="x" size="sm" />
        </button>
      </div>

      <div class="scheme3-console-caption">空间导航</div>
      <nav class="scheme3-console-links">
        <section v-for="section in consoleNavSections" :key="section.id" class="scheme3-console-section">
          <div class="scheme3-console-section-label">{{ section.label }}</div>
          <router-link
            v-for="item in section.items"
            :key="item.path"
            :to="{ path: item.path, query: item.query }"
            class="scheme3-console-link"
            :class="{ 'scheme3-console-link-active': isNavActive(item.path) }"
            @click="closeMobileNav"
          >
            <span class="scheme3-console-link-icon"><Icon :name="item.icon" size="sm" /></span>
            <span class="scheme3-console-link-text">{{ item.label }}</span>
            <span v-if="isNavActive(item.path)" class="scheme3-console-current">当前</span>
          </router-link>
        </section>
      </nav>

      <div class="scheme3-console-foot">
        <div class="scheme3-console-account">
          <span class="scheme3-console-avatar">{{ userInitials }}</span>
          <div class="scheme3-console-account-copy"><strong>{{ userLabel }}</strong><span>已登录</span></div>
        </div>
        <button type="button" class="scheme3-console-action" @click="toggleTheme">
          <Icon :name="isDarkMode ? 'sun' : 'moon'" size="sm" />
          <span>{{ isDarkMode ? '切换浅色' : '切换深色' }}</span>
        </button>
        <button type="button" class="scheme3-console-action" @click="toggleNavCollapse">
          <Icon :name="navCollapsed ? 'chevronRight' : 'chevronLeft'" size="sm" />
          <span>{{ navCollapsed ? '展开导航' : '收起导航' }}</span>
        </button>
        <button type="button" class="scheme3-console-action scheme3-console-logout" @click="logout">
          <Icon name="login" size="sm" />
          <span>退出空间</span>
        </button>
      </div>
    </aside>

    <div v-if="mobileNavOpen" class="scheme3-console-overlay" @click="closeMobileNav"></div>

    <div class="scheme3-console-workspace">
      <header class="scheme3-console-topbar">
        <div class="scheme3-console-topbar-left">
          <button type="button" class="scheme3-console-menu-button" aria-label="打开导航" @click="openMobileNav">
            <Icon name="menu" size="md" />
          </button>
          <div>
            <span class="scheme3-console-topbar-kicker">SHOUR OR TOKEN / 运营空间</span>
            <strong>{{ pageTitle }}</strong>
          </div>
        </div>
        <div class="scheme3-console-topbar-right">
          <AnnouncementBell v-if="authStore.user" class="scheme3-console-tool" />
          <LocaleSwitcher class="scheme3-console-tool scheme3-console-locale-tool" />
          <span class="scheme3-console-balance"><Icon name="dollar" size="xs" />可用 {{ formatMoney(Number(user?.balance || 0)) }}</span>
          <span class="scheme3-console-status"><i></i>会话在线</span>
          <router-link to="/profile" class="scheme3-console-user" aria-label="打开个人资料">
            <span class="scheme3-console-avatar">{{ userInitials }}</span>
            <span>{{ userLabel }}</span>
          </router-link>
        </div>
      </header>

      <main class="scheme3-console-content">
        <div class="scheme3-console-page-frame">
          <slot />
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AnnouncementBell from '@/components/common/AnnouncementBell.vue'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore, useAuthStore } from '@/stores'
import { useBatchImageAccess } from '@/composables/useBatchImageAccess'
import type { CustomMenuItem } from '@/types'
import { FeatureFlags, isFeatureFlagEnabled } from '@/utils/featureFlags'

type ConsoleNavIcon = 'home' | 'key' | 'grid' | 'image' | 'chart' | 'chartBar' | 'server' | 'creditCard' | 'document' | 'gift' | 'users' | 'user'

interface ConsoleNavItem {
  path: string
  query?: Record<string, string>
  label: string
  icon: ConsoleNavIcon
}

interface ConsoleNavSection {
  id: string
  label: string
  items: ConsoleNavItem[]
}

const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const authStore = useAuthStore()
const { canUseBatchImage, refreshBatchImageAccess } = useBatchImageAccess()
const mobileNavOpen = ref(false)
const navCollapsed = ref(localStorage.getItem('scheme3-nav-collapsed') === '1')
const isDarkMode = ref(document.documentElement.classList.contains('dark'))

const user = computed(() => authStore.user)
const userLabel = computed(() => user.value?.username || user.value?.email?.split('@')[0] || '当前账号')
const userInitials = computed(() => userLabel.value.trim().slice(0, 2).toUpperCase() || 'ST')

const pageTitle = computed(() => {
  const titles: Record<string, string> = {
    '/dashboard': '控制台总览',
    '/scheme3-dashboard': '控制台总览',
    '/keys': '密钥中枢',
    '/model-square': '模型广场',
    '/model-plaza': '模型行情',
    '/canvas': '绘图工作站',
    '/chat': 'AI 聊天',
    '/chat-studio': 'AI 聊天',
    '/chat-images': '聊天生图',
    '/chat-images/native': '聊天生图',
    '/studio-bridge/launch': '聊天生图',
    '/image-creator': '图片创作',
    '/image-manager': '图片库',
    '/leaderboard': '总排行榜',
    '/docs/batch-image': '批量生图',
    '/batch-image': '批量生图',
    '/usage': '请求账本',
    '/available-channels': '可用节点',
    '/monitor': '渠道状态',
    '/subscriptions': '我的订阅',
    '/purchase': '购买订阅',
    '/orders': '订单记录',
    '/redeem': '兑换中心',
    '/affiliate': '推广中心',
    '/profile': '个人资料',
    '/payment/qrcode': '扫码支付',
    '/payment/stripe': '在线支付',
    '/payment/airwallex': '在线支付',
  }
  if (route.path.startsWith('/custom/')) return '扩展页面'
  return titles[route.path] || '控制台'
})

const consoleNavSections = computed<ConsoleNavSection[]>(() => {
  const simpleMode = authStore.isSimpleMode
  const workspace: ConsoleNavItem[] = [
    { path: '/dashboard', label: '运行总览', icon: 'home' },
    { path: '/keys', label: '密钥中枢', icon: 'key' },
    { path: '/model-square', label: '模型广场', icon: 'grid' },
    { path: '/canvas', label: '绘图工作站', icon: 'image' },
    { path: '/leaderboard', label: '总排行榜', icon: 'chart' },
  ]
  if (isFeatureFlagEnabled(FeatureFlags.modelPlaza)) {
    workspace.push({ path: '/model-plaza', query: { embedded: '1' }, label: '模型行情', icon: 'grid' })
  }

  const traffic: ConsoleNavItem[] = []
  if (!simpleMode && canUseBatchImage.value) traffic.push({ path: '/batch-image', label: '批量生图', icon: 'image' })
  if (!simpleMode) traffic.push({ path: '/usage', label: '请求账本', icon: 'chartBar' })
  if (!simpleMode && isFeatureFlagEnabled(FeatureFlags.availableChannels)) traffic.push({ path: '/available-channels', label: '可用节点', icon: 'server' })
  if (isFeatureFlagEnabled(FeatureFlags.channelMonitor)) traffic.push({ path: '/monitor', label: '渠道状态', icon: 'server' })

  const account: ConsoleNavItem[] = []
  if (!simpleMode) account.push({ path: '/subscriptions', label: '我的订阅', icon: 'creditCard' })
  if (!simpleMode && isFeatureFlagEnabled(FeatureFlags.payment)) {
    account.push({ path: '/purchase', label: '购买订阅', icon: 'creditCard' })
    account.push({ path: '/orders', label: '订单记录', icon: 'document' })
  }
  if (!simpleMode) account.push({ path: '/redeem', label: '兑换中心', icon: 'gift' })
  if (!simpleMode && isFeatureFlagEnabled(FeatureFlags.affiliate)) account.push({ path: '/affiliate', label: '推广中心', icon: 'users' })
  account.push({ path: '/profile', label: '个人资料', icon: 'user' })

  const customMenuItems = (appStore.cachedPublicSettings?.custom_menu_items ?? [])
    .filter((item: CustomMenuItem) => item.visibility === 'user')
    .sort((a: CustomMenuItem, b: CustomMenuItem) => a.sort_order - b.sort_order)
    .map((item: CustomMenuItem): ConsoleNavItem => ({ path: `/custom/${item.id}`, label: item.label, icon: 'grid' }))

  return [
    { id: 'workspace', label: '工作台', items: workspace },
    { id: 'traffic', label: '调用与资源', items: traffic },
    { id: 'account', label: '账户与订阅', items: account },
    { id: 'custom', label: '扩展入口', items: customMenuItems },
  ].filter((section) => section.items.length > 0)
})

function isNavActive(path: string) {
  return path === '/dashboard' ? route.path === path : route.path === path || route.path.startsWith(`${path}/`)
}

function openMobileNav() { mobileNavOpen.value = true }
function closeMobileNav() { mobileNavOpen.value = false }
function toggleNavCollapse() {
  navCollapsed.value = !navCollapsed.value
  localStorage.setItem('scheme3-nav-collapsed', navCollapsed.value ? '1' : '0')
}
function toggleTheme() {
  isDarkMode.value = !isDarkMode.value
  document.documentElement.classList.toggle('dark', isDarkMode.value)
  localStorage.setItem('theme', isDarkMode.value ? 'dark' : 'light')
}
function formatMoney(value: number) {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 2, maximumFractionDigits: 4 }).format(value)
}
async function logout() {
  await authStore.logout().catch(() => undefined)
  await router.push('/login')
}

watch(() => route.fullPath, closeMobileNav)
onMounted(() => {
  document.body.classList.add('scheme3-user-context')
  void refreshBatchImageAccess()
})
onBeforeUnmount(() => {
  document.body.classList.remove('scheme3-user-context')
})
</script>

<style scoped>
.scheme3-console-layout {
  --scheme3-ink: #16150f;
  --scheme3-muted: #6b695f;
  --scheme3-line: #dad5c8;
  --scheme3-paper: #f4f2ec;
  --scheme3-card: #fbfaf6;
  --scheme3-subtle: #f1eee6;
  display: grid;
  grid-template-columns: 15.5rem minmax(0, 1fr);
  min-height: 100vh;
  background: var(--scheme3-paper);
  color: var(--scheme3-ink);
  transition: grid-template-columns 220ms ease;
}
.scheme3-console-layout-collapsed { grid-template-columns: 5.2rem minmax(0, 1fr); }
.scheme3-console-sidebar { position: sticky; top: 0; z-index: 50; display: flex; height: 100vh; min-width: 0; flex-direction: column; border-right: 1px solid var(--scheme3-line); background: rgba(251,250,246,.93); backdrop-filter: blur(22px); }
.scheme3-console-brand { display: flex; min-height: 5.25rem; align-items: center; gap: .7rem; border-bottom: 1px solid var(--scheme3-line); padding: 1rem; }
.scheme3-console-brand-mark { display: inline-flex; width: 2.35rem; height: 2.35rem; flex-shrink: 0; align-items: center; justify-content: center; overflow: hidden; border: 1px solid rgba(30,92,66,.35); border-radius: 9px; background: #1e5c42; color: #f4f2ec; font-size: .68rem; font-weight: 800; letter-spacing: .08em; box-shadow: 0 8px 18px rgba(30,92,66,.18); }
.scheme3-console-brand-mark img { width: 100%; height: 100%; object-fit: contain; }
.scheme3-console-brand-copy { min-width: 0; }
.scheme3-console-brand-copy strong { display: block; overflow: hidden; color: var(--scheme3-ink); font-size: .76rem; font-weight: 800; letter-spacing: .02em; text-overflow: ellipsis; white-space: nowrap; }
.scheme3-console-brand-copy span { display: block; margin-top: .18rem; color: var(--scheme3-muted); font-size: .62rem; }
.scheme3-console-close { display: none; margin-left: auto; border: 0; background: transparent; color: var(--scheme3-muted); }
.scheme3-console-caption { padding: 1.2rem 1rem .5rem; color: var(--scheme3-muted); font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; font-size: .58rem; font-weight: 700; letter-spacing: .16em; }
.scheme3-console-links { display: flex; min-height: 0; flex: 1; flex-direction: column; gap: .8rem; overflow-y: auto; padding: 0 .65rem .9rem; }
.scheme3-console-section { display: flex; flex-direction: column; gap: .26rem; }
.scheme3-console-section-label { padding: .5rem .65rem .2rem; color: var(--scheme3-muted); font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; font-size: .54rem; font-weight: 700; letter-spacing: .13em; }
.scheme3-console-link { display: flex; min-width: 0; min-height: 2.5rem; align-items: center; gap: .65rem; border: 1px solid transparent; border-radius: 7px; padding: .55rem .65rem; color: var(--scheme3-muted); font-size: .72rem; font-weight: 700; transition: color 150ms ease,background-color 150ms ease,border-color 150ms ease,transform 150ms ease; }
.scheme3-console-link:hover { border-color: var(--scheme3-line); background: rgba(255,255,255,.72); color: var(--scheme3-ink); transform: translateX(2px); }
.scheme3-console-link:active { transform: translateX(2px) scale(.98); }
.scheme3-console-link-active { border-color: rgba(30,92,66,.28); background: rgba(30,92,66,.08); color: #1e5c42; box-shadow: inset 3px 0 0 #1e5c42; }
.scheme3-console-link-icon { display: inline-flex; flex-shrink: 0; color: currentColor; }
.scheme3-console-link-text { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.scheme3-console-current { margin-left: auto; border-radius: 999px; padding: .16rem .35rem; background: rgba(30,92,66,.1); color: #1e5c42; font-size: .54rem; font-weight: 800; }
.scheme3-console-foot { margin-top: auto; border-top: 1px solid var(--scheme3-line); padding: .8rem .65rem; }
.scheme3-console-account { display: flex; min-width: 0; align-items: center; gap: .55rem; padding: .45rem; }
.scheme3-console-avatar { display: inline-flex; width: 1.85rem; height: 1.85rem; flex-shrink: 0; align-items: center; justify-content: center; border-radius: 7px; background: #16150f; color: #f4f2ec; font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; font-size: .6rem; font-weight: 800; }
.scheme3-console-account-copy { min-width: 0; }
.scheme3-console-account-copy strong { display: block; overflow: hidden; color: var(--scheme3-ink); font-size: .67rem; text-overflow: ellipsis; white-space: nowrap; }
.scheme3-console-account-copy span { display: block; margin-top: .12rem; color: var(--scheme3-muted); font-size: .58rem; }
.scheme3-console-action { display: flex; width: 100%; min-height: 2.2rem; align-items: center; gap: .6rem; border: 1px solid transparent; border-radius: 7px; padding: .45rem .65rem; background: transparent; color: var(--scheme3-muted); font-size: .65rem; font-weight: 700; text-align: left; transition: color 150ms ease,background-color 150ms ease,transform 150ms ease; }
.scheme3-console-action:hover { background: rgba(255,255,255,.72); color: var(--scheme3-ink); }
.scheme3-console-action:active { transform: scale(.98); }
.scheme3-console-logout { color: #9e4d3d; }
.scheme3-console-logout:hover { color: #7d372d; background: rgba(158,77,61,.1); }
.scheme3-console-workspace { min-width: 0; }
.scheme3-console-topbar { position: sticky; top: 0; z-index: 30; display: flex; min-height: 4.25rem; align-items: center; justify-content: space-between; gap: 1rem; border-bottom: 1px solid var(--scheme3-line); padding: .75rem 1.5rem; background: rgba(244,242,236,.92); backdrop-filter: blur(20px); }
.scheme3-console-topbar-left { display: flex; min-width: 0; align-items: center; gap: .75rem; }
.scheme3-console-topbar-kicker { display: block; color: var(--scheme3-muted); font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; font-size: .55rem; letter-spacing: .13em; }
.scheme3-console-topbar-left strong { display: block; margin-top: .18rem; color: var(--scheme3-ink); font-family: Georgia,'Times New Roman',serif; font-size: 1.15rem; font-weight: 400; }
.scheme3-console-topbar-right { display: flex; min-width: 0; align-items: center; gap: .65rem; }
.scheme3-console-balance,.scheme3-console-status { display: inline-flex; align-items: center; gap: .28rem; color: var(--scheme3-muted); font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; font-size: .61rem; white-space: nowrap; }
.scheme3-console-balance { color: #b7791f; }
.scheme3-console-status i { width: .36rem; height: .36rem; border-radius: 999px; background: #1e5c42; box-shadow: 0 0 0 .2rem rgba(30,92,66,.13); }
.scheme3-console-user { display: inline-flex; min-width: 0; align-items: center; gap: .45rem; color: var(--scheme3-ink); font-size: .67rem; font-weight: 700; }
.scheme3-console-menu-button { display: none; align-items: center; justify-content: center; border: 1px solid var(--scheme3-line); border-radius: 7px; padding: .45rem; background: var(--scheme3-card); color: var(--scheme3-ink); }
.scheme3-console-overlay { display: none; }
.scheme3-console-content { min-width: 0; padding: 1.15rem 1.35rem 1.6rem; }
.scheme3-console-page-frame { min-height: calc(100vh - 7rem); overflow-x: hidden; }
.scheme3-console-page-frame :deep(.card) { border-color: var(--scheme3-line); border-radius: 8px; background: var(--scheme3-card); box-shadow: 0 10px 24px rgba(54,48,34,.06); }
.scheme3-console-page-frame :deep(.btn-primary) { background: #1e5c42; box-shadow: 0 8px 18px rgba(30,92,66,.16); }
.scheme3-console-page-frame :deep(.btn-primary:hover) { background: #174a35; }
.scheme3-console-page-frame :deep(.btn-secondary) { border-color: var(--scheme3-line); background: var(--scheme3-card); color: var(--scheme3-ink); }
.scheme3-console-page-frame :deep(.input),.scheme3-console-page-frame :deep(select),.scheme3-console-page-frame :deep(textarea) { border-color: var(--scheme3-line); background: var(--scheme3-card); color: var(--scheme3-ink); }
.scheme3-console-page-frame :deep(.input:focus),.scheme3-console-page-frame :deep(select:focus),.scheme3-console-page-frame :deep(textarea:focus) { border-color: #1e5c42; box-shadow: 0 0 0 3px rgba(30,92,66,.12); }
.scheme3-console-page-frame :deep(.bg-white) { background-color: var(--scheme3-card); }
.scheme3-console-page-frame :deep(.border-gray-100),.scheme3-console-page-frame :deep(.border-gray-200) { border-color: var(--scheme3-line); }
.scheme3-console-page-frame :deep(.text-gray-950),.scheme3-console-page-frame :deep(.text-gray-900),.scheme3-console-page-frame :deep(.text-gray-800),.scheme3-console-page-frame :deep(.text-gray-700) { color: var(--scheme3-ink) !important; }
.scheme3-console-page-frame :deep(.text-gray-600),.scheme3-console-page-frame :deep(.text-gray-500),.scheme3-console-page-frame :deep(.text-gray-400) { color: var(--scheme3-muted) !important; }
.scheme3-console-page-frame :deep(.text-primary-900),.scheme3-console-page-frame :deep(.text-primary-800),.scheme3-console-page-frame :deep(.text-primary-700),.scheme3-console-page-frame :deep(.text-primary-600),.scheme3-console-page-frame :deep(.text-primary-500) { color: #1e5c42 !important; }
.scheme3-console-page-frame :deep(.bg-gray-50),.scheme3-console-page-frame :deep(.bg-gray-100),.scheme3-console-page-frame :deep(.bg-primary-50),.scheme3-console-page-frame :deep(.bg-primary-100) { background-color: var(--scheme3-subtle) !important; }
.scheme3-console-page-frame :deep(.bg-primary-500),.scheme3-console-page-frame :deep(.bg-primary-600),.scheme3-console-page-frame :deep(.bg-primary-700) { background-color: #1e5c42 !important; }
.scheme3-console-page-frame :deep(.border-primary-100),.scheme3-console-page-frame :deep(.border-primary-200),.scheme3-console-page-frame :deep(.border-primary-300) { border-color: rgba(30,92,66,.25) !important; }
.scheme3-console-page-frame :deep(.ring-primary-200),.scheme3-console-page-frame :deep(.ring-primary-500) { --tw-ring-color: rgba(30,92,66,.22) !important; }
.scheme3-console-page-frame :deep(table thead),.scheme3-console-page-frame :deep(thead) { background: var(--scheme3-subtle); }
.scheme3-console-page-frame :deep(th),.scheme3-console-page-frame :deep(td) { border-color: var(--scheme3-line); color: var(--scheme3-ink); }

:global(.dark .scheme3-console-layout) { --scheme3-ink: #f4f2ec; --scheme3-muted: #aaa69a; --scheme3-line: #47443a; --scheme3-paper: #1b1b18; --scheme3-card: #24231f; --scheme3-subtle: #2b2924; background: var(--scheme3-paper); }
:global(.dark .scheme3-console-sidebar) { background: rgba(36,35,31,.95); }
:global(.dark .scheme3-console-topbar) { background: rgba(27,27,24,.93); }
:global(.dark .scheme3-console-link:hover),:global(.dark .scheme3-console-action:hover) { background: rgba(143,194,165,.08); }
:global(.dark .scheme3-console-brand-mark) { border-color: rgba(143,194,165,.38); background: #8fc2a5; color: #1b1b18; }
:global(.dark .scheme3-console-avatar) { background: #f4f2ec; color: #1b1b18; }
:global(.dark .scheme3-console-link-active) { border-color: rgba(143,194,165,.3); background: rgba(143,194,165,.1); color: #8fc2a5; box-shadow: inset 3px 0 0 #8fc2a5; }
:global(.dark .scheme3-console-current) { background: rgba(143,194,165,.12); color: #8fc2a5; }
:global(.dark .scheme3-console-page-frame :deep(.card)),:global(.dark .scheme3-console-page-frame :deep(.bg-white)) { background-color: var(--scheme3-card); }
:global(.dark .scheme3-console-page-frame :deep(.bg-dark-800)),:global(.dark .scheme3-console-page-frame :deep(.bg-dark-900)),:global(.dark .scheme3-console-page-frame :deep(.bg-dark-950)) { background-color: var(--scheme3-card) !important; }
:global(.dark .scheme3-console-page-frame :deep(.text-dark-100)),:global(.dark .scheme3-console-page-frame :deep(.text-dark-200)),:global(.dark .scheme3-console-page-frame :deep(.text-dark-300)),:global(.dark .scheme3-console-page-frame :deep(.text-dark-400)) { color: var(--scheme3-muted) !important; }
:global(.dark .scheme3-console-page-frame :deep(.text-primary-900)),:global(.dark .scheme3-console-page-frame :deep(.text-primary-800)),:global(.dark .scheme3-console-page-frame :deep(.text-primary-700)),:global(.dark .scheme3-console-page-frame :deep(.text-primary-600)),:global(.dark .scheme3-console-page-frame :deep(.text-primary-500)) { color: #8fc2a5 !important; }

/* Teleported user controls do not inherit the page-frame variables. Scope
   these overrides to the user console body marker so admin dialogs retain
   their upstream styling. */
:global(body.scheme3-user-context .select-dropdown-portal) { border-color: #dad5c8 !important; background: #fbfaf6 !important; color: #16150f !important; border-radius: 7px; box-shadow: 0 16px 34px rgba(54,48,34,.16) !important; }
:global(body.scheme3-user-context .select-dropdown-portal .select-search) { border-bottom-color: #dad5c8 !important; }
:global(body.scheme3-user-context .select-dropdown-portal .select-search-input) { color: #16150f !important; }
:global(body.scheme3-user-context .select-dropdown-portal .select-option) { color: #6b695f !important; }
:global(body.scheme3-user-context .select-dropdown-portal .select-option:hover),:global(body.scheme3-user-context .select-dropdown-portal .select-option-focused) { background: #f1eee6 !important; color: #16150f !important; }
:global(body.scheme3-user-context .select-dropdown-portal .select-option-selected) { background: rgba(30,92,66,.1) !important; color: #1e5c42 !important; }
:global(body.scheme3-user-context .select-dropdown-portal .select-option-group) { background: #f1eee6 !important; color: #6b695f !important; }
:global(body.scheme3-user-context .select-dropdown-portal .select-empty) { color: #6b695f !important; }
:global(html.dark body.scheme3-user-context .select-dropdown-portal) { border-color: #47443a !important; background: #24231f !important; color: #f4f2ec !important; box-shadow: 0 18px 38px rgba(0,0,0,.28) !important; }
:global(html.dark body.scheme3-user-context .select-dropdown-portal .select-search) { border-bottom-color: #47443a !important; }
:global(html.dark body.scheme3-user-context .select-dropdown-portal .select-search-input) { color: #f4f2ec !important; }
:global(html.dark body.scheme3-user-context .select-dropdown-portal .select-option) { color: #aaa69a !important; }
:global(html.dark body.scheme3-user-context .select-dropdown-portal .select-option:hover),:global(html.dark body.scheme3-user-context .select-dropdown-portal .select-option-focused) { background: #2b2924 !important; color: #f4f2ec !important; }
:global(html.dark body.scheme3-user-context .select-dropdown-portal .select-option-selected) { background: rgba(143,194,165,.12) !important; color: #8fc2a5 !important; }
:global(html.dark body.scheme3-user-context .select-dropdown-portal .select-option-group) { background: #2b2924 !important; color: #aaa69a !important; }
:global(html.dark body.scheme3-user-context .select-dropdown-portal .select-empty) { color: #aaa69a !important; }

:global(body.scheme3-user-context .modal-overlay) { background: rgba(21,20,16,.48) !important; backdrop-filter: blur(5px); }
:global(body.scheme3-user-context .modal-content) { border: 1px solid #dad5c8; border-radius: 8px; background: #fbfaf6 !important; color: #16150f; box-shadow: 0 24px 60px rgba(54,48,34,.18) !important; }
:global(body.scheme3-user-context .modal-header) { border-bottom-color: #dad5c8 !important; }
:global(body.scheme3-user-context .modal-footer) { border-top-color: #dad5c8 !important; background: #f1eee6 !important; }
:global(body.scheme3-user-context .modal-title) { color: #16150f !important; }
:global(body.scheme3-user-context .modal-content .btn-primary) { background: #1e5c42 !important; }
:global(body.scheme3-user-context .modal-content .btn-secondary) { border-color: #dad5c8 !important; background: #fbfaf6 !important; color: #16150f !important; }
:global(body.scheme3-user-context .modal-content .input) { border-color: #dad5c8 !important; background: #fbfaf6 !important; color: #16150f !important; }
:global(html.dark body.scheme3-user-context .modal-overlay) { background: rgba(0,0,0,.62) !important; }
:global(html.dark body.scheme3-user-context .modal-content) { border-color: #47443a; background: #24231f !important; color: #f4f2ec; box-shadow: 0 24px 60px rgba(0,0,0,.32) !important; }
:global(html.dark body.scheme3-user-context .modal-header) { border-bottom-color: #47443a !important; }
:global(html.dark body.scheme3-user-context .modal-footer) { border-top-color: #47443a !important; background: #2b2924 !important; }
:global(html.dark body.scheme3-user-context .modal-title) { color: #f4f2ec !important; }
:global(html.dark body.scheme3-user-context .modal-content .btn-secondary) { border-color: #47443a !important; background: #24231f !important; color: #f4f2ec !important; }
:global(html.dark body.scheme3-user-context .modal-content .input) { border-color: #47443a !important; background: #24231f !important; color: #f4f2ec !important; }

:global(body.scheme3-user-context .scheme3-toast) { border: 1px solid #dad5c8 !important; border-left-width: 3px !important; border-radius: 7px !important; background: #fbfaf6 !important; color: #16150f; box-shadow: 0 16px 34px rgba(54,48,34,.16) !important; }
:global(body.scheme3-user-context .scheme3-toast .text-gray-900),:global(body.scheme3-user-context .scheme3-toast .text-gray-600) { color: #16150f !important; }
:global(body.scheme3-user-context .scheme3-toast .text-gray-400) { color: #6b695f !important; }
:global(body.scheme3-user-context .scheme3-toast .bg-gray-100) { background: #f1eee6 !important; }
:global(body.scheme3-user-context .scheme3-toast .hover\:bg-gray-100:hover) { background: #f1eee6 !important; }
:global(body.scheme3-user-context .scheme3-toast.border-green-500) { border-left-color: #1e5c42 !important; }
:global(body.scheme3-user-context .scheme3-toast.border-red-500) { border-left-color: #9e4d3d !important; }
:global(body.scheme3-user-context .scheme3-toast.border-yellow-500) { border-left-color: #b7791f !important; }
:global(body.scheme3-user-context .scheme3-toast.border-blue-500) { border-left-color: #1e5c42 !important; }
:global(body.scheme3-user-context .scheme3-toast .text-blue-500) { color: #1e5c42 !important; }
:global(body.scheme3-user-context .scheme3-toast .bg-green-500),:global(body.scheme3-user-context .scheme3-toast .bg-blue-500) { background: #1e5c42 !important; }
:global(body.scheme3-user-context .scheme3-toast .bg-yellow-500) { background: #b7791f !important; }

:global(html.dark body.scheme3-user-context .scheme3-toast) { border-color: #47443a !important; background: #24231f !important; color: #f4f2ec; box-shadow: 0 18px 38px rgba(0,0,0,.3) !important; }
:global(html.dark body.scheme3-user-context .scheme3-toast .text-gray-900),:global(html.dark body.scheme3-user-context .scheme3-toast .text-gray-600),:global(html.dark body.scheme3-user-context .scheme3-toast .text-white) { color: #f4f2ec !important; }
:global(html.dark body.scheme3-user-context .scheme3-toast .text-gray-400),:global(html.dark body.scheme3-user-context .scheme3-toast .text-gray-500) { color: #aaa69a !important; }
:global(html.dark body.scheme3-user-context .scheme3-toast .bg-gray-100),:global(html.dark body.scheme3-user-context .scheme3-toast .bg-dark-700) { background: #2b2924 !important; }
:global(html.dark body.scheme3-user-context .scheme3-toast .hover\:bg-dark-700:hover) { background: #2b2924 !important; }
:global(html.dark body.scheme3-user-context .scheme3-toast.border-green-500),:global(html.dark body.scheme3-user-context .scheme3-toast.border-blue-500) { border-left-color: #8fc2a5 !important; }
:global(html.dark body.scheme3-user-context .scheme3-toast.border-red-500) { border-left-color: #c76b5d !important; }
:global(html.dark body.scheme3-user-context .scheme3-toast.border-yellow-500) { border-left-color: #d3a45c !important; }
:global(html.dark body.scheme3-user-context .scheme3-toast .text-blue-500),:global(html.dark body.scheme3-user-context .scheme3-toast .text-green-500) { color: #8fc2a5 !important; }
:global(html.dark body.scheme3-user-context .scheme3-toast .bg-green-500),:global(html.dark body.scheme3-user-context .scheme3-toast .bg-blue-500) { background: #8fc2a5 !important; }
:global(html.dark body.scheme3-user-context .scheme3-toast .bg-yellow-500) { background: #d3a45c !important; }

/* Shared user components still carry a few semantic blue/purple utility
   classes. Re-map those utilities only inside the user console; admin pages
   keep the upstream palette because they never receive this body marker. */
:global(body.scheme3-user-context [class*="bg-blue-50"]),
:global(body.scheme3-user-context [class*="bg-indigo-50"]),
:global(body.scheme3-user-context [class*="bg-purple-50"]),
:global(body.scheme3-user-context [class*="bg-violet-50"]),
:global(body.scheme3-user-context [class*="bg-sky-50"]),
:global(body.scheme3-user-context [class*="bg-cyan-50"]) { background-color: #f1eee6 !important; }

:global(body.scheme3-user-context [class*="bg-blue-100"]),
:global(body.scheme3-user-context [class*="bg-indigo-100"]),
:global(body.scheme3-user-context [class*="bg-purple-100"]),
:global(body.scheme3-user-context [class*="bg-violet-100"]),
:global(body.scheme3-user-context [class*="bg-sky-100"]),
:global(body.scheme3-user-context [class*="bg-cyan-100"]) { background-color: rgba(30,92,66,.1) !important; }

:global(body.scheme3-user-context [class*="bg-blue-500"]),
:global(body.scheme3-user-context [class*="bg-indigo-500"]),
:global(body.scheme3-user-context [class*="bg-purple-500"]),
:global(body.scheme3-user-context [class*="bg-violet-500"]) { background-color: #1e5c42 !important; }

:global(body.scheme3-user-context [class*="text-blue-"]),
:global(body.scheme3-user-context [class*="text-indigo-"]),
:global(body.scheme3-user-context [class*="text-purple-"]),
:global(body.scheme3-user-context [class*="text-violet-"]),
:global(body.scheme3-user-context [class*="text-sky-"]),
:global(body.scheme3-user-context [class*="text-cyan-"]) { color: #1e5c42 !important; }

:global(body.scheme3-user-context [class*="border-blue-"]),
:global(body.scheme3-user-context [class*="border-indigo-"]),
:global(body.scheme3-user-context [class*="border-purple-"]),
:global(body.scheme3-user-context [class*="border-violet-"]),
:global(body.scheme3-user-context [class*="border-sky-"]),
:global(body.scheme3-user-context [class*="border-cyan-"]) { border-color: rgba(30,92,66,.25) !important; }

:global(body.scheme3-user-context .select-trigger),
:global(body.scheme3-user-context .date-picker-trigger) { border-color: #dad5c8 !important; border-radius: 7px; background: #fbfaf6 !important; color: #16150f !important; }
:global(body.scheme3-user-context .select-trigger:hover),
:global(body.scheme3-user-context .select-trigger-open),
:global(body.scheme3-user-context .date-picker-trigger:hover),
:global(body.scheme3-user-context .date-picker-trigger-open) { border-color: #1e5c42 !important; box-shadow: 0 0 0 3px rgba(30,92,66,.11); }

:global(html.dark body.scheme3-user-context [class*="dark:bg-blue-"]),
:global(html.dark body.scheme3-user-context [class*="dark:bg-indigo-"]),
:global(html.dark body.scheme3-user-context [class*="dark:bg-purple-"]),
:global(html.dark body.scheme3-user-context [class*="dark:bg-violet-"]),
:global(html.dark body.scheme3-user-context [class*="dark:bg-sky-"]),
:global(html.dark body.scheme3-user-context [class*="dark:bg-cyan-"]) { background-color: rgba(143,194,165,.1) !important; }

:global(html.dark body.scheme3-user-context [class*="dark:text-blue-"]),
:global(html.dark body.scheme3-user-context [class*="dark:text-indigo-"]),
:global(html.dark body.scheme3-user-context [class*="dark:text-purple-"]),
:global(html.dark body.scheme3-user-context [class*="dark:text-violet-"]),
:global(html.dark body.scheme3-user-context [class*="dark:text-sky-"]),
:global(html.dark body.scheme3-user-context [class*="dark:text-cyan-"]) { color: #8fc2a5 !important; }

:global(html.dark body.scheme3-user-context [class*="dark:border-blue-"]),
:global(html.dark body.scheme3-user-context [class*="dark:border-indigo-"]),
:global(html.dark body.scheme3-user-context [class*="dark:border-purple-"]),
:global(html.dark body.scheme3-user-context [class*="dark:border-violet-"]),
:global(html.dark body.scheme3-user-context [class*="dark:border-sky-"]),
:global(html.dark body.scheme3-user-context [class*="dark:border-cyan-"]) { border-color: rgba(143,194,165,.25) !important; }

:global(html.dark body.scheme3-user-context .select-trigger),
:global(html.dark body.scheme3-user-context .date-picker-trigger) { border-color: #47443a !important; background: #24231f !important; color: #f4f2ec !important; }
:global(html.dark body.scheme3-user-context .select-trigger:hover),
:global(html.dark body.scheme3-user-context .select-trigger-open),
:global(html.dark body.scheme3-user-context .date-picker-trigger:hover),
:global(html.dark body.scheme3-user-context .date-picker-trigger-open) { border-color: #8fc2a5 !important; box-shadow: 0 0 0 3px rgba(143,194,165,.13); }

.scheme3-console-layout-collapsed .scheme3-console-brand-copy,.scheme3-console-layout-collapsed .scheme3-console-caption,.scheme3-console-layout-collapsed .scheme3-console-section-label,.scheme3-console-layout-collapsed .scheme3-console-link-text,.scheme3-console-layout-collapsed .scheme3-console-current,.scheme3-console-layout-collapsed .scheme3-console-account-copy,.scheme3-console-layout-collapsed .scheme3-console-action span { display: none; }
.scheme3-console-layout-collapsed .scheme3-console-brand,.scheme3-console-layout-collapsed .scheme3-console-account { justify-content: center; }
.scheme3-console-layout-collapsed .scheme3-console-link,.scheme3-console-layout-collapsed .scheme3-console-action { justify-content: center; padding-right: .4rem; padding-left: .4rem; }

@media (max-width: 1023px) {
  .scheme3-console-layout,.scheme3-console-layout-collapsed { display: block; }
  .scheme3-console-sidebar { position: fixed; left: 0; top: 0; width: 15.5rem; transform: translateX(-102%); transition: transform 220ms ease; }
  .scheme3-console-sidebar.scheme3-console-sidebar-open { transform: translateX(0); box-shadow: 18px 0 40px rgba(14,30,58,.18); }
  .scheme3-console-close { display: inline-flex; align-items: center; justify-content: center; }
  .scheme3-console-overlay { position: fixed; z-index: 40; inset: 0; display: block; background: rgba(9,18,32,.38); backdrop-filter: blur(2px); }
  .scheme3-console-menu-button { display: inline-flex; }
  .scheme3-console-topbar { padding-right: 1rem; padding-left: 1rem; }
}
@media (max-width: 640px) {
  .scheme3-console-content { padding: .7rem .65rem 1rem; }
  .scheme3-console-topbar { min-height: 3.85rem; padding: .65rem; }
  .scheme3-console-topbar-kicker { font-size: .49rem; }
  .scheme3-console-topbar-left strong { font-size: 1rem; }
  .scheme3-console-topbar-right { gap: .45rem; }
  .scheme3-console-balance,.scheme3-console-topbar-right .scheme3-console-locale-tool { display: none; }
  .scheme3-console-user > span:last-child { display: none; }
  :global(body.scheme3-user-context .scheme3-toast) { min-width: 0 !important; width: calc(100vw - 2rem); }
}
</style>

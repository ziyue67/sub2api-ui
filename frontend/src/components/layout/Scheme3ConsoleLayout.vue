<template>
  <div
    class="scheme3-console-layout"
    :class="{
      'scheme3-console-layout-collapsed': navCollapsed,
      'scheme3-console-layout-admin': adminMode,
    }"
  >
    <aside class="scheme3-console-sidebar" :class="{ 'scheme3-console-sidebar-open': mobileNavOpen }" :aria-label="adminMode ? '运维导航' : '空间导航'">
      <div class="scheme3-console-brand">
        <router-link :to="homePath" class="scheme3-console-brand-mark" :aria-label="adminMode ? '返回运维总览' : '返回控制台'" @click="handleNavItemClick(homePath)">
          <img v-if="siteLogo" :src="siteLogo" alt="" />
          <span v-else>ST</span>
        </router-link>
        <div class="scheme3-console-brand-copy">
          <strong>{{ siteName }}</strong>
          <span>{{ adminMode ? '运维控制空间' : '个人工作空间' }}</span>
        </div>
        <VersionBadge :version="appStore.siteVersion" class="scheme3-console-version-control" />
        <button type="button" class="scheme3-console-close" aria-label="关闭导航" @click="closeMobileNav">
          <Icon name="x" size="sm" />
        </button>
      </div>

      <div class="scheme3-console-caption">{{ adminMode ? '运维导航' : '空间导航' }}</div>
      <nav ref="sidebarNavRef" class="scheme3-console-links">
        <section v-for="section in consoleNavSections" :key="section.id" class="scheme3-console-section">
          <div v-if="section.showLabel" class="scheme3-console-section-label">{{ section.label }}</div>
          <template v-for="item in section.items" :key="item.path">
            <button
              v-if="item.children?.length"
              type="button"
              class="scheme3-console-link scheme3-console-group"
              :class="{
                'scheme3-console-link-active': isGroupActive(item) && !isGroupExpanded(item),
                'scheme3-console-link-collapsed': navCollapsed,
              }"
              :title="navCollapsed ? item.label : undefined"
              @click="toggleNavGroup(item)"
            >
              <span v-if="item.iconSvg" class="scheme3-console-link-icon scheme3-console-svg-icon" v-html="sanitizeSvg(item.iconSvg)"></span>
              <span v-else class="scheme3-console-link-icon"><Icon :name="item.icon" size="sm" /></span>
              <span class="scheme3-console-link-text" :aria-hidden="navCollapsed ? 'true' : 'false'">{{ item.label }}</span>
              <Icon v-if="!navCollapsed" name="chevronDown" size="xs" class="scheme3-console-group-chevron" :class="{ 'is-expanded': isGroupExpanded(item) }" />
            </button>
            <div v-if="item.children?.length && !navCollapsed && isGroupExpanded(item)" class="scheme3-console-subnav">
              <router-link
                v-for="child in item.children"
                :key="child.path"
                :to="{ path: child.path, query: child.query }"
                class="scheme3-console-link scheme3-console-subnav-link"
                :class="{ 'scheme3-console-link-active': isNavActive(child.path, true) }"
                @click="handleNavItemClick(child.path)"
              >
                <span v-if="child.iconSvg" class="scheme3-console-link-icon scheme3-console-svg-icon" v-html="sanitizeSvg(child.iconSvg)"></span>
                <span v-else class="scheme3-console-link-icon"><Icon :name="child.icon" size="sm" /></span>
                <span class="scheme3-console-link-text">{{ child.label }}</span>
                <span v-if="isNavActive(child.path, true)" class="scheme3-console-current">当前</span>
              </router-link>
            </div>
            <router-link
              v-else-if="!item.children?.length"
              :to="{ path: item.path, query: item.query }"
              class="scheme3-console-link"
              :class="{ 'scheme3-console-link-active': isNavActive(item.path), 'scheme3-console-link-collapsed': navCollapsed }"
              :title="navCollapsed ? item.label : undefined"
              :id="navTourId(item.path)"
              :data-tour="item.path === '/keys' ? 'sidebar-my-keys' : undefined"
              @click="handleNavItemClick(item.path)"
            >
              <span v-if="item.iconSvg" class="scheme3-console-link-icon scheme3-console-svg-icon" v-html="sanitizeSvg(item.iconSvg)"></span>
              <span v-else class="scheme3-console-link-icon"><Icon :name="item.icon" size="sm" /></span>
              <span class="scheme3-console-link-text" :aria-hidden="navCollapsed ? 'true' : 'false'">{{ item.label }}</span>
              <span v-if="isNavActive(item.path)" class="scheme3-console-current">当前</span>
            </router-link>
          </template>
        </section>
      </nav>

      <div class="scheme3-console-foot">
        <router-link to="/profile" class="scheme3-console-account" aria-label="打开个人资料" :title="userEmail">
          <span class="scheme3-console-avatar">
            <img v-if="avatarUrl" :src="avatarUrl" :alt="userLabel" />
            <span v-else>{{ userInitials }}</span>
          </span>
          <div class="scheme3-console-account-copy">
            <strong>{{ userLabel }}</strong>
            <span>{{ accountRoleLabel }}<template v-if="userEmail"> · {{ userEmail }}</template></span>
          </div>
        </router-link>
        <button type="button" class="scheme3-console-action" @click="toggleTheme">
          <Icon :name="isDarkMode ? 'sun' : 'moon'" size="sm" />
          <span>{{ isDarkMode ? '切换浅色' : '切换深色' }}</span>
        </button>
        <button type="button" class="scheme3-console-action" @click="toggleNavCollapse">
          <Icon :name="navCollapsed ? 'chevronRight' : 'chevronLeft'" size="sm" />
          <span>{{ navCollapsed ? '展开导航' : '收起导航' }}</span>
        </button>
        <button
          v-if="showOnboardingButton"
          type="button"
          class="scheme3-console-action"
          @click="handleReplayGuide"
        >
          <Icon name="questionCircle" size="sm" />
          <span>{{ t('onboarding.restartTour') }}</span>
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
            <span class="scheme3-console-topbar-kicker">{{ adminMode ? 'SHOUR OR TOKEN / 运维空间' : 'SHOUR OR TOKEN / 运营空间' }}</span>
            <strong>{{ pageTitle }}</strong>
            <small v-if="pageDescription">{{ pageDescription }}</small>
          </div>
        </div>
        <div class="scheme3-console-topbar-right">
          <AnnouncementBell v-if="authStore.user" class="scheme3-console-tool" />
          <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="scheme3-console-doc-link">
            <Icon name="book" size="sm" />
            <span>{{ t('nav.docs') }}</span>
          </a>
          <router-link v-if="authStore.user && modelPlazaEnabled" :to="{ path: '/model-plaza', query: { embedded: '1' } }" class="scheme3-console-doc-link">
            <Icon name="grid" size="sm" />
            <span>{{ t('nav.modelPlaza') }}</span>
          </router-link>
          <LocaleSwitcher class="scheme3-console-tool scheme3-console-locale-tool" />
          <SubscriptionProgressMini v-if="authStore.user" class="scheme3-console-subscription" />
          <span class="scheme3-console-balance"><Icon name="dollar" size="xs" />{{ t('common.availableBalance') }} {{ formatMoney(Number(user?.balance || 0)) }}</span>
          <span class="scheme3-console-status"><i></i>{{ t('common.online') }}</span>
          <div ref="accountMenuRef" class="scheme3-console-account-menu">
            <button
              :id="accountMenuButtonId"
              ref="accountMenuButtonRef"
              type="button"
              class="scheme3-console-user"
              aria-haspopup="menu"
              :aria-controls="accountMenuId"
              :aria-expanded="accountMenuOpen"
              :aria-label="t('common.userMenu')"
              :title="userEmail"
              @click.stop="toggleAccountMenu"
            >
              <span class="scheme3-console-avatar">
                <img v-if="avatarUrl" :src="avatarUrl" :alt="userLabel" />
                <span v-else>{{ userInitials }}</span>
              </span>
              <span class="scheme3-console-user-copy"><strong>{{ userLabel }}</strong><small>{{ accountRoleLabel }}</small></span>
              <Icon name="chevronDown" size="xs" class="scheme3-console-user-chevron" :class="{ 'is-open': accountMenuOpen }" />
            </button>
            <div
              v-if="accountMenuOpen"
              :id="accountMenuId"
              class="scheme3-console-account-popover"
              role="menu"
              tabindex="-1"
              @keydown="handleAccountMenuKeydown"
            >
              <div class="scheme3-console-account-summary" role="presentation">
                <strong>{{ userLabel }}</strong>
                <span>{{ userEmail }}</span>
                <small>{{ accountRoleLabel }}</small>
              </div>
              <div class="scheme3-console-account-balance" role="presentation">
                <div><span>{{ t('common.availableBalance') }}</span><strong>{{ formatMoney(availableBalance) }}</strong></div>
                <div v-if="frozenBalance > 0"><span>{{ t('common.frozenBalance') }}</span><strong>{{ formatMoney(frozenBalance) }}</strong></div>
                <div class="scheme3-console-account-total"><span>{{ t('common.totalBalance') }}</span><strong>{{ formatMoney(totalBalance) }}</strong></div>
              </div>
              <div class="scheme3-console-account-links" role="presentation">
                <router-link to="/profile" role="menuitem" @click="closeAccountMenu"><Icon name="user" size="sm" />{{ t('nav.profile') }}</router-link>
                <router-link to="/keys" role="menuitem" @click="closeAccountMenu"><Icon name="key" size="sm" />{{ t('nav.apiKeys') }}</router-link>
                <a v-if="authStore.isAdmin" href="https://github.com/ShourGG/sub2api" target="_blank" rel="noopener noreferrer" role="menuitem" @click="closeAccountMenu()"><Icon name="externalLink" size="sm" />{{ t('nav.github') }}</a>
                <button v-if="showOnboardingButton" type="button" role="menuitem" @click="handleReplayGuide"><Icon name="questionCircle" size="sm" />{{ t('onboarding.restartTour') }}</button>
              </div>
              <div v-if="contactInfo" class="scheme3-console-account-contact" role="presentation"><Icon name="chatBubble" size="sm" /><span>{{ t('common.contactSupport') }}: {{ contactInfo }}</span></div>
              <button type="button" class="scheme3-console-account-logout" role="menuitem" @click="logout"><Icon name="login" size="sm" />{{ t('nav.logout') }}</button>
            </div>
          </div>
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
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import AnnouncementBell from '@/components/common/AnnouncementBell.vue'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import SubscriptionProgressMini from '@/components/common/SubscriptionProgressMini.vue'
import VersionBadge from '@/components/common/VersionBadge.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAdminSettingsStore, useAppStore, useAuthStore, useOnboardingStore } from '@/stores'
import { useBatchImageAccess } from '@/composables/useBatchImageAccess'
import type { CustomMenuItem } from '@/types'
import { resolveDisplaySiteName } from '@/utils/branding'
import { FeatureFlags, makeSidebarFlag } from '@/utils/featureFlags'
import { sanitizeSvg } from '@/utils/sanitize'
import { sanitizeUrl } from '@/utils/url'

type ConsoleNavIcon =
  | 'home'
  | 'key'
  | 'grid'
  | 'image'
  | 'chart'
  | 'chartBar'
  | 'server'
  | 'creditCard'
  | 'document'
  | 'gift'
  | 'users'
  | 'user'
  | 'folder'
  | 'globe'
  | 'bell'
  | 'shield'
  | 'cog'
  | 'book'

interface ConsoleNavItem {
  path: string
  query?: Record<string, string>
  label: string
  icon: ConsoleNavIcon
  iconSvg?: string
  hideInSimpleMode?: boolean
  featureFlag?: () => boolean | undefined
  children?: ConsoleNavItem[]
  expandOnly?: boolean
}

interface ConsoleNavSection {
  id: string
  label: string
  showLabel?: boolean
  items: ConsoleNavItem[]
}

const props = withDefaults(defineProps<{ adminMode?: boolean }>(), { adminMode: false })
const adminMode = computed(() => props.adminMode)
const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const authStore = useAuthStore()
const adminSettingsStore = useAdminSettingsStore()
const onboardingStore = useOnboardingStore()
const { canUseBatchImage, refreshBatchImageAccess } = useBatchImageAccess()
const mobileNavOpen = ref(false)
const navCollapsed = ref(localStorage.getItem('scheme3-nav-collapsed') === '1')
const isDarkMode = ref(document.documentElement.classList.contains('dark'))
const expandedNavGroups = ref<Set<string>>(new Set())
const sidebarNavRef = ref<HTMLElement | null>(null)
const accountMenuRef = ref<HTMLElement | null>(null)
const accountMenuButtonRef = ref<HTMLButtonElement | null>(null)
const accountMenuOpen = ref(false)
const accountMenuId = 'scheme3-account-menu'
const accountMenuButtonId = 'scheme3-account-menu-button'

const user = computed(() => authStore.user)
const homePath = computed(() => (adminMode.value ? '/admin/dashboard' : '/dashboard'))
const userLabel = computed(() => user.value?.username || user.value?.email?.split('@')[0] || t('profile.user'))
const userEmail = computed(() => user.value?.email || '')
const accountRoleLabel = computed(() => t(authStore.isAdmin ? 'common.administratorIdentity' : 'common.userIdentity'))
const userInitials = computed(() => userLabel.value.trim().slice(0, 2).toUpperCase() || 'ST')
const showOnboardingButton = computed(() => adminMode.value && !authStore.isSimpleMode && authStore.isAdmin)
const siteName = computed(() => resolveDisplaySiteName(appStore.siteName))
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const docUrl = computed(() => sanitizeUrl(appStore.docUrl || ''))
const avatarUrl = computed(() => sanitizeUrl(user.value?.avatar_url?.trim() || '', { allowRelative: true, allowDataUrl: true }))
const contactInfo = computed(() => appStore.contactInfo || '')
const availableBalance = computed(() => Number(user.value?.balance || 0))
const frozenBalance = computed(() => Number(user.value?.frozen_balance || 0))
const totalBalance = computed(() => availableBalance.value + frozenBalance.value)

const pageTitle = computed(() => {
  if (route.name === 'CustomPage' || route.path.startsWith('/custom/')) {
    const id = String(route.params.id || route.path.split('/').pop() || '')
    const publicItem = appStore.cachedPublicSettings?.custom_menu_items?.find((item: CustomMenuItem) => item.id === id)
    const adminItem = authStore.isAdmin
      ? adminSettingsStore.customMenuItems.find((item: CustomMenuItem) => item.id === id)
      : undefined
    if (publicItem?.label || adminItem?.label) return publicItem?.label || adminItem?.label || '扩展页面'
    return '扩展页面'
  }
  const titleKey = route.meta.titleKey as string | undefined
  if (titleKey) return t(titleKey)
  return (route.meta.title as string | undefined) || '控制台'
})

const pageDescription = computed(() => {
  const descriptionKey = route.meta.descriptionKey as string | undefined
  if (descriptionKey) return t(descriptionKey)
  return (route.meta.description as string | undefined) || ''
})

const flagChannelMonitor = makeSidebarFlag(FeatureFlags.channelMonitor)
const flagAvailableChannels = makeSidebarFlag(FeatureFlags.availableChannels)
const flagModelPlaza = makeSidebarFlag(FeatureFlags.modelPlaza)
const flagPayment = makeSidebarFlag(FeatureFlags.payment)
const flagAffiliate = makeSidebarFlag(FeatureFlags.affiliate)
const flagRiskControl = makeSidebarFlag(FeatureFlags.riskControl)
const flagOpsMonitoring = () => adminSettingsStore.opsMonitoringEnabled
const flagAdminPayment = () => adminSettingsStore.paymentEnabled
const flagBatchImageAccess = () => canUseBatchImage.value
const modelPlazaEnabled = computed(() => flagModelPlaza())

function applyNavVisibility(items: ConsoleNavItem[]): ConsoleNavItem[] {
  const visible: ConsoleNavItem[] = []
  for (const item of items) {
    if (authStore.isSimpleMode && item.hideInSimpleMode) continue
    if (item.featureFlag && item.featureFlag() === false) continue
    if (item.children) {
      const children = applyNavVisibility(item.children)
      if (children.length > 0) visible.push({ ...item, children })
      continue
    }
    visible.push(item)
  }
  return visible
}

function buildSelfNavItems(withDashboard: boolean): ConsoleNavItem[] {
  const items: ConsoleNavItem[] = []
  if (withDashboard) items.push({ path: '/dashboard', label: t('nav.dashboard'), icon: 'home' })
  items.push(
    { path: '/keys', label: t('nav.apiKeys'), icon: 'key' },
    { path: '/model-square', label: '模型广场', icon: 'grid' },
    { path: '/canvas', label: '绘图工作站', icon: 'image' },
    { path: '/leaderboard', label: '总排行榜', icon: 'chart' },
    { path: '/batch-image', label: t('nav.batchImage'), icon: 'image', hideInSimpleMode: true, featureFlag: flagBatchImageAccess },
    { path: '/usage', label: t('nav.usage'), icon: 'chartBar', hideInSimpleMode: true },
    { path: '/available-channels', label: t('nav.availableChannels'), icon: 'server', hideInSimpleMode: true, featureFlag: flagAvailableChannels },
    { path: '/monitor', label: t('nav.channelStatus'), icon: 'server', featureFlag: flagChannelMonitor },
    { path: '/subscriptions', label: t('nav.mySubscriptions'), icon: 'creditCard', hideInSimpleMode: true },
    { path: '/purchase', label: t('nav.buySubscription'), icon: 'creditCard', hideInSimpleMode: true, featureFlag: flagPayment },
    { path: '/orders', label: t('nav.myOrders'), icon: 'document', hideInSimpleMode: true, featureFlag: flagPayment },
    { path: '/redeem', label: t('nav.redeem'), icon: 'gift', hideInSimpleMode: true },
    { path: '/affiliate', label: t('nav.affiliate'), icon: 'users', hideInSimpleMode: true, featureFlag: flagAffiliate },
    { path: '/profile', label: t('nav.profile'), icon: 'user' },
  )

  const customUserItems = (appStore.cachedPublicSettings?.custom_menu_items ?? [])
    .filter((item: CustomMenuItem) => item.visibility === 'user')
    .sort((a: CustomMenuItem, b: CustomMenuItem) => a.sort_order - b.sort_order)
    .map((item: CustomMenuItem): ConsoleNavItem => ({ path: `/custom/${item.id}`, label: item.label, icon: 'grid', iconSvg: item.icon_svg }))
  items.push(...customUserItems)
  return items
}

function buildAdminNavItems(): ConsoleNavItem[] {
  const items: ConsoleNavItem[] = [
    { path: '/admin/dashboard', label: t('nav.dashboard'), icon: 'home' },
    { path: '/model-square', label: '模型广场', icon: 'grid' },
    { path: '/admin/ops', label: t('nav.ops'), icon: 'chartBar', featureFlag: flagOpsMonitoring },
    { path: '/admin/users', label: t('nav.users'), icon: 'users', hideInSimpleMode: true },
    { path: '/admin/groups', label: t('nav.groups'), icon: 'folder', hideInSimpleMode: true },
    {
      path: '/admin/channels',
      label: t('nav.channelManagement'),
      icon: 'grid',
      hideInSimpleMode: true,
      expandOnly: true,
      children: [
        { path: '/admin/channels/pricing', label: t('nav.channelPricing'), icon: 'creditCard' },
        { path: '/admin/channels/monitor', label: t('nav.channelMonitor'), icon: 'server', featureFlag: flagChannelMonitor },
      ],
    },
    { path: '/admin/subscriptions', label: t('nav.subscriptions'), icon: 'creditCard', hideInSimpleMode: true },
    { path: '/admin/accounts', label: t('nav.accounts'), icon: 'globe' },
    { path: '/admin/announcements', label: t('nav.announcements'), icon: 'bell' },
    { path: '/admin/proxies', label: t('nav.proxies'), icon: 'server' },
    {
      path: '/admin/security-audit',
      label: t('nav.securityAudit'),
      icon: 'shield',
      expandOnly: true,
      featureFlag: flagRiskControl,
      children: [
        { path: '/admin/risk-control', label: t('nav.contentModeration'), icon: 'shield' },
        { path: '/admin/prompt-audit', label: t('nav.promptAudit'), icon: 'shield' },
      ],
    },
    { path: '/admin/redeem', label: t('nav.redeemCodes'), icon: 'gift', hideInSimpleMode: true },
    { path: '/admin/promo-codes', label: t('nav.promoCodes'), icon: 'gift', hideInSimpleMode: true },
    {
      path: '/admin/affiliates',
      label: t('nav.affiliateManagement'),
      icon: 'users',
      hideInSimpleMode: true,
      expandOnly: true,
      featureFlag: flagAffiliate,
      children: [
        { path: '/admin/affiliates/invites', label: t('nav.affiliateInviteRecords'), icon: 'users' },
        { path: '/admin/affiliates/rebates', label: t('nav.affiliateRebateRecords'), icon: 'chart' },
        { path: '/admin/affiliates/transfers', label: t('nav.affiliateTransferRecords'), icon: 'creditCard' },
      ],
    },
    {
      path: '/admin/orders',
      label: t('nav.orderManagement'),
      icon: 'document',
      hideInSimpleMode: true,
      expandOnly: true,
      featureFlag: flagAdminPayment,
      children: [
        { path: '/admin/orders/dashboard', label: t('nav.paymentDashboard'), icon: 'chartBar' },
        { path: '/admin/orders', label: t('nav.orderManagement'), icon: 'document' },
        { path: '/admin/orders/plans', label: t('nav.paymentPlans'), icon: 'creditCard' },
      ],
    },
    { path: '/admin/usage', label: t('nav.usage'), icon: 'chart' },
    { path: '/admin/leaderboard', label: t('nav.leaderboard'), icon: 'chart' },
    { path: '/admin/audit-logs', label: t('nav.auditLogs'), icon: 'shield', hideInSimpleMode: true },
  ]

  const customAdminItems = adminSettingsStore.customMenuItems
    .filter((item: CustomMenuItem) => item.visibility === 'admin')
    .sort((a: CustomMenuItem, b: CustomMenuItem) => a.sort_order - b.sort_order)
    .map((item: CustomMenuItem): ConsoleNavItem => ({ path: `/custom/${item.id}`, label: item.label, icon: 'grid', iconSvg: item.icon_svg }))

  const visible = applyNavVisibility(items)
  if (authStore.isSimpleMode) {
    // Keep the upstream simple-mode order: own keys, system settings, custom items.
    visible.push({ path: '/keys', label: t('nav.apiKeys'), icon: 'key' })
    visible.push({ path: '/admin/settings', label: t('nav.settings'), icon: 'cog' })
    visible.push(...customAdminItems)
    return visible
  }

  visible.push({ path: '/admin/settings', label: t('nav.settings'), icon: 'cog' })
  visible.push(...customAdminItems)
  return visible
}

const consoleNavSections = computed<ConsoleNavSection[]>(() => {
  if (adminMode.value) {
    const sections: ConsoleNavSection[] = [
      { id: 'admin-routing', label: '管理空间', showLabel: false, items: buildAdminNavItems() },
    ]
    if (!authStore.isSimpleMode) {
      sections.push({ id: 'personal', label: t('nav.myAccount'), showLabel: true, items: applyNavVisibility(buildSelfNavItems(false)) })
    }
    return sections.filter((section) => section.items.length > 0)
  }

  if (appStore.backendModeEnabled) return []

  // The upstream sidebar has one flat user section. Keep the same order and
  // avoid making dashboard and every other authenticated route use different tabs.
  const items = applyNavVisibility(buildSelfNavItems(true))
  return items.length > 0 ? [{ id: 'user', label: '用户导航', showLabel: false, items }] : []
})

function isNavActive(path: string, exact = false) {
  if (path === '/dashboard' && route.path === '/scheme3-dashboard') return true
  if (exact || path === '/dashboard') return route.path === path
  return route.path === path || route.path.startsWith(`${path}/`)
}

function navTourId(path: string): string | undefined {
  if (path === '/admin/accounts') return 'sidebar-channel-manage'
  if (path === '/admin/groups') return 'sidebar-group-manage'
  if (path === '/admin/redeem') return 'sidebar-wallet'
  return undefined
}

function handleNavItemClick(itemPath: string) {
  closeMobileNav()

  const selector: Record<string, string> = {
    '/admin/groups': '#sidebar-group-manage',
    '/admin/accounts': '#sidebar-channel-manage',
    '/keys': '[data-tour="sidebar-my-keys"]',
  }
  const tourSelector = selector[itemPath]
  if (tourSelector && onboardingStore.isCurrentStep(tourSelector)) {
    void onboardingStore.nextStep(500).catch(() => undefined)
  }
}

function handleReplayGuide() {
  closeAccountMenu()
  onboardingStore.replay()
}

function isGroupActive(item: ConsoleNavItem): boolean {
  return Boolean(item.children?.some((child) => isNavActive(child.path, true)))
}

function isGroupExpanded(item: ConsoleNavItem): boolean {
  return expandedNavGroups.value.has(item.path) || isGroupActive(item)
}

function toggleNavGroup(item: ConsoleNavItem) {
  if (navCollapsed.value) return
  if (!item.expandOnly) {
    if (route.path !== item.path) void router.push({ path: item.path, query: item.query })
    const next = new Set(expandedNavGroups.value)
    next.add(item.path)
    expandedNavGroups.value = next
    return
  }
  const next = new Set(expandedNavGroups.value)
  if (next.has(item.path)) next.delete(item.path)
  else next.add(item.path)
  expandedNavGroups.value = next
}

function openMobileNav() { mobileNavOpen.value = true }
function closeMobileNav() { mobileNavOpen.value = false }
function accountMenuItems(): HTMLElement[] {
  return Array.from(accountMenuRef.value?.querySelectorAll<HTMLElement>('[role="menuitem"]') ?? [])
}
function focusAccountMenuItem(index: number) {
  const items = accountMenuItems()
  if (!items.length) return
  items[Math.max(0, Math.min(index, items.length - 1))]?.focus()
}
function toggleAccountMenu() {
  accountMenuOpen.value = !accountMenuOpen.value
  if (accountMenuOpen.value) void nextTick(() => focusAccountMenuItem(0))
}
function closeAccountMenu(options: { restoreFocus?: boolean } = {}) {
  accountMenuOpen.value = false
  if (options.restoreFocus) void nextTick(() => accountMenuButtonRef.value?.focus())
}
function handleAccountMenuOutside(event: MouseEvent) {
  if (accountMenuRef.value && !accountMenuRef.value.contains(event.target as Node)) closeAccountMenu()
}
function handleAccountMenuEscape(event: KeyboardEvent) {
  if (event.key === 'Escape' && accountMenuOpen.value) {
    event.preventDefault()
    closeAccountMenu({ restoreFocus: true })
  }
}
function handleAccountMenuKeydown(event: KeyboardEvent) {
  if (!accountMenuOpen.value) return
  const items = accountMenuItems()
  if (!items.length) return
  const currentIndex = Math.max(0, items.indexOf(document.activeElement as HTMLElement))
  if (event.key === 'ArrowDown') {
    event.preventDefault()
    focusAccountMenuItem((currentIndex + 1) % items.length)
  } else if (event.key === 'ArrowUp') {
    event.preventDefault()
    focusAccountMenuItem((currentIndex - 1 + items.length) % items.length)
  } else if (event.key === 'Home') {
    event.preventDefault()
    focusAccountMenuItem(0)
  } else if (event.key === 'End') {
    event.preventDefault()
    focusAccountMenuItem(items.length - 1)
  } else if (event.key === 'Escape') {
    event.preventDefault()
    event.stopPropagation()
    closeAccountMenu({ restoreFocus: true })
  }
}
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
  if (!Number.isFinite(value)) return '$0.00'
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 2, maximumFractionDigits: 4 }).format(value)
}
async function logout() {
  closeAccountMenu()
  await authStore.logout().catch(() => undefined)
  await router.push('/login')
}

watch(() => route.fullPath, () => {
  closeMobileNav()
  closeAccountMenu()
})
watch(
  adminMode,
  (enabled) => {
    if (enabled) void adminSettingsStore.fetch()
  },
  { immediate: true },
)
onMounted(() => {
  document.body.classList.add(adminMode.value ? 'scheme3-admin-context' : 'scheme3-user-context')
  // API-key-backed capabilities belong to the authenticated user, including
  // an admin account. The upstream sidebar evaluates this for both identities.
  void refreshBatchImageAccess()
  if (adminMode.value) void adminSettingsStore.fetch()
  if (appStore.sidebarScrollTop > 0) {
    void nextTick(() => {
      if (sidebarNavRef.value) sidebarNavRef.value.scrollTop = appStore.sidebarScrollTop
    })
  }
  document.addEventListener('click', handleAccountMenuOutside)
  document.addEventListener('keydown', handleAccountMenuEscape)
})
onBeforeUnmount(() => {
  if (sidebarNavRef.value) appStore.sidebarScrollTop = sidebarNavRef.value.scrollTop
  document.removeEventListener('click', handleAccountMenuOutside)
  document.removeEventListener('keydown', handleAccountMenuEscape)
  document.body.classList.remove(adminMode.value ? 'scheme3-admin-context' : 'scheme3-user-context')
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
.scheme3-console-layout-admin { --scheme3-admin-accent: #1e5c42; }
.scheme3-console-layout-admin .scheme3-console-sidebar { border-right-color: #cfc8b8; }
.scheme3-console-layout-admin .scheme3-console-brand-copy strong { letter-spacing: .045em; }
.scheme3-console-layout-admin .scheme3-console-caption { color: #8b8578; }
.scheme3-console-layout-admin .scheme3-console-section-label { color: #8b8578; }
.scheme3-console-layout-admin .scheme3-console-link { min-height: 2.35rem; border-radius: 5px; font-size: .68rem; }
.scheme3-console-layout-admin .scheme3-console-link-active { box-shadow: inset 3px 0 0 var(--scheme3-admin-accent); }
.scheme3-console-layout-admin .scheme3-console-topbar-left strong { font-size: 1.05rem; }
.scheme3-console-layout-admin .scheme3-console-content { padding-top: 1.35rem; }
.scheme3-console-layout-admin .scheme3-console-page-frame { min-height: calc(100vh - 7.2rem); }
.scheme3-console-sidebar { position: sticky; top: 0; z-index: 50; display: flex; height: 100vh; min-width: 0; flex-direction: column; border-right: 1px solid var(--scheme3-line); background: rgba(251,250,246,.93); backdrop-filter: blur(22px); }
.scheme3-console-brand { display: flex; min-height: 5.25rem; align-items: center; gap: .7rem; border-bottom: 1px solid var(--scheme3-line); padding: 1rem; }
.scheme3-console-brand-mark { display: inline-flex; width: 2.35rem; height: 2.35rem; flex-shrink: 0; align-items: center; justify-content: center; overflow: hidden; border: 1px solid rgba(30,92,66,.35); border-radius: 9px; background: #1e5c42; color: #f4f2ec; font-size: .68rem; font-weight: 800; letter-spacing: .08em; box-shadow: 0 8px 18px rgba(30,92,66,.18); }
.scheme3-console-brand-mark img { width: 100%; height: 100%; object-fit: contain; }
.scheme3-console-brand-copy { min-width: 0; }
.scheme3-console-brand-copy strong { display: block; overflow: hidden; color: var(--scheme3-ink); font-size: .76rem; font-weight: 800; letter-spacing: .02em; text-overflow: ellipsis; white-space: nowrap; }
.scheme3-console-brand-copy span { display: block; margin-top: .18rem; color: var(--scheme3-muted); font-size: .62rem; }
.scheme3-console-version-control { min-width: 0; flex-shrink: 0; }
.scheme3-console-version-control :deep(.scheme3-version-static) { color: var(--scheme3-muted); font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; font-size: .54rem; font-weight: 800; }
.scheme3-console-version-control :deep(.scheme3-version-trigger) { display: inline-flex; min-height: 1.55rem; align-items: center; gap: .3rem; border: 1px solid var(--scheme3-line); border-radius: 5px; padding: .22rem .38rem; background: transparent; color: var(--scheme3-muted); font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; font-size: .54rem; font-weight: 800; }
.scheme3-console-version-control :deep(.scheme3-version-trigger:hover) { border-color: #1e5c42; background: rgba(30,92,66,.08); color: #1e5c42; }
.scheme3-console-version-control :deep(.scheme3-version-dropdown) { top: calc(100% + .5rem); left: auto; right: 0; z-index: 90; border: 1px solid var(--scheme3-line); border-radius: 7px; background: var(--scheme3-card); box-shadow: 0 1rem 2.5rem rgba(54,48,34,.16); }
.scheme3-console-version-control :deep(.scheme3-version-dropdown button) { border-radius: 5px; }
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
.scheme3-console-svg-icon :deep(svg) { width: 1rem; height: 1rem; }
.scheme3-console-link-text { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.scheme3-console-current { margin-left: auto; border-radius: 999px; padding: .16rem .35rem; background: rgba(30,92,66,.1); color: #1e5c42; font-size: .54rem; font-weight: 800; }
.scheme3-console-group { width: 100%; cursor: pointer; text-align: left; }
.scheme3-console-group-chevron { flex-shrink: 0; margin-left: auto; transition: transform 160ms ease; }
.scheme3-console-group-chevron.is-expanded { transform: rotate(180deg); }
.scheme3-console-subnav { display: flex; flex-direction: column; gap: .2rem; margin: 0 0 .2rem .82rem; border-left: 1px solid var(--scheme3-line); padding-left: .42rem; }
.scheme3-console-subnav-link { min-height: 2.2rem; padding-top: .42rem; padding-bottom: .42rem; font-size: .67rem; }
.scheme3-console-subnav-link .scheme3-console-link-icon :deep(svg) { width: .88rem; height: .88rem; }
.scheme3-console-doc-link { display: inline-flex; align-items: center; gap: .32rem; border: 1px solid transparent; border-radius: 6px; padding: .32rem .45rem; color: var(--scheme3-muted); font-size: .62rem; font-weight: 700; white-space: nowrap; }
.scheme3-console-doc-link:hover { border-color: var(--scheme3-line); background: rgba(255,255,255,.72); color: var(--scheme3-ink); }
.scheme3-console-subscription { flex-shrink: 0; }
.scheme3-console-foot { margin-top: auto; border-top: 1px solid var(--scheme3-line); padding: .8rem .65rem; }
.scheme3-console-account { display: flex; min-width: 0; align-items: center; gap: .55rem; border-radius: 7px; padding: .45rem; color: inherit; text-decoration: none; }
.scheme3-console-account:hover { background: rgba(255,255,255,.72); }
.scheme3-console-avatar { display: inline-flex; width: 1.85rem; height: 1.85rem; flex-shrink: 0; align-items: center; justify-content: center; overflow: hidden; border-radius: 7px; background: #16150f; color: #f4f2ec; font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; font-size: .6rem; font-weight: 800; }
.scheme3-console-avatar img { width: 100%; height: 100%; object-fit: cover; }
.scheme3-console-account-copy { min-width: 0; }
.scheme3-console-account-copy strong { display: block; overflow: hidden; color: var(--scheme3-ink); font-size: .67rem; text-overflow: ellipsis; white-space: nowrap; }
.scheme3-console-account-copy span { display: block; overflow: hidden; margin-top: .12rem; color: var(--scheme3-muted); font-size: .56rem; text-overflow: ellipsis; white-space: nowrap; }
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
.scheme3-console-topbar-left small { display: block; margin-top: .18rem; max-width: 34rem; overflow: hidden; color: var(--scheme3-muted); font-size: .58rem; text-overflow: ellipsis; white-space: nowrap; }
.scheme3-console-account-menu { position: relative; min-width: 0; }
.scheme3-console-user { display: inline-flex; min-width: 0; align-items: center; gap: .45rem; border: 1px solid transparent; border-radius: 7px; padding: .18rem .3rem; background: transparent; color: var(--scheme3-ink); font-size: .67rem; font-weight: 700; text-decoration: none; cursor: pointer; }
.scheme3-console-user:hover,.scheme3-console-user[aria-expanded="true"] { border-color: var(--scheme3-line); background: rgba(255,255,255,.72); }
.scheme3-console-user-copy { display: grid; min-width: 0; gap: .08rem; }
.scheme3-console-user-copy strong { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.scheme3-console-user-copy small { color: var(--scheme3-muted); font-size: .52rem; font-weight: 700; }
.scheme3-console-user-chevron { flex-shrink: 0; color: var(--scheme3-muted); transition: transform 160ms ease; }
.scheme3-console-user-chevron.is-open { transform: rotate(180deg); }
.scheme3-console-account-popover { position: absolute; top: calc(100% + .55rem); right: 0; z-index: 80; width: min(17rem, calc(100vw - 1.3rem)); max-height: min(32rem, calc(100vh - 4.5rem)); max-height: min(32rem, calc(100dvh - 4.5rem)); overflow-x: hidden; overflow-y: auto; overscroll-behavior: contain; border: 1px solid var(--scheme3-line); background: var(--scheme3-card); box-shadow: 0 1rem 2.5rem rgba(54,48,34,.16); scrollbar-width: thin; }
.scheme3-console-account-summary { display: grid; gap: .18rem; border-bottom: 1px solid var(--scheme3-line); padding: .85rem 1rem .75rem; }
.scheme3-console-account-summary strong { overflow: hidden; color: var(--scheme3-ink); font-size: .78rem; text-overflow: ellipsis; white-space: nowrap; }
.scheme3-console-account-summary span,.scheme3-console-account-summary small { overflow: hidden; color: var(--scheme3-muted); font-size: .6rem; text-overflow: ellipsis; white-space: nowrap; }
.scheme3-console-account-summary small { color: #1e5c42; font-weight: 800; }
.scheme3-console-account-balance { display: grid; gap: .38rem; border-bottom: 1px solid var(--scheme3-line); padding: .7rem 1rem; color: var(--scheme3-muted); font-size: .61rem; }
.scheme3-console-account-balance div { display: flex; align-items: center; justify-content: space-between; gap: .8rem; }
.scheme3-console-account-balance strong { color: var(--scheme3-ink); font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; font-size: .65rem; }
.scheme3-console-account-total { border-top: 1px solid var(--scheme3-line); padding-top: .4rem; }
.scheme3-console-account-links { display: grid; gap: .16rem; padding: .55rem; }
.scheme3-console-account-links a,.scheme3-console-account-links button,.scheme3-console-account-logout { display: flex; width: 100%; min-height: 2.15rem; align-items: center; gap: .5rem; border: 0; padding: .42rem .5rem; background: transparent; color: var(--scheme3-muted); font-size: .66rem; font-weight: 700; text-align: left; text-decoration: none; cursor: pointer; }
.scheme3-console-account-links a:hover,.scheme3-console-account-links button:hover { background: rgba(30,92,66,.08); color: var(--scheme3-ink); }
.scheme3-console-account-contact { display: flex; align-items: center; gap: .45rem; border-top: 1px solid var(--scheme3-line); padding: .65rem 1rem; color: var(--scheme3-muted); font-size: .59rem; }
.scheme3-console-account-contact span { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.scheme3-console-account-logout { border-top: 1px solid var(--scheme3-line); padding: .65rem 1rem; color: #9e4d3d; }
.scheme3-console-account-logout:hover { background: rgba(158,77,61,.08); color: #7d372d; }
.scheme3-console-menu-button { display: none; align-items: center; justify-content: center; border: 1px solid var(--scheme3-line); border-radius: 7px; padding: .45rem; background: var(--scheme3-card); color: var(--scheme3-ink); }
.scheme3-console-overlay { display: none; }
.scheme3-console-content { min-width: 0; padding: 1.15rem 1.35rem 1.6rem; }
.scheme3-console-page-frame { min-height: calc(100vh - 7rem); overflow-x: hidden; }
.scheme3-console-page-frame :deep(.card) { border-color: var(--scheme3-line); border-radius: 7px; background: var(--scheme3-card); box-shadow: none; }
.scheme3-console-page-frame :deep(.rounded-3xl),
.scheme3-console-page-frame :deep(.rounded-2xl),
.scheme3-console-page-frame :deep(.rounded-xl),
.scheme3-console-page-frame :deep(.rounded-lg) { border-radius: 7px !important; }
.scheme3-console-page-frame :deep(.ring-1) { --tw-ring-color: var(--scheme3-line) !important; }
.scheme3-console-page-frame :deep(.shadow),
.scheme3-console-page-frame :deep(.shadow-sm),
.scheme3-console-page-frame :deep(.shadow-2xl),
.scheme3-console-page-frame :deep(.shadow-xl),
.scheme3-console-page-frame :deep(.shadow-lg),
.scheme3-console-page-frame :deep(.shadow-md) { box-shadow: 0 5px 14px rgba(54,48,34,.07) !important; }
.scheme3-console-page-frame :deep(.bg-gradient-to-r),
.scheme3-console-page-frame :deep(.bg-gradient-to-l),
.scheme3-console-page-frame :deep(.bg-gradient-to-b),
.scheme3-console-page-frame :deep(.bg-gradient-to-t),
.scheme3-console-page-frame :deep(.bg-gradient-to-br),
.scheme3-console-page-frame :deep(.bg-gradient-to-bl),
.scheme3-console-page-frame :deep(.bg-gradient-to-tr),
.scheme3-console-page-frame :deep(.bg-gradient-to-tl) { background-image: none !important; }
.scheme3-console-page-frame :deep(.btn) {
  border-radius: 7px !important;
  box-shadow: none !important;
  font-size: .7rem;
  letter-spacing: .01em;
}
.scheme3-console-page-frame :deep(.btn-primary) { background: #1e5c42; box-shadow: 0 8px 18px rgba(30,92,66,.16); }
.scheme3-console-page-frame :deep(.btn-primary:hover) { background: #174a35; }
.scheme3-console-page-frame :deep(.btn-secondary) { border-color: var(--scheme3-line); background: var(--scheme3-card); color: var(--scheme3-ink); }
.scheme3-console-page-frame :deep(.btn-danger) { border-color: #9e4d3d; background: #9e4d3d; color: #fffaf5; }
.scheme3-console-page-frame :deep(.btn-success) { border-color: #1e5c42; background: #1e5c42; color: #fffaf5; }
.scheme3-console-page-frame :deep(.btn-warning) { border-color: #b7791f; background: #b7791f; color: #fffaf5; }
.scheme3-console-page-frame :deep(.btn-stripe) { border-color: #1e5c42; background: #1e5c42; color: #fffaf5; }
.scheme3-console-page-frame :deep(.input),.scheme3-console-page-frame :deep(select),.scheme3-console-page-frame :deep(textarea),.scheme3-console-page-frame :deep(input:not([type='checkbox']):not([type='radio']):not([type='range']):not([type='color'])) { border-color: var(--scheme3-line); border-radius: 6px !important; background: var(--scheme3-card); color: var(--scheme3-ink); box-shadow: none; }
.scheme3-console-page-frame :deep(.input:focus),.scheme3-console-page-frame :deep(select:focus),.scheme3-console-page-frame :deep(textarea:focus) { border-color: #1e5c42; box-shadow: 0 0 0 3px rgba(30,92,66,.12); }
.scheme3-console-page-frame :deep(.bg-white) { background-color: var(--scheme3-card); }
.scheme3-console-page-frame :deep(.border-gray-100),.scheme3-console-page-frame :deep(.border-gray-200),.scheme3-console-page-frame :deep(.border-gray-300),.scheme3-console-page-frame :deep(.border-gray-400) { border-color: var(--scheme3-line) !important; }
.scheme3-console-page-frame :deep(.text-gray-950),.scheme3-console-page-frame :deep(.text-gray-900),.scheme3-console-page-frame :deep(.text-gray-800),.scheme3-console-page-frame :deep(.text-gray-700) { color: var(--scheme3-ink) !important; }
.scheme3-console-page-frame :deep(.text-gray-600),.scheme3-console-page-frame :deep(.text-gray-500),.scheme3-console-page-frame :deep(.text-gray-400) { color: var(--scheme3-muted) !important; }
.scheme3-console-page-frame :deep(.input-hint) { color: var(--scheme3-muted) !important; }
.scheme3-console-page-frame :deep(.text-primary-900),.scheme3-console-page-frame :deep(.text-primary-800),.scheme3-console-page-frame :deep(.text-primary-700),.scheme3-console-page-frame :deep(.text-primary-600),.scheme3-console-page-frame :deep(.text-primary-500) { color: #1e5c42 !important; }
.scheme3-console-page-frame :deep(.bg-gray-50),.scheme3-console-page-frame :deep(.bg-gray-100),.scheme3-console-page-frame :deep(.bg-gray-200),.scheme3-console-page-frame :deep(.bg-gray-300),.scheme3-console-page-frame :deep(.bg-primary-50),.scheme3-console-page-frame :deep(.bg-primary-100),.scheme3-console-page-frame :deep(.bg-primary-200),.scheme3-console-page-frame :deep(.bg-primary-300) { background-color: var(--scheme3-subtle) !important; }
.scheme3-console-page-frame :deep(.bg-primary-400) { background-color: rgba(30,92,66,.18) !important; }
.scheme3-console-page-frame :deep(.bg-primary-500),.scheme3-console-page-frame :deep(.bg-primary-600),.scheme3-console-page-frame :deep(.bg-primary-700) { background-color: #1e5c42 !important; }
.scheme3-console-page-frame :deep(.bg-primary-800),.scheme3-console-page-frame :deep(.bg-primary-900) { background-color: #174a35 !important; }
.scheme3-console-page-frame :deep(.border-primary-50),.scheme3-console-page-frame :deep(.border-primary-100),.scheme3-console-page-frame :deep(.border-primary-200),.scheme3-console-page-frame :deep(.border-primary-300),.scheme3-console-page-frame :deep(.border-primary-400),.scheme3-console-page-frame :deep(.border-primary-500),.scheme3-console-page-frame :deep(.border-primary-600),.scheme3-console-page-frame :deep(.border-primary-700),.scheme3-console-page-frame :deep(.border-primary-800),.scheme3-console-page-frame :deep(.border-primary-900) { border-color: rgba(30,92,66,.32) !important; }
.scheme3-console-page-frame :deep(.border-t-transparent),.scheme3-console-page-frame :deep(.border-r-transparent),.scheme3-console-page-frame :deep(.border-b-transparent),.scheme3-console-page-frame :deep(.border-l-transparent) { border-color: transparent !important; }
.scheme3-console-page-frame :deep(.ring-primary-50),.scheme3-console-page-frame :deep(.ring-primary-100),.scheme3-console-page-frame :deep(.ring-primary-200),.scheme3-console-page-frame :deep(.ring-primary-300),.scheme3-console-page-frame :deep(.ring-primary-400),.scheme3-console-page-frame :deep(.ring-primary-500),.scheme3-console-page-frame :deep(.ring-primary-600),.scheme3-console-page-frame :deep(.ring-primary-700),.scheme3-console-page-frame :deep(.ring-primary-800),.scheme3-console-page-frame :deep(.ring-primary-900) { --tw-ring-color: rgba(30,92,66,.22) !important; }
.scheme3-console-page-frame :deep(.bg-cyan-50),.scheme3-console-page-frame :deep(.bg-cyan-100),.scheme3-console-page-frame :deep(.bg-teal-50),.scheme3-console-page-frame :deep(.bg-teal-100),.scheme3-console-page-frame :deep(.bg-blue-50),.scheme3-console-page-frame :deep(.bg-blue-100),.scheme3-console-page-frame :deep(.bg-indigo-50),.scheme3-console-page-frame :deep(.bg-indigo-100),.scheme3-console-page-frame :deep(.bg-violet-50),.scheme3-console-page-frame :deep(.bg-violet-100),.scheme3-console-page-frame :deep(.bg-purple-50),.scheme3-console-page-frame :deep(.bg-purple-100) { background-color: rgba(30,92,66,.08) !important; }
.scheme3-console-page-frame :deep(.text-cyan-500),.scheme3-console-page-frame :deep(.text-cyan-600),.scheme3-console-page-frame :deep(.text-cyan-700),.scheme3-console-page-frame :deep(.text-teal-500),.scheme3-console-page-frame :deep(.text-teal-600),.scheme3-console-page-frame :deep(.text-teal-700),.scheme3-console-page-frame :deep(.text-blue-500),.scheme3-console-page-frame :deep(.text-blue-600),.scheme3-console-page-frame :deep(.text-blue-700),.scheme3-console-page-frame :deep(.text-indigo-500),.scheme3-console-page-frame :deep(.text-indigo-600),.scheme3-console-page-frame :deep(.text-indigo-700),.scheme3-console-page-frame :deep(.text-violet-500),.scheme3-console-page-frame :deep(.text-violet-600),.scheme3-console-page-frame :deep(.text-violet-700),.scheme3-console-page-frame :deep(.text-purple-500),.scheme3-console-page-frame :deep(.text-purple-600),.scheme3-console-page-frame :deep(.text-purple-700) { color: #1e5c42 !important; }
.scheme3-console-page-frame :deep(.border-cyan-100),.scheme3-console-page-frame :deep(.border-cyan-200),.scheme3-console-page-frame :deep(.border-teal-100),.scheme3-console-page-frame :deep(.border-teal-200),.scheme3-console-page-frame :deep(.border-blue-100),.scheme3-console-page-frame :deep(.border-blue-200),.scheme3-console-page-frame :deep(.border-indigo-100),.scheme3-console-page-frame :deep(.border-indigo-200),.scheme3-console-page-frame :deep(.border-violet-100),.scheme3-console-page-frame :deep(.border-violet-200),.scheme3-console-page-frame :deep(.border-purple-100),.scheme3-console-page-frame :deep(.border-purple-200) { border-color: rgba(30,92,66,.25) !important; }
.scheme3-console-page-frame :deep(.bg-amber-50),.scheme3-console-page-frame :deep(.bg-amber-100),.scheme3-console-page-frame :deep(.bg-yellow-50),.scheme3-console-page-frame :deep(.bg-yellow-100),.scheme3-console-page-frame :deep(.bg-orange-50),.scheme3-console-page-frame :deep(.bg-orange-100) { background-color: rgba(183,121,31,.08) !important; }
.scheme3-console-page-frame :deep(.text-amber-500),.scheme3-console-page-frame :deep(.text-amber-600),.scheme3-console-page-frame :deep(.text-amber-700),.scheme3-console-page-frame :deep(.text-amber-800),.scheme3-console-page-frame :deep(.text-yellow-500),.scheme3-console-page-frame :deep(.text-yellow-600),.scheme3-console-page-frame :deep(.text-yellow-700),.scheme3-console-page-frame :deep(.text-orange-500),.scheme3-console-page-frame :deep(.text-orange-600),.scheme3-console-page-frame :deep(.text-orange-700) { color: #8b5d14 !important; }
.scheme3-console-page-frame :deep(.border-amber-100),.scheme3-console-page-frame :deep(.border-amber-200),.scheme3-console-page-frame :deep(.border-amber-300),.scheme3-console-page-frame :deep(.border-yellow-100),.scheme3-console-page-frame :deep(.border-yellow-200),.scheme3-console-page-frame :deep(.border-orange-100),.scheme3-console-page-frame :deep(.border-orange-200) { border-color: rgba(183,121,31,.32) !important; }
.scheme3-console-page-frame :deep(.bg-green-50),.scheme3-console-page-frame :deep(.bg-green-100),.scheme3-console-page-frame :deep(.bg-emerald-50),.scheme3-console-page-frame :deep(.bg-emerald-100) { background-color: rgba(30,92,66,.08) !important; }
.scheme3-console-page-frame :deep(.text-green-500),.scheme3-console-page-frame :deep(.text-green-600),.scheme3-console-page-frame :deep(.text-green-700),.scheme3-console-page-frame :deep(.text-emerald-500),.scheme3-console-page-frame :deep(.text-emerald-600),.scheme3-console-page-frame :deep(.text-emerald-700) { color: #1e5c42 !important; }
.scheme3-console-page-frame :deep(.border-green-100),.scheme3-console-page-frame :deep(.border-green-200),.scheme3-console-page-frame :deep(.border-green-300),.scheme3-console-page-frame :deep(.border-emerald-100),.scheme3-console-page-frame :deep(.border-emerald-200),.scheme3-console-page-frame :deep(.border-emerald-300) { border-color: rgba(30,92,66,.28) !important; }
.scheme3-console-page-frame :deep(.bg-red-50),.scheme3-console-page-frame :deep(.bg-red-100),.scheme3-console-page-frame :deep(.bg-rose-50),.scheme3-console-page-frame :deep(.bg-rose-100) { background-color: rgba(158,77,61,.07) !important; }
.scheme3-console-page-frame :deep(.text-red-500),.scheme3-console-page-frame :deep(.text-red-600),.scheme3-console-page-frame :deep(.text-red-700),.scheme3-console-page-frame :deep(.text-rose-500),.scheme3-console-page-frame :deep(.text-rose-600),.scheme3-console-page-frame :deep(.text-rose-700) { color: #9e4d3d !important; }
.scheme3-console-page-frame :deep(.border-red-100),.scheme3-console-page-frame :deep(.border-red-200),.scheme3-console-page-frame :deep(.border-red-300),.scheme3-console-page-frame :deep(.border-rose-100),.scheme3-console-page-frame :deep(.border-rose-200),.scheme3-console-page-frame :deep(.border-rose-300) { border-color: rgba(158,77,61,.28) !important; }
.scheme3-console-page-frame :deep(table thead),.scheme3-console-page-frame :deep(thead) { background: var(--scheme3-subtle); }
.scheme3-console-page-frame :deep(th),.scheme3-console-page-frame :deep(td) { border-color: var(--scheme3-line); color: var(--scheme3-ink); }

:global(.dark .scheme3-console-layout) { --scheme3-ink: #f4f2ec; --scheme3-muted: #aaa69a; --scheme3-line: #47443a; --scheme3-paper: #1b1b18; --scheme3-card: #24231f; --scheme3-subtle: #2b2924; background: var(--scheme3-paper); }
:global(.dark .scheme3-console-layout-admin) { --scheme3-admin-accent: #8fc2a5; }
:global(.dark .scheme3-console-layout-admin .scheme3-console-sidebar) { border-right-color: #47443a; }
:global(.dark .scheme3-console-layout-admin .scheme3-console-caption),:global(.dark .scheme3-console-layout-admin .scheme3-console-section-label) { color: #827e72; }
:global(.dark .scheme3-console-layout-admin .scheme3-console-link) { color: #aaa69a; }
:global(.dark .scheme3-console-layout-admin .scheme3-console-link-active) { box-shadow: inset 3px 0 0 var(--scheme3-admin-accent); }
:global(.dark .scheme3-console-sidebar) { background: rgba(36,35,31,.95); }
:global(.dark .scheme3-console-topbar) { background: rgba(27,27,24,.93); }
:global(.dark .scheme3-console-link:hover),:global(.dark .scheme3-console-action:hover) { background: rgba(143,194,165,.08); }
:global(.dark .scheme3-console-brand-mark) { border-color: rgba(143,194,165,.38); background: #8fc2a5; color: #1b1b18; }
:global(.dark .scheme3-console-avatar) { background: #f4f2ec; color: #1b1b18; }
:global(.dark .scheme3-console-version-control .scheme3-version-trigger) { border-color: #47443a; background: transparent; color: #aaa69a; }
:global(.dark .scheme3-console-version-control .scheme3-version-trigger:hover) { border-color: #8fc2a5; background: rgba(143,194,165,.1); color: #8fc2a5; }
:global(.dark .scheme3-console-version-control .scheme3-version-static) { color: #aaa69a; }
:global(.dark .scheme3-console-version-control .scheme3-version-dropdown) { border-color: #47443a; background: #24231f; box-shadow: 0 1rem 2.5rem rgba(0,0,0,.34); }
:global(.dark .scheme3-console-user:hover),:global(.dark .scheme3-console-user[aria-expanded="true"]) { border-color: #47443a; background: rgba(143,194,165,.08); }
:global(.dark .scheme3-console-account-popover) { background: #24231f; box-shadow: 0 1rem 2.5rem rgba(0,0,0,.34); }
:global(.dark .scheme3-console-account-summary small) { color: #8fc2a5; }
:global(.dark .scheme3-console-account-links a:hover),:global(.dark .scheme3-console-account-links button:hover) { background: rgba(143,194,165,.1); color: #f4f2ec; }
:global(.dark .scheme3-console-account-logout:hover) { background: rgba(238,149,129,.1); color: #ee9581; }
:global(.dark .scheme3-console-link-active) { border-color: rgba(143,194,165,.3); background: rgba(143,194,165,.1); color: #8fc2a5; box-shadow: inset 3px 0 0 #8fc2a5; }
:global(.dark .scheme3-console-current) { background: rgba(143,194,165,.12); color: #8fc2a5; }
:global(.dark .scheme3-console-page-frame :deep(.card)),:global(.dark .scheme3-console-page-frame :deep(.bg-white)) { background-color: var(--scheme3-card); }
:global(.dark .scheme3-console-page-frame :deep(.bg-dark-800)),:global(.dark .scheme3-console-page-frame :deep(.bg-dark-900)),:global(.dark .scheme3-console-page-frame :deep(.bg-dark-950)) { background-color: var(--scheme3-card) !important; }
:global(.dark .scheme3-console-page-frame :deep(.text-dark-100)),:global(.dark .scheme3-console-page-frame :deep(.text-dark-200)),:global(.dark .scheme3-console-page-frame :deep(.text-dark-300)),:global(.dark .scheme3-console-page-frame :deep(.text-dark-400)) { color: var(--scheme3-muted) !important; }
:global(.dark .scheme3-console-page-frame :deep(.input-hint)) { color: var(--scheme3-muted) !important; }
:global(.dark .scheme3-console-page-frame :deep(.text-primary-900)),:global(.dark .scheme3-console-page-frame :deep(.text-primary-800)),:global(.dark .scheme3-console-page-frame :deep(.text-primary-700)),:global(.dark .scheme3-console-page-frame :deep(.text-primary-600)),:global(.dark .scheme3-console-page-frame :deep(.text-primary-500)) { color: #8fc2a5 !important; }
:global(.dark .scheme3-console-page-frame :deep(.bg-primary-200)),:global(.dark .scheme3-console-page-frame :deep(.bg-primary-300)) { background-color: #2b2924 !important; }
:global(.dark .scheme3-console-page-frame :deep(.bg-primary-400)) { background-color: rgba(143,194,165,.18) !important; }
:global(.dark .scheme3-console-page-frame :deep(.bg-primary-500)),:global(.dark .scheme3-console-page-frame :deep(.bg-primary-600)),:global(.dark .scheme3-console-page-frame :deep(.bg-primary-700)) { background-color: #8fc2a5 !important; }
:global(.dark .scheme3-console-page-frame :deep(.bg-primary-800)),:global(.dark .scheme3-console-page-frame :deep(.bg-primary-900)) { background-color: #6fa887 !important; }
:global(.dark .scheme3-console-page-frame :deep(.border-primary-50)),:global(.dark .scheme3-console-page-frame :deep(.border-primary-100)),:global(.dark .scheme3-console-page-frame :deep(.border-primary-200)),:global(.dark .scheme3-console-page-frame :deep(.border-primary-300)),:global(.dark .scheme3-console-page-frame :deep(.border-primary-400)),:global(.dark .scheme3-console-page-frame :deep(.border-primary-500)),:global(.dark .scheme3-console-page-frame :deep(.border-primary-600)),:global(.dark .scheme3-console-page-frame :deep(.border-primary-700)),:global(.dark .scheme3-console-page-frame :deep(.border-primary-800)),:global(.dark .scheme3-console-page-frame :deep(.border-primary-900)) { border-color: rgba(143,194,165,.34) !important; }
:global(.dark .scheme3-console-page-frame :deep(.border-t-transparent)),:global(.dark .scheme3-console-page-frame :deep(.border-r-transparent)),:global(.dark .scheme3-console-page-frame :deep(.border-b-transparent)),:global(.dark .scheme3-console-page-frame :deep(.border-l-transparent)) { border-color: transparent !important; }
:global(.dark .scheme3-console-page-frame :deep(.ring-primary-50)),:global(.dark .scheme3-console-page-frame :deep(.ring-primary-100)),:global(.dark .scheme3-console-page-frame :deep(.ring-primary-200)),:global(.dark .scheme3-console-page-frame :deep(.ring-primary-300)),:global(.dark .scheme3-console-page-frame :deep(.ring-primary-400)),:global(.dark .scheme3-console-page-frame :deep(.ring-primary-500)),:global(.dark .scheme3-console-page-frame :deep(.ring-primary-600)),:global(.dark .scheme3-console-page-frame :deep(.ring-primary-700)),:global(.dark .scheme3-console-page-frame :deep(.ring-primary-800)),:global(.dark .scheme3-console-page-frame :deep(.ring-primary-900)) { --tw-ring-color: rgba(143,194,165,.24) !important; }

/* Teleported user controls do not inherit the page-frame variables. Scope
   these overrides to the user console body marker so admin dialogs retain
   their upstream styling. */
:global(body.scheme3-user-context .select-dropdown-portal) { border-color: #dad5c8 !important; background: #fbfaf6 !important; color: #16150f !important; border-radius: 7px; box-shadow: 0 16px 34px rgba(54,48,34,.16) !important; }
:global(body.scheme3-user-context .select-dropdown-portal .select-search) { border-bottom-color: #dad5c8 !important; }
:global(body.scheme3-user-context .select-dropdown-portal .select-search-input) { color: #16150f !important; }
:global(body.scheme3-user-context .select-dropdown-portal .select-option) { color: #6b695f !important; }
:global(body.scheme3-user-context .select-dropdown-portal .select-option:hover),:global(body.scheme3-user-context .select-dropdown-portal .select-option-focused) { background: #f1eee6 !important; color: #16150f !important; }
:global(body.scheme3-user-context .select-dropdown-portal .select-option-selected) { background: rgba(30,92,66,.1) !important; color: #1e5c42 !important; }
:global(body.scheme3-user-context .select-dropdown-portal .select-option-check) { color: #1e5c42 !important; }
:global(body.scheme3-user-context .select-dropdown-portal .select-option-group) { background: #f1eee6 !important; color: #6b695f !important; }
:global(body.scheme3-user-context .select-dropdown-portal .select-empty) { color: #6b695f !important; }
:global(html.dark body.scheme3-user-context .select-dropdown-portal) { border-color: #47443a !important; background: #24231f !important; color: #f4f2ec !important; box-shadow: 0 18px 38px rgba(0,0,0,.28) !important; }
:global(html.dark body.scheme3-user-context .select-dropdown-portal .select-search) { border-bottom-color: #47443a !important; }
:global(html.dark body.scheme3-user-context .select-dropdown-portal .select-search-input) { color: #f4f2ec !important; }
:global(html.dark body.scheme3-user-context .select-dropdown-portal .select-option) { color: #aaa69a !important; }
:global(html.dark body.scheme3-user-context .select-dropdown-portal .select-option:hover),:global(html.dark body.scheme3-user-context .select-dropdown-portal .select-option-focused) { background: #2b2924 !important; color: #f4f2ec !important; }
:global(html.dark body.scheme3-user-context .select-dropdown-portal .select-option-selected) { background: rgba(143,194,165,.12) !important; color: #8fc2a5 !important; }
:global(html.dark body.scheme3-user-context .select-dropdown-portal .select-option-check) { color: #8fc2a5 !important; }
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

/* Shared components still carry a few semantic utility classes. Re-map those
   utilities inside the active third-version context so teleported controls and
   pages that still use the compatibility branch keep the same visual language. */
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

/* Administrator pages all use the third-version shell now, but a number of
   upstream views still render the legacy DataTable/Select/Tab class names.
   These shared rules are deliberately scoped to the body marker and use
   !important so component-scoped upstream rules cannot reintroduce the gray
   table header, white Select surface, or blue tab treatment at runtime. */
:global(body.scheme3-admin-context .select-trigger),
:global(body.scheme3-admin-context .scheme3-select-trigger),
:global(body.scheme3-admin-context .date-picker-trigger) {
  border-color: #dad5c8 !important;
  border-radius: 7px !important;
  background: #fbfaf6 !important;
  color: #16150f !important;
  box-shadow: none !important;
}
:global(body.scheme3-admin-context .select-trigger:hover),
:global(body.scheme3-admin-context .select-trigger-open),
:global(body.scheme3-admin-context .scheme3-select-trigger:hover),
:global(body.scheme3-admin-context .scheme3-select-trigger-open),
:global(body.scheme3-admin-context .date-picker-trigger:hover),
:global(body.scheme3-admin-context .date-picker-trigger-open) {
  border-color: #1e5c42 !important;
  box-shadow: 0 0 0 3px rgba(30,92,66,.11) !important;
}
:global(body.scheme3-admin-context .select-trigger-disabled),
:global(body.scheme3-admin-context .scheme3-select-trigger-disabled),
:global(body.scheme3-admin-context .date-picker-trigger:disabled) {
  background: #f1eee6 !important;
  color: #9a9588 !important;
  box-shadow: none !important;
  cursor: not-allowed;
}
:global(body.scheme3-admin-context .date-picker-dropdown) {
  border-color: #dad5c8 !important;
  border-radius: 7px !important;
  background: #fbfaf6 !important;
  box-shadow: 0 16px 34px rgba(54,48,34,.16) !important;
}
:global(body.scheme3-admin-context .date-picker-preset:hover),
:global(body.scheme3-admin-context .date-picker-preset-active) {
  border-radius: 5px !important;
  background: #f1eee6 !important;
  color: #1e5c42 !important;
}
:global(body.scheme3-admin-context .date-picker-input) {
  border-color: #dad5c8 !important;
  border-radius: 6px !important;
  background: #fffefa !important;
  color: #16150f !important;
}

:global(body.scheme3-admin-context .table-wrapper) {
  border-color: #dad5c8 !important;
  background: #fbfaf6 !important;
  color: #27251f !important;
}
:global(body.scheme3-admin-context .table-wrapper .table-header),
:global(body.scheme3-admin-context .table-header),
:global(body.scheme3-admin-context .sticky-header-cell) {
  border-color: #dad5c8 !important;
  background: #f1eee6 !important;
  color: #777266 !important;
}
:global(body.scheme3-admin-context .table-body) {
  border-color: #ebe7dc !important;
  background: #fbfaf6 !important;
  color: #27251f !important;
}
:global(body.scheme3-admin-context .table-body th),
:global(body.scheme3-admin-context .table-body td),
:global(body.scheme3-admin-context .table-wrapper th),
:global(body.scheme3-admin-context .table-wrapper td) {
  border-color: #ebe7dc !important;
}
:global(body.scheme3-admin-context .table-body .sticky-col),
:global(body.scheme3-admin-context tbody .sticky-col) {
  background: #fbfaf6 !important;
}
:global(body.scheme3-admin-context .table-body tr:hover),
:global(body.scheme3-admin-context .table-body tr:hover .sticky-col),
:global(body.scheme3-admin-context tbody tr:hover .sticky-col) {
  background: #f1eee6 !important;
}
:global(body.scheme3-admin-context .tabs) {
  gap: .2rem !important;
  border: 1px solid #dad5c8 !important;
  border-radius: 7px !important;
  background: #f1eee6 !important;
  padding: .25rem !important;
}
:global(body.scheme3-admin-context .tab) {
  border-radius: 5px !important;
  color: #6b695f !important;
  box-shadow: none !important;
}
:global(body.scheme3-admin-context .tab:hover) {
  background: #fffefa !important;
  color: #16150f !important;
}
:global(body.scheme3-admin-context .tab-active) {
  border: 1px solid rgba(30,92,66,.25) !important;
  background: #fbfaf6 !important;
  color: #1e5c42 !important;
  box-shadow: 0 1px 2px rgba(54,48,34,.08) !important;
}

:global(html.dark body.scheme3-admin-context .select-trigger),
:global(html.dark body.scheme3-admin-context .scheme3-select-trigger),
:global(html.dark body.scheme3-admin-context .date-picker-trigger) {
  border-color: #47443a !important;
  background: #24231f !important;
  color: #f4f2ec !important;
}
:global(html.dark body.scheme3-admin-context .select-trigger:hover),
:global(html.dark body.scheme3-admin-context .select-trigger-open),
:global(html.dark body.scheme3-admin-context .scheme3-select-trigger:hover),
:global(html.dark body.scheme3-admin-context .scheme3-select-trigger-open),
:global(html.dark body.scheme3-admin-context .date-picker-trigger:hover),
:global(html.dark body.scheme3-admin-context .date-picker-trigger-open) {
  border-color: #8fc2a5 !important;
  box-shadow: 0 0 0 3px rgba(143,194,165,.13) !important;
}
:global(html.dark body.scheme3-admin-context .select-trigger-disabled),
:global(html.dark body.scheme3-admin-context .scheme3-select-trigger-disabled),
:global(html.dark body.scheme3-admin-context .date-picker-trigger:disabled) {
  background: #2b2924 !important;
  color: #827e72 !important;
}
:global(html.dark body.scheme3-admin-context .date-picker-dropdown) {
  border-color: #47443a !important;
  background: #24231f !important;
  box-shadow: 0 18px 38px rgba(0,0,0,.28) !important;
}
:global(html.dark body.scheme3-admin-context .date-picker-preset:hover),
:global(html.dark body.scheme3-admin-context .date-picker-preset-active) {
  background: #2b2924 !important;
  color: #8fc2a5 !important;
}
:global(html.dark body.scheme3-admin-context .date-picker-input) {
  border-color: #47443a !important;
  background: #2b2924 !important;
  color: #f4f2ec !important;
}
:global(html.dark body.scheme3-admin-context .table-wrapper) {
  border-color: #47443a !important;
  background: #24231f !important;
  color: #f4f2ec !important;
}
:global(html.dark body.scheme3-admin-context .table-wrapper .table-header),
:global(html.dark body.scheme3-admin-context .table-header),
:global(html.dark body.scheme3-admin-context .sticky-header-cell) {
  border-color: #47443a !important;
  background: #2b2924 !important;
  color: #aaa69a !important;
}
:global(html.dark body.scheme3-admin-context .table-body) {
  border-color: #3a3830 !important;
  background: #24231f !important;
  color: #f4f2ec !important;
}
:global(html.dark body.scheme3-admin-context .table-body th),
:global(html.dark body.scheme3-admin-context .table-body td),
:global(html.dark body.scheme3-admin-context .table-wrapper th),
:global(html.dark body.scheme3-admin-context .table-wrapper td) {
  border-color: #3a3830 !important;
}
:global(html.dark body.scheme3-admin-context .table-body .sticky-col),
:global(html.dark body.scheme3-admin-context tbody .sticky-col) {
  background: #24231f !important;
}
:global(html.dark body.scheme3-admin-context .table-body tr:hover),
:global(html.dark body.scheme3-admin-context .table-body tr:hover .sticky-col),
:global(html.dark body.scheme3-admin-context tbody tr:hover .sticky-col) {
  background: #2b2924 !important;
}
:global(html.dark body.scheme3-admin-context .tabs) {
  border-color: #47443a !important;
  background: #2b2924 !important;
}
:global(html.dark body.scheme3-admin-context .tab) { color: #aaa69a !important; }
:global(html.dark body.scheme3-admin-context .tab:hover) { background: #24231f !important; color: #f4f2ec !important; }
:global(html.dark body.scheme3-admin-context .tab-active) {
  border-color: rgba(143,194,165,.3) !important;
  background: #24231f !important;
  color: #8fc2a5 !important;
  box-shadow: 0 1px 2px rgba(0,0,0,.2) !important;
}

:global(body.scheme3-admin-context .badge) {
  border: 1px solid transparent !important;
  border-radius: 5px !important;
  padding: .2rem .45rem !important;
  background: #f1eee6 !important;
  color: #6b695f !important;
  font-size: .62rem;
  font-weight: 800;
}
:global(body.scheme3-admin-context .badge-primary),
:global(body.scheme3-admin-context .badge-success) {
  border-color: rgba(30,92,66,.25) !important;
  background: rgba(30,92,66,.09) !important;
  color: #1e5c42 !important;
}
:global(body.scheme3-admin-context .badge-warning) {
  border-color: rgba(183,121,31,.3) !important;
  background: rgba(183,121,31,.09) !important;
  color: #8b5d14 !important;
}
:global(body.scheme3-admin-context .badge-danger) {
  border-color: rgba(158,77,61,.3) !important;
  background: rgba(158,77,61,.08) !important;
  color: #9e4d3d !important;
}
:global(body.scheme3-admin-context .badge-purple) {
  border-color: rgba(30,92,66,.25) !important;
  background: rgba(30,92,66,.09) !important;
  color: #1e5c42 !important;
}
:global(html.dark body.scheme3-admin-context .badge) {
  border-color: #47443a !important;
  background: #2b2924 !important;
  color: #aaa69a !important;
}
:global(html.dark body.scheme3-admin-context .badge-primary),
:global(html.dark body.scheme3-admin-context .badge-success) {
  border-color: rgba(143,194,165,.28) !important;
  background: rgba(143,194,165,.1) !important;
  color: #8fc2a5 !important;
}
:global(html.dark body.scheme3-admin-context .badge-warning) {
  border-color: rgba(211,164,92,.32) !important;
  background: rgba(211,164,92,.1) !important;
  color: #d3a45c !important;
}
:global(html.dark body.scheme3-admin-context .badge-danger) {
  border-color: rgba(211,139,121,.35) !important;
  background: rgba(211,139,121,.1) !important;
  color: #d38b79 !important;
}
:global(html.dark body.scheme3-admin-context .badge-purple) {
  border-color: rgba(143,194,165,.28) !important;
  background: rgba(143,194,165,.1) !important;
  color: #8fc2a5 !important;
}

/* The admin monitor uses the same third-version shell. Keep components that
   render through body (selects, dialogs, toasts and the onboarding driver)
   on the same paper palette instead of falling back to the upstream skin. */
:global(body.scheme3-admin-context .select-dropdown-portal) { border-color: #dad5c8 !important; background: #fbfaf6 !important; color: #16150f !important; border-radius: 7px; box-shadow: 0 16px 34px rgba(54,48,34,.16) !important; }
:global(body.scheme3-admin-context .select-dropdown-portal .select-search) { border-bottom-color: #dad5c8 !important; }
:global(body.scheme3-admin-context .select-dropdown-portal .select-search-input) { color: #16150f !important; }
:global(body.scheme3-admin-context .select-dropdown-portal .select-option) { color: #6b695f !important; }
:global(body.scheme3-admin-context .select-dropdown-portal .select-option:hover),:global(body.scheme3-admin-context .select-dropdown-portal .select-option-focused) { background: #f1eee6 !important; color: #16150f !important; }
:global(body.scheme3-admin-context .select-dropdown-portal .select-option-selected) { background: rgba(30,92,66,.1) !important; color: #1e5c42 !important; }
:global(body.scheme3-admin-context .select-dropdown-portal .select-option-check) { color: #1e5c42 !important; }
:global(body.scheme3-admin-context .select-dropdown-portal .select-option-group) { background: #f1eee6 !important; color: #6b695f !important; }
:global(body.scheme3-admin-context .select-dropdown-portal .select-empty) { color: #6b695f !important; }
:global(html.dark body.scheme3-admin-context .select-dropdown-portal) { border-color: #47443a !important; background: #24231f !important; color: #f4f2ec !important; box-shadow: 0 18px 38px rgba(0,0,0,.28) !important; }
:global(html.dark body.scheme3-admin-context .select-dropdown-portal .select-search) { border-bottom-color: #47443a !important; }
:global(html.dark body.scheme3-admin-context .select-dropdown-portal .select-search-input) { color: #f4f2ec !important; }
:global(html.dark body.scheme3-admin-context .select-dropdown-portal .select-option) { color: #aaa69a !important; }
:global(html.dark body.scheme3-admin-context .select-dropdown-portal .select-option:hover),:global(html.dark body.scheme3-admin-context .select-dropdown-portal .select-option-focused) { background: #2b2924 !important; color: #f4f2ec !important; }
:global(html.dark body.scheme3-admin-context .select-dropdown-portal .select-option-selected) { background: rgba(143,194,165,.12) !important; color: #8fc2a5 !important; }
:global(html.dark body.scheme3-admin-context .select-dropdown-portal .select-option-check) { color: #8fc2a5 !important; }
:global(html.dark body.scheme3-admin-context .select-dropdown-portal .select-option-group) { background: #2b2924 !important; color: #aaa69a !important; }
:global(html.dark body.scheme3-admin-context .select-dropdown-portal .select-empty) { color: #aaa69a !important; }

:global(body.scheme3-admin-context .modal-overlay) { background: rgba(21,20,16,.48) !important; backdrop-filter: blur(5px); }
:global(body.scheme3-admin-context .modal-content) { border: 1px solid #dad5c8; border-radius: 8px; background: #fbfaf6 !important; color: #16150f; box-shadow: 0 24px 60px rgba(54,48,34,.18) !important; }
:global(body.scheme3-admin-context .modal-header) { border-bottom-color: #dad5c8 !important; }
:global(body.scheme3-admin-context .modal-footer) { border-top-color: #dad5c8 !important; background: #f1eee6 !important; }
:global(body.scheme3-admin-context .modal-title) { color: #16150f !important; }
:global(body.scheme3-admin-context .modal-content .btn-primary) { background: #1e5c42 !important; }
:global(body.scheme3-admin-context .modal-content .btn-secondary) { border-color: #dad5c8 !important; background: #fbfaf6 !important; color: #16150f !important; }
:global(body.scheme3-admin-context .modal-content .input) { border-color: #dad5c8 !important; background: #fbfaf6 !important; color: #16150f !important; }
:global(html.dark body.scheme3-admin-context .modal-overlay) { background: rgba(0,0,0,.62) !important; }
:global(html.dark body.scheme3-admin-context .modal-content) { border-color: #47443a; background: #24231f !important; color: #f4f2ec; box-shadow: 0 24px 60px rgba(0,0,0,.32) !important; }
:global(html.dark body.scheme3-admin-context .modal-header) { border-bottom-color: #47443a !important; }
:global(html.dark body.scheme3-admin-context .modal-footer) { border-top-color: #47443a !important; background: #2b2924 !important; }
:global(html.dark body.scheme3-admin-context .modal-title) { color: #f4f2ec !important; }
:global(html.dark body.scheme3-admin-context .modal-content .btn-secondary) { border-color: #47443a !important; background: #24231f !important; color: #f4f2ec !important; }
:global(html.dark body.scheme3-admin-context .modal-content .input) { border-color: #47443a !important; background: #24231f !important; color: #f4f2ec !important; }

/* Monitor dialogs use a restrained operator palette even when their reusable
   form components are rendered through BaseDialog's body teleport. */
:global(body.scheme3-admin-context .scheme3-monitor-label) { display: block; margin-bottom: .35rem; color: #6b695f; font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; font-size: .6rem; font-weight: 800; letter-spacing: .04em; }
:global(body.scheme3-admin-context .scheme3-monitor-input) { width: 100%; min-height: 2.35rem; border: 1px solid #dad5c8; border-radius: 6px; background: #fbfaf6; color: #16150f; padding: .5rem .72rem; font-size: .72rem; outline: none; transition: border-color 150ms ease,box-shadow 150ms ease; }
:global(body.scheme3-admin-context .scheme3-monitor-input:focus) { border-color: #1e5c42; box-shadow: 0 0 0 3px rgba(30,92,66,.12); }
:global(body.scheme3-admin-context .scheme3-monitor-button) { display: inline-flex; min-height: 2.25rem; align-items: center; justify-content: center; gap: .4rem; border: 1px solid #dad5c8; border-radius: 6px; background: #fbfaf6; color: #6b695f; padding: .45rem .72rem; font-size: .68rem; font-weight: 800; transition: border-color 150ms ease,background-color 150ms ease,color 150ms ease,transform 120ms ease; }
:global(body.scheme3-admin-context .scheme3-monitor-button:hover:not(:disabled)) { border-color: #1e5c42; color: #1e5c42; }
:global(body.scheme3-admin-context .scheme3-monitor-button:active:not(:disabled)) { transform: scale(.98); }
:global(body.scheme3-admin-context .scheme3-monitor-button.is-primary) { border-color: #1e5c42; background: #1e5c42; color: #fffefa; box-shadow: 0 8px 18px rgba(30,92,66,.16); }
:global(body.scheme3-admin-context .scheme3-monitor-button.is-primary:hover:not(:disabled)) { background: #174a35; color: #fffefa; }
:global(body.scheme3-admin-context .scheme3-monitor-button.is-danger) { border-color: rgba(158,77,61,.35); color: #9e4d3d; }
:global(body.scheme3-admin-context .scheme3-monitor-button:focus-visible),:global(body.scheme3-admin-context .scheme3-monitor-choice:focus-visible) { outline: 2px solid rgba(30,92,66,.28); outline-offset: 2px; }
:global(body.scheme3-admin-context .scheme3-monitor-copy) { color: #6b695f; }
:global(body.scheme3-admin-context .scheme3-monitor-empty) { color: #777266; }
:global(body.scheme3-admin-context .scheme3-monitor-hint) { color: #9a9588; }
:global(body.scheme3-admin-context .scheme3-monitor-error),:global(body.scheme3-admin-context .scheme3-monitor-required) { color: #9e4d3d; }
:global(body.scheme3-admin-context .scheme3-monitor-summary) { color: #6b695f; }
:global(body.scheme3-admin-context .scheme3-monitor-input-icon) { color: #9a9588; pointer-events: none; }
:global(body.scheme3-admin-context .scheme3-monitor-warning) { color: #9e4d3d; }
:global(body.scheme3-admin-context .scheme3-monitor-provider-badge) { border: 1px solid currentColor; border-radius: 5px; background: transparent; }
:global(body.scheme3-admin-context .scheme3-monitor-provider-badge.is-openai) { border-color: rgba(30,92,66,.3); background: rgba(30,92,66,.09); color: #1e5c42; }
:global(body.scheme3-admin-context .scheme3-monitor-provider-badge.is-anthropic) { border-color: rgba(158,77,61,.3); background: rgba(158,77,61,.08); color: #9e4d3d; }
:global(body.scheme3-admin-context .scheme3-monitor-provider-badge.is-gemini) { border-color: rgba(183,121,31,.3); background: rgba(183,121,31,.09); color: #8b5d14; }
:global(body.scheme3-admin-context .scheme3-monitor-provider-badge.is-grok),:global(body.scheme3-admin-context .scheme3-monitor-provider-badge.is-unknown) { border-color: #dad5c8; background: #f1eee6; color: #777266; }
:global(body.scheme3-admin-context .scheme3-monitor-table-wrap) { border: 1px solid #dad5c8; border-radius: 7px; background: #fbfaf6; }
:global(body.scheme3-admin-context .scheme3-monitor-table-head) { border-bottom: 1px solid #dad5c8; background: #f1eee6; color: #777266; }
:global(body.scheme3-admin-context .scheme3-monitor-table-body) { color: #6b695f; }
:global(body.scheme3-admin-context .scheme3-monitor-table-row) { border-bottom: 1px solid #ebe7dc; transition: background-color 120ms ease; }
:global(body.scheme3-admin-context .scheme3-monitor-table-row:last-child) { border-bottom: 0; }
:global(body.scheme3-admin-context .scheme3-monitor-table-row:hover) { background: #f1eee6; }
:global(body.scheme3-admin-context .scheme3-monitor-table-primary),:global(body.scheme3-admin-context .scheme3-monitor-model-name) { color: #27251f; }
:global(body.scheme3-admin-context .scheme3-monitor-table-muted) { color: #777266; }
:global(body.scheme3-admin-context .scheme3-monitor-selection-toolbar) { color: #777266; }
:global(body.scheme3-admin-context .scheme3-monitor-inline-action.is-muted) { color: #777266; }
:global(body.scheme3-admin-context .scheme3-monitor-selection-list) { border: 1px solid #dad5c8; border-radius: 7px; background: #fbfaf6; }
:global(body.scheme3-admin-context .scheme3-monitor-selection-row) { border-bottom: 1px solid #ebe7dc; transition: background-color 120ms ease; }
:global(body.scheme3-admin-context .scheme3-monitor-selection-row:last-child) { border-bottom: 0; }
:global(body.scheme3-admin-context .scheme3-monitor-selection-row:hover) { background: #f1eee6; }
:global(body.scheme3-admin-context .scheme3-monitor-checkbox) { accent-color: #1e5c42; }
:global(body.scheme3-admin-context .scheme3-monitor-disabled-badge) { border: 1px solid #dad5c8; border-radius: 4px; background: #f1eee6; color: #777266; }
:global(body.scheme3-admin-context .scheme3-monitor-result-row) { border: 1px solid #dad5c8; border-radius: 7px; background: #fbfaf6; }
:global(body.scheme3-admin-context .scheme3-monitor-status-badge) { border: 1px solid currentColor; border-radius: 999px; font-weight: 800; }
:global(body.scheme3-admin-context .scheme3-monitor-status-badge.is-operational) { border-color: rgba(30,92,66,.3); background: rgba(30,92,66,.09); color: #1e5c42; }
:global(body.scheme3-admin-context .scheme3-monitor-status-badge.is-degraded) { border-color: rgba(183,121,31,.35); background: rgba(183,121,31,.1); color: #8b5d14; }
:global(body.scheme3-admin-context .scheme3-monitor-status-badge.is-failed),:global(body.scheme3-admin-context .scheme3-monitor-status-badge.is-error) { border-color: rgba(158,77,61,.35); background: rgba(158,77,61,.08); color: #9e4d3d; }
:global(body.scheme3-admin-context .scheme3-monitor-status-badge.is-unknown) { border-color: #dad5c8; background: #f1eee6; color: #777266; }
:global(body.scheme3-admin-context .scheme3-monitor-dialog-message) { color: #6b695f; font-size: .76rem; line-height: 1.55; }
:global(body.scheme3-admin-context .scheme3-monitor-tooltip-content) { color: #f4f2ec; }
:global(body.scheme3-admin-context .scheme3-monitor-tooltip-title),:global(body.scheme3-admin-context .scheme3-monitor-tooltip-primary) { color: #f4f2ec; }
:global(body.scheme3-admin-context .scheme3-monitor-tooltip-muted) { color: #c7c2b6; }
:global(body.scheme3-admin-context .scheme3-monitor-tooltip) { border: 1px solid #1e5c42 !important; border-radius: 7px !important; background: #27251f !important; color: #f4f2ec !important; box-shadow: 0 18px 40px rgba(54,48,34,.24) !important; }
:global(body.scheme3-admin-context .scheme3-tooltip-trigger) { color: #9e4d3d; }
:global(body.scheme3-admin-context .scheme3-tooltip-trigger:hover) { color: #b65f45; }
:global(body.scheme3-admin-context .scheme3-monitor-tooltip-arrow) { position: absolute; bottom: -.25rem; left: 50%; width: .5rem; height: .5rem; transform: translateX(-50%) rotate(45deg); border-right: 1px solid #1e5c42; border-bottom: 1px solid #1e5c42; background: #27251f; }
:global(body.scheme3-admin-context .scheme3-tooltip-close) { position: absolute; top: .35rem; right: .35rem; border: 1px solid transparent; border-radius: 5px; padding: .22rem; color: #aaa69a; }
:global(body.scheme3-admin-context .scheme3-tooltip-close:hover) { border-color: #47443a; background: #2b2924; color: #f4f2ec; }
:global(body.scheme3-admin-context .scheme3-monitor-choice) { display: inline-flex; min-height: 2.35rem; align-items: center; justify-content: center; gap: .45rem; border: 1px solid #dad5c8; border-radius: 6px; background: #fbfaf6; color: #6b695f; padding: .5rem .72rem; font-size: .72rem; font-weight: 700; transition: border-color 150ms ease,background-color 150ms ease,color 150ms ease; }
:global(body.scheme3-admin-context .scheme3-monitor-choice:hover) { border-color: rgba(30,92,66,.55); color: #1e5c42; }
:global(body.scheme3-admin-context .scheme3-monitor-choice.is-active) { border-color: #1e5c42; background: rgba(30,92,66,.1); color: #1e5c42; box-shadow: inset 0 0 0 1px rgba(30,92,66,.15); }
:global(body.scheme3-admin-context .scheme3-monitor-notice) { border: 1px solid rgba(30,92,66,.25); border-radius: 7px; background: rgba(30,92,66,.07); color: #1e5c42; padding: .75rem; }
:global(body.scheme3-admin-context .scheme3-monitor-details) { border: 1px solid #dad5c8; border-radius: 7px; background: #f1eee6; padding: .75rem; }
:global(body.scheme3-admin-context .scheme3-monitor-details summary) { color: #27251f; }
:global(body.scheme3-admin-context .scheme3-monitor-panel) { border: 1px solid #dad5c8; border-radius: 7px; background: #fbfaf6; }
:global(body.scheme3-admin-context .scheme3-monitor-badge) { display: inline-flex; align-items: center; border: 1px solid rgba(30,92,66,.25); border-radius: 999px; background: rgba(30,92,66,.08); color: #1e5c42; padding: .18rem .45rem; font-size: .58rem; font-weight: 800; }
:global(body.scheme3-admin-context .scheme3-monitor-model-tags > div > div:first-child) { border-color: #dad5c8 !important; border-radius: 6px; background: #fbfaf6 !important; }
:global(body.scheme3-admin-context .scheme3-monitor-model-tags > div > div:first-child > span) { border: 1px solid rgba(30,92,66,.25); border-radius: 5px; background: rgba(30,92,66,.08); color: #1e5c42; }
:global(body.scheme3-admin-context .scheme3-monitor-model-tags input) { color: #16150f; }
:global(body.scheme3-admin-context .scheme3-monitor-dialog .select-trigger) { border-color: #dad5c8; border-radius: 6px; background: #fbfaf6; color: #16150f; box-shadow: none; }
:global(body.scheme3-admin-context .scheme3-monitor-dialog .select-trigger:hover),:global(body.scheme3-admin-context .scheme3-monitor-dialog .select-trigger-open) { border-color: #1e5c42; box-shadow: 0 0 0 2px rgba(30,92,66,.12); }
:global(body.scheme3-admin-context .scheme3-monitor-inline-action) { border: 0; background: transparent; color: #1e5c42; padding: 0; font-weight: 800; }
:global(body.scheme3-admin-context .scheme3-monitor-inline-action:hover) { text-decoration: underline; }
:global(body.scheme3-admin-context .scheme3-monitor-button-dashed) { border-style: dashed; }
:global(body.scheme3-admin-context .scheme3-monitor-icon-button) { display: inline-flex; align-items: center; justify-content: center; border: 1px solid transparent; border-radius: 5px; padding: .25rem; background: transparent; color: #6b695f; }
:global(body.scheme3-admin-context .scheme3-monitor-icon-button:hover) { border-color: rgba(158,77,61,.35); background: rgba(158,77,61,.08); color: #9e4d3d; }
:global(body.scheme3-admin-context .scheme3-monitor-action-list) { display: flex; align-items: center; gap: .2rem; }
:global(body.scheme3-admin-context .scheme3-monitor-action-button) { display: inline-flex; min-width: 2.4rem; flex-direction: column; align-items: center; gap: .15rem; border: 1px solid transparent; border-radius: 5px; padding: .28rem; background: transparent; color: #6b695f; font-size: .56rem; font-weight: 800; transition: border-color 120ms ease,background-color 120ms ease,color 120ms ease; }
:global(body.scheme3-admin-context .scheme3-monitor-action-button:hover:not(:disabled)) { border-color: rgba(30,92,66,.24); background: rgba(30,92,66,.08); color: #1e5c42; }
:global(body.scheme3-admin-context .scheme3-monitor-action-button.is-danger:hover:not(:disabled)) { border-color: rgba(158,77,61,.3); background: rgba(158,77,61,.08); color: #9e4d3d; }
:global(body.scheme3-admin-context .scheme3-monitor-action-button:disabled) { cursor: not-allowed; opacity: .45; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-label) { color: #aaa69a; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-input) { border-color: #47443a; background: #24231f; color: #f4f2ec; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-input:focus) { border-color: #8fc2a5; box-shadow: 0 0 0 3px rgba(143,194,165,.13); }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-button) { border-color: #47443a; background: #24231f; color: #aaa69a; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-button:hover:not(:disabled)) { border-color: #8fc2a5; color: #8fc2a5; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-button.is-primary) { border-color: #8fc2a5; background: #8fc2a5; color: #1b1b18; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-button.is-primary:hover:not(:disabled)) { background: #a7d0b7; color: #1b1b18; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-button.is-danger) { border-color: rgba(211,139,121,.35); color: #d38b79; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-copy) { color: #aaa69a; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-empty) { color: #aaa69a; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-hint) { color: #827e72; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-error),:global(html.dark body.scheme3-admin-context .scheme3-monitor-required) { color: #d38b79; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-summary) { color: #aaa69a; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-input-icon) { color: #827e72; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-warning) { color: #d38b79; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-provider-badge.is-openai) { border-color: rgba(143,194,165,.3); background: rgba(143,194,165,.1); color: #8fc2a5; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-provider-badge.is-anthropic) { border-color: rgba(211,139,121,.32); background: rgba(211,139,121,.1); color: #d38b79; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-provider-badge.is-gemini) { border-color: rgba(211,164,92,.32); background: rgba(211,164,92,.1); color: #d3a45c; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-provider-badge.is-grok),:global(html.dark body.scheme3-admin-context .scheme3-monitor-provider-badge.is-unknown) { border-color: #47443a; background: #2b2924; color: #aaa69a; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-table-wrap),:global(html.dark body.scheme3-admin-context .scheme3-monitor-selection-list) { border-color: #47443a; background: #24231f; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-table-head) { border-color: #47443a; background: #2b2924; color: #aaa69a; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-table-body) { color: #aaa69a; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-table-row),:global(html.dark body.scheme3-admin-context .scheme3-monitor-selection-row) { border-color: #3a3830; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-table-row:hover),:global(html.dark body.scheme3-admin-context .scheme3-monitor-selection-row:hover) { background: #2b2924; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-table-primary),:global(html.dark body.scheme3-admin-context .scheme3-monitor-model-name) { color: #f4f2ec; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-table-muted),:global(html.dark body.scheme3-admin-context .scheme3-monitor-selection-toolbar),:global(html.dark body.scheme3-admin-context .scheme3-monitor-inline-action.is-muted) { color: #aaa69a; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-selection-row) { color: #f4f2ec; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-checkbox) { accent-color: #8fc2a5; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-disabled-badge) { border-color: #47443a; background: #2b2924; color: #aaa69a; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-result-row) { border-color: #47443a; background: #24231f; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-status-badge.is-operational) { border-color: rgba(143,194,165,.3); background: rgba(143,194,165,.1); color: #8fc2a5; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-status-badge.is-degraded) { border-color: rgba(211,164,92,.35); background: rgba(211,164,92,.1); color: #d3a45c; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-status-badge.is-failed),:global(html.dark body.scheme3-admin-context .scheme3-monitor-status-badge.is-error) { border-color: rgba(211,139,121,.35); background: rgba(211,139,121,.1); color: #d38b79; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-status-badge.is-unknown) { border-color: #47443a; background: #2b2924; color: #aaa69a; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-dialog-message) { color: #aaa69a; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-tooltip) { border-color: #47443a !important; background: #24231f !important; box-shadow: 0 18px 40px rgba(0,0,0,.34) !important; }
:global(html.dark body.scheme3-admin-context .scheme3-tooltip-trigger) { color: #d38b79; }
:global(html.dark body.scheme3-admin-context .scheme3-tooltip-trigger:hover) { color: #e0a08e; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-tooltip-arrow) { border-color: #47443a; background: #24231f; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-choice) { border-color: #47443a; background: #24231f; color: #aaa69a; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-choice:hover) { border-color: rgba(143,194,165,.65); color: #8fc2a5; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-choice.is-active) { border-color: #8fc2a5; background: rgba(143,194,165,.12); color: #8fc2a5; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-notice) { border-color: rgba(143,194,165,.28); background: rgba(143,194,165,.1); color: #8fc2a5; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-details) { border-color: #47443a; background: #2b2924; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-details summary) { color: #f4f2ec; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-panel) { border-color: #47443a; background: #24231f; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-badge) { border-color: rgba(143,194,165,.28); background: rgba(143,194,165,.1); color: #8fc2a5; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-model-tags > div > div:first-child) { border-color: #47443a !important; background: #24231f !important; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-model-tags > div > div:first-child > span) { border-color: rgba(143,194,165,.28); background: rgba(143,194,165,.1); color: #8fc2a5; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-model-tags input) { color: #f4f2ec; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-dialog .select-trigger) { border-color: #47443a; background: #24231f; color: #f4f2ec; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-dialog .select-trigger:hover),:global(html.dark body.scheme3-admin-context .scheme3-monitor-dialog .select-trigger-open) { border-color: #8fc2a5; box-shadow: 0 0 0 2px rgba(143,194,165,.13); }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-inline-action) { color: #8fc2a5; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-icon-button) { color: #aaa69a; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-icon-button:hover) { border-color: rgba(211,139,121,.35); background: rgba(211,139,121,.1); color: #d38b79; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-action-button) { color: #aaa69a; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-action-button:hover:not(:disabled)) { border-color: rgba(143,194,165,.28); background: rgba(143,194,165,.1); color: #8fc2a5; }
:global(html.dark body.scheme3-admin-context .scheme3-monitor-action-button.is-danger:hover:not(:disabled)) { border-color: rgba(211,139,121,.35); background: rgba(211,139,121,.1); color: #d38b79; }

:global(body.scheme3-admin-context .scheme3-toast) { border: 1px solid #dad5c8 !important; border-left-width: 3px !important; border-radius: 7px !important; background: #fbfaf6 !important; color: #16150f; box-shadow: 0 16px 34px rgba(54,48,34,.16) !important; }
:global(html.dark body.scheme3-admin-context .scheme3-toast) { border-color: #47443a !important; background: #24231f !important; color: #f4f2ec; box-shadow: 0 18px 38px rgba(0,0,0,.3) !important; }
:global(body.scheme3-admin-context .scheme3-toast .text-gray-900),:global(body.scheme3-admin-context .scheme3-toast .text-gray-600) { color: #16150f !important; }
:global(body.scheme3-admin-context .scheme3-toast .text-gray-400),:global(body.scheme3-admin-context .scheme3-toast .text-gray-500) { color: #6b695f !important; }
:global(html.dark body.scheme3-admin-context .scheme3-toast .text-gray-900),:global(html.dark body.scheme3-admin-context .scheme3-toast .text-gray-600),:global(html.dark body.scheme3-admin-context .scheme3-toast .text-white) { color: #f4f2ec !important; }
:global(html.dark body.scheme3-admin-context .scheme3-toast .text-gray-400),:global(html.dark body.scheme3-admin-context .scheme3-toast .text-gray-500) { color: #aaa69a !important; }
:global(body.scheme3-admin-context .scheme3-toast .bg-gray-100),:global(body.scheme3-admin-context .scheme3-toast .hover\:bg-gray-100:hover) { background: #f1eee6 !important; }
:global(html.dark body.scheme3-admin-context .scheme3-toast .bg-gray-100),:global(html.dark body.scheme3-admin-context .scheme3-toast .bg-dark-700),:global(html.dark body.scheme3-admin-context .scheme3-toast .hover\:bg-dark-700:hover) { background: #2b2924 !important; }
:global(body.scheme3-admin-context .scheme3-toast.border-green-500),:global(body.scheme3-admin-context .scheme3-toast.border-blue-500) { border-left-color: #1e5c42 !important; }
:global(html.dark body.scheme3-admin-context .scheme3-toast.border-green-500),:global(html.dark body.scheme3-admin-context .scheme3-toast.border-blue-500) { border-left-color: #8fc2a5 !important; }

:global(body.scheme3-admin-context [class*="bg-blue-"]),:global(body.scheme3-admin-context [class*="bg-indigo-"]),:global(body.scheme3-admin-context [class*="bg-purple-"]),:global(body.scheme3-admin-context [class*="bg-violet-"]),:global(body.scheme3-admin-context [class*="bg-sky-"]) { background-color: #f1eee6 !important; }
:global(body.scheme3-admin-context [class*="text-blue-"]),:global(body.scheme3-admin-context [class*="text-indigo-"]),:global(body.scheme3-admin-context [class*="text-purple-"]),:global(body.scheme3-admin-context [class*="text-violet-"]),:global(body.scheme3-admin-context [class*="text-sky-"]) { color: #1e5c42 !important; }
:global(body.scheme3-admin-context [class*="border-blue-"]),:global(body.scheme3-admin-context [class*="border-indigo-"]),:global(body.scheme3-admin-context [class*="border-purple-"]),:global(body.scheme3-admin-context [class*="border-violet-"]),:global(body.scheme3-admin-context [class*="border-sky-"]) { border-color: rgba(30,92,66,.25) !important; }
:global(html.dark body.scheme3-admin-context [class*="dark:bg-blue-"]),:global(html.dark body.scheme3-admin-context [class*="dark:bg-indigo-"]),:global(html.dark body.scheme3-admin-context [class*="dark:bg-purple-"]),:global(html.dark body.scheme3-admin-context [class*="dark:bg-violet-"]),:global(html.dark body.scheme3-admin-context [class*="dark:bg-sky-"]) { background-color: rgba(143,194,165,.1) !important; }
:global(html.dark body.scheme3-admin-context [class*="dark:text-blue-"]),:global(html.dark body.scheme3-admin-context [class*="dark:text-indigo-"]),:global(html.dark body.scheme3-admin-context [class*="dark:text-purple-"]),:global(html.dark body.scheme3-admin-context [class*="dark:text-violet-"]),:global(html.dark body.scheme3-admin-context [class*="dark:text-sky-"]) { color: #8fc2a5 !important; }
:global(html.dark body.scheme3-admin-context [class*="dark:border-blue-"]),:global(html.dark body.scheme3-admin-context [class*="dark:border-indigo-"]),:global(html.dark body.scheme3-admin-context [class*="dark:border-purple-"]),:global(html.dark body.scheme3-admin-context [class*="dark:border-violet-"]),:global(html.dark body.scheme3-admin-context [class*="dark:border-sky-"]) { border-color: rgba(143,194,165,.25) !important; }

/* Some upstream templates only define a light utility (for example
   `text-blue-500`) and rely on the inherited dark surface.  In dark mode that
   leaves the original blue/purple accent visible because no `dark:text-*`
   class exists to override it.  Normalize both base and dark utility names
   here so every authenticated third-version route uses the same palette. */
:global(html.dark body.scheme3-user-context [class*="bg-blue-"]),
:global(html.dark body.scheme3-user-context [class*="bg-indigo-"]),
:global(html.dark body.scheme3-user-context [class*="bg-purple-"]),
:global(html.dark body.scheme3-user-context [class*="bg-violet-"]),
:global(html.dark body.scheme3-user-context [class*="bg-sky-"]),
:global(html.dark body.scheme3-user-context [class*="bg-cyan-"]),
:global(html.dark body.scheme3-user-context [class*="bg-teal-"]),
:global(html.dark body.scheme3-admin-context [class*="bg-blue-"]),
:global(html.dark body.scheme3-admin-context [class*="bg-indigo-"]),
:global(html.dark body.scheme3-admin-context [class*="bg-purple-"]),
:global(html.dark body.scheme3-admin-context [class*="bg-violet-"]),
:global(html.dark body.scheme3-admin-context [class*="bg-sky-"]),
:global(html.dark body.scheme3-admin-context [class*="bg-cyan-"]),
:global(html.dark body.scheme3-admin-context [class*="bg-teal-"]) {
  background-color: rgba(143,194,165,.1) !important;
  background-image: none !important;
}
:global(html.dark body.scheme3-user-context [class*="bg-blue-5"]),
:global(html.dark body.scheme3-user-context [class*="bg-blue-6"]),
:global(html.dark body.scheme3-user-context [class*="bg-blue-7"]),
:global(html.dark body.scheme3-user-context [class*="bg-indigo-5"]),
:global(html.dark body.scheme3-user-context [class*="bg-indigo-6"]),
:global(html.dark body.scheme3-user-context [class*="bg-purple-5"]),
:global(html.dark body.scheme3-user-context [class*="bg-purple-6"]),
:global(html.dark body.scheme3-user-context [class*="bg-violet-5"]),
:global(html.dark body.scheme3-user-context [class*="bg-violet-6"]),
:global(html.dark body.scheme3-user-context [class*="bg-cyan-5"]),
:global(html.dark body.scheme3-user-context [class*="bg-teal-5"]),
:global(html.dark body.scheme3-admin-context [class*="bg-blue-5"]),
:global(html.dark body.scheme3-admin-context [class*="bg-blue-6"]),
:global(html.dark body.scheme3-admin-context [class*="bg-blue-7"]),
:global(html.dark body.scheme3-admin-context [class*="bg-indigo-5"]),
:global(html.dark body.scheme3-admin-context [class*="bg-indigo-6"]),
:global(html.dark body.scheme3-admin-context [class*="bg-purple-5"]),
:global(html.dark body.scheme3-admin-context [class*="bg-purple-6"]),
:global(html.dark body.scheme3-admin-context [class*="bg-violet-5"]),
:global(html.dark body.scheme3-admin-context [class*="bg-violet-6"]),
:global(html.dark body.scheme3-admin-context [class*="bg-cyan-5"]),
:global(html.dark body.scheme3-admin-context [class*="bg-teal-5"]) {
  background-color: #8fc2a5 !important;
  background-image: none !important;
}
:global(html.dark body.scheme3-user-context [class*="text-blue-"]),
:global(html.dark body.scheme3-user-context [class*="text-indigo-"]),
:global(html.dark body.scheme3-user-context [class*="text-purple-"]),
:global(html.dark body.scheme3-user-context [class*="text-violet-"]),
:global(html.dark body.scheme3-user-context [class*="text-sky-"]),
:global(html.dark body.scheme3-user-context [class*="text-cyan-"]),
:global(html.dark body.scheme3-user-context [class*="text-teal-"]),
:global(html.dark body.scheme3-user-context [class*="dark:text-blue-"]),
:global(html.dark body.scheme3-user-context [class*="dark:text-indigo-"]),
:global(html.dark body.scheme3-user-context [class*="dark:text-purple-"]),
:global(html.dark body.scheme3-user-context [class*="dark:text-violet-"]),
:global(html.dark body.scheme3-user-context [class*="dark:text-sky-"]),
:global(html.dark body.scheme3-user-context [class*="dark:text-cyan-"]),
:global(html.dark body.scheme3-user-context [class*="dark:text-teal-"]),
:global(html.dark body.scheme3-admin-context [class*="text-blue-"]),
:global(html.dark body.scheme3-admin-context [class*="text-indigo-"]),
:global(html.dark body.scheme3-admin-context [class*="text-purple-"]),
:global(html.dark body.scheme3-admin-context [class*="text-violet-"]),
:global(html.dark body.scheme3-admin-context [class*="text-sky-"]),
:global(html.dark body.scheme3-admin-context [class*="text-cyan-"]),
:global(html.dark body.scheme3-admin-context [class*="text-teal-"]),
:global(html.dark body.scheme3-admin-context [class*="dark:text-blue-"]),
:global(html.dark body.scheme3-admin-context [class*="dark:text-indigo-"]),
:global(html.dark body.scheme3-admin-context [class*="dark:text-purple-"]),
:global(html.dark body.scheme3-admin-context [class*="dark:text-violet-"]),
:global(html.dark body.scheme3-admin-context [class*="dark:text-sky-"]),
:global(html.dark body.scheme3-admin-context [class*="dark:text-cyan-"]),
:global(html.dark body.scheme3-admin-context [class*="dark:text-teal-"]) {
  color: #8fc2a5 !important;
}
:global(html.dark body.scheme3-user-context [class*="border-blue-"]),
:global(html.dark body.scheme3-user-context [class*="border-indigo-"]),
:global(html.dark body.scheme3-user-context [class*="border-purple-"]),
:global(html.dark body.scheme3-user-context [class*="border-violet-"]),
:global(html.dark body.scheme3-user-context [class*="border-sky-"]),
:global(html.dark body.scheme3-user-context [class*="border-cyan-"]),
:global(html.dark body.scheme3-user-context [class*="border-teal-"]),
:global(html.dark body.scheme3-user-context [class*="dark:border-blue-"]),
:global(html.dark body.scheme3-user-context [class*="dark:border-indigo-"]),
:global(html.dark body.scheme3-user-context [class*="dark:border-purple-"]),
:global(html.dark body.scheme3-user-context [class*="dark:border-violet-"]),
:global(html.dark body.scheme3-user-context [class*="dark:border-sky-"]),
:global(html.dark body.scheme3-user-context [class*="dark:border-cyan-"]),
:global(html.dark body.scheme3-user-context [class*="dark:border-teal-"]),
:global(html.dark body.scheme3-admin-context [class*="border-blue-"]),
:global(html.dark body.scheme3-admin-context [class*="border-indigo-"]),
:global(html.dark body.scheme3-admin-context [class*="border-purple-"]),
:global(html.dark body.scheme3-admin-context [class*="border-violet-"]),
:global(html.dark body.scheme3-admin-context [class*="border-sky-"]),
:global(html.dark body.scheme3-admin-context [class*="border-cyan-"]),
:global(html.dark body.scheme3-admin-context [class*="border-teal-"]),
:global(html.dark body.scheme3-admin-context [class*="dark:border-blue-"]),
:global(html.dark body.scheme3-admin-context [class*="dark:border-indigo-"]),
:global(html.dark body.scheme3-admin-context [class*="dark:border-purple-"]),
:global(html.dark body.scheme3-admin-context [class*="dark:border-violet-"]),
:global(html.dark body.scheme3-admin-context [class*="dark:border-sky-"]),
:global(html.dark body.scheme3-admin-context [class*="dark:border-cyan-"]),
:global(html.dark body.scheme3-admin-context [class*="dark:border-teal-"]) {
  border-color: rgba(143,194,165,.28) !important;
}
:global(html.dark body.scheme3-user-context [class*="hover:bg-blue-"]:hover),
:global(html.dark body.scheme3-user-context [class*="hover:bg-indigo-"]:hover),
:global(html.dark body.scheme3-user-context [class*="hover:bg-purple-"]:hover),
:global(html.dark body.scheme3-user-context [class*="hover:bg-violet-"]:hover),
:global(html.dark body.scheme3-user-context [class*="hover:bg-cyan-"]:hover),
:global(html.dark body.scheme3-admin-context [class*="hover:bg-blue-"]:hover),
:global(html.dark body.scheme3-admin-context [class*="hover:bg-indigo-"]:hover),
:global(html.dark body.scheme3-admin-context [class*="hover:bg-purple-"]:hover),
:global(html.dark body.scheme3-admin-context [class*="hover:bg-violet-"]:hover),
:global(html.dark body.scheme3-admin-context [class*="hover:bg-cyan-"]:hover) {
  background-color: #2b2924 !important;
  background-image: none !important;
}
:global(html.dark body.scheme3-user-context [class*="focus:ring-blue-"]),
:global(html.dark body.scheme3-user-context [class*="focus:ring-indigo-"]),
:global(html.dark body.scheme3-user-context [class*="focus:ring-purple-"]),
:global(html.dark body.scheme3-user-context [class*="focus:ring-violet-"]),
:global(html.dark body.scheme3-admin-context [class*="focus:ring-blue-"]),
:global(html.dark body.scheme3-admin-context [class*="focus:ring-indigo-"]),
:global(html.dark body.scheme3-admin-context [class*="focus:ring-purple-"]),
:global(html.dark body.scheme3-admin-context [class*="focus:ring-violet-"]) {
  --tw-ring-color: rgba(143,194,165,.28) !important;
}

/* Navigation feedback keeps its motion, but uses the scheme3 accent instead
   of the upstream teal gradient. */
:global(body.scheme3-user-context .navigation-progress-bar),
:global(body.scheme3-admin-context .navigation-progress-bar) {
  background: #1e5c42 !important;
}
:global(html.dark body.scheme3-user-context .navigation-progress-bar),
:global(html.dark body.scheme3-admin-context .navigation-progress-bar) {
  background: #8fc2a5 !important;
}

/* Hide the upstream auto-tour on the monitor route and keep manual tours,
   when present elsewhere, visually consistent with the third-version shell. */
:global(body.scheme3-user-context .driver-popover),:global(body.scheme3-admin-context .driver-popover) { border: 1px solid #dad5c8 !important; border-radius: 8px !important; background: #fbfaf6 !important; color: #16150f !important; box-shadow: 0 24px 60px rgba(54,48,34,.2) !important; }
:global(body.scheme3-user-context .driver-popover-title),:global(body.scheme3-admin-context .driver-popover-title) { color: #16150f !important; font-family: Georgia,'Times New Roman',serif !important; }
:global(body.scheme3-user-context .driver-popover-description),:global(body.scheme3-admin-context .driver-popover-description) { color: #6b695f !important; }
:global(body.scheme3-user-context .driver-popover-footer button),:global(body.scheme3-admin-context .driver-popover-footer button) { border: 1px solid #dad5c8 !important; border-radius: 6px !important; background: #f1eee6 !important; color: #6b695f !important; text-shadow: none !important; }
:global(body.scheme3-user-context .driver-popover-footer .driver-popover-next-btn),:global(body.scheme3-admin-context .driver-popover-footer .driver-popover-next-btn) { border-color: #1e5c42 !important; background: #1e5c42 !important; color: #f4f2ec !important; }
:global(html.dark body.scheme3-user-context .driver-popover),:global(html.dark body.scheme3-admin-context .driver-popover) { border-color: #47443a !important; background: #24231f !important; color: #f4f2ec !important; box-shadow: 0 24px 60px rgba(0,0,0,.32) !important; }
:global(html.dark body.scheme3-user-context .driver-popover-title),:global(html.dark body.scheme3-admin-context .driver-popover-title) { color: #f4f2ec !important; }
:global(html.dark body.scheme3-user-context .driver-popover-description),:global(html.dark body.scheme3-admin-context .driver-popover-description) { color: #aaa69a !important; }
:global(html.dark body.scheme3-user-context .driver-popover-footer button),:global(html.dark body.scheme3-admin-context .driver-popover-footer button) { border-color: #47443a !important; background: #2b2924 !important; color: #aaa69a !important; }
:global(html.dark body.scheme3-user-context .driver-popover-footer .driver-popover-next-btn),:global(html.dark body.scheme3-admin-context .driver-popover-footer .driver-popover-next-btn) { border-color: #8fc2a5 !important; background: #8fc2a5 !important; color: #1b1b18 !important; }

/* EmptyState keeps its descriptive text in a semantic class instead of a
   Tailwind utility, so the page-frame utility remap cannot reach it. Apply
   the paper palette here for both in-shell pages and empty states rendered by
   shared children. */
:global(body.scheme3-user-context .empty-state > div:first-child),
:global(body.scheme3-admin-context .empty-state > div:first-child) {
  border: 1px solid #dad5c8;
  border-radius: 7px !important;
  background: #f1eee6 !important;
}
:global(body.scheme3-user-context .empty-state-icon),
:global(body.scheme3-admin-context .empty-state-icon) { color: #a49e90 !important; }
:global(body.scheme3-user-context .empty-state-title),
:global(body.scheme3-admin-context .empty-state-title) { color: #27251f !important; }
:global(body.scheme3-user-context .empty-state-description),
:global(body.scheme3-admin-context .empty-state-description) { color: #6b695f !important; }
:global(html.dark body.scheme3-user-context .empty-state > div:first-child),
:global(html.dark body.scheme3-admin-context .empty-state > div:first-child) {
  border-color: #47443a;
  background: #2b2924 !important;
}
:global(html.dark body.scheme3-user-context .empty-state-icon),
:global(html.dark body.scheme3-admin-context .empty-state-icon) { color: #827e72 !important; }
:global(html.dark body.scheme3-user-context .empty-state-title),
:global(html.dark body.scheme3-admin-context .empty-state-title) { color: #f4f2ec !important; }
:global(html.dark body.scheme3-user-context .empty-state-description),
:global(html.dark body.scheme3-admin-context .empty-state-description) { color: #aaa69a !important; }

/* BaseDialog teleports its panel to <body>. Keep legacy utility classes used
   by older admin forms on the same Scheme3 surface instead of relying on the
   page-frame descendant rules (which cannot cross the teleport boundary). */
:global(body.scheme3-user-context .modal-content),
:global(body.scheme3-admin-context .modal-content) {
  background-image: none !important;
}
:global(body.scheme3-user-context .modal-content .card),
:global(body.scheme3-admin-context .modal-content .card) {
  border-color: #dad5c8 !important;
  border-radius: 7px !important;
  background: #fbfaf6 !important;
  box-shadow: none !important;
}
:global(body.scheme3-user-context .modal-content [class~="rounded-xl"]),
:global(body.scheme3-user-context .modal-content [class~="rounded-2xl"]),
:global(body.scheme3-user-context .modal-content [class~="rounded-3xl"]),
:global(body.scheme3-admin-context .modal-content [class~="rounded-xl"]),
:global(body.scheme3-admin-context .modal-content [class~="rounded-2xl"]),
:global(body.scheme3-admin-context .modal-content [class~="rounded-3xl"]) { border-radius: 7px !important; }
:global(body.scheme3-user-context .modal-content [class*="bg-gradient-"]),
:global(body.scheme3-admin-context .modal-content [class*="bg-gradient-"]) { background-image: none !important; }
:global(body.scheme3-user-context .modal-content [class*="bg-gray-"]),
:global(body.scheme3-admin-context .modal-content [class*="bg-gray-"]) { background-color: #f1eee6 !important; }
:global(body.scheme3-user-context .modal-content [class*="text-gray-"]),
:global(body.scheme3-admin-context .modal-content [class*="text-gray-"]) { color: #6b695f !important; }
:global(body.scheme3-user-context .modal-content [class*="text-blue-"]),
:global(body.scheme3-user-context .modal-content [class*="text-indigo-"]),
:global(body.scheme3-user-context .modal-content [class*="text-purple-"]),
:global(body.scheme3-admin-context .modal-content [class*="text-blue-"]),
:global(body.scheme3-admin-context .modal-content [class*="text-indigo-"]),
:global(body.scheme3-admin-context .modal-content [class*="text-purple-"]) { color: #1e5c42 !important; }
:global(html.dark body.scheme3-user-context .modal-content),
:global(html.dark body.scheme3-admin-context .modal-content) { background-image: none !important; }
:global(html.dark body.scheme3-user-context .modal-content .card),
:global(html.dark body.scheme3-admin-context .modal-content .card) { border-color: #47443a !important; background: #24231f !important; }
:global(html.dark body.scheme3-user-context .modal-content [class*="bg-gray-"]),
:global(html.dark body.scheme3-admin-context .modal-content [class*="bg-gray-"]) { background-color: #2b2924 !important; }
:global(html.dark body.scheme3-user-context .modal-content [class*="text-gray-"]),
:global(html.dark body.scheme3-admin-context .modal-content [class*="text-gray-"]) { color: #aaa69a !important; }
:global(html.dark body.scheme3-user-context .modal-content [class*="text-blue-"]),
:global(html.dark body.scheme3-user-context .modal-content [class*="text-indigo-"]),
:global(html.dark body.scheme3-user-context .modal-content [class*="text-purple-"]),
:global(html.dark body.scheme3-admin-context .modal-content [class*="text-blue-"]),
:global(html.dark body.scheme3-admin-context .modal-content [class*="text-indigo-"]),
:global(html.dark body.scheme3-admin-context .modal-content [class*="text-purple-"]) { color: #8fc2a5 !important; }

/* A few upstream account controls still use slate/cyan/teal utilities and
   some teleported dialogs rely on dark-500/600 surfaces. Those selectors sit
   outside the page frame, so normalize them at the authenticated body scope. */
:global(body.scheme3-user-context [class*="bg-slate-"]),
:global(body.scheme3-admin-context [class*="bg-slate-"]) {
  background-color: #f1eee6 !important;
  background-image: none !important;
}
:global(body.scheme3-user-context [class*="bg-cyan-"]),
:global(body.scheme3-user-context [class*="bg-teal-"]),
:global(body.scheme3-admin-context [class*="bg-cyan-"]),
:global(body.scheme3-admin-context [class*="bg-teal-"]) {
  background-color: rgba(30,92,66,.1) !important;
  background-image: none !important;
}
:global(body.scheme3-user-context [class*="text-slate-"]),
:global(body.scheme3-admin-context [class*="text-slate-"]) { color: #6b695f !important; }
:global(body.scheme3-user-context [class*="text-cyan-"]),
:global(body.scheme3-user-context [class*="text-teal-"]),
:global(body.scheme3-admin-context [class*="text-cyan-"]),
:global(body.scheme3-admin-context [class*="text-teal-"]) { color: #1e5c42 !important; }
:global(body.scheme3-user-context [class*="border-slate-"]),
:global(body.scheme3-admin-context [class*="border-slate-"]) { border-color: #dad5c8 !important; }
:global(body.scheme3-user-context [class*="border-cyan-"]),
:global(body.scheme3-user-context [class*="border-teal-"]),
:global(body.scheme3-admin-context [class*="border-cyan-"]),
:global(body.scheme3-admin-context [class*="border-teal-"]) { border-color: rgba(30,92,66,.25) !important; }

:global(html.dark body.scheme3-user-context [class*="bg-slate-"]),
:global(html.dark body.scheme3-admin-context [class*="bg-slate-"]) {
  background-color: #2b2924 !important;
  background-image: none !important;
}
:global(html.dark body.scheme3-user-context [class*="bg-cyan-"]),
:global(html.dark body.scheme3-user-context [class*="bg-teal-"]),
:global(html.dark body.scheme3-admin-context [class*="bg-cyan-"]),
:global(html.dark body.scheme3-admin-context [class*="bg-teal-"]) {
  background-color: rgba(143,194,165,.1) !important;
  background-image: none !important;
}
:global(html.dark body.scheme3-user-context [class*="text-slate-"]),
:global(html.dark body.scheme3-admin-context [class*="text-slate-"]) { color: #aaa69a !important; }
:global(html.dark body.scheme3-user-context [class*="text-cyan-"]),
:global(html.dark body.scheme3-user-context [class*="text-teal-"]),
:global(html.dark body.scheme3-admin-context [class*="text-cyan-"]),
:global(html.dark body.scheme3-admin-context [class*="text-teal-"]) { color: #8fc2a5 !important; }
:global(html.dark body.scheme3-user-context [class*="border-slate-"]),
:global(html.dark body.scheme3-admin-context [class*="border-slate-"]) { border-color: #47443a !important; }
:global(html.dark body.scheme3-user-context [class*="border-cyan-"]),
:global(html.dark body.scheme3-user-context [class*="border-teal-"]),
:global(html.dark body.scheme3-admin-context [class*="border-cyan-"]),
:global(html.dark body.scheme3-admin-context [class*="border-teal-"]) { border-color: rgba(143,194,165,.28) !important; }
:global(html.dark body.scheme3-user-context [class*="dark:bg-dark-500"]),
:global(html.dark body.scheme3-user-context [class*="dark:bg-dark-600"]),
:global(html.dark body.scheme3-admin-context [class*="dark:bg-dark-500"]),
:global(html.dark body.scheme3-admin-context [class*="dark:bg-dark-600"]) { background-color: #2b2924 !important; }
:global(html.dark body.scheme3-user-context [class*="dark:text-dark-500"]),
:global(html.dark body.scheme3-user-context [class*="dark:text-dark-600"]),
:global(html.dark body.scheme3-admin-context [class*="dark:text-dark-500"]),
:global(html.dark body.scheme3-admin-context [class*="dark:text-dark-600"]) { color: #aaa69a !important; }
:global(html.dark body.scheme3-user-context [class*="dark:border-dark-500"]),
:global(html.dark body.scheme3-user-context [class*="dark:border-dark-600"]),
:global(html.dark body.scheme3-admin-context [class*="dark:border-dark-500"]),
:global(html.dark body.scheme3-admin-context [class*="dark:border-dark-600"]) { border-color: #47443a !important; }

/* Finish the dark utility remap for legacy components rendered inside the
   third-version shell. These classes are common in settings, groups, keys,
   and teleported dialogs, so keeping them on the original slate surfaces
   makes the old skin visible even after the page-frame remap. */
:global(html.dark body.scheme3-user-context [class*="dark:bg-dark-700"]),
:global(html.dark body.scheme3-admin-context [class*="dark:bg-dark-700"]),
:global(html.dark body.scheme3-user-context [class*="dark:bg-gray-700"]),
:global(html.dark body.scheme3-admin-context [class*="dark:bg-gray-700"]) {
  background-color: #2b2924 !important;
  background-image: none !important;
}
:global(html.dark body.scheme3-user-context [class*="dark:bg-dark-800"]),
:global(html.dark body.scheme3-admin-context [class*="dark:bg-dark-800"]),
:global(html.dark body.scheme3-user-context [class*="dark:bg-gray-800"]),
:global(html.dark body.scheme3-admin-context [class*="dark:bg-gray-800"]) {
  background-color: #24231f !important;
  background-image: none !important;
}
:global(html.dark body.scheme3-user-context [class*="dark:border-dark-700"]),
:global(html.dark body.scheme3-admin-context [class*="dark:border-dark-700"]),
:global(html.dark body.scheme3-user-context [class*="dark:border-gray-700"]),
:global(html.dark body.scheme3-admin-context [class*="dark:border-gray-700"]),
:global(html.dark body.scheme3-user-context [class*="dark:border-gray-800"]),
:global(html.dark body.scheme3-admin-context [class*="dark:border-gray-800"]) {
  border-color: #47443a !important;
}
:global(html.dark body.scheme3-user-context [class*="dark:text-gray-100"]),
:global(html.dark body.scheme3-admin-context [class*="dark:text-gray-100"]),
:global(html.dark body.scheme3-user-context [class*="dark:text-gray-200"]),
:global(html.dark body.scheme3-admin-context [class*="dark:text-gray-200"]) {
  color: #f4f2ec !important;
}
:global(html.dark body.scheme3-user-context [class*="dark:text-gray-300"]),
:global(html.dark body.scheme3-admin-context [class*="dark:text-gray-300"]),
:global(html.dark body.scheme3-user-context [class*="dark:text-gray-400"]),
:global(html.dark body.scheme3-admin-context [class*="dark:text-gray-400"]),
:global(html.dark body.scheme3-user-context [class*="dark:text-gray-500"]),
:global(html.dark body.scheme3-admin-context [class*="dark:text-gray-500"]) {
  color: #aaa69a !important;
}
:global(html.dark body.scheme3-user-context [class*="dark:hover:bg-dark-700"]:hover),
:global(html.dark body.scheme3-admin-context [class*="dark:hover:bg-dark-700"]:hover),
:global(html.dark body.scheme3-user-context [class*="dark:hover:bg-gray-700"]:hover),
:global(html.dark body.scheme3-admin-context [class*="dark:hover:bg-gray-700"]:hover) {
  background-color: #2b2924 !important;
  background-image: none !important;
}
:global(html.dark body.scheme3-user-context [class*="dark:hover:bg-dark-800"]:hover),
:global(html.dark body.scheme3-admin-context [class*="dark:hover:bg-dark-800"]:hover),
:global(html.dark body.scheme3-user-context [class*="dark:hover:bg-gray-800"]:hover),
:global(html.dark body.scheme3-admin-context [class*="dark:hover:bg-gray-800"]:hover) {
  background-color: #24231f !important;
  background-image: none !important;
}
:global(html.dark body.scheme3-user-context [class*="dark:hover:text-gray-300"]:hover),
:global(html.dark body.scheme3-admin-context [class*="dark:hover:text-gray-300"]:hover),
:global(html.dark body.scheme3-user-context [class*="dark:hover:text-gray-400"]:hover),
:global(html.dark body.scheme3-admin-context [class*="dark:hover:text-gray-400"]:hover) {
  color: #f4f2ec !important;
}

.scheme3-console-layout-collapsed .scheme3-console-brand-copy,.scheme3-console-layout-collapsed .scheme3-console-version-control,.scheme3-console-layout-collapsed .scheme3-console-caption,.scheme3-console-layout-collapsed .scheme3-console-section-label,.scheme3-console-layout-collapsed .scheme3-console-link-text,.scheme3-console-layout-collapsed .scheme3-console-current,.scheme3-console-layout-collapsed .scheme3-console-account-copy,.scheme3-console-layout-collapsed .scheme3-console-action span { display: none; }
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
  .scheme3-console-status { display: none; }
  .scheme3-console-topbar-right .scheme3-console-locale-tool { flex-shrink: 0; }
  .scheme3-console-version-control :deep(.scheme3-version-dropdown) { left: 0; right: auto; width: min(16rem, calc(100vw - 1.3rem)); max-width: calc(100vw - 1.3rem); }
}
@media (max-width: 767px) {
  .scheme3-console-doc-link,
  .scheme3-console-topbar-right .scheme3-console-subscription,
  .scheme3-console-balance { display: none; }
}
@media (max-width: 640px) {
  .scheme3-console-content { padding: .7rem .65rem 1rem; }
  .scheme3-console-topbar { min-height: 3.85rem; padding: .65rem; }
  .scheme3-console-topbar-kicker { font-size: .49rem; }
  .scheme3-console-topbar-left strong { font-size: 1rem; }
  .scheme3-console-topbar-left small { display: none; }
  .scheme3-console-topbar-right { gap: .45rem; }
  .scheme3-console-topbar-right .scheme3-console-locale-tool { display: none; }
  .scheme3-console-user-copy { display: none; }
  .scheme3-console-account-popover { right: -.2rem; }
  :global(body.scheme3-user-context .scheme3-toast) { min-width: 0 !important; width: calc(100vw - 2rem); }
}
</style>

import { createPinia, setActivePinia } from 'pinia'
import { mount, type VueWrapper } from '@vue/test-utils'
import { createMemoryHistory, createRouter, type Router } from 'vue-router'
import { defineComponent, nextTick, toRaw } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { useAdminSettingsStore, useAppStore, useAuthStore } from '@/stores'
import Scheme3ConsoleLayout from '../Scheme3ConsoleLayout.vue'

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

const batchImageAccess = vi.hoisted(() => ({
  value: false,
  refresh: vi.fn(async () => false),
}))

vi.mock('@/composables/useBatchImageAccess', () => ({
  useBatchImageAccess: () => ({
    canUseBatchImage: {
      get value() {
        return batchImageAccess.value
      },
    },
    refreshBatchImageAccess: batchImageAccess.refresh,
  }),
}))

const EmptyRoute = defineComponent({ template: '<div />' })
const routes = [
  '/dashboard',
  '/scheme3-dashboard',
  '/keys',
  '/model-square',
  '/model-plaza',
  '/canvas',
  '/leaderboard',
  '/batch-image',
  '/usage',
  '/available-channels',
  '/monitor',
  '/subscriptions',
  '/purchase',
  '/orders',
  '/redeem',
  '/affiliate',
  '/profile',
  '/admin/dashboard',
  '/admin/ops',
  '/admin/users',
  '/admin/groups',
  '/admin/channels/pricing',
  '/admin/channels/monitor',
  '/admin/subscriptions',
  '/admin/accounts',
  '/admin/announcements',
  '/admin/proxies',
  '/admin/risk-control',
  '/admin/prompt-audit',
  '/admin/redeem',
  '/admin/promo-codes',
  '/admin/affiliates/invites',
  '/admin/affiliates/rebates',
  '/admin/affiliates/transfers',
  '/admin/orders/dashboard',
  '/admin/orders',
  '/admin/orders/plans',
  '/admin/usage',
  '/admin/leaderboard',
  '/admin/audit-logs',
  '/admin/settings',
].map((path) => ({ path, component: EmptyRoute }))

routes.push({ path: '/custom/:id', component: EmptyRoute, name: 'CustomPage' })

type PublicFlags = {
  backend_mode_enabled?: boolean
  channel_monitor_enabled?: boolean
  available_channels_enabled?: boolean
  model_plaza_enabled?: boolean
  payment_enabled?: boolean
  affiliate_enabled?: boolean
  risk_control_enabled?: boolean
}

function makeRouter(): Router {
  return createRouter({ history: createMemoryHistory(), routes })
}

async function mountLayout(options: {
  admin?: boolean
  simple?: boolean
  route?: string
  flags?: PublicFlags
  adminPayment?: boolean
  opsMonitoring?: boolean
  batchImage?: boolean
} = {}): Promise<{ wrapper: VueWrapper; router: Router }> {
  const authStore = useAuthStore()
  authStore.token = 'test-token'
  authStore.user = {
    id: options.admin ? 1 : 2,
    role: options.admin ? 'admin' : 'user',
    username: options.admin ? 'admin' : 'user',
    email: options.admin ? 'admin@example.com' : 'user@example.com',
    balance: 10,
  } as never
  // `runMode` is intentionally readonly in the public auth-store API. The
  // setup-store target retains the writable ref for deterministic fixtures.
  const rawAuthStore = toRaw(authStore) as unknown as { runMode?: { value: 'standard' | 'simple' } }
  if (rawAuthStore.runMode) {
    const writableRunMode = toRaw(rawAuthStore.runMode)
    writableRunMode.value = options.simple ? 'simple' : 'standard'
  }

  const appStore = useAppStore()
  appStore.siteName = 'Sub2API'
  appStore.cachedPublicSettings = {
    backend_mode_enabled: false,
    channel_monitor_enabled: true,
    available_channels_enabled: true,
    model_plaza_enabled: true,
    payment_enabled: true,
    affiliate_enabled: true,
    risk_control_enabled: true,
    custom_menu_items: [
      { id: 'user-entry', label: '用户扩展', visibility: 'user', sort_order: 10, icon_svg: '' },
      { id: 'admin-public-entry', label: '公开管理扩展', visibility: 'admin', sort_order: 20, icon_svg: '' },
    ],
    ...options.flags,
  } as never

  const adminSettingsStore = useAdminSettingsStore()
  adminSettingsStore.loaded = true
  adminSettingsStore.paymentEnabled = options.adminPayment ?? true
  adminSettingsStore.opsMonitoringEnabled = options.opsMonitoring ?? true
  adminSettingsStore.customMenuItems = [
    { id: 'admin-entry', label: '管理员扩展', visibility: 'admin', sort_order: 10, icon_svg: '' },
    { id: 'admin-user-entry', label: '管理侧用户扩展', visibility: 'user', sort_order: 20, icon_svg: '' },
  ]

  batchImageAccess.value = options.batchImage ?? true
  batchImageAccess.refresh.mockResolvedValue(batchImageAccess.value)

  const router = makeRouter()
  await router.push(options.route ?? (options.admin ? '/admin/dashboard' : '/dashboard'))
  await router.isReady()

  const wrapper = mount(Scheme3ConsoleLayout, {
    props: { adminMode: options.admin ?? false },
    slots: { default: '<div data-test="page-content" />' },
    global: {
      plugins: [router],
      stubs: {
        AnnouncementBell: true,
        LocaleSwitcher: true,
        SubscriptionProgressMini: true,
        Icon: { template: '<i />' },
      },
    },
  })
  await nextTick()
  return { wrapper, router }
}

function sectionHrefs(wrapper: VueWrapper, index = 0): string[] {
  return wrapper.findAll('.scheme3-console-section')[index]
    ?.findAll('a.scheme3-console-link')
    .map((link) => link.attributes('href')) ?? []
}

async function expandAllGroups(wrapper: VueWrapper): Promise<void> {
  for (const group of wrapper.findAll('.scheme3-console-group')) {
    if (!group.classes().includes('scheme3-console-link-active')) await group.trigger('click')
  }
  await nextTick()
}

describe('Scheme3ConsoleLayout navigation contract', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    batchImageAccess.value = false
    batchImageAccess.refresh.mockClear()
  })

  afterEach(() => {
    document.body.classList.remove('scheme3-user-context', 'scheme3-admin-context')
  })

  it('shows the complete upstream user menu in the same order when features are enabled', async () => {
    const { wrapper } = await mountLayout()

    expect(sectionHrefs(wrapper)).toEqual([
      '/dashboard',
      '/keys',
      '/model-square',
      '/canvas',
      '/leaderboard',
      '/batch-image',
      '/usage',
      '/available-channels',
      '/monitor',
      '/subscriptions',
      '/purchase',
      '/orders',
      '/redeem',
      '/affiliate',
      '/profile',
      '/custom/user-entry',
    ])
    expect(wrapper.find('.scheme3-console-topbar-right a[href="/model-plaza?embedded=1"]').exists()).toBe(true)
    expect(wrapper.find('.scheme3-console-topbar-right a[href="/model-plaza?embedded=1"] span').text()).not.toBe('模型行情')
    expect(wrapper.text()).toContain('总排行榜')
    expect(wrapper.text()).not.toContain('后台总排行榜')
    expect(wrapper.text()).not.toContain('公开管理扩展')
    wrapper.unmount()
  })

  it('hides every user entry whose public feature switch is disabled', async () => {
    const { wrapper } = await mountLayout({
      batchImage: false,
      flags: {
        channel_monitor_enabled: false,
        available_channels_enabled: false,
        model_plaza_enabled: false,
        payment_enabled: false,
        affiliate_enabled: false,
      },
    })

    expect(sectionHrefs(wrapper)).toEqual([
      '/dashboard',
      '/keys',
      '/model-square',
      '/canvas',
      '/leaderboard',
      '/usage',
      '/subscriptions',
      '/redeem',
      '/profile',
      '/custom/user-entry',
    ])
    expect(wrapper.find('.scheme3-console-topbar-right a[href^="/model-plaza"]').exists()).toBe(false)
    wrapper.unmount()
  })

  it('uses the upstream simple-mode subset for regular users', async () => {
    const { wrapper } = await mountLayout({ simple: true })

    expect(sectionHrefs(wrapper)).toEqual([
      '/dashboard',
      '/keys',
      '/model-square',
      '/canvas',
      '/leaderboard',
      '/monitor',
      '/profile',
      '/custom/user-entry',
    ])
    wrapper.unmount()
  })

  it('does not render a user sidebar in backend mode', async () => {
    const { wrapper } = await mountLayout({ flags: { backend_mode_enabled: true } })

    expect(wrapper.findAll('.scheme3-console-section')).toHaveLength(0)
    wrapper.unmount()
  })

  it('shows both the admin menu and the admin account user menu in standard mode', async () => {
    const { wrapper } = await mountLayout({ admin: true })
    await expandAllGroups(wrapper)

    expect(sectionHrefs(wrapper, 0)).toEqual([
      '/admin/dashboard',
      '/model-square',
      '/admin/ops',
      '/admin/users',
      '/admin/groups',
      '/admin/channels/pricing',
      '/admin/channels/monitor',
      '/admin/subscriptions',
      '/admin/accounts',
      '/admin/announcements',
      '/admin/proxies',
      '/admin/risk-control',
      '/admin/prompt-audit',
      '/admin/redeem',
      '/admin/promo-codes',
      '/admin/affiliates/invites',
      '/admin/affiliates/rebates',
      '/admin/affiliates/transfers',
      '/admin/orders/dashboard',
      '/admin/orders',
      '/admin/orders/plans',
      '/admin/usage',
      '/admin/leaderboard',
      '/admin/audit-logs',
      '/admin/settings',
      '/custom/admin-entry',
    ])
    expect(sectionHrefs(wrapper, 1)).toEqual([
      '/keys',
      '/model-square',
      '/canvas',
      '/leaderboard',
      '/batch-image',
      '/usage',
      '/available-channels',
      '/monitor',
      '/subscriptions',
      '/purchase',
      '/orders',
      '/redeem',
      '/affiliate',
      '/profile',
      '/custom/user-entry',
    ])
    expect(wrapper.find('#sidebar-channel-manage').attributes('href')).toBe('/admin/accounts')
    expect(wrapper.find('#sidebar-group-manage').attributes('href')).toBe('/admin/groups')
    expect(wrapper.find('#sidebar-wallet').attributes('href')).toBe('/admin/redeem')
    expect(wrapper.findAll('.scheme3-console-section')[1].find('.scheme3-console-section-label').exists()).toBe(true)
    expect(wrapper.findAll('button').some((button) => button.text() === 'onboarding.restartTour')).toBe(true)
    wrapper.unmount()
  })

  it('applies independent admin and public switches to an admin account', async () => {
    const { wrapper } = await mountLayout({
      admin: true,
      adminPayment: false,
      opsMonitoring: false,
      batchImage: false,
      flags: {
        channel_monitor_enabled: false,
        available_channels_enabled: false,
        model_plaza_enabled: false,
        payment_enabled: false,
        affiliate_enabled: false,
        risk_control_enabled: false,
      },
    })
    await expandAllGroups(wrapper)

    expect(sectionHrefs(wrapper, 0)).toEqual([
      '/admin/dashboard',
      '/model-square',
      '/admin/users',
      '/admin/groups',
      '/admin/channels/pricing',
      '/admin/subscriptions',
      '/admin/accounts',
      '/admin/announcements',
      '/admin/proxies',
      '/admin/redeem',
      '/admin/promo-codes',
      '/admin/usage',
      '/admin/leaderboard',
      '/admin/audit-logs',
      '/admin/settings',
      '/custom/admin-entry',
    ])
    expect(sectionHrefs(wrapper, 1)).toEqual([
      '/keys',
      '/model-square',
      '/canvas',
      '/leaderboard',
      '/usage',
      '/subscriptions',
      '/redeem',
      '/profile',
      '/custom/user-entry',
    ])
    wrapper.unmount()
  })

  it('uses the upstream admin simple-mode order and omits the personal section', async () => {
    const { wrapper } = await mountLayout({ admin: true, simple: true })
    await expandAllGroups(wrapper)

    expect(wrapper.findAll('.scheme3-console-section')).toHaveLength(1)
    expect(sectionHrefs(wrapper)).toEqual([
      '/admin/dashboard',
      '/model-square',
      '/admin/ops',
      '/admin/accounts',
      '/admin/announcements',
      '/admin/proxies',
      '/admin/risk-control',
      '/admin/prompt-audit',
      '/admin/usage',
      '/admin/leaderboard',
      '/keys',
      '/admin/settings',
      '/custom/admin-entry',
    ])
    expect(wrapper.text()).not.toContain('我的账户')
    expect(wrapper.text()).not.toContain('用户扩展')
    expect(wrapper.findAll('button').some((button) => button.text() === 'onboarding.restartTour')).toBe(false)
    wrapper.unmount()
  })

  it('marks only the exact order child active while keeping its parent expanded', async () => {
    const { wrapper } = await mountLayout({ admin: true, route: '/admin/orders/plans' })

    const orderGroup = wrapper.findAll('.scheme3-console-group').at(3)
    expect(orderGroup.classes()).not.toContain('scheme3-console-link-active')
    expect(orderGroup.classes()).toContain('scheme3-console-group')

    const activeOrderLinks = wrapper.findAll('.scheme3-console-subnav-link.scheme3-console-link-active')
    expect(activeOrderLinks.map((link) => link.attributes('href'))).toEqual(['/admin/orders/plans'])
    wrapper.unmount()
  })

  it('shows the configured custom-menu label as the page title', async () => {
    const { wrapper } = await mountLayout({ route: '/custom/user-entry' })

    expect(wrapper.find('.scheme3-console-topbar-left strong').text()).toBe('用户扩展')
    wrapper.unmount()
  })

  it('treats the Scheme 3 dashboard compatibility route as the dashboard tab', async () => {
    const { wrapper } = await mountLayout({ route: '/scheme3-dashboard' })

    expect(wrapper.find('a.scheme3-console-link[href="/dashboard"]').classes()).toContain('scheme3-console-link-active')
    wrapper.unmount()
  })
})

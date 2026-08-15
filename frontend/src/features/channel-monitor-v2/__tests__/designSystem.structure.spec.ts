/**
 * Structure contracts for the third-version channel monitor surfaces.
 * These checks intentionally reject the old Sub chrome so a future upstream
 * merge cannot silently reintroduce the original card/page-header skin.
 */
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const root = resolve(__dirname, '../../..')

function read(rel: string) {
  return readFileSync(resolve(root, rel), 'utf8')
}

function expectNoLegacyChrome(src: string) {
  expect(src).not.toContain('page-header')
  expect(src).not.toContain('page-title')
  expect(src).not.toContain('rounded-3xl')
  expect(src).not.toContain('ring-1 ring-gray-900/5')
  expect(src).not.toContain('dark:bg-dark-800')
  expect(src).not.toContain('dark:!bg-dark-800')
  expect(src).not.toContain('class="card')
  expect(src).not.toContain('btn btn-primary')
  expect(src).not.toContain('stat-card')
}

describe('channel-monitor-v2 third-version design structure', () => {
  it('user shell uses ledger controls, KPI strip, and internal table scroll', () => {
    const src = read('views/user/ChannelStatusV2View.vue')
    expect(src).toContain('scheme3-channel-status-v2')
    expect(src).toContain('scheme3-v2-control-sheet')
    expect(src).toContain('scheme3-v2-ledger')
    expect(src).toContain('scheme3-v2-toolbar')
    expect(src).toContain('scheme3-v2-kpi-grid')
    expect(src).toContain('scheme3-v2-data-panel')
    expect(src).toContain('scheme3-v2-segmented')
    expect(src).toContain('scheme3-v2-status-dot')
    expect(src).toContain('monitor-toolbar')
    expect(src).toContain('clearDimensions')
    expect(src).toContain('healthModeOptions')
    expect(src).toContain("'cache'")
    expect(src.indexOf('summaryAria')).toBeLessThan(src.indexOf('MonitorTrendChart'))
    expect(src).not.toMatch(/min-width:\s*980px/)
    expect(src).not.toMatch(/min-w-\[980px\]/)
    expect(src).toMatch(/max-h-\[min\(52vh/)
    expect(src).toContain('overflow-auto')
    expect(src).toContain('trendView')
    expect(src).toContain("'platform_group'")
    expect(src).toContain('MonitorTrendChart')
    expectNoLegacyChrome(src)
  })

  it('pulse matrix uses third-version panel chrome and hover tooltips', () => {
    const src = read('features/channel-monitor-v2/RelayPulseMatrix.vue')
    expect(src).toContain('scheme3-v2-panel')
    expect(src).toContain('scheme3-v2-panel-header')
    expect(src).toContain('scheme3-v2-matrix-scroll')
    expect(src).toContain('matrix-scroll')
    expect(src).toMatch(/max-h-\[min\(42vh/)
    expect(src).toContain('overflow-auto')
    expect(src).toContain('pulse-tooltip')
    expect(src).toContain('scheme3-v2-empty')
    expect(src).not.toContain('@/components/common/EmptyState.vue')
    expect(src).toContain('scheme3-v2-status-dot')
    expect(src).not.toContain('modal-overlay')
    expect(src).not.toContain('modal-content')
    expectNoLegacyChrome(src)
  })

  it('metric cells use third-version metric tokens', () => {
    const src = read('features/channel-monitor-v2/MetricCell.vue')
    expect(src).toContain('scheme3-v2-metric-cell')
    expect(src).toContain('scheme3-v2-metric-label')
    expect(src).toContain('scheme3-v2-metric-value')
    expect(src).toContain('is-healthy')
    expectNoLegacyChrome(src)
  })

  it('trend chart uses third-version chart shell tokens', () => {
    const src = read('features/channel-monitor-v2/MonitorTrendChart.vue')
    expect(src).toContain('scheme3-v2-panel')
    expect(src).toContain('scheme3-v2-panel-action')
    expect(src).toContain('scheme3-v2-legend-dot')
    expect(src).toContain('scheme3-v2-empty')
    expect(src).not.toContain('@/components/common/EmptyState.vue')
    expect(src).toContain('min-h-[360px]')
    expectNoLegacyChrome(src)
  })

  it('filter picker uses third-version trigger and teleported dropdown tokens', () => {
    const src = read('features/channel-monitor-v2/FilterMultiSelect.vue')
    expect(src).toContain('scheme3-v2-select-trigger')
    expect(src).toContain('scheme3-v2-filter-dropdown')
    expect(src).toContain('scheme3-v2-filter-option')
    expect(src).toContain('scheme3-v2-checkbox')
    expect(src).toContain('scheme3-v2-dropdown')
    expect(src).not.toContain('select-dropdown-portal')
    expect(src).not.toContain('dropdown-item')
    expect(src).not.toContain('select-option-selected')
    expect(src).not.toContain('select-trigger-open')
    expect(src).not.toContain('rounded-xl')
    expectNoLegacyChrome(src)
  })

  it('settings panel uses third-version panels, rows, and input tokens', () => {
    const src = read('features/channel-monitor-v2/MonitorSettingsPanel.vue')
    const toggle = read('features/channel-monitor-v2/Scheme3V2Toggle.vue')
    expect(src).toContain('scheme3-v2-settings')
    expect(src).toContain('scheme3-v2-settings-header')
    expect(src).toContain('scheme3-v2-settings-panel')
    expect(src).toContain('scheme3-v2-settings-row')
    expect(src).toContain('scheme3-v2-settings-input')
    expect(src).toContain('scheme3-v2-settings-segment')
    expect(src).toContain('Scheme3V2Toggle')
    expect(src).not.toContain("@/components/common/Toggle.vue")
    expect(toggle).toContain('scheme3-v2-toggle')
    expect(toggle).not.toContain('bg-primary-600')
    expect(toggle).not.toContain('ring-0')
    expect(src).toMatch(/max-h-\[min\(40vh/)
    expectNoLegacyChrome(src)
  })

  it('admin monitor header and legacy table use third-version route chrome', () => {
    const src = read('views/admin/ChannelMonitorView.vue')
    const filters = read('components/admin/monitor/MonitorFiltersBar.vue')
    expect(src).toContain('scheme3-admin-monitor')
    expect(src).toContain('scheme3-admin-monitor-header')
    expect(src).toContain('scheme3-admin-monitor-segments')
    expect(src).toContain('scheme3-admin-monitor-segment')
    expect(src).toContain('scheme3-admin-monitor-legacy')
    expect(src).toContain('scheme3-admin-monitor-table-frame')
    expect(src).toContain('scheme3-admin-monitor-empty')
    expect(src).toContain('MonitorSettingsPanel')
    expect(src).not.toContain('TablePageLayout')
    expect(src).not.toContain('EmptyState')
    expect(src).not.toContain("@/components/common/Toggle.vue")
    expect(filters).toContain('scheme3-admin-monitor-filterbar')
    expect(filters).not.toContain('btn btn-')
    expect(filters).not.toContain('class="input')
    expectNoLegacyChrome(src)
    expectNoLegacyChrome(filters)
  })

  it('admin V2 route uses the third-version console shell instead of the legacy app shell', () => {
    const layout = read('components/layout/AppLayout.vue')
    const shell = read('components/layout/Scheme3ConsoleLayout.vue')
    expect(layout).toContain(':admin-mode="useScheme3AdminLayout"')
    expect(layout).toContain('authStore.isAuthenticated && isAdmin.value')
    expect(layout).not.toContain("route.path === '/admin/channels/monitor'")
    expect(layout).toContain("route.path === '/monitor'")
    expect(layout).toContain('useScheme3MonitorLayout')
    expect(shell).toContain('adminMode')
    expect(shell).toContain('admin-routing')
    expect(shell).toContain('scheme3-admin-context')
    expect(shell).not.toContain('AppSidebar')
    expect(shell).not.toContain('AppHeader')
    expect(shell).not.toContain('class="glass')
    expect(shell).not.toContain('page-header')
  })
})

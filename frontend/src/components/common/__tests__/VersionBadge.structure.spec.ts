import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const srcRoot = resolve(__dirname, '../../..')

function read(rel: string) {
  return readFileSync(resolve(srcRoot, rel), 'utf8')
}

describe('VersionBadge viewport positioning contract', () => {
  it('teleports the dropdown and clamps both panel widths to the viewport', () => {
    const src = read('components/common/VersionBadge.vue')

    expect(src).toContain('<Teleport to="body">')
    expect(src).toContain('ref="triggerRef"')
    expect(src).toContain(':style="dropdownStyle"')
    expect(src).toContain('class="scheme3-version-dropdown fixed')
    expect(src).toContain('const DROPDOWN_GUTTER = 8')
    expect(src).toContain('const DEFAULT_DROPDOWN_WIDTH = 256')
    expect(src).toContain('const ROLLBACK_DROPDOWN_WIDTH = 320')
    expect(src).toContain('triggerRect.right - width')
    expect(src).toContain('viewportWidth - width - DROPDOWN_GUTTER')
    expect(src).toContain('viewportHeight - DROPDOWN_GUTTER')
    expect(src).toContain("window.addEventListener('scroll', scheduleDropdownPosition, true)")
  })

  it('keeps Scheme3 visual styles without restoring sidebar-relative positioning', () => {
    const src = read('components/layout/Scheme3ConsoleLayout.vue')

    expect(src).toContain('body.scheme3-user-context .scheme3-version-dropdown')
    expect(src).toContain('body.scheme3-admin-context .scheme3-version-dropdown')
    expect(src).toContain('html.dark body.scheme3-user-context .scheme3-version-dropdown')
    expect(src).not.toContain(
      '.scheme3-console-version-control :deep(.scheme3-version-dropdown)'
    )
    expect(src).not.toContain('width: min(16rem, calc(100vw - 1.3rem))')
  })
})

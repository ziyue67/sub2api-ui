/**
 * Keep scoped Scheme3 dark selectors in a compiler-safe form.
 *
 * Vue's scoped CSS treats `:global(html.dark) .selector` as a global root
 * selector and drops the descendant in the generated rule. The complete
 * selector must live inside one `:global(...)` expression instead.
 */
import { readdirSync, readFileSync } from 'node:fs'
import { join, resolve } from 'node:path'
import { compileStyle, parse } from 'vue/compiler-sfc'
import { describe, expect, it } from 'vitest'

const srcRoot = resolve(__dirname, '../../..')

function vueFilesIn(directory: string): string[] {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = join(directory, entry.name)
    if (entry.isDirectory()) return vueFilesIn(path)
    return entry.isFile() && path.endsWith('.vue') ? [path] : []
  })
}

function hasTrailingDarkGlobalSelector(css: string): boolean {
  const globalStart = ':global('
  let cursor = 0

  while ((cursor = css.indexOf(globalStart, cursor)) !== -1) {
    const contentStart = cursor + globalStart.length
    let depth = 1
    let end = contentStart

    for (; end < css.length && depth > 0; end += 1) {
      if (css[end] === '(') depth += 1
      if (css[end] === ')') depth -= 1
    }

    if (depth !== 0) return true

    const globalContent = css.slice(contentStart, end - 1).trimStart()
    if (/^(?:html\.dark|\.dark)(?:\b|\s|[.#:[>+~])/.test(globalContent)) {
      let next = end
      while (/\s/.test(css[next] ?? '')) next += 1
      if (css[next] !== '{' && css[next] !== ',') return true
    }

    cursor = end
  }

  return false
}

describe('Scheme3 dark selector structure', () => {
  it('recognizes dark :global() selectors that leave descendants outside', () => {
    expect(hasTrailingDarkGlobalSelector(':global(.dark) .panel {}')).toBe(true)
    expect(hasTrailingDarkGlobalSelector(':global(.dark .panel) .child {}')).toBe(true)
    expect(hasTrailingDarkGlobalSelector(':global(html.dark .panel) > .child {}')).toBe(true)
    expect(hasTrailingDarkGlobalSelector(':global(.dark .panel:hover:not(:disabled)) {}')).toBe(
      false
    )
    expect(
      hasTrailingDarkGlobalSelector(
        ':global(.dark .panel), :global(html.dark .other .child:hover) {}'
      )
    ).toBe(false)
  })

  it('keeps every dark descendant selector fully inside :global()', () => {
    const offenders: string[] = []

    for (const path of vueFilesIn(srcRoot)) {
      const filename = path.slice(srcRoot.length + 1)
      const source = readFileSync(path, 'utf8')
      const { descriptor, errors } = parse(source, { filename: path })
      if (errors.length > 0) offenders.push(`${filename}: SFC parse failed`)

      descriptor.styles.forEach((style, index) => {
        if (!style.scoped) return
        const css = style.content.replace(/\/\*[\s\S]*?\*\//g, '')
        if (hasTrailingDarkGlobalSelector(css)) {
          offenders.push(`${filename}: scoped style ${index + 1}`)
        }
      })
    }

    expect(offenders).toEqual([])
  })

  it('compiles scoped dark selectors without dropped descendants or leaked pseudos', () => {
    const failures: string[] = []

    for (const path of vueFilesIn(srcRoot)) {
      const filename = path.slice(srcRoot.length + 1)
      const source = readFileSync(path, 'utf8')
      const { descriptor } = parse(source, { filename: path })

      descriptor.styles.forEach((style, index) => {
        if (!style.scoped) return
        const result = compileStyle({
          filename: path,
          id: `data-v-dark-selector-${index}`,
          source: style.content,
          scoped: true
        })
        const hasLeakedPseudo = /:(?:global|deep|slotted)\(/.test(result.code)
        const hasBareDarkRule = /(?:^|})\s*(?:html\.dark|\.dark)\s*\{/m.test(result.code)
        if (result.errors.length > 0 || hasLeakedPseudo || hasBareDarkRule) {
          failures.push(`${filename}: scoped style ${index + 1}`)
        }
      })
    }

    expect(failures).toEqual([])
  })
})

import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const source = readFileSync(resolve(__dirname, '../ImageCreatorView.vue'), 'utf8')

describe('ImageCreator authenticated preview hydration', () => {
  it('does not expose protected file URLs to img before authenticated fetch completes', () => {
    expect(source).toContain("url: shouldFetchImageUrl(sourceUrl) ? '' : sourceUrl")
    expect(source).toContain('<img v-if="item.url" :src="item.url"')
    expect(source).toContain('class="btn btn-secondary btn-sm" :disabled="!item.url"')
    expect(source).toContain('const sourceUrl = item.sourceUrl || item.url')
    expect(source).toContain('const blob = await downloadImageFile(sourceUrl)')
    expect(source).toContain('current.url = objectUrl')
    expect(source).toContain("if (item.url && !shouldFetchImageUrl(item.url))")
    expect(source).toContain('let componentMounted = false')
    expect(source).toContain('if (!componentMounted || token !== imagePreviewLoadToken || current !== item)')
    expect(source).toContain("appStore.showError(t('imageCreator.downloadFailed'))")
  })
})

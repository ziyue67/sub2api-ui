import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const source = readFileSync(
  resolve(dirname(fileURLToPath(import.meta.url)), '../MonitorRunResultDialog.vue'),
  'utf8',
)

describe('MonitorRunResultDialog responsive structure', () => {
  it('stacks quota and status on narrow dialogs, then restores the row layout at sm', () => {
    expect(source).toContain('flex flex-col items-stretch gap-2')
    expect(source).toContain('sm:flex-row sm:items-center sm:justify-between')
    expect(source).toContain('min-w-0 flex flex-col')
    expect(source).toContain('flex shrink-0 items-center justify-end gap-2 sm:justify-start')
    expect(source).not.toContain('scheme3-monitor-result-row flex items-center justify-between')
  })
})

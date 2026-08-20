import { describe, expect, it } from 'vitest'
import { renderMarkdown } from '~/composables/useMarkdown'

describe('renderMarkdown', () => {
  it('renders basic markdown', () => {
    expect(renderMarkdown('**bold**')).toContain('<strong>bold</strong>')
  })

  it('sanitizes injected scripts', () => {
    const html = renderMarkdown('hello <script>alert(1)</script>')
    expect(html).not.toContain('<script')
    expect(html).toContain('hello')
  })

  it('strips event handler attributes', () => {
    expect(renderMarkdown('<img src="x" onerror="alert(1)">')).not.toContain('onerror')
  })

  it('forces links to open in a new tab with rel=noopener', () => {
    const html = renderMarkdown('[x](https://example.com)')
    expect(html).toContain('target="_blank"')
    expect(html).toContain('rel="noopener"')
  })

  it('returns an empty string for blank input', () => {
    expect(renderMarkdown('   ')).toBe('')
  })
})

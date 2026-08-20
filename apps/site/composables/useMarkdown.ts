import { marked } from 'marked'
import DOMPurify from 'dompurify'

/**
 * Render markdown to sanitized html. Links open in a new tab.
 * Client-side only (DOMPurify needs a DOM) — call from onMounted.
 */
export function renderMarkdown(md: string): string {
  if (!md.trim()) return ''
  const html = (marked.parse(md, { async: false }) as string).replace(
    /<a /g,
    '<a target="_blank" rel="noopener" ',
  )
  return DOMPurify.sanitize(html, { ADD_ATTR: ['target'] })
}

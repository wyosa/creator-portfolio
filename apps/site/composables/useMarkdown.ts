import { marked } from 'marked'
import DOMPurify from 'isomorphic-dompurify'

/**
 * Render markdown to sanitized html. Links open in a new tab.
 * Isomorphic: works during SSR (jsdom-backed on the server via
 * isomorphic-dompurify's node export, plain window dompurify on the client).
 */
export function renderMarkdown(md: string): string {
  if (!md.trim()) return ''
  const html = (marked.parse(md, { async: false }) as string).replace(
    /<a /g,
    '<a target="_blank" rel="noopener" ',
  )
  return DOMPurify.sanitize(html, { ADD_ATTR: ['target'] })
}

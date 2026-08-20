import type { MediaItem, UploadResponse } from '~/types/media'

/**
 * Public, ordered media list.
 * Degrades to an empty array when the api is unreachable so pages
 * still render (and SSR never 500s because the backend is down).
 */
export function useMedia() {
  return useFetch<MediaItem[]>('/api/media', {
    key: 'media-list',
    default: () => [] as MediaItem[],
  })
}

/** local file path of a looped/self-hosted video (by extension) */
export function isVideoPath(path: string): boolean {
  return /\.(mp4|webm|mov|m4v)$/i.test(path)
}

/** still frame served by youtube for the given video id */
export function youtubeThumb(id: string): string {
  return `https://i.ytimg.com/vi/${id}/hqdefault.jpg`
}

/** human-readable fallback label: title → path → external id → #id */
export function mediaLabel(item: MediaItem): string {
  return item.title || item.path || item.external_id || `#${item.id}`
}

/** upload a single file; resolves to the stored paths and dimensions */
export async function uploadFile(
  file: File,
  api: <T = unknown>(url: string, opts?: Parameters<typeof $fetch>[1]) => Promise<T> = $fetch,
): Promise<UploadResponse> {
  const form = new FormData()
  form.append('file', file)
  return api<UploadResponse>('/api/upload', { method: 'POST', body: form })
}

/**
 * Iframe embed url for a youtube/vimeo item.
 * 'grid' — muted inline loop without player chrome; 'player' — lightbox
 * player with sound controls.
 */
export function buildEmbedUrl(item: MediaItem, mode: 'grid' | 'player'): string | null {
  if (item.type !== 'video' || !item.external_id) return null
  const id = item.external_id

  if (item.source === 'youtube') {
    return mode === 'grid'
      ? `https://www.youtube-nocookie.com/embed/${id}?autoplay=1&mute=1&loop=1&playlist=${id}&controls=0&rel=0&modestbranding=1&disablekb=1&fs=0&iv_load_policy=3&playsinline=1`
      : `https://www.youtube-nocookie.com/embed/${id}?autoplay=1&controls=1&rel=0&modestbranding=1&playsinline=1`
  }

  if (item.source === 'vimeo') {
    return mode === 'grid'
      ? `https://player.vimeo.com/video/${id}?autoplay=1&muted=1&loop=1&background=1&byline=0&title=0&api=1`
      : `https://player.vimeo.com/video/${id}?autoplay=1&byline=0&title=0`
  }

  return null
}

/**
 * Vimeo background embeds auto-pause when the tab or the player itself gets
 * covered (e.g. by the lightbox). Ping them with the player api to resume.
 */
export function resumeVimeoEmbeds(root: HTMLElement | null) {
  if (!root) return
  root.querySelectorAll<HTMLIFrameElement>('iframe[src*="player.vimeo.com"]').forEach((f) => {
    f.contentWindow?.postMessage(JSON.stringify({ method: 'play' }), 'https://player.vimeo.com')
  })
}

/**
 * Parse a youtube or vimeo url into an external id.
 * Returns null when the url is not recognized for the given source.
 */
export function parseVideoUrl(
  raw: string,
  source: 'youtube' | 'vimeo',
): { source: 'youtube' | 'vimeo'; externalId: string } | null {
  const url = raw.trim()
  if (!url) return null

  if (source === 'youtube') {
    const match =
      url.match(/youtube\.com\/watch\?[^#]*v=([\w-]{6,})/) ||
      url.match(/(?:youtu\.be\/|youtube\.com\/(?:shorts|embed)\/)([\w-]{6,})/)
    return match ? { source, externalId: match[1] } : null
  }

  const match = url.match(/(?:player\.)?vimeo\.com\/(?:video\/)?(\d+)/)
  return match ? { source, externalId: match[1] } : null
}

export type MediaType = 'photo' | 'video'

export type MediaSource = 'upload' | 'youtube' | 'vimeo'

export interface MediaItem {
  id: number
  type: MediaType
  source: MediaSource
  path: string
  /** tiny blurred placeholder; absent on older records (derived by extension then) */
  thumb?: string
  external_id: string
  title: string
  description: string
  instagram_url: string
  youtube_url: string
  vimeo_url: string
  preview_path: string
  width: number
  height: number
  featured: boolean
  position: number
  created_at: string
  /** field → lang → text */
  translations: Record<string, Record<string, string>>
}

export interface UploadResponse {
  path: string
  kind: string
  thumb: string
  width: number
  height: number
}

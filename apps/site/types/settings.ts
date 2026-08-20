export interface InfoLink {
  label: string
  url: string
  enabled: boolean
}

export interface SiteSettings {
  site_name: string
  site_subtitle: string
  info_text: string
  info_links: InfoLink[]
  languages: string[]
  /** field → lang → text */
  translations: Record<string, Record<string, string>>
}

/** info link in the admin form, with a stable client-side key */
export interface InfoLinkDraft extends InfoLink {
  cid: number
}

/** working copy of the site settings edited in the admin form */
export interface SettingsForm {
  site_name: string
  site_subtitle: string
  info_text: string
  info_links: InfoLinkDraft[]
  languages: string[]
  /** field → lang → text */
  tr: Record<'site_name' | 'site_subtitle' | 'info_text', Record<string, string>>
}

export const defaultSettings: SiteSettings = {
  site_name: 'your name',
  site_subtitle: 'photographer & videographer',
  info_text:
    "i'm a photographer & videographer. this site is a selection of personal and commissioned work.",
  info_links: [
    { label: 'email', url: 'mailto:hello@example.com', enabled: true },
    { label: 'instagram', url: 'https://instagram.com/', enabled: true },
    { label: 'telegram', url: 'https://t.me/', enabled: true },
  ],
  languages: ['en'],
  translations: {},
}

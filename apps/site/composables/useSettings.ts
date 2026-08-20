import type { MaybeRefOrGetter } from 'vue'
import type { InfoLink, InfoLinkDraft, SiteSettings } from '~/types/settings'
import { defaultSettings } from '~/types/settings'

/**
 * Site-wide settings (site name, info page content, ...).
 * Cached under a fixed key so header + pages share one request.
 * Degrades to defaults when the api is unreachable.
 */
export function useSettings() {
  return useFetch<SiteSettings>('/api/settings', {
    key: 'site-settings',
    default: () => ({ ...defaultSettings }),
  })
}

let nextLinkCid = 0

/** info link → admin form draft with a stable key for v-for */
export function toLinkDraft(link: InfoLink): InfoLinkDraft {
  return { ...link, cid: ++nextLinkCid }
}

/** document title: "<label> — <site name>" or just the site name */
export function usePageTitle(label?: MaybeRefOrGetter<string>) {
  const { data: settings } = useSettings()
  const { pick } = useI18n()
  useHead({
    title: computed(() => {
      const name = pick(
        settings.value?.translations?.site_name,
        settings.value?.site_name || 'portfolio',
      )
      const l = label ? toValue(label) : ''
      return l ? `${l} — ${name}` : name
    }),
  })
}

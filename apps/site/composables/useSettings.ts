import type { ComputedRef, MaybeRefOrGetter } from 'vue'
import type { InfoLink, InfoLinkDraft, SettingsForm, SiteSettings } from '~/types/settings'
import { defaultSettings } from '~/types/settings'

/**
 * Site-wide settings (site name, info page content, ...).
 * Cached under a fixed key so header + pages share one request.
 * Degrades to defaults when the api is unreachable.
 */
export function useSettings() {
  return useFetch<SiteSettings>('/api/settings', {
    key: 'site-settings',
    // deep copy: drafts must not share the default info_links array by reference
    default: () => structuredClone(defaultSettings),
  })
}

let nextLinkCid = 0

/** info link → admin form draft with a stable key for v-for */
export function toLinkDraft(link: InfoLink): InfoLinkDraft {
  return { ...link, cid: ++nextLinkCid }
}

/**
 * Keep base columns and per-language translations consistent before saving.
 * Multi-language: base mirrors the primary-lang translation (edited there).
 * Single-language: the inputs edit base columns, so mirror base → tr —
 * otherwise a stale translation would keep shadowing base on the site.
 */
export function syncSettingsTranslations(form: SettingsForm): void {
  const primary = form.languages.includes('en') ? 'en' : (form.languages[0] ?? 'en')
  for (const f of ['site_name', 'site_subtitle', 'info_text'] as const) {
    if (form.languages.length > 1) {
      const v = form.tr[f][primary]
      if (v !== undefined) form[f] = v
    } else {
      form.tr[f][primary] = form[f]
    }
  }
}

/** "<label> — <site name>" title text, shared by usePageTitle and useSeo */
export function useSiteTitle(label?: MaybeRefOrGetter<string>): ComputedRef<string> {
  const { data: settings } = useSettings()
  const { pick } = useI18n()
  return computed(() => {
    const name = pick(
      settings.value?.translations?.site_name,
      settings.value?.site_name || 'portfolio',
    )
    const l = label ? toValue(label) : ''
    return l ? `${l} — ${name}` : name
  })
}

/** document title: "<label> — <site name>" or just the site name */
export function usePageTitle(label?: MaybeRefOrGetter<string>) {
  useHead({ title: useSiteTitle(label) })
}

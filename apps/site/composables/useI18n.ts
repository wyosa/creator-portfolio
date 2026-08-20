import type { ComputedRef } from 'vue'
import { SUPPORTED_LOCALES, type Locale } from '~/types/i18n'
import { messages, pickTranslation, type MessageKey } from '~/i18n/messages'

const STORAGE_KEY = 'site-locale'

/* read the saved locale from localStorage once per app, not per useI18n() call */
let storageInitHooked = false

export function useI18n() {
  const { data: settings } = useSettings()

  /** languages the owner enabled in the admin */
  const active: ComputedRef<Locale[]> = computed(() => {
    const langs = settings.value?.languages ?? []
    const valid = langs.filter((l): l is Locale =>
      (SUPPORTED_LOCALES as readonly string[]).includes(l),
    )
    return valid.length ? valid : ['en']
  })

  /*
   * preferred = raw preference (saved choice or accept-language), possibly
   * null until known. locale is COMPUTED from it + active langs, so there is
   * no race: whenever settings arrive, the effective locale re-resolves.
   */
  const preferred = useState<string | null>('site-locale-pref', () => null)
  const manual = useState<boolean>('site-locale-manual', () => false)

  if (import.meta.server && preferred.value === null && !manual.value) {
    const headers = useRequestHeaders(['accept-language'])
    preferred.value = headers['accept-language'] ?? null
  }

  if (import.meta.client && !storageInitHooked) {
    storageInitHooked = true
    onMounted(() => {
      if (manual.value) return
      let saved: string | null
      try {
        saved = localStorage.getItem(STORAGE_KEY)
      } catch {
        saved = null
      }
      if (saved) {
        preferred.value = saved
        manual.value = true
        return
      }
      if (preferred.value === null) preferred.value = navigator.language
    })
  }

  const locale: ComputedRef<Locale> = computed(() => {
    const act = active.value
    if (preferred.value) {
      const short = preferred.value.toLowerCase().split(/[-_,]/)[0] as Locale
      if ((act as string[]).includes(short)) return short
    }
    if ((act as string[]).includes('en')) return 'en'
    return act[0]
  })

  function setLocale(l: Locale) {
    preferred.value = l
    manual.value = true
    try {
      localStorage.setItem(STORAGE_KEY, l)
    } catch {
      // private mode etc
    }
  }

  function t(key: MessageKey): string {
    return messages[locale.value]?.[key] ?? messages.en[key] ?? key
  }

  /** translated field → english → base column */
  function pick(tr: Record<string, string> | undefined | null, base: string): string {
    return pickTranslation(tr, base, locale.value)
  }

  return { locale, active, setLocale, t, pick }
}

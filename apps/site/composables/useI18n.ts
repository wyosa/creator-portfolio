import type { ComputedRef } from 'vue'
import { SUPPORTED_LOCALES, type Locale } from '~/types/i18n'

const messages: Record<Locale, Record<string, string>> = {
  en: {
    featured: 'featured',
    film: 'film',
    photo: 'photo',
    info: 'info',
    all: 'all',
    pageNotFound: 'page not found',
    backHome: 'back home',
    nothingHere: 'nothing here yet',
    unavailable: 'temporarily unavailable, please try again later',
  },
  ru: {
    featured: 'избранное',
    film: 'видео',
    photo: 'фото',
    info: 'инфо',
    all: 'всё',
    pageNotFound: 'страница не найдена',
    backHome: 'на главную',
    nothingHere: 'пока пусто',
    unavailable: 'временно недоступно, попробуйте позже',
  },
  es: {
    featured: 'destacados',
    film: 'vídeo',
    photo: 'foto',
    info: 'info',
    all: 'todo',
    pageNotFound: 'página no encontrada',
    backHome: 'volver al inicio',
    nothingHere: 'aún no hay nada',
    unavailable: 'temporalmente no disponible, inténtalo más tarde',
  },
  et: {
    featured: 'valitud',
    film: 'film',
    photo: 'foto',
    info: 'info',
    all: 'kõik',
    pageNotFound: 'lehte ei leitud',
    backHome: 'avalehele',
    nothingHere: 'siia pole veel midagi lisatud',
    unavailable: 'ajutiselt kättesaamatu, proovi hiljem uuesti',
  },
  de: {
    featured: 'ausgewählt',
    film: 'film',
    photo: 'foto',
    info: 'info',
    all: 'alles',
    pageNotFound: 'seite nicht gefunden',
    backHome: 'zur startseite',
    nothingHere: 'noch nichts hier',
    unavailable: 'vorübergehend nicht verfügbar, bitte später erneut versuchen',
  },
}

const STORAGE_KEY = 'site-locale'

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

  onMounted(() => {
    if (manual.value) return
    let saved: string | null = null
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

  function t(key: string): string {
    return messages[locale.value]?.[key] ?? messages.en[key] ?? key
  }

  /** translated field → english → base column */
  function pick(tr: Record<string, string> | undefined | null, base: string): string {
    if (!tr) return base
    return tr[locale.value] || tr.en || base
  }

  return { locale, active, setLocale, t, pick }
}

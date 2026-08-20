export const SUPPORTED_LOCALES = ['en', 'ru', 'es', 'et', 'de'] as const

export type Locale = (typeof SUPPORTED_LOCALES)[number]

export const LOCALE_NAMES: Record<Locale, string> = {
  en: 'english',
  ru: 'русский',
  es: 'español',
  et: 'eesti',
  de: 'deutsch',
}

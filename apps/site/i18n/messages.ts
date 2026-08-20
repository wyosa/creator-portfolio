import type { Locale } from '~/types/i18n'

/*
 * Ui strings per locale. `en` is the reference: every other locale must
 * provide exactly the same keys (enforced by the Record type below and by
 * the parity test).
 */
const en = {
  featured: 'featured',
  film: 'film',
  photo: 'photo',
  info: 'info',
  all: 'all',
  pageNotFound: 'page not found',
  somethingWentWrong: 'something went wrong',
  backHome: 'back home',
  nothingHere: 'nothing here yet',
  unavailable: 'temporarily unavailable, please try again later',
  tooManyAttempts: 'too many attempts, please try again later',
  toggleMenu: 'toggle menu',
  openPhoto: 'open photo',
  playVideo: 'play video with sound',
  close: 'close',
  previous: 'previous',
  next: 'next',
}

export type MessageKey = keyof typeof en

export const messages: Record<Locale, Record<MessageKey, string>> = {
  en,
  ru: {
    featured: 'избранное',
    film: 'видео',
    photo: 'фото',
    info: 'инфо',
    all: 'всё',
    pageNotFound: 'страница не найдена',
    somethingWentWrong: 'что-то пошло не так',
    backHome: 'на главную',
    nothingHere: 'пока пусто',
    unavailable: 'временно недоступно, попробуйте позже',
    tooManyAttempts: 'слишком много попыток, попробуйте позже',
    toggleMenu: 'открыть или закрыть меню',
    openPhoto: 'открыть фото',
    playVideo: 'смотреть видео со звуком',
    close: 'закрыть',
    previous: 'назад',
    next: 'вперёд',
  },
  es: {
    featured: 'destacados',
    film: 'vídeo',
    photo: 'foto',
    info: 'info',
    all: 'todo',
    pageNotFound: 'página no encontrada',
    somethingWentWrong: 'algo salió mal',
    backHome: 'volver al inicio',
    nothingHere: 'aún no hay nada',
    unavailable: 'temporalmente no disponible, inténtalo más tarde',
    tooManyAttempts: 'demasiados intentos, inténtalo más tarde',
    toggleMenu: 'abrir o cerrar el menú',
    openPhoto: 'abrir foto',
    playVideo: 'reproducir vídeo con sonido',
    close: 'cerrar',
    previous: 'anterior',
    next: 'siguiente',
  },
  et: {
    featured: 'valitud',
    film: 'film',
    photo: 'foto',
    info: 'info',
    all: 'kõik',
    pageNotFound: 'lehte ei leitud',
    somethingWentWrong: 'midagi läks valesti',
    backHome: 'avalehele',
    nothingHere: 'siia pole veel midagi lisatud',
    unavailable: 'ajutiselt kättesaamatu, proovi hiljem uuesti',
    tooManyAttempts: 'liiga palju katseid, proovi hiljem uuesti',
    toggleMenu: 'ava või sulge menüü',
    openPhoto: 'ava foto',
    playVideo: 'esita video heliga',
    close: 'sulge',
    previous: 'eelmine',
    next: 'järgmine',
  },
  de: {
    featured: 'ausgewählt',
    film: 'film',
    photo: 'foto',
    info: 'info',
    all: 'alles',
    pageNotFound: 'seite nicht gefunden',
    somethingWentWrong: 'etwas ist schiefgelaufen',
    backHome: 'zur startseite',
    nothingHere: 'noch nichts hier',
    unavailable: 'vorübergehend nicht verfügbar, bitte später erneut versuchen',
    tooManyAttempts: 'zu viele versuche, bitte später erneut versuchen',
    toggleMenu: 'menü öffnen oder schließen',
    openPhoto: 'foto öffnen',
    playVideo: 'video mit ton abspielen',
    close: 'schließen',
    previous: 'zurück',
    next: 'weiter',
  },
}

/** translated field → english → base column */
export function pickTranslation(
  tr: Record<string, string> | undefined | null,
  base: string,
  locale: Locale,
): string {
  if (!tr) return base
  return tr[locale] || tr.en || base
}

import { describe, expect, it } from 'vitest'
import { messages, pickTranslation } from '~/i18n/messages'
import { SUPPORTED_LOCALES } from '~/types/i18n'

describe('i18n messages', () => {
  const reference = Object.keys(messages.en).sort()

  for (const locale of SUPPORTED_LOCALES) {
    it(`${locale} has exactly the reference (en) keys`, () => {
      expect(Object.keys(messages[locale]).sort()).toEqual(reference)
    })
  }
})

describe('pickTranslation', () => {
  it('prefers the current locale', () => {
    expect(pickTranslation({ ru: 'привет', en: 'hello' }, 'base', 'ru')).toBe('привет')
  })

  it('falls back to english when the locale is missing or empty', () => {
    expect(pickTranslation({ en: 'hello' }, 'base', 'ru')).toBe('hello')
    expect(pickTranslation({ en: 'hello', ru: '' }, 'base', 'ru')).toBe('hello')
  })

  it('falls back to the base column otherwise', () => {
    expect(pickTranslation({ ru: 'привет' }, 'base', 'de')).toBe('base')
    expect(pickTranslation({}, 'base', 'de')).toBe('base')
    expect(pickTranslation(undefined, 'base', 'de')).toBe('base')
    expect(pickTranslation(null, 'base', 'de')).toBe('base')
  })
})

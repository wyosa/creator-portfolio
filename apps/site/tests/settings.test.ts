import { describe, expect, it } from 'vitest'
import { syncSettingsTranslations } from '~/composables/useSettings'
import type { SettingsForm } from '~/types/settings'

function makeForm(overrides: Partial<SettingsForm> = {}): SettingsForm {
  return {
    site_name: 'base name',
    site_subtitle: 'base subtitle',
    info_text: 'base info',
    info_links: [],
    languages: ['en'],
    tr: { site_name: {}, site_subtitle: {}, info_text: {} },
    ...overrides,
  }
}

describe('syncSettingsTranslations', () => {
  it('single language: mirrors base → tr so a stale tr cannot shadow base', () => {
    const form = makeForm({
      languages: ['en'],
      site_name: 'new name',
      tr: { site_name: { en: 'old name' }, site_subtitle: {}, info_text: {} },
    })
    syncSettingsTranslations(form)
    expect(form.site_name).toBe('new name')
    expect(form.tr.site_name.en).toBe('new name')
  })

  it('single non-english language: mirrors base into that language', () => {
    const form = makeForm({ languages: ['ru'], site_name: 'новое имя' })
    syncSettingsTranslations(form)
    expect(form.tr.site_name.ru).toBe('новое имя')
  })

  it('multi language: base mirrors the primary (en) translation', () => {
    const form = makeForm({
      languages: ['en', 'ru'],
      site_name: 'stale base',
      tr: {
        site_name: { en: 'english name', ru: 'русское имя' },
        site_subtitle: {},
        info_text: {},
      },
    })
    syncSettingsTranslations(form)
    expect(form.site_name).toBe('english name')
    expect(form.tr.site_name.ru).toBe('русское имя')
  })

  it('multi language without en: primary is the first active language', () => {
    const form = makeForm({
      languages: ['ru', 'de'],
      tr: { site_name: { ru: 'имя', de: 'name' }, site_subtitle: {}, info_text: {} },
    })
    syncSettingsTranslations(form)
    expect(form.site_name).toBe('имя')
  })
})

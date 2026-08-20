<script setup lang="ts">
import type { ComputedRef } from 'vue'
import { SUPPORTED_LOCALES, type Locale } from '~/types/i18n'
import type { SettingsForm } from '~/types/settings'

const props = defineProps<{ form: SettingsForm }>()

const orderedLangs: ComputedRef<Locale[]> = computed(() =>
  SUPPORTED_LOCALES.filter((l) => props.form.languages.includes(l)),
)
const multiLang = computed(() => props.form.languages.length > 1)

/* which language the settings fields currently edit */
const settingsLang = ref('en')
const editLang = computed(() =>
  orderedLangs.value.some((l) => l === settingsLang.value)
    ? settingsLang.value
    : (orderedLangs.value[0] ?? 'en'),
)

// prefill per-lang fields: stored translation, base value for english
for (const f of ['site_name', 'site_subtitle', 'info_text'] as const) {
  for (const l of orderedLangs.value) {
    if (!props.form.tr[f][l]) props.form.tr[f][l] = l === 'en' ? props.form[f] : ''
  }
}

function toggleLang(l: Locale) {
  const i = props.form.languages.indexOf(l)
  if (i >= 0) {
    if (props.form.languages.length > 1) props.form.languages.splice(i, 1)
  } else {
    props.form.languages.push(l)
    for (const f of ['site_name', 'site_subtitle', 'info_text'] as const) {
      if (!props.form.tr[f][l]) props.form.tr[f][l] = l === 'en' ? props.form[f] : ''
    }
  }
}
</script>

<template>
  <div class="admin__settings-part">
    <span class="tiny admin__add-label">languages</span>
    <div class="admin__langs">
      <label v-for="l in SUPPORTED_LOCALES" :key="l" class="row__check">
        <input type="checkbox" :checked="form.languages.includes(l)" @change="toggleLang(l)" />
        <span class="tiny">{{ l }}</span>
      </label>
    </div>

    <span class="tiny admin__add-label">site</span>

    <div v-if="multiLang" class="admin__lang-tabs">
      <button
        v-for="l in orderedLangs"
        :key="l"
        type="button"
        class="seg__btn"
        :class="{ 'seg__btn--active': editLang === l }"
        @click="settingsLang = l"
      >
        {{ l }}
      </button>
    </div>

    <template v-if="multiLang">
      <div class="admin__settings-grid">
        <label class="field">
          <span>site name</span>
          <input v-model="form.tr.site_name[editLang]" type="text" />
        </label>
        <label class="field">
          <span>site subtitle</span>
          <input v-model="form.tr.site_subtitle[editLang]" type="text" />
        </label>
      </div>
      <label class="field">
        <span>info text (markdown)</span>
        <textarea v-model="form.tr.info_text[editLang]" rows="4" />
      </label>
    </template>

    <template v-else>
      <div class="admin__settings-grid">
        <label class="field">
          <span>site name</span>
          <input v-model="form.site_name" type="text" />
        </label>
        <label class="field">
          <span>site subtitle</span>
          <input v-model="form.site_subtitle" type="text" />
        </label>
      </div>

      <label class="field">
        <span>info text (markdown)</span>
        <textarea v-model="form.info_text" rows="4" />
      </label>
    </template>
  </div>
</template>

<style scoped>
.admin__settings-part {
  align-self: stretch;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 1.75rem;
}

.admin__settings-grid {
  width: 100%;
  max-width: 40rem;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem 2rem;
}

.admin__settings-part .field {
  width: 100%;
  max-width: 40rem;
}

.admin__langs {
  display: flex;
  gap: 1.25rem;
}

.admin__lang-tabs {
  display: flex;
  gap: 1rem;
}

.admin__add-label {
  color: var(--muted);
}

.row__check {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  color: var(--muted);
  cursor: pointer;
  white-space: nowrap;
}

.row__check input {
  accent-color: var(--color);
  margin: 0;
  cursor: pointer;
}
</style>

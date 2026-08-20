<script setup lang="ts">
import type { MediaItem } from '~/types/media'
import type { SettingsForm } from '~/types/settings'
import { SUPPORTED_LOCALES } from '~/types/i18n'

definePageMeta({ middleware: 'auth' })
usePageTitle('admin')

const { data: items, pending, refresh } = await useMedia()
const { data: settings } = await useSettings()

const busy = ref(false)
const actionError = ref('')
const saved = ref(false)
let savedTimer: ReturnType<typeof setTimeout> | undefined

function showSaved() {
  saved.value = true
  clearTimeout(savedTimer)
  savedTimer = setTimeout(() => {
    saved.value = false
  }, 1500)
}

onBeforeUnmount(() => clearTimeout(savedTimer))

/** run a mutation, then refetch the list; onError may roll back optimistic state */
async function runMutation(fn: () => Promise<unknown>, onError?: () => void): Promise<boolean> {
  if (busy.value) return false
  busy.value = true
  actionError.value = ''
  try {
    await fn()
    await refresh()
    showSaved()
    return true
  } catch {
    onError?.()
    actionError.value = 'something went wrong'
    return false
  } finally {
    busy.value = false
  }
}

/* --- site settings (fields live in AdminSettings/AdminLinks) --- */

const settingsForm: SettingsForm = reactive({
  site_name: settings.value.site_name,
  site_subtitle: settings.value.site_subtitle,
  info_text: settings.value.info_text,
  info_links: settings.value.info_links.map(toLinkDraft),
  languages: [...(settings.value.languages?.length ? settings.value.languages : ['en'])],
  tr: {
    site_name: { ...(settings.value.translations?.site_name ?? {}) },
    site_subtitle: { ...(settings.value.translations?.site_subtitle ?? {}) },
    info_text: { ...(settings.value.translations?.info_text ?? {}) },
  },
})

async function saveSettings() {
  // mirror base (fallback) columns from english (or the first active lang)
  const primary = settingsForm.languages.includes('en') ? 'en' : settingsForm.languages[0]
  for (const f of ['site_name', 'site_subtitle', 'info_text'] as const) {
    const v = settingsForm.tr[f][primary]
    if (settingsForm.languages.length > 1 && v !== undefined) settingsForm[f] = v
  }
  await runMutation(async () => {
    const res = await $fetch('/api/settings', {
      method: 'PUT',
      body: {
        site_name: settingsForm.site_name,
        site_subtitle: settingsForm.site_subtitle,
        info_text: settingsForm.info_text,
        info_links: settingsForm.info_links.map(({ cid: _cid, ...l }) => l),
        languages: SUPPORTED_LOCALES.filter((l) => settingsForm.languages.includes(l)),
        translations: settingsForm.tr,
      },
    })
    settings.value = res as typeof settings.value
  })
}

/* --- media list actions --- */

function removeItem(item: MediaItem) {
  if (!confirm(`delete "${mediaLabel(item)}"?`)) return
  void runMutation(() => $fetch(`/api/media/${item.id}`, { method: 'DELETE' }))
}

function onReorder(from: number, to: number) {
  if (!items.value.length) return

  const backup = [...items.value]
  const list = [...items.value]
  const [moved] = list.splice(from, 1)
  list.splice(to, 0, moved)
  items.value = list

  void runMutation(
    () =>
      $fetch('/api/media/reorder', {
        method: 'PUT',
        body: { ids: list.map((m) => m.id) },
      }),
    () => {
      items.value = backup
    },
  )
}

async function onMediaChanged() {
  await refresh()
  showSaved()
}

/* --- edit modal --- */

const editingItem = ref<MediaItem | null>(null)
</script>

<template>
  <div class="admin">
    <section class="admin__settings" aria-label="site settings">
      <AdminSettings :form="settingsForm" />
      <AdminLinks :links="settingsForm.info_links" />
      <div>
        <button type="button" class="btn" :disabled="busy" @click="saveSettings">
          save settings
        </button>
      </div>
    </section>

    <div class="admin__bar">
      <span class="tiny">media</span>
      <span class="admin__bar-right">
        <span v-if="saved" class="tiny admin__saved">saved</span>
        <span v-if="actionError" class="tiny">{{ actionError }}</span>
      </span>
    </div>

    <AdminMediaAdd
      :busy="busy"
      :run-mutation="runMutation"
      @changed="onMediaChanged"
      @failed="actionError = $event"
    />

    <AdminMediaList
      :items="items"
      :pending="pending"
      :busy="busy"
      @reorder="onReorder"
      @delete="removeItem"
      @edit="editingItem = $event"
    />

    <AdminEditModal
      v-if="editingItem"
      :item="editingItem"
      @close="editingItem = null"
      @saved="refresh"
    />
  </div>
</template>

<style scoped>
.admin {
  max-width: 64rem;
  margin: 0 auto;
  padding: 1rem 0 4rem;
}

.admin__settings {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 1.75rem;
  padding: 2rem 0 3.5rem;
}

.admin__bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 3rem 0 1rem;
}

.admin__bar-right {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.admin__saved {
  color: var(--muted);
}
</style>

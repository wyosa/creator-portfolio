<script setup lang="ts">
import type { MediaItem } from '~/types/media'

const props = defineProps<{ item: MediaItem }>()
const emit = defineEmits<{ close: []; saved: [] }>()

const { data: settings } = useSettings()
const langs = computed(() =>
  settings.value?.languages?.length ? settings.value.languages : ['en'],
)
const multiLang = computed(() => langs.value.length > 1)
const curLang = ref(langs.value[0])

/* per-language title/description; base columns mirror the primary lang */
const trForm = reactive({
  title: { ...(props.item.translations?.title ?? {}) },
  description: { ...(props.item.translations?.description ?? {}) },
})
const primaryLang = computed(() => (langs.value.includes('en') ? 'en' : langs.value[0]))
if (!trForm.title[primaryLang.value]) trForm.title[primaryLang.value] = props.item.title
if (!trForm.description[primaryLang.value])
  trForm.description[primaryLang.value] = props.item.description

const form = reactive({
  instagram_url: props.item.instagram_url,
  youtube_url: props.item.youtube_url,
  vimeo_url: props.item.vimeo_url,
  featured: props.item.featured,
})
const previewPath = ref(props.item.preview_path)
const previewInput = ref<HTMLInputElement | null>(null)
const busy = ref(false)
const error = ref('')

const isExternalVideo = computed(
  () => props.item.type === 'video' && props.item.source !== 'upload',
)

async function save() {
  busy.value = true
  error.value = ''
  try {
    await $fetch(`/api/media/${props.item.id}`, {
      method: 'PUT',
      body: {
        title: (trForm.title[primaryLang.value] ?? '').trim(),
        description: trForm.description[primaryLang.value] ?? '',
        instagram_url: form.instagram_url.trim(),
        youtube_url: form.youtube_url.trim(),
        vimeo_url: form.vimeo_url.trim(),
        featured: form.featured,
        preview_path: previewPath.value,
        translations: { title: trForm.title, description: trForm.description },
      },
    })
    emit('saved')
    emit('close')
  } catch {
    error.value = 'save failed'
  } finally {
    busy.value = false
  }
}

async function onPreviewPicked(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  busy.value = true
  error.value = ''
  try {
    const { path } = await uploadFile(file)
    previewPath.value = path
  } catch {
    error.value = 'preview upload failed'
  } finally {
    busy.value = false
  }
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') emit('close')
}

onMounted(() => window.addEventListener('keydown', onKeydown))
onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown))
</script>

<template>
  <Teleport to="body">
    <div
      class="emodal"
      role="dialog"
      aria-modal="true"
      aria-label="edit media"
      @click.self="emit('close')"
    >
      <div class="emodal__panel">
        <div class="emodal__head">
          <span class="tiny">edit — {{ item.type }} / {{ item.source }}</span>
          <button type="button" class="emodal__close" aria-label="close" @click="emit('close')">
            ×
          </button>
        </div>

        <div class="emodal__thumb">
          <MediaThumb :item="item" :label="`vimeo #${item.external_id}`" />
        </div>

        <div v-if="multiLang" class="emodal__langs">
          <button
            v-for="l in langs"
            :key="l"
            type="button"
            class="emodal__lang"
            :class="{ 'emodal__lang--active': curLang === l }"
            @click="curLang = l"
          >
            {{ l }}
          </button>
        </div>

        <label class="field">
          <span>title</span>
          <input v-model="trForm.title[curLang]" type="text" />
        </label>

        <label class="field">
          <span>description (markdown)</span>
          <textarea
            v-model="trForm.description[curLang]"
            rows="4"
            placeholder="**bold**, [links](https://...), lists..."
          />
        </label>

        <label v-if="item.type === 'photo'" class="field">
          <span>instagram url</span>
          <input
            v-model="form.instagram_url"
            type="text"
            placeholder="https://instagram.com/p/..."
          />
        </label>

        <template v-if="item.type === 'video'">
          <label class="field">
            <span>youtube url (icon on the card)</span>
            <input
              v-model="form.youtube_url"
              type="text"
              placeholder="https://youtube.com/watch?v=..."
            />
          </label>
          <label class="field">
            <span>vimeo url (icon on the card)</span>
            <input v-model="form.vimeo_url" type="text" placeholder="https://vimeo.com/..." />
          </label>
        </template>

        <div v-if="isExternalVideo" class="emodal__preview">
          <span class="tiny">preview (looped in the grid instead of the embed)</span>
          <div class="emodal__preview-row">
            <div class="emodal__preview-thumb">
              <template v-if="previewPath">
                <video
                  v-if="isVideoPath(previewPath)"
                  :src="previewPath"
                  muted
                  playsinline
                  preload="metadata"
                />
                <img v-else :src="previewPath" alt="preview" />
              </template>
              <span v-else class="tiny" style="color: var(--muted)">none</span>
            </div>
            <button
              type="button"
              class="btn btn--text"
              :disabled="busy"
              @click="previewInput?.click()"
            >
              {{ previewPath ? 'replace' : 'upload' }}
            </button>
            <button
              v-if="previewPath"
              type="button"
              class="btn btn--text"
              :disabled="busy"
              @click="previewPath = ''"
            >
              remove
            </button>
          </div>
          <input
            ref="previewInput"
            type="file"
            accept="image/*,video/*"
            hidden
            @change="onPreviewPicked"
          />
        </div>

        <label class="emodal__featured">
          <input v-model="form.featured" type="checkbox" />
          <span class="tiny">featured</span>
        </label>

        <p v-if="error" class="tiny">{{ error }}</p>

        <div class="emodal__foot">
          <button type="button" class="btn btn--text" :disabled="busy" @click="emit('close')">
            cancel
          </button>
          <button type="button" class="btn" :disabled="busy" @click="save">
            {{ busy ? 'saving...' : 'save' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.emodal {
  position: fixed;
  inset: 0;
  z-index: 600;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: 6vh 1rem 2rem;
  overflow-y: auto;
}

.emodal__panel {
  width: min(100%, 44rem);
  background: var(--background);
  color: var(--color);
  border: 1px solid var(--line);
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  transition:
    background-color 0.3s var(--transition-smooth),
    color 0.3s var(--transition-smooth);
}

.emodal__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: var(--muted);
}

.emodal__langs {
  display: flex;
  gap: 1rem;
}

.emodal__lang {
  font-size: 0.75rem;
  letter-spacing: 0.08em;
  text-transform: lowercase;
  color: var(--muted);
  padding: 0.2em 0;
  transition: color 0.2s var(--transition-smooth);
}

.emodal__lang:hover {
  color: var(--color);
}

.emodal__lang--active {
  color: var(--color);
  text-decoration: underline;
  text-underline-offset: 3px;
}

.emodal__close {
  font-size: 1.1rem;
  line-height: 1;
  color: var(--color);
}

.emodal__thumb {
  width: 100%;
  aspect-ratio: 16 / 9;
  background: var(--surface);
  overflow: hidden;
}

.emodal__thumb :deep(.media-thumb__label) {
  font-size: 0.7rem;
  color: var(--muted);
}

.emodal__preview {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.emodal__preview-row {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.emodal__preview-thumb {
  width: 120px;
  height: 68px;
  background: var(--surface);
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.emodal__preview-thumb img,
.emodal__preview-thumb video {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.emodal__featured {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  color: var(--muted);
  cursor: pointer;
}

.emodal__featured input {
  accent-color: var(--color);
  margin: 0;
  cursor: pointer;
}

.emodal__foot {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  align-items: center;
}
</style>

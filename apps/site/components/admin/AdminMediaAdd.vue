<script setup lang="ts">
const props = defineProps<{
  busy: boolean
  runMutation: (fn: () => Promise<unknown>) => Promise<boolean>
}>()
const emit = defineEmits<{ changed: []; failed: [message: string] }>()

const api = useApi()

/* --- uploads (add photo / add video) --- */

const photoInput = ref<HTMLInputElement | null>(null)
const videoInput = ref<HTMLInputElement | null>(null)
const uploading = ref(false)

async function uploadAndCreate(file: File, type: 'photo' | 'video') {
  uploading.value = true
  try {
    // thumb is the blurred placeholder path ('' when the api could not make one)
    const { path, thumb, width, height } = await uploadFile(file, api)
    await api('/api/media', {
      method: 'POST',
      body: { type, source: 'upload', path, thumb, width, height },
    })
    emit('changed')
  } catch (e) {
    emit('failed', e instanceof ApiError && e.status === 429 ? e.message : 'upload failed')
  } finally {
    uploading.value = false
  }
}

function onFilePicked(e: Event, type: 'photo' | 'video') {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (file) void uploadAndCreate(file, type)
  input.value = ''
}

/* --- external video (youtube / vimeo) --- */

const videoTab = ref<'upload' | 'youtube' | 'vimeo'>('upload')
const videoUrl = ref('')
const videoUrlError = ref('')

function selectTab(tab: 'upload' | 'youtube' | 'vimeo') {
  videoTab.value = tab
  videoUrlError.value = ''
}

async function addExternalVideo() {
  const tab = videoTab.value
  if (tab === 'upload') return
  videoUrlError.value = ''
  const parsed = parseVideoUrl(videoUrl.value, tab)
  if (!parsed) {
    videoUrlError.value = `unrecognized ${tab} url`
    return
  }
  const ok = await props.runMutation(() =>
    api('/api/media', {
      method: 'POST',
      body: { type: 'video', source: parsed.source, external_id: parsed.externalId },
    }),
  )
  if (ok) videoUrl.value = ''
}
</script>

<template>
  <section class="admin__add" aria-label="add media">
    <div class="admin__add-group">
      <span class="tiny admin__add-label">photo</span>
      <button type="button" class="btn" :disabled="uploading" @click="photoInput?.click()">
        {{ uploading ? 'uploading...' : 'add photo' }}
      </button>
      <input
        ref="photoInput"
        type="file"
        accept="image/*"
        hidden
        @change="onFilePicked($event, 'photo')"
      />
    </div>

    <div class="admin__add-group">
      <span class="tiny admin__add-label">video</span>
      <div class="admin__add-video-row">
        <div class="seg">
          <button
            v-for="tab in ['upload', 'youtube', 'vimeo'] as const"
            :key="tab"
            type="button"
            class="seg__btn"
            :class="{ 'seg__btn--active': videoTab === tab }"
            @click="selectTab(tab)"
          >
            {{ tab }}
          </button>
        </div>

        <template v-if="videoTab === 'upload'">
          <button type="button" class="btn" :disabled="uploading" @click="videoInput?.click()">
            {{ uploading ? 'uploading...' : 'add video' }}
          </button>
          <input
            ref="videoInput"
            type="file"
            accept="video/*"
            hidden
            @change="onFilePicked($event, 'video')"
          />
        </template>

        <form v-else class="admin__url-row" @submit.prevent="addExternalVideo">
          <input
            v-model="videoUrl"
            class="admin__url"
            type="text"
            inputmode="url"
            :placeholder="
              videoTab === 'youtube'
                ? 'https://www.youtube.com/watch?v=...'
                : 'https://vimeo.com/...'
            "
          />
          <button type="submit" class="btn" :disabled="busy">add video</button>
        </form>
      </div>
      <p v-if="videoUrlError" class="tiny">{{ videoUrlError }}</p>
    </div>
  </section>
</template>

<style scoped>
.admin__add {
  display: flex;
  flex-wrap: wrap;
  gap: 2.5rem;
  padding: 1.5rem 0 2rem;
  border-top: 1px solid var(--surface);
}

.admin__add-group {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.75rem;
}

.admin__add-video-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 1.5rem;
}

.admin__add-label {
  color: var(--muted);
}

.admin__url-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.admin__url {
  width: min(360px, 60vw);
  border: none;
  border-bottom: 1px solid var(--line);
  border-radius: 0;
  padding: 0.5em 0;
  font-size: 0.8rem;
  letter-spacing: 0.04em;
  background: transparent;
  outline: none;
  transition: border-color 0.2s var(--transition-smooth);
}

.admin__url:focus {
  border-bottom-color: var(--color);
}
</style>

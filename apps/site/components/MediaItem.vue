<script setup lang="ts">
import type { MediaItem } from '~/types/media'

const props = withDefaults(defineProps<{ item: MediaItem; fixed?: boolean }>(), { fixed: false })
const emit = defineEmits<{ open: []; measure: [id: number, width: number, height: number] }>()

/* markdown description (per-locale), rendered + sanitized client-side only */
const { pick, locale } = useI18n()
const descHtml = ref('')

function renderDesc() {
  const d = pick(props.item.translations?.description, props.item.description)?.trim()
  descHtml.value = d ? renderMarkdown(d) : ''
}

onMounted(renderDesc)
watch(locale, renderDesc)

/* --- photos: tiny blurred placeholder until the full image loads --- */

const thumbSrc = computed(() => props.item.path.replace(/\.[a-z0-9]+$/i, '.thumb.jpg'))
const thumbOk = ref(true)
const loaded = ref(false)

const thumbImg = ref<HTMLImageElement | null>(null)
const fullImg = ref<HTMLImageElement | null>(null)

onMounted(() => {
  // images may finish loading before hydration attaches @load/@error
  if (fullImg.value?.complete && fullImg.value.naturalWidth > 0) {
    loaded.value = true
    emit('measure', props.item.id, fullImg.value.naturalWidth, fullImg.value.naturalHeight)
  }
  if (thumbImg.value?.complete && thumbImg.value.naturalWidth === 0) thumbOk.value = false
})

function onPhotoLoad(e: Event) {
  loaded.value = true
  const img = e.target as HTMLImageElement
  emit('measure', props.item.id, img.naturalWidth, img.naturalHeight)
}

/* --- videos: local looped preview instead of external embeds --- */

const preview = computed(() =>
  props.item.type === 'video' && props.item.preview_path ? props.item.preview_path : null,
)
const previewIsVideo = computed(() => (preview.value ? isVideoPath(preview.value) : false))

const embedUrl = computed(() => {
  if (preview.value) return null
  return buildEmbedUrl(props.item, 'grid')
})

/*
 * Youtube embeds flash their own chrome (title bar, channel info) for the
 * first seconds of autoplay and nothing in the iframe api can suppress it.
 * Keep the iframe behind a skeleton until that transient ui has faded.
 * Vimeo's background mode has no such chrome — show it immediately.
 * Youtube also re-flashes chrome when the browser tab regains focus, so we
 * briefly re-hide on visibilitychange.
 */
const embedVisible = ref(props.item.source !== 'youtube')
let embedTimer: ReturnType<typeof setTimeout> | undefined

function scheduleEmbedShow() {
  clearTimeout(embedTimer)
  embedTimer = setTimeout(() => {
    embedVisible.value = true
  }, 4500)
}

function onVisibility() {
  if (!embedUrl.value || props.item.source !== 'youtube') return
  if (document.visibilityState === 'visible') {
    embedVisible.value = false
    scheduleEmbedShow()
  }
}

onMounted(() => {
  if (!embedUrl.value) return
  if (props.item.source === 'youtube') {
    scheduleEmbedShow()
    document.addEventListener('visibilitychange', onVisibility)
  }
})

onBeforeUnmount(() => {
  clearTimeout(embedTimer)
  document.removeEventListener('visibilitychange', onVisibility)
})
</script>

<template>
  <figure
    class="media-item"
    :class="{
      'media-item--video': item.type === 'video',
      'media-item--photo': item.type === 'photo',
      'media-item--fixed': fixed,
    }"
  >
    <div
      class="media-item__frame"
      :class="{ 'media-item__frame--waiting': embedUrl && !embedVisible }"
    >
      <template v-if="item.type === 'photo'">
        <img
          v-if="thumbOk"
          ref="thumbImg"
          class="media-item__thumb"
          :src="thumbSrc"
          alt=""
          aria-hidden="true"
          @error="thumbOk = false"
        />
        <img
          ref="fullImg"
          class="media-item__media"
          :class="{
            'media-item__media--over': thumbOk,
            'media-item__media--loaded': loaded || !thumbOk,
          }"
          :src="item.path"
          :alt="item.title || 'photo'"
          loading="lazy"
          @load="onPhotoLoad"
        />
      </template>

      <video
        v-else-if="preview && previewIsVideo"
        class="media-item__media"
        :src="preview"
        autoplay
        muted
        loop
        playsinline
        preload="auto"
      />

      <img
        v-else-if="preview"
        class="media-item__media"
        :src="preview"
        :alt="item.title || 'video preview'"
        loading="lazy"
      />

      <video
        v-else-if="item.source === 'upload'"
        class="media-item__media"
        :src="item.path"
        autoplay
        muted
        loop
        playsinline
        preload="metadata"
      />

      <iframe
        v-else-if="embedUrl"
        class="media-item__media media-item__embed"
        :class="{ 'media-item__embed--visible': embedVisible }"
        :src="embedUrl"
        :title="item.title || 'video'"
        loading="lazy"
        allow="autoplay; fullscreen; picture-in-picture; encrypted-media"
        tabindex="-1"
      />
    </div>

    <button
      v-if="item.type !== 'photo' || loaded"
      type="button"
      class="media-item__click"
      :class="{ 'media-item__click--zoom': item.type === 'photo' }"
      :aria-label="item.type === 'photo' ? 'open photo' : 'play video with sound'"
      @click="$emit('open')"
    />

    <div v-if="descHtml" class="media-item__desc">
      <div class="media-item__desc-inner" v-html="descHtml" />
    </div>

    <div class="media-item__icons">
      <a
        v-if="item.instagram_url"
        class="media-item__icon"
        :href="item.instagram_url"
        target="_blank"
        rel="noopener"
        aria-label="view on instagram"
      >
        <InstaIcon />
      </a>
      <a
        v-if="item.youtube_url"
        class="media-item__icon"
        :href="item.youtube_url"
        target="_blank"
        rel="noopener"
        aria-label="watch on youtube"
      >
        <svg
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="1.5"
          aria-hidden="true"
        >
          <rect x="2.5" y="5.5" width="19" height="13" rx="3" />
          <path d="M10.2 9.3v5.4l4.8-2.7z" fill="currentColor" stroke="none" />
        </svg>
      </a>
      <a
        v-if="item.vimeo_url"
        class="media-item__icon"
        :href="item.vimeo_url"
        target="_blank"
        rel="noopener"
        aria-label="watch on vimeo"
      >
        <svg
          width="15"
          height="15"
          viewBox="0 0 1000 903.7354126"
          fill="currentColor"
          aria-hidden="true"
        >
          <path
            d="M853.98,0C668.22,0,580.27,240.17,580.27,240.17l14.38,20.86c16.49-9.87,36.22-20.86,59.06-20.86c24.8,0,45.38,18.32,45.38,66.67c0,80.06-91.47,276.81-173.64,276.81c-36.5,0-52.85-59.2-76.53-166.17c-8.03-31.29-22.13-111.49-42.57-240.31C387.32,57.79,373.78,10.01,306.41,10.01C151.8,10.01,0,225.37,0,225.37l21.99,45.81c16.49-9.87,36.08-21,64.83-21c54.83,0,60.32,60.89,91.75,173.64c28.33,103.45,56.24,205.08,84.99,310.36c28.89,105.29,88.23,169.56,161.24,169.56c109.51,0,214.94-92.88,358.14-278.22c66.67-85.41,136.86-200.14,173.64-280.9C993.8,263.14,1000,193.8,1000,161.95C1000,69.49,963.5,0,853.98,0z"
          />
        </svg>
      </a>
    </div>
  </figure>
</template>

<style scoped>
.media-item {
  position: relative;
  display: block;
  width: min(100%, 62rem);
  margin: 0 auto;
}

.media-item__frame {
  position: relative;
  width: 100%;
  overflow: hidden;
  background: var(--surface);
}

.media-item--video .media-item__frame {
  aspect-ratio: 16 / 9;
}

/* skeleton shimmer while an external embed hides its first-seconds chrome */

.media-item__frame--waiting {
  background: linear-gradient(110deg, var(--surface) 30%, var(--line) 45%, var(--surface) 60%);
  background-size: 200% 100%;
  animation: skeletonShimmer 1.4s linear infinite;
}

@keyframes skeletonShimmer {
  from {
    background-position: 200% 0;
  }

  to {
    background-position: -200% 0;
  }
}

/* fixed mode: the parent cell dictates width & height (justified grid) */

.media-item--fixed {
  width: 100%;
  height: 100%;
  margin: 0;
}

.media-item--fixed .media-item__frame {
  position: absolute;
  inset: 0;
  aspect-ratio: auto;
}

.media-item--fixed .media-item__thumb {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.media-item--fixed .media-item__media {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.media-item__thumb {
  display: block;
  width: 100%;
  height: auto;
  filter: blur(14px);
  transform: scale(1.04);
}

.media-item__media {
  display: block;
  width: 100%;
  height: auto;
  border: none;
  pointer-events: none;
}

.media-item__media--over {
  position: absolute;
  inset: 0;
  height: 100%;
  object-fit: cover;
  opacity: 0;
  transition: opacity 0.5s var(--transition-smooth);
}

.media-item__media--over.media-item__media--loaded {
  opacity: 1;
}

.media-item--video .media-item__media {
  position: absolute;
  inset: 0;
  height: 100%;
  object-fit: cover;
}

.media-item__embed {
  opacity: 0;
  transition: opacity 0.6s var(--transition-smooth);
}

.media-item__embed--visible {
  opacity: 1;
}

.media-item__click {
  position: absolute;
  inset: 0;
  z-index: 1;
  cursor: pointer;
}

.media-item__click--zoom {
  cursor: zoom-in;
}

.media-item__icons {
  position: absolute;
  top: 12px;
  right: 12px;
  z-index: 3;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.media-item__icon {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--background);
  color: var(--color);
  display: flex;
  align-items: center;
  justify-content: center;
  transition:
    background-color 0.3s var(--transition-smooth),
    color 0.3s var(--transition-smooth);
}

/* markdown description shown on hover; links stay clickable above the
   click layer (pointer-events none on the overlay, auto on the text) */

.media-item__desc {
  position: absolute;
  inset: 0;
  z-index: 2;
  display: flex;
  align-items: flex-end;
  padding: 1.25rem;
  background: rgba(0, 0, 0, 0.32);
  color: #fff;
  opacity: 0;
  transition: opacity 0.3s var(--transition-smooth);
  pointer-events: none;
}

.media-item:hover .media-item__desc,
.media-item:focus-within .media-item__desc {
  opacity: 1;
}

.media-item__desc-inner {
  pointer-events: auto;
  max-height: 100%;
  overflow-y: auto;
  font-size: 0.75rem;
  line-height: 1.6;
  letter-spacing: 0.04em;
}

.media-item__desc-inner :deep(p) {
  margin: 0 0 0.75rem;
}

.media-item__desc-inner :deep(p:last-child) {
  margin-bottom: 0;
}

.media-item__desc-inner :deep(a) {
  color: #fff;
  text-decoration: underline;
  text-underline-offset: 2px;
}

.media-item__desc-inner :deep(ul),
.media-item__desc-inner :deep(ol) {
  padding-left: 1.2em;
  margin: 0 0 0.75rem;
  list-style: disc;
}

.media-item__desc-inner :deep(ol) {
  list-style: decimal;
}

.media-item__desc-inner :deep(strong) {
  font-weight: 600;
}

.media-item__desc-inner :deep(code) {
  font-family: monospace;
  font-size: 0.9em;
}
</style>

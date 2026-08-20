<script setup lang="ts">
import type { MediaItem } from '~/types/media'

const props = defineProps<{
  items: MediaItem[]
  index: number | null
}>()

const emit = defineEmits<{
  close: []
  'update:index': [index: number]
}>()

const current = computed(() => (props.index === null ? null : (props.items[props.index] ?? null)))

const many = computed(() => props.items.length > 1)

function prev() {
  if (props.index === null || !props.items.length) return
  emit('update:index', (props.index - 1 + props.items.length) % props.items.length)
}

function next() {
  if (props.index === null || !props.items.length) return
  emit('update:index', (props.index + 1) % props.items.length)
}

/** full player embed (sound adjustable) used inside the overlay */
const playerUrl = computed(() => (current.value ? buildEmbedUrl(current.value, 'player') : null))

function onKeydown(e: KeyboardEvent) {
  if (props.index === null) return
  if (e.key === 'Escape') emit('close')
  else if (e.key === 'ArrowLeft') prev()
  else if (e.key === 'ArrowRight') next()
}

useBodyScrollLock(computed(() => props.index !== null))

onMounted(() => window.addEventListener('keydown', onKeydown))
onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown))
</script>

<template>
  <Teleport to="body">
    <Transition name="lightbox">
      <div
        v-if="current"
        class="lightbox"
        role="dialog"
        aria-modal="true"
        :aria-label="current.title || 'media viewer'"
        @click.self="emit('close')"
      >
        <button type="button" class="lightbox__close" aria-label="close" @click="emit('close')">
          <svg
            width="18"
            height="18"
            viewBox="0 0 18 18"
            fill="none"
            stroke="currentColor"
            stroke-width="1.5"
            aria-hidden="true"
          >
            <path d="M2 2l14 14M16 2L2 16" />
          </svg>
        </button>

        <button
          v-if="many"
          type="button"
          class="lightbox__chevron lightbox__chevron--prev"
          aria-label="previous"
          @click="prev"
        >
          <svg
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="1.5"
            aria-hidden="true"
          >
            <path d="M15 4l-8 8 8 8" />
          </svg>
        </button>

        <div class="lightbox__stage" @click.self="emit('close')">
          <img
            v-if="current.type === 'photo'"
            :key="`img-${current.id}`"
            class="lightbox__photo"
            :src="current.path"
            :alt="current.title || 'photo'"
          />

          <div v-else :key="`vid-${current.id}`" class="lightbox__video">
            <video
              v-if="current.source === 'upload'"
              :src="current.path"
              controls
              autoplay
              playsinline
            />
            <iframe
              v-else-if="playerUrl"
              :src="playerUrl"
              :title="current.title || 'video'"
              allow="autoplay; fullscreen; picture-in-picture; encrypted-media"
              allowfullscreen
            />
          </div>
        </div>

        <button
          v-if="many"
          type="button"
          class="lightbox__chevron lightbox__chevron--next"
          aria-label="next"
          @click="next"
        >
          <svg
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="1.5"
            aria-hidden="true"
          >
            <path d="M9 4l8 8-8 8" />
          </svg>
        </button>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.lightbox {
  position: fixed;
  inset: 0;
  z-index: 500;
  background: rgba(0, 0, 0, 0.92);
  display: flex;
  align-items: center;
  justify-content: center;
}

.lightbox__stage {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  padding: 4rem;
}

.lightbox__photo {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.lightbox__video {
  width: min(100%, calc((100vh - 8rem) * 16 / 9));
  aspect-ratio: 16 / 9;
}

.lightbox__video video,
.lightbox__video iframe {
  width: 100%;
  height: 100%;
  border: none;
  background: #000;
}

.lightbox__close {
  position: absolute;
  top: 1rem;
  right: 1rem;
  z-index: 2;
  width: 44px;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  opacity: 0.7;
  transition: opacity 0.2s var(--transition-smooth);
}

.lightbox__chevron {
  position: absolute;
  top: 50%;
  z-index: 2;
  width: 56px;
  height: 56px;
  margin-top: -28px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  opacity: 0.7;
  transition: opacity 0.2s var(--transition-smooth);
}

.lightbox__chevron--prev {
  left: 0.5rem;
}

.lightbox__chevron--next {
  right: 0.5rem;
}

.lightbox__close:hover,
.lightbox__chevron:hover {
  opacity: 1;
}

@media (max-width: 767px) {
  .lightbox__stage {
    padding: 3rem 0.5rem;
  }
}

.lightbox-enter-active,
.lightbox-leave-active {
  transition: opacity 0.25s var(--transition-smooth);
}

.lightbox-enter-from,
.lightbox-leave-to {
  opacity: 0;
}
</style>

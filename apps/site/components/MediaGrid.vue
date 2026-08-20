<script setup lang="ts">
import type { MediaItem } from '~/types/media'

defineProps<{ items: MediaItem[] }>()

const root = ref<HTMLElement | null>(null)

/* opacity-only reveal, staggered above the fold */
useRevealOnScroll(root, '.media-item')

/* lightbox */
const lightboxIndex = ref<number | null>(null)

// vimeo grid embeds pause while the lightbox (or another tab) covers them
watch(lightboxIndex, (v: number | null) => {
  if (v === null) nextTick(() => resumeVimeoEmbeds(root.value))
})

function onVisible() {
  if (document.visibilityState === 'visible') resumeVimeoEmbeds(root.value)
}

onMounted(() => document.addEventListener('visibilitychange', onVisible))
onBeforeUnmount(() => document.removeEventListener('visibilitychange', onVisible))
</script>

<template>
  <div ref="root" class="media-grid">
    <MediaItem
      v-for="(item, index) in items"
      :key="item.id"
      :item="item"
      @open="lightboxIndex = index"
    />

    <MediaLightbox
      :items="items"
      :index="lightboxIndex"
      @close="lightboxIndex = null"
      @update:index="lightboxIndex = $event"
    />
  </div>
</template>

<style scoped>
.media-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--gap);
  width: 100%;
}

.media-grid :deep(.media-item) {
  opacity: 0;
  transition: opacity 0.6s var(--transition-smooth);
}

.media-grid :deep(.media-item.is-visible) {
  opacity: 1;
}
</style>

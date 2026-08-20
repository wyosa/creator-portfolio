<script setup lang="ts">
import type { MediaItem } from '~/types/media'

/*
 * Small media preview for admin rows/modals. `preview` allows the custom
 * looped preview file to stand in for external embeds (media list); `label`
 * overrides the fallback plaque text.
 */
defineProps<{ item: MediaItem; preview?: boolean; label?: string }>()
</script>

<template>
  <img v-if="item.type === 'photo'" :src="item.path" :alt="item.title" loading="lazy" />
  <video
    v-else-if="item.source === 'upload'"
    :src="item.path"
    muted
    playsinline
    preload="metadata"
  />
  <template v-else-if="preview && item.preview_path">
    <img
      v-if="!isVideoPath(item.preview_path)"
      :src="item.preview_path"
      :alt="item.title"
      loading="lazy"
    />
    <video v-else :src="item.preview_path" muted playsinline preload="metadata" />
  </template>
  <img
    v-else-if="item.source === 'youtube'"
    :src="youtubeThumb(item.external_id)"
    :alt="item.title"
    loading="lazy"
  />
  <span v-else class="media-thumb__label">{{ label ?? 'vimeo' }}</span>
</template>

<style scoped>
img,
video {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

/* theme colors are set by the parent via :deep(.media-thumb__label) */
.media-thumb__label {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  letter-spacing: 0.12em;
  text-transform: lowercase;
}
</style>

import type { Ref } from 'vue'

/**
 * Shared lightbox wiring for the media grids: the open index plus resuming
 * vimeo grid embeds, which auto-pause while the lightbox (or another tab)
 * covers them.
 */
export function useLightbox(root: Ref<HTMLElement | null>) {
  const lightboxIndex = ref<number | null>(null)

  watch(lightboxIndex, (v: number | null) => {
    if (v === null) nextTick(() => resumeVimeoEmbeds(root.value))
  })

  function onVisible() {
    if (document.visibilityState === 'visible') resumeVimeoEmbeds(root.value)
  }

  onMounted(() => document.addEventListener('visibilitychange', onVisible))
  onBeforeUnmount(() => document.removeEventListener('visibilitychange', onVisible))

  return { lightboxIndex }
}

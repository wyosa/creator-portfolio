import type { Ref } from 'vue'

/** lock body scrolling while `open` is true; always restore on unmount */
export function useBodyScrollLock(open: Ref<boolean>) {
  watch(open, (v: boolean) => {
    if (import.meta.client) document.body.style.overflow = v ? 'hidden' : ''
  })

  onBeforeUnmount(() => {
    if (import.meta.client) document.body.style.overflow = ''
  })
}

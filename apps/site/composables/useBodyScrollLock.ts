import { onBeforeUnmount, onMounted, watch, type Ref } from 'vue'

let lockCount = 0
let savedOverflow: string | null = null

/** lock body scroll; concurrent locks (menu + lightbox) stack via a counter */
export function lockBodyScroll() {
  if (typeof document === 'undefined') return
  if (lockCount === 0) {
    savedOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
  }
  lockCount++
}

/** release one lock; the original overflow is restored with the last release */
export function unlockBodyScroll() {
  if (typeof document === 'undefined' || lockCount === 0) return
  lockCount--
  if (lockCount === 0) {
    document.body.style.overflow = savedOverflow ?? ''
    savedOverflow = null
  }
}

/** lock body scrolling while `open` is true; always restore on unmount */
export function useBodyScrollLock(open: Ref<boolean>) {
  watch(open, (v: boolean) => (v ? lockBodyScroll() : unlockBodyScroll()))

  onMounted(() => {
    if (open.value) lockBodyScroll()
  })

  onBeforeUnmount(() => {
    if (open.value) unlockBodyScroll()
  })
}

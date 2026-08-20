import { nextTick, onBeforeUnmount, onMounted, watch, type Ref } from 'vue'

const FOCUSABLE = [
  'a[href]',
  'button:not([disabled])',
  'input:not([disabled])',
  'textarea:not([disabled])',
  'select:not([disabled])',
  '[tabindex]:not([tabindex="-1"])',
].join(', ')

/**
 * Focus trap for modal overlays: on activation focus moves into the
 * container, Tab/Shift+Tab cycle inside it, and on deactivation focus
 * returns to the element that had it before.
 */
export function useFocusTrap(container: Ref<HTMLElement | null>, active: Ref<boolean>) {
  let previous: HTMLElement | null = null

  function focusFirst() {
    const el = container.value
    if (!el) return
    const target = el.querySelector<HTMLElement>(FOCUSABLE) ?? el
    if (target === el && !el.hasAttribute('tabindex')) el.setAttribute('tabindex', '-1')
    target.focus()
  }

  function onKeydown(e: KeyboardEvent) {
    const el = container.value
    if (e.key !== 'Tab' || !el) return
    const items = Array.from(el.querySelectorAll<HTMLElement>(FOCUSABLE))
    const first = items[0]
    const last = items[items.length - 1]
    if (!first || !last) {
      e.preventDefault()
      return
    }
    const current = document.activeElement
    const inside = current instanceof HTMLElement && el.contains(current)
    if (e.shiftKey && (!inside || current === first)) {
      e.preventDefault()
      last.focus()
    } else if (!e.shiftKey && (!inside || current === last)) {
      e.preventDefault()
      first.focus()
    }
  }

  function activate() {
    if (typeof document === 'undefined') return
    previous = document.activeElement as HTMLElement | null
    document.addEventListener('keydown', onKeydown, true)
    void nextTick(focusFirst)
  }

  function deactivate() {
    if (typeof document === 'undefined') return
    document.removeEventListener('keydown', onKeydown, true)
    previous?.focus()
    previous = null
  }

  onMounted(() => {
    if (active.value) activate()
  })

  watch(active, (v) => {
    if (v) activate()
    else deactivate()
  })

  onBeforeUnmount(() => {
    if (active.value) deactivate()
  })
}

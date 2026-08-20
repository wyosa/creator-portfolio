import type { Ref } from 'vue'

/*
 * Reveal on scroll: opacity-only fade. Above-the-fold elements fade in on
 * load with a small stagger, the rest fade in when scrolled into view via an
 * IntersectionObserver. Call the returned observe() again (after nextTick)
 * whenever the set of matched elements changes.
 */
export function useRevealOnScroll(root: Ref<HTMLElement | null>, selector: string) {
  let observer: IntersectionObserver | null = null

  function observe() {
    if (!root.value) return
    const els = Array.from(root.value.querySelectorAll<HTMLElement>(selector))
    if (!els.length) return

    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      els.forEach((el) => el.classList.add('is-visible'))
      return
    }

    const vh = window.innerHeight
    let delay = 0

    for (const el of els) {
      if (el.classList.contains('is-visible')) continue
      if (el.getBoundingClientRect().top < vh) {
        el.style.transitionDelay = `${delay}s`
        delay += 0.08
        el.addEventListener(
          'transitionend',
          () => {
            el.style.transitionDelay = ''
          },
          { once: true },
        )
        requestAnimationFrame(() => el.classList.add('is-visible'))
      } else {
        observer?.observe(el)
      }
    }
  }

  onMounted(() => {
    observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (!entry.isIntersecting) continue
          ;(entry.target as HTMLElement).classList.add('is-visible')
          observer?.unobserve(entry.target)
        }
      },
      { rootMargin: '0px 0px -8% 0px' },
    )
    observe()
  })

  onBeforeUnmount(() => {
    observer?.disconnect()
    observer = null
  })

  return { observe }
}

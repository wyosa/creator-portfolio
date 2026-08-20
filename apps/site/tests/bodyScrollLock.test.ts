import { mount } from '@vue/test-utils'
import { defineComponent, h, nextTick, ref } from 'vue'
import { describe, expect, it } from 'vitest'
import { useBodyScrollLock } from '~/composables/useBodyScrollLock'

function locker(open = ref(true)) {
  return defineComponent({
    setup() {
      useBodyScrollLock(open)
      return () => h('div')
    },
  })
}

describe('useBodyScrollLock', () => {
  it('stacks nested locks and restores the original overflow after the last one', () => {
    document.body.style.overflow = 'auto'

    const a = mount(locker())
    const b = mount(locker())
    expect(document.body.style.overflow).toBe('hidden')

    a.unmount()
    expect(document.body.style.overflow).toBe('hidden')

    b.unmount()
    expect(document.body.style.overflow).toBe('auto')
  })

  it('locks on mount when open and unlocks when open flips to false', async () => {
    document.body.style.overflow = ''
    const open = ref(true)
    const wrapper = mount(locker(open))
    expect(document.body.style.overflow).toBe('hidden')

    open.value = false
    await nextTick()
    expect(document.body.style.overflow).toBe('')

    wrapper.unmount()
    expect(document.body.style.overflow).toBe('')
  })
})

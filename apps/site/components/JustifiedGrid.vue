<script setup lang="ts">
import type { MediaItem } from '~/types/media'

/*
 * Justified ("masonry rows") layout: every row fills the full width, each
 * item keeps its own aspect ratio — no holes, no cropping beyond the row
 * height fit. Aspect ratios come from the api (photo pixel size), videos
 * are 16:9; items without stored dimensions are corrected on image load.
 */

const props = defineProps<{ items: MediaItem[] }>()

const GAP = 4

const root = ref<HTMLElement | null>(null)
const containerWidth = ref(1280)

/* measured-at-runtime ratios for items missing width/height in the db */
const measured = ref<Record<number, number>>({})

function aspect(item: MediaItem): number {
  const m = measured.value[item.id]
  if (m) return m
  if (item.width > 0 && item.height > 0) return item.width / item.height
  return item.type === 'video' ? 16 / 9 : 3 / 2
}

function onMeasure(id: number, width: number, height: number) {
  if (!width || !height) return
  const item = props.items.find((i) => i.id === id)
  if (item && item.width > 0 && item.height > 0) return // db ratio is authoritative
  const r = width / height
  if (Math.abs((measured.value[id] ?? 0) - r) < 0.001) return
  measured.value = { ...measured.value, [id]: r }
}

interface Row {
  height: number
  items: MediaItem[]
}

const rows = computed<Row[]>(() => {
  const W = containerWidth.value
  const target = W < 768 ? 230 : 420

  const out: Row[] = []
  let row: MediaItem[] = []
  let sum = 0

  const flush = (isLast: boolean) => {
    if (!row.length || sum <= 0) return
    let h = (W - (row.length - 1) * GAP) / sum
    if (isLast) h = Math.min(h, target) // keep the trailing row calm
    out.push({ height: h, items: row })
    row = []
    sum = 0
  }

  for (const item of props.items) {
    row.push(item)
    sum += aspect(item)
    const h = (W - (row.length - 1) * GAP) / sum
    if (h <= target * 1.35) flush(false)
  }
  flush(true)
  return out
})

/* lightbox over the flat item list */
const { lightboxIndex } = useLightbox(root)

function flatIndex(item: MediaItem): number {
  return props.items.indexOf(item)
}

/* container width tracking */
let ro: ResizeObserver | null = null

/* opacity-only reveal, same behavior as the classic grid */
const { observe: observeRows } = useRevealOnScroll(root, '.jgrid__row')

watch(rows, () => nextTick(observeRows))

onMounted(() => {
  if (root.value) {
    containerWidth.value = root.value.clientWidth
    ro = new ResizeObserver((entries) => {
      const w = entries[0]?.contentRect.width
      if (w && Math.abs(w - containerWidth.value) > 1) containerWidth.value = w
    })
    ro.observe(root.value)
  }
})

onBeforeUnmount(() => {
  ro?.disconnect()
  ro = null
})
</script>

<template>
  <div ref="root" class="jgrid">
    <div
      v-for="row in rows"
      :key="row.items[0].id"
      class="jgrid__row"
      :style="{ height: `${row.height}px` }"
    >
      <div
        v-for="item in row.items"
        :key="item.id"
        class="jgrid__cell"
        :style="{ width: `${row.height * aspect(item)}px` }"
      >
        <MediaItem
          :item="item"
          fixed
          @open="lightboxIndex = flatIndex(item)"
          @measure="onMeasure"
        />
      </div>
    </div>

    <MediaLightbox
      :items="items"
      :index="lightboxIndex"
      @close="lightboxIndex = null"
      @update:index="lightboxIndex = $event"
    />
  </div>
</template>

<style scoped>
.jgrid {
  width: 100%;
}

.jgrid__row {
  display: flex;
  gap: var(--gap);
  margin-bottom: var(--gap);
  opacity: 0;
  transition: opacity 0.6s var(--transition-smooth);
}

.jgrid__row.is-visible {
  opacity: 1;
}

.jgrid__cell {
  position: relative;
  flex: 0 0 auto;
}
</style>

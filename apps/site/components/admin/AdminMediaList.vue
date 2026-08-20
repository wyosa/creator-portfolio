<script setup lang="ts">
import type { MediaItem } from '~/types/media'

defineProps<{ items: MediaItem[]; pending: boolean; busy: boolean }>()
const emit = defineEmits<{
  reorder: [from: number, to: number]
  delete: [item: MediaItem]
  edit: [item: MediaItem]
}>()

/* drag & drop reorder — the page owns the actual list mutation */
const { dragFrom, dragOver, handleFor, onDragStart, onDragOver, onDragLeave, onDragEnd, onDrop } =
  useDragReorder((from, to) => emit('reorder', from, to))
</script>

<template>
  <p v-if="pending && !items.length" class="tiny">loading...</p>
  <p v-else-if="!items.length" class="tiny">no media yet</p>

  <div v-else class="admin__list">
    <ul class="rows">
      <li
        v-for="(item, index) in items"
        :key="item.id"
        class="row"
        :class="{ 'row--over': dragOver === index && dragFrom !== index }"
        :draggable="handleFor === item.id"
        @dragstart="onDragStart(index, $event)"
        @dragover="onDragOver(index, $event)"
        @dragleave="onDragLeave"
        @drop="onDrop(index, $event)"
        @dragend="onDragEnd"
      >
        <DragHandle @grab="handleFor = item.id" @release="handleFor = null" />

        <div class="row__thumb">
          <MediaThumb :item="item" preview />
        </div>

        <span class="row__title">
          <span class="row__title-text">{{ mediaLabel(item) }}</span>
          <svg
            v-if="item.featured"
            class="row__star"
            width="13"
            height="13"
            viewBox="0 0 24 24"
            fill="currentColor"
            aria-label="featured"
          >
            <path
              d="M12 2.5l2.9 6 6.6.9-4.8 4.6 1.2 6.5L12 17.4 6.1 20.5l1.2-6.5L2.5 9.4l6.6-.9z"
            />
          </svg>
        </span>

        <button type="button" class="row__preview-btn" @click="emit('edit', item)">edit</button>

        <button type="button" class="row__delete" :disabled="busy" @click="emit('delete', item)">
          delete
        </button>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.admin__list {
  overflow-x: auto;
}

.rows {
  min-width: 560px;
  display: flex;
  flex-direction: column;
}

.row {
  display: grid;
  grid-template-columns: 24px 96px minmax(0, 1fr) auto auto;
  gap: 16px;
  align-items: center;
  padding: 10px 0;
  border-top: 1px solid var(--surface);
  transition: opacity 0.2s var(--transition-smooth);
}

.row--over {
  border-top-color: var(--color);
}

.row__thumb {
  width: 96px;
  height: 60px;
  background: var(--surface);
  overflow: hidden;
}

.row__thumb :deep(.media-thumb__label) {
  background: #111;
  color: #fff;
  font-size: 0.6rem;
}

.row__title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-width: 0;
  font-size: 0.8rem;
  letter-spacing: 0.04em;
}

.row__title-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.row__star {
  flex: 0 0 auto;
  color: var(--color);
}

.row__preview-btn {
  font-size: 0.7rem;
  letter-spacing: 0.12em;
  text-transform: lowercase;
  text-decoration: underline;
  text-underline-offset: 3px;
  color: var(--muted);
  transition: color 0.2s var(--transition-smooth);
}

.row__preview-btn:hover {
  color: var(--color);
}

.row__delete {
  font-size: 0.7rem;
  letter-spacing: 0.12em;
  text-transform: lowercase;
  text-decoration: underline;
  text-underline-offset: 3px;
  color: var(--muted);
  transition: color 0.2s var(--transition-smooth);
}

.row__delete:hover {
  color: var(--color);
}

.row__delete:disabled {
  opacity: 0.4;
  cursor: default;
}
</style>

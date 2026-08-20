<script setup lang="ts">
import type { InfoLinkDraft } from '~/types/settings'

const props = defineProps<{ links: InfoLinkDraft[] }>()

/* drag & drop reorder (same interaction as media rows) */
const { dragFrom, dragOver, handleFor, onDragStart, onDragOver, onDragLeave, onDragEnd, onDrop } =
  useDragReorder((from, to) => {
    const [moved] = props.links.splice(from, 1)
    props.links.splice(to, 0, moved)
  })

function addLink() {
  props.links.push(toLinkDraft({ label: '', url: '', enabled: true }))
}
</script>

<template>
  <div class="admin__links">
    <span class="tiny admin__add-label">info links</span>
    <div
      v-for="(link, i) in links"
      :key="link.cid"
      class="admin__link-row"
      :class="{ 'row--over': dragOver === i && dragFrom !== i }"
      :draggable="handleFor === i"
      @dragstart="onDragStart(i, $event)"
      @dragover="onDragOver(i, $event)"
      @dragleave="onDragLeave"
      @drop="onDrop(i, $event)"
      @dragend="onDragEnd"
    >
      <DragHandle @grab="handleFor = i" @release="handleFor = null" />
      <label class="row__check">
        <input v-model="link.enabled" type="checkbox" />
        <span class="tiny">show</span>
      </label>
      <input v-model="link.label" class="row__input" type="text" placeholder="label" />
      <input v-model="link.url" class="row__input" type="text" placeholder="url" />
      <button
        type="button"
        class="row__preview-clear"
        aria-label="remove link"
        @click="links.splice(i, 1)"
      >
        ×
      </button>
    </div>
    <button type="button" class="btn btn--text" @click="addLink">add link</button>
  </div>
</template>

<style scoped>
.admin__links {
  align-self: stretch;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 1.75rem;
}

.admin__add-label {
  color: var(--muted);
}

.admin__link-row {
  width: 100%;
  max-width: 40rem;
  display: grid;
  grid-template-columns: auto auto minmax(0, 140px) 1fr auto;
  gap: 0.75rem;
  align-items: center;
  border-top: 1px solid transparent;
}

.row--over {
  border-top-color: var(--color);
}

.row__check {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  color: var(--muted);
  cursor: pointer;
  white-space: nowrap;
}

.row__check input {
  accent-color: var(--color);
  margin: 0;
  cursor: pointer;
}

.row__input {
  width: 100%;
  border: none;
  border-bottom: 1px solid transparent;
  border-radius: 0;
  padding: 0.35em 0;
  font-size: 0.8rem;
  letter-spacing: 0.04em;
  background: transparent;
  outline: none;
  transition: border-color 0.2s var(--transition-smooth);
}

.row__input:hover {
  border-bottom-color: var(--line);
}

.row__input:focus {
  border-bottom-color: var(--color);
}

.row__preview-clear {
  font-size: 0.9rem;
  line-height: 1;
  color: var(--muted);
  transition: color 0.2s var(--transition-smooth);
}

.row__preview-clear:hover {
  color: var(--color);
}
</style>

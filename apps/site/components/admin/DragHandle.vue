<script setup lang="ts">
const emit = defineEmits<{ grab: []; release: [] }>()
</script>

<template>
  <button
    type="button"
    class="drag-handle"
    aria-label="drag to reorder"
    @mousedown="emit('grab')"
    @mouseup="emit('release')"
  >
    <svg width="10" height="16" viewBox="0 0 10 16" fill="currentColor" aria-hidden="true">
      <circle cx="2" cy="2" r="1.3" />
      <circle cx="8" cy="2" r="1.3" />
      <circle cx="2" cy="8" r="1.3" />
      <circle cx="8" cy="8" r="1.3" />
      <circle cx="2" cy="14" r="1.3" />
      <circle cx="8" cy="14" r="1.3" />
    </svg>
  </button>
</template>

<style scoped>
.drag-handle {
  width: 24px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--muted);
  cursor: grab;
  touch-action: none;
}
</style>

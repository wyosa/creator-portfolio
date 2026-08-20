/*
 * Shared drag & drop row-reorder mechanics (admin media list, info links).
 * Rows become draggable only while their handle is held (handleFor); the
 * drop callback receives source/target indexes and performs the actual move.
 */
export function useDragReorder(onMove: (from: number, to: number) => void) {
  const dragFrom = ref<number | null>(null)
  const dragOver = ref<number | null>(null)
  const handleFor = ref<number | null>(null)

  function onDragStart(index: number, e: DragEvent) {
    dragFrom.value = index
    if (e.dataTransfer) {
      e.dataTransfer.effectAllowed = 'move'
      e.dataTransfer.setData('text/plain', String(index))
    }
  }

  function onDragOver(index: number, e: DragEvent) {
    e.preventDefault()
    if (e.dataTransfer) e.dataTransfer.dropEffect = 'move'
    dragOver.value = index
  }

  function onDragLeave(e: DragEvent) {
    const row = e.currentTarget as HTMLElement
    if (!row.contains(e.relatedTarget as Node)) dragOver.value = null
  }

  function onDragEnd() {
    dragFrom.value = null
    dragOver.value = null
    handleFor.value = null
  }

  function onDrop(index: number, e: DragEvent) {
    e.preventDefault()
    const from = dragFrom.value
    onDragEnd()
    if (from === null || from === index) return
    onMove(from, index)
  }

  return { dragFrom, dragOver, handleFor, onDragStart, onDragOver, onDragLeave, onDragEnd, onDrop }
}

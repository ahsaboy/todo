import { ref, type Ref } from 'vue'

export function useDragSort<T>(
  items: Ref<T[]>,
  onReorder: (newOrder: T[]) => Promise<void>,
) {
  const draggedIndex = ref<number | null>(null)
  const dragOverIndex = ref<number | null>(null)

  function onDragStart(index: number, e: DragEvent) {
    draggedIndex.value = index
    if (e.dataTransfer) e.dataTransfer.effectAllowed = 'move'
  }

  function onDragOver(index: number) {
    dragOverIndex.value = index
  }

  async function onDrop(targetIndex: number) {
    const from = draggedIndex.value
    if (from === null || from === targetIndex) { onDragEnd(); return }
    const list = [...items.value]
    const [moved] = list.splice(from, 1)
    list.splice(targetIndex, 0, moved)
    try {
      await onReorder(list)
    } catch (e) {
      throw e
    } finally {
      onDragEnd()
    }
  }

  function onDragEnd() {
    draggedIndex.value = null
    dragOverIndex.value = null
  }

  return { draggedIndex, dragOverIndex, onDragStart, onDragOver, onDrop, onDragEnd }
}

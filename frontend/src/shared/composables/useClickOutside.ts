import { onMounted, onBeforeUnmount, type Ref } from 'vue'

export function useClickOutside(target: Ref<HTMLElement | null>, handler: () => void) {
  const listener = (e: MouseEvent) => {
    if (!target.value || target.value.contains(e.target as Node)) return
    handler()
  }

  onMounted(() => document.addEventListener('click', listener))
  onBeforeUnmount(() => document.removeEventListener('click', listener))
}

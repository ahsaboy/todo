<template>
  <component :is="tag" ref="el" class="motion-stagger">
    <slot />
  </component>
</template>

<script setup lang="ts">
import { ref, onMounted, onUpdated, nextTick, onBeforeUnmount } from 'vue'

withDefaults(defineProps<{
  tag?: string
}>(), {
  tag: 'div',
})

const el = ref<HTMLElement | null>(null)

function updateIndices() {
  if (!el.value) return
  const children = el.value.children
  for (let i = 0; i < children.length; i++) {
    ;(children[i] as HTMLElement).style.setProperty('--motion-index', String(i))
  }
}

let observer: MutationObserver | null = null

onMounted(() => {
  updateIndices()
  observer = new MutationObserver(updateIndices)
  if (el.value) {
    observer.observe(el.value, { childList: true })
  }
})

onUpdated(() => {
  nextTick(updateIndices)
})

onBeforeUnmount(() => {
  observer?.disconnect()
})
</script>

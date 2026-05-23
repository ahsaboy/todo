<template>
  <Transition name="collapse" @enter="onEnter" @leave="onLeave">
    <slot />
  </Transition>
</template>

<script setup lang="ts">
function onEnter(el: Element) {
  const cell = el as HTMLElement
  cell.style.display = 'grid'
  cell.style.gridTemplateRows = '0fr'
  requestAnimationFrame(() => {
    cell.style.gridTemplateRows = '1fr'
  })
  el.addEventListener('transitionend', () => {
    cell.style.display = ''
    cell.style.gridTemplateRows = ''
  }, { once: true })
}

function onLeave(el: Element, done: () => void) {
  const cell = el as HTMLElement
  cell.style.display = 'grid'
  cell.style.gridTemplateRows = '1fr'
  requestAnimationFrame(() => {
    cell.style.gridTemplateRows = '0fr'
  })
  el.addEventListener('transitionend', done, { once: true })
}
</script>

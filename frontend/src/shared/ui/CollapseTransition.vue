<template>
  <Transition name="collapse" @enter="onEnter" @leave="onLeave">
    <slot />
  </Transition>
</template>

<script setup lang="ts">
function onEnter(el: Element) {
  const cell = el as HTMLElement
  let cleaned = false
  const cleanup = () => {
    if (cleaned) return
    cleaned = true
    cell.style.display = ''
    cell.style.gridTemplateRows = ''
  }

  cell.style.display = 'grid'
  cell.style.gridTemplateRows = '0fr'
  requestAnimationFrame(() => {
    cell.style.gridTemplateRows = '1fr'
  })

  el.addEventListener('transitionend', (e) => {
    if (e.propertyName === 'grid-template-rows') cleanup()
  }, { once: true })
  el.addEventListener('transitioncancel', cleanup, { once: true })
  setTimeout(cleanup, 400)
}

function onLeave(el: Element, done: () => void) {
  const cell = el as HTMLElement
  let completed = false
  const finish = () => {
    if (completed) return
    completed = true
    done()
  }

  cell.style.display = 'grid'
  cell.style.gridTemplateRows = '1fr'
  requestAnimationFrame(() => {
    cell.style.gridTemplateRows = '0fr'
  })

  el.addEventListener('transitionend', (e) => {
    if (e.propertyName === 'grid-template-rows') finish()
  }, { once: true })
  el.addEventListener('transitioncancel', finish, { once: true })
  setTimeout(finish, 400)
}
</script>

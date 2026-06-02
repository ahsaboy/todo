<script setup lang="ts">
defineProps<{
  label: string
}>()

defineEmits<{
  click: []
}>()
</script>

<template>
  <button
    class="mobile-fab"
    :aria-label="label"
    @click="$emit('click')"
  >
    <slot />
  </button>
</template>

<style scoped>
.mobile-fab {
  display: none;
  position: fixed;
  bottom: calc(var(--bottom-nav-height, 56px) + 1rem);
  right: 1rem;
  width: 52px;
  height: 52px;
  border-radius: 50%;
  background: var(--color-primary);
  color: #fff;
  border: none;
  box-shadow: var(--shadow-panel);
  cursor: pointer;
  align-items: center;
  justify-content: center;
  z-index: 100;
  transition: transform var(--motion-duration-fast) var(--motion-ease-standard),
              opacity var(--motion-duration-fast) var(--motion-ease-standard);
}

.mobile-fab:active {
  transform: scale(0.92);
}

@media (max-width: 767px) {
  .mobile-fab {
    display: flex;
  }
}
</style>

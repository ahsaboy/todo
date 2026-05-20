<template>
  <div class="sk-block" :style="{ width, height, borderRadius }" />
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  width?: string
  height?: string
  borderRadius?: string
}>(), {
  width: '100%',
  height: '16px',
  borderRadius: '4px',
})
</script>

<style scoped>
.sk-block {
  background: linear-gradient(
    90deg,
    var(--color-surface-muted) 25%,
    var(--color-border) 37%,
    var(--color-surface-muted) 63%
  );
  background-size: 400% 100%;
  animation: sk-shimmer 1.4s ease infinite;
}

@keyframes sk-shimmer {
  0% { background-position: 100% 50%; }
  100% { background-position: 0 50%; }
}

@media (prefers-reduced-motion: reduce) {
  .sk-block {
    animation: none;
    background: var(--color-surface-muted);
  }
}
</style>

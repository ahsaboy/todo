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
    135deg,
    var(--color-surface-muted) 25%,
    var(--color-border) 37%,
    var(--color-surface-muted) 63%
  );
  background-size: 400% 100%;
  animation: sk-shimmer var(--motion-duration-shimmer, 1.4s) var(--motion-ease-standard) infinite;
}
</style>

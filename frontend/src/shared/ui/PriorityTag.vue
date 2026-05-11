<template>
  <span class="priority-tag" :class="level">
    {{ label }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  priority: number
}>()

const level = computed(() => {
  if (props.priority === 1) return 'high'
  if (props.priority === 2) return 'medium'
  return 'low'
})

const label = computed(() => {
  if (props.priority === 1) return '高'
  if (props.priority === 2) return '中'
  return '低'
})
</script>

<style scoped>
.priority-tag {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  border: 1px solid transparent;
  box-shadow: inset 0 0 0 1px transparent;
}

.priority-tag.high {
  background: color-mix(in srgb, var(--color-glow-danger) 84%, transparent);
  border-color: color-mix(in srgb, var(--color-danger) 26%, transparent);
  box-shadow: 0 0 12px color-mix(in srgb, var(--color-glow-danger) 80%, transparent);
  color: var(--color-danger);
}

.priority-tag.medium {
  background: color-mix(in srgb, var(--color-glow-warning) 84%, transparent);
  border-color: color-mix(in srgb, var(--color-warning) 24%, transparent);
  box-shadow: 0 0 12px color-mix(in srgb, var(--color-glow-warning) 78%, transparent);
  color: var(--color-warning);
}

.priority-tag.low {
  background: color-mix(in srgb, var(--color-glow-info) 78%, transparent);
  border-color: color-mix(in srgb, var(--color-info) 18%, transparent);
  box-shadow: 0 0 10px color-mix(in srgb, var(--color-glow-info) 68%, transparent);
  color: color-mix(in srgb, var(--color-info) 90%, var(--color-text) 10%);
}
</style>

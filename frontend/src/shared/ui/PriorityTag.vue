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
  padding: 1px 6px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
  flex-shrink: 0;
}

.priority-tag.high {
  background: color-mix(in srgb, var(--color-danger) 16%, transparent);
  color: var(--color-danger);
}

.priority-tag.medium {
  background: color-mix(in srgb, var(--color-warning) 16%, transparent);
  color: var(--color-warning);
}

.priority-tag.low {
  background: color-mix(in srgb, var(--color-info) 12%, transparent);
  color: var(--color-info);
}
</style>

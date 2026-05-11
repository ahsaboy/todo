<template>
  <span class="status-tag" :class="status">
    {{ label }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  completed: boolean
  overdue?: boolean
}>()

const status = computed(() => {
  if (props.overdue) return 'overdue'
  if (props.completed) return 'completed'
  return 'pending'
})

const label = computed(() => {
  if (props.overdue) return '逾期'
  if (props.completed) return '已完成'
  return '待处理'
})
</script>

<style scoped>
.status-tag {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  border: 1px solid transparent;
  box-shadow: inset 0 0 0 1px transparent;
}

.status-tag.pending {
  background: color-mix(in srgb, var(--color-glow-info) 84%, transparent);
  border-color: color-mix(in srgb, var(--color-info) 22%, transparent);
  box-shadow: 0 0 10px color-mix(in srgb, var(--color-glow-info) 72%, transparent);
  color: var(--color-info);
}

.status-tag.completed {
  background: color-mix(in srgb, var(--color-glow-success) 72%, transparent);
  border-color: color-mix(in srgb, var(--color-success) 18%, transparent);
  box-shadow: 0 0 10px color-mix(in srgb, var(--color-glow-success) 56%, transparent);
  color: color-mix(in srgb, var(--color-success) 88%, var(--color-text) 12%);
}

.status-tag.overdue {
  background: color-mix(in srgb, var(--color-glow-danger) 84%, transparent);
  border-color: color-mix(in srgb, var(--color-danger) 26%, transparent);
  box-shadow: 0 0 12px color-mix(in srgb, var(--color-glow-danger) 80%, transparent);
  color: var(--color-danger);
}
</style>

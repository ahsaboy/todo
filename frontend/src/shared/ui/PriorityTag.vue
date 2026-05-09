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
  if (props.priority === 3) return 'high'
  if (props.priority === 2) return 'medium'
  return 'low'
})

const label = computed(() => {
  if (props.priority === 3) return '高'
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
}

.priority-tag.high {
  background: #fee2e2;
  color: #dc2626;
}

.priority-tag.medium {
  background: #ffedd5;
  color: #ea580c;
}

.priority-tag.low {
  background: #e0e7ff;
  color: #4f46e5;
}
</style>

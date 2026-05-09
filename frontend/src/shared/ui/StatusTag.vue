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
}

.status-tag.pending {
  background: #f1f5f9;
  color: #475569;
}

.status-tag.completed {
  background: #dcfce7;
  color: #16a34a;
}

.status-tag.overdue {
  background: #fee2e2;
  color: #dc2626;
}
</style>

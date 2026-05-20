<template>
  <MotionStagger class="task-board">
    <TaskBoardColumn title="待处理" :tasks="pendingTasks" @cardClick="$emit('cardClick', $event)" />
    <TaskBoardColumn
      title="已完成"
      :tasks="completedTasks"
      @cardClick="$emit('cardClick', $event)"
    />
  </MotionStagger>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Task } from '@/entities/task/model'
import TaskBoardColumn from './TaskBoardColumn.vue'
import MotionStagger from '@/shared/ui/MotionStagger.vue'

const props = defineProps<{
  tasks: Task[]
}>()

defineEmits<{
  cardClick: [task: Task]
}>()

const pendingTasks = computed(() => props.tasks.filter((t) => !t.completed))
const completedTasks = computed(() => props.tasks.filter((t) => t.completed))
</script>

<style scoped>
.task-board {
  display: flex;
  gap: 16px;
  overflow-x: auto;
  padding-bottom: 16px;
}

/* 移动端：垂直布局 */
@media (max-width: 767px) {
  .task-board {
    flex-direction: column;
  }
}
</style>

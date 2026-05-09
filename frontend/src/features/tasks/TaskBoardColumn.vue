<template>
  <div class="board-column" :class="{ collapsed }">
    <div class="column-header" @click="collapsed = !collapsed">
      <span class="column-arrow" :class="{ collapsed }">▼</span>
      <span class="column-title">{{ title }}</span>
      <span class="column-count">{{ tasks.length }}</span>
    </div>

    <div v-if="!collapsed" class="column-cards">
      <TaskBoardCard
        v-for="task in tasks"
        :key="task.id"
        :task="task"
        @click="$emit('cardClick', task)"
      />

      <div v-if="tasks.length === 0" class="column-empty">暂无任务</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { Task } from '@/entities/task/model'
import TaskBoardCard from './TaskBoardCard.vue'

defineProps<{
  title: string
  tasks: Task[]
}>()

defineEmits<{
  cardClick: [task: Task]
}>()

const collapsed = ref(false)
</script>

<style scoped>
.board-column {
  flex: 1;
  min-width: 280px;
  background: var(--color-surface-muted);
  border-radius: 8px;
  padding: 12px;
}

.column-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px;
  cursor: pointer;
  user-select: none;
}

.column-arrow {
  font-size: 12px;
  transition: transform 200ms;
}

.column-arrow.collapsed {
  transform: rotate(-90deg);
}

.column-title {
  font-weight: 500;
  color: var(--color-text);
}

.column-count {
  font-size: 12px;
  color: var(--color-text-muted);
  background: var(--color-border);
  padding: 2px 6px;
  border-radius: 10px;
}

.column-cards {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 8px;
}

.column-empty {
  text-align: center;
  padding: 24px 12px;
  color: var(--color-text-muted);
  font-size: 13px;
}
</style>

<template>
  <div class="sk-grouped-list">
    <div v-for="g in groups" :key="g" class="sk-group">
      <div class="sk-group-header">
        <SkeletonBlock width="14px" height="14px" />
        <SkeletonBlock :width="`${40 + g * 12}px`" height="14px" />
        <SkeletonBlock width="24px" height="12px" />
      </div>
      <div class="sk-group-tasks">
        <div v-for="n in g" :key="n" class="sk-task-card">
          <SkeletonBlock width="20px" height="20px" borderRadius="50%" />
          <div class="sk-task-body">
            <SkeletonBlock :width="`${45 + n * 10}%`" height="14px" />
            <SkeletonBlock width="30%" height="12px" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import SkeletonBlock from '@/shared/ui/SkeletonBlock.vue'

const groups = [3, 2]
</script>

<style scoped>
.sk-grouped-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.sk-group-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: var(--color-surface-muted);
  border-radius: 6px;
}

.sk-group-tasks {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.sk-task-card {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 6px;
}

.sk-task-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
</style>

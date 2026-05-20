<template>
  <div class="sk-board">
    <div v-for="col in 2" :key="col" class="sk-column">
      <div class="sk-column-header">
        <SkeletonBlock width="60px" height="16px" />
        <SkeletonBlock width="24px" height="16px" borderRadius="999px" />
      </div>
      <div class="sk-column-body">
        <div v-for="n in 3" :key="n" class="sk-board-card">
          <SkeletonBlock :width="`${50 + n * 8}%`" height="14px" />
          <SkeletonBlock width="40%" height="12px" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import SkeletonBlock from '@/shared/ui/SkeletonBlock.vue'
</script>

<style scoped>
.sk-board {
  display: flex;
  gap: 16px;
}

.sk-column {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 12px;
  background: var(--color-surface-muted);
  border-radius: 8px;
  min-width: 0;
}

.sk-column-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 0;
}

.sk-column-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.sk-board-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 6px;
}

@media (max-width: 767px) {
  .sk-board {
    flex-direction: column;
  }
}
</style>

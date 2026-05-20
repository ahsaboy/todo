<template>
  <div class="sk-task-list">
    <!-- 移动端：卡片骨架 -->
    <div class="sk-cards">
      <div v-for="n in count" :key="n" class="sk-card">
        <SkeletonBlock width="20px" height="20px" borderRadius="50%" />
        <div class="sk-card-body">
          <SkeletonBlock :width="`${50 + (n % 3) * 15}%`" height="14px" />
          <SkeletonBlock width="40%" height="12px" />
        </div>
      </div>
    </div>

    <!-- 桌面端：表格骨架 -->
    <div class="sk-table">
      <div class="sk-table-head">
        <SkeletonBlock width="48px" height="14px" />
        <SkeletonBlock width="40%" height="14px" />
        <SkeletonBlock width="60px" height="14px" />
        <SkeletonBlock width="100px" height="14px" />
        <SkeletonBlock width="100px" height="14px" />
        <SkeletonBlock width="60px" height="14px" />
        <SkeletonBlock width="100px" height="14px" />
      </div>
      <div v-for="n in count" :key="n" class="sk-table-row">
        <SkeletonBlock width="20px" height="20px" borderRadius="50%" />
        <SkeletonBlock :width="`${40 + (n % 4) * 12}%`" height="14px" />
        <SkeletonBlock width="48px" height="20px" borderRadius="999px" />
        <SkeletonBlock width="80px" height="14px" />
        <SkeletonBlock width="80px" height="14px" />
        <SkeletonBlock width="48px" height="14px" />
        <SkeletonBlock width="80px" height="14px" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import SkeletonBlock from '@/shared/ui/SkeletonBlock.vue'

withDefaults(defineProps<{ count?: number }>(), { count: 5 })
</script>

<style scoped>
.sk-task-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* 移动端卡片 */
.sk-cards {
  display: none;
  flex-direction: column;
  gap: 12px;
}

.sk-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
}

.sk-card-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

/* 桌面端表格 */
.sk-table {
  display: flex;
  flex-direction: column;
}

.sk-table-head {
  display: grid;
  grid-template-columns: 48px 1fr 70px 110px 110px 70px 110px;
  align-items: center;
  gap: 12px;
  padding: 0 12px;
  height: 44px;
  border-bottom: 1px solid var(--color-border);
}

.sk-table-row {
  display: grid;
  grid-template-columns: 48px 1fr 70px 110px 110px 70px 110px;
  align-items: center;
  gap: 12px;
  padding: 0 12px;
  height: 44px;
  border-bottom: 1px solid var(--color-border);
}

.sk-table-row:last-child {
  border-bottom: none;
}

@media (max-width: 767px) {
  .sk-cards {
    display: flex;
  }

  .sk-table {
    display: none;
  }
}
</style>

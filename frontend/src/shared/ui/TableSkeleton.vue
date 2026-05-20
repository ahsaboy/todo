<template>
  <div class="sk-table-generic">
    <!-- 移动端：卡片骨架 -->
    <div class="sk-cards">
      <div v-for="n in rows" :key="n" class="sk-card">
        <div class="sk-card-header">
          <div class="sk-card-heading">
            <SkeletonBlock :width="`${60 + (n % 3) * 15}px`" height="15px" />
            <SkeletonBlock width="120px" height="12px" />
          </div>
          <SkeletonBlock width="44px" height="20px" borderRadius="999px" />
        </div>
        <div class="sk-card-meta">
          <SkeletonBlock width="100%" height="12px" />
          <SkeletonBlock width="70%" height="12px" />
        </div>
      </div>
    </div>

    <!-- 桌面端：表格骨架 -->
    <div class="sk-table-wrap">
      <table class="sk-table">
        <thead>
          <tr>
            <th v-for="c in columns" :key="c" :style="{ width: colWidths[c - 1] || 'auto' }">
              <SkeletonBlock width="60%" height="13px" />
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="n in rows" :key="n">
            <td v-for="c in columns" :key="c">
              <SkeletonBlock :width="`${40 + ((n + c) % 5) * 12}%`" height="14px" />
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import SkeletonBlock from '@/shared/ui/SkeletonBlock.vue'

withDefaults(defineProps<{
  columns?: number
  rows?: number
  colWidths?: string[]
}>(), {
  columns: 5,
  rows: 4,
  colWidths: () => [],
})
</script>

<style scoped>
.sk-table-generic {
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
  flex-direction: column;
  gap: 12px;
  padding: 14px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
}

.sk-card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.sk-card-heading {
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex: 1;
}

.sk-card-meta {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

/* 桌面端表格 */
.sk-table-wrap {
  overflow-x: auto;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 6px;
}

.sk-table {
  width: 100%;
  border-collapse: collapse;
}

.sk-table th,
.sk-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid var(--color-border);
  vertical-align: top;
}

.sk-table th {
  background: var(--color-surface-muted);
}

.sk-table tr:last-child td {
  border-bottom: none;
}

@media (max-width: 767px) {
  .sk-cards {
    display: flex;
  }

  .sk-table-wrap {
    display: none;
  }
}
</style>

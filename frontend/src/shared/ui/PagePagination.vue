<script setup lang="ts">
import { computed } from 'vue'
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'

const props = defineProps<{
  page: number
  total: number
  totalPages: number
}>()

const emit = defineEmits<{
  'update:page': [value: number]
}>()

const visiblePages = computed(() => {
  const { page, totalPages } = props
  if (totalPages <= 7) {
    return Array.from({ length: totalPages }, (_, i) => i + 1)
  }
  const pages: (number | '...')[] = [1]
  const start = Math.max(2, page - 1)
  const end = Math.min(totalPages - 1, page + 1)
  if (start > 2) pages.push('...')
  for (let i = start; i <= end; i++) pages.push(i)
  if (end < totalPages - 1) pages.push('...')
  pages.push(totalPages)
  return pages
})

function go(p: number) {
  if (p >= 1 && p <= props.totalPages && p !== props.page) {
    emit('update:page', p)
  }
}
</script>

<template>
  <div v-if="total > 0" class="page-pagination">
    <span class="pagination-info">共 {{ total }} 条</span>
    <button class="btn btn-sm" :disabled="page <= 1" @click="go(page - 1)">
      <ChevronLeft :size="14" />
    </button>
    <template v-for="(p, i) in visiblePages" :key="i">
      <span v-if="p === '...'" class="pagination-ellipsis">...</span>
      <button
        v-else
        class="btn btn-sm"
        :class="{ 'btn-primary': p === page }"
        @click="go(p)"
      >
        {{ p }}
      </button>
    </template>
    <button class="btn btn-sm" :disabled="page >= totalPages" @click="go(page + 1)">
      <ChevronRight :size="14" />
    </button>
  </div>
</template>

<style scoped>
.page-pagination {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: 1rem;
  font-size: 0.85rem;
  color: var(--color-text-muted, #888);
}

.pagination-info {
  white-space: nowrap;
  margin-right: 0.25rem;
}

.pagination-ellipsis {
  padding: 0 0.2rem;
}
</style>

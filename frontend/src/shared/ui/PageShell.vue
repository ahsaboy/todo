<script setup lang="ts">
import { type Component } from 'vue'
import SkeletonBlock from './SkeletonBlock.vue'

withDefaults(defineProps<{
  loading: boolean
  error?: string | null
  empty?: boolean
  skeleton?: Component
  emptyTitle?: string
  emptyDescription?: string
  emptyAction?: { label: string; onClick: () => void }
  errorRetry?: () => void
}>(), {
  error: null,
  empty: false,
  skeleton: undefined,
  emptyTitle: '暂无数据',
  emptyDescription: '',
  emptyAction: undefined,
  errorRetry: undefined,
})
</script>

<template>
  <Transition name="sk-fade" mode="out-in">
    <!-- Loading -->
    <div v-if="loading" key="skeleton">
      <component :is="skeleton ?? SkeletonBlock" />
    </div>

    <div v-else key="content">
      <!-- Error -->
      <div v-if="error" class="page-error">
        <p>{{ error }}</p>
        <button v-if="errorRetry" class="btn btn-primary" @click="errorRetry">重试</button>
      </div>

      <!-- Empty -->
      <div v-else-if="empty" class="page-empty">
        <slot name="empty-icon" />
        <p class="page-empty-title">{{ emptyTitle }}</p>
        <p v-if="emptyDescription" class="page-empty-desc">{{ emptyDescription }}</p>
        <button
          v-if="emptyAction"
          class="btn btn-primary"
          @click="emptyAction.onClick"
        >
          {{ emptyAction.label }}
        </button>
      </div>

      <!-- Content -->
      <slot v-else />
    </div>
  </Transition>
</template>

<style scoped>
.page-error {
  text-align: center;
  padding: 2rem 1rem;
  color: var(--color-danger);
}

.page-error p {
  margin-bottom: 0.75rem;
}

.page-empty {
  text-align: center;
  padding: 3rem 1rem;
  color: var(--color-text-muted);
}

.page-empty-title {
  font-size: 1rem;
  margin-bottom: 0.25rem;
}

.page-empty-desc {
  font-size: 0.85rem;
  margin-bottom: 0.75rem;
}
</style>

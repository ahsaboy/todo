<template>
  <div class="quick-create">
    <input
      v-model="title"
      type="text"
      placeholder="快速添加任务..."
      class="quick-input"
      @keyup.enter="handleCreate"
    />
    <button class="quick-btn" :disabled="!title.trim()" @click="handleCreate">添加</button>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const emit = defineEmits<{
  create: [title: string]
}>()

const title = ref('')

function handleCreate() {
  if (title.value.trim()) {
    emit('create', title.value.trim())
    title.value = ''
  }
}
</script>

<style scoped>
.quick-create {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}

.quick-input {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  font-size: 14px;
}

.quick-input:focus {
  outline: none;
  border-color: var(--color-primary);
}

.quick-btn {
  padding: 8px 16px;
  background: var(--color-primary);
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
}

.quick-btn:hover:not(:disabled) {
  background: var(--color-primary-hover);
}

.quick-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>

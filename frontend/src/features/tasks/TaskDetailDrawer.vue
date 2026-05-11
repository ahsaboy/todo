<template>
  <Teleport to="body">
    <Transition name="drawer">
      <div v-if="visible" class="drawer-overlay" @click.self="handleClose">
        <div class="drawer-content">
          <div class="drawer-header">
            <h3>{{ title || (task ? '编辑任务' : '创建任务') }}</h3>
            <div class="drawer-header-actions">
              <button v-if="task" class="btn-delete" type="button" @click="handleDelete">删除</button>
              <button class="btn-close" type="button" @click="handleClose">×</button>
            </div>
          </div>

          <div class="drawer-body">
            <slot v-if="$slots.default"></slot>
            <TaskForm
              v-else
              :initial-data="task || {}"
              :mode="task ? 'edit' : 'create'"
              @submit="handleSubmit"
              @cancel="handleClose"
            />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import type { Task } from '@/entities/task/model'
import type { CreateTaskPayload, UpdateTaskPayload } from '@/entities/task/model'
import TaskForm from './TaskForm.vue'

const props = defineProps<{
  visible: boolean
  title?: string
  task?: Task | null
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  submit: [payload: CreateTaskPayload | UpdateTaskPayload]
  delete: [id: number]
}>()

const hasChanges = ref(false)

watch(
  () => props.visible,
  (val) => {
    if (val) {
      hasChanges.value = false
    }
  },
)

function handleClose() {
  if (hasChanges.value) {
    if (!confirm('确定要关闭吗？未保存的更改将丢失。')) {
      return
    }
  }
  emit('update:visible', false)
}

function handleDelete() {
  if (props.task) {
    emit('delete', props.task.id)
  }
}

function handleSubmit(payload: CreateTaskPayload | UpdateTaskPayload) {
  emit('submit', payload)
  emit('update:visible', false)
}
</script>

<style scoped>
.drawer-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--el-mask-color);
  z-index: 1000;
  display: flex;
  justify-content: flex-end;
}

.drawer-content {
  width: 520px;
  max-width: 100%;
  height: 100%;
  background: var(--color-surface);
  display: flex;
  flex-direction: column;
}

.drawer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--color-border);
}

.drawer-header h3 {
  margin: 0;
  font-size: 16px;
}

.drawer-header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn-delete {
  padding: 4px 12px;
  background: none;
  border: 1px solid var(--color-danger);
  border-radius: 6px;
  color: var(--color-danger);
  font-size: 13px;
  cursor: pointer;
}

.btn-delete:hover {
  background: color-mix(in srgb, var(--color-danger) 10%, transparent);
}

.btn-close {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  color: var(--color-text-muted);
}

.drawer-body {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}

/* 动画 */
.drawer-enter-active,
.drawer-leave-active {
  transition: opacity 200ms;
}

.drawer-enter-active .drawer-content,
.drawer-leave-active .drawer-content {
  transition: transform 200ms;
}

.drawer-enter-from,
.drawer-leave-to {
  opacity: 0;
}

.drawer-enter-from .drawer-content,
.drawer-leave-to .drawer-content {
  transform: translateX(100%);
}

/* 移动端全屏 */
@media (max-width: 767px) {
  .drawer-content {
    width: 100%;
  }
}
</style>

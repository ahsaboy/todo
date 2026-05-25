<template>
  <Teleport to="body">
    <Transition name="overlay-motion" appear>
      <div v-if="visible" class="dialog-overlay" @click.self="handleSkip">
        <div class="dialog-content motion-panel motion-panel--scale">
          <div class="dialog-header">
            <h3>记录专注时长</h3>
            <button class="btn-close" type="button" @click="handleSkip">×</button>
          </div>

          <div class="dialog-body">
            <p class="task-name">{{ taskTitle }}</p>
            <div class="form-group">
              <label for="focus-duration">专注时长（分钟）</label>
              <input
                id="focus-duration"
                v-model.number="minutes"
                type="number"
                min="1"
                placeholder="例如：25"
                autofocus
                @keydown.enter="handleConfirm"
              />
            </div>
          </div>

          <div class="dialog-footer">
            <button class="btn-secondary" type="button" @click="handleSkip">跳过</button>
            <button class="btn-primary" type="button" :disabled="!isValid" @click="handleConfirm">记录</button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'

const props = defineProps<{
  visible: boolean
  taskTitle: string
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  confirm: [duration: number | null]
}>()

const minutes = ref<number | null>(null)

const isValid = computed(() => minutes.value !== null && minutes.value > 0)

watch(() => props.visible, (v) => {
  if (v) {
    minutes.value = null
  }
})

function handleConfirm() {
  if (isValid.value) {
    emit('confirm', minutes.value)
    emit('update:visible', false)
  }
}

function handleSkip() {
  emit('confirm', null)
  emit('update:visible', false)
}
</script>

<style scoped>
.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--el-mask-color);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
}

.dialog-content {
  width: 400px;
  max-width: 90vw;
  background: var(--color-surface);
  border-radius: 8px;
  box-shadow: var(--shadow-panel);
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--color-border);
}

.dialog-header h3 {
  margin: 0;
}

.btn-close {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
}

.dialog-body {
  padding: 20px;
}

.task-name {
  margin: 0 0 16px;
  font-size: 14px;
  color: var(--color-text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-group label {
  font-size: 13px;
  font-weight: 500;
}

.form-group input {
  padding: 8px 12px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-surface-muted);
  color: var(--color-text);
  font-size: 14px;
}

.form-group input:focus {
  outline: none;
  border-color: var(--color-primary);
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding: 12px 20px;
  border-top: 1px solid var(--color-border);
}

.btn-primary {
  padding: 8px 16px;
  background: var(--color-primary);
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  padding: 8px 16px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  cursor: pointer;
}
</style>

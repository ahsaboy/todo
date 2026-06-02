<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import BaseDialog from '@/shared/ui/BaseDialog.vue'

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

watch(() => props.visible, (v) => { if (v) minutes.value = null })

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

<template>
  <BaseDialog :visible="visible" title="记录专注时长" max-width="400px" @update:visible="handleSkip">
    <p class="task-name">{{ taskTitle }}</p>
    <div class="form-group">
      <label for="focus-duration">专注时长（分钟）</label>
      <input
        id="focus-duration"
        v-model.number="minutes"
        type="number"
        min="1"
        placeholder="例如：25"
        class="form-input"
        autofocus
        @keydown.enter="handleConfirm"
      />
    </div>
    <template #footer>
      <button class="btn" type="button" @click="handleSkip">跳过</button>
      <button class="btn btn-primary" type="button" :disabled="!isValid" @click="handleConfirm">记录</button>
    </template>
  </BaseDialog>
</template>

<style scoped>
.task-name {
  margin: 0;
  font-size: 0.85rem;
  color: var(--color-text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>

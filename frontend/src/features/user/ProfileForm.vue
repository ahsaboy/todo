<template>
  <form class="profile-form" @submit.prevent="handleSubmit">
    <div class="form-group">
      <label>用户名</label>
      <input :value="user?.username" disabled />
    </div>

    <div class="form-group">
      <label>邮箱</label>
      <input v-model="email" type="email" placeholder="your@email.com" />
    </div>

    <div class="form-group">
      <label>注册时间</label>
      <input :value="formatDate(user?.createdAt)" disabled />
    </div>

    <div class="form-actions">
      <button type="submit" class="btn-primary" :disabled="submitting || email === user?.email">
        {{ submitting ? '保存中...' : '保存' }}
      </button>
    </div>
  </form>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import type { User } from '@/entities/user/model'

const props = defineProps<{
  user: User | null
}>()

const emit = defineEmits<{
  submit: [email: string]
}>()

const email = ref(props.user?.email || '')
const submitting = ref(false)

watch(
  () => props.user,
  (val) => {
    if (val) email.value = val.email
  },
)

function formatDate(dateStr?: string): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('zh-CN')
}

async function handleSubmit() {
  submitting.value = true
  try {
    emit('submit', email.value)
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.profile-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
  max-width: 400px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-group label {
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text);
}

.form-group input {
  padding: 8px 12px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  font-size: 14px;
}

.form-group input:disabled {
  background: var(--color-surface-muted);
  color: var(--color-text-muted);
}

.form-actions {
  padding-top: 8px;
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
</style>

<template>
  <form class="password-form" @submit.prevent="handleSubmit">
    <div class="form-group">
      <label>当前密码 *</label>
      <input v-model="form.old_password" type="password" required />
    </div>

    <div class="form-group">
      <label>新密码 *</label>
      <input v-model="form.new_password" type="password" required minlength="6" />
    </div>

    <div class="form-group">
      <label>确认新密码 *</label>
      <input v-model="confirmPassword" type="password" required />
      <span v-if="error" class="error-text">{{ error }}</span>
    </div>

    <div class="form-actions">
      <button type="submit" class="btn-primary" :disabled="submitting">
        {{ submitting ? '修改中...' : '修改密码' }}
      </button>
    </div>
  </form>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import type { ChangePasswordPayload } from '@/entities/user/model'

const emit = defineEmits<{
  submit: [payload: ChangePasswordPayload]
}>()

const form = reactive<ChangePasswordPayload>({
  old_password: '',
  new_password: '',
})

const confirmPassword = ref('')
const error = ref('')
const submitting = ref(false)

function validate(): boolean {
  if (form.new_password !== confirmPassword.value) {
    error.value = '两次输入的密码不一致'
    return false
  }
  if (form.new_password.length < 6) {
    error.value = '密码长度至少 6 位'
    return false
  }
  error.value = ''
  return true
}

async function handleSubmit() {
  if (!validate()) return

  submitting.value = true
  try {
    emit('submit', { ...form })
    form.old_password = ''
    form.new_password = ''
    confirmPassword.value = ''
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.password-form {
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

.error-text {
  font-size: 12px;
  color: var(--color-danger);
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

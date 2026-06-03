<template>
  <form class="password-form" @submit.prevent="state.handleSubmit">
    <div class="form-group">
      <label for="password-current">当前密码 *</label>
      <input
        id="password-current"
        v-model="state.form.old_password"
        name="current_password"
        type="password"
        required
        autocomplete="current-password"
      />
    </div>

    <div class="form-group">
      <label for="password-new">新密码 *</label>
      <input
        id="password-new"
        v-model="state.form.new_password"
        name="new_password"
        type="password"
        required
        minlength="6"
        autocomplete="new-password"
      />
    </div>

    <div class="form-group">
      <label for="password-confirm">确认新密码 *</label>
      <input
        id="password-confirm"
        v-model="confirmPassword"
        name="confirm_password"
        type="password"
        required
        autocomplete="new-password"
      />
      <span v-if="state.error.value" class="error-text">{{ state.error.value }}</span>
    </div>

    <div class="form-actions">
      <button type="submit" class="btn-primary" :disabled="state.submitting.value">
        {{ state.submitting.value ? '修改中...' : '修改密码' }}
      </button>
    </div>
  </form>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useFormState } from '@/shared/composables/useFormState'
import type { ChangePasswordPayload } from '@/entities/user/model'

const emit = defineEmits<{
  submit: [payload: ChangePasswordPayload]
}>()

const confirmPassword = ref('')

const state = useFormState<ChangePasswordPayload & { _confirm?: string }>({
  initialData: { old_password: '', new_password: '' },
  validate: (data) => {
    if (data.new_password !== confirmPassword.value) return '两次输入的密码不一致'
    if (data.new_password.length < 6) return '密码长度至少 6 位'
    return null
  },
  onSubmit: async (data) => {
    emit('submit', { old_password: data.old_password, new_password: data.new_password })
    state.reset()
    confirmPassword.value = ''
  },
})
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

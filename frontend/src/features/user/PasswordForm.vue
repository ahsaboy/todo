<template>
  <form class="password-form" @submit.prevent="handleSubmit">
    <div v-if="hasPassword" class="form-group">
      <label for="password-current">当前密码 *</label>
      <input
        id="password-current"
        v-model="oldPassword"
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
        v-model="newPassword"
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
      <span v-if="error" class="error-text">{{ error }}</span>
    </div>

    <div class="form-actions">
      <button type="submit" class="btn-primary" :disabled="submitting">
        {{ submitting ? '处理中...' : (hasPassword ? '修改密码' : '设置密码') }}
      </button>
    </div>
  </form>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{
  hasPassword: boolean
  onChangePassword?: (payload: { old_password: string; new_password: string }) => Promise<void>
  onSetPassword?: (newPassword: string) => Promise<void>
}>()

const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const submitting = ref(false)
const error = ref('')

async function handleSubmit() {
  error.value = ''
  if (props.hasPassword && !oldPassword.value) {
    error.value = '请输入当前密码'
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    error.value = '两次输入的密码不一致'
    return
  }
  if (newPassword.value.length < 6) {
    error.value = '密码长度至少 6 位'
    return
  }

  submitting.value = true
  try {
    if (props.hasPassword && props.onChangePassword) {
      await props.onChangePassword({ old_password: oldPassword.value, new_password: newPassword.value })
    } else if (!props.hasPassword && props.onSetPassword) {
      await props.onSetPassword(newPassword.value)
    }
    oldPassword.value = ''
    newPassword.value = ''
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
  gap: 1rem;
  max-width: 400px;
}

.error-text {
  font-size: var(--text-xs);
  color: var(--color-danger);
}

.form-actions {
  padding-top: 0.25rem;
}

.btn-primary {
  min-height: 36px;
  padding: 0 1.25rem;
}

.btn-primary:disabled {
  opacity: var(--opacity-disabled);
  cursor: not-allowed;
}
</style>

<template>
  <div class="profile-page">
    <h2>个人资料</h2>

    <div class="form-section">
      <h3>基本信息</h3>
      <ProfileForm :user="user" @submit="handleProfileSubmit" />
    </div>

    <div class="form-section">
      <h3>修改密码</h3>
      <PasswordForm @submit="handlePasswordSubmit" />
    </div>

    <div v-if="message" class="message" :class="messageType">
      {{ message }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getProfile, updateProfile, changePassword } from '@/entities/user/api'
import { toUser } from '@/entities/user/mapper'
import type { User, ChangePasswordPayload } from '@/entities/user/model'
import ProfileForm from '@/features/user/ProfileForm.vue'
import PasswordForm from '@/features/user/PasswordForm.vue'

const user = ref<User | null>(null)
const message = ref('')
const messageType = ref<'success' | 'error'>('success')

onMounted(async () => {
  try {
    const response = await getProfile()
    user.value = toUser(response.data)
  } catch {
    showMessage('加载用户信息失败', 'error')
  }
})

function showMessage(msg: string, type: 'success' | 'error') {
  message.value = msg
  messageType.value = type
  setTimeout(() => {
    message.value = ''
  }, 3000)
}

async function handleProfileSubmit(email: string) {
  try {
    const response = await updateProfile({ email })
    user.value = toUser(response.data)
    showMessage('保存成功', 'success')
  } catch {
    showMessage('保存失败', 'error')
  }
}

async function handlePasswordSubmit(payload: ChangePasswordPayload) {
  try {
    await changePassword(payload)
    showMessage('密码修改成功', 'success')
  } catch {
    showMessage('密码修改失败', 'error')
  }
}
</script>

<style scoped>
.profile-page {
  display: flex;
  flex-direction: column;
  gap: 24px;
  max-width: 640px;
}

.profile-page h2 {
  margin: 0;
  font-size: 20px;
}

.form-section {
  padding: 20px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
}

.form-section h3 {
  margin: 0 0 16px;
  font-size: 16px;
}

.message {
  padding: 12px;
  border-radius: 6px;
  font-size: 14px;
}

.message.success {
  background: #dcfce7;
  color: #16a34a;
}

.message.error {
  background: #fee2e2;
  color: #dc2626;
}
</style>

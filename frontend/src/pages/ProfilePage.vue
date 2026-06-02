<template>
  <div class="profile-page">
    <h2>个人资料</h2>

    <PageShell
      :loading="loading"
      :error="loadError"
      :skeleton="ProfileSkeleton"
      :error-retry="loadProfile"
    >
      <div class="form-section">
        <h3>基本信息</h3>
        <ProfileForm :user="user" @submit="handleProfileSubmit" />
      </div>

      <div class="form-section">
        <h3>修改密码</h3>
        <PasswordForm @submit="handlePasswordSubmit" />
      </div>
    </PageShell>

    <div v-if="message" class="message" :class="messageType">
      {{ message }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { getProfile, updateProfile, changePassword } from '@/entities/user/api'
import { toUser } from '@/entities/user/mapper'
import type { User, ChangePasswordPayload } from '@/entities/user/model'
import { useFetch } from '@/shared/composables/useFetch'
import ProfileForm from '@/features/user/ProfileForm.vue'
import PasswordForm from '@/features/user/PasswordForm.vue'
import ProfileSkeleton from '@/shared/ui/ProfileSkeleton.vue'
import PageShell from '@/shared/ui/PageShell.vue'

const { data: user, isLoading: loading, error: loadError, load: loadProfile } = useFetch<User>({
  fetcher: async () => {
    const response = await getProfile()
    return toUser(response.data)
  },
  errorPrefix: '加载用户信息',
})

const message = ref('')
const messageType = ref<'success' | 'error'>('success')

function showMessage(msg: string, type: 'success' | 'error') {
  message.value = msg
  messageType.value = type
  setTimeout(() => { message.value = '' }, 3000)
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
  margin: 0 auto;
}

@media (min-width: 1024px) {
  .profile-page { padding-top: 20px; }
}

.profile-page h2 { margin: 0; font-size: 20px; }

.form-section {
  padding: 20px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  margin-bottom: 16px;
}

.form-section:last-child { margin-bottom: 0; }
.form-section h3 { margin: 0 0 16px; font-size: 16px; }

.message {
  padding: 12px;
  border-radius: 6px;
  font-size: 14px;
}

.message.success {
  background: color-mix(in srgb, var(--color-glow-success) 72%, transparent);
  border: 1px solid color-mix(in srgb, var(--color-success) 18%, transparent);
  color: var(--color-success);
}

.message.error {
  background: color-mix(in srgb, var(--color-glow-danger) 84%, transparent);
  border: 1px solid color-mix(in srgb, var(--color-danger) 26%, transparent);
  color: var(--color-danger);
}
</style>

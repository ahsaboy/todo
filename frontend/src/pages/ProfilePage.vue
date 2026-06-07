<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import PageShell from '@/shared/ui/PageShell.vue'
import ProfileSkeleton from '@/shared/ui/ProfileSkeleton.vue'
import ProfileForm from '@/features/user/ProfileForm.vue'
import EmailManager from '@/features/user/EmailManager.vue'
import OAuthAccountsManager from '@/features/user/OAuthAccountsManager.vue'
import PasswordForm from '@/features/user/PasswordForm.vue'
import { useFetch } from '@/shared/composables/useFetch'
import { ApiError } from '@/shared/api/errors'
import { getProfile, updateProfile, changePassword, setPassword } from '@/entities/user/api'
import { toProfile } from '@/entities/user/mapper'
import { getOAuthProviders, getEmailStatus } from '@/entities/auth/api'
import type { Profile } from '@/entities/user/model'
import type { OAuthProvider } from '@/entities/auth/model'

const route = useRoute()

const {
  data: profile,
  isLoading: loading,
  error: loadError,
  load: loadProfile,
} = useFetch<Profile>({
  fetcher: async () => {
    const res = await getProfile()
    return toProfile(res.data)
  },
  errorPrefix: '加载用户信息',
})

const oauthProviders = ref<OAuthProvider[]>([])
const emailServiceEnabled = ref(false)

const message = ref('')
const messageType = ref<'success' | 'error'>('success')
let messageTimer: ReturnType<typeof setTimeout> | null = null

function showMessage(msg: string, type: 'success' | 'error') {
  if (messageTimer) clearTimeout(messageTimer)
  message.value = msg
  messageType.value = type
  messageTimer = setTimeout(() => { message.value = ''; messageTimer = null }, 3000)
}

onMounted(async () => {
  const [providersRes, emailRes] = await Promise.allSettled([
    getOAuthProviders(),
    getEmailStatus(),
  ])

  if (providersRes.status === 'fulfilled') {
    oauthProviders.value = providersRes.value.data ?? []
  }
  if (emailRes.status === 'fulfilled') {
    emailServiceEnabled.value = emailRes.value.data?.available ?? false
  }

  // 检查 OAuth 绑定回调结果
  const params = new URLSearchParams(window.location.hash.split('?')[1] || '')
  const linked = params.get('oauth_linked')
  const linkError = params.get('oauth_link_error')
  if (linked) {
    showMessage(`${linked} 账号绑定成功`, 'success')
    loadProfile()
    cleanHashParams()
  } else if (linkError) {
    const errorMessages: Record<string, string> = {
      already_linked: '该账号已经绑定过了',
      linked_to_other: '该账号已绑定到其他用户',
      link_failed: '绑定失败，请重试',
    }
    showMessage(errorMessages[linkError] || `绑定失败: ${linkError}`, 'error')
    cleanHashParams()
  }
})

function cleanHashParams() {
  const hash = window.location.hash.split('?')[0]
  history.replaceState(null, '', hash)
}

async function handleEmailSubmit(email: string, code?: string) {
  try {
    const res = await updateProfile({ email, code })
    profile.value = toProfile(res.data)
    showMessage('邮箱保存成功', 'success')
  } catch (e) {
    showMessage(e instanceof ApiError ? e.message : '邮箱保存失败', 'error')
  }
}

async function handlePasswordChange(payload: { old_password: string; new_password: string }) {
  try {
    await changePassword(payload)
    showMessage('密码修改成功', 'success')
    loadProfile()
  } catch (e) {
    showMessage(e instanceof ApiError ? e.message : '密码修改失败', 'error')
  }
}

async function handlePasswordSet(newPassword: string) {
  try {
    await setPassword(newPassword)
    showMessage('密码设置成功', 'success')
    loadProfile()
  } catch (e) {
    showMessage(e instanceof ApiError ? e.message : '密码设置失败', 'error')
  }
}

function handleOAuthUnlinked() {
  showMessage('已解绑', 'success')
  loadProfile()
}

function handleOAuthError(msg: string) {
  showMessage(msg, 'error')
}
</script>

<template>
  <div class="page-container">
    <h1 class="page-title">个人资料</h1>

    <PageShell :loading="loading" :error="loadError" :skeleton="ProfileSkeleton" :error-retry="loadProfile">
      <div class="profile-sections">
        <!-- 头像 + 基本信息 -->
        <section class="profile-card">
          <ProfileForm :user="profile?.user ?? null" />
        </section>

        <!-- 邮箱管理 -->
        <section class="profile-card">
          <h2 class="card-title">邮箱</h2>
          <EmailManager
            :current-email="profile?.user?.email ?? ''"
            :email-service-enabled="emailServiceEnabled"
            @submit="handleEmailSubmit"
          />
        </section>

        <!-- OAuth 账号管理 -->
        <section v-if="oauthProviders.length > 0" class="profile-card">
          <h2 class="card-title">社交账号</h2>
          <OAuthAccountsManager
            :accounts="profile?.oauthAccounts ?? []"
            :providers="oauthProviders"
            :has-password="profile?.hasPassword ?? true"
            @unlinked="handleOAuthUnlinked"
            @error="handleOAuthError"
          />
        </section>

        <!-- 密码 -->
        <section class="profile-card">
          <h2 class="card-title">{{ profile?.hasPassword === false ? '设置密码' : '修改密码' }}</h2>
          <PasswordForm
            :has-password="profile?.hasPassword ?? true"
            :on-change-password="handlePasswordChange"
            :on-set-password="handlePasswordSet"
          />
        </section>
      </div>
    </PageShell>

    <!-- 全局提示 -->
    <Transition name="toast-slide">
      <div v-if="message" class="toast" :class="messageType">
        {{ message }}
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.page-container {
  max-width: 640px;
  margin: 0 auto;
}

@media (min-width: 1024px) {
  .page-container { padding-top: 20px; }
}

.page-title {
  margin: 0 0 1.25rem;
  font-size: var(--text-2xl);
  font-weight: 600;
  color: var(--color-text);
}

.profile-sections {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.profile-card {
  padding: 1.25rem;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
}

.card-title {
  margin: 0 0 1rem;
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-text);
}

.toast {
  position: fixed;
  bottom: 2rem;
  left: 50%;
  transform: translateX(-50%);
  padding: 0.75rem 1.25rem;
  border-radius: var(--radius-md);
  font-size: var(--text-sm);
  font-weight: 500;
  z-index: 1000;
  pointer-events: none;
}

.toast.success {
  background: var(--color-success);
  color: #fff;
}

.toast.error {
  background: var(--color-danger);
  color: #fff;
}

.toast-slide-enter-active,
.toast-slide-leave-active {
  transition: all var(--motion-duration-base) var(--motion-ease-standard);
}

.toast-slide-enter-from,
.toast-slide-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(1rem);
}
</style>

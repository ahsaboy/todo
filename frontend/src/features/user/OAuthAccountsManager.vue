<script setup lang="ts">
import { computed } from 'vue'
import { Github, Chrome, Terminal, Globe, Link, Unlink, AlertCircle } from 'lucide-vue-next'
import type { OAuthAccount } from '@/entities/user/model'
import type { OAuthProvider } from '@/entities/auth/model'
import { unlinkOAuthAccount } from '@/entities/user/api'
import { api } from '@/shared/api/client'
import type { ApiResponse } from '@/shared/api/types'
import { getBaseRedirectUri } from '@/shared/utils/url'

const props = defineProps<{
  accounts: OAuthAccount[]
  providers: OAuthProvider[]
  hasPassword: boolean
}>()

const emit = defineEmits<{
  unlinked: []
  error: [msg: string]
}>()

const iconComponents: Record<string, typeof Github> = {
  github: Github,
  google: Chrome,
  globe: Globe,
  terminal: Terminal,
  message: Globe,
  user: Globe,
}

const providerLabels: Record<string, string> = {
  github: 'GitHub',
  google: 'Google',
  linuxdo: 'LinuxDo',
}

const linkedProviders = computed(() => new Set(props.accounts.map(a => a.provider)))

const availableProviders = computed(() =>
  props.providers.filter(p => !linkedProviders.value.has(p.name)),
)

const canUnlink = computed(() => props.hasPassword || props.accounts.length > 1)

async function handleUnlink(account: OAuthAccount) {
  try {
    await unlinkOAuthAccount(account.id)
    emit('unlinked')
  } catch {
    emit('error', '解绑失败')
  }
}

async function handleLink(provider: string) {
  try {
    const redirectUri = encodeURIComponent(getBaseRedirectUri())
    const res = await api.get<ApiResponse<{ auth_url: string }>>(`/user/oauth/${provider}/link?redirect_uri=${redirectUri}`)
    if (res.data?.auth_url) {
      window.location.href = res.data.auth_url
    }
  } catch {
    emit('error', '发起绑定失败')
  }
}
</script>

<template>
  <div class="oauth-manager">
    <!-- 已绑定列表 -->
    <div v-if="accounts.length > 0" class="oauth-list">
      <div v-for="account in accounts" :key="account.id" class="oauth-item">
        <component :is="iconComponents[account.provider] || Globe" :size="20" :stroke-width="1.8" class="oauth-item-icon" />
        <span class="oauth-item-label">{{ providerLabels[account.provider] || account.provider }}</span>
        <span class="oauth-item-name">{{ account.displayName }}</span>
        <button
          class="oauth-unlink-btn"
          :disabled="!canUnlink"
          :title="canUnlink ? '解绑' : '至少需要保留一种登录方式'"
          @click="handleUnlink(account)"
        >
          <Unlink :size="14" />
          <span>解绑</span>
        </button>
      </div>
    </div>
    <p v-else class="oauth-empty">尚未绑定任何社交账号</p>

    <!-- 解绑保护提示 -->
    <p v-if="!canUnlink && accounts.length > 0" class="oauth-hint">
      <AlertCircle :size="14" />
      <span>当前账号无密码，无法解绑唯一的登录方式。请先设置密码。</span>
    </p>

    <!-- 可绑定列表 -->
    <div v-if="availableProviders.length > 0" class="oauth-link-section">
      <p class="oauth-link-label">绑定新账号</p>
      <div class="oauth-link-buttons">
        <button
          v-for="provider in availableProviders"
          :key="provider.name"
          class="oauth-link-btn"
          @click="handleLink(provider.name)"
        >
          <component :is="iconComponents[provider.icon] || Globe" :size="16" :stroke-width="1.8" />
          <span>{{ providerLabels[provider.name] || provider.label }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.oauth-manager {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.oauth-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.oauth-item {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.625rem 0.75rem;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
}

.oauth-item-icon {
  color: var(--color-text-muted);
  flex-shrink: 0;
}

.oauth-item-label {
  font-weight: 600;
  font-size: var(--text-sm);
  color: var(--color-text);
  min-width: 56px;
}

.oauth-item-name {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.oauth-unlink-btn {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.25rem 0.5rem;
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  background: transparent;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  cursor: pointer;
  flex-shrink: 0;
  transition: all var(--motion-duration-fast) var(--motion-ease-standard);
}

.oauth-unlink-btn:hover:not(:disabled) {
  color: var(--color-danger);
  border-color: var(--color-danger);
}

.oauth-unlink-btn:disabled {
  opacity: var(--opacity-disabled);
  cursor: not-allowed;
}

.oauth-empty {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  margin: 0;
}

.oauth-hint {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: var(--text-xs);
  color: var(--color-warning, var(--color-text-muted));
  margin: 0;
  padding: 0.5rem 0.75rem;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
}

.oauth-link-section {
  padding-top: 0.5rem;
  border-top: 1px solid var(--color-border);
}

.oauth-link-label {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  margin: 0 0 0.5rem;
}

.oauth-link-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.oauth-link-btn {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.5rem 0.875rem;
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-text);
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  cursor: pointer;
  min-height: 40px;
  transition: all var(--motion-duration-fast) var(--motion-ease-standard);
}

.oauth-link-btn:hover {
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.oauth-link-btn:active {
  transform: scale(0.98);
}
</style>

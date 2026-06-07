<script setup lang="ts">
import { ref, watch } from 'vue'
import { Copy, Check, Github, Chrome, Terminal, Globe } from 'lucide-vue-next'
import PageShell from '@/shared/ui/PageShell.vue'
import BaseDialog from '@/shared/ui/BaseDialog.vue'
import ConfigGroupCard from './ConfigGroupCard.vue'
import { useFetch } from '@/shared/composables/useFetch'
import { getConfig } from '@/entities/app-config/api'
import type { ConfigGroup } from '@/entities/app-config/model'
import { adminApi } from '@/shared/api/admin-client'
import type { ApiResponse } from '@/shared/api/types'

const { data: groups, isLoading, error, load } = useFetch<ConfigGroup[]>({
  fetcher: getConfig,
  errorPrefix: '加载系统配置',
})

const restartVisible = ref(false)

function onSaved(restartRequired: boolean) {
  if (restartRequired) restartVisible.value = true
  load()
}

const groupTitles: Record<string, string> = {
  server: '服务器',
  i18n: '国际化',
  reminder: '提醒',
  cors: '跨域 CORS',
  rate_limit: '限流',
  logging: '日志',
  email: '邮箱配置',
  static: '静态资源',
  database: '数据库',
  admin: '管理员',
  oauth: 'OAuth 社交登录',
}

// OAuth 回调地址
interface CallbackInfo {
  provider: string
  label: string
  icon: string
  callback_url: string
}
const callbackURLs = ref<CallbackInfo[]>([])
const copiedKey = ref<string | null>(null)
const hasOAuthGroup = ref(false)

const iconComponents: Record<string, typeof Github> = {
  github: Github,
  chrome: Chrome,
  terminal: Terminal,
  globe: Globe,
}

watch(
  () => groups.value,
  (g) => {
    if (g && g.some(item => item.group === 'oauth')) {
      hasOAuthGroup.value = true
      loadCallbackURLs()
    }
  },
  { once: true },
)

async function loadCallbackURLs() {
  try {
    const res = await adminApi.get<ApiResponse<{ base_url: string; callbacks: CallbackInfo[] }>>('/oauth/callback-urls')
    callbackURLs.value = res.data?.callbacks ?? []
  } catch {
    callbackURLs.value = []
  }
}

async function copyURL(url: string, key: string) {
  try {
    await navigator.clipboard.writeText(url)
  } catch {
    const ta = document.createElement('textarea')
    ta.value = url
    document.body.appendChild(ta)
    ta.select()
    document.execCommand('copy')
    document.body.removeChild(ta)
  }
  copiedKey.value = key
  setTimeout(() => { copiedKey.value = null }, 2000)
}
</script>

<template>
  <div class="page-container">
    <h1 class="admin-page-title">系统配置</h1>

    <PageShell :loading="isLoading" :error="error" :error-retry="load">
      <div class="config-groups">
        <ConfigGroupCard
          v-for="g in groups ?? []"
          :key="g.group"
          :group="g"
          :title="groupTitles[g.group] ?? g.group"
          @saved="onSaved"
        />
      </div>

      <div v-if="hasOAuthGroup && callbackURLs.length > 0" class="oauth-callbacks-card">
        <h3 class="oauth-callbacks-title">OAuth 回调地址</h3>
        <p class="oauth-callbacks-desc">将以下地址填入对应 OAuth 提供商的回调 URL 设置中</p>
        <div class="oauth-callbacks-list">
          <div v-for="cb in callbackURLs" :key="cb.provider" class="oauth-callback-item">
            <component :is="iconComponents[cb.icon] || Globe" :size="16" :stroke-width="1.8" class="oauth-callback-icon" />
            <span class="oauth-callback-label">{{ cb.label }}</span>
            <code class="oauth-callback-url">{{ cb.callback_url }}</code>
            <button
              class="oauth-copy-btn"
              :class="{ copied: copiedKey === cb.provider }"
              @click="copyURL(cb.callback_url, cb.provider)"
              :title="copiedKey === cb.provider ? '已复制' : '复制地址'"
            >
              <Check v-if="copiedKey === cb.provider" :size="14" />
              <Copy v-else :size="14" />
            </button>
          </div>
        </div>
      </div>

      <p class="config-hint">
        配置优先级：命令行参数 &gt; 数据库 &gt; 配置文件。
        <span class="badge badge-info badge-xs">重启</span> 表示需重启服务生效，其余均为保存即生效。
      </p>
    </PageShell>

    <BaseDialog v-model:visible="restartVisible" title="保存成功">
      <p>配置已保存到数据库。其中包含<strong>需重启生效</strong>的项，请重启服务后生效。</p>
      <template #footer="{ close }">
        <button class="btn btn-primary" @click="close">知道了</button>
      </template>
    </BaseDialog>
  </div>
</template>

<style scoped>
.config-groups {
  display: flex;
  flex-direction: column;
}

.config-hint {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  margin-top: 0.5rem;
  padding-top: 1rem;
  border-top: 1px solid var(--color-border);
  line-height: 1.8;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.2rem;
}

.badge-xs {
  font-size: 0.65rem;
  padding: 0.05rem 0.35rem;
  border-radius: var(--radius-sm);
  font-weight: 600;
  vertical-align: middle;
}

/* ── OAuth 回调地址卡片 ── */
.oauth-callbacks-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: 1.25rem;
  margin-top: 1rem;
}

.oauth-callbacks-title {
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-text);
  margin: 0 0 0.25rem;
}

.oauth-callbacks-desc {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  margin: 0 0 1rem;
}

.oauth-callbacks-list {
  display: flex;
  flex-direction: column;
  gap: 0.625rem;
}

.oauth-callback-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem 0.75rem;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
}

.oauth-callback-label {
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--color-text);
  min-width: 56px;
  flex-shrink: 0;
}

.oauth-callback-icon {
  color: var(--color-text-muted);
  flex-shrink: 0;
}

.oauth-callback-url {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  word-break: break-all;
  flex: 1;
  font-family: var(--font-mono, monospace);
  background: none;
  padding: 0;
}

.oauth-copy-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-surface);
  color: var(--color-text-muted);
  cursor: pointer;
  flex-shrink: 0;
  transition: all var(--motion-duration-fast) var(--motion-ease-standard);
}

.oauth-copy-btn:hover {
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.oauth-copy-btn.copied {
  border-color: var(--color-success, var(--color-primary));
  color: var(--color-success, var(--color-primary));
}
</style>

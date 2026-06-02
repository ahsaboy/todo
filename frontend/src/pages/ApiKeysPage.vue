<script setup lang="ts">
import { ref, computed } from 'vue'
import { api } from '@/shared/api/client'
import { useSimpleList } from '@/shared/composables/useSimpleList'
import { toApiKeyInfo } from '@/entities/api-key/mapper'
import type { ApiKeyInfo } from '@/entities/api-key/model'
import { deleteApiKey } from '@/entities/api-key/api'
import PageShell from '@/shared/ui/PageShell.vue'
import TableSkeleton from '@/shared/ui/TableSkeleton.vue'
import ApiKeyTable from '@/features/api-keys/ApiKeyTable.vue'
import ApiKeyCreateDialog from '@/features/api-keys/ApiKeyCreateDialog.vue'

const list = useSimpleList<ApiKeyInfo>({
  client: api,
  endpoint: '/user/keys',
  mapItem: toApiKeyInfo,
  errorPrefix: '加载 API Key',
})

const showCreate = ref(false)
const endpointCopied = ref(false)
const configCopied = ref(false)

const mcpEndpoint = computed(() => {
  if (typeof window === 'undefined') return '/mcp'
  return window.location.origin + '/mcp'
})

const mcpConfigJson = computed(() => JSON.stringify({
  mcpServers: {
    todo: {
      url: mcpEndpoint.value,
      headers: { 'api-key': '<your-api-key>' },
    },
  },
}, null, 2))

async function handleRevoke(id: number) {
  if (!confirm('确定要撤销这个 API Key 吗？撤销后将无法恢复。')) return
  await deleteApiKey(id)
  await list.load()
}

async function copyToClipboard(text: string): Promise<boolean> {
  try { await navigator.clipboard.writeText(text); return true } catch { return false }
}

async function copyEndpoint() {
  if (await copyToClipboard(mcpEndpoint.value)) {
    endpointCopied.value = true
    setTimeout(() => { endpointCopied.value = false }, 2000)
  }
}

async function copyConfig() {
  if (await copyToClipboard(mcpConfigJson.value)) {
    configCopied.value = true
    setTimeout(() => { configCopied.value = false }, 2000)
  }
}
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h2>API Key</h2>
      <button class="btn-primary" type="button" @click="showCreate = true">生成 API Key</button>
    </div>

    <div class="info-banner">
      API Key 用于第三方应用访问你的任务数据。请妥善保管，创建后仅显示一次。
      登录时自动生成的 "login" 类型 Key 在最后使用超过 24 小时后会自动清理。
    </div>

    <details class="mcp-card">
      <summary class="mcp-summary">
        <span class="mcp-summary-title">MCP 配置 / 使用说明</span>
        <span class="mcp-summary-hint">点击展开</span>
      </summary>
      <div class="mcp-body">
        <p class="mcp-intro">
          本服务支持通过 <strong>MCP（Model Context Protocol）</strong>协议让 AI 客户端（如 Claude Desktop、Cursor、VS Code Copilot 等）直接读写你的任务。
        </p>
        <div class="mcp-section">
          <div class="mcp-label">端点 URL</div>
          <div class="mcp-inline-row">
            <code class="mcp-code-inline">{{ mcpEndpoint }}</code>
            <button class="mcp-copy-btn" type="button" @click="copyEndpoint">{{ endpointCopied ? '已复制 ✓' : '复制' }}</button>
          </div>
        </div>
        <div class="mcp-section">
          <div class="mcp-label">认证方式</div>
          <div class="mcp-text">
            请求头携带 <code class="mcp-code-inline">api-key: &lt;你在上方创建的 Key&gt;</code>（也兼容 <code class="mcp-code-inline">Authorization: Bearer &lt;key&gt;</code>）。
          </div>
        </div>
        <div class="mcp-section">
          <div class="mcp-label">可用工具（共 12 个）</div>
          <ul class="mcp-tool-list">
            <li><strong>任务管理</strong>：创建 / 查询 / 更新 / 删除 / 切换完成 / 详情（6 个）</li>
            <li><strong>提醒配置</strong>：列表 / 新建 / 详情 / 更新 / 删除（5 个，需开启 <code class="mcp-code-inline">X-MCP-Include-Reminders</code>）</li>
            <li><strong>用户信息</strong>：获取当前用户资料（1 个）</li>
          </ul>
        </div>
        <div class="mcp-section">
          <div class="mcp-label">高级：输出格式与工具范围</div>
          <ul class="mcp-tool-list">
            <li><code class="mcp-code-inline">X-MCP-Include-Reminders</code>：默认隐藏 5 个提醒配置工具；设为任意非空值后显示并允许调用。</li>
            <li><code class="mcp-code-inline">X-MCP-Structured-Output</code>：默认 <code class="mcp-code-inline">content[0].text</code> 直接是完整 JSON 字符串；设为任意非空值后改走 <code class="mcp-code-inline">structuredContent</code>。</li>
          </ul>
          <p class="mcp-note">两个 Header 都按请求实时判定，无需重新 <code class="mcp-code-inline">initialize</code>。</p>
        </div>
        <div class="mcp-section">
          <div class="mcp-label-row">
            <div class="mcp-label">客户端配置示例</div>
            <button class="mcp-copy-btn" type="button" @click="copyConfig">{{ configCopied ? '已复制 ✓' : '复制配置' }}</button>
          </div>
          <pre class="mcp-code-block"><code>{{ mcpConfigJson }}</code></pre>
          <p class="mcp-note">将 <code class="mcp-code-inline">&lt;your-api-key&gt;</code> 替换为上方列表中的实际 Key。如需开启 reminder 工具或 structured 输出，在 <code class="mcp-code-inline">headers</code> 中追加对应 Header。</p>
        </div>
        <p class="mcp-footer">想了解协议细节？查看 <a href="https://modelcontextprotocol.io" target="_blank" rel="noopener noreferrer">modelcontextprotocol.io</a>。</p>
      </div>
    </details>

    <PageShell
      :loading="list.isLoading.value"
      :error="list.error.value"
      :empty="list.items.value.length === 0"
      :skeleton="TableSkeleton"
      empty-title="暂无 API Key"
      :error-retry="list.load"
    >
      <ApiKeyTable :keys="list.items.value" @revoke="handleRevoke" />
    </PageShell>

    <ApiKeyCreateDialog v-model:visible="showCreate" @created="list.load" />
  </div>
</template>

<style scoped>
.info-banner {
  background: color-mix(in srgb, var(--color-glow-info) 78%, transparent);
  border: 1px solid color-mix(in srgb, var(--color-info) 22%, transparent);
  border-radius: 6px;
  padding: 12px;
  color: var(--color-info);
  font-size: 14px;
}

.mcp-card {
  margin-top: 12px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  overflow: hidden;
}

.mcp-summary {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 14px;
  cursor: pointer;
  user-select: none;
  list-style: none;
  color: var(--color-text);
  font-size: 14px;
  font-weight: 600;
}

.mcp-summary::-webkit-details-marker { display: none; }
.mcp-summary:hover { background: var(--color-surface-muted); }

.mcp-summary-title::before {
  content: '▸';
  display: inline-block;
  margin-right: 8px;
  color: var(--color-text-muted);
  transition: transform var(--motion-duration-fast) var(--motion-ease-standard);
}

.mcp-card[open] .mcp-summary-title::before { content: '▾'; }
.mcp-summary-hint { color: var(--color-text-muted); font-size: 12px; font-weight: 400; }
.mcp-card[open] .mcp-summary-hint { display: none; }

.mcp-body { display: flex; flex-direction: column; gap: 14px; padding: 4px 14px 16px; border-top: 1px solid var(--color-border); }
.mcp-intro { margin: 12px 0 0; color: var(--color-text); font-size: 14px; line-height: 1.6; }
.mcp-section { display: flex; flex-direction: column; gap: 6px; }
.mcp-label { color: var(--color-text-muted); font-size: 12px; font-weight: 600; letter-spacing: 0.02em; text-transform: uppercase; }
.mcp-label-row { display: flex; align-items: center; justify-content: space-between; gap: 8px; }
.mcp-text { color: var(--color-text); font-size: 14px; line-height: 1.6; }
.mcp-inline-row { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }

.mcp-code-inline {
  display: inline-block;
  padding: 2px 6px;
  background: var(--color-surface-muted);
  border-radius: 4px;
  color: var(--color-text);
  font-family: monospace;
  font-size: 13px;
  overflow-wrap: anywhere;
}

.mcp-tool-list { margin: 0; padding-left: 20px; color: var(--color-text); font-size: 14px; line-height: 1.7; }

.mcp-code-block {
  margin: 0;
  padding: 12px;
  background: var(--color-surface-muted);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  overflow-x: auto;
  color: var(--color-text);
  font-family: monospace;
  font-size: 13px;
  line-height: 1.5;
}

.mcp-code-block code { background: transparent; padding: 0; font-family: inherit; }

.mcp-copy-btn {
  padding: 4px 10px;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  background: var(--color-surface);
  color: var(--color-text);
  font-size: 12px;
  cursor: pointer;
  transition: background-color var(--motion-duration-fast) var(--motion-ease-standard);
}

.mcp-copy-btn:hover { background: var(--color-surface-muted); }
.mcp-note { margin: 0; color: var(--color-text-muted); font-size: 12px; line-height: 1.5; }
.mcp-footer { margin: 0; color: var(--color-text-muted); font-size: 13px; }
.mcp-footer a { color: var(--color-info); text-decoration: none; }
.mcp-footer a:hover { text-decoration: underline; }
</style>

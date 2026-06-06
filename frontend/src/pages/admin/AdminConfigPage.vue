<script setup lang="ts">
import { ref } from 'vue'
import PageShell from '@/shared/ui/PageShell.vue'
import BaseDialog from '@/shared/ui/BaseDialog.vue'
import ConfigGroupCard from './ConfigGroupCard.vue'
import { useFetch } from '@/shared/composables/useFetch'
import { getConfig } from '@/entities/app-config/api'
import type { ConfigGroup } from '@/entities/app-config/model'

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

      <p class="config-hint">
        配置优先级：命令行参数 &gt; 数据库 &gt; 配置文件。
        仅显示非默认来源：
        <span class="badge badge-warning badge-xs">命令行</span>
        <span class="badge badge-primary badge-xs">数据库</span>。
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
</style>

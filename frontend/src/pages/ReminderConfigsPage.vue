<script setup lang="ts">
import { ref } from 'vue'
import { api } from '@/shared/api/client'
import { useSimpleList } from '@/shared/composables/useSimpleList'
import { toReminderConfig } from '@/entities/reminder-config/mapper'
import {
  createReminderConfig,
  updateReminderConfig,
} from '@/entities/reminder-config/api'
import type {
  ReminderConfig,
  CreateReminderConfigPayload,
  UpdateReminderConfigPayload,
} from '@/entities/reminder-config/model'
import PageShell from '@/shared/ui/PageShell.vue'
import TableSkeleton from '@/shared/ui/TableSkeleton.vue'
import ReminderConfigTable from '@/features/reminder-configs/ReminderConfigTable.vue'
import ReminderConfigForm from '@/features/reminder-configs/ReminderConfigForm.vue'
import TaskDetailDrawer from '@/features/tasks/TaskDetailDrawer.vue'

const list = useSimpleList<ReminderConfig>({
  client: api,
  endpoint: '/user/reminder-configs',
  mapItem: toReminderConfig,
  errorPrefix: '加载提醒配置',
})

const drawerVisible = ref(false)
const editingConfig = ref<ReminderConfig | null>(null)

function openCreate() {
  editingConfig.value = null
  drawerVisible.value = true
}

function editConfig(config: ReminderConfig) {
  editingConfig.value = config
  drawerVisible.value = true
}

function toPayload(config: ReminderConfig): Partial<CreateReminderConfigPayload> {
  return {
    name: config.name,
    channel_type: config.channelType as CreateReminderConfigPayload['channel_type'],
    webhook_url: config.webhookUrl,
    webhook_method: config.webhookMethod as CreateReminderConfigPayload['webhook_method'],
    webhook_headers: config.webhookHeaders,
    webhook_body_template: config.webhookBodyTemplate,
    max_retries: config.maxRetries,
    retry_delay_seconds: config.retryDelaySeconds,
    enabled: config.enabled,
  }
}

async function handleSubmit(payload: CreateReminderConfigPayload | UpdateReminderConfigPayload) {
  if (editingConfig.value) {
    await updateReminderConfig(editingConfig.value.id, payload as UpdateReminderConfigPayload)
  } else {
    await createReminderConfig(payload as CreateReminderConfigPayload)
  }
  drawerVisible.value = false
  await list.load()
}
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h2>提醒配置</h2>
      <button class="btn-primary" type="button" @click="openCreate">新增配置</button>
    </div>

    <PageShell
      :loading="list.isLoading.value"
      :error="list.error.value"
      :empty="list.items.value.length === 0"
      :skeleton="TableSkeleton"
      empty-title="暂无提醒配置"
      :empty-action="{ label: '创建第一个配置', onClick: openCreate }"
      :error-retry="list.load"
    >
      <ReminderConfigTable
        :configs="list.items.value"
        @edit="editConfig"
        @delete="(id) => list.deleteItem(`/user/reminder-configs/${id}`, '确定要删除这个配置吗？')"
      />
    </PageShell>

    <TaskDetailDrawer v-model:visible="drawerVisible" :title="editingConfig ? '编辑配置' : '新增配置'">
      <ReminderConfigForm
        :initial-data="editingConfig ? toPayload(editingConfig) : undefined"
        @submit="handleSubmit"
        @cancel="drawerVisible = false"
      />
    </TaskDetailDrawer>
  </div>
</template>

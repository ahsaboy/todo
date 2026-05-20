<template>
  <div class="page">
    <div class="page-header">
      <h2>提醒配置</h2>
      <button class="btn-primary" type="button" @click="openCreate">新增配置</button>
    </div>

    <Transition name="sk-fade" mode="out-in">
      <TableSkeleton v-if="loading" key="skeleton" :columns="7" :col-widths="['120px', '90px', '60px', '80px', '100px', '200px', '100px']" />

      <template v-else key="content">
        <div v-if="error" class="page-error">
          <p>{{ error }}</p>
          <button type="button" @click="fetchConfigs">重试</button>
        </div>

        <div v-else-if="configs.length === 0" class="page-empty">
          <p>暂无提醒配置</p>
          <button class="btn-primary" type="button" @click="openCreate">创建第一个配置</button>
        </div>

        <ReminderConfigTable v-else :configs="configs" @edit="editConfig" @delete="handleDelete" />
      </template>
    </Transition>

    <!-- 抽屉 -->
    <TaskDetailDrawer
      v-model:visible="drawerVisible"
      :title="editingConfig ? '编辑配置' : '新增配置'"
    >
      <ReminderConfigForm
        :initial-data="editingConfig ? toPayload(editingConfig) : undefined"
        @submit="handleSubmit"
        @cancel="drawerVisible = false"
      />
    </TaskDetailDrawer>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  getReminderConfigs,
  createReminderConfig,
  updateReminderConfig,
  deleteReminderConfig,
} from '@/entities/reminder-config/api'
import { toReminderConfig } from '@/entities/reminder-config/mapper'
import type {
  ReminderConfig,
  CreateReminderConfigPayload,
  UpdateReminderConfigPayload,
} from '@/entities/reminder-config/model'
import ReminderConfigTable from '@/features/reminder-configs/ReminderConfigTable.vue'
import ReminderConfigForm from '@/features/reminder-configs/ReminderConfigForm.vue'
import TableSkeleton from '@/shared/ui/TableSkeleton.vue'
import TaskDetailDrawer from '@/features/tasks/TaskDetailDrawer.vue'

const configs = ref<ReminderConfig[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const drawerVisible = ref(false)
const editingConfig = ref<ReminderConfig | null>(null)

onMounted(() => {
  fetchConfigs()
})

async function fetchConfigs() {
  loading.value = true
  error.value = null
  try {
    const response = await getReminderConfigs()
    const data = Array.isArray(response.data) ? response.data : []
    configs.value = data.map(toReminderConfig)
  } catch (e) {
    error.value = e instanceof Error ? e.message : '加载失败'
  } finally {
    loading.value = false
  }
}

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
  await fetchConfigs()
}

async function handleDelete(id: number) {
  if (!confirm('确定要删除这个配置吗？')) return
  await deleteReminderConfig(id)
  await fetchConfigs()
}
</script>

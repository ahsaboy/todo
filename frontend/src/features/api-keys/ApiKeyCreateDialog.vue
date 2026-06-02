<script setup lang="ts">
import { ref } from 'vue'
import BaseDialog from '@/shared/ui/BaseDialog.vue'
import { createApiKey } from '@/entities/api-key/api'

defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  created: []
}>()

const name = ref('')
const newKey = ref<string | null>(null)
const loading = ref(false)

async function handleCreate() {
  loading.value = true
  try {
    const response = await createApiKey({ name: name.value || undefined })
    const generatedKey = response.data.api_key ?? response.data.key
    if (!generatedKey) throw new Error('接口未返回 API Key')
    newKey.value = generatedKey
    emit('created')
  } catch (e) {
    window.alert('创建失败：' + (e instanceof Error ? e.message : '未知错误'))
  } finally {
    loading.value = false
  }
}

async function copyKey() {
  if (!newKey.value) return
  try {
    await navigator.clipboard.writeText(newKey.value)
  } catch {
    const textarea = document.createElement('textarea')
    textarea.value = newKey.value
    textarea.style.position = 'fixed'
    textarea.style.opacity = '0'
    document.body.appendChild(textarea)
    textarea.select()
    document.execCommand('copy')
    document.body.removeChild(textarea)
  }
  window.alert('已复制到剪贴板')
}

function handleClose() {
  if (newKey.value && !confirm('确定关闭吗？请确保已保存 Key。')) return
  newKey.value = null
  name.value = ''
  emit('update:visible', false)
}
</script>

<template>
  <BaseDialog :visible="visible" title="创建 API Key" max-width="480px" @update:visible="handleClose">
    <div v-if="!newKey" class="create-form">
      <div class="form-group">
        <label for="key-name">名称（可选）</label>
        <input
          id="key-name"
          v-model="name"
          name="api_key_name"
          type="text"
          class="form-input"
          placeholder="例如：生产环境"
          autocomplete="off"
        />
      </div>
      <button class="btn btn-primary" type="button" :disabled="loading" @click="handleCreate" style="width:100%">
        {{ loading ? '创建中...' : '创建' }}
      </button>
    </div>

    <div v-else class="key-display">
      <div class="warning-banner">请保存此 Key，关闭后将无法再次查看！</div>
      <div class="key-value">
        <code>{{ newKey }}</code>
        <button class="btn btn-primary btn-sm" type="button" @click="copyKey">复制</button>
      </div>
    </div>

    <template #footer>
      <button class="btn" type="button" @click="handleClose">
        {{ newKey ? '我已保存' : '取消' }}
      </button>
    </template>
  </BaseDialog>
</template>

<style scoped>
.warning-banner {
  background: color-mix(in srgb, var(--color-warning) 12%, transparent);
  border: 1px solid color-mix(in srgb, var(--color-warning) 24%, transparent);
  border-radius: 6px;
  padding: 0.75rem;
  margin-bottom: 1rem;
  color: var(--color-warning);
  font-size: 0.85rem;
}

.key-value {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: var(--color-surface-muted, rgba(0,0,0,0.04));
  padding: 0.75rem;
  border-radius: 6px;
}

.key-value code {
  flex: 1;
  word-break: break-all;
  font-size: 0.8rem;
}
</style>

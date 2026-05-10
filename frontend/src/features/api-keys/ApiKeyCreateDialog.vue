<template>
  <Teleport to="body">
    <Transition name="dialog">
      <div v-if="visible" class="dialog-overlay" @click.self="handleClose">
        <div class="dialog-content">
          <div class="dialog-header">
            <h3>创建 API Key</h3>
            <button class="btn-close" type="button" @click="handleClose">×</button>
          </div>

          <div class="dialog-body">
            <div v-if="!newKey" class="create-form">
              <div class="form-group">
                <label for="key-name">名称（可选）</label>
                <input
                  id="key-name"
                  v-model="name"
                  name="api_key_name"
                  type="text"
                  placeholder="例如：生产环境"
                  autocomplete="off"
                />
              </div>
              <button class="btn-primary" type="button" :disabled="loading" @click="handleCreate">
                {{ loading ? '创建中...' : '创建' }}
              </button>
            </div>

            <div v-else class="key-display">
              <div class="warning-banner">请保存此 Key，关闭后将无法再次查看！</div>
              <div class="key-value">
                <code>{{ newKey }}</code>
                <button class="btn-copy" type="button" @click="copyKey">复制</button>
              </div>
            </div>
          </div>

          <div class="dialog-footer">
            <button class="btn-secondary" type="button" @click="handleClose">
              {{ newKey ? '我已保存' : '取消' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref } from 'vue'
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
    if (!generatedKey) {
      throw new Error('接口未返回 API Key')
    }
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
    // fallback for non-HTTPS environments
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
  if (newKey.value && !confirm('确定关闭吗？请确保已保存 Key。')) {
    return
  }
  newKey.value = null
  name.value = ''
  emit('update:visible', false)
}
</script>

<style scoped>
.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
}

.dialog-content {
  width: 480px;
  max-width: 90vw;
  background: var(--color-surface);
  border-radius: 8px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.2);
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--color-border);
}

.dialog-header h3 {
  margin: 0;
}

.btn-close {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
}

.dialog-body {
  padding: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 16px;
}

.form-group label {
  font-size: 13px;
  font-weight: 500;
}

.form-group input {
  padding: 8px 12px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
}

.warning-banner {
  background: #fef3c7;
  border: 1px solid #fcd34d;
  border-radius: 6px;
  padding: 12px;
  margin-bottom: 16px;
  color: #92400e;
  font-size: 14px;
}

.key-value {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--color-surface-muted);
  padding: 12px;
  border-radius: 6px;
}

.key-value code {
  flex: 1;
  word-break: break-all;
  font-size: 13px;
}

.btn-copy {
  padding: 6px 12px;
  background: var(--color-primary);
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  white-space: nowrap;
}

.btn-primary {
  width: 100%;
  padding: 10px;
  background: var(--color-primary);
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
}

.btn-primary:disabled {
  opacity: 0.5;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  padding: 12px 20px;
  border-top: 1px solid var(--color-border);
}

.btn-secondary {
  padding: 8px 16px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  cursor: pointer;
}
</style>

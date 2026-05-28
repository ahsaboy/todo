<template>
  <div class="tag-manager-page">
    <header class="page-header">
      <div>
        <h1>标签管理</h1>
        <p class="page-desc">为任务自定义彩色标签,支持图标与排序;改名/删除会同步更新所有任务上的标签。</p>
      </div>
      <button class="btn-primary desktop-only" @click="openCreate">+ 新建标签</button>
    </header>

    <div v-if="store.loading && store.tags.length === 0" class="state-text">加载中...</div>
    <div v-else-if="store.error" class="state-text error">{{ store.error }}</div>
    <div v-else-if="store.tags.length === 0" class="empty">
      <p>还没有任何标签,点右上角"新建"开始吧。</p>
    </div>

    <ul v-else class="tag-list">
      <li
        v-for="(tag, index) in store.tags"
        :key="tag.id"
        class="tag-row"
        :draggable="true"
        @dragstart="onDragStart(index, $event)"
        @dragover.prevent="onDragOver(index)"
        @drop.prevent="onDrop(index)"
        @dragend="onDragEnd"
        :class="{ 'drag-over': dragOverIndex === index && draggedIndex !== index }"
      >
        <span class="drag-handle" title="拖拽排序">⋮⋮</span>
        <TagChip :name="tag.name" size="md" />
        <span class="tag-meta">
          <span class="meta-color" :style="{ background: tag.color }"></span>
          <code class="meta-hex">{{ tag.color }}</code>
          <span v-if="tag.icon" class="meta-icon">{{ tag.icon }}</span>
        </span>
        <div class="tag-actions">
          <button class="btn-link" @click="openEdit(tag)">编辑</button>
          <button class="btn-link danger" @click="confirmDelete(tag)">删除</button>
        </div>
      </li>
    </ul>

    <!-- 移动端浮动按钮 -->
    <button class="fab" type="button" aria-label="新建标签" @click="openCreate"><Plus :size="24" /></button>

    <!-- 编辑/新建 dialog -->
    <div v-if="dialogOpen" class="dialog-mask" @click.self="closeDialog">
      <div class="dialog">
        <h2>{{ editing ? '编辑标签' : '新建标签' }}</h2>
        <div class="form-group">
          <label>名称 *</label>
          <input
            v-model.trim="form.name"
            type="text"
            maxlength="32"
            placeholder="如:工作 / 学习 / 紧急"
            :class="{ error: !!formError.name }"
          />
          <span v-if="formError.name" class="error-text">{{ formError.name }}</span>
        </div>
        <div class="form-group">
          <label>颜色</label>
          <ColorPicker v-model="form.color" />
        </div>
        <div class="form-group">
          <label>图标</label>
          <IconPicker v-model="form.icon" />
        </div>
        <div class="dialog-preview">
          <span class="preview-label">预览:</span>
          <span class="preview-chip" :style="previewChipStyle">
            <component v-if="previewIcon" :is="previewIcon" :size="14" />
            <span>{{ form.name || '未命名' }}</span>
          </span>
        </div>
        <div class="dialog-actions">
          <button class="btn-secondary" @click="closeDialog">取消</button>
          <button class="btn-primary" :disabled="submitting" @click="onSubmit">
            {{ submitting ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { Plus } from 'lucide-vue-next'
import ColorPicker from '@/shared/ui/ColorPicker.vue'
import IconPicker from '@/shared/ui/IconPicker.vue'
import TagChip from '@/features/tags/TagChip.vue'
import { useTagStore } from '@/entities/tag/store'
import { CURATED_ICONS, DEFAULT_TAG_COLOR } from '@/shared/icons/curated'
import type { Tag } from '@/entities/tag/model'

const store = useTagStore()

const dialogOpen = ref(false)
const editing = ref<Tag | null>(null)
const submitting = ref(false)

const form = reactive({
  name: '',
  color: DEFAULT_TAG_COLOR,
  icon: '',
})

const formError = reactive({ name: '' })

const previewIcon = computed(() => (form.icon ? CURATED_ICONS[form.icon] ?? null : null))
const previewChipStyle = computed(() => ({
  color: form.color,
  background: `color-mix(in srgb, ${form.color} 14%, transparent)`,
  borderColor: `color-mix(in srgb, ${form.color} 30%, transparent)`,
}))

function openCreate() {
  editing.value = null
  form.name = ''
  form.color = DEFAULT_TAG_COLOR
  form.icon = ''
  formError.name = ''
  dialogOpen.value = true
}

function openEdit(tag: Tag) {
  editing.value = tag
  form.name = tag.name
  form.color = tag.color
  form.icon = tag.icon
  formError.name = ''
  dialogOpen.value = true
}

function closeDialog() {
  dialogOpen.value = false
  editing.value = null
}

async function onSubmit() {
  if (!form.name.trim()) {
    formError.name = '标签名不能为空'
    return
  }
  if (!/^#[0-9a-fA-F]{6}$/.test(form.color)) {
    alert('颜色必须是 #RRGGBB 形式')
    return
  }
  submitting.value = true
  try {
    if (editing.value) {
      await store.updateTag(editing.value.id, {
        name: form.name,
        color: form.color,
        icon: form.icon,
      })
    } else {
      await store.createTag({
        name: form.name,
        color: form.color,
        icon: form.icon,
      })
    }
    closeDialog()
  } catch (e) {
    const msg = e instanceof Error ? e.message : '保存失败'
    alert(msg)
  } finally {
    submitting.value = false
  }
}

async function confirmDelete(tag: Tag) {
  if (!confirm(`确认删除标签"${tag.name}"?该标签会从所有任务上摘除。`)) return
  try {
    const affected = await store.deleteTag(tag.id)
    if (affected > 0) {
      // 简单提示,不抢占主流程
      console.info(`已从 ${affected} 个任务上摘除标签`)
    }
  } catch (e) {
    alert(e instanceof Error ? e.message : '删除失败')
  }
}

// 拖拽排序
const draggedIndex = ref<number | null>(null)
const dragOverIndex = ref<number | null>(null)

function onDragStart(index: number, e: DragEvent) {
  draggedIndex.value = index
  if (e.dataTransfer) e.dataTransfer.effectAllowed = 'move'
}

function onDragOver(index: number) {
  dragOverIndex.value = index
}

async function onDrop(targetIndex: number) {
  const from = draggedIndex.value
  if (from === null || from === targetIndex) {
    onDragEnd()
    return
  }
  const list = [...store.tags]
  const [moved] = list.splice(from, 1)
  list.splice(targetIndex, 0, moved)
  // 持久化新的 sort_order;若中途失败则从服务端重新加载,回滚到一致状态
  try {
    for (let i = 0; i < list.length; i++) {
      const t = list[i]
      if (t.sortOrder !== i) {
        await store.updateTag(t.id, { sort_order: i })
      }
    }
  } catch (e) {
    alert(e instanceof Error ? e.message : '排序保存失败,已回滚')
    store.fetchTags(true)
  } finally {
    onDragEnd()
  }
}

function onDragEnd() {
  draggedIndex.value = null
  dragOverIndex.value = null
}

onMounted(() => {
  store.fetchTags()
})
</script>

<style scoped>
.tag-manager-page {
  padding: 24px;
  max-width: 880px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* 桌面端/移动端显示控制 */
.desktop-only {
  display: block;
}
.mobile-only {
  display: none;
}

@media (max-width: 767px) {
  .desktop-only {
    display: none;
  }
  .mobile-only {
    display: block;
  }
}

.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.page-header h1 {
  margin: 0 0 4px;
  font-size: 20px;
}

.page-desc {
  margin: 0;
  font-size: 13px;
  color: var(--color-text-muted);
}

.tag-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.tag-row {
  display: grid;
  grid-template-columns: 24px auto 1fr auto;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  background: var(--color-surface);
}

/* 移动端浮动按钮 */
.fab {
  display: none;
}

@media (max-width: 767px) {
  .tag-row {
    grid-template-columns: 20px 1fr auto;
    gap: 10px;
    padding: 12px 14px;
  }
  .tag-meta {
    display: none;
  }

  .fab {
    display: flex;
    position: fixed;
    right: 16px;
    bottom: calc(var(--bottom-nav-height) + 16px);
    width: 56px;
    height: 56px;
    background: var(--color-primary);
    color: white;
    border: none;
    border-radius: 50%;
    font-size: 24px;
    align-items: center;
    justify-content: center;
    box-shadow: var(--shadow-glow-primary);
    z-index: 50;
    cursor: pointer;
  }
  .fab:hover {
    background: var(--color-primary-hover);
  }

  .tag-manager-page {
    padding-bottom: 80px;
  }
}

.tag-row.drag-over {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 1px var(--color-primary) inset;
}

.drag-handle {
  cursor: grab;
  color: var(--color-text-muted);
  user-select: none;
}
.drag-handle:active {
  cursor: grabbing;
}

.tag-meta {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--color-text-muted);
}

.meta-color {
  display: inline-block;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  border: 1px solid var(--color-border);
}

.meta-hex {
  font-family: ui-monospace, monospace;
  font-size: 11px;
}

.meta-icon {
  font-family: ui-monospace, monospace;
  font-size: 11px;
  opacity: 0.7;
}

.tag-actions {
  display: flex;
  gap: 4px;
}

@media (max-width: 767px) {
  .tag-actions {
    gap: 8px;
  }
  .tag-actions .btn-link {
    padding: 6px 10px;
    font-size: 14px;
  }
}

.btn-link {
  background: transparent;
  border: 0;
  color: var(--color-primary);
  cursor: pointer;
  font-size: 13px;
  padding: 4px 8px;
}
.btn-link:hover {
  text-decoration: underline;
}
.btn-link.danger {
  color: var(--color-danger);
}

.empty,
.state-text {
  padding: 32px;
  text-align: center;
  color: var(--color-text-muted);
}
.state-text.error {
  color: var(--color-danger);
}

/* dialog */
.dialog-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
}

.dialog {
  background: var(--color-bg);
  border-radius: 10px;
  padding: 20px;
  width: 480px;
  max-width: calc(100vw - 32px);
  max-height: calc(100vh - 32px);
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 14px;
  box-shadow: 0 8px 28px rgba(0, 0, 0, 0.25);
}

@media (max-width: 767px) {
  .dialog {
    width: calc(100vw - 32px);
    max-height: 80vh;
    margin: auto 16px;
    padding: 24px;
  }
}

.dialog h2 {
  margin: 0;
  font-size: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-group label {
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text);
}

.form-group input {
  padding: 8px 12px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  font-size: 14px;
  background: var(--color-surface);
}

.form-group input.error {
  border-color: var(--color-danger);
}

.error-text {
  font-size: 12px;
  color: var(--color-danger);
}

.dialog-preview {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--color-text-muted);
}

.preview-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 10px;
  border: 1px solid;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 500;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding-top: 8px;
  border-top: 1px solid var(--color-border);
}

.btn-primary,
.btn-secondary {
  padding: 8px 16px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
}

.btn-primary {
  background: var(--color-primary);
  color: white;
  border: none;
}
.btn-primary:hover:not(:disabled) {
  background: var(--color-primary-hover);
}
.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  color: var(--color-text);
}
</style>

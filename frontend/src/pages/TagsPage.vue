<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Plus } from 'lucide-vue-next'
import BaseDialog from '@/shared/ui/BaseDialog.vue'
import MobileFab from '@/shared/ui/MobileFab.vue'
import PageShell from '@/shared/ui/PageShell.vue'
import ColorPicker from '@/shared/ui/ColorPicker.vue'
import IconPicker from '@/shared/ui/IconPicker.vue'
import TagChip from '@/features/tags/TagChip.vue'
import { useTagStore } from '@/entities/tag/store'
import { useFormState } from '@/shared/composables/useFormState'
import { useDragSort } from '@/shared/composables/useDragSort'
import { CURATED_ICONS, DEFAULT_TAG_COLOR } from '@/shared/icons/curated'
import type { Tag } from '@/entities/tag/model'

const store = useTagStore()

const dialogOpen = ref(false)
const editing = ref<Tag | null>(null)

const tagForm = useFormState({
  initialData: { name: '', color: DEFAULT_TAG_COLOR, icon: '' },
  validate: (d) => {
    if (!d.name.trim()) return '标签名不能为空'
    if (!/^#[0-9a-fA-F]{6}$/.test(d.color)) return '颜色必须是 #RRGGBB 形式'
    return null
  },
  onSubmit: async (data) => {
    if (editing.value) {
      await store.updateTag(editing.value.id, { name: data.name, color: data.color, icon: data.icon })
    } else {
      await store.createTag({ name: data.name, color: data.color, icon: data.icon })
    }
    dialogOpen.value = false
    editing.value = null
  },
})

const previewIcon = computed(() => (tagForm.form.icon ? CURATED_ICONS[tagForm.form.icon] ?? null : null))
const previewChipStyle = computed(() => ({
  color: tagForm.form.color,
  background: `color-mix(in srgb, ${tagForm.form.color} 14%, transparent)`,
  borderColor: `color-mix(in srgb, ${tagForm.form.color} 30%, transparent)`,
}))

function openCreate() {
  editing.value = null
  tagForm.resetTo({ name: '', color: DEFAULT_TAG_COLOR, icon: '' })
  dialogOpen.value = true
}

function openEdit(tag: Tag) {
  editing.value = tag
  tagForm.resetTo({ name: tag.name, color: tag.color, icon: tag.icon })
  dialogOpen.value = true
}

async function confirmDelete(tag: Tag) {
  if (!confirm(`确认删除标签"${tag.name}"?该标签会从所有任务上摘除。`)) return
  try {
    const affected = await store.deleteTag(tag.id)
    if (affected > 0) console.info(`已从 ${affected} 个任务上摘除标签`)
  } catch (e) {
    alert(e instanceof Error ? e.message : '删除失败')
  }
}

// 拖拽排序
const { draggedIndex, dragOverIndex, onDragStart, onDragOver, onDrop, onDragEnd } = useDragSort(
  computed(() => store.tags),
  async (newOrder) => {
    try {
      for (let i = 0; i < newOrder.length; i++) {
        const t = newOrder[i]
        if (t.sortOrder !== i) await store.updateTag(t.id, { sort_order: i })
      }
    } catch (e) {
      alert(e instanceof Error ? e.message : '排序保存失败,已回滚')
      store.fetchTags(true)
    }
  },
)

onMounted(() => store.fetchTags())
</script>

<template>
  <div class="tag-manager-page">
    <header class="page-header">
      <div>
        <h1>标签管理</h1>
        <p class="page-desc">为任务自定义彩色标签,支持图标与排序;改名/删除会同步更新所有任务上的标签。</p>
      </div>
      <button class="btn-primary desktop-only" @click="openCreate">+ 新建标签</button>
    </header>

    <PageShell
      :loading="store.loading && store.tags.length === 0"
      :error="store.error"
      :empty="store.tags.length === 0"
      empty-title="还没有任何标签"
      empty-description='点右上角"新建"开始吧。'
    >
    <ul class="tag-list motion-stagger">
      <li
        v-for="(tag, index) in store.tags"
        :key="tag.id"
        class="tag-row"
        :draggable="true"
        :class="{ 'drag-over': dragOverIndex === index && draggedIndex !== index }"
        @dragstart="onDragStart(index, $event)"
        @dragover.prevent="onDragOver(index)"
        @drop.prevent="onDrop(index)"
        @dragend="onDragEnd"
      >
        <span class="drag-handle" title="拖拽排序">⋮⋮</span>
        <TagChip :name="tag.name" size="md" />
        <div class="tag-actions">
          <button class="btn-link" @click="openEdit(tag)">编辑</button>
          <button class="btn-link danger" @click="confirmDelete(tag)">删除</button>
        </div>
      </li>
    </ul>
    </PageShell>

    <MobileFab label="新建标签" @click="openCreate"><Plus :size="24" /></MobileFab>

    <BaseDialog v-model:visible="dialogOpen" :title="editing ? '编辑标签' : '新建标签'" max-width="480px">
      <div class="form-group">
        <label>名称 *</label>
        <input v-model.trim="tagForm.form.name" type="text" class="form-input" maxlength="32" placeholder="如:工作 / 学习 / 紧急" />
      </div>
      <div class="form-group">
        <label>颜色</label>
        <ColorPicker v-model="tagForm.form.color" />
      </div>
      <div class="form-group">
        <label>图标</label>
        <IconPicker v-model="tagForm.form.icon" />
      </div>
      <div class="dialog-preview">
        <span class="preview-label">预览:</span>
        <span class="preview-chip" :style="previewChipStyle">
          <component v-if="previewIcon" :is="previewIcon" :size="14" />
          <span>{{ tagForm.form.name || '未命名' }}</span>
        </span>
      </div>
      <div v-if="tagForm.error.value" class="error-message">{{ tagForm.error.value }}</div>
      <template #footer="{ close }">
        <button class="btn" @click="close">取消</button>
        <button class="btn btn-primary" :disabled="tagForm.submitting.value" @click="tagForm.handleSubmit">
          {{ tagForm.submitting.value ? '保存中...' : '保存' }}
        </button>
      </template>
    </BaseDialog>
  </div>
</template>

<style scoped>
.tag-manager-page {
  padding: 24px;
  max-width: 880px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.desktop-only { display: block; }
.mobile-only { display: none; }

@media (max-width: 767px) {
  .desktop-only { display: none; }
  .mobile-only { display: block; }
  .tag-manager-page { padding-bottom: 80px; }
}

.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.page-header h1 { margin: 0 0 4px; font-size: 20px; }
.page-desc { margin: 0; font-size: 13px; color: var(--color-text-muted); }

.tag-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

@media (max-width: 767px) {
  .tag-list { gap: 12px; }
  .tag-list :deep(.tag-chip) { font-size: 13px; padding: 4px 12px; max-width: none; flex: 1; }
}

.tag-row {
  display: grid;
  grid-template-columns: 24px 1fr auto;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  background: var(--color-surface);
}

@media (max-width: 767px) {
  .tag-row { display: flex; align-items: center; gap: 12px; padding: 14px 16px; border-radius: 12px; box-shadow: 0 1px 3px rgba(0,0,0,0.08); }
  .drag-handle { display: none; }
  .tag-actions { gap: 8px; margin-left: auto; }
  .tag-actions .btn-link { padding: 8px 12px; font-size: 13px; font-weight: 500; border-radius: 8px; background: var(--color-surface); border: 1px solid var(--color-border); transition: background-color var(--motion-duration-fast) var(--motion-ease-standard), border-color var(--motion-duration-fast) var(--motion-ease-standard), color var(--motion-duration-fast) var(--motion-ease-standard); }
  .tag-actions .btn-link:hover { background: var(--color-primary); color: white; border-color: var(--color-primary); }
  .tag-actions .btn-link.danger:hover { background: var(--color-danger); border-color: var(--color-danger); color: white; }
}

.tag-row.drag-over { border-color: var(--color-primary); box-shadow: 0 0 0 1px var(--color-primary) inset; }

.drag-handle { cursor: grab; color: var(--color-text-muted); user-select: none; }
.drag-handle:active { cursor: grabbing; }

.tag-actions { display: flex; gap: 4px; }

.btn-link { background: transparent; border: 0; color: var(--color-primary); cursor: pointer; font-size: 13px; padding: 4px 8px; }
.btn-link:hover { text-decoration: underline; }
.btn-link.danger { color: var(--color-danger); }

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
</style>

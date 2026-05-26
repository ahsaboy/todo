<template>
  <div class="color-picker">
    <div class="preset-grid">
      <button
        v-for="c in presetColors"
        :key="c"
        type="button"
        class="color-swatch"
        :class="{ active: modelValue.toLowerCase() === c.toLowerCase() }"
        :style="{ background: c }"
        :aria-label="`选择颜色 ${c}`"
        @click="onPick(c)"
      >
        <span v-if="modelValue.toLowerCase() === c.toLowerCase()" class="check">✓</span>
      </button>
    </div>
    <div class="custom-row">
      <span class="custom-label">自定义</span>
      <input
        ref="hexInput"
        type="text"
        class="hex-input"
        :value="modelValue"
        maxlength="7"
        placeholder="#3b82f6"
        @input="onHexInput"
      />
      <span class="preview" :style="{ background: isValid ? modelValue : 'transparent', borderColor: isValid ? modelValue : 'var(--color-border)' }"></span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { PRESET_COLORS } from '@/shared/icons/curated'

const props = defineProps<{
  modelValue: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const presetColors = PRESET_COLORS

const isValid = computed(() => /^#[0-9a-fA-F]{6}$/.test(props.modelValue))

function onPick(c: string) {
  emit('update:modelValue', c)
}

function onHexInput(event: Event) {
  const v = (event.target as HTMLInputElement).value
  // 透传原始值给 v-model,TagsPage 在提交时统一做校验;
  // preview 通过 isValid 计算属性保护,非法 hex 会显示为透明
  emit('update:modelValue', v)
}
</script>

<style scoped>
.color-picker {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.preset-grid {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 8px;
}

.color-swatch {
  width: 100%;
  aspect-ratio: 1;
  border-radius: 8px;
  border: 2px solid transparent;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 14px;
  font-weight: 600;
  transition: transform 0.12s ease, box-shadow 0.12s ease;
}

.color-swatch:hover {
  transform: scale(1.06);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.15);
}

.color-swatch.active {
  border-color: var(--color-text);
  box-shadow: 0 0 0 2px var(--color-surface), 0 0 0 4px var(--color-text);
}

.check {
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.4);
}

.custom-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.custom-label {
  font-size: 12px;
  color: var(--color-text-muted);
}

.hex-input {
  flex: 1;
  padding: 6px 10px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  font-size: 13px;
  font-family: ui-monospace, monospace;
  background: var(--color-surface);
}

.preview {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  border: 1px solid var(--color-border);
  flex-shrink: 0;
}
</style>

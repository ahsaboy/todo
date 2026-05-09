<template>
  <div class="json-editor" :class="{ 'is-disabled': disabled }">
    <div class="json-editor-toolbar">
      <button type="button" class="btn-format" :disabled="disabled" @click="format">格式化</button>
    </div>
    <div class="json-editor-container" :style="{ '--json-editor-rows': rows }">
      <pre class="json-highlight" aria-hidden="true" v-html="highlighted"></pre>
      <textarea
        ref="textareaRef"
        :id="id"
        :value="internalValue"
        :placeholder="placeholder"
        :rows="rows"
        class="json-textarea"
        :disabled="disabled"
        spellcheck="false"
        @input="onInput"
        @scroll="onScroll"
        @focus="$emit('focus', $event)"
        @blur="$emit('blur', $event)"
      ></textarea>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, nextTick } from 'vue'

const props = withDefaults(
  defineProps<{
    modelValue: string
    id?: string
    placeholder?: string
    rows?: number
    disabled?: boolean
  }>(),
  {
    id: '',
    placeholder: '',
    rows: 4,
    disabled: false,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
  focus: [event: FocusEvent]
  blur: [event: FocusEvent]
}>()

const textareaRef = ref<HTMLTextAreaElement>()
const highlighted = ref('')
const internalValue = ref(props.modelValue)
let isTyping = false

onMounted(() => {
  if (props.modelValue) {
    internalValue.value = props.modelValue
  }
  highlight()
})

watch(
  () => props.modelValue,
  (val) => {
    if (isTyping) return
    internalValue.value = val
    highlight()
  },
)

function emitValue(val: string) {
  emit('update:modelValue', val)
}

function format() {
  if (props.disabled) return
  if (!internalValue.value.trim()) return
  try {
    const parsed = JSON.parse(internalValue.value)
    internalValue.value = JSON.stringify(parsed, null, 2)
    emitValue(internalValue.value)
    highlight()
    nextTick(syncScroll)
  } catch {
    // invalid JSON, ignore
  }
}

function onInput(e: Event) {
  if (props.disabled) return
  const val = (e.target as HTMLTextAreaElement).value
  isTyping = true
  internalValue.value = val
  emitValue(val)
  highlight()
  requestAnimationFrame(() => {
    isTyping = false
  })
}

function syncScroll() {
  if (!textareaRef.value) return
  const pre = textareaRef.value.previousElementSibling as HTMLElement | null
  if (pre) {
    pre.scrollTop = textareaRef.value.scrollTop
    pre.scrollLeft = textareaRef.value.scrollLeft
  }
}

function onScroll() {
  syncScroll()
}

function highlight() {
  highlighted.value = highlightJson(internalValue.value)
}

function highlightJson(raw: string): string {
  if (!raw) return ''
  const escaped = escapeHtml(raw)
  return escaped.replace(
    /("(?:\\.|[^"\\])*")\s*(:)?|(\btrue\b|\bfalse\b|\bnull\b)|(-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)|([{}[\],])/g,
    (match, str, colon, bool, num, punct) => {
      if (str) {
        if (colon) {
          return `<span class="json-key">${str}</span>:`
        }
        return `<span class="json-string">${str}</span>`
      }
      if (bool) return `<span class="json-bool">${bool}</span>`
      if (num) return `<span class="json-number">${num}</span>`
      if (punct) return `<span class="json-punct">${punct}</span>`
      return match
    },
  )
}

function escapeHtml(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}

defineExpose({ textareaRef, format })
</script>

<style scoped>
.json-editor {
  position: relative;
}

.json-editor-toolbar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 8px;
}

.btn-format {
  padding: 6px 12px;
  font-size: 12px;
  border: 1px solid var(--color-border);
  border-radius: 999px;
  background: var(--color-surface);
  color: var(--color-text);
  cursor: pointer;
  transition: background-color 150ms, border-color 150ms, color 150ms;
}

.btn-format:hover {
  background: var(--color-surface-muted);
  border-color: var(--color-primary);
}

.btn-format:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.json-editor-container {
  display: grid;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  background: var(--color-surface-muted);
  overflow: hidden;
  transition: border-color 150ms, box-shadow 150ms, background-color 150ms;
}

.json-editor:focus-within .json-editor-container {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgb(37 111 108 / 12%);
}

.json-editor.is-disabled {
  opacity: 0.8;
}

.json-highlight,
.json-textarea {
  grid-area: 1 / 1;
  margin: 0;
  padding: 10px 12px;
  min-height: calc(var(--json-editor-rows, 4) * 1.6em + 20px);
  font-family:
    ui-monospace, SFMono-Regular, Menlo, Consolas, 'Liberation Mono', monospace;
  font-size: 13px;
  line-height: 1.6;
  tab-size: 2;
  white-space: pre;
  overflow: auto;
  scrollbar-gutter: stable;
}

.json-highlight {
  pointer-events: none;
  user-select: none;
  background: transparent;
  color: transparent;
}

.json-highlight {
  scrollbar-width: none;
}

.json-highlight::-webkit-scrollbar {
  display: none;
}

.json-highlight :deep(.json-key) {
  color: var(--color-primary);
}

.json-highlight :deep(.json-string) {
  color: var(--color-success);
}

.json-highlight :deep(.json-number) {
  color: var(--color-warning);
}

.json-highlight :deep(.json-bool) {
  color: #8f5cf6;
}

.json-highlight :deep(.json-punct) {
  color: var(--color-text-muted);
}

.json-textarea {
  width: 100%;
  border: 0;
  resize: vertical;
  background: transparent;
  color: transparent;
  caret-color: var(--color-primary);
  outline: none;
  text-shadow: none;
  -webkit-text-fill-color: transparent;
}

.json-textarea::selection {
  background: rgb(37 111 108 / 20%);
}

.json-textarea::placeholder {
  color: var(--color-text-muted);
}

.json-textarea:disabled {
  cursor: not-allowed;
}
</style>

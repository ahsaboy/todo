<template>
  <div
    class="json-editor"
    :class="{ 'is-disabled': disabled, 'is-focused': isFocused, 'is-invalid': hasInvalidJson }"
  >
    <div class="json-editor-toolbar">
      <button type="button" class="btn-format" :disabled="disabled" @click="format">格式化</button>
    </div>
    <div class="json-editor-container">
      <div class="json-highlight" aria-hidden="true">
        <pre ref="highlightContentRef" class="json-highlight-content"></pre>
      </div>
      <textarea
        ref="textareaRef"
        :id="id"
        :value="internalValue"
        :placeholder="placeholder"
        :rows="rows"
        wrap="off"
        class="json-textarea"
        :disabled="disabled"
        spellcheck="false"
        @input="onInput"
        @scroll="onScroll"
        @focus="handleFocus"
        @blur="handleBlur"
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
const highlightContentRef = ref<HTMLElement>()
const highlighted = ref('')
const internalValue = ref(props.modelValue)
const isFocused = ref(false)
const hasInvalidJson = ref(false)
let isTyping = false

onMounted(() => {
  if (props.modelValue) {
    internalValue.value = props.modelValue
  }
  highlight()
  nextTick(() => {
    syncHeight()
    syncScroll()
  })
})

watch(
  () => props.modelValue,
  (val) => {
    if (isTyping) return
    internalValue.value = val
    highlight()
    nextTick(() => {
      syncHeight()
      syncScroll()
    })
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
    nextTick(() => {
      syncHeight()
      syncScroll()
    })
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
  syncHeight()
  requestAnimationFrame(() => {
    isTyping = false
  })
}

function syncHeight() {
  if (!textareaRef.value) return
  textareaRef.value.style.height = 'auto'
  textareaRef.value.style.height = `${textareaRef.value.scrollHeight}px`
}

function syncScroll() {
  if (!textareaRef.value || !highlightContentRef.value) return
  const { scrollTop, scrollLeft } = textareaRef.value
  highlightContentRef.value.style.transform = `translate(${-scrollLeft}px, ${-scrollTop}px)`
}

function onScroll() {
  syncScroll()
}

function highlight() {
  const raw = internalValue.value

  if (!raw) {
    hasInvalidJson.value = false
    highlighted.value = ''
    renderHighlight()
    return
  }

  try {
    JSON.parse(raw)
    hasInvalidJson.value = false
    highlighted.value = tokenizeJson(raw)
  } catch {
    hasInvalidJson.value = true
    highlighted.value = wrapToken('json-invalid', raw)
  }
  renderHighlight()
}

function renderHighlight() {
  if (!highlightContentRef.value) return
  highlightContentRef.value.innerHTML = highlighted.value
}

function handleFocus(event: FocusEvent) {
  isFocused.value = true
  emit('focus', event)
}

function handleBlur(event: FocusEvent) {
  isFocused.value = false
  emit('blur', event)
}

function escapeHtml(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}

type JsonContext =
  | { type: 'object'; phase: 'keyOrEnd' | 'colon' | 'value' | 'commaOrEnd' }
  | { type: 'array'; phase: 'valueOrEnd' | 'commaOrEnd' }

function tokenizeJson(raw: string): string {
  const parts: string[] = []
  const stack: JsonContext[] = []
  let index = 0

  while (index < raw.length) {
    const char = raw[index]

    if (isWhitespace(char)) {
      parts.push(char)
      index += 1
      continue
    }

    if (char === '"') {
      const token = readString(raw, index)
      if (isObjectKeyPosition(stack)) {
        parts.push(wrapToken('json-key', token.value))
        setObjectPhase(stack, 'colon')
      } else {
        parts.push(wrapToken('json-string', token.value))
        advanceAfterValue(stack)
      }
      index = token.nextIndex
      continue
    }

    if (char === '{') {
      parts.push(wrapToken('json-punct', char))
      stack.push({ type: 'object', phase: 'keyOrEnd' })
      index += 1
      continue
    }

    if (char === '[') {
      parts.push(wrapToken('json-punct', char))
      stack.push({ type: 'array', phase: 'valueOrEnd' })
      index += 1
      continue
    }

    if (char === '}' || char === ']') {
      parts.push(wrapToken('json-punct', char))
      stack.pop()
      advanceAfterValue(stack)
      index += 1
      continue
    }

    if (char === ':') {
      parts.push(wrapToken('json-punct', char))
      setObjectPhase(stack, 'value')
      index += 1
      continue
    }

    if (char === ',') {
      parts.push(wrapToken('json-punct', char))
      setNextEntryPhase(stack)
      index += 1
      continue
    }

    if (char === 't' && raw.startsWith('true', index)) {
      parts.push(wrapToken('json-bool', 'true'))
      advanceAfterValue(stack)
      index += 4
      continue
    }

    if (char === 'f' && raw.startsWith('false', index)) {
      parts.push(wrapToken('json-bool', 'false'))
      advanceAfterValue(stack)
      index += 5
      continue
    }

    if (char === 'n' && raw.startsWith('null', index)) {
      parts.push(wrapToken('json-null', 'null'))
      advanceAfterValue(stack)
      index += 4
      continue
    }

    const numberToken = readNumber(raw, index)
    parts.push(wrapToken('json-number', numberToken.value))
    advanceAfterValue(stack)
    index = numberToken.nextIndex
  }

  return parts.join('')
}

function readString(raw: string, startIndex: number) {
  let index = startIndex + 1
  let escaped = false

  while (index < raw.length) {
    const char = raw[index]

    if (escaped) {
      escaped = false
      index += 1
      continue
    }

    if (char === '\\') {
      escaped = true
      index += 1
      continue
    }

    if (char === '"') {
      index += 1
      break
    }

    index += 1
  }

  return {
    value: raw.slice(startIndex, index),
    nextIndex: index,
  }
}

function readNumber(raw: string, startIndex: number) {
  let index = startIndex

  while (index < raw.length && isNumberChar(raw[index])) {
    index += 1
  }

  return {
    value: raw.slice(startIndex, index),
    nextIndex: index,
  }
}

function wrapToken(className: string, value: string) {
  return `<span class="${className}">${escapeHtml(value)}</span>`
}

function isObjectKeyPosition(stack: JsonContext[]) {
  const current = getCurrentContext(stack)
  return current?.type === 'object' && current.phase === 'keyOrEnd'
}

function setObjectPhase(stack: JsonContext[], phase: Extract<JsonContext, { type: 'object' }>['phase']) {
  const current = getCurrentContext(stack)
  if (current?.type === 'object') {
    current.phase = phase
  }
}

function advanceAfterValue(stack: JsonContext[]) {
  const current = getCurrentContext(stack)
  if (!current) return

  if (current.type === 'object' && current.phase === 'value') {
    current.phase = 'commaOrEnd'
  }

  if (current.type === 'array' && current.phase === 'valueOrEnd') {
    current.phase = 'commaOrEnd'
  }
}

function setNextEntryPhase(stack: JsonContext[]) {
  const current = getCurrentContext(stack)
  if (!current) return

  if (current.type === 'object') {
    current.phase = 'keyOrEnd'
  }

  if (current.type === 'array') {
    current.phase = 'valueOrEnd'
  }
}

function getCurrentContext(stack: JsonContext[]) {
  return stack[stack.length - 1]
}

function isWhitespace(char: string) {
  return char === ' ' || char === '\n' || char === '\r' || char === '\t'
}

function isNumberChar(char: string) {
  return (
    char === '-' ||
    char === '+' ||
    char === '.' ||
    char === 'e' ||
    char === 'E' ||
    (char >= '0' && char <= '9')
  )
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
  transition: background-color var(--motion-duration-fast), border-color var(--motion-duration-fast), color var(--motion-duration-fast);
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
  --json-editor-font-family:
    ui-monospace, SFMono-Regular, Menlo, Consolas, 'Liberation Mono', monospace;
  --json-editor-font-size: 13px;
  --json-editor-line-height: 1.6;
  position: relative;
  display: grid;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  background: var(--color-surface-muted);
  overflow: hidden;
  font-family: var(--json-editor-font-family);
  font-size: var(--json-editor-font-size);
  line-height: var(--json-editor-line-height);
  transition: border-color var(--motion-duration-fast), box-shadow var(--motion-duration-fast), background-color var(--motion-duration-fast);
}

.json-editor:focus-within .json-editor-container {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-primary) 12%, transparent);
}

.json-editor.is-disabled {
  opacity: 0.8;
}

.json-highlight,
.json-textarea {
  margin: 0;
  font-family: inherit;
  font-size: inherit;
  line-height: inherit;
  font-weight: 400;
  tab-size: 2;
  letter-spacing: 0;
  font-kerning: none;
  font-variant-ligatures: none;
  font-synthesis: none;
}

.json-highlight {
  position: absolute;
  inset: 0;
  pointer-events: none;
  user-select: none;
  overflow: hidden;
  min-height: 0;
  padding: 10px 12px;
  background: transparent;
  color: transparent;
  opacity: 1;
  transition: opacity var(--motion-duration-fast) var(--motion-ease-standard);
}

.json-highlight-content {
  margin: 0;
  font-family: inherit;
  font-size: inherit;
  line-height: inherit;
  font-weight: inherit;
  letter-spacing: inherit;
  font-kerning: inherit;
  font-variant-ligatures: inherit;
  font-synthesis: inherit;
  tab-size: inherit;
  white-space: pre;
  transform: translate(0, 0);
  will-change: transform;
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
  color: var(--color-info);
}

.json-highlight :deep(.json-null) {
  color: var(--color-text-muted);
}

.json-highlight :deep(.json-punct) {
  color: var(--color-text-muted);
}

.json-highlight :deep(.json-invalid) {
  color: var(--color-danger);
}

.json-textarea {
  position: relative;
  display: block;
  width: 100%;
  min-height: 0;
  padding: 10px 12px;
  border: 0;
  resize: vertical;
  background: transparent;
  color: transparent;
  caret-color: var(--color-text);
  outline: none;
  text-shadow: none;
  -webkit-text-fill-color: transparent;
  white-space: pre;
  word-break: normal;
  overflow-wrap: normal;
  overflow-x: auto;
  overflow-y: hidden;
  scrollbar-gutter: stable;
}

.json-editor.is-focused .json-highlight {
  opacity: 0;
}

.json-editor.is-focused .json-textarea {
  color: var(--color-text);
  -webkit-text-fill-color: var(--color-text);
}

.json-editor.is-focused.is-invalid .json-textarea {
  color: var(--color-danger);
  -webkit-text-fill-color: var(--color-danger);
  caret-color: var(--color-danger);
}

.json-textarea::selection {
  background: color-mix(in srgb, var(--color-primary) 20%, transparent);
}

.json-textarea::placeholder {
  color: var(--color-text-muted);
}

.json-textarea:disabled {
  cursor: not-allowed;
}
</style>

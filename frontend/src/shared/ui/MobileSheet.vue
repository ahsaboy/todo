<template>
  <Teleport to="body">
    <Transition name="overlay-motion" appear>
      <div v-if="visible" class="sheet-overlay" @click.self="$emit('update:visible', false)">
        <div class="sheet-content motion-panel motion-panel--bottom">
          <div class="sheet-handle"></div>
          <div class="sheet-body">
            <slot></slot>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
defineProps<{
  visible: boolean
}>()

defineEmits<{
  'update:visible': [value: boolean]
}>()
</script>

<style scoped>
.sheet-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--el-mask-color);
  z-index: 1000;
  display: flex;
  align-items: flex-end;
}

.sheet-content {
  width: 100%;
  max-height: 80vh;
  background: var(--color-surface);
  border-radius: 16px 16px 0 0;
  overflow: hidden;
}

.sheet-handle {
  width: 40px;
  height: 4px;
  background: var(--color-border);
  border-radius: 2px;
  margin: 12px auto;
}

.sheet-body {
  padding: 0 20px 20px;
  overflow-y: auto;
  max-height: calc(80vh - 40px);
}
</style>

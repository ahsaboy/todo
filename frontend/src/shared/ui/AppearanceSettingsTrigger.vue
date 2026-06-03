<script setup lang="ts">
import { ref } from 'vue'
import { Settings2 } from 'lucide-vue-next'
import AppearancePanel from './AppearancePanel.vue'

const isOpen = ref(false)
const wrapperRef = ref<HTMLElement | null>(null)

function toggle() {
  isOpen.value = !isOpen.value
}
</script>

<template>
  <div ref="wrapperRef" class="appearance-trigger">
    <button
      class="btn-icon theme-toggle-btn"
      type="button"
      aria-label="外观设置"
      @click="toggle"
    >
      <Settings2 :size="16" />
    </button>
    <AppearancePanel
      :visible="isOpen"
      :trigger-ref="wrapperRef"
      @update:visible="isOpen = $event"
    />
  </div>
</template>

<style scoped>
.appearance-trigger {
  position: relative;
  display: flex;
  align-items: center;
}
</style>

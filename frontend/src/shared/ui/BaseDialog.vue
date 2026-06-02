<script setup lang="ts">
import { computed, ref, watch, onMounted, onUnmounted } from 'vue'

const props = withDefaults(defineProps<{
  visible: boolean
  title?: string
  maxWidth?: string
  closeOnOverlay?: boolean
}>(), {
  maxWidth: '440px',
  closeOnOverlay: true,
})

const emit = defineEmits<{
  'update:visible': [value: boolean]
  close: []
}>()

const isMobile = ref(false)
let mq: MediaQueryList | null = null

function syncMobile() {
  isMobile.value = mq?.matches ?? false
}

const panelClass = computed(() =>
  isMobile.value ? 'motion-panel--bottom' : 'motion-panel--scale',
)

function handleOverlayClick() {
  if (props.closeOnOverlay) close()
}

function close() {
  emit('update:visible', false)
  emit('close')
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') close()
}

watch(() => props.visible, (val) => {
  if (val) document.addEventListener('keydown', onKeydown)
  else document.removeEventListener('keydown', onKeydown)
}, { immediate: true })

onMounted(() => {
  mq = window.matchMedia('(max-width: 767px)')
  syncMobile()
  mq.addEventListener('change', syncMobile)
})

onUnmounted(() => {
  document.removeEventListener('keydown', onKeydown)
  mq?.removeEventListener('change', syncMobile)
})
</script>

<template>
  <Teleport to="body">
    <Transition name="overlay-motion" appear>
      <div
        v-if="visible"
        class="admin-modal-overlay"
        :class="{ 'admin-modal-overlay--mobile': isMobile }"
        @click.self="handleOverlayClick"
      >
        <div
          class="admin-modal motion-panel"
          :class="[panelClass, { 'admin-modal--sheet': isMobile }]"
          :style="isMobile ? undefined : { maxWidth }"
          role="dialog"
          aria-modal="true"
          :aria-label="title"
        >
          <div v-if="title || $slots.header">
            <slot name="header">
              <h3>{{ title }}</h3>
            </slot>
          </div>

          <div class="dialog-body">
            <slot />
          </div>

          <div v-if="$slots.footer" class="modal-actions">
            <slot name="footer" :close="close" />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.dialog-body {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.admin-modal-overlay--mobile {
  align-items: flex-end;
}

.admin-modal--sheet {
  width: 100%;
  max-width: 100% !important;
  border-radius: 16px 16px 0 0;
  padding: 1.25rem 1.25rem calc(1.25rem + env(safe-area-inset-bottom));
}
</style>

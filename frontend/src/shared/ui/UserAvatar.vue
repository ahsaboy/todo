<script setup lang="ts">
import { ref, computed, watch } from 'vue'

const props = withDefaults(defineProps<{
  avatarUrl?: string
  username?: string
  size?: 'sm' | 'md' | 'lg' | 'xl'
}>(), {
  size: 'md',
})

const imgError = ref(false)

watch(() => props.avatarUrl, () => { imgError.value = false })

const showImage = computed(() => props.avatarUrl && !imgError.value)

const initial = computed(() => {
  const name = props.username ?? ''
  return name.charAt(0).toUpperCase() || '?'
})

const sizeMap = { sm: 32, md: 48, lg: 72, xl: 96 }
const fontSizeMap = { sm: 13, md: 16, lg: 24, xl: 32 }

const sizeStyle = computed(() => {
  const px = sizeMap[props.size]
  return { width: `${px}px`, height: `${px}px`, fontSize: `${fontSizeMap[props.size]}px` }
})

function onImgError() {
  imgError.value = true
}
</script>

<template>
  <div class="user-avatar" :class="`user-avatar--${size}`" :style="sizeStyle">
    <img
      v-if="showImage"
      :src="avatarUrl"
      :alt="username || 'avatar'"
      class="user-avatar-img"
      @error="onImgError"
    />
    <span v-else class="user-avatar-initial">{{ initial }}</span>
  </div>
</template>

<style scoped>
.user-avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-pill);
  background: var(--color-primary);
  color: var(--color-btn-primary-text, #fff);
  font-weight: 600;
  flex-shrink: 0;
  overflow: hidden;
  user-select: none;
}

.user-avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.user-avatar-initial {
  line-height: 1;
}
</style>

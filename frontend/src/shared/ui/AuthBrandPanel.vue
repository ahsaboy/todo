<script setup lang="ts">
import { ref } from 'vue'
import { useParticleNetwork } from '@/shared/composables/useParticleNetwork'

withDefaults(defineProps<{
  tagline?: string
  appName?: string
}>(), {
  tagline: '让每一个任务井然有序',
  appName: 'TODO',
})

const canvasRef = ref<HTMLCanvasElement | null>(null)
const { particleCount } = useParticleNetwork(canvasRef, {
  particleCount: 200,
  connectionDistance: 120,
  repulseRadius: 800,
  repulseStrength: 0.08,
})
</script>

<template>
  <div class="auth-brand">
    <!-- Mobile: compact header -->
    <div class="auth-brand__mobile-header">
      <img src="/favicon.svg" alt="" class="auth-brand__mobile-logo" />
      <span class="auth-brand__mobile-name">{{ appName }}</span>
    </div>

    <!-- Desktop: full brand experience -->
    <div class="auth-brand__content">
      <canvas ref="canvasRef" class="auth-brand__canvas"></canvas>

      <div class="auth-brand__logo-wrap">
        <img src="/favicon.svg" :alt="appName" class="auth-brand__logo" />
      </div>

      <h2 class="auth-brand__title">{{ appName }}</h2>
      <p class="auth-brand__tagline">{{ tagline }}</p>
    </div>
  </div>
</template>

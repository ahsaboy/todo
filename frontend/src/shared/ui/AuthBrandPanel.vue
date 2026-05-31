<script setup lang="ts">
withDefaults(defineProps<{
  tagline?: string
  appName?: string
}>(), {
  tagline: '让每一个任务井然有序',
  appName: 'TODO',
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
      <div class="auth-brand__logo-orbit auth-brand__logo-orbit--enter auth-brand__logo-orbit--interactive">
        <img src="/favicon.svg" :alt="appName" class="auth-brand__logo auth-brand__logo--glow auth-brand__logo--enter" />
        <div class="auth-brand__orbit-ring auth-brand__orbit-ring--enter"></div>
        <div class="auth-brand__orbit-dot auth-brand__orbit-dot--1 auth-brand__orbit-dot--enter"></div>
        <div class="auth-brand__orbit-dot auth-brand__orbit-dot--2 auth-brand__orbit-dot--enter"></div>
        <div class="auth-brand__orbit-dot auth-brand__orbit-dot--3 auth-brand__orbit-dot--enter"></div>
      </div>

      <h2 class="auth-brand__title auth-brand__title--enter">{{ appName }}</h2>
      <p class="auth-brand__tagline auth-brand__tagline--enter">{{ tagline }}</p>

      <div class="auth-brand__particles" aria-hidden="true">
        <span class="auth-brand__particle auth-brand__particle--1"></span>
        <span class="auth-brand__particle auth-brand__particle--2"></span>
        <span class="auth-brand__particle auth-brand__particle--3"></span>
        <span class="auth-brand__particle auth-brand__particle--4"></span>
        <span class="auth-brand__particle auth-brand__particle--5"></span>
      </div>
    </div>
  </div>
</template>

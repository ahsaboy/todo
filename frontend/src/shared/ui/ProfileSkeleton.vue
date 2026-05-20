<template>
  <div class="sk-profile">
    <SkeletonBlock width="120px" height="20px" />

    <div class="sk-form-section">
      <SkeletonBlock width="80px" height="16px" />
      <div class="sk-form-field">
        <SkeletonBlock width="40px" height="12px" />
        <SkeletonBlock width="100%" height="36px" borderRadius="6px" />
      </div>
      <div class="sk-form-field">
        <SkeletonBlock width="40px" height="12px" />
        <SkeletonBlock width="100%" height="36px" borderRadius="6px" />
      </div>
      <div class="sk-form-actions">
        <SkeletonBlock width="64px" height="36px" borderRadius="6px" />
      </div>
    </div>

    <div class="sk-form-section">
      <SkeletonBlock width="80px" height="16px" />
      <div class="sk-form-field">
        <SkeletonBlock width="56px" height="12px" />
        <SkeletonBlock width="100%" height="36px" borderRadius="6px" />
      </div>
      <div class="sk-form-field">
        <SkeletonBlock width="56px" height="12px" />
        <SkeletonBlock width="100%" height="36px" borderRadius="6px" />
      </div>
      <div class="sk-form-field">
        <SkeletonBlock width="72px" height="12px" />
        <SkeletonBlock width="100%" height="36px" borderRadius="6px" />
      </div>
      <div class="sk-form-actions">
        <SkeletonBlock width="80px" height="36px" borderRadius="6px" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import SkeletonBlock from '@/shared/ui/SkeletonBlock.vue'
</script>

<style scoped>
.sk-profile {
  display: flex;
  flex-direction: column;
  gap: 24px;
  max-width: 640px;
}

.sk-form-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 20px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
}

.sk-form-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.sk-form-actions {
  display: flex;
  justify-content: flex-end;
}
</style>

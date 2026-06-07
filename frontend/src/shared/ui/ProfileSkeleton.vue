<template>
  <div class="sk-profile">
    <!-- 头像卡片骨架 -->
    <div class="sk-form-section sk-avatar-section">
      <SkeletonBlock width="96px" height="96px" borderRadius="var(--radius-pill)" />
      <SkeletonBlock width="120px" height="20px" />
      <div class="sk-meta-rows">
        <div class="sk-meta-row">
          <SkeletonBlock width="48px" height="12px" />
          <SkeletonBlock width="32px" height="12px" />
        </div>
        <div class="sk-meta-row">
          <SkeletonBlock width="56px" height="12px" />
          <SkeletonBlock width="96px" height="12px" />
        </div>
      </div>
    </div>

    <!-- 邮箱骨架 -->
    <div class="sk-form-section">
      <SkeletonBlock width="40px" height="16px" />
      <div class="sk-form-field">
        <SkeletonBlock width="32px" height="12px" />
        <SkeletonBlock width="100%" height="36px" borderRadius="6px" />
      </div>
      <div class="sk-form-actions">
        <SkeletonBlock width="64px" height="36px" borderRadius="6px" />
      </div>
    </div>

    <!-- OAuth 骨架 -->
    <div class="sk-form-section">
      <SkeletonBlock width="64px" height="16px" />
      <div class="sk-oauth-item">
        <SkeletonBlock width="20px" height="20px" borderRadius="var(--radius-pill)" />
        <SkeletonBlock width="56px" height="12px" />
        <SkeletonBlock width="80px" height="12px" />
      </div>
      <div class="sk-oauth-item">
        <SkeletonBlock width="20px" height="20px" borderRadius="var(--radius-pill)" />
        <SkeletonBlock width="48px" height="12px" />
        <SkeletonBlock width="64px" height="12px" />
      </div>
    </div>

    <!-- 密码骨架 -->
    <div class="sk-form-section">
      <SkeletonBlock width="64px" height="16px" />
      <div class="sk-form-field">
        <SkeletonBlock width="56px" height="12px" />
        <SkeletonBlock width="100%" height="36px" borderRadius="6px" />
      </div>
      <div class="sk-form-field">
        <SkeletonBlock width="48px" height="12px" />
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
  gap: 1rem;
  max-width: 640px;
}

.sk-form-section {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1.25rem;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
}

.sk-avatar-section {
  align-items: center;
  gap: 0.75rem;
}

.sk-meta-rows {
  width: 100%;
  max-width: 320px;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.sk-meta-row {
  display: flex;
  justify-content: space-between;
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

.sk-oauth-item {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.625rem 0.75rem;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
}
</style>

<template>
  <div class="profile-info">
    <div class="profile-avatar-section">
      <UserAvatar :avatar-url="user?.avatarUrl" :username="user?.username" size="xl" />
      <h3 class="profile-username">{{ user?.username || '—' }}</h3>
    </div>
    <div class="profile-meta">
      <div class="meta-row">
        <span class="meta-label">用户 ID</span>
        <span class="meta-value">{{ user?.id }}</span>
      </div>
      <div class="meta-row">
        <span class="meta-label">注册时间</span>
        <span class="meta-value">{{ formatDate(user?.createdAt) }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { User } from '@/entities/user/model'
import UserAvatar from '@/shared/ui/UserAvatar.vue'
import { formatDateFull as formatDate } from '@/shared/utils/date'

defineProps<{
  user: User | null
}>()
</script>

<style scoped>
.profile-info {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1.25rem;
}

.profile-avatar-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
}

.profile-username {
  margin: 0;
  font-size: var(--text-xl);
  font-weight: 600;
  color: var(--color-text);
}

.profile-meta {
  width: 100%;
  max-width: 320px;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.meta-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.375rem 0;
}

.meta-label {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
}

.meta-value {
  font-size: var(--text-sm);
  color: var(--color-text);
  font-weight: 500;
}
</style>

<template>
  <form class="profile-form" @submit.prevent="handleSubmit">
    <div class="form-group">
      <label for="profile-username">用户名</label>
      <input id="profile-username" :value="user?.username" name="profile_username" disabled />
    </div>

    <div class="form-group">
      <label for="profile-email">邮箱</label>
      <input
        id="profile-email"
        v-model="email"
        name="profile_email"
        type="email"
        placeholder="your@email.com"
        autocomplete="email"
      />
    </div>

    <div class="form-group">
      <label for="profile-created-at">注册时间</label>
      <input
        id="profile-created-at"
        :value="formatDate(user?.createdAt)"
        name="profile_created_at"
        disabled
      />
    </div>

    <div class="form-actions">
      <button type="submit" class="btn-primary" :disabled="submitting || email === user?.email">
        {{ submitting ? '保存中...' : '保存' }}
      </button>
    </div>
  </form>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import type { User } from '@/entities/user/model'
import { formatDateFull as formatDate } from '@/shared/utils/date'

const props = defineProps<{
  user: User | null
}>()

const emit = defineEmits<{
  submit: [email: string]
}>()

const email = ref(props.user?.email || '')
const submitting = ref(false)

watch(
  () => props.user,
  (val) => {
    if (val) email.value = val.email
  },
)

async function handleSubmit() {
  submitting.value = true
  try {
    emit('submit', email.value)
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.profile-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
  max-width: 400px;
}

.form-actions {
  padding-top: 8px;
  border-top: none;
}
</style>

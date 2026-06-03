<script setup lang="ts">
import type { Component } from 'vue'

defineProps<{
  icon: Component
  iconClass: string
  label: string
  value: string | number
  badgeText?: string
  badgeClass?: string
}>()
</script>

<template>
  <div class="admin-stat-card">
    <div class="stat-icon" :class="iconClass">
      <component :is="icon" :size="24" />
    </div>
    <div class="stat-body">
      <div class="stat-label">{{ label }}</div>
      <div class="stat-value">{{ value }}</div>
      <div v-if="badgeText" class="stat-extra">
        <span class="badge" :class="badgeClass">{{ badgeText }}</span>
      </div>
    </div>
  </div>
</template>

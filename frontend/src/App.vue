<script setup lang="ts">
import {
  resolveShellKey,
  resolveShellTransitionName,
} from '@/shared/motion/routes'
</script>

<template>
  <div class="route-shell-stage">
    <router-view v-slot="{ Component, route }">
      <Transition
        :name="resolveShellTransitionName(route)"
        mode="out-in"
        appear
      >
        <div
          :key="resolveShellKey(route)"
          class="route-shell-view"
        >
          <component :is="Component" />
        </div>
      </Transition>
    </router-view>
  </div>
</template>

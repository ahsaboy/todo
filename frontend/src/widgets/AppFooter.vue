<script setup lang="ts">
import { computed } from 'vue'
import { getVersionString } from '@/shared/utils/version'

const DEFAULT_REPOSITORY_URL = 'https://github.com/ahsaboy/todo'

const repositoryUrl = computed(() => {
  return import.meta.env.VITE_APP_REPOSITORY_URL?.trim() || DEFAULT_REPOSITORY_URL
})

const issuesUrl = computed(() => `${repositoryUrl.value}/issues`)
const versionText = computed(() => getVersionString())
</script>

<template>
  <footer class="app-footer">
    <div class="app-footer__meta">
      <strong class="app-footer__title">TODO</strong>
      <span
        class="app-footer__separator"
        aria-hidden="true"
      >/</span>
      <span>{{ versionText }}</span>
    </div>
    <div class="app-footer__links">
      <a
        :href="repositoryUrl"
        target="_blank"
        rel="noreferrer"
      >仓库地址</a>
      <a
        :href="issuesUrl"
        target="_blank"
        rel="noreferrer"
      >问题反馈</a>
    </div>
  </footer>
</template>

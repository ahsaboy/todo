<script setup lang="ts">
defineProps<{ fields: [string, string][] }>()
</script>

<template>
  <div v-if="fields.length" class="extra-fields-cell">
    <div v-for="([k, v], i) in fields" :key="i" class="extra-field">
      <span class="extra-key">{{ k }}</span>
      <span class="extra-val" :title="v">{{ v }}</span>
    </div>
  </div>
  <span v-else>&mdash;</span>
</template>

<style scoped>
.extra-fields-cell {
  font-size: 0.75rem;
  font-family: monospace;
  vertical-align: top;
}

.extra-field {
  display: flex;
  gap: 4px;
  line-height: 1.5;
}

.extra-key {
  color: var(--color-text-muted);
  flex-shrink: 0;
}

.extra-val {
  color: var(--color-text);
  word-break: break-all;
}
</style>

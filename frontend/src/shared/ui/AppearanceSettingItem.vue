<script setup lang="ts">
import { Check } from 'lucide-vue-next'
import type {
  AppearanceSetting,
  AppearanceSettingToggle,
  ToggleOption,
} from './appearance.data'

const props = defineProps<{
  setting: AppearanceSetting
}>()

function handleToggleClick(setting: AppearanceSettingToggle, option: ToggleOption, event: MouseEvent) {
  if (option.value === setting.currentValue()) return

  const apply = () => setting.setValue(option.value)

  if (setting.viewTransition) {
    setting.viewTransition(event, apply)
  } else {
    apply()
  }
}
</script>

<template>
  <div class="setting-item">
    <div class="setting-item__header">
      <component :is="props.setting.icon" :size="14" />
      <span class="setting-item__label">{{ props.setting.label }}</span>
    </div>

    <!-- toggle 类型：圆形分段按钮组 -->
    <div v-if="props.setting.type === 'toggle'" class="setting-item__toggle-group">
      <button
        v-for="opt in props.setting.options"
        :key="opt.value"
        type="button"
        class="toggle-option"
        :class="{ 'is-active': opt.value === props.setting.currentValue() }"
        @click="(e: MouseEvent) => handleToggleClick(props.setting as AppearanceSettingToggle, opt, e)"
      >
        <component :is="opt.icon" :size="16" />
        <span class="toggle-option__label">{{ opt.label }}</span>
      </button>
    </div>

    <!-- select 类型：选项列表 -->
    <div v-else-if="props.setting.type === 'select'" class="setting-item__select-list">
      <button
        v-for="opt in props.setting.options"
        :key="opt.value"
        type="button"
        class="select-option"
        :class="{ 'is-active': opt.value === props.setting.currentValue() }"
        @click="props.setting.setValue(opt.value)"
      >
        <span
          class="select-option__label"
          :style="opt.preview ? { fontFamily: opt.preview } : undefined"
        >{{ opt.label }}</span>
        <Check
          v-if="opt.value === props.setting.currentValue()"
          :size="14"
          class="select-option__check"
        />
      </button>
    </div>
  </div>
</template>

<style scoped>
.setting-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.setting-item + .setting-item {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--color-border);
}

.setting-item__header {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--color-text-muted);
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.02em;
  text-transform: uppercase;
}

.setting-item__label {
  line-height: 1;
}

/* --- toggle 类型 --- */

.setting-item__toggle-group {
  display: flex;
  gap: 8px;
}

.toggle-option {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 10px 8px;
  border: 1.5px solid var(--color-border);
  border-radius: 10px;
  background: transparent;
  color: var(--color-text-muted);
  cursor: pointer;
  transition:
    border-color var(--motion-duration-fast) var(--motion-ease-standard),
    background var(--motion-duration-fast) var(--motion-ease-standard),
    color var(--motion-duration-fast) var(--motion-ease-standard),
    box-shadow var(--motion-duration-fast) var(--motion-ease-standard);
}

.toggle-option:hover {
  border-color: color-mix(in srgb, var(--color-primary) 40%, transparent);
  color: var(--color-text);
}

.toggle-option.is-active {
  border-color: var(--color-primary);
  background: var(--color-glow-primary);
  color: var(--color-primary);
  box-shadow: 0 0 0 1px color-mix(in srgb, var(--color-primary) 20%, transparent);
}

.toggle-option__label {
  font-size: 11px;
  font-weight: 500;
  line-height: 1;
}

/* --- select 类型 --- */

.setting-item__select-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.select-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  width: 100%;
  padding: 8px 10px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--color-text);
  font: inherit;
  font-size: 13px;
  text-align: left;
  cursor: pointer;
  transition:
    background var(--motion-duration-fast) var(--motion-ease-standard),
    color var(--motion-duration-fast) var(--motion-ease-standard);
}

.select-option:hover {
  background: var(--color-surface-muted);
}

.select-option.is-active {
  color: var(--color-primary);
  font-weight: 600;
}

.select-option__label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.select-option__check {
  flex-shrink: 0;
  color: var(--color-primary);
}
</style>

import type { Component } from 'vue'
import { Sun, Moon, Palette, Type } from 'lucide-vue-next'
import { useThemeStore } from '@/app/stores/theme.store'
import { useFontStore } from '@/app/stores/font.store'
import type { ThemeMode } from '@/app/stores/theme.store'
import type { FontMode } from '@/app/stores/font.store'
import { revealThemeTransition } from '@/shared/utils/viewTransition'

/* ---------- 类型定义 ---------- */

export interface ToggleOption {
  value: string
  icon: Component
  label: string
}

export interface AppearanceSettingToggle {
  key: string
  type: 'toggle'
  label: string
  icon: Component
  options: ToggleOption[]
  currentValue: () => string
  setValue: (value: string) => void
  viewTransition?: (event: MouseEvent, apply: () => void) => void
}

export interface SelectOption {
  value: string
  label: string
  preview?: string
}

export interface AppearanceSettingSelect {
  key: string
  type: 'select'
  label: string
  icon: Component
  options: SelectOption[]
  currentValue: () => string
  setValue: (value: string) => void
}

export type AppearanceSetting = AppearanceSettingToggle | AppearanceSettingSelect

/* ---------- 字体定义列表（追加即可扩展） ---------- */

export interface FontDefinition {
  value: string
  label: string
  preview: string
}

export const FONT_DEFINITIONS: FontDefinition[] = [
  {
    value: 'sans',
    label: '无衬线体',
    preview: "'Inter', 'Noto Sans SC', sans-serif",
  },
  {
    value: 'mono',
    label: '等宽字体',
    preview: "'JetBrains Mono', 'Fira Code', 'Consolas', 'Noto Sans SC', monospace",
  },
  {
    value: 'serif',
    label: '衬线体',
    preview: "'Source Serif 4', 'Noto Serif SC', '宋体', serif",
  },
]

/* ---------- 配置工厂（需在 setup 内调用） ---------- */

export function createAppearanceSettings(): AppearanceSetting[] {
  const themeStore = useThemeStore()
  const fontStore = useFontStore()

  return [
    {
      key: 'theme',
      type: 'toggle',
      label: '主题',
      icon: Palette,
      options: [
        { value: 'light', icon: Sun, label: '浅色' },
        { value: 'dark', icon: Moon, label: '深色' },
      ],
      currentValue: () => themeStore.mode,
      setValue: (v) => themeStore.setTheme(v as ThemeMode),
      viewTransition: (event, apply) => revealThemeTransition(event, apply),
    },
    {
      key: 'font',
      type: 'select',
      label: '字体',
      icon: Type,
      options: FONT_DEFINITIONS.map((f) => ({
        value: f.value,
        label: f.label,
        preview: f.preview,
      })),
      currentValue: () => fontStore.mode,
      setValue: (v) => fontStore.setFont(v as FontMode),
    },
  ]
}

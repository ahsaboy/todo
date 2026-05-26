// 精选 Lucide 图标词典,与后端 internal/models/tag.go 的 CuratedIcons 保持一致。
// key 为 kebab-case 的图标名,value 为 lucide-vue-next 的组件。
// 分类用于 IconPicker UI 的 tab 切换。

import {
  AlarmClock,
  AlertCircle,
  AlertTriangle,
  Activity,
  Baby,
  Bike,
  Book,
  BookOpen,
  Bookmark,
  Brain,
  Briefcase,
  Bug,
  Building,
  Building2,
  CalendarClock,
  Calendar,
  Camera,
  Car,
  CheckCircle,
  ClipboardList,
  Clock,
  Code,
  Coffee,
  Compass,
  Dumbbell,
  Factory,
  FileText,
  Film,
  Flag,
  Flame,
  Folder,
  Gamepad2,
  Gift,
  GitBranch,
  GraduationCap,
  Hash,
  Heart,
  Home,
  Library,
  Lightbulb,
  Map,
  MapPin,
  Music,
  Notebook,
  Package,
  Palette,
  Pencil,
  Pill,
  Pizza,
  Plane,
  Presentation,
  ShoppingBag,
  ShoppingCart,
  Star,
  Stethoscope,
  Tag as TagIcon,
  Target,
  Terminal,
  User,
  Users,
  Utensils,
  Wrench,
  Zap,
} from 'lucide-vue-next'
import type { Component } from 'vue'

export interface IconCategory {
  key: string
  label: string
  icons: string[]
}

export const CURATED_ICONS: Record<string, Component> = {
  // 工作 / 项目
  'briefcase': Briefcase,
  'building': Building,
  'building2': Building2,
  'factory': Factory,
  'presentation': Presentation,
  'file-text': FileText,
  'folder': Folder,
  'clipboard-list': ClipboardList,
  'calendar': Calendar,
  'calendar-clock': CalendarClock,
  // 学习 / 阅读
  'book': Book,
  'book-open': BookOpen,
  'graduation-cap': GraduationCap,
  'library': Library,
  'pencil': Pencil,
  'notebook': Notebook,
  'lightbulb': Lightbulb,
  'brain': Brain,
  // 生活 / 家庭
  'home': Home,
  'heart': Heart,
  'users': Users,
  'user': User,
  'baby': Baby,
  'shopping-cart': ShoppingCart,
  'shopping-bag': ShoppingBag,
  'utensils': Utensils,
  'coffee': Coffee,
  'pizza': Pizza,
  // 健康 / 运动
  'dumbbell': Dumbbell,
  'bike': Bike,
  'activity': Activity,
  'pill': Pill,
  'stethoscope': Stethoscope,
  // 旅行 / 外出
  'plane': Plane,
  'car': Car,
  'map': Map,
  'map-pin': MapPin,
  'compass': Compass,
  // 状态 / 标记
  'star': Star,
  'flag': Flag,
  'bookmark': Bookmark,
  'alert-triangle': AlertTriangle,
  'alert-circle': AlertCircle,
  'check-circle': CheckCircle,
  'clock': Clock,
  'zap': Zap,
  'flame': Flame,
  'target': Target,
  // 创作 / 技术
  'code': Code,
  'terminal': Terminal,
  'git-branch': GitBranch,
  'bug': Bug,
  'wrench': Wrench,
  'palette': Palette,
  'music': Music,
  'camera': Camera,
  'film': Film,
  'gamepad-2': Gamepad2,
  // 其他
  'gift': Gift,
  'tag': TagIcon,
  'hash': Hash,
  'package': Package,
}

export const ICON_CATEGORIES: IconCategory[] = [
  {
    key: 'work',
    label: '工作',
    icons: ['briefcase', 'building', 'building2', 'factory', 'presentation', 'file-text', 'folder', 'clipboard-list', 'calendar', 'calendar-clock'],
  },
  {
    key: 'study',
    label: '学习',
    icons: ['book', 'book-open', 'graduation-cap', 'library', 'pencil', 'notebook', 'lightbulb', 'brain'],
  },
  {
    key: 'life',
    label: '生活',
    icons: ['home', 'heart', 'users', 'user', 'baby', 'shopping-cart', 'shopping-bag', 'utensils', 'coffee', 'pizza'],
  },
  {
    key: 'health',
    label: '健康',
    icons: ['dumbbell', 'bike', 'activity', 'pill', 'stethoscope'],
  },
  {
    key: 'travel',
    label: '旅行',
    icons: ['plane', 'car', 'map', 'map-pin', 'compass'],
  },
  {
    key: 'status',
    label: '标记',
    icons: ['star', 'flag', 'bookmark', 'alert-triangle', 'alert-circle', 'check-circle', 'clock', 'zap', 'flame', 'target'],
  },
  {
    key: 'creative',
    label: '创作',
    icons: ['code', 'terminal', 'git-branch', 'bug', 'wrench', 'palette', 'music', 'camera', 'film', 'gamepad-2'],
  },
  {
    key: 'other',
    label: '其它',
    icons: ['gift', 'tag', 'hash', 'package'],
  },
]

export function isValidIconKey(key: string): boolean {
  return key === '' || Object.prototype.hasOwnProperty.call(CURATED_ICONS, key)
}

// 默认色板:12 个克制的颜色,搭配深浅两色主题
export const PRESET_COLORS: string[] = [
  '#ef4444', // red-500
  '#f97316', // orange-500
  '#f59e0b', // amber-500
  '#eab308', // yellow-500
  '#84cc16', // lime-500
  '#22c55e', // green-500
  '#14b8a6', // teal-500
  '#06b6d4', // cyan-500
  '#3b82f6', // blue-500
  '#6366f1', // indigo-500
  '#a855f7', // purple-500
  '#ec4899', // pink-500
]

export const DEFAULT_TAG_COLOR = '#3b82f6'

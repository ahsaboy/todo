import type { Component } from 'vue'

export interface DataTableColumn<T = unknown> {
  key: string & keyof T
  label: string
  formatter?: (value: T[keyof T], row: T) => string
  /** Render cell as a Vue component */
  component?: Component
  /** Props for the component. Receives the row. */
  componentProps?: (row: T) => Record<string, unknown>
  width?: string
  cellClass?: string | ((row: T) => string)
  truncate?: boolean
}

export interface DataTableAction<T = unknown> {
  id: string
  label: string | ((row: T) => string)
  variant?: 'default' | 'primary' | 'danger'
  onClick: (row: T) => void
  visible?: (row: T) => boolean
  icon?: Component
}

export interface ToolbarFilter {
  id: string
  type: 'text' | 'select' | 'number'
  placeholder?: string
  options?: Array<{ label: string; value: string }>
  value: string
  width?: 'narrow' | 'normal' | 'wide'
}

export interface DataTableConfig<T = unknown> {
  columns: DataTableColumn<T>[]
  actions?: DataTableAction<T>[]
  actionsWidth?: string
  filters?: ToolbarFilter[]
  filterButtonText?: string
  emptyText?: string
  loadingText?: string
  /** Mobile card layout config. If absent, all columns render as meta rows. */
  mobileCard?: {
    titleKey: keyof T & string
    subtitleKey?: keyof T & string
    /** Column key to render as a badge in the card header (e.g. status) */
    badgeKey?: keyof T & string
    metaKeys?: (keyof T & string)[]
  }
}

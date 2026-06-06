// DTO (后端 snake_case)
export interface ConfigFieldDto {
  key: string
  label: string
  type: string
  enum?: string[]
  editable: boolean
  hot_reload: boolean
  value: unknown
  source: string
}

export interface ConfigGroupDto {
  group: string
  fields: ConfigFieldDto[]
}

// Model (前端 camelCase)
export type ConfigFieldType = 'string' | 'int' | 'float' | 'bool' | 'enum'
export type ConfigSource = 'cli' | 'db' | 'config'

export interface ConfigField {
  key: string
  label: string
  type: ConfigFieldType
  enum: string[]
  editable: boolean
  hotReload: boolean
  value: unknown
  source: ConfigSource
}

export interface ConfigGroup {
  group: string
  fields: ConfigField[]
}

export interface ConfigUpdate {
  key: string
  value: unknown
}

export interface UpdateConfigResult {
  restartRequired: boolean
  updated: string[]
}

import type {
  ConfigField,
  ConfigFieldDto,
  ConfigFieldType,
  ConfigGroup,
  ConfigGroupDto,
  ConfigSource,
} from './model'

export function toConfigField(dto: ConfigFieldDto): ConfigField {
  return {
    key: dto.key,
    label: dto.label,
    type: (dto.type as ConfigFieldType) ?? 'string',
    enum: dto.enum ?? [],
    editable: dto.editable ?? false,
    hotReload: dto.hot_reload ?? false,
    value: dto.value,
    source: (dto.source as ConfigSource) ?? 'config',
  }
}

export function toConfigGroup(dto: ConfigGroupDto): ConfigGroup {
  return {
    group: dto.group,
    fields: (dto.fields ?? []).map(toConfigField),
  }
}

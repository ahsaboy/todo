import { API_BASE_URL } from '@/shared/config/api'
import type { LoggerRuntimeConfig } from './types'

interface RuntimeConfigResponse {
  logging?: {
    frontend?: Partial<LoggerRuntimeConfig>
  }
}

const safeDefaultConfig: LoggerRuntimeConfig = {
  consoleEnabled: false,
  fileEnabled: false,
  level: 'warn',
}

let runtimeConfigPromise: Promise<LoggerRuntimeConfig> | null = null
let runtimeConfig: LoggerRuntimeConfig = safeDefaultConfig

export function getLoggerConfig(): LoggerRuntimeConfig {
  return runtimeConfig
}

export async function loadLoggerConfig(): Promise<LoggerRuntimeConfig> {
  if (!runtimeConfigPromise) {
    runtimeConfigPromise = fetch(`${API_BASE_URL}/runtime-config`)
      .then((response) => response.json() as Promise<RuntimeConfigResponse>)
      .then((response) => {
        const frontend = response.logging?.frontend
        runtimeConfig = {
          consoleEnabled: frontend?.consoleEnabled ?? false,
          fileEnabled: frontend?.fileEnabled ?? false,
          level: normalizeLevel(frontend?.level),
        }
        return runtimeConfig
      })
      .catch(() => {
        runtimeConfig = safeDefaultConfig
        return runtimeConfig
      })
  }

  return runtimeConfigPromise
}

function normalizeLevel(level: unknown): LoggerRuntimeConfig['level'] {
  if (level === 'debug' || level === 'info' || level === 'warn' || level === 'error') {
    return level
  }
  return 'warn'
}

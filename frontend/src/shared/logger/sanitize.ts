import type { LogContext } from './types'

const SENSITIVE_KEYS = new Set(['authorization', 'api_key', 'password', 'token', 'cookie'])

function isPlainObject(value: unknown): value is Record<string, unknown> {
  return Object.prototype.toString.call(value) === '[object Object]'
}

function sanitizeValue(value: unknown): unknown {
  if (Array.isArray(value)) {
    return value.map((item) => sanitizeValue(item))
  }

  if (isPlainObject(value)) {
    const result: Record<string, unknown> = {}
    for (const [key, nestedValue] of Object.entries(value)) {
      if (SENSITIVE_KEYS.has(key.toLowerCase())) {
        result[key] = '[REDACTED]'
        continue
      }
      result[key] = sanitizeValue(nestedValue)
    }
    return result
  }

  return value
}

export function sanitizeContext(context?: LogContext): LogContext | undefined {
  if (!context) {
    return undefined
  }

  return sanitizeValue(context) as LogContext
}


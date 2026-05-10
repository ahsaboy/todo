import { getLoggerConfig } from './config'
import { sanitizeContext } from './sanitize'
import { enqueueLog } from './transport'
import type { LogContext, LogLevel } from './types'

const LEVEL_WEIGHT: Record<LogLevel, number> = {
  debug: 10,
  info: 20,
  warn: 30,
  error: 40,
}

function shouldLog(level: LogLevel): boolean {
  return LEVEL_WEIGHT[level] >= LEVEL_WEIGHT[getLoggerConfig().level]
}

function writeConsole(level: LogLevel, message: string, context?: LogContext): void {
  if (!getLoggerConfig().consoleEnabled) {
    return
  }

  const payload = context ? { message, ...context } : message
  switch (level) {
    case 'debug':
      console.debug(payload)
      break
    case 'info':
      console.info(payload)
      break
    case 'warn':
      console.warn(payload)
      break
    case 'error':
      console.error(payload)
      break
  }
}

function writeLog(level: LogLevel, message: string, context?: LogContext): void {
  if (!shouldLog(level)) {
    return
  }

  const sanitizedContext = sanitizeContext(context)
  writeConsole(level, message, sanitizedContext)

  if (!getLoggerConfig().fileEnabled) {
    return
  }

  enqueueLog({
    level,
    message,
    context: sanitizedContext,
    timestamp: new Date().toISOString(),
  })
}

export const logger = {
  debug(message: string, context?: LogContext): void {
    writeLog('debug', message, context)
  },
  info(message: string, context?: LogContext): void {
    writeLog('info', message, context)
  },
  warn(message: string, context?: LogContext): void {
    writeLog('warn', message, context)
  },
  error(message: string, context?: LogContext): void {
    writeLog('error', message, context)
  },
}


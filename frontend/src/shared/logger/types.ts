export type LogLevel = 'debug' | 'info' | 'warn' | 'error'

export interface LoggerRuntimeConfig {
  consoleEnabled: boolean
  fileEnabled: boolean
  level: LogLevel
}

export interface LogContext {
  [key: string]: unknown
}

export interface LogEntry {
  level: LogLevel
  message: string
  context?: LogContext
  timestamp: string
}

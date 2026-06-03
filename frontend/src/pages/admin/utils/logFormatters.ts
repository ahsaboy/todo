export interface LogEntry {
  ts: string
  level: string
  msg: string
  caller?: string
  [key: string]: unknown
}

export interface LogEntryRow extends LogEntry {
  _extraFields: [string, string][]
}

export interface LogFile {
  filename: string
  date: string
  size_bytes: number
}

export function extraFieldEntries(entry: LogEntry): [string, string][] {
  const reserved = new Set(['ts', 'level', 'msg', 'caller'])
  const pairs: [string, string][] = []
  for (const [k, v] of Object.entries(entry)) {
    if (!reserved.has(k)) pairs.push([k, typeof v === 'string' ? v : JSON.stringify(v)])
  }
  return pairs
}

export function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

export function truncateMsg(msg: unknown): string {
  if (typeof msg !== 'string') return String(msg ?? '')
  return msg.length > 120 ? msg.slice(0, 120) + '…' : msg
}

export function levelBadgeClass(row: LogEntryRow): string {
  switch (row.level?.toLowerCase()) {
    case 'debug': return 'badge badge-level-debug'
    case 'info': return 'badge badge-level-info'
    case 'warn': return 'badge badge-level-warn'
    case 'error': return 'badge badge-level-error'
    default: return 'badge'
  }
}

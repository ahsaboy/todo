/**
 * 日期格式化工具 — 统一处理 ISO 8601 ↔ 本地显示格式的转换
 */

const pad = (n: number) => String(n).padStart(2, '0')

/** ISO 8601 UTC → "YYYY-MM-DD HH:mm"（本地时间，用于 vue-datepicker） */
export function isoToDateTimeLocal(iso: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

/** ISO 8601 UTC → "YYYY-MM-DD"（本地日期，用于 vue-datepicker 日期模式） */
export function isoToDateLocal(iso: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

/** "YYYY-MM-DD HH:mm" 本地时间 → ISO 8601 UTC（提交 API 用） */
export function dateTimeLocalToISOString(value: string): string {
  if (!value) return ''
  const d = new Date(value)
  return Number.isNaN(d.getTime()) ? '' : d.toISOString()
}

/** "YYYY-MM-DD" 本地日期 → ISO 8601 UTC（当天 23:59:59.999，浏览器本地时区） */
export function dateToEndOfDayISOString(value: string): string {
  if (!value) return ''
  const parts = value.split('-')
  if (parts.length !== 3) return ''
  const [y, m, d] = parts.map(Number)
  const dt = new Date(y, m - 1, d, 23, 59, 59, 999)
  return Number.isNaN(dt.getTime()) ? '' : dt.toISOString()
}

// ---- 显示格式化 ----

const WEEKDAYS = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']

/** "M月D日 HH:mm" — 通用日期时间显示 */
export function formatDateTime(dateStr: string): string {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  if (Number.isNaN(d.getTime())) return dateStr
  return `${d.getMonth() + 1}月${d.getDate()}日 ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

/** "M月D日" — 短日期显示 */
export function formatDateShort(dateStr: string): string {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  if (Number.isNaN(d.getTime())) return dateStr
  return `${d.getMonth() + 1}月${d.getDate()}日`
}

/** "YYYY年M月D日" — 完整日期显示 */
export function formatDateFull(dateStr?: string): string {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  if (Number.isNaN(d.getTime())) return dateStr
  return `${d.getFullYear()}年${d.getMonth() + 1}月${d.getDate()}日`
}

/** "YYYY年M月D日 HH:mm" — 完整日期时间显示 */
export function formatDateTimeFull(dateStr: string): string {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  if (Number.isNaN(d.getTime())) return dateStr
  return `${d.getFullYear()}年${d.getMonth() + 1}月${d.getDate()}日 ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

/** "今天/明天/M月D日 . 周X HH:mm" — 相对日期显示（卡片用） */
export function formatDueRelative(dateStr: string): string {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  if (Number.isNaN(date.getTime())) return dateStr
  const now = new Date()
  const isToday = date.toDateString() === now.toDateString()
  const tomorrow = new Date(now)
  tomorrow.setDate(tomorrow.getDate() + 1)
  const isTomorrow = date.toDateString() === tomorrow.toDateString()

  let dayPart: string
  if (isToday) {
    dayPart = '今天'
  } else if (isTomorrow) {
    dayPart = '明天'
  } else {
    dayPart = `${date.getMonth() + 1}月${date.getDate()}日`
  }

  const weekday = WEEKDAYS[date.getDay()]
  const time = `${pad(date.getHours())}:${pad(date.getMinutes())}`
  return `${dayPart} ${time} . ${weekday}`
}

/** "今天/明天/M月D日" — 相对日期显示（列表用） */
export function formatDateRelative(dateStr: string): string {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  if (Number.isNaN(date.getTime())) return dateStr
  const now = new Date()
  const isToday = date.toDateString() === now.toDateString()
  const tomorrow = new Date(now)
  tomorrow.setDate(tomorrow.getDate() + 1)
  const isTomorrow = date.toDateString() === tomorrow.toDateString()

  if (isToday) return '今天'
  if (isTomorrow) return '明天'
  return `${date.getMonth() + 1}月${date.getDate()}日`
}

/** 判断任务是否逾期 */
export function isOverdue(dueAt: string | null, completed: boolean): boolean {
  if (completed || !dueAt) return false
  return new Date(dueAt) < new Date()
}

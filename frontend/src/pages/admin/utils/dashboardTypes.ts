export interface DayCount {
  date: string
  count: number
}

export interface PriorityCount {
  priority: number
  count: number
}

export interface TagCount {
  tag: string
  count: number
}

export interface Stats {
  total_users: number
  total_tasks: number
  completed_tasks: number
  total_reminder_configs: number
  total_reminder_logs: number
  today_new_tasks: number
  today_new_users: number
  active_users_7d: number
  priority_dist: PriorityCount[]
  completion_trend: DayCount[]
  top_tags: TagCount[]
}

export interface Trends {
  tasks_per_day: DayCount[]
  users_per_day: DayCount[]
  reminder_status_dist: Record<string, number>
}

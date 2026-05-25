// DTO (后端 snake_case)
export interface TaskDto {
  completed: boolean
  created_at: string
  description: string
  due_at: string
  focus_duration: number | null
  id: number
  priority: number
  remind_at: string
  reminder_sent: boolean
  reminder_sent_at: string
  repeat_end_date: string
  repeat_interval: number
  repeat_type: string
  title: string
  updated_at: string
  user_id: number
}

// Model (前端 camelCase)
export interface Task {
  completed: boolean
  createdAt: string
  description: string
  dueAt: string
  focusDuration: number | null
  id: number
  priority: number
  remindAt: string
  reminderSent: boolean
  reminderSentAt: string
  repeatEndDate: string
  repeatInterval: number
  repeatType: string
  title: string
  updatedAt: string
  userId: number
}

// Payload (请求体，保持 snake_case)
export interface CreateTaskPayload {
  description?: string
  due_at?: string
  priority?: 1 | 2 | 3
  remind_at?: string
  repeat_end_date?: string
  repeat_interval?: number
  repeat_type?: 'none' | 'daily' | 'weekly' | 'monthly' | 'yearly'
  title: string
}

export interface UpdateTaskPayload {
  description?: string
  due_at?: string
  priority?: 1 | 2 | 3
  remind_at?: string
  repeat_end_date?: string
  repeat_interval?: number
  repeat_type?: 'none' | 'daily' | 'weekly' | 'monthly' | 'yearly'
  title?: string
}

package models

import "time"

type Task struct {
	ID             int64   `json:"id"`
	UserID         int64   `json:"user_id"`
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	Completed      bool    `json:"completed"`
	Priority       int     `json:"priority"`
	DueAt          *string `json:"due_at"`
	RemindAt       *string `json:"remind_at"`
	RepeatType     string  `json:"repeat_type"`
	RepeatInterval int     `json:"repeat_interval"`
	RepeatEndDate  *string `json:"repeat_end_date"`
	ReminderSent   bool    `json:"reminder_sent"`
	ReminderSentAt *string `json:"reminder_sent_at"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// TemplateData 用于 Webhook 模板渲染
type TemplateData struct {
	TaskID        int64
	Title         string
	Description   string
	Priority      int
	PriorityText  string
	DueAt         string
	RemindAt      string
	RepeatType    string
	CreatedAt     string
}

func (t *Task) ToTemplateData() TemplateData {
	priorityMap := map[int]string{1: "高", 2: "中", 3: "低"}
	dueAt := ""
	if t.DueAt != nil {
		dueAt = *t.DueAt
	}
	remindAt := ""
	if t.RemindAt != nil {
		remindAt = *t.RemindAt
	}
	return TemplateData{
		TaskID:       t.ID,
		Title:        t.Title,
		Description:  t.Description,
		Priority:     t.Priority,
		PriorityText: priorityMap[t.Priority],
		DueAt:        dueAt,
		RemindAt:     remindAt,
		RepeatType:   t.RepeatType,
		CreatedAt:    t.CreatedAt,
	}
}

type CreateTaskRequest struct {
	Title          string  `json:"title" binding:"required,min=1,max=255"`
	Description    string  `json:"description" binding:"max=1000"`
	Priority       *int    `json:"priority" binding:"omitempty,oneof=1 2 3"`
	DueAt          *string `json:"due_at"`
	RemindAt       *string `json:"remind_at"`
	RepeatType     *string `json:"repeat_type" binding:"omitempty,oneof=none daily weekly monthly yearly"`
	RepeatInterval *int    `json:"repeat_interval" binding:"omitempty,min=1,max=365"`
	RepeatEndDate  *string `json:"repeat_end_date"`
}

type UpdateTaskRequest struct {
	Title          *string `json:"title" binding:"omitempty,min=1,max=255"`
	Description    *string `json:"description" binding:"omitempty,max=1000"`
	Priority       *int    `json:"priority" binding:"omitempty,oneof=1 2 3"`
	DueAt          *string `json:"due_at"`
	RemindAt       *string `json:"remind_at"`
	RepeatType     *string `json:"repeat_type" binding:"omitempty,oneof=none daily weekly monthly yearly"`
	RepeatInterval *int    `json:"repeat_interval" binding:"omitempty,min=1,max=365"`
	RepeatEndDate  *string `json:"repeat_end_date"`
}

type TaskFilters struct {
	Status    string
	Priority  int
	DueBefore string
	DueAfter  string
	Search    string
}

type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

// CalculateNextDueDate 计算重复任务的下一次日期，返回 UTC RFC3339。
func CalculateNextDueDate(current string, repeatType string, interval int) (string, error) {
	t, err := time.Parse(time.RFC3339, current)
	if err != nil {
		return "", err
	}
	t = t.UTC()

	switch repeatType {
	case "daily":
		t = t.AddDate(0, 0, interval)
	case "weekly":
		t = t.AddDate(0, 0, 7*interval)
	case "monthly":
		t = t.AddDate(0, interval, 0)
	case "yearly":
		t = t.AddDate(interval, 0, 0)
	default:
		return t.Format(time.RFC3339), nil
	}

	return t.Format(time.RFC3339), nil
}

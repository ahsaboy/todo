package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"todo/internal/models"
)

type TaskRepo struct {
	db *sql.DB
}

func NewTaskRepo(db *sql.DB) *TaskRepo {
	return &TaskRepo{db: db}
}

func (r *TaskRepo) Create(ctx context.Context, req models.CreateTaskRequest) (*models.Task, error) {
	priority := 3
	if req.Priority != nil {
		priority = *req.Priority
	}
	repeatType := "none"
	if req.RepeatType != nil {
		repeatType = *req.RepeatType
	}
	repeatInterval := 1
	if req.RepeatInterval != nil {
		repeatInterval = *req.RepeatInterval
	}

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO tasks (title, description, priority, due_at, remind_at, repeat_type, repeat_interval, repeat_end_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		req.Title, req.Description, priority, req.DueAt, req.RemindAt, repeatType, repeatInterval, req.RepeatEndDate,
	)
	if err != nil {
		return nil, fmt.Errorf("insert task: %w", err)
	}

	id, _ := result.LastInsertId()
	return r.GetByID(ctx, id)
}

func (r *TaskRepo) GetByID(ctx context.Context, id int64) (*models.Task, error) {
	task := &models.Task{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, title, description, completed, priority, due_at, remind_at,
		       repeat_type, repeat_interval, repeat_end_date, reminder_sent, reminder_sent_at,
		       created_at, updated_at
		FROM tasks WHERE id = ?`, id).Scan(
		&task.ID, &task.Title, &task.Description, &task.Completed, &task.Priority,
		&task.DueAt, &task.RemindAt, &task.RepeatType, &task.RepeatInterval,
		&task.RepeatEndDate, &task.ReminderSent, &task.ReminderSentAt,
		&task.CreatedAt, &task.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}
	return task, nil
}

func (r *TaskRepo) List(ctx context.Context, filters models.TaskFilters, page, limit int, sortField, sortOrder string) ([]models.Task, int64, error) {
	where, args := buildWhereClause(filters)

	var total int64
	countQuery := "SELECT COUNT(*) FROM tasks" + where
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count tasks: %w", err)
	}

	allowedSortFields := map[string]bool{
		"created_at": true, "updated_at": true, "due_at": true, "priority": true, "id": true,
	}
	if !allowedSortFields[sortField] {
		sortField = "created_at"
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	offset := (page - 1) * limit
	query := fmt.Sprintf("SELECT id, title, description, completed, priority, due_at, remind_at, repeat_type, repeat_interval, repeat_end_date, reminder_sent, reminder_sent_at, created_at, updated_at FROM tasks%s ORDER BY %s %s LIMIT ? OFFSET ?", where, sortField, sortOrder)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Completed, &t.Priority, &t.DueAt, &t.RemindAt, &t.RepeatType, &t.RepeatInterval, &t.RepeatEndDate, &t.ReminderSent, &t.ReminderSentAt, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, total, nil
}

func (r *TaskRepo) Update(ctx context.Context, id int64, req models.UpdateTaskRequest) (*models.Task, error) {
	setClauses := []string{}
	args := []any{}

	if req.Title != nil {
		setClauses = append(setClauses, "title = ?")
		args = append(args, *req.Title)
	}
	if req.Description != nil {
		setClauses = append(setClauses, "description = ?")
		args = append(args, *req.Description)
	}
	if req.Priority != nil {
		setClauses = append(setClauses, "priority = ?")
		args = append(args, *req.Priority)
	}
	if req.DueAt != nil {
		setClauses = append(setClauses, "due_at = ?")
		args = append(args, *req.DueAt)
	}
	if req.RemindAt != nil {
		setClauses = append(setClauses, "remind_at = ?, reminder_sent = 0, reminder_sent_at = NULL")
		args = append(args, *req.RemindAt)
	}
	if req.RepeatType != nil {
		setClauses = append(setClauses, "repeat_type = ?")
		args = append(args, *req.RepeatType)
	}
	if req.RepeatInterval != nil {
		setClauses = append(setClauses, "repeat_interval = ?")
		args = append(args, *req.RepeatInterval)
	}
	if req.RepeatEndDate != nil {
		setClauses = append(setClauses, "repeat_end_date = ?")
		args = append(args, *req.RepeatEndDate)
	}

	if len(setClauses) == 0 {
		return r.GetByID(ctx, id)
	}

	setClauses = append(setClauses, "updated_at = datetime('now','localtime')")
	args = append(args, id)

	query := "UPDATE tasks SET " + strings.Join(setClauses, ", ") + " WHERE id = ?"
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, nil
	}
	return r.GetByID(ctx, id)
}

func (r *TaskRepo) Delete(ctx context.Context, id int64) (bool, error) {
	result, err := r.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return false, fmt.Errorf("delete task: %w", err)
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}

func (r *TaskRepo) ToggleComplete(ctx context.Context, id int64) (*models.Task, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE tasks SET completed = CASE WHEN completed = 0 THEN 1 ELSE 0 END, updated_at = datetime('now','localtime') WHERE id = ?`, id)
	if err != nil {
		return nil, fmt.Errorf("toggle complete: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, nil
	}
	return r.GetByID(ctx, id)
}

func (r *TaskRepo) GetPendingReminders(ctx context.Context) ([]models.Task, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, description, completed, priority, due_at, remind_at,
		       repeat_type, repeat_interval, repeat_end_date, reminder_sent, reminder_sent_at,
		       created_at, updated_at
		FROM tasks
		WHERE remind_at IS NOT NULL AND reminder_sent = 0 AND remind_at <= datetime('now','localtime')`)
	if err != nil {
		return nil, fmt.Errorf("query pending reminders: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Completed, &t.Priority, &t.DueAt, &t.RemindAt, &t.RepeatType, &t.RepeatInterval, &t.RepeatEndDate, &t.ReminderSent, &t.ReminderSentAt, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan reminder task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *TaskRepo) MarkReminderSent(ctx context.Context, id int64) (bool, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE tasks SET reminder_sent = 1, reminder_sent_at = datetime('now','localtime') WHERE id = ? AND reminder_sent = 0`, id)
	if err != nil {
		return false, fmt.Errorf("mark reminder sent: %w", err)
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}

func (r *TaskRepo) CreateRepeatTask(ctx context.Context, t *models.Task) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO tasks (title, description, priority, due_at, remind_at, repeat_type, repeat_interval, repeat_end_date, reminder_sent)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0)`,
		t.Title, t.Description, t.Priority, t.DueAt, t.RemindAt, t.RepeatType, t.RepeatInterval, t.RepeatEndDate,
	)
	return err
}

func buildWhereClause(f models.TaskFilters) (string, []any) {
	clauses := []string{}
	args := []any{}

	if f.Status == "completed" {
		clauses = append(clauses, "completed = 1")
	} else if f.Status == "pending" {
		clauses = append(clauses, "completed = 0")
	}
	if f.Priority > 0 {
		clauses = append(clauses, "priority = ?")
		args = append(args, f.Priority)
	}
	if f.DueBefore != "" {
		clauses = append(clauses, "due_at <= ?")
		args = append(args, f.DueBefore)
	}
	if f.DueAfter != "" {
		clauses = append(clauses, "due_at >= ?")
		args = append(args, f.DueAfter)
	}
	if f.Search != "" {
		clauses = append(clauses, "(INSTR(title, ?) > 0 OR INSTR(description, ?) > 0)")
		args = append(args, f.Search, f.Search)
	}

	if len(clauses) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

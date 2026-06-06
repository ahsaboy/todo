package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"todo/internal/models"
	"todo/internal/utils"
)

type TaskRepo struct {
	db          *sql.DB
	gracePeriod time.Duration
}

const taskRepositoryName = "task_repo"

func NewTaskRepo(db *sql.DB, gracePeriod time.Duration) *TaskRepo {
	if gracePeriod <= 0 {
		gracePeriod = 10 * time.Minute
	}
	return &TaskRepo{db: db, gracePeriod: gracePeriod}
}

func (r *TaskRepo) SetGracePeriod(d time.Duration) {
	if d <= 0 {
		d = 10 * time.Minute
	}
	r.gracePeriod = d
}

// marshalTags 把 []string 序列化为 JSON 字符串,nil/空数组都写 "[]"。
func marshalTags(tags []string) string {
	if len(tags) == 0 {
		return "[]"
	}
	b, err := json.Marshal(tags)
	if err != nil {
		return "[]"
	}
	return string(b)
}

// unmarshalTags 把数据库的 tags 列反序列化为 []string,容忍空串/旧数据。
func unmarshalTags(raw string) []string {
	if raw == "" {
		return []string{}
	}
	var tags []string
	if err := json.Unmarshal([]byte(raw), &tags); err != nil {
		return []string{}
	}
	if tags == nil {
		return []string{}
	}
	return tags
}

func (r *TaskRepo) Create(ctx context.Context, userID int64, req models.CreateTaskRequest) (*models.Task, error) {
	log := beginDBOperation(ctx, taskRepositoryName, "create_task",
		zap.Int64("user_id", userID),
		zap.String("title", req.Title),
	)
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

	now := time.Now().UTC().Format(time.RFC3339)
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO tasks (user_id, title, description, priority, due_at, remind_at, repeat_type, repeat_interval, repeat_end_date, tags, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, req.Title, req.Description, priority, req.DueAt, req.RemindAt, repeatType, repeatInterval, req.RepeatEndDate, marshalTags(req.Tags), now, now,
	)
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("insert task: %w", err)
	}

	id := lastInsertID(result)
	log.complete(nil,
		zap.Int64("task_id", id),
		zap.Int64("rows_affected", rowsAffected(result)),
	)
	return r.GetByID(ctx, userID, id)
}

func (r *TaskRepo) GetByID(ctx context.Context, userID, id int64) (*models.Task, error) {
	log := beginDBOperation(ctx, taskRepositoryName, "get_task_by_id",
		zap.Int64("user_id", userID),
		zap.Int64("task_id", id),
	)
	task := &models.Task{}
	var tagsRaw string
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, title, description, completed, priority, due_at, remind_at,
		       repeat_type, repeat_interval, repeat_end_date, reminder_sent, reminder_sent_at,
		       focus_duration, tags, created_at, updated_at
		FROM tasks WHERE id = ? AND user_id = ?`, id, userID).Scan(
		&task.ID, &task.UserID, &task.Title, &task.Description, &task.Completed, &task.Priority,
		&task.DueAt, &task.RemindAt, &task.RepeatType, &task.RepeatInterval,
		&task.RepeatEndDate, &task.ReminderSent, &task.ReminderSentAt,
		&task.FocusDuration, &tagsRaw, &task.CreatedAt, &task.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		log.complete(nil, zap.Bool("found", false))
		return nil, nil
	}
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("get task: %w", err)
	}
	task.Tags = unmarshalTags(tagsRaw)
	log.complete(nil,
		zap.Bool("found", true),
		zap.Bool("completed", task.Completed),
	)
	return task, nil
}

func (r *TaskRepo) List(ctx context.Context, userID int64, filters models.TaskFilters, page, limit int, sortField, sortOrder string) ([]models.Task, int64, error) {
	log := beginDBOperation(ctx, taskRepositoryName, "list_tasks",
		zap.Int64("user_id", userID),
		zap.Int("page", page),
		zap.Int("limit", limit),
		zap.String("sort_field", sortField),
		zap.String("sort_order", sortOrder),
		zap.String("status_filter", filters.Status),
		zap.Int("priority_filter", filters.Priority),
		zap.Bool("has_due_before", filters.DueBefore != ""),
		zap.Bool("has_due_after", filters.DueAfter != ""),
		zap.Bool("has_search", filters.Search != ""),
		zap.Int("tags_filter_count", len(filters.Tags)),
		zap.Int("tags_all_filter_count", len(filters.TagsAll)),
	)
	where, args := buildWhereClause(filters)

	// 添加 user_id 过滤
	userWhere := " WHERE user_id = ?"
	if where == "" {
		where = userWhere
		args = append(args, userID)
	} else {
		where = " WHERE user_id = ? AND " + where[7:] // 去掉原有的 " WHERE "
		args = append([]any{userID}, args...)
	}

	var total int64
	countQuery := "SELECT COUNT(*) FROM tasks" + where
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		log.complete(err)
		return nil, 0, fmt.Errorf("count tasks: %w", err)
	}

	orderBy := buildTaskListOrderBy(sortField, sortOrder)

	offset := (page - 1) * limit
	query := fmt.Sprintf("SELECT id, user_id, title, description, completed, priority, due_at, remind_at, repeat_type, repeat_interval, repeat_end_date, reminder_sent, reminder_sent_at, focus_duration, tags, created_at, updated_at FROM tasks%s ORDER BY %s LIMIT ? OFFSET ?", where, orderBy)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		log.complete(err, zap.Int64("count", total))
		return nil, 0, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		var tagsRaw string
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Completed, &t.Priority, &t.DueAt, &t.RemindAt, &t.RepeatType, &t.RepeatInterval, &t.RepeatEndDate, &t.ReminderSent, &t.ReminderSentAt, &t.FocusDuration, &tagsRaw, &t.CreatedAt, &t.UpdatedAt); err != nil {
			log.complete(err, zap.Int64("count", total))
			return nil, 0, fmt.Errorf("scan task: %w", err)
		}
		t.Tags = unmarshalTags(tagsRaw)
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		log.complete(err, zap.Int64("count", total))
		return nil, 0, fmt.Errorf("rows iteration: %w", err)
	}
	log.complete(nil,
		zap.Int64("count", total),
		zap.Int("result_size", len(tasks)),
	)
	return tasks, total, nil
}

func buildTaskListOrderBy(sortField, sortOrder string) string {
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	switch sortField {
	case "task_center":
		return "completed ASC, " +
			"CASE WHEN due_at IS NULL OR due_at = '' THEN 1 ELSE 0 END ASC, " +
			"CASE WHEN completed = 0 THEN due_at END ASC, " +
			"CASE WHEN completed = 1 THEN due_at END DESC, " +
			"id DESC"
	case "created_at", "updated_at", "due_at", "priority", "id":
		return fmt.Sprintf("%s %s", sortField, sortOrder)
	default:
		return "created_at desc"
	}
}

func (r *TaskRepo) Update(ctx context.Context, userID, id int64, req models.UpdateTaskRequest) (*models.Task, error) {
	log := beginDBOperation(ctx, taskRepositoryName, "update_task",
		zap.Int64("user_id", userID),
		zap.Int64("task_id", id),
		zap.Bool("update_title", req.Title != nil),
		zap.Bool("update_description", req.Description != nil),
		zap.Bool("update_priority", req.Priority != nil),
		zap.Bool("update_due_at", req.DueAt != nil),
		zap.Bool("update_remind_at", req.RemindAt != nil),
		zap.Bool("update_repeat_type", req.RepeatType != nil),
		zap.Bool("update_repeat_interval", req.RepeatInterval != nil),
		zap.Bool("update_repeat_end_date", req.RepeatEndDate != nil),
		zap.Bool("update_tags", req.Tags != nil),
	)
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
		if *req.DueAt == "" {
			setClauses = append(setClauses, "due_at = NULL")
		} else {
			setClauses = append(setClauses, "due_at = ?")
			args = append(args, *req.DueAt)
		}
	}
	if req.RemindAt != nil {
		if *req.RemindAt == "" {
			setClauses = append(setClauses, "remind_at = NULL, reminder_sent = 0, reminder_sent_at = NULL")
		} else {
			setClauses = append(setClauses, "remind_at = ?, reminder_sent = 0, reminder_sent_at = NULL")
			args = append(args, *req.RemindAt)
		}
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
	if req.Tags != nil {
		setClauses = append(setClauses, "tags = ?")
		args = append(args, marshalTags(*req.Tags))
	}

	if len(setClauses) == 0 {
		log.complete(nil, zap.String("result", "no_fields_to_update"))
		return r.GetByID(ctx, userID, id)
	}

	setClauses = append(setClauses, "updated_at = ?")
	args = append(args, time.Now().UTC().Format(time.RFC3339), id, userID)

	query := "UPDATE tasks SET " + strings.Join(setClauses, ", ") + " WHERE id = ? AND user_id = ?"

	if req.RemindAt != nil {
		// 更新 remind_at 时需同步删除 reminder_logs，用事务保证原子性
		tx, err := r.db.BeginTx(ctx, nil)
		if err != nil {
			log.complete(err)
			return nil, fmt.Errorf("begin update task tx: %w", err)
		}
		defer tx.Rollback()

		result, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			log.complete(err)
			return nil, fmt.Errorf("update task: %w", err)
		}
		rows := rowsAffected(result)
		if rows == 0 {
			log.complete(nil, zap.Int64("rows_affected", rows), zap.Bool("found", false))
			return nil, nil
		}

		deleteResult, err := tx.ExecContext(ctx, `DELETE FROM reminder_logs WHERE task_id = ?`, id)
		if err != nil {
			log.complete(err, zap.Int64("rows_affected", rows))
			return nil, fmt.Errorf("clear reminder logs: %w", err)
		}

		if err := tx.Commit(); err != nil {
			log.complete(err)
			return nil, fmt.Errorf("commit update task: %w", err)
		}

		log.complete(nil,
			zap.Int64("rows_affected", rows),
			zap.Int64("cleared_reminder_logs", rowsAffected(deleteResult)),
		)
		return r.GetByID(ctx, userID, id)
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("update task: %w", err)
	}
	rows := rowsAffected(result)
	if rows == 0 {
		log.complete(nil, zap.Int64("rows_affected", rows), zap.Bool("found", false))
		return nil, nil
	}
	log.complete(nil, zap.Int64("rows_affected", rows))
	return r.GetByID(ctx, userID, id)
}

func (r *TaskRepo) Delete(ctx context.Context, userID, id int64) (bool, error) {
	log := beginDBOperation(ctx, taskRepositoryName, "delete_task",
		zap.Int64("user_id", userID),
		zap.Int64("task_id", id),
	)
	result, err := r.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		log.complete(err)
		return false, fmt.Errorf("delete task: %w", err)
	}
	rows := rowsAffected(result)
	log.complete(nil,
		zap.Int64("rows_affected", rows),
		zap.Bool("deleted", rows > 0),
	)
	return rows > 0, nil
}

func (r *TaskRepo) ToggleComplete(ctx context.Context, userID, id int64, focusDuration *int) (*models.Task, error) {
	log := beginDBOperation(ctx, taskRepositoryName, "toggle_task_complete",
		zap.Int64("user_id", userID),
		zap.Int64("task_id", id),
	)
	result, err := r.db.ExecContext(ctx, `
		UPDATE tasks SET completed = CASE WHEN completed = 0 THEN 1 ELSE 0 END, focus_duration = ?, updated_at = ? WHERE id = ? AND user_id = ?`,
		focusDuration, time.Now().UTC().Format(time.RFC3339), id, userID)
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("toggle complete: %w", err)
	}
	rows := rowsAffected(result)
	if rows == 0 {
		log.complete(nil, zap.Int64("rows_affected", rows), zap.Bool("found", false))
		return nil, nil
	}
	log.complete(nil, zap.Int64("rows_affected", rows))
	return r.GetByID(ctx, userID, id)
}

func (r *TaskRepo) GetPendingReminders(ctx context.Context, now time.Time) ([]models.Task, error) {
	log := beginDBOperation(ctx, taskRepositoryName, "get_pending_reminders",
		zap.Time("due_before", now.UTC()),
	)
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, title, description, completed, priority, due_at, remind_at,
		       repeat_type, repeat_interval, repeat_end_date, reminder_sent, reminder_sent_at,
		       focus_duration, tags, created_at, updated_at
		FROM tasks
		WHERE user_id IS NOT NULL
		  AND remind_at IS NOT NULL
		  AND reminder_sent = 0
		  AND completed = 0`)
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("query pending reminders: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	nowUTC := now.UTC()
	for rows.Next() {
		var t models.Task
		var tagsRaw string
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Completed, &t.Priority, &t.DueAt, &t.RemindAt, &t.RepeatType, &t.RepeatInterval, &t.RepeatEndDate, &t.ReminderSent, &t.ReminderSentAt, &t.FocusDuration, &tagsRaw, &t.CreatedAt, &t.UpdatedAt); err != nil {
			log.complete(err)
			return nil, fmt.Errorf("scan reminder task: %w", err)
		}
		t.Tags = unmarshalTags(tagsRaw)
		// 运行时兼容：解析 remind_at（支持 RFC3339 和旧格式），在 Go 侧比较
		if t.RemindAt != nil {
			remindTime, err := utils.ParseDBTime(*t.RemindAt)
			if err != nil {
				continue
			}
			if remindTime.After(nowUTC) {
				continue
			}
			// 提醒时间已过超过宽限期，视为过期，跳过
			if nowUTC.Sub(remindTime) > r.gracePeriod {
				continue
			}
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		log.complete(err)
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	log.complete(nil, zap.Int("result_size", len(tasks)))
	return tasks, nil
}

func (r *TaskRepo) MarkReminderSent(ctx context.Context, id int64) (bool, error) {
	log := beginDBOperation(ctx, taskRepositoryName, "mark_reminder_sent",
		zap.Int64("task_id", id),
	)
	result, err := r.db.ExecContext(ctx, `
		UPDATE tasks SET reminder_sent = 1, reminder_sent_at = ? WHERE id = ? AND reminder_sent = 0`,
		time.Now().UTC().Format(time.RFC3339), id)
	if err != nil {
		log.complete(err)
		return false, fmt.Errorf("mark reminder sent: %w", err)
	}
	rows := rowsAffected(result)
	log.complete(nil,
		zap.Int64("rows_affected", rows),
		zap.Bool("updated", rows > 0),
	)
	return rows > 0, nil
}

func (r *TaskRepo) CreateRepeatTask(ctx context.Context, t *models.Task) error {
	log := beginDBOperation(ctx, taskRepositoryName, "create_repeat_task",
		zap.Int64("user_id", t.UserID),
		zap.Int64("source_task_id", t.ID),
	)
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO tasks (user_id, title, description, priority, due_at, remind_at, repeat_type, repeat_interval, repeat_end_date, tags, reminder_sent, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?)`,
		t.UserID, t.Title, t.Description, t.Priority, t.DueAt, t.RemindAt, t.RepeatType, t.RepeatInterval, t.RepeatEndDate, marshalTags(t.Tags), now, now,
	)
	if err != nil {
		log.complete(err)
		return err
	}
	log.complete(nil,
		zap.Int64("task_id", lastInsertID(result)),
		zap.Int64("rows_affected", rowsAffected(result)),
	)
	return nil
}

// ToggleCompleteAndCreateRepeat 在单个事务中切换完成状态并可选地创建下一次重复任务。
// next 为 nil 时仅切换完成状态。
func (r *TaskRepo) ToggleCompleteAndCreateRepeat(ctx context.Context, userID, id int64, next *models.Task, focusDuration *int) (*models.Task, error) {
	log := beginDBOperation(ctx, taskRepositoryName, "toggle_complete_and_create_repeat",
		zap.Int64("user_id", userID),
		zap.Int64("task_id", id),
		zap.Bool("has_next", next != nil),
	)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		UPDATE tasks SET completed = CASE WHEN completed = 0 THEN 1 ELSE 0 END, focus_duration = ?, updated_at = ? WHERE id = ? AND user_id = ?`,
		focusDuration, time.Now().UTC().Format(time.RFC3339), id, userID)
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("toggle complete: %w", err)
	}
	rows := rowsAffected(result)
	if rows == 0 {
		log.complete(nil, zap.Int64("rows_affected", rows), zap.Bool("found", false))
		return nil, nil
	}

	if next != nil {
		now := time.Now().UTC().Format(time.RFC3339)
		_, err = tx.ExecContext(ctx, `
			INSERT INTO tasks (user_id, title, description, priority, due_at, remind_at, repeat_type, repeat_interval, repeat_end_date, tags, reminder_sent, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?)`,
			next.UserID, next.Title, next.Description, next.Priority, next.DueAt, next.RemindAt,
			next.RepeatType, next.RepeatInterval, next.RepeatEndDate, marshalTags(next.Tags), now, now,
		)
		if err != nil {
			log.complete(err)
			return nil, fmt.Errorf("create repeat task: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.complete(err)
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	log.complete(nil, zap.Int64("rows_affected", rows))
	return r.GetByID(ctx, userID, id)
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
	if len(f.Tags) > 0 {
		// OR 语义:任一命中即匹配。使用 SQLite JSON1 的 json_each 展开 tasks.tags。
		placeholders := strings.Repeat("?,", len(f.Tags))
		placeholders = placeholders[:len(placeholders)-1]
		clauses = append(clauses, fmt.Sprintf("EXISTS (SELECT 1 FROM json_each(tasks.tags) WHERE json_each.value IN (%s))", placeholders))
		for _, name := range f.Tags {
			args = append(args, name)
		}
	}
	if len(f.TagsAll) > 0 {
		// AND 语义:每个 tag 都必须存在
		for _, name := range f.TagsAll {
			clauses = append(clauses, "EXISTS (SELECT 1 FROM json_each(tasks.tags) WHERE json_each.value = ?)")
			args = append(args, name)
		}
	}

	if len(clauses) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func (r *TaskRepo) ListAll(ctx context.Context, userID int64, filters models.TaskFilters, page, limit int) ([]models.Task, int64, error) {
	log := beginDBOperation(ctx, taskRepositoryName, "admin_list_all_tasks",
		zap.Int64("user_id", userID),
		zap.Int("page", page),
		zap.Int("limit", limit),
	)
	where, args := buildWhereClause(filters)
	if userID > 0 {
		if where == "" {
			where = " WHERE user_id = ?"
			args = append(args, userID)
		} else {
			where = " WHERE user_id = ? AND " + where[7:]
			args = append([]any{userID}, args...)
		}
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks"+where, args...).Scan(&total); err != nil {
		log.complete(err)
		return nil, 0, fmt.Errorf("count tasks: %w", err)
	}

	offset := (page - 1) * limit
	query := fmt.Sprintf(`SELECT id, user_id, title, description, completed, priority, due_at, remind_at,
		repeat_type, repeat_interval, repeat_end_date, reminder_sent, reminder_sent_at, focus_duration, tags, created_at, updated_at
		FROM tasks%s ORDER BY created_at DESC LIMIT ? OFFSET ?`, where)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		log.complete(err, zap.Int64("count", total))
		return nil, 0, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]models.Task, 0)
	for rows.Next() {
		var t models.Task
		var tagsRaw string
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Completed, &t.Priority,
			&t.DueAt, &t.RemindAt, &t.RepeatType, &t.RepeatInterval, &t.RepeatEndDate,
			&t.ReminderSent, &t.ReminderSentAt, &t.FocusDuration, &tagsRaw, &t.CreatedAt, &t.UpdatedAt); err != nil {
			log.complete(err, zap.Int64("count", total))
			return nil, 0, fmt.Errorf("scan task: %w", err)
		}
		t.Tags = unmarshalTags(tagsRaw)
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		log.complete(err, zap.Int64("count", total))
		return nil, 0, fmt.Errorf("rows iteration: %w", err)
	}
	log.complete(nil, zap.Int64("count", total), zap.Int("result_size", len(tasks)))
	return tasks, total, nil
}

func (r *TaskRepo) AdminGetByID(ctx context.Context, id int64) (*models.Task, error) {
	log := beginDBOperation(ctx, taskRepositoryName, "admin_get_task_by_id", zap.Int64("task_id", id))
	task := &models.Task{}
	var tagsRaw string
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, title, description, completed, priority, due_at, remind_at,
		       repeat_type, repeat_interval, repeat_end_date, reminder_sent, reminder_sent_at,
		       focus_duration, tags, created_at, updated_at
		FROM tasks WHERE id = ?`, id).Scan(
		&task.ID, &task.UserID, &task.Title, &task.Description, &task.Completed, &task.Priority,
		&task.DueAt, &task.RemindAt, &task.RepeatType, &task.RepeatInterval,
		&task.RepeatEndDate, &task.ReminderSent, &task.ReminderSentAt,
		&task.FocusDuration, &tagsRaw, &task.CreatedAt, &task.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		log.complete(nil, zap.Bool("found", false))
		return nil, nil
	}
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("admin get task: %w", err)
	}
	task.Tags = unmarshalTags(tagsRaw)
	log.complete(nil, zap.Bool("found", true))
	return task, nil
}

func (r *TaskRepo) AdminToggleComplete(ctx context.Context, id int64) (*models.Task, error) {
	log := beginDBOperation(ctx, taskRepositoryName, "admin_toggle_task_complete", zap.Int64("task_id", id))
	result, err := r.db.ExecContext(ctx, `
		UPDATE tasks SET completed = CASE WHEN completed = 0 THEN 1 ELSE 0 END, updated_at = ? WHERE id = ?`,
		time.Now().UTC().Format(time.RFC3339), id)
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("admin toggle complete: %w", err)
	}
	rows := rowsAffected(result)
	if rows == 0 {
		log.complete(nil, zap.Int64("rows_affected", rows), zap.Bool("found", false))
		return nil, nil
	}
	log.complete(nil, zap.Int64("rows_affected", rows))
	return r.AdminGetByID(ctx, id)
}

func (r *TaskRepo) AdminUpdate(ctx context.Context, id int64, req models.UpdateTaskRequest) (*models.Task, error) {
	log := beginDBOperation(ctx, taskRepositoryName, "admin_update_task",
		zap.Int64("task_id", id),
		zap.Bool("update_title", req.Title != nil),
		zap.Bool("update_description", req.Description != nil),
		zap.Bool("update_priority", req.Priority != nil),
		zap.Bool("update_due_at", req.DueAt != nil),
	)
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
		if *req.DueAt == "" {
			setClauses = append(setClauses, "due_at = NULL")
		} else {
			setClauses = append(setClauses, "due_at = ?")
			args = append(args, *req.DueAt)
		}
	}
	if req.RemindAt != nil {
		if *req.RemindAt == "" {
			setClauses = append(setClauses, "remind_at = NULL, reminder_sent = 0, reminder_sent_at = NULL")
		} else {
			setClauses = append(setClauses, "remind_at = ?, reminder_sent = 0, reminder_sent_at = NULL")
			args = append(args, *req.RemindAt)
		}
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
	if req.Tags != nil {
		setClauses = append(setClauses, "tags = ?")
		args = append(args, marshalTags(*req.Tags))
	}

	if len(setClauses) == 0 {
		log.complete(nil, zap.String("result", "no_fields_to_update"))
		return r.AdminGetByID(ctx, id)
	}

	setClauses = append(setClauses, "updated_at = ?")
	args = append(args, time.Now().UTC().Format(time.RFC3339), id)

	query := "UPDATE tasks SET " + strings.Join(setClauses, ", ") + " WHERE id = ?"
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("admin update task: %w", err)
	}
	rows := rowsAffected(result)
	if rows == 0 {
		log.complete(nil, zap.Int64("rows_affected", rows), zap.Bool("found", false))
		return nil, nil
	}
	log.complete(nil, zap.Int64("rows_affected", rows))
	return r.AdminGetByID(ctx, id)
}

func (r *TaskRepo) AdminDelete(ctx context.Context, id int64) (bool, error) {
	log := beginDBOperation(ctx, taskRepositoryName, "admin_delete_task", zap.Int64("task_id", id))
	result, err := r.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		log.complete(err)
		return false, fmt.Errorf("admin delete task: %w", err)
	}
	rows := rowsAffected(result)
	log.complete(nil, zap.Int64("rows_affected", rows), zap.Bool("deleted", rows > 0))
	return rows > 0, nil
}

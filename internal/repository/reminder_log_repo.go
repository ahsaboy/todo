package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"

	"todo/internal/models"
)

type ReminderLogRepo struct {
	db *sql.DB
}

const reminderLogRepositoryName = "reminder_log_repo"

type CreateReminderLogParams struct {
	UserID           int64
	TaskID           int64
	ReminderConfigID int64
	ChannelName      string
	ChannelType      string
	Status           string
	Attempts         int
	ErrorMessage     string
}

func NewReminderLogRepo(db *sql.DB) *ReminderLogRepo {
	return &ReminderLogRepo{db: db}
}

func (r *ReminderLogRepo) Upsert(ctx context.Context, p CreateReminderLogParams) error {
	log := beginDBOperation(ctx, reminderLogRepositoryName, "upsert_reminder_log",
		zap.Int64("user_id", p.UserID),
		zap.Int64("task_id", p.TaskID),
		zap.Int64("reminder_config_id", p.ReminderConfigID),
		zap.String("channel_type", p.ChannelType),
		zap.String("status", p.Status),
		zap.Int("attempts", p.Attempts),
		zap.Bool("has_error_message", p.ErrorMessage != ""),
	)
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO reminder_logs (
			user_id, task_id, reminder_config_id, channel_name, channel_type,
			status, attempts, error_message, created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(task_id, reminder_config_id) WHERE reminder_config_id IS NOT NULL
		DO UPDATE SET
			channel_name = excluded.channel_name,
			channel_type = excluded.channel_type,
			status = excluded.status,
			attempts = excluded.attempts,
			error_message = excluded.error_message,
			created_at = excluded.created_at`,
		p.UserID,
		p.TaskID,
		p.ReminderConfigID,
		p.ChannelName,
		p.ChannelType,
		p.Status,
		p.Attempts,
		p.ErrorMessage,
		time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		log.complete(err)
		return fmt.Errorf("upsert reminder log: %w", err)
	}
	log.complete(nil, zap.Int64("rows_affected", rowsAffected(result)))
	return nil
}

func (r *ReminderLogRepo) HasResultForTaskConfig(ctx context.Context, taskID, configID int64) (bool, error) {
	log := beginDBOperation(ctx, reminderLogRepositoryName, "has_result_for_task_config",
		zap.Int64("task_id", taskID),
		zap.Int64("reminder_config_id", configID),
	)
	var count int
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM reminder_logs WHERE task_id = ? AND reminder_config_id = ?`,
		taskID, configID,
	).Scan(&count); err != nil {
		log.complete(err)
		return false, fmt.Errorf("count reminder log: %w", err)
	}
	log.complete(nil,
		zap.Int("count", count),
		zap.Bool("exists", count > 0),
	)
	return count > 0, nil
}

func (r *ReminderLogRepo) ListByUserID(ctx context.Context, userID int64, page, limit int) ([]models.ReminderLog, int64, error) {
	log := beginDBOperation(ctx, reminderLogRepositoryName, "list_reminder_logs",
		zap.Int64("user_id", userID),
		zap.Int("page", page),
		zap.Int("limit", limit),
	)
	var total int64
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM reminder_logs WHERE user_id = ?`, userID,
	).Scan(&total); err != nil {
		log.complete(err)
		return nil, 0, fmt.Errorf("count reminder logs: %w", err)
	}

	offset := (page - 1) * limit
	rows, err := r.db.QueryContext(ctx, `
		SELECT l.id, l.user_id, l.task_id, COALESCE(t.title, ''),
		       l.reminder_config_id, l.channel_name, l.channel_type, l.status,
		       l.attempts, l.error_message, l.created_at
		FROM reminder_logs l
		LEFT JOIN tasks t ON t.id = l.task_id
		WHERE l.user_id = ?
		ORDER BY l.created_at DESC, l.id DESC
		LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		log.complete(err, zap.Int64("count", total))
		return nil, 0, fmt.Errorf("list reminder logs: %w", err)
	}
	defer rows.Close()

	logs := make([]models.ReminderLog, 0)
	for rows.Next() {
		var item models.ReminderLog
		var configID sql.NullInt64
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.TaskID,
			&item.TaskTitle,
			&configID,
			&item.ChannelName,
			&item.ChannelType,
			&item.Status,
			&item.Attempts,
			&item.ErrorMessage,
			&item.CreatedAt,
		); err != nil {
			log.complete(err, zap.Int64("count", total))
			return nil, 0, fmt.Errorf("scan reminder log: %w", err)
		}
		if configID.Valid {
			item.ReminderConfigID = &configID.Int64
		}
		logs = append(logs, item)
	}
	if err := rows.Err(); err != nil {
		log.complete(err, zap.Int64("count", total))
		return nil, 0, fmt.Errorf("rows iteration: %w", err)
	}
	log.complete(nil,
		zap.Int64("count", total),
		zap.Int("result_size", len(logs)),
	)
	return logs, total, nil
}

func (r *ReminderLogRepo) DeleteByTaskID(ctx context.Context, taskID int64) error {
	log := beginDBOperation(ctx, reminderLogRepositoryName, "delete_reminder_logs_by_task_id",
		zap.Int64("task_id", taskID),
	)
	result, err := r.db.ExecContext(ctx, `DELETE FROM reminder_logs WHERE task_id = ?`, taskID)
	if err != nil {
		log.complete(err)
		return fmt.Errorf("delete reminder logs: %w", err)
	}
	log.complete(nil, zap.Int64("rows_affected", rowsAffected(result)))
	return nil
}

func (r *ReminderLogRepo) ListAll(ctx context.Context, page, limit int) ([]models.ReminderLog, int64, error) {
	log := beginDBOperation(ctx, reminderLogRepositoryName, "admin_list_all_reminder_logs",
		zap.Int("page", page),
		zap.Int("limit", limit),
	)
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM reminder_logs`).Scan(&total); err != nil {
		log.complete(err)
		return nil, 0, fmt.Errorf("count reminder logs: %w", err)
	}

	offset := (page - 1) * limit
	rows, err := r.db.QueryContext(ctx, `
		SELECT l.id, l.user_id, l.task_id, COALESCE(t.title, ''),
		       l.reminder_config_id, l.channel_name, l.channel_type, l.status,
		       l.attempts, l.error_message, l.created_at
		FROM reminder_logs l
		LEFT JOIN tasks t ON t.id = l.task_id
		ORDER BY l.created_at DESC, l.id DESC
		LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		log.complete(err, zap.Int64("count", total))
		return nil, 0, fmt.Errorf("list all reminder logs: %w", err)
	}
	defer rows.Close()

	logs := make([]models.ReminderLog, 0)
	for rows.Next() {
		var item models.ReminderLog
		var configID sql.NullInt64
		if err := rows.Scan(
			&item.ID, &item.UserID, &item.TaskID, &item.TaskTitle,
			&configID, &item.ChannelName, &item.ChannelType,
			&item.Status, &item.Attempts, &item.ErrorMessage, &item.CreatedAt,
		); err != nil {
			log.complete(err, zap.Int64("count", total))
			return nil, 0, fmt.Errorf("scan reminder log: %w", err)
		}
		if configID.Valid {
			item.ReminderConfigID = &configID.Int64
		}
		logs = append(logs, item)
	}
	if err := rows.Err(); err != nil {
		log.complete(err, zap.Int64("count", total))
		return nil, 0, fmt.Errorf("rows iteration: %w", err)
	}
	log.complete(nil, zap.Int64("count", total), zap.Int("result_size", len(logs)))
	return logs, total, nil
}

func (r *ReminderLogRepo) AdminListFiltered(ctx context.Context, page, limit int, userID int64, status string) ([]models.ReminderLog, int64, error) {
	log := beginDBOperation(ctx, reminderLogRepositoryName, "admin_list_filtered_reminder_logs",
		zap.Int("page", page),
		zap.Int("limit", limit),
		zap.Int64("user_id", userID),
		zap.String("status", status),
	)
	where := ""
	args := []any{}
	if userID > 0 {
		where += " WHERE l.user_id = ?"
		args = append(args, userID)
	}
	if status != "" {
		if where == "" {
			where = " WHERE l.status = ?"
		} else {
			where += " AND l.status = ?"
		}
		args = append(args, status)
	}

	var total int64
	countArgs := make([]any, len(args))
	copy(countArgs, args)
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM reminder_logs l`+where, countArgs...).Scan(&total); err != nil {
		log.complete(err)
		return nil, 0, fmt.Errorf("count reminder logs: %w", err)
	}

	offset := (page - 1) * limit
	query := fmt.Sprintf(`
		SELECT l.id, l.user_id, l.task_id, COALESCE(t.title, ''),
		       l.reminder_config_id, l.channel_name, l.channel_type, l.status,
		       l.attempts, l.error_message, l.created_at
		FROM reminder_logs l
		LEFT JOIN tasks t ON t.id = l.task_id
		%s
		ORDER BY l.created_at DESC, l.id DESC
		LIMIT ? OFFSET ?`, where)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		log.complete(err, zap.Int64("count", total))
		return nil, 0, fmt.Errorf("list filtered reminder logs: %w", err)
	}
	defer rows.Close()

	logs := make([]models.ReminderLog, 0)
	for rows.Next() {
		var item models.ReminderLog
		var configID sql.NullInt64
		if err := rows.Scan(
			&item.ID, &item.UserID, &item.TaskID, &item.TaskTitle,
			&configID, &item.ChannelName, &item.ChannelType,
			&item.Status, &item.Attempts, &item.ErrorMessage, &item.CreatedAt,
		); err != nil {
			log.complete(err, zap.Int64("count", total))
			return nil, 0, fmt.Errorf("scan reminder log: %w", err)
		}
		if configID.Valid {
			item.ReminderConfigID = &configID.Int64
		}
		logs = append(logs, item)
	}
	if err := rows.Err(); err != nil {
		log.complete(err, zap.Int64("count", total))
		return nil, 0, fmt.Errorf("rows iteration: %w", err)
	}
	log.complete(nil, zap.Int64("count", total), zap.Int("result_size", len(logs)))
	return logs, total, nil
}

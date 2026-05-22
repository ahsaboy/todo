package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"

	"todo/internal/models"
)

type ReminderConfigRepo struct {
	db *sql.DB
}

const reminderConfigRepositoryName = "reminder_config_repo"

func NewReminderConfigRepo(db *sql.DB) *ReminderConfigRepo {
	return &ReminderConfigRepo{db: db}
}

func (r *ReminderConfigRepo) Create(ctx context.Context, cfg *models.UserReminderConfig) (*models.UserReminderConfig, error) {
	log := beginDBOperation(ctx, reminderConfigRepositoryName, "create_reminder_config",
		zap.Int64("user_id", cfg.UserID),
		zap.String("name", cfg.Name),
		zap.String("channel_type", cfg.ChannelType),
		zap.Bool("enabled", cfg.Enabled),
	)
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO user_reminder_configs
			(user_id, name, channel_type, webhook_url, webhook_method, webhook_headers, webhook_body_template, max_retries, retry_delay_seconds, enabled)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		cfg.UserID, cfg.Name, cfg.ChannelType, cfg.WebhookURL, cfg.WebhookMethod,
		cfg.GetWebhookHeadersJSON(), cfg.WebhookBodyTemplate, cfg.MaxRetries, cfg.RetryDelaySeconds, cfg.Enabled,
	)
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("insert reminder config: %w", err)
	}
	id := lastInsertID(result)
	log.complete(nil,
		zap.Int64("reminder_config_id", id),
		zap.Int64("rows_affected", rowsAffected(result)),
	)
	return r.GetByID(ctx, id, cfg.UserID)
}

func (r *ReminderConfigRepo) GetByID(ctx context.Context, id, userID int64) (*models.UserReminderConfig, error) {
	log := beginDBOperation(ctx, reminderConfigRepositoryName, "get_reminder_config_by_id",
		zap.Int64("reminder_config_id", id),
		zap.Int64("user_id", userID),
	)
	c := &models.UserReminderConfig{}
	var headersRaw string
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, name, channel_type, webhook_url, webhook_method, webhook_headers,
		       webhook_body_template, max_retries, retry_delay_seconds, enabled, created_at, updated_at
		FROM user_reminder_configs WHERE id = ? AND user_id = ?`, id, userID,
	).Scan(&c.ID, &c.UserID, &c.Name, &c.ChannelType, &c.WebhookURL, &c.WebhookMethod,
		&headersRaw, &c.WebhookBodyTemplate, &c.MaxRetries, &c.RetryDelaySeconds, &c.Enabled, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		log.complete(nil, zap.Bool("found", false))
		return nil, nil
	}
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("get reminder config: %w", err)
	}
	c.WebhookHeaders = models.ParseWebhookHeaders(headersRaw)
	log.complete(nil,
		zap.Bool("found", true),
		zap.Bool("enabled", c.Enabled),
	)
	return c, nil
}

func (r *ReminderConfigRepo) GetByUserID(ctx context.Context, userID int64) ([]models.UserReminderConfig, error) {
	log := beginDBOperation(ctx, reminderConfigRepositoryName, "list_reminder_configs",
		zap.Int64("user_id", userID),
	)
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, name, channel_type, webhook_url, webhook_method, webhook_headers,
		       webhook_body_template, max_retries, retry_delay_seconds, enabled, created_at, updated_at
		FROM user_reminder_configs WHERE user_id = ? ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("list reminder configs: %w", err)
	}
	defer rows.Close()

	configs := make([]models.UserReminderConfig, 0)
	for rows.Next() {
		var c models.UserReminderConfig
		var headersRaw string
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.ChannelType, &c.WebhookURL, &c.WebhookMethod,
			&headersRaw, &c.WebhookBodyTemplate, &c.MaxRetries, &c.RetryDelaySeconds, &c.Enabled, &c.CreatedAt, &c.UpdatedAt); err != nil {
			log.complete(err)
			return nil, fmt.Errorf("scan reminder config: %w", err)
		}
		c.WebhookHeaders = models.ParseWebhookHeaders(headersRaw)
		configs = append(configs, c)
	}
	if err := rows.Err(); err != nil {
		log.complete(err)
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	log.complete(nil, zap.Int("result_size", len(configs)))
	return configs, nil
}

func (r *ReminderConfigRepo) Update(ctx context.Context, id, userID int64, req models.UpdateReminderConfigRequest) (*models.UserReminderConfig, error) {
	log := beginDBOperation(ctx, reminderConfigRepositoryName, "update_reminder_config",
		zap.Int64("reminder_config_id", id),
		zap.Int64("user_id", userID),
		zap.Bool("update_name", req.Name != nil),
		zap.Bool("update_channel_type", req.ChannelType != nil),
		zap.Bool("update_webhook_url", req.WebhookURL != nil),
		zap.Bool("update_webhook_method", req.WebhookMethod != nil),
		zap.Bool("update_webhook_headers", req.WebhookHeaders != nil),
		zap.Bool("update_body_template", req.WebhookBodyTemplate != nil),
		zap.Bool("update_max_retries", req.MaxRetries != nil),
		zap.Bool("update_retry_delay_seconds", req.RetryDelaySeconds != nil),
		zap.Bool("update_enabled", req.Enabled != nil),
	)
	setClauses := []string{}
	args := []any{}

	if req.Name != nil {
		setClauses = append(setClauses, "name = ?")
		args = append(args, *req.Name)
	}
	if req.ChannelType != nil {
		setClauses = append(setClauses, "channel_type = ?")
		args = append(args, *req.ChannelType)
	}
	if req.WebhookURL != nil {
		setClauses = append(setClauses, "webhook_url = ?")
		args = append(args, *req.WebhookURL)
	}
	if req.WebhookMethod != nil {
		setClauses = append(setClauses, "webhook_method = ?")
		args = append(args, *req.WebhookMethod)
	}
	if req.WebhookHeaders != nil {
		setClauses = append(setClauses, "webhook_headers = ?")
		args = append(args, (&models.UserReminderConfig{WebhookHeaders: req.WebhookHeaders}).GetWebhookHeadersJSON())
	}
	if req.WebhookBodyTemplate != nil {
		setClauses = append(setClauses, "webhook_body_template = ?")
		args = append(args, *req.WebhookBodyTemplate)
	}
	if req.MaxRetries != nil {
		setClauses = append(setClauses, "max_retries = ?")
		args = append(args, *req.MaxRetries)
	}
	if req.RetryDelaySeconds != nil {
		setClauses = append(setClauses, "retry_delay_seconds = ?")
		args = append(args, *req.RetryDelaySeconds)
	}
	if req.Enabled != nil {
		setClauses = append(setClauses, "enabled = ?")
		args = append(args, *req.Enabled)
	}

	if len(setClauses) == 0 {
		log.complete(nil, zap.String("result", "no_fields_to_update"))
		return r.GetByID(ctx, id, userID)
	}

	setClauses = append(setClauses, "updated_at = ?")
	args = append(args, time.Now().UTC().Format(time.RFC3339), id, userID)

	query := "UPDATE user_reminder_configs SET " + joinClauses(setClauses) + " WHERE id = ? AND user_id = ?"
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("update reminder config: %w", err)
	}
	rows := rowsAffected(result)
	if rows == 0 {
		log.complete(nil, zap.Int64("rows_affected", rows), zap.Bool("found", false))
		return nil, nil
	}
	log.complete(nil, zap.Int64("rows_affected", rows))
	return r.GetByID(ctx, id, userID)
}

func (r *ReminderConfigRepo) Delete(ctx context.Context, id, userID int64) (bool, error) {
	log := beginDBOperation(ctx, reminderConfigRepositoryName, "delete_reminder_config",
		zap.Int64("reminder_config_id", id),
		zap.Int64("user_id", userID),
	)
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM user_reminder_configs WHERE id = ? AND user_id = ?`, id, userID,
	)
	if err != nil {
		log.complete(err)
		return false, fmt.Errorf("delete reminder config: %w", err)
	}
	rows := rowsAffected(result)
	log.complete(nil,
		zap.Int64("rows_affected", rows),
		zap.Bool("deleted", rows > 0),
	)
	return rows > 0, nil
}

func (r *ReminderConfigRepo) HasEnabledByUserID(ctx context.Context, userID int64) (bool, error) {
	log := beginDBOperation(ctx, reminderConfigRepositoryName, "has_enabled_reminder_config",
		zap.Int64("user_id", userID),
	)
	var count int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM user_reminder_configs WHERE user_id = ? AND enabled = 1`,
		userID,
	).Scan(&count); err != nil {
		log.complete(err)
		return false, fmt.Errorf("count enabled reminder configs: %w", err)
	}
	log.complete(nil,
		zap.Int("count", count),
		zap.Bool("has_enabled", count > 0),
	)
	return count > 0, nil
}

func (r *ReminderConfigRepo) ListAll(ctx context.Context, page, limit int) ([]models.UserReminderConfig, int64, error) {
	log := beginDBOperation(ctx, reminderConfigRepositoryName, "admin_list_all_reminder_configs",
		zap.Int("page", page),
		zap.Int("limit", limit),
	)
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_reminder_configs`).Scan(&total); err != nil {
		log.complete(err)
		return nil, 0, fmt.Errorf("count reminder configs: %w", err)
	}

	offset := (page - 1) * limit
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, name, channel_type, webhook_url, webhook_method, webhook_headers,
		       webhook_body_template, max_retries, retry_delay_seconds, enabled, created_at, updated_at
		FROM user_reminder_configs ORDER BY user_id ASC, id ASC LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		log.complete(err, zap.Int64("count", total))
		return nil, 0, fmt.Errorf("list all reminder configs: %w", err)
	}
	defer rows.Close()

	configs := make([]models.UserReminderConfig, 0)
	for rows.Next() {
		var c models.UserReminderConfig
		var headersRaw string
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.ChannelType, &c.WebhookURL, &c.WebhookMethod,
			&headersRaw, &c.WebhookBodyTemplate, &c.MaxRetries, &c.RetryDelaySeconds, &c.Enabled, &c.CreatedAt, &c.UpdatedAt); err != nil {
			log.complete(err, zap.Int64("count", total))
			return nil, 0, fmt.Errorf("scan reminder config: %w", err)
		}
		c.WebhookHeaders = models.ParseWebhookHeaders(headersRaw)
		configs = append(configs, c)
	}
	if err := rows.Err(); err != nil {
		log.complete(err, zap.Int64("count", total))
		return nil, 0, fmt.Errorf("rows iteration: %w", err)
	}
	log.complete(nil, zap.Int64("count", total), zap.Int("result_size", len(configs)))
	return configs, total, nil
}

func joinClauses(clauses []string) string {
	s := ""
	for i, c := range clauses {
		if i > 0 {
			s += ", "
		}
		s += c
	}
	return s
}

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"todo/internal/models"
)

type ReminderConfigRepo struct {
	db *sql.DB
}

func NewReminderConfigRepo(db *sql.DB) *ReminderConfigRepo {
	return &ReminderConfigRepo{db: db}
}

func (r *ReminderConfigRepo) Create(ctx context.Context, cfg *models.UserReminderConfig) (*models.UserReminderConfig, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO user_reminder_configs
			(user_id, name, channel_type, webhook_url, webhook_method, webhook_headers, webhook_body_template, max_retries, retry_delay_seconds, enabled)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		cfg.UserID, cfg.Name, cfg.ChannelType, cfg.WebhookURL, cfg.WebhookMethod,
		cfg.GetWebhookHeadersJSON(), cfg.WebhookBodyTemplate, cfg.MaxRetries, cfg.RetryDelaySeconds, cfg.Enabled,
	)
	if err != nil {
		return nil, fmt.Errorf("insert reminder config: %w", err)
	}
	id, _ := result.LastInsertId()
	return r.GetByID(ctx, id, cfg.UserID)
}

func (r *ReminderConfigRepo) GetByID(ctx context.Context, id, userID int64) (*models.UserReminderConfig, error) {
	c := &models.UserReminderConfig{}
	var headersRaw string
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, name, channel_type, webhook_url, webhook_method, webhook_headers,
		       webhook_body_template, max_retries, retry_delay_seconds, enabled, created_at, updated_at
		FROM user_reminder_configs WHERE id = ? AND user_id = ?`, id, userID,
	).Scan(&c.ID, &c.UserID, &c.Name, &c.ChannelType, &c.WebhookURL, &c.WebhookMethod,
		&headersRaw, &c.WebhookBodyTemplate, &c.MaxRetries, &c.RetryDelaySeconds, &c.Enabled, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get reminder config: %w", err)
	}
	c.WebhookHeaders = models.ParseWebhookHeaders(headersRaw)
	return c, nil
}

func (r *ReminderConfigRepo) GetByUserID(ctx context.Context, userID int64) ([]models.UserReminderConfig, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, name, channel_type, webhook_url, webhook_method, webhook_headers,
		       webhook_body_template, max_retries, retry_delay_seconds, enabled, created_at, updated_at
		FROM user_reminder_configs WHERE user_id = ? ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list reminder configs: %w", err)
	}
	defer rows.Close()

	configs := make([]models.UserReminderConfig, 0)
	for rows.Next() {
		var c models.UserReminderConfig
		var headersRaw string
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.ChannelType, &c.WebhookURL, &c.WebhookMethod,
			&headersRaw, &c.WebhookBodyTemplate, &c.MaxRetries, &c.RetryDelaySeconds, &c.Enabled, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan reminder config: %w", err)
		}
		c.WebhookHeaders = models.ParseWebhookHeaders(headersRaw)
		configs = append(configs, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	return configs, nil
}

func (r *ReminderConfigRepo) Update(ctx context.Context, id, userID int64, req models.UpdateReminderConfigRequest) (*models.UserReminderConfig, error) {
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
		return r.GetByID(ctx, id, userID)
	}

	setClauses = append(setClauses, "updated_at = ?")
	args = append(args, time.Now().UTC().Format(time.RFC3339), id, userID)

	query := "UPDATE user_reminder_configs SET " + joinClauses(setClauses) + " WHERE id = ? AND user_id = ?"
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("update reminder config: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, nil
	}
	return r.GetByID(ctx, id, userID)
}

func (r *ReminderConfigRepo) Delete(ctx context.Context, id, userID int64) (bool, error) {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM user_reminder_configs WHERE id = ? AND user_id = ?`, id, userID,
	)
	if err != nil {
		return false, fmt.Errorf("delete reminder config: %w", err)
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}

func (r *ReminderConfigRepo) HasEnabledByUserID(ctx context.Context, userID int64) (bool, error) {
	var count int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM user_reminder_configs WHERE user_id = ? AND enabled = 1`,
		userID,
	).Scan(&count); err != nil {
		return false, fmt.Errorf("count enabled reminder configs: %w", err)
	}
	return count > 0, nil
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

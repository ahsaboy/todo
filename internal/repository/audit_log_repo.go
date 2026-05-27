package repository

import (
	"context"
	"database/sql"
	"fmt"

	"go.uber.org/zap"
)

const auditLogRepositoryName = "audit_log_repo"

type AuditLog struct {
	ID          int64  `json:"id"`
	AdminUserID int64  `json:"admin_user_id"`
	AdminName   string `json:"admin_name"`
	Action      string `json:"action"`
	TargetType  string `json:"target_type"`
	TargetID    *int64 `json:"target_id"`
	Detail      string `json:"detail"`
	CreatedAt   string `json:"created_at"`
}

type AuditLogRepo struct {
	db *sql.DB
}

func NewAuditLogRepo(db *sql.DB) *AuditLogRepo {
	return &AuditLogRepo{db: db}
}

func (r *AuditLogRepo) Create(ctx context.Context, adminUserID int64, action, targetType string, targetID *int64, detail string) error {
	log := beginDBOperation(ctx, auditLogRepositoryName, "create_audit_log",
		zap.Int64("admin_user_id", adminUserID),
		zap.String("action", action),
		zap.String("target_type", targetType),
	)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO admin_audit_logs (admin_user_id, action, target_type, target_id, detail) VALUES (?, ?, ?, ?, ?)`,
		adminUserID, action, targetType, targetID, detail,
	)
	if err != nil {
		log.complete(err)
		return fmt.Errorf("create audit log: %w", err)
	}
	log.complete(nil)
	return nil
}

func (r *AuditLogRepo) ListAll(ctx context.Context, page, limit int) ([]AuditLog, int64, error) {
	log := beginDBOperation(ctx, auditLogRepositoryName, "list_audit_logs",
		zap.Int("page", page),
		zap.Int("limit", limit),
	)
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM admin_audit_logs`).Scan(&total); err != nil {
		log.complete(err)
		return nil, 0, fmt.Errorf("count audit logs: %w", err)
	}

	offset := (page - 1) * limit
	rows, err := r.db.QueryContext(ctx, `
		SELECT l.id, l.admin_user_id, COALESCE(u.username, ''), l.action, l.target_type, l.target_id, l.detail, l.created_at
		FROM admin_audit_logs l
		LEFT JOIN users u ON u.id = l.admin_user_id
		ORDER BY l.created_at DESC, l.id DESC
		LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		log.complete(err, zap.Int64("count", total))
		return nil, 0, fmt.Errorf("list audit logs: %w", err)
	}
	defer rows.Close()

	logs := make([]AuditLog, 0)
	for rows.Next() {
		var item AuditLog
		var targetID sql.NullInt64
		if err := rows.Scan(
			&item.ID, &item.AdminUserID, &item.AdminName,
			&item.Action, &item.TargetType, &targetID, &item.Detail, &item.CreatedAt,
		); err != nil {
			log.complete(err, zap.Int64("count", total))
			return nil, 0, fmt.Errorf("scan audit log: %w", err)
		}
		if targetID.Valid {
			item.TargetID = &targetID.Int64
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

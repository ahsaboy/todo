package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"

	"todo/internal/models"
)

type APIKeyRepo struct {
	db *sql.DB
}

const apiKeyRepositoryName = "api_key_repo"

func NewAPIKeyRepo(db *sql.DB) *APIKeyRepo {
	return &APIKeyRepo{db: db}
}

func (r *APIKeyRepo) Create(ctx context.Context, userID int64, keyHash, name string) (*models.APIKey, error) {
	log := beginDBOperation(ctx, apiKeyRepositoryName, "create_api_key",
		zap.Int64("user_id", userID),
		zap.String("name", name),
		zap.String("key_fingerprint", apiKeyFingerprint(keyHash)),
	)
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO user_api_keys (user_id, key_hash, name) VALUES (?, ?, ?)`,
		userID, keyHash, name,
	)
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("insert api key: %w", err)
	}
	id := lastInsertID(result)
	log.complete(nil,
		zap.Int64("api_key_id", id),
		zap.Int64("rows_affected", rowsAffected(result)),
	)
	return &models.APIKey{ID: id, UserID: userID, KeyHash: keyHash, Name: name}, nil
}

func (r *APIKeyRepo) GetByUserID(ctx context.Context, userID int64) ([]models.APIKey, error) {
	log := beginDBOperation(ctx, apiKeyRepositoryName, "list_api_keys",
		zap.Int64("user_id", userID),
	)
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, key_hash, name, last_used_at, created_at FROM user_api_keys WHERE user_id = ? ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	defer rows.Close()

	keys := make([]models.APIKey, 0)
	for rows.Next() {
		var k models.APIKey
		if err := rows.Scan(&k.ID, &k.UserID, &k.KeyHash, &k.Name, &k.LastUsedAt, &k.CreatedAt); err != nil {
			log.complete(err)
			return nil, fmt.Errorf("scan api key: %w", err)
		}
		keys = append(keys, k)
	}
	if err := rows.Err(); err != nil {
		log.complete(err)
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	log.complete(nil, zap.Int("result_size", len(keys)))
	return keys, nil
}

func (r *APIKeyRepo) Delete(ctx context.Context, id, userID int64) (bool, error) {
	log := beginDBOperation(ctx, apiKeyRepositoryName, "delete_api_key",
		zap.Int64("api_key_id", id),
		zap.Int64("user_id", userID),
	)
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM user_api_keys WHERE id = ? AND user_id = ?`, id, userID,
	)
	if err != nil {
		log.complete(err)
		return false, fmt.Errorf("delete api key: %w", err)
	}
	rows := rowsAffected(result)
	log.complete(nil,
		zap.Int64("rows_affected", rows),
		zap.Bool("deleted", rows > 0),
	)
	return rows > 0, nil
}

func (r *APIKeyRepo) ValidateKey(ctx context.Context, keyHash string) (int64, error) {
	log := beginDBOperation(ctx, apiKeyRepositoryName, "validate_api_key",
		zap.String("key_fingerprint", apiKeyFingerprint(keyHash)),
	)
	var userID int64
	err := r.db.QueryRowContext(ctx,
		`SELECT user_id FROM user_api_keys WHERE key_hash = ?`, keyHash,
	).Scan(&userID)
	if err == sql.ErrNoRows {
		log.complete(err, zap.Bool("valid", false))
		return 0, fmt.Errorf("invalid api key")
	}
	if err != nil {
		log.complete(err)
		return 0, fmt.Errorf("validate api key: %w", err)
	}

	// 更新 last_used_at(非关键统计字段,失败时降级记录但不阻断验证)
	updateResult, updateErr := r.db.ExecContext(ctx,
		`UPDATE user_api_keys SET last_used_at = ? WHERE key_hash = ?`,
		time.Now().UTC().Format(time.RFC3339), keyHash,
	)
	if updateErr != nil {
		log.complete(nil,
			zap.Int64("user_id", userID),
			zap.Bool("valid", true),
			zap.NamedError("last_used_update_error", updateErr),
		)
		return userID, nil
	}

	log.complete(nil,
		zap.Int64("user_id", userID),
		zap.Bool("valid", true),
		zap.Int64("rows_affected", rowsAffected(updateResult)),
	)
	return userID, nil
}

func (r *APIKeyRepo) DeleteByID(ctx context.Context, id, userID int64) (bool, error) {
	log := beginDBOperation(ctx, apiKeyRepositoryName, "delete_api_key_by_id",
		zap.Int64("api_key_id", id),
		zap.Int64("user_id", userID),
	)
	result, err := r.db.ExecContext(ctx, `DELETE FROM user_api_keys WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		log.complete(err)
		return false, fmt.Errorf("delete api key: %w", err)
	}
	rows := rowsAffected(result)
	log.complete(nil,
		zap.Int64("rows_affected", rows),
		zap.Bool("deleted", rows > 0),
	)
	return rows > 0, nil
}

// CleanupExpiredLoginKeys 删除指定用户下名为 "login" 且 last_used_at 超过 24 小时的 API Key。
func (r *APIKeyRepo) CleanupExpiredLoginKeys(ctx context.Context, userID int64) (int64, error) {
	log := beginDBOperation(ctx, apiKeyRepositoryName, "cleanup_expired_login_keys",
		zap.Int64("user_id", userID),
	)
	cutoff := time.Now().UTC().Add(-24 * time.Hour).Format(time.RFC3339)
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM user_api_keys WHERE user_id = ? AND name = 'login' AND last_used_at IS NOT NULL AND last_used_at < ?`,
		userID, cutoff,
	)
	if err != nil {
		log.complete(err)
		return 0, fmt.Errorf("cleanup expired login keys: %w", err)
	}
	rows := rowsAffected(result)
	log.complete(nil, zap.Int64("rows_affected", rows))
	return rows, nil
}

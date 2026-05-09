package repository

import (
	"context"
	"database/sql"
	"fmt"

	"todo/internal/models"
)

type APIKeyRepo struct {
	db *sql.DB
}

func NewAPIKeyRepo(db *sql.DB) *APIKeyRepo {
	return &APIKeyRepo{db: db}
}

func (r *APIKeyRepo) Create(ctx context.Context, userID int64, keyHash, name string) (*models.APIKey, error) {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO user_api_keys (user_id, key_hash, name) VALUES (?, ?, ?)`,
		userID, keyHash, name,
	)
	if err != nil {
		return nil, fmt.Errorf("insert api key: %w", err)
	}
	id, _ := result.LastInsertId()
	return &models.APIKey{ID: id, UserID: userID, KeyHash: keyHash, Name: name}, nil
}

func (r *APIKeyRepo) GetByUserID(ctx context.Context, userID int64) ([]models.APIKey, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, key_hash, name, last_used_at, created_at FROM user_api_keys WHERE user_id = ? ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	defer rows.Close()

	keys := make([]models.APIKey, 0)
	for rows.Next() {
		var k models.APIKey
		if err := rows.Scan(&k.ID, &k.UserID, &k.KeyHash, &k.Name, &k.LastUsedAt, &k.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan api key: %w", err)
		}
		keys = append(keys, k)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	return keys, nil
}

func (r *APIKeyRepo) Delete(ctx context.Context, id, userID int64) (bool, error) {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM user_api_keys WHERE id = ? AND user_id = ?`, id, userID,
	)
	if err != nil {
		return false, fmt.Errorf("delete api key: %w", err)
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}

func (r *APIKeyRepo) ValidateKey(ctx context.Context, keyHash string) (int64, error) {
	var userID int64
	err := r.db.QueryRowContext(ctx,
		`SELECT user_id FROM user_api_keys WHERE key_hash = ?`, keyHash,
	).Scan(&userID)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("invalid api key")
	}
	if err != nil {
		return 0, fmt.Errorf("validate api key: %w", err)
	}

	// 更新 last_used_at
	r.db.ExecContext(ctx,
		`UPDATE user_api_keys SET last_used_at = datetime('now','localtime') WHERE key_hash = ?`, keyHash,
	)

	return userID, nil
}

func (r *APIKeyRepo) DeleteByID(ctx context.Context, id, userID int64) (bool, error) {
	result, err := r.db.ExecContext(ctx, `DELETE FROM user_api_keys WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return false, fmt.Errorf("delete api key: %w", err)
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}

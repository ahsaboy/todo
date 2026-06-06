package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type AppConfigRepo struct {
	db *sql.DB
}

const appConfigRepositoryName = "app_config_repo"

func NewAppConfigRepo(db *sql.DB) *AppConfigRepo {
	return &AppConfigRepo{db: db}
}

// LoadAll 返回 key→JSON 编码 value 的全部配置。
func (r *AppConfigRepo) LoadAll(ctx context.Context) (map[string]string, error) {
	log := beginDBOperation(ctx, appConfigRepositoryName, "load_all")
	rows, err := r.db.QueryContext(ctx, `SELECT key, value FROM app_config`)
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("load all app_config: %w", err)
	}
	defer rows.Close()

	out := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			log.complete(err)
			return nil, fmt.Errorf("scan app_config: %w", err)
		}
		out[k] = v
	}
	if err := rows.Err(); err != nil {
		log.complete(err)
		return nil, fmt.Errorf("iterate app_config: %w", err)
	}
	log.complete(nil, zap.Int("count", len(out)))
	return out, nil
}

// Upsert 写入或更新某个 key 的 JSON 值。
func (r *AppConfigRepo) Upsert(ctx context.Context, key, value string, updatedBy int64) error {
	log := beginDBOperation(ctx, appConfigRepositoryName, "upsert", zap.String("key", key))
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO app_config (key, value, updated_at, updated_by)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			updated_at = excluded.updated_at,
			updated_by = excluded.updated_by`,
		key, value, now, updatedBy,
	)
	if err != nil {
		log.complete(err)
		return fmt.Errorf("upsert app_config: %w", err)
	}
	log.complete(nil)
	return nil
}

// Delete 删除某个 key(等价于恢复默认,回退到配置文件/环境变量)。
func (r *AppConfigRepo) Delete(ctx context.Context, key string) error {
	log := beginDBOperation(ctx, appConfigRepositoryName, "delete", zap.String("key", key))
	_, err := r.db.ExecContext(ctx, `DELETE FROM app_config WHERE key = ?`, key)
	if err != nil {
		log.complete(err)
		return fmt.Errorf("delete app_config: %w", err)
	}
	log.complete(nil)
	return nil
}

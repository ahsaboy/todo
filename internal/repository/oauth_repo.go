package repository

import (
	"context"
	"database/sql"
	"fmt"

	"go.uber.org/zap"

	"todo/internal/models"
)

type OAuthRepo struct {
	db *sql.DB
}

const oauthRepoName = "oauth_repo"

func NewOAuthRepo(db *sql.DB) *OAuthRepo {
	return &OAuthRepo{db: db}
}

func (r *OAuthRepo) Create(ctx context.Context, userID int64, provider, providerUserID, displayName, avatarURL string) error {
	log := beginDBOperation(ctx, oauthRepoName, "create_oauth_account",
		zap.Int64("user_id", userID),
		zap.String("provider", provider),
		zap.String("provider_user_id", providerUserID),
	)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO user_oauth_accounts (user_id, provider, provider_user_id, display_name, avatar_url) VALUES (?, ?, ?, ?, ?)`,
		userID, provider, providerUserID, displayName, avatarURL,
	)
	if err != nil {
		log.complete(err)
		return fmt.Errorf("insert oauth account: %w", err)
	}
	log.complete(nil)
	return nil
}

func (r *OAuthRepo) GetByProvider(ctx context.Context, provider, providerUserID string) (*models.OAuthAccount, error) {
	log := beginDBOperation(ctx, oauthRepoName, "get_oauth_by_provider",
		zap.String("provider", provider),
		zap.String("provider_user_id", providerUserID),
	)
	oa := &models.OAuthAccount{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, provider, provider_user_id, COALESCE(display_name, ''), COALESCE(avatar_url, ''), created_at FROM user_oauth_accounts WHERE provider = ? AND provider_user_id = ?`,
		provider, providerUserID,
	).Scan(&oa.ID, &oa.UserID, &oa.Provider, &oa.ProviderUserID, &oa.DisplayName, &oa.AvatarURL, &oa.CreatedAt)
	if err == sql.ErrNoRows {
		log.complete(nil, zap.Bool("found", false))
		return nil, nil
	}
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("get oauth account: %w", err)
	}
	log.complete(nil, zap.Bool("found", true), zap.Int64("user_id", oa.UserID))
	return oa, nil
}

func (r *OAuthRepo) GetByUserID(ctx context.Context, userID int64) ([]models.OAuthAccount, error) {
	log := beginDBOperation(ctx, oauthRepoName, "get_oauth_by_user_id", zap.Int64("user_id", userID))
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, provider, provider_user_id, COALESCE(display_name, ''), COALESCE(avatar_url, ''), created_at FROM user_oauth_accounts WHERE user_id = ?`,
		userID,
	)
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("list oauth accounts: %w", err)
	}
	defer rows.Close()

	var accounts []models.OAuthAccount
	for rows.Next() {
		var oa models.OAuthAccount
		if err := rows.Scan(&oa.ID, &oa.UserID, &oa.Provider, &oa.ProviderUserID, &oa.DisplayName, &oa.AvatarURL, &oa.CreatedAt); err != nil {
			log.complete(err)
			return nil, fmt.Errorf("scan oauth account: %w", err)
		}
		accounts = append(accounts, oa)
	}
	log.complete(nil, zap.Int("count", len(accounts)))
	return accounts, nil
}

func (r *OAuthRepo) Delete(ctx context.Context, id, userID int64) (bool, error) {
	log := beginDBOperation(ctx, oauthRepoName, "delete_oauth_account",
		zap.Int64("id", id),
		zap.Int64("user_id", userID),
	)
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM user_oauth_accounts WHERE id = ? AND user_id = ?`,
		id, userID,
	)
	if err != nil {
		log.complete(err)
		return false, fmt.Errorf("delete oauth account: %w", err)
	}
	ra := rowsAffected(result)
	log.complete(nil, zap.Int64("rows_affected", ra))
	return ra > 0, nil
}

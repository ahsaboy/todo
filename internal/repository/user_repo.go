package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"todo/internal/models"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, username, email, passwordHash string) (*models.User, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create user: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx,
		`INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)`,
		username, email, passwordHash,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			return nil, ErrUsernameTaken
		}
		return nil, fmt.Errorf("insert user: %w", err)
	}
	id, _ := result.LastInsertId()

	// 升级自单用户版本时，历史任务可能仍未归属到任何用户。
	if _, err := tx.ExecContext(ctx,
		`UPDATE tasks SET user_id = ? WHERE user_id IS NULL`,
		id,
	); err != nil {
		return nil, fmt.Errorf("claim legacy tasks: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create user: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*models.User, error) {
	u := &models.User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE id = ?`, id,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return u, nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	u := &models.User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE username = ?`, username,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return u, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	u := &models.User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE email = ?`, email,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return u, nil
}

func (r *UserRepo) UpdateProfile(ctx context.Context, id int64, email string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET email = ?, updated_at = datetime('now','localtime') WHERE id = ?`,
		email, id,
	)
	return err
}

func (r *UserRepo) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET password_hash = ?, updated_at = datetime('now','localtime') WHERE id = ?`,
		passwordHash, id,
	)
	return err
}

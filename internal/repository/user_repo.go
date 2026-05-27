package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"todo/internal/models"
)

type UserRepo struct {
	db *sql.DB
}

const userRepositoryName = "user_repo"

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, username, email, passwordHash string) (*models.User, error) {
	log := beginDBOperation(ctx, userRepositoryName, "create_user",
		zap.String("username", username),
		zap.String("email", email),
	)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("begin create user: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx,
		`INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)`,
		username, email, passwordHash,
	)
	if err != nil {
		log.complete(err)
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			return nil, ErrUsernameTaken
		}
		return nil, fmt.Errorf("insert user: %w", err)
	}
	id := lastInsertID(result)

	// 升级自单用户版本时，历史任务可能仍未归属到任何用户。
	claimResult, err := tx.ExecContext(ctx,
		`UPDATE tasks SET user_id = ? WHERE user_id IS NULL`,
		id,
	)
	if err != nil {
		log.complete(err, zap.Int64("user_id", id))
		return nil, fmt.Errorf("claim legacy tasks: %w", err)
	}

	if err := tx.Commit(); err != nil {
		log.complete(err, zap.Int64("user_id", id))
		return nil, fmt.Errorf("commit create user: %w", err)
	}

	log.complete(nil,
		zap.Int64("user_id", id),
		zap.Int64("rows_affected", rowsAffected(result)),
		zap.Int64("claimed_legacy_tasks", rowsAffected(claimResult)),
	)
	return r.GetByID(ctx, id)
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*models.User, error) {
	log := beginDBOperation(ctx, userRepositoryName, "get_user_by_id", zap.Int64("user_id", id))
	u := &models.User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, email, password_hash, is_admin, created_at, updated_at FROM users WHERE id = ?`, id,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		log.complete(nil, zap.Bool("found", false))
		return nil, nil
	}
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("get user: %w", err)
	}
	log.complete(nil, zap.Bool("found", true))
	return u, nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	log := beginDBOperation(ctx, userRepositoryName, "get_user_by_username", zap.String("username", username))
	u := &models.User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, email, password_hash, is_admin, created_at, updated_at FROM users WHERE username = ?`, username,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		log.complete(nil, zap.Bool("found", false))
		return nil, nil
	}
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	log.complete(nil,
		zap.Bool("found", true),
		zap.Int64("user_id", u.ID),
	)
	return u, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	log := beginDBOperation(ctx, userRepositoryName, "get_user_by_email", zap.String("email", email))
	u := &models.User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, email, password_hash, is_admin, created_at, updated_at FROM users WHERE email = ?`, email,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		log.complete(nil, zap.Bool("found", false))
		return nil, nil
	}
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	log.complete(nil,
		zap.Bool("found", true),
		zap.Int64("user_id", u.ID),
	)
	return u, nil
}

func (r *UserRepo) UpdateProfile(ctx context.Context, id int64, email string) error {
	log := beginDBOperation(ctx, userRepositoryName, "update_user_profile",
		zap.Int64("user_id", id),
		zap.String("email", email),
	)
	result, err := r.db.ExecContext(ctx,
		`UPDATE users SET email = ?, updated_at = ? WHERE id = ?`,
		email, time.Now().UTC().Format(time.RFC3339), id,
	)
	if err != nil {
		log.complete(err)
		return err
	}
	log.complete(nil, zap.Int64("rows_affected", rowsAffected(result)))
	return nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	log := beginDBOperation(ctx, userRepositoryName, "update_user_password",
		zap.Int64("user_id", id),
	)
	result, err := r.db.ExecContext(ctx,
		`UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?`,
		passwordHash, time.Now().UTC().Format(time.RFC3339), id,
	)
	if err != nil {
		log.complete(err)
		return err
	}
	log.complete(nil, zap.Int64("rows_affected", rowsAffected(result)))
	return nil
}

func (r *UserRepo) ListAll(ctx context.Context, page, limit int, search string) ([]models.User, int64, error) {
	log := beginDBOperation(ctx, userRepositoryName, "admin_list_all_users",
		zap.Int("page", page),
		zap.Int("limit", limit),
		zap.Bool("has_search", search != ""),
	)

	var total int64
	countQuery := `SELECT COUNT(*) FROM users`
	args := []any{}
	if search != "" {
		countQuery += ` WHERE username LIKE ? OR email LIKE ?`
		args = append(args, "%"+search+"%", "%"+search+"%")
	}
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		log.complete(err)
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	offset := (page - 1) * limit
	query := `SELECT id, username, email, password_hash, is_admin, created_at, updated_at FROM users`
	if search != "" {
		query += ` WHERE username LIKE ? OR email LIKE ?`
	}
	query += ` ORDER BY id DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		log.complete(err, zap.Int64("count", total))
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]models.User, 0)
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt); err != nil {
			log.complete(err, zap.Int64("count", total))
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		log.complete(err, zap.Int64("count", total))
		return nil, 0, fmt.Errorf("rows iteration: %w", err)
	}
	log.complete(nil, zap.Int64("count", total), zap.Int("result_size", len(users)))
	return users, total, nil
}

func (r *UserRepo) Delete(ctx context.Context, id int64) error {
	log := beginDBOperation(ctx, userRepositoryName, "admin_delete_user", zap.Int64("user_id", id))
	result, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	if err != nil {
		log.complete(err)
		return fmt.Errorf("delete user: %w", err)
	}
	ra := rowsAffected(result)
	if ra == 0 {
		log.complete(nil, zap.Int64("rows_affected", 0))
		return ErrNotFound
	}
	log.complete(nil, zap.Int64("rows_affected", ra))
	return nil
}

func (r *UserRepo) ForceResetPassword(ctx context.Context, id int64, passwordHash string) error {
	log := beginDBOperation(ctx, userRepositoryName, "admin_force_reset_password", zap.Int64("user_id", id))
	result, err := r.db.ExecContext(ctx,
		`UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?`,
		passwordHash, time.Now().UTC().Format(time.RFC3339), id,
	)
	if err != nil {
		log.complete(err)
		return fmt.Errorf("force reset password: %w", err)
	}
	ra := rowsAffected(result)
	if ra == 0 {
		log.complete(nil, zap.Int64("rows_affected", 0))
		return ErrNotFound
	}
	log.complete(nil, zap.Int64("rows_affected", ra))
	return nil
}

func (r *UserRepo) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	log := beginDBOperation(ctx, userRepositoryName, "check_is_admin", zap.Int64("user_id", userID))
	var isAdmin int
	err := r.db.QueryRowContext(ctx, `SELECT is_admin FROM users WHERE id = ?`, userID).Scan(&isAdmin)
	if err == sql.ErrNoRows {
		log.complete(nil, zap.Bool("found", false))
		return false, nil
	}
	if err != nil {
		log.complete(err)
		return false, fmt.Errorf("check is_admin: %w", err)
	}
	log.complete(nil, zap.Bool("is_admin", isAdmin == 1))
	return isAdmin == 1, nil
}

func (r *UserRepo) SetIsAdmin(ctx context.Context, userID int64, isAdmin bool) error {
	log := beginDBOperation(ctx, userRepositoryName, "set_is_admin",
		zap.Int64("user_id", userID),
		zap.Bool("is_admin", isAdmin),
	)
	val := 0
	if isAdmin {
		val = 1
	}
	result, err := r.db.ExecContext(ctx,
		`UPDATE users SET is_admin = ?, updated_at = ? WHERE id = ?`,
		val, time.Now().UTC().Format(time.RFC3339), userID,
	)
	if err != nil {
		log.complete(err)
		return fmt.Errorf("set is_admin: %w", err)
	}
	rows := rowsAffected(result)
	log.complete(nil, zap.Int64("rows_affected", rows))
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

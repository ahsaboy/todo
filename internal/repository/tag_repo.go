package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"todo/internal/models"
)

// ErrTagNameTaken 表示同一用户下已存在同名标签。service 层会包装为面向调用方的错误。
var ErrTagNameTaken = errors.New("tag name already exists")

type TagRepo struct {
	db *sql.DB
}

const tagRepositoryName = "tag_repo"

func NewTagRepo(db *sql.DB) *TagRepo {
	return &TagRepo{db: db}
}

func (r *TagRepo) Create(ctx context.Context, t *models.UserTag) (*models.UserTag, error) {
	log := beginDBOperation(ctx, tagRepositoryName, "create_tag",
		zap.Int64("user_id", t.UserID),
		zap.String("name", t.Name),
		zap.String("color", t.Color),
		zap.String("icon", t.Icon),
	)
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO user_tags (user_id, name, color, icon, sort_order, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		t.UserID, t.Name, t.Color, t.Icon, t.SortOrder, now, now,
	)
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("insert tag: %w", err)
	}
	id := lastInsertID(result)
	log.complete(nil,
		zap.Int64("tag_id", id),
		zap.Int64("rows_affected", rowsAffected(result)),
	)
	return r.GetByID(ctx, id, t.UserID)
}

func (r *TagRepo) GetByID(ctx context.Context, id, userID int64) (*models.UserTag, error) {
	log := beginDBOperation(ctx, tagRepositoryName, "get_tag_by_id",
		zap.Int64("tag_id", id),
		zap.Int64("user_id", userID),
	)
	t := &models.UserTag{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, name, color, icon, sort_order, created_at, updated_at
		FROM user_tags WHERE id = ? AND user_id = ?`, id, userID,
	).Scan(&t.ID, &t.UserID, &t.Name, &t.Color, &t.Icon, &t.SortOrder, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		log.complete(nil, zap.Bool("found", false))
		return nil, nil
	}
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("get tag: %w", err)
	}
	log.complete(nil, zap.Bool("found", true))
	return t, nil
}

func (r *TagRepo) GetByName(ctx context.Context, userID int64, name string) (*models.UserTag, error) {
	log := beginDBOperation(ctx, tagRepositoryName, "get_tag_by_name",
		zap.Int64("user_id", userID),
		zap.String("name", name),
	)
	t := &models.UserTag{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, name, color, icon, sort_order, created_at, updated_at
		FROM user_tags WHERE user_id = ? AND name = ?`, userID, name,
	).Scan(&t.ID, &t.UserID, &t.Name, &t.Color, &t.Icon, &t.SortOrder, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		log.complete(nil, zap.Bool("found", false))
		return nil, nil
	}
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("get tag by name: %w", err)
	}
	log.complete(nil, zap.Bool("found", true))
	return t, nil
}

func (r *TagRepo) ListByUserID(ctx context.Context, userID int64) ([]models.UserTag, error) {
	log := beginDBOperation(ctx, tagRepositoryName, "list_tags",
		zap.Int64("user_id", userID),
	)
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, name, color, icon, sort_order, created_at, updated_at
		FROM user_tags WHERE user_id = ?
		ORDER BY sort_order ASC, id ASC`, userID)
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("list tags: %w", err)
	}
	defer rows.Close()

	tags := make([]models.UserTag, 0)
	for rows.Next() {
		var t models.UserTag
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Color, &t.Icon, &t.SortOrder, &t.CreatedAt, &t.UpdatedAt); err != nil {
			log.complete(err)
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, t)
	}
	if err := rows.Err(); err != nil {
		log.complete(err)
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	log.complete(nil, zap.Int("result_size", len(tags)))
	return tags, nil
}

// GetNamesSet 一次性返回该用户全部标签名集合,供 service 层做批量校验。
func (r *TagRepo) GetNamesSet(ctx context.Context, userID int64) (map[string]struct{}, error) {
	log := beginDBOperation(ctx, tagRepositoryName, "get_tag_names_set",
		zap.Int64("user_id", userID),
	)
	rows, err := r.db.QueryContext(ctx, `SELECT name FROM user_tags WHERE user_id = ?`, userID)
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("query tag names: %w", err)
	}
	defer rows.Close()

	names := make(map[string]struct{})
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			log.complete(err)
			return nil, fmt.Errorf("scan tag name: %w", err)
		}
		names[n] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		log.complete(err)
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	log.complete(nil, zap.Int("result_size", len(names)))
	return names, nil
}

// Update 不处理改名时任务侧的 JSON 同步,该同步由 service 层在事务里完成。
func (r *TagRepo) Update(ctx context.Context, id, userID int64, name *string, color *string, icon *string, sortOrder *int) (*models.UserTag, error) {
	log := beginDBOperation(ctx, tagRepositoryName, "update_tag",
		zap.Int64("tag_id", id),
		zap.Int64("user_id", userID),
		zap.Bool("update_name", name != nil),
		zap.Bool("update_color", color != nil),
		zap.Bool("update_icon", icon != nil),
		zap.Bool("update_sort_order", sortOrder != nil),
	)
	setClauses := []string{}
	args := []any{}

	if name != nil {
		setClauses = append(setClauses, "name = ?")
		args = append(args, *name)
	}
	if color != nil {
		setClauses = append(setClauses, "color = ?")
		args = append(args, *color)
	}
	if icon != nil {
		setClauses = append(setClauses, "icon = ?")
		args = append(args, *icon)
	}
	if sortOrder != nil {
		setClauses = append(setClauses, "sort_order = ?")
		args = append(args, *sortOrder)
	}

	if len(setClauses) == 0 {
		log.complete(nil, zap.String("result", "no_fields_to_update"))
		return r.GetByID(ctx, id, userID)
	}

	setClauses = append(setClauses, "updated_at = ?")
	args = append(args, time.Now().UTC().Format(time.RFC3339), id, userID)

	query := "UPDATE user_tags SET " + strings.Join(setClauses, ", ") + " WHERE id = ? AND user_id = ?"
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("update tag: %w", err)
	}
	rows := rowsAffected(result)
	if rows == 0 {
		log.complete(nil, zap.Int64("rows_affected", rows), zap.Bool("found", false))
		return nil, nil
	}
	log.complete(nil, zap.Int64("rows_affected", rows))
	return r.GetByID(ctx, id, userID)
}

func (r *TagRepo) Delete(ctx context.Context, id, userID int64) (bool, error) {
	log := beginDBOperation(ctx, tagRepositoryName, "delete_tag",
		zap.Int64("tag_id", id),
		zap.Int64("user_id", userID),
	)
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM user_tags WHERE id = ? AND user_id = ?`, id, userID,
	)
	if err != nil {
		log.complete(err)
		return false, fmt.Errorf("delete tag: %w", err)
	}
	rows := rowsAffected(result)
	log.complete(nil,
		zap.Int64("rows_affected", rows),
		zap.Bool("deleted", rows > 0),
	)
	return rows > 0, nil
}

// RenameWithTaskSync 在单事务内重命名标签,并把该用户所有任务 tags JSON 中匹配的旧名替换为新名。
// 返回更新后的标签;若 id 不存在则返回 nil。新名重复返回 (nil, ErrTagNameTaken)。
func (r *TagRepo) RenameWithTaskSync(ctx context.Context, id, userID int64, newName string) (*models.UserTag, error) {
	log := beginDBOperation(ctx, tagRepositoryName, "rename_tag_with_task_sync",
		zap.Int64("tag_id", id),
		zap.Int64("user_id", userID),
		zap.String("new_name", newName),
	)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// 读旧名
	var oldName string
	err = tx.QueryRowContext(ctx, `SELECT name FROM user_tags WHERE id = ? AND user_id = ?`, id, userID).Scan(&oldName)
	if err == sql.ErrNoRows {
		log.complete(nil, zap.Bool("found", false))
		return nil, nil
	}
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("query old tag name: %w", err)
	}

	if oldName == newName {
		// 名字未变,直接提交并返回最新值
		if err := tx.Commit(); err != nil {
			log.complete(err)
			return nil, fmt.Errorf("commit tx: %w", err)
		}
		log.complete(nil, zap.String("result", "no_op"))
		return r.GetByID(ctx, id, userID)
	}

	// 检查新名是否已存在
	var conflictID int64
	err = tx.QueryRowContext(ctx, `SELECT id FROM user_tags WHERE user_id = ? AND name = ?`, userID, newName).Scan(&conflictID)
	if err != nil && err != sql.ErrNoRows {
		log.complete(err)
		return nil, fmt.Errorf("check name conflict: %w", err)
	}
	if err == nil {
		log.complete(nil, zap.Bool("conflict", true), zap.Int64("conflict_id", conflictID))
		return nil, ErrTagNameTaken
	}

	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := tx.ExecContext(ctx,
		`UPDATE user_tags SET name = ?, updated_at = ? WHERE id = ? AND user_id = ?`,
		newName, now, id, userID,
	); err != nil {
		log.complete(err)
		return nil, fmt.Errorf("update tag name: %w", err)
	}

	// 同步该用户所有任务 tags JSON 中的旧名 -> 新名
	syncResult, err := tx.ExecContext(ctx, `
		UPDATE tasks
		SET tags = COALESCE(
		    (SELECT json_group_array(CASE WHEN value = ? THEN ? ELSE value END) FROM json_each(tasks.tags)),
		    '[]'
		),
		updated_at = ?
		WHERE user_id = ?
		  AND EXISTS (SELECT 1 FROM json_each(tasks.tags) WHERE value = ?)`,
		oldName, newName, now, userID, oldName,
	)
	if err != nil {
		log.complete(err)
		return nil, fmt.Errorf("sync task tags on rename: %w", err)
	}

	if err := tx.Commit(); err != nil {
		log.complete(err)
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	log.complete(nil,
		zap.Int64("tasks_updated", rowsAffected(syncResult)),
		zap.String("old_name", oldName),
	)
	return r.GetByID(ctx, id, userID)
}

// DeleteWithTaskSync 在单事务内删除标签,并把该用户所有任务 tags JSON 中匹配的名字摘除。
func (r *TagRepo) DeleteWithTaskSync(ctx context.Context, id, userID int64) (bool, int64, error) {
	log := beginDBOperation(ctx, tagRepositoryName, "delete_tag_with_task_sync",
		zap.Int64("tag_id", id),
		zap.Int64("user_id", userID),
	)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.complete(err)
		return false, 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var name string
	err = tx.QueryRowContext(ctx, `SELECT name FROM user_tags WHERE id = ? AND user_id = ?`, id, userID).Scan(&name)
	if err == sql.ErrNoRows {
		log.complete(nil, zap.Bool("found", false))
		return false, 0, nil
	}
	if err != nil {
		log.complete(err)
		return false, 0, fmt.Errorf("query tag name: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM user_tags WHERE id = ? AND user_id = ?`, id, userID); err != nil {
		log.complete(err)
		return false, 0, fmt.Errorf("delete tag: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	syncResult, err := tx.ExecContext(ctx, `
		UPDATE tasks
		SET tags = COALESCE(
		    (SELECT json_group_array(value) FROM json_each(tasks.tags) WHERE value != ?),
		    '[]'
		),
		updated_at = ?
		WHERE user_id = ?
		  AND EXISTS (SELECT 1 FROM json_each(tasks.tags) WHERE value = ?)`,
		name, now, userID, name,
	)
	if err != nil {
		log.complete(err)
		return false, 0, fmt.Errorf("sync task tags on delete: %w", err)
	}

	if err := tx.Commit(); err != nil {
		log.complete(err)
		return false, 0, fmt.Errorf("commit tx: %w", err)
	}

	tasksUpdated := rowsAffected(syncResult)
	log.complete(nil,
		zap.String("name", name),
		zap.Int64("tasks_updated", tasksUpdated),
	)
	return true, tasksUpdated, nil
}

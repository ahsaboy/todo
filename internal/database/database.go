package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT,
    password_hash TEXT NOT NULL,
    is_admin INTEGER DEFAULT 0,
    created_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE IF NOT EXISTS user_api_keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    key_hash TEXT NOT NULL UNIQUE,
    name TEXT DEFAULT 'default',
    last_used_at TEXT,
    created_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS user_reminder_configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL DEFAULT 'default',
    channel_type TEXT NOT NULL DEFAULT 'webhook'
        CHECK(channel_type IN ('webhook','feishu','dingtalk','wecom','slack')),
    webhook_url TEXT NOT NULL,
    webhook_method TEXT DEFAULT 'POST',
    webhook_headers TEXT,
    webhook_body_template TEXT,
    max_retries INTEGER DEFAULT 3,
    retry_delay_seconds INTEGER DEFAULT 5,
    enabled INTEGER DEFAULT 1,
    created_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    title TEXT NOT NULL,
    description TEXT DEFAULT '',
    completed INTEGER DEFAULT 0,
    priority INTEGER DEFAULT 3 CHECK(priority IN (1, 2, 3)),
    due_at TEXT,
    remind_at TEXT,
    repeat_type TEXT DEFAULT 'none' CHECK(repeat_type IN ('none','daily','weekly','monthly','yearly')),
    repeat_interval INTEGER DEFAULT 1,
    repeat_end_date TEXT,
    reminder_sent INTEGER DEFAULT 0,
    reminder_sent_at TEXT,
    focus_duration INTEGER,
    tags TEXT NOT NULL DEFAULT '[]',
    created_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS user_tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    color TEXT NOT NULL DEFAULT '#3b82f6',
    icon TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    UNIQUE(user_id, name),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS reminder_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    task_id INTEGER NOT NULL,
    reminder_config_id INTEGER,
    channel_name TEXT NOT NULL,
    channel_type TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('success','failed')),
    attempts INTEGER NOT NULL DEFAULT 1,
    error_message TEXT DEFAULT '',
    created_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (reminder_config_id) REFERENCES user_reminder_configs(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS admin_audit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    admin_user_id INTEGER NOT NULL,
    action TEXT NOT NULL,
    target_type TEXT NOT NULL,
    target_id INTEGER,
    detail TEXT DEFAULT '',
    created_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    FOREIGN KEY (admin_user_id) REFERENCES users(id) ON DELETE CASCADE
);
`

const indexes = `
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_user_api_keys_key_hash ON user_api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_user_api_keys_user_id ON user_api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_user_reminder_configs_user_id ON user_reminder_configs(user_id);
CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks(user_id);
CREATE INDEX IF NOT EXISTS idx_tasks_reminder_pending
    ON tasks(remind_at, reminder_sent)
    WHERE remind_at IS NOT NULL AND reminder_sent = 0;
CREATE INDEX IF NOT EXISTS idx_tasks_completed ON tasks(completed);
CREATE INDEX IF NOT EXISTS idx_tasks_due_at ON tasks(due_at);
CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority);
CREATE INDEX IF NOT EXISTS idx_user_tags_user_id ON user_tags(user_id);
CREATE INDEX IF NOT EXISTS idx_user_tags_user_sort ON user_tags(user_id, sort_order, id);
CREATE INDEX IF NOT EXISTS idx_reminder_logs_user_id_created_at ON reminder_logs(user_id, created_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_reminder_logs_task_config
    ON reminder_logs(task_id, reminder_config_id)
    WHERE reminder_config_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_admin_audit_logs_created_at ON admin_audit_logs(created_at DESC);
`

func Init(dbPath string) (*sql.DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(on)&_pragma=journal_size_limit(67108864)")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("exec schema: %w", err)
	}

	// 迁移：为旧版 tasks 表补齐 user_id，并尽量回填已有数据。
	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migrate schema: %w", err)
	}

	if _, err := db.Exec(indexes); err != nil {
		return nil, fmt.Errorf("exec indexes: %w", err)
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin migration: %w", err)
	}
	defer tx.Rollback()

	var hasUserID int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='user_id'`).Scan(&hasUserID); err != nil {
		return fmt.Errorf("check tasks.user_id: %w", err)
	}
	if hasUserID == 0 {
		if _, err := tx.Exec(`ALTER TABLE tasks ADD COLUMN user_id INTEGER REFERENCES users(id) ON DELETE SET NULL`); err != nil {
			return fmt.Errorf("add tasks.user_id: %w", err)
		}
	}

	var legacyOwner sql.NullInt64
	err = tx.QueryRow(`SELECT id FROM users ORDER BY id LIMIT 1`).Scan(&legacyOwner)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("find legacy task owner: %w", err)
	}
	if legacyOwner.Valid {
		if _, err := tx.Exec(`UPDATE tasks SET user_id = ? WHERE user_id IS NULL`, legacyOwner.Int64); err != nil {
			return fmt.Errorf("backfill tasks.user_id: %w", err)
		}
	}

	var hasFocusDuration int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='focus_duration'`).Scan(&hasFocusDuration); err != nil {
		return fmt.Errorf("check tasks.focus_duration: %w", err)
	}
	if hasFocusDuration == 0 {
		if _, err := tx.Exec(`ALTER TABLE tasks ADD COLUMN focus_duration INTEGER`); err != nil {
			return fmt.Errorf("add tasks.focus_duration: %w", err)
		}
	}

	var hasTags int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='tags'`).Scan(&hasTags); err != nil {
		return fmt.Errorf("check tasks.tags: %w", err)
	}
	if hasTags == 0 {
		if _, err := tx.Exec(`ALTER TABLE tasks ADD COLUMN tags TEXT NOT NULL DEFAULT '[]'`); err != nil {
			return fmt.Errorf("add tasks.tags: %w", err)
		}
	}

	var hasIsAdmin int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='is_admin'`).Scan(&hasIsAdmin); err != nil {
		return fmt.Errorf("check users.is_admin: %w", err)
	}
	if hasIsAdmin == 0 {
		if _, err := tx.Exec(`ALTER TABLE users ADD COLUMN is_admin INTEGER DEFAULT 0`); err != nil {
			return fmt.Errorf("add users.is_admin: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration: %w", err)
	}
	return nil
}

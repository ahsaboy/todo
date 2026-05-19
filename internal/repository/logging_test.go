package repository

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	_ "modernc.org/sqlite"

	appLogging "todo/internal/logging"
	"todo/internal/models"
)

func TestAPIKeyRepoValidateKeyLogsSanitizedSummary(t *testing.T) {
	db := openRepositoryTestDB(t)
	ctx, buf := newRepositoryTestContext()

	seedRepositoryTestData(t, db)

	repo := NewAPIKeyRepo(db)
	userID, err := repo.ValidateKey(ctx, "sensitivehashvalue123456")
	if err != nil {
		t.Fatalf("ValidateKey: %v", err)
	}
	if userID != 1 {
		t.Fatalf("ValidateKey userID = %d, want 1", userID)
	}

	entries := decodeRepositoryEntries(t, buf)
	entry := findRepositoryEntry(t, entries, apiKeyRepositoryName, "validate_api_key")
	if got := entry["key_fingerprint"]; got != "sensitiv" {
		t.Fatalf("key_fingerprint = %v, want sensitiv", got)
	}
	if got := entry["valid"]; got != true {
		t.Fatalf("valid = %v, want true", got)
	}
	if _, ok := entry["rows_affected"]; !ok {
		t.Fatalf("expected rows_affected in entry: %#v", entry)
	}
	raw := buf.String()
	if strings.Contains(raw, "sensitivehashvalue123456") {
		t.Fatalf("raw log should not contain full api key hash: %s", raw)
	}
}

func TestTaskRepoCreateLogsOperationSummary(t *testing.T) {
	db := openRepositoryTestDB(t)
	ctx, buf := newRepositoryTestContext()

	seedRepositoryTestData(t, db)

	repo := NewTaskRepo(db)
	task, err := repo.Create(ctx, 1, models.CreateTaskRequest{
		Title:       "ship logs",
		Description: "ensure repository logs exist",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if task == nil || task.ID == 0 {
		t.Fatalf("expected created task, got %#v", task)
	}

	entries := decodeRepositoryEntries(t, buf)
	entry := findRepositoryEntry(t, entries, taskRepositoryName, "create_task")
	if got := entry["user_id"]; got != float64(1) {
		t.Fatalf("user_id = %v, want 1", got)
	}
	if got := entry["title"]; got != "ship logs" {
		t.Fatalf("title = %v, want ship logs", got)
	}
	if got := entry["task_id"]; got != float64(task.ID) {
		t.Fatalf("task_id = %v, want %d", got, task.ID)
	}
	if _, ok := entry["rows_affected"]; !ok {
		t.Fatalf("expected rows_affected in entry: %#v", entry)
	}
	if _, ok := entry["duration"]; !ok {
		t.Fatalf("expected duration in entry: %#v", entry)
	}
}

func TestReminderLogRepoUpsertLogsWithoutErrorMessageBody(t *testing.T) {
	db := openRepositoryTestDB(t)
	ctx, buf := newRepositoryTestContext()

	seedRepositoryTestData(t, db)

	repo := NewReminderLogRepo(db)
	err := repo.Upsert(ctx, CreateReminderLogParams{
		UserID:           1,
		TaskID:           1,
		ReminderConfigID: 1,
		ChannelName:      "ops",
		ChannelType:      "webhook",
		Status:           "failed",
		Attempts:         2,
		ErrorMessage:     "secret downstream payload",
	})
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	entries := decodeRepositoryEntries(t, buf)
	entry := findRepositoryEntry(t, entries, reminderLogRepositoryName, "upsert_reminder_log")
	if got := entry["has_error_message"]; got != true {
		t.Fatalf("has_error_message = %v, want true", got)
	}
	if got := entry["status"]; got != "failed" {
		t.Fatalf("status = %v, want failed", got)
	}
	raw := buf.String()
	if strings.Contains(raw, "secret downstream payload") {
		t.Fatalf("raw log should not contain reminder error message body: %s", raw)
	}
}

func openRepositoryTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	statements := []string{
		`CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			email TEXT,
			password_hash TEXT NOT NULL,
			created_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
			updated_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
		)`,
		`CREATE TABLE tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			title TEXT NOT NULL,
			description TEXT DEFAULT '',
			completed INTEGER DEFAULT 0,
			priority INTEGER DEFAULT 3,
			due_at TEXT,
			remind_at TEXT,
			repeat_type TEXT DEFAULT 'none',
			repeat_interval INTEGER DEFAULT 1,
			repeat_end_date TEXT,
			reminder_sent INTEGER DEFAULT 0,
			reminder_sent_at TEXT,
			created_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
			updated_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
		)`,
		`CREATE TABLE user_api_keys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			key_hash TEXT NOT NULL UNIQUE,
			name TEXT DEFAULT 'default',
			last_used_at TEXT,
			created_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
		)`,
		`CREATE TABLE user_reminder_configs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			name TEXT NOT NULL DEFAULT 'default',
			channel_type TEXT NOT NULL DEFAULT 'webhook',
			webhook_url TEXT NOT NULL,
			webhook_method TEXT DEFAULT 'POST',
			webhook_headers TEXT,
			webhook_body_template TEXT,
			max_retries INTEGER DEFAULT 3,
			retry_delay_seconds INTEGER DEFAULT 5,
			enabled INTEGER DEFAULT 1,
			created_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
			updated_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
		)`,
		`CREATE TABLE reminder_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			task_id INTEGER NOT NULL,
			reminder_config_id INTEGER,
			channel_name TEXT NOT NULL,
			channel_type TEXT NOT NULL,
			status TEXT NOT NULL,
			attempts INTEGER NOT NULL DEFAULT 1,
			error_message TEXT DEFAULT '',
			created_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
		)`,
		`CREATE UNIQUE INDEX reminder_logs_task_config_idx ON reminder_logs(task_id, reminder_config_id) WHERE reminder_config_id IS NOT NULL`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("exec schema statement: %v", err)
		}
	}

	return db
}

func seedRepositoryTestData(t *testing.T, db *sql.DB) {
	t.Helper()

	statements := []string{
		`INSERT INTO users (id, username, email, password_hash, created_at, updated_at) VALUES (1, 'alice', 'alice@example.com', 'hashed-password', '2026-05-19T00:00:00Z', '2026-05-19T00:00:00Z')`,
		`INSERT INTO tasks (id, user_id, title, description, completed, priority, created_at, updated_at) VALUES (1, 1, 'existing task', 'seed', 0, 3, '2026-05-19T00:00:00Z', '2026-05-19T00:00:00Z')`,
		`INSERT INTO user_api_keys (id, user_id, key_hash, name, created_at) VALUES (1, 1, 'sensitivehashvalue123456', 'default', '2026-05-19T00:00:00Z')`,
		`INSERT INTO user_reminder_configs (id, user_id, name, channel_type, webhook_url, webhook_method, enabled, created_at, updated_at) VALUES (1, 1, 'primary', 'webhook', 'https://example.com/hook', 'POST', 1, '2026-05-19T00:00:00Z', '2026-05-19T00:00:00Z')`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("exec seed statement: %v", err)
		}
	}
}

func newRepositoryTestContext() (context.Context, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		MessageKey:     "msg",
		CallerKey:      "caller",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	logger := zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), zapcore.AddSync(buf), zap.DebugLevel))
	return appLogging.WithContext(context.Background(), logger), buf
}

func decodeRepositoryEntries(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	entries := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("unmarshal log entry %q: %v", line, err)
		}
		entries = append(entries, entry)
	}
	return entries
}

func findRepositoryEntry(t *testing.T, entries []map[string]any, repositoryName, operation string) map[string]any {
	t.Helper()

	for _, entry := range entries {
		if entry["repository"] == repositoryName && entry["operation"] == operation {
			return entry
		}
	}
	t.Fatalf("missing log entry for repository=%s operation=%s in %#v", repositoryName, operation, entries)
	return nil
}

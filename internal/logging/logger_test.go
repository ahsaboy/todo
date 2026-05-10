package logging

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"todo/internal/config"
)

func TestBackendLogPath(t *testing.T) {
	got := BackendLogPath("logs", time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC))
	if want := filepath.Join("logs", "backend-2026-05-10.log"); got != want {
		t.Fatalf("BackendLogPath() = %q, want %q", got, want)
	}
}

func TestCleanupOldLogsRemovesExpiredLogsAndKeepsFreshLogs(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	oldTime := now.Add(-9 * 24 * time.Hour)
	freshTime := now.Add(-2 * 24 * time.Hour)

	oldBackend := filepath.Join(dir, "backend-2026-04-30.log")
	freshBackend := filepath.Join(dir, "backend-2026-05-09.log")
	oldFrontend := filepath.Join(dir, "frontend-2026-04-29.log")
	freshFrontend := filepath.Join(dir, "frontend-2026-05-10.log")
	otherFile := filepath.Join(dir, "note.txt")

	for _, path := range []string{oldBackend, freshBackend, oldFrontend, freshFrontend, otherFile} {
		if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	for _, path := range []string{oldBackend, oldFrontend, freshBackend, freshFrontend, otherFile} {
		if err := os.Chtimes(path, freshTime, freshTime); err != nil {
			t.Fatalf("chtimes %s: %v", path, err)
		}
	}

	_ = oldTime

	if err := CleanupOldLogs(dir, 7, now); err != nil {
		t.Fatalf("CleanupOldLogs: %v", err)
	}

	for _, path := range []string{oldBackend, oldFrontend} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("%s should be removed, stat err = %v", path, err)
		}
	}
	for _, path := range []string{freshBackend, freshFrontend, otherFile} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("%s should remain, stat err = %v", path, err)
		}
	}
}

func TestCleanupOldLogsDoesNotRemoveNonLogFiles(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)

	nonLog := filepath.Join(dir, "backend-not-a-log.txt")
	if err := os.WriteFile(nonLog, []byte("test"), 0o644); err != nil {
		t.Fatalf("write non-log: %v", err)
	}
	if err := CleanupOldLogs(dir, 7, now); err != nil {
		t.Fatalf("CleanupOldLogs: %v", err)
	}

	if _, err := os.Stat(nonLog); err != nil {
		t.Fatalf("non-log file should remain, stat err = %v", err)
	}
}

func TestNewLoggerWritesBackendFile(t *testing.T) {
	dir := t.TempDir()
	cfg := defaultConfig()
	cfg.Path = dir
	cfg.Backend.ConsoleEnabled = false
	cfg.Backend.FileEnabled = true

	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	defer logger.Sync()

	logger.Info("hello world")
	_ = logger.Sync()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	found := false
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "backend-") && strings.HasSuffix(entry.Name(), ".log") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected backend log file in %s", dir)
	}
}

func TestNewLoggerDoesNotWriteToStdoutWhenConsoleDisabled(t *testing.T) {
	dir := t.TempDir()
	cfg := defaultConfig()
	cfg.Path = dir
	cfg.Backend.ConsoleEnabled = false
	cfg.Backend.FileEnabled = true

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	logger.Info("stdout should stay empty")
	_ = logger.Sync()
	_ = w.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("ReadFrom: %v", err)
	}
	_ = r.Close()

	if strings.TrimSpace(buf.String()) != "" {
		t.Fatalf("expected stdout to stay empty, got %q", buf.String())
	}
}

func defaultConfig() config.LoggingConfig {
	return config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Path:   "./logs",
		Backend: config.LoggingOutputConfig{
			ConsoleEnabled: true,
			FileEnabled:    false,
		},
		Frontend: config.FrontendLoggingConfig{
			ConsoleEnabled: false,
			FileEnabled:    false,
			Level:          "warn",
		},
	}
}

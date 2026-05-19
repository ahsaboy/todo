package logging

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

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
	otherFile := filepath.Join(dir, "note.txt")

	for _, path := range []string{oldBackend, freshBackend, otherFile} {
		if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	for _, path := range []string{oldBackend, freshBackend, otherFile} {
		if err := os.Chtimes(path, freshTime, freshTime); err != nil {
			t.Fatalf("chtimes %s: %v", path, err)
		}
	}

	_ = oldTime

	if err := CleanupOldLogs(dir, 7, now); err != nil {
		t.Fatalf("CleanupOldLogs: %v", err)
	}

	for _, path := range []string{oldBackend} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("%s should be removed, stat err = %v", path, err)
		}
	}
	for _, path := range []string{freshBackend, otherFile} {
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
	cfg.FileEnabled = true

	logger, syncer, err := NewManagedLogger(cfg)
	if err != nil {
		t.Fatalf("NewManagedLogger: %v", err)
	}
	defer syncer.Sync()

	logger.Info("hello world")
	_ = syncer.Sync()

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

func TestAccessLoggerKeepsDetailedFieldsForAPIRequests(t *testing.T) {
	t.Setenv("GIN_MODE", gin.TestMode)
	gin.SetMode(gin.TestMode)

	logger, buf := newBufferedLogger()
	router := gin.New()
	router.Use(AccessLogger(logger))
	router.GET("/api/v1/tasks", func(c *gin.Context) {
		c.Set("user_id", 42)
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks?status=open", nil)
	req.Header.Set("User-Agent", "logger-test")
	req.Header.Set("Referer", "https://example.com/tasks")
	req.RemoteAddr = "203.0.113.7:4567"
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	entry, raw := decodeSingleLogEntry(t, buf)
	for _, key := range []string{"method", "path", "query", "status", "latency", "client_ip", "user_agent", "referer", "body_bytes", "response_bytes", "route", "user_id"} {
		if _, ok := entry[key]; !ok {
			t.Fatalf("expected API access log to include %q: %s", key, raw)
		}
	}
	assertSingleKeyOccurrence(t, raw, "method")
	assertSingleKeyOccurrence(t, raw, "path")
}

func TestAccessLoggerUsesCompactFieldsForStaticAssets(t *testing.T) {
	t.Setenv("GIN_MODE", gin.TestMode)
	gin.SetMode(gin.TestMode)

	logger, buf := newBufferedLogger()
	router := gin.New()
	router.Use(AccessLogger(logger))
	router.GET("/assets/*filepath", func(c *gin.Context) {
		c.String(http.StatusOK, "asset")
	})

	req := httptest.NewRequest(http.MethodGet, "/assets/app.js?v=123", nil)
	req.Header.Set("User-Agent", "logger-test")
	req.Header.Set("Referer", "https://example.com")
	req.RemoteAddr = "203.0.113.8:5678"
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	entry, raw := decodeSingleLogEntry(t, buf)
	for _, key := range []string{"method", "path", "status", "latency", "response_bytes"} {
		if _, ok := entry[key]; !ok {
			t.Fatalf("expected static access log to include %q: %s", key, raw)
		}
	}
	for _, key := range []string{"query", "client_ip", "user_agent", "referer", "body_bytes", "route", "user_id"} {
		if _, ok := entry[key]; ok {
			t.Fatalf("expected static access log to omit %q: %s", key, raw)
		}
	}
	assertSingleKeyOccurrence(t, raw, "method")
	assertSingleKeyOccurrence(t, raw, "path")
}

func TestIsStaticAssetPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{path: "/assets/app.js", want: true},
		{path: "/favicon.svg", want: true},
		{path: "/icons.svg", want: true},
		{path: "/styles/site.css", want: true},
		{path: "/api/v1/tasks", want: false},
		{path: "/mcp", want: false},
	}

	for _, tt := range tests {
		if got := isStaticAssetPath(tt.path); got != tt.want {
			t.Fatalf("isStaticAssetPath(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func defaultConfig() config.LoggingConfig {
	return config.LoggingConfig{
		Path:        "./logs",
		MaxDays:     7,
		FileEnabled: false,
	}
}

func newBufferedLogger() (*zap.Logger, *bytes.Buffer) {
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
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), zapcore.AddSync(buf), zap.NewAtomicLevelAt(zap.InfoLevel))
	return zap.New(core), buf
}

func decodeSingleLogEntry(t *testing.T, buf *bytes.Buffer) (map[string]any, string) {
	t.Helper()

	raw := strings.TrimSpace(buf.String())
	if raw == "" {
		t.Fatal("expected access log output")
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(raw), &entry); err != nil {
		t.Fatalf("decode log entry: %v\nraw: %s", err, raw)
	}
	return entry, raw
}

func assertSingleKeyOccurrence(t *testing.T, raw string, key string) {
	t.Helper()

	token := `"` + key + `"`
	if count := strings.Count(raw, token); count != 1 {
		t.Fatalf("expected %q to appear once, got %d in %s", key, count, raw)
	}
}

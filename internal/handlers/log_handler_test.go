package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"todo/internal/config"
	"todo/internal/logging"
)

func TestRuntimeConfigOmitsSensitiveFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewLogHandler(config.LoggingConfig{
		Frontend: config.FrontendLoggingConfig{
			ConsoleEnabled: false,
			FileEnabled:    true,
			Level:          "warn",
		},
		Path:    "./logs",
		MaxDays: 7,
		Backend: config.LoggingOutputConfig{
			ConsoleEnabled: true,
			FileEnabled:    true,
		},
	}, nil)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/runtime-config", nil)

	handler.RuntimeConfig(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	loggingObj, ok := payload["logging"].(map[string]any)
	if !ok {
		t.Fatalf("missing logging object: %v", payload)
	}
	if _, ok := loggingObj["path"]; ok {
		t.Fatalf("runtime config must not include logging.path")
	}
	if _, ok := loggingObj["max_days"]; ok {
		t.Fatalf("runtime config must not include logging.max_days")
	}
	if _, ok := loggingObj["backend"]; ok {
		t.Fatalf("runtime config must not include backend config")
	}

	frontend, ok := loggingObj["frontend"].(map[string]any)
	if !ok {
		t.Fatalf("missing frontend config: %v", loggingObj)
	}
	if got := frontend["console_enabled"]; got != false {
		t.Fatalf("console_enabled = %v, want false", got)
	}
	if got := frontend["file_enabled"]; got != true {
		t.Fatalf("file_enabled = %v, want true", got)
	}
	if got := frontend["level"]; got != "warn" {
		t.Fatalf("level = %v, want warn", got)
	}
}

func TestFrontendLogsWriteFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dir := t.TempDir()
	handler := NewLogHandler(config.LoggingConfig{
		Path: dir,
		Frontend: config.FrontendLoggingConfig{
			FileEnabled: true,
		},
	}, nil)

	body := `{"level":"info","message":"hello","context":{"endpoint":"/api"}}`
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/logs/frontend", strings.NewReader(body))

	handler.FrontendLogs(ctx)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}

	path := logging.FrontendLogPath(dir, time.Now())
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if !strings.Contains(string(data), `"message":"hello"`) {
		t.Fatalf("log file does not contain entry: %s", string(data))
	}
	if !strings.Contains(string(data), `"endpoint":"/api"`) {
		t.Fatalf("log file does not contain context: %s", string(data))
	}
}

func TestFrontendLogsRejectsTooLongFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewLogHandler(config.LoggingConfig{
		Frontend: config.FrontendLoggingConfig{
			FileEnabled: true,
		},
	}, nil)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	tooLong := strings.Repeat("a", maxFrontendLogMessageBytes+1)
	body := `{"level":"info","message":"` + tooLong + `"}`
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/logs/frontend", strings.NewReader(body))

	handler.FrontendLogs(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestFrontendLogsDisabledDoesNotCreateFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dir := t.TempDir()
	handler := NewLogHandler(config.LoggingConfig{
		Path: dir,
		Frontend: config.FrontendLoggingConfig{
			FileEnabled: false,
		},
	}, nil)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/logs/frontend", strings.NewReader(`{"level":"info","message":"hello"}`))

	handler.FrontendLogs(ctx)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}

	if entries, err := os.ReadDir(dir); err != nil {
		t.Fatalf("ReadDir: %v", err)
	} else if len(entries) != 0 {
		t.Fatalf("expected no files in %s, got %d", dir, len(entries))
	}
}

func TestWriteFrontendLogEntriesUsesDailyFileName(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 5, 10, 9, 0, 0, 0, time.UTC)

	err := writeFrontendLogEntries(config.LoggingConfig{
		Path: dir,
	}, []FrontendLogEntry{{Level: "info", Message: "hello"}}, now)
	if err != nil {
		t.Fatalf("writeFrontendLogEntries: %v", err)
	}

	want := filepath.Join(dir, "frontend-2026-05-10.log")
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("expected %s to exist: %v", want, err)
	}
}

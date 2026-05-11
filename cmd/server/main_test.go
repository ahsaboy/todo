package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"todo/internal/config"
)

func TestApplyLogOutputMode(t *testing.T) {
	tests := []struct {
		name    string
		mode    string
		console bool
		file    bool
		wantErr bool
	}{
		{name: "console", mode: "console", console: true, file: false},
		{name: "file", mode: "file", console: false, file: true},
		{name: "both", mode: "both", console: true, file: true},
		{name: "off", mode: "off", console: false, file: false},
		{name: "invalid", mode: "invalid", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			console, file, err := applyLogOutputMode(tt.mode)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				if !strings.Contains(err.Error(), "console, file, both, off") {
					t.Fatalf("error = %v, want options hint", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("applyLogOutputMode: %v", err)
			}
			if console != tt.console || file != tt.file {
				t.Fatalf("got console=%v file=%v, want console=%v file=%v", console, file, tt.console, tt.file)
			}
		})
	}
}

func TestApplyLoggingOverrides(t *testing.T) {
	cfg := &config.Config{
		Logging: config.LoggingConfig{
			Path:    "./logs",
			MaxDays: 7,
			Backend: config.LoggingOutputConfig{
				ConsoleEnabled: true,
				FileEnabled:    false,
			},
			Frontend: config.FrontendLoggingConfig{
				ConsoleEnabled: false,
				FileEnabled:    false,
				Level:          "warn",
			},
		},
	}

	if err := applyLoggingOverrides(cfg, "./tmp-logs", 3, "both", "file", "debug"); err != nil {
		t.Fatalf("applyLoggingOverrides: %v", err)
	}

	if got, want := cfg.Logging.Path, "./tmp-logs"; got != want {
		t.Fatalf("logging path = %q, want %q", got, want)
	}
	if got, want := cfg.Logging.MaxDays, 3; got != want {
		t.Fatalf("logging max days = %d, want %d", got, want)
	}
	if !cfg.Logging.Backend.ConsoleEnabled || !cfg.Logging.Backend.FileEnabled {
		t.Fatalf("backend flags = %+v, want both enabled", cfg.Logging.Backend)
	}
	if cfg.Logging.Frontend.ConsoleEnabled || !cfg.Logging.Frontend.FileEnabled {
		t.Fatalf("frontend flags = %+v, want console disabled file enabled", cfg.Logging.Frontend)
	}
	if got, want := cfg.Logging.Frontend.Level, "debug"; got != want {
		t.Fatalf("frontend level = %q, want %q", got, want)
	}
}

func TestApplyLoggingOverridesRejectsInvalidMode(t *testing.T) {
	cfg := &config.Config{}

	err := applyLoggingOverrides(cfg, "", 0, "bad", "", "")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "无效的 --backend-log 值") {
		t.Fatalf("error = %v, want backend-log hint", err)
	}
}

func TestAllowedCORSOrigin(t *testing.T) {
	tests := []struct {
		name           string
		requestOrigin  string
		allowedOrigins []string
		want           string
		wantOK         bool
	}{
		{
			name:           "matches one of multiple origins",
			requestOrigin:  "https://app.example.com",
			allowedOrigins: []string{"https://admin.example.com", "https://app.example.com"},
			want:           "https://app.example.com",
			wantOK:         true,
		},
		{
			name:           "wildcard echoes request origin",
			requestOrigin:  "https://app.example.com",
			allowedOrigins: []string{"*"},
			want:           "https://app.example.com",
			wantOK:         true,
		},
		{
			name:           "wildcard without request origin",
			allowedOrigins: []string{"*"},
			want:           "*",
			wantOK:         true,
		},
		{
			name:           "rejects unlisted origin",
			requestOrigin:  "https://evil.example.com",
			allowedOrigins: []string{"https://app.example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := allowedCORSOrigin(tt.requestOrigin, tt.allowedOrigins)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if got != tt.want {
				t.Fatalf("origin = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCORSMiddlewareUsesMatchingRequestOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(corsMiddleware([]string{"https://admin.example.com", "https://app.example.com"}))
	r.GET("/ping", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "https://app.example.com")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if got, want := rec.Header().Get("Access-Control-Allow-Origin"), "https://app.example.com"; got != want {
		t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, want)
	}
	if got, want := rec.Header().Get("Vary"), "Origin"; got != want {
		t.Fatalf("Vary = %q, want %q", got, want)
	}
}

func TestCORSMiddlewareOmitsUnmatchedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(corsMiddleware([]string{"https://app.example.com"}))
	r.GET("/ping", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want empty", got)
	}
}

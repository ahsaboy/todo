package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"todo/internal/config"
)

func TestApplyLoggingOverrides(t *testing.T) {
	cfg := &config.Config{
		Logging: config.LoggingConfig{
			Path:        "./logs",
			MaxDays:     7,
			FileEnabled: false,
		},
	}

	if err := applyLoggingOverrides(cfg, "./tmp-logs", 3, true, false); err != nil {
		t.Fatalf("applyLoggingOverrides: %v", err)
	}

	if got, want := cfg.Logging.Path, "./tmp-logs"; got != want {
		t.Fatalf("logging path = %q, want %q", got, want)
	}
	if got, want := cfg.Logging.MaxDays, 3; got != want {
		t.Fatalf("logging max days = %d, want %d", got, want)
	}
	if !cfg.Logging.FileEnabled {
		t.Fatalf("file enabled = false, want true")
	}
}

func TestApplyLoggingOverridesRejectsConflictingFlags(t *testing.T) {
	cfg := &config.Config{}

	err := applyLoggingOverrides(cfg, "", 0, true, true)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "--log-file-enabled 与 --log-file-disabled 不能同时使用") {
		t.Fatalf("error = %v, want conflicting flags hint", err)
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

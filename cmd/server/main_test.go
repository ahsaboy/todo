package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

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

func TestApplyStaticFilesOverrides(t *testing.T) {
	cfg := &config.Config{StaticFiles: false}

	if err := applyStaticFilesOverrides(cfg, true, false); err != nil {
		t.Fatalf("applyStaticFilesOverrides: %v", err)
	}
	if !cfg.StaticFiles {
		t.Fatalf("static files = false, want true")
	}
}

func TestApplyStaticFilesOverridesRejectsConflictingFlags(t *testing.T) {
	cfg := &config.Config{}

	err := applyStaticFilesOverrides(cfg, true, true)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "--static-files-enabled 与 --static-files-disabled 不能同时使用") {
		t.Fatalf("error = %v, want conflicting flags hint", err)
	}
}

func TestApplyEnvOverrides(t *testing.T) {
	t.Setenv("HOST", "127.0.0.1")
	t.Setenv("PORT", "9090")
	t.Setenv("STATIC_FILES", "false")
	t.Setenv("CORS", "https://todo.example.com, https://admin.example.com")

	cfg := &config.Config{
		Server:      config.ServerConfig{Host: "0.0.0.0", Port: 8080},
		StaticFiles: true,
		CORS:        config.CORSConfig{Enabled: false},
	}

	if err := applyEnvOverrides(cfg); err != nil {
		t.Fatalf("applyEnvOverrides: %v", err)
	}

	if got, want := cfg.Server.Host, "127.0.0.1"; got != want {
		t.Fatalf("host = %q, want %q", got, want)
	}
	if got, want := cfg.Server.Port, 9090; got != want {
		t.Fatalf("port = %d, want %d", got, want)
	}
	if cfg.StaticFiles {
		t.Fatalf("static files = true, want false")
	}
	if !cfg.CORS.Enabled {
		t.Fatalf("cors enabled = false, want true")
	}
	if got, want := strings.Join(cfg.CORS.AllowedOrigins, ","), "https://todo.example.com,https://admin.example.com"; got != want {
		t.Fatalf("cors origins = %q, want %q", got, want)
	}
}

func TestApplyEnvOverridesSupportsBooleanCORS(t *testing.T) {
	t.Setenv("CORS", "false")

	cfg := &config.Config{
		CORS: config.CORSConfig{
			Enabled:        true,
			AllowedOrigins: []string{"*"},
		},
	}

	if err := applyEnvOverrides(cfg); err != nil {
		t.Fatalf("applyEnvOverrides: %v", err)
	}

	if cfg.CORS.Enabled {
		t.Fatalf("cors enabled = true, want false")
	}
	if cfg.CORS.AllowedOrigins != nil {
		t.Fatalf("cors allowed origins = %v, want nil", cfg.CORS.AllowedOrigins)
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

func TestRegisterOptionalRoutesWithStaticFilesDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	registerOptionalRoutes(r, zap.NewNop(), false)

	docReq := httptest.NewRequest(http.MethodGet, "/docs/index.html", nil)
	docRec := httptest.NewRecorder()
	r.ServeHTTP(docRec, docReq)
	if got, want := docRec.Code, http.StatusNotFound; got != want {
		t.Fatalf("docs status = %d, want %d", got, want)
	}
	if body := docRec.Body.String(); !strings.Contains(body, `"code":"NOT_FOUND"`) {
		t.Fatalf("docs body = %q, want NOT_FOUND payload", body)
	}

	pageReq := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	pageRec := httptest.NewRecorder()
	r.ServeHTTP(pageRec, pageReq)
	if got, want := pageRec.Code, http.StatusNotFound; got != want {
		t.Fatalf("page status = %d, want %d", got, want)
	}
}

func TestRegisterOptionalRoutesWithStaticFilesEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	registerOptionalRoutes(r, zap.NewNop(), true)

	docReq := httptest.NewRequest(http.MethodGet, "/docs/index.html", nil)
	docRec := httptest.NewRecorder()
	r.ServeHTTP(docRec, docReq)
	if got := docRec.Code; got == http.StatusNotFound {
		t.Fatalf("docs status = %d, want registered route", got)
	}

	pageReq := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	pageRec := httptest.NewRecorder()
	r.ServeHTTP(pageRec, pageReq)
	if got, want := pageRec.Code, http.StatusFound; got != want {
		t.Fatalf("page status = %d, want %d", got, want)
	}
	if loc := pageRec.Header().Get("Location"); loc != "/#/dashboard" {
		t.Fatalf("page redirect Location = %q, want %q", loc, "/#/dashboard")
	}
}

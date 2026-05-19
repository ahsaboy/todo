package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAppliesLoggingDefaultsAndNormalization(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	err := os.WriteFile(path, []byte(`
logging:
  file_enabled: true
  path: ""
  max_days: 0
`), 0o600)
	if err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if got, want := cfg.Logging.Path, "./logs"; got != want {
		t.Fatalf("logging path = %q, want %q", got, want)
	}
	if got, want := cfg.Logging.MaxDays, 7; got != want {
		t.Fatalf("logging max days = %d, want %d", got, want)
	}
	if !cfg.Logging.FileEnabled {
		t.Fatalf("file enabled = false, want true")
	}
}

func TestLoadKeepsLoggingDefaultsWhenMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	err := os.WriteFile(path, []byte(`
logging: {}
`), 0o600)
	if err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if got, want := cfg.Logging.Path, "./logs"; got != want {
		t.Fatalf("logging path = %q, want %q", got, want)
	}
	if got, want := cfg.Logging.MaxDays, 7; got != want {
		t.Fatalf("logging max days = %d, want %d", got, want)
	}
	if !cfg.Logging.FileEnabled {
		t.Fatalf("file enabled = false, want true")
	}
}

func TestLoadUsesEmbeddedDefaultWhenConfigFileIsMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing.yaml")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load missing config: %v", err)
	}

	if got, want := cfg.Server.Host, "0.0.0.0"; got != want {
		t.Fatalf("server host = %q, want %q", got, want)
	}
	if got, want := cfg.Server.Port, 8080; got != want {
		t.Fatalf("server port = %d, want %d", got, want)
	}
	if got, want := cfg.Database.Path, "./data/tasks.db"; got != want {
		t.Fatalf("database path = %q, want %q", got, want)
	}
	if !cfg.CORS.Enabled {
		t.Fatalf("cors enabled = false, want true")
	}
	if !cfg.Logging.FileEnabled {
		t.Fatalf("file enabled = false, want true from embedded default")
	}
	if !cfg.StaticFiles {
		t.Fatalf("static files = false, want true from embedded default")
	}
}

func TestLoadPrefersStaticFilesField(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	err := os.WriteFile(path, []byte(`
static_files: false
swagger: true
`), 0o600)
	if err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.StaticFiles {
		t.Fatalf("static files = true, want false")
	}
}

func TestLoadFallsBackToLegacySwaggerField(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	err := os.WriteFile(path, []byte(`
swagger: false
`), 0o600)
	if err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.StaticFiles {
		t.Fatalf("static files = true, want false from legacy swagger field")
	}
}

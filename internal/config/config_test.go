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
  level: "TRACE"
  format: "yaml"
  path: ""
  max_days: 0
  frontend:
    level: ""
`), 0o600)
	if err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if got, want := cfg.Logging.Level, "info"; got != want {
		t.Fatalf("logging level = %q, want %q", got, want)
	}
	if got, want := cfg.Logging.Format, "console"; got != want {
		t.Fatalf("logging format = %q, want %q", got, want)
	}
	if got, want := cfg.Logging.Path, "./logs"; got != want {
		t.Fatalf("logging path = %q, want %q", got, want)
	}
	if got, want := cfg.Logging.MaxDays, 7; got != want {
		t.Fatalf("logging max days = %d, want %d", got, want)
	}
	if !cfg.Logging.Backend.ConsoleEnabled {
		t.Fatalf("backend console enabled = false, want true")
	}
	if cfg.Logging.Backend.FileEnabled {
		t.Fatalf("backend file enabled = true, want false")
	}
	if cfg.Logging.Frontend.ConsoleEnabled {
		t.Fatalf("frontend console enabled = true, want false")
	}
	if cfg.Logging.Frontend.FileEnabled {
		t.Fatalf("frontend file enabled = true, want false")
	}
	if got, want := cfg.Logging.Frontend.Level, "info"; got != want {
		t.Fatalf("frontend level = %q, want %q", got, want)
	}
}

func TestLoadKeepsDefaultFrontendLevelWhenMissing(t *testing.T) {
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

	if got, want := cfg.Logging.Level, "info"; got != want {
		t.Fatalf("logging level = %q, want %q", got, want)
	}
	if got, want := cfg.Logging.Format, "json"; got != want {
		t.Fatalf("logging format = %q, want %q", got, want)
	}
	if got, want := cfg.Logging.Path, "./logs"; got != want {
		t.Fatalf("logging path = %q, want %q", got, want)
	}
	if got, want := cfg.Logging.MaxDays, 7; got != want {
		t.Fatalf("logging max days = %d, want %d", got, want)
	}
	if got, want := cfg.Logging.Frontend.Level, "warn"; got != want {
		t.Fatalf("frontend level = %q, want %q", got, want)
	}
}

func TestLoadNormalizesInvalidFrontendLevel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	err := os.WriteFile(path, []byte(`
logging:
  level: "debug"
  frontend:
    level: "verbose"
`), 0o600)
	if err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if got, want := cfg.Logging.Level, "debug"; got != want {
		t.Fatalf("logging level = %q, want %q", got, want)
	}
	if got, want := cfg.Logging.Frontend.Level, "info"; got != want {
		t.Fatalf("frontend level = %q, want %q", got, want)
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
	if got, want := cfg.Logging.Frontend.Level, "warn"; got != want {
		t.Fatalf("frontend level = %q, want %q", got, want)
	}
}

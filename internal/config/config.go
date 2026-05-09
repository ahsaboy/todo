package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Auth     AuthConfig     `yaml:"auth"`
	Reminder ReminderConfig `yaml:"reminder"`
	CORS     CORSConfig     `yaml:"cors"`
	Logging  LoggingConfig  `yaml:"logging"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type AuthConfig struct {
	APIKey string `yaml:"api_key"`
}

type ReminderConfig struct {
	Enabled              bool              `yaml:"enabled"`
	ScanIntervalSeconds  int               `yaml:"scan_interval_seconds"`
	WebhookURL           string            `yaml:"webhook_url"`
	WebhookMethod        string            `yaml:"webhook_method"`
	WebhookHeaders       map[string]string `yaml:"webhook_headers"`
	WebhookBodyTemplate  string            `yaml:"webhook_body_template"`
	WebhookTimeoutSeconds int              `yaml:"webhook_timeout_seconds"`
	MaxRetries           int               `yaml:"max_retries"`
	RetryDelaySeconds    int               `yaml:"retry_delay_seconds"`
}

type CORSConfig struct {
	Enabled        bool     `yaml:"enabled"`
	AllowedOrigins []string `yaml:"allowed_origins"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
			Mode: "debug",
		},
		Database: DatabaseConfig{
			Path: "./data/tasks.db",
		},
		Reminder: ReminderConfig{
			ScanIntervalSeconds:   30,
			WebhookMethod:         "POST",
			WebhookBodyTemplate:   `{"task_id":{{.TaskID}},"title":"{{.Title}}","priority":"{{.PriorityText}}","due_at":"{{.DueAt}}"}`,
			WebhookTimeoutSeconds: 10,
			MaxRetries:            3,
			RetryDelaySeconds:     5,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

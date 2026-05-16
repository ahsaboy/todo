package config

import (
	_ "embed"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed default_config.yaml
var defaultConfigYAML []byte

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Reminder ReminderConfig `yaml:"reminder"`
	CORS     CORSConfig     `yaml:"cors"`
	Logging  LoggingConfig  `yaml:"logging"`
	Swagger  bool           `yaml:"swagger"`
}

type ServerConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Mode     string `yaml:"mode"`
	Timezone string `yaml:"timezone"` // 时间出参的目标时区。空 / "Local" 表示服务器本地;"UTC";IANA 名(Asia/Shanghai);固定偏移(+08:00)
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type ReminderConfig struct {
	Enabled               bool                       `yaml:"enabled"`
	ScanIntervalSeconds   int                        `yaml:"scan_interval_seconds"`
	WebhookBodyTemplate   string                     `yaml:"webhook_body_template"`
	WebhookTimeoutSeconds int                        `yaml:"webhook_timeout_seconds"`
	MaxRetries            int                        `yaml:"max_retries"`
	RetryDelaySeconds     int                        `yaml:"retry_delay_seconds"`
	DefaultTemplates      map[string]DefaultTemplate `yaml:"default_templates"`
}

type DefaultTemplate struct {
	ChannelType         string            `yaml:"channel_type" json:"channel_type"`
	WebhookURL          string            `yaml:"webhook_url" json:"webhook_url"`
	WebhookMethod       string            `yaml:"webhook_method" json:"webhook_method"`
	WebhookHeaders      map[string]string `yaml:"webhook_headers" json:"webhook_headers"`
	WebhookBodyTemplate string            `yaml:"webhook_body_template" json:"webhook_body_template"`
}

type CORSConfig struct {
	Enabled        bool     `yaml:"enabled"`
	AllowedOrigins []string `yaml:"allowed_origins"`
}

type LoggingConfig struct {
	Level    string                `yaml:"level"`
	Format   string                `yaml:"format"`
	Path     string                `yaml:"path"`
	MaxDays  int                   `yaml:"max_days"`
	Backend  LoggingOutputConfig   `yaml:"backend"`
	Frontend FrontendLoggingConfig `yaml:"frontend"`
}

type LoggingOutputConfig struct {
	ConsoleEnabled bool `yaml:"console_enabled"`
	FileEnabled    bool `yaml:"file_enabled"`
}

type FrontendLoggingConfig struct {
	ConsoleEnabled bool   `yaml:"console_enabled"`
	FileEnabled    bool   `yaml:"file_enabled"`
	Level          string `yaml:"level"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		data = defaultConfigYAML
	}

	cfg := &Config{
		Server: ServerConfig{
			Host:     "0.0.0.0",
			Port:     8080,
			Mode:     "debug",
			Timezone: "",
		},
		Database: DatabaseConfig{
			Path: "./data/tasks.db",
		},
		Reminder: ReminderConfig{
			ScanIntervalSeconds:   30,
			WebhookBodyTemplate:   `{"task_id":{{.TaskID}},"title":"{{.Title}}","priority":"{{.PriorityText}}","due_at":"{{.DueAt}}"}`,
			WebhookTimeoutSeconds: 10,
			MaxRetries:            3,
			RetryDelaySeconds:     5,
		},
		Logging: LoggingConfig{
			Level:   "info",
			Format:  "json",
			Path:    "./logs",
			MaxDays: 7,
			Backend: LoggingOutputConfig{
				ConsoleEnabled: true,
				FileEnabled:    false,
			},
			Frontend: FrontendLoggingConfig{
				ConsoleEnabled: false,
				FileEnabled:    false,
				Level:          "warn",
			},
		},
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	normalizeLoggingConfig(&cfg.Logging)

	return cfg, nil
}

func normalizeLoggingConfig(cfg *LoggingConfig) {
	if cfg.Path == "" {
		cfg.Path = "./logs"
	}

	if cfg.MaxDays < 1 {
		cfg.MaxDays = 7
	}

	cfg.Level = normalizeLogLevel(cfg.Level)
	cfg.Format = normalizeLogFormat(cfg.Format)

	if cfg.Frontend.Level == "" {
		cfg.Frontend.Level = cfg.Level
	} else {
		cfg.Frontend.Level = normalizeLogLevel(cfg.Frontend.Level)
	}
}

func normalizeLogLevel(level string) string {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug", "info", "warn", "error":
		return strings.ToLower(strings.TrimSpace(level))
	default:
		return "info"
	}
}

func normalizeLogFormat(format string) string {
	if strings.EqualFold(strings.TrimSpace(format), "json") {
		return "json"
	}

	return "console"
}

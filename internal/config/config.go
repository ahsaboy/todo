package config

import (
	_ "embed"
	"os"

	"gopkg.in/yaml.v3"
)

//go:embed default_config.yaml
var defaultConfigYAML []byte

type Config struct {
	Server      ServerConfig    `yaml:"server"`
	Database    DatabaseConfig  `yaml:"database"`
	Reminder    ReminderConfig  `yaml:"reminder"`
	CORS        CORSConfig      `yaml:"cors"`
	Logging     LoggingConfig   `yaml:"logging"`
	RateLimit   RateLimitConfig `yaml:"rate_limit"`
	Admin       AdminConfig     `yaml:"admin"`
	StaticFiles bool            `yaml:"static_files"`
}

type AdminConfig struct {
	Enabled   bool   `yaml:"enabled"`
	TokenHash string `yaml:"token_hash"` // Deprecated: 管理后台已改为用户名密码认证
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	Email     string `yaml:"email"`
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
	WorkerCount           int                        `yaml:"worker_count"`
	GracePeriodMinutes    int                        `yaml:"grace_period_minutes"`
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

type RateLimitConfig struct {
	Enabled           bool    `yaml:"enabled"`
	ReqsPerSecond     float64 `yaml:"reqs_per_second"`
	Burst             int     `yaml:"burst"`
	AuthReqsPerSecond float64 `yaml:"auth_reqs_per_second"` // 认证端点单独限制（更严格）
	AuthBurst         int     `yaml:"auth_burst"`
}

type LoggingConfig struct {
	FileEnabled bool   `yaml:"file_enabled"`
	Path        string `yaml:"path"`
	MaxDays     int    `yaml:"max_days"`
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
			WorkerCount:           5,
			GracePeriodMinutes:    10,
		},
		Logging: LoggingConfig{
			FileEnabled: true,
			Path:        "./logs",
			MaxDays:     7,
		},
		Admin: AdminConfig{
			Enabled: false,
		},
		StaticFiles: true,
	}

	var raw struct {
		StaticFiles *bool `yaml:"static_files"`
		Swagger     *bool `yaml:"swagger"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if raw.StaticFiles == nil && raw.Swagger != nil {
		cfg.StaticFiles = *raw.Swagger
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
}

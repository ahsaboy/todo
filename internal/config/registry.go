package config

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// FieldType 描述配置项的值类型,供前端选择渲染控件、后端做类型校验。
type FieldType string

const (
	TypeString FieldType = "string"
	TypeInt    FieldType = "int"
	TypeFloat  FieldType = "float"
	TypeBool   FieldType = "bool"
	TypeEnum   FieldType = "enum"
)

// FieldSpec 是单个配置项的元数据,作为「DB→cfg 覆盖」「CLI 锁定」「GET 输出」的单一真相源。
// 新增可管理配置项 = 往 Registry 追加一条;前端从 GET /admin/api/config 拿结构动态渲染,不硬编码字段表。
type FieldSpec struct {
	Key       string // 点分路径,同 DB key,如 "reminder.max_retries"
	Group     string // 分组: server/reminder/cors/logging/rate_limit/i18n/static/database/admin
	Label     string
	Type      FieldType
	Enum      []string                  // Type==TypeEnum 时的可选值
	Editable  bool                      // 引导类配置为 false(只读,仍由文件+CLI 管理)
	HotReload bool                      // true 表示保存后即时生效,无需重启
	Get       func(c *Config) any       // 读取 cfg 当前值
	Set       func(c *Config, v any) error // 类型转换 + 写回 cfg 字段(引导类用 readOnlySet)
}

// Registry 的顺序决定前端分组卡片的展示顺序(同 group 的字段需连续)。
var Registry = []FieldSpec{
	// —— server ——
	{Key: "server.host", Group: "server", Label: "监听地址", Type: TypeString, Editable: false, Get: func(c *Config) any { return c.Server.Host }, Set: readOnlySet},
	{Key: "server.port", Group: "server", Label: "监听端口", Type: TypeInt, Editable: false, Get: func(c *Config) any { return c.Server.Port }, Set: readOnlySet},
	{Key: "server.mode", Group: "server", Label: "运行模式", Type: TypeEnum, Enum: []string{"debug", "release"}, Editable: false, Get: func(c *Config) any { return c.Server.Mode }, Set: readOnlySet},
	{Key: "server.timezone", Group: "server", Label: "时区", Type: TypeString, Editable: true, HotReload: true, Get: func(c *Config) any { return c.Server.Timezone }, Set: func(c *Config, v any) error {
		s, err := toString(v)
		if err != nil {
			return err
		}
		c.Server.Timezone = s
		return nil
	}},

	// —— i18n ——
	{Key: "i18n.default_lang", Group: "i18n", Label: "默认语言", Type: TypeEnum, Enum: []string{"zh-CN", "en"}, Editable: true, HotReload: true, Get: func(c *Config) any { return c.I18n.DefaultLang }, Set: func(c *Config, v any) error {
		s, err := toString(v)
		if err != nil {
			return err
		}
		c.I18n.DefaultLang = s
		return nil
	}},

	// —— reminder ——
	{Key: "reminder.enabled", Group: "reminder", Label: "启用提醒", Type: TypeBool, Editable: true, HotReload: true, Get: func(c *Config) any { return c.Reminder.Enabled }, Set: func(c *Config, v any) error {
		b, err := toBool(v)
		if err != nil {
			return err
		}
		c.Reminder.Enabled = b
		return nil
	}},
	{Key: "reminder.scan_interval_seconds", Group: "reminder", Label: "扫描间隔(秒)", Type: TypeInt, Editable: true, Get: func(c *Config) any { return c.Reminder.ScanIntervalSeconds }, Set: func(c *Config, v any) error {
		n, err := toInt(v)
		if err != nil {
			return err
		}
		c.Reminder.ScanIntervalSeconds = n
		return nil
	}},
	{Key: "reminder.webhook_timeout_seconds", Group: "reminder", Label: "Webhook 超时(秒)", Type: TypeInt, Editable: true, Get: func(c *Config) any { return c.Reminder.WebhookTimeoutSeconds }, Set: func(c *Config, v any) error {
		n, err := toInt(v)
		if err != nil {
			return err
		}
		c.Reminder.WebhookTimeoutSeconds = n
		return nil
	}},
	{Key: "reminder.max_retries", Group: "reminder", Label: "最大重试次数", Type: TypeInt, Editable: true, Get: func(c *Config) any { return c.Reminder.MaxRetries }, Set: func(c *Config, v any) error {
		n, err := toInt(v)
		if err != nil {
			return err
		}
		c.Reminder.MaxRetries = n
		return nil
	}},
	{Key: "reminder.retry_delay_seconds", Group: "reminder", Label: "重试间隔(秒)", Type: TypeInt, Editable: true, Get: func(c *Config) any { return c.Reminder.RetryDelaySeconds }, Set: func(c *Config, v any) error {
		n, err := toInt(v)
		if err != nil {
			return err
		}
		c.Reminder.RetryDelaySeconds = n
		return nil
	}},
	{Key: "reminder.worker_count", Group: "reminder", Label: "工作线程数", Type: TypeInt, Editable: true, HotReload: true, Get: func(c *Config) any { return c.Reminder.WorkerCount }, Set: func(c *Config, v any) error {
		n, err := toInt(v)
		if err != nil {
			return err
		}
		c.Reminder.WorkerCount = n
		return nil
	}},
	{Key: "reminder.grace_period_minutes", Group: "reminder", Label: "提醒宽限期(分)", Type: TypeInt, Editable: true, HotReload: true, Get: func(c *Config) any { return c.Reminder.GracePeriodMinutes }, Set: func(c *Config, v any) error {
		n, err := toInt(v)
		if err != nil {
			return err
		}
		c.Reminder.GracePeriodMinutes = n
		return nil
	}},
	{Key: "reminder.webhook_body_template", Group: "reminder", Label: "默认 Webhook 消息模板", Type: TypeString, Editable: true, HotReload: true, Get: func(c *Config) any { return c.Reminder.WebhookBodyTemplate }, Set: func(c *Config, v any) error {
		s, err := toString(v)
		if err != nil {
			return err
		}
		c.Reminder.WebhookBodyTemplate = s
		return nil
	}},

	// —— cors ——
	{Key: "cors.enabled", Group: "cors", Label: "启用 CORS", Type: TypeBool, Editable: true, HotReload: true, Get: func(c *Config) any { return c.CORS.Enabled }, Set: func(c *Config, v any) error {
		b, err := toBool(v)
		if err != nil {
			return err
		}
		c.CORS.Enabled = b
		return nil
	}},
	{Key: "cors.allowed_origins", Group: "cors", Label: "允许的来源(逗号分隔)", Type: TypeString, Editable: true, HotReload: true, Get: func(c *Config) any { return strings.Join(c.CORS.AllowedOrigins, ", ") }, Set: func(c *Config, v any) error {
		s, err := toString(v)
		if err != nil {
			return err
		}
		c.CORS.AllowedOrigins = SplitCommaSeparated(s)
		return nil
	}},

	// —— rate_limit ——
	{Key: "rate_limit.enabled", Group: "rate_limit", Label: "启用限流", Type: TypeBool, Editable: true, Get: func(c *Config) any { return c.RateLimit.Enabled }, Set: func(c *Config, v any) error {
		b, err := toBool(v)
		if err != nil {
			return err
		}
		c.RateLimit.Enabled = b
		return nil
	}},
	{Key: "rate_limit.reqs_per_second", Group: "rate_limit", Label: "普通限流速率(req/s)", Type: TypeFloat, Editable: true, Get: func(c *Config) any { return c.RateLimit.ReqsPerSecond }, Set: func(c *Config, v any) error {
		f, err := toFloat(v)
		if err != nil {
			return err
		}
		c.RateLimit.ReqsPerSecond = f
		return nil
	}},
	{Key: "rate_limit.burst", Group: "rate_limit", Label: "普通爆发上限", Type: TypeInt, Editable: true, Get: func(c *Config) any { return c.RateLimit.Burst }, Set: func(c *Config, v any) error {
		n, err := toInt(v)
		if err != nil {
			return err
		}
		c.RateLimit.Burst = n
		return nil
	}},
	{Key: "rate_limit.auth_reqs_per_second", Group: "rate_limit", Label: "认证端点限流速率(req/s)", Type: TypeFloat, Editable: true, Get: func(c *Config) any { return c.RateLimit.AuthReqsPerSecond }, Set: func(c *Config, v any) error {
		f, err := toFloat(v)
		if err != nil {
			return err
		}
		c.RateLimit.AuthReqsPerSecond = f
		return nil
	}},
	{Key: "rate_limit.auth_burst", Group: "rate_limit", Label: "认证端点爆发上限", Type: TypeInt, Editable: true, Get: func(c *Config) any { return c.RateLimit.AuthBurst }, Set: func(c *Config, v any) error {
		n, err := toInt(v)
		if err != nil {
			return err
		}
		c.RateLimit.AuthBurst = n
		return nil
	}},

	// —— logging ——
	{Key: "logging.file_enabled", Group: "logging", Label: "写入日志文件", Type: TypeBool, Editable: true, Get: func(c *Config) any { return c.Logging.FileEnabled }, Set: func(c *Config, v any) error {
		b, err := toBool(v)
		if err != nil {
			return err
		}
		c.Logging.FileEnabled = b
		return nil
	}},
	{Key: "logging.path", Group: "logging", Label: "日志目录", Type: TypeString, Editable: true, Get: func(c *Config) any { return c.Logging.Path }, Set: func(c *Config, v any) error {
		s, err := toString(v)
		if err != nil {
			return err
		}
		c.Logging.Path = s
		return nil
	}},
	{Key: "logging.max_days", Group: "logging", Label: "日志保留天数", Type: TypeInt, Editable: true, Get: func(c *Config) any { return c.Logging.MaxDays }, Set: func(c *Config, v any) error {
		n, err := toInt(v)
		if err != nil {
			return err
		}
		c.Logging.MaxDays = n
		return nil
	}},

	// —— email ——
	{Key: "email.enabled", Group: "email", Label: "启用邮箱服务", Type: TypeBool, Editable: true, HotReload: true, Get: func(c *Config) any { return c.Email.Enabled }, Set: func(c *Config, v any) error {
		b, err := toBool(v)
		if err != nil {
			return err
		}
		c.Email.Enabled = b
		return nil
	}},
	{Key: "email.smtp_host", Group: "email", Label: "SMTP 服务器", Type: TypeString, Editable: true, HotReload: true, Get: func(c *Config) any { return c.Email.SMTPHost }, Set: func(c *Config, v any) error {
		s, err := toString(v)
		if err != nil {
			return err
		}
		c.Email.SMTPHost = s
		return nil
	}},
	{Key: "email.smtp_port", Group: "email", Label: "SMTP 端口", Type: TypeInt, Editable: true, HotReload: true, Get: func(c *Config) any { return c.Email.SMTPPort }, Set: func(c *Config, v any) error {
		n, err := toInt(v)
		if err != nil {
			return err
		}
		if n < 1 || n > 65535 {
			return fmt.Errorf("端口范围 1-65535,得到 %d", n)
		}
		c.Email.SMTPPort = n
		return nil
	}},
	{Key: "email.smtp_username", Group: "email", Label: "SMTP 用户名", Type: TypeString, Editable: true, HotReload: true, Get: func(c *Config) any { return c.Email.SMTPUsername }, Set: func(c *Config, v any) error {
		s, err := toString(v)
		if err != nil {
			return err
		}
		c.Email.SMTPUsername = s
		return nil
	}},
	{Key: "email.smtp_password", Group: "email", Label: "SMTP 密码", Type: TypeString, Editable: true, HotReload: true, Get: func(c *Config) any { return c.Email.SMTPPassword }, Set: func(c *Config, v any) error {
		s, err := toString(v)
		if err != nil {
			return err
		}
		c.Email.SMTPPassword = s
		return nil
	}},
	{Key: "email.from_address", Group: "email", Label: "发件人邮箱", Type: TypeString, Editable: true, HotReload: true, Get: func(c *Config) any { return c.Email.FromAddress }, Set: func(c *Config, v any) error {
		s, err := toString(v)
		if err != nil {
			return err
		}
		c.Email.FromAddress = s
		return nil
	}},
	{Key: "email.from_name", Group: "email", Label: "发件人名称", Type: TypeString, Editable: true, HotReload: true, Get: func(c *Config) any { return c.Email.FromName }, Set: func(c *Config, v any) error {
		s, err := toString(v)
		if err != nil {
			return err
		}
		c.Email.FromName = s
		return nil
	}},

	// —— static ——
	{Key: "static_files", Group: "static", Label: "前端静态资源与 Swagger(不含管理后台)", Type: TypeBool, Editable: true, Get: func(c *Config) any { return c.StaticFiles }, Set: func(c *Config, v any) error {
		b, err := toBool(v)
		if err != nil {
			return err
		}
		c.StaticFiles = b
		return nil
	}},

	// —— database(引导类,只读) ——
	{Key: "database.path", Group: "database", Label: "数据库路径", Type: TypeString, Editable: false, Get: func(c *Config) any { return c.Database.Path }, Set: readOnlySet},

	// —— admin(引导类,只读;password 在 handler 脱敏) ——
	{Key: "admin.enabled", Group: "admin", Label: "启用管理后台", Type: TypeBool, Editable: false, Get: func(c *Config) any { return c.Admin.Enabled }, Set: readOnlySet},
	{Key: "admin.username", Group: "admin", Label: "管理员用户名", Type: TypeString, Editable: false, Get: func(c *Config) any { return c.Admin.Username }, Set: readOnlySet},
	{Key: "admin.email", Group: "admin", Label: "管理员邮箱", Type: TypeString, Editable: false, Get: func(c *Config) any { return c.Admin.Email }, Set: readOnlySet},
	{Key: "admin.password", Group: "admin", Label: "管理员密码", Type: TypeString, Editable: false, Get: func(c *Config) any { return c.Admin.Password }, Set: readOnlySet},
}

// RegistryByKey 返回 key→*FieldSpec 的索引(指向 Registry 元素)。
func RegistryByKey() map[string]*FieldSpec {
	m := make(map[string]*FieldSpec, len(Registry))
	for i := range Registry {
		m[Registry[i].Key] = &Registry[i]
	}
	return m
}

// ApplyDBOverrides 把数据库配置覆盖到 cfg,跳过未知/引导类/被 CLI 锁定的 key。
// 单项 JSON 解析或写回失败时忽略该项(不阻断启动),返回被跳过的 key 列表供调用方记录。
func ApplyDBOverrides(cfg *Config, dbValues map[string]string, locked map[string]bool) []string {
	byKey := RegistryByKey()
	var skipped []string
	for key, raw := range dbValues {
		spec, ok := byKey[key]
		if !ok || !spec.Editable || locked[key] {
			continue
		}
		var v any
		if err := json.Unmarshal([]byte(raw), &v); err != nil {
			skipped = append(skipped, key)
			continue
		}
		if err := spec.Set(cfg, v); err != nil {
			skipped = append(skipped, key)
			continue
		}
	}
	return skipped
}

func readOnlySet(c *Config, v any) error {
	return fmt.Errorf("配置项为只读,由配置文件/命令行管理")
}

// SplitCommaSeparated 把逗号分隔的字符串切成去空白、去空项的列表。
func SplitCommaSeparated(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func toInt(v any) (int, error) {
	switch n := v.(type) {
	case float64:
		if n != math.Trunc(n) {
			return 0, fmt.Errorf("期望整数,得到 %v", n)
		}
		return int(n), nil
	case int:
		return n, nil
	case int64:
		return int(n), nil
	case string:
		i, err := strconv.Atoi(strings.TrimSpace(n))
		if err != nil {
			return 0, fmt.Errorf("期望整数,得到 %q", n)
		}
		return i, nil
	default:
		return 0, fmt.Errorf("期望整数,得到 %T", v)
	}
}

func toFloat(v any) (float64, error) {
	switch n := v.(type) {
	case float64:
		return n, nil
	case int:
		return float64(n), nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(n), 64)
		if err != nil {
			return 0, fmt.Errorf("期望数字,得到 %q", n)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("期望数字,得到 %T", v)
	}
}

func toBool(v any) (bool, error) {
	switch b := v.(type) {
	case bool:
		return b, nil
	case string:
		parsed, err := strconv.ParseBool(strings.TrimSpace(b))
		if err != nil {
			return false, fmt.Errorf("期望布尔值,得到 %q", b)
		}
		return parsed, nil
	default:
		return false, fmt.Errorf("期望布尔值,得到 %T", v)
	}
}

func toString(v any) (string, error) {
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("期望字符串,得到 %T", v)
	}
	return s, nil
}

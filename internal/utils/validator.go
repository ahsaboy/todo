package utils

import (
	"time"
)

// ValidateTimeFormat 保留旧函数用于向后兼容，新代码应使用 ParseAPITime。
func ValidateTimeFormat(s string) bool {
	_, err := time.Parse("2006-01-02 15:04:05", s)
	if err == nil {
		return true
	}
	_, err = time.Parse(time.RFC3339, s)
	return err == nil
}

// ParseAPITime 解析 RFC3339 字符串并返回 UTC 时间。空字符串由调用方决定是否允许。
func ParseAPITime(value string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

// FormatDBTime 将 time.Time 格式化为 UTC RFC3339 字符串，用于数据库存储。
func FormatDBTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// NormalizeAPITime 解析 API 入参并返回标准 UTC RFC3339 字符串。
func NormalizeAPITime(value string) (string, error) {
	t, err := ParseAPITime(value)
	if err != nil {
		return "", err
	}
	return FormatDBTime(t), nil
}

// ParseDBTime 兼容读取数据库中的时间字符串。
// 支持两种格式：
//   - RFC3339（新格式）：例如 "2026-05-10T10:30:00Z"
//   - 旧格式（向后兼容）：例如 "2026-05-10 10:30:00"，视为 UTC
//
// 读取旧格式时不会自动转换为新格式；应通过迁移脚本统一数据。
func ParseDBTime(value string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, value)
	if err == nil {
		return t.UTC(), nil
	}
	// 兼容旧格式，视为 UTC
	t, err = time.Parse("2006-01-02 15:04:05", value)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

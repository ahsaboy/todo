package utils

import (
	"fmt"
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

// ParseAPITime 解析 RFC3339 字符串并返回 UTC 时间。
// 支持格式：
//   - RFC3339："2026-05-10T18:30:00Z"、"2026-05-10T18:30:00+08:00"
//   - ISO8601 无冒号时区偏移："2026-05-10T18:30:00+0800"、"2026-05-10T18:30:00-0500"
//
// 空字符串由调用方决定是否允许。
func ParseAPITime(value string) (time.Time, error) {
	// 尝试标准 RFC3339 格式
	t, err := time.Parse(time.RFC3339, value)
	if err == nil {
		return t.UTC(), nil
	}

	// 尝试 ISO8601 无冒号时区偏移格式（+0800、-0500 等）
	t, err = parseISO8601NoColonTimezone(value)
	if err == nil {
		return t.UTC(), nil
	}

	return time.Time{}, fmt.Errorf("不支持的时间格式，请使用 RFC3339 格式")
}

// parseISO8601NoColonTimezone 解析 ISO8601 格式但时区偏移没有冒号的情况
// 例如：
//   - 2026-05-10T18:30:00+0800（T 分隔 + 无冒号时区）
//   - 2026-05-10 18:30:00+0800（空格分隔 + 无冒号时区）
//   - 2026-05-10T18:30:00-0500
//   - 2026-05-10 18:30:00-0500
func parseISO8601NoColonTimezone(value string) (time.Time, error) {
	// ISO8601 无冒号时区偏移的格式布局（支持 T 分隔和空格分隔）
	layouts := []string{
		"2006-01-02T15:04:05-0700",
		"2006-01-02T15:04:05+0700",
		"2006-01-02 15:04:05-0700",
		"2006-01-02 15:04:05+0700",
		"2006-01-02T15:04:05.000-0700",
		"2006-01-02T15:04:05.000+0700",
		"2006-01-02 15:04:05.000-0700",
		"2006-01-02 15:04:05.000+0700",
	}

	for _, layout := range layouts {
		t, err := time.Parse(layout, value)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("不支持的时间格式")
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

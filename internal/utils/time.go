package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	_ "time/tzdata"
)

// FormatOutputTime 把 DB 字符串以目标时区重新格式化为 RFC3339(带偏移)。
//
// 语义:
//   - 空字符串 → 返回 ""
//   - loc == nil → 兜底用 time.UTC,避免 panic
//   - 解析失败 → 原样返回(best-effort,不 panic、不 error)
//   - 成功 → t.In(loc).Format(time.RFC3339),例如 "2026-05-10T18:30:00+08:00"
//
// 支持读取两种 DB 格式:
//   - RFC3339(新格式):  2026-05-10T10:30:00Z
//   - 旧格式(向后兼容): 2026-05-10 10:30:00 (视为 UTC)
func FormatOutputTime(value string, loc *time.Location) string {
	if value == "" {
		return ""
	}
	if loc == nil {
		loc = time.UTC
	}
	t, err := ParseDBTime(value)
	if err != nil {
		return value
	}
	return t.In(loc).Format(time.RFC3339)
}

// FormatOutputTimePtr 处理 *string 字段,便于在 view 函数中避免散落 nil 检查。
// nil 输入 → nil 输出;否则返回新分配的指针(避免共享底层 string)。
func FormatOutputTimePtr(p *string, loc *time.Location) *string {
	if p == nil {
		return nil
	}
	formatted := FormatOutputTime(*p, loc)
	return &formatted
}

// ResolveTimezone 解析配置中的时区名称,返回一个永远非 nil 的 *time.Location。
//
// 语义:
//   - "" / "Local"        → time.Local, nil
//   - "UTC"               → time.UTC, nil
//   - IANA 名(Asia/Shanghai 等) → time.LoadLocation(name)
//   - 固定偏移(+08:00 / -05:00 / +0800 / -0500) → time.FixedZone
//   - 失败              → time.Local, err(返回值仍非 nil,err 由调用方决定如何处理)
func ResolveTimezone(name string) (*time.Location, error) {
	name = strings.TrimSpace(name)
	if name == "" || strings.EqualFold(name, "Local") {
		return time.Local, nil
	}
	if strings.EqualFold(name, "UTC") {
		return time.UTC, nil
	}
	if loc, err := time.LoadLocation(name); err == nil {
		return loc, nil
	}
	if loc, err := parseFixedOffset(name); err == nil {
		return loc, nil
	}
	return time.Local, fmt.Errorf("unrecognized timezone %q", name)
}

// parseFixedOffset 把固定偏移字符串(±HH:MM / ±HHMM / ±HH)转为 time.FixedZone。
// 例如:
//   - "+08:00"  → "UTC+08:00", 8*3600
//   - "+0800"   → "UTC+08:00", 8*3600
//   - "-05:30"  → "UTC-05:30", -5*3600 - 30*60
func parseFixedOffset(name string) (*time.Location, error) {
	if len(name) < 3 {
		return nil, fmt.Errorf("offset too short: %q", name)
	}
	sign := name[0]
	if sign != '+' && sign != '-' {
		return nil, fmt.Errorf("offset must start with +/-: %q", name)
	}
	rest := name[1:]
	rest = strings.ReplaceAll(rest, ":", "")

	var hh, mm int
	var err error
	switch len(rest) {
	case 2: // "+08"
		hh, err = strconv.Atoi(rest)
		if err != nil {
			return nil, fmt.Errorf("invalid offset hours: %q", name)
		}
	case 4: // "+0800"
		hh, err = strconv.Atoi(rest[:2])
		if err != nil {
			return nil, fmt.Errorf("invalid offset hours: %q", name)
		}
		mm, err = strconv.Atoi(rest[2:])
		if err != nil {
			return nil, fmt.Errorf("invalid offset minutes: %q", name)
		}
	default:
		return nil, fmt.Errorf("unsupported offset length: %q", name)
	}
	if hh < 0 || hh > 23 || mm < 0 || mm > 59 {
		return nil, fmt.Errorf("offset out of range: %q", name)
	}
	offset := hh*3600 + mm*60
	if sign == '-' {
		offset = -offset
	}
	// 名称统一格式化为 "UTC±HH:MM",便于在日志和调试中识别。
	label := fmt.Sprintf("UTC%c%02d:%02d", sign, hh, mm)
	return time.FixedZone(label, offset), nil
}

package utils

import (
	"testing"
	"time"
)

func TestParseAPITime(t *testing.T) {
	asiaShanghai, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load Asia/Shanghai: %v", err)
	}
	tests := []struct {
		name    string
		input   string
		loc     *time.Location
		want    time.Time
		wantErr bool
	}{
		{"RFC3339 UTC", "2026-05-10T10:30:00Z", nil, time.Date(2026, 5, 10, 10, 30, 0, 0, time.UTC), false},
		{"RFC3339 +08:00", "2026-05-10T18:30:00+08:00", nil, time.Date(2026, 5, 10, 10, 30, 0, 0, time.UTC), false},
		{"ISO8601 T分隔 +0800", "2026-05-10T18:30:00+0800", nil, time.Date(2026, 5, 10, 10, 30, 0, 0, time.UTC), false},
		{"ISO8601 空格分隔 +0800", "2026-05-10 18:30:00+0800", nil, time.Date(2026, 5, 10, 10, 30, 0, 0, time.UTC), false},
		{"ISO8601 T分隔 -0500", "2026-05-10T18:30:00-0500", nil, time.Date(2026, 5, 10, 23, 30, 0, 0, time.UTC), false},
		{"ISO8601 空格分隔 -0500", "2026-05-10 18:30:00-0500", nil, time.Date(2026, 5, 10, 23, 30, 0, 0, time.UTC), false},
		{"无时区 T分隔 → Asia/Shanghai", "2026-05-10T18:30:00", asiaShanghai, time.Date(2026, 5, 10, 10, 30, 0, 0, time.UTC), false},
		{"无时区 空格分隔 → Asia/Shanghai", "2026-05-10 18:30:00", asiaShanghai, time.Date(2026, 5, 10, 10, 30, 0, 0, time.UTC), false},
		{"无时区 + 毫秒 T分隔 → Asia/Shanghai", "2026-05-10T18:30:00.123", asiaShanghai, time.Date(2026, 5, 10, 10, 30, 0, 123000000, time.UTC), false},
		{"无时区 + 毫秒 空格分隔 → Asia/Shanghai", "2026-05-10 18:30:00.123", asiaShanghai, time.Date(2026, 5, 10, 10, 30, 0, 123000000, time.UTC), false},
		{"无时区 + loc=UTC", "2026-05-10T18:30:00", time.UTC, time.Date(2026, 5, 10, 18, 30, 0, 0, time.UTC), false},
		{"空字符串", "", nil, time.Time{}, true},
		{"非法字符串", "not a time", nil, time.Time{}, true},
		{"只有日期", "2026-05-10", asiaShanghai, time.Time{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAPITime(tt.input, tt.loc)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseAPITime(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && !got.Equal(tt.want) {
				t.Fatalf("ParseAPITime(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatDBTime(t *testing.T) {
	input := time.Date(2026, 5, 10, 10, 30, 0, 0, time.UTC)
	got := FormatDBTime(input)
	want := "2026-05-10T10:30:00Z"
	if got != want {
		t.Fatalf("FormatDBTime() = %q, want %q", got, want)
	}
}

func TestNormalizeAPITime(t *testing.T) {
	asiaShanghai, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load Asia/Shanghai: %v", err)
	}
	tests := []struct {
		name    string
		input   string
		loc     *time.Location
		want    string
		wantErr bool
	}{
		{"RFC3339 UTC", "2026-05-10T10:30:00Z", nil, "2026-05-10T10:30:00Z", false},
		{"RFC3339 +08:00", "2026-05-10T18:30:00+08:00", nil, "2026-05-10T10:30:00Z", false},
		{"ISO8601 T分隔 +0800", "2026-05-10T18:30:00+0800", nil, "2026-05-10T10:30:00Z", false},
		{"ISO8601 空格分隔 +0800", "2026-05-10 18:30:00+0800", nil, "2026-05-10T10:30:00Z", false},
		{"ISO8601 T分隔 -0500", "2026-05-10T18:30:00-0500", nil, "2026-05-10T23:30:00Z", false},
		{"ISO8601 空格分隔 -0500", "2026-05-10 18:30:00-0500", nil, "2026-05-10T23:30:00Z", false},
		{"无时区 T分隔 → Asia/Shanghai", "2026-05-10T18:30:00", asiaShanghai, "2026-05-10T10:30:00Z", false},
		{"无时区 空格分隔 → Asia/Shanghai", "2026-05-10 18:30:00", asiaShanghai, "2026-05-10T10:30:00Z", false},
		{"无时区 → UTC", "2026-05-10T18:30:00", time.UTC, "2026-05-10T18:30:00Z", false},
		{"空字符串", "", nil, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeAPITime(tt.input, tt.loc)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NormalizeAPITime(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Fatalf("NormalizeAPITime(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseDBTime(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{"RFC3339", "2026-05-10T10:30:00Z", time.Date(2026, 5, 10, 10, 30, 0, 0, time.UTC), false},
		{"旧格式", "2026-05-10 10:30:00", time.Date(2026, 5, 10, 10, 30, 0, 0, time.UTC), false},
		{"空字符串", "", time.Time{}, true},
		{"非法字符串", "xyz", time.Time{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDBTime(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseDBTime(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && !got.Equal(tt.want) {
				t.Fatalf("ParseDBTime(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatOutputTime(t *testing.T) {
	asiaShanghai, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load Asia/Shanghai: %v", err)
	}
	newYork, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("load America/New_York: %v", err)
	}
	fixed8, err := ResolveTimezone("+08:00")
	if err != nil {
		t.Fatalf("ResolveTimezone(+08:00): %v", err)
	}

	tests := []struct {
		name  string
		input string
		loc   *time.Location
		want  string
	}{
		{"RFC3339 UTC → Asia/Shanghai", "2026-05-10T10:30:00Z", asiaShanghai, "2026-05-10T18:30:00+08:00"},
		{"RFC3339 +08:00 → Asia/Shanghai", "2026-05-10T18:30:00+08:00", asiaShanghai, "2026-05-10T18:30:00+08:00"},
		{"旧格式 → Asia/Shanghai", "2026-05-10 10:30:00", asiaShanghai, "2026-05-10T18:30:00+08:00"},
		{"RFC3339 UTC → America/New_York (DST)", "2026-05-10T10:30:00Z", newYork, "2026-05-10T06:30:00-04:00"},
		{"RFC3339 UTC → 固定偏移 +08:00", "2026-05-10T10:30:00Z", fixed8, "2026-05-10T18:30:00+08:00"},
		{"空字符串保持空", "", asiaShanghai, ""},
		{"非法字符串原样返回", "not a time", asiaShanghai, "not a time"},
		{"loc=nil 兜底 UTC", "2026-05-10T10:30:00Z", nil, "2026-05-10T10:30:00Z"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatOutputTime(tt.input, tt.loc)
			if got != tt.want {
				t.Fatalf("FormatOutputTime(%q, %v) = %q, want %q", tt.input, tt.loc, got, tt.want)
			}
		})
	}
}

func TestFormatOutputTimePtr(t *testing.T) {
	asiaShanghai, _ := time.LoadLocation("Asia/Shanghai")

	t.Run("nil 输入返回 nil", func(t *testing.T) {
		if got := FormatOutputTimePtr(nil, asiaShanghai); got != nil {
			t.Fatalf("expected nil, got %v", *got)
		}
	})
	t.Run("非 nil 输入返回新指针", func(t *testing.T) {
		in := "2026-05-10T10:30:00Z"
		got := FormatOutputTimePtr(&in, asiaShanghai)
		if got == nil {
			t.Fatal("expected non-nil")
		}
		if &in == got {
			t.Fatal("expected new pointer, got same pointer")
		}
		if *got != "2026-05-10T18:30:00+08:00" {
			t.Fatalf("got %q, want %q", *got, "2026-05-10T18:30:00+08:00")
		}
		// 原对象未被修改
		if in != "2026-05-10T10:30:00Z" {
			t.Fatalf("original modified: %q", in)
		}
	})
	t.Run("空字符串指针保持空字符串语义", func(t *testing.T) {
		empty := ""
		got := FormatOutputTimePtr(&empty, asiaShanghai)
		if got == nil {
			t.Fatal("expected non-nil")
		}
		if *got != "" {
			t.Fatalf("got %q, want empty", *got)
		}
	})
}

func TestResolveTimezone(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantName  string // 期望的 loc.String();为 "" 时不校验
		wantOff   int    // 期望的 UTC 偏移秒数;为 -1 时不校验
		wantErr   bool
		nonNilLoc bool // 期望返回的 loc 非 nil
	}{
		{name: "空", input: "", wantName: "Local", wantOff: -1, nonNilLoc: true},
		{name: "Local 大小写", input: "local", wantName: "Local", wantOff: -1, nonNilLoc: true},
		{name: "UTC", input: "UTC", wantName: "UTC", wantOff: 0, nonNilLoc: true},
		{name: "IANA Asia/Shanghai", input: "Asia/Shanghai", wantName: "Asia/Shanghai", wantOff: 8 * 3600, nonNilLoc: true},
		{name: "固定偏移 +08:00", input: "+08:00", wantName: "UTC+08:00", wantOff: 8 * 3600, nonNilLoc: true},
		{name: "固定偏移 +0800", input: "+0800", wantName: "UTC+08:00", wantOff: 8 * 3600, nonNilLoc: true},
		{name: "固定偏移 -05:00", input: "-05:00", wantName: "UTC-05:00", wantOff: -5 * 3600, nonNilLoc: true},
		{name: "固定偏移 -0530", input: "-05:30", wantName: "UTC-05:30", wantOff: -5*3600 - 30*60, nonNilLoc: true},
		{name: "非法 → fallback Local", input: "Mars/Olympus", wantName: "Local", wantOff: -1, wantErr: true, nonNilLoc: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := ResolveTimezone(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ResolveTimezone(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if tt.nonNilLoc && loc == nil {
				t.Fatalf("expected non-nil loc")
			}
			if tt.wantName != "" && loc.String() != tt.wantName {
				t.Fatalf("ResolveTimezone(%q).String() = %q, want %q", tt.input, loc.String(), tt.wantName)
			}
			if tt.wantOff >= 0 {
				// 用 2026 年 1 月避免 DST 干扰
				_, off := time.Date(2026, 1, 15, 12, 0, 0, 0, loc).Zone()
				if off != tt.wantOff {
					t.Fatalf("ResolveTimezone(%q) offset = %d, want %d", tt.input, off, tt.wantOff)
				}
			}
		})
	}
}


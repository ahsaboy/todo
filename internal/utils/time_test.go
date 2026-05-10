package utils

import (
	"testing"
	"time"
)

func TestParseAPITime(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{"RFC3339 UTC", "2026-05-10T10:30:00Z", time.Date(2026, 5, 10, 10, 30, 0, 0, time.UTC), false},
		{"RFC3339 +08:00", "2026-05-10T18:30:00+08:00", time.Date(2026, 5, 10, 10, 30, 0, 0, time.UTC), false},
		{"空字符串", "", time.Time{}, true},
		{"旧格式", "2026-05-10 18:30:00", time.Time{}, true},
		{"非法字符串", "not a time", time.Time{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAPITime(tt.input)
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
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"RFC3339 UTC", "2026-05-10T10:30:00Z", "2026-05-10T10:30:00Z", false},
		{"RFC3339 +08:00", "2026-05-10T18:30:00+08:00", "2026-05-10T10:30:00Z", false},
		{"旧格式", "2026-05-10 18:30:00", "", true},
		{"空字符串", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeAPITime(tt.input)
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

package models

import (
	"testing"
)

func TestCalculateNextDueDate(t *testing.T) {
	tests := []struct {
		name     string
		current  string
		repeat   string
		interval int
		want     string
		wantErr  bool
	}{
		{
			name:     "daily +1",
			current:  "2026-05-10T10:30:00Z",
			repeat:   "daily",
			interval: 1,
			want:     "2026-05-11T10:30:00Z",
		},
		{
			name:     "weekly +1",
			current:  "2026-05-10T10:30:00Z",
			repeat:   "weekly",
			interval: 1,
			want:     "2026-05-17T10:30:00Z",
		},
		{
			name:     "monthly +1",
			current:  "2026-05-10T10:30:00Z",
			repeat:   "monthly",
			interval: 1,
			want:     "2026-06-10T10:30:00Z",
		},
		{
			name:     "yearly +1",
			current:  "2026-05-10T10:30:00Z",
			repeat:   "yearly",
			interval: 1,
			want:     "2027-05-10T10:30:00Z",
		},
		{
			name:     "带时区输入转 UTC",
			current:  "2026-05-10T18:30:00+08:00",
			repeat:   "daily",
			interval: 1,
			want:     "2026-05-11T10:30:00Z",
		},
		{
			name:    "旧格式输入报错",
			current: "2026-05-10 10:30:00",
			repeat:  "daily",
			wantErr: true,
		},
		{
			name:    "非法输入报错",
			current: "not-a-time",
			repeat:  "daily",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateNextDueDate(tt.current, tt.repeat, tt.interval)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

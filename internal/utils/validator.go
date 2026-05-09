package utils

import (
	"time"
)

func ValidateTimeFormat(s string) bool {
	_, err := time.Parse("2006-01-02 15:04:05", s)
	if err == nil {
		return true
	}
	_, err = time.Parse(time.RFC3339, s)
	return err == nil
}

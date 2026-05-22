package service

import "errors"

var (
	ErrUsernameTaken          = errors.New("username already taken")
	ErrReminderChannelMissing = errors.New("at least one enabled reminder channel is required")
	ErrInvalidTime            = errors.New("invalid time format, expected RFC3339")
	ErrInvalidCredentials     = errors.New("invalid username or password")
	ErrInvalidOldPassword     = errors.New("invalid old password")
	ErrUserNotFound           = errors.New("user not found")
)

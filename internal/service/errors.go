package service

import "errors"

var (
	ErrUsernameTaken          = errors.New("username already taken")
	ErrReminderChannelMissing = errors.New("at least one enabled reminder channel is required")
	ErrInvalidTime            = errors.New("invalid time format, expected RFC3339")
	ErrInvalidCredentials     = errors.New("invalid username or password")
	ErrInvalidOldPassword     = errors.New("invalid old password")
	ErrUserNotFound           = errors.New("user not found")
	ErrTagNameEmpty           = errors.New("tag name cannot be empty")
	ErrInvalidTagColor        = errors.New("invalid tag color, expected #RRGGBB hex")
	ErrInvalidTagIcon         = errors.New("invalid tag icon, must be in the curated list")
	ErrUnknownTag             = errors.New("task references a tag that does not exist for this user")
)

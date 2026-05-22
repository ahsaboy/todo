package repository

import "errors"

var ErrUsernameTaken = errors.New("username already taken")

// ErrNotFound is returned when a record does not exist.
var ErrNotFound = errors.New("record not found")

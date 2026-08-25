package domain

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrInvalid      = errors.New("invalid input")
	ErrConflict     = errors.New("version conflict")
	ErrInvalidState = errors.New("invalid state transition")
	ErrFrozen       = errors.New("configuration is frozen")
)

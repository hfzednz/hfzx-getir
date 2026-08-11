package domain

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrAlreadyExists     = errors.New("already exists")
	ErrInvalidArgument   = errors.New("invalid argument")
	ErrConflict          = errors.New("conflict")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrIllegalTransition = errors.New("illegal status transition")
	ErrSandboxViolation  = errors.New("sandbox violation")
	ErrRateLimited       = errors.New("rate limited")
	ErrNotReady          = errors.New("technology not ready")
)

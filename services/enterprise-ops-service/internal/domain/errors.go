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
	ErrRateLimited       = errors.New("rate limited")
	ErrApprovalRequired  = errors.New("approval required")
)

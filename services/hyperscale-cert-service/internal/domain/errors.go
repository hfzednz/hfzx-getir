package domain

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrAlreadyExists     = errors.New("already exists")
	ErrInvalidArgument   = errors.New("invalid argument")
	ErrConflict          = errors.New("conflict")
	ErrForbidden         = errors.New("forbidden")
	ErrIllegalTransition = errors.New("illegal status transition")
	ErrRateLimited       = errors.New("rate limited")
	ErrGateFailed        = errors.New("hyperscale gate failed")
)

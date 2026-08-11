package domain

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrInvalidArgument   = errors.New("invalid argument")
	ErrConflict          = errors.New("conflict")
	ErrIllegalTransition = errors.New("illegal status transition")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
)

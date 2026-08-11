package domain

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrAlreadyExists       = errors.New("already exists")
	ErrInvalidArgument     = errors.New("invalid argument")
	ErrInvariant           = errors.New("invariant violation")
	ErrConflict            = errors.New("conflict")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrIdempotencyConflict = errors.New("idempotency key conflict")
	ErrIllegalTransition   = errors.New("illegal status transition")
	ErrPeriodClosed        = errors.New("accounting period closed")
	ErrUnbalanced          = errors.New("journal unbalanced")
	ErrApprovalRequired    = errors.New("approval required")
)

package domain

import "errors"

// Sentinel domain errors. Adapters map these to HTTP/gRPC codes.
var (
	ErrNotFound            = errors.New("not found")
	ErrAlreadyExists       = errors.New("already exists")
	ErrInvalidArgument     = errors.New("invalid argument")
	ErrInvariant           = errors.New("invariant violation")
	ErrConflict            = errors.New("conflict")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrCurrencyMismatch    = errors.New("currency mismatch")
	ErrNegativeMoney       = errors.New("money amount must be non-negative")
	ErrIdempotencyConflict = errors.New("idempotency key conflict")
	ErrOverdraft           = errors.New("insufficient available balance")
	ErrHoldNotFound        = errors.New("hold not found")
	ErrHoldReleased        = errors.New("hold already released")
	ErrLimitExceeded       = errors.New("wallet limit exceeded")
)

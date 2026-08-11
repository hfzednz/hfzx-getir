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
	ErrSuppressed          = errors.New("notification suppressed")
	ErrProviderFailed      = errors.New("provider failed")
	ErrIdempotencyConflict = errors.New("idempotency key conflict")
	ErrMaxRetries          = errors.New("max retries exceeded")
)

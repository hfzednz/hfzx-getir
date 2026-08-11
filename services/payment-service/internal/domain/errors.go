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
	ErrInvalidTransition   = errors.New("invalid status transition")
	ErrVersionConflict     = errors.New("optimistic version conflict")
	ErrCurrencyMismatch    = errors.New("currency mismatch")
	ErrNegativeMoney       = errors.New("money amount must be non-negative")
	ErrIdempotencyConflict = errors.New("idempotency key conflict")
	ErrFraudBlocked        = errors.New("fraud blocked")
	ErrInsufficientCapture = errors.New("capture exceeds authorized amount")
	ErrRefundExceeds       = errors.New("refund exceeds capturable amount")
	ErrPSPFailed           = errors.New("psp provider failed")
	ErrNoProviderRoute     = errors.New("no provider route available")
	ErrOverdraft           = errors.New("insufficient available balance")
)

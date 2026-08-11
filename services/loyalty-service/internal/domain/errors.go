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
	ErrInsufficientPoints  = errors.New("insufficient points")
	ErrSelfReferral        = errors.New("self referral not allowed")
	ErrReferralInvalid     = errors.New("referral invalid")
	ErrCashbackState       = errors.New("invalid cashback state")
	ErrIdempotencyConflict = errors.New("idempotency key conflict")
)

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
	ErrUnbalancedJournal   = errors.New("journal unbalanced: sum debit must equal sum credit")
	ErrJournalImmutable    = errors.New("journal is immutable after post")
	ErrInvoiceImmutable    = errors.New("invoice cannot be mutated in current status")
)

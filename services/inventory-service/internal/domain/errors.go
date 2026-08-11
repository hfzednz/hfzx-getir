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
	ErrInsufficientStock   = errors.New("insufficient available stock")
	ErrVersionConflict     = errors.New("optimistic version conflict")
	ErrReservationExpired  = errors.New("reservation expired")
	ErrReservationInactive = errors.New("reservation not active")
	ErrLotNotAllocatable   = errors.New("lot not allocatable")
	ErrNegativeQuantity    = errors.New("quantity must be positive")
)

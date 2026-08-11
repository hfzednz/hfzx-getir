package domain

import "errors"

// Sentinel domain errors. Adapters map these to HTTP/gRPC codes.
var (
	ErrNotFound          = errors.New("not found")
	ErrAlreadyExists     = errors.New("already exists")
	ErrInvalidArgument   = errors.New("invalid argument")
	ErrInvariant         = errors.New("invariant violation")
	ErrConflict          = errors.New("conflict")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrBarcodeMismatch   = errors.New("barcode mismatch")
	ErrOverpick          = errors.New("quantity exceeds remaining")
	ErrRemainingQty      = errors.New("remaining quantity not fulfilled")
	ErrTaskNotClaimable  = errors.New("task not claimable from current status")
	ErrStationBusy       = errors.New("station not available")
	ErrScanRejected      = errors.New("pick scan rejected")
	ErrNotSealed         = errors.New("pack session not sealed")
	ErrWeightMismatch    = errors.New("pack weight outside tolerance")
	ErrAlreadyTerminal   = errors.New("entity already in terminal state")
)

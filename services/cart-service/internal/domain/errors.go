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
	ErrVersionConflict   = errors.New("optimistic version conflict")
	ErrCurrencyMismatch  = errors.New("currency mismatch")
	ErrNegativeMoney     = errors.New("money amount must be non-negative")
	ErrCartInactive      = errors.New("cart is not active")
	ErrMaxQtyExceeded    = errors.New("quantity exceeds max available")
	ErrCouponInvalid     = errors.New("coupon invalid")
	ErrCouponNotApplied  = errors.New("coupon not applied")
	ErrEmptyCart         = errors.New("cart has no lines")
	ErrOwnerRequired     = errors.New("guest token or principal required")
)

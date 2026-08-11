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
	ErrCampaignInactive  = errors.New("campaign is not active")
	ErrCouponInvalid     = errors.New("coupon invalid")
	ErrCouponExhausted   = errors.New("coupon usage exhausted")
	ErrCouponRedeemed    = errors.New("coupon already redeemed")
	ErrVoucherInvalid    = errors.New("voucher invalid")
	ErrVoucherExhausted  = errors.New("voucher usage exhausted")
	ErrUsageLimit        = errors.New("usage limit reached")
	ErrNotEligible       = errors.New("not eligible for promotion")
)

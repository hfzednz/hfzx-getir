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
	ErrPrincipalLocked   = errors.New("principal locked")
	ErrPrincipalSuspended = errors.New("principal suspended")
	ErrPrincipalDeleted  = errors.New("principal deleted")
	ErrSessionExpired    = errors.New("session expired")
	ErrSessionRevoked    = errors.New("session revoked")
	ErrTokenExpired      = errors.New("token expired")
	ErrTokenRevoked      = errors.New("token revoked")
	ErrTokenReuse        = errors.New("refresh token reuse detected")
	ErrDeviceRevoked     = errors.New("device revoked")
	ErrMFARequired       = errors.New("mfa required")
	ErrMFAFailed         = errors.New("mfa verification failed")
	ErrRiskBlocked       = errors.New("blocked by risk policy")
	ErrSecurityViolation = errors.New("security violation")
	ErrRateLimited       = errors.New("rate limited")
)

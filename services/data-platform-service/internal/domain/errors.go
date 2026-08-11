package domain

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrAlreadyExists       = errors.New("already exists")
	ErrInvalidArgument     = errors.New("invalid argument")
	ErrInvariant           = errors.New("invariant violation")
	ErrConflict            = errors.New("conflict")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrIdempotencyConflict = errors.New("idempotency key conflict")
	ErrSchemaIncompatible  = errors.New("schema incompatible")
	ErrQualityFailed       = errors.New("data quality failed")
)

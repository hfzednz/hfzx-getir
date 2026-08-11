package domain

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrAlreadyExists     = errors.New("already exists")
	ErrInvalidArgument   = errors.New("invalid argument")
	ErrConflict          = errors.New("conflict")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrIllegalTransition = errors.New("illegal status transition")
	ErrCountryInactive   = errors.New("country inactive")
	ErrMissingTranslation = errors.New("missing translation")
)

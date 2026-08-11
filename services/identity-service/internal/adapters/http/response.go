package httpadapter

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/nexora/identity-service/internal/domain"
)

// APIError is the NEXORA error envelope.
type APIError struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody is the nested error object.
type ErrorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	TraceID   string `json:"traceId"`
	Retriable bool   `json:"retriable"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeOK(w http.ResponseWriter, v any) {
	writeJSON(w, http.StatusOK, v)
}

func writeCreated(w http.ResponseWriter, v any) {
	writeJSON(w, http.StatusCreated, v)
}

func writeNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func writeErr(w http.ResponseWriter, r *http.Request, err error) {
	traceID := RequestIDFromContext(r.Context())
	code, status, retriable, msg := mapError(err)
	writeJSON(w, status, APIError{Error: ErrorBody{
		Code:      code,
		Message:   msg,
		TraceID:   traceID,
		Retriable: retriable,
	}})
}

func mapError(err error) (code string, status int, retriable bool, message string) {
	if err == nil {
		return "internal_error", http.StatusInternalServerError, true, "unexpected error"
	}
	message = err.Error()
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return "not_found", http.StatusNotFound, false, message
	case errors.Is(err, domain.ErrAlreadyExists):
		return "already_exists", http.StatusConflict, false, message
	case errors.Is(err, domain.ErrInvalidArgument):
		return "invalid_argument", http.StatusBadRequest, false, message
	case errors.Is(err, domain.ErrUnauthorized), errors.Is(err, domain.ErrTokenExpired), errors.Is(err, domain.ErrTokenRevoked):
		return "unauthorized", http.StatusUnauthorized, false, message
	case errors.Is(err, domain.ErrForbidden), errors.Is(err, domain.ErrPrincipalLocked), errors.Is(err, domain.ErrPrincipalSuspended):
		return "forbidden", http.StatusForbidden, false, message
	case errors.Is(err, domain.ErrMFARequired):
		return "mfa_required", http.StatusUnauthorized, false, message
	case errors.Is(err, domain.ErrMFAFailed):
		return "mfa_failed", http.StatusUnauthorized, false, message
	case errors.Is(err, domain.ErrRateLimited):
		return "rate_limited", http.StatusTooManyRequests, true, message
	case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrTokenReuse):
		return "conflict", http.StatusConflict, false, message
	case errors.Is(err, domain.ErrRiskBlocked):
		return "risk_blocked", http.StatusForbidden, false, message
	default:
		return "internal_error", http.StatusInternalServerError, true, message
	}
}

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	return nil
}

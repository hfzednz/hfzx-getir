package httpadapter

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/nexora/settlement-service/internal/domain"
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

func writeErr(w http.ResponseWriter, r *http.Request, err error) {
	traceID := RequestIDFromContext(r.Context())
	code, status, retriable, msg := mapError(err)
	writeJSON(w, status, APIError{Error: ErrorBody{
		Code: code, Message: msg, TraceID: traceID, Retriable: retriable,
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
	case errors.Is(err, domain.ErrInvalidArgument), errors.Is(err, domain.ErrNegativeMoney),
		errors.Is(err, domain.ErrCurrencyMismatch), errors.Is(err, domain.ErrBatchNotEmpty):
		return "invalid_argument", http.StatusBadRequest, false, message
	case errors.Is(err, domain.ErrDualControl), errors.Is(err, domain.ErrForbidden):
		return "forbidden", http.StatusForbidden, false, message
	case errors.Is(err, domain.ErrInvalidTransition):
		return "invalid_transition", http.StatusConflict, false, message
	case errors.Is(err, domain.ErrUnauthorized):
		return "unauthorized", http.StatusUnauthorized, false, message
	case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrVersionConflict),
		errors.Is(err, domain.ErrIdempotencyConflict):
		return "conflict", http.StatusConflict, false, message
	case errors.Is(err, domain.ErrInvariant):
		return "invariant_violation", http.StatusUnprocessableEntity, false, message
	default:
		return "internal_error", http.StatusInternalServerError, true, message
	}
}

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

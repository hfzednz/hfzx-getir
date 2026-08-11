package httpadapter

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/nexora/quality-service/internal/domain"
)

type APIError struct {
	Error ErrorBody `json:"error"`
}

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

func writeOK(w http.ResponseWriter, v any)      { writeJSON(w, http.StatusOK, v) }
func writeCreated(w http.ResponseWriter, v any) { writeJSON(w, http.StatusCreated, v) }

func writeErr(w http.ResponseWriter, r *http.Request, err error) {
	code, status, retriable, msg := mapError(err)
	writeJSON(w, status, APIError{Error: ErrorBody{
		Code: code, Message: msg, TraceID: RequestIDFromContext(r.Context()), Retriable: retriable,
	}})
}

func mapError(err error) (code string, status int, retriable bool, message string) {
	if err == nil {
		return "internal_error", 500, true, "unexpected"
	}
	message = err.Error()
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return "not_found", 404, false, message
	case errors.Is(err, domain.ErrInvalidArgument):
		return "invalid_argument", 400, false, message
	case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrAlreadyExists):
		return "conflict", 409, false, message
	case errors.Is(err, domain.ErrGateFailed), errors.Is(err, domain.ErrNotCertified):
		return "quality_gate_failed", 422, false, message
	case errors.Is(err, domain.ErrForbidden), errors.Is(err, domain.ErrUnauthorized):
		return "forbidden", 403, false, message
	case errors.Is(err, domain.ErrIllegalTransition):
		return "illegal_transition", 409, false, message
	default:
		return "internal_error", 500, true, message
	}
}

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

package httpadapter

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/nexora/crm-service/internal/domain"
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

func ticketDTO(t domain.Ticket) map[string]any {
	out := map[string]any{
		"id":          t.ID.String(),
		"tenantId":    t.TenantID.String(),
		"customerId":  t.CustomerID.String(),
		"status":      t.Status,
		"priority":    t.Priority,
		"category":    t.Category,
		"subject":     t.Subject,
		"description": t.Description,
		"slaBreached": t.SLABreached,
		"tags":        t.Tags,
		"createdAt":   t.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":   t.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if t.AssigneeID != nil {
		out["assigneeId"] = t.AssigneeID.String()
	}
	if t.TeamID != nil {
		out["teamId"] = t.TeamID.String()
	}
	if t.ResolvedAt != nil {
		out["resolvedAt"] = t.ResolvedAt.UTC().Format(time.RFC3339)
	}
	if t.ClosedAt != nil {
		out["closedAt"] = t.ClosedAt.UTC().Format(time.RFC3339)
	}
	return out
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
	case errors.Is(err, domain.ErrInvalidArgument):
		return "invalid_argument", http.StatusBadRequest, false, message
	case errors.Is(err, domain.ErrForbidden), errors.Is(err, domain.ErrUnauthorized):
		return "forbidden", http.StatusForbidden, false, message
	case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrIdempotencyConflict):
		return "conflict", http.StatusConflict, false, message
	case errors.Is(err, domain.ErrIllegalTransition), errors.Is(err, domain.ErrInvariant):
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

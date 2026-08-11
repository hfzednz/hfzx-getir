package httpadapter

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/settlement-service/internal/app"
	"github.com/nexora/settlement-service/internal/domain"
	"github.com/nexora/settlement-service/internal/ratelimit"
)

// Handler serves settlement REST endpoints.
type Handler struct {
	Deps  *app.Deps
	Ready func(*http.Request) error
	Live  func(*http.Request) error
}

// ServerConfig configures the HTTP server.
type ServerConfig struct {
	Addr               string
	Deps               *app.Deps
	Limiter            ratelimit.Limiter
	RateLimitPerMinute int
	CORSOrigins        []string
	Log                *slog.Logger
	Ready              func(*http.Request) error
	Live               func(*http.Request) error
}

// NewHandler returns a fully wired http.Handler.
func NewHandler(cfg ServerConfig) http.Handler {
	h := &Handler{Deps: cfg.Deps, Ready: cfg.Ready, Live: cfg.Live}
	mux := http.NewServeMux()
	const base = "/v1/settlements"

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /ready", h.ready)
	mux.HandleFunc("GET "+base+"/health", h.health)
	mux.HandleFunc("GET "+base+"/ready", h.ready)

	mux.HandleFunc("POST "+base+"/batches", tenant(h.createBatch))
	mux.HandleFunc("GET "+base+"/batches", tenant(h.listBatches))
	mux.HandleFunc("GET "+base+"/batches/{id}", tenant(h.getBatch))
	mux.HandleFunc("POST "+base+"/batches/{id}/lines", tenant(h.addLine))
	mux.HandleFunc("POST "+base+"/batches/{id}/submit", tenant(h.submitBatch))
	mux.HandleFunc("POST "+base+"/batches/{id}/approve", tenant(h.approveBatch))
	mux.HandleFunc("POST "+base+"/batches/{id}/execute", tenant(h.executePayouts))
	mux.HandleFunc("POST "+base+"/batches/{id}/reconcile", tenant(h.reconcile))
	mux.HandleFunc("POST "+base+"/outbox/publish", tenant(h.publishOutbox))

	return chain(mux,
		requestIDMiddleware,
		recoverMiddleware(cfg.Log),
		loggingMiddleware(cfg.Log),
		corsMiddleware(cfg.CORSOrigins),
		rateLimitMiddleware(cfg.Limiter, cfg.RateLimitPerMinute),
	)
}

// NewServer builds an *http.Server with sensible timeouts.
func NewServer(cfg ServerConfig) *http.Server {
	return &http.Server{
		Addr:              cfg.Addr,
		Handler:           NewHandler(cfg),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func tenant(next func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantMiddleware(http.HandlerFunc(next)).ServeHTTP(w, r)
	}
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	if h.Live != nil {
		if err := h.Live(r); err != nil {
			writeErr(w, r, err)
			return
		}
	}
	writeOK(w, map[string]string{"status": "ok"})
}

func (h *Handler) ready(w http.ResponseWriter, r *http.Request) {
	if h.Ready != nil {
		if err := h.Ready(r); err != nil {
			writeErr(w, r, err)
			return
		}
	}
	writeOK(w, map[string]string{"status": "ready"})
}

func (h *Handler) tenant(r *http.Request) uuid.UUID {
	tid, _ := TenantIDFromContext(r.Context())
	return tid
}

func (h *Handler) actor(r *http.Request, bodyActor *uuid.UUID) uuid.UUID {
	if uid, ok := UserIDFromContext(r.Context()); ok {
		return uid
	}
	if bodyActor != nil {
		return *bodyActor
	}
	return uuid.Nil
}

func parseUUIDParam(r *http.Request, name string) (uuid.UUID, error) {
	return uuid.Parse(r.PathValue(name))
}

func idemKey(r *http.Request, bodyKey string) string {
	if k := r.Header.Get("Idempotency-Key"); k != "" {
		return k
	}
	return bodyKey
}

func (h *Handler) createBatch(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Currency       string     `json:"currency"`
		PeriodStart    time.Time  `json:"periodStart"`
		PeriodEnd      time.Time  `json:"periodEnd"`
		Description    string     `json:"description"`
		IdempotencyKey string     `json:"idempotencyKey"`
		ActorID        *uuid.UUID `json:"actorId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	b, err := h.Deps.CreateBatch(r.Context(), app.CreateBatchInput{
		TenantID: h.tenant(r), Currency: body.Currency,
		PeriodStart: body.PeriodStart, PeriodEnd: body.PeriodEnd,
		Description: body.Description, IdempotencyKey: idemKey(r, body.IdempotencyKey),
		ActorID: h.actor(r, body.ActorID),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, batchDTO(b))
}

func (h *Handler) listBatches(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	var status *domain.BatchStatus
	if s := r.URL.Query().Get("status"); s != "" {
		st := domain.BatchStatus(s)
		status = &st
	}
	items, total, err := h.Deps.List(r.Context(), app.ListBatchesInput{
		TenantID: h.tenant(r), Status: status, Limit: limit, Offset: offset,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	out := make([]any, 0, len(items))
	for _, b := range items {
		out = append(out, batchDTO(b))
	}
	writeOK(w, map[string]any{"items": out, "total": total})
}

func (h *Handler) getBatch(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	b, err := h.Deps.GetBatch(r.Context(), h.tenant(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, batchDTO(b))
}

func (h *Handler) addLine(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		PayeeType   string `json:"payeeType"`
		PayeeRef    string `json:"payeeRef"`
		AmountMinor int64  `json:"amountMinor"`
		ExternalRef string `json:"externalRef"`
		Memo        string `json:"memo"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	b, err := h.Deps.AddLine(r.Context(), app.AddLineInput{
		TenantID: h.tenant(r), BatchID: id, PayeeType: domain.PayeeType(body.PayeeType),
		PayeeRef: body.PayeeRef, AmountMinor: body.AmountMinor,
		ExternalRef: body.ExternalRef, Memo: body.Memo,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, batchDTO(b))
}

func (h *Handler) submitBatch(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		ActorID *uuid.UUID `json:"actorId"`
	}
	_ = decodeJSON(r, &body)
	b, err := h.Deps.Submit(r.Context(), app.SubmitInput{
		TenantID: h.tenant(r), BatchID: id, ActorID: h.actor(r, body.ActorID),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, batchDTO(b))
}

func (h *Handler) approveBatch(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		ActorID *uuid.UUID `json:"actorId"`
	}
	_ = decodeJSON(r, &body)
	b, err := h.Deps.Approve(r.Context(), app.ApproveInput{
		TenantID: h.tenant(r), BatchID: id, ActorID: h.actor(r, body.ActorID),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, batchDTO(b))
}

func (h *Handler) executePayouts(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		ActorID *uuid.UUID `json:"actorId"`
	}
	_ = decodeJSON(r, &body)
	b, err := h.Deps.ExecutePayouts(r.Context(), app.ExecutePayoutsInput{
		TenantID: h.tenant(r), BatchID: id, ActorID: h.actor(r, body.ActorID),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, batchDTO(b))
}

func (h *Handler) reconcile(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		ProviderRef   string     `json:"providerRef"`
		ReportedMinor int64      `json:"reportedMinor"`
		ActorID       *uuid.UUID `json:"actorId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.ReconcileProviderReport(r.Context(), app.ReconcileProviderReportInput{
		TenantID: h.tenant(r), BatchID: id, ProviderRef: body.ProviderRef,
		ReportedMinor: body.ReportedMinor, ActorID: h.actor(r, body.ActorID),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	out := map[string]any{
		"matched": res.Matched,
		"reconciliation": map[string]any{
			"id": res.Reconciliation.ID.String(), "expectedMinor": res.Reconciliation.ExpectedMinor,
			"reportedMinor": res.Reconciliation.ReportedMinor, "providerRef": res.Reconciliation.ProviderRef,
		},
	}
	if res.Mismatch != nil {
		out["mismatch"] = map[string]any{
			"id": res.Mismatch.ID.String(), "deltaMinor": res.Mismatch.DeltaMinor, "detail": res.Mismatch.Detail,
		}
	}
	writeOK(w, out)
}

func (h *Handler) publishOutbox(w http.ResponseWriter, r *http.Request) {
	n, err := h.Deps.PublishPending(r.Context(), 100)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"published": n})
}

func batchDTO(b domain.SettlementBatch) map[string]any {
	lines := make([]map[string]any, 0, len(b.Lines))
	for _, l := range b.Lines {
		lines = append(lines, map[string]any{
			"id": l.ID.String(), "payeeType": string(l.PayeeType), "payeeRef": l.PayeeRef,
			"amountMinor": l.AmountMinor, "currency": l.Currency, "externalRef": l.ExternalRef, "memo": l.Memo,
		})
	}
	out := map[string]any{
		"id": b.ID.String(), "status": string(b.Status), "currency": b.Currency,
		"description": b.Description, "totalMinor": b.TotalMinor, "lines": lines,
		"periodStart": b.PeriodStart.UTC().Format(time.RFC3339),
		"periodEnd":   b.PeriodEnd.UTC().Format(time.RFC3339),
	}
	if b.SubmittedBy != nil {
		out["submittedBy"] = b.SubmittedBy.String()
	}
	if b.ApprovedBy != nil {
		out["approvedBy"] = b.ApprovedBy.String()
	}
	return out
}

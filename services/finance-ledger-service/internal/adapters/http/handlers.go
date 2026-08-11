package httpadapter

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/finance-ledger-service/internal/app"
	"github.com/nexora/finance-ledger-service/internal/domain"
	"github.com/nexora/finance-ledger-service/internal/ratelimit"
)

// Handler serves ledger REST endpoints.
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
	const base = "/v1/ledger"

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /ready", h.ready)
	mux.HandleFunc("GET "+base+"/health", h.health)
	mux.HandleFunc("GET "+base+"/ready", h.ready)

	mux.HandleFunc("POST "+base+"/accounts", tenant(h.ensureAccount))
	mux.HandleFunc("GET "+base+"/accounts/{id}/balance", tenant(h.getBalance))
	mux.HandleFunc("POST "+base+"/journals", tenant(h.postJournal))
	mux.HandleFunc("GET "+base+"/journals", tenant(h.listJournals))
	mux.HandleFunc("GET "+base+"/journals/{id}", tenant(h.getJournal))
	mux.HandleFunc("POST "+base+"/invoices", tenant(h.createInvoice))
	mux.HandleFunc("POST "+base+"/invoices/{id}/credit-notes", tenant(h.issueCreditNote))
	mux.HandleFunc("POST "+base+"/tax/calculate", tenant(h.taxCalculate))
	mux.HandleFunc("POST "+base+"/tax-rules", tenant(h.upsertTaxRule))
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

func parseUUIDParam(r *http.Request, name string) (uuid.UUID, error) {
	return uuid.Parse(r.PathValue(name))
}

func idemKey(r *http.Request, bodyKey string) string {
	if k := r.Header.Get("Idempotency-Key"); k != "" {
		return k
	}
	return bodyKey
}

func (h *Handler) ensureAccount(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code     string `json:"code"`
		Name     string `json:"name"`
		Type     string `json:"type"`
		Currency string `json:"currency"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	acc, err := h.Deps.EnsureAccount(r.Context(), app.EnsureAccountInput{
		TenantID: h.tenant(r), Code: body.Code, Name: body.Name,
		Type: domain.AccountType(body.Type), Currency: body.Currency,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, accountDTO(acc))
}

func (h *Handler) getBalance(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	bal, err := h.Deps.GetBalance(r.Context(), app.GetBalanceInput{TenantID: h.tenant(r), AccountID: id})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{
		"accountId": bal.AccountID.String(), "currency": bal.Currency, "balanceMinor": bal.BalanceMinor,
	})
}

func (h *Handler) postJournal(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Currency       string `json:"currency"`
		Reference      string `json:"reference"`
		Description    string `json:"description"`
		IdempotencyKey string `json:"idempotencyKey"`
		Lines          []struct {
			AccountID   *uuid.UUID `json:"accountId"`
			AccountCode string     `json:"accountCode"`
			DebitMinor  int64      `json:"debitMinor"`
			CreditMinor int64      `json:"creditMinor"`
			Memo        string     `json:"memo"`
		} `json:"lines"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	lines := make([]app.JournalLineInput, 0, len(body.Lines))
	for _, l := range body.Lines {
		li := app.JournalLineInput{
			AccountCode: l.AccountCode, DebitMinor: l.DebitMinor, CreditMinor: l.CreditMinor, Memo: l.Memo,
		}
		if l.AccountID != nil {
			li.AccountID = *l.AccountID
		}
		lines = append(lines, li)
	}
	j, err := h.Deps.PostJournal(r.Context(), app.PostJournalInput{
		TenantID: h.tenant(r), Currency: body.Currency, Reference: body.Reference,
		Description: body.Description, IdempotencyKey: idemKey(r, body.IdempotencyKey), Lines: lines,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, journalDTO(j))
}

func (h *Handler) listJournals(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	var status *domain.JournalStatus
	if s := r.URL.Query().Get("status"); s != "" {
		st := domain.JournalStatus(s)
		status = &st
	}
	items, total, err := h.Deps.ListJournals(r.Context(), app.ListJournalsInput{
		TenantID: h.tenant(r), Status: status, Limit: limit, Offset: offset,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	out := make([]any, 0, len(items))
	for _, j := range items {
		out = append(out, journalDTO(j))
	}
	writeOK(w, map[string]any{"items": out, "total": total})
}

func (h *Handler) getJournal(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	j, err := h.Deps.GetJournal(r.Context(), h.tenant(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, journalDTO(j))
}

func (h *Handler) createInvoice(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Currency        string `json:"currency"`
		CounterpartyRef string `json:"counterpartyRef"`
		ExternalRef     string `json:"externalRef"`
		IdempotencyKey  string `json:"idempotencyKey"`
		Issue           bool   `json:"issue"`
		DefaultTaxCode  string `json:"defaultTaxCode"`
		Lines           []struct {
			Description string `json:"description"`
			Qty         int64  `json:"qty"`
			UnitMinor   int64  `json:"unitMinor"`
			TaxCode     string `json:"taxCode"`
		} `json:"lines"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	lines := make([]app.InvoiceLineInput, 0, len(body.Lines))
	for _, l := range body.Lines {
		lines = append(lines, app.InvoiceLineInput{
			Description: l.Description, Qty: l.Qty, UnitMinor: l.UnitMinor, TaxCode: l.TaxCode,
		})
	}
	inv, err := h.Deps.CreateInvoice(r.Context(), app.CreateInvoiceInput{
		TenantID: h.tenant(r), Currency: body.Currency, CounterpartyRef: body.CounterpartyRef,
		ExternalRef: body.ExternalRef, IdempotencyKey: idemKey(r, body.IdempotencyKey),
		Issue: body.Issue, DefaultTaxCode: body.DefaultTaxCode, Lines: lines,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, invoiceDTO(inv))
}

func (h *Handler) issueCreditNote(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		AmountMinor    int64  `json:"amountMinor"`
		Reason         string `json:"reason"`
		IdempotencyKey string `json:"idempotencyKey"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	cn, err := h.Deps.IssueCreditNote(r.Context(), app.IssueCreditNoteInput{
		TenantID: h.tenant(r), InvoiceID: id, AmountMinor: body.AmountMinor,
		Reason: body.Reason, IdempotencyKey: idemKey(r, body.IdempotencyKey),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, map[string]any{
		"id": cn.ID.String(), "invoiceId": cn.InvoiceID.String(),
		"currency": cn.Currency, "amountMinor": cn.AmountMinor, "reason": cn.Reason,
	})
}

func (h *Handler) taxCalculate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		BaseMinor int64  `json:"baseMinor"`
		Currency  string `json:"currency"`
		TaxCode   string `json:"taxCode"`
		RateBps   *int64 `json:"rateBps"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.TaxCalculate(r.Context(), app.TaxCalculateInput{
		TenantID: h.tenant(r), BaseMinor: body.BaseMinor, Currency: body.Currency,
		TaxCode: body.TaxCode, RateBps: body.RateBps,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{
		"baseMinor": res.BaseMinor, "taxMinor": res.TaxMinor, "totalMinor": res.TotalMinor,
		"rateBps": res.RateBps, "taxCode": res.TaxCode, "currency": res.Currency,
	})
}

func (h *Handler) upsertTaxRule(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code     string `json:"code"`
		Name     string `json:"name"`
		RateBps  int64  `json:"rateBps"`
		Currency string `json:"currency"`
		Active   *bool  `json:"active"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	active := true
	if body.Active != nil {
		active = *body.Active
	}
	rule, err := h.Deps.UpsertTaxRule(r.Context(), app.UpsertTaxRuleInput{
		TenantID: h.tenant(r), Code: body.Code, Name: body.Name,
		RateBps: body.RateBps, Currency: body.Currency, Active: active,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, map[string]any{
		"id": rule.ID.String(), "code": rule.Code, "name": rule.Name,
		"rateBps": rule.RateBps, "currency": rule.Currency, "active": rule.Active,
	})
}

func (h *Handler) publishOutbox(w http.ResponseWriter, r *http.Request) {
	n, err := h.Deps.PublishPending(r.Context(), 100)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"published": n})
}

func accountDTO(a domain.Account) map[string]any {
	return map[string]any{
		"id": a.ID.String(), "code": a.Code, "name": a.Name,
		"type": string(a.Type), "currency": a.Currency, "active": a.Active,
	}
}

func journalDTO(j domain.Journal) map[string]any {
	lines := make([]map[string]any, 0, len(j.Lines))
	for _, l := range j.Lines {
		lines = append(lines, map[string]any{
			"id": l.ID.String(), "accountId": l.AccountID.String(), "accountCode": l.AccountCode,
			"debitMinor": l.DebitMinor, "creditMinor": l.CreditMinor, "currency": l.Currency, "memo": l.Memo,
		})
	}
	out := map[string]any{
		"id": j.ID.String(), "status": string(j.Status), "currency": j.Currency,
		"reference": j.Reference, "description": j.Description,
		"debitTotal": j.DebitTotal(), "creditTotal": j.CreditTotal(), "lines": lines,
	}
	if j.PostedAt != nil {
		out["postedAt"] = j.PostedAt.UTC().Format(time.RFC3339)
	}
	return out
}

func invoiceDTO(inv domain.Invoice) map[string]any {
	lines := make([]map[string]any, 0, len(inv.Lines))
	for _, l := range inv.Lines {
		lines = append(lines, map[string]any{
			"id": l.ID.String(), "description": l.Description, "qty": l.Qty,
			"unitMinor": l.UnitMinor, "taxMinor": l.TaxMinor, "totalMinor": l.TotalMinor, "taxCode": l.TaxCode,
		})
	}
	return map[string]any{
		"id": inv.ID.String(), "status": string(inv.Status), "currency": inv.Currency,
		"counterpartyRef": inv.CounterpartyRef, "externalRef": inv.ExternalRef,
		"subtotalMinor": inv.SubtotalMinor, "taxMinor": inv.TaxMinor, "totalMinor": inv.TotalMinor,
		"lines": lines,
	}
}

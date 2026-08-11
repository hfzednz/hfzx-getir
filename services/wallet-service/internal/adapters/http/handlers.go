package httpadapter

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/wallet-service/internal/app"
	"github.com/nexora/wallet-service/internal/app/ports"
	"github.com/nexora/wallet-service/internal/domain"
	"github.com/nexora/wallet-service/internal/ratelimit"
)

// Handler serves wallet REST endpoints.
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
	const base = "/v1/wallets"

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /ready", h.ready)
	mux.HandleFunc("GET "+base+"/health", h.health)
	mux.HandleFunc("GET "+base+"/ready", h.ready)

	mux.HandleFunc("POST "+base, tenant(h.getOrCreate))
	mux.HandleFunc("GET "+base+"/{id}", tenant(h.getWallet))
	mux.HandleFunc("POST "+base+"/{id}/credit", tenant(h.credit))
	mux.HandleFunc("POST "+base+"/{id}/debit", tenant(h.debit))
	mux.HandleFunc("POST "+base+"/{id}/hold", tenant(h.hold))
	mux.HandleFunc("POST "+base+"/holds/{holdId}/release", tenant(h.release))
	mux.HandleFunc("POST "+base+"/{id}/transfer", tenant(h.transfer))
	mux.HandleFunc("GET "+base+"/{id}/history", tenant(h.history))
	mux.HandleFunc("POST "+base+"/{id}/admin/adjust", tenant(h.adminAdjust))
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

func (h *Handler) tenantID(r *http.Request) uuid.UUID {
	tid, _ := TenantIDFromContext(r.Context())
	return tid
}

func (h *Handler) principal(r *http.Request, bodyPrincipal *uuid.UUID) uuid.UUID {
	if uid, ok := UserIDFromContext(r.Context()); ok {
		return uid
	}
	if bodyPrincipal != nil {
		return *bodyPrincipal
	}
	return uuid.Nil
}

func idemKey(r *http.Request, bodyKey string) string {
	if v := r.Header.Get("Idempotency-Key"); v != "" {
		return v
	}
	return bodyKey
}

func walletDTO(v ports.WalletView) map[string]any {
	accts := make([]map[string]any, 0, len(v.Accounts))
	for _, a := range v.Accounts {
		accts = append(accts, map[string]any{
			"id": a.ID, "accountType": a.AccountType,
			"balanceMinor": a.BalanceMinor, "heldMinor": a.HeldMinor,
			"availableMinor": a.Available(), "currency": a.Currency,
		})
	}
	return map[string]any{
		"id": v.Wallet.ID, "tenantId": v.Wallet.TenantID, "principalId": v.Wallet.PrincipalID,
		"currency": v.Wallet.Currency, "active": v.Wallet.Active, "accounts": accts,
		"createdAt": v.Wallet.CreatedAt, "updatedAt": v.Wallet.UpdatedAt,
	}
}

func (h *Handler) getOrCreate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PrincipalID *uuid.UUID `json:"principalId"`
		Currency    string     `json:"currency"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	view, err := h.Deps.GetOrCreate(r.Context(), app.GetOrCreateInput{
		TenantID: h.tenantID(r), PrincipalID: h.principal(r, body.PrincipalID), Currency: body.Currency,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, walletDTO(view))
}

func (h *Handler) getWallet(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	view, err := h.Deps.GetWallet(r.Context(), h.tenantID(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, walletDTO(view))
}

func (h *Handler) credit(w http.ResponseWriter, r *http.Request) {
	h.moneyOp(w, r, true)
}

func (h *Handler) debit(w http.ResponseWriter, r *http.Request) {
	h.moneyOp(w, r, false)
}

func (h *Handler) moneyOp(w http.ResponseWriter, r *http.Request, credit bool) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		AccountType    domain.AccountType `json:"accountType"`
		AmountMinor    int64              `json:"amountMinor"`
		IdempotencyKey string             `json:"idempotencyKey"`
		Reference      string             `json:"reference"`
	}
	_ = decodeJSON(r, &body)
	in := app.MoneyInput{
		TenantID: h.tenantID(r), WalletID: id, AccountType: body.AccountType,
		AmountMinor: body.AmountMinor, IdempotencyKey: idemKey(r, body.IdempotencyKey),
		Reference: body.Reference,
	}
	var acct domain.Account
	var entry domain.Entry
	if credit {
		acct, entry, err = h.Deps.Credit(r.Context(), in)
	} else {
		acct, entry, err = h.Deps.Debit(r.Context(), in)
	}
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{
		"account": map[string]any{
			"id": acct.ID, "accountType": acct.AccountType,
			"balanceMinor": acct.BalanceMinor, "heldMinor": acct.HeldMinor,
			"availableMinor": acct.Available(),
		},
		"entry": entry,
	})
}

func (h *Handler) hold(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		AccountType    domain.AccountType `json:"accountType"`
		AmountMinor    int64              `json:"amountMinor"`
		IdempotencyKey string             `json:"idempotencyKey"`
		Reference      string             `json:"reference"`
	}
	_ = decodeJSON(r, &body)
	hold, acct, err := h.Deps.Hold(r.Context(), app.HoldInput{
		TenantID: h.tenantID(r), WalletID: id, AccountType: body.AccountType,
		AmountMinor: body.AmountMinor, IdempotencyKey: idemKey(r, body.IdempotencyKey),
		Reference: body.Reference,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, map[string]any{"hold": hold, "availableMinor": acct.Available(), "heldMinor": acct.HeldMinor})
}

func (h *Handler) release(w http.ResponseWriter, r *http.Request) {
	holdID, err := uuid.Parse(r.PathValue("holdId"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	hold, acct, err := h.Deps.Release(r.Context(), app.ReleaseInput{TenantID: h.tenantID(r), HoldID: holdID})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"hold": hold, "availableMinor": acct.Available(), "heldMinor": acct.HeldMinor})
}

func (h *Handler) transfer(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		FromAccountType domain.AccountType `json:"fromAccountType"`
		ToWalletID      *uuid.UUID         `json:"toWalletId"`
		ToAccountType   domain.AccountType `json:"toAccountType"`
		AmountMinor     int64              `json:"amountMinor"`
		IdempotencyKey  string             `json:"idempotencyKey"`
		Reference       string             `json:"reference"`
	}
	_ = decodeJSON(r, &body)
	toWallet := id
	if body.ToWalletID != nil {
		toWallet = *body.ToWalletID
	}
	xfer, err := h.Deps.Transfer(r.Context(), app.TransferInput{
		TenantID: h.tenantID(r), FromWalletID: id, FromAccountType: body.FromAccountType,
		ToWalletID: toWallet, ToAccountType: body.ToAccountType,
		AmountMinor: body.AmountMinor, IdempotencyKey: idemKey(r, body.IdempotencyKey),
		Reference: body.Reference,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, xfer)
}

func (h *Handler) history(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, total, err := h.Deps.History(r.Context(), app.HistoryInput{
		TenantID: h.tenantID(r), WalletID: id, Limit: limit, Offset: offset,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items, "total": total})
}

func (h *Handler) adminAdjust(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		AccountType    domain.AccountType `json:"accountType"`
		AmountMinor    int64              `json:"amountMinor"`
		IdempotencyKey string             `json:"idempotencyKey"`
		Reason         string             `json:"reason"`
	}
	_ = decodeJSON(r, &body)
	var actor *uuid.UUID
	if uid, ok := UserIDFromContext(r.Context()); ok {
		actor = &uid
	}
	acct, entry, err := h.Deps.AdminAdjust(r.Context(), app.AdminAdjustInput{
		TenantID: h.tenantID(r), WalletID: id, AccountType: body.AccountType,
		AmountMinor: body.AmountMinor, IdempotencyKey: idemKey(r, body.IdempotencyKey),
		Reason: body.Reason, ActorID: actor,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{
		"account": map[string]any{
			"balanceMinor": acct.BalanceMinor, "heldMinor": acct.HeldMinor, "availableMinor": acct.Available(),
		},
		"entry": entry,
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

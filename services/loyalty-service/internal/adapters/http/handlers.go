package httpadapter

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/app"
	"github.com/nexora/loyalty-service/internal/domain"
	"github.com/nexora/loyalty-service/internal/ratelimit"
)

// Handler serves loyalty REST endpoints.
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
	const base = "/v1/loyalty"

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /ready", h.ready)
	mux.HandleFunc("GET "+base+"/health", h.health)
	mux.HandleFunc("GET "+base+"/ready", h.ready)

	mux.HandleFunc("POST "+base+"/accounts", tenant(h.ensureAccount))
	mux.HandleFunc("GET "+base+"/accounts/{id}", tenant(h.getAccount))
	mux.HandleFunc("GET "+base+"/accounts/{id}/history", tenant(h.pointsHistory))
	mux.HandleFunc("POST "+base+"/points/earn", tenant(h.earnPoints))
	mux.HandleFunc("POST "+base+"/points/redeem", tenant(h.redeemPoints))
	mux.HandleFunc("GET "+base+"/accounts/{id}/membership", tenant(h.getMembership))
	mux.HandleFunc("POST "+base+"/accounts/{id}/membership/evaluate", tenant(h.evaluateMembership))
	mux.HandleFunc("GET "+base+"/rewards", tenant(h.listRewards))
	mux.HandleFunc("POST "+base+"/rewards/unlock", tenant(h.unlockReward))
	mux.HandleFunc("POST "+base+"/rewards/redeem", tenant(h.redeemReward))
	mux.HandleFunc("POST "+base+"/cashback/grant", tenant(h.grantCashback))
	mux.HandleFunc("POST "+base+"/cashback/{id}/confirm", tenant(h.confirmCashback))
	mux.HandleFunc("POST "+base+"/referrals/code", tenant(h.createReferralCode))
	mux.HandleFunc("POST "+base+"/referrals/apply", tenant(h.applyReferral))
	mux.HandleFunc("POST "+base+"/referrals/complete", tenant(h.completeReferral))
	mux.HandleFunc("POST "+base+"/missions/track", tenant(h.trackMission))
	mux.HandleFunc("POST "+base+"/missions/complete", tenant(h.completeMission))
	mux.HandleFunc("POST "+base+"/streaks", tenant(h.updateStreak))
	mux.HandleFunc("POST "+base+"/spin", tenant(h.spin))
	mux.HandleFunc("POST "+base+"/achievements/unlock", tenant(h.unlockAchievement))
	mux.HandleFunc("GET "+base+"/leaderboard", tenant(h.leaderboard))
	mux.HandleFunc("GET "+base+"/accounts/{id}/ai-scores", tenant(h.aiScores))
	mux.HandleFunc("POST "+base+"/admin/grant", tenant(h.adminGrant))
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

func accountDTO(a domain.Account) map[string]any {
	return map[string]any{
		"id": a.ID, "tenantId": a.TenantID, "principalId": a.PrincipalID,
		"points": a.Points, "tierPoints": a.TierPoints, "xp": a.XP, "level": a.Level,
		"active": a.Active, "createdAt": a.CreatedAt, "updatedAt": a.UpdatedAt,
	}
}

func (h *Handler) ensureAccount(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PrincipalID *uuid.UUID `json:"principalId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	acct, err := h.Deps.EnsureAccount(r.Context(), app.EnsureAccountInput{
		TenantID: h.tenantID(r), PrincipalID: h.principal(r, body.PrincipalID),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, accountDTO(acct))
}

func (h *Handler) getAccount(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	acct, err := h.Deps.GetAccount(r.Context(), h.tenantID(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, accountDTO(acct))
}

func (h *Handler) pointsHistory(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	entries, total, err := h.Deps.PointsHistory(r.Context(), h.tenantID(r), id, limit, offset)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": entries, "total": total})
}

func (h *Handler) earnPoints(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PrincipalID *uuid.UUID `json:"principalId"`
		OrderID     uuid.UUID  `json:"orderId"`
		Points      int64      `json:"points"`
		SpendMinor  int64      `json:"spendMinor"`
		Reference   string     `json:"reference"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	acct, entry, err := h.Deps.EarnPoints(r.Context(), app.EarnPointsInput{
		TenantID: h.tenantID(r), PrincipalID: h.principal(r, body.PrincipalID),
		OrderID: body.OrderID, Points: body.Points, SpendMinor: body.SpendMinor, Reference: body.Reference,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"account": accountDTO(acct), "entry": entry})
}

func (h *Handler) redeemPoints(w http.ResponseWriter, r *http.Request) {
	var body struct {
		AccountID      uuid.UUID `json:"accountId"`
		Points         int64     `json:"points"`
		IdempotencyKey string    `json:"idempotencyKey"`
		Reference      string    `json:"reference"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	if v := r.Header.Get("Idempotency-Key"); v != "" {
		body.IdempotencyKey = v
	}
	acct, entry, err := h.Deps.RedeemPoints(r.Context(), app.RedeemPointsInput{
		TenantID: h.tenantID(r), AccountID: body.AccountID, Points: body.Points,
		IdempotencyKey: body.IdempotencyKey, Reference: body.Reference,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"account": accountDTO(acct), "entry": entry})
}

func (h *Handler) getMembership(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	m, err := h.Deps.GetMembership(r.Context(), h.tenantID(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, m)
}

func (h *Handler) evaluateMembership(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	m, upgraded, err := h.Deps.EvaluateMembership(r.Context(), app.EvaluateMembershipInput{
		TenantID: h.tenantID(r), AccountID: id,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"membership": m, "upgraded": upgraded})
}

func (h *Handler) listRewards(w http.ResponseWriter, r *http.Request) {
	list, err := h.Deps.Rewards.ListRewards(r.Context(), h.tenantID(r))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list})
}

func (h *Handler) unlockReward(w http.ResponseWriter, r *http.Request) {
	var body struct {
		AccountID uuid.UUID `json:"accountId"`
		RewardID  uuid.UUID `json:"rewardId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	red, err := h.Deps.UnlockReward(r.Context(), app.UnlockRewardInput{
		TenantID: h.tenantID(r), AccountID: body.AccountID, RewardID: body.RewardID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, red)
}

func (h *Handler) redeemReward(w http.ResponseWriter, r *http.Request) {
	var body struct {
		AccountID    uuid.UUID `json:"accountId"`
		RewardID     uuid.UUID `json:"rewardId"`
		RedemptionID uuid.UUID `json:"redemptionId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	red, err := h.Deps.RedeemReward(r.Context(), app.RedeemRewardInput{
		TenantID: h.tenantID(r), AccountID: body.AccountID, RewardID: body.RewardID, RedemptionID: body.RedemptionID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, red)
}

func (h *Handler) grantCashback(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PrincipalID    *uuid.UUID `json:"principalId"`
		AmountMinor    int64      `json:"amountMinor"`
		Currency       string     `json:"currency"`
		AccountType    string     `json:"accountType"`
		OrderID        *uuid.UUID `json:"orderId"`
		IdempotencyKey string     `json:"idempotencyKey"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	if v := r.Header.Get("Idempotency-Key"); v != "" {
		body.IdempotencyKey = v
	}
	g, err := h.Deps.GrantCashback(r.Context(), app.GrantCashbackInput{
		TenantID: h.tenantID(r), PrincipalID: h.principal(r, body.PrincipalID),
		AmountMinor: body.AmountMinor, Currency: body.Currency, AccountType: body.AccountType,
		OrderID: body.OrderID, IdempotencyKey: body.IdempotencyKey,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, g)
}

func (h *Handler) confirmCashback(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	g, err := h.Deps.ConfirmCashbackToWallet(r.Context(), app.ConfirmCashbackInput{
		TenantID: h.tenantID(r), GrantID: id,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, g)
}

func (h *Handler) createReferralCode(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PrincipalID *uuid.UUID `json:"principalId"`
		Code        string     `json:"code"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	c, err := h.Deps.CreateReferralCode(r.Context(), app.CreateReferralCodeInput{
		TenantID: h.tenantID(r), PrincipalID: h.principal(r, body.PrincipalID), Code: body.Code,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, c)
}

func (h *Handler) applyReferral(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PrincipalID *uuid.UUID `json:"principalId"`
		Code        string     `json:"code"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	ev, err := h.Deps.ApplyReferral(r.Context(), app.ApplyReferralInput{
		TenantID: h.tenantID(r), RefereePrincipal: h.principal(r, body.PrincipalID), Code: body.Code,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, ev)
}

func (h *Handler) completeReferral(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PrincipalID *uuid.UUID `json:"principalId"`
		OrderID     uuid.UUID  `json:"orderId"`
		Points      int64      `json:"points"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	ev, err := h.Deps.CompleteReferral(r.Context(), app.CompleteReferralInput{
		TenantID: h.tenantID(r), RefereePrincipal: h.principal(r, body.PrincipalID),
		OrderID: body.OrderID, Points: body.Points,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, ev)
}

func (h *Handler) trackMission(w http.ResponseWriter, r *http.Request) {
	var body struct {
		AccountID uuid.UUID `json:"accountId"`
		MissionID uuid.UUID `json:"missionId"`
		Delta     int64     `json:"delta"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	p, err := h.Deps.TrackMission(r.Context(), app.TrackMissionInput{
		TenantID: h.tenantID(r), AccountID: body.AccountID, MissionID: body.MissionID, Delta: body.Delta,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, p)
}

func (h *Handler) completeMission(w http.ResponseWriter, r *http.Request) {
	var body struct {
		AccountID uuid.UUID `json:"accountId"`
		MissionID uuid.UUID `json:"missionId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	p, err := h.Deps.CompleteMission(r.Context(), app.CompleteMissionInput{
		TenantID: h.tenantID(r), AccountID: body.AccountID, MissionID: body.MissionID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, p)
}

func (h *Handler) updateStreak(w http.ResponseWriter, r *http.Request) {
	var body struct {
		AccountID uuid.UUID `json:"accountId"`
		Action    string    `json:"action"`
		Date      string    `json:"date"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	s, err := h.Deps.UpdateStreak(r.Context(), app.UpdateStreakInput{
		TenantID: h.tenantID(r), AccountID: body.AccountID, Action: body.Action, Date: body.Date,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, s)
}

func (h *Handler) spin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		AccountID  uuid.UUID `json:"accountId"`
		CampaignID uuid.UUID `json:"campaignId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.Spin(r.Context(), app.SpinInput{
		TenantID: h.tenantID(r), AccountID: body.AccountID, CampaignID: body.CampaignID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, res)
}

func (h *Handler) unlockAchievement(w http.ResponseWriter, r *http.Request) {
	var body struct {
		AccountID uuid.UUID `json:"accountId"`
		Code      string    `json:"code"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	u, err := h.Deps.UnlockAchievement(r.Context(), app.UnlockAchievementInput{
		TenantID: h.tenantID(r), AccountID: body.AccountID, Code: body.Code,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, u)
}

func (h *Handler) leaderboard(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := h.Deps.GetLeaderboard(r.Context(), h.tenantID(r), limit)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) aiScores(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	s, err := h.Deps.GetAIScores(r.Context(), h.tenantID(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, s)
}

func (h *Handler) adminGrant(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PrincipalID *uuid.UUID `json:"principalId"`
		Points      int64      `json:"points"`
		Reason      string     `json:"reason"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var actor *uuid.UUID
	if uid, ok := UserIDFromContext(r.Context()); ok {
		actor = &uid
	}
	acct, entry, err := h.Deps.AdminManualGrant(r.Context(), app.AdminManualGrantInput{
		TenantID: h.tenantID(r), PrincipalID: h.principal(r, body.PrincipalID),
		Points: body.Points, Reason: body.Reason, ActorID: actor,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"account": accountDTO(acct), "entry": entry})
}

func (h *Handler) publishOutbox(w http.ResponseWriter, r *http.Request) {
	n, err := h.Deps.PublishPending(r.Context(), 100)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"published": n})
}

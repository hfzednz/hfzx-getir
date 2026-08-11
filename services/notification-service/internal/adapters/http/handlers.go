package httpadapter

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/notification-service/internal/app"
	"github.com/nexora/notification-service/internal/domain"
	"github.com/nexora/notification-service/internal/ratelimit"
)

// Handler serves notification REST endpoints.
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
	const base = "/v1/notifications"

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /ready", h.ready)
	mux.HandleFunc("GET "+base+"/health", h.health)
	mux.HandleFunc("GET "+base+"/ready", h.ready)

	mux.HandleFunc("POST "+base+"/send", tenant(h.send))
	mux.HandleFunc("POST "+base+"/send/bulk", tenant(h.sendBulk))
	mux.HandleFunc("POST "+base+"/events", tenant(h.handleEvent))

	mux.HandleFunc("POST "+base+"/templates", tenant(h.upsertTemplate))
	mux.HandleFunc("POST "+base+"/templates/{id}/approve", tenant(h.approveTemplate))
	mux.HandleFunc("POST "+base+"/templates/preview", tenant(h.previewTemplate))

	mux.HandleFunc("PUT "+base+"/preferences/{principalId}", tenant(h.setPreferences))
	mux.HandleFunc("GET "+base+"/preferences/{principalId}", tenant(h.getPreferences))

	mux.HandleFunc("POST "+base+"/devices", tenant(h.registerDevice))

	mux.HandleFunc("GET "+base+"/inbox/{principalId}", tenant(h.listInbox))
	mux.HandleFunc("POST "+base+"/inbox/{id}/read", tenant(h.markInboxRead))

	mux.HandleFunc("GET "+base+"/deliveries/{id}", tenant(h.getDelivery))
	mux.HandleFunc("POST "+base+"/deliveries/{id}/retry", tenant(h.retryFailed))
	mux.HandleFunc("POST "+base+"/deliveries/{id}/dlq", tenant(h.moveDLQ))

	mux.HandleFunc("POST "+base+"/schedules", tenant(h.scheduleSend))
	mux.HandleFunc("POST "+base+"/schedules/{id}/cancel", tenant(h.cancelSchedule))
	mux.HandleFunc("POST "+base+"/schedules/process-due", tenant(h.processDue))

	mux.HandleFunc("GET "+base+"/ai/best-send-time/{principalId}", tenant(h.bestSendTime))
	mux.HandleFunc("GET "+base+"/ai/recommend-channel/{principalId}", tenant(h.recommendChannel))

	mux.HandleFunc("GET "+base+"/admin/stats", tenant(h.adminStats))
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

func requireTenant(r *http.Request) (uuid.UUID, error) {
	tid, ok := TenantIDFromContext(r.Context())
	if !ok {
		return uuid.Nil, domain.ErrInvalidArgument
	}
	return tid, nil
}

func parseUUIDParam(r *http.Request, name string) (uuid.UUID, error) {
	raw := r.PathValue(name)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, domain.ErrInvalidArgument
	}
	return id, nil
}

func principalFrom(r *http.Request, bodyID string) (uuid.UUID, error) {
	if uid, ok := UserIDFromContext(r.Context()); ok {
		return uid, nil
	}
	if bodyID != "" {
		id, err := uuid.Parse(bodyID)
		if err != nil {
			return uuid.Nil, domain.ErrInvalidArgument
		}
		return id, nil
	}
	return uuid.Nil, domain.ErrInvalidArgument
}

type sendBody struct {
	PrincipalID    string            `json:"principalId"`
	OrderID        string            `json:"orderId"`
	Channel        string            `json:"channel"`
	Priority       string            `json:"priority"`
	TemplateKey    string            `json:"templateKey"`
	Locale         string            `json:"locale"`
	Recipient      string            `json:"recipient"`
	Subject        string            `json:"subject"`
	Body           string            `json:"body"`
	Vars           map[string]string `json:"vars"`
	IdempotencyKey string            `json:"idempotencyKey"`
}

func (h *Handler) send(w http.ResponseWriter, r *http.Request) {
	tid, err := requireTenant(r)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	var body sendBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	pid, err := principalFrom(r, body.PrincipalID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	idem := body.IdempotencyKey
	if idem == "" {
		idem = r.Header.Get("Idempotency-Key")
	}
	in := app.SendInput{
		TenantID: tid, PrincipalID: pid, Channel: domain.Channel(body.Channel),
		Priority: domain.Priority(body.Priority), TemplateKey: body.TemplateKey,
		Locale: body.Locale, Recipient: body.Recipient, Subject: body.Subject,
		Body: body.Body, Vars: body.Vars, IdempotencyKey: idem,
	}
	if body.OrderID != "" {
		oid, err := uuid.Parse(body.OrderID)
		if err != nil {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
		in.OrderID = &oid
	}
	msg, err := h.Deps.Send(r.Context(), in)
	if err != nil && msg.ID == uuid.Nil {
		writeErr(w, r, err)
		return
	}
	if err != nil {
		writeOK(w, map[string]any{"message": msgDTO(msg), "warning": err.Error()})
		return
	}
	writeCreated(w, map[string]any{"message": msgDTO(msg)})
}

type sendBulkBody struct {
	PrincipalIDs   []string          `json:"principalIds"`
	Channel        string            `json:"channel"`
	Priority       string            `json:"priority"`
	TemplateKey    string            `json:"templateKey"`
	Locale         string            `json:"locale"`
	Vars           map[string]string `json:"vars"`
	IdempotencyKey string            `json:"idempotencyKey"`
}

func (h *Handler) sendBulk(w http.ResponseWriter, r *http.Request) {
	tid, err := requireTenant(r)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	var body sendBulkBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	pids := make([]uuid.UUID, 0, len(body.PrincipalIDs))
	for _, s := range body.PrincipalIDs {
		id, err := uuid.Parse(s)
		if err != nil {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
		pids = append(pids, id)
	}
	msgs, err := h.Deps.SendBulk(r.Context(), app.SendBulkInput{
		TenantID: tid, PrincipalIDs: pids, Channel: domain.Channel(body.Channel),
		Priority: domain.Priority(body.Priority), TemplateKey: body.TemplateKey,
		Locale: body.Locale, Vars: body.Vars, IdempotencyKey: body.IdempotencyKey,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	out := make([]any, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, msgDTO(m))
	}
	writeCreated(w, map[string]any{"messages": out})
}

type eventBody struct {
	PrincipalID    string            `json:"principalId"`
	OrderID        string            `json:"orderId"`
	EventType      string            `json:"eventType"`
	Channel        string            `json:"channel"`
	Priority       string            `json:"priority"`
	Recipient      string            `json:"recipient"`
	Locale         string            `json:"locale"`
	Vars           map[string]string `json:"vars"`
	IdempotencyKey string            `json:"idempotencyKey"`
}

func (h *Handler) handleEvent(w http.ResponseWriter, r *http.Request) {
	tid, err := requireTenant(r)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	var body eventBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	pid, err := principalFrom(r, body.PrincipalID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	in := app.HandleDomainEventInput{
		TenantID: tid, PrincipalID: pid, EventType: body.EventType,
		Channel: domain.Channel(body.Channel), Priority: domain.Priority(body.Priority),
		Recipient: body.Recipient, Locale: body.Locale, Vars: body.Vars,
		IdempotencyKey: body.IdempotencyKey,
	}
	if body.OrderID != "" {
		oid, err := uuid.Parse(body.OrderID)
		if err != nil {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
		in.OrderID = &oid
	}
	msg, err := h.Deps.HandleDomainEvent(r.Context(), in)
	if err != nil && msg.ID == uuid.Nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, map[string]any{"message": msgDTO(msg)})
}

type templateBody struct {
	Key      string `json:"key"`
	Channel  string `json:"channel"`
	Locale   string `json:"locale"`
	Subject  string `json:"subject"`
	Body     string `json:"body"`
	Variant  string `json:"variant"`
	Activate bool   `json:"activate"`
}

func (h *Handler) upsertTemplate(w http.ResponseWriter, r *http.Request) {
	tid, err := requireTenant(r)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	var body templateBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	t, err := h.Deps.UpsertTemplate(r.Context(), app.UpsertTemplateInput{
		TenantID: tid, Key: body.Key, Channel: domain.Channel(body.Channel),
		Locale: body.Locale, Subject: body.Subject, Body: body.Body,
		Variant: body.Variant, Activate: body.Activate,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, map[string]any{"template": tplDTO(t)})
}

func (h *Handler) approveTemplate(w http.ResponseWriter, r *http.Request) {
	tid, err := requireTenant(r)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	t, err := h.Deps.ApproveTemplate(r.Context(), app.ApproveTemplateInput{TenantID: tid, TemplateID: id})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"template": tplDTO(t)})
}

type previewBody struct {
	TemplateID string            `json:"templateId"`
	Key        string            `json:"key"`
	Channel    string            `json:"channel"`
	Locale     string            `json:"locale"`
	Vars       map[string]string `json:"vars"`
}

func (h *Handler) previewTemplate(w http.ResponseWriter, r *http.Request) {
	tid, err := requireTenant(r)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	var body previewBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	in := app.PreviewTemplateInput{
		TenantID: tid, Key: body.Key, Channel: domain.Channel(body.Channel),
		Locale: body.Locale, Vars: body.Vars,
	}
	if body.TemplateID != "" {
		id, err := uuid.Parse(body.TemplateID)
		if err != nil {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
		in.TemplateID = id
	}
	res, err := h.Deps.PreviewTemplate(r.Context(), in)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, res)
}

type prefsBody struct {
	ChannelOptOut map[string]bool `json:"channelOptOut"`
	QuietStart    *int            `json:"quietStart"`
	QuietEnd      *int            `json:"quietEnd"`
}

func (h *Handler) setPreferences(w http.ResponseWriter, r *http.Request) {
	tid, err := requireTenant(r)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	pid, err := parseUUIDParam(r, "principalId")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	var body prefsBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	p, err := h.Deps.SetPreferences(r.Context(), app.SetPreferencesInput{
		TenantID: tid, PrincipalID: pid, ChannelOptOut: body.ChannelOptOut,
		QuietStart: body.QuietStart, QuietEnd: body.QuietEnd,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"preferences": prefDTO(p)})
}

func (h *Handler) getPreferences(w http.ResponseWriter, r *http.Request) {
	tid, err := requireTenant(r)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	pid, err := parseUUIDParam(r, "principalId")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	p, err := h.Deps.GetPreferences(r.Context(), tid, pid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"preferences": prefDTO(p)})
}

type deviceBody struct {
	PrincipalID string `json:"principalId"`
	Platform    string `json:"platform"`
	Token       string `json:"token"`
	Locale      string `json:"locale"`
}

func (h *Handler) registerDevice(w http.ResponseWriter, r *http.Request) {
	tid, err := requireTenant(r)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	var body deviceBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	pid, err := principalFrom(r, body.PrincipalID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	d, err := h.Deps.RegisterDevice(r.Context(), app.RegisterDeviceInput{
		TenantID: tid, PrincipalID: pid, Platform: domain.DevicePlatform(body.Platform),
		Token: body.Token, Locale: body.Locale,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, map[string]any{"device": map[string]any{
		"id": d.ID.String(), "principalId": d.PrincipalID.String(),
		"platform": string(d.Platform), "token": d.Token, "active": d.Active,
	}})
}

func (h *Handler) listInbox(w http.ResponseWriter, r *http.Request) {
	tid, err := requireTenant(r)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	pid, err := parseUUIDParam(r, "principalId")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	include := r.URL.Query().Get("includeArchived") == "true"
	items, err := h.Deps.ListInbox(r.Context(), tid, pid, include)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	out := make([]any, 0, len(items))
	for _, it := range items {
		out = append(out, map[string]any{
			"id": it.ID.String(), "messageId": it.MessageID.String(),
			"title": it.Title, "body": it.Body, "read": it.Read, "archived": it.Archived,
			"createdAt": it.CreatedAt,
		})
	}
	writeOK(w, map[string]any{"items": out})
}

func (h *Handler) markInboxRead(w http.ResponseWriter, r *http.Request) {
	tid, err := requireTenant(r)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	it, err := h.Deps.MarkInboxRead(r.Context(), tid, id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"item": map[string]any{
		"id": it.ID.String(), "read": it.Read, "readAt": it.ReadAt,
	}})
}

func (h *Handler) getDelivery(w http.ResponseWriter, r *http.Request) {
	tid, err := requireTenant(r)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	res, err := h.Deps.GetDelivery(r.Context(), tid, id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	attempts := make([]any, 0, len(res.Attempts))
	for _, a := range res.Attempts {
		attempts = append(attempts, map[string]any{
			"id": a.ID.String(), "attemptNo": a.AttemptNo, "provider": a.Provider,
			"status": a.Status, "providerRef": a.ProviderRef, "error": a.Error,
		})
	}
	writeOK(w, map[string]any{"message": msgDTO(res.Message), "attempts": attempts})
}

func (h *Handler) retryFailed(w http.ResponseWriter, r *http.Request) {
	tid, err := requireTenant(r)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	msg, err := h.Deps.RetryFailed(r.Context(), app.RetryFailedInput{TenantID: tid, MessageID: id})
	if err != nil && msg.ID == uuid.Nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"message": msgDTO(msg)})
}

func (h *Handler) moveDLQ(w http.ResponseWriter, r *http.Request) {
	tid, err := requireTenant(r)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	item, err := h.Deps.MoveToDLQ(r.Context(), app.MoveToDLQInput{TenantID: tid, MessageID: id, Reason: "manual"})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"dlq": map[string]any{
		"id": item.ID.String(), "messageId": item.MessageID.String(), "reason": item.Reason,
	}})
}

type scheduleBody struct {
	PrincipalID    string            `json:"principalId"`
	Channel        string            `json:"channel"`
	Priority       string            `json:"priority"`
	TemplateKey    string            `json:"templateKey"`
	Locale         string            `json:"locale"`
	Recipient      string            `json:"recipient"`
	Subject        string            `json:"subject"`
	Body           string            `json:"body"`
	Vars           map[string]string `json:"vars"`
	IdempotencyKey string            `json:"idempotencyKey"`
	SendAt         time.Time         `json:"sendAt"`
}

func (h *Handler) scheduleSend(w http.ResponseWriter, r *http.Request) {
	tid, err := requireTenant(r)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	var body scheduleBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	pid, err := principalFrom(r, body.PrincipalID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	s, err := h.Deps.ScheduleSend(r.Context(), app.ScheduleSendInput{
		TenantID: tid, PrincipalID: pid, Channel: domain.Channel(body.Channel),
		Priority: domain.Priority(body.Priority), TemplateKey: body.TemplateKey,
		Locale: body.Locale, Recipient: body.Recipient, Subject: body.Subject,
		Body: body.Body, Vars: body.Vars, IdempotencyKey: body.IdempotencyKey,
		SendAt: body.SendAt,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, map[string]any{"schedule": map[string]any{
		"id": s.ID.String(), "sendAt": s.SendAt, "status": string(s.Status),
	}})
}

func (h *Handler) cancelSchedule(w http.ResponseWriter, r *http.Request) {
	tid, err := requireTenant(r)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	s, err := h.Deps.CancelSchedule(r.Context(), tid, id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"schedule": map[string]any{
		"id": s.ID.String(), "status": string(s.Status),
	}})
}

func (h *Handler) processDue(w http.ResponseWriter, r *http.Request) {
	n, err := h.Deps.ProcessDueSchedules(r.Context(), 100)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"processed": n})
}

func (h *Handler) bestSendTime(w http.ResponseWriter, r *http.Request) {
	tid, err := requireTenant(r)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	pid, err := parseUUIDParam(r, "principalId")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	res, err := h.Deps.BestSendTime(r.Context(), tid, pid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, res)
}

func (h *Handler) recommendChannel(w http.ResponseWriter, r *http.Request) {
	tid, err := requireTenant(r)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	pid, err := parseUUIDParam(r, "principalId")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	priority := domain.Priority(r.URL.Query().Get("priority"))
	res, err := h.Deps.RecommendChannel(r.Context(), tid, pid, priority)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, res)
}

func (h *Handler) adminStats(w http.ResponseWriter, r *http.Request) {
	tid, err := requireTenant(r)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	res, err := h.Deps.AdminStats(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, res)
}

func (h *Handler) publishOutbox(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	n, err := h.Deps.PublishPending(r.Context(), limit)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"published": n})
}

func msgDTO(m domain.Message) map[string]any {
	dto := map[string]any{
		"id": m.ID.String(), "tenantId": m.TenantID.String(), "principalId": m.PrincipalID.String(),
		"channel": string(m.Channel), "priority": string(m.Priority), "status": string(m.Status),
		"subject": m.Subject, "body": m.Body, "recipient": m.Recipient,
		"templateKey": m.TemplateKey, "attempts": m.Attempts, "provider": m.Provider,
		"providerRef": m.ProviderRef, "suppressReason": m.SuppressReason,
		"idempotencyKey": m.IdempotencyKey, "createdAt": m.CreatedAt, "updatedAt": m.UpdatedAt,
	}
	if m.OrderID != nil {
		dto["orderId"] = m.OrderID.String()
	}
	if m.SentAt != nil {
		dto["sentAt"] = m.SentAt
	}
	return dto
}

func tplDTO(t domain.Template) map[string]any {
	return map[string]any{
		"id": t.ID.String(), "key": t.Key, "channel": string(t.Channel),
		"locale": t.Locale, "version": t.Version, "status": string(t.Status),
		"subject": t.Subject, "body": t.Body, "variant": t.Variant,
	}
}

func prefDTO(p domain.Preference) map[string]any {
	opt := map[string]bool{}
	for k, v := range p.ChannelOptOut {
		opt[string(k)] = v
	}
	return map[string]any{
		"principalId": p.PrincipalID.String(), "channelOptOut": opt,
		"quietStart": p.QuietStart, "quietEnd": p.QuietEnd, "updatedAt": p.UpdatedAt,
	}
}

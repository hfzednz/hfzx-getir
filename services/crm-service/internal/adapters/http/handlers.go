package httpadapter

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/crm-service/internal/app"
	"github.com/nexora/crm-service/internal/domain"
	"github.com/nexora/crm-service/internal/ratelimit"
)

// Handler serves CRM REST endpoints.
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
	const base = "/v1/crm"

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /ready", h.ready)
	mux.HandleFunc("GET "+base+"/health", h.health)
	mux.HandleFunc("GET "+base+"/ready", h.ready)

	mux.HandleFunc("POST "+base+"/tickets", tenant(h.createTicket))
	mux.HandleFunc("GET "+base+"/tickets", tenant(h.listTickets))
	mux.HandleFunc("GET "+base+"/tickets/{id}", tenant(h.getTicket))
	mux.HandleFunc("POST "+base+"/tickets/{id}/assign", tenant(h.assignTicket))
	mux.HandleFunc("POST "+base+"/tickets/{id}/notes", tenant(h.addNote))
	mux.HandleFunc("POST "+base+"/tickets/{id}/escalate", tenant(h.escalateTicket))
	mux.HandleFunc("POST "+base+"/tickets/{id}/resolve", tenant(h.resolveTicket))
	mux.HandleFunc("POST "+base+"/tickets/{id}/close", tenant(h.closeTicket))
	mux.HandleFunc("POST "+base+"/tickets/{id}/reopen", tenant(h.reopenTicket))
	mux.HandleFunc("POST "+base+"/tickets/merge", tenant(h.mergeTickets))

	mux.HandleFunc("POST "+base+"/chats", tenant(h.startChat))
	mux.HandleFunc("POST "+base+"/chats/{id}/messages", tenant(h.postMessage))
	mux.HandleFunc("POST "+base+"/chats/{id}/transfer", tenant(h.transferChat))
	mux.HandleFunc("POST "+base+"/chats/{id}/end", tenant(h.endChat))

	mux.HandleFunc("POST "+base+"/ai/assist", tenant(h.aiAssist))

	mux.HandleFunc("POST "+base+"/kb/articles", tenant(h.upsertArticle))
	mux.HandleFunc("POST "+base+"/kb/articles/{id}/publish", tenant(h.publishArticle))
	mux.HandleFunc("GET "+base+"/kb/search", tenant(h.searchKB))

	mux.HandleFunc("POST "+base+"/cases", tenant(h.createCase))
	mux.HandleFunc("PATCH "+base+"/cases/{id}", tenant(h.updateCase))

	mux.HandleFunc("POST "+base+"/feedback/csat", tenant(h.submitCSAT))
	mux.HandleFunc("POST "+base+"/feedback/nps", tenant(h.submitNPS))

	mux.HandleFunc("GET "+base+"/customers/{customerId}/360", tenant(h.customer360))

	mux.HandleFunc("POST "+base+"/sla/policies", tenant(h.upsertSLA))
	mux.HandleFunc("POST "+base+"/sla/evaluate", tenant(h.evaluateSLA))
	mux.HandleFunc("POST "+base+"/sla/breach-escalate", tenant(h.breachEscalate))

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

func requireTenant(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	tid, ok := TenantIDFromContext(r.Context())
	if !ok {
		writeErr(w, r, domain.ErrInvalidArgument)
		return uuid.Nil, false
	}
	return tid, true
}

func parsePathID(r *http.Request, key string) (uuid.UUID, error) {
	return uuid.Parse(r.PathValue(key))
}

func (h *Handler) createTicket(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		CustomerID     uuid.UUID `json:"customerId"`
		Subject        string    `json:"subject"`
		Description    string    `json:"description"`
		Priority       string    `json:"priority"`
		Category       string    `json:"category"`
		IdempotencyKey string    `json:"idempotencyKey"`
		Tags           []string  `json:"tags"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	if key := r.Header.Get("Idempotency-Key"); key != "" && body.IdempotencyKey == "" {
		body.IdempotencyKey = key
	}
	var actor *uuid.UUID
	if uid, ok := UserIDFromContext(r.Context()); ok {
		actor = &uid
	}
	t, err := h.Deps.CreateTicket(r.Context(), app.CreateTicketInput{
		TenantID: tid, CustomerID: body.CustomerID, Subject: body.Subject,
		Description: body.Description, Priority: body.Priority, Category: body.Category,
		IdempotencyKey: body.IdempotencyKey, Tags: body.Tags, ActorID: actor,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, ticketDTO(t))
}

func (h *Handler) listTickets(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	if h.Deps == nil || h.Deps.Tickets == nil {
		writeOK(w, map[string]any{"items": []any{}, "total": 0})
		return
	}
	list, err := h.Deps.Tickets.ListOpen(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	items := make([]any, 0, len(list))
	for _, t := range list {
		items = append(items, ticketDTO(t))
	}
	writeOK(w, map[string]any{"items": items, "total": len(items)})
}

func (h *Handler) getTicket(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := parsePathID(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	t, err := h.Deps.Tickets.Get(r.Context(), tid, id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, ticketDTO(t))
}

func (h *Handler) assignTicket(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := parsePathID(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		AgentID uuid.UUID  `json:"agentId"`
		TeamID  *uuid.UUID `json:"teamId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	t, err := h.Deps.AssignTicket(r.Context(), app.AssignTicketInput{
		TenantID: tid, TicketID: id, AgentID: body.AgentID, TeamID: body.TeamID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, t)
}

func (h *Handler) addNote(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := parsePathID(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Body string `json:"body"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	author, _ := UserIDFromContext(r.Context())
	n, err := h.Deps.AddNote(r.Context(), app.AddNoteInput{
		TenantID: tid, TicketID: id, AuthorID: author, Body: body.Body,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, n)
}

func (h *Handler) escalateTicket(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := parsePathID(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Reason string `json:"reason"`
	}
	_ = decodeJSON(r, &body)
	t, esc, err := h.Deps.Escalate(r.Context(), app.EscalateTicketInput{
		TenantID: tid, TicketID: id, Reason: body.Reason,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"ticket": t, "escalation": esc})
}

func (h *Handler) resolveTicket(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := parsePathID(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Note string `json:"note"`
	}
	_ = decodeJSON(r, &body)
	t, err := h.Deps.Resolve(r.Context(), app.ResolveTicketInput{TenantID: tid, TicketID: id, Note: body.Note})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, t)
}

func (h *Handler) closeTicket(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := parsePathID(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	t, err := h.Deps.Close(r.Context(), app.CloseTicketInput{TenantID: tid, TicketID: id})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, t)
}

func (h *Handler) reopenTicket(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := parsePathID(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Reason string `json:"reason"`
	}
	_ = decodeJSON(r, &body)
	t, err := h.Deps.Reopen(r.Context(), app.ReopenTicketInput{TenantID: tid, TicketID: id, Reason: body.Reason})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, t)
}

func (h *Handler) mergeTickets(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		SourceTicketID uuid.UUID `json:"sourceTicketId"`
		TargetTicketID uuid.UUID `json:"targetTicketId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	t, err := h.Deps.MergeTickets(r.Context(), app.MergeTicketsInput{
		TenantID: tid, SourceTicketID: body.SourceTicketID, TargetTicketID: body.TargetTicketID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, t)
}

func (h *Handler) startChat(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		CustomerID uuid.UUID  `json:"customerId"`
		Channel    string     `json:"channel"`
		TicketID   *uuid.UUID `json:"ticketId"`
		AgentID    *uuid.UUID `json:"agentId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	c, err := h.Deps.StartChat(r.Context(), app.StartChatInput{
		TenantID: tid, CustomerID: body.CustomerID, Channel: body.Channel,
		TicketID: body.TicketID, AgentID: body.AgentID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, c)
}

func (h *Handler) postMessage(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := parsePathID(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		SenderRole string     `json:"senderRole"`
		SenderID   *uuid.UUID `json:"senderId"`
		Body       string     `json:"body"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	msg, conv, err := h.Deps.PostMessage(r.Context(), app.PostMessageInput{
		TenantID: tid, ConversationID: id, SenderRole: body.SenderRole,
		SenderID: body.SenderID, Body: body.Body,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, map[string]any{"message": msg, "conversation": conv})
}

func (h *Handler) transferChat(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := parsePathID(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		ToAgentID uuid.UUID `json:"toAgentId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	c, err := h.Deps.TransferChat(r.Context(), app.TransferChatInput{
		TenantID: tid, ConversationID: id, ToAgentID: body.ToAgentID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, c)
}

func (h *Handler) endChat(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := parsePathID(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	c, err := h.Deps.EndChat(r.Context(), app.EndChatInput{TenantID: tid, ConversationID: id})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, c)
}

func (h *Handler) aiAssist(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		ConversationID uuid.UUID `json:"conversationId"`
		CustomerID     uuid.UUID `json:"customerId"`
		Text           string    `json:"text"`
		AutoEscalate   bool      `json:"autoEscalate"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.AIAssist(r.Context(), app.AIAssistInput{
		TenantID: tid, ConversationID: body.ConversationID, CustomerID: body.CustomerID,
		Text: body.Text, AutoEscalate: body.AutoEscalate,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, res)
}

func (h *Handler) upsertArticle(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		ArticleID uuid.UUID `json:"articleId"`
		Slug      string    `json:"slug"`
		Title     string    `json:"title"`
		Body      string    `json:"body"`
		Locale    string    `json:"locale"`
		Tags      []string  `json:"tags"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	a, err := h.Deps.UpsertArticle(r.Context(), app.UpsertArticleInput{
		TenantID: tid, ArticleID: body.ArticleID, Slug: body.Slug,
		Title: body.Title, Body: body.Body, Locale: body.Locale, Tags: body.Tags,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, a)
}

func (h *Handler) publishArticle(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := parsePathID(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	a, err := h.Deps.PublishArticle(r.Context(), tid, id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, a)
}

func (h *Handler) searchKB(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	q := r.URL.Query().Get("q")
	hits, err := h.Deps.SearchKB(r.Context(), tid, q)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"articles": hits})
}

func (h *Handler) createCase(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		CustomerID uuid.UUID  `json:"customerId"`
		TicketID   *uuid.UUID `json:"ticketId"`
		Type       string     `json:"type"`
		Title      string     `json:"title"`
		Details    string     `json:"details"`
		AssigneeID *uuid.UUID `json:"assigneeId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	c, err := h.Deps.CreateCase(r.Context(), app.CreateCaseInput{
		TenantID: tid, CustomerID: body.CustomerID, TicketID: body.TicketID,
		Type: body.Type, Title: body.Title, Details: body.Details, AssigneeID: body.AssigneeID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, c)
}

func (h *Handler) updateCase(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := parsePathID(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Status     string     `json:"status"`
		Title      string     `json:"title"`
		Details    string     `json:"details"`
		AssigneeID *uuid.UUID `json:"assigneeId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	c, err := h.Deps.UpdateCase(r.Context(), app.UpdateCaseInput{
		TenantID: tid, CaseID: id, Status: body.Status, Title: body.Title,
		Details: body.Details, AssigneeID: body.AssigneeID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, c)
}

func (h *Handler) submitCSAT(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		CustomerID     uuid.UUID  `json:"customerId"`
		TicketID       *uuid.UUID `json:"ticketId"`
		ConversationID *uuid.UUID `json:"conversationId"`
		Score          int        `json:"score"`
		Comment        string     `json:"comment"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	c, err := h.Deps.SubmitCSAT(r.Context(), app.SubmitCSATInput{
		TenantID: tid, CustomerID: body.CustomerID, TicketID: body.TicketID,
		ConversationID: body.ConversationID, Score: body.Score, Comment: body.Comment,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, c)
}

func (h *Handler) submitNPS(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		CustomerID uuid.UUID  `json:"customerId"`
		TicketID   *uuid.UUID `json:"ticketId"`
		Score      int        `json:"score"`
		Comment    string     `json:"comment"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	f, err := h.Deps.SubmitNPS(r.Context(), app.SubmitNPSInput{
		TenantID: tid, CustomerID: body.CustomerID, TicketID: body.TicketID,
		Score: body.Score, Comment: body.Comment,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, f)
}

func (h *Handler) customer360(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	cid, err := parsePathID(r, "customerId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	view, err := h.Deps.GetCustomer360(r.Context(), tid, cid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, view)
}

func (h *Handler) upsertSLA(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		Name                 string `json:"name"`
		Priority             string `json:"priority"`
		FirstResponseMinutes int    `json:"firstResponseMinutes"`
		ResolveMinutes       int    `json:"resolveMinutes"`
		Active               bool   `json:"active"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	p, err := h.Deps.UpsertSLAPolicy(r.Context(), app.UpsertSLAPolicyInput{
		TenantID: tid, Name: body.Name, Priority: body.Priority,
		FirstResponseMinutes: body.FirstResponseMinutes, ResolveMinutes: body.ResolveMinutes,
		Active: body.Active,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, p)
}

func (h *Handler) evaluateSLA(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		Escalate bool `json:"escalate"`
	}
	_ = decodeJSON(r, &body)
	breached, err := h.Deps.EvaluateSLA(r.Context(), tid, body.Escalate)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"breached": breached})
}

func (h *Handler) breachEscalate(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	breached, err := h.Deps.BreachEscalation(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"breached": breached})
}

func (h *Handler) adminStats(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	stats, err := h.Deps.AdminStats(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, stats)
}

func (h *Handler) publishOutbox(w http.ResponseWriter, r *http.Request) {
	n, err := h.Deps.PublishPending(r.Context(), 100)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"published": n})
}

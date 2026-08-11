package memory

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/crm-service/internal/app/ports"
	"github.com/nexora/crm-service/internal/domain"
)

// TicketRepo is an in-memory TicketRepo.
type TicketRepo struct{ S *Store }

func (r *TicketRepo) Save(_ context.Context, t domain.Ticket) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Tickets[t.ID] = t
	if t.IdempotencyKey != "" {
		r.S.TicketIdemKey[tenantKey(t.TenantID, t.IdempotencyKey)] = t.ID
	}
	return nil
}

func (r *TicketRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.Ticket, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	t, ok := r.S.Tickets[id]
	if !ok || t.TenantID != tenantID {
		return domain.Ticket{}, domain.ErrNotFound
	}
	return t, nil
}

func (r *TicketRepo) GetByIdempotencyKey(_ context.Context, tenantID uuid.UUID, key string) (domain.Ticket, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.TicketIdemKey[tenantKey(tenantID, key)]
	if !ok {
		return domain.Ticket{}, domain.ErrNotFound
	}
	t, ok := r.S.Tickets[id]
	if !ok || t.TenantID != tenantID {
		return domain.Ticket{}, domain.ErrNotFound
	}
	return t, nil
}

func (r *TicketRepo) ListByCustomer(_ context.Context, tenantID, customerID uuid.UUID) ([]domain.Ticket, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.Ticket
	for _, t := range r.S.Tickets {
		if t.TenantID == tenantID && t.CustomerID == customerID {
			out = append(out, t)
		}
	}
	return out, nil
}

func (r *TicketRepo) AddEvent(_ context.Context, e domain.TicketEvent) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.TicketEvents = append(r.S.TicketEvents, e)
	return nil
}

func (r *TicketRepo) ListEvents(_ context.Context, tenantID, ticketID uuid.UUID) ([]domain.TicketEvent, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.TicketEvent
	for _, e := range r.S.TicketEvents {
		if e.TenantID == tenantID && e.TicketID == ticketID {
			out = append(out, e)
		}
	}
	return out, nil
}

func (r *TicketRepo) AddNote(_ context.Context, n domain.TicketNote) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.TicketNotes = append(r.S.TicketNotes, n)
	return nil
}

func (r *TicketRepo) ListNotes(_ context.Context, tenantID, ticketID uuid.UUID) ([]domain.TicketNote, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.TicketNote
	for _, n := range r.S.TicketNotes {
		if n.TenantID == tenantID && n.TicketID == ticketID {
			out = append(out, n)
		}
	}
	return out, nil
}

func (r *TicketRepo) ListOpen(_ context.Context, tenantID uuid.UUID) ([]domain.Ticket, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.Ticket
	for _, t := range r.S.Tickets {
		if t.TenantID == tenantID && t.Status != domain.TicketStatusClosed {
			out = append(out, t)
		}
	}
	return out, nil
}

var _ ports.TicketRepo = (*TicketRepo)(nil)

// ChatRepo is an in-memory ChatRepo.
type ChatRepo struct{ S *Store }

func (r *ChatRepo) SaveConversation(_ context.Context, c domain.Conversation) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Conversations[c.ID] = c
	return nil
}

func (r *ChatRepo) GetConversation(_ context.Context, tenantID, id uuid.UUID) (domain.Conversation, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	c, ok := r.S.Conversations[id]
	if !ok || c.TenantID != tenantID {
		return domain.Conversation{}, domain.ErrNotFound
	}
	return c, nil
}

func (r *ChatRepo) AddMessage(_ context.Context, m domain.Message) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Messages = append(r.S.Messages, m)
	return nil
}

func (r *ChatRepo) ListMessages(_ context.Context, tenantID, conversationID uuid.UUID) ([]domain.Message, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.Message
	for _, m := range r.S.Messages {
		if m.TenantID == tenantID && m.ConversationID == conversationID {
			out = append(out, m)
		}
	}
	return out, nil
}

var _ ports.ChatRepo = (*ChatRepo)(nil)

// AgentRepo is an in-memory AgentRepo.
type AgentRepo struct{ S *Store }

func (r *AgentRepo) SaveAgent(_ context.Context, a domain.Agent) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Agents[a.ID] = a
	return nil
}

func (r *AgentRepo) GetAgent(_ context.Context, tenantID, id uuid.UUID) (domain.Agent, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	a, ok := r.S.Agents[id]
	if !ok || a.TenantID != tenantID {
		return domain.Agent{}, domain.ErrNotFound
	}
	return a, nil
}

func (r *AgentRepo) SaveTeam(_ context.Context, t domain.Team) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Teams[t.ID] = t
	return nil
}

func (r *AgentRepo) GetTeam(_ context.Context, tenantID, id uuid.UUID) (domain.Team, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	t, ok := r.S.Teams[id]
	if !ok || t.TenantID != tenantID {
		return domain.Team{}, domain.ErrNotFound
	}
	return t, nil
}

func (r *AgentRepo) SaveSkill(_ context.Context, s domain.Skill) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Skills[s.ID] = s
	return nil
}

func (r *AgentRepo) ListAgents(_ context.Context, tenantID uuid.UUID) ([]domain.Agent, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.Agent
	for _, a := range r.S.Agents {
		if a.TenantID == tenantID {
			out = append(out, a)
		}
	}
	return out, nil
}

var _ ports.AgentRepo = (*AgentRepo)(nil)

// KBRepo is an in-memory KBRepo.
type KBRepo struct{ S *Store }

func (r *KBRepo) SaveArticle(_ context.Context, a domain.Article) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Articles[a.ID] = a
	if a.Slug != "" {
		r.S.ArticleSlug[tenantKey(a.TenantID, a.Slug)] = a.ID
	}
	return nil
}

func (r *KBRepo) GetArticle(_ context.Context, tenantID, id uuid.UUID) (domain.Article, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	a, ok := r.S.Articles[id]
	if !ok || a.TenantID != tenantID {
		return domain.Article{}, domain.ErrNotFound
	}
	return a, nil
}

func (r *KBRepo) GetBySlug(_ context.Context, tenantID uuid.UUID, slug string) (domain.Article, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.ArticleSlug[tenantKey(tenantID, slug)]
	if !ok {
		return domain.Article{}, domain.ErrNotFound
	}
	a, ok := r.S.Articles[id]
	if !ok || a.TenantID != tenantID {
		return domain.Article{}, domain.ErrNotFound
	}
	return a, nil
}

func (r *KBRepo) Search(_ context.Context, tenantID uuid.UUID, query string) ([]domain.Article, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	q := strings.ToLower(strings.TrimSpace(query))
	var out []domain.Article
	for _, a := range r.S.Articles {
		if a.TenantID != tenantID || a.Status != domain.ArticleStatusPublished {
			continue
		}
		if q == "" || strings.Contains(strings.ToLower(a.Title), q) ||
			strings.Contains(strings.ToLower(a.Body), q) ||
			strings.Contains(strings.ToLower(a.Slug), q) {
			out = append(out, a)
			continue
		}
		for _, tag := range a.Tags {
			if strings.Contains(strings.ToLower(tag), q) {
				out = append(out, a)
				break
			}
		}
	}
	return out, nil
}

func (r *KBRepo) SaveVersion(_ context.Context, v domain.ArticleVersion) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.ArticleVers = append(r.S.ArticleVers, v)
	return nil
}

var _ ports.KBRepo = (*KBRepo)(nil)

// CaseRepo is an in-memory CaseRepo.
type CaseRepo struct{ S *Store }

func (r *CaseRepo) Save(_ context.Context, c domain.Case) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Cases[c.ID] = c
	return nil
}

func (r *CaseRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.Case, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	c, ok := r.S.Cases[id]
	if !ok || c.TenantID != tenantID {
		return domain.Case{}, domain.ErrNotFound
	}
	return c, nil
}

func (r *CaseRepo) ListByCustomer(_ context.Context, tenantID, customerID uuid.UUID) ([]domain.Case, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.Case
	for _, c := range r.S.Cases {
		if c.TenantID == tenantID && c.CustomerID == customerID {
			out = append(out, c)
		}
	}
	return out, nil
}

var _ ports.CaseRepo = (*CaseRepo)(nil)

// FeedbackRepo is an in-memory FeedbackRepo.
type FeedbackRepo struct{ S *Store }

func (r *FeedbackRepo) SaveFeedback(_ context.Context, f domain.Feedback) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Feedback = append(r.S.Feedback, f)
	return nil
}

func (r *FeedbackRepo) SaveCSAT(_ context.Context, c domain.CSATResponse) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.CSAT = append(r.S.CSAT, c)
	return nil
}

func (r *FeedbackRepo) ListCSAT(_ context.Context, tenantID uuid.UUID) ([]domain.CSATResponse, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.CSATResponse
	for _, c := range r.S.CSAT {
		if c.TenantID == tenantID {
			out = append(out, c)
		}
	}
	return out, nil
}

func (r *FeedbackRepo) ListFeedback(_ context.Context, tenantID uuid.UUID) ([]domain.Feedback, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.Feedback
	for _, f := range r.S.Feedback {
		if f.TenantID == tenantID {
			out = append(out, f)
		}
	}
	return out, nil
}

var _ ports.FeedbackRepo = (*FeedbackRepo)(nil)

// SLARepo is an in-memory SLARepo.
type SLARepo struct{ S *Store }

func (r *SLARepo) SavePolicy(_ context.Context, p domain.SLAPolicy) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.SLAPolicies[p.ID] = p
	r.S.PolicyByPrio[tenantKey(p.TenantID, p.Priority)] = p.ID
	return nil
}

func (r *SLARepo) GetPolicyByPriority(_ context.Context, tenantID uuid.UUID, priority string) (domain.SLAPolicy, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.PolicyByPrio[tenantKey(tenantID, priority)]
	if !ok {
		return domain.SLAPolicy{}, domain.ErrNotFound
	}
	p, ok := r.S.SLAPolicies[id]
	if !ok || p.TenantID != tenantID {
		return domain.SLAPolicy{}, domain.ErrNotFound
	}
	return p, nil
}

func (r *SLARepo) ListPolicies(_ context.Context, tenantID uuid.UUID) ([]domain.SLAPolicy, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.SLAPolicy
	for _, p := range r.S.SLAPolicies {
		if p.TenantID == tenantID {
			out = append(out, p)
		}
	}
	return out, nil
}

func (r *SLARepo) SaveEscalation(_ context.Context, e domain.Escalation) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Escalations = append(r.S.Escalations, e)
	return nil
}

func (r *SLARepo) ListEscalations(_ context.Context, tenantID, ticketID uuid.UUID) ([]domain.Escalation, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.Escalation
	for _, e := range r.S.Escalations {
		if e.TenantID == tenantID && e.TicketID == ticketID {
			out = append(out, e)
		}
	}
	return out, nil
}

var _ ports.SLARepo = (*SLARepo)(nil)

// OutboxRepo is an in-memory OutboxRepository.
type OutboxRepo struct{ S *Store }

func (r *OutboxRepo) Enqueue(_ context.Context, m domain.OutboxMessage) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Outbox = append(r.S.Outbox, m)
	return nil
}

func (r *OutboxRepo) ListPending(_ context.Context, limit int) ([]domain.OutboxMessage, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.OutboxMessage
	for _, m := range r.S.Outbox {
		if m.Status == domain.OutboxStatusPending {
			out = append(out, m)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (r *OutboxRepo) Update(_ context.Context, m domain.OutboxMessage) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	for i := range r.S.Outbox {
		if r.S.Outbox[i].ID == m.ID {
			r.S.Outbox[i] = m
			return nil
		}
	}
	return domain.ErrNotFound
}

var _ ports.OutboxRepository = (*OutboxRepo)(nil)

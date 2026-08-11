package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/crm-service/internal/domain"
)

// Clock provides the current time.
type Clock interface {
	Now() time.Time
}

// IDGen generates UUIDs.
type IDGen interface {
	New() uuid.UUID
}

// EventPublisher publishes domain events.
type EventPublisher interface {
	Publish(ctx context.Context, topic, key string, payload any) error
}

// TicketRepo persists tickets, events, and notes.
type TicketRepo interface {
	Save(ctx context.Context, t domain.Ticket) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Ticket, error)
	GetByIdempotencyKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.Ticket, error)
	ListByCustomer(ctx context.Context, tenantID, customerID uuid.UUID) ([]domain.Ticket, error)
	AddEvent(ctx context.Context, e domain.TicketEvent) error
	ListEvents(ctx context.Context, tenantID, ticketID uuid.UUID) ([]domain.TicketEvent, error)
	AddNote(ctx context.Context, n domain.TicketNote) error
	ListNotes(ctx context.Context, tenantID, ticketID uuid.UUID) ([]domain.TicketNote, error)
	ListOpen(ctx context.Context, tenantID uuid.UUID) ([]domain.Ticket, error)
}

// ChatRepo persists conversations and messages.
type ChatRepo interface {
	SaveConversation(ctx context.Context, c domain.Conversation) error
	GetConversation(ctx context.Context, tenantID, id uuid.UUID) (domain.Conversation, error)
	AddMessage(ctx context.Context, m domain.Message) error
	ListMessages(ctx context.Context, tenantID, conversationID uuid.UUID) ([]domain.Message, error)
}

// AgentRepo persists agents, teams, and skills.
type AgentRepo interface {
	SaveAgent(ctx context.Context, a domain.Agent) error
	GetAgent(ctx context.Context, tenantID, id uuid.UUID) (domain.Agent, error)
	SaveTeam(ctx context.Context, t domain.Team) error
	GetTeam(ctx context.Context, tenantID, id uuid.UUID) (domain.Team, error)
	SaveSkill(ctx context.Context, s domain.Skill) error
	ListAgents(ctx context.Context, tenantID uuid.UUID) ([]domain.Agent, error)
}

// KBRepo persists knowledge-base articles and versions.
type KBRepo interface {
	SaveArticle(ctx context.Context, a domain.Article) error
	GetArticle(ctx context.Context, tenantID, id uuid.UUID) (domain.Article, error)
	GetBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (domain.Article, error)
	Search(ctx context.Context, tenantID uuid.UUID, query string) ([]domain.Article, error)
	SaveVersion(ctx context.Context, v domain.ArticleVersion) error
}

// CaseRepo persists investigation cases.
type CaseRepo interface {
	Save(ctx context.Context, c domain.Case) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Case, error)
	ListByCustomer(ctx context.Context, tenantID, customerID uuid.UUID) ([]domain.Case, error)
}

// FeedbackRepo persists CSAT/NPS/CES responses.
type FeedbackRepo interface {
	SaveFeedback(ctx context.Context, f domain.Feedback) error
	SaveCSAT(ctx context.Context, c domain.CSATResponse) error
	ListCSAT(ctx context.Context, tenantID uuid.UUID) ([]domain.CSATResponse, error)
	ListFeedback(ctx context.Context, tenantID uuid.UUID) ([]domain.Feedback, error)
}

// SLARepo persists SLA policies and escalations.
type SLARepo interface {
	SavePolicy(ctx context.Context, p domain.SLAPolicy) error
	GetPolicyByPriority(ctx context.Context, tenantID uuid.UUID, priority string) (domain.SLAPolicy, error)
	ListPolicies(ctx context.Context, tenantID uuid.UUID) ([]domain.SLAPolicy, error)
	SaveEscalation(ctx context.Context, e domain.Escalation) error
	ListEscalations(ctx context.Context, tenantID, ticketID uuid.UUID) ([]domain.Escalation, error)
}

// OutboxRepository persists outbox messages.
type OutboxRepository interface {
	Enqueue(ctx context.Context, m domain.OutboxMessage) error
	ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
	Update(ctx context.Context, m domain.OutboxMessage) error
}

// ProfileSummary is an opaque customer profile read model.
type ProfileSummary struct {
	CustomerID uuid.UUID      `json:"customerId"`
	DisplayName string        `json:"displayName"`
	Email      string         `json:"email"`
	Phone      string         `json:"phone"`
	Tier       string         `json:"tier"`
	Extra      map[string]any `json:"extra,omitempty"`
}

// OrderSummary is an opaque order read model.
type OrderSummary struct {
	OrderID   uuid.UUID `json:"orderId"`
	Status    string    `json:"status"`
	Total     string    `json:"total"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"createdAt"`
}

// ProfileReadClient reads customer profile SoT (stub).
type ProfileReadClient interface {
	GetProfile(ctx context.Context, tenantID, customerID uuid.UUID) (ProfileSummary, error)
}

// OrderReadClient reads order summaries (stub).
type OrderReadClient interface {
	ListOrders(ctx context.Context, tenantID, customerID uuid.UUID, limit int) ([]OrderSummary, error)
}

// NotificationClient requests notifications without owning delivery.
type NotificationClient interface {
	Notify(ctx context.Context, tenantID, principalID uuid.UUID, templateKey string, data map[string]any) error
}

// RefundRequest is a refund request payload (execution owned elsewhere).
type RefundRequest struct {
	TenantID   uuid.UUID
	CustomerID uuid.UUID
	OrderID    uuid.UUID
	TicketID   uuid.UUID
	Amount     string
	Currency   string
	Reason     string
}

// RefundRequestClient submits refund requests to OMS/payment ports.
type RefundRequestClient interface {
	RequestRefund(ctx context.Context, req RefundRequest) (string, error)
}

// IntentResult is LLM intent classification.
type IntentResult struct {
	Intent     string
	Confidence float64
}

// ReplyResult is an LLM draft reply.
type ReplyResult struct {
	Reply      string
	Confidence float64
	Sources    []string
}

// LLMClient is the AI assistant port (mockable).
type LLMClient interface {
	DetectIntent(ctx context.Context, text string) (IntentResult, error)
	DraftReply(ctx context.Context, text string, kbSnippets []string) (ReplyResult, error)
	AnalyzeSentiment(ctx context.Context, text string) (string, error)
	Summarize(ctx context.Context, text string) (string, error)
}

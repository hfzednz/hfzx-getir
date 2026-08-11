package memory

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/crm-service/internal/domain"
)

// Store is the shared in-memory state.
type Store struct {
	mu sync.RWMutex

	Tickets        map[uuid.UUID]domain.Ticket
	TicketIdemKey  map[string]uuid.UUID // tenant|key -> id
	TicketEvents   []domain.TicketEvent
	TicketNotes    []domain.TicketNote
	Conversations  map[uuid.UUID]domain.Conversation
	Messages       []domain.Message
	Agents         map[uuid.UUID]domain.Agent
	Teams          map[uuid.UUID]domain.Team
	Skills         map[uuid.UUID]domain.Skill
	Articles       map[uuid.UUID]domain.Article
	ArticleSlug    map[string]uuid.UUID // tenant|slug
	ArticleVers    []domain.ArticleVersion
	Cases          map[uuid.UUID]domain.Case
	Feedback       []domain.Feedback
	CSAT           []domain.CSATResponse
	SLAPolicies    map[uuid.UUID]domain.SLAPolicy
	PolicyByPrio   map[string]uuid.UUID // tenant|priority
	Escalations    []domain.Escalation
	Attachments    []domain.AttachmentMeta
	Outbox         []domain.OutboxMessage
	PublishedEvents []map[string]any
}

// NewStore creates an empty store.
func NewStore() *Store {
	return &Store{
		Tickets:       make(map[uuid.UUID]domain.Ticket),
		TicketIdemKey: make(map[string]uuid.UUID),
		Conversations: make(map[uuid.UUID]domain.Conversation),
		Agents:        make(map[uuid.UUID]domain.Agent),
		Teams:         make(map[uuid.UUID]domain.Team),
		Skills:        make(map[uuid.UUID]domain.Skill),
		Articles:      make(map[uuid.UUID]domain.Article),
		ArticleSlug:   make(map[string]uuid.UUID),
		Cases:         make(map[uuid.UUID]domain.Case),
		SLAPolicies:   make(map[uuid.UUID]domain.SLAPolicy),
		PolicyByPrio:  make(map[string]uuid.UUID),
	}
}

func tenantKey(tenantID uuid.UUID, parts ...string) string {
	k := tenantID.String()
	for _, p := range parts {
		k += "|" + p
	}
	return k
}

// Clock is a mutable test clock.
type Clock struct{ T time.Time }

// Now returns the fixed/mutated time.
func (c *Clock) Now() time.Time {
	if c.T.IsZero() {
		return time.Now().UTC()
	}
	return c.T.UTC()
}

// Advance moves the clock forward.
func (c *Clock) Advance(d time.Duration) { c.T = c.T.Add(d) }

// IDGen wraps uuid.New.
type IDGen struct{}

// New returns a random UUID.
func (IDGen) New() uuid.UUID { return uuid.New() }

// Repos bundles memory repositories.
type Repos struct {
	Tickets  *TicketRepo
	Chats    *ChatRepo
	Agents   *AgentRepo
	KB       *KBRepo
	Cases    *CaseRepo
	Feedback *FeedbackRepo
	SLA      *SLARepo
	Outbox   *OutboxRepo
}

// NewRepos wires repos to a shared store.
func NewRepos(s *Store) *Repos {
	return &Repos{
		Tickets:  &TicketRepo{S: s},
		Chats:    &ChatRepo{S: s},
		Agents:   &AgentRepo{S: s},
		KB:       &KBRepo{S: s},
		Cases:    &CaseRepo{S: s},
		Feedback: &FeedbackRepo{S: s},
		SLA:      &SLARepo{S: s},
		Outbox:   &OutboxRepo{S: s},
	}
}

package postgres

import "database/sql"

// Repos groups CRM persistence adapters.
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

// NewRepos constructs postgres-backed repositories.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Tickets:  &TicketRepo{DB: db},
		Chats:    &ChatRepo{DB: db},
		Agents:   &AgentRepo{DB: db},
		KB:       &KBRepo{DB: db},
		Cases:    &CaseRepo{DB: db},
		Feedback: &FeedbackRepo{DB: db},
		SLA:      &SLARepo{DB: db},
		Outbox:   &OutboxRepo{DB: db},
	}
}

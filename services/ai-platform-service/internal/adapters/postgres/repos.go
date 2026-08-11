package postgres

import "database/sql"

// Repos groups ai-platform persistence adapters.
type Repos struct {
	Features   *FeatureRepo
	Models     *ModelRepo
	Prompts    *PromptRepo
	Memory     *MemoryRepo
	Agents     *AgentRepo
	Automation *AutomationRepo
	Drift      *DriftRepo
	Outbox     *OutboxRepo
}

// NewRepos constructs postgres-backed repositories.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Features:   &FeatureRepo{DB: db},
		Models:     &ModelRepo{DB: db},
		Prompts:    &PromptRepo{DB: db},
		Memory:     &MemoryRepo{DB: db},
		Agents:     &AgentRepo{DB: db},
		Automation: &AutomationRepo{DB: db},
		Drift:      &DriftRepo{DB: db},
		Outbox:     &OutboxRepo{DB: db},
	}
}

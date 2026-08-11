package postgres

import "database/sql"

// Repos groups enterprise-ops persistence adapters.
type Repos struct {
	Org        *OrgRepo
	Policies   *PolicyRepo
	Portfolios *PortfolioRepo
	Programs   *ProgramRepo
	Projects   *ProjectRepo
	Milestones *MilestoneRepo
	Objectives *ObjectiveRepo
	KeyResults *KeyResultRepo
	KPIs       *KPIRepo
	Risks      *RiskRepo
	Continuity *ContinuityRepo
	Audits     *AuditRepo
	Findings   *FindingRepo
	Meetings   *MeetingRepo
	Decisions  *DecisionRepo
	Knowledge  *KnowledgeRepo
	Resources  *ResourceRepo
	Outbox     *OutboxRepo
}

// NewRepos constructs postgres-backed repositories.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Org: &OrgRepo{DB: db}, Policies: &PolicyRepo{DB: db}, Portfolios: &PortfolioRepo{DB: db},
		Programs: &ProgramRepo{DB: db}, Projects: &ProjectRepo{DB: db}, Milestones: &MilestoneRepo{DB: db},
		Objectives: &ObjectiveRepo{DB: db}, KeyResults: &KeyResultRepo{DB: db}, KPIs: &KPIRepo{DB: db},
		Risks: &RiskRepo{DB: db}, Continuity: &ContinuityRepo{DB: db}, Audits: &AuditRepo{DB: db},
		Findings: &FindingRepo{DB: db}, Meetings: &MeetingRepo{DB: db}, Decisions: &DecisionRepo{DB: db},
		Knowledge: &KnowledgeRepo{DB: db}, Resources: &ResourceRepo{DB: db}, Outbox: &OutboxRepo{DB: db},
	}
}

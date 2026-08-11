package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/crm-service/internal/app/ports"
	"github.com/nexora/crm-service/internal/domain"
)

// AgentRepo persists agents, teams, and skills.
type AgentRepo struct{ DB *sql.DB }

func (r *AgentRepo) SaveAgent(ctx context.Context, a domain.Agent) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO agents (id, tenant_id, user_id, display_name, team_id, status, skill_ids, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET
			user_id=EXCLUDED.user_id,
			display_name=EXCLUDED.display_name,
			team_id=EXCLUDED.team_id,
			status=EXCLUDED.status,
			skill_ids=EXCLUDED.skill_ids,
			updated_at=EXCLUDED.updated_at`,
		a.ID, a.TenantID, a.UserID, a.DisplayName, nullUUID(a.TeamID), a.Status,
		UUIDArray(a.SkillIDs), a.CreatedAt.UTC(), a.UpdatedAt.UTC())
	return err
}

func (r *AgentRepo) GetAgent(ctx context.Context, tenantID, id uuid.UUID) (domain.Agent, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, user_id, display_name, team_id, status, skill_ids, created_at, updated_at
		FROM agents WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var a domain.Agent
	var team uuid.NullUUID
	var skills UUIDArray
	err := row.Scan(&a.ID, &a.TenantID, &a.UserID, &a.DisplayName, &team, &a.Status, &skills, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return domain.Agent{}, mapNotFound(err)
	}
	a.TeamID = scanNullUUID(team)
	a.SkillIDs = []uuid.UUID(skills)
	a.CreatedAt = a.CreatedAt.UTC()
	a.UpdatedAt = a.UpdatedAt.UTC()
	return a, nil
}

func (r *AgentRepo) SaveTeam(ctx context.Context, t domain.Team) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO teams (id, tenant_id, name, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (id) DO UPDATE SET
			name=EXCLUDED.name,
			updated_at=EXCLUDED.updated_at`,
		t.ID, t.TenantID, t.Name, t.CreatedAt.UTC(), t.UpdatedAt.UTC())
	return err
}

func (r *AgentRepo) GetTeam(ctx context.Context, tenantID, id uuid.UUID) (domain.Team, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, created_at, updated_at FROM teams WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var t domain.Team
	err := row.Scan(&t.ID, &t.TenantID, &t.Name, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return domain.Team{}, mapNotFound(err)
	}
	t.CreatedAt = t.CreatedAt.UTC()
	t.UpdatedAt = t.UpdatedAt.UTC()
	return t, nil
}

func (r *AgentRepo) SaveSkill(ctx context.Context, s domain.Skill) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO skills (id, tenant_id, name, created_at)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name`,
		s.ID, s.TenantID, s.Name, s.CreatedAt.UTC())
	return err
}

func (r *AgentRepo) ListAgents(ctx context.Context, tenantID uuid.UUID) ([]domain.Agent, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, user_id, display_name, team_id, status, skill_ids, created_at, updated_at
		FROM agents WHERE tenant_id=$1 ORDER BY display_name ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Agent{}
	for rows.Next() {
		var a domain.Agent
		var team uuid.NullUUID
		var skills UUIDArray
		if err := rows.Scan(&a.ID, &a.TenantID, &a.UserID, &a.DisplayName, &team, &a.Status, &skills, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		a.TeamID = scanNullUUID(team)
		a.SkillIDs = []uuid.UUID(skills)
		a.CreatedAt = a.CreatedAt.UTC()
		a.UpdatedAt = a.UpdatedAt.UTC()
		out = append(out, a)
	}
	return out, rows.Err()
}

var _ ports.AgentRepo = (*AgentRepo)(nil)

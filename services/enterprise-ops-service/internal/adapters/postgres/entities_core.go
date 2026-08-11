package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/enterprise-ops-service/internal/app/ports"
	"github.com/nexora/enterprise-ops-service/internal/domain"
)

type scannable interface{ Scan(dest ...any) error }

type OrgRepo struct{ DB *sql.DB }

func (r *OrgRepo) Save(ctx context.Context, n domain.OrgNode) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO eo_org_nodes (
			id, tenant_id, kind, code, name, parent_id, manager_ref, country_code, active, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (tenant_id, code) DO UPDATE SET
			id=EXCLUDED.id, kind=EXCLUDED.kind, name=EXCLUDED.name, parent_id=EXCLUDED.parent_id,
			manager_ref=EXCLUDED.manager_ref, country_code=EXCLUDED.country_code, active=EXCLUDED.active,
			updated_at=EXCLUDED.updated_at`,
		n.ID, n.TenantID, n.Kind, n.Code, n.Name, nullUUID(n.ParentID), n.ManagerRef, n.CountryCode,
		n.Active, n.CreatedAt.UTC(), n.UpdatedAt.UTC())
	return err
}

func (r *OrgRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.OrgNode, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, kind, code, name, parent_id, manager_ref, country_code, active, created_at, updated_at
		FROM eo_org_nodes WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	return scanOrg(row)
}

func (r *OrgRepo) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.OrgNode, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, kind, code, name, parent_id, manager_ref, country_code, active, created_at, updated_at
		FROM eo_org_nodes WHERE tenant_id=$1 AND code=$2`, tenantID, code)
	return scanOrg(row)
}

func (r *OrgRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.OrgNode, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, kind, code, name, parent_id, manager_ref, country_code, active, created_at, updated_at
		FROM eo_org_nodes WHERE tenant_id=$1 ORDER BY code ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.OrgNode{}
	for rows.Next() {
		n, err := scanOrg(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (r *OrgRepo) ListChildren(ctx context.Context, tenantID, parentID uuid.UUID) ([]domain.OrgNode, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, kind, code, name, parent_id, manager_ref, country_code, active, created_at, updated_at
		FROM eo_org_nodes WHERE tenant_id=$1 AND parent_id=$2 ORDER BY code ASC`, tenantID, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.OrgNode{}
	for rows.Next() {
		n, err := scanOrg(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func scanOrg(row scannable) (domain.OrgNode, error) {
	var n domain.OrgNode
	var parent uuid.NullUUID
	err := row.Scan(&n.ID, &n.TenantID, &n.Kind, &n.Code, &n.Name, &parent, &n.ManagerRef, &n.CountryCode,
		&n.Active, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.OrgNode{}, domain.ErrNotFound
		}
		return domain.OrgNode{}, err
	}
	n.ParentID = scanNullUUID(parent)
	n.CreatedAt = n.CreatedAt.UTC()
	n.UpdatedAt = n.UpdatedAt.UTC()
	return n, nil
}

type PolicyRepo struct{ DB *sql.DB }

func (r *PolicyRepo) Save(ctx context.Context, p domain.Policy) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO eo_policies (
			id, tenant_id, key, title, kind, status, version, body_uri, owner_ref, approved_by, approved_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		ON CONFLICT (tenant_id, key) DO UPDATE SET
			id=EXCLUDED.id, title=EXCLUDED.title, kind=EXCLUDED.kind, status=EXCLUDED.status,
			version=EXCLUDED.version, body_uri=EXCLUDED.body_uri, owner_ref=EXCLUDED.owner_ref,
			approved_by=EXCLUDED.approved_by, approved_at=EXCLUDED.approved_at, updated_at=EXCLUDED.updated_at`,
		p.ID, p.TenantID, p.Key, p.Title, p.Kind, p.Status, p.Version, p.BodyURI, p.OwnerRef,
		p.ApprovedBy, nullTime(p.ApprovedAt), p.CreatedAt.UTC(), p.UpdatedAt.UTC())
	return err
}

func (r *PolicyRepo) GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.Policy, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, key, title, kind, status, version, body_uri, owner_ref, approved_by, approved_at, created_at, updated_at
		FROM eo_policies WHERE tenant_id=$1 AND key=$2`, tenantID, key)
	return scanPolicy(row)
}

func (r *PolicyRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.Policy, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, key, title, kind, status, version, body_uri, owner_ref, approved_by, approved_at, created_at, updated_at
		FROM eo_policies WHERE tenant_id=$1 ORDER BY key ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Policy{}
	for rows.Next() {
		p, err := scanPolicy(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func scanPolicy(row scannable) (domain.Policy, error) {
	var p domain.Policy
	var approved sql.NullTime
	err := row.Scan(&p.ID, &p.TenantID, &p.Key, &p.Title, &p.Kind, &p.Status, &p.Version, &p.BodyURI,
		&p.OwnerRef, &p.ApprovedBy, &approved, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.Policy{}, domain.ErrNotFound
		}
		return domain.Policy{}, err
	}
	p.ApprovedAt = scanNullTime(approved)
	p.CreatedAt = p.CreatedAt.UTC()
	p.UpdatedAt = p.UpdatedAt.UTC()
	return p, nil
}

type PortfolioRepo struct{ DB *sql.DB }

func (r *PortfolioRepo) Save(ctx context.Context, p domain.Portfolio) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO eo_portfolios (id, tenant_id, code, name, owner_ref, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (id) DO UPDATE SET code=EXCLUDED.code, name=EXCLUDED.name, owner_ref=EXCLUDED.owner_ref`,
		p.ID, p.TenantID, p.Code, p.Name, p.OwnerRef, p.CreatedAt.UTC())
	return err
}

func (r *PortfolioRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Portfolio, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, code, name, owner_ref, created_at
		FROM eo_portfolios WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var p domain.Portfolio
	err := row.Scan(&p.ID, &p.TenantID, &p.Code, &p.Name, &p.OwnerRef, &p.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.Portfolio{}, domain.ErrNotFound
		}
		return domain.Portfolio{}, err
	}
	p.CreatedAt = p.CreatedAt.UTC()
	return p, nil
}

func (r *PortfolioRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.Portfolio, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, code, name, owner_ref, created_at
		FROM eo_portfolios WHERE tenant_id=$1 ORDER BY code ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Portfolio{}
	for rows.Next() {
		var p domain.Portfolio
		if err := rows.Scan(&p.ID, &p.TenantID, &p.Code, &p.Name, &p.OwnerRef, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.CreatedAt = p.CreatedAt.UTC()
		out = append(out, p)
	}
	return out, rows.Err()
}

type ProgramRepo struct{ DB *sql.DB }

func (r *ProgramRepo) Save(ctx context.Context, p domain.Program) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO eo_programs (id, tenant_id, portfolio_id, code, name, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (id) DO UPDATE SET portfolio_id=EXCLUDED.portfolio_id, code=EXCLUDED.code, name=EXCLUDED.name`,
		p.ID, p.TenantID, p.PortfolioID, p.Code, p.Name, p.CreatedAt.UTC())
	return err
}

func (r *ProgramRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Program, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, portfolio_id, code, name, created_at
		FROM eo_programs WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var p domain.Program
	err := row.Scan(&p.ID, &p.TenantID, &p.PortfolioID, &p.Code, &p.Name, &p.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.Program{}, domain.ErrNotFound
		}
		return domain.Program{}, err
	}
	p.CreatedAt = p.CreatedAt.UTC()
	return p, nil
}

func (r *ProgramRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.Program, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, portfolio_id, code, name, created_at
		FROM eo_programs WHERE tenant_id=$1 ORDER BY code ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Program{}
	for rows.Next() {
		var p domain.Program
		if err := rows.Scan(&p.ID, &p.TenantID, &p.PortfolioID, &p.Code, &p.Name, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.CreatedAt = p.CreatedAt.UTC()
		out = append(out, p)
	}
	return out, rows.Err()
}

type ProjectRepo struct{ DB *sql.DB }

func (r *ProjectRepo) Save(ctx context.Context, p domain.Project) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO eo_projects (
			id, tenant_id, program_id, code, name, status, budget_minor, currency, health, owner_ref, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (id) DO UPDATE SET
			program_id=EXCLUDED.program_id, code=EXCLUDED.code, name=EXCLUDED.name, status=EXCLUDED.status,
			budget_minor=EXCLUDED.budget_minor, currency=EXCLUDED.currency, health=EXCLUDED.health,
			owner_ref=EXCLUDED.owner_ref, updated_at=EXCLUDED.updated_at`,
		p.ID, p.TenantID, p.ProgramID, p.Code, p.Name, p.Status, p.BudgetMinor, p.Currency, p.Health,
		p.OwnerRef, p.CreatedAt.UTC(), p.UpdatedAt.UTC())
	return err
}

func (r *ProjectRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Project, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, program_id, code, name, status, budget_minor, currency, health, owner_ref, created_at, updated_at
		FROM eo_projects WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var p domain.Project
	err := row.Scan(&p.ID, &p.TenantID, &p.ProgramID, &p.Code, &p.Name, &p.Status, &p.BudgetMinor,
		&p.Currency, &p.Health, &p.OwnerRef, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.Project{}, domain.ErrNotFound
		}
		return domain.Project{}, err
	}
	p.CreatedAt = p.CreatedAt.UTC()
	p.UpdatedAt = p.UpdatedAt.UTC()
	return p, nil
}

func (r *ProjectRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.Project, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, program_id, code, name, status, budget_minor, currency, health, owner_ref, created_at, updated_at
		FROM eo_projects WHERE tenant_id=$1 ORDER BY code ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Project{}
	for rows.Next() {
		var p domain.Project
		if err := rows.Scan(&p.ID, &p.TenantID, &p.ProgramID, &p.Code, &p.Name, &p.Status, &p.BudgetMinor,
			&p.Currency, &p.Health, &p.OwnerRef, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.CreatedAt = p.CreatedAt.UTC()
		p.UpdatedAt = p.UpdatedAt.UTC()
		out = append(out, p)
	}
	return out, rows.Err()
}

type MilestoneRepo struct{ DB *sql.DB }

func (r *MilestoneRepo) Save(ctx context.Context, m domain.Milestone) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO eo_milestones (id, tenant_id, project_id, name, due_at, done, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (id) DO UPDATE SET
			project_id=EXCLUDED.project_id, name=EXCLUDED.name, due_at=EXCLUDED.due_at, done=EXCLUDED.done`,
		m.ID, m.TenantID, m.ProjectID, m.Name, m.DueAt.UTC(), m.Done, m.CreatedAt.UTC())
	return err
}

func (r *MilestoneRepo) ListByProject(ctx context.Context, tenantID, projectID uuid.UUID) ([]domain.Milestone, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, project_id, name, due_at, done, created_at
		FROM eo_milestones WHERE tenant_id=$1 AND project_id=$2 ORDER BY due_at ASC`, tenantID, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Milestone{}
	for rows.Next() {
		var m domain.Milestone
		if err := rows.Scan(&m.ID, &m.TenantID, &m.ProjectID, &m.Name, &m.DueAt, &m.Done, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.DueAt = m.DueAt.UTC()
		m.CreatedAt = m.CreatedAt.UTC()
		out = append(out, m)
	}
	return out, rows.Err()
}

type ObjectiveRepo struct{ DB *sql.DB }

func (r *ObjectiveRepo) Save(ctx context.Context, o domain.Objective) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO eo_objectives (id, tenant_id, period, title, owner_ref, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (id) DO UPDATE SET period=EXCLUDED.period, title=EXCLUDED.title, owner_ref=EXCLUDED.owner_ref`,
		o.ID, o.TenantID, o.Period, o.Title, o.OwnerRef, o.CreatedAt.UTC())
	return err
}

func (r *ObjectiveRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.Objective, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, period, title, owner_ref, created_at
		FROM eo_objectives WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Objective{}
	for rows.Next() {
		var o domain.Objective
		if err := rows.Scan(&o.ID, &o.TenantID, &o.Period, &o.Title, &o.OwnerRef, &o.CreatedAt); err != nil {
			return nil, err
		}
		o.CreatedAt = o.CreatedAt.UTC()
		out = append(out, o)
	}
	return out, rows.Err()
}

type KeyResultRepo struct{ DB *sql.DB }

func (r *KeyResultRepo) Save(ctx context.Context, kr domain.KeyResult) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO eo_key_results (id, tenant_id, objective_id, title, target, current, unit, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET
			objective_id=EXCLUDED.objective_id, title=EXCLUDED.title, target=EXCLUDED.target,
			current=EXCLUDED.current, unit=EXCLUDED.unit, updated_at=EXCLUDED.updated_at`,
		kr.ID, kr.TenantID, kr.ObjectiveID, kr.Title, kr.Target, kr.Current, kr.Unit, kr.UpdatedAt.UTC())
	return err
}

func (r *KeyResultRepo) ListByObjective(ctx context.Context, tenantID, objectiveID uuid.UUID) ([]domain.KeyResult, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, objective_id, title, target, current, unit, updated_at
		FROM eo_key_results WHERE tenant_id=$1 AND objective_id=$2 ORDER BY title ASC`, tenantID, objectiveID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.KeyResult{}
	for rows.Next() {
		var kr domain.KeyResult
		if err := rows.Scan(&kr.ID, &kr.TenantID, &kr.ObjectiveID, &kr.Title, &kr.Target, &kr.Current, &kr.Unit, &kr.UpdatedAt); err != nil {
			return nil, err
		}
		kr.UpdatedAt = kr.UpdatedAt.UTC()
		out = append(out, kr)
	}
	return out, rows.Err()
}

var (
	_ ports.OrgRepo        = (*OrgRepo)(nil)
	_ ports.PolicyRepo     = (*PolicyRepo)(nil)
	_ ports.PortfolioRepo  = (*PortfolioRepo)(nil)
	_ ports.ProgramRepo    = (*ProgramRepo)(nil)
	_ ports.ProjectRepo    = (*ProjectRepo)(nil)
	_ ports.MilestoneRepo  = (*MilestoneRepo)(nil)
	_ ports.ObjectiveRepo  = (*ObjectiveRepo)(nil)
	_ ports.KeyResultRepo  = (*KeyResultRepo)(nil)
)

package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/nexora/innovation-service/internal/app/ports"
	"github.com/nexora/innovation-service/internal/domain"
)

type ModuleRepo struct{ DB *sql.DB }

func (r *ModuleRepo) Save(ctx context.Context, m domain.InnovationModule) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO inv_modules (
			id, tenant_id, key, name, domain, status, trl, score, sandbox_only, description, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (tenant_id, key) DO UPDATE SET
			id=EXCLUDED.id, name=EXCLUDED.name, domain=EXCLUDED.domain, status=EXCLUDED.status,
			trl=EXCLUDED.trl, score=EXCLUDED.score, sandbox_only=EXCLUDED.sandbox_only,
			description=EXCLUDED.description, updated_at=EXCLUDED.updated_at`,
		m.ID, m.TenantID, m.Key, m.Name, m.Domain, m.Status, int(m.TRL), m.Score, m.SandboxOnly,
		m.Description, m.CreatedAt.UTC(), m.UpdatedAt.UTC())
	return err
}

func (r *ModuleRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.InnovationModule, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, key, name, domain, status, trl, score, sandbox_only, description, created_at, updated_at
		FROM inv_modules WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	return scanModule(row)
}

func (r *ModuleRepo) GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.InnovationModule, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, key, name, domain, status, trl, score, sandbox_only, description, created_at, updated_at
		FROM inv_modules WHERE tenant_id=$1 AND key=$2`, tenantID, key)
	return scanModule(row)
}

func (r *ModuleRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.InnovationModule, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, key, name, domain, status, trl, score, sandbox_only, description, created_at, updated_at
		FROM inv_modules WHERE tenant_id=$1 ORDER BY key ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.InnovationModule{}
	for rows.Next() {
		m, err := scanModule(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

type scannable interface{ Scan(dest ...any) error }

func scanModule(row scannable) (domain.InnovationModule, error) {
	var m domain.InnovationModule
	var trl int
	err := row.Scan(&m.ID, &m.TenantID, &m.Key, &m.Name, &m.Domain, &m.Status, &trl, &m.Score,
		&m.SandboxOnly, &m.Description, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.InnovationModule{}, domain.ErrNotFound
		}
		return domain.InnovationModule{}, err
	}
	m.TRL = domain.TRL(trl)
	m.CreatedAt = m.CreatedAt.UTC()
	m.UpdatedAt = m.UpdatedAt.UTC()
	return m, nil
}

type ExperimentRepo struct{ DB *sql.DB }

func (r *ExperimentRepo) Save(ctx context.Context, e domain.ResearchExperiment) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO inv_experiments (
			id, tenant_id, module_id, name, hypothesis, status, created_at, completed_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET
			module_id=EXCLUDED.module_id, name=EXCLUDED.name, hypothesis=EXCLUDED.hypothesis,
			status=EXCLUDED.status, completed_at=EXCLUDED.completed_at`,
		e.ID, e.TenantID, e.ModuleID, e.Name, e.Hypothesis, e.Status, e.CreatedAt.UTC(), nullTime(e.CompletedAt))
	return err
}

func (r *ExperimentRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.ResearchExperiment, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, module_id, name, hypothesis, status, created_at, completed_at
		FROM inv_experiments WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ResearchExperiment{}
	for rows.Next() {
		var e domain.ResearchExperiment
		var completed sql.NullTime
		if err := rows.Scan(&e.ID, &e.TenantID, &e.ModuleID, &e.Name, &e.Hypothesis, &e.Status, &e.CreatedAt, &completed); err != nil {
			return nil, err
		}
		e.CompletedAt = scanNullTime(completed)
		e.CreatedAt = e.CreatedAt.UTC()
		out = append(out, e)
	}
	return out, rows.Err()
}

type SimulationRepo struct{ DB *sql.DB }

func (r *SimulationRepo) Save(ctx context.Context, s domain.SimulationRun) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO inv_simulations (
			id, tenant_id, kind, name, status, params, accuracy, result_summary, started_at, completed_at, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO UPDATE SET
			kind=EXCLUDED.kind, name=EXCLUDED.name, status=EXCLUDED.status, params=EXCLUDED.params,
			accuracy=EXCLUDED.accuracy, result_summary=EXCLUDED.result_summary,
			started_at=EXCLUDED.started_at, completed_at=EXCLUDED.completed_at`,
		s.ID, s.TenantID, s.Kind, s.Name, s.Status, JSONMap(s.Params), s.Accuracy, s.ResultSummary,
		nullTime(s.StartedAt), nullTime(s.CompletedAt), s.CreatedAt.UTC())
	return err
}

func (r *SimulationRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.SimulationRun, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, kind, name, status, params, accuracy, result_summary, started_at, completed_at, created_at
		FROM inv_simulations WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	return scanSimulation(row)
}

func (r *SimulationRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.SimulationRun, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, kind, name, status, params, accuracy, result_summary, started_at, completed_at, created_at
		FROM inv_simulations WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.SimulationRun{}
	for rows.Next() {
		s, err := scanSimulation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func scanSimulation(row scannable) (domain.SimulationRun, error) {
	var s domain.SimulationRun
	var params JSONMap
	var started, completed sql.NullTime
	err := row.Scan(&s.ID, &s.TenantID, &s.Kind, &s.Name, &s.Status, &params, &s.Accuracy, &s.ResultSummary,
		&started, &completed, &s.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.SimulationRun{}, domain.ErrNotFound
		}
		return domain.SimulationRun{}, err
	}
	s.Params = map[string]any(params)
	s.StartedAt = scanNullTime(started)
	s.CompletedAt = scanNullTime(completed)
	s.CreatedAt = s.CreatedAt.UTC()
	return s, nil
}

type TwinRepo struct{ DB *sql.DB }

func (r *TwinRepo) Save(ctx context.Context, t domain.DigitalTwin) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO inv_twins (
			id, tenant_id, kind, ref_key, name, model_uri, version, active, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (id) DO UPDATE SET
			kind=EXCLUDED.kind, ref_key=EXCLUDED.ref_key, name=EXCLUDED.name, model_uri=EXCLUDED.model_uri,
			version=EXCLUDED.version, active=EXCLUDED.active, updated_at=EXCLUDED.updated_at`,
		t.ID, t.TenantID, t.Kind, t.RefKey, t.Name, t.ModelURI, t.Version, t.Active,
		t.CreatedAt.UTC(), t.UpdatedAt.UTC())
	return err
}

func (r *TwinRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.DigitalTwin, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, kind, ref_key, name, model_uri, version, active, created_at, updated_at
		FROM inv_twins WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.DigitalTwin{}
	for rows.Next() {
		var t domain.DigitalTwin
		if err := rows.Scan(&t.ID, &t.TenantID, &t.Kind, &t.RefKey, &t.Name, &t.ModelURI, &t.Version,
			&t.Active, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		t.CreatedAt = t.CreatedAt.UTC()
		t.UpdatedAt = t.UpdatedAt.UTC()
		out = append(out, t)
	}
	return out, rows.Err()
}

type EdgeRepo struct{ DB *sql.DB }

func (r *EdgeRepo) Save(ctx context.Context, n domain.EdgeNode) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO inv_edge_nodes (id, tenant_id, key, region, capabilities, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (tenant_id, key) DO UPDATE SET
			id=EXCLUDED.id, region=EXCLUDED.region, capabilities=EXCLUDED.capabilities, status=EXCLUDED.status`,
		n.ID, n.TenantID, n.Key, n.Region, textArray(n.Caps), n.Status, n.CreatedAt.UTC())
	return err
}

func (r *EdgeRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.EdgeNode, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, key, region, capabilities, status, created_at
		FROM inv_edge_nodes WHERE tenant_id=$1 ORDER BY key ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.EdgeNode{}
	for rows.Next() {
		var n domain.EdgeNode
		var caps []string
		if err := rows.Scan(&n.ID, &n.TenantID, &n.Key, &n.Region, pq.Array(&caps), &n.Status, &n.CreatedAt); err != nil {
			return nil, err
		}
		n.Caps = caps
		n.CreatedAt = n.CreatedAt.UTC()
		out = append(out, n)
	}
	return out, rows.Err()
}

type IoTRepo struct{ DB *sql.DB }

func (r *IoTRepo) Save(ctx context.Context, d domain.IoTDevice) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO inv_iot_devices (
			id, tenant_id, device_key, kind, location, connected, last_seen_at, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (tenant_id, device_key) DO UPDATE SET
			id=EXCLUDED.id, kind=EXCLUDED.kind, location=EXCLUDED.location,
			connected=EXCLUDED.connected, last_seen_at=EXCLUDED.last_seen_at`,
		d.ID, d.TenantID, d.DeviceKey, d.Kind, d.Location, d.Connected, nullTime(d.LastSeenAt), d.CreatedAt.UTC())
	return err
}

func (r *IoTRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.IoTDevice, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, device_key, kind, location, connected, last_seen_at, created_at
		FROM inv_iot_devices WHERE tenant_id=$1 ORDER BY device_key ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.IoTDevice{}
	for rows.Next() {
		var d domain.IoTDevice
		var lastSeen sql.NullTime
		if err := rows.Scan(&d.ID, &d.TenantID, &d.DeviceKey, &d.Kind, &d.Location, &d.Connected, &lastSeen, &d.CreatedAt); err != nil {
			return nil, err
		}
		d.LastSeenAt = scanNullTime(lastSeen)
		d.CreatedAt = d.CreatedAt.UTC()
		out = append(out, d)
	}
	return out, rows.Err()
}

var (
	_ ports.ModuleRepo     = (*ModuleRepo)(nil)
	_ ports.ExperimentRepo = (*ExperimentRepo)(nil)
	_ ports.SimulationRepo = (*SimulationRepo)(nil)
	_ ports.TwinRepo       = (*TwinRepo)(nil)
	_ ports.EdgeRepo       = (*EdgeRepo)(nil)
	_ ports.IoTRepo        = (*IoTRepo)(nil)
)

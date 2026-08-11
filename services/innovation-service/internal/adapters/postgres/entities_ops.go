package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/nexora/innovation-service/internal/app/ports"
	"github.com/nexora/innovation-service/internal/domain"
)

type RobotRepo struct{ DB *sql.DB }

func (r *RobotRepo) Save(ctx context.Context, robot domain.Robot) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO inv_robots (id, tenant_id, key, kind, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (tenant_id, key) DO UPDATE SET
			id=EXCLUDED.id, kind=EXCLUDED.kind, status=EXCLUDED.status`,
		robot.ID, robot.TenantID, robot.Key, robot.Kind, robot.Status, robot.CreatedAt.UTC())
	return err
}

func (r *RobotRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Robot, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, key, kind, status, created_at
		FROM inv_robots WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var robot domain.Robot
	err := row.Scan(&robot.ID, &robot.TenantID, &robot.Key, &robot.Kind, &robot.Status, &robot.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.Robot{}, domain.ErrNotFound
		}
		return domain.Robot{}, err
	}
	robot.CreatedAt = robot.CreatedAt.UTC()
	return robot, nil
}

func (r *RobotRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.Robot, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, key, kind, status, created_at
		FROM inv_robots WHERE tenant_id=$1 ORDER BY key ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Robot{}
	for rows.Next() {
		var robot domain.Robot
		if err := rows.Scan(&robot.ID, &robot.TenantID, &robot.Key, &robot.Kind, &robot.Status, &robot.CreatedAt); err != nil {
			return nil, err
		}
		robot.CreatedAt = robot.CreatedAt.UTC()
		out = append(out, robot)
	}
	return out, rows.Err()
}

type AssignmentRepo struct{ DB *sql.DB }

func (r *AssignmentRepo) Save(ctx context.Context, a domain.RobotAssignment) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO inv_robot_assignments (id, tenant_id, robot_id, task_ref, status, assigned_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (id) DO UPDATE SET
			robot_id=EXCLUDED.robot_id, task_ref=EXCLUDED.task_ref, status=EXCLUDED.status, assigned_at=EXCLUDED.assigned_at`,
		a.ID, a.TenantID, a.RobotID, a.TaskRef, a.Status, a.AssignedAt.UTC())
	return err
}

func (r *AssignmentRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.RobotAssignment, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, robot_id, task_ref, status, assigned_at
		FROM inv_robot_assignments WHERE tenant_id=$1 ORDER BY assigned_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.RobotAssignment{}
	for rows.Next() {
		var a domain.RobotAssignment
		if err := rows.Scan(&a.ID, &a.TenantID, &a.RobotID, &a.TaskRef, &a.Status, &a.AssignedAt); err != nil {
			return nil, err
		}
		a.AssignedAt = a.AssignedAt.UTC()
		out = append(out, a)
	}
	return out, rows.Err()
}

type DroneRepo struct{ DB *sql.DB }

func (r *DroneRepo) Save(ctx context.Context, m domain.DroneMission) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO inv_drone_missions (
			id, tenant_id, drone_key, order_ref, landing_zone, status, compliance, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET
			drone_key=EXCLUDED.drone_key, order_ref=EXCLUDED.order_ref, landing_zone=EXCLUDED.landing_zone,
			status=EXCLUDED.status, compliance=EXCLUDED.compliance`,
		m.ID, m.TenantID, m.DroneKey, m.OrderRef, m.LandingZone, m.Status, textArray(m.Compliance), m.CreatedAt.UTC())
	return err
}

func (r *DroneRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.DroneMission, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, drone_key, order_ref, landing_zone, status, compliance, created_at
		FROM inv_drone_missions WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.DroneMission{}
	for rows.Next() {
		var m domain.DroneMission
		var compliance []string
		if err := rows.Scan(&m.ID, &m.TenantID, &m.DroneKey, &m.OrderRef, &m.LandingZone, &m.Status,
			pq.Array(&compliance), &m.CreatedAt); err != nil {
			return nil, err
		}
		m.Compliance = compliance
		m.CreatedAt = m.CreatedAt.UTC()
		out = append(out, m)
	}
	return out, rows.Err()
}

type BlockchainRepo struct{ DB *sql.DB }

func (r *BlockchainRepo) Save(ctx context.Context, h domain.BlockchainHook) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO inv_blockchain_hooks (id, tenant_id, purpose, chain_ref, active, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (id) DO UPDATE SET
			purpose=EXCLUDED.purpose, chain_ref=EXCLUDED.chain_ref, active=EXCLUDED.active`,
		h.ID, h.TenantID, h.Purpose, h.ChainRef, h.Active, h.CreatedAt.UTC())
	return err
}

func (r *BlockchainRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.BlockchainHook, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, purpose, chain_ref, active, created_at
		FROM inv_blockchain_hooks WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.BlockchainHook{}
	for rows.Next() {
		var h domain.BlockchainHook
		if err := rows.Scan(&h.ID, &h.TenantID, &h.Purpose, &h.ChainRef, &h.Active, &h.CreatedAt); err != nil {
			return nil, err
		}
		h.CreatedAt = h.CreatedAt.UTC()
		out = append(out, h)
	}
	return out, rows.Err()
}

type XRRepo struct{ DB *sql.DB }

func (r *XRRepo) Save(ctx context.Context, x domain.XRExperience) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO inv_xr (id, tenant_id, kind, asset_uri, active, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (id) DO UPDATE SET
			kind=EXCLUDED.kind, asset_uri=EXCLUDED.asset_uri, active=EXCLUDED.active`,
		x.ID, x.TenantID, x.Kind, x.AssetURI, x.Active, x.CreatedAt.UTC())
	return err
}

func (r *XRRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.XRExperience, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, kind, asset_uri, active, created_at
		FROM inv_xr WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.XRExperience{}
	for rows.Next() {
		var x domain.XRExperience
		if err := rows.Scan(&x.ID, &x.TenantID, &x.Kind, &x.AssetURI, &x.Active, &x.CreatedAt); err != nil {
			return nil, err
		}
		x.CreatedAt = x.CreatedAt.UTC()
		out = append(out, x)
	}
	return out, rows.Err()
}

type MultimodalRepo struct{ DB *sql.DB }

func (r *MultimodalRepo) Save(ctx context.Context, s domain.MultimodalSession) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO inv_multimodal (id, tenant_id, subject_id, modes, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (id) DO UPDATE SET
			subject_id=EXCLUDED.subject_id, modes=EXCLUDED.modes, status=EXCLUDED.status`,
		s.ID, s.TenantID, s.SubjectID, textArray(s.Modes), s.Status, s.CreatedAt.UTC())
	return err
}

func (r *MultimodalRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.MultimodalSession, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, subject_id, modes, status, created_at
		FROM inv_multimodal WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.MultimodalSession{}
	for rows.Next() {
		var s domain.MultimodalSession
		var modes []string
		if err := rows.Scan(&s.ID, &s.TenantID, &s.SubjectID, pq.Array(&modes), &s.Status, &s.CreatedAt); err != nil {
			return nil, err
		}
		s.Modes = modes
		s.CreatedAt = s.CreatedAt.UTC()
		out = append(out, s)
	}
	return out, rows.Err()
}

type GreenRepo struct{ DB *sql.DB }

func (r *GreenRepo) Save(ctx context.Context, g domain.GreenMetric) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO inv_green (
			id, tenant_id, period, carbon_grams, energy_wh, savings_percent, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (tenant_id, period) DO UPDATE SET
			id=EXCLUDED.id, carbon_grams=EXCLUDED.carbon_grams, energy_wh=EXCLUDED.energy_wh,
			savings_percent=EXCLUDED.savings_percent, updated_at=EXCLUDED.updated_at`,
		g.ID, g.TenantID, g.Period, g.CarbonGrams, g.EnergyWh, g.SavingsPercent, g.UpdatedAt.UTC())
	return err
}

func (r *GreenRepo) GetByPeriod(ctx context.Context, tenantID uuid.UUID, period string) (domain.GreenMetric, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, period, carbon_grams, energy_wh, savings_percent, updated_at
		FROM inv_green WHERE tenant_id=$1 AND period=$2`, tenantID, period)
	var g domain.GreenMetric
	err := row.Scan(&g.ID, &g.TenantID, &g.Period, &g.CarbonGrams, &g.EnergyWh, &g.SavingsPercent, &g.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.GreenMetric{}, domain.ErrNotFound
		}
		return domain.GreenMetric{}, err
	}
	g.UpdatedAt = g.UpdatedAt.UTC()
	return g, nil
}

func (r *GreenRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.GreenMetric, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, period, carbon_grams, energy_wh, savings_percent, updated_at
		FROM inv_green WHERE tenant_id=$1 ORDER BY period DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.GreenMetric{}
	for rows.Next() {
		var g domain.GreenMetric
		if err := rows.Scan(&g.ID, &g.TenantID, &g.Period, &g.CarbonGrams, &g.EnergyWh, &g.SavingsPercent, &g.UpdatedAt); err != nil {
			return nil, err
		}
		g.UpdatedAt = g.UpdatedAt.UTC()
		out = append(out, g)
	}
	return out, rows.Err()
}

type QuantumRepo struct{ DB *sql.DB }

func (r *QuantumRepo) Save(ctx context.Context, q domain.QuantumHook) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO inv_quantum (id, tenant_id, kind, adapter, active, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (id) DO UPDATE SET
			kind=EXCLUDED.kind, adapter=EXCLUDED.adapter, active=EXCLUDED.active`,
		q.ID, q.TenantID, q.Kind, q.Adapter, q.Active, q.CreatedAt.UTC())
	return err
}

func (r *QuantumRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.QuantumHook, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, kind, adapter, active, created_at
		FROM inv_quantum WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.QuantumHook{}
	for rows.Next() {
		var q domain.QuantumHook
		if err := rows.Scan(&q.ID, &q.TenantID, &q.Kind, &q.Adapter, &q.Active, &q.CreatedAt); err != nil {
			return nil, err
		}
		q.CreatedAt = q.CreatedAt.UTC()
		out = append(out, q)
	}
	return out, rows.Err()
}

var (
	_ ports.RobotRepo      = (*RobotRepo)(nil)
	_ ports.AssignmentRepo = (*AssignmentRepo)(nil)
	_ ports.DroneRepo      = (*DroneRepo)(nil)
	_ ports.BlockchainRepo = (*BlockchainRepo)(nil)
	_ ports.XRRepo         = (*XRRepo)(nil)
	_ ports.MultimodalRepo = (*MultimodalRepo)(nil)
	_ ports.GreenRepo      = (*GreenRepo)(nil)
	_ ports.QuantumRepo    = (*QuantumRepo)(nil)
)

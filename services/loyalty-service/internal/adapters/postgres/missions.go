package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/app/ports"
	"github.com/nexora/loyalty-service/internal/domain"
)

// MissionRepo persists missions and progress.
type MissionRepo struct{ DB *sql.DB }

func (r *MissionRepo) CreateMission(ctx context.Context, m domain.Mission) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO missions
		  (id, tenant_id, code, title, target_count, reward_points, achievement, active, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		m.ID, m.TenantID, m.Code, m.Title, m.TargetCount, m.RewardPoints, m.Achievement, m.Active, m.CreatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *MissionRepo) GetMission(ctx context.Context, tenantID, missionID uuid.UUID) (domain.Mission, error) {
	return r.scanMission(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, code, title, target_count, reward_points, achievement, active, created_at
		FROM missions WHERE id=$1 AND tenant_id=$2`, missionID, tenantID))
}

func (r *MissionRepo) GetMissionByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.Mission, error) {
	return r.scanMission(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, code, title, target_count, reward_points, achievement, active, created_at
		FROM missions WHERE tenant_id=$1 AND code=$2`, tenantID, code))
}

func (r *MissionRepo) GetProgress(ctx context.Context, tenantID, accountID, missionID uuid.UUID) (domain.MissionProgress, error) {
	var p domain.MissionProgress
	var status string
	var completed sql.NullTime
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, account_id, mission_id, progress, status, updated_at, completed_at
		FROM mission_progress WHERE tenant_id=$1 AND account_id=$2 AND mission_id=$3`,
		tenantID, accountID, missionID).Scan(
		&p.ID, &p.TenantID, &p.AccountID, &p.MissionID, &p.Progress, &status, &p.UpdatedAt, &completed)
	if isNoRows(err) {
		return domain.MissionProgress{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.MissionProgress{}, err
	}
	p.Status = domain.MissionStatus(status)
	p.UpdatedAt = p.UpdatedAt.UTC()
	p.CompletedAt = scanNullTime(completed)
	return p, nil
}

func (r *MissionRepo) UpsertProgress(ctx context.Context, p domain.MissionProgress) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO mission_progress
		  (id, tenant_id, account_id, mission_id, progress, status, updated_at, completed_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (tenant_id, account_id, mission_id) DO UPDATE SET
		  progress = EXCLUDED.progress,
		  status = EXCLUDED.status,
		  updated_at = EXCLUDED.updated_at,
		  completed_at = EXCLUDED.completed_at`,
		p.ID, p.TenantID, p.AccountID, p.MissionID, p.Progress, string(p.Status),
		p.UpdatedAt.UTC(), nullTime(p.CompletedAt))
	return err
}

func (r *MissionRepo) scanMission(row *sql.Row) (domain.Mission, error) {
	var m domain.Mission
	err := row.Scan(
		&m.ID, &m.TenantID, &m.Code, &m.Title, &m.TargetCount, &m.RewardPoints,
		&m.Achievement, &m.Active, &m.CreatedAt)
	if isNoRows(err) {
		return domain.Mission{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Mission{}, err
	}
	m.CreatedAt = m.CreatedAt.UTC()
	return m, nil
}

var _ ports.MissionRepo = (*MissionRepo)(nil)

package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/location-service/internal/app/ports"
	"github.com/nexora/location-service/internal/domain"
)

// HistoryRepo persists capped location history.
type HistoryRepo struct{ DB *sql.DB }

func (r *HistoryRepo) Ingest(ctx context.Context, h domain.LocationHistory) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO location_history (id, tenant_id, subject_type, subject_id, lat, lng, recorded_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		h.ID, h.TenantID, string(h.SubjectType), h.SubjectID, h.Lat, h.Lng, h.RecordedAt.UTC())
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		DELETE FROM location_history
		WHERE id IN (
			SELECT id FROM location_history
			WHERE tenant_id=$1 AND subject_type=$2::subject_type AND subject_id=$3
			ORDER BY recorded_at DESC
			OFFSET $4
		)`, h.TenantID, string(h.SubjectType), h.SubjectID, domain.MaxHistoryPerSubject)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *HistoryRepo) List(ctx context.Context, tenantID uuid.UUID, subjectType domain.SubjectType, subjectID string, limit int) ([]domain.LocationHistory, error) {
	if limit <= 0 {
		limit = domain.MaxHistoryPerSubject
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, subject_type, subject_id, lat, lng, recorded_at
		FROM location_history
		WHERE tenant_id=$1 AND subject_type=$2::subject_type AND subject_id=$3
		ORDER BY recorded_at DESC
		LIMIT $4`, tenantID, string(subjectType), subjectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.LocationHistory, 0)
	for rows.Next() {
		var h domain.LocationHistory
		var st string
		if err := rows.Scan(&h.ID, &h.TenantID, &st, &h.SubjectID, &h.Lat, &h.Lng, &h.RecordedAt); err != nil {
			return nil, err
		}
		h.SubjectType = domain.SubjectType(st)
		h.RecordedAt = h.RecordedAt.UTC()
		out = append(out, h)
	}
	return out, rows.Err()
}

var _ ports.HistoryRepo = (*HistoryRepo)(nil)

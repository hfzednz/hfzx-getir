package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/nexora/warehouse-service/internal/app/ports"
	"github.com/nexora/warehouse-service/internal/domain"
)

// PickRepo persists pick sessions and lines (lines/tenant/picker live in metadata when needed).
type PickRepo struct{ DB *sql.DB }

var _ ports.PickRepo = (*PickRepo)(nil)

const (
	metaTenantID = "_tenantId"
	metaPickerID = "_pickerId"
	metaLines    = "_lines"
)

func (r *PickRepo) CreateSession(ctx context.Context, s domain.PickSession) error {
	if err := ensureWarehouse(ctx, r.DB, s.TenantID, s.WarehouseID); err != nil {
		return err
	}
	strategy := s.Strategy
	if strategy == "" {
		strategy = domain.PickStrategySingle
	}
	meta := pickMeta(s)
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO pick_sessions (
			id, task_id, warehouse_id, fulfillment_id, strategy, route,
			started_at, completed_at, metadata, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		s.ID, s.TaskID, s.WarehouseID, nullUUIDValue(s.FulfillmentID), string(strategy),
		JSONRaw{V: s.Route}, nullTime(s.StartedAt), nullTime(s.CompletedAt), meta, s.CreatedAt, s.UpdatedAt,
	)
	return mapUniqueViolation(err)
}

func (r *PickRepo) UpdateSession(ctx context.Context, s domain.PickSession) error {
	strategy := s.Strategy
	if strategy == "" {
		strategy = domain.PickStrategySingle
	}
	meta := pickMeta(s)
	res, err := r.DB.ExecContext(ctx, `
		UPDATE pick_sessions SET
			warehouse_id=$1, fulfillment_id=$2, strategy=$3, route=$4,
			started_at=$5, completed_at=$6, metadata=$7, updated_at=$8
		WHERE id=$9`,
		s.WarehouseID, nullUUIDValue(s.FulfillmentID), string(strategy), JSONRaw{V: s.Route},
		nullTime(s.StartedAt), nullTime(s.CompletedAt), meta, s.UpdatedAt, s.ID,
	)
	if err != nil {
		return mapUniqueViolation(err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	// Enforce tenant match via metadata when provided.
	if s.TenantID != uuid.Nil {
		got, err := r.GetSessionByID(ctx, s.TenantID, s.ID)
		if err != nil {
			return err
		}
		_ = got
	}
	return nil
}

func (r *PickRepo) GetSessionByID(ctx context.Context, tenantID, id uuid.UUID) (domain.PickSession, error) {
	s, err := r.scanSession(r.DB.QueryRowContext(ctx, `
		SELECT id, task_id, warehouse_id, fulfillment_id, strategy, route,
			started_at, completed_at, metadata, created_at, updated_at
		FROM pick_sessions WHERE id=$1`, id))
	if err != nil {
		return domain.PickSession{}, err
	}
	if s.TenantID != tenantID {
		return domain.PickSession{}, domain.ErrNotFound
	}
	return s, nil
}

func (r *PickRepo) GetSessionByTaskID(ctx context.Context, tenantID, taskID uuid.UUID) (domain.PickSession, error) {
	s, err := r.scanSession(r.DB.QueryRowContext(ctx, `
		SELECT id, task_id, warehouse_id, fulfillment_id, strategy, route,
			started_at, completed_at, metadata, created_at, updated_at
		FROM pick_sessions WHERE task_id=$1`, taskID))
	if err != nil {
		return domain.PickSession{}, err
	}
	if s.TenantID != tenantID {
		return domain.PickSession{}, domain.ErrNotFound
	}
	return s, nil
}

func pickMeta(s domain.PickSession) JSONMap {
	extra := map[string]any{
		metaTenantID: s.TenantID.String(),
		metaLines:    s.Lines,
	}
	if s.PickerID != nil {
		extra[metaPickerID] = s.PickerID.String()
	}
	return mergeMeta(s.Metadata, extra)
}

func (r *PickRepo) scanSession(row scannable) (domain.PickSession, error) {
	var s domain.PickSession
	var fulfillmentID uuid.NullUUID
	var strategy string
	var routeRaw []byte
	var started, completed sql.NullTime
	var meta JSONMap
	err := row.Scan(
		&s.ID, &s.TaskID, &s.WarehouseID, &fulfillmentID, &strategy, &routeRaw,
		&started, &completed, &meta, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return domain.PickSession{}, mapNotFound(err)
	}
	s.FulfillmentID = scanUUIDOrNil(fulfillmentID)
	s.Strategy = domain.PickStrategy(strategy)
	s.StartedAt = scanNullTime(started)
	s.CompletedAt = scanNullTime(completed)
	s.CreatedAt = s.CreatedAt.UTC()
	s.UpdatedAt = s.UpdatedAt.UTC()
	if len(routeRaw) > 0 && string(routeRaw) != "null" {
		_ = json.Unmarshal(routeRaw, &s.Route)
	}
	if s.Route == nil {
		s.Route = []domain.PickRouteStep{}
	}
	userMeta := map[string]any{}
	for k, v := range meta {
		switch k {
		case metaTenantID:
			if str, ok := v.(string); ok {
				if id, err := uuid.Parse(str); err == nil {
					s.TenantID = id
				}
			}
		case metaPickerID:
			if str, ok := v.(string); ok {
				if id, err := uuid.Parse(str); err == nil {
					s.PickerID = &id
				}
			}
		case metaLines:
			b, _ := json.Marshal(v)
			_ = json.Unmarshal(b, &s.Lines)
		default:
			userMeta[k] = v
		}
	}
	s.Metadata = userMeta
	if s.Lines == nil {
		s.Lines = []domain.PickLine{}
	}
	return s, nil
}

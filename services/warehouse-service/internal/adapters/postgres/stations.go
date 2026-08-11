package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/warehouse-service/internal/app/ports"
	"github.com/nexora/warehouse-service/internal/domain"
)

const metaClaimedBy = "_claimedBy"

// StationRepo persists stations.
type StationRepo struct{ DB *sql.DB }

var _ ports.StationRepo = (*StationRepo)(nil)

func (r *StationRepo) Create(ctx context.Context, s domain.Station) error {
	if err := ensureWarehouse(ctx, r.DB, s.TenantID, s.WarehouseID); err != nil {
		return err
	}
	status := s.Status
	if status == "" {
		status = domain.StationStatusAvailable
	}
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO stations (
			id, warehouse_id, code, type, status, name, zone_code, metadata, created_at, updated_at, deleted_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		s.ID, s.WarehouseID, s.Code, string(s.Type), string(status), s.Name, zoneCode(s),
		stationMeta(s), s.CreatedAt, s.UpdatedAt, nullTime(s.DeletedAt),
	)
	return mapUniqueViolation(err)
}

func (r *StationRepo) Update(ctx context.Context, s domain.Station) error {
	status := s.Status
	if status == "" {
		status = domain.StationStatusAvailable
	}
	res, err := r.DB.ExecContext(ctx, `
		UPDATE stations SET
			code=$1, type=$2, status=$3, name=$4, zone_code=$5, metadata=$6, updated_at=$7, deleted_at=$8
		WHERE id=$9 AND warehouse_id=$10`,
		s.Code, string(s.Type), string(status), s.Name, zoneCode(s), stationMeta(s), s.UpdatedAt, nullTime(s.DeletedAt),
		s.ID, s.WarehouseID,
	)
	return rowsAffectedOrNotFound(res, err)
}

func (r *StationRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Station, error) {
	return r.scanStation(r.DB.QueryRowContext(ctx, `
		SELECT s.id, w.tenant_id, s.warehouse_id, s.code, s.type, s.status, s.name, s.zone_code,
			s.metadata, s.created_at, s.updated_at, s.deleted_at
		FROM stations s
		JOIN warehouses w ON w.id = s.warehouse_id
		WHERE s.id=$1 AND w.tenant_id=$2`, id, tenantID))
}

func (r *StationRepo) ListByWarehouse(ctx context.Context, tenantID, warehouseID uuid.UUID) ([]domain.Station, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT s.id, w.tenant_id, s.warehouse_id, s.code, s.type, s.status, s.name, s.zone_code,
			s.metadata, s.created_at, s.updated_at, s.deleted_at
		FROM stations s
		JOIN warehouses w ON w.id = s.warehouse_id
		WHERE s.warehouse_id=$1 AND w.tenant_id=$2 AND s.deleted_at IS NULL
		ORDER BY s.code`, warehouseID, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Station, 0)
	for rows.Next() {
		s, err := scanStationRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func stationMeta(s domain.Station) JSONMap {
	extra := map[string]any{}
	if s.ClaimedBy != nil {
		extra[metaClaimedBy] = s.ClaimedBy.String()
	}
	return mergeMeta(s.Metadata, extra)
}

func (r *StationRepo) scanStation(row scannable) (domain.Station, error) {
	s, err := scanStationRow(row)
	if err != nil {
		return domain.Station{}, mapNotFound(err)
	}
	return s, nil
}

func scanStationRow(row scannable) (domain.Station, error) {
	var s domain.Station
	var typ, status string
	var meta JSONMap
	var deleted sql.NullTime
	err := row.Scan(
		&s.ID, &s.TenantID, &s.WarehouseID, &s.Code, &typ, &status, &s.Name, &s.ZoneCode,
		&meta, &s.CreatedAt, &s.UpdatedAt, &deleted,
	)
	if err != nil {
		return domain.Station{}, err
	}
	s.Type = domain.StationType(typ)
	s.Status = domain.StationStatus(status)
	s.Zone = s.ZoneCode
	s.DeletedAt = scanNullTime(deleted)
	s.CreatedAt = s.CreatedAt.UTC()
	s.UpdatedAt = s.UpdatedAt.UTC()
	userMeta := map[string]any{}
	for k, v := range meta {
		if k == metaClaimedBy {
			if str, ok := v.(string); ok {
				if id, err := uuid.Parse(str); err == nil {
					s.ClaimedBy = &id
				}
			}
			continue
		}
		userMeta[k] = v
	}
	s.Metadata = userMeta
	return s, nil
}

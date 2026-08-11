package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/inventory-service/internal/app/ports"
	"github.com/nexora/inventory-service/internal/domain"
)

type WarehouseRepo struct{ DB *sql.DB }

func (r *WarehouseRepo) Create(ctx context.Context, w domain.Warehouse) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO warehouses (id, tenant_id, code, name, region_id, timezone, status, metadata, created_at, updated_at, deleted_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7::warehouse_status,$8,$9,$10,$11)`,
		w.ID, w.TenantID, w.Code, w.Name, nullUUID(w.RegionID), w.Timezone, string(w.Status), JSONMap(w.Metadata),
		w.CreatedAt, w.UpdatedAt, nullTime(w.DeletedAt))
	return mapUniqueViolation(err)
}

func (r *WarehouseRepo) Update(ctx context.Context, w domain.Warehouse) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE warehouses SET code=$2, name=$3, region_id=$4, timezone=$5, status=$6::warehouse_status,
			metadata=$7, updated_at=$8, deleted_at=$9
		WHERE id=$1 AND tenant_id=$10`,
		w.ID, w.Code, w.Name, nullUUID(w.RegionID), w.Timezone, string(w.Status),
		JSONMap(w.Metadata), w.UpdatedAt, nullTime(w.DeletedAt), w.TenantID)
	if err != nil {
		return mapUniqueViolation(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *WarehouseRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Warehouse, error) {
	w, err := scanWarehouse(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, code, name, region_id, timezone, status::text, metadata, created_at, updated_at, deleted_at
		FROM warehouses WHERE id=$1 AND tenant_id=$2 AND deleted_at IS NULL`, id, tenantID))
	if err != nil {
		return domain.Warehouse{}, mapNotFound(err)
	}
	return w, nil
}

func (r *WarehouseRepo) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.Warehouse, error) {
	w, err := scanWarehouse(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, code, name, region_id, timezone, status::text, metadata, created_at, updated_at, deleted_at
		FROM warehouses WHERE tenant_id=$1 AND code=$2 AND deleted_at IS NULL`, tenantID, code))
	if err != nil {
		return domain.Warehouse{}, mapNotFound(err)
	}
	return w, nil
}

func (r *WarehouseRepo) List(ctx context.Context, f ports.WarehouseFilter) ([]domain.Warehouse, int, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	args := []any{f.TenantID}
	where := `tenant_id=$1 AND deleted_at IS NULL`
	if f.Status != nil {
		args = append(args, string(*f.Status))
		where += ` AND status=$` + itoa(len(args)) + `::warehouse_status`
	}
	if f.Query != "" {
		args = append(args, "%"+f.Query+"%")
		n := itoa(len(args))
		where += ` AND (code ILIKE $` + n + ` OR name ILIKE $` + n + `)`
	}
	var total int
	if err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM warehouses WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, limit, f.Offset)
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, code, name, region_id, timezone, status::text, metadata, created_at, updated_at, deleted_at
		FROM warehouses WHERE `+where+` ORDER BY code ASC LIMIT $`+itoa(len(args)-1)+` OFFSET $`+itoa(len(args)), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.Warehouse{}
	for rows.Next() {
		w, err := scanWarehouse(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, w)
	}
	return out, total, rows.Err()
}

func (r *WarehouseRepo) Delete(ctx context.Context, tenantID, id uuid.UUID, at time.Time) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE warehouses SET status='closed'::warehouse_status, deleted_at=$3, updated_at=$3
		WHERE id=$1 AND tenant_id=$2 AND deleted_at IS NULL`, id, tenantID, at)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

type warehouseScanner interface {
	Scan(dest ...any) error
}

func scanWarehouse(s warehouseScanner) (domain.Warehouse, error) {
	var w domain.Warehouse
	var region uuid.NullUUID
	var status string
	var meta JSONMap
	var deleted sql.NullTime
	err := s.Scan(&w.ID, &w.TenantID, &w.Code, &w.Name, &region, &w.Timezone, &status, &meta, &w.CreatedAt, &w.UpdatedAt, &deleted)
	if err != nil {
		return domain.Warehouse{}, err
	}
	w.RegionID = scanNullUUID(region)
	w.Status = domain.WarehouseStatus(status)
	w.Metadata = map[string]any(meta)
	w.DeletedAt = scanNullTime(deleted)
	return w, nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

var _ ports.WarehouseRepository = (*WarehouseRepo)(nil)

type LocationRepo struct{ DB *sql.DB }

func (r *LocationRepo) Create(ctx context.Context, l domain.Location) error {
	var zone any
	if l.ZoneType != nil {
		zone = string(*l.ZoneType)
	}
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO locations (
			id, warehouse_id, parent_id, kind, zone_type, code, path, depth, name, is_pickable, is_active,
			metadata, created_at, updated_at, deleted_at
		) VALUES ($1,$2,$3,$4::location_kind,$5::zone_type,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		l.ID, l.WarehouseID, nullUUID(l.ParentID), string(l.Kind), zone, l.Code, l.Path, l.Depth, l.Name, l.IsPickable, l.IsActive,
		JSONMap(l.Metadata), l.CreatedAt, l.UpdatedAt, nullTime(l.DeletedAt))
	return mapUniqueViolation(err)
}

func (r *LocationRepo) Update(ctx context.Context, l domain.Location) error {
	var zone any
	if l.ZoneType != nil {
		zone = string(*l.ZoneType)
	}
	res, err := r.DB.ExecContext(ctx, `
		UPDATE locations SET parent_id=$2, kind=$3::location_kind, zone_type=$4::zone_type, code=$5, path=$6, depth=$7,
			name=$8, is_pickable=$9, is_active=$10, metadata=$11, updated_at=$12, deleted_at=$13
		WHERE id=$1`,
		l.ID, nullUUID(l.ParentID), string(l.Kind), zone, l.Code, l.Path, l.Depth,
		l.Name, l.IsPickable, l.IsActive, JSONMap(l.Metadata), l.UpdatedAt, nullTime(l.DeletedAt))
	if err != nil {
		return mapUniqueViolation(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *LocationRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Location, error) {
	l, err := scanLocation(r.DB.QueryRowContext(ctx, `
		SELECT id, warehouse_id, parent_id, kind::text, zone_type::text, code, path, depth, name, is_pickable, is_active,
			metadata, created_at, updated_at, deleted_at
		FROM locations WHERE id=$1 AND deleted_at IS NULL`, id))
	if err != nil {
		return domain.Location{}, mapNotFound(err)
	}
	return l, nil
}

func (r *LocationRepo) ListByWarehouse(ctx context.Context, warehouseID uuid.UUID) ([]domain.Location, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, warehouse_id, parent_id, kind::text, zone_type::text, code, path, depth, name, is_pickable, is_active,
			metadata, created_at, updated_at, deleted_at
		FROM locations WHERE warehouse_id=$1 AND deleted_at IS NULL ORDER BY path`, warehouseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Location{}
	for rows.Next() {
		l, err := scanLocation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *LocationRepo) ListChildren(ctx context.Context, parentID uuid.UUID) ([]domain.Location, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, warehouse_id, parent_id, kind::text, zone_type::text, code, path, depth, name, is_pickable, is_active,
			metadata, created_at, updated_at, deleted_at
		FROM locations WHERE parent_id=$1 AND deleted_at IS NULL ORDER BY code`, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Location{}
	for rows.Next() {
		l, err := scanLocation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *LocationRepo) Delete(ctx context.Context, id uuid.UUID, at time.Time) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE locations SET deleted_at=$2, is_active=FALSE, updated_at=$2 WHERE id=$1 AND deleted_at IS NULL`, id, at)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

type locationScanner interface {
	Scan(dest ...any) error
}

func scanLocation(s locationScanner) (domain.Location, error) {
	var l domain.Location
	var parent uuid.NullUUID
	var kind string
	var zone sql.NullString
	var meta JSONMap
	var deleted sql.NullTime
	err := s.Scan(&l.ID, &l.WarehouseID, &parent, &kind, &zone, &l.Code, &l.Path, &l.Depth, &l.Name, &l.IsPickable, &l.IsActive,
		&meta, &l.CreatedAt, &l.UpdatedAt, &deleted)
	if err != nil {
		return domain.Location{}, err
	}
	l.ParentID = scanNullUUID(parent)
	l.Kind = domain.LocationKind(kind)
	if zone.Valid && zone.String != "" {
		z := domain.ZoneType(zone.String)
		l.ZoneType = &z
	}
	l.Metadata = map[string]any(meta)
	l.DeletedAt = scanNullTime(deleted)
	return l, nil
}

var _ ports.LocationRepository = (*LocationRepo)(nil)

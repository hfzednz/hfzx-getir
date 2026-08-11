package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/location-service/internal/app/ports"
	"github.com/nexora/location-service/internal/domain"
)

// POIRepo persists spatial POI index entries.
type POIRepo struct{ DB *sql.DB }

func (r *POIRepo) Upsert(ctx context.Context, p domain.POI) error {
	meta := JSONMap(p.Meta)
	if meta == nil {
		meta = JSONMap{}
	}
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO pois
		  (id, tenant_id, kind, ref_id, name, lat, lng, meta, active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO UPDATE SET
		  tenant_id=EXCLUDED.tenant_id, kind=EXCLUDED.kind, ref_id=EXCLUDED.ref_id,
		  name=EXCLUDED.name, lat=EXCLUDED.lat, lng=EXCLUDED.lng, meta=EXCLUDED.meta,
		  active=EXCLUDED.active, updated_at=EXCLUDED.updated_at`,
		p.ID, p.TenantID, string(p.Kind), p.RefID, p.Name, p.Lat, p.Lng, meta, p.Active,
		p.CreatedAt.UTC(), p.UpdatedAt.UTC())
	return err
}

func (r *POIRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.POI, error) {
	p, err := r.scanOne(ctx, `
		SELECT id, tenant_id, kind, ref_id, name, lat, lng, meta, active, created_at, updated_at
		FROM pois WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	if err != nil {
		if isNoRows(err) {
			return domain.POI{}, fmt.Errorf("%w: poi", domain.ErrNotFound)
		}
		return domain.POI{}, err
	}
	return p, nil
}

func (r *POIRepo) Nearby(ctx context.Context, q domain.NearbyQuery) ([]domain.POI, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}
	distExpr := `(6371000.0 * 2 * asin(sqrt(
		power(sin(radians(($1 - lat) / 2)), 2) +
		cos(radians($1)) * cos(radians(lat)) * power(sin(radians(($2 - lng) / 2)), 2)
	)))`
	sqlQ := `
		SELECT id, tenant_id, kind, ref_id, name, lat, lng, meta, active, created_at, updated_at
		FROM pois
		WHERE tenant_id=$3 AND active=TRUE
		  AND ($4::text IS NULL OR kind = $4::poi_kind)
		  AND ` + distExpr + ` <= $5
		ORDER BY ` + distExpr + ` ASC
		LIMIT $6`
	var kind any
	if q.Kind != nil {
		kind = string(*q.Kind)
	}
	return r.scanMany(ctx, sqlQ, q.Center.Lat, q.Center.Lng, q.TenantID, kind, q.RadiusM, limit)
}

func (r *POIRepo) NearestOfKind(ctx context.Context, tenantID uuid.UUID, kind domain.POIKind, lat, lng float64, limit int) ([]domain.POI, error) {
	if limit <= 0 {
		limit = 1
	}
	distExpr := `(6371000.0 * 2 * asin(sqrt(
		power(sin(radians(($1 - lat) / 2)), 2) +
		cos(radians($1)) * cos(radians(lat)) * power(sin(radians(($2 - lng) / 2)), 2)
	)))`
	sqlQ := `
		SELECT id, tenant_id, kind, ref_id, name, lat, lng, meta, active, created_at, updated_at
		FROM pois
		WHERE tenant_id=$3 AND active=TRUE AND kind=$4::poi_kind
		ORDER BY ` + distExpr + ` ASC
		LIMIT $5`
	return r.scanMany(ctx, sqlQ, lat, lng, tenantID, string(kind), limit)
}

func (r *POIRepo) InBBox(ctx context.Context, tenantID uuid.UUID, bbox domain.BBox, kind *domain.POIKind, limit int) ([]domain.POI, error) {
	if limit <= 0 {
		limit = 50
	}
	var kindArg any
	if kind != nil {
		kindArg = string(*kind)
	}
	return r.scanMany(ctx, `
		SELECT id, tenant_id, kind, ref_id, name, lat, lng, meta, active, created_at, updated_at
		FROM pois
		WHERE tenant_id=$1 AND active=TRUE
		  AND ($2::text IS NULL OR kind = $2::poi_kind)
		  AND lat BETWEEN $3 AND $4 AND lng BETWEEN $5 AND $6
		LIMIT $7`,
		tenantID, kindArg, bbox.MinLat, bbox.MaxLat, bbox.MinLng, bbox.MaxLng, limit)
}

func (r *POIRepo) Count(ctx context.Context, tenantID uuid.UUID) (int, error) {
	var n int
	err := r.DB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM pois WHERE tenant_id=$1 AND active=TRUE`, tenantID).Scan(&n)
	return n, err
}

func (r *POIRepo) scanOne(ctx context.Context, q string, args ...any) (domain.POI, error) {
	var p domain.POI
	var kind string
	var meta JSONMap
	err := r.DB.QueryRowContext(ctx, q, args...).Scan(
		&p.ID, &p.TenantID, &kind, &p.RefID, &p.Name, &p.Lat, &p.Lng, &meta, &p.Active, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return domain.POI{}, err
	}
	p.Kind = domain.POIKind(kind)
	p.Meta = map[string]any(meta)
	p.CreatedAt = p.CreatedAt.UTC()
	p.UpdatedAt = p.UpdatedAt.UTC()
	return p, nil
}

func (r *POIRepo) scanMany(ctx context.Context, q string, args ...any) ([]domain.POI, error) {
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.POI, 0)
	for rows.Next() {
		var p domain.POI
		var kind string
		var meta JSONMap
		if err := rows.Scan(
			&p.ID, &p.TenantID, &kind, &p.RefID, &p.Name, &p.Lat, &p.Lng, &meta, &p.Active, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.Kind = domain.POIKind(kind)
		p.Meta = map[string]any(meta)
		p.CreatedAt = p.CreatedAt.UTC()
		p.UpdatedAt = p.UpdatedAt.UTC()
		out = append(out, p)
	}
	return out, rows.Err()
}

var _ ports.POIRepo = (*POIRepo)(nil)

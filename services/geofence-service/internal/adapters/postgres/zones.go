package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/geofence-service/internal/app/ports"
	"github.com/nexora/geofence-service/internal/domain"
)

// ZoneRepo persists geofence zones.
type ZoneRepo struct{ DB *sql.DB }

func (r *ZoneRepo) Create(ctx context.Context, z domain.Zone) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO zones (
			id, tenant_id, name, city, kind, vertices, center_lat, center_lng, radius_m,
			active, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		z.ID, z.TenantID, z.Name, z.City, string(z.Kind), VerticesJSON(z.Vertices),
		nullFloat(z.CenterLat), nullFloat(z.CenterLng), nullFloat(z.RadiusM),
		z.Active, z.CreatedAt.UTC(), z.UpdatedAt.UTC())
	if err != nil {
		if mapUniqueViolation(err) == domain.ErrAlreadyExists {
			return fmt.Errorf("%w: zone %s", domain.ErrAlreadyExists, z.ID)
		}
		return err
	}
	return nil
}

func (r *ZoneRepo) Update(ctx context.Context, z domain.Zone) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE zones SET
			name=$3, city=$4, kind=$5, vertices=$6, center_lat=$7, center_lng=$8, radius_m=$9,
			active=$10, updated_at=$11
		WHERE id=$1 AND tenant_id=$2`,
		z.ID, z.TenantID, z.Name, z.City, string(z.Kind), VerticesJSON(z.Vertices),
		nullFloat(z.CenterLat), nullFloat(z.CenterLng), nullFloat(z.RadiusM),
		z.Active, z.UpdatedAt.UTC())
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("%w: zone %s", domain.ErrNotFound, z.ID)
	}
	return nil
}

func (r *ZoneRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	res, err := r.DB.ExecContext(ctx, `DELETE FROM zones WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("%w: zone %s", domain.ErrNotFound, id)
	}
	return nil
}

func (r *ZoneRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Zone, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, city, kind, vertices, center_lat, center_lng, radius_m,
			active, created_at, updated_at
		FROM zones WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	z, err := scanZone(row)
	if err != nil {
		if isNoRows(err) {
			return domain.Zone{}, fmt.Errorf("%w: zone %s", domain.ErrNotFound, id)
		}
		return domain.Zone{}, err
	}
	return z, nil
}

func (r *ZoneRepo) List(ctx context.Context, tenantID uuid.UUID, city string, kind domain.ZoneKind, limit, offset int) ([]domain.Zone, int, error) {
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	where := []string{"tenant_id=$1"}
	args := []any{tenantID}
	argN := 2
	if city != "" {
		where = append(where, fmt.Sprintf("city=$%d", argN))
		args = append(args, city)
		argN++
	}
	if kind != "" {
		where = append(where, fmt.Sprintf("kind=$%d", argN))
		args = append(args, string(kind))
		argN++
	}
	clause := strings.Join(where, " AND ")

	var total int
	countQ := "SELECT COUNT(*) FROM zones WHERE " + clause
	if err := r.DB.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listArgs := append(append([]any{}, args...), limit, offset)
	listQ := fmt.Sprintf(`
		SELECT id, tenant_id, name, city, kind, vertices, center_lat, center_lng, radius_m,
			active, created_at, updated_at
		FROM zones WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, clause, argN, argN+1)

	rows, err := r.DB.QueryContext(ctx, listQ, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []domain.Zone{}
	for rows.Next() {
		z, err := scanZone(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, z)
	}
	return out, total, rows.Err()
}

func (r *ZoneRepo) ListActive(ctx context.Context, tenantID uuid.UUID, city string) ([]domain.Zone, error) {
	where := []string{"tenant_id=$1", "active=TRUE"}
	args := []any{tenantID}
	if city != "" {
		where = append(where, "(city='' OR city=$2)")
		args = append(args, city)
	}
	q := `
		SELECT id, tenant_id, name, city, kind, vertices, center_lat, center_lng, radius_m,
			active, created_at, updated_at
		FROM zones WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY created_at DESC`

	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Zone{}
	for rows.Next() {
		z, err := scanZone(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, z)
	}
	return out, rows.Err()
}

type scannable interface {
	Scan(dest ...any) error
}

func scanZone(row scannable) (domain.Zone, error) {
	var z domain.Zone
	var kind string
	var vertices VerticesJSON
	var centerLat, centerLng, radiusM sql.NullFloat64
	err := row.Scan(
		&z.ID, &z.TenantID, &z.Name, &z.City, &kind, &vertices,
		&centerLat, &centerLng, &radiusM,
		&z.Active, &z.CreatedAt, &z.UpdatedAt)
	if err != nil {
		return domain.Zone{}, err
	}
	z.Kind = domain.ZoneKind(kind)
	z.Vertices = []domain.Point(vertices)
	z.CenterLat = scanNullFloat(centerLat)
	z.CenterLng = scanNullFloat(centerLng)
	z.RadiusM = scanNullFloat(radiusM)
	z.CreatedAt = z.CreatedAt.UTC()
	z.UpdatedAt = z.UpdatedAt.UTC()
	return z, nil
}

var _ ports.ZoneRepo = (*ZoneRepo)(nil)

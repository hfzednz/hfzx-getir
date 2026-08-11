package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/notification-service/internal/app/ports"
	"github.com/nexora/notification-service/internal/domain"
)

// DeviceRepo persists push devices.
type DeviceRepo struct{ DB *sql.DB }

func (r *DeviceRepo) Upsert(ctx context.Context, d domain.Device) (domain.Device, error) {
	var existingID uuid.UUID
	err := r.DB.QueryRowContext(ctx, `
		SELECT id FROM devices WHERE tenant_id=$1 AND token=$2`, d.TenantID, d.Token).Scan(&existingID)
	if err == nil {
		_, err = r.DB.ExecContext(ctx, `
			UPDATE devices SET principal_id=$3, platform=$4, locale=$5, active=true, updated_at=$6
			WHERE id=$1 AND tenant_id=$2`,
			existingID, d.TenantID, d.PrincipalID, string(d.Platform), d.Locale, d.UpdatedAt.UTC())
		if err != nil {
			return domain.Device{}, err
		}
		return r.getByID(ctx, d.TenantID, existingID)
	}
	if !isNoRows(err) {
		return domain.Device{}, err
	}
	_, err = r.DB.ExecContext(ctx, `
		INSERT INTO devices (id, tenant_id, principal_id, platform, token, locale, active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		d.ID, d.TenantID, d.PrincipalID, string(d.Platform), d.Token, d.Locale, d.Active,
		d.CreatedAt.UTC(), d.UpdatedAt.UTC())
	if err != nil {
		return domain.Device{}, mapUniqueViolation(err)
	}
	return d, nil
}

func (r *DeviceRepo) ListActive(ctx context.Context, tenantID, principalID uuid.UUID) ([]domain.Device, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, principal_id, platform, token, locale, active, created_at, updated_at
		FROM devices WHERE tenant_id=$1 AND principal_id=$2 AND active=true`, tenantID, principalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Device{}
	for rows.Next() {
		d, err := scanDevice(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *DeviceRepo) Deactivate(ctx context.Context, tenantID, id uuid.UUID) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE devices SET active=false, updated_at=now() WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *DeviceRepo) getByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Device, error) {
	d, err := scanDevice(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, principal_id, platform, token, locale, active, created_at, updated_at
		FROM devices WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if isNoRows(err) {
		return domain.Device{}, domain.ErrNotFound
	}
	return d, err
}

func scanDevice(row scanner) (domain.Device, error) {
	var d domain.Device
	var platform string
	err := row.Scan(
		&d.ID, &d.TenantID, &d.PrincipalID, &platform, &d.Token, &d.Locale, &d.Active,
		&d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return domain.Device{}, err
	}
	d.Platform = domain.DevicePlatform(platform)
	d.CreatedAt = d.CreatedAt.UTC()
	d.UpdatedAt = d.UpdatedAt.UTC()
	return d, nil
}

var _ ports.DeviceRepo = (*DeviceRepo)(nil)

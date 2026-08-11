package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/identity-service/internal/app/ports"
	"github.com/nexora/identity-service/internal/domain"
)

type DeviceRepo struct{ DB *sql.DB }

func (r *DeviceRepo) Create(ctx context.Context, d domain.Device) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO devices (id, principal_id, fingerprint, platform, name, trusted, trusted_at, last_seen_at, created_at, updated_at, revoked_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		d.ID, d.PrincipalID, d.Fingerprint, d.Platform, d.Name, d.Trusted, d.TrustedAt, d.LastSeenAt, d.CreatedAt, d.UpdatedAt, d.RevokedAt)
	return err
}

func (r *DeviceRepo) Update(ctx context.Context, d domain.Device) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE devices SET platform=$2, name=$3, trusted=$4, trusted_at=$5, last_seen_at=$6, updated_at=$7, revoked_at=$8
		WHERE id=$1`, d.ID, d.Platform, d.Name, d.Trusted, d.TrustedAt, d.LastSeenAt, d.UpdatedAt, d.RevokedAt)
	return err
}

func (r *DeviceRepo) scan(row interface{ Scan(dest ...any) error }) (domain.Device, error) {
	var d domain.Device
	err := row.Scan(&d.ID, &d.PrincipalID, &d.Fingerprint, &d.Platform, &d.Name, &d.Trusted, &d.TrustedAt, &d.LastSeenAt, &d.CreatedAt, &d.UpdatedAt, &d.RevokedAt)
	if err != nil {
		return domain.Device{}, mapNotFound(err)
	}
	return d, nil
}

func (r *DeviceRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Device, error) {
	return r.scan(r.DB.QueryRowContext(ctx, `
		SELECT id, principal_id, fingerprint, platform, name, trusted, trusted_at, last_seen_at, created_at, updated_at, revoked_at
		FROM devices WHERE id=$1`, id))
}

func (r *DeviceRepo) FindByFingerprint(ctx context.Context, principalID uuid.UUID, fingerprint string) (domain.Device, error) {
	return r.scan(r.DB.QueryRowContext(ctx, `
		SELECT id, principal_id, fingerprint, platform, name, trusted, trusted_at, last_seen_at, created_at, updated_at, revoked_at
		FROM devices WHERE principal_id=$1 AND fingerprint=$2`, principalID, fingerprint))
}

func (r *DeviceRepo) ListByPrincipal(ctx context.Context, principalID uuid.UUID) ([]domain.Device, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, principal_id, fingerprint, platform, name, trusted, trusted_at, last_seen_at, created_at, updated_at, revoked_at
		FROM devices WHERE principal_id=$1 ORDER BY last_seen_at DESC`, principalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Device{}
	for rows.Next() {
		d, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

var _ ports.DeviceRepository = (*DeviceRepo)(nil)

package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/notification-service/internal/app/ports"
	"github.com/nexora/notification-service/internal/domain"
)

// PreferenceRepo persists preferences and consents.
type PreferenceRepo struct{ DB *sql.DB }

func (r *PreferenceRepo) Get(ctx context.Context, tenantID, principalID uuid.UUID) (domain.Preference, error) {
	var p domain.Preference
	var opt JSONBoolMap
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, principal_id, channel_opt_out, quiet_start, quiet_end, updated_at
		FROM preferences WHERE tenant_id=$1 AND principal_id=$2`, tenantID, principalID).Scan(
		&p.ID, &p.TenantID, &p.PrincipalID, &opt, &p.QuietStart, &p.QuietEnd, &p.UpdatedAt)
	if isNoRows(err) {
		return domain.Preference{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Preference{}, err
	}
	p.ChannelOptOut = channelOptOutFromJSON(opt)
	p.UpdatedAt = p.UpdatedAt.UTC()
	return p, nil
}

func (r *PreferenceRepo) Upsert(ctx context.Context, p domain.Preference) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO preferences (id, tenant_id, principal_id, channel_opt_out, quiet_start, quiet_end, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (tenant_id, principal_id) DO UPDATE SET
			id=EXCLUDED.id, channel_opt_out=EXCLUDED.channel_opt_out, quiet_start=EXCLUDED.quiet_start,
			quiet_end=EXCLUDED.quiet_end, updated_at=EXCLUDED.updated_at`,
		p.ID, p.TenantID, p.PrincipalID, channelOptOutToJSON(p.ChannelOptOut),
		p.QuietStart, p.QuietEnd, p.UpdatedAt.UTC())
	return err
}

func (r *PreferenceRepo) RecordConsent(ctx context.Context, c domain.Consent) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO consents (id, tenant_id, principal_id, purpose, granted, source, recorded_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		c.ID, c.TenantID, c.PrincipalID, c.Purpose, c.Granted, c.Source, c.RecordedAt.UTC())
	return err
}

func (r *PreferenceRepo) ListConsents(ctx context.Context, tenantID, principalID uuid.UUID) ([]domain.Consent, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, principal_id, purpose, granted, source, recorded_at
		FROM consents WHERE tenant_id=$1 AND principal_id=$2 ORDER BY recorded_at ASC`, tenantID, principalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Consent{}
	for rows.Next() {
		var c domain.Consent
		if err := rows.Scan(&c.ID, &c.TenantID, &c.PrincipalID, &c.Purpose, &c.Granted, &c.Source, &c.RecordedAt); err != nil {
			return nil, err
		}
		c.RecordedAt = c.RecordedAt.UTC()
		out = append(out, c)
	}
	return out, rows.Err()
}

var _ ports.PreferenceRepo = (*PreferenceRepo)(nil)

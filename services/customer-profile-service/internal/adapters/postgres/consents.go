package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/app/ports"
	"github.com/nexora/customer-profile-service/internal/domain"
)

// ConsentRepo persists consents.
type ConsentRepo struct{ DB *sql.DB }

var _ ports.ConsentRepository = (*ConsentRepo)(nil)

func (r *ConsentRepo) Upsert(ctx context.Context, c domain.Consent) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO consents (
			id, profile_id, tenant_id, channel, granted, source, recorded_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (profile_id, channel) DO UPDATE SET
			id=EXCLUDED.id, granted=EXCLUDED.granted, source=EXCLUDED.source,
			recorded_at=EXCLUDED.recorded_at, updated_at=EXCLUDED.updated_at`,
		c.ID, c.ProfileID, c.TenantID, string(c.Channel), c.Granted, c.Source, c.RecordedAt, c.CreatedAt, c.UpdatedAt,
	)
	return err
}

func (r *ConsentRepo) List(ctx context.Context, profileID uuid.UUID) ([]domain.Consent, error) {
	rows, err := r.DB.QueryContext(ctx, consentSelect+` WHERE profile_id=$1 ORDER BY channel`, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Consent, 0)
	for rows.Next() {
		c, err := scanConsentRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *ConsentRepo) Get(ctx context.Context, profileID uuid.UUID, channel domain.ConsentChannel) (domain.Consent, error) {
	c, err := scanConsentRow(r.DB.QueryRowContext(ctx, consentSelect+` WHERE profile_id=$1 AND channel=$2`, profileID, string(channel)))
	if err != nil {
		return domain.Consent{}, mapNotFound(err)
	}
	return c, nil
}

const consentSelect = `
	SELECT id, profile_id, tenant_id, channel, granted, source, recorded_at, created_at, updated_at
	FROM consents`

func scanConsentRow(row scannable) (domain.Consent, error) {
	var c domain.Consent
	var channel string
	err := row.Scan(&c.ID, &c.ProfileID, &c.TenantID, &channel, &c.Granted, &c.Source, &c.RecordedAt, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return domain.Consent{}, err
	}
	c.Channel = domain.ConsentChannel(channel)
	c.RecordedAt = c.RecordedAt.UTC()
	c.CreatedAt = c.CreatedAt.UTC()
	c.UpdatedAt = c.UpdatedAt.UTC()
	return c, nil
}

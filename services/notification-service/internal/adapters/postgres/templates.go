package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/notification-service/internal/app/ports"
	"github.com/nexora/notification-service/internal/domain"
)

// TemplateRepo persists notification templates.
type TemplateRepo struct{ DB *sql.DB }

func (r *TemplateRepo) Upsert(ctx context.Context, t domain.Template) (domain.Template, error) {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO templates (
			id, tenant_id, key, channel, locale, version, status, subject, body, variant, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (id) DO UPDATE SET
			key=EXCLUDED.key, channel=EXCLUDED.channel, locale=EXCLUDED.locale, version=EXCLUDED.version,
			status=EXCLUDED.status, subject=EXCLUDED.subject, body=EXCLUDED.body, variant=EXCLUDED.variant,
			updated_at=EXCLUDED.updated_at`,
		t.ID, t.TenantID, t.Key, string(t.Channel), t.Locale, t.Version, string(t.Status),
		t.Subject, t.Body, t.Variant, t.CreatedAt.UTC(), t.UpdatedAt.UTC())
	if err != nil {
		return domain.Template{}, mapUniqueViolation(err)
	}
	return t, nil
}

func (r *TemplateRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Template, error) {
	t, err := r.scan(r.DB.QueryRowContext(ctx, templateSelect+` WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if isNoRows(err) {
		return domain.Template{}, domain.ErrNotFound
	}
	return t, err
}

func (r *TemplateRepo) GetByKey(ctx context.Context, tenantID uuid.UUID, key string, channel domain.Channel, locale string) (domain.Template, error) {
	t, err := r.scan(r.DB.QueryRowContext(ctx, templateSelect+`
		WHERE tenant_id=$1 AND key=$2 AND channel=$3 AND locale=$4
		ORDER BY version DESC LIMIT 1`, tenantID, key, string(channel), locale))
	if isNoRows(err) {
		return domain.Template{}, domain.ErrNotFound
	}
	return t, err
}

func (r *TemplateRepo) Approve(ctx context.Context, tenantID, id uuid.UUID, now time.Time) (domain.Template, error) {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE templates SET status=$3, updated_at=$4 WHERE id=$1 AND tenant_id=$2`,
		id, tenantID, string(domain.TemplateActive), now.UTC())
	if err != nil {
		return domain.Template{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.Template{}, domain.ErrNotFound
	}
	return r.Get(ctx, tenantID, id)
}

func (r *TemplateRepo) List(ctx context.Context, tenantID uuid.UUID, key string) ([]domain.Template, error) {
	q := templateSelect + ` WHERE tenant_id=$1`
	args := []any{tenantID}
	if key != "" {
		q += ` AND key=$2`
		args = append(args, key)
	}
	q += ` ORDER BY key, channel, locale, version DESC`
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Template{}
	for rows.Next() {
		t, err := scanTemplate(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

const templateSelect = `
	SELECT id, tenant_id, key, channel, locale, version, status, subject, body, variant, created_at, updated_at
	FROM templates`

type scanner interface {
	Scan(dest ...any) error
}

func (r *TemplateRepo) scan(row *sql.Row) (domain.Template, error) {
	return scanTemplate(row)
}

func scanTemplate(row scanner) (domain.Template, error) {
	var t domain.Template
	var channel, status string
	err := row.Scan(
		&t.ID, &t.TenantID, &t.Key, &channel, &t.Locale, &t.Version, &status,
		&t.Subject, &t.Body, &t.Variant, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return domain.Template{}, err
	}
	t.Channel = domain.Channel(channel)
	t.Status = domain.TemplateStatus(status)
	t.CreatedAt = t.CreatedAt.UTC()
	t.UpdatedAt = t.UpdatedAt.UTC()
	return t, nil
}

var _ ports.TemplateRepo = (*TemplateRepo)(nil)

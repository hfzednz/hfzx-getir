package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/autonomy-service/internal/app/ports"
	"github.com/nexora/autonomy-service/internal/domain"
)

type GenesisRepo struct{ DB *sql.DB }

var _ ports.GenesisRepo = (*GenesisRepo)(nil)

func (r *GenesisRepo) Save(ctx context.Context, c domain.GenesisCertificate) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO auto_genesis (id, tenant_id, version, status, gates, issued_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (id) DO UPDATE SET
			tenant_id=EXCLUDED.tenant_id, version=EXCLUDED.version, status=EXCLUDED.status,
			gates=EXCLUDED.gates, issued_at=EXCLUDED.issued_at, created_at=EXCLUDED.created_at`,
		c.ID, c.TenantID, c.Version, c.Status, JSONBoolMap(c.Gates), nullTime(c.IssuedAt), c.CreatedAt,
	)
	return mapUniqueViolation(err)
}

func (r *GenesisRepo) Latest(ctx context.Context, tenantID uuid.UUID) (domain.GenesisCertificate, error) {
	var c domain.GenesisCertificate
	var gates JSONBoolMap
	var issued sql.NullTime
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, version, status, gates, issued_at, created_at
		FROM auto_genesis WHERE tenant_id=$1
		ORDER BY created_at DESC LIMIT 1`, tenantID).Scan(
		&c.ID, &c.TenantID, &c.Version, &c.Status, &gates, &issued, &c.CreatedAt,
	)
	if isNoRows(err) {
		return domain.GenesisCertificate{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.GenesisCertificate{}, err
	}
	c.Gates = map[string]bool(gates)
	if c.Gates == nil {
		c.Gates = map[string]bool{}
	}
	c.IssuedAt = scanNullTime(issued)
	c.CreatedAt = c.CreatedAt.UTC()
	return c, nil
}

func (r *GenesisRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.GenesisCertificate, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, version, status, gates, issued_at, created_at
		FROM auto_genesis WHERE tenant_id=$1 ORDER BY created_at ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.GenesisCertificate, 0)
	for rows.Next() {
		var c domain.GenesisCertificate
		var gates JSONBoolMap
		var issued sql.NullTime
		if err := rows.Scan(&c.ID, &c.TenantID, &c.Version, &c.Status, &gates, &issued, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.Gates = map[string]bool(gates)
		if c.Gates == nil {
			c.Gates = map[string]bool{}
		}
		c.IssuedAt = scanNullTime(issued)
		c.CreatedAt = c.CreatedAt.UTC()
		out = append(out, c)
	}
	return out, rows.Err()
}

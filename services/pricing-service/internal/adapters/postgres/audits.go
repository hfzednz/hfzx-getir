package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/pricing-service/internal/app/ports"
	"github.com/nexora/pricing-service/internal/domain"
)

// QuoteAuditRepo persists quote audit snapshots.
type QuoteAuditRepo struct{ DB *sql.DB }

func (r *QuoteAuditRepo) Create(ctx context.Context, a domain.QuoteAudit) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO quote_audit (id, tenant_id, quote_id, cart_id, simulated, request, response, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		a.ID, a.TenantID, a.QuoteID, nullUUID(a.CartID), a.Simulated,
		JSONMap(a.Request), JSONMap(a.Response), a.CreatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *QuoteAuditRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.QuoteAudit, error) {
	a, err := scanQuoteAudit(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, quote_id, cart_id, simulated, request, response, created_at
		FROM quote_audit WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if isNoRows(err) {
		return domain.QuoteAudit{}, fmt.Errorf("%w: quote audit", domain.ErrNotFound)
	}
	return a, err
}

func (r *QuoteAuditRepo) List(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.QuoteAudit, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, quote_id, cart_id, simulated, request, response, created_at
		FROM quote_audit WHERE tenant_id=$1 ORDER BY created_at DESC LIMIT $2`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.QuoteAudit, 0)
	for rows.Next() {
		a, err := scanQuoteAudit(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func scanQuoteAudit(row scanner) (domain.QuoteAudit, error) {
	var a domain.QuoteAudit
	var cartID uuid.NullUUID
	var req, resp JSONMap
	err := row.Scan(&a.ID, &a.TenantID, &a.QuoteID, &cartID, &a.Simulated, &req, &resp, &a.CreatedAt)
	if err != nil {
		return domain.QuoteAudit{}, err
	}
	a.CartID = scanNullUUID(cartID)
	a.Request = map[string]any(req)
	a.Response = map[string]any(resp)
	a.CreatedAt = a.CreatedAt.UTC()
	return a, nil
}

var _ ports.QuoteAuditRepo = (*QuoteAuditRepo)(nil)

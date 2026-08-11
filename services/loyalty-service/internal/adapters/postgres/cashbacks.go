package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/app/ports"
	"github.com/nexora/loyalty-service/internal/domain"
)

// CashbackRepo persists cashback grants.
type CashbackRepo struct{ DB *sql.DB }

func (r *CashbackRepo) CreateGrant(ctx context.Context, g domain.CashbackGrant) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO cashback_grants
		  (id, tenant_id, account_id, principal_id, amount_minor, currency, account_type, status,
		   order_id, idempotency_key, wallet_ref, failure_reason, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		g.ID, g.TenantID, g.AccountID, g.PrincipalID, g.AmountMinor, g.Currency, g.AccountType, string(g.Status),
		nullUUID(g.OrderID), g.IdempotencyKey, g.WalletRef, g.FailureReason, g.CreatedAt.UTC(), g.UpdatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *CashbackRepo) GetGrant(ctx context.Context, tenantID, grantID uuid.UUID) (domain.CashbackGrant, error) {
	return r.scanGrant(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, account_id, principal_id, amount_minor, currency, account_type, status,
		       order_id, idempotency_key, wallet_ref, failure_reason, created_at, updated_at
		FROM cashback_grants WHERE id=$1 AND tenant_id=$2`, grantID, tenantID))
}

func (r *CashbackRepo) GetGrantByIdempotency(ctx context.Context, tenantID uuid.UUID, key string) (domain.CashbackGrant, error) {
	return r.scanGrant(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, account_id, principal_id, amount_minor, currency, account_type, status,
		       order_id, idempotency_key, wallet_ref, failure_reason, created_at, updated_at
		FROM cashback_grants WHERE tenant_id=$1 AND idempotency_key=$2`, tenantID, key))
}

func (r *CashbackRepo) UpdateGrant(ctx context.Context, g domain.CashbackGrant) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE cashback_grants
		SET status=$3, order_id=$4, wallet_ref=$5, failure_reason=$6, updated_at=$7
		WHERE id=$1 AND tenant_id=$2`,
		g.ID, g.TenantID, string(g.Status), nullUUID(g.OrderID), g.WalletRef, g.FailureReason, g.UpdatedAt.UTC())
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *CashbackRepo) scanGrant(row *sql.Row) (domain.CashbackGrant, error) {
	var g domain.CashbackGrant
	var status string
	var orderID uuid.NullUUID
	err := row.Scan(
		&g.ID, &g.TenantID, &g.AccountID, &g.PrincipalID, &g.AmountMinor, &g.Currency, &g.AccountType, &status,
		&orderID, &g.IdempotencyKey, &g.WalletRef, &g.FailureReason, &g.CreatedAt, &g.UpdatedAt)
	if isNoRows(err) {
		return domain.CashbackGrant{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.CashbackGrant{}, err
	}
	g.Status = domain.CashbackStatus(status)
	g.OrderID = scanNullUUID(orderID)
	g.CreatedAt = g.CreatedAt.UTC()
	g.UpdatedAt = g.UpdatedAt.UTC()
	return g, nil
}

var _ ports.CashbackRepo = (*CashbackRepo)(nil)

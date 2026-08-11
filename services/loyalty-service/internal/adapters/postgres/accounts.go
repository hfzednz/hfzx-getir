package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/app/ports"
	"github.com/nexora/loyalty-service/internal/domain"
)

// AccountRepo persists loyalty accounts, point ledger, audit, and stats.
type AccountRepo struct{ DB *sql.DB }

func (r *AccountRepo) CreateAccount(ctx context.Context, a domain.Account) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO loyalty_accounts
		  (id, tenant_id, principal_id, points, tier_points, xp, level, active, version, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		a.ID, a.TenantID, a.PrincipalID, a.Points, a.TierPoints, a.XP, a.Level, a.Active, a.Version,
		a.CreatedAt.UTC(), a.UpdatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *AccountRepo) GetAccount(ctx context.Context, tenantID, accountID uuid.UUID) (domain.Account, error) {
	return r.scanAccount(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, principal_id, points, tier_points, xp, level, active, version, created_at, updated_at
		FROM loyalty_accounts WHERE id=$1 AND tenant_id=$2`, accountID, tenantID))
}

func (r *AccountRepo) GetAccountByPrincipal(ctx context.Context, tenantID, principalID uuid.UUID) (domain.Account, error) {
	return r.scanAccount(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, principal_id, points, tier_points, xp, level, active, version, created_at, updated_at
		FROM loyalty_accounts WHERE tenant_id=$1 AND principal_id=$2`, tenantID, principalID))
}

func (r *AccountRepo) UpdateAccount(ctx context.Context, a domain.Account) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE loyalty_accounts
		SET points=$3, tier_points=$4, xp=$5, level=$6, active=$7, version=$8, updated_at=$9
		WHERE id=$1 AND tenant_id=$2`,
		a.ID, a.TenantID, a.Points, a.TierPoints, a.XP, a.Level, a.Active, a.Version, a.UpdatedAt.UTC())
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *AccountRepo) CreateLedgerEntry(ctx context.Context, e domain.PointLedgerEntry) error {
	meta := JSONMap(e.Metadata)
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO point_ledger
		  (id, tenant_id, account_id, kind, points, balance_after, order_id, reference,
		   idempotency_key, metadata, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		e.ID, e.TenantID, e.AccountID, string(e.Kind), e.Points, e.BalanceAfter, nullUUID(e.OrderID),
		e.Reference, e.IdempotencyKey, meta, e.CreatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *AccountRepo) GetLedgerByIdempotency(ctx context.Context, tenantID uuid.UUID, key string) (domain.PointLedgerEntry, error) {
	return r.scanLedger(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, account_id, kind, points, balance_after, order_id, reference,
		       idempotency_key, metadata, created_at
		FROM point_ledger WHERE tenant_id=$1 AND idempotency_key=$2`, tenantID, key))
}

func (r *AccountRepo) GetLedgerByOrder(ctx context.Context, tenantID, accountID, orderID uuid.UUID) (domain.PointLedgerEntry, error) {
	return r.scanLedger(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, account_id, kind, points, balance_after, order_id, reference,
		       idempotency_key, metadata, created_at
		FROM point_ledger
		WHERE tenant_id=$1 AND account_id=$2 AND order_id=$3
		ORDER BY created_at DESC LIMIT 1`, tenantID, accountID, orderID))
}

func (r *AccountRepo) ListLedger(ctx context.Context, tenantID, accountID uuid.UUID, limit, offset int) ([]domain.PointLedgerEntry, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	var total int
	if err := r.DB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM point_ledger WHERE tenant_id=$1 AND account_id=$2`,
		tenantID, accountID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, account_id, kind, points, balance_after, order_id, reference,
		       idempotency_key, metadata, created_at
		FROM point_ledger
		WHERE tenant_id=$1 AND account_id=$2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`, tenantID, accountID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []domain.PointLedgerEntry
	for rows.Next() {
		e, err := scanLedgerRow(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}

func (r *AccountRepo) CreateAudit(ctx context.Context, a domain.AuditEntry) error {
	detail := JSONMap(a.Detail)
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO loyalty_audit (id, tenant_id, account_id, action, actor_id, detail, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		a.ID, a.TenantID, nullUUIDValue(a.AccountID), a.Action, nullUUID(a.ActorID), detail, a.CreatedAt.UTC())
	return err
}

func (r *AccountRepo) IncrStat(ctx context.Context, tenantID, accountID uuid.UUID, key string, delta int64) (int64, error) {
	var value int64
	err := r.DB.QueryRowContext(ctx, `
		INSERT INTO account_stats (tenant_id, account_id, stat_key, value, updated_at)
		VALUES ($1,$2,$3,$4,now())
		ON CONFLICT (tenant_id, account_id, stat_key)
		DO UPDATE SET value = account_stats.value + EXCLUDED.value, updated_at = now()
		RETURNING value`, tenantID, accountID, key, delta).Scan(&value)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (r *AccountRepo) GetStat(ctx context.Context, tenantID, accountID uuid.UUID, key string) (int64, error) {
	var value int64
	err := r.DB.QueryRowContext(ctx, `
		SELECT value FROM account_stats WHERE tenant_id=$1 AND account_id=$2 AND stat_key=$3`,
		tenantID, accountID, key).Scan(&value)
	if isNoRows(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (r *AccountRepo) scanAccount(row *sql.Row) (domain.Account, error) {
	var a domain.Account
	err := row.Scan(
		&a.ID, &a.TenantID, &a.PrincipalID, &a.Points, &a.TierPoints, &a.XP, &a.Level,
		&a.Active, &a.Version, &a.CreatedAt, &a.UpdatedAt)
	if isNoRows(err) {
		return domain.Account{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Account{}, err
	}
	a.CreatedAt = a.CreatedAt.UTC()
	a.UpdatedAt = a.UpdatedAt.UTC()
	return a, nil
}

func (r *AccountRepo) scanLedger(row *sql.Row) (domain.PointLedgerEntry, error) {
	var e domain.PointLedgerEntry
	var kind string
	var orderID uuid.NullUUID
	var meta JSONMap
	err := row.Scan(
		&e.ID, &e.TenantID, &e.AccountID, &kind, &e.Points, &e.BalanceAfter, &orderID,
		&e.Reference, &e.IdempotencyKey, &meta, &e.CreatedAt)
	if isNoRows(err) {
		return domain.PointLedgerEntry{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.PointLedgerEntry{}, err
	}
	e.Kind = domain.PointEntryKind(kind)
	e.OrderID = scanNullUUID(orderID)
	e.Metadata = map[string]any(meta)
	e.CreatedAt = e.CreatedAt.UTC()
	return e, nil
}

func scanLedgerRow(row scannable) (domain.PointLedgerEntry, error) {
	var e domain.PointLedgerEntry
	var kind string
	var orderID uuid.NullUUID
	var meta JSONMap
	if err := row.Scan(
		&e.ID, &e.TenantID, &e.AccountID, &kind, &e.Points, &e.BalanceAfter, &orderID,
		&e.Reference, &e.IdempotencyKey, &meta, &e.CreatedAt); err != nil {
		return domain.PointLedgerEntry{}, err
	}
	e.Kind = domain.PointEntryKind(kind)
	e.OrderID = scanNullUUID(orderID)
	e.Metadata = map[string]any(meta)
	e.CreatedAt = e.CreatedAt.UTC()
	return e, nil
}

var _ ports.AccountRepo = (*AccountRepo)(nil)

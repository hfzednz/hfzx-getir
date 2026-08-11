package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/wallet-service/internal/app/ports"
	"github.com/nexora/wallet-service/internal/domain"
)

// WalletRepo persists wallets and related money entities.
type WalletRepo struct{ DB *sql.DB }

func (r *WalletRepo) CreateWallet(ctx context.Context, w domain.Wallet, accounts []domain.Account) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO wallets (id, tenant_id, principal_id, currency, active, version, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		w.ID, w.TenantID, w.PrincipalID, w.Currency, w.Active, w.Version, w.CreatedAt.UTC(), w.UpdatedAt.UTC())
	if err != nil {
		return mapUniqueViolation(err)
	}
	for _, a := range accounts {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO wallet_accounts
			  (id, wallet_id, tenant_id, account_type, balance_minor, held_minor, currency, version, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
			a.ID, a.WalletID, a.TenantID, string(a.AccountType), a.BalanceMinor, a.HeldMinor,
			a.Currency, a.Version, a.UpdatedAt.UTC())
		if err != nil {
			return mapUniqueViolation(err)
		}
	}
	return tx.Commit()
}

func (r *WalletRepo) GetWallet(ctx context.Context, tenantID, walletID uuid.UUID) (domain.Wallet, error) {
	var w domain.Wallet
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, principal_id, currency, active, version, created_at, updated_at
		FROM wallets WHERE id=$1 AND tenant_id=$2`, walletID, tenantID).Scan(
		&w.ID, &w.TenantID, &w.PrincipalID, &w.Currency, &w.Active, &w.Version, &w.CreatedAt, &w.UpdatedAt)
	if isNoRows(err) {
		return domain.Wallet{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Wallet{}, err
	}
	w.CreatedAt = w.CreatedAt.UTC()
	w.UpdatedAt = w.UpdatedAt.UTC()
	return w, nil
}

func (r *WalletRepo) GetWalletByPrincipal(ctx context.Context, tenantID, principalID uuid.UUID) (domain.Wallet, error) {
	var w domain.Wallet
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, principal_id, currency, active, version, created_at, updated_at
		FROM wallets WHERE tenant_id=$1 AND principal_id=$2`, tenantID, principalID).Scan(
		&w.ID, &w.TenantID, &w.PrincipalID, &w.Currency, &w.Active, &w.Version, &w.CreatedAt, &w.UpdatedAt)
	if isNoRows(err) {
		return domain.Wallet{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Wallet{}, err
	}
	w.CreatedAt = w.CreatedAt.UTC()
	w.UpdatedAt = w.UpdatedAt.UTC()
	return w, nil
}

func (r *WalletRepo) UpdateWallet(ctx context.Context, w domain.Wallet) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE wallets SET currency=$3, active=$4, version=$5, updated_at=$6
		WHERE id=$1 AND tenant_id=$2`,
		w.ID, w.TenantID, w.Currency, w.Active, w.Version, w.UpdatedAt.UTC())
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *WalletRepo) GetAccount(ctx context.Context, tenantID, accountID uuid.UUID) (domain.Account, error) {
	return r.scanAccount(r.DB.QueryRowContext(ctx, `
		SELECT id, wallet_id, tenant_id, account_type, balance_minor, held_minor, currency, version, updated_at
		FROM wallet_accounts WHERE id=$1 AND tenant_id=$2`, accountID, tenantID))
}

func (r *WalletRepo) GetAccountByType(ctx context.Context, tenantID, walletID uuid.UUID, t domain.AccountType) (domain.Account, error) {
	return r.scanAccount(r.DB.QueryRowContext(ctx, `
		SELECT id, wallet_id, tenant_id, account_type, balance_minor, held_minor, currency, version, updated_at
		FROM wallet_accounts WHERE tenant_id=$1 AND wallet_id=$2 AND account_type=$3`,
		tenantID, walletID, string(t)))
}

func (r *WalletRepo) ListAccounts(ctx context.Context, tenantID, walletID uuid.UUID) ([]domain.Account, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, wallet_id, tenant_id, account_type, balance_minor, held_minor, currency, version, updated_at
		FROM wallet_accounts WHERE tenant_id=$1 AND wallet_id=$2 ORDER BY account_type`, tenantID, walletID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Account
	for rows.Next() {
		a, err := scanAccountRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *WalletRepo) UpdateAccount(ctx context.Context, a domain.Account) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE wallet_accounts
		SET balance_minor=$3, held_minor=$4, currency=$5, version=$6, updated_at=$7
		WHERE id=$1 AND tenant_id=$2`,
		a.ID, a.TenantID, a.BalanceMinor, a.HeldMinor, a.Currency, a.Version, a.UpdatedAt.UTC())
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *WalletRepo) CreateEntry(ctx context.Context, e domain.Entry) error {
	meta := JSONMap(e.Metadata)
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO wallet_entries
		  (id, wallet_id, account_id, tenant_id, kind, amount_minor, currency,
		   balance_after, held_after, reference, idempotency_key, metadata, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		e.ID, e.WalletID, e.AccountID, e.TenantID, string(e.Kind), e.AmountMinor, e.Currency,
		e.BalanceAfter, e.HeldAfter, e.Reference, e.IdempotencyKey, meta, e.CreatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *WalletRepo) GetEntryByIdempotency(ctx context.Context, tenantID uuid.UUID, key string) (domain.Entry, error) {
	var e domain.Entry
	var kind string
	var meta JSONMap
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, wallet_id, account_id, tenant_id, kind, amount_minor, currency,
		       balance_after, held_after, reference, idempotency_key, metadata, created_at
		FROM wallet_entries WHERE tenant_id=$1 AND idempotency_key=$2`, tenantID, key).Scan(
		&e.ID, &e.WalletID, &e.AccountID, &e.TenantID, &kind, &e.AmountMinor, &e.Currency,
		&e.BalanceAfter, &e.HeldAfter, &e.Reference, &e.IdempotencyKey, &meta, &e.CreatedAt)
	if isNoRows(err) {
		return domain.Entry{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Entry{}, err
	}
	e.Kind = domain.EntryKind(kind)
	e.Metadata = map[string]any(meta)
	e.CreatedAt = e.CreatedAt.UTC()
	return e, nil
}

func (r *WalletRepo) ListEntries(ctx context.Context, tenantID, walletID uuid.UUID, limit, offset int) ([]domain.Entry, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	var total int
	if err := r.DB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM wallet_entries WHERE tenant_id=$1 AND wallet_id=$2`, tenantID, walletID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, wallet_id, account_id, tenant_id, kind, amount_minor, currency,
		       balance_after, held_after, reference, idempotency_key, metadata, created_at
		FROM wallet_entries WHERE tenant_id=$1 AND wallet_id=$2
		ORDER BY created_at DESC LIMIT $3 OFFSET $4`, tenantID, walletID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []domain.Entry
	for rows.Next() {
		var e domain.Entry
		var kind string
		var meta JSONMap
		if err := rows.Scan(
			&e.ID, &e.WalletID, &e.AccountID, &e.TenantID, &kind, &e.AmountMinor, &e.Currency,
			&e.BalanceAfter, &e.HeldAfter, &e.Reference, &e.IdempotencyKey, &meta, &e.CreatedAt); err != nil {
			return nil, 0, err
		}
		e.Kind = domain.EntryKind(kind)
		e.Metadata = map[string]any(meta)
		e.CreatedAt = e.CreatedAt.UTC()
		out = append(out, e)
	}
	return out, total, rows.Err()
}

func (r *WalletRepo) CreateHold(ctx context.Context, h domain.Hold) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO wallet_holds
		  (id, wallet_id, account_id, tenant_id, amount_minor, currency, status,
		   reference, idempotency_key, created_at, released_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		h.ID, h.WalletID, h.AccountID, h.TenantID, h.AmountMinor, h.Currency, string(h.Status),
		h.Reference, h.IdempotencyKey, h.CreatedAt.UTC(), nullTime(h.ReleasedAt))
	return mapUniqueViolation(err)
}

func (r *WalletRepo) GetHold(ctx context.Context, tenantID, holdID uuid.UUID) (domain.Hold, error) {
	h, err := r.scanHold(r.DB.QueryRowContext(ctx, `
		SELECT id, wallet_id, account_id, tenant_id, amount_minor, currency, status,
		       reference, idempotency_key, created_at, released_at
		FROM wallet_holds WHERE id=$1 AND tenant_id=$2`, holdID, tenantID))
	if isNoRows(err) {
		return domain.Hold{}, domain.ErrHoldNotFound
	}
	return h, err
}

func (r *WalletRepo) GetHoldByIdempotency(ctx context.Context, tenantID uuid.UUID, key string) (domain.Hold, error) {
	h, err := r.scanHold(r.DB.QueryRowContext(ctx, `
		SELECT id, wallet_id, account_id, tenant_id, amount_minor, currency, status,
		       reference, idempotency_key, created_at, released_at
		FROM wallet_holds WHERE tenant_id=$1 AND idempotency_key=$2`, tenantID, key))
	if isNoRows(err) {
		return domain.Hold{}, domain.ErrNotFound
	}
	return h, err
}

func (r *WalletRepo) UpdateHold(ctx context.Context, h domain.Hold) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE wallet_holds SET status=$3, released_at=$4
		WHERE id=$1 AND tenant_id=$2`,
		h.ID, h.TenantID, string(h.Status), nullTime(h.ReleasedAt))
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrHoldNotFound
	}
	return nil
}

func (r *WalletRepo) CreateTransfer(ctx context.Context, t domain.Transfer) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO wallet_transfers
		  (id, tenant_id, from_wallet_id, from_account_id, to_wallet_id, to_account_id,
		   amount_minor, currency, idempotency_key, reference, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		t.ID, t.TenantID, t.FromWalletID, t.FromAccountID, t.ToWalletID, t.ToAccountID,
		t.AmountMinor, t.Currency, t.IdempotencyKey, t.Reference, t.CreatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *WalletRepo) GetTransferByIdempotency(ctx context.Context, tenantID uuid.UUID, key string) (domain.Transfer, error) {
	var t domain.Transfer
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, from_wallet_id, from_account_id, to_wallet_id, to_account_id,
		       amount_minor, currency, idempotency_key, reference, created_at
		FROM wallet_transfers WHERE tenant_id=$1 AND idempotency_key=$2`, tenantID, key).Scan(
		&t.ID, &t.TenantID, &t.FromWalletID, &t.FromAccountID, &t.ToWalletID, &t.ToAccountID,
		&t.AmountMinor, &t.Currency, &t.IdempotencyKey, &t.Reference, &t.CreatedAt)
	if isNoRows(err) {
		return domain.Transfer{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Transfer{}, err
	}
	t.CreatedAt = t.CreatedAt.UTC()
	return t, nil
}

func (r *WalletRepo) UpsertLimit(ctx context.Context, l domain.Limit) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO wallet_limits
		  (id, wallet_id, tenant_id, limit_type, amount_minor, currency, window_key, used_minor, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (wallet_id, limit_type, window_key) DO UPDATE SET
		  amount_minor=EXCLUDED.amount_minor,
		  currency=EXCLUDED.currency,
		  used_minor=EXCLUDED.used_minor,
		  updated_at=EXCLUDED.updated_at,
		  tenant_id=EXCLUDED.tenant_id`,
		l.ID, l.WalletID, l.TenantID, l.LimitType, l.AmountMinor, l.Currency, l.WindowKey, l.UsedMinor, l.UpdatedAt.UTC())
	return err
}

func (r *WalletRepo) GetLimit(ctx context.Context, tenantID, walletID uuid.UUID, limitType, windowKey string) (domain.Limit, error) {
	var l domain.Limit
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, wallet_id, tenant_id, limit_type, amount_minor, currency, window_key, used_minor, updated_at
		FROM wallet_limits WHERE tenant_id=$1 AND wallet_id=$2 AND limit_type=$3 AND window_key=$4`,
		tenantID, walletID, limitType, windowKey).Scan(
		&l.ID, &l.WalletID, &l.TenantID, &l.LimitType, &l.AmountMinor, &l.Currency, &l.WindowKey, &l.UsedMinor, &l.UpdatedAt)
	if isNoRows(err) {
		return domain.Limit{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Limit{}, err
	}
	l.UpdatedAt = l.UpdatedAt.UTC()
	return l, nil
}

func (r *WalletRepo) CreateAudit(ctx context.Context, a domain.AuditEntry) error {
	detail := JSONMap(a.Detail)
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO wallet_audit
		  (id, tenant_id, wallet_id, action, actor_id, amount_minor, currency, detail, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		a.ID, a.TenantID, a.WalletID, a.Action, nullUUID(a.ActorID), a.AmountMinor, a.Currency, detail, a.CreatedAt.UTC())
	return err
}

func (r *WalletRepo) scanAccount(row *sql.Row) (domain.Account, error) {
	var a domain.Account
	var typ string
	err := row.Scan(&a.ID, &a.WalletID, &a.TenantID, &typ, &a.BalanceMinor, &a.HeldMinor, &a.Currency, &a.Version, &a.UpdatedAt)
	if isNoRows(err) {
		return domain.Account{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Account{}, err
	}
	a.AccountType = domain.AccountType(typ)
	a.UpdatedAt = a.UpdatedAt.UTC()
	return a, nil
}

func scanAccountRow(rows *sql.Rows) (domain.Account, error) {
	var a domain.Account
	var typ string
	err := rows.Scan(&a.ID, &a.WalletID, &a.TenantID, &typ, &a.BalanceMinor, &a.HeldMinor, &a.Currency, &a.Version, &a.UpdatedAt)
	if err != nil {
		return domain.Account{}, err
	}
	a.AccountType = domain.AccountType(typ)
	a.UpdatedAt = a.UpdatedAt.UTC()
	return a, nil
}

func (r *WalletRepo) scanHold(row *sql.Row) (domain.Hold, error) {
	var h domain.Hold
	var status string
	var released sql.NullTime
	err := row.Scan(
		&h.ID, &h.WalletID, &h.AccountID, &h.TenantID, &h.AmountMinor, &h.Currency, &status,
		&h.Reference, &h.IdempotencyKey, &h.CreatedAt, &released)
	if err != nil {
		return domain.Hold{}, err
	}
	h.Status = domain.HoldStatus(status)
	h.CreatedAt = h.CreatedAt.UTC()
	h.ReleasedAt = scanNullTime(released)
	return h, nil
}

var _ ports.WalletRepo = (*WalletRepo)(nil)

package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/finance-ledger-service/internal/app/ports"
	"github.com/nexora/finance-ledger-service/internal/domain"
)

// AccountRepo persists chart-of-accounts.
type AccountRepo struct{ DB *sql.DB }

func (r *AccountRepo) Create(ctx context.Context, a domain.Account) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO chart_of_accounts (
			id, tenant_id, code, name, account_type, currency, active, created_at, updated_at, version
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		a.ID, a.TenantID, a.Code, a.Name, string(a.Type), a.Currency, a.Active,
		a.CreatedAt.UTC(), a.UpdatedAt.UTC(), a.Version)
	return mapUniqueViolation(err)
}

func (r *AccountRepo) Update(ctx context.Context, a domain.Account) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE chart_of_accounts SET name=$3, account_type=$4, currency=$5, active=$6, updated_at=$7, version=$8
		WHERE id=$1 AND tenant_id=$2`,
		a.ID, a.TenantID, a.Name, string(a.Type), a.Currency, a.Active, a.UpdatedAt.UTC(), a.Version)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *AccountRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Account, error) {
	return r.scan(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, code, name, account_type, currency, active, created_at, updated_at, version
		FROM chart_of_accounts WHERE id=$1 AND tenant_id=$2`, id, tenantID))
}

func (r *AccountRepo) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.Account, error) {
	return r.scan(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, code, name, account_type, currency, active, created_at, updated_at, version
		FROM chart_of_accounts WHERE tenant_id=$1 AND code=$2`, tenantID, code))
}

func (r *AccountRepo) scan(row *sql.Row) (domain.Account, error) {
	var a domain.Account
	var typ string
	err := row.Scan(&a.ID, &a.TenantID, &a.Code, &a.Name, &typ, &a.Currency, &a.Active, &a.CreatedAt, &a.UpdatedAt, &a.Version)
	if isNoRows(err) {
		return domain.Account{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Account{}, err
	}
	a.Type = domain.AccountType(typ)
	a.CreatedAt = a.CreatedAt.UTC()
	a.UpdatedAt = a.UpdatedAt.UTC()
	return a, nil
}

// JournalRepo persists journals and lines.
type JournalRepo struct{ DB *sql.DB }

func (r *JournalRepo) Create(ctx context.Context, j domain.Journal) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO journals (
			id, tenant_id, status, currency, reference, description, idempotency_key,
			posted_at, created_at, updated_at, version
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		j.ID, j.TenantID, string(j.Status), j.Currency, j.Reference, j.Description, nullString(j.IdempotencyKey),
		nullTime(j.PostedAt), j.CreatedAt.UTC(), j.UpdatedAt.UTC(), j.Version)
	if err != nil {
		return mapUniqueViolation(err)
	}
	if err := r.insertLines(ctx, tx, j); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *JournalRepo) Update(ctx context.Context, j domain.Journal) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	res, err := tx.ExecContext(ctx, `
		UPDATE journals SET status=$3, currency=$4, reference=$5, description=$6, idempotency_key=$7,
			posted_at=$8, updated_at=$9, version=$10
		WHERE id=$1 AND tenant_id=$2`,
		j.ID, j.TenantID, string(j.Status), j.Currency, j.Reference, j.Description, nullString(j.IdempotencyKey),
		nullTime(j.PostedAt), j.UpdatedAt.UTC(), j.Version)
	if err != nil {
		return mapUniqueViolation(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM journal_lines WHERE journal_id=$1 AND tenant_id=$2`, j.ID, j.TenantID); err != nil {
		return err
	}
	if err := r.insertLines(ctx, tx, j); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *JournalRepo) insertLines(ctx context.Context, tx *sql.Tx, j domain.Journal) error {
	for _, l := range j.Lines {
		id := l.ID
		if id == uuid.Nil {
			id = uuid.New()
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO journal_lines (
				id, journal_id, tenant_id, account_id, account_code, debit_minor, credit_minor, currency, memo
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
			id, j.ID, j.TenantID, l.AccountID, l.AccountCode, l.DebitMinor, l.CreditMinor, l.Currency, l.Memo)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *JournalRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Journal, error) {
	j, err := r.scanHeader(r.DB.QueryRowContext(ctx, journalSelect+` WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if isNoRows(err) {
		return domain.Journal{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Journal{}, err
	}
	lines, err := r.loadLines(ctx, tenantID, id)
	if err != nil {
		return domain.Journal{}, err
	}
	j.Lines = lines
	return j, nil
}

func (r *JournalRepo) GetByIdempotencyKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.Journal, error) {
	j, err := r.scanHeader(r.DB.QueryRowContext(ctx, journalSelect+` WHERE tenant_id=$1 AND idempotency_key=$2`, tenantID, key))
	if isNoRows(err) {
		return domain.Journal{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Journal{}, err
	}
	lines, err := r.loadLines(ctx, tenantID, j.ID)
	if err != nil {
		return domain.Journal{}, err
	}
	j.Lines = lines
	return j, nil
}

func (r *JournalRepo) List(ctx context.Context, f ports.JournalFilter) ([]domain.Journal, int, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}
	where := `WHERE tenant_id=$1`
	args := []any{f.TenantID}
	if f.Status != nil {
		where += ` AND status=$2`
		args = append(args, string(*f.Status))
	}
	var total int
	if err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM journals `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	q := journalSelect + ` ` + where + fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, len(args)-1, len(args))
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.Journal{}
	for rows.Next() {
		j, err := scanJournalRow(rows)
		if err != nil {
			return nil, 0, err
		}
		lines, err := r.loadLines(ctx, f.TenantID, j.ID)
		if err != nil {
			return nil, 0, err
		}
		j.Lines = lines
		out = append(out, j)
	}
	return out, total, rows.Err()
}

func (r *JournalRepo) BalanceMinor(ctx context.Context, tenantID, accountID uuid.UUID) (int64, error) {
	var bal sql.NullInt64
	err := r.DB.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(jl.debit_minor - jl.credit_minor), 0)
		FROM journal_lines jl
		JOIN journals j ON j.id = jl.journal_id
		WHERE jl.tenant_id=$1 AND jl.account_id=$2 AND j.status='posted'`, tenantID, accountID).Scan(&bal)
	if err != nil {
		return 0, err
	}
	if !bal.Valid {
		return 0, nil
	}
	return bal.Int64, nil
}

const journalSelect = `
	SELECT id, tenant_id, status, currency, reference, description, COALESCE(idempotency_key,''),
		posted_at, created_at, updated_at, version FROM journals`

func (r *JournalRepo) scanHeader(row *sql.Row) (domain.Journal, error) {
	return scanJournalRow(row)
}

type journalScanner interface {
	Scan(dest ...any) error
}

func scanJournalRow(row journalScanner) (domain.Journal, error) {
	var j domain.Journal
	var status string
	var posted sql.NullTime
	err := row.Scan(
		&j.ID, &j.TenantID, &status, &j.Currency, &j.Reference, &j.Description, &j.IdempotencyKey,
		&posted, &j.CreatedAt, &j.UpdatedAt, &j.Version)
	if err != nil {
		return domain.Journal{}, err
	}
	j.Status = domain.JournalStatus(status)
	j.PostedAt = scanNullTime(posted)
	j.CreatedAt = j.CreatedAt.UTC()
	j.UpdatedAt = j.UpdatedAt.UTC()
	return j, nil
}

func (r *JournalRepo) loadLines(ctx context.Context, tenantID, journalID uuid.UUID) ([]domain.JournalLine, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, account_id, account_code, debit_minor, credit_minor, currency, memo
		FROM journal_lines WHERE tenant_id=$1 AND journal_id=$2`, tenantID, journalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.JournalLine{}
	for rows.Next() {
		var l domain.JournalLine
		if err := rows.Scan(&l.ID, &l.AccountID, &l.AccountCode, &l.DebitMinor, &l.CreditMinor, &l.Currency, &l.Memo); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

var (
	_ ports.AccountRepository = (*AccountRepo)(nil)
	_ ports.JournalRepository = (*JournalRepo)(nil)
)

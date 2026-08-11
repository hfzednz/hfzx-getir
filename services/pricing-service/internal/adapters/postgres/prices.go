package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/pricing-service/internal/app/ports"
	"github.com/nexora/pricing-service/internal/domain"
)

// PriceRepo persists price books and entries.
type PriceRepo struct{ DB *sql.DB }

func (r *PriceRepo) UpsertBook(ctx context.Context, b domain.PriceBook) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO price_books (id, tenant_id, name, currency, active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (id) DO UPDATE SET
			name=EXCLUDED.name, currency=EXCLUDED.currency, active=EXCLUDED.active,
			updated_at=EXCLUDED.updated_at`,
		b.ID, b.TenantID, b.Name, b.Currency, b.Active, b.CreatedAt.UTC(), b.UpdatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *PriceRepo) GetBook(ctx context.Context, tenantID, bookID uuid.UUID) (domain.PriceBook, error) {
	b, err := scanPriceBook(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, currency, active, created_at, updated_at
		FROM price_books WHERE id=$1 AND tenant_id=$2`, bookID, tenantID))
	if isNoRows(err) {
		return domain.PriceBook{}, fmt.Errorf("%w: price book", domain.ErrNotFound)
	}
	return b, err
}

func (r *PriceRepo) ListBooks(ctx context.Context, tenantID uuid.UUID) ([]domain.PriceBook, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, name, currency, active, created_at, updated_at
		FROM price_books WHERE tenant_id=$1 ORDER BY created_at ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.PriceBook, 0)
	for rows.Next() {
		b, err := scanPriceBook(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *PriceRepo) UpsertEntry(ctx context.Context, e domain.PriceEntry) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO price_entries (
			id, tenant_id, price_book_id, variant_id, scope, scope_id,
			amount_minor, currency, valid_from, valid_to, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (id) DO UPDATE SET
			price_book_id=EXCLUDED.price_book_id, variant_id=EXCLUDED.variant_id,
			scope=EXCLUDED.scope, scope_id=EXCLUDED.scope_id,
			amount_minor=EXCLUDED.amount_minor, currency=EXCLUDED.currency,
			valid_from=EXCLUDED.valid_from, valid_to=EXCLUDED.valid_to,
			updated_at=EXCLUDED.updated_at`,
		e.ID, e.TenantID, e.PriceBookID, e.VariantID, string(e.Scope), nullUUID(e.ScopeID),
		e.AmountMinor, e.Currency, e.ValidFrom.UTC(), nullTime(e.ValidTo),
		e.CreatedAt.UTC(), e.UpdatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *PriceRepo) GetEntry(ctx context.Context, tenantID, entryID uuid.UUID) (domain.PriceEntry, error) {
	e, err := scanPriceEntry(r.DB.QueryRowContext(ctx, priceEntrySelect+` WHERE id=$1 AND tenant_id=$2`, entryID, tenantID))
	if isNoRows(err) {
		return domain.PriceEntry{}, fmt.Errorf("%w: price entry", domain.ErrNotFound)
	}
	return e, err
}

func (r *PriceRepo) ListEntries(ctx context.Context, tenantID uuid.UUID, bookID, variantID *uuid.UUID) ([]domain.PriceEntry, error) {
	q := priceEntrySelect + ` WHERE tenant_id=$1`
	args := []any{tenantID}
	if bookID != nil {
		args = append(args, *bookID)
		q += fmt.Sprintf(` AND price_book_id=$%d`, len(args))
	}
	if variantID != nil {
		args = append(args, *variantID)
		q += fmt.Sprintf(` AND variant_id=$%d`, len(args))
	}
	q += ` ORDER BY created_at ASC`
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPriceEntries(rows)
}

func (r *PriceRepo) ListEntriesForVariant(ctx context.Context, tenantID, variantID uuid.UUID) ([]domain.PriceEntry, error) {
	return r.ListEntries(ctx, tenantID, nil, &variantID)
}

const priceEntrySelect = `
	SELECT id, tenant_id, price_book_id, variant_id, scope, scope_id,
		amount_minor, currency, valid_from, valid_to, created_at, updated_at
	FROM price_entries`

func scanPriceBook(row scanner) (domain.PriceBook, error) {
	var b domain.PriceBook
	err := row.Scan(&b.ID, &b.TenantID, &b.Name, &b.Currency, &b.Active, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return domain.PriceBook{}, err
	}
	b.CreatedAt = b.CreatedAt.UTC()
	b.UpdatedAt = b.UpdatedAt.UTC()
	return b, nil
}

func scanPriceEntry(row scanner) (domain.PriceEntry, error) {
	var e domain.PriceEntry
	var scope string
	var scopeID uuid.NullUUID
	var validTo sql.NullTime
	err := row.Scan(
		&e.ID, &e.TenantID, &e.PriceBookID, &e.VariantID, &scope, &scopeID,
		&e.AmountMinor, &e.Currency, &e.ValidFrom, &validTo, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return domain.PriceEntry{}, err
	}
	e.Scope = domain.PriceScope(scope)
	e.ScopeID = scanNullUUID(scopeID)
	e.ValidTo = scanNullTime(validTo)
	e.ValidFrom = e.ValidFrom.UTC()
	e.CreatedAt = e.CreatedAt.UTC()
	e.UpdatedAt = e.UpdatedAt.UTC()
	return e, nil
}

func scanPriceEntries(rows *sql.Rows) ([]domain.PriceEntry, error) {
	out := make([]domain.PriceEntry, 0)
	for rows.Next() {
		e, err := scanPriceEntry(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

var _ ports.PriceRepo = (*PriceRepo)(nil)

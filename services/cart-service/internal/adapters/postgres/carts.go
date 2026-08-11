package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/cart-service/internal/app/ports"
	"github.com/nexora/cart-service/internal/domain"
)

// CartRepo persists the cart aggregate in PostgreSQL.
type CartRepo struct{ DB *sql.DB }

var _ ports.CartRepository = (*CartRepo)(nil)

const cartColumns = `
	id, tenant_id, guest_token, principal_id, city_id, status, currency,
	version, merged_into_id, metadata, created_at, updated_at,
	abandoned_at, converted_at, merged_at
`

func (r *CartRepo) Create(ctx context.Context, c domain.Cart) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO carts (
			id, tenant_id, guest_token, principal_id, city_id, status, currency,
			version, merged_into_id, metadata, created_at, updated_at,
			abandoned_at, converted_at, merged_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,
			$8,$9,$10,$11,$12,
			$13,$14,$15
		)`,
		c.ID, c.TenantID, c.GuestToken, nullUUID(c.PrincipalID), nullUUID(c.CityID),
		string(c.Status), c.Currency, c.Version, nullUUID(c.MergedIntoID), JSONMap(c.Metadata),
		c.CreatedAt, c.UpdatedAt, nullTime(c.AbandonedAt), nullTime(c.ConvertedAt), nullTime(c.MergedAt),
	)
	if err != nil {
		return mapUniqueViolation(err)
	}
	if err := replaceCartChildren(ctx, tx, c); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *CartRepo) Update(ctx context.Context, c domain.Cart) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, `
		UPDATE carts SET
			guest_token=$1, principal_id=$2, city_id=$3, status=$4, currency=$5,
			version=$6, merged_into_id=$7, metadata=$8, updated_at=$9,
			abandoned_at=$10, converted_at=$11, merged_at=$12
		WHERE id=$13 AND tenant_id=$14`,
		c.GuestToken, nullUUID(c.PrincipalID), nullUUID(c.CityID), string(c.Status), c.Currency,
		c.Version, nullUUID(c.MergedIntoID), JSONMap(c.Metadata), c.UpdatedAt,
		nullTime(c.AbandonedAt), nullTime(c.ConvertedAt), nullTime(c.MergedAt),
		c.ID, c.TenantID,
	)
	if err != nil {
		return mapUniqueViolation(err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	if err := replaceCartChildren(ctx, tx, c); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *CartRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Cart, error) {
	c, err := r.scanCart(ctx, r.DB.QueryRowContext(ctx, `
		SELECT `+cartColumns+` FROM carts WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if err != nil {
		return domain.Cart{}, err
	}
	if err := loadCartChildren(ctx, r.DB, &c); err != nil {
		return domain.Cart{}, err
	}
	return c, nil
}

func (r *CartRepo) GetActiveByGuest(ctx context.Context, tenantID uuid.UUID, guestToken string) (domain.Cart, error) {
	c, err := r.scanCart(ctx, r.DB.QueryRowContext(ctx, `
		SELECT `+cartColumns+` FROM carts
		WHERE tenant_id=$1 AND guest_token=$2 AND status='active'
		ORDER BY updated_at DESC
		LIMIT 1`, tenantID, guestToken))
	if err != nil {
		return domain.Cart{}, err
	}
	if err := loadCartChildren(ctx, r.DB, &c); err != nil {
		return domain.Cart{}, err
	}
	return c, nil
}

func (r *CartRepo) GetActiveByPrincipal(ctx context.Context, tenantID, principalID uuid.UUID) (domain.Cart, error) {
	c, err := r.scanCart(ctx, r.DB.QueryRowContext(ctx, `
		SELECT `+cartColumns+` FROM carts
		WHERE tenant_id=$1 AND principal_id=$2 AND status='active'
		ORDER BY updated_at DESC
		LIMIT 1`, tenantID, principalID))
	if err != nil {
		return domain.Cart{}, err
	}
	if err := loadCartChildren(ctx, r.DB, &c); err != nil {
		return domain.Cart{}, err
	}
	return c, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func (r *CartRepo) scanCart(ctx context.Context, row rowScanner) (domain.Cart, error) {
	_ = ctx
	c, err := scanCartRow(row)
	if isNoRows(err) {
		return domain.Cart{}, domain.ErrNotFound
	}
	return c, err
}

func scanCartRow(row rowScanner) (domain.Cart, error) {
	var c domain.Cart
	var status string
	var principal, city, merged uuid.NullUUID
	var meta JSONMap
	var abandoned, converted, mergedAt sql.NullTime
	err := row.Scan(
		&c.ID, &c.TenantID, &c.GuestToken, &principal, &city, &status, &c.Currency,
		&c.Version, &merged, &meta, &c.CreatedAt, &c.UpdatedAt,
		&abandoned, &converted, &mergedAt,
	)
	if err != nil {
		return domain.Cart{}, err
	}
	c.Status = domain.CartStatus(status)
	c.PrincipalID = scanNullUUID(principal)
	c.CityID = scanNullUUID(city)
	c.MergedIntoID = scanNullUUID(merged)
	c.Metadata = map[string]any(meta)
	if c.Metadata == nil {
		c.Metadata = map[string]any{}
	}
	c.AbandonedAt = scanNullTime(abandoned)
	c.ConvertedAt = scanNullTime(converted)
	c.MergedAt = scanNullTime(mergedAt)
	c.CreatedAt = c.CreatedAt.UTC()
	c.UpdatedAt = c.UpdatedAt.UTC()
	return c, nil
}

func replaceCartChildren(ctx context.Context, tx *sql.Tx, c domain.Cart) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM cart_lines WHERE cart_id=$1`, c.ID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM cart_coupons WHERE cart_id=$1`, c.ID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM cart_quotes WHERE cart_id=$1`, c.ID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM cart_reservation_refs WHERE cart_id=$1`, c.ID); err != nil {
		return err
	}
	for _, l := range c.Lines {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO cart_lines (
				id, cart_id, tenant_id, variant_id, qty, max_qty, notes, addons,
				replacement_pref, created_at, updated_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
			l.ID, c.ID, c.TenantID, l.VariantID, l.Qty, l.MaxQty, l.Notes,
			JSONAddons(l.Addons), l.ReplacementPref, l.CreatedAt, l.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}
	for _, cp := range c.Coupons {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO cart_coupons (
				id, cart_id, tenant_id, code, discount_minor, applied_at, metadata
			) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			uuid.New(), c.ID, c.TenantID, cp.Code, cp.DiscountMinor, cp.AppliedAt, JSONMap(cp.Metadata),
		)
		if err != nil {
			return err
		}
	}
	if c.Quote != nil {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO cart_quotes (
				id, cart_id, tenant_id, quote_id, currency,
				subtotal_minor, discount_minor, tax_minor, delivery_minor,
				service_minor, packaging_minor, tip_minor, total_minor,
				line_quotes, quoted_at, created_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
			uuid.New(), c.ID, c.TenantID, c.Quote.QuoteID, c.Quote.Currency,
			c.Quote.SubtotalMinor, c.Quote.DiscountMinor, c.Quote.TaxMinor, c.Quote.DeliveryMinor,
			c.Quote.ServiceMinor, c.Quote.PackagingMinor, c.Quote.TipMinor, c.Quote.TotalMinor,
			JSONLineQuotes(c.Quote.LineQuotes), c.Quote.QuotedAt, c.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}
	for _, ref := range c.ReservationRefs {
		id := ref.ID
		if id == uuid.Nil {
			id = uuid.New()
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO cart_reservation_refs (
				id, cart_id, tenant_id, reservation_ref, idempotency_key,
				expires_at, created_at, released_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			id, c.ID, c.TenantID, ref.ReservationRef, ref.IdempotencyKey,
			nullTime(ref.ExpiresAt), ref.CreatedAt, nullTime(ref.ReleasedAt),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadCartChildren(ctx context.Context, db *sql.DB, c *domain.Cart) error {
	lines, err := loadCartLines(ctx, db, c.ID)
	if err != nil {
		return err
	}
	c.Lines = lines

	coupons, err := loadCartCoupons(ctx, db, c.ID)
	if err != nil {
		return err
	}
	c.Coupons = coupons

	quote, err := loadCartQuote(ctx, db, c.ID)
	if err != nil {
		return err
	}
	c.Quote = quote

	refs, err := loadReservationRefs(ctx, db, c.ID)
	if err != nil {
		return err
	}
	c.ReservationRefs = refs
	return nil
}

func loadCartLines(ctx context.Context, db *sql.DB, cartID uuid.UUID) ([]domain.CartLine, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, cart_id, tenant_id, variant_id, qty, max_qty, notes, addons,
			replacement_pref, created_at, updated_at
		FROM cart_lines WHERE cart_id=$1 ORDER BY created_at`, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.CartLine, 0)
	for rows.Next() {
		var l domain.CartLine
		var addons JSONAddons
		if err := rows.Scan(
			&l.ID, &l.CartID, &l.TenantID, &l.VariantID, &l.Qty, &l.MaxQty, &l.Notes, &addons,
			&l.ReplacementPref, &l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return nil, err
		}
		l.Addons = []domain.LineAddon(addons)
		l.CreatedAt = l.CreatedAt.UTC()
		l.UpdatedAt = l.UpdatedAt.UTC()
		out = append(out, l)
	}
	return out, rows.Err()
}

func loadCartCoupons(ctx context.Context, db *sql.DB, cartID uuid.UUID) ([]domain.AppliedCoupon, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT code, discount_minor, applied_at, metadata
		FROM cart_coupons WHERE cart_id=$1 ORDER BY applied_at`, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.AppliedCoupon, 0)
	for rows.Next() {
		var cp domain.AppliedCoupon
		var meta JSONMap
		if err := rows.Scan(&cp.Code, &cp.DiscountMinor, &cp.AppliedAt, &meta); err != nil {
			return nil, err
		}
		cp.Metadata = map[string]any(meta)
		if cp.Metadata == nil {
			cp.Metadata = map[string]any{}
		}
		cp.AppliedAt = cp.AppliedAt.UTC()
		out = append(out, cp)
	}
	return out, rows.Err()
}

func loadCartQuote(ctx context.Context, db *sql.DB, cartID uuid.UUID) (*domain.QuoteSnapshot, error) {
	var q domain.QuoteSnapshot
	var lineQuotes JSONLineQuotes
	err := db.QueryRowContext(ctx, `
		SELECT quote_id, currency, subtotal_minor, discount_minor, tax_minor, delivery_minor,
			service_minor, packaging_minor, tip_minor, total_minor, line_quotes, quoted_at
		FROM cart_quotes WHERE cart_id=$1`, cartID).Scan(
		&q.QuoteID, &q.Currency, &q.SubtotalMinor, &q.DiscountMinor, &q.TaxMinor, &q.DeliveryMinor,
		&q.ServiceMinor, &q.PackagingMinor, &q.TipMinor, &q.TotalMinor, &lineQuotes, &q.QuotedAt,
	)
	if isNoRows(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	q.LineQuotes = []domain.LineQuote(lineQuotes)
	q.QuotedAt = q.QuotedAt.UTC()
	return &q, nil
}

func loadReservationRefs(ctx context.Context, db *sql.DB, cartID uuid.UUID) ([]domain.ReservationRef, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, reservation_ref, idempotency_key, expires_at, created_at, released_at
		FROM cart_reservation_refs WHERE cart_id=$1 ORDER BY created_at`, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.ReservationRef, 0)
	for rows.Next() {
		var ref domain.ReservationRef
		var expires, released sql.NullTime
		if err := rows.Scan(
			&ref.ID, &ref.ReservationRef, &ref.IdempotencyKey, &expires, &ref.CreatedAt, &released,
		); err != nil {
			return nil, err
		}
		ref.ExpiresAt = scanNullTime(expires)
		ref.ReleasedAt = scanNullTime(released)
		ref.CreatedAt = ref.CreatedAt.UTC()
		out = append(out, ref)
	}
	return out, rows.Err()
}

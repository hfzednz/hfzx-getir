package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/checkout-service/internal/app/ports"
	"github.com/nexora/checkout-service/internal/domain"
)

// SessionRepo persists checkout sessions in PostgreSQL.
type SessionRepo struct{ DB *sql.DB }

var _ ports.CheckoutRepo = (*SessionRepo)(nil)

const sessionColumns = `
	id, tenant_id, cart_id, principal_id, status, delivery_option,
	address_snapshot, slot_snapshot, gift_prefs, invoice_prefs, substitutions,
	notes, tip_minor, currency, validation_results, quote_snapshot,
	order_id, idempotency_key, recovery_token, city_id, coupon_codes, version, metadata,
	created_at, updated_at, completed_at, abandoned_at, failed_at
`

func (r *SessionRepo) Create(ctx context.Context, s domain.Session) error {
	delivery := nullString(string(s.DeliveryOption))
	subs := string(s.Substitutions)
	if subs == "" {
		subs = string(domain.SubstitutionAsk)
	}
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO checkout_sessions (
			id, tenant_id, cart_id, principal_id, status, delivery_option,
			address_snapshot, slot_snapshot, gift_prefs, invoice_prefs, substitutions,
			notes, tip_minor, currency, validation_results, quote_snapshot,
			order_id, idempotency_key, recovery_token, city_id, coupon_codes, version, metadata,
			created_at, updated_at, completed_at, abandoned_at, failed_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,
			$7,$8,$9,$10,$11,
			$12,$13,$14,$15,$16,
			$17,$18,$19,$20,$21,$22,$23,
			$24,$25,$26,$27,$28
		)`,
		s.ID, s.TenantID, s.CartID, s.PrincipalID, string(s.Status), delivery,
		JSONAddress(s.Address), JSONSlot(s.Slot), JSONGift(s.Gift), JSONInvoice(s.Invoice),
		subs,
		s.Notes, s.TipMinor, s.Currency, JSONValidation(s.Validation), JSONQuote(s.Quote),
		s.OrderID, s.IdempotencyKey, s.RecoveryToken, s.CityID, StringArray(s.CouponCodes), s.Version, JSONMap(s.Metadata),
		s.CreatedAt, s.UpdatedAt, nullTime(s.CompletedAt), nullTime(s.AbandonedAt), nullTime(s.FailedAt),
	)
	return mapUniqueViolation(err)
}

func (r *SessionRepo) Update(ctx context.Context, s domain.Session) error {
	var curVersion int64
	err := r.DB.QueryRowContext(ctx,
		`SELECT version FROM checkout_sessions WHERE id=$1 AND tenant_id=$2`,
		s.ID, s.TenantID,
	).Scan(&curVersion)
	if isNoRows(err) {
		return domain.ErrNotFound
	}
	if err != nil {
		return err
	}
	if s.Version < curVersion {
		return domain.ErrVersionConflict
	}

	delivery := nullString(string(s.DeliveryOption))
	subs := string(s.Substitutions)
	if subs == "" {
		subs = string(domain.SubstitutionAsk)
	}
	res, err := r.DB.ExecContext(ctx, `
		UPDATE checkout_sessions SET
			cart_id=$1, principal_id=$2, status=$3, delivery_option=$4,
			address_snapshot=$5, slot_snapshot=$6, gift_prefs=$7, invoice_prefs=$8, substitutions=$9,
			notes=$10, tip_minor=$11, currency=$12, validation_results=$13, quote_snapshot=$14,
			order_id=$15, recovery_token=$16, city_id=$17, coupon_codes=$18, version=$19, metadata=$20,
			updated_at=$21, completed_at=$22, abandoned_at=$23, failed_at=$24
		WHERE id=$25 AND tenant_id=$26`,
		s.CartID, s.PrincipalID, string(s.Status), delivery,
		JSONAddress(s.Address), JSONSlot(s.Slot), JSONGift(s.Gift), JSONInvoice(s.Invoice),
		subs,
		s.Notes, s.TipMinor, s.Currency, JSONValidation(s.Validation), JSONQuote(s.Quote),
		s.OrderID, s.RecoveryToken, s.CityID, StringArray(s.CouponCodes), s.Version, JSONMap(s.Metadata),
		s.UpdatedAt, nullTime(s.CompletedAt), nullTime(s.AbandonedAt), nullTime(s.FailedAt),
		s.ID, s.TenantID,
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
	return nil
}

func (r *SessionRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Session, error) {
	return r.scanSession(ctx, r.DB.QueryRowContext(ctx, `
		SELECT `+sessionColumns+` FROM checkout_sessions WHERE id=$1 AND tenant_id=$2`, id, tenantID))
}

func (r *SessionRepo) GetByIdempotencyKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.Session, error) {
	return r.scanSession(ctx, r.DB.QueryRowContext(ctx, `
		SELECT `+sessionColumns+` FROM checkout_sessions WHERE tenant_id=$1 AND idempotency_key=$2`, tenantID, key))
}

func (r *SessionRepo) GetByRecoveryToken(ctx context.Context, token string) (domain.Session, error) {
	return r.scanSession(ctx, r.DB.QueryRowContext(ctx, `
		SELECT `+sessionColumns+` FROM checkout_sessions WHERE recovery_token=$1 AND recovery_token <> ''`, token))
}

func (r *SessionRepo) List(ctx context.Context, f ports.SessionFilter) ([]domain.Session, int, error) {
	where := []string{"tenant_id = $1"}
	args := []any{f.TenantID}
	argN := 2
	if f.PrincipalID != nil {
		where = append(where, fmt.Sprintf("principal_id = $%d", argN))
		args = append(args, *f.PrincipalID)
		argN++
	}
	if f.Status != nil {
		where = append(where, fmt.Sprintf("status = $%d", argN))
		args = append(args, string(*f.Status))
		argN++
	}
	if f.Query != "" {
		q := "%" + strings.ToLower(f.Query) + "%"
		where = append(where, fmt.Sprintf(
			`(LOWER(id::text) LIKE $%d OR LOWER(cart_id::text) LIKE $%d OR LOWER(order_id) LIKE $%d OR LOWER(status::text) LIKE $%d)`,
			argN, argN, argN, argN))
		args = append(args, q)
		argN++
	}
	whereSQL := strings.Join(where, " AND ")

	var total int
	if err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM checkout_sessions WHERE `+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}
	args = append(args, limit, offset)
	rows, err := r.DB.QueryContext(ctx, `
		SELECT `+sessionColumns+` FROM checkout_sessions WHERE `+whereSQL+`
		ORDER BY created_at DESC
		LIMIT $`+fmt.Sprint(argN)+` OFFSET $`+fmt.Sprint(argN+1), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]domain.Session, 0)
	for rows.Next() {
		s, err := scanSessionRow(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, s)
	}
	return out, total, rows.Err()
}

func (r *SessionRepo) CountByStatus(ctx context.Context, tenantID uuid.UUID) (map[domain.SessionStatus]int, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT status, COUNT(*) FROM checkout_sessions WHERE tenant_id=$1 GROUP BY status`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[domain.SessionStatus]int{}
	for rows.Next() {
		var status string
		var n int
		if err := rows.Scan(&status, &n); err != nil {
			return nil, err
		}
		out[domain.SessionStatus(status)] = n
	}
	return out, rows.Err()
}

func (r *SessionRepo) scanSession(ctx context.Context, row rowScanner) (domain.Session, error) {
	_ = ctx
	s, err := scanSessionRow(row)
	if isNoRows(err) {
		return domain.Session{}, domain.ErrNotFound
	}
	return s, err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanSessionRow(row rowScanner) (domain.Session, error) {
	var s domain.Session
	var status string
	var delivery sql.NullString
	var address JSONAddress
	var slot JSONSlot
	var gift JSONGift
	var invoice JSONInvoice
	var substitutions string
	var validation JSONValidation
	var quote JSONQuote
	var coupons StringArray
	var meta JSONMap
	var completed, abandoned, failed sql.NullTime
	err := row.Scan(
		&s.ID, &s.TenantID, &s.CartID, &s.PrincipalID, &status, &delivery,
		&address, &slot, &gift, &invoice, &substitutions,
		&s.Notes, &s.TipMinor, &s.Currency, &validation, &quote,
		&s.OrderID, &s.IdempotencyKey, &s.RecoveryToken, &s.CityID, &coupons, &s.Version, &meta,
		&s.CreatedAt, &s.UpdatedAt, &completed, &abandoned, &failed,
	)
	if err != nil {
		return domain.Session{}, err
	}
	s.Status = domain.SessionStatus(status)
	s.DeliveryOption = domain.DeliveryOption(scanNullString(delivery))
	s.Address = domain.AddressSnapshot(address)
	s.Slot = domain.SlotSnapshot(slot)
	s.Gift = domain.GiftPrefs(gift)
	s.Invoice = domain.InvoicePrefs(invoice)
	s.Substitutions = domain.SubstitutionPolicy(substitutions)
	s.Validation = domain.ValidationResults(validation)
	s.Quote = domain.QuoteSnapshot(quote)
	s.CouponCodes = []string(coupons)
	if s.CouponCodes == nil {
		s.CouponCodes = []string{}
	}
	s.Metadata = map[string]any(meta)
	if s.Metadata == nil {
		s.Metadata = map[string]any{}
	}
	s.CompletedAt = scanNullTime(completed)
	s.AbandonedAt = scanNullTime(abandoned)
	s.FailedAt = scanNullTime(failed)
	s.CreatedAt = s.CreatedAt.UTC()
	s.UpdatedAt = s.UpdatedAt.UTC()
	if !s.Quote.QuotedAt.IsZero() {
		s.Quote.QuotedAt = s.Quote.QuotedAt.UTC()
	}
	if !s.Validation.CheckedAt.IsZero() {
		s.Validation.CheckedAt = s.Validation.CheckedAt.UTC()
	}
	return s, nil
}

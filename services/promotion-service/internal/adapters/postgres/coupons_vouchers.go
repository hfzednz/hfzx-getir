package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/promotion-service/internal/app/ports"
	"github.com/nexora/promotion-service/internal/domain"
)

// CouponRepo persists coupons and redemptions.
type CouponRepo struct{ DB *sql.DB }

func (r *CouponRepo) Create(ctx context.Context, c domain.Coupon) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO coupons (
			id, tenant_id, promotion_id, code, kind, max_redemptions, redeemed_count, principal_id,
			starts_at, ends_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		c.ID, c.TenantID, c.PromotionID, c.Code, string(c.Kind), c.MaxRedemptions, c.RedeemedCount, nullUUID(c.PrincipalID),
		nullTime(c.StartsAt), nullTime(c.EndsAt), c.CreatedAt.UTC(), c.UpdatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *CouponRepo) Update(ctx context.Context, c domain.Coupon) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE coupons SET kind=$3, max_redemptions=$4, redeemed_count=$5, principal_id=$6,
			starts_at=$7, ends_at=$8, updated_at=$9
		WHERE id=$1 AND tenant_id=$2`,
		c.ID, c.TenantID, string(c.Kind), c.MaxRedemptions, c.RedeemedCount, nullUUID(c.PrincipalID),
		nullTime(c.StartsAt), nullTime(c.EndsAt), c.UpdatedAt.UTC())
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *CouponRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Coupon, error) {
	return r.scan(r.DB.QueryRowContext(ctx, couponSelect+` WHERE id=$1 AND tenant_id=$2`, id, tenantID))
}

func (r *CouponRepo) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.Coupon, error) {
	return r.scan(r.DB.QueryRowContext(ctx, couponSelect+` WHERE tenant_id=$1 AND code=$2`, tenantID, code))
}

func (r *CouponRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Coupon, error) {
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := r.DB.QueryContext(ctx, couponSelect+` WHERE tenant_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Coupon{}
	for rows.Next() {
		c, err := scanCoupon(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *CouponRepo) CreateRedemption(ctx context.Context, red domain.CouponRedemption) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO coupon_redemptions (
			id, tenant_id, coupon_id, principal_id, idempotency_key, order_ref, discount_minor, currency, redeemed_at, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		red.ID, red.TenantID, red.CouponID, red.PrincipalID, red.IdempotencyKey, red.OrderRef,
		red.DiscountMinor, red.Currency, red.RedeemedAt.UTC(), red.CreatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *CouponRepo) GetRedemptionByIdempotency(ctx context.Context, tenantID uuid.UUID, key string) (domain.CouponRedemption, error) {
	var red domain.CouponRedemption
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, coupon_id, principal_id, idempotency_key, order_ref, discount_minor, currency, redeemed_at, created_at
		FROM coupon_redemptions WHERE tenant_id=$1 AND idempotency_key=$2`, tenantID, key).Scan(
		&red.ID, &red.TenantID, &red.CouponID, &red.PrincipalID, &red.IdempotencyKey, &red.OrderRef,
		&red.DiscountMinor, &red.Currency, &red.RedeemedAt, &red.CreatedAt)
	if isNoRows(err) {
		return domain.CouponRedemption{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.CouponRedemption{}, err
	}
	red.RedeemedAt = red.RedeemedAt.UTC()
	red.CreatedAt = red.CreatedAt.UTC()
	return red, nil
}

const couponSelect = `
	SELECT id, tenant_id, promotion_id, code, kind, max_redemptions, redeemed_count, principal_id,
		starts_at, ends_at, created_at, updated_at FROM coupons`

type couponScanner interface {
	Scan(dest ...any) error
}

func (r *CouponRepo) scan(row *sql.Row) (domain.Coupon, error) {
	c, err := scanCoupon(row)
	if isNoRows(err) {
		return domain.Coupon{}, domain.ErrNotFound
	}
	return c, err
}

func scanCoupon(row couponScanner) (domain.Coupon, error) {
	var c domain.Coupon
	var kind string
	var principal uuid.NullUUID
	var starts, ends sql.NullTime
	err := row.Scan(
		&c.ID, &c.TenantID, &c.PromotionID, &c.Code, &kind, &c.MaxRedemptions, &c.RedeemedCount, &principal,
		&starts, &ends, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return domain.Coupon{}, err
	}
	c.Kind = domain.CouponKind(kind)
	c.PrincipalID = scanNullUUID(principal)
	c.StartsAt = scanNullTime(starts)
	c.EndsAt = scanNullTime(ends)
	c.CreatedAt = c.CreatedAt.UTC()
	c.UpdatedAt = c.UpdatedAt.UTC()
	return c, nil
}

// VoucherRepo persists vouchers and redemptions.
type VoucherRepo struct{ DB *sql.DB }

func (r *VoucherRepo) Create(ctx context.Context, v domain.Voucher) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO vouchers (
			id, tenant_id, promotion_id, code, principal_id, status, value_minor, currency, remaining_minor,
			starts_at, ends_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		v.ID, v.TenantID, nullUUID(v.PromotionID), v.Code, v.PrincipalID, string(v.Status), v.ValueMinor, v.Currency,
		v.RemainingMinor, nullTime(v.StartsAt), nullTime(v.EndsAt), v.CreatedAt.UTC(), v.UpdatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *VoucherRepo) Update(ctx context.Context, v domain.Voucher) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE vouchers SET promotion_id=$3, status=$4, value_minor=$5, remaining_minor=$6,
			starts_at=$7, ends_at=$8, updated_at=$9
		WHERE id=$1 AND tenant_id=$2`,
		v.ID, v.TenantID, nullUUID(v.PromotionID), string(v.Status), v.ValueMinor, v.RemainingMinor,
		nullTime(v.StartsAt), nullTime(v.EndsAt), v.UpdatedAt.UTC())
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *VoucherRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Voucher, error) {
	return r.scan(r.DB.QueryRowContext(ctx, voucherSelect+` WHERE id=$1 AND tenant_id=$2`, id, tenantID))
}

func (r *VoucherRepo) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.Voucher, error) {
	return r.scan(r.DB.QueryRowContext(ctx, voucherSelect+` WHERE tenant_id=$1 AND code=$2`, tenantID, code))
}

func (r *VoucherRepo) CreateRedemption(ctx context.Context, red domain.VoucherRedemption) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO voucher_redemptions (
			id, tenant_id, voucher_id, principal_id, idempotency_key, order_ref, amount_minor, currency, redeemed_at, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		red.ID, red.TenantID, red.VoucherID, red.PrincipalID, red.IdempotencyKey, red.OrderRef,
		red.AmountMinor, red.Currency, red.RedeemedAt.UTC(), red.CreatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *VoucherRepo) GetRedemptionByIdempotency(ctx context.Context, tenantID uuid.UUID, key string) (domain.VoucherRedemption, error) {
	var red domain.VoucherRedemption
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, voucher_id, principal_id, idempotency_key, order_ref, amount_minor, currency, redeemed_at, created_at
		FROM voucher_redemptions WHERE tenant_id=$1 AND idempotency_key=$2`, tenantID, key).Scan(
		&red.ID, &red.TenantID, &red.VoucherID, &red.PrincipalID, &red.IdempotencyKey, &red.OrderRef,
		&red.AmountMinor, &red.Currency, &red.RedeemedAt, &red.CreatedAt)
	if isNoRows(err) {
		return domain.VoucherRedemption{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.VoucherRedemption{}, err
	}
	red.RedeemedAt = red.RedeemedAt.UTC()
	red.CreatedAt = red.CreatedAt.UTC()
	return red, nil
}

const voucherSelect = `
	SELECT id, tenant_id, promotion_id, code, principal_id, status, value_minor, currency, remaining_minor,
		starts_at, ends_at, created_at, updated_at FROM vouchers`

func (r *VoucherRepo) scan(row *sql.Row) (domain.Voucher, error) {
	var v domain.Voucher
	var status string
	var promo uuid.NullUUID
	var starts, ends sql.NullTime
	err := row.Scan(
		&v.ID, &v.TenantID, &promo, &v.Code, &v.PrincipalID, &status, &v.ValueMinor, &v.Currency, &v.RemainingMinor,
		&starts, &ends, &v.CreatedAt, &v.UpdatedAt)
	if isNoRows(err) {
		return domain.Voucher{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Voucher{}, err
	}
	v.Status = domain.VoucherStatus(status)
	v.PromotionID = scanNullUUID(promo)
	v.StartsAt = scanNullTime(starts)
	v.EndsAt = scanNullTime(ends)
	v.CreatedAt = v.CreatedAt.UTC()
	v.UpdatedAt = v.UpdatedAt.UTC()
	return v, nil
}

// UsageRepo tracks usage counters.
type UsageRepo struct{ DB *sql.DB }

func (r *UsageRepo) Get(ctx context.Context, tenantID, promotionID uuid.UUID, scope domain.UsageScope, scopeKey string) (domain.UsageCounter, error) {
	var u domain.UsageCounter
	var sc string
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, promotion_id, scope, scope_key, count, updated_at
		FROM usage_counters WHERE tenant_id=$1 AND promotion_id=$2 AND scope=$3 AND scope_key=$4`,
		tenantID, promotionID, string(scope), scopeKey).Scan(
		&u.ID, &u.TenantID, &u.PromotionID, &sc, &u.ScopeKey, &u.Count, &u.UpdatedAt)
	if isNoRows(err) {
		return domain.UsageCounter{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.UsageCounter{}, err
	}
	u.Scope = domain.UsageScope(sc)
	u.UpdatedAt = u.UpdatedAt.UTC()
	return u, nil
}

func (r *UsageRepo) Increment(ctx context.Context, tenantID, promotionID uuid.UUID, scope domain.UsageScope, scopeKey string) (domain.UsageCounter, error) {
	id := uuid.New()
	now := time.Now().UTC()
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO usage_counters (id, tenant_id, promotion_id, scope, scope_key, count, updated_at)
		VALUES ($1,$2,$3,$4,$5,1,$6)
		ON CONFLICT (tenant_id, promotion_id, scope, scope_key)
		DO UPDATE SET count = usage_counters.count + 1, updated_at = EXCLUDED.updated_at`,
		id, tenantID, promotionID, string(scope), scopeKey, now)
	if err != nil {
		return domain.UsageCounter{}, err
	}
	return r.Get(ctx, tenantID, promotionID, scope, scopeKey)
}

// SimulationRepo stores evaluate simulations.
type SimulationRepo struct{ DB *sql.DB }

func (r *SimulationRepo) Create(ctx context.Context, s domain.Simulation) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO simulations (id, tenant_id, request_payload, result_payload, created_at)
		VALUES ($1,$2,$3,$4,$5)`,
		s.ID, s.TenantID, JSONMap(s.RequestPayload), JSONMap(s.ResultPayload), s.CreatedAt.UTC())
	return err
}

func (r *SimulationRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Simulation, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, request_payload, result_payload, created_at
		FROM simulations WHERE tenant_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Simulation{}
	for rows.Next() {
		var s domain.Simulation
		var req, res JSONMap
		if err := rows.Scan(&s.ID, &s.TenantID, &req, &res, &s.CreatedAt); err != nil {
			return nil, err
		}
		s.RequestPayload = map[string]any(req)
		s.ResultPayload = map[string]any(res)
		s.CreatedAt = s.CreatedAt.UTC()
		out = append(out, s)
	}
	return out, rows.Err()
}

// OutboxRepo persists transactional outbox rows.
type OutboxRepo struct{ DB *sql.DB }

func (r *OutboxRepo) Enqueue(ctx context.Context, m domain.OutboxMessage) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO outbox (
			id, tenant_id, aggregate_id, topic, key, payload, status, attempts, last_error,
			created_at, updated_at, published_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		m.ID, m.TenantID, m.AggregateID, m.Topic, m.Key, JSONMap(m.Payload), string(m.Status), m.Attempts, m.LastError,
		m.CreatedAt.UTC(), m.UpdatedAt.UTC(), nullTime(m.PublishedAt))
	return err
}

func (r *OutboxRepo) ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, aggregate_id, topic, key, payload, status, attempts, last_error,
			created_at, updated_at, published_at
		FROM outbox WHERE status='pending' ORDER BY created_at ASC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.OutboxMessage{}
	for rows.Next() {
		var m domain.OutboxMessage
		var status string
		var payload JSONMap
		var published sql.NullTime
		if err := rows.Scan(
			&m.ID, &m.TenantID, &m.AggregateID, &m.Topic, &m.Key, &payload, &status, &m.Attempts, &m.LastError,
			&m.CreatedAt, &m.UpdatedAt, &published); err != nil {
			return nil, err
		}
		m.Status = domain.OutboxStatus(status)
		m.Payload = map[string]any(payload)
		m.CreatedAt = m.CreatedAt.UTC()
		m.UpdatedAt = m.UpdatedAt.UTC()
		m.PublishedAt = scanNullTime(published)
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *OutboxRepo) MarkPublished(ctx context.Context, id uuid.UUID, at time.Time) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE outbox SET status='published', published_at=$2, updated_at=$2 WHERE id=$1`, id, at.UTC())
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// Repos groups promotion persistence adapters.
type Repos struct {
	Campaigns   *CampaignRepo
	Promotions  *PromotionRepo
	Rules       *RuleRepo
	Coupons     *CouponRepo
	Vouchers    *VoucherRepo
	Usage       *UsageRepo
	Simulations *SimulationRepo
	Outbox      *OutboxRepo
}

// NewRepos constructs postgres-backed repositories.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Campaigns:   &CampaignRepo{DB: db},
		Promotions:  &PromotionRepo{DB: db},
		Rules:       &RuleRepo{DB: db},
		Coupons:     &CouponRepo{DB: db},
		Vouchers:    &VoucherRepo{DB: db},
		Usage:       &UsageRepo{DB: db},
		Simulations: &SimulationRepo{DB: db},
		Outbox:      &OutboxRepo{DB: db},
	}
}

var (
	_ ports.CouponRepository     = (*CouponRepo)(nil)
	_ ports.VoucherRepository    = (*VoucherRepo)(nil)
	_ ports.UsageRepository      = (*UsageRepo)(nil)
	_ ports.SimulationRepository = (*SimulationRepo)(nil)
	_ ports.OutboxRepository     = (*OutboxRepo)(nil)
)

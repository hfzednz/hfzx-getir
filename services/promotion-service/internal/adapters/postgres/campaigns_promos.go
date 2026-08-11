package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/promotion-service/internal/app/ports"
	"github.com/nexora/promotion-service/internal/domain"
)

// CampaignRepo persists campaigns.
type CampaignRepo struct{ DB *sql.DB }

func (r *CampaignRepo) Create(ctx context.Context, c domain.Campaign) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO campaigns (id, tenant_id, name, description, status, starts_at, ends_at, created_at, updated_at, version)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		c.ID, c.TenantID, c.Name, c.Description, string(c.Status), nullTime(c.StartsAt), nullTime(c.EndsAt),
		c.CreatedAt.UTC(), c.UpdatedAt.UTC(), c.Version)
	return mapUniqueViolation(err)
}

func (r *CampaignRepo) Update(ctx context.Context, c domain.Campaign) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE campaigns SET name=$3, description=$4, status=$5, starts_at=$6, ends_at=$7, updated_at=$8, version=$9
		WHERE id=$1 AND tenant_id=$2`,
		c.ID, c.TenantID, c.Name, c.Description, string(c.Status), nullTime(c.StartsAt), nullTime(c.EndsAt),
		c.UpdatedAt.UTC(), c.Version)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *CampaignRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Campaign, error) {
	return r.scan(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, description, status, starts_at, ends_at, created_at, updated_at, version
		FROM campaigns WHERE id=$1 AND tenant_id=$2`, id, tenantID))
}

func (r *CampaignRepo) List(ctx context.Context, tenantID uuid.UUID, status *domain.CampaignStatus, limit, offset int) ([]domain.Campaign, error) {
	if limit <= 0 {
		limit = 50
	}
	where := `WHERE tenant_id=$1`
	args := []any{tenantID}
	if status != nil {
		where += ` AND status=$2`
		args = append(args, string(*status))
	}
	args = append(args, limit, offset)
	q := fmt.Sprintf(`SELECT id, tenant_id, name, description, status, starts_at, ends_at, created_at, updated_at, version
		FROM campaigns %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, len(args)-1, len(args))
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCampaigns(rows)
}

func (r *CampaignRepo) ListActive(ctx context.Context, tenantID uuid.UUID, now time.Time) ([]domain.Campaign, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, name, description, status, starts_at, ends_at, created_at, updated_at, version
		FROM campaigns
		WHERE tenant_id=$1 AND status='active'
		  AND (starts_at IS NULL OR starts_at <= $2)
		  AND (ends_at IS NULL OR ends_at > $2)
		ORDER BY created_at DESC`, tenantID, now.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCampaigns(rows)
}

func (r *CampaignRepo) scan(row *sql.Row) (domain.Campaign, error) {
	c, err := scanCampaignRow(row)
	if isNoRows(err) {
		return domain.Campaign{}, domain.ErrNotFound
	}
	return c, err
}

type campScanner interface{ Scan(dest ...any) error }

func scanCampaignRow(row campScanner) (domain.Campaign, error) {
	var c domain.Campaign
	var status string
	var starts, ends sql.NullTime
	err := row.Scan(&c.ID, &c.TenantID, &c.Name, &c.Description, &status, &starts, &ends, &c.CreatedAt, &c.UpdatedAt, &c.Version)
	if err != nil {
		return domain.Campaign{}, err
	}
	c.Status = domain.CampaignStatus(status)
	c.StartsAt = scanNullTime(starts)
	c.EndsAt = scanNullTime(ends)
	c.CreatedAt = c.CreatedAt.UTC()
	c.UpdatedAt = c.UpdatedAt.UTC()
	return c, nil
}

func scanCampaigns(rows *sql.Rows) ([]domain.Campaign, error) {
	out := []domain.Campaign{}
	for rows.Next() {
		c, err := scanCampaignRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// PromotionRepo persists promotions.
type PromotionRepo struct{ DB *sql.DB }

func (r *PromotionRepo) Create(ctx context.Context, p domain.Promotion) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO promotions (
			id, tenant_id, campaign_id, name, type, percent_off, fixed_off_minor, buy_qty, get_qty,
			threshold_minor, gift_variant_id, max_discount_minor, priority, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		p.ID, p.TenantID, p.CampaignID, p.Name, string(p.Type), p.PercentOff, p.FixedOffMinor, p.BuyQty, p.GetQty,
		p.ThresholdMinor, p.GiftVariantID, p.MaxDiscountMinor, p.Priority, p.CreatedAt.UTC(), p.UpdatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *PromotionRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Promotion, error) {
	return r.scan(r.DB.QueryRowContext(ctx, promoSelect+` WHERE id=$1 AND tenant_id=$2`, id, tenantID))
}

func (r *PromotionRepo) ListByCampaign(ctx context.Context, tenantID, campaignID uuid.UUID) ([]domain.Promotion, error) {
	rows, err := r.DB.QueryContext(ctx, promoSelect+` WHERE tenant_id=$1 AND campaign_id=$2 ORDER BY priority DESC`, tenantID, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPromos(rows)
}

func (r *PromotionRepo) ListByIDs(ctx context.Context, tenantID uuid.UUID, ids []uuid.UUID) ([]domain.Promotion, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	out := []domain.Promotion{}
	for _, id := range ids {
		p, err := r.GetByID(ctx, tenantID, id)
		if err != nil {
			if errorsIsNotFound(err) {
				continue
			}
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func errorsIsNotFound(err error) bool {
	return err == domain.ErrNotFound
}

const promoSelect = `
	SELECT id, tenant_id, campaign_id, name, type, percent_off, fixed_off_minor, buy_qty, get_qty,
		threshold_minor, gift_variant_id, max_discount_minor, priority, created_at, updated_at FROM promotions`

func (r *PromotionRepo) scan(row *sql.Row) (domain.Promotion, error) {
	p, err := scanPromoRow(row)
	if isNoRows(err) {
		return domain.Promotion{}, domain.ErrNotFound
	}
	return p, err
}

type promoScanner interface{ Scan(dest ...any) error }

func scanPromoRow(row promoScanner) (domain.Promotion, error) {
	var p domain.Promotion
	var typ string
	err := row.Scan(
		&p.ID, &p.TenantID, &p.CampaignID, &p.Name, &typ, &p.PercentOff, &p.FixedOffMinor, &p.BuyQty, &p.GetQty,
		&p.ThresholdMinor, &p.GiftVariantID, &p.MaxDiscountMinor, &p.Priority, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return domain.Promotion{}, err
	}
	p.Type = domain.PromotionType(typ)
	p.CreatedAt = p.CreatedAt.UTC()
	p.UpdatedAt = p.UpdatedAt.UTC()
	return p, nil
}

func scanPromos(rows *sql.Rows) ([]domain.Promotion, error) {
	out := []domain.Promotion{}
	for rows.Next() {
		p, err := scanPromoRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// RuleRepo persists rules.
type RuleRepo struct{ DB *sql.DB }

func (r *RuleRepo) Create(ctx context.Context, rule domain.Rule) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO rules (
			id, tenant_id, promotion_id, priority, stack_group, stackable, exclude_promotion_ids,
			variant_ids, category_ids, brand_ids, segment_ids, global_limit, per_user_limit,
			per_order_limit, per_device_limit, min_qty, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`,
		rule.ID, rule.TenantID, rule.PromotionID, rule.Priority, rule.StackGroup, rule.Stackable,
		UUIDArray(rule.ExcludePromotionIDs), TextArray(rule.VariantIDs), TextArray(rule.CategoryIDs),
		TextArray(rule.BrandIDs), TextArray(rule.SegmentIDs), rule.GlobalLimit, rule.PerUserLimit,
		rule.PerOrderLimit, rule.PerDeviceLimit, rule.MinQty, rule.CreatedAt.UTC(), rule.UpdatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *RuleRepo) GetByPromotionID(ctx context.Context, tenantID, promotionID uuid.UUID) (domain.Rule, error) {
	return r.scan(r.DB.QueryRowContext(ctx, ruleSelect+` WHERE tenant_id=$1 AND promotion_id=$2`, tenantID, promotionID))
}

func (r *RuleRepo) ListByPromotionIDs(ctx context.Context, tenantID uuid.UUID, promotionIDs []uuid.UUID) ([]domain.Rule, error) {
	out := []domain.Rule{}
	for _, id := range promotionIDs {
		rule, err := r.GetByPromotionID(ctx, tenantID, id)
		if err != nil {
			if errorsIsNotFound(err) {
				continue
			}
			return nil, err
		}
		out = append(out, rule)
	}
	return out, nil
}

const ruleSelect = `
	SELECT id, tenant_id, promotion_id, priority, stack_group, stackable, exclude_promotion_ids,
		variant_ids, category_ids, brand_ids, segment_ids, global_limit, per_user_limit,
		per_order_limit, per_device_limit, min_qty, created_at, updated_at FROM rules`

func (r *RuleRepo) scan(row *sql.Row) (domain.Rule, error) {
	var rule domain.Rule
	var excl UUIDArray
	var variants, cats, brands, segs TextArray
	err := row.Scan(
		&rule.ID, &rule.TenantID, &rule.PromotionID, &rule.Priority, &rule.StackGroup, &rule.Stackable, &excl,
		&variants, &cats, &brands, &segs, &rule.GlobalLimit, &rule.PerUserLimit,
		&rule.PerOrderLimit, &rule.PerDeviceLimit, &rule.MinQty, &rule.CreatedAt, &rule.UpdatedAt)
	if isNoRows(err) {
		return domain.Rule{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Rule{}, err
	}
	rule.ExcludePromotionIDs = []uuid.UUID(excl)
	rule.VariantIDs = []string(variants)
	rule.CategoryIDs = []string(cats)
	rule.BrandIDs = []string(brands)
	rule.SegmentIDs = []string(segs)
	rule.CreatedAt = rule.CreatedAt.UTC()
	rule.UpdatedAt = rule.UpdatedAt.UTC()
	return rule, nil
}

var (
	_ ports.CampaignRepository  = (*CampaignRepo)(nil)
	_ ports.PromotionRepository = (*PromotionRepo)(nil)
	_ ports.RuleRepository      = (*RuleRepo)(nil)
)

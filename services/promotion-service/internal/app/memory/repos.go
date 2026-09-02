package memory

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/promotion-service/internal/app/ports"
	"github.com/nexora/promotion-service/internal/domain"
)

// NewRepos wires all memory repositories from a shared store.
func NewRepos(s *Store) (
	ports.CampaignRepository,
	ports.PromotionRepository,
	ports.RuleRepository,
	ports.CouponRepository,
	ports.VoucherRepository,
	ports.UsageRepository,
	ports.SimulationRepository,
	ports.OutboxRepository,
) {
	return &CampaignRepo{S: s}, &PromotionRepo{S: s}, &RuleRepo{S: s},
		&CouponRepo{S: s}, &VoucherRepo{S: s}, &UsageRepo{S: s},
		&SimulationRepo{S: s}, &OutboxRepo{S: s}
}

// --- Campaign ---

type CampaignRepo struct{ S *Store }

func (r *CampaignRepo) Create(_ context.Context, c domain.Campaign) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Campaigns[c.ID]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.Campaigns[c.ID] = cloneCampaign(c)
	return nil
}

func (r *CampaignRepo) Update(_ context.Context, c domain.Campaign) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	cur, ok := r.S.Campaigns[c.ID]
	if !ok || cur.TenantID != c.TenantID {
		return domain.ErrNotFound
	}
	r.S.Campaigns[c.ID] = cloneCampaign(c)
	return nil
}

func (r *CampaignRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.Campaign, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	c, ok := r.S.Campaigns[id]
	if !ok || c.TenantID != tenantID {
		return domain.Campaign{}, domain.ErrNotFound
	}
	return cloneCampaign(c), nil
}

func (r *CampaignRepo) List(_ context.Context, tenantID uuid.UUID, status *domain.CampaignStatus, limit, offset int) ([]domain.Campaign, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.Campaign
	for _, c := range r.S.Campaigns {
		if c.TenantID != tenantID {
			continue
		}
		if status != nil && c.Status != *status {
			continue
		}
		out = append(out, cloneCampaign(c))
	}
	if offset > len(out) {
		return nil, nil
	}
	out = out[offset:]
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (r *CampaignRepo) ListActive(_ context.Context, tenantID uuid.UUID, now time.Time) ([]domain.Campaign, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.Campaign
	for _, c := range r.S.Campaigns {
		if c.TenantID != tenantID {
			continue
		}
		if c.IsActiveAt(now) {
			out = append(out, cloneCampaign(c))
		}
	}
	return out, nil
}

func cloneCampaign(c domain.Campaign) domain.Campaign {
	out := c
	if c.StartsAt != nil {
		t := *c.StartsAt
		out.StartsAt = &t
	}
	if c.EndsAt != nil {
		t := *c.EndsAt
		out.EndsAt = &t
	}
	return out
}

var _ ports.CampaignRepository = (*CampaignRepo)(nil)

// --- Promotion ---

type PromotionRepo struct{ S *Store }

func (r *PromotionRepo) Create(_ context.Context, p domain.Promotion) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Promotions[p.ID]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.Promotions[p.ID] = p
	return nil
}

func (r *PromotionRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.Promotion, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	p, ok := r.S.Promotions[id]
	if !ok || p.TenantID != tenantID {
		return domain.Promotion{}, domain.ErrNotFound
	}
	return p, nil
}

func (r *PromotionRepo) ListByCampaign(_ context.Context, tenantID, campaignID uuid.UUID) ([]domain.Promotion, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.Promotion
	for _, p := range r.S.Promotions {
		if p.TenantID == tenantID && p.CampaignID == campaignID {
			out = append(out, p)
		}
	}
	return out, nil
}

func (r *PromotionRepo) ListByIDs(_ context.Context, tenantID uuid.UUID, ids []uuid.UUID) ([]domain.Promotion, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.Promotion
	for _, id := range ids {
		if p, ok := r.S.Promotions[id]; ok && p.TenantID == tenantID {
			out = append(out, p)
		}
	}
	return out, nil
}

var _ ports.PromotionRepository = (*PromotionRepo)(nil)

// --- Rule ---

type RuleRepo struct{ S *Store }

func (r *RuleRepo) Create(_ context.Context, rule domain.Rule) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Rules[rule.ID]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.Rules[rule.ID] = cloneRule(rule)
	r.S.RulesByPromo[rule.PromotionID] = rule.ID
	return nil
}

func (r *RuleRepo) GetByPromotionID(_ context.Context, tenantID, promotionID uuid.UUID) (domain.Rule, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	rid, ok := r.S.RulesByPromo[promotionID]
	if !ok {
		return domain.Rule{}, domain.ErrNotFound
	}
	rule, ok := r.S.Rules[rid]
	if !ok || rule.TenantID != tenantID {
		return domain.Rule{}, domain.ErrNotFound
	}
	return cloneRule(rule), nil
}

func (r *RuleRepo) ListByPromotionIDs(_ context.Context, tenantID uuid.UUID, promotionIDs []uuid.UUID) ([]domain.Rule, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.Rule
	for _, pid := range promotionIDs {
		if rid, ok := r.S.RulesByPromo[pid]; ok {
			if rule, ok := r.S.Rules[rid]; ok && rule.TenantID == tenantID {
				out = append(out, cloneRule(rule))
			}
		}
	}
	return out, nil
}

func cloneRule(r domain.Rule) domain.Rule {
	out := r
	out.ExcludePromotionIDs = append([]uuid.UUID(nil), r.ExcludePromotionIDs...)
	out.VariantIDs = append([]string(nil), r.VariantIDs...)
	out.CategoryIDs = append([]string(nil), r.CategoryIDs...)
	out.BrandIDs = append([]string(nil), r.BrandIDs...)
	out.SegmentIDs = append([]string(nil), r.SegmentIDs...)
	return out
}

var _ ports.RuleRepository = (*RuleRepo)(nil)

// --- Coupon ---

type CouponRepo struct{ S *Store }

func (r *CouponRepo) Create(_ context.Context, c domain.Coupon) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Coupons[c.ID]; ok {
		return domain.ErrAlreadyExists
	}
	code := strings.ToUpper(strings.TrimSpace(c.Code))
	ck := tenantCodeKey(c.TenantID, code)
	if _, ok := r.S.CouponsByCode[ck]; ok {
		return domain.ErrAlreadyExists
	}
	c.Code = code
	r.S.Coupons[c.ID] = cloneCoupon(c)
	r.S.CouponsByCode[ck] = c.ID
	return nil
}

func (r *CouponRepo) Update(_ context.Context, c domain.Coupon) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	cur, ok := r.S.Coupons[c.ID]
	if !ok || cur.TenantID != c.TenantID {
		return domain.ErrNotFound
	}
	r.S.Coupons[c.ID] = cloneCoupon(c)
	return nil
}

func (r *CouponRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.Coupon, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	c, ok := r.S.Coupons[id]
	if !ok || c.TenantID != tenantID {
		return domain.Coupon{}, domain.ErrNotFound
	}
	return cloneCoupon(c), nil
}

func (r *CouponRepo) GetByCode(_ context.Context, tenantID uuid.UUID, code string) (domain.Coupon, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	ck := tenantCodeKey(tenantID, strings.ToUpper(strings.TrimSpace(code)))
	id, ok := r.S.CouponsByCode[ck]
	if !ok {
		return domain.Coupon{}, domain.ErrNotFound
	}
	c, ok := r.S.Coupons[id]
	if !ok || c.TenantID != tenantID {
		return domain.Coupon{}, domain.ErrNotFound
	}
	return cloneCoupon(c), nil
}

func (r *CouponRepo) List(_ context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Coupon, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := []domain.Coupon{}
	for _, c := range r.S.Coupons {
		if c.TenantID == tenantID {
			out = append(out, cloneCoupon(c))
		}
	}
	if offset > len(out) {
		return []domain.Coupon{}, nil
	}
	out = out[offset:]
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (r *CouponRepo) CreateRedemption(_ context.Context, red domain.CouponRedemption) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	ik := idemKey(red.TenantID, red.IdempotencyKey)
	if _, ok := r.S.CouponIdem[ik]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.CouponRedemptions[red.ID] = red
	r.S.CouponIdem[ik] = red.ID
	return nil
}

func (r *CouponRepo) GetRedemptionByIdempotency(_ context.Context, tenantID uuid.UUID, key string) (domain.CouponRedemption, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.CouponIdem[idemKey(tenantID, key)]
	if !ok {
		return domain.CouponRedemption{}, domain.ErrNotFound
	}
	return r.S.CouponRedemptions[id], nil
}

func cloneCoupon(c domain.Coupon) domain.Coupon {
	out := c
	if c.PrincipalID != nil {
		p := *c.PrincipalID
		out.PrincipalID = &p
	}
	if c.StartsAt != nil {
		t := *c.StartsAt
		out.StartsAt = &t
	}
	if c.EndsAt != nil {
		t := *c.EndsAt
		out.EndsAt = &t
	}
	return out
}

var _ ports.CouponRepository = (*CouponRepo)(nil)

// --- Voucher ---

type VoucherRepo struct{ S *Store }

func (r *VoucherRepo) Create(_ context.Context, v domain.Voucher) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Vouchers[v.ID]; ok {
		return domain.ErrAlreadyExists
	}
	code := strings.ToUpper(strings.TrimSpace(v.Code))
	ck := tenantCodeKey(v.TenantID, code)
	if _, ok := r.S.VouchersByCode[ck]; ok {
		return domain.ErrAlreadyExists
	}
	v.Code = code
	r.S.Vouchers[v.ID] = cloneVoucher(v)
	r.S.VouchersByCode[ck] = v.ID
	return nil
}

func (r *VoucherRepo) Update(_ context.Context, v domain.Voucher) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	cur, ok := r.S.Vouchers[v.ID]
	if !ok || cur.TenantID != v.TenantID {
		return domain.ErrNotFound
	}
	r.S.Vouchers[v.ID] = cloneVoucher(v)
	return nil
}

func (r *VoucherRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.Voucher, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	v, ok := r.S.Vouchers[id]
	if !ok || v.TenantID != tenantID {
		return domain.Voucher{}, domain.ErrNotFound
	}
	return cloneVoucher(v), nil
}

func (r *VoucherRepo) GetByCode(_ context.Context, tenantID uuid.UUID, code string) (domain.Voucher, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	ck := tenantCodeKey(tenantID, strings.ToUpper(strings.TrimSpace(code)))
	id, ok := r.S.VouchersByCode[ck]
	if !ok {
		return domain.Voucher{}, domain.ErrNotFound
	}
	return cloneVoucher(r.S.Vouchers[id]), nil
}

func (r *VoucherRepo) CreateRedemption(_ context.Context, red domain.VoucherRedemption) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	ik := idemKey(red.TenantID, red.IdempotencyKey)
	if _, ok := r.S.VoucherIdem[ik]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.VoucherRedemptions[red.ID] = red
	r.S.VoucherIdem[ik] = red.ID
	return nil
}

func (r *VoucherRepo) GetRedemptionByIdempotency(_ context.Context, tenantID uuid.UUID, key string) (domain.VoucherRedemption, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.VoucherIdem[idemKey(tenantID, key)]
	if !ok {
		return domain.VoucherRedemption{}, domain.ErrNotFound
	}
	return r.S.VoucherRedemptions[id], nil
}

func cloneVoucher(v domain.Voucher) domain.Voucher {
	out := v
	if v.PromotionID != nil {
		p := *v.PromotionID
		out.PromotionID = &p
	}
	if v.StartsAt != nil {
		t := *v.StartsAt
		out.StartsAt = &t
	}
	if v.EndsAt != nil {
		t := *v.EndsAt
		out.EndsAt = &t
	}
	return out
}

var _ ports.VoucherRepository = (*VoucherRepo)(nil)

// --- Usage ---

type UsageRepo struct{ S *Store }

func (r *UsageRepo) Get(_ context.Context, tenantID, promotionID uuid.UUID, scope domain.UsageScope, scopeKey string) (domain.UsageCounter, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	u, ok := r.S.Usage[usageKey(tenantID, promotionID, scope, scopeKey)]
	if !ok {
		return domain.UsageCounter{}, domain.ErrNotFound
	}
	return u, nil
}

func (r *UsageRepo) Increment(_ context.Context, tenantID, promotionID uuid.UUID, scope domain.UsageScope, scopeKey string) (domain.UsageCounter, error) {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	k := usageKey(tenantID, promotionID, scope, scopeKey)
	u, ok := r.S.Usage[k]
	if !ok {
		u = domain.UsageCounter{
			ID:          uuid.New(),
			TenantID:    tenantID,
			PromotionID: promotionID,
			Scope:       scope,
			ScopeKey:    scopeKey,
			Count:       0,
		}
	}
	u.Count++
	u.UpdatedAt = time.Now().UTC()
	r.S.Usage[k] = u
	return u, nil
}

var _ ports.UsageRepository = (*UsageRepo)(nil)

// --- Simulation ---

type SimulationRepo struct{ S *Store }

func (r *SimulationRepo) Create(_ context.Context, s domain.Simulation) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Simulations[s.ID] = s
	return nil
}

func (r *SimulationRepo) List(_ context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Simulation, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.Simulation
	for _, s := range r.S.Simulations {
		if s.TenantID == tenantID {
			out = append(out, s)
		}
	}
	if offset > len(out) {
		return nil, nil
	}
	out = out[offset:]
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

var _ ports.SimulationRepository = (*SimulationRepo)(nil)

// --- Outbox ---

type OutboxRepo struct{ S *Store }

func (r *OutboxRepo) Enqueue(_ context.Context, msg domain.OutboxMessage) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Outbox[msg.ID] = msg
	return nil
}

func (r *OutboxRepo) ListPending(_ context.Context, limit int) ([]domain.OutboxMessage, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.OutboxMessage
	for _, m := range r.S.Outbox {
		if m.Status == domain.OutboxStatusPending {
			out = append(out, m)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (r *OutboxRepo) MarkPublished(_ context.Context, id uuid.UUID, at time.Time) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	m, ok := r.S.Outbox[id]
	if !ok {
		return domain.ErrNotFound
	}
	m.Status = domain.OutboxStatusPublished
	m.PublishedAt = &at
	m.UpdatedAt = at
	r.S.Outbox[id] = m
	return nil
}

var _ ports.OutboxRepository = (*OutboxRepo)(nil)

// NoopPublisher is a silent event publisher for tests.
type NoopPublisher struct{}

func (NoopPublisher) Publish(context.Context, string, string, any) error { return nil }

var _ ports.EventPublisher = (NoopPublisher{})

package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/pricing-service/internal/domain"
)

// UpsertTaxRuleInput creates or updates a tax display rule.
type UpsertTaxRuleInput struct {
	TenantID  uuid.UUID
	RuleID    *uuid.UUID
	Code      string
	Name      string
	RateBps   int
	Inclusive bool
	RegionID  *uuid.UUID
	Active    *bool
	Priority  int
}

// UpsertTaxRule upserts a tax rule.
func (d *Deps) UpsertTaxRule(ctx context.Context, in UpsertTaxRuleInput) (domain.TaxRule, error) {
	if in.TenantID == uuid.Nil {
		return domain.TaxRule{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	now := d.now()
	var rule domain.TaxRule
	if in.RuleID != nil && *in.RuleID != uuid.Nil {
		existing, err := d.Taxes.Get(ctx, in.TenantID, *in.RuleID)
		if err != nil {
			return domain.TaxRule{}, err
		}
		rule = existing
		if in.Code != "" {
			rule.Code = in.Code
		}
		if in.Name != "" {
			rule.Name = in.Name
		}
		rule.RateBps = in.RateBps
		rule.Inclusive = in.Inclusive
		rule.RegionID = in.RegionID
		rule.Priority = in.Priority
		if in.Active != nil {
			rule.Active = *in.Active
		}
		rule.UpdatedAt = now
	} else {
		active := true
		if in.Active != nil {
			active = *in.Active
		}
		rule = domain.TaxRule{
			ID: d.newID(), TenantID: in.TenantID, Code: in.Code, Name: in.Name,
			RateBps: in.RateBps, Inclusive: in.Inclusive, RegionID: in.RegionID,
			Active: active, Priority: in.Priority, CreatedAt: now, UpdatedAt: now,
		}
	}
	if rule.Name == "" {
		rule.Name = rule.Code
	}
	if err := rule.Validate(); err != nil {
		return domain.TaxRule{}, err
	}
	if err := d.Taxes.Upsert(ctx, rule); err != nil {
		return domain.TaxRule{}, err
	}
	return rule, nil
}

// TaxCalculateInput computes tax on a taxable base.
type TaxCalculateInput struct {
	TenantID   uuid.UUID
	RegionID   *uuid.UUID
	BaseMinor  int64
	Currency   string
}

// TaxCalculateResult is the computed tax.
type TaxCalculateResult struct {
	TaxMinor    int64
	RateBps     int
	RuleCode    string
	Inclusive   bool
	TaxableBase int64
}

// TaxCalculate applies the best matching tax rule to a taxable base.
func (d *Deps) TaxCalculate(ctx context.Context, in TaxCalculateInput) (TaxCalculateResult, error) {
	if in.TenantID == uuid.Nil {
		return TaxCalculateResult{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	if in.BaseMinor < 0 {
		return TaxCalculateResult{}, fmt.Errorf("%w: base_minor", domain.ErrNegativeMoney)
	}
	if in.Currency != "" {
		if _, err := domain.NewMoney(0, strings.ToUpper(in.Currency)); err != nil {
			return TaxCalculateResult{}, err
		}
	}
	rules, err := d.Taxes.List(ctx, in.TenantID)
	if err != nil {
		return TaxCalculateResult{}, err
	}
	rule, ok := domain.SelectTaxRule(rules, in.RegionID)
	if !ok {
		return TaxCalculateResult{TaxMinor: 0, TaxableBase: in.BaseMinor}, nil
	}
	tax := rule.TaxOn(in.BaseMinor)
	return TaxCalculateResult{
		TaxMinor: tax, RateBps: rule.RateBps, RuleCode: rule.Code,
		Inclusive: rule.Inclusive, TaxableBase: in.BaseMinor,
	}, nil
}

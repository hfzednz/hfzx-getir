package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/finance-ledger-service/internal/domain"
)

// TaxCalculateInput computes tax for a base amount.
type TaxCalculateInput struct {
	TenantID  uuid.UUID
	BaseMinor int64
	Currency  string
	TaxCode   string
	RateBps   *int64 // optional override; when nil, loads TaxRule by code
}

// TaxCalculate resolves a tax rule (or override) and returns TaxResult.
func (d *Deps) TaxCalculate(ctx context.Context, in TaxCalculateInput) (domain.TaxResult, error) {
	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	code := strings.TrimSpace(in.TaxCode)
	var rateBps int64
	if in.RateBps != nil {
		rateBps = *in.RateBps
	} else {
		if in.TenantID == uuid.Nil {
			return domain.TaxResult{}, fmt.Errorf("%w: tenant_id", domain.ErrInvalidArgument)
		}
		if code == "" {
			return domain.TaxResult{}, fmt.Errorf("%w: tax_code", domain.ErrInvalidArgument)
		}
		rule, err := d.TaxRules.GetByCode(ctx, in.TenantID, code)
		if err != nil {
			return domain.TaxResult{}, err
		}
		if !rule.Active {
			return domain.TaxResult{}, fmt.Errorf("%w: tax rule inactive", domain.ErrInvalidArgument)
		}
		rateBps = rule.RateBps
	}
	return domain.CalculateTax(in.BaseMinor, rateBps, currency, code)
}

// UpsertTaxRuleInput creates/updates a tax rule.
type UpsertTaxRuleInput struct {
	TenantID uuid.UUID
	Code     string
	Name     string
	RateBps  int64
	Currency string
	Active   bool
}

// UpsertTaxRule stores a tax rule by tenant+code.
func (d *Deps) UpsertTaxRule(ctx context.Context, in UpsertTaxRuleInput) (domain.TaxRule, error) {
	if in.TenantID == uuid.Nil {
		return domain.TaxRule{}, fmt.Errorf("%w: tenant_id", domain.ErrInvalidArgument)
	}
	code := strings.TrimSpace(in.Code)
	now := d.now()
	rule := domain.TaxRule{
		ID:        d.newID(),
		TenantID:  in.TenantID,
		Code:      code,
		Name:      strings.TrimSpace(in.Name),
		RateBps:   in.RateBps,
		Currency:  strings.ToUpper(strings.TrimSpace(in.Currency)),
		Active:    in.Active,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if rule.Name == "" {
		rule.Name = code
	}
	if existing, err := d.TaxRules.GetByCode(ctx, in.TenantID, code); err == nil {
		rule.ID = existing.ID
		rule.CreatedAt = existing.CreatedAt
	}
	if err := rule.Validate(); err != nil {
		return domain.TaxRule{}, err
	}
	if err := d.TaxRules.Upsert(ctx, rule); err != nil {
		return domain.TaxRule{}, err
	}
	return rule, nil
}

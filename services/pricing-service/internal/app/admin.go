package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/pricing-service/internal/domain"
)

// AdminListResult aggregates admin list views.
type AdminListResult struct {
	Books    []domain.PriceBook
	Entries  []domain.PriceEntry
	TaxRules []domain.TaxRule
	Dynamics []domain.DynamicRule
	Audits   []domain.QuoteAudit
}

// AdminListInput filters admin listings.
type AdminListInput struct {
	TenantID  uuid.UUID
	BookID    *uuid.UUID
	VariantID *uuid.UUID
	AuditLimit int
}

// AdminList returns books, entries, tax rules, dynamic rules, and recent audits.
func (d *Deps) AdminList(ctx context.Context, in AdminListInput) (AdminListResult, error) {
	if in.TenantID == uuid.Nil {
		return AdminListResult{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	out := AdminListResult{}
	var err error
	if d.Prices != nil {
		out.Books, err = d.Prices.ListBooks(ctx, in.TenantID)
		if err != nil {
			return AdminListResult{}, err
		}
		out.Entries, err = d.Prices.ListEntries(ctx, in.TenantID, in.BookID, in.VariantID)
		if err != nil {
			return AdminListResult{}, err
		}
	}
	if d.Taxes != nil {
		out.TaxRules, err = d.Taxes.List(ctx, in.TenantID)
		if err != nil {
			return AdminListResult{}, err
		}
	}
	if d.Dynamics != nil {
		out.Dynamics, err = d.Dynamics.List(ctx, in.TenantID)
		if err != nil {
			return AdminListResult{}, err
		}
	}
	if d.Audits != nil {
		limit := in.AuditLimit
		if limit <= 0 {
			limit = 50
		}
		out.Audits, err = d.Audits.List(ctx, in.TenantID, limit)
		if err != nil {
			return AdminListResult{}, err
		}
	}
	return out, nil
}

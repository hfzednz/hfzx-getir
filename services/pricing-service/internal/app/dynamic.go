package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/pricing-service/internal/app/ports"
	"github.com/nexora/pricing-service/internal/domain"
)

// UpsertDynamicRuleInput creates or updates a dynamic pricing rule.
type UpsertDynamicRuleInput struct {
	TenantID           uuid.UUID
	RuleID             *uuid.UUID
	Code               string
	Kind               domain.DynamicKind
	Trigger            domain.DynamicTrigger
	AdjustmentBps      int
	AdjustmentMinor    int64
	StartHour          int
	EndHour            int
	InventoryThreshold int
	Active             *bool
	Priority           int
}

// UpsertDynamicRule upserts a dynamic rule.
func (d *Deps) UpsertDynamicRule(ctx context.Context, in UpsertDynamicRuleInput) (domain.DynamicRule, error) {
	if in.TenantID == uuid.Nil {
		return domain.DynamicRule{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	now := d.now()
	var rule domain.DynamicRule
	if in.RuleID != nil && *in.RuleID != uuid.Nil {
		existing, err := d.Dynamics.Get(ctx, in.TenantID, *in.RuleID)
		if err != nil {
			return domain.DynamicRule{}, err
		}
		rule = existing
		if in.Code != "" {
			rule.Code = in.Code
		}
		if in.Kind != "" {
			rule.Kind = in.Kind
		}
		if in.Trigger != "" {
			rule.Trigger = in.Trigger
		}
		rule.AdjustmentBps = in.AdjustmentBps
		rule.AdjustmentMinor = in.AdjustmentMinor
		rule.StartHour = in.StartHour
		rule.EndHour = in.EndHour
		rule.InventoryThreshold = in.InventoryThreshold
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
		rule = domain.DynamicRule{
			ID: d.newID(), TenantID: in.TenantID, Code: in.Code,
			Kind: in.Kind, Trigger: in.Trigger,
			AdjustmentBps: in.AdjustmentBps, AdjustmentMinor: in.AdjustmentMinor,
			StartHour: in.StartHour, EndHour: in.EndHour,
			InventoryThreshold: in.InventoryThreshold,
			Active: active, Priority: in.Priority, CreatedAt: now, UpdatedAt: now,
		}
	}
	if err := rule.Validate(); err != nil {
		return domain.DynamicRule{}, err
	}
	if err := d.Dynamics.Upsert(ctx, rule); err != nil {
		return domain.DynamicRule{}, err
	}
	return rule, nil
}

// ApplyDynamicInput adjusts a resolved unit price with active dynamic rules.
type ApplyDynamicInput struct {
	TenantID    uuid.UUID
	VariantID   uuid.UUID
	UnitMinor   int64
	Currency    string
	WarehouseID *uuid.UUID
}

// ApplyDynamicResult is the adjusted unit price.
type ApplyDynamicResult struct {
	UnitMinor       int64
	BaseUnitMinor   int64
	AdjustmentMinor int64
	AppliedRules    []string
}

// ApplyDynamic applies percent/fixed time-of-day or inventory_hint rules.
func (d *Deps) ApplyDynamic(ctx context.Context, in ApplyDynamicInput) (ApplyDynamicResult, error) {
	if in.TenantID == uuid.Nil {
		return ApplyDynamicResult{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	if in.UnitMinor < 0 {
		return ApplyDynamicResult{}, fmt.Errorf("%w: unit_minor", domain.ErrNegativeMoney)
	}
	if in.Currency != "" {
		if _, err := domain.NewMoney(0, strings.ToUpper(in.Currency)); err != nil {
			return ApplyDynamicResult{}, err
		}
	}
	now := d.now()
	rules, err := d.Dynamics.ListActive(ctx, in.TenantID)
	if err != nil {
		return ApplyDynamicResult{}, err
	}

	var invQty *int
	if d.Hints != nil && in.VariantID != uuid.Nil {
		hint, herr := d.Hints.Hint(ctx, ports.DynamicHintRequest{
			TenantID: in.TenantID, VariantID: in.VariantID, WarehouseID: in.WarehouseID,
		})
		if herr == nil {
			invQty = hint.AvailableQty
		}
	}

	applicable := domain.SelectDynamicRules(rules, now, invQty)
	unit := in.UnitMinor
	codes := make([]string, 0, len(applicable))
	for _, r := range applicable {
		unit = r.ApplyToUnit(unit)
		codes = append(codes, r.Code)
	}
	return ApplyDynamicResult{
		UnitMinor: unit, BaseUnitMinor: in.UnitMinor,
		AdjustmentMinor: unit - in.UnitMinor, AppliedRules: codes,
	}, nil
}

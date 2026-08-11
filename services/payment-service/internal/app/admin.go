package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/payment-service/internal/app/ports"
	"github.com/nexora/payment-service/internal/domain"
)

// EligibilityInput checks whether payment methods are available (no charge).
type EligibilityInput struct {
	TenantID    uuid.UUID
	PrincipalID uuid.UUID
	AmountMinor int64
	Currency    string
	MethodTypes []domain.PaymentMethodType
}

// EligibilityResult is eligibility without capture.
type EligibilityResult struct {
	Eligible bool
	Reason   string
	Methods  []string
}

// CheckEligibility returns available payment methods for checkout (no charge).
func (d *Deps) CheckEligibility(ctx context.Context, in EligibilityInput) (EligibilityResult, error) {
	if in.TenantID == uuid.Nil {
		return EligibilityResult{}, fmt.Errorf("%w: tenant required", domain.ErrInvalidArgument)
	}
	if _, err := domain.NewMoney(in.AmountMinor, in.Currency); err != nil {
		return EligibilityResult{}, err
	}
	methods := []string{}
	wanted := in.MethodTypes
	if len(wanted) == 0 {
		wanted = []domain.PaymentMethodType{domain.MethodCard, domain.MethodWallet, domain.MethodApple, domain.MethodGoogle}
	}
	for _, m := range wanted {
		routes, err := d.Intents.ListRoutes(ctx, in.TenantID, m, in.Currency)
		if err != nil {
			continue
		}
		if len(routes) > 0 || m == domain.MethodWallet {
			methods = append(methods, string(m))
		}
	}
	if len(methods) == 0 {
		methods = []string{string(domain.MethodCard), string(domain.MethodWallet)}
	}
	return EligibilityResult{Eligible: true, Methods: methods}, nil
}

// RouteProviderInput resolves ordered PSP providers for a method.
type RouteProviderInput struct {
	TenantID   uuid.UUID
	MethodType domain.PaymentMethodType
	Currency   string
}

// RouteProvider returns ordered provider names (primary first).
func (d *Deps) RouteProvider(ctx context.Context, in RouteProviderInput) ([]string, error) {
	if in.TenantID == uuid.Nil {
		return nil, fmt.Errorf("%w: tenant required", domain.ErrInvalidArgument)
	}
	if in.MethodType == "" {
		in.MethodType = domain.MethodCard
	}
	routes, err := d.Intents.ListRoutes(ctx, in.TenantID, in.MethodType, in.Currency)
	if err != nil {
		return nil, err
	}
	if len(routes) == 0 {
		return []string{"mock_primary", "mock_failover"}, nil
	}
	return routes[0].Providers, nil
}

// UpsertRouteInput upserts a provider route.
type UpsertRouteInput struct {
	TenantID   uuid.UUID
	MethodType domain.PaymentMethodType
	Currency   string
	Providers  []string
	Priority   int
}

// UpsertRoute stores a provider preference route.
func (d *Deps) UpsertRoute(ctx context.Context, in UpsertRouteInput) (domain.ProviderRoute, error) {
	if len(in.Providers) == 0 {
		return domain.ProviderRoute{}, fmt.Errorf("%w: providers required", domain.ErrInvalidArgument)
	}
	now := d.now()
	r := domain.ProviderRoute{
		ID: d.newID(), TenantID: in.TenantID, MethodType: in.MethodType,
		Currency: in.Currency, Providers: in.Providers, Active: true,
		Priority: in.Priority, CreatedAt: now, UpdatedAt: now,
	}
	if err := d.Intents.UpsertRoute(ctx, r); err != nil {
		return domain.ProviderRoute{}, err
	}
	return r, nil
}

// RecordChargebackInput records a provider dispute.
type RecordChargebackInput struct {
	TenantID    uuid.UUID
	IntentID    uuid.UUID
	AmountMinor int64
	ReasonCode  string
	Reason      string
	ProviderRef string
}

// RecordChargeback opens a chargeback against a captured payment.
func (d *Deps) RecordChargeback(ctx context.Context, in RecordChargebackInput) (domain.Chargeback, error) {
	intent, err := d.Intents.GetIntent(ctx, in.TenantID, in.IntentID)
	if err != nil {
		return domain.Chargeback{}, err
	}
	amount := in.AmountMinor
	if amount == 0 {
		amount = intent.CapturedMinor
	}
	if amount <= 0 {
		return domain.Chargeback{}, fmt.Errorf("%w: chargeback amount required", domain.ErrInvalidArgument)
	}
	now := d.now()
	cb := domain.Chargeback{
		ID: d.newID(), IntentID: intent.ID, TenantID: intent.TenantID,
		AmountMinor: amount, Currency: intent.Currency,
		Status: domain.ChargebackOpened, Provider: intent.Provider,
		ProviderRef: in.ProviderRef, ReasonCode: in.ReasonCode, Reason: in.Reason,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := d.Intents.CreateChargeback(ctx, cb); err != nil {
		return domain.Chargeback{}, err
	}
	d.emit(ctx, intent, domain.EventChargebackCreated, map[string]any{
		"chargebackId": cb.ID.String(), "amountMinor": amount, "reasonCode": in.ReasonCode,
	})
	d.audit(ctx, intent.TenantID, &intent.ID, "chargeback", amount, intent.Currency, map[string]any{"chargebackId": cb.ID.String()})
	return cb, nil
}

// AdminListInput lists intents for admin explorer.
type AdminListInput struct {
	TenantID    uuid.UUID
	PrincipalID *uuid.UUID
	Status      *domain.IntentStatus
	OrderID     string
	Limit       int
	Offset      int
}

// AdminList returns a page of payment intents.
func (d *Deps) AdminList(ctx context.Context, in AdminListInput) ([]domain.PaymentIntent, int, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = 50
	}
	return d.Intents.ListIntents(ctx, ports.IntentFilter{
		TenantID: in.TenantID, PrincipalID: in.PrincipalID,
		Status: in.Status, OrderID: in.OrderID, Limit: limit, Offset: in.Offset,
	})
}

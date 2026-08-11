package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/payment-service/internal/domain"
)

// CreateIntentInput creates a new payment intent.
type CreateIntentInput struct {
	TenantID        uuid.UUID
	PrincipalID     uuid.UUID
	OrderID         string // opaque
	AmountMinor     int64
	Currency        string
	MethodType      domain.PaymentMethodType
	PaymentMethodID *uuid.UUID
	IdempotencyKey  string
	Metadata        map[string]any
}

// CreateIntent creates (or returns existing) payment intent by idempotency key.
func (d *Deps) CreateIntent(ctx context.Context, in CreateIntentInput) (domain.PaymentIntent, error) {
	if in.TenantID == uuid.Nil || in.PrincipalID == uuid.Nil {
		return domain.PaymentIntent{}, fmt.Errorf("%w: tenant and principal required", domain.ErrInvalidArgument)
	}
	if in.IdempotencyKey == "" {
		return domain.PaymentIntent{}, fmt.Errorf("%w: idempotency_key required", domain.ErrInvalidArgument)
	}
	if in.MethodType == "" {
		in.MethodType = domain.MethodCard
	}
	money, err := domain.NewMoney(in.AmountMinor, in.Currency)
	if err != nil {
		return domain.PaymentIntent{}, err
	}
	if money.AmountMinor == 0 {
		return domain.PaymentIntent{}, fmt.Errorf("%w: amount must be > 0", domain.ErrInvalidArgument)
	}

	if existing, err := d.Intents.GetByIdempotencyKey(ctx, in.TenantID, in.IdempotencyKey); err == nil {
		return existing, nil
	}

	now := d.now()
	intent := domain.PaymentIntent{
		ID:             d.newID(),
		TenantID:       in.TenantID,
		PrincipalID:    in.PrincipalID,
		OrderID:        strings.TrimSpace(in.OrderID),
		Status:         domain.IntentInitiated,
		AmountMinor:    money.AmountMinor,
		Currency:       money.Currency,
		MethodType:     in.MethodType,
		PaymentMethodID: in.PaymentMethodID,
		IdempotencyKey: in.IdempotencyKey,
		Metadata:       in.Metadata,
		Version:        1,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if intent.Metadata == nil {
		intent.Metadata = map[string]any{}
	}
	if err := intent.Validate(); err != nil {
		return domain.PaymentIntent{}, err
	}
	if err := d.Intents.CreateIntent(ctx, intent); err != nil {
		if existing, e2 := d.Intents.GetByIdempotencyKey(ctx, in.TenantID, in.IdempotencyKey); e2 == nil {
			return existing, nil
		}
		return domain.PaymentIntent{}, err
	}
	d.emit(ctx, intent, domain.EventPaymentInitiated, nil)
	d.audit(ctx, intent.TenantID, &intent.ID, "create_intent", intent.AmountMinor, intent.Currency, nil)
	return intent, nil
}

// GetIntent loads an intent by id.
func (d *Deps) GetIntent(ctx context.Context, tenantID, id uuid.UUID) (domain.PaymentIntent, error) {
	return d.Intents.GetIntent(ctx, tenantID, id)
}

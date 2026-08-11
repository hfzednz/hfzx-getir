package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/payment-service/internal/app/ports"
	"github.com/nexora/payment-service/internal/domain"
)

// VoidInput voids an authorized (uncaptured) intent.
type VoidInput struct {
	TenantID       uuid.UUID
	IntentID       uuid.UUID
	IdempotencyKey string
}

// Void releases an authorization without capture.
func (d *Deps) Void(ctx context.Context, in VoidInput) (domain.PaymentIntent, error) {
	if in.IdempotencyKey == "" {
		return domain.PaymentIntent{}, fmt.Errorf("%w: idempotency_key required", domain.ErrInvalidArgument)
	}
	intent, err := d.Intents.GetIntent(ctx, in.TenantID, in.IntentID)
	if err != nil {
		return domain.PaymentIntent{}, err
	}
	if intent.Status == domain.IntentVoided {
		return intent, nil
	}
	if intent.Status != domain.IntentAuthorized {
		return domain.PaymentIntent{}, fmt.Errorf("%w: cannot void from %s", domain.ErrInvalidTransition, intent.Status)
	}
	if intent.CapturedMinor > 0 {
		return domain.PaymentIntent{}, fmt.Errorf("%w: cannot void partially captured intent", domain.ErrInvalidTransition)
	}

	now := d.now()
	if intent.MethodType != domain.MethodWallet && intent.Provider != "wallet" {
		if d.PSP == nil {
			return domain.PaymentIntent{}, fmt.Errorf("%w: psp not configured", domain.ErrInvariant)
		}
		res, err := d.PSP.Void(ctx, ports.VoidRequest{
			IntentID: intent.ID, TenantID: intent.TenantID,
			ProviderRef: intent.ProviderIntentRef, IdempotencyKey: in.IdempotencyKey,
		})
		provider := d.PSP.Name()
		if err != nil || !res.Success {
			msg := res.ErrorMessage
			if msg == "" && err != nil {
				msg = err.Error()
			}
			_ = d.Intents.CreateAttempt(ctx, domain.PaymentAttempt{
				ID: d.newID(), IntentID: intent.ID, TenantID: intent.TenantID,
				Kind: domain.AttemptVoid, Status: domain.AttemptFailed,
				Provider: provider, ErrorCode: res.ErrorCode, ErrorMessage: msg,
				IdempotencyKey: in.IdempotencyKey, CreatedAt: now,
			})
			if err != nil {
				return domain.PaymentIntent{}, err
			}
			return domain.PaymentIntent{}, fmt.Errorf("%w: %s", domain.ErrPSPFailed, msg)
		}
		_ = d.Intents.CreateAttempt(ctx, domain.PaymentAttempt{
			ID: d.newID(), IntentID: intent.ID, TenantID: intent.TenantID,
			Kind: domain.AttemptVoid, Status: domain.AttemptSuccess,
			Provider: provider, ProviderRef: res.ProviderRef,
			IdempotencyKey: in.IdempotencyKey, CreatedAt: now,
		})
	} else {
		_ = d.Intents.CreateAttempt(ctx, domain.PaymentAttempt{
			ID: d.newID(), IntentID: intent.ID, TenantID: intent.TenantID,
			Kind: domain.AttemptVoid, Status: domain.AttemptSuccess,
			Provider: "wallet", IdempotencyKey: in.IdempotencyKey, CreatedAt: now,
		})
	}

	intent.Status = domain.IntentVoided
	intent.VoidedAt = &now
	intent.UpdatedAt = now
	intent.Version++
	if err := d.Intents.UpdateIntent(ctx, intent); err != nil {
		return domain.PaymentIntent{}, err
	}
	d.emit(ctx, intent, domain.EventPaymentVoided, nil)
	d.audit(ctx, intent.TenantID, &intent.ID, "void", intent.AmountMinor, intent.Currency, nil)
	return intent, nil
}

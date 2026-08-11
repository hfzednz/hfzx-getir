package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/payment-service/internal/app/ports"
	"github.com/nexora/payment-service/internal/domain"
)

// CaptureInput captures an authorized intent (full or partial).
type CaptureInput struct {
	TenantID       uuid.UUID
	IntentID       uuid.UUID
	AmountMinor    int64 // 0 = full remaining
	IdempotencyKey string
}

// Capture captures funds. PartialCapture is Capture with AmountMinor < remaining.
func (d *Deps) Capture(ctx context.Context, in CaptureInput) (domain.PaymentIntent, error) {
	if in.IdempotencyKey == "" {
		return domain.PaymentIntent{}, fmt.Errorf("%w: idempotency_key required", domain.ErrInvalidArgument)
	}
	intent, err := d.Intents.GetIntent(ctx, in.TenantID, in.IntentID)
	if err != nil {
		return domain.PaymentIntent{}, err
	}
	if intent.Status == domain.IntentCaptured && intent.RemainingCapturable() == 0 {
		return intent, nil
	}
	if intent.Status != domain.IntentAuthorized && !(intent.Status == domain.IntentCaptured && intent.RemainingCapturable() > 0) {
		return domain.PaymentIntent{}, fmt.Errorf("%w: cannot capture from %s", domain.ErrInvalidTransition, intent.Status)
	}

	amount := in.AmountMinor
	if amount == 0 {
		amount = intent.RemainingCapturable()
	}
	if amount <= 0 {
		return domain.PaymentIntent{}, fmt.Errorf("%w: capture amount must be > 0", domain.ErrInvalidArgument)
	}
	if amount > intent.RemainingCapturable() {
		return domain.PaymentIntent{}, fmt.Errorf("%w: want %d remaining %d", domain.ErrInsufficientCapture, amount, intent.RemainingCapturable())
	}

	now := d.now()

	if intent.MethodType == domain.MethodWallet || intent.Provider == "wallet" {
		// Wallet authorize already moved funds; capture is bookkeeping.
		intent.CapturedMinor += amount
		if intent.CapturedMinor >= intent.AmountMinor {
			intent.Status = domain.IntentCaptured
			intent.CapturedAt = &now
		}
		intent.UpdatedAt = now
		intent.Version++
		_ = d.Intents.UpdateIntent(ctx, intent)
		_ = d.Intents.CreateAttempt(ctx, domain.PaymentAttempt{
			ID: d.newID(), IntentID: intent.ID, TenantID: intent.TenantID,
			Kind: domain.AttemptCapture, Status: domain.AttemptSuccess,
			Provider: "wallet", AmountMinor: amount, Currency: intent.Currency,
			IdempotencyKey: in.IdempotencyKey, CreatedAt: now,
		})
		d.emit(ctx, intent, domain.EventPaymentCaptured, map[string]any{"capturedMinor": amount})
		d.audit(ctx, intent.TenantID, &intent.ID, "capture", amount, intent.Currency, nil)
		return intent, nil
	}

	if d.PSP == nil {
		return domain.PaymentIntent{}, fmt.Errorf("%w: psp not configured", domain.ErrInvariant)
	}
	res, err := d.PSP.Capture(ctx, ports.CaptureRequest{
		IntentID: intent.ID, TenantID: intent.TenantID,
		ProviderRef: intent.ProviderIntentRef, AmountMinor: amount, Currency: intent.Currency,
		IdempotencyKey: in.IdempotencyKey,
	})
	provider := d.PSP.Name()
	if err != nil || !res.Success {
		msg := res.ErrorMessage
		if msg == "" && err != nil {
			msg = err.Error()
		}
		_ = d.Intents.CreateAttempt(ctx, domain.PaymentAttempt{
			ID: d.newID(), IntentID: intent.ID, TenantID: intent.TenantID,
			Kind: domain.AttemptCapture, Status: domain.AttemptFailed,
			Provider: provider, AmountMinor: amount, Currency: intent.Currency,
			ErrorCode: res.ErrorCode, ErrorMessage: msg,
			IdempotencyKey: in.IdempotencyKey, CreatedAt: now,
		})
		if err != nil {
			return domain.PaymentIntent{}, err
		}
		return domain.PaymentIntent{}, fmt.Errorf("%w: %s", domain.ErrPSPFailed, msg)
	}

	intent.CapturedMinor += amount
	if intent.CapturedMinor >= intent.AmountMinor {
		intent.Status = domain.IntentCaptured
		intent.CapturedAt = &now
	} else {
		intent.Status = domain.IntentCaptured // partial still marked captured with remaining capturable tracking
		if intent.CapturedAt == nil {
			intent.CapturedAt = &now
		}
	}
	intent.UpdatedAt = now
	intent.Version++
	if err := d.Intents.UpdateIntent(ctx, intent); err != nil {
		return domain.PaymentIntent{}, err
	}
	_ = d.Intents.CreateAttempt(ctx, domain.PaymentAttempt{
		ID: d.newID(), IntentID: intent.ID, TenantID: intent.TenantID,
		Kind: domain.AttemptCapture, Status: domain.AttemptSuccess,
		Provider: provider, ProviderRef: res.ProviderRef,
		AmountMinor: amount, Currency: intent.Currency,
		IdempotencyKey: in.IdempotencyKey, CreatedAt: now,
	})
	d.emit(ctx, intent, domain.EventPaymentCaptured, map[string]any{"capturedMinor": amount})
	d.audit(ctx, intent.TenantID, &intent.ID, "capture", amount, intent.Currency, map[string]any{"provider": provider})
	d.postLedger(ctx, intent, "capture")
	return intent, nil
}

// PartialCapture is an alias that requires AmountMinor > 0 and < remaining.
func (d *Deps) PartialCapture(ctx context.Context, in CaptureInput) (domain.PaymentIntent, error) {
	if in.AmountMinor <= 0 {
		return domain.PaymentIntent{}, fmt.Errorf("%w: partial capture requires amount", domain.ErrInvalidArgument)
	}
	return d.Capture(ctx, in)
}

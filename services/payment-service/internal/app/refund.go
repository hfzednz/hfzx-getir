package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/payment-service/internal/app/ports"
	"github.com/nexora/payment-service/internal/domain"
)

// RefundInput refunds a captured intent (partial or full).
type RefundInput struct {
	TenantID       uuid.UUID
	IntentID       uuid.UUID
	AmountMinor    int64 // 0 = full remaining refundable
	Reason         string
	IdempotencyKey string
}

// Refund creates a refund against captured funds.
func (d *Deps) Refund(ctx context.Context, in RefundInput) (domain.Refund, domain.PaymentIntent, error) {
	if in.IdempotencyKey == "" {
		return domain.Refund{}, domain.PaymentIntent{}, fmt.Errorf("%w: idempotency_key required", domain.ErrInvalidArgument)
	}
	if existing, err := d.Intents.GetRefundByIdempotency(ctx, in.TenantID, in.IdempotencyKey); err == nil {
		intent, _ := d.Intents.GetIntent(ctx, in.TenantID, existing.IntentID)
		return existing, intent, nil
	}

	intent, err := d.Intents.GetIntent(ctx, in.TenantID, in.IntentID)
	if err != nil {
		return domain.Refund{}, domain.PaymentIntent{}, err
	}
	if intent.Status != domain.IntentCaptured && intent.Status != domain.IntentRefunded {
		return domain.Refund{}, domain.PaymentIntent{}, fmt.Errorf("%w: cannot refund from %s", domain.ErrInvalidTransition, intent.Status)
	}

	amount := in.AmountMinor
	if amount == 0 {
		amount = intent.RemainingRefundable()
	}
	if amount <= 0 {
		return domain.Refund{}, domain.PaymentIntent{}, fmt.Errorf("%w: refund amount must be > 0", domain.ErrInvalidArgument)
	}
	if amount > intent.RemainingRefundable() {
		return domain.Refund{}, domain.PaymentIntent{}, fmt.Errorf("%w: want %d remaining %d", domain.ErrRefundExceeds, amount, intent.RemainingRefundable())
	}

	now := d.now()
	refund := domain.Refund{
		ID: d.newID(), IntentID: intent.ID, TenantID: intent.TenantID,
		AmountMinor: amount, Currency: intent.Currency,
		Status: domain.RefundRequested, Provider: intent.Provider,
		Reason: in.Reason, IdempotencyKey: in.IdempotencyKey, CreatedAt: now,
	}
	d.emit(ctx, intent, domain.EventRefundRequested, map[string]any{"refundId": refund.ID.String(), "amountMinor": amount})

	if intent.Provider != "wallet" && intent.MethodType != domain.MethodWallet {
		if d.PSP == nil {
			return domain.Refund{}, domain.PaymentIntent{}, fmt.Errorf("%w: psp not configured", domain.ErrInvariant)
		}
		res, err := d.PSP.Refund(ctx, ports.RefundRequest{
			IntentID: intent.ID, TenantID: intent.TenantID,
			ProviderRef: intent.ProviderIntentRef, AmountMinor: amount, Currency: intent.Currency,
			IdempotencyKey: in.IdempotencyKey, Reason: in.Reason,
		})
		if err != nil || !res.Success {
			msg := res.ErrorMessage
			if msg == "" && err != nil {
				msg = err.Error()
			}
			refund.Status = domain.RefundFailed
			_ = d.Intents.CreateRefund(ctx, refund)
			_ = d.Intents.CreateAttempt(ctx, domain.PaymentAttempt{
				ID: d.newID(), IntentID: intent.ID, TenantID: intent.TenantID,
				Kind: domain.AttemptRefund, Status: domain.AttemptFailed,
				Provider: d.PSP.Name(), AmountMinor: amount, Currency: intent.Currency,
				ErrorCode: res.ErrorCode, ErrorMessage: msg,
				IdempotencyKey: in.IdempotencyKey, CreatedAt: now,
			})
			if err != nil {
				return refund, intent, err
			}
			return refund, intent, fmt.Errorf("%w: %s", domain.ErrPSPFailed, msg)
		}
		refund.ProviderRef = res.ProviderRef
		_ = d.Intents.CreateAttempt(ctx, domain.PaymentAttempt{
			ID: d.newID(), IntentID: intent.ID, TenantID: intent.TenantID,
			Kind: domain.AttemptRefund, Status: domain.AttemptSuccess,
			Provider: d.PSP.Name(), ProviderRef: res.ProviderRef,
			AmountMinor: amount, Currency: intent.Currency,
			IdempotencyKey: in.IdempotencyKey, CreatedAt: now,
		})
	} else {
		_ = d.Intents.CreateAttempt(ctx, domain.PaymentAttempt{
			ID: d.newID(), IntentID: intent.ID, TenantID: intent.TenantID,
			Kind: domain.AttemptRefund, Status: domain.AttemptSuccess,
			Provider: "wallet", AmountMinor: amount, Currency: intent.Currency,
			IdempotencyKey: in.IdempotencyKey, CreatedAt: now,
		})
	}

	refund.Status = domain.RefundCompleted
	refund.CompletedAt = &now
	if err := d.Intents.CreateRefund(ctx, refund); err != nil {
		return domain.Refund{}, domain.PaymentIntent{}, err
	}

	intent.RefundedMinor += amount
	if intent.RemainingRefundable() == 0 {
		intent.Status = domain.IntentRefunded
	}
	intent.UpdatedAt = now
	intent.Version++
	if err := d.Intents.UpdateIntent(ctx, intent); err != nil {
		return domain.Refund{}, domain.PaymentIntent{}, err
	}
	d.emit(ctx, intent, domain.EventRefundCompleted, map[string]any{"refundId": refund.ID.String(), "amountMinor": amount})
	d.audit(ctx, intent.TenantID, &intent.ID, "refund", amount, intent.Currency, map[string]any{"refundId": refund.ID.String()})
	if err := d.postLedger(ctx, intent, "refund", amount); err != nil {
		d.audit(ctx, intent.TenantID, &intent.ID, "ledger_failed", amount, intent.Currency, map[string]any{"action": "refund", "reason": err.Error()})
		return refund, intent, err
	}
	return refund, intent, nil
}

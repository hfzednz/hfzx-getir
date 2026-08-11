package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/payment-service/internal/app/ports"
	"github.com/nexora/payment-service/internal/domain"
)

// AuthorizeInput authorizes an initiated payment intent.
type AuthorizeInput struct {
	TenantID       uuid.UUID
	IntentID       uuid.UUID
	IdempotencyKey string
	Token          string // optional override; else from payment method
}

// Authorize runs fraud check then PSP authorize (or wallet debit).
// Idempotent: re-authorize with same key on already-authorized intent returns it.
func (d *Deps) Authorize(ctx context.Context, in AuthorizeInput) (domain.PaymentIntent, error) {
	if in.IdempotencyKey == "" {
		return domain.PaymentIntent{}, fmt.Errorf("%w: idempotency_key required", domain.ErrInvalidArgument)
	}
	intent, err := d.Intents.GetIntent(ctx, in.TenantID, in.IntentID)
	if err != nil {
		return domain.PaymentIntent{}, err
	}

	// Idempotent: already authorized (or beyond) only when matching authorize key.
	if intent.Status == domain.IntentAuthorized || intent.Status == domain.IntentCaptured {
		attempts, _ := d.Intents.ListAttempts(ctx, in.TenantID, intent.ID)
		for _, a := range attempts {
			if a.Kind == domain.AttemptAuthorize && a.IdempotencyKey == in.IdempotencyKey && a.Status == domain.AttemptSuccess {
				return intent, nil
			}
		}
		return domain.PaymentIntent{}, fmt.Errorf("%w: authorize idempotency key mismatch", domain.ErrConflict)
	}

	if intent.Status != domain.IntentInitiated {
		return domain.PaymentIntent{}, fmt.Errorf("%w: cannot authorize from %s", domain.ErrInvalidTransition, intent.Status)
	}

	// Fraud scoring
	if d.Fraud != nil {
		fr, err := d.Fraud.Score(ctx, ports.FraudRequest{
			TenantID: intent.TenantID, PrincipalID: intent.PrincipalID, IntentID: intent.ID,
			AmountMinor: intent.AmountMinor, Currency: intent.Currency,
			MethodType: intent.MethodType, OrderID: intent.OrderID,
		})
		if err != nil {
			return domain.PaymentIntent{}, err
		}
		intent.FraudScore = fr.Score
		intent.FraudDecision = string(fr.Decision)
		_ = d.Intents.CreateFraudScore(ctx, domain.FraudScore{
			ID: d.newID(), IntentID: intent.ID, TenantID: intent.TenantID,
			Score: fr.Score, Decision: fr.Decision, Reasons: fr.Reasons, CreatedAt: d.now(),
		})
		if fr.Decision == domain.FraudBlock || fr.Score >= d.fraudThreshold() && fr.Decision != domain.FraudAllow {
			now := d.now()
			intent.Status = domain.IntentFailed
			intent.FailureReason = "fraud_blocked"
			intent.FailedAt = &now
			intent.UpdatedAt = now
			intent.Version++
			_ = d.Intents.UpdateIntent(ctx, intent)
			_ = d.Intents.CreateAttempt(ctx, domain.PaymentAttempt{
				ID: d.newID(), IntentID: intent.ID, TenantID: intent.TenantID,
				Kind: domain.AttemptAuthorize, Status: domain.AttemptFailed,
				AmountMinor: intent.AmountMinor, Currency: intent.Currency,
				ErrorCode: "fraud_blocked", ErrorMessage: "fraud decision block",
				IdempotencyKey: in.IdempotencyKey, CreatedAt: now,
			})
			d.emit(ctx, intent, domain.EventPaymentFailed, map[string]any{"reason": "fraud_blocked"})
			d.audit(ctx, intent.TenantID, &intent.ID, "authorize_fraud_block", intent.AmountMinor, intent.Currency, nil)
			return domain.PaymentIntent{}, fmt.Errorf("%w: score=%d decision=%s", domain.ErrFraudBlocked, fr.Score, fr.Decision)
		}
	}

	now := d.now()

	// Wallet path
	if intent.MethodType == domain.MethodWallet {
		if d.Wallet == nil {
			return domain.PaymentIntent{}, fmt.Errorf("%w: wallet client not configured", domain.ErrInvariant)
		}
		wr, err := d.Wallet.Debit(ctx, ports.WalletDebitRequest{
			TenantID: intent.TenantID, PrincipalID: intent.PrincipalID,
			AmountMinor: intent.AmountMinor, Currency: intent.Currency,
			AccountType: "cash", IdempotencyKey: in.IdempotencyKey,
			Reference: intent.ID.String(),
		})
		if err != nil || !wr.Success {
			intent.Status = domain.IntentFailed
			intent.FailureReason = wr.Reason
			if intent.FailureReason == "" && err != nil {
				intent.FailureReason = err.Error()
			}
			intent.FailedAt = &now
			intent.UpdatedAt = now
			intent.Version++
			_ = d.Intents.UpdateIntent(ctx, intent)
			_ = d.Intents.CreateAttempt(ctx, domain.PaymentAttempt{
				ID: d.newID(), IntentID: intent.ID, TenantID: intent.TenantID,
				Kind: domain.AttemptAuthorize, Status: domain.AttemptFailed,
				Provider: "wallet", AmountMinor: intent.AmountMinor, Currency: intent.Currency,
				ErrorMessage: intent.FailureReason, IdempotencyKey: in.IdempotencyKey, CreatedAt: now,
			})
			d.emit(ctx, intent, domain.EventPaymentFailed, map[string]any{"reason": intent.FailureReason})
			if err != nil {
				return domain.PaymentIntent{}, err
			}
			return domain.PaymentIntent{}, fmt.Errorf("%w: %s", domain.ErrOverdraft, wr.Reason)
		}
		intent.Status = domain.IntentAuthorized
		intent.Provider = "wallet"
		intent.ProviderIntentRef = wr.EntryID
		intent.AuthorizedAt = &now
		intent.UpdatedAt = now
		intent.Version++
		_ = d.Intents.UpdateIntent(ctx, intent)
		_ = d.Intents.CreateAttempt(ctx, domain.PaymentAttempt{
			ID: d.newID(), IntentID: intent.ID, TenantID: intent.TenantID,
			Kind: domain.AttemptAuthorize, Status: domain.AttemptSuccess,
			Provider: "wallet", ProviderRef: wr.EntryID,
			AmountMinor: intent.AmountMinor, Currency: intent.Currency,
			IdempotencyKey: in.IdempotencyKey, CreatedAt: now,
		})
		d.emit(ctx, intent, domain.EventPaymentAuthorized, nil)
		d.audit(ctx, intent.TenantID, &intent.ID, "authorize_wallet", intent.AmountMinor, intent.Currency, nil)
		d.postLedger(ctx, intent, "authorize")
		return intent, nil
	}

	// Card / PSP path
	token := in.Token
	if token == "" && intent.PaymentMethodID != nil {
		m, err := d.Intents.GetMethod(ctx, intent.TenantID, *intent.PaymentMethodID)
		if err == nil {
			token = m.Token
		}
	}
	if token == "" {
		token = "tok_test"
	}
	if d.PSP == nil {
		return domain.PaymentIntent{}, fmt.Errorf("%w: psp not configured", domain.ErrInvariant)
	}

	res, err := d.PSP.Authorize(ctx, ports.AuthorizeRequest{
		IntentID: intent.ID, TenantID: intent.TenantID,
		AmountMinor: intent.AmountMinor, Currency: intent.Currency,
		Token: token, IdempotencyKey: in.IdempotencyKey,
	})
	provider := d.PSP.Name()
	if err != nil || !res.Success {
		msg := res.ErrorMessage
		if msg == "" && err != nil {
			msg = err.Error()
		}
		intent.Status = domain.IntentFailed
		intent.FailureReason = msg
		intent.FailedAt = &now
		intent.UpdatedAt = now
		intent.Version++
		_ = d.Intents.UpdateIntent(ctx, intent)
		_ = d.Intents.CreateAttempt(ctx, domain.PaymentAttempt{
			ID: d.newID(), IntentID: intent.ID, TenantID: intent.TenantID,
			Kind: domain.AttemptAuthorize, Status: domain.AttemptFailed,
			Provider: provider, AmountMinor: intent.AmountMinor, Currency: intent.Currency,
			ErrorCode: res.ErrorCode, ErrorMessage: msg,
			IdempotencyKey: in.IdempotencyKey, CreatedAt: now,
		})
		d.emit(ctx, intent, domain.EventPaymentFailed, map[string]any{"reason": msg})
		if err != nil {
			return domain.PaymentIntent{}, err
		}
		return domain.PaymentIntent{}, fmt.Errorf("%w: %s", domain.ErrPSPFailed, msg)
	}

	intent.Status = domain.IntentAuthorized
	intent.Provider = provider
	intent.ProviderIntentRef = res.ProviderRef
	intent.AuthorizedAt = &now
	intent.UpdatedAt = now
	intent.Version++
	if err := d.Intents.UpdateIntent(ctx, intent); err != nil {
		return domain.PaymentIntent{}, err
	}
	_ = d.Intents.CreateAttempt(ctx, domain.PaymentAttempt{
		ID: d.newID(), IntentID: intent.ID, TenantID: intent.TenantID,
		Kind: domain.AttemptAuthorize, Status: domain.AttemptSuccess,
		Provider: provider, ProviderRef: res.ProviderRef,
		AmountMinor: intent.AmountMinor, Currency: intent.Currency,
		IdempotencyKey: in.IdempotencyKey, CreatedAt: now,
	})
	d.emit(ctx, intent, domain.EventPaymentAuthorized, nil)
	d.audit(ctx, intent.TenantID, &intent.ID, "authorize", intent.AmountMinor, intent.Currency, map[string]any{"provider": provider})
	d.postLedger(ctx, intent, "authorize")
	return intent, nil
}

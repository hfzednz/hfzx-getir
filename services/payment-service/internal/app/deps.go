package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/payment-service/internal/app/ports"
	"github.com/nexora/payment-service/internal/domain"
)

// Deps aggregates application ports for payment use cases.
type Deps struct {
	Intents   ports.IntentRepo
	Outbox    ports.OutboxRepository
	Publisher ports.EventPublisher
	PSP       ports.PSPClient // usually Failover router wrapping MockPSPs
	Fraud     ports.FraudClient
	Wallet    ports.WalletClient
	Ledger    ports.LedgerClient
	Clock     ports.Clock
	IDs       ports.IDGen

	// FraudBlockThreshold: scores >= this with decision block are rejected (default 80).
	FraudBlockThreshold int
}

// SystemClock is a real-time Clock.
type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }

// UUIDGen generates random UUIDs.
type UUIDGen struct{}

func (UUIDGen) New() uuid.UUID { return uuid.New() }

func (d *Deps) now() time.Time {
	if d.Clock != nil {
		return d.Clock.Now().UTC()
	}
	return time.Now().UTC()
}

func (d *Deps) newID() uuid.UUID {
	if d.IDs != nil {
		return d.IDs.New()
	}
	return uuid.New()
}

func (d *Deps) fraudThreshold() int {
	if d.FraudBlockThreshold > 0 {
		return d.FraudBlockThreshold
	}
	return 80
}

func (d *Deps) enqueueOutbox(ctx context.Context, tenantID, intentID uuid.UUID, eventType string, payload map[string]any) {
	if d.Outbox == nil {
		return
	}
	now := d.now()
	msg := domain.OutboxMessage{
		ID:        d.newID(),
		TenantID:  tenantID,
		IntentID:  intentID,
		Topic:     domain.TopicForEvent(eventType),
		Key:       intentID.String(),
		Payload:   payload,
		Status:    domain.OutboxStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_ = d.Outbox.Enqueue(ctx, msg)
}

func (d *Deps) emit(ctx context.Context, intent domain.PaymentIntent, eventType string, extra map[string]any) {
	now := d.now()
	payload := map[string]any{
		"type":        eventType,
		"intentId":    intent.ID.String(),
		"tenantId":    intent.TenantID.String(),
		"principalId": intent.PrincipalID.String(),
		"orderId":     intent.OrderID,
		"status":      intent.Status,
		"amountMinor": intent.AmountMinor,
		"currency":    intent.Currency,
		"occurredAt":  now,
	}
	for k, v := range extra {
		payload[k] = v
	}
	d.enqueueOutbox(ctx, intent.TenantID, intent.ID, eventType, payload)
}

func (d *Deps) audit(ctx context.Context, tenantID uuid.UUID, intentID *uuid.UUID, action string, amount int64, currency string, detail map[string]any) {
	if d.Intents == nil {
		return
	}
	_ = d.Intents.CreateAudit(ctx, domain.AuditEntry{
		ID: d.newID(), TenantID: tenantID, IntentID: intentID,
		Action: action, AmountMinor: amount, Currency: currency,
		Detail: detail, CreatedAt: d.now(),
	})
}

func (d *Deps) postLedger(ctx context.Context, intent domain.PaymentIntent, action string, amountMinor int64) error {
	if d.Ledger == nil {
		return nil
	}
	if amountMinor <= 0 {
		amountMinor = intent.AmountMinor
	}
	lines := ledgerLines(action, amountMinor, intent.Currency)
	if _, err := d.Ledger.PostJournal(ctx, ports.PostJournalRequest{
		TenantID:       intent.TenantID,
		IdempotencyKey: intent.IdempotencyKey + ":" + action,
		Reference:      intent.ID.String(),
		Lines:          lines,
	}); err != nil {
		slog.Default().Error("payment.ledger.post",
			"err", err,
			"intentId", intent.ID.String(),
			"action", action,
			"tenantId", intent.TenantID.String(),
		)
		return fmt.Errorf("%w: %s", domain.ErrLedgerFailed, err.Error())
	}
	return nil
}

func ledgerLines(action string, amountMinor int64, currency string) []ports.JournalLine {
	switch action {
	case "capture":
		return []ports.JournalLine{
			{AccountCode: "liability.customer", DebitMinor: amountMinor, Currency: currency},
			{AccountCode: "revenue.sales", CreditMinor: amountMinor, Currency: currency},
		}
	case "refund":
		return []ports.JournalLine{
			{AccountCode: "revenue.sales", DebitMinor: amountMinor, Currency: currency},
			{AccountCode: "clearing.psp", CreditMinor: amountMinor, Currency: currency},
		}
	default:
		return []ports.JournalLine{
			{AccountCode: "clearing.psp", DebitMinor: amountMinor, Currency: currency},
			{AccountCode: "liability.customer", CreditMinor: amountMinor, Currency: currency},
		}
	}
}

// PublishPending drains the outbox via EventPublisher.
func (d *Deps) PublishPending(ctx context.Context, limit int) (int, error) {
	if d.Outbox == nil || d.Publisher == nil {
		return 0, nil
	}
	if limit <= 0 {
		limit = 100
	}
	msgs, err := d.Outbox.ListPending(ctx, limit)
	if err != nil {
		return 0, err
	}
	n := 0
	now := d.now()
	for _, m := range msgs {
		if err := d.Publisher.Publish(ctx, m.Topic, m.Key, m.Payload); err != nil {
			m.Attempts++
			m.LastError = err.Error()
			m.Status = domain.OutboxStatusFailed
			m.UpdatedAt = now
			_ = d.Outbox.Update(ctx, m)
			continue
		}
		m.Attempts++
		m.Status = domain.OutboxStatusPublished
		m.PublishedAt = &now
		m.UpdatedAt = now
		_ = d.Outbox.Update(ctx, m)
		n++
	}
	return n, nil
}

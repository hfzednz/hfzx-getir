package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/wallet-service/internal/app/ports"
	"github.com/nexora/wallet-service/internal/domain"
)

// Deps aggregates application ports for wallet use cases.
type Deps struct {
	Wallets   ports.WalletRepo
	Outbox    ports.OutboxRepository
	Publisher ports.EventPublisher
	Ledger    ports.LedgerClient
	Clock     ports.Clock
	IDs       ports.IDGen
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

func (d *Deps) enqueueOutbox(ctx context.Context, tenantID, walletID uuid.UUID, eventType string, payload map[string]any) {
	if d.Outbox == nil {
		return
	}
	now := d.now()
	_ = d.Outbox.Enqueue(ctx, domain.OutboxMessage{
		ID: d.newID(), TenantID: tenantID, WalletID: walletID,
		Topic: domain.TopicForEvent(eventType), Key: walletID.String(),
		Payload: payload, Status: domain.OutboxStatusPending,
		CreatedAt: now, UpdatedAt: now,
	})
}

func (d *Deps) emit(ctx context.Context, w domain.Wallet, eventType string, extra map[string]any) {
	now := d.now()
	payload := map[string]any{
		"type": eventType, "walletId": w.ID.String(), "tenantId": w.TenantID.String(),
		"principalId": w.PrincipalID.String(), "occurredAt": now,
	}
	for k, v := range extra {
		payload[k] = v
	}
	d.enqueueOutbox(ctx, w.TenantID, w.ID, eventType, payload)
}

func (d *Deps) postLedger(ctx context.Context, tenantID uuid.UUID, idem, ref string, amount int64, currency string) {
	if d.Ledger == nil {
		return
	}
	_, _ = d.Ledger.PostJournal(ctx, ports.PostJournalRequest{
		TenantID: tenantID, IdempotencyKey: idem, Reference: ref,
		Lines: []ports.JournalLine{
			{AccountCode: "liability.wallet", CreditMinor: amount, Currency: currency},
			{AccountCode: "asset.clearing", DebitMinor: amount, Currency: currency},
		},
	})
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

// Package ports defines application-layer dependency interfaces (hexagonal ports).
package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/settlement-service/internal/domain"
)

// Clock abstracts time for deterministic tests.
type Clock interface {
	Now() time.Time
}

// IDGen abstracts UUID generation.
type IDGen interface {
	New() uuid.UUID
}

// EventPublisher publishes domain events (Kafka adapters).
type EventPublisher interface {
	Publish(ctx context.Context, topic string, key string, payload any) error
}

// BatchFilter lists settlement batches.
type BatchFilter struct {
	TenantID uuid.UUID
	Status   *domain.BatchStatus
	Limit    int
	Offset   int
}

// BatchRepository persists settlement batches and related rows.
type BatchRepository interface {
	Create(ctx context.Context, b domain.SettlementBatch) error
	Update(ctx context.Context, b domain.SettlementBatch) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.SettlementBatch, error)
	GetByIdempotencyKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.SettlementBatch, error)
	List(ctx context.Context, f BatchFilter) ([]domain.SettlementBatch, int, error)
	SavePayout(ctx context.Context, p domain.PayoutInstruction) error
	ListPayouts(ctx context.Context, tenantID, batchID uuid.UUID) ([]domain.PayoutInstruction, error)
	SaveReconciliation(ctx context.Context, r domain.Reconciliation) error
	SaveMismatch(ctx context.Context, m domain.Mismatch) error
	ListMismatches(ctx context.Context, tenantID, batchID uuid.UUID) ([]domain.Mismatch, error)
}

// EventStore persists append-only settlement timeline events.
type EventStore interface {
	Append(ctx context.Context, e domain.SettlementEvent) error
	ListByBatch(ctx context.Context, tenantID, batchID uuid.UUID) ([]domain.SettlementEvent, error)
}

// OutboxRepository persists transactional outbox rows.
type OutboxRepository interface {
	Enqueue(ctx context.Context, m domain.OutboxMessage) error
	Update(ctx context.Context, m domain.OutboxMessage) error
	ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
}

// LedgerPostRequest posts a balancing journal to finance-ledger-service.
type LedgerPostRequest struct {
	TenantID       uuid.UUID
	Currency       string
	Reference      string
	IdempotencyKey string
	DebitAccount   string
	CreditAccount  string
	AmountMinor    int64
}

// LedgerPostResult is the ledger response.
type LedgerPostResult struct {
	JournalID string
	Posted    bool
}

// LedgerClient posts settlement journals (opaque to cart/order).
type LedgerClient interface {
	PostSettlementJournal(ctx context.Context, req LedgerPostRequest) (LedgerPostResult, error)
}

// PayoutRequest is a bank/PSP payout instruction.
type PayoutRequest struct {
	InstructionID uuid.UUID
	TenantID      uuid.UUID
	PayeeType     domain.PayeeType
	PayeeRef      string
	AmountMinor   int64
	Currency      string
}

// PayoutResult is the provider response.
type PayoutResult struct {
	ProviderRef string
	Succeeded   bool
	Error       string
}

// PayoutClient executes payouts via bank/PSP port.
type PayoutClient interface {
	Execute(ctx context.Context, req PayoutRequest) (PayoutResult, error)
}

// Package ports defines application-layer dependency interfaces (hexagonal ports).
package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/finance-ledger-service/internal/domain"
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

// JournalFilter lists journals.
type JournalFilter struct {
	TenantID uuid.UUID
	Status   *domain.JournalStatus
	Limit    int
	Offset   int
}

// AccountRepository persists chart-of-accounts.
type AccountRepository interface {
	Create(ctx context.Context, a domain.Account) error
	Update(ctx context.Context, a domain.Account) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Account, error)
	GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.Account, error)
}

// JournalRepository persists journals and lines.
type JournalRepository interface {
	Create(ctx context.Context, j domain.Journal) error
	Update(ctx context.Context, j domain.Journal) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Journal, error)
	GetByIdempotencyKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.Journal, error)
	List(ctx context.Context, f JournalFilter) ([]domain.Journal, int, error)
	// BalanceMinor returns net balance for an account from posted journal lines.
	// Convention: debit − credit (assets/expenses positive when debited).
	BalanceMinor(ctx context.Context, tenantID, accountID uuid.UUID) (int64, error)
}

// InvoiceRepository persists invoices and credit notes.
type InvoiceRepository interface {
	Create(ctx context.Context, inv domain.Invoice) error
	Update(ctx context.Context, inv domain.Invoice) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Invoice, error)
	GetByIdempotencyKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.Invoice, error)
	CreateCreditNote(ctx context.Context, cn domain.CreditNote) error
	GetCreditNoteByIdempotencyKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.CreditNote, error)
}

// TaxRuleRepository persists tax rules.
type TaxRuleRepository interface {
	Upsert(ctx context.Context, r domain.TaxRule) error
	GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.TaxRule, error)
}

// EventStore persists append-only ledger timeline events.
type EventStore interface {
	Append(ctx context.Context, e domain.LedgerEvent) error
	ListByEntity(ctx context.Context, tenantID, entityID uuid.UUID) ([]domain.LedgerEvent, error)
}

// OutboxRepository persists transactional outbox rows.
type OutboxRepository interface {
	Enqueue(ctx context.Context, m domain.OutboxMessage) error
	Update(ctx context.Context, m domain.OutboxMessage) error
	ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
}

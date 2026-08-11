// Package ports defines application-layer dependency interfaces.
package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/wallet-service/internal/domain"
)

// Clock abstracts time for deterministic tests.
type Clock interface {
	Now() time.Time
}

// IDGen abstracts UUID generation.
type IDGen interface {
	New() uuid.UUID
}

// EventPublisher publishes domain events.
type EventPublisher interface {
	Publish(ctx context.Context, topic string, key string, payload any) error
}

// WalletView is wallet + accounts snapshot.
type WalletView struct {
	Wallet   domain.Wallet
	Accounts []domain.Account
}

// WalletRepo persists wallets, accounts, entries, holds, transfers.
type WalletRepo interface {
	CreateWallet(ctx context.Context, w domain.Wallet, accounts []domain.Account) error
	GetWallet(ctx context.Context, tenantID, walletID uuid.UUID) (domain.Wallet, error)
	GetWalletByPrincipal(ctx context.Context, tenantID, principalID uuid.UUID) (domain.Wallet, error)
	UpdateWallet(ctx context.Context, w domain.Wallet) error

	GetAccount(ctx context.Context, tenantID, accountID uuid.UUID) (domain.Account, error)
	GetAccountByType(ctx context.Context, tenantID, walletID uuid.UUID, t domain.AccountType) (domain.Account, error)
	ListAccounts(ctx context.Context, tenantID, walletID uuid.UUID) ([]domain.Account, error)
	UpdateAccount(ctx context.Context, a domain.Account) error

	CreateEntry(ctx context.Context, e domain.Entry) error
	GetEntryByIdempotency(ctx context.Context, tenantID uuid.UUID, key string) (domain.Entry, error)
	ListEntries(ctx context.Context, tenantID, walletID uuid.UUID, limit, offset int) ([]domain.Entry, int, error)

	CreateHold(ctx context.Context, h domain.Hold) error
	GetHold(ctx context.Context, tenantID, holdID uuid.UUID) (domain.Hold, error)
	GetHoldByIdempotency(ctx context.Context, tenantID uuid.UUID, key string) (domain.Hold, error)
	UpdateHold(ctx context.Context, h domain.Hold) error

	CreateTransfer(ctx context.Context, t domain.Transfer) error
	GetTransferByIdempotency(ctx context.Context, tenantID uuid.UUID, key string) (domain.Transfer, error)

	UpsertLimit(ctx context.Context, l domain.Limit) error
	GetLimit(ctx context.Context, tenantID, walletID uuid.UUID, limitType, windowKey string) (domain.Limit, error)

	CreateAudit(ctx context.Context, a domain.AuditEntry) error
}

// OutboxRepository persists transactional outbox rows.
type OutboxRepository interface {
	Enqueue(ctx context.Context, m domain.OutboxMessage) error
	Update(ctx context.Context, m domain.OutboxMessage) error
	ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
}

// JournalLine is a double-entry line stub.
type JournalLine struct {
	AccountCode string
	DebitMinor  int64
	CreditMinor int64
	Currency    string
}

// PostJournalRequest posts a balanced journal stub.
type PostJournalRequest struct {
	TenantID       uuid.UUID
	IdempotencyKey string
	Reference      string
	Lines          []JournalLine
}

// PostJournalResult is the ledger stub response.
type PostJournalResult struct {
	JournalID string
	Posted    bool
}

// LedgerClient posts journal entries to finance-ledger-service (stub).
type LedgerClient interface {
	PostJournal(ctx context.Context, req PostJournalRequest) (PostJournalResult, error)
}

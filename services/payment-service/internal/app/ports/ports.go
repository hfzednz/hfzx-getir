// Package ports defines application-layer dependency interfaces (hexagonal ports).
package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/payment-service/internal/domain"
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

// IntentFilter lists payment intents.
type IntentFilter struct {
	TenantID    uuid.UUID
	PrincipalID *uuid.UUID
	Status      *domain.IntentStatus
	OrderID     string
	Query       string
	Limit       int
	Offset      int
}

// IntentRepo persists payment intents and related aggregates.
type IntentRepo interface {
	CreateIntent(ctx context.Context, i domain.PaymentIntent) error
	UpdateIntent(ctx context.Context, i domain.PaymentIntent) error
	GetIntent(ctx context.Context, tenantID, id uuid.UUID) (domain.PaymentIntent, error)
	GetByIdempotencyKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.PaymentIntent, error)
	ListIntents(ctx context.Context, f IntentFilter) ([]domain.PaymentIntent, int, error)

	CreateAttempt(ctx context.Context, a domain.PaymentAttempt) error
	ListAttempts(ctx context.Context, tenantID, intentID uuid.UUID) ([]domain.PaymentAttempt, error)

	CreateMethod(ctx context.Context, m domain.PaymentMethod) error
	GetMethod(ctx context.Context, tenantID, id uuid.UUID) (domain.PaymentMethod, error)
	ListMethods(ctx context.Context, tenantID, principalID uuid.UUID) ([]domain.PaymentMethod, error)

	CreateRefund(ctx context.Context, r domain.Refund) error
	UpdateRefund(ctx context.Context, r domain.Refund) error
	GetRefundByIdempotency(ctx context.Context, tenantID uuid.UUID, key string) (domain.Refund, error)
	ListRefunds(ctx context.Context, tenantID, intentID uuid.UUID) ([]domain.Refund, error)

	CreateChargeback(ctx context.Context, c domain.Chargeback) error
	ListChargebacks(ctx context.Context, tenantID uuid.UUID, intentID *uuid.UUID) ([]domain.Chargeback, error)

	UpsertRoute(ctx context.Context, r domain.ProviderRoute) error
	ListRoutes(ctx context.Context, tenantID uuid.UUID, method domain.PaymentMethodType, currency string) ([]domain.ProviderRoute, error)

	CreateFraudScore(ctx context.Context, f domain.FraudScore) error
	CreateAudit(ctx context.Context, a domain.AuditEntry) error
}

// OutboxRepository persists transactional outbox rows.
type OutboxRepository interface {
	Enqueue(ctx context.Context, m domain.OutboxMessage) error
	Update(ctx context.Context, m domain.OutboxMessage) error
	ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
}

// AuthorizeRequest is a PSP authorize call.
type AuthorizeRequest struct {
	IntentID       uuid.UUID
	TenantID       uuid.UUID
	AmountMinor    int64
	Currency       string
	Token          string
	IdempotencyKey string
	Metadata       map[string]any
}

// AuthorizeResult is a PSP authorize response.
type AuthorizeResult struct {
	ProviderRef string
	Success     bool
	ErrorCode   string
	ErrorMessage string
}

// CaptureRequest is a PSP capture call.
type CaptureRequest struct {
	IntentID       uuid.UUID
	TenantID       uuid.UUID
	ProviderRef    string
	AmountMinor    int64
	Currency       string
	IdempotencyKey string
}

// CaptureResult is a PSP capture response.
type CaptureResult struct {
	ProviderRef string
	Success     bool
	ErrorCode   string
	ErrorMessage string
}

// VoidRequest is a PSP void call.
type VoidRequest struct {
	IntentID       uuid.UUID
	TenantID       uuid.UUID
	ProviderRef    string
	IdempotencyKey string
}

// VoidResult is a PSP void response.
type VoidResult struct {
	ProviderRef string
	Success     bool
	ErrorCode   string
	ErrorMessage string
}

// RefundRequest is a PSP refund call.
type RefundRequest struct {
	IntentID       uuid.UUID
	TenantID       uuid.UUID
	ProviderRef    string
	AmountMinor    int64
	Currency       string
	IdempotencyKey string
	Reason         string
}

// RefundResult is a PSP refund response.
type RefundResult struct {
	ProviderRef string
	Success     bool
	ErrorCode   string
	ErrorMessage string
}

// PSPClient talks to a payment service provider.
type PSPClient interface {
	Name() string
	Authorize(ctx context.Context, req AuthorizeRequest) (AuthorizeResult, error)
	Capture(ctx context.Context, req CaptureRequest) (CaptureResult, error)
	Void(ctx context.Context, req VoidRequest) (VoidResult, error)
	Refund(ctx context.Context, req RefundRequest) (RefundResult, error)
}

// FraudRequest asks for a risk score.
type FraudRequest struct {
	TenantID    uuid.UUID
	PrincipalID uuid.UUID
	IntentID    uuid.UUID
	AmountMinor int64
	Currency    string
	MethodType  domain.PaymentMethodType
	OrderID     string
}

// FraudResult is risk scoring outcome.
type FraudResult struct {
	Score     int
	Decision  domain.FraudDecision
	Reasons   []string
}

// FraudClient talks to fraud-service.
type FraudClient interface {
	Score(ctx context.Context, req FraudRequest) (FraudResult, error)
}

// WalletDebitRequest asks wallet-service to debit for pay-with-wallet.
type WalletDebitRequest struct {
	TenantID       uuid.UUID
	PrincipalID    uuid.UUID
	AmountMinor    int64
	Currency       string
	AccountType    string
	IdempotencyKey string
	Reference      string
}

// WalletDebitResult is the wallet debit outcome.
type WalletDebitResult struct {
	WalletID  string
	EntryID   string
	Success   bool
	Reason    string
}

// WalletClient talks to wallet-service (pay-with-wallet only).
type WalletClient interface {
	Debit(ctx context.Context, req WalletDebitRequest) (WalletDebitResult, error)
}

// JournalLine is a double-entry line stub for finance-ledger.
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

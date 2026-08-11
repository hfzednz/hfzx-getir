// Package ports defines application-layer dependency interfaces (hexagonal ports).
package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/pricing-service/internal/domain"
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

// PriceRepo persists price books and entries.
type PriceRepo interface {
	UpsertBook(ctx context.Context, b domain.PriceBook) error
	GetBook(ctx context.Context, tenantID, bookID uuid.UUID) (domain.PriceBook, error)
	ListBooks(ctx context.Context, tenantID uuid.UUID) ([]domain.PriceBook, error)

	UpsertEntry(ctx context.Context, e domain.PriceEntry) error
	GetEntry(ctx context.Context, tenantID, entryID uuid.UUID) (domain.PriceEntry, error)
	ListEntries(ctx context.Context, tenantID uuid.UUID, bookID *uuid.UUID, variantID *uuid.UUID) ([]domain.PriceEntry, error)
	ListEntriesForVariant(ctx context.Context, tenantID, variantID uuid.UUID) ([]domain.PriceEntry, error)
}

// TaxRepo persists tax display rules.
type TaxRepo interface {
	Upsert(ctx context.Context, t domain.TaxRule) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.TaxRule, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.TaxRule, error)
}

// DynamicRepo persists dynamic pricing rules.
type DynamicRepo interface {
	Upsert(ctx context.Context, r domain.DynamicRule) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.DynamicRule, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.DynamicRule, error)
	ListActive(ctx context.Context, tenantID uuid.UUID) ([]domain.DynamicRule, error)
}

// QuoteAuditRepo persists quote audit snapshots.
type QuoteAuditRepo interface {
	Create(ctx context.Context, a domain.QuoteAudit) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.QuoteAudit, error)
	List(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.QuoteAudit, error)
}

// OutboxRepository persists transactional outbox rows.
type OutboxRepository interface {
	Enqueue(ctx context.Context, m domain.OutboxMessage) error
	Update(ctx context.Context, m domain.OutboxMessage) error
	ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
}

// PromoLineInput is a cart line for promotion evaluation.
type PromoLineInput struct {
	VariantID      uuid.UUID
	Qty            int
	UnitPriceMinor int64
	LineTotalMinor int64
}

// PromoEvaluateRequest asks promotion-service to evaluate discounts.
type PromoEvaluateRequest struct {
	TenantID    uuid.UUID
	CustomerID  *uuid.UUID
	Currency    string
	CouponCodes []string
	Lines       []PromoLineInput
	SubtotalMinor int64
}

// PromoDiscountResult is one applicable discount from Evaluate.
type PromoDiscountResult struct {
	PromotionID   string
	Code          string
	DiscountMinor int64
	Description   string
	VariantID     *uuid.UUID // optional line allocation
}

// PromoEvaluateResult is the Evaluate response.
type PromoEvaluateResult struct {
	Discounts      []PromoDiscountResult
	DiscountMinor  int64
}

// PromoClient calls promotion-service Evaluate (no local promo storage).
type PromoClient interface {
	Evaluate(ctx context.Context, req PromoEvaluateRequest) (PromoEvaluateResult, error)
}

// DynamicHintRequest asks for inventory/demand hints for dynamic pricing.
type DynamicHintRequest struct {
	TenantID  uuid.UUID
	VariantID uuid.UUID
	WarehouseID *uuid.UUID
}

// DynamicHintResult is an optional inventory/demand hint.
type DynamicHintResult struct {
	AvailableQty *int
	DemandScore  *float64
}

// DynamicHintClient is an optional port for inventory/demand hints (stub ok).
type DynamicHintClient interface {
	Hint(ctx context.Context, req DynamicHintRequest) (DynamicHintResult, error)
}

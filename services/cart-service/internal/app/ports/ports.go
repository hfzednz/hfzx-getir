// Package ports defines application-layer dependency interfaces (hexagonal ports).
package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/cart-service/internal/domain"
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

// CartRepository persists the cart aggregate.
type CartRepository interface {
	Create(ctx context.Context, c domain.Cart) error
	Update(ctx context.Context, c domain.Cart) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Cart, error)
	GetActiveByGuest(ctx context.Context, tenantID uuid.UUID, guestToken string) (domain.Cart, error)
	GetActiveByPrincipal(ctx context.Context, tenantID, principalID uuid.UUID) (domain.Cart, error)
}

// EventStore persists append-only cart timeline events.
type EventStore interface {
	Append(ctx context.Context, e domain.CartEvent) error
	ListByCart(ctx context.Context, tenantID, cartID uuid.UUID) ([]domain.CartEvent, error)
}

// OutboxRepository persists transactional outbox rows.
type OutboxRepository interface {
	Enqueue(ctx context.Context, m domain.OutboxMessage) error
	Update(ctx context.Context, m domain.OutboxMessage) error
	ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
}

// SavedCartRepository persists save-for-later snapshots.
type SavedCartRepository interface {
	Create(ctx context.Context, s domain.SavedCart) error
	ListByPrincipal(ctx context.Context, tenantID, principalID uuid.UUID) ([]domain.SavedCart, error)
}

// QuoteLineInput is a cart line sent to PricingClient.
type QuoteLineInput struct {
	VariantID uuid.UUID
	Qty       int
}

// QuoteRequest asks pricing-service for a cart quote (minor units).
type QuoteRequest struct {
	TenantID    uuid.UUID
	CartID      uuid.UUID
	Currency    string
	CityID      *uuid.UUID
	CouponCodes []string
	Lines       []QuoteLineInput
}

// QuoteResult is the pricing quote response (all minor units).
type QuoteResult struct {
	QuoteID        uuid.UUID
	Currency       string
	SubtotalMinor  int64
	DiscountMinor  int64
	TaxMinor       int64
	DeliveryMinor  int64
	ServiceMinor   int64
	PackagingMinor int64
	TipMinor       int64
	TotalMinor     int64
	LineQuotes     []domain.LineQuote
	QuotedAt       time.Time
}

// PricingClient talks to pricing-service.
type PricingClient interface {
	Quote(ctx context.Context, req QuoteRequest) (QuoteResult, error)
}

// ATPLine is a variant qty for availability check.
type ATPLine struct {
	VariantID uuid.UUID
	Qty       int
}

// ATPRequest asks inventory-service for available-to-promise.
type ATPRequest struct {
	TenantID uuid.UUID
	CityID   *uuid.UUID
	Lines    []ATPLine
}

// ATPLineResult is per-variant ATP.
type ATPLineResult struct {
	VariantID   uuid.UUID
	Available   int
	WarehouseID *uuid.UUID
}

// ATPResult is the ATP response.
type ATPResult struct {
	Lines []ATPLineResult
}

// SoftReserveLine is a line for inventory soft reservation.
type SoftReserveLine struct {
	VariantID uuid.UUID
	Qty       int
}

// SoftReserveRequest asks inventory-service for a soft hold.
type SoftReserveRequest struct {
	TenantID       uuid.UUID
	CartID         uuid.UUID
	IdempotencyKey string
	Lines          []SoftReserveLine
}

// SoftReserveResult is the opaque reservation ref from inventory.
type SoftReserveResult struct {
	ReservationRef string
	ExpiresAt      *time.Time
}

// ReleaseRequest releases a soft reservation.
type ReleaseRequest struct {
	TenantID       uuid.UUID
	ReservationRef string
	IdempotencyKey string
}

// InventoryClient talks to inventory-service (opaque refs only; no stock ledger).
type InventoryClient interface {
	ATP(ctx context.Context, req ATPRequest) (ATPResult, error)
	SoftReserve(ctx context.Context, req SoftReserveRequest) (SoftReserveResult, error)
	Release(ctx context.Context, req ReleaseRequest) error
}

// RecommendRequest asks for cart recommendations.
type RecommendRequest struct {
	TenantID uuid.UUID
	CartID   uuid.UUID
	CityID   *uuid.UUID
	Limit    int
}

// RecommendItem is a suggested opaque variant.
type RecommendItem struct {
	VariantID uuid.UUID
	Score     float64
	Reason    string
}

// RecommendResult is the recommendation response.
type RecommendResult struct {
	Items []RecommendItem
}

// RecommendClient talks to a recommendation / personalization service.
type RecommendClient interface {
	Recommend(ctx context.Context, req RecommendRequest) (RecommendResult, error)
}

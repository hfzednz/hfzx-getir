// Package ports defines application-layer dependency interfaces (hexagonal ports).
package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/checkout-service/internal/domain"
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

// SessionFilter lists checkout sessions.
type SessionFilter struct {
	TenantID    uuid.UUID
	PrincipalID *uuid.UUID
	Status      *domain.SessionStatus
	Query       string
	Limit       int
	Offset      int
}

// CheckoutRepo persists checkout sessions.
type CheckoutRepo interface {
	Create(ctx context.Context, s domain.Session) error
	Update(ctx context.Context, s domain.Session) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Session, error)
	GetByIdempotencyKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.Session, error)
	GetByRecoveryToken(ctx context.Context, token string) (domain.Session, error)
	List(ctx context.Context, f SessionFilter) ([]domain.Session, int, error)
	CountByStatus(ctx context.Context, tenantID uuid.UUID) (map[domain.SessionStatus]int, error)
}

// EventStore persists append-only checkout timeline events.
type EventStore interface {
	Append(ctx context.Context, e domain.SessionEvent) error
	ListBySession(ctx context.Context, tenantID, sessionID uuid.UUID) ([]domain.SessionEvent, error)
}

// OutboxRepository persists transactional outbox rows.
type OutboxRepository interface {
	Enqueue(ctx context.Context, m domain.OutboxMessage) error
	Update(ctx context.Context, m domain.OutboxMessage) error
	ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
}

// CartLine is a line returned by cart-service.
type CartLine struct {
	VariantID      uuid.UUID
	SKUCode        string
	TitleSnapshot  string
	Qty            int
	UnitPriceMinor int64
	Notes          string
	AgeRestricted  bool
	WarehouseID    *uuid.UUID
}

// CartView is the cart aggregate snapshot needed by checkout.
type CartView struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	PrincipalID uuid.UUID
	GuestID     string
	CityID      string
	Currency    string
	CouponCodes []string
	Lines       []CartLine
	Active      bool
}

// CartClient fetches cart + lines from cart-service.
type CartClient interface {
	GetCart(ctx context.Context, tenantID, cartID uuid.UUID) (CartView, error)
}

// QuoteLineRequest is a line for pricing preview.
type QuoteLineRequest struct {
	VariantID uuid.UUID
	SKUCode   string
	Qty       int
}

// QuoteRequest asks pricing-service for a preview.
type QuoteRequest struct {
	TenantID       uuid.UUID
	CartID         uuid.UUID
	CheckoutID     uuid.UUID
	Currency       string
	CityID         string
	DeliveryOption domain.DeliveryOption
	CouponCodes    []string
	TipMinor       int64
	Lines          []QuoteLineRequest
}

// QuoteResult is a pricing preview in minor units.
type QuoteResult struct {
	QuoteID        string
	Currency       string
	SubtotalMinor  int64
	DiscountMinor  int64
	TaxMinor       int64
	DeliveryMinor  int64
	ServiceMinor   int64
	PackagingMinor int64
	TipMinor       int64
	TotalMinor     int64
	QuotedAt       time.Time
}

// PricingClient talks to pricing-service.
type PricingClient interface {
	Quote(ctx context.Context, req QuoteRequest) (QuoteResult, error)
}

// ATPLineRequest is a line for available-to-promise check.
type ATPLineRequest struct {
	VariantID   uuid.UUID
	SKUCode     string
	Qty         int
	WarehouseID *uuid.UUID
}

// ATPRequest asks inventory-service for ATP (no ledger mutation).
type ATPRequest struct {
	TenantID uuid.UUID
	CityID   string
	Lines    []ATPLineRequest
}

// ATPLineResult is per-line availability.
type ATPLineResult struct {
	VariantID   uuid.UUID
	Available   bool
	AvailableQty int
	Reason      string
}

// ATPResult is the ATP check outcome.
type ATPResult struct {
	AllAvailable bool
	Lines        []ATPLineResult
}

// InventoryClient talks to inventory-service (ATP only; no ledger).
type InventoryClient interface {
	CheckATP(ctx context.Context, req ATPRequest) (ATPResult, error)
}

// GeofenceRequest asks whether an address is in a delivery zone.
type GeofenceRequest struct {
	TenantID       uuid.UUID
	CityID         string
	Lat            float64
	Lng            float64
	DeliveryOption domain.DeliveryOption
}

// GeofenceResult is zone eligibility.
type GeofenceResult struct {
	InZone      bool
	ZoneID      string
	Reason      string
	MinOrderMinor int64
}

// GeofenceClient talks to geofence-service.
type GeofenceClient interface {
	CheckZone(ctx context.Context, req GeofenceRequest) (GeofenceResult, error)
}

// FraudRequest asks for a risk score.
type FraudRequest struct {
	TenantID    uuid.UUID
	PrincipalID uuid.UUID
	CheckoutID  uuid.UUID
	TotalMinor  int64
	Currency    string
	CityID      string
}

// FraudResult is risk scoring outcome.
type FraudResult struct {
	Score     float64
	Decision  string // allow | review | block
	Reason    string
}

// FraudClient talks to fraud-service.
type FraudClient interface {
	Score(ctx context.Context, req FraudRequest) (FraudResult, error)
}

// PaymentEligibilityRequest checks whether payment methods are available.
type PaymentEligibilityRequest struct {
	TenantID    uuid.UUID
	PrincipalID uuid.UUID
	TotalMinor  int64
	Currency    string
}

// PaymentEligibilityResult is eligibility without capture.
type PaymentEligibilityResult struct {
	Eligible bool
	Reason   string
	Methods  []string
}

// PaymentEligibilityClient talks to payment-service eligibility API only.
type PaymentEligibilityClient interface {
	Check(ctx context.Context, req PaymentEligibilityRequest) (PaymentEligibilityResult, error)
}

// CreateFromCheckoutLine is a priced line for order create.
type CreateFromCheckoutLine struct {
	VariantID      uuid.UUID
	SKUCode        string
	TitleSnapshot  string
	Qty            int
	UnitPriceMinor int64
	DiscountsMinor int64
	TaxMinor       int64
	WarehouseID    *uuid.UUID
}

// CreateFromCheckoutRequest asks order-service to create from checkout.
type CreateFromCheckoutRequest struct {
	TenantID            uuid.UUID
	CustomerPrincipalID uuid.UUID
	CheckoutSessionID   uuid.UUID
	CartID              uuid.UUID
	Currency            string
	IdempotencyKey      string
	AddressSnapshot     map[string]any
	Notes               string
	Gift                map[string]any
	DeliveryOption      domain.DeliveryOption
	ScheduledAt         *time.Time
	TipMinor            int64
	DiscountMinor       int64
	ShippingMinor       int64
	TaxMinor            int64
	SubtotalMinor       int64
	TotalMinor          int64
	Lines               []CreateFromCheckoutLine
	Metadata            map[string]any
}

// CreateFromCheckoutResult is the opaque order id from order-service.
type CreateFromCheckoutResult struct {
	OrderID string
	Status  string
}

// PlaceOrderRequest optionally triggers place after create (OMS owns saga).
type PlaceOrderRequest struct {
	TenantID       uuid.UUID
	OrderID        string
	IdempotencyKey string
}

// PlaceOrderResult is the opaque place outcome.
type PlaceOrderResult struct {
	OrderID string
	Status  string
}

// OrderClient talks to order-service (CreateFromCheckout; optional PlaceOrder).
type OrderClient interface {
	CreateFromCheckout(ctx context.Context, req CreateFromCheckoutRequest) (CreateFromCheckoutResult, error)
	PlaceOrder(ctx context.Context, req PlaceOrderRequest) (PlaceOrderResult, error)
}

// PromoRequest validates coupon eligibility at checkout.
type PromoRequest struct {
	TenantID    uuid.UUID
	PrincipalID uuid.UUID
	Codes       []string
	SubtotalMinor int64
	Currency    string
	CityID      string
}

// PromoResult is coupon eligibility.
type PromoResult struct {
	Valid         bool
	Reason        string
	DiscountMinor int64
	CodesApplied  []string
}

// PromoClient talks to promo-service.
type PromoClient interface {
	Validate(ctx context.Context, req PromoRequest) (PromoResult, error)
}

// CustomerCheckRequest validates the shopper is active.
type CustomerCheckRequest struct {
	TenantID    uuid.UUID
	PrincipalID uuid.UUID
}

// CustomerCheckResult is customer eligibility.
type CustomerCheckResult struct {
	Active bool
	Reason string
	Age    int
}

// CustomerClient optionally validates the principal (via cart or profile).
// Implemented by CartClient memory / stub when no dedicated profile call.
type CustomerClient interface {
	Check(ctx context.Context, req CustomerCheckRequest) (CustomerCheckResult, error)
}

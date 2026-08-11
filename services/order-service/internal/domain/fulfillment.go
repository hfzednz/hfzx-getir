package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// FulfillmentStatus is the OMS projection of a warehouse fulfillment split.
type FulfillmentStatus string

const (
	FulfillmentStatusPending    FulfillmentStatus = "pending"
	FulfillmentStatusAssigned   FulfillmentStatus = "assigned"
	FulfillmentStatusPicking    FulfillmentStatus = "picking"
	FulfillmentStatusPacking    FulfillmentStatus = "packing"
	FulfillmentStatusReady      FulfillmentStatus = "ready"
	FulfillmentStatusDispatched FulfillmentStatus = "dispatched"
	FulfillmentStatusDelivered  FulfillmentStatus = "delivered"
	FulfillmentStatusCancelled  FulfillmentStatus = "cancelled"
	FulfillmentStatusFailed     FulfillmentStatus = "failed"
)

// Valid reports whether the fulfillment status is recognized.
func (s FulfillmentStatus) Valid() bool {
	switch s {
	case FulfillmentStatusPending, FulfillmentStatusAssigned, FulfillmentStatusPicking,
		FulfillmentStatusPacking, FulfillmentStatusReady, FulfillmentStatusDispatched,
		FulfillmentStatusDelivered, FulfillmentStatusCancelled, FulfillmentStatusFailed:
		return true
	default:
		return false
	}
}

// Fulfillment is a split fulfillment unit projected from warehouse-service.
type Fulfillment struct {
	ID             uuid.UUID
	OrderID        uuid.UUID
	TenantID       uuid.UUID
	WarehouseID    uuid.UUID
	Status         FulfillmentStatus
	ReservationID  string
	FulfillmentRef string
	Priority       int
	LineIDs        []uuid.UUID
	Metadata       map[string]any
	CreatedAt      time.Time
	UpdatedAt      time.Time
	CancelledAt    *time.Time
	CompletedAt    *time.Time
}

// Validate checks fulfillment invariants.
func (f Fulfillment) Validate() error {
	if f.ID == uuid.Nil {
		return fmt.Errorf("%w: fulfillment id required", ErrInvalidArgument)
	}
	if f.OrderID == uuid.Nil {
		return fmt.Errorf("%w: order_id required", ErrInvalidArgument)
	}
	if f.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if f.WarehouseID == uuid.Nil {
		return fmt.Errorf("%w: warehouse_id required", ErrInvalidArgument)
	}
	if !f.Status.Valid() {
		return fmt.Errorf("%w: invalid fulfillment status %q", ErrInvalidArgument, f.Status)
	}
	return nil
}

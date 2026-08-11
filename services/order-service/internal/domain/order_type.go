package domain

import "fmt"

// OrderType classifies how an order is fulfilled / billed.
type OrderType string

const (
	OrderTypeInstant        OrderType = "instant"
	OrderTypeScheduled      OrderType = "scheduled"
	OrderTypeExpress        OrderType = "express"
	OrderTypePickup         OrderType = "pickup"
	OrderTypeGift           OrderType = "gift"
	OrderTypeSubscription   OrderType = "subscription"
	OrderTypeCorporate      OrderType = "corporate"
	OrderTypeMultiWarehouse OrderType = "multi_warehouse"
	OrderTypeSplit          OrderType = "split"
	OrderTypeReplacement    OrderType = "replacement"
)

// Valid reports whether the order type is recognized.
func (t OrderType) Valid() bool {
	switch t {
	case OrderTypeInstant, OrderTypeScheduled, OrderTypeExpress, OrderTypePickup,
		OrderTypeGift, OrderTypeSubscription, OrderTypeCorporate,
		OrderTypeMultiWarehouse, OrderTypeSplit, OrderTypeReplacement:
		return true
	default:
		return false
	}
}

// Validate returns an error when the type is unknown.
func (t OrderType) Validate() error {
	if !t.Valid() {
		return fmt.Errorf("%w: invalid order type %q", ErrInvalidArgument, t)
	}
	return nil
}

package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/order-service/internal/app/ports"
	"github.com/nexora/order-service/internal/domain"
)

// ApplyLifecycleEventInput applies a warehouse or dispatch lifecycle event.
type ApplyLifecycleEventInput struct {
	TenantID   uuid.UUID
	OrderID    uuid.UUID
	EventType  string
	Payload    map[string]any
	ActorID    *uuid.UUID
	CourierRef string
}

// ApplyWarehouseEvent advances OMS state from warehouse events
// (PickingStarted, PackingCompleted, …).
func (d *Deps) ApplyWarehouseEvent(ctx context.Context, in ApplyLifecycleEventInput) (domain.Order, error) {
	return d.applyLifecycleEvent(ctx, in)
}

// ApplyDispatchEvent advances OMS state from dispatch events
// (CourierAssigned, OutForDelivery, Delivered).
func (d *Deps) ApplyDispatchEvent(ctx context.Context, in ApplyLifecycleEventInput) (domain.Order, error) {
	return d.applyLifecycleEvent(ctx, in)
}

func (d *Deps) applyLifecycleEvent(ctx context.Context, in ApplyLifecycleEventInput) (domain.Order, error) {
	if in.TenantID == uuid.Nil || in.OrderID == uuid.Nil {
		return domain.Order{}, fmt.Errorf("%w: tenant_id and order_id required", domain.ErrInvalidArgument)
	}
	if in.EventType == "" {
		return domain.Order{}, fmt.Errorf("%w: event type required", domain.ErrInvalidArgument)
	}
	o, err := d.Orders.GetByID(ctx, in.TenantID, in.OrderID)
	if err != nil {
		return domain.Order{}, err
	}

	var target domain.OrderStatus
	switch in.EventType {
	case domain.EventPickingStarted:
		target = domain.OrderStatusPicking
	case domain.EventPackingCompleted:
		// packing then ready_for_dispatch
		if o.Status == domain.OrderStatusPicking || o.Status == domain.OrderStatusWarehouseAssigned {
			if o.Status == domain.OrderStatusWarehouseAssigned {
				if err := d.transition(&o, domain.OrderStatusPicking); err != nil {
					return domain.Order{}, err
				}
			}
			if err := d.transition(&o, domain.OrderStatusPacking); err != nil {
				return domain.Order{}, err
			}
		}
		target = domain.OrderStatusReadyForDispatch
	case domain.EventCourierAssigned:
		target = domain.OrderStatusCourierAssigned
		if in.CourierRef != "" {
			o.CourierRef = in.CourierRef
		} else if ref, ok := in.Payload["courierRef"].(string); ok {
			o.CourierRef = ref
		}
	case domain.EventOutForDelivery:
		target = domain.OrderStatusOutForDelivery
	case domain.EventDelivered:
		target = domain.OrderStatusDelivered
	default:
		return domain.Order{}, fmt.Errorf("%w: unsupported lifecycle event %s", domain.ErrInvalidArgument, in.EventType)
	}

	if err := d.transition(&o, target); err != nil {
		return domain.Order{}, err
	}
	if err := d.appendEvent(ctx, o.ID, o.TenantID, in.EventType, in.Payload); err != nil {
		return domain.Order{}, err
	}

	// After packing ready: request dispatch.
	if in.EventType == domain.EventPackingCompleted && d.Dispatch != nil {
		fulRef := ""
		if d.Fulfillments != nil {
			if fs, err := d.Fulfillments.ListByOrder(ctx, o.TenantID, o.ID); err == nil && len(fs) > 0 {
				fulRef = fs[0].FulfillmentRef
			}
		}
		res, err := d.Dispatch.RequestDispatch(ctx, ports.RequestDispatchRequest{
			TenantID: o.TenantID, OrderID: o.ID, FulfillmentRef: fulRef,
			IdempotencyKey: "dispatch:" + o.ID.String(),
		})
		if err == nil && res.DispatchRef != "" {
			o.CourierRef = res.DispatchRef
		}
	}

	// Delivered → Completed path.
	if in.EventType == domain.EventDelivered {
		if err := d.transition(&o, domain.OrderStatusCompleted); err != nil {
			return domain.Order{}, err
		}
		_ = d.appendEvent(ctx, o.ID, o.TenantID, domain.EventCompleted, map[string]any{
			"from": string(domain.OrderStatusDelivered),
		})
	}

	if err := d.Orders.Update(ctx, o); err != nil {
		return domain.Order{}, err
	}
	d.indexOrder(ctx, o)
	return o, nil
}

package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/order-service/internal/app/ports"
	"github.com/nexora/order-service/internal/domain"
)

// SetPriorityInput updates order priority.
type SetPriorityInput struct {
	TenantID uuid.UUID
	OrderID  uuid.UUID
	Priority int
}

// SetPriority sets admin priority on an order.
func (d *Deps) SetPriority(ctx context.Context, in SetPriorityInput) (domain.Order, error) {
	o, err := d.Orders.GetByID(ctx, in.TenantID, in.OrderID)
	if err != nil {
		return domain.Order{}, err
	}
	o.Priority = in.Priority
	o.UpdatedAt = d.now()
	o.Version++
	if err := d.Orders.Update(ctx, o); err != nil {
		return domain.Order{}, err
	}
	d.indexOrder(ctx, o)
	return o, nil
}

// SplitOrderInput creates fulfillment splits by warehouse.
type SplitOrderInput struct {
	TenantID uuid.UUID
	OrderID  uuid.UUID
	Splits   []SplitUnitInput
}

// SplitUnitInput describes one warehouse split.
type SplitUnitInput struct {
	WarehouseID uuid.UUID
	LineIDs     []uuid.UUID
	Priority    int
}

// SplitOrder creates fulfillment split units for an order.
func (d *Deps) SplitOrder(ctx context.Context, in SplitOrderInput) ([]domain.Fulfillment, error) {
	o, err := d.Orders.GetByID(ctx, in.TenantID, in.OrderID)
	if err != nil {
		return nil, err
	}
	if len(in.Splits) == 0 {
		return nil, fmt.Errorf("%w: at least one split required", domain.ErrInvalidArgument)
	}
	if d.Fulfillments == nil {
		return nil, fmt.Errorf("%w: fulfillments repository not configured", domain.ErrInvariant)
	}
	now := d.now()
	out := make([]domain.Fulfillment, 0, len(in.Splits))
	whIDs := make([]uuid.UUID, 0, len(in.Splits))
	for _, s := range in.Splits {
		if s.WarehouseID == uuid.Nil {
			return nil, fmt.Errorf("%w: warehouse_id required", domain.ErrInvalidArgument)
		}
		f := domain.Fulfillment{
			ID: d.newID(), OrderID: o.ID, TenantID: o.TenantID,
			WarehouseID: s.WarehouseID, Status: domain.FulfillmentStatusPending,
			ReservationID: o.ReservationRef, Priority: s.Priority,
			LineIDs: s.LineIDs, Metadata: map[string]any{},
			CreatedAt: now, UpdatedAt: now,
		}
		if err := f.Validate(); err != nil {
			return nil, err
		}
		if err := d.Fulfillments.Create(ctx, f); err != nil {
			return nil, err
		}
		out = append(out, f)
		whIDs = append(whIDs, s.WarehouseID)
	}
	o.Type = domain.OrderTypeSplit
	o.WarehouseIDs = whIDs
	o.UpdatedAt = now
	o.Version++
	_ = d.Orders.Update(ctx, o)
	d.indexOrder(ctx, o)
	return out, nil
}

// InterveneStatusInput is a guarded admin status override.
type InterveneStatusInput struct {
	TenantID   uuid.UUID
	OrderID    uuid.UUID
	NextStatus domain.OrderStatus
	Reason     string
	Force      bool // when true, still requires a legal transition (no illegal jumps)
	ActorID    *uuid.UUID
}

// InterveneStatus applies a guarded status transition (still enforces transition table).
func (d *Deps) InterveneStatus(ctx context.Context, in InterveneStatusInput) (domain.Order, error) {
	o, err := d.Orders.GetByID(ctx, in.TenantID, in.OrderID)
	if err != nil {
		return domain.Order{}, err
	}
	if in.Reason == "" {
		return domain.Order{}, fmt.Errorf("%w: reason required for intervene", domain.ErrInvalidArgument)
	}
	// Guarded: never allow illegal transitions even for admin.
	if err := d.transition(&o, in.NextStatus); err != nil {
		return domain.Order{}, err
	}
	_ = d.appendEvent(ctx, o.ID, o.TenantID, "AdminIntervene", map[string]any{
		"nextStatus": string(in.NextStatus),
		"reason":     in.Reason,
		"force":      in.Force,
	})
	if err := d.Orders.Update(ctx, o); err != nil {
		return domain.Order{}, err
	}
	d.indexOrder(ctx, o)
	return o, nil
}

// Timeline returns the append-only order event timeline.
func (d *Deps) Timeline(ctx context.Context, tenantID, orderID uuid.UUID) ([]domain.OrderEvent, error) {
	if _, err := d.Orders.GetByID(ctx, tenantID, orderID); err != nil {
		return nil, err
	}
	if d.Events == nil {
		return nil, nil
	}
	return d.Events.ListByOrder(ctx, tenantID, orderID)
}

// SearchOrders queries the search index.
func (d *Deps) SearchOrders(ctx context.Context, q ports.SearchQuery) (ports.SearchResult, error) {
	if d.Search == nil {
		return ports.SearchResult{}, fmt.Errorf("%w: search not configured", domain.ErrInvariant)
	}
	if q.Limit <= 0 {
		q.Limit = 50
	}
	return d.Search.Search(ctx, q)
}

// PublishPending drains the outbox (stub publish via EventPublisher).
func (d *Deps) PublishPending(ctx context.Context, limit int) (int, error) {
	if d.Outbox == nil {
		return 0, nil
	}
	if limit <= 0 {
		limit = 100
	}
	msgs, err := d.Outbox.ListPending(ctx, limit)
	if err != nil {
		return 0, err
	}
	published := 0
	for _, m := range msgs {
		if d.Publisher != nil {
			if err := d.Publisher.Publish(ctx, m.Topic, m.Key, m.Payload); err != nil {
				m.Attempts++
				m.LastError = err.Error()
				m.Status = domain.OutboxStatusFailed
				m.UpdatedAt = d.now()
				_ = d.Outbox.Update(ctx, m)
				continue
			}
		}
		now := d.now()
		m.Status = domain.OutboxStatusPublished
		m.PublishedAt = &now
		m.UpdatedAt = now
		m.Attempts++
		_ = d.Outbox.Update(ctx, m)
		published++
	}
	return published, nil
}

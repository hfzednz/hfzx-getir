package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/order-service/internal/app/ports"
	"github.com/nexora/order-service/internal/domain"
)

// CreateLineInput is a priced line snapshot at create time.
type CreateLineInput struct {
	VariantID      uuid.UUID
	SKUCode        string
	TitleSnapshot  string
	Qty            int
	UnitPriceMinor int64
	DiscountsMinor int64
	TaxMinor       int64
	WarehouseID    *uuid.UUID
	SortOrder      int
	Metadata       map[string]any
}

// CreateDraftInput creates a draft order (idempotent).
type CreateDraftInput struct {
	TenantID            uuid.UUID
	CustomerPrincipalID uuid.UUID
	Type                domain.OrderType
	Currency            string
	IdempotencyKey      string
	AddressSnapshot     map[string]any
	Notes               string
	Gift                map[string]any
	Priority            int
	WarehouseIDs        []uuid.UUID
	ScheduledAt         *time.Time
	Lines               []CreateLineInput
	DiscountMinor       int64
	ShippingMinor       int64
	TipMinor            int64
	Metadata            map[string]any
}

// CreateFromCheckoutInput creates an order ready to place (pending_payment).
type CreateFromCheckoutInput struct {
	CreateDraftInput
}

// CreateDraft creates a draft order. Same idempotency key returns the existing order.
func (d *Deps) CreateDraft(ctx context.Context, in CreateDraftInput) (domain.Order, error) {
	return d.createOrder(ctx, in, domain.OrderStatusDraft)
}

// CreateFromCheckout creates an order from checkout in pending_payment status.
func (d *Deps) CreateFromCheckout(ctx context.Context, in CreateFromCheckoutInput) (domain.Order, error) {
	return d.createOrder(ctx, in.CreateDraftInput, domain.OrderStatusPendingPayment)
}

func (d *Deps) createOrder(ctx context.Context, in CreateDraftInput, status domain.OrderStatus) (domain.Order, error) {
	if in.TenantID == uuid.Nil {
		return domain.Order{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	if in.CustomerPrincipalID == uuid.Nil {
		return domain.Order{}, fmt.Errorf("%w: customer_principal_id required", domain.ErrInvalidArgument)
	}
	key := strings.TrimSpace(in.IdempotencyKey)
	if key == "" {
		return domain.Order{}, fmt.Errorf("%w: idempotency_key required", domain.ErrInvalidArgument)
	}
	if existing, err := d.Orders.GetByIdempotencyKey(ctx, in.TenantID, key); err == nil {
		return existing, nil
	}

	ot := in.Type
	if ot == "" {
		ot = domain.OrderTypeInstant
	}
	if err := ot.Validate(); err != nil {
		return domain.Order{}, err
	}
	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if currency == "" {
		return domain.Order{}, fmt.Errorf("%w: currency required", domain.ErrInvalidArgument)
	}
	if len(in.Lines) == 0 {
		return domain.Order{}, fmt.Errorf("%w: at least one line required", domain.ErrInvalidArgument)
	}

	now := d.now()
	orderID := d.newID()
	lines := make([]domain.OrderLine, 0, len(in.Lines))
	var subtotal, tax int64
	whSet := map[uuid.UUID]struct{}{}
	for i, li := range in.Lines {
		lineTotal, err := domain.ComputeLineTotalMinor(li.Qty, li.UnitPriceMinor, li.DiscountsMinor, li.TaxMinor)
		if err != nil {
			return domain.Order{}, fmt.Errorf("line[%d]: %w", i, err)
		}
		gross := li.UnitPriceMinor*int64(li.Qty) - li.DiscountsMinor
		subtotal += gross
		tax += li.TaxMinor
		sort := li.SortOrder
		if sort == 0 {
			sort = i
		}
		line := domain.OrderLine{
			ID:             d.newID(),
			OrderID:        orderID,
			TenantID:       in.TenantID,
			VariantID:      li.VariantID,
			SKUCode:        li.SKUCode,
			TitleSnapshot:  li.TitleSnapshot,
			Qty:            li.Qty,
			UnitPriceMinor: li.UnitPriceMinor,
			DiscountsMinor: li.DiscountsMinor,
			TaxMinor:       li.TaxMinor,
			LineTotalMinor: lineTotal,
			WarehouseID:    li.WarehouseID,
			SortOrder:      sort,
			Metadata:       li.Metadata,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if line.Metadata == nil {
			line.Metadata = map[string]any{}
		}
		if err := line.Validate(); err != nil {
			return domain.Order{}, fmt.Errorf("line[%d]: %w", i, err)
		}
		lines = append(lines, line)
		if li.WarehouseID != nil {
			whSet[*li.WarehouseID] = struct{}{}
		}
	}
	warehouseIDs := in.WarehouseIDs
	if len(warehouseIDs) == 0 {
		for id := range whSet {
			warehouseIDs = append(warehouseIDs, id)
		}
	}
	total := subtotal - in.DiscountMinor + tax + in.ShippingMinor + in.TipMinor
	if total < 0 {
		return domain.Order{}, fmt.Errorf("%w: total negative", domain.ErrNegativeMoney)
	}
	if _, err := domain.NewMoney(total, currency); err != nil {
		return domain.Order{}, err
	}

	o := domain.Order{
		ID:                  orderID,
		TenantID:            in.TenantID,
		CustomerPrincipalID: in.CustomerPrincipalID,
		Status:              status,
		Type:                ot,
		Currency:            currency,
		SubtotalMinor:       subtotal,
		DiscountMinor:       in.DiscountMinor,
		TaxMinor:            tax,
		ShippingMinor:       in.ShippingMinor,
		TipMinor:            in.TipMinor,
		TotalMinor:          total,
		AddressSnapshot:     in.AddressSnapshot,
		Notes:               in.Notes,
		Gift:                in.Gift,
		Priority:            in.Priority,
		WarehouseIDs:        warehouseIDs,
		Version:             1,
		IdempotencyKey:      key,
		ScheduledAt:         in.ScheduledAt,
		Lines:               lines,
		Metadata:            in.Metadata,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	if o.AddressSnapshot == nil {
		o.AddressSnapshot = map[string]any{}
	}
	if o.Metadata == nil {
		o.Metadata = map[string]any{}
	}
	if status == domain.OrderStatusPendingPayment {
		o.PlacedAt = &now
	}
	if err := o.Validate(); err != nil {
		return domain.Order{}, err
	}
	if err := d.Orders.Create(ctx, o); err != nil {
		// Race: another create won — return existing.
		if existing, gerr := d.Orders.GetByIdempotencyKey(ctx, in.TenantID, key); gerr == nil {
			return existing, nil
		}
		return domain.Order{}, err
	}
	_ = d.appendEvent(ctx, o.ID, o.TenantID, domain.EventOrderCreated, map[string]any{
		"status": string(o.Status),
		"total":  o.TotalMinor,
	})
	d.indexOrder(ctx, o)
	return o, nil
}

// GetOrder returns an order by id.
func (d *Deps) GetOrder(ctx context.Context, tenantID, orderID uuid.UUID) (domain.Order, error) {
	return d.Orders.GetByID(ctx, tenantID, orderID)
}

// ListOrders lists orders for a tenant.
func (d *Deps) ListOrders(ctx context.Context, tenantID uuid.UUID, status *domain.OrderStatus, limit, offset int) ([]domain.Order, int, error) {
	return d.Orders.List(ctx, ports.OrderFilter{
		TenantID: tenantID, Status: status, Limit: limit, Offset: offset,
	})
}

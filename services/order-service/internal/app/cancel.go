package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/order-service/internal/app/ports"
	"github.com/nexora/order-service/internal/domain"
)

// CancelOrderInput cancels an order with policy-driven compensations.
type CancelOrderInput struct {
	TenantID       uuid.UUID
	OrderID        uuid.UUID
	Reason         string
	IdempotencyKey string
	ActorID        *uuid.UUID
}

// CancelOrder applies cancel policy and compensates (Release + Void/Refund).
func (d *Deps) CancelOrder(ctx context.Context, in CancelOrderInput) (domain.Order, error) {
	if in.TenantID == uuid.Nil || in.OrderID == uuid.Nil {
		return domain.Order{}, fmt.Errorf("%w: tenant_id and order_id required", domain.ErrInvalidArgument)
	}
	o, err := d.Orders.GetByID(ctx, in.TenantID, in.OrderID)
	if err != nil {
		return domain.Order{}, err
	}
	if o.Status == domain.OrderStatusCancelled {
		return o, nil
	}
	action := domain.EvaluateCancel(o.Status)
	if !action.Allowed {
		return domain.Order{}, fmt.Errorf("%w: %s", domain.ErrCancelNotAllowed, action.Reason)
	}

	if action.ReleaseReservation && o.ReservationRef != "" && d.Inventory != nil {
		_ = d.Inventory.Release(ctx, ports.ReleaseRequest{
			TenantID: o.TenantID, ReservationRef: o.ReservationRef,
			IdempotencyKey: "cancel-release:" + o.ID.String(),
		})
	}
	if action.VoidOrRefundPayment && o.PaymentIntentRef != "" && d.Payment != nil {
		_ = d.Payment.Void(ctx, ports.VoidRequest{
			TenantID: o.TenantID, PaymentIntentRef: o.PaymentIntentRef,
			IdempotencyKey: "cancel-void:" + o.ID.String(),
		})
	}

	if err := d.transition(&o, domain.OrderStatusCancelled); err != nil {
		return domain.Order{}, err
	}
	_ = d.appendEvent(ctx, o.ID, o.TenantID, domain.EventCancelled, map[string]any{
		"reason":  in.Reason,
		"phase":   string(action.Phase),
		"release": action.ReleaseReservation,
		"void":    action.VoidOrRefundPayment,
	})
	if err := d.Orders.Update(ctx, o); err != nil {
		return domain.Order{}, err
	}
	d.indexOrder(ctx, o)
	return o, nil
}

// RequestReturnInput opens a return request.
type RequestReturnInput struct {
	TenantID    uuid.UUID
	OrderID     uuid.UUID
	Reason      string
	Notes       string
	Disposition domain.ReturnDisposition
	ActorID     *uuid.UUID
	Lines       []ReturnLineInput
}

// ReturnLineInput is a line being returned.
type ReturnLineInput struct {
	OrderLineID uuid.UUID
	VariantID   uuid.UUID
	Qty         int
	Disposition domain.ReturnDisposition
	Reason      string
}

// RequestReturn creates a return request (post-delivery).
func (d *Deps) RequestReturn(ctx context.Context, in RequestReturnInput) (domain.Return, error) {
	o, err := d.Orders.GetByID(ctx, in.TenantID, in.OrderID)
	if err != nil {
		return domain.Return{}, err
	}
	if err := domain.AssertReturnAllowed(o.Status); err != nil {
		return domain.Return{}, err
	}
	disp := in.Disposition
	if disp == "" {
		disp = domain.ReturnDispositionPending
	}
	now := d.now()
	retID := d.newID()
	lines := make([]domain.ReturnLine, 0, len(in.Lines))
	for _, li := range in.Lines {
		ld := li.Disposition
		if ld == "" {
			ld = disp
		}
		lines = append(lines, domain.ReturnLine{
			ID: d.newID(), ReturnID: retID, OrderLineID: li.OrderLineID,
			VariantID: li.VariantID, Qty: li.Qty, Disposition: ld,
			Reason: li.Reason, Metadata: map[string]any{}, CreatedAt: now,
		})
	}
	if len(lines) == 0 {
		// Default: all order lines.
		for _, ol := range o.Lines {
			lines = append(lines, domain.ReturnLine{
				ID: d.newID(), ReturnID: retID, OrderLineID: ol.ID,
				VariantID: ol.VariantID, Qty: ol.Qty, Disposition: disp,
				Metadata: map[string]any{}, CreatedAt: now,
			})
		}
	}
	ret := domain.Return{
		ID: retID, OrderID: o.ID, TenantID: o.TenantID,
		Status: domain.ReturnStatusRequested, Disposition: disp,
		Reason: in.Reason, Notes: in.Notes, ActorID: in.ActorID,
		Lines: lines, Metadata: map[string]any{},
		RequestedAt: now, CreatedAt: now, UpdatedAt: now,
	}
	if err := ret.Validate(); err != nil {
		return domain.Return{}, err
	}
	if d.Returns == nil {
		return domain.Return{}, fmt.Errorf("%w: returns repository not configured", domain.ErrInvariant)
	}
	if err := d.Returns.Create(ctx, ret); err != nil {
		return domain.Return{}, err
	}
	_ = d.appendEvent(ctx, o.ID, o.TenantID, domain.EventReturnRequested, map[string]any{
		"returnId": ret.ID.String(),
		"reason":   in.Reason,
	})
	return ret, nil
}

// RequestRefundInput opens a refund request.
type RequestRefundInput struct {
	TenantID    uuid.UUID
	OrderID     uuid.UUID
	AmountMinor int64
	Currency    string
	Method      domain.RefundMethod
	Reason      string
	ReturnID    *uuid.UUID
	ActorID     *uuid.UUID
}

// RequestRefund creates a refund request and calls the payment refund port when possible.
func (d *Deps) RequestRefund(ctx context.Context, in RequestRefundInput) (domain.Refund, error) {
	o, err := d.Orders.GetByID(ctx, in.TenantID, in.OrderID)
	if err != nil {
		return domain.Refund{}, err
	}
	if err := domain.AssertRefundAllowed(o.Status); err != nil {
		return domain.Refund{}, err
	}
	amount := in.AmountMinor
	if amount <= 0 {
		amount = o.TotalMinor
	}
	currency := in.Currency
	if currency == "" {
		currency = o.Currency
	}
	method := in.Method
	if method == "" {
		method = domain.RefundMethodCard
	}
	now := d.now()
	ref := domain.Refund{
		ID: d.newID(), OrderID: o.ID, TenantID: o.TenantID, ReturnID: in.ReturnID,
		AmountMinor: amount, Currency: currency, Method: method,
		Status: domain.RefundStatusPending, Reason: in.Reason, ActorID: in.ActorID,
		Metadata: map[string]any{}, RequestedAt: now, CreatedAt: now, UpdatedAt: now,
	}
	if err := ref.Validate(); err != nil {
		return domain.Refund{}, err
	}
	if d.Refunds == nil {
		return domain.Refund{}, fmt.Errorf("%w: refunds repository not configured", domain.ErrInvariant)
	}
	if err := d.Refunds.Create(ctx, ref); err != nil {
		return domain.Refund{}, err
	}

	if o.CanTransitionTo(domain.OrderStatusRefundPending) {
		_ = d.transition(&o, domain.OrderStatusRefundPending)
	}
	_ = d.appendEvent(ctx, o.ID, o.TenantID, domain.EventRefundRequested, map[string]any{
		"refundId":    ref.ID.String(),
		"amountMinor": amount,
	})

	if o.PaymentIntentRef != "" && d.Payment != nil {
		res, err := d.Payment.Refund(ctx, ports.RefundPaymentRequest{
			TenantID: o.TenantID, PaymentIntentRef: o.PaymentIntentRef,
			AmountMinor: amount, Currency: currency,
			IdempotencyKey: "refund:" + ref.ID.String(),
		})
		if err == nil {
			ref.PaymentRefundRef = res.PaymentRefundRef
			_ = ref.TransitionTo(domain.RefundStatusSucceeded)
			_ = d.Refunds.Update(ctx, ref)
			if o.CanTransitionTo(domain.OrderStatusRefunded) {
				_ = d.transition(&o, domain.OrderStatusRefunded)
			}
			_ = d.appendEvent(ctx, o.ID, o.TenantID, domain.EventRefundCompleted, map[string]any{
				"refundId":         ref.ID.String(),
				"paymentRefundRef": res.PaymentRefundRef,
			})
		}
	}
	_ = d.Orders.Update(ctx, o)
	d.indexOrder(ctx, o)
	return ref, nil
}

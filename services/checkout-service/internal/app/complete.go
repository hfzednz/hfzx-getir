package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/checkout-service/internal/app/ports"
	"github.com/nexora/checkout-service/internal/domain"
)

// CompleteInput confirms checkout and creates an order via OrderClient.
type CompleteInput struct {
	TenantID       uuid.UUID
	SessionID      uuid.UUID
	IdempotencyKey string
	PlaceOrder     bool
}

// Complete is idempotent: ready → completing → completed via OrderClient.CreateFromCheckout.
// Does not run OMS place saga unless PlaceOrder/AutoPlaceOrder is set.
func (d *Deps) Complete(ctx context.Context, in CompleteInput) (domain.Session, error) {
	key := strings.TrimSpace(in.IdempotencyKey)
	var out domain.Session
	err := d.withCompleteLock(ctx, completeLockKey(in.TenantID, in.SessionID, key), func() error {
		s, err := d.Sessions.GetByID(ctx, in.TenantID, in.SessionID)
		if err != nil {
			return err
		}
		// Idempotent: already completed with order id.
		if s.Status == domain.StatusCompleted && s.OrderID != "" {
			out = s
			return nil
		}
		if s.Status == domain.StatusBlocked {
			return fmt.Errorf("%w: cannot complete blocked session", domain.ErrInvalidTransition)
		}
		if s.Status != domain.StatusReady && s.Status != domain.StatusCompleting {
			return fmt.Errorf("%w: cannot complete from status %s", domain.ErrInvalidTransition, s.Status)
		}
		if !s.Validation.Passed {
			return fmt.Errorf("%w: validation not passed", domain.ErrNotReady)
		}

		idem := key
		if idem == "" {
			idem = s.IdempotencyKey
		}
		if idem == "" {
			return fmt.Errorf("%w: idempotency_key required", domain.ErrInvalidArgument)
		}

		if s.Status == domain.StatusReady {
			if err := d.transition(&s, domain.StatusCompleting); err != nil {
				return err
			}
			if err := d.Sessions.Update(ctx, s); err != nil {
				return err
			}
		}

		cart, err := d.Cart.GetCart(ctx, s.TenantID, s.CartID)
		if err != nil {
			_ = d.failSession(ctx, &s, err.Error())
			return err
		}

		lines := make([]ports.CreateFromCheckoutLine, 0, len(cart.Lines))
		for _, l := range cart.Lines {
			lines = append(lines, ports.CreateFromCheckoutLine{
				VariantID: l.VariantID, SKUCode: l.SKUCode, TitleSnapshot: l.TitleSnapshot,
				Qty: l.Qty, UnitPriceMinor: l.UnitPriceMinor, WarehouseID: l.WarehouseID,
			})
		}

		addr := map[string]any{
			"label": s.Address.Label, "line1": s.Address.Line1, "line2": s.Address.Line2,
			"city": s.Address.City, "district": s.Address.District, "postalCode": s.Address.PostalCode,
			"country": s.Address.Country, "lat": s.Address.Lat, "lng": s.Address.Lng,
			"contactName": s.Address.ContactName, "phone": s.Address.Phone,
		}
		gift := map[string]any{
			"enabled": s.Gift.Enabled, "message": s.Gift.Message, "from": s.Gift.From,
		}

		req := ports.CreateFromCheckoutRequest{
			TenantID:            s.TenantID,
			CustomerPrincipalID: s.PrincipalID,
			CheckoutSessionID:   s.ID,
			CartID:              s.CartID,
			Currency:            s.Currency,
			IdempotencyKey:      idem,
			AddressSnapshot:     addr,
			Notes:               s.Notes,
			Gift:                gift,
			DeliveryOption:      s.DeliveryOption,
			ScheduledAt:         s.Slot.StartsAt,
			TipMinor:            s.TipMinor,
			DiscountMinor:       s.Quote.DiscountMinor,
			ShippingMinor:       s.Quote.DeliveryMinor,
			TaxMinor:            s.Quote.TaxMinor,
			SubtotalMinor:       s.Quote.SubtotalMinor,
			TotalMinor:          s.Quote.TotalMinor,
			Lines:               lines,
			Metadata: map[string]any{
				"checkoutSessionId": s.ID.String(),
				"quoteId":           s.Quote.QuoteID,
			},
		}

		if d.Orders == nil {
			_ = d.failSession(ctx, &s, "order client not configured")
			return fmt.Errorf("%w: order client not configured", domain.ErrInvariant)
		}
		res, err := d.Orders.CreateFromCheckout(ctx, req)
		if err != nil {
			_ = d.failSession(ctx, &s, err.Error())
			return err
		}
		s.OrderID = res.OrderID
		if s.OrderID == "" {
			_ = d.failSession(ctx, &s, "empty order id")
			return fmt.Errorf("%w: empty order id from order-service", domain.ErrInvariant)
		}

		if in.PlaceOrder || d.AutoPlaceOrder {
			_, _ = d.Orders.PlaceOrder(ctx, ports.PlaceOrderRequest{
				TenantID: s.TenantID, OrderID: s.OrderID, IdempotencyKey: idem + ":place",
			})
		}

		if err := d.transition(&s, domain.StatusCompleted); err != nil {
			return err
		}
		if err := d.Sessions.Update(ctx, s); err != nil {
			return err
		}
		_ = d.appendEvent(ctx, s.ID, s.TenantID, domain.EventCheckoutCompleted, map[string]any{
			"orderId": s.OrderID,
			"status":  string(s.Status),
		})
		out = s
		return nil
	})
	return out, err
}

func completeLockKey(tenantID, sessionID uuid.UUID, idem string) string {
	if idem != "" {
		return tenantID.String() + ":" + idem
	}
	return tenantID.String() + ":session:" + sessionID.String()
}

func (d *Deps) failSession(ctx context.Context, s *domain.Session, reason string) error {
	if s.Status == domain.StatusCompleted || s.Status == domain.StatusFailed {
		return nil
	}
	if s.Status == domain.StatusReady {
		_ = d.transition(s, domain.StatusCompleting)
	}
	if err := domain.ValidateTransition(s.Status, domain.StatusFailed); err == nil {
		_ = d.transition(s, domain.StatusFailed)
	}
	s.UpdatedAt = d.now()
	if s.Metadata == nil {
		s.Metadata = map[string]any{}
	}
	s.Metadata["failReason"] = reason
	_ = d.Sessions.Update(ctx, *s)
	_ = d.appendEvent(ctx, s.ID, s.TenantID, domain.EventCheckoutFailed, map[string]any{
		"reason": reason,
	})
	return nil
}

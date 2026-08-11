package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/cart-service/internal/domain"
)

// MarkAbandonedInput marks a cart as abandoned.
type MarkAbandonedInput struct {
	TenantID uuid.UUID
	CartID   uuid.UUID
}

// MarkAbandoned sets status to abandoned.
func (d *Deps) MarkAbandoned(ctx context.Context, in MarkAbandonedInput) (domain.Cart, error) {
	c, err := d.loadCart(ctx, in.TenantID, in.CartID)
	if err != nil {
		return domain.Cart{}, err
	}
	if c.Status == domain.CartStatusAbandoned {
		return c, nil
	}
	if c.Status != domain.CartStatusActive {
		return domain.Cart{}, fmt.Errorf("%w: cannot abandon from %s", domain.ErrInvalidTransition, c.Status)
	}
	now := d.now()
	c.Status = domain.CartStatusAbandoned
	c.AbandonedAt = &now
	c.UpdatedAt = now
	c.Version++
	if err := d.Carts.Update(ctx, c); err != nil {
		return domain.Cart{}, err
	}
	_ = d.appendEvent(ctx, c.ID, c.TenantID, domain.EventCartAbandoned, map[string]any{})
	return c, nil
}

// RecoverInput recovers an abandoned cart back to active.
type RecoverInput struct {
	TenantID uuid.UUID
	CartID   uuid.UUID
}

// Recover restores an abandoned cart to active.
func (d *Deps) Recover(ctx context.Context, in RecoverInput) (domain.Cart, error) {
	c, err := d.loadCart(ctx, in.TenantID, in.CartID)
	if err != nil {
		return domain.Cart{}, err
	}
	if c.Status == domain.CartStatusActive {
		return c, nil
	}
	if c.Status != domain.CartStatusAbandoned {
		return domain.Cart{}, fmt.Errorf("%w: cannot recover from %s", domain.ErrInvalidTransition, c.Status)
	}
	now := d.now()
	c.Status = domain.CartStatusActive
	c.AbandonedAt = nil
	c.UpdatedAt = now
	c.Version++
	if err := d.Carts.Update(ctx, c); err != nil {
		return domain.Cart{}, err
	}
	_ = d.appendEvent(ctx, c.ID, c.TenantID, domain.EventCartRecovered, map[string]any{})
	return c, nil
}

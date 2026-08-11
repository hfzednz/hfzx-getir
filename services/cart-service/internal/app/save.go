package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/cart-service/internal/app/ports"
	"github.com/nexora/cart-service/internal/domain"
)

// RecommendationsInput fetches cart recommendations.
type RecommendationsInput struct {
	TenantID uuid.UUID
	CartID   uuid.UUID
	Limit    int
}

// Recommendations returns suggested variants via RecommendClient.
func (d *Deps) Recommendations(ctx context.Context, in RecommendationsInput) (ports.RecommendResult, error) {
	c, err := d.loadCart(ctx, in.TenantID, in.CartID)
	if err != nil {
		return ports.RecommendResult{}, err
	}
	if d.Recommend == nil {
		return ports.RecommendResult{Items: nil}, nil
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 8
	}
	return d.Recommend.Recommend(ctx, ports.RecommendRequest{
		TenantID: c.TenantID, CartID: c.ID, CityID: c.CityID, Limit: limit,
	})
}

// SaveCartInput saves a cart snapshot for later.
type SaveCartInput struct {
	TenantID    uuid.UUID
	CartID      uuid.UUID
	PrincipalID uuid.UUID
	Name        string
}

// SaveCart persists a named save-for-later snapshot.
func (d *Deps) SaveCart(ctx context.Context, in SaveCartInput) (domain.SavedCart, error) {
	if in.PrincipalID == uuid.Nil {
		return domain.SavedCart{}, fmt.Errorf("%w: principal required", domain.ErrUnauthorized)
	}
	c, err := d.loadCart(ctx, in.TenantID, in.CartID)
	if err != nil {
		return domain.SavedCart{}, err
	}
	if d.Saved == nil {
		return domain.SavedCart{}, fmt.Errorf("%w: saved cart repo required", domain.ErrInvariant)
	}
	now := d.now()
	lines := make([]map[string]any, 0, len(c.Lines))
	for _, l := range c.Lines {
		lines = append(lines, map[string]any{
			"variantId": l.VariantID.String(),
			"qty":       l.Qty,
			"maxQty":    l.MaxQty,
			"notes":     l.Notes,
		})
	}
	src := c.ID
	s := domain.SavedCart{
		ID:           d.newID(),
		TenantID:     c.TenantID,
		PrincipalID:  in.PrincipalID,
		SourceCartID: &src,
		Name:         in.Name,
		Snapshot: map[string]any{
			"currency": c.Currency,
			"lines":    lines,
			"coupons":  c.CouponCodes(),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := d.Saved.Create(ctx, s); err != nil {
		return domain.SavedCart{}, err
	}
	_ = d.appendEvent(ctx, c.ID, c.TenantID, domain.EventCartSaved, map[string]any{
		"savedCartId": s.ID.String(),
		"name":        s.Name,
	})
	return s, nil
}

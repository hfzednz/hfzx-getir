package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/checkout-service/internal/domain"
)

// StartFromCartInput creates a checkout session from a cart.
type StartFromCartInput struct {
	TenantID       uuid.UUID
	CartID         uuid.UUID
	PrincipalID    uuid.UUID
	IdempotencyKey string
	DeliveryOption domain.DeliveryOption
	Currency       string
}

// StartFromCart creates a session in started status (idempotent by key).
func (d *Deps) StartFromCart(ctx context.Context, in StartFromCartInput) (domain.Session, error) {
	if in.TenantID == uuid.Nil {
		return domain.Session{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	if in.CartID == uuid.Nil {
		return domain.Session{}, fmt.Errorf("%w: cart_id required", domain.ErrInvalidArgument)
	}
	if in.PrincipalID == uuid.Nil {
		return domain.Session{}, fmt.Errorf("%w: principal_id required", domain.ErrInvalidArgument)
	}
	key := strings.TrimSpace(in.IdempotencyKey)
	if key == "" {
		return domain.Session{}, fmt.Errorf("%w: idempotency_key required", domain.ErrInvalidArgument)
	}
	if existing, err := d.Sessions.GetByIdempotencyKey(ctx, in.TenantID, key); err == nil {
		return existing, nil
	}

	cart, err := d.Cart.GetCart(ctx, in.TenantID, in.CartID)
	if err != nil {
		return domain.Session{}, err
	}
	if !cart.Active {
		return domain.Session{}, fmt.Errorf("%w: cart inactive", domain.ErrInvariant)
	}
	if len(cart.Lines) == 0 {
		return domain.Session{}, fmt.Errorf("%w: cart has no lines", domain.ErrInvalidArgument)
	}
	if cart.PrincipalID != uuid.Nil && cart.PrincipalID != in.PrincipalID {
		return domain.Session{}, fmt.Errorf("%w: cart principal mismatch", domain.ErrForbidden)
	}

	opt := in.DeliveryOption
	if opt == "" {
		opt = domain.DeliveryInstant
	}
	if !opt.Valid() {
		return domain.Session{}, fmt.Errorf("%w: invalid delivery option", domain.ErrInvalidArgument)
	}

	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if currency == "" {
		currency = strings.ToUpper(strings.TrimSpace(cart.Currency))
	}
	if currency == "" {
		currency = "TRY"
	}
	if _, err := domain.NewMoney(0, currency); err != nil {
		return domain.Session{}, err
	}

	now := d.now()
	s := domain.Session{
		ID:             d.newID(),
		TenantID:       in.TenantID,
		CartID:         in.CartID,
		PrincipalID:    in.PrincipalID,
		Status:         domain.StatusStarted,
		DeliveryOption: opt,
		Substitutions:  domain.SubstitutionAsk,
		Currency:       currency,
		IdempotencyKey: key,
		RecoveryToken:  d.newID().String(),
		CityID:         cart.CityID,
		CouponCodes:    append([]string(nil), cart.CouponCodes...),
		Version:        1,
		Metadata:       map[string]any{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.Validate(); err != nil {
		return domain.Session{}, err
	}
	if err := d.Sessions.Create(ctx, s); err != nil {
		if existing, gerr := d.Sessions.GetByIdempotencyKey(ctx, in.TenantID, key); gerr == nil {
			return existing, nil
		}
		return domain.Session{}, err
	}
	_ = d.appendEvent(ctx, s.ID, s.TenantID, domain.EventCheckoutStarted, map[string]any{
		"cartId":  s.CartID.String(),
		"status":  string(s.Status),
		"cityId":  s.CityID,
	})
	return s, nil
}

// GetSession returns a session by id.
func (d *Deps) GetSession(ctx context.Context, tenantID, sessionID uuid.UUID) (domain.Session, error) {
	return d.Sessions.GetByID(ctx, tenantID, sessionID)
}

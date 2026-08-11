package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/checkout-service/internal/domain"
)

// PatchInput updates checkout preferences before complete.
type PatchInput struct {
	TenantID       uuid.UUID
	SessionID      uuid.UUID
	Address        *domain.AddressSnapshot
	Slot           *domain.SlotSnapshot
	Gift           *domain.GiftPrefs
	Invoice        *domain.InvoicePrefs
	Notes          *string
	Substitutions  *domain.SubstitutionPolicy
	TipMinor       *int64
	DeliveryOption *domain.DeliveryOption
	CouponCodes    []string
	ClearCoupons   bool
}

// Patch applies preference updates. Moves ready/blocked back toward needing re-validation
// by clearing ready status (sets to started if was ready/blocked).
func (d *Deps) Patch(ctx context.Context, in PatchInput) (domain.Session, error) {
	s, err := d.Sessions.GetByID(ctx, in.TenantID, in.SessionID)
	if err != nil {
		return domain.Session{}, err
	}
	if !s.Status.CanPatch() {
		return domain.Session{}, fmt.Errorf("%w: cannot patch in status %s", domain.ErrInvalidTransition, s.Status)
	}

	if in.Address != nil {
		s.Address = *in.Address
	}
	if in.Slot != nil {
		s.Slot = *in.Slot
	}
	if in.Gift != nil {
		s.Gift = *in.Gift
	}
	if in.Invoice != nil {
		s.Invoice = *in.Invoice
	}
	if in.Notes != nil {
		s.Notes = *in.Notes
	}
	if in.Substitutions != nil {
		if !in.Substitutions.Valid() {
			return domain.Session{}, fmt.Errorf("%w: invalid substitutions", domain.ErrInvalidArgument)
		}
		s.Substitutions = *in.Substitutions
	}
	if in.TipMinor != nil {
		if *in.TipMinor < 0 {
			return domain.Session{}, fmt.Errorf("%w: tip", domain.ErrNegativeMoney)
		}
		s.TipMinor = *in.TipMinor
	}
	if in.DeliveryOption != nil {
		if !in.DeliveryOption.Valid() {
			return domain.Session{}, fmt.Errorf("%w: invalid delivery option", domain.ErrInvalidArgument)
		}
		s.DeliveryOption = *in.DeliveryOption
	}
	if in.ClearCoupons {
		s.CouponCodes = nil
	} else if in.CouponCodes != nil {
		s.CouponCodes = append([]string(nil), in.CouponCodes...)
	}

	// Preference changes invalidate prior ready/blocked validation.
	if s.Status == domain.StatusReady || s.Status == domain.StatusBlocked || s.Status == domain.StatusValidating {
		s.Status = domain.StatusStarted
		s.Validation = domain.ValidationResults{}
		s.Version++
	}
	s.UpdatedAt = d.now()
	if err := s.Validate(); err != nil {
		return domain.Session{}, err
	}
	if err := d.Sessions.Update(ctx, s); err != nil {
		return domain.Session{}, err
	}
	return s, nil
}

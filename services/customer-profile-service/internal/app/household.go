package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/domain"
)

// CreateHouseholdInput creates a household owned by a profile.
type CreateHouseholdInput struct {
	OwnerProfileID uuid.UUID
	Name           string
	Sharing        domain.HouseholdSharingFlags
}

// AddHouseholdMemberInput adds a member to a household.
type AddHouseholdMemberInput struct {
	HouseholdID uuid.UUID
	ProfileID   uuid.UUID
	Role        domain.HouseholdMemberRole
}

// UpdateSharingInput updates household-level sharing flags.
type UpdateSharingInput struct {
	HouseholdID uuid.UUID
	Sharing     domain.HouseholdSharingFlags
}

// CreateHousehold creates a household with the owner as first member.
func (d *Deps) CreateHousehold(ctx context.Context, in CreateHouseholdInput) (domain.Household, error) {
	owner, err := d.requireActiveProfile(ctx, in.OwnerProfileID)
	if err != nil {
		return domain.Household{}, err
	}
	now := d.now()
	h := domain.Household{
		ID:             d.newID(),
		TenantID:       owner.TenantID,
		Name:           in.Name,
		OwnerProfileID: in.OwnerProfileID,
		Sharing:        in.Sharing,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if h.Name == "" {
		h.Name = "Household"
	}
	if err := h.Validate(); err != nil {
		return domain.Household{}, err
	}
	if err := d.Households.Create(ctx, h); err != nil {
		return domain.Household{}, err
	}
	member := domain.HouseholdMember{
		ID:          d.newID(),
		HouseholdID: h.ID,
		ProfileID:   in.OwnerProfileID,
		Role:        domain.HouseholdRoleOwner,
		JoinedAt:    now,
	}
	if err := d.Households.AddMember(ctx, member); err != nil {
		return domain.Household{}, err
	}
	return h, nil
}

// AddHouseholdMember adds a profile to a household.
func (d *Deps) AddHouseholdMember(ctx context.Context, in AddHouseholdMemberInput) (domain.HouseholdMember, error) {
	h, err := d.Households.Get(ctx, in.HouseholdID)
	if err != nil {
		return domain.HouseholdMember{}, err
	}
	if _, err := d.requireActiveProfile(ctx, in.ProfileID); err != nil {
		return domain.HouseholdMember{}, err
	}
	role := in.Role
	if role == "" {
		role = domain.HouseholdRoleAdult
	}
	m := domain.HouseholdMember{
		ID:          d.newID(),
		HouseholdID: h.ID,
		ProfileID:   in.ProfileID,
		Role:        role,
		JoinedAt:    d.now(),
	}
	if err := m.Validate(); err != nil {
		return domain.HouseholdMember{}, err
	}
	if err := d.Households.AddMember(ctx, m); err != nil {
		return domain.HouseholdMember{}, err
	}
	return m, nil
}

// UpdateSharing updates household sharing flags.
func (d *Deps) UpdateSharing(ctx context.Context, in UpdateSharingInput) (domain.Household, error) {
	h, err := d.Households.Get(ctx, in.HouseholdID)
	if err != nil {
		return domain.Household{}, err
	}
	h.Sharing = in.Sharing
	h.UpdatedAt = d.now()
	if err := h.Validate(); err != nil {
		return domain.Household{}, err
	}
	if err := d.Households.Update(ctx, h); err != nil {
		return domain.Household{}, err
	}
	return h, nil
}

// GetHouseholdByOwner returns the household owned by a profile.
func (d *Deps) GetHouseholdByOwner(ctx context.Context, ownerProfileID uuid.UUID) (domain.Household, error) {
	if ownerProfileID == uuid.Nil {
		return domain.Household{}, fmt.Errorf("%w: owner profile_id required", domain.ErrInvalidArgument)
	}
	return d.Households.GetByOwner(ctx, ownerProfileID)
}

// ListHouseholdMembers returns members of a household.
func (d *Deps) ListHouseholdMembers(ctx context.Context, householdID uuid.UUID) ([]domain.HouseholdMember, error) {
	if householdID == uuid.Nil {
		return nil, fmt.Errorf("%w: household_id required", domain.ErrInvalidArgument)
	}
	return d.Households.ListMembers(ctx, householdID)
}

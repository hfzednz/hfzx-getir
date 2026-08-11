package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/domain"
)

// ProvisionInput creates a profile after IAM principal provisioning.
type ProvisionInput struct {
	TenantID    uuid.UUID
	PrincipalID uuid.UUID
	DisplayName string
	FullName    string
	Nickname    string
	Language    string
	CountryCode string
	City        string
	Timezone    string
	TraceID     string
}

// UpdateProfileInput patches mutable profile fields.
type UpdateProfileInput struct {
	ProfileID     uuid.UUID
	DisplayName   *string
	FullName      *string
	Nickname      *string
	Gender        *domain.Gender
	Language      *string
	CountryCode   *string
	City          *string
	Timezone      *string
	Occupation    *string
	FamilySize    *int
	Dietary       map[string]any
	Accessibility map[string]any
	TraceID       string
}

// Provision creates a CustomerProfile after IAM registration.
func (d *Deps) Provision(ctx context.Context, in ProvisionInput) (domain.CustomerProfile, error) {
	if in.TenantID == uuid.Nil || in.PrincipalID == uuid.Nil {
		return domain.CustomerProfile{}, fmt.Errorf("%w: tenant_id and principal_id required", domain.ErrInvalidArgument)
	}
	if existing, err := d.Profiles.GetByPrincipalID(ctx, in.TenantID, in.PrincipalID); err == nil {
		return existing, nil
	} else if err != domain.ErrNotFound {
		return domain.CustomerProfile{}, err
	}

	now := d.now()
	p := domain.CustomerProfile{
		ID:          d.newID(),
		PrincipalID: in.PrincipalID,
		TenantID:    in.TenantID,
		DisplayName: in.DisplayName,
		FullName:    in.FullName,
		Nickname:    in.Nickname,
		Gender:      domain.GenderUnspecified,
		Language:    in.Language,
		CountryCode: in.CountryCode,
		City:        in.City,
		Timezone:    in.Timezone,
		Status:      domain.ProfileStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := p.Validate(); err != nil {
		return domain.CustomerProfile{}, err
	}
	if err := d.Profiles.Create(ctx, p); err != nil {
		return domain.CustomerProfile{}, err
	}
	d.publishLifecycle(ctx, domain.EventCustomerCreated, p, nil, in.TraceID)
	d.indexProfile(ctx, p)
	return p, nil
}

// GetProfile returns a profile by id.
func (d *Deps) GetProfile(ctx context.Context, profileID uuid.UUID) (domain.CustomerProfile, error) {
	if profileID == uuid.Nil {
		return domain.CustomerProfile{}, fmt.Errorf("%w: profile_id required", domain.ErrInvalidArgument)
	}
	return d.Profiles.GetByID(ctx, profileID)
}

// GetProfileByPrincipal returns a profile by tenant + principal (gateway /me).
func (d *Deps) GetProfileByPrincipal(ctx context.Context, tenantID, principalID uuid.UUID) (domain.CustomerProfile, error) {
	if tenantID == uuid.Nil || principalID == uuid.Nil {
		return domain.CustomerProfile{}, fmt.Errorf("%w: tenant_id and principal_id required", domain.ErrInvalidArgument)
	}
	return d.Profiles.GetByPrincipalID(ctx, tenantID, principalID)
}

// UpdateProfile patches profile attributes and emits CustomerUpdated.
func (d *Deps) UpdateProfile(ctx context.Context, in UpdateProfileInput) (domain.CustomerProfile, error) {
	p, err := d.requireActiveProfile(ctx, in.ProfileID)
	if err != nil {
		return domain.CustomerProfile{}, err
	}
	if in.DisplayName != nil {
		p.DisplayName = *in.DisplayName
	}
	if in.FullName != nil {
		p.FullName = *in.FullName
	}
	if in.Nickname != nil {
		p.Nickname = *in.Nickname
	}
	if in.Gender != nil {
		p.Gender = *in.Gender
	}
	if in.Language != nil {
		p.Language = *in.Language
	}
	if in.CountryCode != nil {
		p.CountryCode = *in.CountryCode
	}
	if in.City != nil {
		p.City = *in.City
	}
	if in.Timezone != nil {
		p.Timezone = *in.Timezone
	}
	if in.Occupation != nil {
		p.Occupation = *in.Occupation
	}
	if in.FamilySize != nil {
		p.FamilySize = *in.FamilySize
	}
	if in.Dietary != nil {
		p.Dietary = in.Dietary
	}
	if in.Accessibility != nil {
		p.Accessibility = in.Accessibility
	}
	p.UpdatedAt = d.now()
	if err := p.Validate(); err != nil {
		return domain.CustomerProfile{}, err
	}
	if err := d.Profiles.Update(ctx, p); err != nil {
		return domain.CustomerProfile{}, err
	}
	d.publishLifecycle(ctx, domain.EventCustomerUpdated, p, nil, in.TraceID)
	d.indexProfile(ctx, p)
	return p, nil
}

// SoftDeleteProfile marks a profile deleted and emits ProfileDeleted.
func (d *Deps) SoftDeleteProfile(ctx context.Context, profileID uuid.UUID, traceID string) error {
	p, err := d.requireActiveProfile(ctx, profileID)
	if err != nil {
		return err
	}
	now := d.now()
	if err := d.Profiles.SoftDelete(ctx, profileID, now); err != nil {
		return err
	}
	p.Status = domain.ProfileStatusDeleted
	p.DeletedAt = &now
	p.UpdatedAt = now
	d.publishLifecycle(ctx, domain.EventProfileDeleted, p, nil, traceID)
	d.deleteProfileIndex(ctx, p.TenantID, profileID)
	return nil
}

package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/domain"
)

// SearchCustomers searches profiles within a tenant.
func (d *Deps) SearchCustomers(ctx context.Context, tenantID uuid.UUID, query string, limit int) ([]domain.CustomerProfile, error) {
	if tenantID == uuid.Nil {
		return nil, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	if limit <= 0 {
		limit = 20
	}
	return d.Profiles.Search(ctx, tenantID, query, limit)
}

// MergeCustomers merges source into target and marks source as merged (soft-delete semantics).
func (d *Deps) MergeCustomers(ctx context.Context, targetID, sourceID uuid.UUID, traceID string) (domain.CustomerProfile, error) {
	if targetID == uuid.Nil || sourceID == uuid.Nil {
		return domain.CustomerProfile{}, fmt.Errorf("%w: target and source required", domain.ErrInvalidArgument)
	}
	if targetID == sourceID {
		return domain.CustomerProfile{}, fmt.Errorf("%w: cannot merge profile into itself", domain.ErrInvalidArgument)
	}
	target, err := d.requireActiveProfile(ctx, targetID)
	if err != nil {
		return domain.CustomerProfile{}, err
	}
	source, err := d.requireActiveProfile(ctx, sourceID)
	if err != nil {
		return domain.CustomerProfile{}, err
	}
	if target.TenantID != source.TenantID {
		return domain.CustomerProfile{}, fmt.Errorf("%w: tenant mismatch", domain.ErrForbidden)
	}

	addrs, _ := d.Addresses.ListByProfile(ctx, sourceID)
	for _, a := range addrs {
		if a.DeletedAt != nil {
			continue
		}
		a.ID = d.newID()
		a.ProfileID = targetID
		a.IsDefault = false
		a.UpdatedAt = d.now()
		_ = d.Addresses.Create(ctx, a)
	}

	tags, _ := d.Tags.List(ctx, sourceID)
	for _, t := range tags {
		t.ProfileID = targetID
		t.AssignedAt = d.now()
		_ = d.Tags.Add(ctx, t)
	}

	now := d.now()
	source.Status = domain.ProfileStatusMerged
	source.UpdatedAt = now
	// Merged profiles must not carry DeletedAt (invariant); soft-delete semantics via status.
	source.DeletedAt = nil
	if err := source.Validate(); err != nil {
		return domain.CustomerProfile{}, err
	}
	if err := d.Profiles.Update(ctx, source); err != nil {
		return domain.CustomerProfile{}, err
	}

	d.publishLifecycle(ctx, domain.EventProfileDeleted, source, map[string]any{
		"mergedInto": targetID,
		"reason":     "merge",
	}, traceID)
	d.deleteProfileIndex(ctx, source.TenantID, sourceID)

	target.UpdatedAt = now
	_ = d.Profiles.Update(ctx, target)
	d.indexProfile(ctx, target)
	return target, nil
}

// FindDuplicates finds profiles sharing display/full name within a tenant.
func (d *Deps) FindDuplicates(ctx context.Context, tenantID uuid.UUID, displayName, fullName string) ([]domain.CustomerProfile, error) {
	if tenantID == uuid.Nil {
		return nil, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	displayName = strings.TrimSpace(displayName)
	fullName = strings.TrimSpace(fullName)
	if displayName == "" && fullName == "" {
		return nil, fmt.Errorf("%w: display_name or full_name required", domain.ErrInvalidArgument)
	}
	return d.Profiles.FindDuplicates(ctx, tenantID, displayName, fullName)
}

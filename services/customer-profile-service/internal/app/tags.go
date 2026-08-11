package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/domain"
)

// AddTagInput assigns a tag to a profile (creates a custom tag definition when TagID is nil).
type AddTagInput struct {
	ProfileID  uuid.UUID
	TagID      uuid.UUID
	Kind       domain.TagKind
	Name       string
	AssignedBy *uuid.UUID
	Note       string
}

// AddTag attaches a tag to a profile.
func (d *Deps) AddTag(ctx context.Context, in AddTagInput) (domain.ProfileTag, error) {
	p, err := d.requireActiveProfile(ctx, in.ProfileID)
	if err != nil {
		return domain.ProfileTag{}, err
	}
	tagID := in.TagID
	if tagID == uuid.Nil {
		kind := in.Kind
		if kind == "" {
			kind = domain.TagKindCustom
		}
		name := in.Name
		if name == "" {
			name = string(kind)
		}
		now := d.now()
		t := domain.Tag{
			ID:        d.newID(),
			TenantID:  p.TenantID,
			Kind:      kind,
			Name:      name,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := t.Validate(); err != nil {
			return domain.ProfileTag{}, err
		}
		if err := d.Tags.UpsertTag(ctx, t); err != nil {
			return domain.ProfileTag{}, err
		}
		tagID = t.ID
	} else if _, err := d.Tags.GetTag(ctx, tagID); err != nil {
		return domain.ProfileTag{}, err
	}

	pt := domain.ProfileTag{
		ProfileID:  in.ProfileID,
		TagID:      tagID,
		AssignedBy: in.AssignedBy,
		AssignedAt: d.now(),
		Note:       in.Note,
	}
	if err := pt.Validate(); err != nil {
		return domain.ProfileTag{}, err
	}
	if err := d.Tags.Add(ctx, pt); err != nil {
		return domain.ProfileTag{}, err
	}
	return pt, nil
}

// RemoveTag removes a tag assignment.
func (d *Deps) RemoveTag(ctx context.Context, profileID, tagID uuid.UUID) error {
	if _, err := d.requireActiveProfile(ctx, profileID); err != nil {
		return err
	}
	if tagID == uuid.Nil {
		return fmt.Errorf("%w: tag_id required", domain.ErrInvalidArgument)
	}
	return d.Tags.Remove(ctx, profileID, tagID)
}

// ListTags returns tag assignments for a profile.
func (d *Deps) ListTags(ctx context.Context, profileID uuid.UUID) ([]domain.ProfileTag, error) {
	if _, err := d.requireActiveProfile(ctx, profileID); err != nil {
		return nil, err
	}
	return d.Tags.List(ctx, profileID)
}

package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/app/ports"
	"github.com/nexora/customer-profile-service/internal/domain"
)

// AssignSegmentInput assigns a profile to a segment.
type AssignSegmentInput struct {
	SegmentID  uuid.UUID
	ProfileID  uuid.UUID
	Source     string
	TraceID    string
}

// EvaluateDynamic evaluates a dynamic segment rule against a profile and assigns if matched.
func (d *Deps) EvaluateDynamic(ctx context.Context, segmentID, profileID uuid.UUID, traceID string) (bool, error) {
	seg, err := d.Segments.GetSegment(ctx, segmentID)
	if err != nil {
		return false, err
	}
	p, err := d.requireActiveProfile(ctx, profileID)
	if err != nil {
		return false, err
	}
	if seg.TenantID != p.TenantID {
		return false, fmt.Errorf("%w: tenant mismatch", domain.ErrForbidden)
	}
	matched := matchDynamicRule(ctx, d, seg, profileID)
	if !matched {
		_ = d.Segments.RemoveMembership(ctx, segmentID, profileID)
		return false, nil
	}
	_, err = d.AssignSegment(ctx, AssignSegmentInput{
		SegmentID: segmentID,
		ProfileID: profileID,
		Source:    "dynamic",
		TraceID:   traceID,
	})
	return err == nil, err
}

func matchDynamicRule(ctx context.Context, d *Deps, seg domain.Segment, profileID uuid.UUID) bool {
	if seg.Kind != domain.SegmentKindDynamic || seg.Rules == nil {
		return false
	}
	if tagKind, ok := seg.Rules["tagKind"].(string); ok && tagKind != "" {
		pts, err := d.Tags.List(ctx, profileID)
		if err != nil {
			return false
		}
		for _, pt := range pts {
			t, err := d.Tags.GetTag(ctx, pt.TagID)
			if err != nil {
				continue
			}
			if string(t.Kind) == tagKind {
				return true
			}
		}
		return false
	}
	if city, ok := seg.Rules["city"].(string); ok && city != "" {
		p, err := d.Profiles.GetByID(ctx, profileID)
		if err != nil {
			return false
		}
		return p.City == city
	}
	return false
}

// AssignSegment assigns membership and emits SegmentChanged.
func (d *Deps) AssignSegment(ctx context.Context, in AssignSegmentInput) (domain.SegmentMembership, error) {
	seg, err := d.Segments.GetSegment(ctx, in.SegmentID)
	if err != nil {
		return domain.SegmentMembership{}, err
	}
	p, err := d.requireActiveProfile(ctx, in.ProfileID)
	if err != nil {
		return domain.SegmentMembership{}, err
	}
	src := in.Source
	if src == "" {
		src = "admin"
	}
	m := domain.SegmentMembership{
		SegmentID: seg.ID,
		ProfileID: in.ProfileID,
		JoinedAt:  d.now(),
		Source:    src,
	}
	if err := m.Validate(); err != nil {
		return domain.SegmentMembership{}, err
	}
	if err := d.Segments.Assign(ctx, m); err != nil {
		return domain.SegmentMembership{}, err
	}
	d.publish(ctx, ports.TopicSegmentEvents, p.ID.String(), map[string]any{
		"eventId": d.newID().String(), "eventType": domain.EventSegmentChanged,
		"occurredAt": d.now(), "tenantId": p.TenantID, "principalId": p.PrincipalID,
		"profileId": p.ID, "segmentId": seg.ID, "segmentName": seg.Name,
		"action": "assigned", "traceId": in.TraceID,
	})
	return m, nil
}

// ListSegmentMembers returns memberships for a segment.
func (d *Deps) ListSegmentMembers(ctx context.Context, segmentID uuid.UUID) ([]domain.SegmentMembership, error) {
	if segmentID == uuid.Nil {
		return nil, fmt.Errorf("%w: segment_id required", domain.ErrInvalidArgument)
	}
	return d.Segments.ListMembers(ctx, segmentID)
}

// ListSegments returns segment definitions for a tenant.
func (d *Deps) ListSegments(ctx context.Context, tenantID uuid.UUID) ([]domain.Segment, error) {
	if tenantID == uuid.Nil {
		return nil, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	return d.Segments.ListSegments(ctx, tenantID)
}

// ListProfileSegments returns segment memberships for a profile.
func (d *Deps) ListProfileSegments(ctx context.Context, profileID uuid.UUID) ([]domain.SegmentMembership, error) {
	if _, err := d.requireActiveProfile(ctx, profileID); err != nil {
		return nil, err
	}
	return d.Segments.ListByProfile(ctx, profileID)
}

// UpsertSegmentDefinition creates or updates a segment definition.
func (d *Deps) UpsertSegmentDefinition(ctx context.Context, s domain.Segment) (domain.Segment, error) {
	if s.ID == uuid.Nil {
		s.ID = d.newID()
	}
	now := d.now()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
	s.UpdatedAt = now
	if err := s.Validate(); err != nil {
		return domain.Segment{}, err
	}
	if err := d.Segments.UpsertSegment(ctx, s); err != nil {
		return domain.Segment{}, err
	}
	return s, nil
}

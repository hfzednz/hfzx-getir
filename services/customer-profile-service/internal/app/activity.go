package app

import (
	"context"
	"fmt"
	"net"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/domain"
)

// RecordActivityInput records a profile-side activity entry.
type RecordActivityInput struct {
	ProfileID    uuid.UUID
	ActorID      *uuid.UUID
	Action       string
	ResourceType string
	ResourceID   *uuid.UUID
	Payload      map[string]any
	IP           net.IP
	UserAgent    string
}

// RecordActivity appends an activity log entry.
func (d *Deps) RecordActivity(ctx context.Context, in RecordActivityInput) (domain.ActivityEntry, error) {
	p, err := d.requireActiveProfile(ctx, in.ProfileID)
	if err != nil {
		return domain.ActivityEntry{}, err
	}
	if in.Action == "" {
		return domain.ActivityEntry{}, fmt.Errorf("%w: action required", domain.ErrInvalidArgument)
	}
	now := d.now()
	e := domain.ActivityEntry{
		ID:           d.newID(),
		ProfileID:    in.ProfileID,
		TenantID:     p.TenantID,
		ActorID:      in.ActorID,
		Action:       in.Action,
		ResourceType: in.ResourceType,
		ResourceID:   in.ResourceID,
		Payload:      in.Payload,
		IP:           in.IP,
		UserAgent:    in.UserAgent,
		OccurredAt:   now,
		CreatedAt:    now,
	}
	if err := e.Validate(); err != nil {
		return domain.ActivityEntry{}, err
	}
	if err := d.Activity.Record(ctx, e); err != nil {
		return domain.ActivityEntry{}, err
	}
	return e, nil
}

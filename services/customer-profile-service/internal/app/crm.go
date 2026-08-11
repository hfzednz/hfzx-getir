package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/domain"
)

// AddNoteInput creates a CRM note.
type AddNoteInput struct {
	ProfileID uuid.UUID
	AuthorID  uuid.UUID
	Body      string
	Pinned    bool
}

// AppendTimelineInput appends a timeline event.
type AppendTimelineInput struct {
	ProfileID uuid.UUID
	Type      string
	Payload   map[string]any
	ActorID   *uuid.UUID
}

// AddNote creates a CRM note on a profile.
func (d *Deps) AddNote(ctx context.Context, in AddNoteInput) (domain.CRMNote, error) {
	p, err := d.requireActiveProfile(ctx, in.ProfileID)
	if err != nil {
		return domain.CRMNote{}, err
	}
	now := d.now()
	n := domain.CRMNote{
		ID:        d.newID(),
		ProfileID: in.ProfileID,
		TenantID:  p.TenantID,
		AuthorID:  in.AuthorID,
		Body:      in.Body,
		Pinned:    in.Pinned,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := n.Validate(); err != nil {
		return domain.CRMNote{}, err
	}
	if err := d.CRM.AddNote(ctx, n); err != nil {
		return domain.CRMNote{}, err
	}
	return n, nil
}

// ListNotes returns CRM notes for a profile.
func (d *Deps) ListNotes(ctx context.Context, profileID uuid.UUID, limit int) ([]domain.CRMNote, error) {
	if _, err := d.requireActiveProfile(ctx, profileID); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 50
	}
	return d.CRM.ListNotes(ctx, profileID, limit)
}

// AppendTimeline appends a timeline event.
func (d *Deps) AppendTimeline(ctx context.Context, in AppendTimelineInput) (domain.TimelineEvent, error) {
	p, err := d.requireActiveProfile(ctx, in.ProfileID)
	if err != nil {
		return domain.TimelineEvent{}, err
	}
	now := d.now()
	e := domain.TimelineEvent{
		ID:         d.newID(),
		ProfileID:  in.ProfileID,
		TenantID:   p.TenantID,
		Type:       in.Type,
		Payload:    in.Payload,
		ActorID:    in.ActorID,
		OccurredAt: now,
		CreatedAt:  now,
	}
	if err := e.Validate(); err != nil {
		return domain.TimelineEvent{}, err
	}
	if err := d.CRM.AppendTimeline(ctx, e); err != nil {
		return domain.TimelineEvent{}, err
	}
	return e, nil
}

// ListTimeline returns timeline events for a profile.
func (d *Deps) ListTimeline(ctx context.Context, profileID uuid.UUID, limit int) ([]domain.TimelineEvent, error) {
	if _, err := d.requireActiveProfile(ctx, profileID); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 50
	}
	return d.CRM.ListTimeline(ctx, profileID, limit)
}

// GetCustomer360 assembles a CRM 360 view.
func (d *Deps) GetCustomer360(ctx context.Context, profileID uuid.UUID) (domain.Customer360, error) {
	if profileID == uuid.Nil {
		return domain.Customer360{}, fmt.Errorf("%w: profile_id required", domain.ErrInvalidArgument)
	}
	p, err := d.Profiles.GetByID(ctx, profileID)
	if err != nil {
		return domain.Customer360{}, err
	}
	view := domain.Customer360{Profile: p}
	if prefs, err := d.Preferences.Get(ctx, profileID); err == nil {
		view.Preferences = &prefs
	}
	if addrs, err := d.Addresses.ListByProfile(ctx, profileID); err == nil {
		view.Addresses = addrs
	}
	if tags, err := d.Tags.List(ctx, profileID); err == nil {
		view.Tags = tags
	}
	if consents, err := d.Consents.List(ctx, profileID); err == nil {
		view.Consents = consents
	}
	if notes, err := d.CRM.ListNotes(ctx, profileID, 20); err == nil {
		view.Notes = notes
	}
	if tl, err := d.CRM.ListTimeline(ctx, profileID, 50); err == nil {
		view.Timeline = tl
	}
	if segs, err := d.Segments.ListByProfile(ctx, profileID); err == nil {
		view.Segments = segs
	}
	if pers, err := d.Personalization.Get(ctx, profileID); err == nil {
		view.Personalization = &pers
	}
	if ai, err := d.AIModels.Get(ctx, profileID); err == nil {
		view.AIModel = &ai
	}
	if h, err := d.Households.GetByOwner(ctx, profileID); err == nil {
		view.Household = &h
	}
	return view, nil
}

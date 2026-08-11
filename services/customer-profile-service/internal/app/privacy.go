package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/app/ports"
	"github.com/nexora/customer-profile-service/internal/domain"
)

// RequestExport creates an export privacy request and emits ExportRequested.
func (d *Deps) RequestExport(ctx context.Context, profileID uuid.UUID, traceID string) (domain.PrivacyRequest, error) {
	p, err := d.requireActiveProfile(ctx, profileID)
	if err != nil {
		return domain.PrivacyRequest{}, err
	}
	now := d.now()
	r := domain.PrivacyRequest{
		ID:        d.newID(),
		ProfileID: profileID,
		TenantID:  p.TenantID,
		Kind:      domain.PrivacyRequestExport,
		Status:    domain.PrivacyStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := r.Validate(); err != nil {
		return domain.PrivacyRequest{}, err
	}
	if err := d.Privacy.Create(ctx, r); err != nil {
		return domain.PrivacyRequest{}, err
	}
	d.publish(ctx, ports.TopicPrivacyEvents, p.ID.String(), map[string]any{
		"eventId": d.newID().String(), "eventType": domain.EventExportRequested,
		"occurredAt": now, "tenantId": p.TenantID, "principalId": p.PrincipalID,
		"profileId": p.ID, "requestId": r.ID, "traceId": traceID,
	})
	return r, nil
}

// RequestDeletion creates a deletion privacy request and emits DeletionRequested.
func (d *Deps) RequestDeletion(ctx context.Context, profileID uuid.UUID, traceID string) (domain.PrivacyRequest, error) {
	p, err := d.requireActiveProfile(ctx, profileID)
	if err != nil {
		return domain.PrivacyRequest{}, err
	}
	now := d.now()
	r := domain.PrivacyRequest{
		ID:        d.newID(),
		ProfileID: profileID,
		TenantID:  p.TenantID,
		Kind:      domain.PrivacyRequestDeletion,
		Status:    domain.PrivacyStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := r.Validate(); err != nil {
		return domain.PrivacyRequest{}, err
	}
	if err := d.Privacy.Create(ctx, r); err != nil {
		return domain.PrivacyRequest{}, err
	}
	d.publish(ctx, ports.TopicPrivacyEvents, p.ID.String(), map[string]any{
		"eventId": d.newID().String(), "eventType": domain.EventDeletionRequested,
		"occurredAt": now, "tenantId": p.TenantID, "principalId": p.PrincipalID,
		"profileId": p.ID, "requestId": r.ID, "traceId": traceID,
	})
	return r, nil
}

// ProcessPrivacyRequest stubs completion of a privacy request.
func (d *Deps) ProcessPrivacyRequest(ctx context.Context, requestID uuid.UUID) (domain.PrivacyRequest, error) {
	if requestID == uuid.Nil {
		return domain.PrivacyRequest{}, fmt.Errorf("%w: request_id required", domain.ErrInvalidArgument)
	}
	r, err := d.Privacy.Get(ctx, requestID)
	if err != nil {
		return domain.PrivacyRequest{}, err
	}
	if r.Status == domain.PrivacyStatusCompleted {
		return r, nil
	}
	now := d.now()
	r.Status = domain.PrivacyStatusProcessing
	r.UpdatedAt = now
	_ = d.Privacy.Update(ctx, r)

	switch r.Kind {
	case domain.PrivacyRequestExport:
		r.PayloadRef = fmt.Sprintf("stub://exports/%s", r.ID.String())
		r.Status = domain.PrivacyStatusCompleted
		r.CompletedAt = &now
		r.UpdatedAt = now
	case domain.PrivacyRequestDelete:
		_ = d.SoftDeleteProfile(ctx, r.ProfileID, "")
		r.Status = domain.PrivacyStatusCompleted
		r.CompletedAt = &now
		r.UpdatedAt = now
	default:
		r.Status = domain.PrivacyStatusFailed
		r.ErrorMessage = "unknown kind"
		r.UpdatedAt = now
	}
	if err := r.Validate(); err != nil {
		return domain.PrivacyRequest{}, err
	}
	if err := d.Privacy.Update(ctx, r); err != nil {
		return domain.PrivacyRequest{}, err
	}
	return r, nil
}

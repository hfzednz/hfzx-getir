package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/app/ports"
	"github.com/nexora/customer-profile-service/internal/domain"
)

// Deps aggregates application ports for use cases.
type Deps struct {
	Profiles        ports.ProfileRepository
	Addresses       ports.AddressRepository
	Preferences     ports.PreferencesRepository
	Tags            ports.TagRepository
	Households      ports.HouseholdRepository
	Consents        ports.ConsentRepository
	CRM             ports.CRMRepository
	Segments        ports.SegmentRepository
	Personalization ports.PersonalizationRepository
	AIModels        ports.AIModelRepository
	Privacy         ports.PrivacyRepository
	Activity        ports.ActivityRepository
	Events          ports.EventPublisher
	Media           ports.MediaStore
	Zones           ports.ZoneValidator
	Search          ports.ProfileSearchIndexer
	Clock           ports.Clock
	IDs             ports.IDGen
}

func (d *Deps) now() time.Time {
	if d.Clock != nil {
		return d.Clock.Now().UTC()
	}
	return time.Now().UTC()
}

func (d *Deps) newID() uuid.UUID {
	if d.IDs != nil {
		return d.IDs.New()
	}
	return uuid.New()
}

func (d *Deps) publish(ctx context.Context, topic, key string, payload any) {
	if d.Events == nil {
		return
	}
	_ = d.Events.Publish(ctx, topic, key, payload)
}

func (d *Deps) publishLifecycle(ctx context.Context, eventType string, p domain.CustomerProfile, extra map[string]any, traceID string) {
	payload := map[string]any{
		"eventId":     d.newID().String(),
		"eventType":   eventType,
		"occurredAt":  d.now(),
		"tenantId":    p.TenantID,
		"principalId": p.PrincipalID,
		"profileId":   p.ID,
		"traceId":     traceID,
	}
	for k, v := range extra {
		payload[k] = v
	}
	d.publish(ctx, ports.TopicProfileLifecycle, p.ID.String(), payload)
}

func (d *Deps) indexProfile(ctx context.Context, p domain.CustomerProfile) {
	if d.Search == nil {
		return
	}
	_ = d.Search.IndexProfile(ctx, p)
}

func (d *Deps) deleteProfileIndex(ctx context.Context, tenantID, profileID uuid.UUID) {
	if d.Search == nil {
		return
	}
	_ = d.Search.DeleteProfile(ctx, tenantID, profileID)
}

func (d *Deps) requireActiveProfile(ctx context.Context, profileID uuid.UUID) (domain.CustomerProfile, error) {
	if profileID == uuid.Nil {
		return domain.CustomerProfile{}, fmt.Errorf("%w: profile_id required", domain.ErrInvalidArgument)
	}
	p, err := d.Profiles.GetByID(ctx, profileID)
	if err != nil {
		return domain.CustomerProfile{}, err
	}
	switch p.Status {
	case domain.ProfileStatusDeleted:
		return domain.CustomerProfile{}, domain.ErrProfileDeleted
	case domain.ProfileStatusMerged:
		return domain.CustomerProfile{}, domain.ErrProfileMerged
	}
	if !p.IsActive() {
		return domain.CustomerProfile{}, domain.ErrProfileDeleted
	}
	return p, nil
}

// SystemClock is a real-time Clock.
type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }

// UUIDGen generates random UUIDs.
type UUIDGen struct{}

func (UUIDGen) New() uuid.UUID { return uuid.New() }

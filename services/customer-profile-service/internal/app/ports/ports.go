// Package ports defines application-layer dependency interfaces (hexagonal ports).
package ports

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/domain"
)

// Clock abstracts time for deterministic tests.
type Clock interface {
	Now() time.Time
}

// IDGen abstracts UUID generation.
type IDGen interface {
	New() uuid.UUID
}

// EventPublisher publishes domain events (Kafka adapters).
type EventPublisher interface {
	Publish(ctx context.Context, topic string, key string, payload any) error
}

// MediaStore stores avatar binaries / refs (media-service port).
type MediaStore interface {
	PutAvatar(ctx context.Context, tenantID, profileID uuid.UUID, filename string, contentType string, r io.Reader) (url string, version int, err error)
	DeleteAvatar(ctx context.Context, tenantID, profileID uuid.UUID) error
}

// ZoneValidator checks whether lat/lng falls in a delivery zone (geofence port).
type ZoneValidator interface {
	ValidateZone(ctx context.Context, tenantID uuid.UUID, lat, lng float64) (zoneID string, ok bool, err error)
}

// ProfileSearchIndexer indexes profiles for admin/search (OpenSearch adapter).
type ProfileSearchIndexer interface {
	IndexProfile(ctx context.Context, p domain.CustomerProfile) error
	DeleteProfile(ctx context.Context, tenantID, profileID uuid.UUID) error
}

// Kafka topic aliases (owned by domain package).
const (
	TopicProfileLifecycle = domain.TopicProfileLifecycle
	TopicAddressEvents    = domain.TopicAddressEvents
	TopicPreferenceEvents = domain.TopicPreferenceEvents
	TopicMediaEvents      = domain.TopicMediaEvents
	TopicConsentEvents    = domain.TopicConsentEvents
	TopicSegmentEvents    = domain.TopicSegmentEvents
	TopicPrivacyEvents    = domain.TopicPrivacyEvents
)

// ---------------------------------------------------------------------------
// Repositories
// ---------------------------------------------------------------------------

// ProfileRepository persists customer profiles.
type ProfileRepository interface {
	Create(ctx context.Context, p domain.CustomerProfile) error
	Update(ctx context.Context, p domain.CustomerProfile) error
	GetByID(ctx context.Context, id uuid.UUID) (domain.CustomerProfile, error)
	GetByPrincipalID(ctx context.Context, tenantID, principalID uuid.UUID) (domain.CustomerProfile, error)
	SoftDelete(ctx context.Context, id uuid.UUID, at time.Time) error
	Search(ctx context.Context, tenantID uuid.UUID, query string, limit int) ([]domain.CustomerProfile, error)
	FindDuplicates(ctx context.Context, tenantID uuid.UUID, displayName, fullName string) ([]domain.CustomerProfile, error)
}

// AddressRepository persists addresses.
type AddressRepository interface {
	Create(ctx context.Context, a domain.Address) error
	Update(ctx context.Context, a domain.Address) error
	GetByID(ctx context.Context, id uuid.UUID) (domain.Address, error)
	ListByProfile(ctx context.Context, profileID uuid.UUID) ([]domain.Address, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ClearDefault(ctx context.Context, profileID uuid.UUID) error
}

// PreferencesRepository persists preferences.
type PreferencesRepository interface {
	Get(ctx context.Context, profileID uuid.UUID) (domain.Preferences, error)
	Upsert(ctx context.Context, p domain.Preferences) error
}

// TagRepository persists tag definitions and profile assignments.
type TagRepository interface {
	UpsertTag(ctx context.Context, t domain.Tag) error
	GetTag(ctx context.Context, id uuid.UUID) (domain.Tag, error)
	Add(ctx context.Context, pt domain.ProfileTag) error
	Remove(ctx context.Context, profileID, tagID uuid.UUID) error
	List(ctx context.Context, profileID uuid.UUID) ([]domain.ProfileTag, error)
}

// HouseholdRepository persists households and members.
type HouseholdRepository interface {
	Create(ctx context.Context, h domain.Household) error
	Update(ctx context.Context, h domain.Household) error
	Get(ctx context.Context, id uuid.UUID) (domain.Household, error)
	GetByOwner(ctx context.Context, ownerProfileID uuid.UUID) (domain.Household, error)
	AddMember(ctx context.Context, m domain.HouseholdMember) error
	UpdateMember(ctx context.Context, m domain.HouseholdMember) error
	ListMembers(ctx context.Context, householdID uuid.UUID) ([]domain.HouseholdMember, error)
	GetMember(ctx context.Context, householdID, profileID uuid.UUID) (domain.HouseholdMember, error)
}

// ConsentRepository persists consents.
type ConsentRepository interface {
	Upsert(ctx context.Context, c domain.Consent) error
	List(ctx context.Context, profileID uuid.UUID) ([]domain.Consent, error)
	Get(ctx context.Context, profileID uuid.UUID, channel domain.ConsentChannel) (domain.Consent, error)
}

// CRMRepository persists CRM notes and timeline.
type CRMRepository interface {
	AddNote(ctx context.Context, n domain.CRMNote) error
	ListNotes(ctx context.Context, profileID uuid.UUID, limit int) ([]domain.CRMNote, error)
	AppendTimeline(ctx context.Context, e domain.TimelineEvent) error
	ListTimeline(ctx context.Context, profileID uuid.UUID, limit int) ([]domain.TimelineEvent, error)
}

// SegmentRepository persists segments and memberships.
type SegmentRepository interface {
	GetSegment(ctx context.Context, id uuid.UUID) (domain.Segment, error)
	ListSegments(ctx context.Context, tenantID uuid.UUID) ([]domain.Segment, error)
	UpsertSegment(ctx context.Context, s domain.Segment) error
	Assign(ctx context.Context, m domain.SegmentMembership) error
	RemoveMembership(ctx context.Context, segmentID, profileID uuid.UUID) error
	ListMembers(ctx context.Context, segmentID uuid.UUID) ([]domain.SegmentMembership, error)
	ListByProfile(ctx context.Context, profileID uuid.UUID) ([]domain.SegmentMembership, error)
}

// PersonalizationRepository persists personalization profiles.
type PersonalizationRepository interface {
	Get(ctx context.Context, profileID uuid.UUID) (domain.Personalization, error)
	Upsert(ctx context.Context, p domain.Personalization) error
}

// AIModelRepository persists AI customer model scores.
type AIModelRepository interface {
	Get(ctx context.Context, profileID uuid.UUID) (domain.AICustomerModel, error)
	Upsert(ctx context.Context, m domain.AICustomerModel) error
}

// PrivacyRepository persists privacy requests.
type PrivacyRepository interface {
	Create(ctx context.Context, r domain.PrivacyRequest) error
	Update(ctx context.Context, r domain.PrivacyRequest) error
	Get(ctx context.Context, id uuid.UUID) (domain.PrivacyRequest, error)
	ListPending(ctx context.Context, limit int) ([]domain.PrivacyRequest, error)
}

// ActivityRepository persists profile-side activity.
type ActivityRepository interface {
	Record(ctx context.Context, e domain.ActivityEntry) error
	List(ctx context.Context, profileID uuid.UUID, limit int) ([]domain.ActivityEntry, error)
}

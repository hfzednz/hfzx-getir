// Package memory provides in-memory port implementations for unit tests.
package memory

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/app/ports"
	"github.com/nexora/customer-profile-service/internal/domain"
)

// Store is a shared in-memory database for all repositories.
type Store struct {
	mu sync.RWMutex

	Profiles        map[uuid.UUID]domain.CustomerProfile
	Addresses       map[uuid.UUID]domain.Address
	Preferences     map[uuid.UUID]domain.Preferences
	Tags            map[uuid.UUID]domain.Tag
	ProfileTags     map[string]domain.ProfileTag // profileID|tagID
	Households      map[uuid.UUID]domain.Household
	Members         map[uuid.UUID]domain.HouseholdMember
	Consents        map[uuid.UUID]domain.Consent
	Notes           map[uuid.UUID]domain.CRMNote
	Timeline        map[uuid.UUID]domain.TimelineEvent
	Segments        map[uuid.UUID]domain.Segment
	Memberships     map[string]domain.SegmentMembership // segmentID|profileID
	Personalization map[uuid.UUID]domain.Personalization
	AIModels        map[uuid.UUID]domain.AICustomerModel
	Privacy         map[uuid.UUID]domain.PrivacyRequest
	Activity        map[uuid.UUID]domain.ActivityEntry
}

// NewStore returns an empty in-memory store.
func NewStore() *Store {
	return &Store{
		Profiles:        make(map[uuid.UUID]domain.CustomerProfile),
		Addresses:       make(map[uuid.UUID]domain.Address),
		Preferences:     make(map[uuid.UUID]domain.Preferences),
		Tags:            make(map[uuid.UUID]domain.Tag),
		ProfileTags:     make(map[string]domain.ProfileTag),
		Households:      make(map[uuid.UUID]domain.Household),
		Members:         make(map[uuid.UUID]domain.HouseholdMember),
		Consents:        make(map[uuid.UUID]domain.Consent),
		Notes:           make(map[uuid.UUID]domain.CRMNote),
		Timeline:        make(map[uuid.UUID]domain.TimelineEvent),
		Segments:        make(map[uuid.UUID]domain.Segment),
		Memberships:     make(map[string]domain.SegmentMembership),
		Personalization: make(map[uuid.UUID]domain.Personalization),
		AIModels:        make(map[uuid.UUID]domain.AICustomerModel),
		Privacy:         make(map[uuid.UUID]domain.PrivacyRequest),
		Activity:        make(map[uuid.UUID]domain.ActivityEntry),
	}
}

func ptKey(profileID, tagID uuid.UUID) string {
	return profileID.String() + "|" + tagID.String()
}

func memKey(segmentID, profileID uuid.UUID) string {
	return segmentID.String() + "|" + profileID.String()
}

// ---------------------------------------------------------------------------
// ProfileRepository
// ---------------------------------------------------------------------------

type ProfileRepo struct{ S *Store }

func (r *ProfileRepo) Create(_ context.Context, p domain.CustomerProfile) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Profiles[p.ID]; ok {
		return domain.ErrAlreadyExists
	}
	for _, existing := range r.S.Profiles {
		if existing.TenantID == p.TenantID && existing.PrincipalID == p.PrincipalID && existing.Status == domain.ProfileStatusActive {
			return domain.ErrAlreadyExists
		}
	}
	r.S.Profiles[p.ID] = p
	return nil
}

func (r *ProfileRepo) Update(_ context.Context, p domain.CustomerProfile) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Profiles[p.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.Profiles[p.ID] = p
	return nil
}

func (r *ProfileRepo) GetByID(_ context.Context, id uuid.UUID) (domain.CustomerProfile, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	p, ok := r.S.Profiles[id]
	if !ok {
		return domain.CustomerProfile{}, domain.ErrNotFound
	}
	return p, nil
}

func (r *ProfileRepo) GetByPrincipalID(_ context.Context, tenantID, principalID uuid.UUID) (domain.CustomerProfile, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	for _, p := range r.S.Profiles {
		if p.TenantID == tenantID && p.PrincipalID == principalID {
			return p, nil
		}
	}
	return domain.CustomerProfile{}, domain.ErrNotFound
}

func (r *ProfileRepo) SoftDelete(_ context.Context, id uuid.UUID, at time.Time) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	p, ok := r.S.Profiles[id]
	if !ok {
		return domain.ErrNotFound
	}
	p.Status = domain.ProfileStatusDeleted
	p.DeletedAt = &at
	p.UpdatedAt = at
	r.S.Profiles[id] = p
	return nil
}

func (r *ProfileRepo) Search(_ context.Context, tenantID uuid.UUID, query string, limit int) ([]domain.CustomerProfile, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	q := strings.ToLower(query)
	out := make([]domain.CustomerProfile, 0)
	for _, p := range r.S.Profiles {
		if p.TenantID != tenantID || p.Status != domain.ProfileStatusActive {
			continue
		}
		if q == "" ||
			strings.Contains(strings.ToLower(p.DisplayName), q) ||
			strings.Contains(strings.ToLower(p.FullName), q) ||
			strings.Contains(strings.ToLower(p.Nickname), q) ||
			strings.Contains(strings.ToLower(p.City), q) ||
			strings.Contains(p.ID.String(), q) {
			out = append(out, p)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (r *ProfileRepo) FindDuplicates(_ context.Context, tenantID uuid.UUID, displayName, fullName string) ([]domain.CustomerProfile, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	dn := strings.ToLower(strings.TrimSpace(displayName))
	fn := strings.ToLower(strings.TrimSpace(fullName))
	out := make([]domain.CustomerProfile, 0)
	for _, p := range r.S.Profiles {
		if p.TenantID != tenantID || p.Status != domain.ProfileStatusActive {
			continue
		}
		if dn != "" && strings.ToLower(p.DisplayName) == dn {
			out = append(out, p)
			continue
		}
		if fn != "" && strings.ToLower(p.FullName) == fn {
			out = append(out, p)
		}
	}
	return out, nil
}

var _ ports.ProfileRepository = (*ProfileRepo)(nil)

// ---------------------------------------------------------------------------
// AddressRepository
// ---------------------------------------------------------------------------

type AddressRepo struct{ S *Store }

func (r *AddressRepo) Create(_ context.Context, a domain.Address) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Addresses[a.ID] = a
	return nil
}

func (r *AddressRepo) Update(_ context.Context, a domain.Address) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Addresses[a.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.Addresses[a.ID] = a
	return nil
}

func (r *AddressRepo) GetByID(_ context.Context, id uuid.UUID) (domain.Address, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	a, ok := r.S.Addresses[id]
	if !ok {
		return domain.Address{}, domain.ErrNotFound
	}
	return a, nil
}

func (r *AddressRepo) ListByProfile(_ context.Context, profileID uuid.UUID) ([]domain.Address, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.Address, 0)
	for _, a := range r.S.Addresses {
		if a.ProfileID == profileID {
			out = append(out, a)
		}
	}
	return out, nil
}

func (r *AddressRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	a, ok := r.S.Addresses[id]
	if !ok {
		return domain.ErrNotFound
	}
	now := time.Now().UTC()
	a.DeletedAt = &now
	a.IsDefault = false
	r.S.Addresses[id] = a
	return nil
}

func (r *AddressRepo) ClearDefault(_ context.Context, profileID uuid.UUID) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	for id, a := range r.S.Addresses {
		if a.ProfileID == profileID && a.IsDefault {
			a.IsDefault = false
			r.S.Addresses[id] = a
		}
	}
	return nil
}

var _ ports.AddressRepository = (*AddressRepo)(nil)

// ---------------------------------------------------------------------------
// PreferencesRepository
// ---------------------------------------------------------------------------

type PreferencesRepo struct{ S *Store }

func (r *PreferencesRepo) Get(_ context.Context, profileID uuid.UUID) (domain.Preferences, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	p, ok := r.S.Preferences[profileID]
	if !ok {
		return domain.Preferences{}, domain.ErrNotFound
	}
	return p, nil
}

func (r *PreferencesRepo) Upsert(_ context.Context, p domain.Preferences) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Preferences[p.ProfileID] = p
	return nil
}

var _ ports.PreferencesRepository = (*PreferencesRepo)(nil)

// ---------------------------------------------------------------------------
// TagRepository
// ---------------------------------------------------------------------------

type TagRepo struct{ S *Store }

func (r *TagRepo) UpsertTag(_ context.Context, t domain.Tag) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Tags[t.ID] = t
	return nil
}

func (r *TagRepo) GetTag(_ context.Context, id uuid.UUID) (domain.Tag, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	t, ok := r.S.Tags[id]
	if !ok {
		return domain.Tag{}, domain.ErrNotFound
	}
	return t, nil
}

func (r *TagRepo) Add(_ context.Context, pt domain.ProfileTag) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	k := ptKey(pt.ProfileID, pt.TagID)
	if _, ok := r.S.ProfileTags[k]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.ProfileTags[k] = pt
	return nil
}

func (r *TagRepo) Remove(_ context.Context, profileID, tagID uuid.UUID) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	k := ptKey(profileID, tagID)
	if _, ok := r.S.ProfileTags[k]; !ok {
		return domain.ErrNotFound
	}
	delete(r.S.ProfileTags, k)
	return nil
}

func (r *TagRepo) List(_ context.Context, profileID uuid.UUID) ([]domain.ProfileTag, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.ProfileTag, 0)
	for _, pt := range r.S.ProfileTags {
		if pt.ProfileID == profileID {
			out = append(out, pt)
		}
	}
	return out, nil
}

var _ ports.TagRepository = (*TagRepo)(nil)

// ---------------------------------------------------------------------------
// HouseholdRepository
// ---------------------------------------------------------------------------

type HouseholdRepo struct{ S *Store }

func (r *HouseholdRepo) Create(_ context.Context, h domain.Household) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Households[h.ID] = h
	return nil
}

func (r *HouseholdRepo) Update(_ context.Context, h domain.Household) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Households[h.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.Households[h.ID] = h
	return nil
}

func (r *HouseholdRepo) Get(_ context.Context, id uuid.UUID) (domain.Household, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	h, ok := r.S.Households[id]
	if !ok {
		return domain.Household{}, domain.ErrNotFound
	}
	return h, nil
}

func (r *HouseholdRepo) GetByOwner(_ context.Context, ownerProfileID uuid.UUID) (domain.Household, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	for _, h := range r.S.Households {
		if h.OwnerProfileID == ownerProfileID && h.DeletedAt == nil {
			return h, nil
		}
	}
	return domain.Household{}, domain.ErrNotFound
}

func (r *HouseholdRepo) AddMember(_ context.Context, m domain.HouseholdMember) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	for _, existing := range r.S.Members {
		if existing.HouseholdID == m.HouseholdID && existing.ProfileID == m.ProfileID && existing.LeftAt == nil {
			return domain.ErrAlreadyExists
		}
	}
	r.S.Members[m.ID] = m
	return nil
}

func (r *HouseholdRepo) UpdateMember(_ context.Context, m domain.HouseholdMember) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Members[m.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.Members[m.ID] = m
	return nil
}

func (r *HouseholdRepo) ListMembers(_ context.Context, householdID uuid.UUID) ([]domain.HouseholdMember, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.HouseholdMember, 0)
	for _, m := range r.S.Members {
		if m.HouseholdID == householdID {
			out = append(out, m)
		}
	}
	return out, nil
}

func (r *HouseholdRepo) GetMember(_ context.Context, householdID, profileID uuid.UUID) (domain.HouseholdMember, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	for _, m := range r.S.Members {
		if m.HouseholdID == householdID && m.ProfileID == profileID {
			return m, nil
		}
	}
	return domain.HouseholdMember{}, domain.ErrNotFound
}

var _ ports.HouseholdRepository = (*HouseholdRepo)(nil)

// ---------------------------------------------------------------------------
// ConsentRepository
// ---------------------------------------------------------------------------

type ConsentRepo struct{ S *Store }

func (r *ConsentRepo) Upsert(_ context.Context, c domain.Consent) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	for id, existing := range r.S.Consents {
		if existing.ProfileID == c.ProfileID && existing.Channel == c.Channel {
			c.ID = existing.ID
			c.CreatedAt = existing.CreatedAt
			delete(r.S.Consents, id)
			break
		}
	}
	r.S.Consents[c.ID] = c
	return nil
}

func (r *ConsentRepo) List(_ context.Context, profileID uuid.UUID) ([]domain.Consent, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.Consent, 0)
	for _, c := range r.S.Consents {
		if c.ProfileID == profileID {
			out = append(out, c)
		}
	}
	return out, nil
}

func (r *ConsentRepo) Get(_ context.Context, profileID uuid.UUID, channel domain.ConsentChannel) (domain.Consent, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	for _, c := range r.S.Consents {
		if c.ProfileID == profileID && c.Channel == channel {
			return c, nil
		}
	}
	return domain.Consent{}, domain.ErrNotFound
}

var _ ports.ConsentRepository = (*ConsentRepo)(nil)

// ---------------------------------------------------------------------------
// CRMRepository
// ---------------------------------------------------------------------------

type CRMRepo struct{ S *Store }

func (r *CRMRepo) AddNote(_ context.Context, n domain.CRMNote) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Notes[n.ID] = n
	return nil
}

func (r *CRMRepo) ListNotes(_ context.Context, profileID uuid.UUID, limit int) ([]domain.CRMNote, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.CRMNote, 0)
	for _, n := range r.S.Notes {
		if n.ProfileID == profileID && n.DeletedAt == nil {
			out = append(out, n)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (r *CRMRepo) AppendTimeline(_ context.Context, e domain.TimelineEvent) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Timeline[e.ID] = e
	return nil
}

func (r *CRMRepo) ListTimeline(_ context.Context, profileID uuid.UUID, limit int) ([]domain.TimelineEvent, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.TimelineEvent, 0)
	for _, e := range r.S.Timeline {
		if e.ProfileID == profileID {
			out = append(out, e)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

var _ ports.CRMRepository = (*CRMRepo)(nil)

// ---------------------------------------------------------------------------
// SegmentRepository
// ---------------------------------------------------------------------------

type SegmentRepo struct{ S *Store }

func (r *SegmentRepo) GetSegment(_ context.Context, id uuid.UUID) (domain.Segment, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	s, ok := r.S.Segments[id]
	if !ok {
		return domain.Segment{}, domain.ErrNotFound
	}
	return s, nil
}

func (r *SegmentRepo) ListSegments(_ context.Context, tenantID uuid.UUID) ([]domain.Segment, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.Segment, 0)
	for _, s := range r.S.Segments {
		if s.TenantID == tenantID {
			out = append(out, s)
		}
	}
	return out, nil
}

func (r *SegmentRepo) UpsertSegment(_ context.Context, s domain.Segment) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Segments[s.ID] = s
	return nil
}

func (r *SegmentRepo) Assign(_ context.Context, m domain.SegmentMembership) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Memberships[memKey(m.SegmentID, m.ProfileID)] = m
	return nil
}

func (r *SegmentRepo) RemoveMembership(_ context.Context, segmentID, profileID uuid.UUID) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	delete(r.S.Memberships, memKey(segmentID, profileID))
	return nil
}

func (r *SegmentRepo) ListMembers(_ context.Context, segmentID uuid.UUID) ([]domain.SegmentMembership, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.SegmentMembership, 0)
	for _, m := range r.S.Memberships {
		if m.SegmentID == segmentID {
			out = append(out, m)
		}
	}
	return out, nil
}

func (r *SegmentRepo) ListByProfile(_ context.Context, profileID uuid.UUID) ([]domain.SegmentMembership, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.SegmentMembership, 0)
	for _, m := range r.S.Memberships {
		if m.ProfileID == profileID {
			out = append(out, m)
		}
	}
	return out, nil
}

var _ ports.SegmentRepository = (*SegmentRepo)(nil)

// ---------------------------------------------------------------------------
// PersonalizationRepository
// ---------------------------------------------------------------------------

type PersonalizationRepo struct{ S *Store }

func (r *PersonalizationRepo) Get(_ context.Context, profileID uuid.UUID) (domain.Personalization, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	p, ok := r.S.Personalization[profileID]
	if !ok {
		return domain.Personalization{}, domain.ErrNotFound
	}
	return p, nil
}

func (r *PersonalizationRepo) Upsert(_ context.Context, p domain.Personalization) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Personalization[p.ProfileID] = p
	return nil
}

var _ ports.PersonalizationRepository = (*PersonalizationRepo)(nil)

// ---------------------------------------------------------------------------
// AIModelRepository
// ---------------------------------------------------------------------------

type AIModelRepo struct{ S *Store }

func (r *AIModelRepo) Get(_ context.Context, profileID uuid.UUID) (domain.AICustomerModel, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	m, ok := r.S.AIModels[profileID]
	if !ok {
		return domain.AICustomerModel{}, domain.ErrNotFound
	}
	return m, nil
}

func (r *AIModelRepo) Upsert(_ context.Context, m domain.AICustomerModel) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.AIModels[m.ProfileID] = m
	return nil
}

var _ ports.AIModelRepository = (*AIModelRepo)(nil)

// ---------------------------------------------------------------------------
// PrivacyRepository
// ---------------------------------------------------------------------------

type PrivacyRepo struct{ S *Store }

func (r *PrivacyRepo) Create(_ context.Context, req domain.PrivacyRequest) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Privacy[req.ID] = req
	return nil
}

func (r *PrivacyRepo) Update(_ context.Context, req domain.PrivacyRequest) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Privacy[req.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.Privacy[req.ID] = req
	return nil
}

func (r *PrivacyRepo) Get(_ context.Context, id uuid.UUID) (domain.PrivacyRequest, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	req, ok := r.S.Privacy[id]
	if !ok {
		return domain.PrivacyRequest{}, domain.ErrNotFound
	}
	return req, nil
}

func (r *PrivacyRepo) ListPending(_ context.Context, limit int) ([]domain.PrivacyRequest, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.PrivacyRequest, 0)
	for _, req := range r.S.Privacy {
		if req.Status == domain.PrivacyStatusPending {
			out = append(out, req)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

var _ ports.PrivacyRepository = (*PrivacyRepo)(nil)

// ---------------------------------------------------------------------------
// ActivityRepository
// ---------------------------------------------------------------------------

type ActivityRepo struct{ S *Store }

func (r *ActivityRepo) Record(_ context.Context, e domain.ActivityEntry) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Activity[e.ID] = e
	return nil
}

func (r *ActivityRepo) List(_ context.Context, profileID uuid.UUID, limit int) ([]domain.ActivityEntry, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.ActivityEntry, 0)
	for _, e := range r.S.Activity {
		if e.ProfileID == profileID {
			out = append(out, e)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

var _ ports.ActivityRepository = (*ActivityRepo)(nil)

// ---------------------------------------------------------------------------
// Small fakes
// ---------------------------------------------------------------------------

type Clock struct{ T time.Time }

func (c *Clock) Now() time.Time {
	if c.T.IsZero() {
		return time.Now().UTC()
	}
	return c.T.UTC()
}

type IDGen struct{}

func (IDGen) New() uuid.UUID { return uuid.New() }

type EventPublisher struct {
	mu     sync.Mutex
	Events []PublishedEvent
}

type PublishedEvent struct {
	Topic   string
	Key     string
	Payload any
}

func (e *EventPublisher) Publish(_ context.Context, topic, key string, payload any) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Events = append(e.Events, PublishedEvent{Topic: topic, Key: key, Payload: payload})
	return nil
}

func (e *EventPublisher) OfType(eventType string) []PublishedEvent {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]PublishedEvent, 0)
	for _, ev := range e.Events {
		if m, ok := ev.Payload.(map[string]any); ok {
			if m["eventType"] == eventType {
				out = append(out, ev)
			}
		}
	}
	return out
}

type MediaStore struct {
	URLs map[string]string
	Ver  map[string]int
}

func NewMediaStore() *MediaStore {
	return &MediaStore{URLs: make(map[string]string), Ver: make(map[string]int)}
}

func (m *MediaStore) PutAvatar(_ context.Context, tenantID, profileID uuid.UUID, filename, contentType string, r io.Reader) (string, int, error) {
	key := profileID.String()
	m.Ver[key]++
	url := fmt.Sprintf("mem://%s/%s/%s", tenantID, profileID, filename)
	if r != nil {
		_, _ = io.Copy(io.Discard, r)
	}
	_ = contentType
	m.URLs[key] = url
	return url, m.Ver[key], nil
}

func (m *MediaStore) DeleteAvatar(_ context.Context, _, profileID uuid.UUID) error {
	key := profileID.String()
	delete(m.URLs, key)
	delete(m.Ver, key)
	return nil
}

type ZoneValidator struct {
	OK     bool
	ZoneID string
	Err    error
}

func (z *ZoneValidator) ValidateZone(_ context.Context, _ uuid.UUID, _, _ float64) (string, bool, error) {
	if z.Err != nil {
		return "", false, z.Err
	}
	id := z.ZoneID
	if id == "" {
		id = "zone-1"
	}
	return id, z.OK, nil
}

var _ ports.Clock = (*Clock)(nil)
var _ ports.IDGen = (IDGen{})
var _ ports.EventPublisher = (*EventPublisher)(nil)
var _ ports.MediaStore = (*MediaStore)(nil)
var _ ports.ZoneValidator = (*ZoneValidator)(nil)
